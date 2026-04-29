package calendar

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/evermeet/evermeet/internal/identity"
)

// FoundingDoc is hashed to produce the permanent calendar ID.
type FoundingDoc struct {
	Type      string   `json:"type"`       // always "calendar"
	Owners    []string `json:"owners"`     // initial did:em: owner list
	CreatedAt string   `json:"created_at"`
	Nonce     string   `json:"nonce"`
	HomeHost  string   `json:"home_host,omitempty"`
}

// Sig is one entry in the sigs array on a mutable state record.
type Sig struct {
	DID string `json:"did"`
	Sig string `json:"sig"`
}

// GovernanceOwner is one entry in the governance block.
type GovernanceOwner struct {
	DID  string `json:"did"`
	Role string `json:"role"` // owner | editor
}

// Governance defines update rules.
type Governance struct {
	Threshold int               `json:"threshold"`
	Owners    []GovernanceOwner `json:"owners"`
}

// MutableState is the signed, updateable part of a calendar record.
type MutableState struct {
	ID          string     `json:"id"`
	Prev        string     `json:"prev,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Avatar      string     `json:"avatar,omitempty"`
	Website     string     `json:"website,omitempty"`
	Governance  Governance `json:"governance"`
	UpdatedAt   string     `json:"updated_at"`
	Sigs        []Sig      `json:"sigs"`
}

// Fields is the user-supplied input for creating or updating a calendar.
type Fields struct {
	Name        string
	Description string
	Avatar      string
	Website     string
}

// New creates a new founding doc and signed initial mutable state.
func New(ownerDID string, priv ed25519.PrivateKey, homeHost string, f Fields) (*FoundingDoc, string, *MutableState, string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, "", nil, "", fmt.Errorf("nonce: %w", err)
	}

	now := time.Now().UTC()
	founding := &FoundingDoc{
		Type:      "calendar",
		Owners:    []string{ownerDID},
		CreatedAt: now.Format(time.RFC3339),
		Nonce:     hex.EncodeToString(nonce),
		HomeHost:  homeHost,
	}

	calID, err := identity.ContentHash(founding)
	if err != nil {
		return nil, "", nil, "", fmt.Errorf("hash founding doc: %w", err)
	}

	state, stateHash, err := buildState(calID, "", ownerDID, priv, f, now)
	if err != nil {
		return nil, "", nil, "", err
	}

	return founding, calID, state, stateHash, nil
}

// Update produces a new signed mutable state.
func Update(current *MutableState, currentHash, signerDID string, priv ed25519.PrivateKey, f Fields) (*MutableState, string, error) {
	if !isOwner(current.Governance, signerDID) {
		return nil, "", fmt.Errorf("signer %s is not an owner", signerDID)
	}
	return buildState(current.ID, currentHash, signerDID, priv, f, time.Now().UTC())
}

// Verify checks all signatures on the mutable state against the governance threshold.
func Verify(state *MutableState, resolveKey func(did string) (ed25519.PublicKey, error)) error {
	if len(state.Sigs) < state.Governance.Threshold {
		return fmt.Errorf("insufficient signatures: have %d, need %d", len(state.Sigs), state.Governance.Threshold)
	}
	unsigned := *state
	unsigned.Sigs = nil
	for _, s := range state.Sigs {
		if !isOwner(state.Governance, s.DID) {
			return fmt.Errorf("signer %s not in governance", s.DID)
		}
		pub, err := resolveKey(s.DID)
		if err != nil {
			return fmt.Errorf("resolve key for %s: %w", s.DID, err)
		}
		ok, err := identity.Verify(pub, unsigned, s.Sig)
		if err != nil || !ok {
			return fmt.Errorf("invalid sig from %s", s.DID)
		}
	}
	return nil
}

func buildState(id, prev, signerDID string, priv ed25519.PrivateKey, f Fields, now time.Time) (*MutableState, string, error) {
	state := &MutableState{
		ID:          id,
		Prev:        prev,
		Name:        f.Name,
		Description: f.Description,
		Avatar:      f.Avatar,
		Website:     f.Website,
		Governance: Governance{
			Threshold: 1,
			Owners:    []GovernanceOwner{{DID: signerDID, Role: "owner"}},
		},
		UpdatedAt: now.Format(time.RFC3339),
		Sigs:      nil,
	}

	sig, err := identity.Sign(priv, state)
	if err != nil {
		return nil, "", fmt.Errorf("sign state: %w", err)
	}
	state.Sigs = []Sig{{DID: signerDID, Sig: sig}}

	hash, err := identity.ContentHash(state)
	if err != nil {
		return nil, "", err
	}
	return state, hash, nil
}

func isOwner(g Governance, did string) bool {
	for _, o := range g.Owners {
		if o.DID == did {
			return true
		}
	}
	return false
}
