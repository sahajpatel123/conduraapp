<script lang="ts">
  /**
   * ErrorState — the shared error surface for every Condura route.
   *
   * Per MOAT §2.6: one component owns all error rendering across Chat,
   * Channels, Skills, Hub, and Audit. Three lines, exactly:
   *   1. Headline (italic display 22) — "We couldn't <verb> the <noun>."
   *   2. Cause (mono 11) — one noun, what failed.
   *   3. Reason (mono 11) — one phrase, likely cause.
   *   + Try again pill (mono-pollen outline), optional Open Settings link,
   *     err-hair that draws in over 600ms.
   *
   * The headline font is italic display (the same headline font as
   * Chat / Skills / Channels) so all four routes read as one product.
   */
  import type { Snippet } from 'svelte';
  import Pulse from './Pulse.svelte';
  import Button from './Button.svelte';

  let {
    head,
    cause,
    reason,
    onretry,
    retryLabel = 'Try again',
    onsettings,
    settingsLabel = 'Open Settings',
    children,
  }: {
    head: string;
    cause: string;
    reason: string;
    onretry?: () => void;
    retryLabel?: string;
    onsettings?: () => void;
    settingsLabel?: string;
    children?: Snippet;
  } = $props();
</script>

<div class="error-state" role="alert" aria-live="polite">
  <div class="row">
    <Pulse phase="error" size={8} />
    <span class="head">{head}</span>
  </div>
  <p class="cause">
    <span class="k">Cause:</span>
    <span class="v">{cause}</span>
  </p>
  <p class="reason">
    <span class="k">Likely reason:</span>
    <span class="v">{reason}</span>
  </p>

  {#if children}
    {@render children()}
  {/if}

  <div class="actions">
    {#if onsettings}
      <button class="link" onclick={onsettings}>{settingsLabel} →</button>
    {/if}
    {#if onretry}
      <Button variant="secondary" size="sm" onclick={onretry}>{retryLabel}</Button>
    {/if}
  </div>

  <div class="err-hair" aria-hidden="true"></div>
</div>

<style>
  .error-state {
    max-width: 560px;
    padding: var(--space-4) 0 var(--space-3);
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .row {
    display: inline-flex;
    align-items: center;
    gap: 10px;
  }
  .head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 22px;
    line-height: 1.15;
    color: var(--content);
    letter-spacing: -0.01em;
  }
  .cause,
  .reason {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--content-soft);
    line-height: 1.55;
    margin: 0;
    display: flex;
    gap: var(--space-2);
  }
  .cause .k,
  .reason .k {
    color: var(--content-faint);
    text-transform: uppercase;
    letter-spacing: 0.14em;
    flex: none;
    min-width: 96px;
  }
  .cause .v,
  .reason .v {
    color: var(--content);
    font-family: var(--font-mono);
  }
  .actions {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-top: var(--space-2);
  }
  .link {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--content-mute);
    background: transparent;
    border: 0;
    padding: 0;
    cursor: pointer;
    transition: color var(--dur) var(--ease);
  }
  .link:hover {
    color: var(--synapse);
  }
  .link:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
    border-radius: var(--r-xs);
  }
  .err-hair {
    height: 1px;
    width: 100%;
    background: linear-gradient(90deg, var(--hair-strong) 0%, var(--hair-strong) 60%, transparent 100%);
    transform: scaleX(0);
    transform-origin: left;
    animation: err-hair-draw 600ms var(--ease) 120ms forwards;
    margin-top: var(--space-3);
  }
  @keyframes err-hair-draw {
    to {
      transform: scaleX(1);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .err-hair {
      transform: scaleX(1);
      animation: none;
    }
  }
</style>