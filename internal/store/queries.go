package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ---- SIWE nonces ----

type SIWENonce struct {
	Nonce     string
	Address   string
	ChainID   string
	CreatedAt time.Time
	ExpiresAt time.Time
	Used      bool
}

type SIWEIdentity struct {
	ChainID string
	Address string
}

func (d *DB) InsertSIWENonce(ctx context.Context, n *SIWENonce) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO siwe_nonces (nonce, address, chain_id, created_at, expires_at, used)
			 VALUES (?, ?, ?, ?, ?, 0)`,
			n.Nonce, n.Address, n.ChainID,
			n.CreatedAt.UTC().Format(time.RFC3339),
			n.ExpiresAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) GetSIWENonce(ctx context.Context, nonce string) (*SIWENonce, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT nonce, address, chain_id, created_at, expires_at, used
		 FROM siwe_nonces WHERE nonce = ?`, nonce)
	var n SIWENonce
	var ca, ea string
	var used int
	if err := row.Scan(&n.Nonce, &n.Address, &n.ChainID, &ca, &ea, &used); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("get siwe nonce: %w", err)
	}
	n.CreatedAt, _ = time.Parse(time.RFC3339, ca)
	n.ExpiresAt, _ = time.Parse(time.RFC3339, ea)
	n.Used = used != 0
	return &n, nil
}

func (d *DB) MarkSIWENonceUsed(ctx context.Context, nonce string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE siwe_nonces SET used = 1 WHERE nonce = ?`, nonce)
		return err
	})
}

func (d *DB) GetLocalSIWEIdentities(ctx context.Context) ([]SIWEIdentity, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT siwe_chain_id, siwe_address FROM user_private WHERE siwe_chain_id != '' AND siwe_address != ''`)
	if err != nil {
		return nil, fmt.Errorf("get local siwe identities: %w", err)
	}
	defer rows.Close()

	var identities []SIWEIdentity
	for rows.Next() {
		var identity SIWEIdentity
		if err := rows.Scan(&identity.ChainID, &identity.Address); err != nil {
			return nil, err
		}
		identities = append(identities, identity)
	}
	return identities, rows.Err()
}

// ---- Cross-instance nonces ----

type CrossInstanceNonce struct {
	Nonce      string
	ForeignURL string
	EventID    string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	Used       bool
}

func (d *DB) InsertNonce(ctx context.Context, n *CrossInstanceNonce) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO cross_instance_nonces (nonce, foreign_url, event_id, created_at, expires_at, used)
			 VALUES (?, ?, ?, ?, ?, 0)`,
			n.Nonce, n.ForeignURL, n.EventID,
			n.CreatedAt.UTC().Format(time.RFC3339),
			n.ExpiresAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) GetNonce(ctx context.Context, nonce string) (*CrossInstanceNonce, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT nonce, foreign_url, event_id, created_at, expires_at, used
		 FROM cross_instance_nonces WHERE nonce = ?`, nonce)
	var n CrossInstanceNonce
	var ca, ea string
	var used int
	if err := row.Scan(&n.Nonce, &n.ForeignURL, &n.EventID, &ca, &ea, &used); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("get nonce: %w", err)
	}
	n.CreatedAt, _ = time.Parse(time.RFC3339, ca)
	n.ExpiresAt, _ = time.Parse(time.RFC3339, ea)
	n.Used = used != 0
	return &n, nil
}

