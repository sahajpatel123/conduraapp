# Condura · MOAT

> The premium-quality test for the Condura shell. Read this before opening
> any surface. Every surface in `app/web/frontend/src/lib/condura/` must
> earn the bar set here. Anything that fails it is "competent but generic"
> — and competent-but-generic is how solo founders lose a year.

The five tests below, in order, are how we close the gap between
"a nice local agent" and "the thing you can't stop thinking about."

---

## 1. THE RESTRAINT TEST — what is OVER-designed

These scream "AI generated this" today. They must be removed or sharply
toned down before v0.1.0 ships. A premium product earns its flourishes; an
AI-generated one performs them.

### 1.1 The Ritual has six keyframes to introduce itself — five too many
**File:** `Ritual.svelte` lines 1550–1564. `voidHold`, `moteDrift`,
`wordReveal`, `fadeUp`, `firstBeat`, `breathe` — six named animations
firing in a 3.5-second window just to land the wordmark. The comment
block (1480–1547) is longer than most components in the app.
**Fix:** Keep `wordReveal` + one ambient `breathe`. Delete the rest. The
awakening is two seconds long. The user's attention for a wordmark
reveal is shorter.

### 1.2 The "err-state" block is copy-pasted verbatim into four components
**Files:** `Chat.svelte:213–221`, `Skills.svelte:71–81`,
`Channels.svelte:150–158`, plus the `err-hair-draw` keyframe re-declared
in each (Chat:388, Skills:254, Channels:369). Same five classes,
same 600ms `transform: scaleX(0→1)`, same 120ms delay.
**Fix:** Extract `ErrorState.svelte` with props `{head, sub, action}`.
Delete the four copies. Adds zero behavior; restores the signal that
things were designed, not generated.

### 1.3 The hero/eyebrow/headline/sub block is hand-rolled in every route
**Files:** `Shell.svelte:194–211`, `Chat.svelte:200–211`,
`Skills.svelte:56–63`, `Channels.svelte:94–102`, `About.svelte:83–105`,
six times inside `Ritual.svelte:488–638`. The CSS for `.title` /
`.headline` / `.eyebrow` / `.sub` is duplicated six different ways with
six different pixel sizes.
**Fix:** One `RouteHero.svelte` with size tokens. The hero already has a
type scale; the routes haven't read it.

### 1.4 The `<Cursor />` pixel quill is rendered unconditionally
**File:** `Shell.svelte:185`. Every page, every frame, every OS, the
cursor is a 20×20 SVG data-URI quill — overriding native cursors for
select, text, resize, even loading states. This is 2005-era Flash
energy. A premium product trusts the OS.
**Fix:** Make `Cursor` an opt-in Settings toggle under "Developer". Off
by default. The brand can live in the wordmark.

### 1.5 The Skills card hover does `rotateX(2deg)`
**File:** `Skills.svelte:283`. `transform: translateY(-4px) rotateX(2deg)`
on hover, then `rotateX(1deg)` on `:active`. There is no 3D surface to
read this against — `perspective: 1000px` on the deck (line 268) creates
no actual depth because the cards have no varied Z. This is the "vibe
coded flex" in its purest form.
**Fix:** Delete `perspective`, `transform-style`, and both `rotateX`
calls. The `-4px translateY` alone reads as elevation. The 3D tilt
reads as a designer showing off.

### 1.6 The constellation SVG in `Ritual` is a decorative load
**File:** `Ritual.svelte:413–442`. Six bezier paths, six circles, a
breathing center, draw-in animations on `wired` and `skipped` — all to
visualize which onboarding steps the user has done. The user is doing
the steps. They know which they've done. A checklist with dots would
teach more in 1/20th the code.
**Fix:** Either (a) make the constellation map to the actual current
state of the system (what permissions are granted, what channels
wired, what skills installed) so it teaches, or (b) replace it with a
plain numbered step list. Decorative orbits do not teach.

