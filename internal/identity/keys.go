package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/multiformats/go-multibase"
	"lukechampine.com/blake3"
)

// Keypair holds a signing keypair and a separate rotation keypair.
// The rotation key is used only to authorize signing key rotations
// and should be kept offline.
type Keypair struct {
	SigningPriv  ed25519.PrivateKey
	SigningPub   ed25519.PublicKey
	RotationPriv ed25519.PrivateKey
	RotationPub  ed25519.PublicKey
}

// DeriveDID computes did:em:{base32(blake3(pubkey))} from an Ed25519 public key.
// This is the stable, permanent identifier for a user — it never changes even
// if the signing key is rotated.
func DeriveDID(pubkey ed25519.PublicKey) string {
	h := blake3.Sum256(pubkey)
	// Truncate to 15 bytes (120 bits) for a 24-character Base32 ID
	encoded, _ := multibase.Encode(multibase.Base32, h[:15])
	// multibase prefixes with 'b' for base32; strip it and use our own prefix
	return "did:em:" + encoded[1:]
}

// Generate creates a new Ed25519 signing keypair and a separate rotation keypair.
func Generate() (*Keypair, error) {
	sigPub, sigPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate signing key: %w", err)
	}
	rotPub, rotPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate rotation key: %w", err)
	}
	return &Keypair{
		SigningPriv:  sigPriv,
		SigningPub:   sigPub,
		RotationPriv: rotPriv,
		RotationPub:  rotPub,
	}, nil
}

// LoadOrGenerate loads keypair files from dir, or generates and saves new ones if absent.
func LoadOrGenerate(dir string) (*Keypair, error) {
	sigPath := filepath.Join(dir, "signing.key")
	rotPath := filepath.Join(dir, "rotation.key")

	if _, err := os.Stat(sigPath); os.IsNotExist(err) {
		kp, err := Generate()
		if err != nil {
			return nil, err
		}
		if err := saveKey(sigPath, kp.SigningPriv); err != nil {
			return nil, fmt.Errorf("save signing key: %w", err)
		}
		if err := saveKey(rotPath, kp.RotationPriv); err != nil {
			return nil, fmt.Errorf("save rotation key: %w", err)
		}
		return kp, nil
	}

	sigPriv, err := loadKey(sigPath)
	if err != nil {
		return nil, fmt.Errorf("load signing key: %w", err)
	}
	rotPriv, err := loadKey(rotPath)
	if err != nil {
		return nil, fmt.Errorf("load rotation key: %w", err)
	}
	return &Keypair{
		SigningPriv:  sigPriv,
		SigningPub:   sigPriv.Public().(ed25519.PublicKey),
		RotationPriv: rotPriv,
		RotationPub:  rotPriv.Public().(ed25519.PublicKey),
	}, nil
}

func saveKey(path string, priv ed25519.PrivateKey) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	block := &pem.Block{Type: "ED25519 PRIVATE KEY", Bytes: priv.Seed()}
	return os.WriteFile(path, pem.EncodeToMemory(block), 0600)
}

func loadKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM in %s", path)
	}
	if len(block.Bytes) != ed25519.SeedSize {
		return nil, fmt.Errorf("unexpected key size in %s", path)
	}
	return ed25519.NewKeyFromSeed(block.Bytes), nil
}
