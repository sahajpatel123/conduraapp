# SCREEN_DELEGATION.md — Condura

> **The screen spec for `#/delegation` (Delegation.svelte).** Source of truth for layout, states, motion, keyboard, data. Implementation lives in `app/web/frontend/src/lib/condura/Delegation.svelte`. Read alongside `DIRECTION.md` (voice), `MOAT.md` (bar), `APPFLOW.md §4.8` (this surface's spec).
>
> **One sentence:** *Condura at the centre, sub-agent CLIs on the ring; threads drawn; live panel for what the Gatekeeper has surfaced.*

---

## 1. LAYOUT & CONTENT

Three-column surface below 820px collapses to two (canvas + live stacked); below 480px collapses to one column. **Constellation is the hero. Live is the receipt.** Bottom CTA always pinned.

| Region | Position | Width | Contents |
|---|---|---|---|
| **Header (`.head`)** | Top, full-width | 100% / max 1100px | eyebrow `— The conductor` (mono 11px faint), h1 `Delegation.` (display 28-40px), sub copy 56ch max |
| **Left rail · CLI list (`.cli-list`)** | Below header, left | 168px fixed (≥820px only) | Auto-detected CLIs from `$PATH` (claude, codex, antigravity, opencode, kilo, hermes, ollama). Each row: status dot (4px) + mono label + Tooltip on hover. Status: `detected` (synapse), `missing` (faint), `busy` (pollen animated). |
| **Main · Constellation (`.canvas`)** | Centre | 440×440 (responsive: 380h below 820px) | Breathing core (96px halo, Pulse `acting`), 7 sub-agent nodes on a 340px dashed orbit ring (170px radius), SVG thread legs (H+V diagonals through core). Empty: single Glyph `delegation` 56px + caption. |
| **Right · Live panel (`.live`)** | Right of canvas | 280-360px (minmax) | Eyebrow `— Live` + count chip + Refresh pill. Rows: badge + label + Approve/Deny/Run. Empty: italic "Nothing in flight." + sub. |
| **Bottom · Spawn CTA (`.spawn`)** | Below surface, full-width | 100% / max 1100px | Sticky. CLI picker (dropdown of detected CLIs), model picker (text input, mono 12px), task textarea (multiline, 80px max), Spawn button (primary, `--action` pollen) + `⌘N` kbd hint. Disabled until CLI + task non-empty. |

### Visual weight (top → bottom)

1. **Constellation core** (96px halo + breathing Pulse — highest weight, draws eye first).
2. **Sub-agent nodes** (96px cards, pulse-dot 6px).
3. **Spawn CTA** (pollen primary — second-highest contrast, signals "do this now").
4. **Live count chip** (`--synapse` mono 11px — always live, signals current state).
5. **Left rail** (faint mono labels — last resort, scannable not readable).

### Density

- Canvas: airy. 340px ring on 440px canvas = 50px breathing room each side.
- Live rows: `space-3` gap, single-line label, button row wraps at 320px.
- Spawn CTA: `space-4` internal padding, three regions stacked.

### What scrolls

- **Live panel** scrolls internally when rows exceed viewport height (`overflow-y: auto`).
- **Nothing else scrolls** — surface is single-screen. If sub-agent list grows beyond 7, canvas must scale ring radius, not paginate.

---

## 2. STATE MATRIX

| State | Constellation | Live panel | Spawn CTA | Header copy | Thread state | Pulse |
|---|---|---|---|---|---|---|
| **empty** | 1 node (core only, single Glyph if zero CLIs) | italic `Nothing in flight.` + sub | enabled, defaulting to first detected CLI | `Delegation.` | no threads | breathing core only |
| **spawning** | new node scales 0→1, thread draws from core | "Spawning `<cli>`…" row with Pulse `thinking` | disabled, label flips to `Spawning…` | unchanged | **draws in** 240ms (line 1→0) | new node breathes on arrival |
| **running** | node breathes (synapse-glow pulse 1.6s), thread brightens to `--synapse-glow` | streams stdout/stderr in mono 12px, 16ms cadence | enabled (parallel spawns allowed up to 5) | unchanged | glow pulse loops | running node = brightest |
| **completed** | node static, pollen check `✓` Glyph fades in over 240ms | final line + `Done · N tokens` summary | enabled | unchanged | **stops drawing**, holds at full | static |
| **error** | node border turns `--danger`, ring scales 1→1.04→1 once (220ms) | first line = error message in `--danger`, retry button below | enabled | unchanged | thread dims to `--danger` 40% | static (Pulse `error` if persistent) |
| **cancelled** | node scales 1→0 over 240ms, thread erases right→left over 320ms | row removed, count decrements with `--count-delta` 200ms | enabled | unchanged | **erases** | static |

### Exact copy

| Where | Copy |
|---|---|
| Header eyebrow | `— The conductor` |
| Header h1 | `Delegation.` |
| Header sub | `Condura at the center, sub-agents on the ring. Each thread is a delegation channel; the Live panel shows what's actually in flight.` |
| Empty canvas (0 sub-agents) | eyebrow `— No sub-agents wired` / body `Install Claude Code, Codex, or Ollama to populate the constellation.` |
| Empty canvas (0 spawned) | italic `Awaiting your first delegation.` |
| Empty live | h `Nothing in flight.` / sub `When a sub-agent asks the Gatekeeper to act, the ticket will land here.` |
| Spawn placeholder (task) | `What should <cli> do?` |
| Spawn placeholder (model) | `default model` |
| Spawn button (idle) | `Spawn sub-agent` + kbd `⌘N` |
| Spawn button (spawning) | `Spawning…` (no kbd) |
| Node tooltip | `<name> · <state>. Run: opens delegate stream. Cancel: stops the current delegation.` |
| Popover footer | `Run opens delegate stream · Cancel stops the current delegation` |
| Error row | `<error.message>` mono 11px `--danger` |
| Retry button | `Retry spawn` |
| Cancel button (in Live row) | `Cancel` (kebab action, mono uppercase) |

---

## 3. MOTION CHOREOGRAPHY

All durations from `--dur-*` tokens. No raw ms. No cubic-beziers outside `--ease` / `--ease-in` / `--ease-pop`.

| Gesture | Property | Duration | Easing | Trigger |
|---|---|---|---|---|
| **Constellation initial draw** | SVG `stroke-dashoffset: 1 → 0` (thread legs) + node `transform: scale(0) → scale(1)` | `--dur-cine` (900ms) | `--ease` (legs), `--ease-pop` (nodes) | route mount |
| **Node stagger** | each node 60ms after previous | derived from `--dur-cine` / 7 | `--ease-pop` | route mount |
| **Spawn: node appearance** | `transform: scale(0) → scale(1)` | 320ms | `--ease-pop` | `delegation.spawn` resolves |
| **Spawn: thread draw** | `stroke-dashoffset: 1 → 0` | 240ms | `--ease` | same |
| **Running: synapse-glow pulse** | `filter: brightness(1) → brightness(1.15)` + dot scale `1 → 1.4 → 1` | 1.6s loop | `--ease` | `state === 'running'` |
| **Live stream cadence** | append-line to scroll container | 16ms (rAF throttled) | n/a | SSE `delta` event |
| **Cancel: node exit** | `transform: scale(1) → scale(0)` | 240ms | `--ease-in` | cancel click |
| **Cancel: thread erase** | `stroke-dashoffset: 0 → 1` (reverse direction) | 320ms | `--ease-in` | same |
| **Error: ring shake** | `transform: translateX(0 → -2px → 2px → 0)` | 220ms | `--ease` | error event |
| **Count delta** | `--count-delta: 1` opacity `0 → 1 → 0` over 200ms | 200ms | `--ease` | Live row add/remove |
| **Hover node** | `transform: scale(1) → scale(1.06)` + border `--hair → --synapse` | `--dur` | `--ease` | mouseenter |
| **Hover thread leg** | stroke `--synapse → --pollen` | `--dur` | `--ease` | node hover |
| **Popover enter** | `opacity: 0 → 1` + `translateY(-4px → 0)` | `--dur-fast` | `--ease` | node hover |
| **Spawn button press** | `scale(0.97)` + `filter: brightness(0.95) saturate(1.1)` + `translateY(0.5px)` | `--dur-fast` | `--ease` | mousedown |
| **Refresh pill** | `translateY(0 → -1px)` + border `--hair-strong → --synapse` | `--dur` | `--ease` | hover |

### Reduced-motion contract

`@media (prefers-reduced-motion: reduce)` (owned exclusively by `condura.css`):

- Constellation initial draw skipped (nodes render at scale 1, threads at full opacity).
- Spawn node scale animation skipped (instant appearance).
- Running pulse replaced with a static `--synapse-glow` border (no filter oscillation).
- Cancel exit skipped (instant removal).
- Error shake skipped (border color transition only).
- Live stream cadence unchanged (it's data, not decoration — the cadence is functional).

### Energy budget

- `data-energy="low"`: drop running pulse to static border, disable hover translateY, skip popover translateY (opacity only).
- `data-energy="balanced"` (default): as above table.
- `data-energy="high"`: all gestures enabled, longest durations.

---

## 4. KEYBOARD

| Key | Action | Region | Visible affordance |
|---|---|---|---|
| **Tab** | Cycle through CLIs in left rail (wraps) | Left rail | Each focused CLI gets synapse ring (2px) + pollen halo (5px) per DIRECTION.md §6 rule |
| **Shift+Tab** | Reverse cycle | Left rail | same |
| **Enter** (on focused CLI) | Focus spawn task input + pre-select that CLI | Global | Spawn CTA CLI picker shows name |
| **⌘N** / **Ctrl+N** | Focus spawn task input (from anywhere on route) | Global | Spawn button shows `⌘N` hint |
| **⌘.** / **Ctrl+. ** | Cancel focused/running sub-agent (most recent first if none focused) | Global | Live row marked with 200ms danger halo |
| **Esc** | Deselect node / close popover (if open) / clear spawn task focus (if empty) | Canvas / Live | focus-visible halo removed |
| **↑ / ↓** | Move focus between Constellation nodes | Canvas | Focus ring follows (matches DIRECTION.md `top: highlightTop` pattern from ⌘K palette) |
| **← / →** | Move focus between Constellation nodes (wrap) | Canvas | same |
| **Enter** (on focused node) | Open popover for that node (if not running) / focus Live row for that CLI (if running) | Canvas | popover slides in |
| **Space** (on focused node) | Toggle node selection (pins popover) | Canvas | popover stays open until Space/Esc |
| **⌘R** | Refresh Live panel | Global | Refresh pill pulses pollen once |
| **⌘K** | Open command palette (global, route-independent) | Global | unchanged |
| **?** | Open keyboard shortcuts sheet | Global | unchanged |
| **g then d** | Go to Delegation route (from any route) | Global | unchanged |

### Focus order (mount order)

1. Left rail: first CLI (Tab to enter region).
2. Constellation: core (Space/Enter → pulse-pause).
3. Constellation: nodes (↑↓←→ to navigate).
4. Live panel: Refresh → row buttons (Approve / Deny / Run).
5. Spawn CTA: CLI picker → model picker → task textarea → Spawn button.

### What does NOT have a keybinding

- Selecting a model directly (text input, not keyboard-driven).
- Reordering CLIs (auto-detected, not user-ordered).

---

## 5. COMPONENTS USED

| Component | Source | Purpose in this surface |
|---|---|---|
| **Constellation** | inline SVG in `Delegation.svelte` (lines 184-213) | Hero canvas — core + orbit ring + 7 nodes + diagonal thread legs |
| **CLIList** | new — extract from `Delegation.svelte` (left rail, currently not split) | Auto-detected CLIs with status dot, mono label, Tooltip |
| **LivePanel** | `<aside class="live">` (lines 282-353) | Eyebrow + count chip + Refresh + rows of pending actions |
| **SpawnCTA** | new — extract to `SpawnCTA.svelte` | CLI picker + model picker + task textarea + Spawn button |
| **Thread** | `Thread.svelte` (existing, 2 instances per node: H leg + V leg) | Diagonal SVG thread legs from core to each node |
| **Pulse** | `Pulse.svelte` (existing) | Core halo (96px, phase `acting`) + per-node pulse-dot (6px, phase varies) + error states (phase `error`) |
| **Glyph** | `Glyph.svelte` (existing) | `delegation` (empty canvas), `orbit` (sub-agent node icon), `spawn` (CTA), `cancel` (Live row), `cli` (left rail) |
| **Button** | `Button.svelte` (existing) | Spawn (primary pollen), Refresh, Cancel, Approve/Deny/Run (per Live row) |
| **Tooltip** | new — extract from `MOAT.md §2.9` | Per-CLI hover, per-node hover (replaces current popover for non-running nodes) |
| **Popover** | new — `.c-popover` primitive per MOAT.md §2.8 | Pinned node popover (Space to pin) |
| **EmptyState** | `EmptyState.svelte` (existing) | Empty canvas + empty live (3-line copy per MOAT.md §2.4) |

### Components NOT used (and why)

- **Spinner** — banned per MOAT.md §4 rule 7. Pulse + Thread only.
- **Toast / celebration modal** — banned per MOAT.md §4 rules 5-6. Live row is the receipt.
- **Modal / Sheet** — Delegation is not modal. Spawn CTA is inline. No overlay.

---

## 6. DATA FETCHED

| RPC | Direction | Cadence | Used for |
|---|---|---|---|
| `delegation.listAgents` | daemon → client | on mount + every 10s | Source of truth for active delegations (currently mapped from `pendingActions` store) |
| `delegation.spawn` | client → daemon | on Spawn click | Creates new sub-agent; returns `{agent_id, status: 'spawning'}` |
| `delegation.cancel` | client → daemon | on Cancel click (Live row or ⌘.) | Stops running sub-agent; returns `{agent_id, status: 'cancelled'}` |
| `delegation.subscribe` | daemon → client (SSE) | stream | Per-line stdout/stderr + status transitions (spawning → running → completed/error) |
| `cli.detect` | daemon → client | on mount + every 30s | Auto-detect CLIs in `$PATH`, return `{claude: bool, codex: bool, ...}` |
| `pendingActions.refresh` | client → daemon | every 5s | Polls Gatekeeper pending tickets (already wired in current `Delegation.svelte:151`) |

### Polling cadence rationale

- `listAgents` 10s — sub-agent state changes are slow (model latency dominates).
- `pendingActions` 5s — Gatekeeper tickets are the UI's heartbeat; matches current implementation.
- `cli.detect` 30s — $PATH only changes on install/uninstall; slow cadence is fine.
- SSE for live stream — never poll stdout/stderr.

### Error states

| Failure | Surface response |
|---|---|
| `delegation.listAgents` rejects | canvas renders empty (single Glyph), Live panel shows italic "Couldn't reach delegation service." with Retry button |
| `delegation.spawn` rejects (spawn failed) | Spawn button label flips back to `Spawn sub-agent`, error appears below textarea in `--danger` mono 11px |
| `delegation.subscribe` SSE drops | Live row for that agent shows `stream disconnected` italic + auto-reconnect badge |
| `cli.detect` rejects | Left rail shows `detection failed` italic in `--warn`; Spawn CTA falls back to text input for CLI name |
| `pendingActions.refresh` rejects | Live panel empty state unchanged (graceful degradation per APPFLOW.md §6.1) |

### Auth / secrets

- No auth headers — daemon is local Unix socket.
- No API keys transit this surface.
- CLI detection runs in daemon sandbox; no user input passed to `$PATH` lookup.

---

## 7. DESIGN DECISIONS

### MOAT compliance

| Test | Pass | Evidence |
|---|---|---|
| **Restraint (§1)** | ✓ | Single core, single ring, single Glyph in empty. No 3D tilt, no glassmorphism, no rainbow. |
| **Detail (§2)** | partial | Empty states teach (3 lines), focus rings tracked (pill on refresh, synapse ring on nodes), loading uses Pulse+label not spinner. **Gap:** Tooltip primitive not yet extracted (MOAT §2.9). |
| **Signature (§3)** | ✓ | The thread is everywhere — core-to-node legs draw on mount, brighten on running, erase on cancel. |
| **Anti-patterns (§4)** | ✓ | No gradient text, no emoji, no glass, no rainbow, no "Welcome" copy, no spinner, no rectangular outlines, no double shadows, no animation without meaning. |
| **$50M feel (§5)** | ✓ | Popover pinned on Space (premium), cancel yields gracefully (draws-out, doesn't snap), empty teaches install. |

### Why this is a power-user surface (and readable)

- **Three affordances, one per region**: left rail = what's available, canvas = what's wired, Live = what's happening, CTA = what's next. A power user can scan all four in 200ms. A new user reads the canvas (visual hierarchy) first, then the CTA.
- **Density is opt-in**: left rail is faint by default (scannable, not readable); Live is full-bleed mono (readable when needed). Power users enable both; new users ignore left rail entirely.
- **No settings surface deeper than this**: autonomy matrix lives in Settings; delegation has no settings — it's all visible at once.

### Why Constellation is the signature viz

- **Teaches the topology**: Condura at center, sub-agents on ring, threads between them. 200ms comprehension.
- **Scales with state**: 1 node = empty, 7 nodes = full ring, lines that draw = live, lines that erase = cancelled. The viz IS the state.
- **Matches the Sync garden metaphor** (nodes + threads) without copying it — Sync is many-to-many, Delegation is one-to-many.

### Why empty TEACHES

- Zero sub-agents → Glyph + "Install Claude Code, Codex, or Ollama to populate the constellation." (per MOAT §2.4: what / why empty / next action).
- Zero spawned but CLIs detected → "Awaiting your first delegation." with cursor in task input.
- Zero pending in Live → "Nothing in flight. When a sub-agent asks the Gatekeeper to act, the ticket will land here." — explains the live panel's purpose (a user might otherwise think it's broken).

### Why Live panel never spinners

- The Live panel is a **stream**, not a fetch. Data arrives on SSE; spinner would be a lie.
- New rows append at the bottom with a 200ms `--count-delta` fade.
- During a stream, the active row pulses the running-node glyph (synapse-glow).
- If SSE drops, the row shows `stream disconnected` italic + reconnect attempt — never a spinner.

### Why a Spawn CTA, not a command

- Power users will use ⌘N. New users won't know it exists. The CTA is for new users; the keyboard is for power users. Both paths reach the same store action.
- The CTA shows the CLI + model + task in one region — the user can see what they're about to spawn before clicking.

### Trade-offs accepted

- **No drag-to-reorder** of sub-agent nodes — auto-positioned by index, deterministic.
- **No cancel-all button** — ⌘. cancels most-recent only (safer than mass-cancel).
- **No model picker dropdown** — text input only (no full provider enumeration in this surface; falls back to default model).
- **No autonomy matrix** — autonomy lives in Settings (per APPFLOW.md §4.10).

---

## 8. DRIFT TABLE

What changed from the current `Delegation.svelte` to this spec.

| ID | What | Status | Rationale |
|---|---|---|---|
| **D-DEL-01** | Add left rail (CLIList) | **ADD** | Current implementation inlines 7 sub-agents in the canvas as static nodes — no visibility into which CLIs are actually detected vs missing. Rail = truthful state. |
| **D-DEL-02** | Replace canvas popover with Tooltip + pinned Popover | **CHANGE** | Current: always-shown popover on hover only. Spec: Tooltip on hover (400ms delay), Popover pinned on Space. Matches MOAT §2.9. |
| **D-DEL-03** | Move Spawn from (non-existent) to bottom CTA | **ADD** | Current `Delegation.svelte` has no spawn affordance — it's read-only. Spec adds SpawnCTA with CLI picker + model + task. |
| **D-DEL-04** | Use `delegation.listAgents` + SSE for live sub-agent state | **CHANGE** | Current implementation derives sub-agent state from `pendingActions` only. Spec sources truth from `delegation.listAgents` (10s poll) + `delegation.subscribe` (SSE) for stdout/stderr. |
| **D-DEL-05** | Rename "Live" eyebrow to `— Live` (no change) | **HOLD** | Current matches. |
| **D-DEL-06** | Add ⌘N global hotkey for Spawn | **ADD** | Per MOAT §2.10 keyboard story. |
| **D-DEL-07** | Add ⌘. for cancel-most-recent | **ADD** | Same. |
| **D-DEL-08** | Constellation initial draw on route mount | **CHANGE** | Current renders threads drawn from start (no entrance animation). Spec adds 900ms draw on mount with 60ms node stagger. |
| **D-DEL-09** | Cancel draws thread out (right→left) | **ADD** | Current cancel is row removal only. Spec adds visual erasure to match MOAT §5.3 spirit ("agent yielding, not just unfocused"). |
| **D-DEL-10** | Reduced-motion drops node scale animation | **CHANGE** | Current reduced-motion handler (lines 850-866) disables transitions globally. Spec narrows: keep transitions, drop scale entrance. |
| **D-DEL-11** | Energy budget hook (`data-energy`) | **ADD** | Current has no energy awareness. Spec adds low/balanced/high tiers. |
| **D-DEL-12** | Error ring shake (220ms) | **ADD** | Currently errors show as static `--danger` border. Spec adds subtle horizontal shake on error event. |
| **D-DEL-13** | Per-node tooltip vs canvas-wide aria-label | **CHANGE** | Current has aria-label on canvas + node. Spec keeps canvas aria-label, adds per-node Tooltip for sighted users (a11y win). |
| **D-DEL-14** | `delegation.subscribe` SSE for stdout/stderr | **ADD** | Current Live panel shows only pending Gatekeeper tickets. Spec adds per-agent stdout/stderr stream under each running node row. |
| **D-DEL-15** | Remove "Refresh" pill from Live header | **CONSIDER** | With SSE live, manual refresh is redundant. **Hold for v0.2.0** — power users may want it. |
| **D-DEL-16** | Keep `pendingActions` polling (5s) | **HOLD** | Current cadence correct per APPFLOW.md §4.8. SSE supplements, doesn't replace. |
| **D-DEL-17** | Orbit ring dashed → solid on hover | **CONSIDER** | Spec holds dashed (current). Could thicken on hover for affordance. Defer. |
| **D-DEL-18** | Popover pins via Space | **ADD** | Currently popover closes on mouseleave. Spec pins via Space, closes on Space/Esc. Matches Linear/Arc pattern. |
| **D-DEL-19** | `cli.detect` polled every 30s | **ADD** | Currently CLI list is static (no detection). Spec polls daemon. |
| **D-DEL-20** | Remove 96px halo from non-running nodes | **HOLD** | Current halo is on core only (correct). Verify no regression. |

### Removed

| What | Why |
|---|---|
| Custom shadow on hover (`Skills.svelte`-style `0 0 0 4px --synapse-glow + --shadow-card`) | Replaced with `--shadow-float` only. Per MOAT §4 rule 9 (no double shadows). |
| Inline empty copy "No sub-agents wired" without action line | Replaced with 3-line pattern (what / why / next). Per MOAT §2.4. |
| Pulse `error` on every error | Replaced with one-shot ring shake + static `--danger` border. Pulse reserved for persistent states (kill switch, awaiting consent). |

### Unchanged from current implementation

- Orbit ring geometry (340px diameter, 170px radius).
- CANVAS_SIZE (440px).
- 7 sub-agent list (claude_code, codex, antigravity, opencode, kilo, hermes, ollama).
- Live panel structure (eyebrow + count + refresh + rows).
- Approve/Deny/Run buttons per row.
- `--synapse-glow` 6px pulse-dot at 2.4s breathing.
- `aria-label` on canvas (`role="img"`).
- Reduced-motion `transform: translate(-50%, -50%) scale(1) !important` for nodes.

---

**Total spec:** 8 sections, ~5,400 words. Implementation estimate: 6-8 hours (left rail + SpawnCTA extraction + SSE wiring + entrance/exit choreography + Tooltip primitive). Touches: `Delegation.svelte`, `Tooltip.svelte` (new), `SpawnCTA.svelte` (new), `delegation.store.ts` (SSE), `cli.detect` daemon RPC (new).