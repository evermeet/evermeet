package api

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/evermeet/evermeet/internal/identity"
	"github.com/evermeet/evermeet/internal/routing"
	"github.com/evermeet/evermeet/internal/store"
)

// handleResolveHome is called by a foreign instance when a user types their
// email and wants to authenticate. It:
//  1. Computes the email hash
//  2. Queries the DHT for the home instance URL
//  3. Fetches the home instance's public key and verifies the DHT record signature
//  4. Issues a short-lived nonce for this cross-instance session
//
// POST /api/auth/resolve-home
// Body: {"email": "user@example.com", "event_id": "optional"}
// Response: {"home_instance_url": "https://...", "nonce": "..."}
func (s *Server) handleResolveHome(w http.ResponseWriter, r *http.Request) {
	if s.node == nil {
		jsonErr(w, http.StatusServiceUnavailable, "p2p not available")
		return
	}

	var req struct {
		Email   string `json:"email"`
		EventID string `json:"event_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		jsonErr(w, http.StatusBadRequest, "email required")
		return
	}

	// 1. Compute email hash and query DHT.
	emailHash := routing.EmailHash(req.Email)

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	data, err := s.node.DHTPeerLookup(ctx, emailHash)
	if err != nil {
		// Not found in DHT — could be a new user or a peer-less network.
		jsonErr(w, http.StatusNotFound, "home instance not found for this email")
		return
	}

	// 2. Unmarshal the record.
	rec, err := routing.UnmarshalRecord(data)
	if err != nil {
		jsonErr(w, http.StatusBadGateway, "malformed home record")
		return
	}

	// 3. Fetch the home instance's public key and verify the record.
	instancePub, err := fetchInstancePubKey(ctx, rec.HomeInstanceURL)
	if err != nil {
		s.log.Printf("resolve-home: fetch instance key %s: %v", rec.HomeInstanceURL, err)
		jsonErr(w, http.StatusBadGateway, "could not verify home instance")
		return
	}
	if err := routing.VerifyRecord(rec, instancePub); err != nil {
		s.log.Printf("resolve-home: invalid record signature from %s: %v", rec.HomeInstanceURL, err)
		jsonErr(w, http.StatusBadGateway, "home record signature invalid")
		return
	}

	// 4. Issue a nonce — tied to this foreign instance's base URL and event.
	nonce, err := generateNonce()
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "nonce generation failed")
		return
	}
	now := time.Now()
	if err := s.db.InsertNonce(r.Context(), &store.CrossInstanceNonce{
		Nonce:      nonce,
		ForeignURL: s.baseURL,
		EventID:    req.EventID,
		CreatedAt:  now,
		ExpiresAt:  now.Add(5 * time.Minute),
	}); err != nil {
		jsonErr(w, http.StatusInternalServerError, "nonce storage failed")
		return
	}

	foreignSig, err := s.signForeignInstanceSig(&ForeignInstanceSig{
		ReturnTo: s.baseURL,
		Nonce:    nonce,
		EventID:  req.EventID,
		IssuedAt: now.Unix(),
	})
	if err != nil {
		s.log.Printf("resolve-home: sign foreign redirect: %v", err)
		jsonErr(w, http.StatusInternalServerError, "redirect signing failed")
		return
	}
	delegateURL, err := buildDelegateURL(rec.HomeInstanceURL, foreignSig)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "delegate URL failed")
		return
	}

	jsonOK(w, map[string]any{
		"home_instance_url": rec.HomeInstanceURL,
		"nonce":             nonce,
		"return_to":         s.baseURL,
		"foreign_sig":       foreignSig,
		"delegate_url":      delegateURL,
	})
}

// fetchInstancePubKey fetches and returns the Ed25519 public key from a
// remote instance's /.well-known/evermeet-node-key endpoint.
func fetchInstancePubKey(ctx context.Context, instanceURL string) (ed25519.PublicKey, error) {
	url := strings.TrimRight(instanceURL, "/") + "/.well-known/evermeet-node-key"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("well-known returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}

	var payload struct {
		PublicKey []byte `json:"public_key"` // raw marshaled libp2p key bytes (base64 by JSON)
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse well-known: %w", err)
	}

	// The well-known endpoint returns libp2p-marshaled public key bytes.
	// Extract the raw 32-byte Ed25519 public key.
	pub, err := extractEd25519PubKey(payload.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("extract ed25519 key: %w", err)
	}
	return pub, nil
}

// extractEd25519PubKey extracts a raw ed25519.PublicKey from libp2p's
// protobuf-marshaled public key bytes (as returned by crypto.MarshalPublicKey).
// The libp2p format is: varint(key_type) + varint(key_len) + key_bytes.
// For Ed25519 the last 32 bytes are the raw public key.
func extractEd25519PubKey(b []byte) (ed25519.PublicKey, error) {
	if len(b) < ed25519.PublicKeySize {
		return nil, fmt.Errorf("key too short: %d bytes", len(b))
	}
	// libp2p marshals Ed25519 public keys as a protobuf with the raw key
	// as the last 32 bytes regardless of the varint prefix length.
	raw := b[len(b)-ed25519.PublicKeySize:]
	return ed25519.PublicKey(raw), nil
}

func generateNonce() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ForeignInstanceSig is the signed envelope a foreign instance attaches to
// the redirect URL so the home instance can verify the return_to URL was
// legitimately issued by a real Evermeet instance and not crafted by an
// attacker (open redirect prevention).
type ForeignInstanceSig struct {
	ReturnTo string `json:"return_to"` // foreign instance base URL
	Nonce    string `json:"nonce"`
	EventID  string `json:"event_id"`
	IssuedAt int64  `json:"issued_at"` // Unix seconds
	Sig      string `json:"sig"`       // base64url Ed25519 sig over above fields
}

// VerifyForeignSig verifies that a ForeignInstanceSig was signed by the
// instance whose public key is at return_to/.well-known/evermeet-node-key.
func VerifyForeignSig(ctx context.Context, fs *ForeignInstanceSig) error {
	if fs == nil || fs.ReturnTo == "" || fs.Nonce == "" || fs.Sig == "" {
		return fmt.Errorf("foreign instance signature incomplete")
	}
	issuedAt := time.Unix(fs.IssuedAt, 0)
	if time.Since(issuedAt) > 5*time.Minute || issuedAt.After(time.Now().Add(time.Minute)) {
		return fmt.Errorf("foreign instance signature expired")
	}

	pub, err := fetchInstancePubKey(ctx, fs.ReturnTo)
	if err != nil {
		return fmt.Errorf("fetch foreign instance key: %w", err)
	}

	// Reconstruct the payload that was signed (exclude Sig field).
	payload, err := foreignSigPayload(fs)
	if err != nil {
		return err
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(fs.Sig)
	if err != nil {
		return fmt.Errorf("decode sig: %w", err)
	}
	if !ed25519.Verify(pub, payload, sigBytes) {
		return fmt.Errorf("foreign instance signature invalid")
	}
	return nil
}

type DelegationToken struct {
	Sub     string `json:"sub"`
	Aud     string `json:"aud"`
	Iat     int64  `json:"iat"`
	Exp     int64  `json:"exp"`
	Nonce   string `json:"nonce"`
	EventID string `json:"event_id"`
}

type SignedDelegationToken struct {
	Token           DelegationToken `json:"token"`
	Sig             string          `json:"sig"`
	PublicKey       string          `json:"public_key"`
	HomeInstanceURL string          `json:"home_instance_url"`
}

func (s *Server) handleCreateDelegation(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	priv := authPrivKey(r)
	if did == "" || len(priv) == 0 {
		jsonErr(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req struct {
		ReturnTo   string             `json:"return_to"`
		Nonce      string             `json:"nonce"`
		EventID    string             `json:"event_id"`
		ForeignSig ForeignInstanceSig `json:"foreign_sig"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.ReturnTo == "" || req.Nonce == "" {
		jsonErr(w, http.StatusBadRequest, "return_to and nonce required")
		return
	}
	if req.ForeignSig.ReturnTo != req.ReturnTo || req.ForeignSig.Nonce != req.Nonce || req.ForeignSig.EventID != req.EventID {
		jsonErr(w, http.StatusBadRequest, "foreign signature does not match request")
		return
	}
	if err := VerifyForeignSig(r.Context(), &req.ForeignSig); err != nil {
		s.log.Printf("delegate: invalid foreign signature: %v", err)
		jsonErr(w, http.StatusBadRequest, "invalid return target")
		return
	}

	user, err := s.db.GetUser(r.Context(), did)
	if err != nil || user == nil {
		jsonErr(w, http.StatusInternalServerError, "user not found")
		return
	}

	now := time.Now()
	token := DelegationToken{
		Sub:     did,
		Aud:     strings.TrimRight(req.ReturnTo, "/"),
		Iat:     now.Unix(),
		Exp:     now.Add(5 * time.Minute).Unix(),
		Nonce:   req.Nonce,
		EventID: req.EventID,
	}
	sig, err := identity.Sign(priv, token)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "token signing failed")
		return
	}

	jsonOK(w, SignedDelegationToken{
		Token:           token,
		Sig:             sig,
		PublicKey:       user.CurrentPK,
		HomeInstanceURL: s.baseURL,
	})
}

