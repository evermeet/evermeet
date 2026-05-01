package identity

import (
	"fmt"
	"strings"
)

// NormalizeEvermeetDID returns a canonical lowercase form of an Evermeet DID
// (did:em:…) for routing and comparisons.
func NormalizeEvermeetDID(did string) (string, error) {
	d := strings.ToLower(strings.TrimSpace(did))
	if !strings.HasPrefix(d, "did:em:") {
		return "", fmt.Errorf("invalid DID: expected did:em: prefix")
	}
	if len(d) < len("did:em:")+8 {
		return "", fmt.Errorf("invalid DID")
	}
	return d, nil
}
