# SCREEN_SYNC — Condura

> **P2P device pairing as a garden of nodes threaded together.** This device
> is the green node at the centre (the agent's presence); each paired device
> is a pollen node around it; the **threads between them are the sync links**
> — they draw in on mount and breathe at a 9 s cadence (a link is alive). No
> central server is visible — only the threads between your machines.
>
> **Route:** `#/sync` · **Component:** `app/web/frontend/src/lib/condura/Sync.svelte`

---

## 1. LAYOUT & CONTENT

Single-column surface, `max-width: 64rem`, `padding: var(--space-6) var(--space-5) var(--space-5)`, vertical scroll. Children top→bottom:

| Zone | Content | Sizing |
|---|---|---|
| **Header** | eyebrow `P2P Sync` + headline "Your machines, *threaded.*" (italic accent on "threaded" via `.alive`) + lead (one sentence) | left text + right status pill |
| **Status pill** (right of header) | `<Pulse>` + mono label (`Sync paused` / `Awaiting a peer` / `N paired`) | pill, `7px 12px`, hair-strong border |
| **Rule** | `<Thread orientation="h">` | 2 px tall |
| **Error rail** (conditional) | italic headline + sub + retry pill + `err-hair` | full width |
| **Canvas** | this-device node (centre) + paired-device nodes (radial) + inter-device SVG threads + whisper at bottom | `flex: 1 1 auto`, `min-height: 420px`, `border-radius: var(--r-lg)` |
| **Peers rail** | eyebrow `Discovered on your LAN` + Refresh button + peer chips | full width |
| **Footer whisper** | italic Instrument Serif: "No central server — just threads between your machines." | centred |

### This-device identity (green node, centre)

| Element | Source | Visual |
|---|---|---|
| 96 × 96 dot | radial gradient `--synapse-glow → --synapse-deep` | glow + inset paper ring |
| 72 × 72 face | white paper card | rounded 14 px |
| QR (when identity ready) | `qrcode.toDataURL({payload: {v:1, device_id, name}, margin:1, width:160})` | covers the face |
| Glyph fallback | `<Glyph name="sync" size=26>` while QR loads | centred |
| Halo | `inset: -10px` ring, 1 px synapse-glow | breathing 9 s |
| Label | `name` + first 12 chars of `device_id` (mono) | under the dot |

### Paired device (pollen node, radial)

Positioning: `RADIUS = 34%` from centre, angle = `(i / n) * 2π − π/2` (12 o'clock first). For `n == 1`, angle is `−π/2` (still top).

| Element | Source | Visual |
|---|---|---|
| 26 × 26 dot | radial gradient `--pollen-light → --pollen` + `--pollen-halo` shadow | spring-in on mount |
| Halo | `inset: -8px` ring, 1 px pollen, 0.3 opacity | static |
| Label | `device_name` + "paired {fmtSeen}" (mono) | under the dot |

### Peers rail (LAN-discovered)

Each peer chip: name + fingerprint/address (mono, first 12 chars) + `<Button variant="primary" size="sm">Pair</Button>`. Refresh button (top-right) re-runs `ipc.syncStatus()` + `sync.refresh()`.

### Footer honest-data note (under everything)

> "Memory, skills, and config flow directly between your devices — encrypted, no central server. **Logs and API keys never leave a machine.**"

| Synced | NOT synced |
|---|---|
| Memory (episodic + semantic) | Audit log |
| Skills | Screenshots (Replay) |
| Config | API keys |
| | Logs |

---

## 2. STATE MATRIX

Every state is reachable. The defaults in `sync.*` derive from polling, not user input.

| State | Trigger | Surface signal | Exact copy |
|---|---|---|---|
| **empty** | `sync.pairs.length === 0 && !pending && !sync.loading && !sync.error` | green node alone at centre + bottom whisper | eyebrow (none) · whisper: "Discovering peers on your LAN…" |
| **loading** | `sync.loading === true && !status` | green node face shows `<Glyph name="sync">` (QR not yet rendered) | header badge: "Awaiting a peer" + `<Pulse phase="thinking">` |
| **loaded** | `status !== null && sync.pairs.length === 0` | green node face shows QR (160 px) | header badge: "Awaiting a peer" + `<Pulse phase="thinking">` |
| **paired** | `sync.pairs.length >= 1` | pollen nodes around centre + threads drawn + breathing | header badge: "N paired" + `<Pulse phase="idle">` |
| **pairing (pending)** | `sync.pendingPin && sync.pendingPeerId` | green node + pollen nodes + centre card | card headline: "Pairing with {peerName}" · centre PIN: "{pendingPin}" (mono 26 px, +0.18em) · TTL: "M:SS" or "expired" · hint: "Read the PIN on the peer device, then type it here to seal the link." · primary CTA: "Seal link" (disabled until `pinReady`) · ghost: "Cancel pairing" |
| **syncing (per-row)** | new thread just drawn | node halo pulses brighter for one breath cycle | no new copy |
| **error** | `sync.error !== null` | italic err-state block above canvas + red Pulse | headline: "We couldn't reach the daemon." · sub: "{error} Try again — sync state lives in the daemon." · CTA: "Try again" |
| **revoke popover** | paired-node click | backdrop + centred popover (320 px) | eyebrow "Paired device" (danger colour) · name · full `device_id` (mono, word-break) · warn: "Revoking severs the thread. The other side loses access on its next sync." · ghost "Keep" · danger "Revoke" |

`pinReady = /^\d{4,8}$/.test(pinInput.trim())` (4–8 digits). The PIN input enforces `inputmode="numeric"`, `maxlength="8"`. The TTL ring depletes over `pendingTotalMs` (set once on pending arm, ticks every 250 ms).

---

## 3. MOTION CHOREOGRAPHY

All durations are tokens (`--dur-fast 140` / `--dur 280` / `--dur-slow 520` / `--dur-cine 900`); all eases are `--ease` (`cubic-bezier(0.22, 1, 0.36, 1)`) unless noted. Reduced-motion short-circuits via the single block in `condura.css`.

| Element | Property | Duration | Easing | Trigger |
|---|---|---|---|---|
| **Thread draw** (this-device ↔ paired) | SVG `stroke-dashoffset 1 → 0` | 1100 ms | `--ease` | node enters / pair complete |
| **Thread breathe** | `opacity 0.55 ↔ 1` (synapse) / `0.12 ↔ 0.34` (glow) | 9 s loop | `ease-in-out` | always-on for live threads; staggered `--breathe-delay: 1.1s + i*0.6s` |
| **Thread fray (revoke)** | `stroke-dasharray 0.18 0.16` + `opacity 0` + `dashoffset 1` | 900 ms | `--ease` | revoke click |
| **This-device halo** | `opacity 0.4` breathing | 9 s loop | `ease-in-out` | mount |
| **Paired node enter** | `opacity 0 → 1` + `scale 0.3 → 1.12 → 1.0` (spring) | 700 ms | `--ease` | pair complete, staggered 0.08 s per index |
| **Paired node revoke-fade** | `opacity 1 → 0` + `scale 1 → 0.5` | 900 ms | `--ease` | revoke click |
| **QR render** | `qrcode.toDataURL` (no entrance; appears in face) | async | — | identity arrives / changes |
| **Pending card enter** | `animation: blur-in` | `--dur-slow` | `--ease` | pending arm |
| **Revoke popover enter** | `animation: blur-in` | `--dur` | `--ease` | node click |
| **TTL ring deplete** | `stroke-dashoffset: 1 → 0` (tweened by `transition: 250ms linear`) | 250 ms per tick | linear | pendingExpiresAt tick |
| **Pollen dot hover** | `transform: scale(1.18)` | `--dur` | `--ease` | mouseenter |
| **Selected pollen dot** | `box-shadow: 0 0 0 4px pollen-tint + halo` | `--dur` | `--ease` | node click |
| **Refresh button hover** | `translateY(-1px)` + tint | `--dur` | `--ease` | mouseenter |
| **Refresh button press** | `scale(0.97)` + `filter: brightness(0.95) saturate(1.1)` + `translateY(0.5px)` | `--dur-fast` | `--ease` | mousedown |
| **PIN input focus** | `border-color: --pollen` + `box-shadow: 0 0 0 4px --pollen-halo` | `--dur` | `--ease` | focus |
| **err-hair** | `transform: scaleX(0 → 1)` | 600 ms | `--ease` (120 ms delay) | error render |
| **Reduced motion** | `*, *::before, *::after { animation-duration: 0.01ms !important; transition-duration: 0.01ms !important; }` + `.paper-grain, .mote, .ambient-thread { display: none; }` | global | — | OS `prefers-reduced-motion: reduce` |

---

## 4. KEYBOARD

| Key | Context | Action |
|---|---|---|
| **Tab** | anywhere in surface | cycles through Refresh → Peer Pair buttons → (none for QR) |
| **Shift+Tab** | anywhere | cycles in reverse |
| **Enter** | focused Pair button | starts pairing (`startPair(p.device_id)`) |
| **Enter** | PIN input (when `pinReady`) | calls `confirmPair()` |
| **Escape** | selected pair popover open | closes popover (`selectedPair = null`) |
| **Escape** | pending card open | cancels pending (`sync.clearPending()`) |
| **⌘P** | anywhere on Sync surface | opens pairing flow (when at least one peer is discovered — first peer's `startPair`) |
| **⌘R** | anywhere | refresh (`refreshAll()`) |
| **⌘,** | anywhere on app | opens Settings (via shell `Shortcuts` overlay) |
| **?** | anywhere on app | opens Shortcuts overlay (lists ⌘P, ⌘R, Esc) |

The QR face has `pointer-events: none` (it's not interactive). The green centre dot is `pointer-events: none` (it's this device, not selectable). Only paired pollen dots + peer chips + Refresh + PIN input are focusable.

**Focus halos:**
- pill buttons → 2 px synapse ring + 5 px pollen halo (no inset line)
- rounded inputs (PIN) → 4 px pollen halo + 1 px synapse inset
- pollen dots → 2 px synapse outline + 4 px offset (rounded, follows the shape)

---

## 5. COMPONENTS USED

| Component | Where | Why |
|---|---|---|
| **`Thread`** | horizontal rule between header and canvas | the signature — completion made visible |
| **`Pulse`** | header status pill, peer chips empty hint, pending-card awaiting, paired-node halo, error row | the only allowed loading indicator (per MOAT §2.5) |
| **`Glyph`** | header dot, QR fallback, refresh button, revoke button | brand line vocabulary; never emoji |
| **`Button`** | Refresh, Pair (per peer), Cancel pairing, Seal link, Keep, Revoke | primary CTA + ghost + danger variants |
| **`qrcode`** (npm) | renders `{v:1, device_id, name}` to a data-URL PNG | 160 px, margin 1, single QR per identity change |

There is no `PairingModal.svelte` here — the pending card is embedded in `Sync.svelte` itself. The older `app/web/frontend/src/lib/components/PairingModal.svelte` is a separate floating variant for ad-hoc flows (interview scenarios); it shares the same RPC contract (`sync.pairWith`, `sync.confirmPairing`).

---

## 6. DATA FETCHED

All via `sync` store (`app/web/frontend/src/lib/stores/sync.svelte.ts`) + `ipc` client.

| Call | Direction | When | Returns |
|---|---|---|---|
| `ipc.syncStatus()` | request | mount + every 5 s | `{running, device_id, name}` |
| `sync.refresh()` | request | mount + every 5 s + Refresh click | `{peers, pairs, pendingPin, pendingPeerId, pendingExpiresAt}` |
| `sync.pairWith(peerId)` | request | peer Pair click | sets `pendingPin`, `pendingPeerId`, `pendingExpiresAt` |
| `sync.confirmPairing(pin)` | request | Seal link / Enter in PIN input | `boolean` (success clears pending; pair appends to `pairs`) |
| `sync.clearPending()` | request | Cancel pairing / Escape | clears pending fields |
| `sync.revoke(deviceId)` | request | Revoke click | `boolean`; success → pair removed |
| `QRCode.toDataURL(payload, opts)` | local (no IPC) | `$effect` on `status.device_id`/`status.name` change | data-URL string |

**SSE subscription:** none on this surface (sync state is polled every 5 s, not streamed; the daemon's bus is read on demand). The store does not currently expose a `sync.subscribe` — polling is the contract.

---

## 7. DESIGN DECISIONS

| Decision | Rationale (MOAT / DIRECTION / APPFLOW) |
|---|---|
| **Garden-of-threads, not a settings list** | APPFLOW §4.4: "garden of nodes threaded together" is the visual grammar — the threads ARE the sync links, not decoration. |
| **This device is the green centre node** | DIRECTION §2: nodes are alive; this device's presence is the agent's pulse. The QR face proves identity (not a logo, not a placeholder). |
| **Paired devices are pollen (warm amber), not synapse** | DIRECTION §4: status ≠ brand. Synapse is for brand links; pollen is for action/identity. Paired devices are *the user's* — they get pollen, not the brand green. |
| **QR + truncated pubkey = identity proof** | MOAT §1 trust: the QR encodes `{v:1, device_id, name}` so the user can verify the identity they are pairing with is the one they see on screen. The truncated 12-char ID under the label matches the QR payload. |
| **Footer honest-data note** | DIRECTION §1: warm, awake, never louder. The note is italic Instrument Serif — paper voice. It states what's synced AND what isn't. No cloud. |
| **Pairing is an embedded card, not a modal** | MOAT §2.8: `.c-modal` is for blocking consent; pairing is a *handshake*, not a confirmation. The centred card with `blur-in` is the right gesture. |
| **Thread draw is the callback to the titlebar** | DIRECTION §5 (Thread is the signature): when a thread draws between two devices, the same gesture the titlebar uses to signal completion. The user learns one motion, sees it everywhere. |
| **TTL ring is pollen, not synapse** | DIRECTION §4 (status vs brand): waiting/awaiting is a *status*, not the brand. Pollen is the action colour; the ring is the action (waiting for the peer). |
| **Polling 5 s, not SSE** | APPFLOW §4.4: mDNS discovery is racy; polling is honest. SSE is reserved for the audit chain and the conversation stream where real-time matters. |
| **Pending PIN is mono-uppercase, +0.18em tracking** | DIRECTION §3: mono is for IDs, paths, timestamps. The PIN is read as a code, not a word. |
| **Revoke is destructive — danger button + danger eyebrow + warn copy** | CLAUDE.md §2.1 invariant #4 (user can always stop the agent): revoking a paired device is a real action. The popover must read as consequential, not casual. |
| **No "spinner" anywhere** | MOAT §2.5 + §4 #7: loading shows a Pulse + mono-uppercase label (`INDEXING…`) or the Thread drawing. No `<Spinner />` exists in the codebase. |
| **Empty state whispers "Discovering peers on your LAN…"** | APPFLOW §6.10: empty teaches (1) what this area is, (2) why it might be empty, (3) the one action that fills it. The whisper + Refresh button covers all three. |
| **mDNS discovery is the only "next action"** | DIRECTION §2 architecture-is-the-skip: there's no "Add device manually" button (LAN discovery is the path; LAN-not-available falls into the same empty whisper). |

---

## 8. DRIFT TABLE

What changed between the spec (APPFLOW §4.4, the original brief) and the implementation in `Sync.svelte`. Every row below is the kind of drift that must be reconciled in a single commit when the doc and the code disagree.

| ID | Spec said | Code does | Status | Resolution |
|---|---|---|---|---|
| D-S-01 | "PairingModal.svelte opens as a sheet from the right" (older PairingModal.svelte in `app/web/frontend/src/lib/components/`) | Pending card is embedded in `Sync.svelte`, centred, not a sheet | **Kept** | PairingModal.svelte remains as the floating interview variant; Sync.svelte owns the in-route card per APPFLOW §4.4 |
| D-S-02 | QR face uses `<Glyph name="sync">` fallback | Glyph fallback renders during identity load (status null) | **Kept** | Truthful: until identity arrives, the centre shows the brand mark, not a fake QR |
| D-S-03 | APPFLOW §4.4 says "thread breathes at 9 s" | 9 s `ease-in-out` loop on `opacity` + `filter: blur(3px)` glow at 0.12–0.34 opacity | **Kept** | Matches |
| D-S-04 | APPFLOW §4.4 says "revoke frays the thread" | `sync-fray` keyframe: `dasharray 0.18 0.16` + `opacity 0` + `dashoffset 1` over 900 ms | **Kept** | Matches |
| D-S-05 | SPEC said "Pair another" affordance in empty state | Empty state shows Refresh button + the bottom whisper; "Pair another" lives implicitly via the Peers rail | **Kept** | The Peers rail + Refresh IS the "Pair another" affordance — peer discovery is continuous |
| D-S-06 | SPEC did not specify a status pill | Header badge added: Pulse + mono label (`Sync paused` / `Awaiting a peer` / `N paired`) | **Added** | New: derives from `!status?.running ? 'error' : pairs.length === 0 ? 'thinking' : 'idle'` |
| D-S-07 | SPEC did not specify a TTL countdown ring | Added: 132 × 132 SVG ring around the PIN, pollen stroke, dashoffset transitions over 250 ms | **Added** | New: signals PIN expiry without an extra text line; reads at a glance |
| D-S-08 | SPEC did not specify node positioning formula | `RADIUS = 34%`, angle = `(i/n)*2π − π/2` (12 o'clock first); `n == 1` → fixed top | **Added** | New: keeps labels inside the canvas; radial spread is even |
| D-S-09 | SPEC said "pin 4–8 digits" | `/^\d{4,8}$/.test(pinInput.trim())`; `inputmode="numeric"`, `maxlength="8"` | **Kept** | Matches |
| D-S-10 | SPEC did not mention `err-hair` | Added: 600 ms `scaleX 0→1` hairline in the error block, 120 ms delay | **Added** | New: per MOAT §3 (err-hair is the error variant of the Thread) |
| D-S-11 | SPEC did not specify reduced-motion behaviour | Per-component `@media (prefers-reduced-motion: reduce)` neutralises thread animations + node halo; lives in `Sync.svelte` style block | **Drift** | **TODO: remove.** Per MOAT §2.3, the single block in `condura.css` owns reduced-motion. Sync.svelte must not redeclare. |
| D-S-12 | SPEC said polling `sync.status()` every 5 s | `setInterval(refreshAll, 5000)` calls `ipc.syncStatus()` + `sync.refresh()` together | **Kept** | Matches |
| D-S-13 | SPEC said "PIN input is 4–8 digits" but did not specify inputmode | `inputmode="numeric"`, `placeholder="000000"`, mono 18 px, +0.18em tracking | **Kept** | New detail |
| D-S-14 | SPEC did not specify a "Cancel pairing" link | Added: bottom-of-card ghost-style link, mono-uppercase, calls `cancelPending()` | **Added** | New: the user must always be able to back out (CLAUDE.md §2.1 invariant #4) |
| D-S-15 | SPEC said "no central server" | Footer whisper + lead copy state this explicitly; no third-party SDKs imported (no firebase, no segment, no posthog, no axios to a remote host) | **Kept** | Matches; this is a load-bearing claim |
| D-S-16 | SPEC said "PIN is generated server-side" | PIN is generated by the daemon (`sync.pendingPin`), arrives via SSE/poll, expires via `pendingExpiresAt` | **Kept** | Matches |
| D-S-17 | SPEC did not specify the revoke-back behaviour | On `sync.revoke()` failure, `frayedIds` is rolled back so the thread re-renders | **Added** | New: user can retry without re-loading the route |
| D-S-18 | MOAT §2.8 says three overlay primitives (modal / sheet / popover) | Revoke popover uses inline `.revoke-pop` + `.pop-backdrop`, not the future `.c-popover` primitive | **Drift** | **TODO:** extract `.c-popover` to `condura.css` per MOAT §2.8; collapse this inline pattern |
| D-S-19 | DIRECTION §5 reduced-motion contract is owned by `condura.css` only | `Sync.svelte:1170–1181` redeclares its own reduced-motion block (sets `stroke-dashoffset: 0`, `opacity: 0.7`, `animation: none`) | **Drift** | **TODO: remove.** The `condura.css` global rule already neutralises durations; this block is redundant and violates the rule. |
| D-S-20 | DIRECTION §5 says "components do not declare their own cubic-beziers" | No new beziers declared in Sync.svelte | **Kept** | Matches |
| D-S-21 | DIRECTION §6 Rule 2: no emoji | All icons are `<Glyph>` (sync, trash, refresh) — no emoji anywhere | **Kept** | Matches |
| D-S-22 | DIRECTION §6 Rule 1: no gradient text | All text is single-color per role | **Kept** | Matches |
| D-S-23 | DIRECTION §6 Rule 7: loading states teach, not spin | All loading states pair Pulse + mono-uppercase label (`Awaiting a peer`, `Refreshing…`, `Sealing…`) | **Kept** | Matches |
| D-S-24 | MOAT §4 #9: no double shadows | `.revoke-pop` uses `--shadow-float` only; `.pending-card` uses `--shadow-float` only | **Kept** | Matches |
| D-S-25 | DIRECTION §3: italic accent reserved for one load-bearing phrase per surface | `.alive` appears once in the headline ("threaded.") — the single italic accent | **Kept** | Matches |

**Net drift to close:** D-S-11, D-S-19 (reduced-motion duplication), D-S-18 (extract `.c-popover` primitive).