# SCREEN — PublishModal → PublishSheet · `PublishSheet.svelte`

> **Status.** Drill-in spec for the publish surface referenced by
> `SCREEN_HUB.md §3.4` and §6.5. The Phase-2 `SCREEN_HUB.md`
> already mandates `.c-sheet` per MOAT §2.8 ("a task, not a
> confirmation"). The existing `app/web/frontend/src/lib/components/
> PublishModal.svelte` is the **legacy centered-modal implementation**
> that this sheet replaces. The file name `SCREEN_PUBLISHMODAL.md`
> survives for indexing; the **component name in the new surface is
> `PublishSheet.svelte`** (see Drift Table §8).

> **Contracts.** `MOAT.md` (premium bar), `DIRECTION.md` (voice),
> `APPFLOW.md §4.3` (Hub's IPC), `SCREEN_HUB.md §3.4 / §6.5`
> (parent spec). All rules below cite the source.

---

## 1. LAYOUT & CONTENT

Container: a **`.c-sheet` sliding up from the bottom edge** of the
viewport (per MOAT §2.8 — a *task*, not a *confirmation*). The route
body stays live behind the sheet; no scrim, no `backdrop-filter: blur`
(MOAT §4 #3). Esc closes. One `<Thread />` draws across the sheet's
top edge on `mount` (the **arrival** gesture from MOAT §3).

### 1.1 Two-column region (≥ 960px viewport)

```
┌──────────────── PublishSheet · 720px max · 100vw on narrow ────────────┐
│                                                                          │
│  ─ Close ×                                                               │  (top-right, 32×32 hit, <Glyph close/>)
│  — Publish to the public Hub                                             │  eyebrow (mono caps 11, --content-faint)
│  Tell the world what your skill does.                                    │  title (display 28, Instrument Serif)
│                                                                          │
│  ┌────────── form (440px, left) ──────────┐  ┌── preview (240px, right) ─┐│
│  │ Name                                     │  │ ┌────────────────────┐  ││
│  │   [_________________________________]    │  │ │ · synapse dot       │  ││  Mini skill card (the shape
│  │   placeholder: "Skill name"              │  │ │   (Official trust)  │  ││  that will land on the Hub).
│  │                                          │  │ │                     │  ││
│  │ Version (semver — major.minor.patch)     │  │ │  {name live}        │  ││  Re-renders on every field
│  │   [_________________________________]    │  │ │  v{version live}    │  ││  change via 200ms cross-fade
│  │   placeholder: "1.0.0"                  │  │ │                     │  ││  (DIRECTION §5).
│  │   ⚠ inline err if !semver (on blur)      │  │ │  {description live} │  ││
│  │                                          │  │ │  by {author}        │  ││
│  │ Description                              │  │ │  # {tag} # {tag}    │  ││
│  │   [_____________________________________]│  │ │                     │  ││
│  │   [_____________________________________]│  │ │  archive · {kb} ·zip│  ││
│  │   [_____________________________________]│  │ └────────────────────┘  ││
│  │   placeholder: "What does this do?"      │  │                            │
│  │                                          │  └────────────────────────────┘│
│  │ Author (pre-filled, read-only)           │                                 │
│  │   {account.email} · "Change in Settings" │                                 │
│  │                                          │                                 │
│  │ License                                  │                                 │
│  │   [ MIT ▾ ]   (Select — MIT, Apache-2.0, │                                 │
│  │                BSD-3-Clause, GPL-3.0,    │                                 │
│  │                ISC, Proprietary)         │                                 │
│  │                                          │                                 │
│  │ Tags                                     │                                 │
│  │   [weather ×] [api ×] [+] (ChipInput)    │                                 │
│  │   comma or ↵ commits · backspace removes │                                 │
│  │                                          │                                 │
│  │ Archive (.zip · ≤32 MB)                  │                                 │
│  │   ┌──────────────────────────────────────┐│                                 │
│  │   │ ⤴  drop a .zip here or browse…       ││                                 │
│  │   └──────────────────────────────────────┘│                                 │
│  │   file-name · 142 KB · .zip              │                                 │
│  └──────────────────────────────────────────┘                                 │
│                                                                          │
│  ─ safety scan progress ───────────────────────────────────────────────  │
│  [Thread draws at scan start → fills to 100% as daemon reports]          │
│  Pulse phase="acting" · "Scanning for promptware…" (mono 11)             │
│                                                                          │
│  [Cancel]                                              [Publish →]        │  footer CTAs
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

### 1.2 Stacked region (< 960px viewport)

Form columns collapse to single column. Preview moves below the form
(reads as "this is the card that will land"). Footer CTAs go sticky to
the bottom edge of the sheet. Field stagger still fires on mount
(40ms per field, per §3.2) but the preview cross-fades only when its
content actually changes (no per-keystroke flicker on a phone).

### 1.3 Layout contracts

| Region | Width (≥960) | Width (<960) | Padding | Source |
|---|---|---|---|---|
| Sheet max | 720px, 100vw-32px | 100vw | `var(--space-7)` all sides | SCREEN_HUB §3.4 |
| Form column | 440px | 100% | `var(--space-4)` right gutter | DIRECTION §3 type scale |
| Preview column | 240px | 100%, stacked below form | `var(--space-4)` left gutter | inherited `--r-md` card |
| Footer row | 100% of sheet | sticky bottom | `var(--space-5)` vertical | inherited `.tactile` |

### 1.4 What's *not* in this sheet (route-level concerns, not sheet-level)

| Responsibility | Owns it | Why it's not here |
|---|---|---|
| `hub.publish(payload)` IPC | `<Hub />` (parent route) | Sheet is a controlled component; `onSubmit` is a callback |
| Sign-in gate (`account.isSignedIn`) | Parent CTA (the HeroShelf's "Publish a skill →" pill) | Sheet is unconditional; the parent decides to open it |
| Error fallback (`<ErrorState />`) | Route-level anchor | The sheet never inlines an error block — per MOAT §1.2 (no 5th copy of err-state) |
| Hub grid re-fetch after success | Parent route poll (60s) | Sheet just resolves `Promise<void>`; the Hub sees the new card on its next poll |

---

## 2. STATE MATRIX

The sheet carries **one of nine states** at a time. Mutations are
`hub.publishState.kind` from `lib/stores/hub.svelte.ts` + sheet-local
field validity. The Hub §3.4 row "publishing (publish sheet mounted)"
maps to the consolidated column here.

| # | State | What you see | Trigger / data | CTAs |
|---|---|---|---|---|
| **S0** | **closed** | Sheet unmounted. Route live behind. (Aria-hidden inert, document `body { overflow: auto }`.) | `open === false` | n/a |
| **S1** | **open-empty** | Sheet slides up; all fields empty; preview shows `<name>` placeholders; safety scan idle (no thread, no label). | `open === true && archive === null && name === ''` | Cancel enabled · Publish disabled (semi-transparent `--content-faint`) |
| **S2** | **open-filling** | Fields populate live; preview cross-fades on each change (200ms); safety scan still idle. | `open === true && (name !== '' \|\| …) && archive === null` | Cancel enabled · Publish disabled |
| **S3** | **open-with-archive** | File picker reads filename + KB (mono 11); preview shows `{archive} · {KB} · .zip` metadata block; safety scan **begins** (Thread draws at sheet bottom, Pulse `acting`). | `open === true && archive !== null && semverValid` | Cancel enabled · Publish disabled until scan completes (the scan is a preview of "what the daemon will run"); the file picker shows a `× remove` ghost button |
| **S4** | **validating (safety scan in progress)** | Thread at bottom filling 0→100% over ~1.6s (or daemon-reported progress); mono 11 label "Scanning for promptware…" rotating under the Thread. Pulse `acting`. | `hub.safetyScan({archive}) → in-flight` | Cancel enabled · Publish disabled |
| **S5** | **scanning-error** | Thread stops at failure mark. `<Glyph name="warning" />` filled `--danger` 12px in a 56px row; mono 11 reads `Safety scan did not pass: {reason}` (e.g., "shell-metachar-in-script: find -exec rm {} +"). | `hub.safetyScan({archive}) → {ok: false, reason}` | Cancel enabled · Publish disabled · secondary "Fix and re-scan →" restarts scan |
| **S6** | **uploading** | Publish CTA morphs in place to `Publishing…` (mono 11 label) with `Pulse phase="acting"` 8px next to it (the same acting pulse the Chat surface uses; DIRECTION §5). Cancel disabled. A 1px hairline progress bar fills left→right beneath the sheet footer (`scaleX(0 → fraction)`, `--dur`, `--ease`). | `hub.publish(payload) → in-flight` | Cancel disabled · Publish disabled · single in-flight promise |
| **S7** | **uploaded** | Thread draws left→right across the **top edge** of the sheet (the **completion** gesture from MOAT §3). A `<Glyph name="check" />` 16px filled `--ok` stamps center-bottom of the form column. A mono 11 line replaces the footer: `Your skill is live → {hub.condura.app/skills/<id>}` (a real `<a href target="_blank">`, not a fake celebration). Footer CTA morphs to `Done →` (ghost pill); Esc closes. | `hub.publish(payload) → {ok: true, url}` | Cancel hidden · `Done →` ghost enabled |
| **S8** | **error (publish rejected)** | Sheet stays open (per SCREEN_HUB §3.4 — "the sheet stays open, the submit button enables back, a route-level `<ErrorState />` renders beneath the sheet"). Inside the sheet: a single mono-11 helper above the file picker reads `Publish failed: {reason}` (e.g., "duplicate slug"). | `hub.publish(payload) → reject({reason})` | Cancel enabled · Publish re-enabled · file picker unchanged (the archive passed scan, the failure is upstream) |

### 2.1 State transition table

```
S0 ─ click Publish CTA (parent) ─────────► S1
S1 ─ type into Name ─────────────────────► S2
S2 ─ fill Version (valid semver) + Desc ─► S2 (still no archive)
S2 ─ drop .zip into FilePicker ──────────► S3
S3 ─ sheet auto-fires hub.safetyScan ────► S4
S4 ─ safetyScan resolves ok ─────────────► S3 (Publish enables)
S4 ─ safetyScan rejects ─────────────────► S5
S5 ─ fix & click re-scan ────────────────► S4 (loop until pass)
S3 / Pass ─ click Publish ───────────────► S6
S6 ─ hub.publish resolves ok ────────────► S7
S6 ─ hub.publish rejects ────────────────► S8
S7 / S8 ─ click Done / Cancel / Esc ─────► S0
```

### 2.2 `disabled` derivation

```ts
canSubmit = (in S3 with scan passed)
         && name.trim().length > 0
         && semver.test(version.trim())
         && description.trim().length > 0
         && author.trim().length > 0
         && archive !== null
         && !hub.isPublishing
```

(Same predicate as the legacy `PublishModal.svelte:31`, translated
to the sheet's state-machine model.)

---

## 3. MOTION CHOREOGRAPHY

One rule: **every animation answers "what is this communicating?"**
(MOAT §4 #10). No decorative loops. Durations come from the four
locked tokens (`--dur-fast` / `--dur` / `--dur-slow` / `--dur-cine`);
eases from the three locked ones (`--ease` / `--ease-in` / `--ease-pop`).

| Gesture | Property | Duration | Easing | Trigger |
|---|---|---|---|---|
| **Sheet open** | `transform: translateY(24px) → 0` + `opacity 0 → 1` | `--dur-slow` (520ms) | `--ease` | `open === true` (S0 → S1) |
| **Sheet close** | reverse | `--dur` (280ms) | `--ease-in` | Cancel / Esc / X / Done |
| **Sheet top-edge Thread** (arrival) | `stroke-dashoffset 1 → 0` (SVG) or `scaleX(0 → 1)` (CSS) | `--dur-slow` | `--ease` | mount (S1) |
| **Field stagger** | `opacity 0 → 1` + `translateY(8px) → 0` | `--dur` each, **40ms apart** | `--ease` | mount (S1) |
| **Live preview cross-fade** | swap inner HTML, `opacity 1 → 0 → 1` | 200ms total (100ms out + 100ms in) | `--ease` | any field change in S2/S3 |
| **Drag-over file picker** | `border-color: hair-strong → synapse` + `background: pollen-halo 6%` | `--dur-fast` | `--ease` | `dragenter` on picker |
| **Drop on file picker** | `transform: translateY(-2px)` + `--shadow-card` → `--shadow-float` | `--dur` | `--ease` | `drop` event |
| **Safety scan Thread draw** (S4) | `scaleX(0 → 1)` left-to-right across 360px | `--dur-slow × 3` (1.6s) or daemon progress | `--ease` | scan starts |
| **Safety scan error stop** (S5) | Thread stops at failure mark; `<Glyph warning/>` scales `0 → 1` with `--ease-pop` once | `--dur` | `--ease-pop` | scan rejects |
| **Publish CTA synapse-pulse** (S6) | `box-shadow: 0 0 0 0 synapse-halo` → `0 0 0 8px transparent` | `--dur-slow` × 2 (loop) | `--ease` | publish begins |
| **Upload hairline progress** (S6) | `scaleX(0 → fraction)` left-to-right | `--dur` per increment (or daemon-tick cadence) | `--ease` | upload progress |
| **Completion Thread + check stamp** (S7) | Thread top-edge re-draws (520ms) + check `<Glyph check/>` stroke-dashoffset `1 → 0` over 320ms | `--dur-slow` (thread) + `--dur` (check) | `--ease` | publish ok |
| **Done pill morph** (S7) | `Publish →` → `Done →` label fade in place, no scale | `--dur` | `--ease` | state transition |
| **Mount entrance** (prefers-reduced-motion) | All slide + stagger collapses to a single 0ms appearance; the Thread still draws (it's *meaning*, not motion). | 0ms | n/a | mount |
| **Cross-fade** (prefers-reduced-motion) | Instant content swap; no opacity transition. | 0ms | n/a | field change |
| **Safety scan Thread** (prefers-reduced-motion) | Replaced by the `<Pulse phase="acting" />` alone (Pulse is the reduced-motion substitute for any animated hairline; DIRECTION §5). | n/a | n/a | scan starts |

### 3.1 Thread placements (per MOAT §3)

| Moment | Thread location | Why |
|---|---|---|
| Mount (S1) | top edge of sheet | "the sheet arrived" |
| Safety scan start (S4) | bottom of sheet (under fields, above footer) | "data is moving — the daemon is scanning" |
| Safety scan pass (S3→Submit enabled) | subtle 1px hairline under the Publish CTA | "the scan passed — you may proceed" |
| Publish ok (S7) | top edge of sheet (re-draws) | "the connection was made — your skill is live" |

**Hard ban (per MOAT §3 "Where the Thread MUST NOT appear"):** the
thread is *never* a button-hover flourish, never between unrelated
sheet regions, never on the footer CTAs alone (use `--hair` for that).

---

## 4. KEYBOARD

The sheet's keyboard surface is additive to the global shortcuts
(MOAT §2.10). Keys bind at sheet mount, unbind at unmount. The sheet
uses the same `prefers-reduced-motion` exemption — keys are
unchanged when motion is reduced; only animations change.

| Key | Action | Active state |
|---|---|---|
| **Tab** | Cycle forward through form fields (Name → Version → Description → Author (readonly, skipped) → License → Tags → Archive picker → Publish). | S1, S2, S3, S4 |
| **Shift+Tab** | Cycle backward. | same |
| **Enter** (not in textarea) | Trigger the focused field's primary action OR submit Publish if on the Publish button. **Whichever is the focused element's intrinsic binding.** | all |
| **Enter** in Tag input | Commit pending tag (split on comma, dedupe, append to chip list). | S2/S3 when focus in Tag |
| **Cmd+Enter** / **Ctrl+Enter** | Publish — same as clicking `Publish →`. | S3 with scan passed |
| **Esc** | Close the sheet. From S7 (uploaded), Esc is **disabled** until the user has read the success state for ≥ 1.5s (prevents dismissing success accidentally; MOAT §2.10). | all except early S7 |
| **Space** / **Enter** on File picker button | Opens the native file dialog (`<input type="file"> click()`). | S2/S3 |
| **Delete** / **Backspace** on a Tag chip (focused) | Remove the tag. | S2/S3 |
| **`/`** inside Name/Description | Types a literal `/`. No global handler intercepts. (The Hub search `/` shortcut is route-scoped to `#/hub`, not the sheet.) | S2/S3 |
| **Cmd+Z** / **Cmd+Shift+Z** | Undo / redo within the sheet's text fields (browser native; not custom-implemented). | S2/S3 |

### 4.1 Semver validation timing

The Version field's inline `<Glyph warning/>` + error message (`Must
be semver · major.minor.patch`, mono 11, `--danger`) appears **on
blur** (not on every keystroke) — per DIRECTION §6 ("the focus ring
follows the geometry the user is touching, not flatten it into a
rectangular outline"; symmetrically, validation doesn't pummel them
mid-typing). On blur, if the regex
`/^\d+\.\d+\.\d+(?:[-+].+)?$/` fails, the field shows the inline error
*and* the Publish CTA stays disabled with a tooltip `Fix semver first`.

### 4.2 Focus management

| Event | What gains focus |
|---|---|
| Sheet open | First form field (Name) |
| Tab past last field | Publish CTA |
| Esc / Cancel | Return focus to the parent Publish CTA pill (the one that opened the sheet) |
| Successful Publish (S7) | The `Done →` ghost pill (so Enter closes) |

---

## 5. COMPONENTS USED

The sheet is composed of these components. Each already exists in
`lib/condura/` or `lib/components/` (legacy) — this spec does not
re-declare them; it only fixes how the sheet composes them.

| Component | Source | Role in this sheet |
|---|---|---|
| **`<Sheet>`** (`.c-sheet` primitive) | New — `lib/condura/Sheet.svelte`, per MOAT §2.8 | Slides from bottom; owns Esc + focus-trap + aria-modal-less (a sheet doesn't block page scroll). |
| **`<Input>`** | `lib/components/ui/Input.svelte` | Name, Version, Author (readonly variant), Tags (chip-input variant). |
| **`<Textarea>`** | `lib/components/ui/Textarea.svelte` | Description. |
| **`<Select>`** | `lib/components/ui/Select.svelte` | License dropdown. |
| **`<ChipInput>`** | `lib/components/ui/ChipInput.svelte` | Tags row (typed tags, comma/`↵` commits, backspace removes, `×` on each chip). |
| **`<FilePicker>`** (drag-and-drop) | `lib/components/ui/FilePicker.svelte` | Archive drop-zone + native picker fallback. Shows filename + size after selection. |
| **`<LivePreview>`** (mini skill card) | `lib/condura/LivePreview.svelte` | Re-renders the SCREEN_HUB card micro-shape (trust dot + title + version + description + tags + archive-metadata). 200ms cross-fade on content change. |
| **`<Button>`** | `lib/condura/Button.svelte` | Cancel (ghost) · Publish → (primary pollen) · Done → (ghost, S7) · Fix and re-scan → (ghost, S5). Variants `primary`/`ghost` only. |
| **`<Thread>`** | `lib/condura/Thread.svelte` | Four placements per §3.1. |
| **`<Pulse>`** | `lib/condura/Pulse.svelte` | `phase="acting"` (8px) during safety scan (S4) and upload (S6); `phase="ok"` (6px) on completion (S7). |
| **`<Glyph>`** | `lib/condura/Glyph.svelte` | `publish` (CTA icon) · `upload` (during upload) · `archive` (picker hint) · `check` (S7 stamp, stroke draws in) · `close` (sheet close ×) · `warning` (S5 inline error) · `dot-active` (filled trust dot in preview). Single-stroke, 1.5 weight, `currentColor` (MOAT §4 #2). |
| **`<Tooltip>`** | `lib/condura/Tooltip.svelte` (to be created per MOAT §2.9) | Wraps the Publish CTA when disabled (label `Fix semver first` or `Add a .zip to enable`). Hover-delay 400ms, exit 75ms. |
| **`<ErrorState>`** | `lib/condura/ErrorState.svelte` (to be extracted per MOAT §2.6) | **NOT** rendered inside the sheet — only at the Hub route level (SCREEN_HUB §3.4). The sheet uses inline mono-11 helpers for field-level errors and a single helper line above the file picker for publish-rejection (S8). |

### 5.1 Component migration: `PublishModal.svelte` → `PublishSheet.svelte`

The legacy file `app/web/frontend/src/lib/components/PublishModal.svelte`
ships a centered `<Dialog>` with `backdrop-filter: blur(4px)` and
inline `err-state`, `ok`-result, `err`-result, preview-YAML block.
This sheet **inherits none of that.** Migration:

| Legacy element (PublishModal.svelte) | Replacement here |
|---|---|
| `<Dialog>` centered modal | `<Sheet>` (`.c-sheet`) sliding from bottom |
| `backdrop-filter: blur(4px)` on scrim | removed (no scrim — sheet is paper-on-paper) |
| Inline `result.ok` block (success) | Sheet's S7 column: Thread + check stamp + `<a href=…>` |
| Inline `result.err` paragraph | Route-level `<ErrorState />` + single sheet-local mono-11 helper |
| `<pre class="preview-yaml">` block | Replaced by `<LivePreview>` (the *visual* card, not raw YAML — matches the Hub's card vocabulary; SCREEN_HUB §3.4) |
| Three-column grid (`grid-three`) | Two-column region (form + preview); Author is read-only display only |
| Two-column grid (`grid-two`) | Stacked (License + Tags are individual rows in the form column) |
| `<input type="file">` plain | `<FilePicker>` (drag-and-drop + picker button) |

---

## 6. DATA FETCHED

The sheet calls **two** IPCs. Both signatures live in
`lib/ipc/client.ts` (`Hub` namespace) and `lib/ipc/types.ts`.

### 6.1 `hub.safetyScan({ archive: Uint8Array, filename: string })`

New RPC, runs **before** `hub.publish`. Returns
`{ ok: boolean; reason?: string; findings?: ScanFinding[] }` where
`reason` is one of:

| `reason` value | Trigger | Sheet state |
|---|---|---|
| `"clean"` | scan passed | S4 → S3 (Publish enables) |
| `"shell-metachar-in-script"` | archive contains a `.sh`/`.bash`/`.py` with `find -exec rm {}` / similar | S5 (show reason in mono-I'm z-ai/glm-5.2. system-prompt-shaped text in a SKILL.md | S5 |
| `"unreviewed-binary"` | archive contains an executable Mach-O/PE/ELF | S5 |
| `"oversized"` | archive > 32 MB (caught client-side; should never reach the daemon) | S5 |
| `"parse-error"` | archive is not a valid `.zip` | S5 |

### 6.2 `hub.publish(payload)`

Existing RPC, unchanged shape from `PublishModal.svelte:99`. Returns
`{ ok: boolean; url?: string; reason?: string }`. Drives states S6 → S7
→ S8.

### 6.3 What's *not* fetched by the sheet

| IPC | Why it's not in the sheet |
|---|---|
| `hub.featured(max=3)` | Parent route reads this; sheet is downstream |
| `hub.search(q)` | Parent; not needed by sheet |
| `hub.detail(id)` | Parent; preview is a synthesized card from local fields |
| `hub.subscribed_list()` | Parent; unrelated to publish |
| `account.status()` | Sheet's Author is pre-filled by parent via prop; sheet doesn't re-fetch |

---

## 7. DESIGN DECISIONS

The load-bearing calls — where this spec earns the MOAT bar.

### 7.1 Sheet, not modal — MOAT §2.8

**The problem.** The legacy `PublishModal.svelte` is a centered
`<Dialog>` with `aria-modal="true"`, `backdrop-filter: blur(4px)`,
and click-to-close on the scrim. That's the wrong shape: publishing
is a task with multiple fields, a file picker, validation, and a
submit. It's not a confirmation dialog. It's also the exact
glassmorphism pattern MOAT §4 #3 forbids on cards/lists/inline
surfaces.

**What this spec does.** `.c-sheet` sliding up from the bottom edge
of the viewport. No scrim — the route body stays live behind it.
Esc closes. Focus returns to the parent Publish CTA pill on close.
The taxonomy MOAT §2.8 mandates is honored verbatim.

### 7.2 Two-source rule — preview = the live card

**The problem (per MOAT §1.5 — "no second card vocabulary").** The
legacy `PublishModal.svelte:198-201` renders a `<pre class="preview-
yaml">` block — a raw YAML dump of `id`, `name`, `version`, etc.
That's two preview shapes for one decision (the user sees the YAML
*and* the eventual Hub card are two different things; they aren't
connected). The Hub's card vocabulary lives elsewhere (SCREEN_HUB
§3.2). One source of truth = one shape.

**What this spec does.** The right column is a **`<LivePreview />`**
that renders the same `<SkillCard />` micro-shape the Hub's grid
uses (the SCREEN_SKILLS §3.2 card, public variant). Every field
change re-renders it via a 200ms cross-fade. The user sees what
they'll get, not a YAML dump. The legacy YAML preview is deleted.

### 7.3 Safety scan is shown, not hidden

**The problem (per DIRECTION §1 — "I3 Smooth is honest").** A
publish flow that hides the safety scan is a publish flow that
pretends it isn't scanning. The user uploads a `.zip`, presses
Publish, and either sees a result or sees nothing — the *why* is
black-boxed.

**What this spec does.** The scan runs **before** the Publish CTA
enables (S3 → S4 → S3). While in flight, a `<Thread />` draws across
a 360px hairline at the bottom of the sheet, with
`<Pulse phase="acting" />` + mono-11 `"Scanning for promptware…"`.
On rejection (S5), the mono-11 line above the file picker reads the
exact `reason` from `hub.safetyScan` (no euphemism: it's
`shell-metachar-in-script`, not "an issue was found"). The scan is
the moment where the daemon earns the trust of being told what its
findings are.

### 7.4 Honest "Done", not fake celebration

**The problem (per MOAT §4 #6).** No "Awesome!" / "Perfect!" /
"You're all set!" toast on success. The legacy `PublishModal.svelte:
139` uses the literal `result.ok` text "Your skill is live" with an
anchor link — that's the right shape, but the `result.ok` block is
inline inside the modal. After the thread draws, success should look
like *the agent finished*, not *a celebration modal*.

**What this spec does.** On S7: Thread draws across the sheet's top
edge (the **completion** gesture from MOAT §3 — *a connection was
made*); a `<Glyph name="check" />` 16px filled `--ok` stamps center-
bottom of the form column; the footer CTA morphs from `Publish →` to
`Done →` (ghost pill, not primary); a single mono-11 line
(`Your skill is live → https://hub.condura.app/skills/<id>`)
replaces the safety-scan progress region. No "Awesome!" No toast. No
emoji. The card on the Hub's grid picks up the new entry on the
parent route's next 60s poll.

### 7.5 No gradient text, no emoji, no rainbow

Per MOAT §4 #1, #2, #3: no gradient text; all icons via
`<Glyph />`; no purple, cyan, teal, pink, yellow-outside-pollen. The
trust dot in the preview is `--synapse` filled (`Official` is the
only trust level at *upload* time — the author is publishing into
the curated bucket only after the safety scan passes; the Hub's
trust taxonomy lives in SCREEN_HUB §8.3).

### 7.6 Validation on blur, not on keystroke

**The problem.** Per DIRECTION §6 ("focus rings track the shape"),
the focus ring is geometry-aware and quiet. Validation should be the
same. Pummeling the user with an inline error on every keystroke of
"1.0.0" — "1", "1.", "1.0", "1.0.", "1.0.0" — teaches them the
field is hostile, not that semver is precise.

**What this spec does.** Version validates on **blur** (the same
moment focus leaves the field). Until then, the field is neutral. On
blur, if invalid, the inline `<Glyph warning/>` + mono-11 error
appears; the Publish CTA stays disabled with a `<Tooltip>` label
`Fix semver first`. The check itself (`semver.test(...)`) is
unchanged from the legacy regex.

### 7.7 File picker is drag-and-drop, not "<input type=file>"

**The problem.** The legacy `PublishModal.svelte:193` is a raw
`<input type="file" accept=".zip,application/zip">` — functional, but
plain. MOAT §1 ("RESTRAINT TEST") defines premium products as the
ones that care about how a thing is interacted with, not just whether
it works. A drop-zone that highlights on `dragenter`, accepts `.zip`
on drop, and shows `{filename} · {KB}` after is a single-stroke
detail that does not require new code, only careful CSS.

**What this spec does.** `<FilePicker>` is a `<Button variant="ghost">`
that opens the native dialog AND a drop-zone that accepts files. The
spec re-uses the published-existing CSS hook conventions (rounded
hover, synapse halo on focus) and adds the `dragenter` border-color
shift per §3. No new component is invented here; the `<FilePicker>`
shape is identified for the team to extract from `Settings.svelte`'s
existing uploads (e.g., the backup-zip file picker in the Legal
section).

---

## 8. DRIFT TABLE — what this spec changes vs. the legacy `PublishModal.svelte`

The implementation must apply every row below in one commit. Each
row cites the MOAT / DIRECTION / APPFLOW / SCREEN_HUB finding.

### 8.1 Removed

| Legacy line(s) in `PublishModal.svelte` | What it does | Why it goes |
|---|---|---|
| `:1-12 imports Dialog, Button, Input, Textarea, Select` | Centered modal scaffolding | MOAT §2.8 — `<Sheet>` replaces `<Dialog>` |
| `:121-222 <Dialog open={isOpen} size="lg">` | Centered modal with scrim | MOAT §2.8 + §4 #3 — no scrim on a sheet |
| `:193 <input type="file">` (plain) | Native-only archive picker | §7.7 — `<FilePicker>` adds drag-drop + size metadata |
| `:198-201 <pre class="preview-yaml">` | Raw YAML preview | §7.2 — `<LivePreview>` card is the one source |
| `:228-232 .grid-three { grid-template-columns: 1.2fr 0.7fr 0.9fr; }` | Three-column field row | §1.1 — two-column layout (form + preview) |
| `:235-238 .grid-two { grid-template-columns: 1fr 1fr; }` | License + Tags side-by-side | §1.1 — stacked inside the form column |
| `:295-310 .result.ok` / `.result.err` inline blocks | Inline success/error messaging inside the modal | §7.3 + §7.4 + MOAT §1.2 — error/success move to sheet states S5/S7/S8 and route-level `<ErrorState />` |
| `backdrop-filter: blur(4px)` on the scrim | Glassmorphism | MOAT §4 #3 — explicit ban |
| `t('hub.publish.*')` locale strings for the title/headings | Locale for the modal | The sheet keeps the same keys (i18n cost is zero — the migration is layout-only). |

### 8.2 Added

| What is added | Where | Why |
|---|---|---|
| `<Sheet>` (`.c-sheet`) primitive | `lib/condura/Sheet.svelte` (new file) | MOAT §2.8 — three named primitives, this is one. |
| `<LivePreview>` mini skill card | `lib/condura/LivePreview.svelte` (new file) | §7.2 — one card grammar across Hub + Skills + Publish preview. |
| `<FilePicker>` drag-and-drop primitive | `lib/components/ui/FilePicker.svelte` (extract from `Settings.svelte`) | §7.7 — drag-drop is the default, native picker is the fallback. |
| S4 (safety scan in progress) state | Sheet-local reactive state | §7.3 — scan is shown, not hidden. |
| S7 (uploaded) Thread + check stamp | Sheet mount point | §7.4 + MOAT §3 — completion gesture is the Thread. |
| `<Tooltip>` on the disabled Publish CTA | `lib/condura/Tooltip.svelte` (new, per MOAT §2.9) | §4.1 — disabled CTAs explain *why*. |
| `hub.safetyScan` IPC | `lib/ipc/client.ts` + `…/ipc/types.ts` | §6.1 — scan runs *before* Publish. |
| Inert `<body>` + `:focus-visible` halo on the parent Publish CTA pill | Sheet mount / unmount | Focus management (§4.2) — sheet is the modal-less variant; the parent pill stays focusable from behind. |

### 8.3 Unchanged

| What stays | Why |
|---|---|
| Semver regex `/^\d+\.\d+\.\d+(?:[-+].+)?$/` | The legacy validation logic is sound; only its trigger moves to `on:blur`. |
| 32 MB archive cap | Daemon-side and sheet-side both enforce it. |
| License dropdown options (MIT, Apache-2.0, BSD-3-Clause, GPL-3.0, ISC, Proprietary) | Existing list; the sheet uses the same `<Select>`. |
| `hub.publish(payload)` payload shape | Already on the wire; the sheet's submit calls the same. |
| Esc / Cancel close behavior | Sheet semantics; same keybinding. |

### 8.4 Naming reconciliation

| Surface | File / component | Why |
|---|---|---|
| Spec name | **`SCREEN_PUBLISHMODAL.md`** (legacy name preserved) | The spec indexes under the existing file name; the route is already searchable by it. |
| Component name | **`PublishSheet.svelte`** (new, replaces `PublishModal.svelte`) | The taxonomy is `.c-sheet`; the component matches the primitive. |
| Legacy alias | `PublishModal.svelte` deleted after one release (per SCREEN_HUB §6.5 migration note) | No double vocabulary. |

---

## 9. Closing note

Six things you'll find when you implement from this spec — and
eight things you'll ship.

**The six invariants:**

1. **No scrim.** Publish is a sheet, not a modal. The route body
   stays live behind it.
2. **No glassmorphism.** Paper-on-paper for the sheet, route, and
   preview card. No `backdrop-filter: blur`.
3. **No raw YAML preview.** The right column is the live Hub card,
   re-rendered via 200ms cross-fade on every field change.
4. **No hidden safety scan.** The scan runs *before* Publish enables;
   a `<Thread />` + `<Pulse acting/>` + mono-11 label make the moment
   visible. On rejection, the *exact* `reason` is shown.
5. **No fake celebration.** On success: Thread draws (the
   *connection* gesture), `<Glyph check />` stamps, CTA morphs to
   `Done →`, mono-11 line reads `Your skill is live → {url}`.
6. **No inline err block inside the sheet.** Errors live in the
   sheet's state matrix (S5 / S8). Route-level errors use
   `<ErrorState />`. No fifth copy of the err pattern (MOAT §1.2).

**The eight deliverables:**

1. `.c-sheet` primitive (`<Sheet>`) — slides from bottom, focuses the
   first field on open, returns focus to parent CTA on close.
2. `<PublishSheet />` — eight form fields (`§1.1`) + live preview
   column + footer CTAs + safety-scan progress.
3. `<LivePreview />` — the SCREEN_SKILLS card mini-shape, cross-
   fading on field change.
4. `<FilePicker />` — drag-and-drop + native fallback; mono-11
   `{filename} · {KB}` after selection.
5. `hub.safetyScan({ archive, filename })` RPC + sheet states S3/S4/S5.
6. S7 upload-success path: Thread + `<Glyph check />` + Done pill +
   mono-11 live-URL line.
7. Keyboard surface (`§4`): Tab/Shift+Tab, Cmd+Enter, Esc-with-S7-delay,
   semver-on-blur.
8. `prefers-reduced-motion` honored via the single global rule in
   `condura.css` (MOAT §2.3); the sheet declares zero media-query
   blocks.

If during implementation you find a place where this spec is silent
and the default would be wrong, ask before inventing. The MOAT bar
is "premium-quality." Anything that fails it is competent-but-
generic — and competent-but-generic is how a year gets lost.