### 1.7 The signature flourish (italic green "alive" span) is used 6+ times
**Files:** `Shell.svelte:292` (the `.dot` after Condura), `Chat.svelte:205`
("alive"), `Ritual.svelte:490, 579, 634`. The same `<span class="alive">`
trick applied to single words to punctuate a sentence. Once: a wink.
Five times: a tic.
**Fix:** Reserve `.alive` for one place per surface — a single load-bearing
phrase. Everywhere else, write the headline well enough that italic-green
isn't needed.

---

## 2. THE DETAIL TEST — 10 micro-polish gaps

Each gap is listed with the exact code to change. None of these are
inventions; each is a pattern premium products ship, and Condura either
doesn't or does wrong.

### 2.1 Focus rings must track rounded shapes, not render as rectangles
**Today:** The universal `:focus-visible` (condura.css:299–302) uses a
single `box-shadow: var(--shadow-focus)` which is `0 0 0 4px pollen-halo,
inset 0 0 0 1px inset`. For square inputs (textareas, the keycap on
hotkey capture) this glows a square halo around a square — wrong.
**Fix:** For elements with `border-radius ≥ 8px`, replace the halo with
`box-shadow: 0 0 0 2px var(--synapse), 0 0 0 5px var(--pollen-halo)`.
For pill-radius ≥ 999px, drop the inset 1px line entirely and let the
synapse ring carry the focus state. Audit `Button.svelte`'s
`:focus-visible` (lines 75–80) — it currently inherits the same rectangular
halo.

### 2.2 Press states need weight, not just shrinkage
**Today:** Every `.btn:active` and `[role=button]:active` shrinks to
`scale(0.97)` (condura.css:321–325; 11 components redeclare the same).
A smaller element looks "smaller" — it doesn't feel pressed.
**Fix:** Add `filter: brightness(0.95) saturate(1.1)` to the global
tactile `:active` rule alongside the scale. The same component also
gets a 1px `translateY(0.5px)` so the surface visibly settles into the
page. Premium touch screens are doing this for free; the web has to
choose to.

### 2.3 `prefers-reduced-motion: reduce` is respected inconsistently
**Today:** condura.css:469–476 sets `animation-duration: 0.01ms` and
hides grain/motes. But `Skills.svelte:513–521`, `Channels.svelte:444–451`,
`About.svelte:409–418`, `Chat.svelte:391–396` each re-declare their own
reduced-motion overrides — and the Rules of the house say they
shouldn't. `TitlebarThread.svelte:12` short-circuits the rAF loop on
`prefers-reduced-motion`, but `ReactiveVisibility` `IntersectionObserver`
(line 58–66) and `Page Visibility` (48–54) each have their own logic.
**Fix:** One `@media (prefers-reduced-motion: reduce)` block in
`condura.css` does the whole work via `*` selectors and attribute
driven animation suppression. Components never repeat the media query.

### 2.4 Empty states must teach, not decorate
**Today:** `Chat.svelte:185–212` (empty state) is a garden of motes plus
a hero titled *"Your computer, alive."*. Beautiful. It doesn't teach
the user what to do next. `Skills.svelte:82–86` says "No skills yet" +
"Run a complex task — Condura will save the procedure as a skill
automatically" — this one is correct. Pattern is broken across the app.
**Fix:** Every empty state's copy has three lines, always:
(1) what this area is, (2) why it might be empty, (3) the one action
that fills it. Example for Chat: *"A quiet place to type to Condura.
Ask it to draft a memo, find a file, summarize a PDF, or watch your
screen. Hotkey to summon it from anywhere."* with a single input focus.

### 2.5 Loading states must feel alive, not generic
**Today:** All loading states (`Skills.svelte:65–69`, `Channels.svelte:147`)
are a `<Pulse phase="thinking" size=8>` plus a fixed uppercase mono
label ("INDEXING…", "PROBING REACH…"). The label is honest. The Pulse
is generic — a single breathing dot.
**Fix:** Each loading state carries a *thread* animation that draws in
from left-to-right over 1.6s (use the existing `drawthread` keyframe).
Connect threads tell the user "data is moving." A single breathing dot
tells the user "data might be moving." The difference is the loom.

