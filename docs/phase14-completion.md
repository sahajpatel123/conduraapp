# Phase 14 — Completion & Verification Checklist

Phase 14 adds the **optional** account, messaging (channels), P2P sync pairing,
Hub publishing, voice onboarding, and the marketing website. Everything here is
**additive**: Synaptic still works fully signed-out, offline, with no channels
and no account. This checklist covers the UI / website / docs slice (Agent 3).

How to run the app for manual checks:

```bash
# Daemon + GUI (dev)
cd app/web/frontend && npm install && npm run dev   # Svelte GUI
# In another shell, run the daemon (see CLAUDE.md "Running locally").

# Website
cd web && npm install && npm run dev                # http://localhost:3000
```

Automated gates (must pass before commit):

```bash
# Frontend
cd app/web/frontend && npm run check && npm run build
# Website
cd web && npm run lint && npm run build
# Go (backend the UI talks to)
go build ./... && go vet ./... && go test ./...
```

---

## 14B — Account UI

- [ ] Signed-out: sidebar footer shows a subtle **Sign in** link.
- [ ] Clicking **Sign in** opens `SignInPanel` (Google, GitHub, email magic link).
- [ ] Entering a valid email enables **Send link**; invalid email keeps it disabled.
- [ ] Google / GitHub buttons open the provider URL in the system browser
      (when OAuth client IDs are configured); otherwise an inline error explains
      the provider isn't configured — the app stays usable.
- [ ] After sign-in, the footer shows an avatar (or initial) + email chip.
- [ ] Clicking the chip opens `AccountMenu` (email, provider, tier).
- [ ] **Sign out** asks for confirmation, then returns to the signed-out state.
- [ ] Settings → **Account** mirrors the state (summary when signed-in; benefits
      list + Sign in when signed-out).
- [ ] Killing the daemon and reloading degrades gracefully to signed-out (no crash).

## 14C — Channels UI

- [ ] Sidebar shows a **Channels** icon; it routes to `#/channels`.
- [ ] Empty state explains how to connect Telegram.
- [ ] A non-`digits:secret` token shows the format hint and keeps **Connect** disabled.
- [ ] A well-formed token enables **Connect**; on success it appears in the list.
- [ ] Connected channel shows a status dot (green/connected, red/error).
- [ ] **Disconnect** confirms, then removes the channel.
- [ ] Status auto-refreshes (~10s) without manual reload.
- [ ] Settings → **Channels** links to the page.

## 14D — Website

- [ ] `web` builds (`npm run build`) and lints (`npm run lint`) with no errors.
- [ ] Landing page renders the hero ("AI on your computer, free"), features,
      OS download buttons, and a demo placeholder.
- [ ] **Manifesto** page renders mission / philosophy / privacy commitment.
- [ ] **Changelog** page renders content from the repo `CHANGELOG.md`
      (and degrades gracefully if missing).
- [ ] **Legal** page renders the EULA from `EULA.md`.
- [ ] Nav bar links work (Home, Manifesto, Changelog, Download, Legal); footer
      shows copyright + GitHub + Discord.

## 14F — Sync Pairing UI

- [ ] Discovered peers list auto-refreshes every 5s.
- [ ] **Pair** opens `PairingModal` instead of a `window.prompt()`.
- [ ] The modal shows a QR of this device's identity, the PIN, and a TTL countdown.
- [ ] Entering the peer's PIN and confirming completes pairing; the device moves
      to **Paired devices**.
- [ ] **Cancel** / Escape clears the pending pairing.
- [ ] **Revoke** confirms and removes a paired device.

## 14G — Hub Publish UI

- [ ] Hub header shows **+ Publish a Skill**.
- [ ] The modal validates: name required, version must be semver, archive required.
- [ ] An archive > 32 MB is rejected with a clear message.
- [ ] Submitting shows an uploading state, then success (with Hub link) or an error.
- [ ] **Done** / close resets the publish state.

## 14H — Voice in Onboarding

- [ ] Ready screen shows a **Set up voice** card reflecting `onboarding.probe_voice`
      (mic available?, wake word on/off + phrase).
- [ ] The card deep-links into Settings.
- [ ] Settings → **Voice**: wake-word toggle, sensitivity slider, hotword field,
      and **Test mic** (reports microphone permission status).
- [ ] Toggling/saving persists (re-open Settings reflects the saved values).

---

## Notes / known limitations

- Account sign-in (OAuth + magic link) requires the hosted auth service /
  configured OAuth client IDs to fully complete; the UI is wired end-to-end and
  degrades to clear inline errors when the backend isn't configured.
- Telegram channel status is reported by the daemon's reach subsystem; a live
  bot connection is required for messages to actually flow.