func (d *DB) MarkNonceUsed(ctx context.Context, nonce string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE cross_instance_nonces SET used = 1 WHERE nonce = ?`, nonce)
		return err
	})
}

// ---- Blobs ----

type BlobRecord struct {
	Hash        string
	ContentType string
	Size        int64
	UploadedBy  string
	CreatedAt   time.Time
}

type BlobSource struct {
	Hash        string
	InstanceURL string
	CreatedAt   time.Time
}

func (d *DB) InsertBlob(ctx context.Context, b *BlobRecord) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO blobs (hash, content_type, size, uploaded_by, created_at) VALUES (?,?,?,?,?)`,
			b.Hash, b.ContentType, b.Size, b.UploadedBy, b.CreatedAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) GetBlob(ctx context.Context, hash string) (*BlobRecord, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT hash, content_type, size, uploaded_by, created_at FROM blobs WHERE hash = ?`, hash)
	var b BlobRecord
	var ca string
	if err := row.Scan(&b.Hash, &b.ContentType, &b.Size, &b.UploadedBy, &ca); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("get blob: %w", err)
	}
	b.CreatedAt, _ = time.Parse(time.RFC3339, ca)
	return &b, nil
}

func (d *DB) InsertBlobSource(ctx context.Context, src *BlobSource) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO blob_sources (hash, instance_url, created_at) VALUES (?, ?, ?)`,
			src.Hash, src.InstanceURL, src.CreatedAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) ListBlobSources(ctx context.Context, hash string) ([]*BlobSource, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT hash, instance_url, created_at FROM blob_sources WHERE hash = ? ORDER BY created_at ASC`, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*BlobSource
	for rows.Next() {
		src := &BlobSource{}
		var createdAt string
		if err := rows.Scan(&src.Hash, &src.InstanceURL, &createdAt); err != nil {
			return nil, err
		}
		src.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		out = append(out, src)
	}
	return out, rows.Err()
}

// ---- Key event log ----

type KELOp struct {
	Hash      string
	DID       string
	Type      string // genesis | rotate | migrate
	Payload   string // canonical JSON
	Prev      string // empty for genesis
	Seq       int
	CreatedAt time.Time
}

func (d *DB) AppendKELOp(ctx context.Context, op *KELOp) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO kel_ops (hash, did, type, payload, prev, seq, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			op.Hash, op.DID, op.Type, op.Payload,
			nullString(op.Prev), op.Seq, op.CreatedAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) GetKELForDID(ctx context.Context, did string) ([]*KELOp, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT hash, did, type, payload, COALESCE(prev,''), seq, created_at
		 FROM kel_ops WHERE did = ? ORDER BY seq ASC`, did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanKELOps(rows)
}

func (d *DB) GetKELOp(ctx context.Context, hash string) (*KELOp, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT hash, did, type, payload, COALESCE(prev,''), seq, created_at FROM kel_ops WHERE hash = ?`, hash)
	return scanKELOp(row)
}

// ---- Users ----

type User struct {
	DID         string
	DisplayName string
	Avatar      string
	Bio         string
	CurrentPK   string // hex-encoded Ed25519 public key
	RotationPK  string
	Endpoint    string
	Sig         string
	UpdatedAt   time.Time
	InstanceID  string
}

func (d *DB) UpsertUser(ctx context.Context, u *User) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO users (did, display_name, avatar, bio, current_pk, rotation_pk, endpoint, sig, updated_at, home_host)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT(did) DO UPDATE SET
			   display_name = excluded.display_name,
			   avatar       = excluded.avatar,
			   bio          = excluded.bio,
			   current_pk   = excluded.current_pk,
			   rotation_pk  = excluded.rotation_pk,
			   endpoint     = excluded.endpoint,
			   sig          = excluded.sig,
			   updated_at   = excluded.updated_at,
			   home_host    = excluded.home_host`,
			u.DID, u.DisplayName, u.Avatar, u.Bio,
			u.CurrentPK, u.RotationPK, u.Endpoint, u.Sig,
			u.UpdatedAt.UTC().Format(time.RFC3339), u.InstanceID,
		)
		return err
	})
}

func (d *DB) GetUser(ctx context.Context, did string) (*User, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT did, display_name, avatar, bio, current_pk, rotation_pk, endpoint, sig, updated_at, COALESCE(home_host,'')
		 FROM users WHERE did = ?`, did)
	u := &User{}
	var updatedAt string
	err := row.Scan(&u.DID, &u.DisplayName, &u.Avatar, &u.Bio,
		&u.CurrentPK, &u.RotationPK, &u.Endpoint, &u.Sig, &updatedAt, &u.InstanceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return u, nil
}

// ---- User private ----

type UserPrivate struct {
	DID           string
	Email         string
	EmailVerified bool
	GoogleSub     string
	SIWEChainID   string
	SIWEAddress   string
	PasskeyIDs    string // JSON array
	EncSigningKey string // AES-256-GCM encrypted Ed25519 seed
}

func (d *DB) UpsertUserPrivate(ctx context.Context, p *UserPrivate) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO user_private (did, email, email_verified, google_sub, siwe_chain_id, siwe_address, passkey_ids, enc_signing_key)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT(did) DO UPDATE SET
			   email          = excluded.email,
			   email_verified = excluded.email_verified,
			   google_sub     = excluded.google_sub,
			   passkey_ids    = excluded.passkey_ids,
			   enc_signing_key = excluded.enc_signing_key,
			   siwe_chain_id  = COALESCE(NULLIF(excluded.siwe_chain_id, ''), siwe_chain_id),
			   siwe_address   = COALESCE(NULLIF(excluded.siwe_address, ''), siwe_address)`,
			p.DID, p.Email, boolInt(p.EmailVerified),
			p.GoogleSub, p.SIWEChainID, p.SIWEAddress, p.PasskeyIDs, p.EncSigningKey,
		)
		return err
	})
}

