-- Ethereum wallets are backup login methods for Evermeet did:em identities.
ALTER TABLE user_private ADD COLUMN siwe_chain_id TEXT NOT NULL DEFAULT '';
ALTER TABLE user_private ADD COLUMN siwe_address TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX idx_user_private_siwe ON user_private(siwe_chain_id, siwe_address)
WHERE siwe_chain_id != '' AND siwe_address != '';
