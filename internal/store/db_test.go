package store

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestDB_OpenAndMigrate(t *testing.T) {
	f, err := os.CreateTemp("", "evermeet-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	db, err := Open(f.Name())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// Migrations should be idempotent — open again.
	db2, err := Open(f.Name())
	if err != nil {
		t.Fatalf("Open second time: %v", err)
	}
	db2.Close()
}

func TestDB_KELRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	op := &KELOp{
		Hash:      "testhash",
		DID:       "did:em:abc",
		Type:      "genesis",
		Payload:   `{"type":"genesis"}`,
		Prev:      "",
		Seq:       0,
		CreatedAt: time.Now(),
	}

	if err := db.AppendKELOp(ctx, op); err != nil {
		t.Fatalf("AppendKELOp: %v", err)
	}

	ops, err := db.GetKELForDID(ctx, "did:em:abc")
	if err != nil {
		t.Fatalf("GetKELForDID: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 op, got %d", len(ops))
	}
	if ops[0].Hash != "testhash" {
		t.Errorf("hash mismatch: %s", ops[0].Hash)
	}
}

func TestDB_EventStateChain(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	if err := db.InsertEventFounding(ctx, &EventFounding{
		ID:      "event1",
		Payload: `{"type":"event"}`,
	}); err != nil {
		t.Fatalf("InsertEventFounding: %v", err)
	}

	s1 := &EventState{
		Hash:      "hash1",
		ID:        "event1",
		Prev:      "",
		Payload:   `{"title":"v1"}`,
		CreatedAt: time.Now(),
	}
	if err := db.AppendEventState(ctx, s1); err != nil {
		t.Fatalf("AppendEventState v1: %v", err)
	}

	current, err := db.GetCurrentEventState(ctx, "event1")
	if err != nil || current == nil {
		t.Fatalf("GetCurrentEventState: %v %v", current, err)
	}
	if current.Hash != "hash1" {
		t.Errorf("expected hash1, got %s", current.Hash)
	}

	// Append a second state — should demote the first.
	s2 := &EventState{
		Hash:      "hash2",
		ID:        "event1",
		Prev:      "hash1",
		Payload:   `{"title":"v2"}`,
		CreatedAt: time.Now(),
	}
	if err := db.AppendEventState(ctx, s2); err != nil {
		t.Fatalf("AppendEventState v2: %v", err)
	}

	current, err = db.GetCurrentEventState(ctx, "event1")
	if err != nil || current == nil {
		t.Fatal("GetCurrentEventState after update")
	}
	if current.Hash != "hash2" {
		t.Errorf("expected hash2 as current, got %s", current.Hash)
	}
}

func TestDB_SessionRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	s := &Session{
		Token:     "tok123",
		DID:       "did:em:abc",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := db.InsertSession(ctx, s); err != nil {
		t.Fatalf("InsertSession: %v", err)
	}

	got, err := db.GetSession(ctx, "tok123")
	if err != nil || got == nil {
		t.Fatalf("GetSession: %v %v", got, err)
	}
	if got.DID != "did:em:abc" {
		t.Errorf("DID mismatch: %s", got.DID)
	}

	if err := db.DeleteSession(ctx, "tok123"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	got, _ = db.GetSession(ctx, "tok123")
	if got != nil {
		t.Error("session should be deleted")
	}
}

func openTestDB(t *testing.T) *DB {
	t.Helper()
	f, err := os.CreateTemp("", "evermeet-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	db, err := Open(f.Name())
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
