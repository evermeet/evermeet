package api

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/evermeet/evermeet/internal/identity"
)

// decryptUserKey retrieves and decrypts the Ed25519 signing key for the given DID.
func (s *Server) decryptUserKey(ctx context.Context, did string) (ed25519.PrivateKey, error) {
	priv, err := s.db.GetUserPrivate(ctx, did)
	if err != nil || priv == nil {
		return nil, fmt.Errorf("user private record not found for %s", did)
	}
	password := identity.SessionPassword(s.serverSecret, did)
	return identity.DecryptKeypair(priv.EncSigningKey, password)
}
