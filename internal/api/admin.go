package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/evermeet/evermeet/internal/calendar"
	"github.com/evermeet/evermeet/internal/events"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/go-chi/chi/v5"
)

func (s *Server) handleAdminOverview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	adminCount, err := s.db.CountAdminAccounts(ctx)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "count admins failed")
		return
	}
	userCount, err := s.db.CountUsers(ctx)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "count users failed")
		return
	}
	eventCount, err := s.db.CountCurrentEvents(ctx)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "count events failed")
		return
	}
	calendarCount, err := s.db.CountCalendars(ctx)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "count calendars failed")
		return
	}
	blobCount, err := s.db.CountBlobs(ctx)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "count blobs failed")
		return
	}

	uptime := time.Since(s.startTime)
	homeID := s.homeHost()
	var p2pStatus any
	if n := s.libp2pNode(); n != nil {
		st := n.Status()
		p2pStatus = map[string]any{
			"id":                   st.ID,
			"addresses":            st.Addresses,
			"peers":                st.Peers,
			"evermeet_instance_id": homeID,
		}
	} else {
		p2pStatus = map[string]any{
			"evermeet_instance_id": homeID,
		}
	}

	jsonOK(w, map[string]any{
		"instance_id": s.homeHost(),
		"base_url":    s.baseURL,
		"version":     "v0.1.0-alpha",
		"started_at":  s.startTime.UTC().Format(time.RFC3339),
		"uptime":      formatUptime(uptime),
		"counts": map[string]int{
			"admins":    adminCount,
			"users":     userCount,
			"events":    eventCount,
			"calendars": calendarCount,
			"blobs":     blobCount,
		},
		"p2p": p2pStatus,
		"config": map[string]any{
			"node": s.cfg.Node,
			"p2p":  s.cfg.P2P,
		},
	})
}

