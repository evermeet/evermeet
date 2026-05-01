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
	if err != nil || user == nil {
		if n := s.libp2pNode(); n != nil {
			// Try fetching from P2P
			if err := n.FetchUser(did); err == nil {
				user, err = s.db.GetUser(ctx, did)
			}
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
		"instance_id":  user.InstanceID,
		"updated_at":   user.UpdatedAt,
	})
}
