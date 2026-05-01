package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/evermeet/evermeet/internal/identity"
	"github.com/evermeet/evermeet/internal/store"
	"golang.org/x/crypto/sha3"
)

const siweStatement = "Sign in to Evermeet."

func (s *Server) handleSIWEStart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address string `json:"address"`
		ChainID string `json:"chain_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid request")
		return
	}
	address, err := normalizeEthereumAddress(req.Address)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid ethereum address")
		return
	}
	if _, err := strconv.ParseInt(req.ChainID, 10, 64); err != nil || req.ChainID == "" {
		jsonErr(w, http.StatusBadRequest, "invalid chain_id")
		return
	}

	nonce := randomHex(16)
	now := time.Now().UTC()
	expiresAt := now.Add(5 * time.Minute)
	if err := s.db.InsertSIWENonce(r.Context(), &store.SIWENonce{
		Nonce:     nonce,
		Address:   address,
		ChainID:   req.ChainID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}); err != nil {
		jsonErr(w, http.StatusInternalServerError, "nonce storage failed")
		return
	}

	message, err := s.buildSIWEMessage(address, req.ChainID, nonce, now, expiresAt)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "message creation failed")
		return
	}
	jsonOK(w, map[string]string{
		"message": message,
		"nonce":   nonce,
	})
}

func (s *Server) handleSIWEFinish(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Message   string `json:"message"`
		Signature string `json:"signature"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" || req.Signature == "" {
		jsonErr(w, http.StatusBadRequest, "message and signature required")
		return
	}

	msg, err := parseSIWEMessage(req.Message)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid siwe message")
		return
	}
	if err := s.verifySIWEMessage(msg); err != nil {
		jsonErr(w, http.StatusUnauthorized, err.Error())
		return
	}

	recovered, err := recoverEthereumAddress(req.Message, req.Signature)
	if err != nil {
		jsonErr(w, http.StatusUnauthorized, "invalid signature")
		return
	}
	if !strings.EqualFold(recovered, msg.Address) {
		jsonErr(w, http.StatusUnauthorized, "signature address mismatch")
		return
	}

	nonce, err := s.db.GetSIWENonce(r.Context(), msg.Nonce)
	if err != nil || nonce == nil || nonce.Used || time.Now().After(nonce.ExpiresAt) {
		jsonErr(w, http.StatusUnauthorized, "invalid or expired nonce")
		return
	}
	recoveredLower := recovered
	if nonce.Address != recoveredLower || nonce.ChainID != msg.ChainID {
		jsonErr(w, http.StatusUnauthorized, "nonce does not match message")
		return
	}
	if err := s.db.MarkSIWENonceUsed(r.Context(), msg.Nonce); err != nil {
		jsonErr(w, http.StatusInternalServerError, "nonce update failed")
		return
	}

	did, err := s.lookupOrCreateSIWEUser(r.Context(), msg.ChainID, recoveredLower, authDID(r))
	if err != nil {
		s.log.Printf("siwe identity: %v", err)
		jsonErr(w, http.StatusInternalServerError, "user storage failed")
		return
	}
	if pub := s.publisher(); pub != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := pub.PublishEthereum(ctx, msg.ChainID, recoveredLower); err != nil {
				s.log.Printf("dht publish siwe %s:%s: %v", msg.ChainID, recoveredLower, err)
			}
		}()
	}

	s.createSession(w, r.Context(), did)
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) lookupOrCreateSIWEUser(ctx context.Context, chainID, address, currentDID string) (string, error) {
	existing, err := s.db.GetUserPrivateBySIWE(ctx, chainID, address)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return existing.DID, nil
	}

	if currentDID != "" {
		if err := s.db.SetUserSIWE(ctx, currentDID, chainID, address); err != nil {
			return "", err
		}
		return currentDID, nil
	}

	kp, err := identity.Generate()
	if err != nil {
		return "", err
	}
	did := identity.DeriveDID(kp.SigningPub)
	password := identity.SessionPassword(s.serverSecret, did)
	encKey, err := identity.EncryptKeypair(kp.SigningPriv, password)
	if err != nil {
		return "", err
	}
	genesisOp, genesisHash, err := identity.BuildGenesisOp(kp.SigningPriv, kp.RotationPriv, s.homeHost())
	if err != nil {
		return "", err
	}
	now := time.Now()
	if err := s.db.AppendKELOp(ctx, &store.KELOp{
		Hash:      genesisHash,
		DID:       did,
		Type:      string(genesisOp.Type),
		Payload:   mustJSON(genesisOp),
		Seq:       0,
		CreatedAt: now,
	}); err != nil {
		return "", err
	}
	if err := s.db.UpsertUser(ctx, &store.User{
		DID:         did,
		DisplayName: shortAddress(address),
		CurrentPK:   hex.EncodeToString(kp.SigningPub),
		RotationPK:  hex.EncodeToString(kp.RotationPub),
		Endpoint:    s.baseURL,
		Sig:         genesisHash,
		UpdatedAt:   now,
		InstanceID:  s.homeHost(),
	}); err != nil {
		return "", err
	}
	if err := s.db.UpsertUserPrivate(ctx, &store.UserPrivate{
		DID:           did,
		SIWEChainID:   chainID,
		SIWEAddress:   address,
		EncSigningKey: encKey,
	}); err != nil {
		return "", err
	}
	return did, nil
}

