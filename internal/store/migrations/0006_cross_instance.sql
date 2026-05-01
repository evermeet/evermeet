-- Nonces issued to foreign instances during cross-instance login.
-- The nonce is embedded in the delegation token to prevent replay across instances.
CREATE TABLE cross_instance_nonces (
    nonce       TEXT PRIMARY KEY,
    foreign_url TEXT NOT NULL,   -- the foreign instance that requested this nonce
    event_id    TEXT NOT NULL DEFAULT '',  -- optional: event the user was trying to RSVP to
    created_at  TEXT NOT NULL,
    expires_at  TEXT NOT NULL,
    used        INTEGER NOT NULL DEFAULT 0
);
