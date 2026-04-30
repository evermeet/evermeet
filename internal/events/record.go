package events

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/evermeet/evermeet/internal/identity"
)

// FoundingDoc is hashed to produce the permanent event ID.
// It is immutable — stored once, never modified.
type FoundingDoc struct {
	Type      string  `json:"type"`       // always "event"
	Organizer string  `json:"organizer"`  // did:em: of creator
	Calendar  *string `json:"calendar"`   // optional calendar ID
	CreatedAt string  `json:"created_at"`
	Nonce     string  `json:"nonce"` // random hex, ensures uniqueness
	InstanceID string `json:"instance_id,omitempty"`
}

// Location is the physical or virtual location of an event.
type Location struct {
	Name    string  `json:"name"`
	Address string  `json:"address,omitempty"`
	Lat     float64 `json:"lat,omitempty"`
	Lon     float64 `json:"lon,omitempty"`
	URL     string  `json:"url,omitempty"` // for virtual events
}

// RSVPPolicy controls RSVP behaviour for an event.
type RSVPPolicy struct {
	Limit    int    `json:"limit,omitempty"`
	Count    int    `json:"count"`
	Deadline string `json:"deadline,omitempty"`
	Approval string `json:"approval"` // auto | manual
}

// GovernanceOwner is one entry in the governance block.
type GovernanceOwner struct {
	DID  string `json:"did"`
	Role string `json:"role"` // owner | editor
}

// Governance defines who can update the record and how many signatures are required.
type Governance struct {
	Threshold int               `json:"threshold"`
	Owners    []GovernanceOwner `json:"owners"`
}

// Sig is one entry in the sigs array on a mutable state record.
type Sig struct {
	DID string `json:"did"`
	Sig string `json:"sig"`
}

