package identity

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"time"
)

// KELOpType identifies the operation type in a key event log entry.
type KELOpType string

const (
	KELOpGenesis  KELOpType = "genesis"
	KELOpRotate   KELOpType = "rotate"
	KELOpMigrate  KELOpType = "migrate"
)

// KELOp is one signed entry in a key event log. The hash of its canonical
// JSON is its stable identity and is used as the prev pointer in the next op.
type KELOp struct {
	Type       KELOpType `json:"type"`
	PK         string    `json:"pk"`                   // hex Ed25519 signing public key
	RotationPK string    `json:"rotation_pk"`          // hex Ed25519 rotation public key
	NewPK      string    `json:"new_pk,omitempty"`     // rotate only
	Endpoint   string    `json:"endpoint,omitempty"`   // genesis + migrate
	Prev       string    `json:"prev,omitempty"`       // empty for genesis
	Seq        int       `json:"seq"`
	CreatedAt  string    `json:"created_at"`
	Sig        string    `json:"sig"` // base64url Ed25519 signature over canonical JSON of the op minus sig
}

// KELState is the current resolved state after replaying a key event log.
type KELState struct {
	DID        string
	CurrentPK  ed25519.PublicKey
	RotationPK ed25519.PublicKey
	Endpoint   string
	Seq        int
	HeadHash   string // hash of the last op
}

