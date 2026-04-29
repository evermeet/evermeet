# Evermeet

Platform and P2P protocol for hosting and publishing events like meetups, conferences or hackathons.

The goal is to create a viable alternative to [Lu.ma](https://lu.ma), [Meetup.com](https://meetup.com), and other similar centralized services and put sovereignty back in the hands of the people — without sacrificing the UX people are used to.

## Features

- **Self-sovereign identities** — every user has an Ed25519 keypair; login with passkey, Google, or email without giving up ownership
- **Events propagated through P2P layer** — no central server, no single point of failure
- **Single self-contained binary** with both node and usable web UI — download, run, done
- **Censorship-resistant** — no platform can deplatform your event, ever
- **Cryptographically signed event records** — every event is tamper-proof and verifiable by anyone
- **Private by default** — email never published, RSVP lists encrypted; only the organizer sees who is coming
- **Federated discovery** — find events across all nodes without a central registry
- **Open protocol** — any client can implement it; the ecosystem belongs to no one

---

## Identity

Every user has a stable, self-certifying identifier derived from their genesis keypair — inspired by KERI. The identifier never changes even if the signing key is rotated or the user migrates to a different instance. No external registry, no blockchain, no central authority.

```
did:em:{base32(blake3(initial_public_key))}
  → immutable forever, self-certifying
```

### Key event log

Identity state is maintained as an append-only signed operation log, replicated across the P2P network on the `em/dids` gossipsub topic. Any node can replay the log to get current state.

```json
// genesis — creates the identity
{ "type": "genesis",  "pk": "pk:z6Mk...f3a", "rotation_pk": "pk:r9Qp...8de",
  "endpoint": "https://evermeet.app", "sig": "..." }
// did:em:2abc...9xyz  ←  derived as base32(blake3(pk)), never stored in the op itself

// rotate — new signing key, signed by rotation_pk
{ "type": "rotate",   "new_pk": "pk:m2Jx...1bc", "rotation_pk": "pk:r9Qp...8de",
  "prev": "<hash_of_prev_op>", "sig": "..." }

// migrate — move to new instance, signed by current signing key
{ "type": "migrate",  "endpoint": "https://home.tree.cz",
  "prev": "<hash_of_prev_op>", "sig": "..." }
```

Each operation references the previous one by hash — making the log tamper-evident. A rotation signed by anything other than the current rotation key is rejected by all nodes.

### Identity survival

Your identity survives your instance disappearing if any of these hold:

- **You hold your keys** — if you have the signing and rotation keys exported (passkey, file, password manager), your identity exists regardless of any server. Spin up a new instance, publish a `migrate` operation, done.
- **Any node cached your log** — every node that ever resolved your identity stored your full key event log. As long as one peer has seen you, your identity can propagate again.
- **You seeded your log elsewhere** — you can push your key event log to any other instance as a backup. They can't impersonate you (no private key) but keep your identity resolvable.

The real risk is **key loss, not instance loss**. If you lose both signing and rotation keys, recovery is impossible — there is no "forgot password." This is the honest tradeoff of self-sovereignty.

For custodial users (email/OAuth login), the instance holds encrypted keys. If the instance disappears before they export their keys, their identity is at risk. The app actively mitigates this by encouraging key export and seeding the identity log to at least one other known node. The path from custodial to sovereign is one click and actively surfaced in the UI.

### Login methods

The protocol only knows Ed25519. Login methods are just different ways to unlock a keypair — the network always sees the same identity regardless of how you authenticated.

| Method | Key lives where | Sovereignty |
|--------|-----------------|-------------|
| **Passkey** | Device secure enclave | Full — instance never sees the key |
| **SIWE** | Ethereum wallet | Full — instance never sees the key |
| **Nostr keypair** | Nostr client / NIP-07 extension | Full — instance never sees the key |
| **Magic link / OAuth** | Instance (encrypted at rest) | Custodial by default, upgradeable |

For passkey and SIWE users, identity is sovereign from day one. For email/Google users, the instance holds the encrypted key until they export it or add a passkey. Either way, every signature on the protocol is Ed25519.

### Invitations and cross-instance identity

When an organizer invites someone by email the app handles three cases:

**1. Known user on this instance** — notification sent, invitation linked to existing identity.

**2. User on a different instance** — invitation email offers two options: sign in with existing Evermeet account to claim it, or create a new account here. RSVP ends up signed by their own key regardless.

**3. Unknown email** — magic link sent, new keypair created silently on first click. They have a real identity from that moment and can add a passkey later to take full custody.

---

## Home instance

Pure P2P is the ideal. In practice, some things are impossible to do purely peer-to-peer — email delivery requires an SMTP server, push notifications require an always-online process, key custody for non-technical users requires somewhere to store an encrypted keypair. The home instance exists to handle exactly these responsibilities, and nothing more.

A home instance is an Evermeet node that a user designates to act on their behalf. It is not their identity — it holds no special authority over their `did:em:` identifier. It is a delegate: the user grants it specific permissions, and can revoke or reassign them at any time.

### What a home instance does

```
Delegated responsibility          Why P2P alone can't do it
─────────────────────────────     ──────────────────────────────────────────
Send transactional email          SMTP requires an always-on server with a domain
Receive email invitations         Someone needs to accept inbound mail for the user
Key custody (optional)            Non-technical users need somewhere to store keys
Publish key event log updates     Node must be online to propagate did:em: changes
Push notifications                Requires persistent connection on behalf of user
Store encrypted RSVPs             Organizer needs always-on inbox for incoming RSVPs
```

### Progressive custody

The home instance is designed around a spectrum — users start wherever they are comfortable and move toward full sovereignty at their own pace:

```
Level 1 — Hosted instance (e.g. evermeet.app)
  · instance holds encrypted keypair
  · user logs in with email / Google / passkey
  · instance sends email, manages key event log, stores RSVPs
  · same UX as Lu.ma — user never thinks about keys

Level 2 — Sovereign keys, hosted instance
  · user holds keys via passkey (device secure enclave), exported key file,
    Ethereum wallet (SIWE), or Nostr keypair (NIP-07) — instance no longer holds the signing key
  · instance still handles email, notifications, and key event log publishing
  · identity is fully sovereign, services are delegated

Level 3 — Self-hosted instance
  · user runs their own Evermeet node
  · full control over all delegated responsibilities
  · instance can be a cheap VPS, a home server, or a friend's node
  · no dependency on any third party
```

At every level, the `did:em:` identifier is the same. Event history, RSVPs, Calendar memberships — all portable. Moving between levels is a migration, not a restart.

### What the home instance cannot do

The home instance never has authority over the user's identity on the protocol. It cannot:

- forge signatures on behalf of the user (it doesn't hold the signing key at Level 2+)
- prevent the user from migrating to a different instance
- revoke or modify the user's `did:em:` identifier
- access encrypted RSVP payloads the user received as an attendee

This is the key difference from a centralized platform: the instance is infrastructure the user controls, not a platform that controls the user.

---

## Primitives

Evermeet has three first-class record types. Each has a stable, deterministic ID derived from its founding document — similar to how Radicle derives repository IDs. No registry, no server, no coordination needed to create one.

```
User        →  did:em:{base32(blake3(initial_public_key))}           — self-sovereign identity, rotation-safe
Calendar    →  base32(blake3(canonical_json(founding_doc)))          — content-addressed, derived from initial owner set
Event       →  base32(blake3(canonical_json(founding_doc)))          — content-addressed, derived from organizer + nonce
```

All three IDs are derived from genesis state and permanently fixed at creation. Mutable state (profile, calendar details, event content) is layered on top as signed updates. An Event's title or date can change — its ID never does.

### Relationships

```
User ──────creates──────────────────────► Event
User ──────owns / co-owns───────────────► Calendar
Calendar ──publishes────────────────────► Event
User ──────subscribes───────────────────► Calendar
User ──────RSVPs────────────────────────► Event
Event ─────belongs to (optional)────────► Calendar
```

An Event can exist without a Calendar — one-off events are first-class. A Calendar can have multiple co-owners — communities outlast individuals. Both are reflected in the schemas below.

---

### Revisions and governance

Every mutable state update — for both Calendars and Events — extends a tamper-evident revision chain. Each update references the previous state by hash, forming an auditable history similar to Git commits or plcbundle's PLC operation log.

```
founding_doc
     │
     ▼
state_v1  { ..., prev: hash(founding_doc), sigs: [...] }
     │
     ▼
state_v2  { ..., prev: hash(state_v1),     sigs: [...] }
     │
     ▼
state_v3  { ..., prev: hash(state_v2),     sigs: [...] }
```

Any node can verify the full history by replaying the chain. A broken hash means tampering. The latest valid state is canonical.

**Governance threshold**

Each Calendar and Event defines a threshold policy in its mutable state:

```json
"governance": {
  "threshold": 2,
  "owners": [
    { "did": "did:em:2abc...9xyz", "role": "owner" },
    { "did": "did:em:5fgh...1klm", "role": "owner" },
    { "did": "did:em:9pqr...4stu", "role": "editor" }
  ]
}
```

`threshold: 1` means any single owner can push an update immediately. `threshold: 2` means two owner signatures are required before an update is accepted by the network. Any value up to the number of owners is valid.

**Proposals**

When a proposed update has not yet collected enough signatures to meet the threshold, it exists as a pending proposal on the network:

```json
{
  "type":   "proposal",
  "target": "4zxy...8qpr",
  "change": { ...new mutable state... },
  "prev":   "hash_of_current_state",
  "sigs": [
    { "did": "did:em:2abc...9xyz", "sig": "3045...ab12" }
  ]
}
```

Other owners cosign the proposal until the threshold is met. Once threshold is reached the proposal is gossiped as the new canonical state with all collected signatures. The full signing timeline is preserved in the chain — auditable by anyone.

Threshold changes themselves must meet the current threshold to apply.

---

### User

A User is a person on the network. Their Ed25519 public key is their identity. The public record contains only what is safe to be world-readable — display name, avatar, bio. **Email is private and never published.**

```json
// public identity record — world readable, replicated on the P2P network
{
  "type":         "user",
  "did":          "did:em:2abc...9xyz",
  "display_name": "Tree",
  "avatar":       "https://...",
  "bio":          "Building sovereign infrastructure.",
  "created_at":   "2025-01-01T00:00:00Z",
  "sig":          "3045...ab12"
}

// private record — stored only on the user's instance, never published
{
  "did":            "did:em:2abc...9xyz",
  "email":          "tree@example.com",
  "email_verified": true,
  "auth":           { "google_sub": "...", "passkey_id": "..." }
}
```

Email lives only on the instance, encrypted at rest. It is used for SMTP notifications and nothing else — no other node ever sees it. A User with no email is a valid identity on the protocol.

---

### Calendar

A Calendar is a persistent identity for an organizer or community. Its ID is derived deterministically from a minimal founding document — the initial owner set, a timestamp, and a nonce. The founding document is hashed once at creation and never changes; all subsequent updates (adding co-owners, changing name/description) are mutable state layered on top of the stable ID.

```json
// founding document — hashed to produce the Calendar ID, immutable
{
  "type":       "calendar",
  "owners":     ["did:em:2abc...9xyz"],
  "created_at": "2025-01-01T00:00:00Z",
  "nonce":      "a3f7...9c2b"
}

// id = base32(blake3(canonical_json(founding_doc)))
// e.g. = "4zxy...8qpr"

// mutable state — signed update, chained by prev hash
{
  "id":          "4zxy...8qpr",
  "prev":        "hash_of_previous_state",
  "name":        "Paralelní Polis Events",
  "description": "Events hosted at or affiliated with Paralelní Polis Prague.",
  "avatar":      "https://...",
  "website":     "https://paralelnipolis.cz",
  "governance": {
    "threshold": 1,
    "owners": [
      { "did": "did:em:2abc...9xyz", "role": "owner" },
      { "did": "did:em:5fgh...1klm", "role": "editor" }
    ]
  },
  "updated_at":  "2025-03-15T00:00:00Z",
  "sigs": [
    { "did": "did:em:2abc...9xyz", "sig": "3045...cd34" }
  ]
}
```

The `nonce` allows the same owner to create multiple distinct Calendars. The `governance` block defines who can update the Calendar and how many signatures are required. Updates form a `prev`-linked chain — full revision history is auditable by any node.

---

### Event

An Event is a signed record created by a User or a Calendar. Like Calendar, its ID is derived from a minimal founding document — the organizer, optional calendar reference, timestamp, and nonce. Mutable content (title, description, dates, location) is layered on top as a signed update.

```json
// founding document — hashed to produce the Event ID, immutable
{
  "type":       "event",
  "organizer":  "did:em:2abc...9xyz",
  "calendar":   "4zxy...8qpr",  // optional
  "created_at": "2025-04-01T00:00:00Z",
  "nonce":      "9c2b...f1e4"
}

// id = base32(blake3(canonical_json(founding_doc)))
// e.g. = "7mnp...3rst"

// mutable state — signed update, chained by prev hash
{
  "id":          "7mnp...3rst",
  "prev":        "hash_of_previous_state",
  "organizer":   "did:em:2abc...9xyz",
  "calendar":    "4zxy...8qpr",
  "title":       "Unfork Prague 2025",
  "description": "An anarchistic, non-commercial conference for decentralized infrastructure builders.",
  "starts_at":   "2025-06-14T10:00:00Z",
  "ends_at":     "2025-06-14T22:00:00Z",
  "location": {
    "name":    "Paralelní Polis",
    "address": "Dělnická 43, Prague",
    "lat":     50.09,
    "lon":     14.44
  },
  "governance": {
    "threshold": 1,
    "owners": [
      { "did": "did:em:2abc...9xyz", "role": "owner" }
    ]
  },
  "rsvp": {
    "limit":    100,
    "count":    0,
    "deadline": "2025-06-10T00:00:00Z",
    "approval": "auto"
  },
  "payment": {
    "hint":    "lightning",
    "address": "lnurl1..."
  },
  "visibility": "public",
  "tickets": [
    { "id": "general", "name": "General Admission", "price": 0, "currency": null, "limit": 100 }
  ],
  "program":    null,
  "tags":       ["hackathon", "decentralization", "prague"],
  "updated_at": "2025-04-01T00:00:00Z",
  "sigs": [
    { "did": "did:em:2abc...9xyz", "sig": "3045...ef56" }
  ]
}
```

The `calendar` field is optional — omit it for standalone events. The `organizer` is always the User who created the event, traceable even under a Calendar. The `governance` block defines who can update the event and the required threshold. Updates form a `prev`-linked revision chain — the ID never changes.

---

### RSVP

RSVP is the most sensitive data in the system. Who is attending a political meetup, a liberty-tech conference, or any sensitive gathering should never be a world-readable ledger — or even visible to uninvolved nodes on the network.

**RSVPs are never gossiped.** They are communicated bilaterally and directly between the attendee's instance and the organizer's instance (the instance where the event is homed). No relay node, no third-party instance, no network observer ever sees an RSVP.

```
attendee's instance  ──direct federation API──►  organizer's instance
                      signed + encrypted              stores privately
```

An RSVP is a private data envelope (see [Private data](#private-data)) with `data_type: "rsvp"`, delivered bilaterally. The full attendee identity and details are inside the encrypted payload — nothing identifying is in the outer envelope beyond the event ID.

```json
{
  "type":      "private_data",
  "context":   "7mnp...3rst",
  "data_type": "rsvp",
  "recipients": ["did:em:2abc...9xyz"],
  "payload":   "<status, message, details — encrypted to organizer's public key>",
  "keys": {
    "did:em:2abc...9xyz": "<symmetric key encrypted to organizer's public key>"
  },
  "created_at": "2025-04-15T00:00:00Z",
  "sig":        "3045...gh78"
}
```

The Event record publishes only an aggregate count, never identities:

```json
"rsvp": {
  "limit":    100,
  "count":    48,
  "deadline": "2025-06-10T00:00:00Z",
  "approval": "auto"
}
```

State transitions (confirmed, waitlisted, rejected, promoted) are returned by the organizer as signed `rsvp_state` private data envelopes delivered back to the attendee's instance. The exact logic of waitlists, approvals, and state transitions will be specified separately — the bilateral delivery model and private data envelope format are firm constraints; the state machine details are not yet final.

---

### Private data

Beyond RSVPs, events generate other private data — organizer notes, attendee messages, check-in records, ticket details, broadcast drafts. Rather than solving privacy separately for each, Evermeet defines a general **private data envelope** mechanism.

Any private record is wrapped in a sealed envelope:

```json
{
  "type":       "private_data",
  "context":    "7mnp...3rst",        // event or calendar this belongs to
  "data_type":  "rsvp" | "note" | "checkin" | "ticket" | "...",
  "recipients": ["did:em:2abc...9xyz", "did:em:5fgh...1klm"],
  "payload":    "<encrypted with per-envelope symmetric key>",
  "keys": {
    "did:em:2abc...9xyz": "<symmetric key encrypted to recipient's public key>",
    "did:em:5fgh...1klm": "<symmetric key encrypted to recipient's public key>"
  },
  "created_at": "2025-04-15T00:00:00Z",
  "sig":        "..."
}
```

The payload is encrypted with a random symmetric key. That key is then encrypted once per authorized recipient. Any recipient can decrypt their copy of the symmetric key and read the payload. Adding a recipient later is just encrypting the symmetric key to their public key — no re-encryption needed.

**Transmission vs replication are separate concerns:**

- **Bilateral delivery** (like RSVPs) — envelope sent directly instance-to-instance, never touches gossipsub
- **Encrypted replication** — envelope gossiped on `em/private/{context_id}` in sealed form; any seeding node stores ciphertext without being able to read it

The organizer chooses per data type which model applies. RSVPs use bilateral delivery. Organizer notes might use encrypted replication so co-owners always have them. The envelope format is identical either way — the difference is only in how it travels.

This gives a single primitive that covers all private data needs now and in the future, without designing a separate privacy mechanism for every feature.

---

### Private events

Events have four visibility levels:

| Level | Who can see it | Publicly indexed |
|-------|---------------|-----------------|
| `public` | Everyone | Yes |
| `unlisted` | Anyone with the direct link | No |
| `private` | Invited recipients only | No |
| `secret` | Trusted peers only, never hits public gossipsub | No |

**Public and unlisted** events have a fully readable event record. Unlisted events are simply not submitted to any indexer — they exist on the network but are undiscoverable without the direct link.

**Private events** encrypt the event payload. The record has a public envelope (so the network can replicate it) and an encrypted payload that only invited recipients can decrypt. Encryption uses a per-event symmetric key, distributed once per recipient:

```json
{
  "id":         "7mnp...3rst",
  "organizer":  "did:em:2abc...9xyz",
  "visibility": "private",
  "payload":    "<event details encrypted with symmetric key>",
  "recipients": [
    { "did": "did:em:5fgh...1klm", "key": "<symmetric key encrypted to recipient's public key>" },
    { "did": "did:em:9pqr...4stu", "key": "<symmetric key encrypted to recipient's public key>" }
  ],
  "sig": "..."
}
```

Each recipient decrypts their copy of the symmetric key using their own Ed25519 key (converted to X25519 for ECDH), then decrypts the payload. Adding a new invitee is just encrypting the symmetric key to their public key — the payload is never re-encrypted. The network replicates ciphertext freely; only invited recipients can read it.

**Secret events** never publish to the main `em/events` gossipsub topic at all. The record is only exchanged directly between the organizer's instance and explicitly trusted peer instances. No indexer ever sees it, no uninvited node can replicate it.

RSVPs for private and secret events follow the same encrypted envelope model as public events — the attendee list remains private regardless of event visibility.

---

### Extended features

The core primitives (User, Calendar, Event, RSVP) are intentionally minimal. More complex features are built on top as optional layers — the protocol stays simple, implementations add richness. The following are planned:

- **Ticket Sales & Registration** — free, paid, and crypto ticket types with per-event limits
- **Program Management** — structured schedule with sessions, time slots, and rooms
- **Speaker Management** — speaker profiles, bios, and session assignments
- **Attendee Management** — check-in, badges, and post-event credentials
- **Payment Processing** — pluggable payment hints (Lightning, on-chain, fiat); the protocol records intent, not transactions
- **Analytics & Reporting** — local analytics computed from signed records; no telemetry, organizers own their data
- **Broadcast Email** — organizer-initiated messages to confirmed attendees via the home instance

---

## How it works

Each Evermeet instance is a full node on the network. It handles everything locally — identity, event storage, P2P sync, search, and email notifications — as a single cohesive application. Users interact with a familiar web UI; the P2P layer runs silently underneath.

```
┌─────────────────────────────────────────────┐
│              Identity layer                  │
│     Ed25519 keypair · did:em: · key event log │
└─────────────────────────────────────────────┘
                      │
┌─────────────────────────────────────────────┐
│              P2P network                     │
│         node · node · node · node           │
│       (every node is an equal peer)         │
└─────────────────────────────────────────────┘
                      │
┌─────────────────────────────────────────────┐
│              Client layer                    │
│       Web app · CLI · Third-party clients   │
└─────────────────────────────────────────────┘
```

**Publishing an event:**
1. Organizer creates an event record (title, date, location, RSVP policy, payment hint)
2. The record is signed with the organizer's Ed25519 key
3. Signed record is stored locally and propagated to the P2P network
4. Every peer that receives it stores and replicates it, making it searchable

**Attending an event:**
1. Attendee discovers the event via any node's search
2. Attendee's instance creates an RSVP private data envelope, signed and encrypted to the organizer's key
3. Envelope delivered directly to the organizer's instance — no gossipsub, no third-party nodes
4. Organizer's instance decrypts and verifies — no fake RSVPs, no platform in between
5. Organizer's instance sends confirmation back to the attendee's instance via federation API
6. Attendee's instance notifies the user via SMTP — organizer never learns the attendee's email

**Email** is a pure notification side-channel. The protocol never touches email — the instance fires plain transactional SMTP for invitations, confirmations, updates, and reminders. It has nothing to do with identity.

**What a signed event record looks like:**

See the full schema in the [Primitives → Event](#event) section. Every event has an immutable founding document (hashed to produce its ID) and a mutable signed state record on top. The ID never changes even if the title, dates, or location are updated.

---

## Protocol

Evermeet uses two complementary communication patterns: **gossipsub** for replicating records across the network, and a **direct instance API** for privacy-sensitive coordination that cannot be broadcast.

### Gossipsub topics

All record replication happens over libp2p gossipsub. Nodes subscribe to topics they care about and receive new records as they are published. Every record is cryptographically signed — a node that receives a forged or tampered record simply drops it.

| Topic | Content | Who subscribes |
|-------|---------|----------------|
| `em/dids` | User identity key event log operations | All nodes |
| `em/profiles` | Public User profile updates (display name, avatar, bio) | All nodes |
| `em/calendars` | Calendar state updates | All nodes |
| `em/events` | Event state updates | All nodes |

`em/dids` carries the identity key event log — genesis, rotate, and migrate operations. Every node stores the full log for every identity it has seen, so key rotation and instance migration propagate automatically without any registry.

`em/profiles` carries mutable profile data separate from identity operations — display name, avatar, bio. Kept separate from `em/dids` so identity log stays minimal and audit-friendly.

Public records (profiles, calendars, events) gossip freely. **RSVPs are never gossiped** — they travel only via direct instance-to-instance federation. Other private data may optionally replicate via `em/private/{context_id}` in encrypted form — seeding nodes store ciphertext without being able to read it.

### Direct instance API

Some operations involve private data that cannot be broadcast. These go over a direct authenticated HTTP API between instances, resolved via the sender's identity key event log.

```
POST /api/federation/notify
     → ask another instance to deliver a notification to one of its users
     → organizer's instance never learns the recipient's email address
     → the recipient's instance handles its own user's SMTP

POST /api/federation/rsvp/confirm
     → organizer's instance confirms or rejects an RSVP
     → delivered directly to the attendee's instance
     → attendee's instance notifies the user

GET  /api/federation/user/{did}
     → resolve whether a did:em: identifier is known to this instance
     → used during invitation flow to detect existing cross-instance users
```

All federation API requests are signed by the requesting instance's key and verified before processing. An instance accepts federation requests only for users it actually hosts.

### Cross-instance RSVP flow

The full flow when a user on one instance RSVPs to an event hosted on another:

```
1. User on evermeet.app views event hosted on events.devcon.org

2. User clicks RSVP
   → evermeet.app signs the RSVP with the user's Ed25519 key
   → payload encrypted to organizer's public key
   → RSVP sent directly via federation API to events.devcon.org
   → no gossipsub, no third-party nodes involved

3. events.devcon.org receives the RSVP
   → verifies signature against attendee's did:em: key event log
   → decrypts payload with organizer's private key
   → records the RSVP privately

4. events.devcon.org sends confirmation
   → POST /api/federation/rsvp/confirm → evermeet.app
   → evermeet.app notifies its user via SMTP
   → events.devcon.org never learns the user's email address
```

### Seeding

Events and Calendars need to be seeded to remain available if the organizer's instance goes offline.

**Automatic peer replication** — nodes that receive a record via gossipsub cache it locally. Popular events naturally become more available as more nodes have seen them. This is the BitTorrent model.

**Explicit seeding** — any user can choose to seed a Calendar, which means their instance replicates all events and Calendar state published under it. Seeding covers public records only — RSVPs are never seeded, as they travel only between attendee and organizer instances directly.

**External archival (optional)** — events can be seeded to Swarm, Filecoin, or Arweave for permanent verifiable archival. The event record's content address is the same regardless of where it is seeded.

### Known limitations

**Identity loss on key loss** — if a user loses both their signing key and rotation key, their identity cannot be recovered. Instance failure alone is not fatal (the key event log is replicated across the network), but key loss is permanent. Custodial users are protected as long as their instance is alive and holds their encrypted keys — the app encourages key export and seeds the identity log to backup nodes proactively.

**RSVP loss on instance failure** — RSVPs are held only by the organizer's instance and each attendee's own instance. If the organizer's instance disappears, the collected RSVP list is gone. Attendees retain their own signed copy and can re-submit to a recovered or migrated instance, but there is no network-level redundancy by design — this is the tradeoff of not gossiping RSVPs. For important events, organizers should ensure their instance is reliable or use a Calendar with co-owners whose instances also act as RSVP backup recipients.

**Waitlists and RSVP state transitions require an online authority** — waitlist promotion and state transitions (confirmed, rejected, promoted) require the organizer's instance to be online to issue signed `rsvp_state` envelopes. If it disappears mid-waitlist, no new transitions can be issued. Confirmed attendees retain their signed proof; waitlisted attendees are in limbo. For events with waitlists, use a Calendar with at least one co-owner — the co-owner instance can take over authority. Standalone events should use auto-approval only.

**Proposals blocked by offline owners** — with a threshold greater than 1, a pending proposal requires co-signatures from multiple owners. If an owner's instance is offline or the owner is unresponsive, the proposal cannot reach threshold and the update is stuck. There is no timeout or override mechanism by design — governance requires active participation. For low-urgency updates, threshold 1 with a trusted single owner is simpler. For shared governance, all co-owners should maintain reliable instances.

**Revision chain conflicts** — if two owners independently push competing updates from the same `prev` hash (a fork), both are initially valid from their own instance's perspective. The network resolves conflicts by preferring the update with the most signatures, then earliest timestamp as a tiebreaker. This is a known edge case in distributed multi-sig systems; threshold > 1 makes it significantly less likely since coordination is required before publishing.

### What nodes verify

Every node independently verifies every record it receives before storing or forwarding it:

```
User record     → sig valid against current pk in user's key event log
Calendar state  → prev hash valid; sigs meet governance threshold; all signers in owners list
Event state     → prev hash valid; sigs meet governance threshold; all signers in owners list
Proposal        → prev hash valid; each sig valid; sigs do not yet meet threshold
RSVP envelope   → sig valid; delivered bilaterally, not verified by relay nodes
Federation req  → sig valid against requesting instance's key event log
```

A record that fails verification is silently dropped. No trusted intermediaries, no certificate authorities — cryptographic proof all the way down.

---

## The network

There is only one type of node. Every Evermeet node does the same things:

- **hosts** its own events and RSVP lists
- **relays** events it hears from peers
- **indexes** everything it receives for local search
- **serves** a web UI and search API to browsers and other clients

This is the BitTorrent model: every participant is a full peer. No special infrastructure nodes, no privileged relays, no separate indexer tier. The network is just nodes, all equal.

The only configuration distinction is **public vs. private**:

| Mode | Behaviour |
|------|-----------|
| `public` | Announces itself to the network, accepts incoming gossip from any peer, serves search API publicly |
| `private` | Only syncs with explicitly trusted peers listed in config, no public discovery |

This is a single config flag — not a different binary, not a different node type.

---

## Implementation

The reference implementation is written in **Go** (backend + P2P node) and **Svelte** (frontend), compiled into a single self-contained binary.

### Why Go

- Single binary compilation — the entire node ships as one executable
- `go-libp2p` — mature, production-tested libp2p implementation
- `embed.FS` — frontend assets baked directly into the binary at build time
- Excellent concurrency primitives for P2P networking
- Already battle-tested in ATScan and plcbundle

### Why Svelte

- Compiles to vanilla JS with no runtime — smallest possible bundle
- Fast, reactive UI without the overhead of a virtual DOM
- SvelteKit in SPA mode for client-side routing — no SSR needed
- Fits the "single binary, zero external dependencies" ethos

### Project structure

```
evermeet/
├── cmd/
│   └── evermeet/        # binary entrypoint, config loading
├── internal/
│   ├── node/            # libp2p host, gossipsub, DHT
│   ├── identity/        # Ed25519 keys, did:em: key event log, login methods
│   ├── store/           # SQLite via modernc.org/sqlite
│   ├── calendar/        # calendar record schema, co-owner management
│   ├── events/          # event record schema, signing, verification
│   ├── rsvp/            # RSVP state machine, bilateral delivery
│   ├── private/         # private data envelope, encryption, replication
│   ├── email/           # SMTP client, invitation and notification flows
│   └── api/             # HTTP API served to the frontend
├── web/                 # Svelte frontend source
│   ├── src/
│   │   ├── routes/      # SvelteKit pages
│   │   ├── lib/         # shared components and stores
│   │   └── api/         # typed API client
│   └── static/
└── embed.go             # embeds compiled web/build into the binary
```

### Key dependencies

| Package | Purpose |
|---------|---------|
| `github.com/libp2p/go-libp2p` | P2P networking, gossipsub, DHT |
| `modernc.org/sqlite` | Embedded SQLite, no CGo required |
| `github.com/multiformats/go-multibase` | base32 encoding for did:em: identifiers |
| `github.com/go-chi/chi` | HTTP router for the API |
| `svelte` + `@sveltejs/kit` | Frontend framework |
| `vite` | Frontend build tool |

### Development setup

```bash
# Prerequisites: Go 1.22+, Node.js 20+

# Clone
git clone https://github.com/evermeet/evermeet
cd evermeet

# Install frontend deps
cd web && npm install && cd ..

# Run in dev mode (Go backend + Vite dev server with HMR)
make dev

# Build production binary (compiles Svelte, embeds into Go binary)
make build

# Run
./evermeet --config evermeet.toml
```

---

## Self-hosting

Evermeet ships as a single static binary. No Docker required, no database to install separately, no Node.js runtime.

```bash
# Download and run
./evermeet --config evermeet.toml

# Open your node
open http://localhost:7331
```

Your node comes with a full web UI at `localhost:7331`. Expose it publicly by putting any reverse proxy (nginx, Caddy) in front of it.

**Config example (`evermeet.toml`):**
```toml
[node]
port     = 7331
data_dir = "./data"
public   = true

[identity]
key_file     = "./keys/signing.key"
rotation_key = "./keys/rotation.key"  # keep this offline

[email]
smtp_host     = "smtp.example.com"
smtp_port     = 587
smtp_user     = "evermeet@example.com"
smtp_password = "..."
from          = "Evermeet <evermeet@example.com>"

[p2p]
bootstrap_peers = [
  "/ip4/bootstrap.evermeet.app/tcp/4001/p2p/12D3..."
]
```

---

## Philosophy

Evermeet is built on a simple premise: **your community should not depend on any company's continued goodwill.**

Every year, platforms deplatform events, raise prices, change their terms, or shut down. Every time that happens, communities lose their history, their attendee lists, and their ability to organize. This is not an accident — it is the business model.

Evermeet makes this impossible by design:

- Events are signed records that belong to a cryptographic identity, not a platform account
- Identity is a keypair — it exists as long as you hold the key
- The network has no center to shut down
- The protocol is open — if evermeet.app disappears tomorrow, all nodes keep working

Sovereignty should not require technical expertise. A person who logs in with Google has a real self-sovereign identity from day one — they just haven't taken custody of the key yet. The path to full sovereignty is always open, never forced.

This is the parallel polis principle applied to event infrastructure: build the alternative, make the old system irrelevant.

---

## Roadmap

**MVP**
- [ ] Core event record schema + Ed25519 signing
- [ ] did:em: key event log, P2P replication via gossipsub
- [ ] Single node binary with embedded web UI
- [ ] Email / Google / passkey login
- [ ] RSVP via bilateral private data envelope (no gossipsub)
- [ ] Cross-instance federation (invitations, RSVPs, notifications)
- [ ] Public, unlisted, and private events
- [ ] Calendar and Event governance (threshold, proposals, revision chain)
- [ ] Transactional email via SMTP

**Extended features**
- [ ] Ticket types (free, paid, crypto)
- [ ] Program / schedule management
- [ ] Speaker management via Verifiable Credentials
- [ ] Attendee check-in and post-event credentials
- [ ] Broadcast email to attendees
- [ ] Analytics (local, no telemetry)
- [ ] Waitlist with co-owner fallback authority

**Integrations & clients**
- [ ] SIWE login
- [ ] Nostr keypair login (NIP-07)
- [ ] Key export and custody migration
- [ ] Mobile client
- [ ] Calendar feed export (iCal)
- [ ] External archival (Swarm, Filecoin, Arweave)
- [ ] Lightning / on-chain payment processing hints

---

## Contributing

Evermeet is an open protocol. The reference implementation is open source. If you want to contribute, run a node, or build a client — you don't need anyone's permission.

---

## License

MIT