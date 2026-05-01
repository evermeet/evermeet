package routing

import (
	"bytes"
	"testing"
)

func TestEmailHashNormalizesEmail(t *testing.T) {
	a := EmailHash("Alice@Example.COM")
	b := EmailHash("  alice@example.com  ")
	if !bytes.Equal(a, b) {
		t.Fatalf("expected normalized emails to produce the same routing key")
	}
}

func TestEmailHashDomainSeparated(t *testing.T) {
	a := EmailHash("alice@example.com")
	b := EmailHash("bob@example.com")
	if bytes.Equal(a, b) {
		t.Fatalf("expected different emails to produce different routing keys")
	}
	if len(a) != 64 {
		t.Fatalf("expected 32-byte key hex encoded to 64 bytes, got %d", len(a))
	}
}

func TestEthereumHashNormalizesIdentity(t *testing.T) {
	a := EthereumHash("1", "0xABCDEFabcdefABCDefAbCdefAbcdefABcdefABCD")
	b := EthereumHash(" 1 ", " 0xabcdefabcdefabcdefabcdefabcdefabcdefabcd ")
	if !bytes.Equal(a, b) {
		t.Fatalf("expected normalized ethereum identities to produce the same routing key")
	}
}

func TestDIDHashNormalizesCase(t *testing.T) {
	a := DIDHash("did:em:ABCDEFabcdef123456789012")
	b := DIDHash("  did:em:abcdefabcdef123456789012  ")
	if string(a) != string(b) {
		t.Errorf("DIDHash not case-stable: %s vs %s", a, b)
	}
}

func TestEthereumHashDomainSeparatedFromEmail(t *testing.T) {
	email := EmailHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	wallet := EthereumHash("1", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	if bytes.Equal(email, wallet) {
		t.Fatalf("expected ethereum and email routing keys to be domain-separated")
	}
	if len(wallet) != 64 {
		t.Fatalf("expected 32-byte key hex encoded to 64 bytes, got %d", len(wallet))
	}
}
