package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/evermeet/evermeet/internal/events"
	"github.com/evermeet/evermeet/internal/identity"
	"github.com/evermeet/evermeet/internal/node"
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
	if (err != nil || founding == nil) && s.node != nil {
		// Try fetching from P2P
		if err := s.node.FetchEvent(id); err == nil {
			founding, err = s.db.GetEventFounding(ctx, id)
		}
	}
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
	var fd events.FoundingDoc
	if err := json.Unmarshal([]byte(founding.Payload), &fd); err == nil {
		if hostURL := s.eventHostURL(ctx, &fd, ms); hostURL != "" {
			if err := proxyPublicEventAPI(w, r, hostURL, "/api/events/"+url.PathEscape(id)); err != nil {
				s.log.Printf("proxy remote event %s via %s failed, using cached replica: %v", id, hostURL, err)
			} else {
				return
			}
		}
	}

	// Private events: only visible to organizer (for now; recipient list comes in Phase 8).
	if ms.Visibility == "private" {
		did := authDID(r)
		if did != ms.Organizer {
			jsonErr(w, http.StatusForbidden, "event is private")
			return
		}
	}
	rsvpCount, err := s.publicRSVPCount(ctx, id, ms.RSVP.Approval)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "load rsvp count failed")
		return
	}
	ms.RSVP.Count = rsvpCount
	payload, err := json.Marshal(ms)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "render event state")
		return
	}

	jsonOK(w, map[string]any{
		"id":      id,
		"state":   json.RawMessage(payload),
		"hash":    state.Hash,
		"founded": json.RawMessage(founding.Payload),
	})
}

func (s *Server) handleEventHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	states, err := s.db.ListEventStatesByID(ctx, id)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "list event history failed")
		return
	}

	type revision struct {
		Hash      string          `json:"hash"`
		Prev      string          `json:"prev,omitempty"`
		IsCurrent bool            `json:"is_current"`
		CreatedAt string          `json:"created_at"`
		State     json.RawMessage `json:"state"`
	}

	out := make([]revision, 0, len(states))
	for _, st := range states {
		out = append(out, revision{
			Hash:      st.Hash,
			Prev:      st.Prev,
			IsCurrent: st.IsCurrent,
			CreatedAt: st.CreatedAt.UTC().Format(time.RFC3339),
			State:     json.RawMessage(st.Payload),
		})
	}

	jsonOK(w, map[string]any{
		"id":        id,
		"revisions": out,
	})
}

