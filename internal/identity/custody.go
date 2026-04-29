package identity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// argon2 params — tuned for interactive login (fast enough, hard enough).
const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
)

// EncryptKeypair encrypts an Ed25519 private key seed using AES-256-GCM with
// an Argon2id-derived key. The output is hex-encoded: salt || nonce || ciphertext.
func EncryptKeypair(priv ed25519.PrivateKey, password []byte) (string, error) {
	seed := priv.Seed()

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	key := argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, argonKeyLen)

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

	ct := gcm.Seal(nil, nonce, seed, nil)

	payload := append(salt, append(nonce, ct...)...)
	return hex.EncodeToString(payload), nil
}

// DecryptKeypair decrypts an Ed25519 private key previously encrypted with EncryptKeypair.
func DecryptKeypair(encrypted string, password []byte) (ed25519.PrivateKey, error) {
	payload, err := hex.DecodeString(encrypted)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	if len(payload) < 16 {
		return nil, fmt.Errorf("payload too short")
	}
	salt := payload[:16]
	rest := payload[16:]

	key := argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ns := gcm.NonceSize()
	if len(rest) < ns {
		return nil, fmt.Errorf("payload too short for nonce")
	}
	nonce, ct := rest[:ns], rest[ns:]

	seed, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("unexpected seed size %d", len(seed))
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

// SessionPassword derives a deterministic encryption password from a server secret + DID.
// Used for custodial users where the "password" is server-controlled, not user-controlled.
func SessionPassword(serverSecret []byte, did string) []byte {
	return argon2.IDKey(serverSecret, []byte(did), 1, 32*1024, 2, 32)
}