func (d *DB) GetUserPrivate(ctx context.Context, did string) (*UserPrivate, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT did, COALESCE(email,''), email_verified, COALESCE(google_sub,''), COALESCE(siwe_chain_id,''), COALESCE(siwe_address,''),
		        COALESCE(passkey_ids,''), COALESCE(enc_signing_key,'')
		 FROM user_private WHERE did = ?`, did)
	p := &UserPrivate{}
	var ev int
	err := row.Scan(&p.DID, &p.Email, &ev, &p.GoogleSub, &p.SIWEChainID, &p.SIWEAddress, &p.PasskeyIDs, &p.EncSigningKey)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.EmailVerified = ev != 0
	return p, nil
}

func (d *DB) GetUserPrivateByEmail(ctx context.Context, email string) (*UserPrivate, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT did, COALESCE(email,''), email_verified, COALESCE(google_sub,''), COALESCE(siwe_chain_id,''), COALESCE(siwe_address,''),
		        COALESCE(passkey_ids,''), COALESCE(enc_signing_key,'')
		 FROM user_private WHERE lower(trim(email)) = lower(trim(?))`, email)
	p := &UserPrivate{}
	var ev int
	err := row.Scan(&p.DID, &p.Email, &ev, &p.GoogleSub, &p.SIWEChainID, &p.SIWEAddress, &p.PasskeyIDs, &p.EncSigningKey)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.EmailVerified = ev != 0
	return p, nil
}

// GetAllUserEmails returns all non-empty email addresses stored on this instance.
// Used by the DHT heartbeat to re-publish home-routing records.
func (d *DB) GetAllUserEmails(ctx context.Context) ([]string, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT email FROM user_private WHERE email != ''`)
	if err != nil {
		return nil, fmt.Errorf("get all emails: %w", err)
	}
	defer rows.Close()
	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, rows.Err()
}

func (d *DB) GetUserPrivateBySIWE(ctx context.Context, chainID, address string) (*UserPrivate, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT did, COALESCE(email,''), email_verified, COALESCE(google_sub,''), COALESCE(siwe_chain_id,''), COALESCE(siwe_address,''),
		        COALESCE(passkey_ids,''), COALESCE(enc_signing_key,'')
		 FROM user_private WHERE siwe_chain_id = ? AND siwe_address = ?`, chainID, address)
	p := &UserPrivate{}
	var ev int
	err := row.Scan(&p.DID, &p.Email, &ev, &p.GoogleSub, &p.SIWEChainID, &p.SIWEAddress, &p.PasskeyIDs, &p.EncSigningKey)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.EmailVerified = ev != 0
	return p, nil
}

func (d *DB) SetUserSIWE(ctx context.Context, did, chainID, address string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE user_private SET siwe_chain_id = ?, siwe_address = ? WHERE did = ?`, chainID, address, did)
		return err
	})
}

func (d *DB) GetUserPrivateByGoogleSub(ctx context.Context, sub string) (*UserPrivate, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT did, COALESCE(email,''), email_verified, COALESCE(google_sub,''), COALESCE(siwe_chain_id,''), COALESCE(siwe_address,''),
		        COALESCE(passkey_ids,''), COALESCE(enc_signing_key,'')
		 FROM user_private WHERE google_sub = ?`, sub)
	p := &UserPrivate{}
	var ev int
	err := row.Scan(&p.DID, &p.Email, &ev, &p.GoogleSub, &p.SIWEChainID, &p.SIWEAddress, &p.PasskeyIDs, &p.EncSigningKey)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.EmailVerified = ev != 0
	return p, nil
}

// ---- Event records ----

type EventFounding struct {
	ID      string
	Payload string // canonical JSON
}

type EventState struct {
	Hash      string
	ID        string
	Prev      string
	Payload   string // canonical JSON
	IsCurrent bool
	CreatedAt time.Time
}

