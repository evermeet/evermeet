package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/evermeet/evermeet/internal/store"
)

func (s *Server) setupGate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/setup/status" || r.URL.Path == "/api/setup/complete" {
			next.ServeHTTP(w, r)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		required, err := s.setupRequired(r.Context())
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, "setup check failed")
			return
		}
		if required {
			jsonErr(w, http.StatusLocked, "setup required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) setupRequired(ctx context.Context) (bool, error) {
	hasAdmins, err := s.db.HasAdminAccounts(ctx)
	if err != nil {
		return false, err
	}
	return !hasAdmins, nil
}

func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	required, err := s.setupRequired(r.Context())
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "setup check failed")
		return
	}
	jsonOK(w, map[string]bool{"required": required})
}

func (s *Server) handleSetupComplete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid request")
		return
	}
	req.Token = strings.TrimSpace(req.Token)
	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.Token == "" || req.Email == "" {
		jsonErr(w, http.StatusBadRequest, "token and email required")
		return
	}

	s.setupMu.Lock()
	defer s.setupMu.Unlock()

	required, err := s.setupRequired(r.Context())
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "setup check failed")
		return
	}
	if !required {
		jsonErr(w, http.StatusConflict, "setup already completed")
		return
	}
	if s.setupToken == "" || subtle.ConstantTimeCompare([]byte(req.Token), []byte(s.setupToken)) != 1 {
		jsonErr(w, http.StatusUnauthorized, "invalid setup token")
		return
	}

	_, did, err := s.lookupOrCreateUser(r.Context(), req.Email)
	if err != nil {
		s.log.Printf("setup user create: %v", err)
		jsonErr(w, http.StatusInternalServerError, "create admin failed")
		return
	}
	if err := s.db.SetUserEmailVerified(r.Context(), did, true); err != nil {
		jsonErr(w, http.StatusInternalServerError, "verify admin email failed")
		return
	}
	if req.DisplayName != "" {
		if err := s.setUserDisplayName(r.Context(), did, req.DisplayName); err != nil {
			jsonErr(w, http.StatusInternalServerError, "update admin profile failed")
			return
		}
	}
	if err := s.db.InsertAdminAccount(r.Context(), &store.AdminAccount{
		DID:       did,
		CreatedAt: time.Now(),
	}); err != nil {
		jsonErr(w, http.StatusInternalServerError, "create admin failed")
		return
	}

	s.setupToken = ""
	s.createSession(w, r.Context(), did)
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) setUserDisplayName(ctx context.Context, did, displayName string) error {
	user, err := s.db.GetUser(ctx, did)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	user.DisplayName = displayName
	user.UpdatedAt = time.Now()
	return s.db.UpsertUser(ctx, user)
}
