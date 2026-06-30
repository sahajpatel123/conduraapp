<!--
  ConsentModal — the ONLY modal in Synaptic.

  Per spec §10.2 (safety): for any WRITE / NETWORK / DESTRUCTIVE action,
  this modal appears. Native-feeling (system dialog colors and sizing).
  Cascades text in with deliberate-stagger so the user is forced to read.

  Per motion agent §3.6: 420ms emphasized timing. Stagger 180ms per item.
  Approve button disabled for 1.2s during the cascade.

  Props:
    verb      — the action verb: "Send", "Delete", "Open", etc.
    target    — the target: email address, file path, URL
    details   — additional context to show the user
    onapprove — handler for "Allow"
    ondeny    — handler for "Don't allow"
-->
<script lang="ts">
  import Button from './Button.svelte';

  interface Props {
    verb: string;
    target: string;
    details?: string;
    onapprove?: () => void;
    ondeny?: () => void;
  }

  let { verb, target, details, onapprove, ondeny }: Props = $props();

  let approveEnabled = $state(false);

  // Per motion agent §3.6: 1.2s forced dwell before approve becomes enabled.
  $effect(() => {
    const id = setTimeout(() => {
      approveEnabled = true;
    }, 1200);
    return () => clearTimeout(id);
  });
</script>

<div class="modal-host" role="presentation">
  <div class="scrim" aria-hidden="true"></div>

  <div
    class="modal"
    role="alertdialog"
    aria-modal="true"
    aria-labelledby="consent-title"
    aria-describedby="consent-body"
  >
    <header class="modal__head">
      <h2 id="consent-title" class="modal__title">
        Synaptic wants to {verb.toLowerCase()}.
      </h2>
    </header>

    <div id="consent-body" class="modal__body">
      <div class="modal__row" style="animation-delay: 0ms">
        <span class="modal__label">Action</span>
        <span class="modal__value">{verb}</span>
      </div>
      <div class="modal__row" style="animation-delay: 180ms">
        <span class="modal__label">Target</span>
        <span class="modal__value modal__value--target">{target}</span>
      </div>
      {#if details}
        <div class="modal__row" style="animation-delay: 360ms">
          <span class="modal__label">Details</span>
          <span class="modal__value modal__value--details">{details}</span>
        </div>
      {/if}
      <div class="modal__row modal__row--meta" style="animation-delay: 540ms">
        <span class="modal__label">You can revoke this any time</span>
        <span class="modal__value">I'll stop the moment you do.</span>
      </div>
    </div>

    <footer class="modal__foot">
      <Button variant="secondary" size="md" onclick={ondeny}>
        Don't allow
      </Button>
      <Button
        variant="primary"
        size="md"
        disabled={!approveEnabled}
        onclick={onapprove}
      >
        Allow this once
      </Button>
    </footer>

    <p class="modal__foot-hint">
      For ongoing access, choose "Always allow" in Settings → Permissions.
    </p>
  </div>
</div>

<style>
  .modal-host {
    position: fixed;
    inset: 0;
    z-index: var(--z-modal);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .scrim {
    position: absolute;
    inset: 0;
    background-color: var(--surface-scrim);
    backdrop-filter: blur(4px);
    animation: scrim-in var(--duration-emphasized) var(--ease-accelerate) both;
  }

  @keyframes scrim-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .modal {
    position: relative;
    width: 480px;
    max-width: calc(100vw - 32px);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-4);
    padding: var(--space-6);
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    animation: modal-in var(--duration-emphasized) var(--ease-emphasized) both;
  }

  @keyframes modal-in {
    from {
      opacity: 0;
      transform: translateY(24px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .modal__head {
    padding-bottom: var(--space-3);
    border-bottom: 1px solid var(--border-subtle);
  }

  .modal__title {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    line-height: 1.3;
    font-weight: 400;
    color: var(--content-primary);
    margin: 0;
  }

  .modal__body {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .modal__row {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    animation: row-in var(--duration-emphasized) var(--ease-emphasized) both;
  }

  @keyframes row-in {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .modal__row--meta {
    padding-top: var(--space-3);
    border-top: 1px solid var(--border-subtle);
    margin-top: var(--space-2);
  }

  .modal__label {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-tertiary);
  }

  .modal__value {
    font-family: var(--font-sans);
    font-size: var(--text-body-size);
    color: var(--content-primary);
    word-break: break-word;
  }

  .modal__value--target {
    font-family: var(--font-mono);
    font-size: var(--text-body-sm-size);
    color: var(--content-secondary);
    padding: var(--space-2) var(--space-3);
    background-color: var(--paper-warm-50);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
  }

  .modal__value--details {
    color: var(--content-secondary);
    line-height: 1.5;
  }

  .modal__foot {
    display: flex;
    gap: var(--space-3);
    justify-content: flex-end;
    padding-top: var(--space-3);
    border-top: 1px solid var(--border-subtle);
  }

  .modal__foot-hint {
    margin: 0;
    font-size: var(--text-caption-size);
    color: var(--content-muted);
    text-align: center;
  }

  @media (prefers-reduced-motion: reduce) {
    .scrim,
    .modal,
    .modal__row {
      animation: none;
      opacity: 1;
      transform: none;
    }
  }
</style>