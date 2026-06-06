# Security Policy

Synaptic performs physical, often irreversible actions on the user's operating system. Security is not a feature — it is a survival requirement. This document describes how to report vulnerabilities and how we handle them.

---

## Supported Versions

| Version | Supported |
|---|---|
| v0.1.0 (current dev) | ✅ Active development |
| v0.0.x (preview) | ❌ No longer supported |
| < v0.0.x | ❌ Not supported |

We commit to security updates for the current major.minor version. Older versions are not patched.

---

## Reporting a Vulnerability

**Please do not report security vulnerabilities via public GitHub Issues.**

Email: **security@synaptic.app**

PGP key: [synaptic.app/.well-known/pgp-key.asc](https://synaptic.app/.well-known/pgp-key.asc) (fingerprint: TBD)

### What to include

1. **Description** of the vulnerability.
2. **Steps to reproduce** (proof of concept preferred).
3. **Affected versions**.
4. **Impact** — what an attacker could achieve.
5. **Environment** — macOS / Windows / Linux version, Synaptic version, any relevant config.

### What to expect

- **Acknowledgment** within 48 hours.
- **Triage** within 5 business days. We will confirm or reject the report.
- **Fix timeline**: critical vulnerabilities patched within 7 days; high within 30 days; medium within 90 days.
- **Coordinated disclosure**: we will work with you on a disclosure timeline. Default is 90 days from report.
- **Credit**: we will credit you in the security advisory and (with your permission) on our website, unless you prefer to remain anonymous.

---

## Our Security Architecture (High-Level)

Synaptic is designed with defense in depth. Even if one layer is compromised, the system remains safe. The layers are:

1. **The Strategist (LLM)** — proposes what to do. Cannot execute anything.
2. **The Gatekeeper (deterministic)** — the only path to physical action. Cannot be prompt-injected.
3. **Model Isolation** — every model handoff is sanitized deterministically.
4. **TCC / OS Permissions** — granted by the user, revocable at any time.
5. **Sandboxed Execution** — shell commands run in a restricted env with an allowlist.
6. **Audit Log (HMAC-chained)** — every action is logged, append-only, tamper-resistant.
7. **Kill Switch (3 layers)** — hard hotkey, watchdog, network isolation. All independent of the agent.
8. **Anomaly Detector** — pauses the agent on suspicious behavior.
9. **Selective Perception** — twin-snapshot verification before any physical action.
10. **Encrypted Storage** — sensitive data at rest is AES-GCM encrypted.

See `CLAUDE.md` Section 2 for the full invariants and Section 5 for the 7 non-negotiables.

---

## Threat Model

We assume:
- The user has ChatGPT Plus, Claude Pro, or other subscriptions that may be compromised.
- The Telegram/Discord bot token may leak.
- A model will be prompt-injected at some point.
- The user's computer is connected to the internet and may be on a hostile network (coffee shop, etc.).
- The user themselves may make mistakes (clicking "Allow" too easily).

We do **not** assume:
- The user's machine is malware-free.
- The user will read every warning.
- The LLM provider is trustworthy (we use allowlists).

---

## Specific Risks We Mitigate

| Risk | Mitigation |
|---|---|
| **Agent clicks wrong button (stale screen)** | Twin-snapshot verification |
| **Agent runs in infinite loop** | Anomaly detector + watchdog |
| **Agent exfiltrates data** | Network isolation (kill switch layer 3) + spend monitor |
| **Model is prompt-injected via tool output** | Threat pattern scanner + delimiter markers |
| **OAuth token leaks** | Encrypted at rest + token rotation |
| **Sensitive data scraped (passwords, banking)** | Hardcoded blocklist + sensitive site detector + TCC tiers |
| **Audit log tampered** | HMAC-chained append-only log |
| **Agent acts while user is away** | Presence tracker + queue-when-absent policy |
| **Malicious skill imported from Hub** | Sandboxed skill execution + safety scanner |
| **Computer use triggers OS-level damage** | Pre-action verification + blast radius classification |

---

## Security Updates

Security updates are released as patch versions (e.g., v0.1.1, v0.1.2). They are auto-updated by default. Users can opt out of auto-updates in Settings, but we strongly recommend keeping auto-updates enabled.

Critical security updates are also announced via:
- In-app notification
- `synaptic.app/security`
- Discord `#security` channel
- Email to registered users (for Synaptic Account holders)

---

## Bug Bounty

We do not currently operate a paid bug bounty program. We do credit reporters (with permission) in security advisories and our Hall of Fame page.

If you are interested in a paid engagement, contact `security@synaptic.app`.

---

## Out of Scope

The following are **out of scope** for security reports:

- Theoretical vulnerabilities without a working proof of concept.
- Social engineering attacks against Synaptic staff.
- Physical attacks against the user's machine.
- Vulnerabilities in third-party software (LLM providers, MCP servers, CLI tools) — please report to those vendors.
- Self-XSS (you attacking your own install).
- Denial of service via resource exhaustion.
- Issues requiring the user to install malicious software first.

---

## Security Hall of Fame

_(No entries yet — be the first!)_

---

## Contact

- **Security reports**: security@synaptic.app
- **General questions**: support@synaptic.app
- **PGP**: [synaptic.app/.well-known/pgp-key.asc](https://synaptic.app/.well-known/pgp-key.asc)

---

**Thank you for helping us keep Synaptic and its users safe.**
