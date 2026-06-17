# Privacy Policy

**Last updated:** 2026-06-06

This Privacy Policy describes how Condura ("we", "us", "our") handles your information when you use our desktop application ("Software"). We take your privacy very seriously. **By design, Condura keeps your data on your computer and does not send it to us.**

---

## TL;DR

- **Your data stays on your computer.** Memory, skills, audit logs, embeddings, API keys — all stored locally, encrypted.
- **Telemetry is OFF by default.** We don't collect anything unless you opt in.
- **The only outbound network calls** are to the LLM provider(s) you configured and to our update server. No other calls.
- **No accounts required** for the core product. Condura Account (for Skills Hub, donations, support) is optional.
- **P2P sync** is device-to-device, end-to-end encrypted. We never see the contents.

---

## Information We Do NOT Collect

By default, we collect **nothing**. Specifically:

- ❌ Your API keys (they're encrypted locally and never leave your machine)
- ❌ Your prompts or conversations
- ❌ Your memory, skills, or audit logs
- ❌ Your screenshots or computer-use actions
- ❌ Your file system contents
- ❌ Your usage patterns
- ❌ Your machine's IP address
- ❌ Your location

---

## Information We DO Collect (Opt-In Only)

If you opt in to telemetry, we collect **only** the following, **anonymized and aggregated**:

- ✅ App version and OS version
- ✅ Number of actions performed per day (no details about what)
- ✅ Crash reports (stack traces, no user data)
- ✅ Feature usage counts (e.g., "user opened settings 3 times")

This helps us prioritize features and fix bugs. You can opt in or out at any time in Settings → Privacy. We will **never** sell or share this data with third parties.

---

## Where Your Data Lives

| Data Type | Storage Location | Encrypted? |
|---|---|---|
| API keys | `~/.synaptic/store.db` | ✅ AES-256-GCM |
| OAuth tokens | `~/.synaptic/store.db` | ✅ AES-256-GCM |
| Memory (episodic, semantic) | `~/.synaptic/store.db` | ✅ at rest |
| Skills | `~/.synaptic/store.db` + `~/.synaptic/skills/` | ✅ at rest |
| Audit log | `~/.synaptic/store.db` | ✅ at rest + HMAC chain |
| User-adaptive model | `~/.synaptic/store.db` | ✅ at rest |
| Configuration | `~/.synaptic/config.yaml` | ⚠️ plaintext (no secrets) |
| Screenshots (replay) | `~/.synaptic/replay/` | ✅ at rest |
| Crash reports | sent to our error tracking service | ⚠️ if opted in |

**Backup location** (on uninstall): `~/Documents/synaptic-backups/`

**P2P sync** (if enabled): your other devices, via E2E encrypted protocol. We never see the contents.

**Cloud sync** (if enabled): your chosen destination (iCloud Drive, Google Drive, Dropbox, or our own E2E encrypted server). Encrypted in transit and at rest.

---

## Network Calls

Condura makes outbound network calls **only** to:

1. **LLM providers you have configured** (Anthropic, OpenAI, Google, xAI, Mistral, DeepSeek, OpenRouter, etc.) — to send your prompts and receive responses.
2. **Our update server** (`updates.condura.app`) — to check for new versions every 6 hours and on launch. This sends your current version and OS type only.
3. **OAuth providers** (if you connect a subscription) — for authentication.
4. **Optional: Skills Hub** (`hub.condura.app`) — if you browse or publish skills.
5. **Optional: P2P sync relays** — only if your devices can't connect directly.
6. **Optional: our error tracking service** — only if you opt in to crash reporting.

We do **not** make any other outbound network calls. We do **not** call home, phone home, or report your usage to anyone.

---

## Local Network Access

Condura does **not** scan your local network, your files, or your apps unless you explicitly ask it to. Computer-use features only access apps and files when:

- You summon the agent and ask it to do something.
- You have pre-approved the app or action in your policy.

The agent will **never** silently scan your system in the background.

---

## Microphone, Camera, Screen Recording

These are sensitive OS permissions. Condura requests them only when needed:

- **Microphone**: requested when you enable voice input. You can deny this and still use the app (text only).
- **Screen Recording**: requested when the agent needs to see the screen (vision CUA fallback, screenshots). Denying this limits computer-use to Accessibility API only.
- **Camera**: Condura does **not** access the camera. If you see a camera permission prompt, something is wrong — please report it.

You can revoke these permissions at any time in System Settings (or your OS's privacy settings).

---

## Children's Privacy

Condura is not intended for children under 13. We do not knowingly collect information from children under 13. If you are a parent and believe your child has used Condura, please contact us at privacy@condura.app.

---

## International Data Transfers

If you use a cloud-synced feature (e.g., cloud backup, P2P relay fallback), your data may transit through servers in various countries. All such transfers are E2E encrypted — we cannot read the contents.

If you are in the EU/UK/EEA, you have the right to:

- Access your personal data (it's all on your computer).
- Rectify inaccurate data.
- Erase your data (delete the app and clear the local store).
- Restrict or object to processing.
- Data portability (export everything via Settings).
- Lodge a complaint with your supervisory authority.

If you are in California (CCPA), you have similar rights.

To exercise any of these rights, contact privacy@condura.app. Since your data is local, you can typically do this yourself by exporting or deleting your local store.

---

## Third-Party Services

Condura integrates with third-party services. Their privacy policies apply to data you send to them:

- **Anthropic** (Claude API) — [anthropic.com/privacy](https://www.anthropic.com/privacy)
- **OpenAI** — [openai.com/policies/privacy-policy](https://openai.com/policies/privacy-policy)
- **Google** (Gemini API) — [policies.google.com/privacy](https://policies.google.com/privacy)
- **xAI** (Grok) — [x.ai/privacy-policy](https://x.ai/privacy-policy)
- **Mistral** — [mistral.ai/terms/privacy](https://mistral.ai/terms/privacy)
- **OpenRouter** — [openrouter.ai/privacy](https://openrouter.ai/privacy)

When you connect a subscription or paste an API key, you are agreeing to the privacy policy of that provider. We encourage you to review them.

---

## Cookies and Tracking

Our website (`condura.app`) uses minimal, privacy-respecting analytics (Plausible or Umami, no cookies, no personal data). We do not use Google Analytics, Facebook Pixel, or any cross-site tracking.

---

## Changes to This Policy

We may update this Privacy Policy from time to time. We will notify you via:

- In-app notification
- `condura.app/privacy` (date updated)
- Email to registered users (for Condura Account holders)

Material changes will be announced at least 30 days in advance, during which you may export and delete your data if you disagree with the changes.

---

## Data Retention

- **Local data**: as long as you keep Condura installed. Deleted when you uninstall (with auto-backup to `~/Documents/synaptic-backups/` for 30 days, then permanently deleted).
- **Opt-in telemetry**: 90 days, then aggregated (anonymized).
- **Cloud backups** (if you use them): as long as you keep the account. Delete your account to delete all backups.
- **OAuth tokens**: rotated automatically. Revoked when you log out or uninstall.

---

## Contact

If you have questions about this Privacy Policy, contact:

- **Email**: privacy@condura.app
- **Web**: condura.app/privacy
- **Mail**: Condura Privacy Team, [address TBD]

---

**By using Condura, you acknowledge that you have read and understood this Privacy Policy.**

**Your data is yours. We just help you use it.**
