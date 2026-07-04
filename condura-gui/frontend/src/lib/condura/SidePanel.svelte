<script lang="ts">
  import type { Snippet } from 'svelte';
  import Thread from './Thread.svelte';

  /**
   * Condura · SidePanel — the .c-sheet slide-in for per-node detail.
   * ──────────────────────────────────────────────────────────────────────
   * Per MOAT §2.8: this is the .c-sheet primitive — slides from the right,
   * doesn't block page scroll, Esc + outside-click + X to close. Width 380px.
   *
   * Per SCREEN_RITUAL.md §3.4:
   *   0ms: side panel slides in (translateX(24px) → 0, opacity 0 → 1) over
   *        --dur-slow (520ms).
   *   80ms: panel content staggers in (60ms per row).
   *   On mount: Thread draws across the bottom edge over --dur-slow.
   *
   * Focus trap: when `open === true`, Tab cycles within the panel.
   * Esc closes. Initial focus lands on the close button.
   */
  let {
    open,
    eyebrow,
    headline,
    onclose,
    children,
  }: {
    open: boolean;
    eyebrow: string;
    headline: string;
    onclose: () => void;
    children?: Snippet;
  } = $props();

  let panelEl = $state<HTMLDivElement | undefined>(undefined);
  let closeBtn = $state<HTMLButtonElement | undefined>(undefined);

  function onkeydown(e: KeyboardEvent): void {
    if (!open) return;
    if (e.key === 'Escape') {
      e.preventDefault();
      onclose();
      return;
    }
    if (e.key !== 'Tab' || !panelEl) return;
    // Focus trap: keep Tab inside the panel.
    const focusable = panelEl.querySelectorAll<HTMLElement>(
      'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
    );
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    const active = document.activeElement as HTMLElement | null;
    if (e.shiftKey && active === first) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && active === last) {
      e.preventDefault();
      first.focus();
    }
  }

  function onbackdropClick(e: MouseEvent): void {
    if (e.target === e.currentTarget) onclose();
  }

  $effect(() => {
    if (open) {
      // Initial focus on the close button after the slide-in begins.
      requestAnimationFrame(() => closeBtn?.focus());
    }
  });
</script>

<svelte:window onkeydown={onkeydown} />

{#if open}
  <!-- The backdrop is a thin scrim; clicks pass through to close. -->
  <div
    class="backdrop"
    onclick={onbackdropClick}
    onkeydown={(e) => {
      if (e.key === 'Escape') onclose();
    }}
    role="presentation"
  >
    <div
      class="panel"
      bind:this={panelEl}
      role="dialog"
      aria-modal="true"
      aria-labelledby="panel-headline"
      tabindex="-1"
    >
      <header class="head">
        <div class="head-text">
          <div class="eyebrow">{eyebrow}</div>
          <h2 id="panel-headline" class="headline">{headline}</h2>
        </div>
        <button
          bind:this={closeBtn}
          class="close"
          onclick={onclose}
          aria-label="Close panel"
          type="button"
        >
          <span aria-hidden="true">×</span>
        </button>
      </header>

      <div class="body">
        {@render children?.()}
      </div>

      <!-- Signature thread-draw across the panel's bottom edge. -->
      <div class="thread-edge">
        <Thread orientation="h" draw={true} glow={false} />
      </div>
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: var(--z-modal);
    display: flex;
    justify-content: flex-end;
    background: color-mix(in oklab, var(--ink) 6%, transparent);
    animation: fade-in var(--dur) var(--ease) both;
  }
  .panel {
    position: relative;
    width: min(380px, 92vw);
    height: 100%;
    background: var(--surface-raised);
    border-left: 1px solid var(--hair);
    box-shadow: var(--shadow-float);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    animation: slide-in var(--dur-slow) var(--ease) both;
  }
  @keyframes slide-in {
    from { transform: translateX(24px); opacity: 0; }
    to   { transform: translateX(0);    opacity: 1; }
  }
  @keyframes fade-in {
    from { opacity: 0; }
    to   { opacity: 1; }
  }

  .head {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    padding: var(--space-5) var(--space-5) var(--space-4);
    border-bottom: 1px solid var(--hair);
  }
  .head-text {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }
  .eyebrow {
    font-family: var(--font-mono);
    font-size: var(--text-caption);
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .headline {
    font-family: var(--font-display);
    font-weight: 400;
    font-size: var(--text-h2);
    line-height: var(--lh-h2);
    letter-spacing: var(--ls-h2);
    color: var(--content);
    margin: 0;
  }
  .close {
    flex: none;
    width: 28px;
    height: 28px;
    display: grid;
    place-items: center;
    background: none;
    border: 1px solid var(--hair);
    border-radius: var(--r-pill);
    color: var(--content-mute);
    font-family: var(--font-display);
    font-size: 18px;
    line-height: 1;
    cursor: pointer;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      transform var(--dur) var(--ease);
  }
  .close:hover {
    color: var(--content);
    background: var(--surface-card);
    border-color: var(--hair-strong);
  }
  .close:active {
    transform: scale(0.97);
  }
  .close:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 2px var(--synapse),
      0 0 0 5px var(--pollen-halo);
  }

  .body {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-5);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .thread-edge {
    position: relative;
    height: 2px;
    flex: none;
  }

  @media (prefers-reduced-motion: reduce) {
    .panel {
      animation: none;
    }
    .backdrop {
      animation: none;
    }
  }
</style>