### 2.6 Error states must guide, not poeticize
**Today:** Every error uses Instrument Serif italic 22px headline +
15px sub. Lovely. `Chat.svelte:217`: *"We couldn't reach the daemon."*
followed by *"The thread stopped mid-sentence."* The user doesn't know
which subsystem (daemon? provider? network?), what retry will do, or
whether their data is safe.
**Fix:** Keep the serif headline. Add three lines below: (1) **what
failed** in one noun (e.g., "Connection to daemon"), (2) **likely cause**
in one phrase ("daemon was restarted"), (3) **next action** as a button
("Restart daemon", "Open Settings → Connection"). All error states in
Chat / Skills / Channels / Settings collapse to one `ErrorState`
component.

### 2.7 The tactile vocabulary must be one thing everywhere
**Today:** `condura.css:309–325` defines the global `.tactile` class
with one transition list. But nine components re-declare their own
`transition: transform var(--dur) var(--ease), background var(--dur)
var(--ease), …`. `Button.svelte:52–58` redeclares; `Channels.svelte:224–229`
redeclares; `Skills.svelte:277–281` redeclares; the ritual redeclares.
This is the press feedback the user is supposed to learn once.
**Fix:** Components that want tactile behavior use `class="tactile"`.
The CSS owns the transition. Components are free to add *meaning* —
a tinted shadow on a primary press, a thread stroke that brightens on
a card lift — but they do not duplicate the timing.

