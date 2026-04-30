-- Calendar ownership index — which calendars a DID owns
CREATE TABLE IF NOT EXISTS calendar_owners (
    calendar_id TEXT NOT NULL REFERENCES calendar_founding(id),
    did         TEXT NOT NULL,
    PRIMARY KEY (calendar_id, did)
);
CREATE INDEX IF NOT EXISTS idx_calendar_owners_did ON calendar_owners(did);

-- Calendar subscriptions — which DIDs follow which calendars
CREATE TABLE IF NOT EXISTS calendar_subscriptions (
    calendar_id  TEXT NOT NULL REFERENCES calendar_founding(id),
    did          TEXT NOT NULL,
    subscribed_at TEXT NOT NULL,
    PRIMARY KEY (calendar_id, did)
);
CREATE INDEX IF NOT EXISTS idx_calendar_subscriptions_did ON calendar_subscriptions(did);
