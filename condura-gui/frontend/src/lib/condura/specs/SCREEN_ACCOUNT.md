# SCREEN_ACCOUNT.md — Condura Account · Screen Architecture

> **The contract for the Account route surface.** This is the screen for the
> optional Synaptic account — the one that unlocks Hub publishing, donations,
> and support tickets. **It is not required to use Condura.** The local agent
> works without it; P2P sync works without it. The screen exists to make this
> honest.
>
> **Reading order for the next agent.** Read §1 (Drift) — what changes. Read
> §2 (Layout) — the geometry. Read §3 (State Matrix) — every state is
> reachable. Read §4 (Motion) — the verb behind every animation. Skip §5–§9
> only if you already know the app.
>
> **Source-of-truth docs** (read before this spec):
> - `MOAT.md` — quality bar (§1 restraint, §2 detail, §3 thread, §4 anti-patterns).
> - `APPFLOW.md` §3.2, §4, §6.9 — Account in the navigation graph.
> - `TEARDOWN.md` §2, §4 — mature 2026 settings-as-document and the
>   "no forced account" pattern from Linear / Notion / Manus.
>
> **Existing implementation.** `Settings.svelte:756–780` (current Account
> chip / sign-in link inside Settings); `components/SignInPanel.svelte`
> (the legacy modal-based OAuth panel — to be deprecated); `components/
> AccountMenu.svelte` (legacy sidebar dropdown menu — to be deprecated);
> `stores/account.svelte.ts` (the data store; do not break its public
> surface). This spec describes the new full-surface Account route at
> `#/account`, replacing both.

---

## Table of Contents