func (s *Server) handleEventAttendees(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	founding, err := s.db.GetEventFounding(ctx, id)
	if err != nil || founding == nil {
		jsonErr(w, http.StatusNotFound, "event not found")
		return
	}
	state, err := s.db.GetCurrentEventState(ctx, id)
	if err != nil || state == nil {
		jsonErr(w, http.StatusNotFound, "event not found")
		return
	}
	var ms events.MutableState
	if err := json.Unmarshal([]byte(state.Payload), &ms); err != nil {
		jsonErr(w, http.StatusInternalServerError, "parse event state")
		return
	}
	var fd events.FoundingDoc
	if err := json.Unmarshal([]byte(founding.Payload), &fd); err == nil {
		if hostURL := s.eventHostURL(ctx, &fd, ms); hostURL != "" {
			if err := proxyPublicEventAPI(w, r, hostURL, "/api/events/"+url.PathEscape(id)+"/attendees"); err != nil {
				s.log.Printf("proxy remote attendees %s via %s failed, using local cache: %v", id, hostURL, err)
			} else {
				return
			}
		}
	}
	if ms.Visibility == "private" {
		did := authDID(r)
		if did != ms.Organizer {
			jsonErr(w, http.StatusForbidden, "event is private")
			return
		}
	}
	if ms.RSVP.Visible != nil && !*ms.RSVP.Visible {
		jsonErr(w, http.StatusForbidden, "attendees are private for this event")
		return
	}

	attendees, count, err := s.publicRSVPAttendees(ctx, id, ms.RSVP.Approval)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "load attendees failed")
		return
	}

	jsonOK(w, map[string]any{
		"attendees": attendees,
		"count":     count,
	})
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	priv := authPrivKey(r)
	ctx := r.Context()

	var req struct {
		Title        string           `json:"title"`
		Description  string           `json:"description"`
		CoverURL     string           `json:"cover_url"`
		StartsAt     string           `json:"starts_at"`
		EndsAt       string           `json:"ends_at"`
		Location     *events.Location `json:"location"`
		CalendarID   *string          `json:"calendar_id"`
		Owners       []string         `json:"owners"`
		Visibility   string           `json:"visibility"`
		RSVPLimit    int              `json:"rsvp_limit"`
		RSVPApproval string           `json:"rsvp_approval"`
		RSVPVisible  *bool            `json:"rsvp_visible"`
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
	var calendarID *string
	if req.CalendarID != nil {
		trimmed := strings.TrimSpace(*req.CalendarID)
		if trimmed != "" {
			calendarID = &trimmed
			calendar, err := s.db.GetCurrentCalendarState(ctx, trimmed)
			if err != nil {
				jsonErr(w, http.StatusInternalServerError, "check calendar failed")
				return
			}
			if calendar == nil {
				jsonErr(w, http.StatusBadRequest, "calendar not found")
				return
			}
			isOwner, err := s.db.IsCalendarOwner(ctx, trimmed, did)
			if err != nil {
				jsonErr(w, http.StatusInternalServerError, "check calendar ownership failed")
				return
			}
			if !isOwner {
				jsonErr(w, http.StatusForbidden, "you are not an owner of this calendar")
				return
			}
		}
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "starts_at must be RFC3339")
		return
	}

	f := events.Fields{
		Title:        req.Title,
		Description:  req.Description,
		CoverURL:     req.CoverURL,
		StartsAt:     startsAt,
		Location:     req.Location,
		CalendarID:   calendarID,
		Visibility:   req.Visibility,
		RSVPLimit:    req.RSVPLimit,
		RSVPApproval: req.RSVPApproval,
		RSVPVisible:  req.RSVPVisible,
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

	founding, eventID, state, stateHash, err := events.New(did, priv, s.homeHost(), s.baseURL, f)
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

	// Broadcast via P2P if node is available
	if s.node != nil {
		gossip := node.GossipEvent{
			Founding: founding,
			State:    state,
		}
		data, _ := json.Marshal(gossip)
		if err := s.node.BroadcastEvent(data); err != nil {
			s.log.Printf("p2p broadcast failed: %v", err)
		}
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
		CoverURL     string           `json:"cover_url"`
		StartsAt     string           `json:"starts_at"`
		EndsAt       string           `json:"ends_at"`
		Location     *events.Location `json:"location"`
		CalendarID   *string          `json:"calendar_id"`
		Owners       []string         `json:"owners"`
		Visibility   string           `json:"visibility"`
		RSVPLimit    int              `json:"rsvp_limit"`
		RSVPApproval string           `json:"rsvp_approval"`
		RSVPVisible  *bool            `json:"rsvp_visible"`
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

	nextCalendar := ms.Calendar
	if req.CalendarID != nil {
		trimmed := strings.TrimSpace(*req.CalendarID)
		if trimmed == "" {
			nextCalendar = nil
		} else {
			calendar, err := s.db.GetCurrentCalendarState(ctx, trimmed)
			if err != nil {
				jsonErr(w, http.StatusInternalServerError, "check calendar failed")
				return
			}
			if calendar == nil {
				jsonErr(w, http.StatusBadRequest, "calendar not found")
				return
			}
			isOwner, err := s.db.IsCalendarOwner(ctx, trimmed, did)
			if err != nil {
				jsonErr(w, http.StatusInternalServerError, "check calendar ownership failed")
				return
			}
			if !isOwner {
				jsonErr(w, http.StatusForbidden, "you are not an owner of this calendar")
				return
			}
			nextCalendar = &trimmed
		}
	}
	var owners []events.GovernanceOwner
	if req.Owners != nil {
		seen := map[string]struct{}{}
		owners = make([]events.GovernanceOwner, 0, len(req.Owners))
		for _, owner := range req.Owners {
			owner = strings.TrimSpace(owner)
			if owner == "" {
				continue
			}
			if _, ok := seen[owner]; ok {
				continue
			}
			seen[owner] = struct{}{}
			owners = append(owners, events.GovernanceOwner{DID: owner, Role: "owner"})
		}
	}
	rsvpVisible := req.RSVPVisible
	if rsvpVisible == nil {
		currentVisible := ms.RSVP.Visible == nil || *ms.RSVP.Visible
		rsvpVisible = &currentVisible
	}

	f := events.Fields{
		Title:        req.Title,
		Description:  req.Description,
		CoverURL:     req.CoverURL,
		StartsAt:     startsAt,
		Location:     req.Location,
		Visibility:   req.Visibility,
		RSVPLimit:    req.RSVPLimit,
		RSVPApproval: req.RSVPApproval,
		RSVPVisible:  rsvpVisible,
		Tags:         req.Tags,
		CalendarID:   nextCalendar,
		Owners:       owners,
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

	// Broadcast update via P2P
	if s.node != nil {
		// For updates, we might only need the state, but GossipEvent expects founding.
		// We can get founding from DB if needed, or update GossipEvent to make it optional.
		// For now, let's fetch founding.
		founding, _ := s.db.GetEventFounding(ctx, id)
		var fd events.FoundingDoc
		if founding != nil {
			json.Unmarshal([]byte(founding.Payload), &fd)
		}

		gossip := node.GossipEvent{
			Founding: &fd,
			State:    newState,
		}
		data, _ := json.Marshal(gossip)
		s.node.BroadcastEvent(data)
	}

	jsonOK(w, map[string]any{
		"id":    id,
		"hash":  newHash,
		"state": json.RawMessage(statePayload),
	})
}

func (s *Server) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	did := authDID(r)
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
	if ms.Organizer != did {
		jsonErr(w, http.StatusForbidden, "only organizer can delete event")
		return
	}

	if err := s.db.DeleteEvent(ctx, id); err != nil {
		jsonErr(w, http.StatusInternalServerError, "delete event failed")
		return
	}

	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleRSVP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	did := authDID(r)
	ctx := r.Context()

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Note  string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// 1. Get event state to find organizer DID
	state, err := s.db.GetCurrentEventState(ctx, id)
	if err != nil || state == nil {
		jsonErr(w, http.StatusNotFound, "event not found")
		return
	}
	var ms events.MutableState
	json.Unmarshal([]byte(state.Payload), &ms)
	if ms.RSVP.Visible != nil && !*ms.RSVP.Visible {
		jsonErr(w, http.StatusForbidden, "rsvp is not visible for this event")
		return
	}
	existing, err := s.db.GetRSVPForEventSender(ctx, id, did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "check rsvp status failed")
		return
	}
	if existing != nil && existing.Status != "cancelled" {
		jsonOK(w, map[string]any{
			"status":        rsvpEffectiveStatus(existing.Status, ms.RSVP.Approval),
			"guest_visible": existing.GuestVisible,
			"received_at":   existing.ReceivedAt,
		})
		return
	}

	// 2. Get organizer's public key
	organizer, err := s.db.GetUser(ctx, ms.Organizer)
	if err != nil || organizer == nil {
		jsonErr(w, http.StatusNotFound, "organizer not found")
		return
	}
	organizerPub, err := hex.DecodeString(organizer.CurrentPK)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "invalid organizer key")
		return
	}

	// 3. Create and encrypt RSVP
	rsvp := events.RSVP{
		EventID:   id,
		SenderDID: did,
		Name:      req.Name,
		Email:     req.Email,
		Note:      req.Note,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	rsvpPayload, _ := json.Marshal(rsvp)
	encrypted, err := identity.EncryptForRecipient(organizerPub, rsvpPayload)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "encryption failed")
		return
	}

	// 4. Save envelope
	envelope := &store.RSVPEnvelope{
		ID:           randomHex(16),
		EventID:      id,
		SenderDID:    did,
		Payload:      encrypted,
		Status:       rsvpInitialStatus(ms.RSVP.Approval),
		GuestVisible: true,
		ReceivedAt:   time.Now(),
	}
	if err := s.db.InsertRSVPEnvelope(ctx, envelope); err != nil {
		jsonErr(w, http.StatusInternalServerError, "save rsvp failed")
		return
	}
	s.emitRSVPReceiptAsync(did, ms, rsvpEffectiveStatus(envelope.Status, ms.RSVP.Approval), envelope.GuestVisible)

	jsonOK(w, map[string]any{
		"status":        envelope.Status,
		"guest_visible": envelope.GuestVisible,
		"received_at":   envelope.ReceivedAt,
	})
}

