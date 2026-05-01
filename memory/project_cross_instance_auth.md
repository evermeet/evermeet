---
name: Cross-instance auth implementation plan
description: DHT blind routing plan for cross-instance login/RSVP — decisions, architecture, and open questions resolved
type: project
---

# Cross-Instance Auth via DHT Blind Routing

**Why:** User on foreign instance (events.devcon.org) wants to RSVP. Foreign instance must discover user's home instance without a central DB and without leaking plaintext emails.

**Core mechanism:** Kademlia DHT. Email hash (blake3 + domain separator) as key, signed home instance URL as value.

## Resolved decisions

- **Bootstrap:** mDNS-only to start. No hardcoded bootstrap peers yet.
- **return_to validation:** Foreign instance signs the return_to URL with its instance Ed25519 key. Home instance verifies before issuing delegation.
- **Token format:** Simple JSON envelope signed with Ed25519. NOT JWT. Consistent with rest of codebase.
- **Email hash:** `blake3("evermeet-home-routing-v1:" + email)` hex-encoded. Public domain separator, not a secret.
- **DHT record value:** `{homeInstanceURL, timestamp}` signed by the **instance** Ed25519 key (not user key).
- **Delegation token signed by:** the **user's** Ed25519 key (unlocked during local auth on home instance).

## Implementation phases

### Phase 1 — DHT Infrastructure
- 1a: Persistent libp2p identity → save/load p2p.key in dataDir
- 1b: Wire Kademlia DHT in internal/node/node.go (go-libp2p-kad-dht)
- 1c: DHTPublish(key, value) and DHTPeerLookup(key) methods on Node

### Phase 2 — Email Hash & Registration
- 2a: Domain separator constant in a new internal/routing package
- 2b: Publish email→homeInstance mapping to DHT on user registration (in lookupOrCreateUser)
- 2c: Background heartbeat re-publishes all users every 12h (DHT TTL ~24h)
- 2d: Implement /.well-known/evermeet-node-key (currently 501) — returns instance public key

### Phase 3 — Foreign Instance Lookup
- 3a: POST /api/auth/resolve-home — takes email, returns homeInstanceURL
- 3b: Verify DHT record signature against instance public key from /.well-known/evermeet-node-key
- 3c: Foreign instance issues nonce (32 bytes, 5min TTL, stored in DB)

### Phase 4 — Redirect & Delegation Token
- 4a: Frontend on foreign instance: email input → resolve-home → show discovered URL → redirect
- 4b: New page /auth/delegate on home instance
- 4c: Delegation token schema:
  { sub: did, aud: foreign_instance_url, iat, exp, nonce, event_id }
  Signed with user Ed25519 private key
- 4d: Foreign instance signs return_to URL with its instance key, included in redirect params
- 4e: POST /api/auth/delegate-verify on foreign instance — verifies token+sig, creates 24h session

### Phase 5 — RSVP Integration
- No RSVP endpoint changes needed once delegation session exists
- /internal/rsvp/ is currently empty — wire up later

## Key files
- internal/node/node.go — add DHT, persistent identity
- internal/node/handlers.go — existing P2P RPC patterns to follow
- internal/api/auth.go — lookupOrCreateUser, session creation
- internal/api/router.go — add new endpoints, fix /.well-known/evermeet-node-key
- internal/store/queries.go — add nonce table, delegation session queries
- User.Endpoint field already exists in DB — stores home instance URL

## Security notes
- Fake home instance cannot forge delegation token (signed by user's Ed25519 key)
- Open redirect prevented by foreign instance signing return_to
- DHT record poisoning mitigated by instance key signature verification
- Nonce in delegation token prevents replay across instances