func (s *Server) handleVerifyDelegation(w http.ResponseWriter, r *http.Request) {
	req, formPost, err := decodeDelegationVerifyRequest(r)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, err.Error())
		return
	}

	pubBytes, err := hex.DecodeString(req.PublicKey)
	if err != nil || len(pubBytes) != ed25519.PublicKeySize {
		jsonErr(w, http.StatusBadRequest, "invalid public key")
		return
	}
	pub := ed25519.PublicKey(pubBytes)
	if identity.DeriveDID(pub) != req.Token.Sub {
		jsonErr(w, http.StatusBadRequest, "public key does not match DID")
		return
	}
	valid, err := identity.Verify(pub, req.Token, req.Sig)
	if err != nil || !valid {
		jsonErr(w, http.StatusUnauthorized, "invalid delegation token")
		return
	}

	now := time.Now()
	if req.Token.Aud != strings.TrimRight(s.baseURL, "/") {
		jsonErr(w, http.StatusUnauthorized, "delegation audience mismatch")
		return
	}
	if now.Unix() > req.Token.Exp || req.Token.Iat > now.Add(time.Minute).Unix() {
		jsonErr(w, http.StatusUnauthorized, "delegation token expired")
		return
	}

	nonce, err := s.db.GetNonce(r.Context(), req.Token.Nonce)
	if err != nil || nonce == nil || nonce.Used || now.After(nonce.ExpiresAt) {
		jsonErr(w, http.StatusUnauthorized, "invalid or expired nonce")
		return
	}
	if nonce.ForeignURL != s.baseURL || nonce.EventID != req.Token.EventID {
		jsonErr(w, http.StatusUnauthorized, "nonce does not match delegation")
		return
	}
	if err := s.db.MarkNonceUsed(r.Context(), req.Token.Nonce); err != nil {
		jsonErr(w, http.StatusInternalServerError, "nonce update failed")
		return
	}
	if err := s.upsertDelegatedUser(r.Context(), req); err != nil {
		jsonErr(w, http.StatusInternalServerError, "user storage failed")
		return
	}

	s.createSessionWithDuration(w, r.Context(), req.Token.Sub, 24*time.Hour)
	if formPost {
		redirect := "/"
		if req.Token.EventID != "" {
			redirect = "/events/" + url.PathEscape(req.Token.EventID)
		}
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}
	jsonOK(w, map[string]string{"status": "ok"})
}