func (d *DB) InsertEventFounding(ctx context.Context, f *EventFounding) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO event_founding (id, payload) VALUES (?, ?)`,
			f.ID, f.Payload,
		)
		return err
	})
}

func (d *DB) AppendEventState(ctx context.Context, s *EventState) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		// Demote previous current state.
		if _, err := tx.Exec(
			`UPDATE event_states SET is_current = 0 WHERE id = ? AND is_current = 1`, s.ID,
		); err != nil {
			return err
		}
		_, err := tx.Exec(
			`INSERT INTO event_states (hash, id, prev, payload, is_current, created_at)
			 VALUES (?, ?, ?, ?, 1, ?)`,
			s.Hash, s.ID, nullString(s.Prev), s.Payload,
			s.CreatedAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) GetCurrentEventState(ctx context.Context, id string) (*EventState, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT hash, id, COALESCE(prev,''), payload, is_current, created_at
		 FROM event_states WHERE id = ? AND is_current = 1`, id)
	return scanEventState(row)
}

func (d *DB) ListEventStatesByID(ctx context.Context, id string) ([]*EventState, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT hash, id, COALESCE(prev,''), payload, is_current, created_at
		 FROM event_states
		 WHERE id = ?
		 ORDER BY created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var states []*EventState
	for rows.Next() {
		s, err := scanEventStateRow(rows)
		if err != nil {
			return nil, err
		}
		states = append(states, s)
	}
	return states, rows.Err()
}

func (d *DB) GetEventFounding(ctx context.Context, id string) (*EventFounding, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT id, payload FROM event_founding WHERE id = ?`, id)
	f := &EventFounding{}
	err := row.Scan(&f.ID, &f.Payload)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return f, err
}

func (d *DB) ListCurrentEventStates(ctx context.Context, limit, offset int) ([]*EventState, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT hash, id, COALESCE(prev,''), payload, is_current, created_at
		 FROM event_states WHERE is_current = 1
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var states []*EventState
	for rows.Next() {
		s, err := scanEventStateRow(rows)
		if err != nil {
			return nil, err
		}
		states = append(states, s)
	}
	return states, rows.Err()
}

func (d *DB) DeleteEvent(ctx context.Context, id string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		if _, err := tx.Exec(`DELETE FROM rsvp_envelopes WHERE event_id = ?`, id); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM event_states WHERE id = ?`, id); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM event_founding WHERE id = ?`, id); err != nil {
			return err
		}
		return nil
	})
}

// ---- Calendar records ----

type CalendarFounding struct {
	ID      string
	Payload string
}

type CalendarState struct {
	Hash      string
	ID        string
	Prev      string
	Payload   string
	IsCurrent bool
	CreatedAt time.Time
}

func (d *DB) InsertCalendarFounding(ctx context.Context, f *CalendarFounding) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO calendar_founding (id, payload) VALUES (?, ?)`,
			f.ID, f.Payload,
		)
		return err
	})
}

func (d *DB) AppendCalendarState(ctx context.Context, s *CalendarState) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		if _, err := tx.Exec(
			`UPDATE calendar_states SET is_current = 0 WHERE id = ? AND is_current = 1`, s.ID,
		); err != nil {
			return err
		}
		_, err := tx.Exec(
			`INSERT INTO calendar_states (hash, id, prev, payload, is_current, created_at)
			 VALUES (?, ?, ?, ?, 1, ?)`,
			s.Hash, s.ID, nullString(s.Prev), s.Payload,
			s.CreatedAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) GetCurrentCalendarState(ctx context.Context, id string) (*CalendarState, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT hash, id, COALESCE(prev,''), payload, is_current, created_at
		 FROM calendar_states WHERE id = ? AND is_current = 1`, id)
	s := &CalendarState{}
	var createdAt string
	err := row.Scan(&s.Hash, &s.ID, &s.Prev, &s.Payload, &s.IsCurrent, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return s, nil
}

func (d *DB) ListCurrentCalendars(ctx context.Context) ([]*CalendarState, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT hash, id, COALESCE(prev,''), payload, is_current, created_at
		FROM calendar_states
		WHERE is_current = 1
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCalendarStates(rows)
}

func (d *DB) InsertCalendarOwner(ctx context.Context, calendarID, did string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO calendar_owners (calendar_id, did) VALUES (?, ?)`,
			calendarID, did,
		)
		return err
	})
}

