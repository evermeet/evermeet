-- Add backup flags to passkeys
ALTER TABLE passkeys ADD COLUMN backup_eligible INTEGER NOT NULL DEFAULT 0;
ALTER TABLE passkeys ADD COLUMN backup_state INTEGER NOT NULL DEFAULT 0;
