<script lang="ts">
  import { onMount } from 'svelte';
  import { onboarding } from '../stores/onboarding.svelte';
  import { FALLBACK_EULA_TEXT, FALLBACK_EULA_VERSION } from './fallbackEula';

  /**
   * Condura · Gate (Screen 1 of the first-run flow)
   * ──────────────────────────────────────────────────────────────────────
   * Legal-first first-run screen. Per SCREEN_RITUAL.md §1.2 + DIRECTION.md §2.2,
   * this IS the arrival signature — no cinematic precedes it. The wax seal stamp
   * is the only CTA; once stamped, the ritual dissolves to the Constellation.
   *
   *  · Scrollable EULA well with 2px synapse progress bar on left edge
   *  · Checkbox + "I have read and accept…" label
   *  · Wax seal — 64×64 radial gradient, `sealBloom` keyframe on stamp
   *  · Bottom-left `not now · quit` skip-note (the only escape from legal)
   *
   * Props:
   *   onAccept()  — fired after `sealBloom` finishes (650ms after stamp click)
   */
  let { onAccept }: { onAccept: () => void } = $props();

  // ── EULA state ────────────────────────────────────────────────────────
  let eulaText = $state('');
  let eulaVersion = $state(FALLBACK_EULA_VERSION);
  let eulaScrolled = $state(false);
  let eulaAccepted = $state(false);
  let eulaEl = $state<HTMLDivElement | undefined>(undefined);
  let eulaReadPct = $state(0);
  let stamped = $state(false);

  function effectiveEula(): { text: string; version: string } {
    const live = onboarding.eula;
    if (live && typeof live.text === 'string' && live.text.trim().length > 200) {
      return { text: live.text, version: live.version || eulaVersion };
    }
    return { text: FALLBACK_EULA_TEXT, version: eulaVersion };
  }

  // Whether the daemon returned a live EULA. When false, we show the offline
  // fallback note so the user knows their acceptance will be replayed later.
  let eulaIsFallback = $derived(
    !(onboarding.eula && typeof onboarding.eula.text === 'string' && onboarding.eula.text.trim().length > 200)
  );

  function recomputeEulaScroll(): void {
    if (!eulaEl) return;
    const max = eulaEl.scrollHeight - eulaEl.clientHeight;
    const ratio = max > 0 ? eulaEl.scrollTop / max : 1;
    eulaReadPct = Math.min(100, Math.max(0, ratio * 100));
    // Self-resolve: short text that fits without scrolling, or the user has
    // reached the bottom (within 8px), arms the stamp.
    if (max <= 4 || max - eulaEl.scrollTop <= 8) {
      if (ratio > 0.4) eulaScrolled = true;
    }
    // Long text: only unlock after the reader has actually scrolled past 85%.
    if (max > 80 && ratio >= 0.85) eulaScrolled = true;
  }

  // After text renders, self-resolve (the text might fit on one page — we
  // don't make the user scroll a one-screen license).
  $effect(() => {
    void eulaText;
    requestAnimationFrame(() => recomputeEulaScroll());
  });

  // Hard-unlock the seal as soon as text is present + non-empty + accepted.
  // Reading the checkbox is the consent; the scroll-thread remains as a
  // visual signal of reading progress.
  let canStamp = $derived(
    !!eulaText && eulaText.trim().length > 100 && eulaAccepted && !onboarding.busy
  );

  async function loadEula(): Promise<void> {
    try {
      await onboarding.loadEula();
    } catch {
      // ignore — fallback below
    }
    const eff = effectiveEula();
    eulaText = eff.text;
    eulaVersion = eff.version;
    requestAnimationFrame(() => recomputeEulaScroll());
  }

  function stampSeal(): void {
    if (!canStamp) return;
    stamped = true;
    void onboarding.acceptEula(eulaVersion).catch((e) => console.error(e));
    // Wait for sealBloom (600ms) + a 650ms linger on the "Accepted" status
    // before dissolving to the Constellation.
    setTimeout(() => onAccept(), 650);
  }

  function quitApp(): void {
    try {
      (window as unknown as { close?: () => void }).close?.();
    } catch {
      /* ignore */
    }
  }

  onMount(() => {
    void loadEula();
  });