type parsedSIWEMessage struct {
	Domain         string
	Address        string
	URI            string
	Version        string
	ChainID        string
	Nonce          string
	IssuedAt       string
	ExpirationTime string
}

func (s *Server) buildSIWEMessage(address, chainID, nonce string, issuedAt, expiresAt time.Time) (string, error) {
	u, err := url.Parse(s.baseURL)
	if err != nil {
		return "", err
	}
	domain := u.Host
	uri := strings.TrimRight(s.baseURL, "/")
	return fmt.Sprintf("%s wants you to sign in with your Ethereum account:\n%s\n\n%s\n\nURI: %s\nVersion: 1\nChain ID: %s\nNonce: %s\nIssued At: %s\nExpiration Time: %s",
		domain,
		address,
		siweStatement,
		uri,
		chainID,
		nonce,
		issuedAt.Format(time.RFC3339),
		expiresAt.Format(time.RFC3339),
	), nil
}

func parseSIWEMessage(raw string) (*parsedSIWEMessage, error) {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	if len(lines) < 10 || !strings.HasSuffix(lines[0], " wants you to sign in with your Ethereum account:") {
		return nil, fmt.Errorf("invalid siwe header")
	}
	address, err := normalizeEthereumAddress(lines[1])
	if err != nil {
		return nil, err
	}
	msg := &parsedSIWEMessage{
		Domain:  strings.TrimSuffix(lines[0], " wants you to sign in with your Ethereum account:"),
		Address: address,
	}
	for _, line := range lines[2:] {
		switch {
		case strings.HasPrefix(line, "URI: "):
			msg.URI = strings.TrimPrefix(line, "URI: ")
		case strings.HasPrefix(line, "Version: "):
			msg.Version = strings.TrimPrefix(line, "Version: ")
		case strings.HasPrefix(line, "Chain ID: "):
			msg.ChainID = strings.TrimPrefix(line, "Chain ID: ")
		case strings.HasPrefix(line, "Nonce: "):
			msg.Nonce = strings.TrimPrefix(line, "Nonce: ")
		case strings.HasPrefix(line, "Issued At: "):
			msg.IssuedAt = strings.TrimPrefix(line, "Issued At: ")
		case strings.HasPrefix(line, "Expiration Time: "):
			msg.ExpirationTime = strings.TrimPrefix(line, "Expiration Time: ")
		}
	}
	if msg.Domain == "" || msg.URI == "" || msg.Version != "1" || msg.ChainID == "" || msg.Nonce == "" {
		return nil, fmt.Errorf("missing siwe fields")
	}
	return msg, nil
}

func (s *Server) verifySIWEMessage(msg *parsedSIWEMessage) error {
	base, err := url.Parse(s.baseURL)
	if err != nil {
		return fmt.Errorf("invalid server URL")
	}
	if msg.Domain != base.Host {
		return fmt.Errorf("domain mismatch")
	}
	if strings.TrimRight(msg.URI, "/") != strings.TrimRight(s.baseURL, "/") {
		return fmt.Errorf("uri mismatch")
	}
	if _, err := strconv.ParseInt(msg.ChainID, 10, 64); err != nil {
		return fmt.Errorf("invalid chain id")
	}
	if msg.ExpirationTime != "" {
		exp, err := time.Parse(time.RFC3339, msg.ExpirationTime)
		if err != nil || time.Now().After(exp) {
			return fmt.Errorf("message expired")
		}
	}
	if msg.IssuedAt != "" {
		iat, err := time.Parse(time.RFC3339, msg.IssuedAt)
		if err != nil || iat.After(time.Now().Add(time.Minute)) {
			return fmt.Errorf("invalid issued-at")
		}
	}
	return nil
}

func recoverEthereumAddress(message, signature string) (string, error) {
	sig, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		return "", err
	}
	if len(sig) != 65 {
		return "", fmt.Errorf("invalid signature length")
	}
	recoveryID := sig[64]
	if recoveryID >= 27 {
		recoveryID -= 27
	}
	if recoveryID > 1 {
		return "", fmt.Errorf("invalid recovery id")
	}

	compact := make([]byte, 65)
	compact[0] = 27 + recoveryID
	copy(compact[1:33], sig[:32])
	copy(compact[33:], sig[32:64])

	pub, _, err := secp256k1ecdsa.RecoverCompact(compact, ethereumMessageHash(message))
	if err != nil {
		return "", err
	}
	serialized := pub.SerializeUncompressed()
	h := sha3.NewLegacyKeccak256()
	h.Write(serialized[1:])
	sum := h.Sum(nil)
	return "0x" + hex.EncodeToString(sum[len(sum)-20:]), nil
}

func shortAddress(address string) string {
	if len(address) <= 12 {
		return address
	}
	return address[:6] + "..." + address[len(address)-4:]
}

func normalizeEthereumAddress(address string) (string, error) {
	address = strings.TrimSpace(address)
	raw := strings.TrimPrefix(address, "0x")
	raw = strings.TrimPrefix(raw, "0X")
	if len(raw) != 40 {
		return "", fmt.Errorf("invalid ethereum address length")
	}
	if _, err := hex.DecodeString(raw); err != nil {
		return "", err
	}
	return "0x" + strings.ToLower(raw), nil
}

func ethereumMessageHash(message string) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(prefix))
	h.Write([]byte(message))
	return h.Sum(nil)
}
