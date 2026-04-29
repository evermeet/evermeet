package api

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/evermeet/evermeet/internal/identity"
	"github.com/evermeet/evermeet/internal/store"
)

func (s *Server) handleMagicLinkRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		jsonErr(w, http.StatusBadRequest, "email required")
		return
	}

	ctx := r.Context()

	priv, did, err := s.lookupOrCreateUser(ctx, req.Email)
	if err != nil {
		s.log.Printf("magic link lookup: %v", err)
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	_ = priv

	token := randomHex(32)
	ml := &store.MagicLink{
		Token:     token,
		Email:     req.Email,
		DID:       did,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.db.InsertMagicLink(ctx, ml); err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", s.baseURL, token)
	if s.email != nil {
		if err := s.email.SendMagicLink(req.Email, verifyURL); err != nil {
			s.log.Printf("send magic link to %s: %v", req.Email, err)
		}
	} else {
		s.log.Printf("MAGIC LINK (dev): %s", verifyURL)
	}

	jsonOK(w, map[string]string{"status": "sent"})
}

func (s *Server) handleMagicLinkVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		jsonErr(w, http.StatusBadRequest, "token required")
		return
	}

	ctx := r.Context()
	ml, err := s.db.GetMagicLink(ctx, token)
	if err != nil || ml == nil || ml.Used || time.Now().After(ml.ExpiresAt) {
		jsonErr(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	if err := s.db.MarkMagicLinkUsed(ctx, token); err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	sessionToken := randomHex(32)
	sess := &store.Session{
		Token:     sessionToken,
		DID:       ml.DID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.db.InsertSession(ctx, sess); err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  sess.ExpiresAt,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	if did == "" {
		jsonErr(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	ctx := r.Context()
	user, err := s.db.GetUser(ctx, did)
	if err != nil || user == nil {
		jsonErr(w, http.StatusInternalServerError, "user not found")
		return
	}

	jsonOK(w, map[string]any{
		"did":          user.DID,
		"display_name": user.DisplayName,
		"avatar":       user.Avatar,
		"bio":          user.Bio,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		s.db.DeleteSession(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
	jsonOK(w, map[string]string{"status": "ok"})
}

// lookupOrCreateUser finds or creates a user for the given email.
// Returns the user's Ed25519 private key (decrypted from custodial storage) and their DID.
func (s *Server) lookupOrCreateUser(ctx context.Context, email string) (ed25519.PrivateKey, string, error) {
	existing, err := s.db.GetUserPrivateByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}

	if existing != nil {
		password := identity.SessionPassword(s.serverSecret, existing.DID)
		priv, err := identity.DecryptKeypair(existing.EncSigningKey, password)
		if err != nil {
			return nil, "", fmt.Errorf("decrypt keypair for %s: %w", existing.DID, err)
		}
		return priv, existing.DID, nil
	}

	// New user — generate keypair and persist everything.
	kp, err := identity.Generate()
	if err != nil {
		return nil, "", err
	}
	did := identity.DeriveDID(kp.SigningPub)
	password := identity.SessionPassword(s.serverSecret, did)

	encKey, err := identity.EncryptKeypair(kp.SigningPriv, password)
	if err != nil {
		return nil, "", err
	}

	genesisOp, genesisHash, err := identity.BuildGenesisOp(kp.SigningPriv, kp.RotationPriv, s.baseURL)
	if err != nil {
		return nil, "", err
	}

	if err := s.db.AppendKELOp(ctx, &store.KELOp{
		Hash:      genesisHash,
		DID:       did,
		Type:      string(genesisOp.Type),
		Payload:   mustJSON(genesisOp),
		Seq:       0,
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, "", err
	}

	if err := s.db.UpsertUser(ctx, &store.User{
		DID:        did,
		CurrentPK:  hex.EncodeToString(kp.SigningPub),
		RotationPK: hex.EncodeToString(kp.RotationPub),
		Endpoint:   s.baseURL,
		Sig:        genesisHash,
		UpdatedAt:  time.Now(),
	}); err != nil {
		return nil, "", err
	}

	if err := s.db.UpsertUserPrivate(ctx, &store.UserPrivate{
		DID:           did,
		Email:         email,
		EmailVerified: false,
		EncSigningKey: encKey,
	}); err != nil {
		return nil, "", err
	}

	return kp.SigningPriv, did, nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