</script>

<div class="gate surface-paper">
  <div class="pw-bloom"></div>
  <div class="pw-grain"></div>

  <!-- bottom-left escape — only route off Gate -->
  <button class="skip-note" onclick={quitApp}>
    <span>not now · quit</span><span class="arr">→</span>
  </button>

  <div class="content">
    <div class="eyebrow">— The terms</div>
    <h1 class="headline">First, the terms.</h1>
    <p class="sub">
      Free for personal and commercial use, no tracking, no lock-in. Read what that
      means — scroll to the bottom, then stamp the seal.
    </p>

    <div class="eula" bind:this={eulaEl} onscroll={recomputeEulaScroll}>
      <div class="eula-read" style:height="{eulaReadPct}%"></div>
      <pre class="eula-text">{eulaText || 'Loading the license…'}</pre>
    </div>

    <label class="check">
      <input
        type="checkbox"
        bind:checked={eulaAccepted}
        disabled={!eulaText}
      />
      <span>I have read and accept the Condura Freeware EULA</span>
    </label>

    {#if eulaIsFallback}
      <p class="eula-offline-note">
        Read offline (daemon unreachable) — your acceptance will be replayed to
        Condura on next boot.
      </p>
    {/if}

    <div class="seal-row">
      <button
        class="seal"
        class:stamped
        disabled={!canStamp}
        onclick={stampSeal}
        aria-label="Stamp to accept"
      >
        <span class="seal-c">C</span>
      </button>
      <div class="seal-text">
        <span class="st-1">
          {#if stamped}
            Accepted · thank you
          {:else if !eulaText}
            Loading the license…
          {:else if !eulaScrolled}
            Scroll to the bottom first
          {:else}
            Stamp to accept
          {/if}
        </span>
        <span class="st-2">a considered act — not a click</span>
      </div>
    </div>
  </div>
</div>

<style>
  .gate {
    position: fixed;
    inset: 0;
    z-index: var(--z-modal);
    overflow: hidden;
    display: grid;
    place-items: center;
  }
  .pw-bloom {
    position: absolute;
    inset: 0;
    z-index: 0;
    background:
      radial-gradient(ellipse at 20% 0%, var(--bloom-1), transparent 50%),
      radial-gradient(ellipse at 92% 8%, var(--bloom-2), transparent 45%),
      radial-gradient(ellipse at 50% 105%, var(--bloom-3), transparent 55%);
    pointer-events: none;
  }
  .pw-grain {
    position: absolute;
    inset: 0;
    z-index: 1;
    pointer-events: none;
    opacity: var(--grain-opacity);
    background-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240'><filter id='n'><feTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='2' stitchTiles='stitch'/><feColorMatrix values='0 0 0 0 0.08  0 0 0 0 0.07  0 0 0 0 0.04  0 0 0 0.06 0'/></filter><rect width='100%25' height='100%25' filter='url(%23n)'/></svg>");
    background-size: 240px 240px;
    mix-blend-mode: multiply;
  }
  :global([data-mode='dark']) .pw-grain {
    mix-blend-mode: screen;
  }

  .skip-note {
    position: absolute;
    left: var(--space-5);
    bottom: var(--space-5);
    z-index: 5;
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    color: var(--content-faint);
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    border-radius: var(--r-sm);
    background: none;
    border: none;
    cursor: pointer;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .skip-note:hover {
    color: var(--pollen);
    background: color-mix(in oklab, var(--pollen) 8%, transparent);
    transform: translateX(2px);
  }
  .skip-note:active {
    transform: scale(0.97);
  }
  .skip-note:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .skip-note .arr {
    color: var(--pollen);
  }

  .content {
    position: relative;
    z-index: 3;
    width: 100%;
    max-width: 560px;
    padding: 0 var(--space-5);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    /* Vertically biased to ~38% of viewport — leaves room above for the
       wax-seal halo and room below for the bottom-left skip-note. */
    margin-top: -8vh;
  }

  .eyebrow {
    font-family: var(--font-mono);
    font-size: var(--text-caption);
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-faint);
    animation: fade-in-up var(--dur) var(--ease) 140ms both;
  }
  .headline {
    font-family: var(--font-display);
    font-weight: 400;
    font-size: var(--text-h1);
    line-height: var(--lh-h1);
    letter-spacing: var(--ls-h1);
    color: var(--content);
    animation: fade-in-up var(--dur) var(--ease) 260ms both;
  }
  .sub {
    font-size: var(--text-lead);
    line-height: var(--lh-lead);
    color: var(--content-soft);
    max-width: 52ch;
    animation: fade-in-up var(--dur-slow) var(--ease) 420ms both;
  }

  .eula {
    position: relative;
    max-height: 280px;
    overflow-y: auto;
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
    padding: var(--space-4) var(--space-5);
    animation: fade-in-up var(--dur-slow) var(--ease) 580ms both;
  }
  .eula-text {
    margin: 0;
    font-family: var(--font-display);
    font-size: 14px;
    line-height: 1.7;
    color: var(--content-soft);
    white-space: pre-wrap;
  }
  .eula-read {
    position: absolute;
    left: 0;
    top: 0;
    width: 2px;
    background: var(--synapse);
    transition: height 100ms linear;
  }

  .check {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: 14px;
    color: var(--content);
    cursor: pointer;
    animation: fade-in-up var(--dur) var(--ease) 760ms both;
  }
  .check input {
    width: 16px;
    height: 16px;
    accent-color: var(--pollen);
    cursor: pointer;
  }
  .check input[disabled] {
    opacity: 0.4;
  }

  .eula-offline-note {
    margin: 2px 0 0 var(--space-5);
    font-family: var(--font-display);
    font-style: italic;
    font-size: 13px;
    line-height: 1.5;
    color: var(--content-faint);
    animation: fade-in-up var(--dur) var(--ease) 820ms both;
  }

  .seal-row {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    animation: fade-in-up var(--dur-cine) var(--ease-pop) 940ms both;
  }
  .seal {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: radial-gradient(circle at 35% 30%, var(--synapse-glow), var(--synapse-deep) 70%);
    color: var(--paper);
    display: grid;
    place-items: center;
    cursor: pointer;
    flex: none;
    box-shadow:
      0 8px 20px -8px color-mix(in oklab, var(--synapse) 50%, transparent),
      inset 0 0 0 2px color-mix(in oklab, var(--paper) 20%, transparent);
    transition:
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease),
      opacity var(--dur) var(--ease);
  }
  .seal:hover:not([disabled]) {
    box-shadow:
      0 8px 20px -8px color-mix(in oklab, var(--synapse) 50%, transparent),
      inset 0 0 0 2px color-mix(in oklab, var(--paper) 20%, transparent),
      var(--pollen-halo);
  }
  .seal[disabled] {
    opacity: 0.35;
    cursor: not-allowed;
  }
  .seal.stamped {
    transform: scale(0.94) translateY(2px);
    animation: sealBloom 600ms var(--ease);
  }
  .seal-c {
    font-family: var(--font-display);
    font-size: 28px;
    line-height: 1;
    color: var(--paper);
  }
  :global([data-mode='dark']) .seal-c {
    color: var(--ink);
  }

  .seal-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .st-1 {
    font-size: 15px;
    color: var(--content);
  }
  .st-2 {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-faint);
  }

  @keyframes sealBloom {
    0% {
      box-shadow:
        0 8px 20px -8px rgba(11, 61, 46, 0.5),
        inset 0 0 0 2px rgba(234, 246, 239, 0.2),
        0 0 0 0 rgba(201, 123, 46, 0.5);
    }
    100% {
      box-shadow:
        0 8px 20px -8px rgba(11, 61, 46, 0.5),
        inset 0 0 0 2px rgba(234, 246, 239, 0.2),
        0 0 0 28px rgba(201, 123, 46, 0);
    }
  }
  @keyframes fade-in-up {
    from { opacity: 0; transform: translateY(8px); }
    to   { opacity: 1; transform: translateY(0); }
  }

  @media (prefers-reduced-motion: reduce) {
    .seal.stamped {
      animation: none;
      transform: scale(0.94) translateY(2px);
    }
    .eyebrow,
    .headline,
    .sub,
    .eula,
    .check,
    .eula-offline-note,
    .seal-row {
      animation: none;
      opacity: 1;
      transform: none;
    }
  }
</style>