package api

import (
	"github.com/evermeet/evermeet/internal/store"
	"github.com/go-webauthn/webauthn/webauthn"
)

// waUser wraps store.User and store.UserPrivate to implement webauthn.User.
type waUser struct {
	user  *store.User
	creds []webauthn.Credential
}

func (u *waUser) WebAuthnID() []byte {
	return []byte(u.user.DID)
}

func (u *waUser) WebAuthnName() string {
	if u.user.DisplayName != "" {
		return u.user.DisplayName
	}
	return u.user.DID
}

func (u *waUser) WebAuthnDisplayName() string {
	return u.WebAuthnName()
}

func (u *waUser) WebAuthnIcon() string {
	return u.user.Avatar
}

func (u *waUser) WebAuthnCredentials() []webauthn.Credential {
	return u.creds
}

func (s *Server) wrapUser(user *store.User, passkeys []*store.Passkey) *waUser {
	creds := make([]webauthn.Credential, len(passkeys))
	for i, p := range passkeys {
		creds[i] = webauthn.Credential{
			ID:              p.ID,
			PublicKey:       p.PublicKey,
			AttestationType: p.AttestationType,
			// Transport: ... (omitted for now, standard default is fine)
			Flags: webauthn.NewCredentialFlags(0), // We'll set the bools manually
			Authenticator: webauthn.Authenticator{
				SignCount: uint32(p.Counter),
			},
		}
		creds[i].Flags.BackupEligible = p.BackupEligible
		creds[i].Flags.BackupState = p.BackupState
	}
	return &waUser{user: user, creds: creds}
}
