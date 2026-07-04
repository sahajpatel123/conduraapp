# Threat Model: Reach & Ecosystem (Phase 12)

## 1. Scope

This document covers the threat model for Phase 12 components:
- **Skills Hub** (hub.condura.app) — public skill marketplace
- **P2P Sync** — device-to-device state synchronization
- **i18n** — internationalization layer
- **TUI** — terminal user interface

## 2. Trust Boundaries

| Boundary | Description |
|----------|-------------|
| **User ↔ TUI** | Local terminal session. No network exposure. |
| **Daemon ↔ Hub** | HTTPS with optional API key. Hub is untrusted content source. |
| **Device ↔ Device** | P2P with E2E encryption. No server in the data path. |
| **User ↔ i18n** | Locale files are local, loaded from embedded FS. |

## 3. Skills Hub Threats

### 3.1 Malicious Skill Download
- **Risk:** Attacker publishes a skill with dangerous steps (shell injection, data exfiltration).
- **Mitigation:** Every downloaded skill passes through `hub.Scan()` which checks for dangerous patterns. The Gatekeeper runtime layer enforces consent for all actions. Users must approve installation.

### 3.2 Checksum Tampering
- **Risk:** Man-in-the-middle modifies the skill archive in transit.
- **Mitigation:** SHA-256 checksum verification after download (`hub.Verify()`). HTTPS transport provides additional integrity.

### 3.3 Supply Chain Attack
- **Risk:** Compromised author publishes a malicious update to a trusted skill.
- **Mitigation:** Provenance tracking (author key, checksum). Users can pin skill versions. Trust levels (official/community/experimental) gate auto-updates.

### 3.4 Hub Server Compromise
- **Risk:** Attacker controls hub.condura.app and serves malicious content.
- **Mitigation:** Skills are scanned locally before installation. Users must explicitly approve. No automatic execution of downloaded skills.

## 4. P2P Sync Threats

### 4.1 Unauthorized Device Pairing
- **Risk:** Attacker pairs with a device and reads synced state.
- **Mitigation:** Ed25519 device identity with explicit pairing confirmation (QR code or 6-digit PIN). Noise XX handshake provides mutual authentication.

### 4.2 Eavesdropping
- **Risk:** Attacker observes sync traffic on the network.
- **Mitigation:** E2E encryption (Noise XX + AES-256-GCM). No server sees plaintext.

### 4.3 Device Loss/Theft
- **Risk:** Lost device has access to synced state.
- **Mitigation:** Device identity is keychain-backed. Revocation protocol allows other devices to invalidate a compromised device. Optional passphrase protection.

### 4.4 CRDT Conflict Manipulation
- **Risk:** Attacker injects malicious entries into the CRDT store.
- **Mitigation:** All entries are signed by the device's Ed25519 key. Signatures are verified on merge. Only entries from paired devices are accepted.

### 4.5 Replay Attack
- **Risk:** Attacker replays old sync messages to revert state.
- **Mitigation:** Vector clocks provide causal ordering. Old messages are rejected by the CRDT merge logic.

## 5. i18n Threats

### 5.1 Locale File Tampering
- **Risk:** Attacker modifies embedded locale files to display incorrect translations.
- **Mitigation:** Locale files are embedded in the binary via `go:embed`. Modification requires recompilation.

### 5.2 Injection via Locale Strings
- **Risk:** Malicious locale file contains executable content.
- **Mitigation:** Locale strings are data-only (JSON). No code execution path. UI renders as plain text.

## 6. TUI Threats

### 6.1 Local Privilege Escalation
- **Risk:** TUI runs with user privileges, no escalation path.
- **Mitigation:** TUI connects to daemon via IPC (Unix socket or TCP on localhost). No remote exposure.

### 6.2 IPC Impersonation
- **Risk:** Attacker connects to daemon IPC and sends malicious commands.
- **Mitigation:** Bearer token authentication on IPC. Unix socket permissions restrict access.

## 7. Non-Negotiables (from MISSION)

1. **Secrets never sync.** API keys, device private keys, and credentials are excluded from P2P sync.
2. **Imported artifacts are untrusted.** Skills from the hub are scanned and gated.
3. **P2P is E2E encrypted.** No server sees plaintext.
4. **i18n strings are data.** No code execution from locale files.
5. **Forcing E2E tests prove bypass-impossible.** Integration tests verify safety invariants.