### 2.8 Three overlays exist; the code uses two names for them
**Today:** Modal vs sheet vs popover are not defined anywhere in
codebase. `PairingModal.svelte` is a sheet (slides from right,
`Sync.svelte`). `Skills.svelte:108–127` builds a slide-from-right sheet
inline with `class="overlay"` + `class="sheet"` + `role="dialog"
aria-modal="true"`. `Channels.svelte` re-implements a different overlay
inline for its detail panel. `ConsentModal.svelte` exists separately.
No taxonomy.
**Fix:** Three named primitives in `condura.css`: `.c-modal` (centered,
focus-trapped, blocks page, dismiss-on-Esc), `.c-sheet` (slides from
edge, doesn't block page scroll, Esc to close), `.c-popover`
(anchored, small, dismiss-on-outside-click). Each gets a Svelte wrapper
that owns focus, keybindings, and aria semantics. The four ad-hoc
overlays in `Skills`, `Channels`, `PairingModal`, `ConsentModal`
collapse to these three.

### 2.9 Tooltip vs popover vs sheet must be distinct
**Today:** No component in condura/ renders a tooltip. Some surface
hints via `title="..."` (e.g., `Shell.svelte:201`, `PairingModal`'s
`title="Command palette"`). Native tooltips are slow, system-themed,
and on macOS they appear below the cursor — defeating the placement.
**Fix:** Add a `<Tooltip label>` component: hover-delay 400ms, exit
75ms, `aria-describedby`, one-line max, no rich content. Use for icon
buttons in the titlebar (theme toggle, kill switch), for truncated
session names, for hash-route hints. Stop using `title=` for anything
that needs to be readable.

### 2.10 The keyboard story is incomplete
**Today:** `Shell.svelte:110–125` registers ⌘K + Shift+O + Shift+P
globally. That is the entire keyboard surface. There is no `⌘,` to
open Settings, no `⌘[` / `⌘]` for back/forward nav, no `g` + `g` /
`s` / `h` style two-key shortcuts, no visible focus on the titlebar
thread, no `?` to open the shortcut sheet, no Escape keybinding at the
shell level to dismiss overlays.
**Fix:** Ship a single `Shortcuts.svelte` overlay triggered by `?`:
- `⌘K` palette
- `⌘P` quick-prompt
- `⌘,` Settings
- `⌘[` back, `⌘]` forward through hash history
- `g` then `s` / `h` / `a` / `c` / `k` Go-to Surface
- `Esc` dismisses the topmost overlay
- `Tab` reaches every interactive element with `focus-visible` halo
- The overlay's table renders with the same `--r-md` card style and
the same mono-uppercase label pattern as the rest of the chrome.

---

## 3. THE SIGNATURE — the one unmistakable thing

### The Thread.

A one-pixel hairline that draws in from left to right, in `--synapse`,
200ms `--dur-slow`, `cubic-bezier(0.22, 1, 0.36, 1)`. That is all.

The thread is the visual grammar the user learns once and sees
everywhere: in the titlebar (`TitlebarThread.svelte`), between chat
turns (`Chat.svelte`'s `Thread` between messages), on the channel
audit-link underline (`Channels.svelte:423–438`), on the About colophon
hairlines (`About.svelte:301–317`), on the spirit-line that follows a
pressed `Send` button on the composer (`Chat.svelte:515–528`), and on
the breathing spine that lives at the bottom of the Ritual.

It is one CSS line repeated:

```css
.condura-thread line,
.rule-ink,
.err-hair {
  stroke: var(--synapse);
  pathLength: 1;
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  transition: stroke-dashoffset var(--dur-slow) var(--ease);
}
```

When a moment needs weight — a message lands, a permission is granted,
a step completes, an error resolves — the thread draws in. Not
animates, not fades, *draws*. The motion is purposeful: a connection
being made. The cadence is `--dur-slow` (520ms) — slow enough that the
user registers the gesture, fast enough that nobody waits for it.

**Commit to it everywhere.** A new surface must ship at least one
thread. A new error state uses `err-hair`. A new completed action draws
the thread. There is no other allowed flourish for "this is now
finished."

A premium product earns one element. This is ours.

---

## 4. WHAT WE WILL NOT DO — 10 anti-patterns, non-negotiable

Any of these landing in a PR is a reason to reject the PR. No
exceptions, no "but it's just this once."

1. **No gradient text.** Anywhere. Ever. The brand voice is paper and ink.
   Gradient text is a 2017-portfolio-site tic.
2. **No emoji as UI icons.** Use `<Glyph name=...>` from `icons.ts`. If we
   need an icon that isn't in `icons.ts`, add it to `icons.ts`, do not
   reach for `🚀`. The Glyph set is single-stroke, 1.5 weight, currentColor.
3. **No glassmorphism unless elevation is earned.** A floating surface
   that needs to look elevated (CommandPalette, Settings sheet, dialog)
   uses `--shadow-float`. Period. `backdrop-filter: blur` is prohibited
   on cards, lists, and inline surfaces — `Channels.svelte:361`'s
   `.overlay` `backdrop-filter: blur(4px)` is the kind of thing we
   should delete, not add.
4. **No rainbow accents.** Brand has `synapse` (green), `pollen`
   (orange), and `--content*`. Status uses `--ok`, `--warn`, `--danger`,
   `--info`. There is no purple, no cyan, no teal. Adding a new color
   requires a CLAUDE.md amendment.
5. **No "Welcome to the future" copy.** No "Let's get started", no
   "Welcome aboard", no "Your AI journey begins". Write copy that
   tells the user what this surface does and what to do next. Period.
   The empty states in §2.4 show the shape.
6. **No fake enthusiasm.** No "Awesome!", no "Great choice!", no
   "Perfect!" in a toast, no "You're all set!" celebration modal. The
   agent may register *satisfaction* in one word of body copy if it is
   earned ("Done — sent."). Never in a UI affordance.
7. **No spinner loaders.** No `<Spinner />` component, no `↻`, no
   three-dots variant, no `LoadingSpinner` import. Loading shows a
   thread drawing in (§2.5) or a hash-route pulse. The momentary
   ambiguity of "is this loading or did it finish?" is a spinner smell.
8. **No rectangular focus outlines.** Not in code, not in screenshots,
   not in the spec. The pollen halo (§2.1) or the synapse ring — nothing
   else. `outline: 1px solid var(--content)` is the deepest failure.
9. **No double shadows.** Components stack `--shadow-card` + a
   one-off `box-shadow: 0 8px 20px …`. Three components already do
   this in hover (`Ritual.svelte:858`, `Button.svelte:91–95`,
   `PairingModal.svelte`'s sheet). One elevation token per surface
   (§3.2 of `condura.css` declares this. It is not followed.). The
   fix is mechanical: remove the layered `box-shadow` and rely on the
   token.
10. **No animation that doesn't carry meaning.** The `voidHold` keyframe
    in §1.1 didn't carry meaning — it filled 800ms between arrival and
    awakening. The Skills `rotateX(2deg)` didn't carry meaning — it
    showed off. Every animation in the codebase must answer: "What is
    this communicating?" The thread communicates *completion*. The
    breathe communicates *alive-ness*. The pollen-float communicates
    *rest*. A spinning loader communicates nothing because there is
    nothing to say.

---

## 5. THE $50M FEEL — 5 concrete moves that signal "we care"

The user shouldn't be able to point at one thing and say "this is
nice." They should feel it across ten things, none of which is
flagged as the special one.

### 5.1 The cursor on hoverable surfaces changes to a pollen ring
**File:** `condura.css:347–349`. A 24×24 SVG drawn on top of the page
cursor with `body[data-hover='1']`, a single `--pollen` dot inside a
synapse stroke. The plumbing exists. There is no `Cursor.svelte`
component actually setting `data-hover='1'` on any element.
**Fix:** A single `hover-region` directive / action, applied via
`use:hoverRegion` on `button:not([disabled])`, `[role=button]`,
selectable rows, links. When the cursor is over a hoverable surface,
the component sets `document.body.dataset.hover = '1'`. The ring is
visible evidence that the surface "catches" your mouse. Premiere
products — Stripe Checkout, Linear's command-K — have something
analogous. Condura now has it.

### 5.2 The composer focus state draws a thread inward
**Already exists at `Chat.svelte:515–528`.** The `::before` line under
the composer card scales from `scaleX(0)` to `scaleX(1)` over
`--dur-slow`, in `--synapse`. This is one of the great micro-moments
in the file. **Move the same gesture to every text input** in the
app — the email field on the Account ritual step, the keycap input on
hotkey capture, the search field on Settings, the query input on
Hub/Skills. The same line, the same easing, the same color. The user
learns one gesture for "I am writing here."

### 5.3 The Stop button during streaming breaks the model, not the page
**Today:** `Chat.svelte:277` swaps the Send button for a "Stop" button
when `conversation.isStreaming`, calling `conversation.cancel()`.
The cursor remains `default`. The user can't tell whether the click
landed. **Fix:** On Stop click, replace the chip with a "Stopping…"
label for 280ms, then with "Stopped" + a small pollen pulse that
dissolves the streaming turn. The thread above the turn draws out
right→left (reverse). The user sees the agent *yielding*, not just a
button becoming unfocused. This is one of the cheapest animations to
ship and one of the most premium things to watch.

### 5.4 The Settings nav rail updates its own active state instantly,
         without a route enter
**Today:** Setting nav within Settings is currently re-rendered on
route entry (Shell applies `route-enter` to the whole route container).
The user's eye has to track the route map after every click.
**Fix:** The active nav item gets a 1px `--synapse` line on its left
edge that draws down from top (`scaleY(0→1)` over `--dur`) when the
route changes — without re-entering the whole container. The route's
content does the existing `blur-in` enter. The nav has *its own*
animation, decoupled. Linear does this. Arc does this. Condura can.

### 5.5 The Ritual's "you are here" must read on a phone screenshot
**Today:** The Ritual is two-column at 1080px (constellation left,
content right, `left: 240px; right: 56px;` in the `.content` style,
`Ritual.svelte:792–802`). Below 768px the constellation takes the same
240px and the content squeezes into a 320px column. A user with a
laptop who screenshots the device setup step will see all of it; a
user with a phone will see a sliver.
**Fix:** Below 768px, hide the constellation entirely (not just shrink
it), widen content to full-width with comfortable padding, and ensure
every interactive element has a 44×44 touch target. The Ritual is a
one-time read; it deserves to work on every device. Mobile responsiveness
is a `$50M move` because the user assumes "of course it works on my
phone" — and right now, it doesn't.

---

## Closing note for the next agent

Premium quality is not an aesthetic. It is a series of 50 micro-decisions
that don't contradict each other. None of the moves above will be noticed
in isolation. Together they are the difference between "an AI app" and
"the thing the user keeps coming back to." Ship them one at a time,
never in batches, and always with the test: *would removing this make
the product worse, or just less decorative?* If removing it would make
it worse, it's earned. Otherwise, delete it.
