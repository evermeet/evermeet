package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/evermeet/evermeet/internal/calendar"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/go-chi/chi/v5"
)

func (s *Server) handleListCalendars(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	did := authDID(r)

	type calendarItem struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Avatar      string `json:"avatar,omitempty"`
		BackdropURL string `json:"backdrop_url,omitempty"`
		Website     string `json:"website,omitempty"`
		Subscribers int    `json:"subscribers"`
	}

	toItem := func(st *store.CalendarState) (*calendarItem, error) {
		var ms calendar.MutableState
		if err := json.Unmarshal([]byte(st.Payload), &ms); err != nil {
			return nil, err
		}
		subs, _ := s.db.CountCalendarSubscribers(ctx, ms.ID)
		return &calendarItem{
			ID:          ms.ID,
			Name:        ms.Name,
			Description: ms.Description,
			Avatar:      ms.Avatar,
			BackdropURL: ms.BackdropURL,
			Website:     ms.Website,
			Subscribers: subs,
		}, nil
	}

	owned, err := s.db.ListOwnedCalendars(ctx, did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "list owned calendars")
		return
	}

	subscribed, err := s.db.ListSubscribedCalendars(ctx, did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "list subscribed calendars")
		return
	}

	ownedItems := make([]*calendarItem, 0, len(owned))
	for _, st := range owned {
		item, err := toItem(st)
		if err != nil {
			continue
		}
		ownedItems = append(ownedItems, item)
	}

	subscribedItems := make([]*calendarItem, 0, len(subscribed))
	for _, st := range subscribed {
		item, err := toItem(st)
		if err != nil {
			continue
		}
		subscribedItems = append(subscribedItems, item)
	}

	jsonOK(w, map[string]any{
		"owned":      ownedItems,
		"subscribed": subscribedItems,
	})
}

func (s *Server) handleDiscoverCalendars(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	did := authDID(r)

	type calendarItem struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Avatar      string `json:"avatar,omitempty"`
		BackdropURL string `json:"backdrop_url,omitempty"`
		Website     string `json:"website,omitempty"`
		Subscribers int    `json:"subscribers"`
		Subscribed  bool   `json:"subscribed"`
	}

	calendars, err := s.db.ListCurrentCalendars(ctx)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "list calendars")
		return
	}

	items := make([]*calendarItem, 0, len(calendars))
	for _, st := range calendars {
		var ms calendar.MutableState
		if err := json.Unmarshal([]byte(st.Payload), &ms); err != nil {
			continue
		}

		subs, _ := s.db.CountCalendarSubscribers(ctx, ms.ID)
		subscribed := false
		if did != "" {
			subscribed, _ = s.db.IsSubscribed(ctx, ms.ID, did)
		}

		items = append(items, &calendarItem{
			ID:          ms.ID,
			Name:        ms.Name,
			Description: ms.Description,
			Avatar:      ms.Avatar,
			BackdropURL: ms.BackdropURL,
			Website:     ms.Website,
			Subscribers: subs,
			Subscribed:  subscribed,
		})
	}

	jsonOK(w, map[string]any{
		"calendars": items,
	})
}

