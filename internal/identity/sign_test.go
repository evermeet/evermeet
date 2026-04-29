package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

// RFC 8785 JCS test vectors — canonical form must match exactly.
func TestCanonicalJSON_RFC8785Vectors(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "null",
			input: nil,
			want:  "null",
		},
		{
			name:  "true",
			input: true,
			want:  "true",
		},
		{
			name:  "false",
			input: false,
			want:  "false",
		},
		{
			name:  "integer",
			input: map[string]any{"n": float64(42)},
			want:  `{"n":42}`,
		},
		{
			name:  "negative integer",
			input: map[string]any{"n": float64(-1)},
			want:  `{"n":-1}`,
		},
		{
			name:  "float",
			input: map[string]any{"n": 3.14},
			want:  `{"n":3.14}`,
		},
		{
			name:  "empty object",
			input: map[string]any{},
			want:  `{}`,
		},
		{
			name:  "keys sorted",
			input: map[string]any{"b": 2.0, "a": 1.0},
			want:  `{"a":1,"b":2}`,
		},
		{
			name:  "nested object keys sorted",
			input: map[string]any{"z": map[string]any{"b": 2.0, "a": 1.0}, "a": "hello"},
			want:  `{"a":"hello","z":{"a":1,"b":2}}`,
		},
		{
			name:  "array preserved order",
			input: []any{3.0, 1.0, 2.0},
			want:  `[3,1,2]`,
		},
		{
			name:  "string escaping",
			input: "hello world",
			want:  `"hello world"`,
		},
		{
			name:  "large integer",
			input: map[string]any{"n": float64(9007199254740992)},
			want:  `{"n":9007199254740992}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CanonicalJSON(tc.input)
			if err != nil {
				t.Fatalf("CanonicalJSON error: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestSignAndVerify(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	payload := map[string]any{
		"type":    "event",
		"title":   "Test Event",
		"version": 1.0,
	}

	sig, err := Sign(priv, payload)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	ok, err := Verify(pub, payload, sig)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Fatal("signature verification failed")
	}
}

func TestVerify_TamperedPayload(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	payload := map[string]any{"title": "Original"}
	sig, err := Sign(priv, payload)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	tampered := map[string]any{"title": "Modified"}
	ok, err := Verify(pub, tampered, sig)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if ok {
		t.Fatal("tampered payload should not verify")
	}
}

func TestContentHash_Deterministic(t *testing.T) {
	v := map[string]any{"z": "last", "a": "first"}
	h1, err := ContentHash(v)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := ContentHash(v)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Errorf("content hash not deterministic: %s != %s", h1, h2)
	}
}

func TestContentHash_KeyOrderIndependent(t *testing.T) {
	// Two maps with same keys/values in different insertion order must hash identically.
	a := map[string]any{"b": 2.0, "a": 1.0}
	b := map[string]any{"a": 1.0, "b": 2.0}

	ha, err := ContentHash(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := ContentHash(b)
	if err != nil {
		t.Fatal(err)
	}
	if ha != hb {
		t.Errorf("key-order dependency: %s != %s", ha, hb)
	}
}

func TestDeriveDID(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	did := DeriveDID(pub)
	if len(did) < 8 || did[:7] != "did:em:" {
		t.Errorf("unexpected DID format: %s", did)
	}
	// Must be deterministic.
	did2 := DeriveDID(pub)
	if did != did2 {
		t.Errorf("DID not deterministic: %s != %s", did, did2)
	}
}
