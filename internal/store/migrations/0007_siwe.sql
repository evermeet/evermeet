-- Short-lived nonces for Sign-In with Ethereum.
CREATE TABLE siwe_nonces (
    nonce      TEXT PRIMARY KEY,
    address    TEXT NOT NULL,
    chain_id   TEXT NOT NULL,
    created_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    used       INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_siwe_nonces_address ON siwe_nonces(address, used);