1. [Spec vs. Implementation Drift](#1-spec-vs-implementation-drift)
2. [Layout](#2-layout)
3. [State Matrix](#3-state-matrix)
4. [Motion Choreography](#4-motion-choreography)
5. [Keyboard](#5-keyboard)
6. [Components Used](#6-components-used)
7. [Data Fetched](#7-data-fetched)
8. [Design Decisions (MOAT Compliance)](#8-design-decisions-moat-compliance)
9. [Accessibility Contract](#9-accessibility-contract)
10. [Implementation Notes](#10-implementation-notes)
11. [Test Plan](#11-test-plan)

---

## 1. Spec vs. Implementation Drift

What this spec changes in the current code. Phase 4 must remove the old,
ship the new, in one atomic diff — no half-states.

| # | Today | Phase 4 (this spec) | Why |
|---|---|---|---|
| 1 | Account is a 3-line section inside `Settings.svelte:756–780` — a chip when signed in, an inline "sign in" link when not. | **Account is its own route at `#/account`** (NavRail item 9, replaces nothing — it's added). The Settings page loses the Account section. | APPFLOW §3.2 — Account earns its own surface; one route per concern. |
| 2 | `components/SignInPanel.svelte` is a modal `<Dialog>` with three tabs (`signin / signup / magic`) and three OAuth providers (Google / GitHub / Apple). | **No modal. The Account route IS the panel.** Single visible state. Tabs collapse; OAuth rows are always visible. | MOAT §4 #1 restraint — no nested modals; route enter is the surface change. |
| 3 | `components/AccountMenu.svelte` is a sidebar dropdown with a 2-step confirm popover for sign-out. | **No sidebar dropdown.** The Account route hosts a confirm popover on the route itself. | APPFLOW §3.2 — the rail has 10 items; duplicating account there adds nothing. |
| 4 | Three OAuth providers wired (Google / GitHub / Apple). | **Two OAuth providers (Google / GitHub) + Email magic-link.** Apple deferred to v0.2.0 per `MOAT §1.6`. | MOAT §1 restraint — ship two, not three. Apple in `SignInPanel.svelte` keeps its wiring in v0.2.0. |
| 5 | Hero copy: "Not signed in. The agent works without an account — sign in to sync skills and hub bookmarks." | **Hero copy: "A Condura account is optional."** Body: "The local agent works without an account. An account unlocks the public Skills Hub, donations, and support." Footer note always renders. | APPFLOW §2.3.8 + MOAT §4 #5 — never fake enthusiasm; never imply necessity. |
| 6 | Tier badge is absent. | **Tier badge** ("free" / "pro" / "team") next to provider when signed in. | The spec calls it out (per task brief). |
| 7 | "Manage account" link opens a new tab. | **"Manage on condura.app →"** is a threadlink (`.thread-link`) that deep-links to `https://condura.app/account` via the system browser. | Consistent with the channel audit-link pattern (`APPFLOW §4.7`). |
| 8 | Sign-out is a two-step confirm inside the dropdown. | **Sign-out is a confirm popover** (`.c-popover`) anchored to the sign-out button, two buttons ("Cancel" / "Sign out"). | MOAT §2.8 — three named primitives; popover is one. |
| 9 | Magic-link row is a single field + button stacked, no Pulse / no pulse-per-row feedback. | **AuthPicker rows each carry a Pulse in `idle` phase** when unchosen. The chosen row morphs its icon into a `Pulse phase="thinking"` during the round-trip. | MOAT §3 (signature is the pulse / thread; loading isn't a spinner). |
| 10 | Empty / error states show generic messages ("The magic link didn't go through"). | **Guide-error** with three lines: what failed / likely cause / next action (per `MOAT §2.6`). | MOAT §2.6 — error states must guide, not poeticize. |
| 11 | No footer note about P2P sync. | **Always-on footer note:** "Your account is only for Hub + donations + support. Sync is P2P — no account needed." | Locked decision §30 (CLAUDE.md). Account surface must be honest about what it's *not*. |

---

## 2. Layout

### 2.1 Geometry

The Account route occupies the **Main region** of the Shell (per
`SCREEN_SHELL.md §1.1`). It does **not** use the right rail; the surface is
document-style, single-column, centered. Width is constrained so the hero
copy reads at a comfortable line-length; on ≥1440px the column is 560px
wide, on 1024–1439px it fills the column, on <1024px it falls back to a
full-bleed layout with comfortable padding.

```
Shell · Account route · signed-out
┌──────────────────────────────────────────────────────────────┐
│  Titlebar (44px)                                             │
│ ┌──────────┬───────────────────────────────────────┬────────┐ │
│ │          │                                       │        │ │
│ │ NavRail  │   (1) Hero — eyebrow + headline + body│        │ │
│ │  240px   │                                       │        │ │
│ │          │   (2) What an account unlocks — 3 rows│        │ │
│ │  Chat    │                                       │        │ │
│ │  Hub     │   (3) AuthPicker — 3 auth rows        │        │ │
│ │  Skills  │       Google · GitHub · Email         │        │ │
│ │  Sync    │                                       │        │ │
│ │  Audit   │   (4) Footer note — P2P clarification │        │ │
│ │  Replay  │                                       │        │ │
│ │  Chnls   │                                       │        │ │
│ │  Deleg   │                                       │        │ │
│ │  Settngs │                                       │        │ │
│ │ •Account │                                       │        │ │
│ │  About   │                                       │        │ │
│ │          │                                       │        │ │
│ └──────────┴───────────────────────────────────────┴────────┘ │
│  StatusBar (28px)                                            │
└──────────────────────────────────────────────────────────────┘
```

```
Shell · Account route · signed-in
┌──────────────────────────────────────────────────────────────┐
│  Titlebar (44px)                                             │
│ ┌──────────┬───────────────────────────────────────┬────────┐ │
│ │          │                                       │        │ │
│ │ NavRail  │   (1) AccountCard                      │        │ │
│ │  240px   │       Avatar + name + email + provider│        │ │
│ │          │       Tier badge                      │        │ │
│ │          │                                       │        │ │
│ │          │   (2) What this account unlocks — 3   │        │ │
│ │          │       rows                            │        │ │
│ │          │                                       │        │ │
│ │          │   (3) Threadlink — "Manage on         │        │ │
│ │          │       condura.app →"                  │        │ │
│ │          │                                       │        │ │
│ │          │   (4) Sign-out button + popover       │        │ │
│ │          │                                       │        │ │
│ │          │   (5) Footer note — P2P clarification │        │ │
│ │          │                                       │        │ │
│ └──────────┴───────────────────────────────────────┴────────┘ │
└──────────────────────────────────────────────────────────────┘
```

### 2.2 The Column

The Account surface is a **flowing column**, not a modal, not a card. There
is no card chrome — only hairline dividers (`--hair`) between sections.
This matches the "Settings as document" pattern (`APPFLOW §5`).

```
┌─────────────────────────────────────┐
│  Eyebrow   (mono, 11px, --content-2) │
│                                     │
│  Headline  (serif, 28px, --content) │ ← "A Condura account is optional."
│                                     │
│  Body      (sans, 15px, --content-1)│ ← one-line descender
│                                     │
│  ─────                              │ ← hairline, 1px, --hair
│                                     │
│  What an account unlocks            │ ← section eyebrow
│                                     │
│   ⌬  Publish to the Skills Hub      │ ← list, glyph + body
│   ⌬  Support the project            │
│   ⌬  Open a support ticket          │
│                                     │
│  ─────                              │
│                                     │
│  Choose a method                    │ ← section eyebrow (signed-out only)
│                                     │
│  ┌───────────────────────────────┐  │
│  │ ◉  Continue with Google       │  │ ← AuthPicker row 1
│  └───────────────────────────────┘  │
│  ┌───────────────────────────────┐  │
│  │ ◉  Continue with GitHub       │  │ ← AuthPicker row 2
│  └───────────────────────────────┘  │
│  ┌───────────────────────────────┐  │
│  │ ◉  Email me a magic link      │  │ ← AuthPicker row 3
│  └───────────────────────────────┘  │
│                                     │
│  ─────                              │
│                                     │
│  Footer note (mono, 11px, --content-2)
│   "Your account is only for Hub +   │
│    donations + support. Sync is     │
│    P2P — no account needed."        │
└─────────────────────────────────────┘
```

### 2.3 The AccountCard (signed-in)

When signed in, the surface replaces the AuthPicker with an AccountCard
and re-flows:

```
┌─────────────────────────────────────┐
│  ┌────┐                             │
│  │ AV │  Alex Rivera                │ ← Avatar 36px, name 17px serif
│  └────┘  alex@example.com           │ ← email 13px mono
│           Via Google · pro          │ ← provider + tier, 11px mono
│                                     │
│  ─────                              │
│                                     │
│  What this account unlocks          │
│   ⌬  Publish to the Skills Hub      │
│   ⌬  Support the project            │
│   ⌬  Open a support ticket          │
│                                     │
│  ─────                              │
│                                     │
│  Manage on condura.app →            │ ← threadlink
│                                     │
│  Sign out                           │ ← Button variant="ghost" size="sm"
│                                     │
│  ─────                              │
│                                     │
│  Footer note                        │
└─────────────────────────────────────┘
```

### 2.4 Vertical rhythm

| Element | Type | Size | Weight | Color | Notes |
|---|---|---|---|---|---|
| Eyebrow | mono | 11px | 500 (medium) | `--content-2` | UPPERCASE, `letter-spacing: 0.12em` |
| Headline | serif italic | 28px | 400 | `--content` | Italic per design system; "italic is the seal" |
| Body | sans | 15px | 400 | `--content-1` | Line-height 1.55 |
| Section eyebrow | mono | 11px | 500 | `--content-2` | UPPERCASE |
| List item | sans | 14px | 400 | `--content` | Glyph 16px in `--content-2` |
| Footer note | mono | 11px | 400 | `--content-2` | Lowercase |
| AuthPicker row | sans | 14px | 500 | `--content` | 44px tall (touch target) |

Vertical spacing uses `--space-7` (32px) between hero and first section,
`--space-5` (20px) between sections, `--space-3` (12px) between list
items.

### 2.5 Width & breakpoints

| Breakpoint | Column behavior |
|---|---|
| ≥ 1440 px | Main column 1fr; Account content centered, max-width 560px |
| 1024–1439 px | Main column 1fr; Account content centered, max-width 560px, padded `--space-7` left/right |
| 768–1023 px | Main column 1fr; NavRail condensed to 56px (icons only); Account content fills, padded `--space-5` |
| < 768 px | NavRail as drawer; Account content full-bleed with `--space-5` padding; all touch targets ≥ 44px |

The 560px cap on the content column keeps the hero copy readable (≈66
characters at 15px). This is the same column width used in `Settings.svelte`
for the autonomy matrix.

---

## 3. State Matrix

The Account surface has **five primary states**, derived from the brief
plus the standard state vocabulary (`APPFLOW §7`):

| # | State | Visual | Trigger | Reachable? |
|---|---|---|---|---|
| S1 | **signed-out** | Hero + AuthPicker + footer | First load; default for new users | ✅ |
| S2 | **signing-in** | Hero dimmed, AuthPicker rows show pulse-per-row, chosen row's icon morphs into `Pulse phase="thinking"` | User clicks an AuthPicker row | ✅ |
| S3 | **signed-in** | AccountCard + list + threadlink + Sign out + footer | OAuth callback resolves; magic link click → `account.handleCallback` returns success | ✅ |
| S4 | **error** | Guide-error block above AuthPicker; the failed row retains its icon (no morph back); other rows remain interactive | RPC rejects (network, daemon down, OAuth state mismatch, email rejected) | ✅ |
| S5 | **signing-out** | Sign-out button shows confirm popover (`.c-popover`); on confirm, card fades out, AuthPicker fades in over 320ms | User clicks Sign out, then confirms | ✅ |

### 3.1 Signed-out (default — S1)

**Visual:**
- Hero: eyebrow "ACCOUNT" → headline "A Condura account is optional." →
  body "The local agent works without one. An account unlocks the public
  Skills Hub, donations, and support."
- Hairline.
- Section eyebrow "WHAT AN ACCOUNT UNLOCKS".
- Three list rows, each a `<Glyph name="...">` + body copy:
  - `Glyph name="hub"` (or `bookmark` if `hub` is taken) — "Publish to the public Skills Hub"
  - `Glyph name="heart"` (or `donate` if exists) — "Support the project on GitHub Sponsors, Open Collective, or Stripe"
  - `Glyph name="lifebuoy"` (or `mail`) — "Open a support ticket via email or Discord"
- Hairline.
- Section eyebrow "CHOOSE A METHOD" (signed-out only).
- `<AuthPicker>` — three rows: Google, GitHub, Email.
- Hairline.
- Footer note: "Your account is only for Hub + donations + support. Sync is P2P — no account needed."

**Interactions:**
- Tab reaches AuthPicker rows in order.
- Enter on a row → starts the corresponding sign-in flow (state → S2).
- Esc — no-op (no overlay open).

### 3.2 Signing-in (S2)

**Visual:**
- Hero remains visible at 100% opacity (per the brief — the auth picker
  is the thing dimming).
- AuthPicker rows each show a `<Pulse phase="idle" size=6>` on the right
  edge (each row has its own pulse, breathing at the row's cadence — not
  a single shared pulse).
- The chosen row's leading icon **morphs** into a `<Pulse phase="thinking"
  size=8>` (icon scales to 0 + opacity 0, pulse scales from 0.4 → 1
  with fade-in over 200ms).
- The **other rows dim to 50% opacity** over `--dur` (280ms) — they
  remain in the tab order but visually recede.
- A `<Thread>` hairline draws in below the chosen row over `--dur-slow`
  (520ms), in `--synapse` (the "flowing connection" — see §8.1).

**Sub-states by row:**
- **Google row chosen:** opens the system browser at the returned
  `account.oauth_url` URL (via `runtime.BrowserOpenURL`, falling back to
  `window.open`); state stays S2 with a `synapse-glow` pulse until the
  callback resolves.
- **GitHub row chosen:** same as Google, different URL.
- **Email row chosen:** the Email row expands inline (height 44px →
  96px) over 240ms with an email input field + a "Send link" button. The
  pulse continues until the RPC returns (success → S3 magic-link-pending;
  error → S4).

**Error path from S2:** if the RPC rejects (e.g., daemon down, OAuth
state mismatch), the chosen row's pulse morphs back to its icon over
280ms, the guide-error block appears above the AuthPicker, and the other
rows return to 100% opacity. State → S4.

### 3.3 Signed-in (S3)

**Visual:**
- The AuthPicker **fades out** (opacity 1 → 0 over `--dur`); the
  AccountCard **slides in from the right** (translateX(16px → 0) over
  `--dur-slow`, opacity 0 → 1).
- The hero copy persists but the eyebrow updates to "ACCOUNT · SIGNED IN".
- `<AccountCard>` renders:
  - `<Avatar size="md">` (36px, circle) on the left. Fallback (no
    `account.avatarURL`) is a single-letter monogram in
    `--synapse-glow` on a paper-tinted background.
  - Name (serif, 17px) + email (mono, 13px) + "Via {provider} · {tier}"
    (mono, 11px) stacked right of the avatar.
- Below the card: "WHAT THIS ACCOUNT UNLOCKS" list (same three rows as
  S1, but now they read as enabled rather than offered).
- Hairline.
- `<Threadlink>` "Manage on condura.app →" — opens the system browser at
  `https://condura.app/account` (consistent with the channel audit-link
  pattern).
- `<Button variant="ghost" size="sm">` "Sign out" — anchored right.
- Footer note (always present).

**Transitions out:**
- Sign out click → state S5 (signing-out).
- Network drop → state S4 (signed-out, daemon down). The AccountCard
  fades out and the AuthPicker fades in.

### 3.4 Error (S4)

**Visual:**
- A guide-error block (`.err-state` block, italic Instrument Serif
  headline 18px + 14px sans body) renders **above** the AuthPicker (in
  signed-out) or above the AccountCard (in signed-out → signed-in
  transition if the callback fails).
- Headline: what failed (one noun). Body: likely cause + next action.

**Exact copy per error source:**

| Source | Headline | Body |
|---|---|---|
| Daemon unreachable | "We couldn't reach the daemon." | "Condura is running locally. If this keeps happening, restart Condura from the menu bar." |
| OAuth state mismatch | "Sign-in was interrupted." | "We didn't recognize the return request. Try again — if it persists, clear your browser cookies for condura.app." |
| Email rejected | "Email didn't go through." | "Check the address — or try Google or GitHub instead." |
| Email rate-limited | "Too many tries." | "Wait a minute, then try again. Or use Google or GitHub." |
| Token exchange failed | "Provider rejected the sign-in." | "The token exchange didn't complete. Try another method, or check Discord for known outages." |

**Interactions:** all rows remain clickable; the guide-error stays
visible until the user clicks another row (state → S2 again) or the next
sign-in succeeds (state → S3).

### 3.5 Signing-out (S5)

**Visual:**
- User clicks the "Sign out" button → a `.c-popover` anchored to the
  button fades in (opacity 0 → 1) and scales (scale 0.96 → 1) over 200ms.
- Popover contents:
  - Body: "Sign out of alex@example.com?" (uses the account email,
    per the MOAT).
  - Two buttons: "Cancel" (ghost) and "Sign out" (danger). Cancel
    closes the popover; Sign out calls `account.signOut()`.
- During the `account.signOut()` RPC, the Sign out button shows
  `<Pulse phase="acting" size=6>` next to "Signing out…".
- On success: the AccountCard fades out (opacity 1 → 0 over `--dur`) and
  the AuthPicker fades in (opacity 0 → 1 over `--dur`) simultaneously.
  The eyebrow reverts to "ACCOUNT" (not "ACCOUNT · SIGNED IN").
- On failure: popover stays open, inline error "Sign out didn't
  complete" appears in the popover; the user can retry.

### 3.6 Sign-in flow sub-state — magic-link sent

When the user picks Email, a sub-state is introduced **before** S3:

| State | Visual | Trigger |
|---|---|---|
| **S2.e** — email entered | Email row expanded with field + "Send link" button | User clicks Email row |
| **S2.e.sent** — link sent | Field becomes read-only mono text "Link sent to alex@example.com · check spam, or use another method →" + a `<Pulse phase="ok" size=6>` | `account.signInWithEmail` returns `sent: true` |
| **S2.e.err** | Same field + button, inline error from §3.4 | RPC rejects |

The `S2.e.sent` state is persistent until the user clicks the row again
(re-arming for a new email) or the magic link callback resolves (→ S3).

---

## 4. Motion Choreography

The motion grammar is the **MOAT.md §3 thread** + the existing
Pulse-based vocabulary. Every animation answers "what just happened?"

### 4.1 Enter (mount from route)

The Account surface is mounted via `Shell.svelte`'s `{#key route}` +
`.route-enter` (per `APPFLOW §3.2`). The route-level enter is a 280ms
opacity 0 → 1 + 8px blur-in.

Inside the surface, on first mount in S1 (signed-out):

| Element | Animation | Duration | Easing | Notes |
|---|---|---|---|---|
| Hero (eyebrow → headline → body) | opacity 0 → 1 | 280ms | `--ease` | Single fade; no slide |
| Hairline below hero | `stroke-dashoffset` 1 → 0 | `--dur-slow` (520ms) | `--ease` | The signature thread |
| Section eyebrow "WHAT AN ACCOUNT UNLOCKS" | opacity 0 → 1 | 280ms, **120ms delay** | `--ease` | After hero |
| List rows (3) | opacity 0 → 1 + translateY(4px → 0) | 240ms each | `--ease` | **Stagger 80ms apart** |
| Hairline below list | `stroke-dashoffset` 1 → 0 | `--dur-slow` | `--ease` | After list |
| Section eyebrow "CHOOSE A METHOD" | opacity 0 → 1 | 280ms, **120ms delay** | `--ease` | |
| AuthPicker rows (3) | opacity 0 → 1 + translateY(4px → 0) | 240ms each | `--ease` | **Stagger 80ms apart** |
| Hairline below picker | `stroke-dashoffset` 1 → 0 | `--dur-slow` | `--ease` | |
| Footer note | opacity 0 → 1 | 280ms | `--ease` | Last |

**Total enter: ~1.6s** (longest path). On `prefers-reduced-motion: reduce`,
all durations drop to 0 and the stagger collapses — the surface appears
instantly.

### 4.2 Signing-in (S1 → S2)

Triggered by row click.

| Element | Animation | Duration | Easing |
|---|---|---|---|
| Chosen row icon | scale 1 → 0 + opacity 1 → 0 | 200ms | `--ease-in` |
| Chosen row, pulse appearing | scale 0.4 → 1 + opacity 0 → 1 | 240ms, **+40ms delay** | `--ease` |
| Other rows (2) | opacity 1 → 0.5 | `--dur` (280ms) | `--ease` |
| Thread under chosen row | `stroke-dashoffset` 1 → 0 | `--dur-slow` (520ms) | `--ease` |

### 4.3 Signed-in (S2 → S3)

Triggered by successful callback resolution.

| Element | Animation | Duration | Easing |
|---|---|---|---|
| AuthPicker (whole block) | opacity 1 → 0 | `--dur` (280ms) | `--ease` |
| Hero eyebrow text | swap "ACCOUNT" → "ACCOUNT · SIGNED IN" | instant (text swap) | — |
| AccountCard | translateX(16px → 0) + opacity 0 → 1 | `--dur-slow` (520ms) | `--ease` |
| Avatar inside card | scale 0.9 → 1 + opacity 0 → 1 | 200ms, **+200ms delay** | `--ease` |
| Card meta (name, email, provider) | opacity 0 → 1 | 240ms, **+260ms delay** | `--ease` |
| Hairline below card | `stroke-dashoffset` 1 → 0 | `--dur-slow` | `--ease` |
| List rows (3, "what this unlocks") | opacity 0 → 1 + translateY(4px → 0) | 240ms each | `--ease` | Stagger 80ms |
| Threadlink | opacity 0 → 1 | 240ms, **+120ms delay** | `--ease` |
| Sign out button | opacity 0 → 1 | 240ms, **+160ms delay** | `--ease` |
| Footer note | opacity 0 → 1 | 240ms | `--ease` |

### 4.4 Error (S1/S2 → S4)

| Element | Animation | Duration | Easing |
|---|---|---|---|
| Guide-error block | opacity 0 → 1 + translateY(4px → 0) | `--dur` (280ms) | `--ease` |
| Chosen row pulse (if any) | scale 1 → 0 + opacity 1 → 0 | 200ms | `--ease-in` |
| Chosen row icon (returns) | scale 0 → 1 + opacity 0 → 1 | 240ms, **+40ms delay** | `--ease` |
| Other rows | opacity 0.5 → 1 | `--dur` | `--ease` |
| Thread under chosen row | `stroke-dashoffset` 0 → 1 (draw out) | `--dur-slow` | `--ease` |

### 4.5 Signing-out (S3 → S5)

| Element | Animation | Duration | Easing |
|---|---|---|---|
| Confirm popover | opacity 0 → 1 + scale 0.96 → 1 | 200ms | `--ease` |
| "Cancel" / "Sign out" buttons | opacity 0 → 1 | 160ms, **+80ms delay** | `--ease` |

On confirm:

| Element | Animation | Duration | Easing |
|---|---|---|---|
| AccountCard | opacity 1 → 0 + translateX(0 → -16px) | `--dur` (280ms) | `--ease-in` |
| AuthPicker | opacity 0 → 1 + translateY(4px → 0) | `--dur`, **+200ms delay** | `--ease` |
| Hero eyebrow text | swap "ACCOUNT · SIGNED IN" → "ACCOUNT" | instant (text swap) | — |

### 4.6 Press states

Per `MOAT §2.2`, press states need weight — not just shrinkage. Applied
to the AuthPicker rows + Sign out button:

```css
.auth-row:active,
.signout-btn:active {
  transform: translateY(0.5px) scale(0.985);
  filter: brightness(0.96) saturate(1.05);
  transition: transform var(--dur-fast) var(--ease), filter var(--dur-fast) var(--ease);
}
```

### 4.7 Focus state

Per `MOAT §2.1`, focus rings track rounded shapes:

```css
.auth-row:focus-visible,
.signout-btn:focus-visible {
  outline: none; /* remove default */
  box-shadow:
    0 0 0 2px var(--synapse),
    0 0 0 5px var(--pollen-halo);
  border-radius: var(--r-md); /* matches row radius */
}
```

### 4.8 Reduced motion

A single global override (`MOAT §2.3`) handles all of the above. The
component does not redeclare any `@media (prefers-reduced-motion)`
blocks. All `translateY`, `scale`, and `opacity` transitions drop to 0ms;
the stagger collapses; the thread draws instantly (no draw-in).

---

## 5. Keyboard

### 5.1 Tab order

When the Account route is active and signed-out (S1):

```
1. AuthPicker row — Google
2. AuthPicker row — GitHub
3. AuthPicker row — Email
   3a. (when Email expanded) Email input field
   3b. (when Email expanded) Send link button
4. Footer note link (if any — currently none, but reserved for future)
```

When signed-in (S3):

```
1. Threadlink — "Manage on condura.app →"
2. Sign out button
```

### 5.2 Keys

| Key | State | Action |
|---|---|---|
| `Tab` / `Shift+Tab` | all | Move focus per §5.1 |
| `Enter` | AuthPicker row | Start that provider's sign-in flow (→ S2) |
| `Enter` | Send link button (S2.e) | Call `account.signInWithEmail` |
| `Enter` | Threadlink (S3) | Open `https://condura.app/account` in system browser |
| `Enter` | Sign out (S3) | Open confirm popover (→ S5) |
| `Space` | AuthPicker row | Same as Enter |
| `Escape` | S5 (popover open) | Close popover, focus returns to Sign out |
| `Escape` | S2.e (email expanded) | Collapse email row, focus returns to AuthPicker Email row |
| `Escape` | S2 (chosen row) | Cancel pending sign-in? — NO. The OAuth round-trip is already in flight; cancelling it server-side requires a separate RPC. For v0.1.0, Esc during S2 closes nothing. |
| `Cmd+,` | any | Open Settings (shell-level shortcut per `MOAT §2.10`) |
| `⌘K` | any | Open Command Palette (shell-level) |

### 5.3 Focus management

- On route enter, focus stays where it was (no auto-focus — the brief
  does not require it and auto-focus on auth forms is a "vibe-coded"
  smell).
- When the AuthPicker → AccountCard transition completes (S2 → S3),
  focus moves to the first focusable element in the AccountCard (the
  Threadlink). This is announced via `aria-live="polite"`.
- When the popover opens (S3 → S5), focus moves to the first button
  (Cancel). Closing the popover returns focus to the Sign out button.
- When the Email row expands (S1 → S2.e), focus moves to the email input.
- On error (any → S4), focus does not move (the guide-error appears
  above; focus stays on the chosen row so the user can retry).

### 5.4 Screen reader announcements

- The route mount announces via `aria-live="polite"`: "Account. A
  Condura account is optional." (or "Account. Signed in as alex@example.com."
  when S3.)
- The signing-in transition (S1 → S2) announces: "Signing in with
  Google." (or GitHub / Email.)
- The signing-in → signed-in transition (S2 → S3) announces:
  "Signed in as alex@example.com."
- The guide-error (→ S4) announces via `role="alert"` (assertive): the
  full error text.
- The signing-out popover (S3 → S5) announces: "Confirm sign out for
  alex@example.com."

---

## 6. Components Used

| Component | Role | Notes |
|---|---|---|
| `<RouteHero>` | Eyebrow + headline + body block | Replaces 6 hand-rolled hero blocks per `MOAT §1.3`. The eyebrow prop accepts "ACCOUNT" or "ACCOUNT · SIGNED IN". |
| `<AuthPicker>` | Three-row OAuth + magic-link picker | New component (does not exist today). Built per this spec. Each row: icon slot + label slot + right-slot for the pulse. Renders three rows by default (Google / GitHub / Email); the brief specifies the magic-link row lives in the picker, not in a separate sub-panel. |
| `<AccountCard>` | Signed-in account summary card | New component (does not exist today). Avatar + name + email + provider + tier badge. Replaces the inline `.account-chip` in `Settings.svelte`. |
| `<Avatar>` | 36px circular avatar | Existing — `components/ui/Avatar.svelte`. Falls back to monogram if `account.avatarURL` is empty. |
| `<Button>` | "Send link", "Sign out", "Cancel" | Existing — `components/ui/Button.svelte`. Variants: `primary`, `ghost`, `danger`. |
| `<Pulse>` | Idle (right-edge per row), `thinking` (chosen row during sign-in), `ok` (magic-link sent), `acting` (during sign-out RPC), `error` (when applicable) | Existing — `condura/Pulse.svelte`. |
| `<Thread>` | Hairline dividers between sections; sub-thread under the chosen AuthPicker row during S2 | Existing — `condura/Thread.svelte`. |
| `<Glyph>` | `account`, `hub`, `heart`, `lifebuoy`, `mail`, `google`, `github`, `sign-out` | Existing — `condura/Glyph.svelte` + `condura/icons.ts`. New icons to add: `google`, `github`, `sign-out`. `google` and `github` use their official mark paths (single-stroke, 1.5 weight) per `MOAT §4 #2` — but see §8.2 for the brand mark caveat. |
| `<Tooltip>` | "Manage on condura.app →" threadlink hint | Existing primitive per `MOAT §2.9`. Hover-delay 400ms. Tooltip text: "Opens in your browser". |
| `<Popover>` | Sign-out confirm popover | New primitive per `MOAT §2.8`. Anchored to the Sign out button. Backdrop-click closes; Esc closes; focus is trapped. |
| `<Threadlink>` | "Manage on condura.app →" | Existing pattern (`Channels.svelte:423–438`). Renders as italic mono text with a hairline under that draws in on hover. |
| `<ErrorState>` | Guide-error block (S4) | Existing primitive per `MOAT §2.6`. Props: `head`, `sub`, `action`. |
| `<Input>` | Email field inside the expanded Email row | Existing — `components/ui/Input.svelte`. |

---

## 7. Data Fetched

### 7.1 IPC methods

All account data flows through the existing `account` store
(`stores/account.svelte.ts`). The Account route reads from the store
and does **not** call IPC directly. On mount:

| Call | Purpose | Source |
|---|---|---|
| `account.checkStatus()` | Populates `status`, `isSignedIn`, `email`, `provider`, `tier`, `avatarURL`, `displayName`, `configuredProviders` | `stores/account.svelte.ts:133` |
| (on row click) `account.signInWithProvider({ provider, redirect_uri, scopes })` | Returns `{ url, state, code_verifier }` for the OAuth redirect | `stores/account.svelte.ts:163` |
| (on Email row click) `account.signInWithEmail(email, locale, redirect_url)` | Returns `{ sent }` or sets `account.error` | `stores/account.svelte.ts:221` |
| (on callback resolve) `account.handleCallback(code, state, redirectURI)` | Exchanges code for tokens; sets `status` | `stores/account.svelte.ts:252` |
| (on Sign out confirm) `account.signOut()` | Clears `status` | `stores/account.svelte.ts:302` |

### 7.2 OAuth redirect URIs

| Provider | Redirect URI |
|---|---|
| Google OAuth | `condura://auth/callback` (matches existing `SignInPanel.svelte:25`) |
| GitHub OAuth | `condura://auth/callback` |
| Magic link | `https://condura.app/auth/verify` (matches existing `SignInPanel.svelte:27`) |

### 7.3 State timing

| State | Wait time before timeout (UX fallback) | Real source |
|---|---|---|
| S1 → S2 (OAuth) | The OAuth round-trip has no client-side timeout — the user controls the flow from their browser. | n/a |
| S2 → S2.e.sent (magic link) | 8s; if `account.signInWithEmail` has not returned by then, show the inline error "Sending… is taking longer than expected. Check your connection, or try another method." | `account.signInWithEmail` resolve |
| S2 → S3 (callback) | The callback URL is `condura://auth/callback?...` — when the OS routes this back to the desktop app, `account.handleCallback` runs. | OS deep-link |
| S3 → S5 → S1 (sign out) | 5s; if `account.signOut` has not returned by then, show "Sign out is taking longer than expected. Try again, or restart Condura." | `account.signOut` resolve |

---

## 8. Design Decisions (MOAT Compliance)

### 8.1 Thread as the signature (MOAT §3)

Every state transition in the Account surface draws a thread. The
MOAT commits to one element. The Account surface does not deviate:

- Between hero and list: hairline divider.
- Between list and AuthPicker: hairline divider.
- Under the chosen AuthPicker row during S2: a sub-thread draws in (the
  "this is the one being acted on" gesture).
- Between AuthPicker and footer: hairline divider.

The thread is the **only** flourish for "this is now finished." If
another animation lands in this file that isn't a thread, the
PR is wrong.

### 8.2 The Glyph brand caveat (MOAT §4 #2 + #4)

Per `MOAT §4 #2`, "no emoji as UI icons" — use `<Glyph name=...>`. For
the Google and GitHub auth providers, the brand marks are commonly drawn
in their official colors (multi-color). Condura's brand is paper + ink +
synapse green. **The auth rows show monochrome single-stroke glyphs** (1.5
weight, currentColor), drawn to match the rest of the icon system. The
brand "color" is implicit in the row hover (the synapse halo) and the
post-sign-in pulse (synapse-green).

The icon paths come from each brand's published single-stroke variant.
If neither is available, fall back to text labels: "Continue with Google"
— **never** the colorful SVG logo (per `MOAT §4 #1`: no rainbow accents,
no gradient text).

### 8.3 Restraint (MOAT §1)

- **No carousel.** No 3 OAuth providers + Apple + Microsoft + Twitter +
  Discord. Two OAuth + Email is enough. Apple defers to v0.2.0.
- **No "Welcome to the community!" copy.** The hero says "A Condura
  account is optional." That is the entire headline.
- **No spinner.** Loading is a thread drawing in (`MOAT §4 #7`) or a
  pulse (`MOAT §2.5`).
- **No fake enthusiasm.** No "Awesome!" toasts. No "🎉 You're signed
  in!" modal. The card slides in; that's the celebration.

### 8.4 Honesty (MOAT §4 #5, #6 + CLAUDE.md locked decisions)

The footer note is the load-bearing copy of this screen. It exists
because the most common error in product onboarding is **implying an
account is required when it isn't.** Condura's account is genuinely
optional:

- The local agent runs without one.
- P2P sync works without one (device-to-device encrypted).
- Skills install locally without one.
- The Hub is the only thing that needs one.

The footer note says so, every time, no matter what state the screen is
in. The note is **always rendered** — never conditional on state, never
hidden behind a click.

### 8.5 Focus + press + tactile (MOAT §2.1, §2.2, §2.7)

- Focus rings track rounded shapes (`§2.1`).
- Press states add weight, not just shrinkage (`§2.2`).
- The `tactile` class is applied to AuthPicker rows and the Sign out
  button — components own meaning, not timing (`§2.7`).

### 8.6 No fake enthusiasm (MOAT §4 #6)

The Account surface has zero "Awesome!", "Great choice!", "You're all
set!" copy. Where the surface must register success (magic link sent),
the pulse and the mono text "Link sent to alex@example.com · check spam,
or use another method →" suffice. The user knows what just happened
because the visual structure changed.

### 8.7 Popover primitive (MOAT §2.8)

The sign-out confirm popover is a `.c-popover` — anchored, small,
dismiss-on-outside-click, dismiss-on-Esc. It is **not** a `.c-modal`
(which would block the page). It is **not** a `.c-sheet` (which slides
from an edge). It is a popover — and there is now exactly one popover in
the codebase (this one). Future popovers for other surfaces should
follow this pattern.

---

## 9. Accessibility Contract

| Surface | `role` | `aria-*` |
|---|---|---|
| Main region | `main` | `aria-labelledby="account-route-hero"` |
| Hero | `h1` (single h1 per route) | `id="account-route-hero"` |
| Section eyebrows | `h2` | `id="account-section-{id}"` |
| AuthPicker | `radiogroup` | `aria-labelledby="account-auth-picker-label"` |
| AuthPicker row | `radio` | `aria-checked`, `aria-busy` during S2 |
| Email input (S2.e) | `input[type="email"]` | `aria-required`, `aria-invalid` on format error |
| AccountCard | `region` | `aria-labelledby="account-card-name"` |
| Tier badge | `span` | `aria-label="Tier: pro"` (or whatever tier) |
| Threadlink | `a` | `aria-label="Manage your account on condura.app — opens in browser"` |
| Sign out button | `button` | `aria-haspopup="dialog"` |
| Confirm popover | `dialog` | `aria-modal="false"`, `aria-labelledby="signout-confirm-q"` |
| Guide-error block | `alert` | `aria-live="assertive"` |
| Footer note | `p` | `aria-label="Account scope note"` |

Color contrast: all text meets WCAG 2.2 AA (4.5:1 for body, 3:1 for
large text and UI). The `--content`, `--content-1`, and `--content-2`
tokens are calibrated for this (see `condura.css` light + dark blocks).

---

## 10. Implementation Notes

### 10.1 File layout

```
condura/
├── Account.svelte                  ← the route entry (replaces inline Settings.svelte:756-780)
├── components/
│   ├── AuthPicker.svelte           ← new (Google / GitHub / Email rows)
│   ├── AccountCard.svelte          ← new (signed-in summary card)
│   └── ...
├── Popover.svelte                  ← new primitive (per MOAT §2.8) — or pulled from a shared /ui/
└── specs/
    └── SCREEN_ACCOUNT.md           ← this file
```

The `Account.svelte` is the route entry, mounted by `Shell.svelte` at
`#/account` (per `APPFLOW §3.2`). It reads from `account` store, manages
local UI state (`viewingProvider`, `confirmingSignOut`, `email`),
and delegates rendering to `<AuthPicker>` and `<AccountCard>`.

### 10.2 Migration from inline + modal

The legacy surfaces must be removed in the same atomic diff:

1. `Settings.svelte:756–780` — delete the Account section + its styles.
2. `components/SignInPanel.svelte` — keep on disk for v0.2.0 (Apple
   support) but stop importing it from `Settings.svelte`.
3. `components/AccountMenu.svelte` — keep on disk for v0.2.0 (sidebar
   dropdown affordance) but stop importing it.
4. `NavRail.svelte` — add the Account item (per `SCREEN_NAVRAIL.md`
   drift #8, this spec assumes that change is already landed; if not,
   it's a one-line addition).

### 10.3 Edge cases

| Case | Behavior |
|---|---|
| Daemon down on mount | The route mounts; `account.checkStatus()` rejects → `account.error` set; the AuthPicker renders with a guide-error block above ("We couldn't reach the daemon"). All rows remain clickable; clicks re-attempt `checkStatus` first. |
| User signs in on the web dashboard, then opens desktop | `account.checkStatus()` resolves with `signed_in: true` → AccountCard renders. The web dashboard sign-in is a separate flow; desktop just reads the existing account. |
| User signs out on the web dashboard while desktop is open | `account.signOut` was called from the web; the desktop's `account.checkStatus()` will detect the change on its next refresh. **For v0.1.0, there is no SSE sync between web and desktop accounts.** The desktop surfaces the new state on next route enter or refresh. |
| Tier changes on the server (free → pro) | Same as above — no SSE in v0.1.0. The AccountCard shows the cached tier; the user can click "Manage on condura.app →" to refresh on the web. |
| User has Apple configured but Google + GitHub + Email aren't | Apple is not surfaced in v0.1.0 (per §8.3). The user is told in the next release note when Apple ships. |

### 10.4 i18n

All user-facing copy lives in `locales/<lang>.json` under `account.route.*`
keys. The English defaults:

```json
{
  "account.route.eyebrow": "Account",
  "account.route.signed_in_eyebrow": "Account · signed in",
  "account.route.headline": "A Condura account is optional.",
  "account.route.body": "The local agent works without one. An account unlocks the public Skills Hub, donations, and support.",
  "account.route.unlocks_heading": "What an account unlocks",
  "account.route.unlock_hub": "Publish to the public Skills Hub",
  "account.route.unlock_donate": "Support the project",
  "account.route.unlock_support": "Open a support ticket",
  "account.route.picker_heading": "Choose a method",
  "account.route.row_google": "Continue with Google",
  "account.route.row_github": "Continue with GitHub",
  "account.route.row_email": "Email me a magic link",
  "account.route.email_label": "Email address",
  "account.route.email_placeholder": "you@example.com",
  "account.route.send_link": "Send link",
  "account.route.sending": "Sending…",
  "account.route.link_sent": "Link sent to {email} · check spam, or use another method →",
  "account.route.footer_note": "Your account is only for Hub + donations + support. Sync is P2P — no account needed.",
  "account.route.manage_link": "Manage on condura.app →",
  "account.route.manage_tooltip": "Opens in your browser",
  "account.route.signout": "Sign out",
  "account.route.signout_confirm": "Sign out of {email}?",
  "account.route.cancel": "Cancel",
  "account.route.signing_out": "Signing out…",
  "account.route.tier_free": "free",
  "account.route.tier_pro": "pro",
  "account.route.tier_team": "team",
  "account.route.via": "Via {provider} · {tier}"
}
```

The "Via {provider}" sentence uses the localized provider label
(`Google`, `GitHub`, `Email`) — same mapping as the existing
`AccountMenu.svelte:24–32`.

---

## 11. Test Plan

### 11.1 Unit / component tests (vitest + @testing-library/svelte)

| Test | Asserts |
|---|---|
| Mounts in signed-out state | Hero reads "A Condura account is optional."; AuthPicker renders 3 rows; footer note is present. |
| Mounts in signed-in state | AccountCard renders; threadlink reads "Manage on condura.app →"; Sign out button is present. |
| Google row click | Calls `account.signInWithGoogle('condura://auth/callback')`; chosen row's icon morphs to a `Pulse phase="thinking"`; other rows dim to 0.5 opacity. |
| GitHub row click | Same as Google, with GitHub redirect URI. |
| Email row click | Email row expands (height increases); email input + Send link button appear; focus moves to input. |
| Email field submit | Calls `account.signInWithEmail(email, 'en', 'https://condura.app/auth/verify')`; on success, shows "Link sent to {email}…" with a `Pulse phase="ok"`. |
| Email field invalid format | Send link button is disabled. |
| OAuth callback resolves | AuthPicker fades out; AccountCard slides in; eyebrow updates to "ACCOUNT · SIGNED IN". |
| Sign out click → confirm opens | `.c-popover` is rendered; focus moves to Cancel button. |
| Confirm cancel | Popover closes; focus returns to Sign out button. |
| Confirm sign out | Calls `account.signOut()`; AccountCard fades out; AuthPicker fades in. |
| Guide-error rendering | When `account.error` is set, the ErrorState block renders above the AuthPicker with the matching copy. |
| Footer note always renders | Whether signed in or signed out, the footer note is present in the DOM. |
| Reduced-motion path | With `prefers-reduced-motion: reduce`, all `transition-duration` values resolve to 0ms. |
| Tab order (signed-out) | Three AuthPicker rows are focusable in order; Email expansion adds input + button. |
| Tab order (signed-in) | Threadlink then Sign out button. |
| Esc on popover | Popover closes; focus returns to Sign out. |
| Esc on email row | Email row collapses; focus returns to AuthPicker Email row. |
| Screen reader announcements | Each state transition triggers an `aria-live` announcement. |

### 11.2 Visual / E2E (Playwright)

| Test | Asserts |
|---|---|
| Route renders from `#/account` | NavRail Account item is active; route content is visible. |
| Full sign-in flow | Click Google → browser opens → callback → AccountCard slides in. |
| Magic-link flow | Click Email → enter email → Send link → "Link sent" appears. |
| Sign-out flow | Click Sign out → confirm → AuthPicker returns. |
| Reduced motion | All animations resolve in <50ms (no visible stagger). |

### 11.3 MOAT compliance checklist

- [ ] No gradient text (`§4 #1`).
- [ ] No emoji as UI icons (`§4 #2`).
- [ ] No glassmorphism (`§4 #3`).
- [ ] No rainbow accents (`§4 #4`).
- [ ] No "Welcome to the future" copy (`§4 #5`).
- [ ] No fake enthusiasm (`§4 #6`).
- [ ] No spinner loaders (`§4 #7`).
- [ ] No rectangular focus outlines (`§4 #8`).
- [ ] No double shadows (`§4 #9`).
- [ ] Every animation carries meaning (`§4 #10`).
- [ ] Thread is the only flourish for "this is now finished" (`§3`).
- [ ] All copy lives in locales/`lang`.json (§10.4).
- [ ] WCAG 2.2 AA contrast on every text + UI element (§9).
