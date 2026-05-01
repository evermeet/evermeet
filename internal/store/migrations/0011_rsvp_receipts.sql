CREATE TABLE rsvp_receipts (
    event_instance_url TEXT NOT NULL,
    event_id           TEXT NOT NULL,
    did                TEXT NOT NULL,
    status             TEXT NOT NULL,
    guest_visible      INTEGER NOT NULL DEFAULT 1,
    event_url          TEXT,
    event_title        TEXT,
    event_starts_at    TEXT,
    issued_at          TEXT NOT NULL,
    updated_at         TEXT NOT NULL,
    sig                TEXT NOT NULL,
    PRIMARY KEY (event_instance_url, event_id, did)
);
CREATE INDEX idx_rsvp_receipts_did_updated ON rsvp_receipts(did, updated_at DESC);
