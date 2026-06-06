# Architecture 08 — P2P Encrypted Sync

> Your data, your devices, end-to-end encrypted. We never see the contents.

---

## The Goal

A user has a Mac, a Windows PC, and an iPad. They want their Synaptic setup — memory, skills, user model, settings — to follow them across all three.

Constraints:

1. **Privacy**: we (Synaptic) must never be able to read the contents.
2. **Offline**: must work without internet. (LAN discovery.)
3. **Resilient**: must work even if the central server is down.
4. **Zero-config for LAN**: devices on the same Wi-Fi discover each other automatically.
5. **Secure for WAN**: devices in different networks sync via a relay, E2E encrypted.
6. **Conflict-free**: if the user edits the same fact on two devices, the merge is sensible.
7. **No accounts required** for LAN sync. Accounts only for WAN relay.

---

## Threat Model

We assume:

- The Synaptic server is honest-but-curious: it can see metadata (who, when, how much) but not contents.
- The network is hostile: a coffee-shop attacker cannot decrypt traffic.
- A device can be lost or stolen: revocation must work.

We do not assume:

- That the user has a Synaptic Account.
- That any device is on the same network as another.
- That the user has a static IP.

---

## The Protocol Stack

```
┌────────────────────────────────────────────────────────────┐
│  Application layer: Memory, Skills, Settings, Audit        │
└──────────────────────────┬─────────────────────────────────┘
                           │  signed, encrypted records
                           ▼
┌────────────────────────────────────────────────────────────┐
│  Sync layer: CRDTs (Yjs / Automerge-style), conflict-free │
└──────────────────────────┬─────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│  Transport layer: libp2p (Noise XX, Yamux, mDNS, DHT)     │
└──────────────────────────┬─────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│  Discovery: mDNS (LAN) + DHT (WAN) + optional relay        │
└────────────────────────────────────────────────────────────┐
```

---

## Device Identity (Ed25519)

Every Synaptic install generates an **Ed25519 keypair** on first launch:

- **Private key**: stored in the OS keychain (Keychain on macOS, Credential Manager on Windows, libsecret on Linux). Never written to disk in plaintext.
- **Public key**: the device's identity. Format: `syn:device:ed25519:<base32-public-key>`.

The user has a **human-readable name** (e.g., "Sahaj's MacBook") that's a separate, signed record.

**Pairing**: to add a device, the user pairs it with an existing one via:

1. **QR code** (preferred, for devices with screens). One device shows a QR; the other scans.
2. **6-digit code** (for headless devices). Displayed on the source device, entered on the new one.
3. **Both devices on same LAN** (zero-config). They auto-discover via mDNS and prompt: "Pair with <name>?"

Pairing exchanges the public keys, **derives a shared Noise XX session**, and **creates a pairing record** in the audit log.

---

## The Noise XX Handshake

`libp2p` uses **Noise XX** by default:

- Mutual authentication (both sides prove they hold the private key).
- Forward secrecy (new ephemeral keys per session).
- Zero round-trip latency after handshake.

After Noise XX, all traffic is **AES-256-GCM encrypted** with a session key.

The server (relay) sees only ciphertext + the public keys of the endpoints. It cannot read the contents.

---

## Discovery

### LAN: mDNS

On the same network, devices advertise themselves via **mDNS** (`_synaptic._tcp.local`):

- Service type: `_synaptic._tcp`
- Port: 7666 (default; configurable)
- TXT records: `device_id`, `device_name`, `device_type`, `version`

When the user opens the "Devices" panel in the overlay, they see all Synaptic devices on the LAN. Tapping initiates pairing.

### WAN: DHT (libp2p's Kademlia)

For devices on different networks, we use libp2p's **DHT** to find peers. Each device publishes its peer ID to the DHT. Other devices can look up the peer ID and connect.

**DHT lookups are slow (seconds) but work over the open internet.** Behind NAT, devices need to use a relay.

### Relay (Optional)

For devices behind NAT (most home networks), a **relay** is needed. Options:

1. **User's own relay**: the user can run a relay on a VPS (we provide a Docker image). This is the most private.
2. **Synaptic relay** (default if no other option): a relay we run. We see only ciphertext + metadata. Users can opt out entirely.
3. **No relay**: if all devices are on LAN, no relay is used.

