<script lang="ts">
  import Pulse from './Pulse.svelte';
  import Button from './Button.svelte';
  import Glyph from './Glyph.svelte';

  // The "agent went insane" / halted state. Calm but firm. The Pulse goes red
  // and fast. Motion here is unreducible (16ms) — safety beats aesthetics.
  let {
    reason = 'user requested',
    onresume,
  }: { reason?: string; onresume?: () => void } = $props();
</script>

<div class="kill-overlay">
  <div class="kill-card">
    <Pulse phase="error" size={12} />
    <div class="kill-eyebrow">— Halted · kill switch engaged</div>
    <h2 class="kill-title">Condura has stopped.</h2>
    <p class="kill-body">
      Every active stream was canceled. The agent is not running. Reason:
      <span class="kill-reason">{reason}</span>.
    </p>
    <p class="kill-note">
      Resuming mints a ticket you confirm from the CLI — the GUI never
      auto-restarts a halted agent. Auto-recovery is the enemy.
    </p>
    <div class="kill-foot">
      <Button variant="secondary" onclick={onresume}>Mint resume ticket</Button>
    </div>
  </div>
</div>

<style>
  .kill-overlay {
    position: fixed;
    inset: 0;
    z-index: var(--z-modal);
    background: color-mix(in oklab, var(--surface-ink) 86%, transparent);
    backdrop-filter: blur(8px);
    display: grid;
    place-items: center;
    animation: blur-in var(--dur) var(--ease);
  }
  .kill-card {
    max-width: 460px;
    padding: var(--space-8);
    color: var(--paper);
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .kill-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: color-mix(in oklab, var(--paper) 50%, transparent);
    margin-top: var(--space-3);
  }
  .kill-title {
    font-family: var(--font-display);
    font-size: 40px;
    line-height: 1.05;
    letter-spacing: -0.03em;
    color: var(--paper);
  }
  .kill-body {
    font-size: 16px;
    line-height: 1.55;
    color: color-mix(in oklab, var(--paper) 66%, transparent);
  }
  .kill-reason {
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--pollen);
  }
  .kill-note {
    font-size: 13px;
    color: color-mix(in oklab, var(--paper) 40%, transparent);
    font-style: italic;
  }
  .kill-foot {
    margin-top: var(--space-4);
  }
  :global(.kill-foot .btn-secondary) {
    background: color-mix(in oklab, var(--paper) 8%, transparent);
    color: var(--paper);
    border-color: color-mix(in oklab, var(--paper) 20%, transparent);
  }
</style>