func (d *DB) ReplaceCalendarOwners(ctx context.Context, calendarID string, dids []string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		if _, err := tx.Exec(`DELETE FROM calendar_owners WHERE calendar_id = ?`, calendarID); err != nil {
			return err
		}
		for _, did := range dids {
			if _, err := tx.Exec(
				`INSERT OR IGNORE INTO calendar_owners (calendar_id, did) VALUES (?, ?)`,
				calendarID, did,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *DB) IsCalendarOwner(ctx context.Context, calendarID, did string) (bool, error) {
	var count int
	err := d.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM calendar_owners WHERE calendar_id = ? AND did = ?`,
		calendarID, did,
	).Scan(&count)
	return count > 0, err
}

func (d *DB) ListOwnedCalendars(ctx context.Context, did string) ([]*CalendarState, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT cs.hash, cs.id, COALESCE(cs.prev,''), cs.payload, cs.is_current, cs.created_at
		FROM calendar_states cs
		JOIN calendar_owners co ON co.calendar_id = cs.id
		WHERE co.did = ? AND cs.is_current = 1
		ORDER BY cs.created_at DESC`, did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCalendarStates(rows)
}

func (d *DB) SubscribeCalendar(ctx context.Context, calendarID, did string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO calendar_subscriptions (calendar_id, did, subscribed_at) VALUES (?, ?, ?)`,
			calendarID, did, time.Now().UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) UnsubscribeCalendar(ctx context.Context, calendarID, did string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`DELETE FROM calendar_subscriptions WHERE calendar_id = ? AND did = ?`,
			calendarID, did,
		)
		return err
	})
}

func (d *DB) IsSubscribed(ctx context.Context, calendarID, did string) (bool, error) {
	var count int
	err := d.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM calendar_subscriptions WHERE calendar_id = ? AND did = ?`,
		calendarID, did,
	).Scan(&count)
	return count > 0, err
}

func (d *DB) ListSubscribedCalendars(ctx context.Context, did string) ([]*CalendarState, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT cs.hash, cs.id, COALESCE(cs.prev,''), cs.payload, cs.is_current, cs.created_at
		FROM calendar_states cs
		JOIN calendar_subscriptions sub ON sub.calendar_id = cs.id
		WHERE sub.did = ? AND cs.is_current = 1
		ORDER BY sub.subscribed_at DESC`, did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCalendarStates(rows)
}

func (d *DB) CountCalendarSubscribers(ctx context.Context, calendarID string) (int, error) {
	var count int
	err := d.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM calendar_subscriptions WHERE calendar_id = ?`, calendarID,
	).Scan(&count)
	return count, err
}

func (d *DB) ListEventsByCalendar(ctx context.Context, calendarID string) ([]*EventState, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT es.hash, es.id, COALESCE(es.prev,''), es.payload, es.is_current, es.created_at
		FROM event_states es
		JOIN event_founding ef ON ef.id = es.id
		WHERE json_extract(es.payload, '$.calendar') = ? AND es.is_current = 1
		ORDER BY json_extract(es.payload, '$.starts_at') ASC`, calendarID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*EventState
	for rows.Next() {
		s, err := scanEventStateRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func scanCalendarStates(rows *sql.Rows) ([]*CalendarState, error) {
	var out []*CalendarState
	for rows.Next() {
		s := &CalendarState{}
		var createdAt string
		if err := rows.Scan(&s.Hash, &s.ID, &s.Prev, &s.Payload, &s.IsCurrent, &createdAt); err != nil {
			return nil, err
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		out = append(out, s)
	}
	return out, rows.Err()
}

// ---- RSVP envelopes ----

type RSVPEnvelope struct {
	ID           string
	EventID      string
	SenderDID    string
	Payload      string // encrypted blob (JSON of private.Envelope)
	Status       string // pending | confirmed | rejected | waitlisted | cancelled
	GuestVisible bool
	ReceivedAt   time.Time
}

func (d *DB) InsertRSVPEnvelope(ctx context.Context, e *RSVPEnvelope) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO rsvp_envelopes (id, event_id, sender_did, payload, status, guest_visible, received_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			e.ID, e.EventID, e.SenderDID, e.Payload, e.Status,
			boolInt(e.GuestVisible),
			e.ReceivedAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) UpdateRSVPStatus(ctx context.Context, id, status string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE rsvp_envelopes SET status = ? WHERE id = ?`, status, id)
		return err
	})
}

