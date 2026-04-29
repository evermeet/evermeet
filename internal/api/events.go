package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/evermeet/evermeet/internal/events"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/go-chi/chi/v5"
)

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	states, err := s.db.ListCurrentEventStates(r.Context(), limit, offset)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]json.RawMessage, 0, len(states))
	for _, st := range states {
		// Only return public/unlisted events in the listing.
		var ms events.MutableState
		if err := json.Unmarshal([]byte(st.Payload), &ms); err != nil {
			continue
		}
		if ms.Visibility == "private" {
			continue
		}
		out = append(out, json.RawMessage(st.Payload))
	}

	jsonOK(w, map[string]any{"events": out, "limit": limit, "offset": offset})
}

func (s *Server) handleGetEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	founding, err := s.db.GetEventFounding(ctx, id)
	if err != nil || founding == nil {
		jsonErr(w, http.StatusNotFound, "event not found")
		return
	}

	state, err := s.db.GetCurrentEventState(ctx, id)
	if err != nil || state == nil {
		jsonErr(w, http.StatusNotFound, "event state not found")
		return
	}

	var ms events.MutableState
	if err := json.Unmarshal([]byte(state.Payload), &ms); err != nil {
		jsonErr(w, http.StatusInternalServerError, "parse event state")
		return
	}

	// Private events: only visible to organizer (for now; recipient list comes in Phase 8).
	if ms.Visibility == "private" {
		did := authDID(r)
		if did != ms.Organizer {
			jsonErr(w, http.StatusForbidden, "event is private")
			return
		}
	}

	jsonOK(w, map[string]any{
		"id":      id,
		"state":   json.RawMessage(state.Payload),
		"hash":    state.Hash,
		"founded": json.RawMessage(founding.Payload),
	})
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	priv := authPrivKey(r)
	ctx := r.Context()

	var req struct {
		Title        string           `json:"title"`
		Description  string           `json:"description"`
		StartsAt     string           `json:"starts_at"`
		EndsAt       string           `json:"ends_at"`
		Location     *events.Location `json:"location"`
		CalendarID   *string          `json:"calendar_id"`
		Visibility   string           `json:"visibility"`
		RSVPLimit    int              `json:"rsvp_limit"`
		RSVPApproval string           `json:"rsvp_approval"`
		Tags         []string         `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Title == "" {
		jsonErr(w, http.StatusBadRequest, "title required")
		return
	}
	if req.StartsAt == "" {
		jsonErr(w, http.StatusBadRequest, "starts_at required")
		return
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "starts_at must be RFC3339")
		return
	}

	f := events.Fields{
		Title:        req.Title,
		Description:  req.Description,
		StartsAt:     startsAt,
		Location:     req.Location,
		CalendarID:   req.CalendarID,
		Visibility:   req.Visibility,
		RSVPLimit:    req.RSVPLimit,
		RSVPApproval: req.RSVPApproval,
		Tags:         req.Tags,
	}
	if req.EndsAt != "" {
		t, err := time.Parse(time.RFC3339, req.EndsAt)
		if err != nil {
			jsonErr(w, http.StatusBadRequest, "ends_at must be RFC3339")
			return
		}
		f.EndsAt = &t
	}

	founding, eventID, state, stateHash, err := events.New(did, priv, f)
	if err != nil {
		s.log.Printf("create event: %v", err)
		jsonErr(w, http.StatusInternalServerError, "create event failed")
		return
	}

	foundingPayload := mustJSON(founding)
	statePayload := mustJSON(state)

	if err := s.db.InsertEventFounding(ctx, &store.EventFounding{
		ID:      eventID,
		Payload: foundingPayload,
	}); err != nil {
		jsonErr(w, http.StatusInternalServerError, "persist event failed")
		return
	}

	if err := s.db.AppendEventState(ctx, &store.EventState{
		Hash:      stateHash,
		ID:        eventID,
		Prev:      "",
		Payload:   statePayload,
		CreatedAt: time.Now(),
	}); err != nil {
		jsonErr(w, http.StatusInternalServerError, "persist event state failed")
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]any{
		"id":    eventID,
		"hash":  stateHash,
		"state": json.RawMessage(statePayload),
	})
}

func (s *Server) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	did := authDID(r)
	priv := authPrivKey(r)
	ctx := r.Context()

	current, err := s.db.GetCurrentEventState(ctx, id)
	if err != nil || current == nil {
		jsonErr(w, http.StatusNotFound, "event not found")
		return
	}

	var ms events.MutableState
	if err := json.Unmarshal([]byte(current.Payload), &ms); err != nil {
		jsonErr(w, http.StatusInternalServerError, "parse current state")
		return
	}

	var req struct {
		Title        string           `json:"title"`
		Description  string           `json:"description"`
		StartsAt     string           `json:"starts_at"`
		EndsAt       string           `json:"ends_at"`
		Location     *events.Location `json:"location"`
		Visibility   string           `json:"visibility"`
		RSVPLimit    int              `json:"rsvp_limit"`
		RSVPApproval string           `json:"rsvp_approval"`
		Tags         []string         `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "starts_at must be RFC3339")
		return
	}

	f := events.Fields{
		Title:        req.Title,
		Description:  req.Description,
		StartsAt:     startsAt,
		Location:     req.Location,
		Visibility:   req.Visibility,
		RSVPLimit:    req.RSVPLimit,
		RSVPApproval: req.RSVPApproval,
		Tags:         req.Tags,
		CalendarID:   ms.Calendar,
	}
	if req.EndsAt != "" {
		t, err := time.Parse(time.RFC3339, req.EndsAt)
		if err != nil {
			jsonErr(w, http.StatusBadRequest, "ends_at must be RFC3339")
			return
		}
		f.EndsAt = &t
	}

	newState, newHash, err := events.Update(&ms, current.Hash, did, priv, f)
	if err != nil {
		jsonErr(w, http.StatusForbidden, err.Error())
		return
	}

	statePayload := mustJSON(newState)
	if err := s.db.AppendEventState(ctx, &store.EventState{
		Hash:      newHash,
		ID:        id,
		Prev:      current.Hash,
		Payload:   statePayload,
		CreatedAt: time.Now(),
	}); err != nil {
		jsonErr(w, http.StatusInternalServerError, "persist update failed")
		return
	}

	jsonOK(w, map[string]any{
		"id":    id,
		"hash":  newHash,
		"state": json.RawMessage(statePayload),
	})
}
