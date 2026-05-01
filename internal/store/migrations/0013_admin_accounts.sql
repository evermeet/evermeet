CREATE TABLE admin_accounts (
    did        TEXT PRIMARY KEY REFERENCES users(did) ON DELETE CASCADE,
    created_at TEXT NOT NULL
);
