package api

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/evermeet/evermeet/internal/identity"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/go-webauthn/webauthn/webauthn"
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
	pollToken := randomHex(32)
	ml := &store.MagicLink{
		Token:     token,
		PollToken: pollToken,
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

	jsonOK(w, map[string]string{"status": "sent", "poll_token": pollToken})
}

func (s *Server) handleMagicLinkVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		jsonErr(w, http.StatusBadRequest, "token required")
		return
	}

	ctx := r.Context()
	ml, err := s.db.GetMagicLink(ctx, token)
	if err != nil || ml == nil || time.Now().After(ml.ExpiresAt) {
		jsonErr(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	if !ml.Used {
		if err := s.db.MarkMagicLinkUsed(ctx, token); err != nil {
			jsonErr(w, http.StatusInternalServerError, "internal error")
			return
		}
	}
	sess, err := s.ensureMagicLinkSession(ctx, ml)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	s.setSessionCookie(w, sess)

	http.Redirect(w, r, "/auth/verify?approved=1", http.StatusFound)
}

func (s *Server) handleMagicLinkStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PollToken string `json:"poll_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PollToken == "" {
		jsonErr(w, http.StatusBadRequest, "poll_token required")
		return
	}

	ctx := r.Context()
	ml, err := s.db.GetMagicLinkByPollToken(ctx, req.PollToken)
	if err != nil || ml == nil {
		jsonErr(w, http.StatusNotFound, "magic link request not found")
		return
	}
	if time.Now().After(ml.ExpiresAt) {
		jsonErr(w, http.StatusUnauthorized, "magic link expired")
		return
	}
	if !ml.Used {
		jsonOK(w, map[string]string{"status": "pending"})
		return
	}

	sess, err := s.ensureMagicLinkSession(ctx, ml)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	s.setSessionCookie(w, sess)

	jsonOK(w, map[string]string{"status": "signed_in"})
}

func (s *Server) ensureMagicLinkSession(ctx context.Context, ml *store.MagicLink) (*store.Session, error) {
	sessionToken := ml.SessionToken
	if sessionToken == "" {
		sessionToken = randomHex(32)
		sess := &store.Session{
			Token:     sessionToken,
			DID:       ml.DID,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}
		if err := s.db.InsertSession(ctx, sess); err != nil {
			return nil, err
		}
		if ml.PollToken != "" {
			if err := s.db.SetMagicLinkSession(ctx, ml.PollToken, sessionToken); err != nil {
				return nil, err
			}
		}
		return sess, nil
	}

	sess, err := s.db.GetSession(ctx, sessionToken)
	if err != nil || sess == nil {
		return nil, fmt.Errorf("magic link session missing")
	}
	return sess, nil
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

	homeInstanceURL := user.Endpoint
	if homeInstanceURL == "" {
		homeInstanceURL = s.baseURL
	}
	isLocal := strings.TrimRight(homeInstanceURL, "/") == strings.TrimRight(s.baseURL, "/")
	isAdmin, err := s.db.IsAdmin(ctx, did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "admin lookup failed")
		return
	}
	adminRole := ""
	isOwner := false
	if isAdmin {
		adminRole, err = s.db.GetAdminRole(ctx, did)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, "admin role lookup failed")
			return
		}
		isOwner = adminRole == "owner"
	}

	jsonOK(w, map[string]any{
		"did":               user.DID,
		"display_name":      user.DisplayName,
		"avatar":            user.Avatar,
		"bio":               user.Bio,
		"is_local":          isLocal,
		"auth_kind":         map[bool]string{true: "local", false: "remote"}[isLocal],
		"home_instance_url": homeInstanceURL,
		"is_admin":          isAdmin,
		"admin_role":        adminRole,
		"is_owner":          isOwner,
	})
}

func (s *Server) handleAuthMethods(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	if did == "" {
		jsonErr(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	ctx := r.Context()
	private, err := s.db.GetUserPrivate(ctx, did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "private user lookup failed")
		return
	}
	passkeys, err := s.db.GetPasskeysByDID(ctx, did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "passkey lookup failed")
		return
	}

	passkeyItems := make([]map[string]any, 0, len(passkeys))
	for _, passkey := range passkeys {
		passkeyItems = append(passkeyItems, map[string]any{
			"id":               base64.RawURLEncoding.EncodeToString(passkey.ID),
			"attestation_type": passkey.AttestationType,
			"counter":          passkey.Counter,
			"backup_eligible":  passkey.BackupEligible,
			"backup_state":     passkey.BackupState,
			"user_verified":    passkey.UserVerified,
			"user_present":     passkey.UserPresent,
			"created_at":       passkey.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	email := ""
	emailVerified := false
	siweChainID := ""
	siweAddress := ""
	if private != nil {
		email = private.Email
		emailVerified = private.EmailVerified
		siweChainID = private.SIWEChainID
		siweAddress = private.SIWEAddress
	}

	jsonOK(w, map[string]any{
		"email": map[string]any{
			"address":  email,
			"verified": emailVerified,
			"linked":   email != "",
		},
		"ethereum": map[string]any{
			"chain_id": siweChainID,
			"address":  siweAddress,
			"linked":   siweChainID != "" && siweAddress != "",
		},
		"passkeys": passkeyItems,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := s.sessionCookie(r); err == nil {
		s.db.DeleteSession(r.Context(), cookie.Value)
	}
	s.clearSessionCookie(w)
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	if did == "" {
		jsonErr(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
		Bio         string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid request")
		return
	}

	ctx := r.Context()
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
		jsonErr(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	if priv := authPrivKey(r); len(priv) > 0 {
		if n := s.libp2pNode(); n != nil {
			if err := n.BroadcastUserProfile(user, priv); err != nil {
				s.log.Printf("broadcast profile: %v", err)
			}
		}
	}

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

	genesisOp, genesisHash, err := identity.BuildGenesisOp(kp.SigningPriv, kp.RotationPriv, s.homeHost())
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
		InstanceID: s.homeHost(),
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

	// Publish email → home instance mapping to the DHT so foreign instances
	// can discover where this user's home is. Fire-and-forget: a failure here
	// does not block registration; the heartbeat will retry within 12h.
	if pub := s.publisher(); pub != nil {
		go func() {
			pubCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := pub.Publish(pubCtx, email); err != nil {
				s.log.Printf("dht publish for new user %s: %v", did, err)
			}
		}()
	}

	return kp.SigningPriv, did, nil
}

func (s *Server) handlePasskeyRegisterStart(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	ctx := r.Context()

	user, err := s.db.GetUser(ctx, did)
	if err != nil || user == nil {
		jsonErr(w, http.StatusNotFound, "user not found")
		return
	}

	passkeys, err := s.db.GetPasskeysByDID(ctx, did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	waUser := s.wrapUser(user, passkeys)
	options, sessionData, err := s.webauthn.BeginRegistration(waUser)
	if err != nil {
		s.log.Printf("webauthn registration start: %v", err)
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	token := randomHex(32)
	sd, _ := json.Marshal(sessionData)
	ws := &store.WebAuthnSession{
		Token:     token,
		DID:       did,
		Data:      sd,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if err := s.db.InsertWebAuthnSession(ctx, ws); err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("X-WebAuthn-Session", token)
	jsonOK(w, options)
}

func (s *Server) handlePasskeyRegisterFinish(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-WebAuthn-Session")
	if token == "" {
		jsonErr(w, http.StatusBadRequest, "session token required")
		return
	}

	ctx := r.Context()
	ws, err := s.db.GetWebAuthnSession(ctx, token)
	if err != nil || ws == nil || time.Now().After(ws.ExpiresAt) {
		jsonErr(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	defer s.db.DeleteWebAuthnSession(ctx, token)

	var sessionData webauthn.SessionData
	if err := json.Unmarshal(ws.Data, &sessionData); err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	user, err := s.db.GetUser(ctx, ws.DID)
	if err != nil || user == nil {
		jsonErr(w, http.StatusNotFound, "user not found")
		return
	}

	passkeys, err := s.db.GetPasskeysByDID(ctx, ws.DID)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	waUser := s.wrapUser(user, passkeys)
	credential, err := s.webauthn.FinishRegistration(waUser, sessionData, r)
	if err != nil {
		s.log.Printf("webauthn registration finish: %v", err)
		jsonErr(w, http.StatusBadRequest, "registration failed")
		return
	}

	p := &store.Passkey{
		ID:              credential.ID,
		DID:             ws.DID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Counter:         int64(credential.Authenticator.SignCount),
		BackupEligible:  credential.Flags.BackupEligible,
		BackupState:     credential.Flags.BackupState,
		UserPresent:     true,
		UserVerified:    true,
		CreatedAt:       time.Now(),
	}
	if err := s.db.InsertPasskey(ctx, p); err != nil {
		jsonErr(w, http.StatusInternalServerError, "failed to save passkey")
		return
	}

	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handlePasskeySignupStart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Generate a new custodial identity for this prospective user
	kp, err := identity.Generate()
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "identity generation failed")
		return
	}
	did := identity.DeriveDID(kp.SigningPub)

	// 2. Prepare WebAuthn registration
	waUser := &waUser{user: &store.User{DID: did}}
	options, sessionData, err := s.webauthn.BeginRegistration(waUser)
	if err != nil {
		s.log.Printf("webauthn signup start: %v", err)
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	// 3. Encrypt the new signing key immediately (custodial Level 1)
	password := identity.SessionPassword(s.serverSecret, did)
	encKey, err := identity.EncryptKeypair(kp.SigningPriv, password)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "key encryption failed")
		return
	}

	// 4. Store session state along with the encrypted key to be persisted on finish
	token := randomHex(32)
	state := map[string]any{
		"sd":  sessionData,
		"enc": encKey,
		"pk":  hex.EncodeToString(kp.SigningPub),
		"rp":  hex.EncodeToString(kp.RotationPub),
	}
	sd, _ := json.Marshal(state)
	ws := &store.WebAuthnSession{
		Token:     token,
		DID:       did,
		Data:      sd,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if err := s.db.InsertWebAuthnSession(ctx, ws); err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("X-WebAuthn-Session", token)
	jsonOK(w, options)
}

func (s *Server) handlePasskeySignupFinish(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-WebAuthn-Session")
	if token == "" {
		jsonErr(w, http.StatusBadRequest, "session token required")
		return
	}

	ctx := r.Context()
	ws, err := s.db.GetWebAuthnSession(ctx, token)
	if err != nil || ws == nil || time.Now().After(ws.ExpiresAt) {
		jsonErr(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	defer s.db.DeleteWebAuthnSession(ctx, token)

	var state struct {
		SD  webauthn.SessionData `json:"sd"`
		Enc string               `json:"enc"`
		PK  string               `json:"pk"`
		RP  string               `json:"rp"`
	}
	if err := json.Unmarshal(ws.Data, &state); err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	waUser := &waUser{user: &store.User{DID: ws.DID}}
	credential, err := s.webauthn.FinishRegistration(waUser, state.SD, r)
	if err != nil {
		s.log.Printf("webauthn signup finish: %v", err)
		jsonErr(w, http.StatusBadRequest, "registration failed")
		return
	}

	// Persist the new user!
	genesisHash := "signup-" + randomHex(8) // Placeholder for real KEL genesis
	if err := s.db.UpsertUser(ctx, &store.User{
		DID:        ws.DID,
		CurrentPK:  state.PK,
		RotationPK: state.RP,
		Endpoint:   s.baseURL,
		Sig:        genesisHash,
		UpdatedAt:  time.Now(),
		InstanceID: s.homeHost(),
	}); err != nil {
		jsonErr(w, http.StatusInternalServerError, "failed to save user")
		return
	}

	if err := s.db.UpsertUserPrivate(ctx, &store.UserPrivate{
		DID:           ws.DID,
		EncSigningKey: state.Enc,
	}); err != nil {
		jsonErr(w, http.StatusInternalServerError, "failed to save private data")
		return
	}

	p := &store.Passkey{
		ID:              credential.ID,
		DID:             ws.DID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Counter:         int64(credential.Authenticator.SignCount),
		BackupEligible:  credential.Flags.BackupEligible,
		BackupState:     credential.Flags.BackupState,
		UserPresent:     true,
		UserVerified:    true,
		CreatedAt:       time.Now(),
	}
	if err := s.db.InsertPasskey(ctx, p); err != nil {
		jsonErr(w, http.StatusInternalServerError, "failed to save passkey")
		return
	}

	// Create session and log them in
	s.createSession(w, ctx, ws.DID)
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handlePasskeyLoginStart(w http.ResponseWriter, r *http.Request) {
	// Discovery-based login (no email needed)
	options, sessionData, err := s.webauthn.BeginDiscoverableLogin()
	if err != nil {
		s.log.Printf("webauthn login start: %v", err)
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	token := randomHex(32)
	sd, _ := json.Marshal(sessionData)
	ws := &store.WebAuthnSession{
		Token:     token,
		DID:       "", // will be filled on finish
		Data:      sd,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if err := s.db.InsertWebAuthnSession(r.Context(), ws); err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("X-WebAuthn-Session", token)
	jsonOK(w, options)
}

func (s *Server) handlePasskeyLoginFinish(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-WebAuthn-Session")
	if token == "" {
		jsonErr(w, http.StatusBadRequest, "session token required")
		return
	}

	ctx := r.Context()
	ws, err := s.db.GetWebAuthnSession(ctx, token)
	if err != nil || ws == nil || time.Now().After(ws.ExpiresAt) {
		jsonErr(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	defer s.db.DeleteWebAuthnSession(ctx, token)

	var sessionData webauthn.SessionData
	if err := json.Unmarshal(ws.Data, &sessionData); err != nil {
		jsonErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	// FinishPasskeyLogin uses the discoverable handler to find the user
	user, credential, err := s.webauthn.FinishPasskeyLogin(func(rawID, userHandle []byte) (webauthn.User, error) {
		did := string(userHandle)
		s.log.Printf("discoverable login userHandle: %s", did)
		u, err := s.db.GetUser(context.Background(), did)
		if err != nil || u == nil {
			s.log.Printf("discoverable login: user %s not found", did)
			return nil, fmt.Errorf("user not found")
		}
		pk, err := s.db.GetPasskeysByDID(context.Background(), did)
		if err != nil {
			s.log.Printf("discoverable login: failed to get passkeys for %s: %v", did, err)
			return nil, err
		}
		return s.wrapUser(u, pk), nil
	}, sessionData, r)

	if err != nil {
		s.log.Printf("webauthn login finish error: %v", err)
		jsonErr(w, http.StatusUnauthorized, "login failed")
		return
	}

	wa := user.(*waUser)
	s.db.UpdatePasskeyCounter(ctx, credential.ID, int64(credential.Authenticator.SignCount))
	s.db.UpdatePasskeyBackupFlags(ctx, credential.ID, credential.Flags.BackupEligible, credential.Flags.BackupState)
	s.createSession(w, ctx, wa.user.DID)

	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) createSession(w http.ResponseWriter, ctx context.Context, did string) {
	s.createSessionWithDuration(w, ctx, did, 30*24*time.Hour)
}

func (s *Server) createSessionWithDuration(w http.ResponseWriter, ctx context.Context, did string, duration time.Duration) {
	sessionToken := randomHex(32)
	sess := &store.Session{
		Token:     sessionToken,
		DID:       did,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}
	if err := s.db.InsertSession(ctx, sess); err != nil {
		s.log.Printf("failed to insert session: %v", err)
		return
	}

	s.setSessionCookie(w, sess)
}

const legacySessionCookieName = "session"

func (s *Server) sessionCookieName() string {
	if s.instanceID == "" {
		return legacySessionCookieName
	}
	return "session_" + s.instanceID
}

func (s *Server) sessionCookie(r *http.Request) (*http.Cookie, error) {
	if cookie, err := r.Cookie(s.sessionCookieName()); err == nil {
		return cookie, nil
	}
	return r.Cookie(legacySessionCookieName)
}

func (s *Server) setSessionCookie(w http.ResponseWriter, sess *store.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.sessionCookieName(),
		Value:    sess.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  sess.ExpiresAt,
	})

	if s.sessionCookieName() != legacySessionCookieName {
		http.SetCookie(w, &http.Cookie{
			Name:    legacySessionCookieName,
			Value:   "",
			Path:    "/",
			MaxAge:  -1,
			Expires: time.Unix(0, 0),
		})
	}
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	for _, name := range []string{s.sessionCookieName(), legacySessionCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:    name,
			Value:   "",
			Path:    "/",
			MaxAge:  -1,
			Expires: time.Unix(0, 0),
		})
	}
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