type adminObjectItem struct {
	Type      string         `json:"type"`
	ID        string         `json:"id"`
	Label     string         `json:"label"`
	Subtitle  string         `json:"subtitle,omitempty"`
	UpdatedAt string         `json:"updated_at,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

type adminObjectGroup struct {
	Type  string            `json:"type"`
	Label string            `json:"label"`
	Count int               `json:"count"`
	Items []adminObjectItem `json:"items"`
}

type adminObjectSummary struct {
	Type       string `json:"type"`
	Label      string `json:"label"`
	Count      int    `json:"count"`
	HostedHere int    `json:"hosted_here"`
	Bytes      int64  `json:"bytes"`
}

type adminAccountResponse struct {
	DID         string `json:"did"`
	DisplayName string `json:"display_name"`
	Endpoint    string `json:"endpoint,omitempty"`
	Role        string `json:"role"`
	CreatedAt   string `json:"created_at"`
}

func (s *Server) handleAdminObjectsOverview(w http.ResponseWriter, r *http.Request) {
	summaries := make([]adminObjectSummary, 0, 4)
	for _, objectType := range []string{"users", "events", "calendars", "blobs"} {
		summary, err := s.adminObjectSummary(r, objectType)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, "load "+objectType+" failed")
			return
		}
		summaries = append(summaries, summary)
	}
	jsonOK(w, map[string]any{"objects": summaries})
}

func (s *Server) handleAdminAdminsList(w http.ResponseWriter, r *http.Request) {
	admins, err := s.db.ListAdminAccounts(r.Context())
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "load admins failed")
		return
	}
	myRole, err := s.db.GetAdminRole(r.Context(), authDID(r))
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "admin role lookup failed")
		return
	}
	myRole = normalizeAdminRole(myRole)
	out := make([]adminAccountResponse, 0, len(admins))
	for _, admin := range admins {
		displayName := strings.TrimSpace(admin.Name)
		if displayName == "" {
			displayName = shortDID(admin.DID)
		}
		out = append(out, adminAccountResponse{
			DID:         admin.DID,
			DisplayName: displayName,
			Endpoint:    admin.Endpoint,
			Role:        normalizeAdminRole(admin.Role),
			CreatedAt:   admin.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	jsonOK(w, map[string]any{
		"my_role": myRole,
		"admins":  out,
	})
}

func (s *Server) handleAdminAdminsCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DID   string `json:"did"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid request")
		return
	}
	role := normalizeAdminRole(req.Role)
	if role == "" {
		role = "admin"
	}
	var did string
	if strings.TrimSpace(req.Email) != "" {
		_, resolvedDID, err := s.lookupOrCreateUser(r.Context(), strings.TrimSpace(req.Email))
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, "resolve user failed")
			return
		}
		did = resolvedDID
	} else {
		did = strings.TrimSpace(req.DID)
	}
	if did == "" {
		jsonErr(w, http.StatusBadRequest, "email or did required")
		return
	}
	exists, err := s.db.IsAdmin(r.Context(), did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "admin lookup failed")
		return
	}
	if exists {
		jsonErr(w, http.StatusConflict, "admin already exists")
		return
	}
	if err := s.db.InsertAdminAccount(r.Context(), &store.AdminAccount{
		DID:       did,
		Role:      role,
		CreatedAt: time.Now(),
	}); err != nil {
		jsonErr(w, http.StatusInternalServerError, "create admin failed")
		return
	}
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleAdminAdminsSetRole(w http.ResponseWriter, r *http.Request) {
	did := strings.TrimSpace(chi.URLParam(r, "did"))
	if did == "" {
		jsonErr(w, http.StatusBadRequest, "did required")
		return
	}
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid request")
		return
	}
	role := normalizeAdminRole(req.Role)
	if role == "" {
		jsonErr(w, http.StatusBadRequest, "invalid role")
		return
	}
	currentRole, err := s.db.GetAdminRole(r.Context(), did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "admin role lookup failed")
		return
	}
	if currentRole == "" {
		jsonErr(w, http.StatusNotFound, "admin not found")
		return
	}
	if currentRole == "owner" && role != "owner" {
		owners, err := s.db.CountAdminsByRole(r.Context(), "owner")
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, "owner count failed")
			return
		}
		if owners <= 1 {
			jsonErr(w, http.StatusBadRequest, "at least one owner is required")
			return
		}
	}
	if err := s.db.SetAdminRole(r.Context(), did, role); err != nil {
		jsonErr(w, http.StatusInternalServerError, "update role failed")
		return
	}
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleAdminObjectsByType(w http.ResponseWriter, r *http.Request) {
	objectType := normalizeAdminObjectType(chi.URLParam(r, "type"))
	if objectType == "" {
		jsonErr(w, http.StatusNotFound, "unknown object type")
		return
	}
	limit := queryInt(r, "limit", 50, 1, 100)
	offset := queryInt(r, "offset", 0, 0, 1000000)
	group, err := s.adminObjectGroup(r, objectType, limit, offset)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "load objects failed")
		return
	}
	jsonOK(w, map[string]any{
		"type":   group.Type,
		"label":  group.Label,
		"count":  group.Count,
		"items":  group.Items,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Server) adminObjectGroup(r *http.Request, objectType string, limit, offset int) (adminObjectGroup, error) {
	ctx := r.Context()
	switch objectType {
	case "users":
		count, err := s.db.CountUsers(ctx)
		if err != nil {
			return adminObjectGroup{}, err
		}
		users, err := s.db.ListUsers(ctx, limit, offset)
		if err != nil {
			return adminObjectGroup{}, err
		}
		items := make([]adminObjectItem, 0, len(users))
		for _, user := range users {
			label := strings.TrimSpace(user.DisplayName)
			if label == "" {
				label = shortDID(user.DID)
			}
			items = append(items, adminObjectItem{
				Type:      "users",
				ID:        user.DID,
				Label:     label,
				Subtitle:  user.Endpoint,
				UpdatedAt: user.UpdatedAt.UTC().Format(time.RFC3339),
				Meta: map[string]any{
					"instance_id": user.InstanceID,
				},
			})
		}
		return adminObjectGroup{Type: "users", Label: "Users", Count: count, Items: items}, nil
	case "events":
		count, err := s.db.CountCurrentEvents(ctx)
		if err != nil {
			return adminObjectGroup{}, err
		}
		states, err := s.db.ListCurrentEventStates(ctx, limit, offset)
		if err != nil {
			return adminObjectGroup{}, err
		}
		items := make([]adminObjectItem, 0, len(states))
		for _, state := range states {
			var ms events.MutableState
			_ = json.Unmarshal([]byte(state.Payload), &ms)
			label := strings.TrimSpace(ms.Title)
			if label == "" {
				label = state.ID
			}
			items = append(items, adminObjectItem{
				Type:      "events",
				ID:        state.ID,
				Label:     label,
				Subtitle:  ms.StartsAt,
				UpdatedAt: state.CreatedAt.UTC().Format(time.RFC3339),
				Meta: map[string]any{
					"visibility": ms.Visibility,
					"organizer":  ms.Organizer,
				},
			})
		}
		return adminObjectGroup{Type: "events", Label: "Events", Count: count, Items: items}, nil
	case "calendars":
		count, err := s.db.CountCalendars(ctx)
		if err != nil {
			return adminObjectGroup{}, err
		}
		states, err := s.db.ListCurrentCalendarsPaginated(ctx, limit, offset)
		if err != nil {
			return adminObjectGroup{}, err
		}
		items := make([]adminObjectItem, 0, len(states))
		for _, state := range states {
			var ms calendar.MutableState
			_ = json.Unmarshal([]byte(state.Payload), &ms)
			label := strings.TrimSpace(ms.Name)
			if label == "" {
				label = state.ID
			}
			items = append(items, adminObjectItem{
				Type:      "calendars",
				ID:        state.ID,
				Label:     label,
				Subtitle:  ms.Description,
				UpdatedAt: state.CreatedAt.UTC().Format(time.RFC3339),
				Meta: map[string]any{
					"owners": len(ms.Governance.Owners),
				},
			})
		}
		return adminObjectGroup{Type: "calendars", Label: "Calendars", Count: count, Items: items}, nil
	case "blobs":
		count, err := s.db.CountBlobs(ctx)
		if err != nil {
			return adminObjectGroup{}, err
		}
		blobs, err := s.db.ListBlobs(ctx, limit, offset)
		if err != nil {
			return adminObjectGroup{}, err
		}
		items := make([]adminObjectItem, 0, len(blobs))
		for _, blob := range blobs {
			items = append(items, adminObjectItem{
				Type:      "blobs",
				ID:        blob.Hash,
				Label:     blob.Hash,
				Subtitle:  blob.ContentType,
				UpdatedAt: blob.CreatedAt.UTC().Format(time.RFC3339),
				Meta: map[string]any{
					"size":        blob.Size,
					"uploaded_by": blob.UploadedBy,
				},
			})
		}
		return adminObjectGroup{Type: "blobs", Label: "Blobs", Count: count, Items: items}, nil
	default:
		return adminObjectGroup{}, fmt.Errorf("unknown object type %q", objectType)
	}
}

func (s *Server) adminObjectSummary(r *http.Request, objectType string) (adminObjectSummary, error) {
	ctx := r.Context()
	switch objectType {
	case "users":
		count, err := s.db.CountUsers(ctx)
		if err != nil {
			return adminObjectSummary{}, err
		}
		hostedHere, err := s.db.CountLocalUsers(ctx, s.baseURL)
		if err != nil {
			return adminObjectSummary{}, err
		}
		bytes, err := s.db.EstimateUserBytes(ctx)
		if err != nil {
			return adminObjectSummary{}, err
		}
		return adminObjectSummary{
			Type:       "users",
			Label:      "Users",
			Count:      count,
			HostedHere: hostedHere,
			Bytes:      bytes,
		}, nil
	case "events":
		count, err := s.db.CountCurrentEvents(ctx)
		if err != nil {
			return adminObjectSummary{}, err
		}
		bytes, err := s.db.EstimateCurrentEventBytes(ctx)
		if err != nil {
			return adminObjectSummary{}, err
		}
		hostedHere, err := s.countHostedEvents(ctx, count)
		if err != nil {
			return adminObjectSummary{}, err
		}
		return adminObjectSummary{
			Type:       "events",
			Label:      "Events",
			Count:      count,
			HostedHere: hostedHere,
			Bytes:      bytes,
		}, nil
	case "calendars":
		count, err := s.db.CountCalendars(ctx)
		if err != nil {
			return adminObjectSummary{}, err
		}
		bytes, err := s.db.EstimateCurrentCalendarBytes(ctx)
		if err != nil {
			return adminObjectSummary{}, err
		}
		hostedHere, err := s.countHostedCalendars(ctx, count)
		if err != nil {
			return adminObjectSummary{}, err
		}
		return adminObjectSummary{
			Type:       "calendars",
			Label:      "Calendars",
			Count:      count,
			HostedHere: hostedHere,
			Bytes:      bytes,
		}, nil
	case "blobs":
		count, err := s.db.CountBlobs(ctx)
		if err != nil {
			return adminObjectSummary{}, err
		}
		bytes, err := s.db.SumBlobBytes(ctx)
		if err != nil {
			return adminObjectSummary{}, err
		}
		hostedHere, err := s.db.CountLocalBlobs(ctx)
		if err != nil {
			return adminObjectSummary{}, err
		}
		return adminObjectSummary{
			Type:       "blobs",
			Label:      "Blobs",
			Count:      count,
			HostedHere: hostedHere,
			Bytes:      bytes,
		}, nil
	default:
		return adminObjectSummary{}, fmt.Errorf("unknown object type %q", objectType)
	}
}

func (s *Server) countHostedEvents(ctx context.Context, count int) (int, error) {
	if count == 0 {
		return 0, nil
	}
	states, err := s.db.ListCurrentEventStates(ctx, count, 0)
	if err != nil {
		return 0, err
	}
	hosted := 0
	for _, state := range states {
		founding, err := s.db.GetEventFounding(ctx, state.ID)
		if err != nil || founding == nil {
			if err != nil {
				return 0, err
			}
			continue
		}
		var fd events.FoundingDoc
		if err := json.Unmarshal([]byte(founding.Payload), &fd); err != nil {
			continue
		}
		if s.eventFoundingHostedHere(fd) {
			hosted++
		}
	}
	return hosted, nil
}

func (s *Server) countHostedCalendars(ctx context.Context, count int) (int, error) {
	if count == 0 {
		return 0, nil
	}
	states, err := s.db.ListCurrentCalendarsPaginated(ctx, count, 0)
	if err != nil {
		return 0, err
	}
	hosted := 0
	for _, state := range states {
		founding, err := s.db.GetCalendarFounding(ctx, state.ID)
		if err != nil || founding == nil {
			if err != nil {
				return 0, err
			}
			continue
		}
		var fd calendar.FoundingDoc
		if err := json.Unmarshal([]byte(founding.Payload), &fd); err != nil {
			continue
		}
		if s.isLocalInstanceID(fd.InstanceID) {
			hosted++
		}
	}
	return hosted, nil
}

func (s *Server) eventFoundingHostedHere(fd events.FoundingDoc) bool {
	localBase := strings.TrimRight(s.baseURL, "/")
	if instanceURL := strings.TrimRight(fd.InstanceURL, "/"); instanceURL != "" {
		return instanceURL == localBase
	}
	return s.isLocalInstanceID(fd.InstanceID)
}

func normalizeAdminObjectType(objectType string) string {
	switch strings.ToLower(strings.TrimSpace(objectType)) {
	case "user", "users":
		return "users"
	case "event", "events":
		return "events"
	case "calendar", "calendars":
		return "calendars"
	case "blob", "blobs":
		return "blobs"
	default:
		return ""
	}
}

func normalizeAdminRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner":
		return "owner"
	case "admin":
		return "admin"
	default:
		return ""
	}
}

func queryInt(r *http.Request, key string, fallback, min, max int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
