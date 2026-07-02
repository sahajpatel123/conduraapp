# SCREEN_FLOATINGINTERVIEW — Popover interview (clarifying ambiguity)

> **Surface:** lightweight helper popover that appears **only when needed** —
> ambiguity detected in the user's request, or the user explicitly opts in
> ("help me shape this"). Never unsolicited. Anchored to the composer (or a
> settings field, or a skill description) via a hairline Thread. The eight
> tests below describe the structural skeleton, the state, the motion, the
> keyboard, the components, the data, the design rationale, and the drift.

> **Why:** FloatingInterview.svelte today is the 9-step onboarding wizard
> (eula → permissions → hotkey → complete), routed via the daemon's onboarding
> state machine. This spec describes the *different* surface in the same
> conversational family — a 2–4 question clarifier that shapes an ambiguous
> ask into something the agent can act on, anchored to the context it
> qualifies. Two surfaces, one component vocabulary.

---

## 1. LAYOUT & CONTENT

A small paper-card popover anchored to the calling context (composer top-end,
settings field ring, skill description pencil). Two to four questions max;
never a wizard. One question visible at any moment. Below the question, a
free-text input or a 3–5 chip picker — depending on what the question
demands. Above the question, a thin Thread hairline draws left→right to
visually tie the popover back to its anchor.

| Region | Content | Notes |
|---|---|---|
| **Anchor thread** | 1.25px synapse line connecting the popover to the context | DIRECTION.md §3 — the Thread; never an arrow |
| **Eyebrow** | "shaping this — {N} of {TOTAL}" in JetBrains Mono 10px, +0.100em, uppercase, `--content-faint` | Only visible if `TOTAL > 1` |
| **Prompt** | Question in Inter 17px / 1.55 / −0.010em, `--ink-soft` | Loaded from `interview.start` or per-question stream |
| **Free-text input** | `<Input>` primitive (paper-sunken well, 1px hair border, 2px synapse focus ring on rounded 10px radius) | One input only; multi-line hidden unless question opts in |
| **Chip picker** | 3–5 `<Chip>` rows, `--r-pill`, JetBrains Mono 11px, `--content` on `--paper-2`; selected chip = `--pollen` on `--pollen-light` | Visible only when `chips.length > 0` |
| **Progress dots** | Inline left-aligned, 6px circles, `--hair` resting, `--pollen` filled current, hairline ring after current | One dot per question; never a percent bar |
| **Skip** | Bottom-left, JetBrains Mono 11px, +0.12em uppercase, `--content-faint` → `--ink` on hover | Closes popover; question is *remembered* as skipped in the summary |
| **Next →** | Bottom-right, pollen primary `<Button>`, magnetic | Disabled until the active question has a value (free-text non-empty *or* chip selected) |
| **Generate →** | Bottom-right, replaces Next on the **last** question | Same disabled rule; cargo "submit, don't continue" |
| **Footer hint** | `↵ to continue · Esc to close · Tab between fields` in JetBrains Mono 10px, `--content-faint` | Hidden when an active question has focus |

### Spacing tokens (read from `condura.css`, never invent new)

| Element | Token |
|---|---|
| Popover width | `min(420px, 92vw)` |
| Card padding | `--space-7` horizontal, `--space-6` vertical |
| Internal gap (eyebrow → prompt → input → dots → footer) | `--space-4` |
| Chip-row gap | `--space-2` |
| Footer separator (above Skip / Next) | 1px `--hair` (the only sanctioned divider) |
| Anchor-thread offset from popover edge | `--space-4` |

### Information hierarchy (top to bottom)

1. **Eyebrow** — *where am I in this conversation?* (micro-meta)
2. **Prompt** — *what is being asked?* (the only thing the user must read)
3. **Input or Chips** — *what do I do?* (one of two affordances)
4. **Progress dots** — *how much remains?* (ambient; not a clock)
5. **Footer** — *Skip · Next · keyboard hint* (the actions)

---

## 2. STATE MATRIX