func (d *DB) UpdateRSVPGuestVisible(ctx context.Context, id string, visible bool) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE rsvp_envelopes SET guest_visible = ? WHERE id = ?`, boolInt(visible), id)
		return err
	})
}

func (d *DB) ListRSVPsForEvent(ctx context.Context, eventID string) ([]*RSVPEnvelope, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT id, event_id, sender_did, payload, status, COALESCE(guest_visible, 1), received_at
		 FROM rsvp_envelopes WHERE event_id = ? ORDER BY received_at ASC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*RSVPEnvelope
	for rows.Next() {
		e := &RSVPEnvelope{}
		var ra string
		var guestVisible int
		if err := rows.Scan(&e.ID, &e.EventID, &e.SenderDID, &e.Payload, &e.Status, &guestVisible, &ra); err != nil {
			return nil, err
		}
		e.GuestVisible = guestVisible != 0
		e.ReceivedAt, _ = time.Parse(time.RFC3339, ra)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (d *DB) GetRSVPForEventSender(ctx context.Context, eventID, senderDID string) (*RSVPEnvelope, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT id, event_id, sender_did, payload, status, COALESCE(guest_visible, 1), received_at
		 FROM rsvp_envelopes
		 WHERE event_id = ? AND sender_did = ?
		 ORDER BY received_at DESC
		 LIMIT 1`, eventID, senderDID)
	e := &RSVPEnvelope{}
	var receivedAt string
	var guestVisible int
	if err := row.Scan(&e.ID, &e.EventID, &e.SenderDID, &e.Payload, &e.Status, &guestVisible, &receivedAt); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	e.GuestVisible = guestVisible != 0
	e.ReceivedAt, _ = time.Parse(time.RFC3339, receivedAt)
	return e, nil
}

// ---- Mirrored RSVP receipts ----

type RSVPReceipt struct {
	EventInstanceURL string
	EventID          string
	DID              string
	Status           string
	GuestVisible     bool
	EventURL         string
	EventTitle       string
	EventStartsAt    string
	IssuedAt         time.Time
	UpdatedAt        time.Time
	Sig              string
}

func (d *DB) UpsertRSVPReceipt(ctx context.Context, r *RSVPReceipt) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO rsvp_receipts (
				event_instance_url, event_id, did, status, guest_visible,
				event_url, event_title, event_starts_at, issued_at, updated_at, sig
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(event_instance_url, event_id, did) DO UPDATE SET
				status = excluded.status,
				guest_visible = excluded.guest_visible,
				event_url = excluded.event_url,
				event_title = excluded.event_title,
				event_starts_at = excluded.event_starts_at,
				issued_at = excluded.issued_at,
				updated_at = excluded.updated_at,
				sig = excluded.sig
			WHERE excluded.updated_at >= rsvp_receipts.updated_at`,
			r.EventInstanceURL, r.EventID, r.DID, r.Status, boolInt(r.GuestVisible),
			nullString(r.EventURL), nullString(r.EventTitle), nullString(r.EventStartsAt),
			r.IssuedAt.UTC().Format(time.RFC3339),
			r.UpdatedAt.UTC().Format(time.RFC3339),
			r.Sig,
		)
		return err
	})
}

func (d *DB) ListRSVPReceiptsForDID(ctx context.Context, did string) ([]*RSVPReceipt, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT event_instance_url, event_id, did, status, guest_visible,
		        COALESCE(event_url, ''), COALESCE(event_title, ''), COALESCE(event_starts_at, ''),
		        issued_at, updated_at, sig
		 FROM rsvp_receipts
		 WHERE did = ?
		 ORDER BY updated_at DESC`, did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*RSVPReceipt
	for rows.Next() {
		r := &RSVPReceipt{}
		var guestVisible int
		var issuedAt, updatedAt string
		if err := rows.Scan(
			&r.EventInstanceURL, &r.EventID, &r.DID, &r.Status, &guestVisible,
			&r.EventURL, &r.EventTitle, &r.EventStartsAt,
			&issuedAt, &updatedAt, &r.Sig,
		); err != nil {
			return nil, err
		}
		r.GuestVisible = guestVisible != 0
		r.IssuedAt, _ = time.Parse(time.RFC3339, issuedAt)
		r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---- Sessions ----

type Session struct {
	Token     string
	DID       string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (d *DB) InsertSession(ctx context.Context, s *Session) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO sessions (token, did, created_at, expires_at) VALUES (?, ?, ?, ?)`,
			s.Token, s.DID,
			s.CreatedAt.UTC().Format(time.RFC3339),
			s.ExpiresAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) GetSession(ctx context.Context, token string) (*Session, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT token, did, created_at, expires_at FROM sessions WHERE token = ?`, token)
	s := &Session{}
	var ca, ea string
	err := row.Scan(&s.Token, &s.DID, &ca, &ea)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, ca)
	s.ExpiresAt, _ = time.Parse(time.RFC3339, ea)
	return s, nil
}

