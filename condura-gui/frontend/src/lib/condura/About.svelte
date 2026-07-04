<script lang="ts">
  // Condura · About — the colophon. A flowing document, not a settings tab,
  // not a modal. The seven non-negotiable invariants from CLAUDE.md §2.1
  // surfaced as the visible promise of the agent. The spec is the contract:
  // changes to invariant copy or the hero title require a CLAUDE.md
  // amendment (SCREEN_ABOUT.md §1.3, §7 D2, "Provenance").

  import { onMount, onDestroy } from 'svelte';
  import Pulse from './Pulse.svelte';
  import Thread from './Thread.svelte';
  import { ROUTE_HASH } from './NavRail.svelte';
  import { ipc } from '../ipc/client';

  // ── State ───────────────────────────────────────────────────────────
  // hover     : which row is under the cursor (drives the armor rect)
  // visible   : rows whose hairline Thread has been drawn in (IO-driven)
  // colophonIn: footer breath-thread draws after first rAF
  // drawn     : rows whose entry stagger slot has elapsed (mount choreography)
  // active    : the row whose center is closest to viewport center
  let hover = $state<string | null>(null);
  let visible = $state<Set<string>>(new Set());
  let drawn = $state<Set<string>>(new Set());
  let colophonIn = $state(false);
  let activeRow = $state<string | null>(null);
  let wordmarkIn = $state(true); // start true; SSR/no-JS case renders static
  let reduceMotion = $state(false);
  let buildInfo = $state<{ version: string; commit: string } | null>(null);

  // ── The seven invariants (CLAUDE.md §2.1, verbatim) ────────────────
  // Each row also carries a monospace citation naming the file path in
  // the daemon that enforces the invariant. Spec target citations
  // (SCREEN_ABOUT.md §1.4). The spec guarantees these.
  const invariants = [
    {
      n: '01',
      title: 'The Strategist and the Gatekeeper are separate.',
      body: 'The Strategist is any model. The Gatekeeper is deterministic code. They are never the same system.',
      citation: 'internal/gatekeeper/policy.go:42',
    },
    {
      n: '02',
      title: 'The Gatekeeper is the only path to physical action.',
      body: 'No model output flows to a click, type, or shell exec without passing the Gatekeeper.',
      citation: 'internal/gatekeeper/gate.go:17',
    },
    {
      n: '03',
      title: 'Destructive actions require a real human at the keyboard.',
      body: 'A native modal that halts execution until the human physically allows. No exceptions.',
      citation: 'internal/gatekeeper/consent.go:88',
    },
    {
      n: '04',
      title: 'The user can always stop the agent.',
      body: 'A hard hotkey, a watchdog timer, network isolation, a menu-bar kill. Four independent mechanisms.',
      citation: 'internal/halt/halt.go:31',
    },
    {
      n: '05',
      title: 'Every action is auditable.',
      body: 'HMAC-chained, append-only, tamper-resistant. If something goes wrong, we can prove exactly what happened.',
      citation: 'internal/audit/chain.go:55',
    },
    {
      n: '06',
      title: 'The agent is a guest, not an owner.',
      body: 'It requests permission to enter rooms. The user grants or denies. We never escalate, never bypass.',
      citation: 'internal/safety/permission.go:14',
    },
    {
      n: '07',
      title: 'OS permissions are granted by the user, on their machine.',
      body: "We don't have access. We ask, they grant. The onboarding makes this easy and clear.",
      citation: 'internal/permissions/permissions.go:23',
    },
  ];

  // ── Constants ───────────────────────────────────────────────────────
  const STAGGER_MS = 60; // SCREEN_ABOUT.md §3.1 / §3.4 — stagger; collapses to 0 under reduced-motion
  const DONATE_URL = 'https://synaptic.app/donate';

  // ── Element refs ────────────────────────────────────────────────────
  // Per-row HTMLElement registry, keyed by invariant id. `bind:this`
  // can't take a method call, so we use a Svelte action that hands the
  // mounted/unmounted node back into the map via the `node` parameter.
  const rowEls = new Map<string, HTMLElement>();
  function trackRow(id: string) {
    return (node: HTMLElement) => {
      rowEls.set(id, node);
      return {
        destroy() {
          rowEls.delete(id);
        },
      };
    };
  }
  let io: IntersectionObserver | null = null;
  let scrollRaf = 0;

  // ── Mount: IO + stagger + colophon draw + version fetch ────────────
  onMount(() => {
    // Reduced-motion respect (SPEC §3.4). Per MOAT §2.3 the global
    // condura.css block is the source of truth; we read it here so the
    // JS-driven stagger and wordmark draw can collapse to 0.
    reduceMotion = matchMedia('(prefers-reduced-motion: reduce)').matches;

    // Wordmark draw: starts the SVG stroke-dashoffset → 0 transition.
    // Always begins at 0 — it's a one-shot CSS animation; reduced-motion
    // users see the static wordmark instead (the CSS @keyframes completes
    // immediately).
    if (reduceMotion) {
      wordmarkIn = true;
    } else {
      // delay one frame so the browser commits the offset=1 state, then
      // the transition has something to animate from.
      requestAnimationFrame(() => (wordmarkIn = true));
    }

    // IntersectionObserver draws each row's hairline Thread once the row
    // crosses 0.35 visibility (SCREEN_ABOUT.md §3.1).
    if (typeof IntersectionObserver !== 'undefined') {
      io = new IntersectionObserver(
        (entries) => {
          for (const entry of entries) {
            if (entry.isIntersecting && entry.target instanceof HTMLElement) {
              const id = entry.target.dataset.line ?? '';
              if (id) visible = new Set([...visible, id]);
            }
          }
        },
        { threshold: 0.35 }
      );
      for (const el of rowEls.values()) io.observe(el);
    } else {
      // No IO support: reveal all rows immediately.
      visible = new Set(invariants.map((r) => r.n));
    }

    // Mount choreography — stagger each row in by 60ms.
    // Reduced-motion collapses the delay to 0 so all rows land together.
    const runDraw = (n: string, ms: number) => {
      const delay = reduceMotion ? 0 : ms;
      setTimeout(() => {
        drawn = new Set([...drawn, n]);
      }, delay);
    };
    invariants.forEach((inv, i) => runDraw(inv.n, i * STAGGER_MS));

    // Scroll-linked active row (SPEC §3.2) — the row whose center is
    // closest to viewport center gets a left-border synapse accent.
    // rAF-throttled so we don't thrash on every scroll tick.
    const onScroll = () => {
      if (scrollRaf) return;
      scrollRaf = requestAnimationFrame(() => {
        scrollRaf = 0;
        computeActiveRow();
      });
    };
    window.addEventListener('scroll', onScroll, { passive: true });
    window.addEventListener('resize', onScroll);
    // initial pass
    computeActiveRow();

    // Footer breath-thread draws in on next frame (SPEC §3.1).
    requestAnimationFrame(() => (colophonIn = true));

    // Keyboard: ⌘D / Ctrl+D → open donate URL (SPEC §4).
    const onKey = (e: KeyboardEvent) => {
      const cmd = e.metaKey || e.ctrlKey;
      if (cmd && (e.key === 'd' || e.key === 'D')) {
        e.preventDefault();
        window.open(DONATE_URL, '_blank', 'noopener,noreferrer');
      }
    };
    window.addEventListener('keydown', onKey);

    // Version manifest (SPEC §6) — gracefully degrade if RPC is down.
    // The page still renders; build hash substring becomes "version unavailable".
    ipc
      .version()
      .then((v) => {
        buildInfo = { version: v.version, commit: v.commit };
      })
      .catch(() => {
        buildInfo = null;
      });

    return () => {
      window.removeEventListener('scroll', onScroll);
      window.removeEventListener('resize', onScroll);
      window.removeEventListener('keydown', onKey);
    };
  });

  onDestroy(() => {
    io?.disconnect();
    if (scrollRaf) cancelAnimationFrame(scrollRaf);
  });

  function computeActiveRow() {
    const vh = window.innerHeight;
    const center = vh / 2;
    let best: { id: string; dist: number } | null = null;
    for (const [id, el] of rowEls) {
      const r = el.getBoundingClientRect();
      const rowCenter = (r.top + r.bottom) / 2;
      const d = Math.abs(rowCenter - center);
      if (best === null || d < best.dist) best = { id, dist: d };
    }
    activeRow = best ? best.id : null;
  }

  // ── Colophon text builders ─────────────────────────────────────────
  function fmtVersion(): string {
    if (!buildInfo) return 'version unavailable';
    const v = buildInfo.version || 'v?';
    const c = buildInfo.commit ? buildInfo.commit.slice(0, 7) : '';
    return c ? `${v} · ${c}` : v;
  }
