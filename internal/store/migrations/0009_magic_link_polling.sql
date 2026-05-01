-- Magic links approve a login request started in another browser tab.
ALTER TABLE magic_links ADD COLUMN poll_token TEXT NOT NULL DEFAULT '';
ALTER TABLE magic_links ADD COLUMN session_token TEXT NOT NULL DEFAULT '';
ALTER TABLE magic_links ADD COLUMN approved_at TEXT;

CREATE UNIQUE INDEX idx_magic_links_poll_token ON magic_links(poll_token)
WHERE poll_token != '';
