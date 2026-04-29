package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
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
		ID:         randomHex(16),
		EventID:    id,
		SenderDID:  did,
		Payload:    encrypted,
		Status:     "pending",
		ReceivedAt: time.Now(),
	}
	if err := s.db.InsertRSVPEnvelope(ctx, envelope); err != nil {
		jsonErr(w, http.StatusInternalServerError, "save rsvp failed")
		return
	}

	jsonOK(w, map[string]string{"status": "ok"})
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
			"status":      env.Status,
			"received_at": env.ReceivedAt,
			"rsvp":        rsvp,
		})
	}

	jsonOK(w, out)
}
