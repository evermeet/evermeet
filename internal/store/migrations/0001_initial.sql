-- Key event log operations (gossipsub topic: em/dids)
CREATE TABLE kel_ops (
    hash        TEXT PRIMARY KEY,
    did         TEXT NOT NULL,
    type        TEXT NOT NULL CHECK(type IN ('genesis','rotate','migrate')),
    payload     TEXT NOT NULL,
    prev        TEXT,
    seq         INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL
);
CREATE INDEX idx_kel_ops_did ON kel_ops(did, seq);

-- User public records (gossipsub topic: em/profiles)
CREATE TABLE users (
    did          TEXT PRIMARY KEY,
    display_name TEXT,
    avatar       TEXT,
    bio          TEXT,
    current_pk   TEXT NOT NULL,
    rotation_pk  TEXT NOT NULL,
    endpoint     TEXT,
    sig          TEXT NOT NULL,
    updated_at   TEXT NOT NULL
);

-- Private user data — never replicated, local instance only
CREATE TABLE user_private (
    did             TEXT PRIMARY KEY REFERENCES users(did),
    email           TEXT,
    email_verified  INTEGER NOT NULL DEFAULT 0,
    google_sub      TEXT,
    passkey_ids     TEXT,
    enc_signing_key TEXT
);

-- Calendar founding documents — immutable after creation
CREATE TABLE calendar_founding (
    id      TEXT PRIMARY KEY,
    payload TEXT NOT NULL
);

-- Calendar mutable state — prev-hash revision chain
CREATE TABLE calendar_states (
    hash       TEXT PRIMARY KEY,
    id         TEXT NOT NULL REFERENCES calendar_founding(id),
    prev       TEXT,
    payload    TEXT NOT NULL,
    is_current INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);
CREATE INDEX idx_calendar_states_id ON calendar_states(id, is_current);

-- Event founding documents — immutable after creation
CREATE TABLE event_founding (
    id      TEXT PRIMARY KEY,
    payload TEXT NOT NULL
);

-- Event mutable state — prev-hash revision chain
CREATE TABLE event_states (
    hash       TEXT PRIMARY KEY,
    id         TEXT NOT NULL REFERENCES event_founding(id),
    prev       TEXT,
    payload    TEXT NOT NULL,
    is_current INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);
CREATE INDEX idx_event_states_id ON event_states(id, is_current);

-- RSVP envelopes — private, never gossiped
CREATE TABLE rsvp_envelopes (
    id          TEXT PRIMARY KEY,
    event_id    TEXT NOT NULL,
    sender_did  TEXT NOT NULL,
    payload     TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    received_at TEXT NOT NULL
);
CREATE INDEX idx_rsvp_envelopes_event ON rsvp_envelopes(event_id);

-- Multi-sig governance proposals (threshold > 1)
CREATE TABLE proposals (
    hash        TEXT PRIMARY KEY,
    target_id   TEXT NOT NULL,
    target_type TEXT NOT NULL CHECK(target_type IN ('calendar','event')),
    payload     TEXT NOT NULL,
    sigs        TEXT NOT NULL,
    prev        TEXT NOT NULL,
    created_at  TEXT NOT NULL
);

-- Magic link tokens for email login
CREATE TABLE magic_links (
    token      TEXT PRIMARY KEY,
    email      TEXT NOT NULL,
    did        TEXT,
    expires_at TEXT NOT NULL,
    used       INTEGER NOT NULL DEFAULT 0
);

-- HTTP session tokens
CREATE TABLE sessions (
    token      TEXT PRIMARY KEY,
    did        TEXT NOT NULL,
    created_at TEXT NOT NULL,
    expires_at TEXT NOT NULL
);
CREATE INDEX idx_sessions_did ON sessions(did);
