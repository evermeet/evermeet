package identity

import "testing"

func TestNormalizeEvermeetDID(t *testing.T) {
	raw := "  DiD:Em:abcdefghijklmnopqrstuvwxyz1234  "
	got, err := NormalizeEvermeetDID(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := "did:em:abcdefghijklmnopqrstuvwxyz1234"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if _, err := NormalizeEvermeetDID("not-a-did"); err == nil {
		t.Fatal("expected error")
	}
}
