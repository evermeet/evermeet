package api

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evermeet/evermeet/internal/identity"
)

func TestVerifySignedRSVPReceipt(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	instance := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/evermeet-node-key" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"instance_id": "test",
			"public_key":  []byte(pub),
		})
	}))
	defer instance.Close()

	now := time.Now().UTC()
	receipt := RSVPReceiptToken{
		EventInstanceURL: instance.URL,
		EventID:          "event1",
		DID:              "did:em:abc",
		Status:           "confirmed",
		GuestVisible:     true,
		EventURL:         instance.URL + "/events/event1",
		EventTitle:       "Event",
		IssuedAt:         now.Unix(),
		UpdatedAt:        now.Unix(),
	}
	sig, err := identity.Sign(priv, receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifySignedRSVPReceipt(context.Background(), &SignedRSVPReceipt{Receipt: receipt, Sig: sig}); err != nil {
		t.Fatalf("verifySignedRSVPReceipt: %v", err)
	}

	receipt.Status = "cancelled"
	if err := verifySignedRSVPReceipt(context.Background(), &SignedRSVPReceipt{Receipt: receipt, Sig: sig}); err == nil {
		t.Fatal("expected modified receipt to fail verification")
	}
}
