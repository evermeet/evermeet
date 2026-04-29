package events

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/evermeet/evermeet/internal/identity"
)

func newKeypair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return pub, priv
}

func TestEvent_NewAndVerify(t *testing.T) {
	pub, priv := newKeypair(t)
	did := identity.DeriveDID(pub)

	f := Fields{
		Title:    "Test Event",
		StartsAt: time.Now().Add(24 * time.Hour),
	}

	founding, eventID, state, stateHash, err := New(did, priv, f)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if founding.Type != "event" {
		t.Errorf("founding type: %s", founding.Type)
	}
	if eventID == "" {
		t.Error("empty event ID")
	}
	if stateHash == "" {
		t.Error("empty state hash")
	}
	if state.Title != "Test Event" {
		t.Errorf("title: %s", state.Title)
	}
	if state.ID != eventID {
		t.Errorf("state.ID mismatch: %s != %s", state.ID, eventID)
	}
	if state.Prev != "" {
		t.Errorf("initial state should have no prev, got %s", state.Prev)
	}

	// Verify signature.
	err = Verify(state, func(signerDID string) (ed25519.PublicKey, error) {
		if signerDID == did {
			return pub, nil
		}
		return nil, nil
	})
	if err != nil {
		t.Errorf("Verify: %v", err)
	}
}

func TestEvent_Update(t *testing.T) {
	pub, priv := newKeypair(t)
	did := identity.DeriveDID(pub)

	founding, eventID, state, stateHash, err := New(did, priv, Fields{
		Title:    "Original Title",
		StartsAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = founding
	_ = eventID

	newState, newHash, err := Update(state, stateHash, did, priv, Fields{
		Title:    "Updated Title",
		StartsAt: time.Now().Add(48 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if newState.Title != "Updated Title" {
		t.Errorf("title not updated: %s", newState.Title)
	}
	if newState.Prev != stateHash {
		t.Errorf("prev not set correctly: %s", newState.Prev)
	}
	if newState.ID != state.ID {
		t.Errorf("event ID changed: %s -> %s", state.ID, newState.ID)
	}
	if newHash == stateHash {
		t.Error("state hash did not change after update")
	}
}

func TestEvent_UnauthorizedUpdate(t *testing.T) {
	pub, priv := newKeypair(t)
	did := identity.DeriveDID(pub)

	_, _, state, stateHash, err := New(did, priv, Fields{
		Title:    "My Event",
		StartsAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Attacker keypair.
	attackerPub, attackerPriv := newKeypair(t)
	attackerDID := identity.DeriveDID(attackerPub)

	_, _, err = Update(state, stateHash, attackerDID, attackerPriv, Fields{
		Title:    "Hacked",
		StartsAt: time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("expected error for unauthorized update, got nil")
	}
}

func TestEvent_IDStability(t *testing.T) {
	pub, priv := newKeypair(t)
	did := identity.DeriveDID(pub)

	_, eventID, state, stateHash, _ := New(did, priv, Fields{
		Title:    "Stable",
		StartsAt: time.Now().Add(24 * time.Hour),
	})

	updatedState, _, _ := Update(state, stateHash, did, priv, Fields{
		Title:    "Changed",
		StartsAt: time.Now().Add(48 * time.Hour),
	})

	if updatedState.ID != eventID {
		t.Errorf("event ID changed after update: %s -> %s", eventID, updatedState.ID)
	}
}