</script>

<article class="about">
  <!-- HERO ────────────────────────────────────────────────────────── -->
  <header class="head">
    <div class="eyebrow">— The colophon</div>

    <!--
      Wordmark: "Condura" rendered as SVG <text> with stroke + fill.
      The stroke uses pathLength=1 / stroke-dasharray=1 so
      stroke-dashoffset 1→0 over --dur-cine (900ms) draws the outline
      left to right, then the fill is made opaque via a delayed class
      toggle. Reduced-motion users skip the animation entirely.
    -->
    <h1 class="title" aria-label="Condura">
      <span class="wordmark-svg" aria-hidden="true">
        <svg viewBox="0 0 320 64" preserveAspectRatio="xMinYMid meet">
          <text
            class="wordmark-stroke"
            x="0" y="48"
            font-family="'Instrument Serif', Georgia, serif"
            font-size="56"
            font-weight="400"
            fill="var(--content)"
            stroke="var(--content)"
            stroke-width="1"
            pathLength="1"
            stroke-dasharray="1"
            stroke-dashoffset={wordmarkIn ? 0 : 1}
            >Condura</text
          >
        </svg>
      </span>
      <span class="wordmark-fallback" aria-hidden="true">Condura</span>
      <span class="title-alive">{` `}</span>
      <span class="title-mantissa">— made by a human and an AI, in partnership.</span>
    </h1>

    <p class="sub">
      Condura is a free, local-first desktop agent. v0.1.0. No telemetry. No lock-in.
      The seven promises below are non-negotiable — they are the safety we built on,
      and the reason you can trust this thing on your machine.
    </p>
    <div class="cred">
      <Pulse phase="thinking" size={6} />
      <span>
        thinking in public ·
        <a class="cred-link" href={ROUTE_HASH.about}>v0.1.0 · changelog</a>
      </span>
    </div>
  </header>

  <!-- LEDGER ──────────────────────────────────────────────────────── -->
  <section class="ledger" aria-label="The seven non-negotiable invariants">
    <div class="ledger-eyebrow">— The seven invariants</div>

    {#each invariants as inv (inv.n)}
      <div
        class="row"
        class:hover={hover === inv.n}
        class:active={activeRow === inv.n}
        class:drawn={drawn.has(inv.n)}
        use:trackRow={inv.n}
        data-line={inv.n}
        role="group"
        aria-label={`Invariant ${inv.n}: ${inv.title}`}
        onmouseenter={() => (hover = inv.n)}
        onmouseleave={() => (hover = null)}
      >
        <!-- hairline Thread that draws in L→R once the row enters viewport -->
        <svg class="hairline" preserveAspectRatio="none" aria-hidden="true">
          <line
            x1="0" y1="0.5" x2="1" y2="0.5"
            pathLength="1"
            vector-effect="non-scaling-stroke"
            stroke-dasharray="1"
            stroke-dashoffset={visible.has(inv.n) ? 0 : 1}
          />
        </svg>

        <span class="row-n">{inv.n}</span>
        <div class="row-body">
          <div class="row-title">{inv.title}</div>
          <div class="row-text">{inv.body}</div>
          <!-- Monospace citation (SCREEN_ABOUT.md §1.4): hairline-faint at rest, -->
          <!-- reveals on hover via the Thread underline (SPEC §3.3).              -->
          <a
            class="citation"
            href="#/about"
            tabindex="0"
            title={`Enforced in ${inv.citation}`}
            onclick={(e) => e.preventDefault()}
          >
            {inv.citation}
          </a>
        </div>

        <!-- armor rect (the protection gesture) paints in on hover -->
        <svg class="armor" preserveAspectRatio="none" aria-hidden="true">
          <rect
            x="1" y="1" width="98%" height="92%"
            rx="14" ry="14"
            fill="none"
            stroke="var(--synapse-glow)"
            stroke-width="1.5"
            pathLength="1"
            vector-effect="non-scaling-stroke"
            stroke-dasharray="1"
            stroke-dashoffset={hover === inv.n ? 0 : 1}
          />
        </svg>
      </div>
    {/each}
  </section>

  <!-- FOOTER ──────────────────────────────────────────────────────── -->
  <footer class="foot">
    <div class="breath-thread" aria-hidden="true">
      <Thread orientation="h" draw={colophonIn} />
      <div class="breath-pulse"><Pulse phase="idle" size={6} /></div>
    </div>

    <p class="colophon">
      Condura · {fmtVersion()} · free for personal and commercial use.
      <br />
      <a href="#/about">EULA</a>
      <span class="dot">·</span>
      <a href="#/privacy">Privacy</a>
      <span class="dot">·</span>
      <a
        class="donate"
        href={DONATE_URL}
        target="_blank"
        rel="noopener noreferrer"
        aria-keyshortcuts="Meta+D Control+D"
      >
        Support Condura <kbd>⌘D</kbd>
      </a>
    </p>
  </footer>
</article>

<noscript>
  <style>
    /* No-JS fallback (SPEC §2): no IO, no stagger, no animation. */
    .row { opacity: 1 !important; transform: none !important; }
    .hairline line { stroke-dashoffset: 0 !important; }
    .armor rect { stroke-dashoffset: 1 !important; }
    .wordmark-svg { display: none; }
    .wordmark-fallback { display: inline; }
  </style>
</noscript>

<style>
  /* ── Container ─────────────────────────────────────────────────── */
  .about {
    max-width: 760px; /* SPEC §1 — do not regress from spec target ~720 */
    margin: 0 auto;
    padding: var(--space-9) var(--space-5) var(--space-10);
  }

  /* ── HERO ──────────────────────────────────────────────────────── */
  .head { margin-bottom: var(--space-8); }
  .eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .title {
    font-family: var(--font-display);
    font-size: clamp(32px, 4vw, 48px);
    line-height: 1.04;
    letter-spacing: -0.035em;
    color: var(--content);
    margin: var(--space-4) 0 var(--space-3);
    /* upright, NOT italic-synapse (DIRECTION §6 rule "no gradient", SPEC §1.1, D3) */
  }
  /*
   * The wordmark SVG renders one glyph path; the duplicate text node is
   * its no-JS fallback. With JS, the inline SVG <text> element takes the
   * visible slot. Both share the same baseline grid (the SVG viewBox sits
   * flush with the display font size).
   */
  .wordmark-svg {
    display: inline-block;
    width: clamp(180px, 22vw, 260px);
    height: 0.92em; /* visually matches clamp(32,4vw,48) baseline */
    vertical-align: -0.08em;
    line-height: 1;
  }
  .wordmark-svg svg { width: 100%; height: 100%; display: block; }
  .wordmark-stroke {
    transition: stroke-dashoffset var(--dur-cine) var(--ease),
                fill-opacity var(--dur-cine) var(--ease) calc(var(--dur-cine) * 0.5);
    fill-opacity: 1;
  }
  .wordmark-fallback {
    display: none; /* SVG handles it; revealed via <noscript> override */
  }
  .title-mantissa {
    font-family: var(--font-display);
    font-style: italic;
    font-size: clamp(20px, 2.4vw, 28px);
    letter-spacing: -0.02em;
    color: var(--content);
    margin-left: var(--space-3);
    white-space: nowrap;
  }
  /* under 540px the mantissa drops to a new line */
  @media (max-width: 540px) {
    .title-mantissa {
      display: block;
      margin-left: 0;
      margin-top: var(--space-2);
      font-size: 18px;
      white-space: normal;
    }
  }
  .sub {
    font-size: 16px;
    line-height: 1.6;
    color: var(--content-soft);
    max-width: 56ch;
  }
  .cred {
    display: flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-top: var(--space-4);
  }
  .cred-link {
    color: var(--synapse);
    text-transform: none;
    letter-spacing: normal;
    font-family: var(--font-sans);
    font-size: 13px;
    padding: 1px 4px;
    border-radius: var(--r-xs);
    text-decoration: none;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease);
  }
  .cred-link:hover {
    background: color-mix(in oklab, var(--synapse) 8%, transparent);
  }
  .cred-link:active { transform: scale(0.97); }
  .cred-link:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  /* ── LEDGER ────────────────────────────────────────────────────── */
  .ledger {
    margin-top: var(--space-8);
    border-top: 1px solid var(--hair);
    border-bottom: 1px solid var(--hair);
    padding: var(--space-5) 0;
  }
  .ledger-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-bottom: var(--space-4);
  }
  .row {
    position: relative;
    width: 100%;
    text-align: left;
    display: grid;
    grid-template-columns: 56px 1fr;
    column-gap: var(--space-4);
    padding: var(--space-4);
    background: transparent;
    border: 0;
    border-left: 2px solid transparent;
    border-radius: var(--r-md);
    cursor: default;
    transition:
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      padding-left var(--dur) var(--ease),
      opacity var(--dur-slow) var(--ease);
    /* Mount choreography — start invisible; stagger flips `drawn` to fade in */
    opacity: 0;
    transform: translateY(6px);
  }
  .row.drawn {
    opacity: 1;
    transform: translateY(0);
  }
  .row:hover {
    background: var(--surface-card);
    /* per SPEC §7 D5: 1px lift, not 4px. text-led, not card-led. */
    transform: translateY(-1px);
  }
  .row:hover.drawn {
    transform: translateY(-1px);
  }
  .row:active { transform: scale(0.99); }
  /* SPEC §3.2 — scroll-linked active state: synapse left-border accent
     on the row whose center is closest to viewport center. */
  .row.active {
    border-left-color: var(--synapse);
    padding-left: calc(var(--space-4) + 2px);
  }
  .row:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  .row-n {
    font-family: var(--font-display);
    font-size: 28px;
    line-height: 1;
    color: var(--synapse);
    opacity: 0.55;
    transition: opacity var(--dur) var(--ease);
  }
  .row:hover .row-n,
  .row.active .row-n {
    opacity: 1;
  }

  .row-body {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0; /* allow text wrap inside grid */
  }
  .row-title {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 18px;
    line-height: 1.25;
    color: var(--content);
  }
  .row-text {
    font-size: 14px;
    line-height: 1.55;
    color: var(--content-soft);
  }

  /* Monospace citation (SCREEN_ABOUT.md §1.4).
     Idle = hairline-faint, the user sees the promise on first read.
     On hover → reveals via the Thread underline (SPEC §3.3). */
  .citation {
    align-self: flex-start;
    margin-top: var(--space-2);
    font-family: var(--font-mono);
    font-size: 12px;
    line-height: 1;
    letter-spacing: 0.12em;
    color: var(--content-ghost);
    text-decoration: none;
    position: relative;
    padding: 2px 0;
    transition:
      color var(--dur) var(--ease),
      transform var(--dur-fast) var(--ease);
  }
  .citation::after {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: -1px;
    height: 1px;
    background: currentColor;
    transform: scaleX(0);
    transform-origin: left;
    transition: transform var(--dur) var(--ease);
  }
  .row:hover .citation,
  .row:focus-within .citation {
    color: var(--content-faint);
  }
  .citation:hover,
  .citation:focus-visible {
    color: var(--synapse);
    outline: none;
  }
  .citation:hover::after,
  .citation:focus-visible::after,
  .row:hover .citation::after,
  .row:focus-within .citation::after {
    transform: scaleX(1);
  }

  /* the 1-px hairline that draws in L→R when the row scrolls into view (SPEC §3.1) */
  .hairline {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    width: 100%;
    height: 1px;
    pointer-events: none;
  }
  .hairline line {
    fill: none;
    stroke: var(--synapse);
    stroke-width: 1;
    stroke-linecap: round;
    transition: stroke-dashoffset var(--dur-slow) var(--ease);
  }

  /* the protective synapse armor rect that paints on hover */
  .armor {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }
  .armor rect {
    transition: stroke-dashoffset var(--dur-slow) var(--ease);
  }

  /* ── FOOTER ────────────────────────────────────────────────────── */
  .foot {
    margin-top: var(--space-9);
    text-align: center;
    position: relative;
  }
  .breath-thread {
    position: relative;
    height: 24px;
    margin-bottom: var(--space-5);
  }
  .breath-thread :global(.condura-thread) {
    height: 2px;
    width: 100%;
  }
  .breath-pulse {
    position: absolute;
    left: 50%;
    top: 0;
    transform: translateX(-50%);
  }
  .colophon {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    line-height: 1.8;
    color: var(--content-soft);
  }
  .colophon a {
    color: var(--content);
    margin: 0 4px;
    padding: 2px 4px;
    cursor: pointer;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
    text-decoration: none;
    position: relative;
    border-radius: var(--r-xs);
  }
  .colophon a::after {
    content: '';
    position: absolute;
    left: 4px;
    right: 4px;
    bottom: 2px;
    height: 1px;
    background: currentColor;
    transform: scaleX(0);
    transform-origin: left;
    transition: transform var(--dur) var(--ease);
  }
  .colophon a:hover {
    color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 8%, transparent);
  }
  .colophon a:active { transform: scale(0.97); }
  .colophon a:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .colophon a:hover::after,
  .colophon a:focus-visible::after {
    transform: scaleX(1);
  }
  /* SPEC §7 D6 — the donate link earns the only pollen-colored CTA in the doc. */
  .colophon a.donate {
    color: var(--pollen);
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .colophon a.donate:hover {
    color: var(--pollen);
    background: color-mix(in oklab, var(--pollen) 10%, transparent);
  }
  .colophon a.donate kbd {
    font-family: var(--font-mono);
    font-size: 10px;
    line-height: 1;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    background: var(--surface-card);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-xs);
    padding: 2px 5px;
    color: var(--content-soft);
    /* kbd inherits motion from .colophon a; it does not get its own hover */
  }
  .colophon .dot {
    color: var(--content-faint);
    margin: 0 4px;
  }

  /* ── Reduced-motion contract ─────────────────────────────────────
     Owned by condura.css globally, but the row hover tint, the
     wordmark draw, and the stagger delta are surface-specific (SPEC
     §3.4 / D8 / MOAT §2.3). Skipped duration → instant final state. */
  @media (prefers-reduced-motion: reduce) {
    .row {
      /* stagger collapses to 0; all rows land together already visible */
      opacity: 1;
      transform: none;
    }
    .row:hover { background: transparent; transform: none; }
    .wordmark-stroke {
      transition: none;
      stroke-dashoffset: 0;
    }
    .armor rect,
    .hairline line,
    .colophon a::after,
    .citation::after {
      transition: none !important;
    }
    .citation::after { transform: scaleX(1); }
  }
</style>
