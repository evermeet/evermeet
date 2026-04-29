package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func generateKeypair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return pub, priv
}

func TestKEL_GenesisReplay(t *testing.T) {
	_, sigPriv := generateKeypair(t)
	_, rotPriv := generateKeypair(t)

	op, hash, err := BuildGenesisOp(sigPriv, rotPriv, "https://example.com")
	if err != nil {
		t.Fatalf("BuildGenesisOp: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	state, err := ReplayKEL([]*KELOp{op})
	if err != nil {
		t.Fatalf("ReplayKEL: %v", err)
	}

	expectedDID := DeriveDID(sigPriv.Public().(ed25519.PublicKey))
	if state.DID != expectedDID {
		t.Errorf("DID mismatch: got %s, want %s", state.DID, expectedDID)
	}
	if state.Endpoint != "https://example.com" {
		t.Errorf("endpoint mismatch: %s", state.Endpoint)
	}
	if state.Seq != 0 {
		t.Errorf("seq should be 0, got %d", state.Seq)
	}
	if state.HeadHash != hash {
		t.Errorf("head hash mismatch")
	}
}

func TestKEL_Migrate(t *testing.T) {
	_, sigPriv := generateKeypair(t)
	_, rotPriv := generateKeypair(t)

	genesis, genesisHash, err := BuildGenesisOp(sigPriv, rotPriv, "https://instance-a.com")
	if err != nil {
		t.Fatal(err)
	}

	state, _ := ReplayKEL([]*KELOp{genesis})

	migrateOp, migrateHash, err := BuildMigrateOp(state, sigPriv, "https://instance-b.com", genesisHash)
	if err != nil {
		t.Fatalf("BuildMigrateOp: %v", err)
	}

	finalState, err := ReplayKEL([]*KELOp{genesis, migrateOp})
	if err != nil {
		t.Fatalf("ReplayKEL with migrate: %v", err)
	}

	if finalState.Endpoint != "https://instance-b.com" {
		t.Errorf("endpoint not updated: %s", finalState.Endpoint)
	}
	if finalState.Seq != 1 {
		t.Errorf("seq should be 1, got %d", finalState.Seq)
	}
	if finalState.HeadHash != migrateHash {
		t.Errorf("head hash mismatch after migrate")
	}
}

func TestKEL_TamperedGenesis(t *testing.T) {
	_, sigPriv := generateKeypair(t)
	_, rotPriv := generateKeypair(t)

	op, _, err := BuildGenesisOp(sigPriv, rotPriv, "https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	// Tamper: change the endpoint after signing.
	op.Endpoint = "https://evil.com"

	_, err = ReplayKEL([]*KELOp{op})
	if err == nil {
		t.Fatal("expected error for tampered genesis, got nil")
	}
}

func TestKEL_MigrateWrongPrev(t *testing.T) {
	_, sigPriv := generateKeypair(t)
	_, rotPriv := generateKeypair(t)

	genesis, _, err := BuildGenesisOp(sigPriv, rotPriv, "https://instance-a.com")
	if err != nil {
		t.Fatal(err)
	}
	state, _ := ReplayKEL([]*KELOp{genesis})

	// Use a wrong prev hash.
	migrateOp, _, err := BuildMigrateOp(state, sigPriv, "https://instance-b.com", "wronghash")
	if err != nil {
		t.Fatal(err)
	}

	_, err = ReplayKEL([]*KELOp{genesis, migrateOp})
	if err == nil {
		t.Fatal("expected error for wrong prev hash, got nil")
	}
}

func TestKEL_ValidateIncomingOp(t *testing.T) {
	_, sigPriv := generateKeypair(t)
	_, rotPriv := generateKeypair(t)

	genesis, genesisHash, err := BuildGenesisOp(sigPriv, rotPriv, "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	state, _ := ReplayKEL([]*KELOp{genesis})
	_ = genesisHash

	migrateOp, _, err := BuildMigrateOp(state, sigPriv, "https://new.com", state.HeadHash)
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateKELOp(migrateOp, state); err != nil {
		t.Errorf("ValidateKELOp: %v", err)
	}
}
