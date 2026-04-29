package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/evermeet/evermeet/internal/store"
	"github.com/go-chi/chi/v5"
)

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	did := chi.URLParam(r, "did")
	user, err := s.db.GetUser(r.Context(), did)
	if err != nil || user == nil {
		jsonErr(w, http.StatusNotFound, "user not found")
		return
	}
	jsonOK(w, map[string]any{
		"did":          user.DID,
		"display_name": user.DisplayName,
		"avatar":       user.Avatar,
		"bio":          user.Bio,
		"endpoint":     user.Endpoint,
	})
}

func (s *Server) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	ctx := r.Context()

	var req struct {
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
		Bio         string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	user, err := s.db.GetUser(ctx, did)
	if err != nil || user == nil {
		jsonErr(w, http.StatusNotFound, "user not found")
		return
	}

	user.DisplayName = req.DisplayName
	user.Avatar = req.Avatar
	user.Bio = req.Bio
	user.UpdatedAt = time.Now()

	if err := s.db.UpsertUser(ctx, user); err != nil {
		jsonErr(w, http.StatusInternalServerError, "update failed")
		return
	}

	jsonOK(w, &store.User{
		DID:         user.DID,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		Bio:         user.Bio,
	})
}