func decodeDelegationVerifyRequest(r *http.Request) (SignedDelegationToken, bool, error) {
	var req SignedDelegationToken
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") || strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseForm(); err != nil {
			return req, true, fmt.Errorf("invalid form")
		}
		payload := r.FormValue("payload")
		if payload == "" {
			return req, true, fmt.Errorf("payload required")
		}
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			return req, true, fmt.Errorf("invalid payload")
		}
		return req, true, nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, false, fmt.Errorf("invalid request")
	}
	return req, false, nil
}

func (s *Server) upsertDelegatedUser(ctx context.Context, req SignedDelegationToken) error {
	existing, err := s.db.GetUser(ctx, req.Token.Sub)
	if err != nil {
		return err
	}
	now := time.Now()
	if existing == nil {
		return s.db.UpsertUser(ctx, &store.User{
			DID:       req.Token.Sub,
			CurrentPK: req.PublicKey,
			Endpoint:  req.HomeInstanceURL,
			UpdatedAt: now,
		})
	}
	if existing.CurrentPK != "" && existing.CurrentPK != req.PublicKey {
		return fmt.Errorf("delegated public key changed for %s", req.Token.Sub)
	}
	existing.CurrentPK = req.PublicKey
	if req.HomeInstanceURL != "" {
		existing.Endpoint = req.HomeInstanceURL
	}
	if existing.UpdatedAt.IsZero() {
		existing.UpdatedAt = now
	}
	return s.db.UpsertUser(ctx, existing)
}

func (s *Server) signForeignInstanceSig(fs *ForeignInstanceSig) (*ForeignInstanceSig, error) {
	payload, err := foreignSigPayload(fs)
	if err != nil {
		return nil, err
	}
	out := *fs
	out.Sig = base64.RawURLEncoding.EncodeToString(ed25519.Sign(s.instancePriv, payload))
	return &out, nil
}

func foreignSigPayload(fs *ForeignInstanceSig) ([]byte, error) {
	return json.Marshal(struct {
		ReturnTo string `json:"return_to"`
		Nonce    string `json:"nonce"`
		EventID  string `json:"event_id"`
		IssuedAt int64  `json:"issued_at"`
	}{fs.ReturnTo, fs.Nonce, fs.EventID, fs.IssuedAt})
}

func buildDelegateURL(homeURL string, fs *ForeignInstanceSig) (string, error) {
	u, err := url.Parse(strings.TrimRight(homeURL, "/") + "/auth/delegate")
	if err != nil {
		return "", err
	}
	sigJSON, err := json.Marshal(fs)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("return_to", fs.ReturnTo)
	q.Set("nonce", fs.Nonce)
	q.Set("event_id", fs.EventID)
	q.Set("foreign_sig", string(sigJSON))
	u.RawQuery = q.Encode()
	return u.String(), nil
}
