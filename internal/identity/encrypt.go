package identity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"

	"filippo.io/edwards25519"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

// EncryptForRecipient encrypts a message using a shared secret derived via ECDH
// between an ephemeral keypair and the recipient's public key.
// The recipient public key must be an Ed25519 key, which we convert to X25519.
// Returns base64url-encoded: ephemeral_pubkey(32) || nonce(12) || ciphertext.
func EncryptForRecipient(recipientPub ed25519.PublicKey, message []byte) (string, error) {
	// 1. Convert Ed25519 public key to X25519
	recipX, err := ed25519ToCurve25519Pub(recipientPub)
	if err != nil {
		return "", fmt.Errorf("convert recipient key: %w", err)
	}

	// 2. Generate ephemeral X25519 keypair
	var ephPriv, ephPub [32]byte
	if _, err := io.ReadFull(rand.Reader, ephPriv[:]); err != nil {
		return "", err
	}
	curve25519.ScalarBaseMult(&ephPub, &ephPriv)

	// 3. Derive shared secret via ECDH
	shared, err := curve25519.X25519(ephPriv[:], recipX[:])
	if err != nil {
		return "", err
	}

	// 4. Derive AES-256-GCM key from shared secret + ephemeral pubkey using HKDF
	kdf := hkdf.New(sha256.New, shared, ephPub[:], []byte("evermeet-rsvp-v1"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(kdf, key); err != nil {
		return "", err
	}

	// 5. Encrypt message with AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, message, nil)

	// Result: ephPub(32) + nonce(12) + ciphertext
	res := append(ephPub[:], append(nonce, ciphertext...)...)
	return base64.RawURLEncoding.EncodeToString(res), nil
}

// DecryptForRecipient decrypts a message using the recipient's Ed25519 private key.
// The private key is converted to X25519 for ECDH.
func DecryptForRecipient(recipientPriv ed25519.PrivateKey, encrypted string) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if len(data) < 32+12+16 {
		return nil, fmt.Errorf("data too short")
	}

	ephPub := data[:32]
	nonce := data[32 : 32+12]
	ciphertext := data[32+12:]

	// 1. Convert Ed25519 private key to X25519
	privX := ed25519ToCurve25519Priv(recipientPriv)

	// 2. Derive shared secret via ECDH
	shared, err := curve25519.X25519(privX[:], ephPub)
	if err != nil {
		return nil, err
	}

	// 3. Derive AES-256-GCM key from shared secret + ephemeral pubkey using HKDF
	kdf := hkdf.New(sha256.New, shared, ephPub, []byte("evermeet-rsvp-v1"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(kdf, key); err != nil {
		return nil, err
	}

	// 4. Decrypt message with AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}

func ed25519ToCurve25519Pub(pk ed25519.PublicKey) ([32]byte, error) {
	p, err := new(edwards25519.Point).SetBytes(pk)
	if err != nil {
		return [32]byte{}, err
	}
	return [32]byte(p.BytesMontgomery()), nil
}

func ed25519ToCurve25519Priv(sk ed25519.PrivateKey) [32]byte {
	h := sha512.Sum512(sk.Seed())
	var priv [32]byte
	copy(priv[:], h[:32])
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64
	return priv
}