func (s *Server) handleCancelRSVP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	did := authDID(r)
	ctx := r.Context()

	state, err := s.db.GetCurrentEventState(ctx, id)
	if err != nil || state == nil {
		jsonErr(w, http.StatusNotFound, "event not found")
		return
	}
	var ms events.MutableState
	json.Unmarshal([]byte(state.Payload), &ms)

	envelope, err := s.db.GetRSVPForEventSender(ctx, id, did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "check rsvp status failed")
		return
	}
	if envelope == nil {
		jsonOK(w, map[string]any{"has_rsvp": false})
		return
	}
	if envelope.Status != "cancelled" {
		if err := s.db.UpdateRSVPStatus(ctx, envelope.ID, "cancelled"); err != nil {
			jsonErr(w, http.StatusInternalServerError, "cancel rsvp failed")
			return
		}
		envelope.Status = "cancelled"
	}
	s.emitRSVPReceiptAsync(did, ms, envelope.Status, envelope.GuestVisible)

	jsonOK(w, map[string]any{
		"has_rsvp":      true,
		"status":        envelope.Status,
		"guest_visible": envelope.GuestVisible,
		"received_at":   envelope.ReceivedAt,
	})
}

func (s *Server) handleUpdateRSVPGuestVisibility(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	did := authDID(r)
	ctx := r.Context()

	var req struct {
		Visible bool `json:"visible"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	state, err := s.db.GetCurrentEventState(ctx, id)
	if err != nil || state == nil {
		jsonErr(w, http.StatusNotFound, "event not found")
		return
	}
	var ms events.MutableState
	json.Unmarshal([]byte(state.Payload), &ms)

	envelope, err := s.db.GetRSVPForEventSender(ctx, id, did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "check rsvp status failed")
		return
	}
	if envelope == nil || envelope.Status == "cancelled" {
		jsonErr(w, http.StatusNotFound, "active rsvp not found")
		return
	}
	if err := s.db.UpdateRSVPGuestVisible(ctx, envelope.ID, req.Visible); err != nil {
		jsonErr(w, http.StatusInternalServerError, "update guest list visibility failed")
		return
	}
	envelope.GuestVisible = req.Visible
	s.emitRSVPReceiptAsync(did, ms, rsvpEffectiveStatus(envelope.Status, ms.RSVP.Approval), envelope.GuestVisible)

	jsonOK(w, map[string]any{
		"has_rsvp":      true,
		"status":        rsvpEffectiveStatus(envelope.Status, ms.RSVP.Approval),
		"guest_visible": envelope.GuestVisible,
		"received_at":   envelope.ReceivedAt,
	})
}

func (s *Server) handleMyRSVPStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	did := authDID(r)
	ctx := r.Context()

	state, err := s.db.GetCurrentEventState(ctx, id)
	if err != nil || state == nil {
		jsonErr(w, http.StatusNotFound, "event not found")
		return
	}
	var ms events.MutableState
	json.Unmarshal([]byte(state.Payload), &ms)

	envelope, err := s.db.GetRSVPForEventSender(ctx, id, did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "check rsvp status failed")
		return
	}
	if envelope == nil {
		jsonOK(w, map[string]any{"has_rsvp": false})
		return
	}

	jsonOK(w, map[string]any{
		"has_rsvp":      true,
		"status":        rsvpEffectiveStatus(envelope.Status, ms.RSVP.Approval),
		"guest_visible": envelope.GuestVisible,
		"received_at":   envelope.ReceivedAt,
	})
}

func (s *Server) handleListRSVPs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	did := authDID(r)
	priv := authPrivKey(r)
	ctx := r.Context()

	// 1. Verify requester is organizer
	state, err := s.db.GetCurrentEventState(ctx, id)
	if err != nil || state == nil {
		jsonErr(w, http.StatusNotFound, "event not found")
		return
	}
	var ms events.MutableState
	json.Unmarshal([]byte(state.Payload), &ms)
	if ms.Organizer != did {
		jsonErr(w, http.StatusForbidden, "only organizer can list RSVPs")
		return
	}

	// 2. List envelopes
	envelopes, err := s.db.ListRSVPsForEvent(ctx, id)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "list rsvps failed")
		return
	}

	// 3. Decrypt payloads
	out := make([]any, 0, len(envelopes))
	for _, env := range envelopes {
		decrypted, err := identity.DecryptForRecipient(priv, env.Payload)
		if err != nil {
			s.log.Printf("decrypt rsvp %s: %v", env.ID, err)
			continue
		}
		var rsvp events.RSVP
		if err := json.Unmarshal(decrypted, &rsvp); err != nil {
			continue
		}
		out = append(out, map[string]any{
			"id":          env.ID,
			"sender_did":  env.SenderDID,
			"status":      rsvpEffectiveStatus(env.Status, ms.RSVP.Approval),
			"received_at": env.ReceivedAt,
			"rsvp":        rsvp,
		})
	}

	jsonOK(w, out)
}

func rsvpInitialStatus(approval string) string {
	if approval == "manual" {
		return "pending"
	}
	return "confirmed"
}

func rsvpEffectiveStatus(status, approval string) string {
	if status == "pending" && approval != "manual" {
		return "confirmed"
	}
	return status
}

func (s *Server) publicRSVPCount(ctx context.Context, eventID, approval string) (int, error) {
	latest, err := s.latestRSVPEnvelopes(ctx, eventID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, env := range latest {
		if rsvpEffectiveStatus(env.Status, approval) == "confirmed" {
			count++
		}
	}
	return count, nil
}

func (s *Server) eventHostURL(ctx context.Context, founding *events.FoundingDoc, ms events.MutableState) string {
	localBase := strings.TrimRight(s.baseURL, "/")
	if founding != nil {
		if instanceURL := strings.TrimRight(founding.InstanceURL, "/"); instanceURL != "" && instanceURL != localBase {
			return instanceURL
		}
		if s.isLocalInstanceID(founding.InstanceID) {
			return ""
		}
		if founding.InstanceID != "" && founding.InstanceID != s.homeHost() {
			if instanceURL := s.instanceIDURL(founding.InstanceID); instanceURL != "" && instanceURL != localBase {
				return instanceURL
			}
			user, err := s.db.GetUser(ctx, ms.Organizer)
			if err == nil && user != nil {
				if endpoint := strings.TrimRight(user.Endpoint, "/"); endpoint != "" && endpoint != localBase {
					return endpoint
				}
			}
		}
	}
	return ""
}

func (s *Server) isLocalInstanceID(instanceID string) bool {
	if instanceID == "" {
		return false
	}
	if instanceID == s.homeHost() {
		return true
	}
	id, _, ok := strings.Cut(instanceID, "@")
	return ok && id == s.instanceID
}

func (s *Server) instanceIDURL(instanceID string) string {
	_, host, ok := strings.Cut(instanceID, "@")
	if !ok || host == "" {
		return ""
	}
	scheme := "http"
	if u, err := url.Parse(s.baseURL); err == nil && u.Scheme != "" {
		scheme = u.Scheme
	}
	return scheme + "://" + host
}

func proxyPublicEventAPI(w http.ResponseWriter, r *http.Request, hostURL, path string) error {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, strings.TrimRight(hostURL, "/")+path, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("remote returned %s", resp.Status)
	}
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}

type publicRSVPAttendee struct {
	DID         string `json:"did"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
}

func (s *Server) publicRSVPAttendees(ctx context.Context, eventID, approval string) ([]publicRSVPAttendee, int, error) {
	latest, err := s.latestRSVPEnvelopes(ctx, eventID)
	if err != nil {
		return nil, 0, err
	}

	count := 0
	attendees := make([]publicRSVPAttendee, 0, len(latest))
	for _, env := range latest {
		if rsvpEffectiveStatus(env.Status, approval) != "confirmed" {
			continue
		}
		count++
		if !env.GuestVisible {
			continue
		}
		attendee := publicRSVPAttendee{
			DID:         env.SenderDID,
			DisplayName: shortDID(env.SenderDID),
		}
		user, err := s.db.GetUser(ctx, env.SenderDID)
		if err != nil {
			return nil, 0, err
		}
		if user != nil {
			attendee.DisplayName = strings.TrimSpace(user.DisplayName)
			if attendee.DisplayName == "" {
				attendee.DisplayName = shortDID(env.SenderDID)
			}
			attendee.Avatar = user.Avatar
		}
		attendees = append(attendees, attendee)
	}
	return attendees, count, nil
}

func (s *Server) latestRSVPEnvelopes(ctx context.Context, eventID string) ([]*store.RSVPEnvelope, error) {
	envelopes, err := s.db.ListRSVPsForEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	latestBySender := make(map[string]*store.RSVPEnvelope, len(envelopes))
	for _, env := range envelopes {
		latestBySender[env.SenderDID] = env
	}
	latest := make([]*store.RSVPEnvelope, 0, len(latestBySender))
	for _, env := range latestBySender {
		latest = append(latest, env)
	}
	sort.Slice(latest, func(i, j int) bool {
		return latest[i].ReceivedAt.Before(latest[j].ReceivedAt)
	})
	return latest, nil
}

func shortDID(did string) string {
	if len(did) <= 16 {
		return did
	}
	return did[:16] + "..."
}