func (d *DB) DeleteSession(ctx context.Context, token string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`DELETE FROM sessions WHERE token = ?`, token)
		return err
	})
}

// ---- Magic links ----

type MagicLink struct {
	Token        string
	PollToken    string
	SessionToken string
	Email        string
	DID          string // empty if new user
	ExpiresAt    time.Time
	ApprovedAt   time.Time
	Used         bool
}

func (d *DB) InsertMagicLink(ctx context.Context, ml *MagicLink) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO magic_links (token, poll_token, email, did, expires_at, used) VALUES (?, ?, ?, ?, ?, 0)`,
			ml.Token, ml.PollToken, ml.Email, nullString(ml.DID),
			ml.ExpiresAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) GetMagicLink(ctx context.Context, token string) (*MagicLink, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT token, COALESCE(poll_token,''), COALESCE(session_token,''), email, COALESCE(did,''), expires_at, COALESCE(approved_at,''), used
		 FROM magic_links WHERE token = ?`, token)
	return scanMagicLink(row)
}

func (d *DB) GetMagicLinkByPollToken(ctx context.Context, pollToken string) (*MagicLink, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT token, COALESCE(poll_token,''), COALESCE(session_token,''), email, COALESCE(did,''), expires_at, COALESCE(approved_at,''), used
		 FROM magic_links WHERE poll_token = ?`, pollToken)
	return scanMagicLink(row)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanMagicLink(row rowScanner) (*MagicLink, error) {
	ml := &MagicLink{}
	var ea, approvedAt string
	var used int
	err := row.Scan(&ml.Token, &ml.PollToken, &ml.SessionToken, &ml.Email, &ml.DID, &ea, &approvedAt, &used)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ml.ExpiresAt, _ = time.Parse(time.RFC3339, ea)
	if approvedAt != "" {
		ml.ApprovedAt, _ = time.Parse(time.RFC3339, approvedAt)
	}
	ml.Used = used != 0
	return ml, nil
}

func (d *DB) MarkMagicLinkUsed(ctx context.Context, token string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE magic_links SET used = 1, approved_at = ? WHERE token = ?`,
			time.Now().UTC().Format(time.RFC3339), token)
		return err
	})
}

