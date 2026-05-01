package api

import (
	"encoding/hex"
	"testing"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/sha3"
)

func TestRecoverEthereumAddress(t *testing.T) {
	priv, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	message := "example.org wants you to sign in with your Ethereum account:\n0x0000000000000000000000000000000000000000"
	compact := secp256k1ecdsa.SignCompact(priv, ethereumMessageHash(message), false)
	ethSig := make([]byte, 65)
	copy(ethSig[:64], compact[1:])
	ethSig[64] = compact[0]

	recovered, err := recoverEthereumAddress(message, "0x"+hex.EncodeToString(ethSig))
	if err != nil {
		t.Fatalf("recover address: %v", err)
	}

	pub := priv.PubKey().SerializeUncompressed()
	h := sha3.NewLegacyKeccak256()
	h.Write(pub[1:])
	sum := h.Sum(nil)
	want := "0x" + hex.EncodeToString(sum[len(sum)-20:])
	if recovered != want {
		t.Fatalf("recovered address %s, want %s", recovered, want)
	}
}

func TestNormalizeEthereumAddress(t *testing.T) {
	got, err := normalizeEthereumAddress("0xABcDEFabcdefABCDefAbCdefAbcdefABcdefABCD")
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	want := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"
	if got != want {
		t.Fatalf("normalized address %s, want %s", got, want)
	}
}
