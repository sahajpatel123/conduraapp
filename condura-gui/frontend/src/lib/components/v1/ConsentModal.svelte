<!--
  ConsentModal — the ONLY modal in Synaptic.

  Per spec §10.2 (safety): for any WRITE / NETWORK / DESTRUCTIVE action,
  this modal appears. Native-feeling (system dialog colors and sizing).
  Cascades text in with deliberate-stagger so the user is forced to read.

  Per motion agent §3.6: 420ms emphasized timing. Stagger 180ms per item.
  Approve button disabled for 1.2s during the cascade.

  Refinements:
    - Blast-radius icon prefix indicates the type of action at a glance
    - Target row visually emphasized (the thing being acted upon)
    - "About to:" preview line uses italic serif for the agent's voice
    - Approve button shows a kbd hint (the shortcut for "always allow")
    - Subtle ambient plum glow behind the modal for the brand presence

  Props:
    verb      — the action verb: "Send", "Delete", "Open", etc.
    target    — the target: email address, file path, URL
    details   — additional context to show the user
    blastRadius — 'read' | 'write' | 'network' | 'destructive' (default: derived from verb)
    onapprove — handler for "Allow"
    ondeny    — handler for "Don't allow"
-->
<script lang="ts">
  import Button from './Button.svelte';
  import Icon, { type IconName } from './icons/Icon.svelte';
  import KeyCombo from './KeyCombo.svelte';

  type BlastRadius = 'read' | 'write' | 'network' | 'destructive';

  interface Props {
    verb: string;
    target: string;
    details?: string;
    blastRadius?: BlastRadius;
    onapprove?: () => void;
    ondeny?: () => void;
  }

  let { verb, target, details, blastRadius, onapprove, ondeny }: Props = $props();

  let approveEnabled = $state(false);

  // Derive blast radius from verb if not provided
  const DESTRUCTIVE_VERBS = ['delete', 'remove', 'destroy', 'wipe', 'erase', 'transfer', 'send-money', 'purchase'];
  const NETWORK_VERBS = ['send', 'post', 'publish', 'submit', 'email', 'message', 'tweet', 'upload'];
  const WRITE_VERBS = ['create', 'write', 'edit', 'save', 'rename', 'move', 'delete-file'];
  // (READ_VERBS = all else)

  let effectiveRadius = $derived<BlastRadius>(
    blastRadius ?? (
      DESTRUCTIVE_VERBS.includes(verb.toLowerCase()) ? 'destructive' :
      NETWORK_VERBS.includes(verb.toLowerCase()) ? 'network' :
      WRITE_VERBS.includes(verb.toLowerCase()) ? 'write' :
      'read'
    )
  );

  // Per-radius icon and color
  const RADIUS_META: Record<BlastRadius, { icon: IconName; label: string; color: string }> = {
    read:        { icon: 'eye',     label: 'Read',        color: 'var(--ink-cool-500)' },
    write:       { icon: 'edit',    label: 'Write',       color: 'var(--info-500)' },
    network:     { icon: 'globe',   label: 'Network',     color: 'var(--warning-500)' },
    destructive: { icon: 'trash',   label: 'Destructive', color: 'var(--error-500)' },
  };

  let meta = $derived(RADIUS_META[effectiveRadius]);

  // Per motion agent §3.6: 1.2s forced dwell before approve becomes enabled.
  $effect(() => {
    const id = setTimeout(() => {
      approveEnabled = true;
    }, 1200);
    return () => clearTimeout(id);
  });

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      ondeny?.();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="modal-host" role="presentation">
  <div class="scrim" aria-hidden="true"></div>

  <!-- Ambient plum bloom behind the modal — subtle brand presence -->
  <div class="modal-bloom" aria-hidden="true"></div>

  <div
    class="modal"
    role="alertdialog"
    aria-modal="true"
    aria-labelledby="consent-title"
    aria-describedby="consent-body"
  >
    <header class="modal__head">
      <div class="modal__blast" style="--blast-color: {meta.color};">
        <Icon name={meta.icon} size="sm" />
        <span>{meta.label}</span>
      </div>
      <h2 id="consent-title" class="modal__title">
        Synaptic wants to <em>{verb.toLowerCase()}</em>.
      </h2>
      <p class="modal__about">
        <em>About to {verb.toLowerCase()}</em>
        <code class="modal__about-target">{target}</code>
      </p>
    </header>

    <div id="consent-body" class="modal__body">
      <div class="modal__row modal__row--target" style="animation-delay: 0ms">
        <span class="modal__label">Target</span>
        <span class="modal__value modal__value--target">{target}</span>
      </div>

      {#if details}
        <div class="modal__row" style="animation-delay: 180ms">
          <span class="modal__label">Details</span>
          <span class="modal__value modal__value--details">{details}</span>
        </div>
      {/if}

      <div class="modal__row modal__row--meta" style="animation-delay: {details ? 360 : 180}ms">
        <span class="modal__label">You can revoke this any time</span>
        <span class="modal__value">I'll stop the moment you do.</span>
      </div>
    </div>

    <footer class="modal__foot">
      <Button variant="secondary" size="md" onclick={ondeny}>
        Don't allow
      </Button>
      <div class="modal__approve">
        <Button
          variant="primary"
          size="md"
          disabled={!approveEnabled}
          onclick={onapprove}
        >
          {approveEnabled ? 'Allow this once' : 'Reading…'}
        </Button>
        <span class="modal__approve-hint">
          <KeyCombo combo="⌘↩" size="sm" /> to allow
        </span>
      </div>
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

  /* The ambient brand bloom — subtle plum glow behind the modal.
     Reinforces "this is Synaptic asking" without being aggressive. */
  .modal-bloom {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 800px;
    height: 800px;
    background: radial-gradient(
      circle,
      var(--content-accent) 0%,
      transparent 60%
    );
    opacity: 0.04;
    pointer-events: none;
    animation: bloom-in 800ms var(--ease-decelerate) 200ms both;
  }

  @keyframes bloom-in {
    from {
      opacity: 0;
      transform: translate(-50%, -50%) scale(0.8);
    }
    to {
      opacity: 0.04;
      transform: translate(-50%, -50%) scale(1);
    }
  }

  .modal {
    position: relative;
    width: 520px;
    max-width: calc(100vw - 32px);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-4);
    padding: var(--space-7);
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    animation: modal-in var(--duration-emphasized) var(--ease-emphasized) both;
  }

  @keyframes modal-in {
    from {
      opacity: 0;
      transform: translateY(24px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  /* ── Head ──────────────────────────────────────────────── */
  .modal__head {
    padding-bottom: var(--space-3);
    border-bottom: 1px solid var(--border-subtle);
  }

  /* Blast-radius badge — color-coded by action type */
  .modal__blast {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) var(--space-3);
    background-color: color-mix(in srgb, var(--blast-color) 10%, transparent);
    color: var(--blast-color);
    border: 1px solid color-mix(in srgb, var(--blast-color) 20%, transparent);
    border-radius: var(--radius-pill);
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    margin-bottom: var(--space-3);
  }

  .modal__title {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    line-height: 1.3;
    font-weight: 400;
    color: var(--content-primary);
    margin: 0 0 var(--space-3) 0;
  }

  .modal__title em {
    font-style: italic;
    color: var(--content-accent);
    font-weight: 500;
  }

  /* The "About to" preview — italic serif, the agent speaking */
  .modal__about {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    margin: 0;
    font-family: var(--font-serif);
    font-style: italic;
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
  }

  .modal__about-target {
    font-family: var(--font-mono);
    font-size: var(--text-body-sm-size);
    font-style: normal;
    color: var(--content-primary);
    background-color: var(--paper-warm-50);
    border: 1px solid var(--border-subtle);
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-sm);
    word-break: break-all;
  }

  /* ── Body ─────────────────────────────────────────────── */
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

  .modal__row--target {
    padding: var(--space-3);
    background-color: var(--plum-50);
    border: 1px solid var(--plum-200);
    border-radius: var(--radius-md);
    /* The target is the most important piece — it's what gets acted upon.
       Make it visually unmistakable. */
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
    line-height: 1.5;
  }

  .modal__value--target {
    font-family: var(--font-mono);
    font-size: var(--text-body-size);
    color: var(--plum-900);
    font-weight: 500;
  }

  .modal__value--details {
    color: var(--content-secondary);
  }

  /* ── Foot ─────────────────────────────────────────────── */
  .modal__foot {
    display: flex;
    gap: var(--space-3);
    justify-content: space-between;
    align-items: center;
    padding-top: var(--space-3);
    border-top: 1px solid var(--border-subtle);
  }

  .modal__approve {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .modal__approve-hint {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    letter-spacing: 0.02em;
  }

  .modal__foot-hint {
    margin: 0;
    font-size: var(--text-caption-size);
    color: var(--content-muted);
    text-align: center;
    font-style: italic;
    font-family: var(--font-serif);
  }

  /* ── Reduced motion ───────────────────────────────────── */
  @media (prefers-reduced-motion: reduce) {
    .scrim,
    .modal,
    .modal-bloom,
    .modal__row {
      animation: none;
      opacity: 1;
      transform: none;
    }
    .modal-bloom {
      opacity: 0.04;
    }
  }
</style>