package api

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/evermeet/evermeet/internal/events"
	"github.com/evermeet/evermeet/internal/identity"
	"github.com/evermeet/evermeet/internal/store"
)

type RSVPReceiptToken struct {
	EventInstanceURL string `json:"event_instance_url"`
	EventID          string `json:"event_id"`
	DID              string `json:"did"`
	Status           string `json:"status"`
	GuestVisible     bool   `json:"guest_visible"`
	EventURL         string `json:"event_url,omitempty"`
	EventTitle       string `json:"event_title,omitempty"`
	EventStartsAt    string `json:"event_starts_at,omitempty"`
	IssuedAt         int64  `json:"issued_at"`
	UpdatedAt        int64  `json:"updated_at"`
}

type SignedRSVPReceipt struct {
	Receipt RSVPReceiptToken `json:"receipt"`
	Sig     string           `json:"sig"`
}

func (s *Server) handleReceiveRSVPReceipt(w http.ResponseWriter, r *http.Request) {
	var req SignedRSVPReceipt
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := verifySignedRSVPReceipt(r.Context(), &req); err != nil {
		s.log.Printf("rsvp receipt: invalid receipt: %v", err)
		jsonErr(w, http.StatusBadRequest, "invalid receipt")
		return
	}

	receipt := req.Receipt
	user, err := s.db.GetUser(r.Context(), receipt.DID)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "check user failed")
		return
	}
	if user == nil {
		jsonErr(w, http.StatusNotFound, "user not found")
		return
	}
	userHomeURL := strings.TrimRight(user.Endpoint, "/")
	if userHomeURL != "" && userHomeURL != strings.TrimRight(s.baseURL, "/") {
		jsonErr(w, http.StatusForbidden, "user is not local to this instance")
		return
	}

	now := time.Now()
	issuedAt := time.Unix(receipt.IssuedAt, 0).UTC()
	updatedAt := time.Unix(receipt.UpdatedAt, 0).UTC()
	if updatedAt.IsZero() {
		updatedAt = now.UTC()
	}

	if err := s.db.UpsertRSVPReceipt(r.Context(), &store.RSVPReceipt{
		EventInstanceURL: strings.TrimRight(receipt.EventInstanceURL, "/"),
		EventID:          receipt.EventID,
		DID:              receipt.DID,
		Status:           receipt.Status,
		GuestVisible:     receipt.GuestVisible,
		EventURL:         receipt.EventURL,
		EventTitle:       receipt.EventTitle,
		EventStartsAt:    receipt.EventStartsAt,
		IssuedAt:         issuedAt,
		UpdatedAt:        updatedAt,
		Sig:              req.Sig,
	}); err != nil {
		jsonErr(w, http.StatusInternalServerError, "store receipt failed")
		return
	}

	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleListRSVPReceipts(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)
	receipts, err := s.db.ListRSVPReceiptsForDID(r.Context(), did)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "list receipts failed")
		return
	}

	out := make([]map[string]any, 0, len(receipts))
	for _, receipt := range receipts {
		out = append(out, map[string]any{
			"event_instance_url": receipt.EventInstanceURL,
			"event_id":           receipt.EventID,
			"did":                receipt.DID,
			"status":             receipt.Status,
			"guest_visible":      receipt.GuestVisible,
			"event_url":          receipt.EventURL,
			"event_title":        receipt.EventTitle,
			"event_starts_at":    receipt.EventStartsAt,
			"issued_at":          receipt.IssuedAt.UTC().Format(time.RFC3339),
			"updated_at":         receipt.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	jsonOK(w, map[string]any{"receipts": out})
}

func verifySignedRSVPReceipt(ctx context.Context, req *SignedRSVPReceipt) error {
	receipt := req.Receipt
	if strings.TrimSpace(receipt.EventInstanceURL) == "" || strings.TrimSpace(receipt.EventID) == "" || strings.TrimSpace(receipt.DID) == "" {
		return fmt.Errorf("receipt missing required fields")
	}
	if !validRSVPReceiptStatus(receipt.Status) {
		return fmt.Errorf("invalid status")
	}
	if req.Sig == "" {
		return fmt.Errorf("missing signature")
	}
	issuedAt := time.Unix(receipt.IssuedAt, 0)
	if time.Since(issuedAt) > 24*time.Hour || issuedAt.After(time.Now().Add(5*time.Minute)) {
		return fmt.Errorf("receipt signature expired")
	}
	if receipt.UpdatedAt <= 0 {
		return fmt.Errorf("missing updated_at")
	}

	pub, err := fetchInstancePubKey(ctx, receipt.EventInstanceURL)
	if err != nil {
		return fmt.Errorf("fetch event instance key: %w", err)
	}
	ok, err := identity.Verify(pub, receipt, req.Sig)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("signature invalid")
	}
	return nil
}

func validRSVPReceiptStatus(status string) bool {
	switch status {
	case "pending", "confirmed", "rejected", "waitlisted", "cancelled":
		return true
	default:
		return false
	}
}

func (s *Server) signRSVPReceipt(receipt RSVPReceiptToken) (*SignedRSVPReceipt, error) {
	if len(s.instancePriv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid instance private key")
	}
	sig, err := identity.Sign(s.instancePriv, receipt)
	if err != nil {
		return nil, err
	}
	return &SignedRSVPReceipt{Receipt: receipt, Sig: sig}, nil
}

func (s *Server) emitRSVPReceiptAsync(did string, ms events.MutableState, status string, guestVisible bool) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.emitRSVPReceipt(ctx, did, ms, status, guestVisible); err != nil {
			s.log.Printf("rsvp receipt: emit failed for %s/%s: %v", ms.ID, did, err)
		}
	}()
}

func (s *Server) emitRSVPReceipt(ctx context.Context, did string, ms events.MutableState, status string, guestVisible bool) error {
	user, err := s.db.GetUser(ctx, did)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}
	homeURL := strings.TrimRight(user.Endpoint, "/")
	if homeURL == "" || homeURL == strings.TrimRight(s.baseURL, "/") {
		return nil
	}

	now := time.Now().UTC()
	receipt := RSVPReceiptToken{
		EventInstanceURL: strings.TrimRight(s.baseURL, "/"),
		EventID:          ms.ID,
		DID:              did,
		Status:           status,
		GuestVisible:     guestVisible,
		EventURL:         eventURL(s.baseURL, ms.ID),
		EventTitle:       ms.Title,
		EventStartsAt:    ms.StartsAt,
		IssuedAt:         now.Unix(),
		UpdatedAt:        now.Unix(),
	}
	signed, err := s.signRSVPReceipt(receipt)
	if err != nil {
		return err
	}

	body, err := json.Marshal(signed)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, homeURL+"/api/federation/rsvp-receipts", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("home instance returned %s", resp.Status)
	}
	return nil
}

func eventURL(baseURL, eventID string) string {
	return strings.TrimRight(baseURL, "/") + "/events/" + url.PathEscape(eventID)
}
