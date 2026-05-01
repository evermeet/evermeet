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

func TestDB_RSVPReceiptUpsert(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	receipt := &RSVPReceipt{
		EventInstanceURL: "http://event.example",
		EventID:          "event1",
		DID:              "did:em:abc",
		Status:           "confirmed",
		GuestVisible:     true,
		EventURL:         "http://event.example/events/event1",
		EventTitle:       "First title",
		EventStartsAt:    "2026-05-01T12:00:00Z",
		IssuedAt:         time.Now(),
		UpdatedAt:        time.Now(),
		Sig:              "sig1",
	}
	if err := db.UpsertRSVPReceipt(ctx, receipt); err != nil {
		t.Fatalf("UpsertRSVPReceipt: %v", err)
	}

	receipt.Status = "cancelled"
	receipt.GuestVisible = false
	receipt.EventTitle = "Updated title"
	receipt.Sig = "sig2"
	receipt.UpdatedAt = receipt.UpdatedAt.Add(time.Minute)
	if err := db.UpsertRSVPReceipt(ctx, receipt); err != nil {
		t.Fatalf("UpsertRSVPReceipt update: %v", err)
	}
	older := *receipt
	older.Status = "pending"
	older.UpdatedAt = receipt.UpdatedAt.Add(-2 * time.Minute)
	if err := db.UpsertRSVPReceipt(ctx, &older); err != nil {
		t.Fatalf("UpsertRSVPReceipt older update: %v", err)
	}

	got, err := db.ListRSVPReceiptsForDID(ctx, "did:em:abc")
	if err != nil {
		t.Fatalf("ListRSVPReceiptsForDID: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 receipt, got %d", len(got))
	}
	if got[0].Status != "cancelled" || got[0].GuestVisible || got[0].EventTitle != "Updated title" || got[0].Sig != "sig2" {
		t.Fatalf("receipt was not updated: %+v", got[0])
	}
}

func TestDB_BlobSourcesRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	src := &BlobSource{
		Hash:        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		InstanceURL: "http://localhost:8080",
		CreatedAt:   time.Now(),
	}
	if err := db.InsertBlobSource(ctx, src); err != nil {
		t.Fatalf("InsertBlobSource: %v", err)
	}
	if err := db.InsertBlobSource(ctx, src); err != nil {
		t.Fatalf("InsertBlobSource duplicate: %v", err)
	}

	got, err := db.ListBlobSources(ctx, src.Hash)
	if err != nil {
		t.Fatalf("ListBlobSources: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 source, got %d", len(got))
	}
	if got[0].InstanceURL != src.InstanceURL {
		t.Fatalf("source URL mismatch: %s", got[0].InstanceURL)
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
