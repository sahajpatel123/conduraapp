# ADR-0005: P2P Sync Over a Central Server

- **Status**: Accepted
- **Date**: 2026-06-06
- **Deciders**: Condura core team
- **Supersedes**: —
- **Superseded by**: —

---

## Context

A user has multiple devices. They want their Condura data (memory, skills, settings, user model) on all of them.

The data is **highly personal**: API keys, OAuth tokens, conversation history, user-model beliefs. **We (Condura) must not be able to read it.**

Three architectures are possible:

1. **Central server** (like iCloud, Dropbox): our servers hold the encrypted blobs. We can see metadata (when, how much) but not contents.
2. **P2P** (like Syncthing, Resilio): devices talk directly. Our servers are only used for discovery and (optional) relay.
3. **P2P with end-to-end encryption** (like Matrix, Signal): devices talk directly or via relay, all traffic E2E encrypted. Our servers see only metadata.

We considered all three.

## Decision

**We use P2P with end-to-end encryption (option 3), built on libp2p.**

Our servers are used for:

- DHT bootstrap (helps devices find each other).
- Optional relay (for users who don't run their own).
- Version checks and updates.
- Skills Hub.

We do **not** use our servers for:

- Storing memory, skills, or user model contents.
- Storing audit logs.
- Storing OAuth tokens or API keys.

## Rationale

### Why not a central server

- **Trust**: users must trust us to not look at the data, to not be compelled by a government, to not have a bug that leaks data.
- **Single point of failure**: if our server is down, sync is down.
- **Cost**: we have to store and serve every user's data. That's expensive at scale.
- **Data residency**: some users want their data to live in their country. P2P solves this.
- **Privacy**: even with E2E encryption, a central server knows "user X has Y GB of data, syncs Z times a day." That's metadata we don't want to know.

### Why P2P

- **Trust**: we don't see contents. The user is the only one with the keys.
- **Resilience**: if our relay is down, LAN sync still works. If both devices are online, they sync directly.
- **Cost**: dramatically lower. We run only the bootstrap and (optional) relay.
- **Speed**: LAN sync is < 100ms. WAN sync is < 1s.
- **Sovereignty**: the user's data is on their devices. Period.

### Why libp2p

- **Mature**: used by IPFS, Filecoin, Ethereum 2.0, Polkadot.
- **Cross-platform**: works on every OS we support.
- **Crypto built-in**: Noise XX, Ed25519, all the standards.
- **Discovery**: mDNS for LAN, Kademlia DHT for WAN.
- **Relay**: built-in circuit relay.
- **Streams**: multiplexed streams over a single connection.
- **License**: MIT (we can ship it).

### Why not Syncthing-style (BEP)

- **No E2E encryption of contents by default** (transport is TLS, but devices can decrypt).
- **No fine-grained conflict resolution** (file-level, not record-level).
- **No streaming sync** (full-file).

### Why not Matrix

- **Federation**: not what we want. We want **P2P**, not "users can run their own server."
- **Complexity**: Matrix is a chat protocol; using it for sync is overkill.
- **No built-in CRDTs**: we'd have to add them.

### Why not custom

- **Time**: libp2p is 5+ years of work. We can't replicate that.
- **Bugs**: libp2p is battle-tested. Custom = unknown attack surface.
- **Audits**: libp2p is audited. Custom would need to be.

## Consequences

### Positive

- We don't see the user's data. Ever.
- The system is resilient to our outages.
- The user is the only one with the keys.
- LAN sync is fast and zero-config.

### Negative

- **NAT traversal**: not all devices are reachable from the internet. We rely on relays for those.
- **Discovery latency**: DHT lookups are slow. The first sync after a long offline period can be slow.
- **Relays needed**: we run a default relay. Users can run their own.
- **Mobile devices**: iOS and Android limit background networking. We'll have to use platform-specific push (APNs, FCM) to wake up mobile clients. (Deferred to v0.2.)
- **Conflict resolution**: CRDTs are powerful but not magic. Some conflicts need user resolution.

### Neutral

- We commit to **libp2p** as the transport.
- We commit to **Noise XX** for handshake.
- We commit to **AES-256-GCM** for transport encryption.
- We commit to **Ed25519** for device identity.
- We commit to **mDNS** for LAN discovery.

---

## The Sync Protocol Stack

```
┌────────────────────────────────────────────────────┐
│  Application: Memory, Skills, Settings, Audit      │
└──────────────────────────┬─────────────────────────┘
                           │  signed, encrypted records
                           ▼
┌────────────────────────────────────────────────────┐
│  Sync: CRDTs (Yjs-style), vector clocks, LWW       │
└──────────────────────────┬─────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────┐
│  libp2p: Noise XX, Yamux, mDNS, Kademlia DHT,      │
│          Circuit Relay v2                          │
└────────────────────────────────────────────────────┐
```

---

## Threat Model

We assume:

- Our servers are **honest-but-curious**: they see metadata, not contents.
- The network is **hostile**: an attacker can see all traffic.
- A device can be **lost or stolen**: revocation must work.
- The user may be on an **untrusted network** (coffee shop).

We do not assume:

- That the user has a Condura Account.
- That any device is on the same network.
- That the user has a static IP.

---

## The Server's Role (Minimal)

Our servers run:

1. **DHT bootstrap nodes**: 5-10 nodes, geographically distributed. They help new devices find peers.
2. **Optional relay**: for users behind NAT. End-to-end encrypted.
3. **Version check**: `GET https://updates.condura.app/v0.1.0/check?version=X` — returns "newer version available."
4. **Skills Hub**: `https://hub.condura.app` — public marketplace for skills.
5. **Crash reports**: opt-in only.

We do **not** run:

1. A central memory store.
2. A central user model.
3. A central audit log.
4. A central OAuth token store.
5. A central skill store (the Hub is public, not user-specific).

---

## The User's Choices

```yaml
sync:
  enabled: true
  mode: p2p                # p2p | cloud | both | off
  discovery:
    lan: true              # mDNS
    wan: true              # DHT
  relay:
    use_synaptic_relay: true   # default relay
    custom_relay: null         # or a URL
  bandwidth:
    cap_mb_per_day: 100
    on_cellular: false
    on_metered: false
  conflicts:
    auto_resolve: false
    notify: true
```

---

## When P2P Isn't Enough (Future)

For v0.2, we may add:

- **Mobile push** (APNs, FCM) to wake up mobile clients.
- **Cloud relay** (a paid relay with better uptime).
- **Snapshot restore** (a way to restore from a friend's device if all of yours are lost).

None of these require us to see the user's data. They're all E2E encrypted.

---

## Related Docs

- [00-overview.md](../architecture/00-overview.md) — The conductor pattern
- [08-sync.md](../architecture/08-sync.md) — P2P sync protocol details
- [07-memory.md](../architecture/07-memory.md) — What gets synced
- [PRIVACY.md](../../PRIVACY.md) — Privacy guarantees
- [SECURITY.md](../../SECURITY.md) — Revocation and threat model