The user can also disable WAN sync entirely in Settings.

---

## The Data: CRDTs

For conflict-free merging, we use a **CRDT (Conflict-free Replicated Data Type)** approach, similar to Yjs or Automerge.

Each record is:

```json
{
  "id": "fact-123",
  "type": "fact",
  "subject": "user",
  "predicate": "works_at",
  "object": "Acme Corp",
  "confidence": 0.95,
  "updated_at": 1749823200,
  "updated_by": "syn:device:ed25519:abc123",
  "lamport": 42,
  "prev": "..."
}
```

CRDTs guarantee that if two devices edit the same record independently, the merge is **deterministic** and **preserves intent**. E.g., if device A increments a counter and device B increments the same counter, the result is +2, not "last write wins."

For **semantic facts**, we use **LWW (last-write-wins) with vector clocks** to detect concurrent edits. The user is asked to resolve conflicts when detected ("two devices said different things — which is right?").

For **episodes** and **skills**, we use **add-only CRDTs** (no deletions propagate; only the user's explicit "forget" command deletes).

For **audit log**, we use **append-only CRDTs** — no entries are ever deleted, only added.

---

## Sync Schedule

By default:

- **LAN**: instant (when a change is made, push to all paired LAN devices immediately).
- **WAN**: every 5 minutes, or on user action, or when the user re-opens Synaptic.

The user can configure:

- **Push only on Wi-Fi** (default).
- **Sync over cellular** (off by default, opt-in).
- **Manual sync only** (the user clicks "Sync now").
- **Bandwidth limit** for sync (e.g., 1 MB/day).

---

## Conflict Resolution

If two devices edit the same fact:

```
Device A: "User works at Acme Corp." (lamport 41, 14:32)
Device B: "User works at NewCo."      (lamport 42, 14:35)

Sync:
  A sees B's edit (lamport 42 > 41).
  A stores BOTH as a conflict.
  User is prompted: "I see two different facts. Which is correct?"
  User picks: NewCo.
  A and B converge to "User works at NewCo."
```

The conflict UI is non-blocking. Until the user resolves, **both facts exist** and the agent is told: "There is a conflict. Ask the user."

---

## Revocation

If a device is lost or stolen:

1. User goes to `Settings → Devices → <lost device> → Revoke`.
2. A revocation record is broadcast to all other devices.
3. The lost device's public key is added to a revocation list (CRDT set).
4. All other devices refuse to sync with the revoked device.
5. The user is prompted to **rotate the device-identity keypair** on all remaining devices (this re-pairs them).

The revocation is **cryptographically signed** by any of the user's currently-active devices. A single device cannot unilaterally revoke another (would require collusion or out-of-band confirmation).

---

## Storage on Each Device

Each device stores:

- The full local store (memory, skills, settings, audit).
- A **CRDT log** of recent changes (for fast sync).
- A **vector clock** of which records it has seen from which devices.
- The pairing records (one per paired device, with public keys + metadata).

Total CRDT log size: ~10MB for a heavy user. Compressed and pruned daily.

---

## The Server-Side (Synaptic's Own Servers)

We run:

- **DHT bootstrap nodes** (helps new devices find peers).
- **Optional relay** (for users who don't run their own).

We do **not** run:

- A central memory store.
- A central user model.
- A central audit log.

**Even our relay cannot read your data.** Noise XX + AES-256-GCM is end-to-end. If you don't trust our relay, run your own (we provide the Docker image and a one-line `docker run`).

---

## The Trust Hierarchy

```
You ─── trust ───► your own devices
        ─── trust ───► your own relay (if you run one)
        ─── may not trust ───► Synaptic relay (sees only metadata)
        ─── must not trust ───► us, the network, the OS vendor
```

This is the **zero-trust architecture**. The system is safe even if every server is compromised.

---

## Related Docs

- [00-overview.md](00-overview.md) — The conductor pattern
- [07-memory.md](07-memory.md) — What gets synced
- [PRIVACY.md](../PRIVACY.md) — Privacy guarantees for synced data
- [SECURITY.md](../SECURITY.md) — Revocation and threat model