Eight reachable states. Each has a defined rendering path; nothing falls to
a dead wall. The default for any unknown error is **honest degradation** —
`<ErrorState>` with the err-hair, the italic headline, and a single retry.

| # | State | Popover visibility | Anchor | Affordance | Trigger | Exits to |
|---|---|---|---|---|---|---|
| 1 | **closed** | hidden | normal | — | initial / Esc / outside-click / Skip / Generate done | open-q1 |
| 2 | **open-q1** | mounted, scrolled to top | active | input focused, placeholder blinking | context detected ambiguity or user opted in | open-qN / skipping / done / error |
| 3 | **open-qN** | mounted, question N visible | active | input + chips (per question type) | Next pressed, free-text value, chip selected | open-qN+1 / skipping / done / error |
| 4 | **skipping** | mounted, current fades + next slides in | active | nothing interactive | Skip pressed | open-qN+1 / done (if last was skipped) |
| 5 | **done** | collapsing (paper-fold + 8px y-down + thread draws to anchor) | thread tying popover to context (250ms `--ease-in`) | replaced by summary thread anchoring the popover's footprint | Generate pressed, all questions answered or skipped | closed |
| 6 | **error** | mounted, eyebrow → "couldn't shape this", italic headline + err-hair + retry | unchanged | retry button + Skip still works | `interview.start` or `interview.next` failed | retry → open-q{N} |
| 7 | **shadow** (input blurred but value retained) | mounted, focus halo deintensifies 280→140ms | unchanged | unchanged | focus lost | open-qN |
| 8 | **hover-chip** (chip picker hover) | unchanged | unchanged | chip border `--hair-strong`, scale(1.02), translateY(-1px) | pointermove over a chip | open-qN |

### Reachable combinations (audit row)

| Trigger | New state | Side effects |
|---|---|---|
| Context calls `interview.start(anchorContext)` | open-q1 | footer scrolls to bottom, input gets focus next tick |
| User types in free-text | open-qN (shadow→focused) | Next enabled when `value.trim().length > 0` |
| User picks chip | open-qN (shadow→focused) | chip set, Next enabled, *no auto-advance* (user still confirms) |
| User presses Enter | open-qN → open-qN+1 (or done) | `interview.next(answer)`; fade + slide transition |
| User presses Esc | closed (if value empty) or open-qN → done-with-current-as-skipped | `interview.skip({ q: current, reason: 'user_escape' })` |
| User clicks outside | closed | values retained in the conversation's draft context for 90s |
| User presses Skip | skipping → open-q{N+1} or done | `interview.skip({ q: current })` |
| Free-text submit on last question | open-qN → done | `interview.complete({ answers: [...] })` |
| Chip select on last question | open-qN → done | `interview.complete({ answers: [...], final: 'chip' })` |
| Error from any RPC | open-qN → error | inline retry only; never silently re-fires |

---

## 3. MOTION CHOREOGRAPHY

Read `DIRECTION.md §5` first. Every duration below resolves to one of the four
durations (`--dur-fast` 140ms / `--dur` 280ms / `--dur-slow` 520ms / `--dur-cine` 900ms).
Every ease resolves to one of three (`--ease` / `--ease-in` / `--ease-pop`).

