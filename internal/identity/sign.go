package identity

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/multiformats/go-multibase"
	"lukechampine.com/blake3"
)

// CanonicalJSON serializes v to JCS (RFC 8785) canonical JSON:
// sorted object keys, no whitespace, deterministic number encoding.
// This is the format used for all signatures in the Evermeet protocol.
func CanonicalJSON(v any) ([]byte, error) {
	// Marshal to get a generic representation, then re-serialize canonically.
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, err
	}
	return canonicalize(generic)
}

func canonicalize(v any) ([]byte, error) {
	switch val := v.(type) {
	case nil:
		return []byte("null"), nil
	case bool:
		if val {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	case float64:
		return canonicalNumber(val)
	case string:
		return json.Marshal(val) // stdlib handles Unicode correctly
	case []any:
		var sb strings.Builder
		sb.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				sb.WriteByte(',')
			}
			b, err := canonicalize(item)
			if err != nil {
				return nil, err
			}
			sb.Write(b)
		}
		sb.WriteByte(']')
		return []byte(sb.String()), nil
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var sb strings.Builder
		sb.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				sb.WriteByte(',')
			}
			keyBytes, err := json.Marshal(k)
			if err != nil {
				return nil, err
			}
			sb.Write(keyBytes)
			sb.WriteByte(':')
			valBytes, err := canonicalize(val[k])
			if err != nil {
				return nil, err
			}
			sb.Write(valBytes)
		}
		sb.WriteByte('}')
		return []byte(sb.String()), nil
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
	}
}

// canonicalNumber serializes a float64 per JCS (RFC 8785 §3.2.2.3):
// integers are rendered without a decimal point or exponent when possible;
// other values use ES6 number-to-string semantics (shortest representation).
func canonicalNumber(f float64) ([]byte, error) {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return nil, fmt.Errorf("JSON does not support Inf or NaN")
	}
	// If the value is an integer in safe integer range, emit as integer.
	if f == math.Trunc(f) && f >= -9007199254740992 && f <= 9007199254740992 {
		return []byte(fmt.Sprintf("%.0f", f)), nil
	}
	// Otherwise use strconv shortest representation (matches ES6 behavior for normal floats).
	// json.Marshal produces shortest-round-trip representation.
	return json.Marshal(f)
}

// ContentHash returns the base32-encoded BLAKE3 hash of the canonical JSON of v.
// This is used for record IDs and prev-chain links.
func ContentHash(v any) (string, error) {
	b, err := CanonicalJSON(v)
	if err != nil {
		return "", err
	}
	return ContentHashBytes(b), nil
}

// ContentHashBytes returns the base32-encoded BLAKE3 hash of raw bytes.
func ContentHashBytes(b []byte) string {
	h := blake3.Sum256(b)
	encoded, _ := multibase.Encode(multibase.Base32, h[:])
	return encoded[1:] // strip multibase prefix byte
}

// Sign signs the canonical JSON of v with the given Ed25519 private key.
// Returns a base64-encoded signature.
func Sign(priv ed25519.PrivateKey, v any) (string, error) {
	b, err := CanonicalJSON(v)
	if err != nil {
		return "", fmt.Errorf("canonical JSON: %w", err)
	}
	sig := ed25519.Sign(priv, b)
	return base64.RawURLEncoding.EncodeToString(sig), nil
}

// SignBytes signs raw bytes directly.
func SignBytes(priv ed25519.PrivateKey, b []byte) string {
	sig := ed25519.Sign(priv, b)
	return base64.RawURLEncoding.EncodeToString(sig)
}

// Verify verifies an Ed25519 signature over the canonical JSON of v.
// sig is expected to be base64url-encoded (no padding).
func Verify(pub ed25519.PublicKey, v any, sig string) (bool, error) {
	b, err := CanonicalJSON(v)
	if err != nil {
		return false, fmt.Errorf("canonical JSON: %w", err)
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return false, fmt.Errorf("decode sig: %w", err)
	}
	return ed25519.Verify(pub, b, sigBytes), nil
}

// VerifyBytes verifies an Ed25519 signature over raw bytes.
func VerifyBytes(pub ed25519.PublicKey, b []byte, sig string) (bool, error) {
	sigBytes, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return false, fmt.Errorf("decode sig: %w", err)
	}
	return ed25519.Verify(pub, b, sigBytes), nil
}