// BuildGenesisOp creates the first entry in a key event log, signed by the signing key.
func BuildGenesisOp(sigPriv, rotPriv ed25519.PrivateKey, endpoint string) (*KELOp, string, error) {
	sigPub := sigPriv.Public().(ed25519.PublicKey)
	rotPub := rotPriv.Public().(ed25519.PublicKey)

	op := &KELOp{
		Type:       KELOpGenesis,
		PK:         hex.EncodeToString(sigPub),
		RotationPK: hex.EncodeToString(rotPub),
		Endpoint:   endpoint,
		Seq:        0,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	sig, err := signKELOp(sigPriv, op)
	if err != nil {
		return nil, "", err
	}
	op.Sig = sig

	hash, err := ContentHash(op)
	if err != nil {
		return nil, "", err
	}
	return op, hash, nil
}

// BuildRotateOp creates a key rotation op. Must be signed by the current rotation key.
func BuildRotateOp(current *KELState, newSigPriv, newRotPriv ed25519.PrivateKey, rotationPriv ed25519.PrivateKey, prevHash string) (*KELOp, string, error) {
	newSigPub := newSigPriv.Public().(ed25519.PublicKey)
	newRotPub := newRotPriv.Public().(ed25519.PublicKey)

	op := &KELOp{
		Type:       KELOpRotate,
		PK:         hex.EncodeToString(current.CurrentPK),
		RotationPK: hex.EncodeToString(current.RotationPK),
		NewPK:      hex.EncodeToString(newSigPub),
		Prev:       prevHash,
		Seq:        current.Seq + 1,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	// Update rotation_pk to new rotation key in this op.
	op.RotationPK = hex.EncodeToString(newRotPub)

	sig, err := signKELOp(rotationPriv, op)
	if err != nil {
		return nil, "", err
	}
	op.Sig = sig

	hash, err := ContentHash(op)
	if err != nil {
		return nil, "", err
	}
	return op, hash, nil
}

// BuildMigrateOp creates an endpoint migration op. Signed by the current signing key.
func BuildMigrateOp(current *KELState, sigPriv ed25519.PrivateKey, newEndpoint, prevHash string) (*KELOp, string, error) {
	op := &KELOp{
		Type:       KELOpMigrate,
		PK:         hex.EncodeToString(current.CurrentPK),
		RotationPK: hex.EncodeToString(current.RotationPK),
		Endpoint:   newEndpoint,
		Prev:       prevHash,
		Seq:        current.Seq + 1,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	sig, err := signKELOp(sigPriv, op)
	if err != nil {
		return nil, "", err
	}
	op.Sig = sig

	hash, err := ContentHash(op)
	if err != nil {
		return nil, "", err
	}
	return op, hash, nil
}

// ReplayKEL replays a sequence of ops (in ascending seq order) and returns the
// current state. Returns an error if any op fails verification.
func ReplayKEL(ops []*KELOp) (*KELState, error) {
	if len(ops) == 0 {
		return nil, fmt.Errorf("empty key event log")
	}

	genesis := ops[0]
	if genesis.Type != KELOpGenesis {
		return nil, fmt.Errorf("first op must be genesis, got %s", genesis.Type)
	}
	if genesis.Prev != "" {
		return nil, fmt.Errorf("genesis op must have no prev")
	}

	sigPub, err := hexPubKey(genesis.PK)
	if err != nil {
		return nil, fmt.Errorf("genesis pk: %w", err)
	}
	if err := verifyKELOp(sigPub, genesis); err != nil {
		return nil, fmt.Errorf("genesis sig: %w", err)
	}

	rotPub, err := hexPubKey(genesis.RotationPK)
	if err != nil {
		return nil, fmt.Errorf("genesis rotation_pk: %w", err)
	}

	did := DeriveDID(sigPub)
	genesisHash, err := ContentHash(genesis)
	if err != nil {
		return nil, err
	}

	state := &KELState{
		DID:        did,
		CurrentPK:  sigPub,
		RotationPK: rotPub,
		Endpoint:   genesis.Endpoint,
		Seq:        0,
		HeadHash:   genesisHash,
	}

	for i, op := range ops[1:] {
		expectedSeq := i + 1
		if op.Seq != expectedSeq {
			return nil, fmt.Errorf("op %d: expected seq %d, got %d", i+1, expectedSeq, op.Seq)
		}
		if op.Prev != state.HeadHash {
			return nil, fmt.Errorf("op %d: prev hash mismatch", i+1)
		}

		switch op.Type {
		case KELOpRotate:
			// Rotate must be signed by the current rotation key.
			if err := verifyKELOp(state.RotationPK, op); err != nil {
				return nil, fmt.Errorf("rotate op %d sig: %w", i+1, err)
			}
			newPK, err := hexPubKey(op.NewPK)
			if err != nil {
				return nil, fmt.Errorf("rotate new_pk: %w", err)
			}
			newRotPK, err := hexPubKey(op.RotationPK)
			if err != nil {
				return nil, fmt.Errorf("rotate rotation_pk: %w", err)
			}
			state.CurrentPK = newPK
			state.RotationPK = newRotPK

		case KELOpMigrate:
			// Migrate must be signed by the current signing key.
			if err := verifyKELOp(state.CurrentPK, op); err != nil {
				return nil, fmt.Errorf("migrate op %d sig: %w", i+1, err)
			}
			state.Endpoint = op.Endpoint

		default:
			return nil, fmt.Errorf("op %d: unknown type %s", i+1, op.Type)
		}

		state.Seq = op.Seq
		state.HeadHash, err = ContentHash(op)
		if err != nil {
			return nil, err
		}
	}

	return state, nil
}

// ValidateKELOp validates a single incoming op against the current state.
// Used when receiving ops over the gossipsub network.
func ValidateKELOp(op *KELOp, current *KELState) error {
	if op.Seq != current.Seq+1 {
		return fmt.Errorf("seq mismatch: expected %d, got %d", current.Seq+1, op.Seq)
	}
	if op.Prev != current.HeadHash {
		return fmt.Errorf("prev hash mismatch")
	}

	switch op.Type {
	case KELOpRotate:
		return verifyKELOp(current.RotationPK, op)
	case KELOpMigrate:
		return verifyKELOp(current.CurrentPK, op)
	default:
		return fmt.Errorf("unknown op type: %s", op.Type)
	}
}

// signKELOp signs the op with the given key. The sig field must be empty before signing.
func signKELOp(priv ed25519.PrivateKey, op *KELOp) (string, error) {
	// Clone without sig field for signing.
	unsigned := *op
	unsigned.Sig = ""
	return Sign(priv, unsigned)
}

// verifyKELOp verifies the op signature using the given public key.
func verifyKELOp(pub ed25519.PublicKey, op *KELOp) error {
	sig := op.Sig
	unsigned := *op
	unsigned.Sig = ""
	ok, err := Verify(pub, unsigned, sig)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

func hexPubKey(s string) (ed25519.PublicKey, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("expected %d bytes, got %d", ed25519.PublicKeySize, len(b))
	}
	return ed25519.PublicKey(b), nil
}
