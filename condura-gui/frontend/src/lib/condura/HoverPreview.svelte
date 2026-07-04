<script lang="ts">
  /**
   * Condura · HoverPreview — the 60px-tall strip below the constellation ring.
   * ──────────────────────────────────────────────────────────────────────
   * Per DIRECTION.md §2.2 + SCREEN_RITUAL.md §1.8: fades in over 220ms with
   * an 80ms delay when the user hovers a node. The copy is per-node; for the
   * Summon node the preview includes a one-line keycap row.
   *
   * Renders nothing when no node is hovered (the parent passes `null`).
   */
  import type { NodeId } from './ConstellationNode.svelte';

  let { active }: { active: NodeId | null } = $props();

  // Per-node preview copy. The single load-bearing phrase per surface — written
  // here, not re-decorated with italic-green (MOAT §1.7).
  const COPY: Record<NodeId, string> = {
    perceive: 'Accessibility · Screen Recording. Condura reads only what it must to act.',
    power: 'Local Ollama · an API key · a subscription. Your model, your call.',
    summon: 'Press your combo to call Condura.',
    voice: '"hey condura" — say the wake word; it listens.',
    threads: 'Telegram is ready. The rest when you wire them.',
    account: 'Optional. Sign in for skills sync, donations, support.',
  };

  let visible = $derived(active !== null);
</script>

<div class="preview" class:visible aria-live="polite">
  <div class="preview-inner">
    {#if visible && active}
      <span class="copy">{COPY[active]}</span>
      {#if active === 'summon'}
        <span class="keycap-row" aria-hidden="true">
          <span class="kc">⌘</span><span class="kc">⇧</span><span class="kc">Space</span>
        </span>
      {/if}
    {/if}
  </div>
</div>

<style>
  .preview {
    width: 100%;
    max-width: 560px;
    height: 60px;
    margin: 0 auto;
    opacity: 0;
    transform: translateY(4px);
    transition:
      opacity var(--dur) var(--ease),
      transform var(--dur) var(--ease);
    transition-delay: 80ms;
    pointer-events: none;
  }
  .preview.visible {
    opacity: 1;
    transform: translateY(0);
  }
  .preview-inner {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    padding: 0 var(--space-5);
    border-top: 1px solid var(--hair);
    border-bottom: 1px solid var(--hair);
    background: color-mix(in oklab, var(--surface-card) 60%, transparent);
  }
  .copy {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 14px;
    line-height: 1.4;
    color: var(--content-soft);
    text-align: center;
    max-width: 52ch;
  }
  .keycap-row {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    flex: none;
  }
  .kc {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 3px 8px;
    border: 1px solid var(--hair-strong);
    border-bottom-width: 2px;
    border-radius: var(--r-sm);
    background: var(--surface-card);
    color: var(--content);
    box-shadow: var(--shadow-paper);
  }

  @media (prefers-reduced-motion: reduce) {
    .preview {
      transition: none;
      transform: none;
    }
  }
</style>