// MutableState is the signed, updateable part of an event record.
type MutableState struct {
	ID          string      `json:"id"`
	Prev        string      `json:"prev,omitempty"`
	Organizer   string      `json:"organizer"`
	Calendar    *string     `json:"calendar,omitempty"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	CoverURL    string      `json:"cover_url,omitempty"`
	StartsAt    string      `json:"starts_at"`
	EndsAt      string      `json:"ends_at,omitempty"`
	Location    *Location   `json:"location,omitempty"`
	Governance  Governance  `json:"governance"`
	RSVP        RSVPPolicy  `json:"rsvp"`
	Visibility  string      `json:"visibility"` // public | unlisted | private
	Tags        []string    `json:"tags,omitempty"`
	UpdatedAt   string      `json:"updated_at"`
	Sigs        []Sig       `json:"sigs"`
}

// Fields is the user-supplied input when creating or updating an event.
type Fields struct {
	Title        string
	Description  string
	CoverURL     string
	StartsAt     time.Time
	EndsAt       *time.Time
	Location     *Location
	CalendarID   *string
	Visibility   string
	RSVPLimit    int
	RSVPApproval string
	RSVPDeadline *time.Time
	Tags         []string
}

// New creates a new founding doc and signed initial mutable state.
// Returns (foundingDoc, eventID, mutableState, stateHash, error).
func New(organizerDID string, priv ed25519.PrivateKey, homeHost string, f Fields) (*FoundingDoc, string, *MutableState, string, error) {
	if f.CalendarID == nil || *f.CalendarID == "" {
		return nil, "", nil, "", fmt.Errorf("calendar_id is required")
	}

	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, "", nil, "", fmt.Errorf("nonce: %w", err)
	}

	now := time.Now().UTC()
	founding := &FoundingDoc{
		Type:      "event",
		Organizer: organizerDID,
		Calendar:  f.CalendarID,
		CreatedAt: now.Format(time.RFC3339),
		Nonce:     hex.EncodeToString(nonce),
		InstanceID: homeHost,
	}

	eventID, err := identity.ContentHash(founding)
	if err != nil {
		return nil, "", nil, "", fmt.Errorf("hash founding doc: %w", err)
	}

	state, stateHash, err := buildState(eventID, "", organizerDID, priv, f, now)
	if err != nil {
		return nil, "", nil, "", err
	}

	return founding, eventID, state, stateHash, nil
}

// Update produces a new signed mutable state updating the current one.
func Update(current *MutableState, currentHash string, signerDID string, priv ed25519.PrivateKey, f Fields) (*MutableState, string, error) {
	now := time.Now().UTC()

	// Verify the signer is an authorized owner.
	if !isOwner(current.Governance, signerDID) {
		return nil, "", fmt.Errorf("signer %s is not an owner of event %s", signerDID, current.ID)
	}

	state, hash, err := buildState(current.ID, currentHash, signerDID, priv, f, now)
	if err != nil {
		return nil, "", err
	}

	// Preserve RSVP count from current state.
	state.RSVP.Count = current.RSVP.Count

	return state, hash, nil
}

// Verify checks that the mutable state has a valid signature from an authorized owner.
// For threshold > 1, all required sigs must be valid.
func Verify(state *MutableState, resolveKey func(did string) (ed25519.PublicKey, error)) error {
	if len(state.Sigs) < state.Governance.Threshold {
		return fmt.Errorf("insufficient signatures: have %d, need %d", len(state.Sigs), state.Governance.Threshold)
	}

	// Build the unsigned version for verification.
	unsigned := *state
	unsigned.Sigs = nil

	for _, s := range state.Sigs {
		if !isOwner(state.Governance, s.DID) {
			return fmt.Errorf("signer %s is not in governance owners", s.DID)
		}
		pub, err := resolveKey(s.DID)
		if err != nil {
			return fmt.Errorf("resolve key for %s: %w", s.DID, err)
		}
		ok, err := identity.Verify(pub, unsigned, s.Sig)
		if err != nil {
			return fmt.Errorf("verify sig from %s: %w", s.DID, err)
		}
		if !ok {
			return fmt.Errorf("invalid sig from %s", s.DID)
		}
	}
	return nil
}

func buildState(id, prev, signerDID string, priv ed25519.PrivateKey, f Fields, now time.Time) (*MutableState, string, error) {
	vis := f.Visibility
	if vis == "" {
		vis = "public"
	}
	approval := f.RSVPApproval
	if approval == "" {
		approval = "auto"
	}

	rsvp := RSVPPolicy{
		Limit:    f.RSVPLimit,
		Approval: approval,
	}
	if f.RSVPDeadline != nil {
		rsvp.Deadline = f.RSVPDeadline.UTC().Format(time.RFC3339)
	}

	var endsAt string
	if f.EndsAt != nil {
		endsAt = f.EndsAt.UTC().Format(time.RFC3339)
	}

	state := &MutableState{
		ID:          id,
		Prev:        prev,
		Organizer:   signerDID,
		Calendar:    f.CalendarID,
		Title:       f.Title,
		Description: f.Description,
		CoverURL:    f.CoverURL,
		StartsAt:    f.StartsAt.UTC().Format(time.RFC3339),
		EndsAt:      endsAt,
		Location:    f.Location,
		Governance: Governance{
			Threshold: 1,
			Owners:    []GovernanceOwner{{DID: signerDID, Role: "owner"}},
		},
		RSVP:       rsvp,
		Visibility: vis,
		Tags:       f.Tags,
		UpdatedAt:  now.Format(time.RFC3339),
		Sigs:       nil, // populated below
	}

	sig, err := identity.Sign(priv, state)
	if err != nil {
		return nil, "", fmt.Errorf("sign state: %w", err)
	}
	state.Sigs = []Sig{{DID: signerDID, Sig: sig}}

	// Hash includes the sigs.
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

// RSVP is the private content of an RSVP envelope.
// It is encrypted for the event organizer.
type RSVP struct {
	EventID   string `json:"event_id"`
	SenderDID string `json:"sender_did"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Note      string `json:"note,omitempty"`
	Timestamp string `json:"timestamp"`
}
