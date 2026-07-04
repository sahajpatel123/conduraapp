<!--
  Toast — single notification message.

  Per spec: every action produces a "receipt" of what changed. The Toast
  is the in-app version of that — surfaces non-blocking feedback for
  save events, errors, and the rare "Synaptic is thinking" tip.

  Five variants:
    - info:    neutral, low-attention
    - success: muted green, completion confirmation
    - warning: muted amber, soft attention
    - error:   muted red, requires attention
    - agent:   accent plum, the agent speaking

  Props:
    variant   — 'info' | 'success' | 'warning' | 'error' | 'agent'
    title     — primary line
    message   — secondary line (optional)
    icon      — optional custom icon name (defaults to variant icon)
    duration  — auto-dismiss in ms (0 = sticky, default 4000)
    onclose   — handler when dismissed
-->
<script lang="ts">
  import Icon, { type IconName } from './icons/Icon.svelte';
  import IconButton from './IconButton.svelte';

  type Variant = 'info' | 'success' | 'warning' | 'error' | 'agent';

  interface Props {
    variant?: Variant;
    title: string;
    message?: string;
    icon?: IconName;
    duration?: number;
    onclose?: () => void;
  }

  let {
    variant = 'info',
    title,
    message,
    icon,
    duration = 4000,
    onclose,
  }: Props = $props();

  // Default icon per variant — only if not explicitly overridden
  const VARIANT_ICON: Record<Variant, IconName> = {
    info: 'sparkle',
    success: 'check',
    warning: 'sparkle',
    error: 'x',
    agent: 'chat',
  };

  let effectiveIcon = $derived(icon ?? VARIANT_ICON[variant]);

  // Auto-dismiss timer
  $effect(() => {
    if (duration > 0) {
      const id = setTimeout(() => onclose?.(), duration);
      return () => clearTimeout(id);
    }
  });
</script>

<div class="toast toast--{variant}" role="status" aria-live="polite">
  <div class="toast__icon" aria-hidden="true">
    <Icon name={effectiveIcon} size="sm" />
  </div>

  <div class="toast__body">
    <div class="toast__title">{title}</div>
    {#if message}
      <div class="toast__message">{message}</div>
    {/if}
  </div>

  {#if onclose}
    <div class="toast__action">
      <IconButton name="x" label="Dismiss" size={28} variant="ghost" onclick={onclose} />
    </div>
  {/if}
</div>

<style>
  .toast {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-3) var(--space-3) var(--space-4);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-left: 3px solid var(--content-muted);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-3);
    min-width: 280px;
    max-width: 420px;
    font-family: var(--font-sans);
    color: var(--content-primary);
    /* The entrance — slide up + fade in */
    animation: toast-in var(--duration-base) var(--ease-decelerate) both;
  }

  @keyframes toast-in {
    from {
      opacity: 0;
      transform: translateY(8px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  /* ── Variant colors — left border + icon color ──────────────── */
  .toast--info {
    border-left-color: var(--info-500);
  }
  .toast--info .toast__icon {
    color: var(--info-500);
  }

  .toast--success {
    border-left-color: var(--success-500);
  }
  .toast--success .toast__icon {
    color: var(--success-500);
  }

  .toast--warning {
    border-left-color: var(--warning-500);
  }
  .toast--warning .toast__icon {
    color: var(--warning-500);
  }

  .toast--error {
    border-left-color: var(--error-500);
  }
  .toast--error .toast__icon {
    color: var(--error-500);
  }

  .toast--agent {
    border-left-color: var(--content-accent);
  }
  .toast--agent .toast__icon {
    color: var(--content-accent);
  }

  /* ── Icon ─────────────────────────────────────────────── */
  .toast__icon {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    margin-top: 2px;
  }

  /* ── Body ────────────────────────────────────────────── */
  .toast__body {
    flex: 1;
    min-width: 0;
  }

  .toast__title {
    font-size: var(--text-body-size);
    font-weight: 500;
    color: var(--content-primary);
    line-height: 1.4;
  }

  .toast__message {
    margin-top: 2px;
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    line-height: 1.5;
  }

  /* ── Action ───────────────────────────────────────────── */
  .toast__action {
    flex-shrink: 0;
    margin: -4px -4px 0 0;
  }

  @media (prefers-reduced-motion: reduce) {
    .toast {
      animation: none;
    }
  }
</style>