| Motion | Property changed | Duration | Ease | Trigger |
|---|---|---|---|---|
| **Popover open** | `opacity 0 → 1` + `transform: scale(0.96 → 1)` + `transform-origin: var(--anchor-x) var(--anchor-y)` | 200ms | `--ease-out` | `interview.start` |
| **Anchor thread draw** | `stroke-dashoffset 1 → 0` (SVG path, popover edge → context edge) | `--dur-slow` (520ms) | `--ease-out` | popover open |
| **Question out (left)** | `opacity 1 → 0` + `translateX(0 → -8px)` | 160ms | `--ease-in` | Next / Skip pressed |
| **Question in (right)** | `opacity 0 → 1` + `translateX(8px → 0)` | 200ms | `--ease-out`, 60ms stagger | after out completes |
| **Progress dot fill (current)** | `transform: scale(0.8 → 1)` + `background --hair → --pollen` | `--dur` (280ms) | `--ease-pop` (one-time per surface) | question change |
| **Progress dot ring (already-answered)** | `box-shadow: 0 0 0 1px --synapse` (the synapse armor `rect` for `MOAT` §2.1) | 140ms | `--ease` | answer accepted |
| **Chip hover** | `transform: translateY(-1px)` + `border-color: var(--hair-strong)` | `--dur` | `--ease` | pointerenter |
| **Chip select** | `background: var(--paper-2) → var(--pollen-light)` + `border-color: var(--pollen)` | `--dur` | `--ease` | click / Space / 1–5 key |
| **Next → enable** | `transform: scale(0.97) → 1` + pollen halo brightens (pulse once, 280ms) | `--dur` | `--ease-pop` (the one allowed pop per surface) | value entered |
| **Popover collapse (done)** | `transform: scale(1 → 0.94)` + `opacity 1 → 0` + `translateY(0 → 8px)` | `--dur-slow` | `--ease-in` | Generate pressed |
| **Summary thread (done)** | synapse path **draws from popover's last point to the context anchor** then **the popover dissolves**, leaving the thread only | `--dur-slow` then `--dur` | `--ease` then `--ease-in` | done → thread anchored |
| **Error entry** | eyebrow swaps to italic Instrument Serif 22px, err-hair `stroke-dashoffset 1 → 0` | `--dur-slow` | `--ease` | RPC reject |
| **Pollen mote (idle, q1 of TOTAL>1)** | one 8px mote drifts from top-right corner toward the Skip button, 1.6s, then dissolves; never recurs until question changes | 1600ms | `linear` | popover idle > 4s |
| **Pulse on Next when disabled** | none — **disabled means no performance** | — | — | — |
| **Reduced-motion: skip scale & slide** | popover open = `opacity 0 → 1` only, duration 120ms; no slide; no dot-pop; thread still draws (it's meaning, not ornament) | 120ms / `--dur-slow` | `--ease` | `prefers-reduced-motion: reduce` |

### Rule of one (`MOAT.md §4 rule 4`)

One metaphor per component. The Thread carries completion. The Pulse carries
aliveness. The pollen mote carries *rest*. There is **no spinner** here
(MOAT §4 rule 7). Loading is a thread drawing in or a thread backwashing out
— never `↻`, never three dots.

---

## 4. KEYBOARD

Every key reads from `condura.css :focus-visible` with the shape-tracking
halo from `DIRECTION.md §6 rule 6`. The popover owns focus on open,
restores focus to the anchor on close (the calling context node receives
`document.activeElement = anchor`).

| Key (in popover scope) | Action | Notes |
|---|---|---|
| **Enter** | advance to next question (or Generate on last) | IME-safe via `event.isComposing` guard |
| **Shift+Enter** | newline in free-text (when input is `<textarea>`) | only when question is multi-line opt-in |
| **Esc** | close (or, if value typed + last question, advance-as-skipped) | never silently discards typed values — preserves them in the conversation draft for 90s |
| **Tab** | move between input, chip-row, Skip, Next | order: input → chips (in order) → Skip → Next; reverse with Shift+Tab |
| **↑ ↓** (when chip-picker focused) | cycle chip focus | loops; pollen-halo follows |
| **1–5** (when chip-picker focused) | quick-select chip N | respects chip count; no-op if N exceeds length |
| **← Back** (Backspace on Android, Delete on Mac, fn-Backspace on iPadOS) | previous question, if not first | only enabled when `q > 1`; preserved answers remounted |
| **?** | open shortcuts sheet (the `?`-bound overlay from `MOAT.md §2.10`) | **not** rendered inside the popover — same overlay as the rest of the app |
| **⌘. (Cmd+Period on Mac / Ctrl+. on Win)** | always closes | the universal "stop" key, app-wide |

### Focus ownership

- On open: `firstFocusable.focus()` next `requestAnimationFrame` (so the
  anchor-thread has a frame to draw). Prefers the free-text input; falls
  back to the chip-row.
- On question change (Next / Skip): focus moves to the **new** question's
  primary affordance (not the Next button — the user is now reading).
- On Generate / collapse: focus returns to **the calling anchor** (composer,
  field, skill description pencil). The Thread remains as the visible
  callback for 8s, then dims to `--hair`.

---

## 5. COMPONENTS USED

Every name below is a file under `app/web/frontend/src/lib/condura/`. No
new primitives invented. No emoji (MOAT §4 rule 2). No gradient text
(MOAT §4 rule 1). No rainbow accents (MOAT §4 rule 3).

| Component | Role here | Source |
|---|---|---|
| **Popover** (primitive) | Card surface + scrim-free anchor + outside-click dismiss | `condura/Popover.svelte` *(to add; merge PairingModal sheet into the taxonomy of MOAT §2.8)* |
| **QuestionPrompt** | The eyebrow + prompt + input/chip-picker composition | `condura/QuestionPrompt.svelte` *(to add)* |
| **ChipPicker** | One row of 3–5 chips; tab/arrow/1-5 keyboard | `condura/ChipPicker.svelte` *(to add; or compose from existing `Chip` rows in `Channels.svelte`)* |
| **Input** | Free-text input with paper-sunken well, `--shadow-focus` rounded halo | `condura/Input.svelte` *(to add; reuse the `Ritual.svelte:230–245` keycap pattern + `Chat.svelte:515–528` composer thread)* |
| **Button** | Primary pollen `magnetic` for Next / Generate | `condura/Button.svelte` *(exists)* |
| **ProgressDots** | Inline 6px dots, current pollen-filled, synapse hairline-ring after | `condura/ProgressDots.svelte` *(to add; or extract the dot triple from `Settings.svelte:614–672`)* |
| **Thread** | The signature hairline that ties popover to anchor | the SVG pattern from `DIRECTION.md §5` |
| **Pulse** | Used **only** in the WaitingChip when the RPC is in-flight (skipping / done / error) | `condura/Pulse.svelte` *(exists)* — never in idle, never in default |
| **Glyph** (icons) | `arrow` for Next, `skip` for the Skip link, `check` for Generate | `condura/Glyph.svelte` *(exists)* — single-stroke, 1.5 weight, currentColor |
| **Tooltip** | `?` shortcut hint chips, hover-previews on truncated chips | `condura/Tooltip.svelte` *(to add per `MOAT.md §2.9`)* |

### Component composition

```
<Popover anchor={...} onclose={...} role="dialog" aria-label="Shaping this">
  <Thread path={anchorPath} />            // MOAT §3: the visual signature
  <QuestionPrompt
    eyebrow={i18n('shaping.this.eyebrow', { i, total })}
    prompt={q.prompt}
    type={q.type}                          // 'text' | 'chips'
    value={answers[q.id]}
    onchange={...}
  >
    {#if q.type === 'text'}
      <Input value={answers[q.id]} placeholder={q.placeholder} autofocus />
    {:else}
      <ChipPicker chips={q.chips} value={answers[q.id]} onchange={...} />
    {/if}
  </QuestionPrompt>
  <ProgressDots total={total} current={i} answered={answered} />
  <footer class="fc-foot">
    <button class="link-skip" onclick={skip}>Skip</button>
    <Button variant="primary" magnetic disabled={!canAdvance} onclick={advance}>
      {#if isLast}Generate{:else}Next →{/if}
    </Button>
  </footer>
</Popover>
```

---

## 6. DATA FETCHED

| RPC | Direction | Payload | Returned | When | Failure mode |
|---|---|---|---|---|---|
| `interview.start` | client → daemon | `{ anchorKind: 'composer' \| 'settings_field' \| 'skill_description'; anchorContext: { route?: string; field?: string; skillId?: string; draft?: string } }` | `{ interviewId: string; questions: Question[]; total: number; suggestionReason: 'ambiguity_detected' \| 'user_opted_in' \| 'low_confidence' \| 'missing_field' }` | popover mounts | throw → state error; popover stays closed |
| `interview.next` | client → daemon | `{ interviewId: string; qId: string; answer: string \| { chip: string } \| null }` | `{ accepted: boolean; nextQuestion?: Question; finished?: boolean; partial?: { summaryDraft: string } }` | each Next press | throw → state error; retry |
| `interview.skip` | client → daemon | `{ interviewId: string; qId: string; reason?: 'user_skip' \| 'user_escape' }` | `{ nextQuestion?: Question; finished?: boolean }` | Skip / Esc | throw → state error |
| `interview.complete` | client → daemon | `{ interviewId: string; answers: Answer[]; durationMs: number }` | `{ accepted: boolean; shapedContextId: string; threadId: string }` | Generate | throw → state error; values preserved in draft |

### Question schema (interview-side)

```ts
type Question = {
  id: string;
  prompt: string;
  type: 'text' | 'chips';
  placeholder?: string;       // only for 'text'
  chips?: string[];           // only for 'chips' — 3..5 entries, ≤24 chars each
  maxLength?: number;         // optional clamp on free-text (default 280)
  required?: boolean;         // default false; "required" just means Skip is disabled
};
```

### Idempotency

`interview.next` carries `qId`, not `index`. If the daemon receives a
duplicate `qId` (e.g. user double-tapped Enter), it returns the cached
`nextQuestion` and bumps a `dedupeCount` in the audit log. **No silent
double-submit.**

### Audit footprint

Every state transition writes one event to `audit` (`internal/audit`):

| Event | Severity | Notes |
|---|---|---|
| `interview.started` | info | includes `suggestionReason` |
| `interview.answered` | info | redacts free-text (PII guard from `internal/perception.PIIRedactor`) |
| `interview.skipped` | info | reason enum |
| `interview.completed` | info | includes `answersCount`, not the answers themselves |
| `interview.failed` | warn | RPC error, never block (per `MOAT` §5 — *help, never lock*) |

No free-text answers land in the audit chain. The shape (`shapedContextId`)
is what downstream components consume.

---

## 7. DESIGN DECISIONS

### MOAT compliance

| Test | Verdict | Evidence |
|---|---|---|
| **§1 Restraint** — what's over-designed? | The 6-keyframe wizard was deleted (`Ritual.svelte`) — see DIRECTION.md §2. This popover has **two** gestures: the Thread (draws) and the pollen mote (one, idle, ≤4s). Nothing decorative. |
| **§2 Detail** — micro-polish gaps | Follows §2.1 shape-tracking halo; §2.2 press states; §2.3 single reduced-motion owner (`condura.css`); §2.4 empty states teach (the popover has no empty state); §2.5 loading states teach (thread drawing); §2.6 errors guide (see state #6); §2.7 tactile vocabulary is `--dur` only; §2.8 collapses `PairingModal` into the `Popover` primitive; §2.9 uses the new `<Tooltip>` (no `title=`); §2.10 keyboard complete (this spec's §4). |
| **§3 Signature** — the Thread | The anchor thread between popover and context *is* the Thread. It draws on open. It persists as the summary callback on done. |
| **§4 What we won't do** — 10 anti-patterns | No gradient text. No emoji (Glyph only). No glassmorphism (paper-card; `--shadow-float` for elevation, not blur). No rainbow. No "Welcome to the future" copy ("shaping this — 1 of 3"). No fake enthusiasm. No spinner (Thread draws instead). No rectangular outline (shape-tracking halo). No double shadows (one token per surface). No animation without meaning (every motion has a row in §3). |

### HELPER contract (the single most important decision)

The popover is a **helper**, not a wizard. It must obey:

1. **It appears only when needed.** Three trigger conditions, all of them
   earned: (a) the daemon flagged `low_confidence` on the most recent draft,
   (b) a required field has no value and the user pressed Send, (c) the user
   explicitly opted in ("help me shape this" — available in `⌘K` palette as
   a *modal-only-when-asked* command). Never on first paint. Never on idle.
2. **It feels like conversation, not form.** Prompts are written in second
   person ("what kind of tone?"), never first person ("please select tone").
   The eyebrow ("shaping this — 1 of 3") is descriptive, not celebratory.
3. **No fake enthusiasm.** "Great question!" never. "Sure, picking that up"
   never. The popover can say "shaping this" or "almost there" — that's the
   ceiling. Success in the summary thread is a draw, not a stamp.
4. **The summary Thread is the only callback.** When Generate collapses the
   popover, what remains is the synapse hairline tying the popover's last
   point to the context anchor. The user sees *the connection was made.*
   This is the same gesture the titlebar carries (`TitlebarThread.svelte`),
   so the vocabulary is consistent app-wide.

### Voice of the prompts (paper, not chrome)

| Question type | Voice example | Anti-pattern (forbidden) |
|---|---|---|
| Tone | "what kind of voice — formal, casual, or terse?" | "Please choose your preferred tone" |
| Audience | "who's this for?" | "Please specify your target audience" |
| Format | "a memo, an email, or just notes to yourself?" | "Select output format" |
| Length | "how long — a paragraph, a page, or a one-liner?" | "Enter desired length" |
| Trigger | "what should prompt this — a hotkey, a slash, or nothing?" | "Configure trigger method" |

The first word is always a lowercase interrogative or "what/how/which" —
matches the existing Chat composer voice (`Chat.svelte:470–520`).

### Anchor contract

| Anchor kind | Where the thread ties | Special handling |
|---|---|---|
| **composer** | top of `<Input>` in `Chat.svelte`, just above the model `<select>` | thread extends right→left to the model's pollen dot |
| **settings field** | right edge of the focused input | thread extends down-right; field gets a synapse hairline ring instead of its focus halo |
| **skill description** | left of the `<Glyph name="edit" />` pencil | thread extends up-left; pencil receives the pollen halo once |

### Anti-features (non-goals)

- **No multi-paragraph answers.** One chip or one short text. If the user
  needs a paragraph, they should not be in the popover — they should be in
  the composer.
- **No backtracking pages.** The popover shows one question at a time,
  remembered in order. Re-edit happens via `Backspace` (see §4); it is
  not a browse-back.
- **No save-state independence.** Closing the popover with `Esc` after
  typing in the first question does *not* save the answer. Closed = gone.
  The conversation's draft retains the in-flight summary for 90s as a
  courtesy, but the per-question answers are ephemeral.
- **No "skip all" button.** Skipping is per question. If the user wants
  to abandon, they hit Esc — the thread tells the daemon nothing was
  shaped, and the original (ambiguous) prompt continues downstream.

---

## 8. DRIFT TABLE

| What was drifted | Removed | Added | Why |
|---|---|---|---|
| `FloatingInterview.svelte` as the **9-step onboarding wizard** (eula → permissions → hotkey → complete) | The wizard-stripping; this spec renames the *intent* to clarify. Two surfaces, one vocabulary. | — | The existing file is correct (it renders the real `onboarding.*` RPCs); it's just miscategorized as "interview." This spec is the *shaping interview* — the helper popover — not the first-run wizard. |
| `PairingModal.svelte` (a `.c-sheet`) | inline `slide-from-right`, four ad-hoc overlays (`Skills.svelte:108–127`, `Channels.svelte`, `PairingModal`, `ConsentModal`) | one `Popover.svelte` primitive in the taxonomy from `MOAT.md §2.8` | One anchor primitive for anchor surfaces (popovers, tooltips, hint cards). |
| `↻` rotation animation (any spinner fallback) | never existed; banned by `MOAT` §4 rule 7 | Thread draws in on the only loading case (RPC in-flight) | "Thread = connection made"; a spinner says "wait." |
| Free-text answers cached in `localStorage` | incidental pattern that could PII-leak | ephemeral answers; only `shapedContextId` persists | The popover is a helper, not a notebook. PII redaction owned by `internal/perception.PIIRedactor` runs before audit write. |
| The original one-line eyebrow "Step 1 of 4" (uppercase, JetBrains Mono) | The 22% of `FloatingInterview.svelte:199` that wrote `Step {i} of {N}` in title-case shell | Lowercase descriptive eyebrow ("shaping this — 1 of 3") | Per `MOAT` §4 rule 5 ("Configure, not comply"); the eyebrow is ambient, not a progress meter. |
| Progress as a **percent bar** (`.fc-prog` SVG line in `FloatingInterview.svelte:281–287`) | the 1.5px `synapse` hairline + `step-count` | inline `ProgressDots` (one per question, pollen-filled current) | A bar reads as a timer; dots read as ambient acknowledgment. Per `DIRECTION.md §5` (no decorative ornaments). |
| `link-back` (back-arrow text) | kept as text-only if `q > 1`, but rendered via `Backspace` only | `← Back` exposed as a chip in the footer only when question > 1 | Matches `MOAT` §2.10 keyboard completeness; arrows in UI compete with the Thread as a navigation metaphor. |
| `QuickPromptOverlay.svelte`'s **auto-dismiss** (5s timer) | timer for the shaping interview | no auto-dismiss | Helpers should not disappear mid-question. The overlay pattern is for ephemeral sends; the shaping interview is for thinking. |
| `sealBloom` (the radial EULA bloom — `MOAT` §5 reserves it for legal) | a copy attempt to give the popover its own bloom | one `--ease-pop` on Next-enable (the single allowed per-surface pop) | The seal is sacred (`MOAT` §1.7); no other CTA gets a bloom. |
| `wc-thread` SVG style inline | one `.fc-thread` element with hardcoded `M 0 12 L 9999 12` | the canonical `<Thread path={...} />` (MOAT §3 signature) | The Thread is the contract; inline SVGs are the violation. |
| "Accept & continue →" button label (calls `onboarding.acceptEula`) | unchanged — this surface does not own that call | the **Generate →** label, loaded only on the last question | Vocabulary split: "Accept" is legal (`Ritual.svelte:211`); "Generate" is shaping. |
| Pollen mote on every transition | the original two-decoration pattern (think + breathe) | zero pollen in the default state; one idle mote after 4s only when `total > 1` | Per `MOAT §4 rule 10`: every animation must carry meaning. The idle mote carries "tap me if you forgot," nothing more. |
| `eulaScrolled / eulaAccepted` derived state (specific to legal) | nothing — this surface does not own legal | — | EULA is the wizard's job, not the interview's. |
| `permPoll = setInterval(refreshPerms, 2000)` (in `FloatingInterview.svelte:175`) | OS-permission polling, not interview polling | one-shot RPC per question | The interview's questions are daemon-side state; polling OS permissions here would be a conflation of two surfaces. |

### What does NOT drift

- The Thread. Never. (MOAT §3.)
- The `--space-*` tokens (always from `condura.css`, never reinvented).
- The four duration tokens (`--dur-fast` / `--dur` / `--dur-slow` / `--dur-cine`).
- The three ease tokens (`--ease` / `--ease-in` / `--ease-pop`).
- The four shadow tokens (`--shadow-paper` / `--shadow-card` / `--shadow-float` / `--shadow-focus`).
- The Glyph set (single-stroke, 1.5 weight, currentColor).
- The seed palette (paper, ink, synapse, pollen + the four status colors).
- The reduced-motion contract (single owner in `condura.css`).
- The keyboard shortcuts (`⌘.`, `Esc`, `Enter`, `Tab`, `?`, the four chips).

---

**Closing test (from `DIRECTION.md`):**

> *Does this read like a paper notebook that learned to listen — warm,
> awake, and never louder than the room it's in?*

If yes, ship it. The popover is the gentlest helper the product has —
it appears only when the user needs to think, and it leaves a thread
where it stood. That's the whole surface.