func (s *Server) handleCreateCalendar(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	priv := authPrivKey(r)
	ctx := r.Context()

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Avatar      string `json:"avatar"`
		BackdropURL string `json:"backdrop_url"`
		Website     string `json:"website"`
		Owners      []string `json:"owners"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" {
		jsonErr(w, http.StatusBadRequest, "name required")
		return
	}

	f := calendar.Fields{
		Name:        req.Name,
		Description: req.Description,
		Avatar:      req.Avatar,
		BackdropURL: req.BackdropURL,
		Website:     req.Website,
	}

	founding, calID, state, stateHash, err := calendar.New(did, priv, s.homeHost(), f)
	if err != nil {
		s.log.Printf("create calendar: %v", err)
		jsonErr(w, http.StatusInternalServerError, "create calendar failed")
		return
	}

	foundingPayload := mustJSON(founding)
	statePayload := mustJSON(state)

	if err := s.db.InsertCalendarFounding(ctx, &store.CalendarFounding{
		ID:      calID,
		Payload: string(foundingPayload),
	}); err != nil {
		s.log.Printf("insert calendar founding: %v", err)
		jsonErr(w, http.StatusInternalServerError, "store failed")
		return
	}

	if err := s.db.AppendCalendarState(ctx, &store.CalendarState{
		Hash:      stateHash,
		ID:        calID,
		Payload:   string(statePayload),
		CreatedAt: time.Now(),
	}); err != nil {
		s.log.Printf("append calendar state: %v", err)
		jsonErr(w, http.StatusInternalServerError, "store failed")
		return
	}

	if err := s.db.InsertCalendarOwner(ctx, calID, did); err != nil {
		s.log.Printf("insert calendar owner: %v", err)
	}

	jsonOK(w, map[string]any{"id": calID, "state": state})
}

func (s *Server) handleGetCalendar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	did := authDID(r)

	current, err := s.db.GetCurrentCalendarState(ctx, id)
	if err != nil || current == nil {
		jsonErr(w, http.StatusNotFound, "calendar not found")
		return
	}

	var ms calendar.MutableState
	if err := json.Unmarshal([]byte(current.Payload), &ms); err != nil {
		jsonErr(w, http.StatusInternalServerError, "parse state")
		return
	}

	subs, _ := s.db.CountCalendarSubscribers(ctx, id)

	subscribed := false
	if did != "" {
		subscribed, _ = s.db.IsSubscribed(ctx, id, did)
	}

	events, err := s.db.ListEventsByCalendar(ctx, id)
	if err != nil {
		s.log.Printf("list events for calendar %s: %v", id, err)
	}

	type eventSummary struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		StartsAt string `json:"starts_at"`
		EndsAt   string `json:"ends_at,omitempty"`
		Location any    `json:"location,omitempty"`
		CoverURL string `json:"cover_url,omitempty"`
		Hosts    []string `json:"hosts,omitempty"`
	}

	eventList := make([]eventSummary, 0, len(events))
	for _, ev := range events {
		var raw map[string]any
		if err := json.Unmarshal([]byte(ev.Payload), &raw); err != nil {
			continue
		}
		es := eventSummary{
			ID:       ev.ID,
			Location: raw["location"],
		}
		if v, ok := raw["title"].(string); ok {
			es.Title = v
		}
		if v, ok := raw["starts_at"].(string); ok {
			es.StartsAt = v
		}
		if v, ok := raw["ends_at"].(string); ok {
			es.EndsAt = v
		}
		if v, ok := raw["cover_url"].(string); ok {
			es.CoverURL = v
		}
		if gov, ok := raw["governance"].(map[string]any); ok {
			if owners, ok := gov["owners"].([]any); ok {
				hosts := make([]string, 0, len(owners))
				for _, owner := range owners {
					ownerMap, ok := owner.(map[string]any)
					if !ok {
						continue
					}
					if did, ok := ownerMap["did"].(string); ok && did != "" {
						hosts = append(hosts, did)
					}
				}
				es.Hosts = hosts
			}
		}
		eventList = append(eventList, es)
	}

	jsonOK(w, map[string]any{
		"id":           ms.ID,
		"name":         ms.Name,
		"description":  ms.Description,
		"avatar":       ms.Avatar,
		"backdrop_url": ms.BackdropURL,
		"website":      ms.Website,
		"governance":   ms.Governance,
		"updated_at":   ms.UpdatedAt,
		"subscribers":  subs,
		"subscribed":   subscribed,
		"events":       eventList,
	})
}

func (s *Server) handleUpdateCalendar(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	priv := authPrivKey(r)
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	current, err := s.db.GetCurrentCalendarState(ctx, id)
	if err != nil || current == nil {
		jsonErr(w, http.StatusNotFound, "calendar not found")
		return
	}

	var ms calendar.MutableState
	if err := json.Unmarshal([]byte(current.Payload), &ms); err != nil {
		jsonErr(w, http.StatusInternalServerError, "parse state")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Avatar      string `json:"avatar"`
		BackdropURL string `json:"backdrop_url"`
		Website     string `json:"website"`
		Owners      []string `json:"owners"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" {
		jsonErr(w, http.StatusBadRequest, "name required")
		return
	}

	var owners []calendar.GovernanceOwner
	if req.Owners != nil {
		seen := map[string]struct{}{}
		owners = make([]calendar.GovernanceOwner, 0, len(req.Owners))
		for _, owner := range req.Owners {
			owner = strings.TrimSpace(owner)
			if owner == "" {
				continue
			}
			if _, ok := seen[owner]; ok {
				continue
			}
			seen[owner] = struct{}{}
			owners = append(owners, calendar.GovernanceOwner{DID: owner, Role: "owner"})
		}
	}

	newState, newHash, err := calendar.Update(&ms, current.Hash, did, priv, calendar.Fields{
		Name:        req.Name,
		Description: req.Description,
		Avatar:      req.Avatar,
		BackdropURL: req.BackdropURL,
		Website:     req.Website,
		Owners:      owners,
	})
	if err != nil {
		jsonErr(w, http.StatusForbidden, err.Error())
		return
	}

	if err := s.db.AppendCalendarState(ctx, &store.CalendarState{
		Hash:      newHash,
		ID:        id,
		Prev:      current.Hash,
		Payload:   string(mustJSON(newState)),
		CreatedAt: time.Now(),
	}); err != nil {
		jsonErr(w, http.StatusInternalServerError, "store failed")
		return
	}

	ownerDIDs := make([]string, 0, len(newState.Governance.Owners))
	for _, owner := range newState.Governance.Owners {
		ownerDIDs = append(ownerDIDs, owner.DID)
	}
	if err := s.db.ReplaceCalendarOwners(ctx, id, ownerDIDs); err != nil {
		jsonErr(w, http.StatusInternalServerError, "store owners failed")
		return
	}

	jsonOK(w, map[string]any{"id": id, "state": newState})
}

func (s *Server) handleSubscribeCalendar(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	current, err := s.db.GetCurrentCalendarState(ctx, id)
	if err != nil || current == nil {
		jsonErr(w, http.StatusNotFound, "calendar not found")
		return
	}

	if err := s.db.SubscribeCalendar(ctx, id, did); err != nil {
		jsonErr(w, http.StatusInternalServerError, "subscribe failed")
		return
	}
	jsonOK(w, map[string]any{"subscribed": true})
}

func (s *Server) handleUnsubscribeCalendar(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if err := s.db.UnsubscribeCalendar(ctx, id, did); err != nil {
		jsonErr(w, http.StatusInternalServerError, "unsubscribe failed")
		return
	}
	jsonOK(w, map[string]any{"subscribed": false})
}
