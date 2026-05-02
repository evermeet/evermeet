# Evermeet

A self-hosted, federated platform for publishing and discovering events — meetups, conferences, hackathons, and more.

A viable alternative to Lu.ma and Meetup.com that puts sovereignty back in the hands of organizers, without sacrificing the UX people are used to.

---

## Features

- **Single binary** — one file runs the full node with web UI included
- **Self-sovereign identities** — every user has an Ed25519 keypair; login with passkey or email
- **Federated discovery** — find events across all nodes via a DHT-based P2P network
- **Cryptographically signed events** — every event record is tamper-proof and verifiable
- **Censorship-resistant** — no platform can take down your event
- **Private by default** — RSVPs are encrypted; only the organizer sees who is coming
- **Open protocol** — any client can implement it

---

## Quick start

**From binary:**
```bash
./evermeet --config evermeet.toml
```

**From source:**
```bash
make setup   # install web dependencies
make build   # build web + Go binary
make run
```

**With Docker:**
```bash
docker run -p 7331:7331 -p 4001:4001 -v $(pwd)/data:/data evermeet/evermeet:latest
```

**Build locally:**
```bash
docker build -t evermeet .
docker run -p 7331:7331 -p 4001:4001 -v $(pwd)/data:/data evermeet
```

**With Docker Compose:**
```bash
docker compose up
```

Open `http://localhost:7331` — on first run you will be prompted to create an admin account.

---

## Configuration

Copy the defaults file and edit to taste:

```bash
cp evermeet.defaults.toml evermeet.toml
```

Key settings:

| Key | Default | Description |
|-----|---------|-------------|
| `node.port` | `7331` | HTTP listen port |
| `node.base_url` | `http://localhost:7331` | Public URL (required for federation) |
| `node.public` | `true` | Allow unauthenticated event browsing |
| `node.data_dir` | `./data` | Database, keys, and blob storage |
| `p2p.listen_port` | `4001` | libp2p DHT port |
| `p2p.bootstrap_peers` | `[]` | Other Evermeet nodes to connect to |
| `email.smtp_host` | — | SMTP server (falls back to local sendmail) |

The config can also be edited live from the admin UI at `/admin/config` — the server restarts automatically after saving.

---

## Bootstrap node

A bootstrap node helps other instances discover each other via the DHT. It runs without a database or web UI — just the P2P layer.

```bash
./evermeet --bootstrap --p2p-port 4001 --data ./data
```

On startup it prints its multiaddr:
```
listen /ip4/1.2.3.4/tcp/4001/p2p/12D3KooW...
```

Add that address to other instances' `bootstrap_peers` in their config.

---

## Admin UI

The web UI includes an admin panel at `/admin` with:

- **Overview** — instance stats, uptime, memory usage
- **Network** — P2P peers, DHT routing table stats
- **Objects** — users, events, calendars, blobs
- **Email** — mail transport status and test sender
- **Admins** — manage admin accounts and roles
- **Config** — live TOML editor with restart

---

## Deployment

See [`docker-compose.yml`](docker-compose.yml) for a production-ready setup including Watchtower for automatic updates.

Tagged releases are automatically built and pushed to [Docker Hub](https://hub.docker.com/r/evermeet/evermeet):

```bash
docker pull evermeet/evermeet:latest
```

To cut a release:
```bash
make release
```

---

## License

MIT
