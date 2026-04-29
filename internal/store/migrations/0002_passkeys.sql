-- WebAuthn Credentials
CREATE TABLE passkeys (
    id              BLOB PRIMARY KEY, -- Credential ID
    did             TEXT NOT NULL REFERENCES users(did),
    public_key      BLOB NOT NULL,
    attestation_type TEXT NOT NULL,
    transport       TEXT, -- JSON array of strings
    counter         INTEGER NOT NULL DEFAULT 0,
    user_present    INTEGER NOT NULL DEFAULT 1,
    user_verified   INTEGER NOT NULL DEFAULT 0,
    backup_eligible INTEGER NOT NULL DEFAULT 0,
    backup_state    INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL
);
CREATE INDEX idx_passkeys_did ON passkeys(did);

-- Temporary storage for WebAuthn session data (between start and finish)
CREATE TABLE webauthn_sessions (
    token       TEXT PRIMARY KEY, -- The session ID
    did         TEXT NOT NULL,
    data        BLOB NOT NULL, -- Serialized webauthn.SessionData
    expires_at  TEXT NOT NULL
);