func (d *DB) SetMagicLinkSession(ctx context.Context, pollToken, sessionToken string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE magic_links SET session_token = ? WHERE poll_token = ?`, sessionToken, pollToken)
		return err
	})
}

// ---- Passkeys (WebAuthn) ----

type Passkey struct {
	ID              []byte
	DID             string
	PublicKey       []byte
	AttestationType string
	Transport       string // JSON array
	Counter         int64
	UserPresent     bool
	UserVerified    bool
	BackupEligible  bool
	BackupState     bool
	CreatedAt       time.Time
}

func (d *DB) InsertPasskey(ctx context.Context, p *Passkey) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO passkeys (id, did, public_key, attestation_type, transport, counter, user_present, user_verified, backup_eligible, backup_state, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			p.ID, p.DID, p.PublicKey, p.AttestationType, p.Transport, p.Counter,
			boolInt(p.UserPresent), boolInt(p.UserVerified), boolInt(p.BackupEligible), boolInt(p.BackupState), p.CreatedAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) GetPasskeysByDID(ctx context.Context, did string) ([]*Passkey, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT id, did, public_key, attestation_type, COALESCE(transport,''), counter, user_present, user_verified, backup_eligible, backup_state, created_at
		 FROM passkeys WHERE did = ?`, did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Passkey
	for rows.Next() {
		p := &Passkey{}
		var ca string
		var up, uv, be, bs int
		if err := rows.Scan(&p.ID, &p.DID, &p.PublicKey, &p.AttestationType, &p.Transport, &p.Counter, &up, &uv, &be, &bs, &ca); err != nil {
			return nil, err
		}
		p.UserPresent = up != 0
		p.UserVerified = uv != 0
		p.BackupEligible = be != 0
		p.BackupState = bs != 0
		p.CreatedAt, _ = time.Parse(time.RFC3339, ca)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (d *DB) GetPasskeyByID(ctx context.Context, id []byte) (*Passkey, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT id, did, public_key, attestation_type, COALESCE(transport,''), counter, user_present, user_verified, backup_eligible, backup_state, created_at
		 FROM passkeys WHERE id = ?`, id)
	p := &Passkey{}
	var ca string
	var up, uv, be, bs int
	err := row.Scan(&p.ID, &p.DID, &p.PublicKey, &p.AttestationType, &p.Transport, &p.Counter, &up, &uv, &be, &bs, &ca)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.UserPresent = up != 0
	p.UserVerified = uv != 0
	p.BackupEligible = be != 0
	p.BackupState = bs != 0
	p.CreatedAt, _ = time.Parse(time.RFC3339, ca)
	return p, nil
}

func (d *DB) UpdatePasskeyCounter(ctx context.Context, id []byte, counter int64) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`UPDATE passkeys SET counter = ? WHERE id = ?`, counter, id)
		return err
	})
}

func (d *DB) UpdatePasskeyBackupFlags(ctx context.Context, id []byte, eligible, state bool) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`UPDATE passkeys SET backup_eligible = ?, backup_state = ? WHERE id = ?`,
			boolInt(eligible), boolInt(state), id,
		)
		return err
	})
}

// ---- WebAuthn Sessions ----

type WebAuthnSession struct {
	Token     string
	DID       string
	Data      []byte
	ExpiresAt time.Time
}

func (d *DB) InsertWebAuthnSession(ctx context.Context, s *WebAuthnSession) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO webauthn_sessions (token, did, data, expires_at) VALUES (?, ?, ?, ?)`,
			s.Token, s.DID, s.Data, s.ExpiresAt.UTC().Format(time.RFC3339),
		)
		return err
	})
}

func (d *DB) GetWebAuthnSession(ctx context.Context, token string) (*WebAuthnSession, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT token, did, data, expires_at FROM webauthn_sessions WHERE token = ?`, token)
	s := &WebAuthnSession{}
	var ea string
	err := row.Scan(&s.Token, &s.DID, &s.Data, &ea)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.ExpiresAt, _ = time.Parse(time.RFC3339, ea)
	return s, nil
}

func (d *DB) DeleteWebAuthnSession(ctx context.Context, token string) error {
	return d.Write(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec(`DELETE FROM webauthn_sessions WHERE token = ?`, token)
		return err
	})
}

// ---- helpers ----

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func scanKELOps(rows *sql.Rows) ([]*KELOp, error) {
	var ops []*KELOp
	for rows.Next() {
		op := &KELOp{}
		var ca string
		if err := rows.Scan(&op.Hash, &op.DID, &op.Type, &op.Payload, &op.Prev, &op.Seq, &ca); err != nil {
			return nil, err
		}
		op.CreatedAt, _ = time.Parse(time.RFC3339, ca)
		ops = append(ops, op)
	}
	return ops, rows.Err()
}

func scanKELOp(row *sql.Row) (*KELOp, error) {
	op := &KELOp{}
	var ca string
	err := row.Scan(&op.Hash, &op.DID, &op.Type, &op.Payload, &op.Prev, &op.Seq, &ca)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	op.CreatedAt, _ = time.Parse(time.RFC3339, ca)
	return op, nil
}

func scanEventState(row *sql.Row) (*EventState, error) {
	s := &EventState{}
	var ca string
	var isCurrent int
	err := row.Scan(&s.Hash, &s.ID, &s.Prev, &s.Payload, &isCurrent, &ca)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan event state: %w", err)
	}
	s.IsCurrent = isCurrent != 0
	s.CreatedAt, _ = time.Parse(time.RFC3339, ca)
	return s, nil
}

func scanEventStateRow(rows *sql.Rows) (*EventState, error) {
	s := &EventState{}
	var ca string
	var isCurrent int
	err := rows.Scan(&s.Hash, &s.ID, &s.Prev, &s.Payload, &isCurrent, &ca)
	if err != nil {
		return nil, err
	}
	s.IsCurrent = isCurrent != 0
	s.CreatedAt, _ = time.Parse(time.RFC3339, ca)
	return s, nil
}
