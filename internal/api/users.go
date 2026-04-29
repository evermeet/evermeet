package api

import (
	"net/http"
	"github.com/go-chi/chi/v5"
)

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	did := chi.URLParam(r, "did")
	if did == "" {
		jsonErr(w, http.StatusBadRequest, "did required")
		return
	}

	ctx := r.Context()
	user, err := s.db.GetUser(ctx, did)
	if (err != nil || user == nil) && s.node != nil {
		// Try fetching from P2P
		if err := s.node.FetchUser(did); err == nil {
			user, err = s.db.GetUser(ctx, did)
		}
	}

	if err != nil || user == nil {
		jsonErr(w, http.StatusNotFound, "user not found")
		return
	}

	jsonOK(w, map[string]any{
		"did":          user.DID,
		"display_name": user.DisplayName,
		"avatar":       user.Avatar,
		"bio":          user.Bio,
		"home_host":    user.HomeHost,
		"updated_at":   user.UpdatedAt,
	})
}
