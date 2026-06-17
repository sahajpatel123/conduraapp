/**
 * ConsentModal — native-looking Gatekeeper consent dialog.
 *
 * The modal appears when the daemon posts a pending consent ticket.
 * It shows the action, the target app/context, and two buttons:
 * Allow and Deny. A countdown bar shows the 5-minute timeout.
 */
<script lang="ts">
  import { consent } from '../stores/consent.svelte'

  function formatAction(kind: string): string {
    switch (kind?.toLowerCase()) {
      case 'read':
        return 'read from'
      case 'write':
        return 'write to'
      case 'network':
        return 'send data through'
      case 'destructive':
        return 'perform a destructive action on'
      default:
        return 'act on'
    }
  }

  function formatCountdown(ms: number): string {
    const s = Math.max(0, Math.ceil(ms / 1000))
    const m = Math.floor(s / 60)
    const r = s % 60
    return `${m}:${r.toString().padStart(2, '0')}`
  }
</script>

{#if consent.ticket}
  <div class="consent-backdrop" role="presentation">
    <div
      class="consent-modal"
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="consent-title"
      aria-describedby="consent-body"
    >
      <div class="consent-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          <path d="M12 8v4M12 16h.01" />
        </svg>
      </div>

      <h2 id="consent-title" class="consent-title">Allow this action?</h2>

      <p id="consent-body" class="consent-body">
        <strong>Condura</strong> wants to
        <strong>{formatAction(consent.ticket.action_kind)}</strong>
        {#if consent.ticket.detail}
          {consent.ticket.detail}
        {:else}
          an application
        {/if}
        {#if consent.ticket.actor}
          <span class="consent-meta">via {consent.ticket.actor}</span>
        {/if}
      </p>

      <div class="consent-countdown">
        <span class="consent-countdown-label">Expires in {formatCountdown(consent.timer)}</span>
        <div class="consent-countdown-bar">
          <div
            class="consent-countdown-fill"
            style:width="{(consent.timer / 300000) * 100}%"
          ></div>
        </div>
      </div>

      {#if consent.error}
        <p class="consent-error">{consent.error}</p>
      {/if}

      <div class="consent-actions">
        <button class="consent-deny" onclick={() => consent.deny()}>
          Deny
        </button>
        <button class="consent-allow" onclick={() => consent.approve()}>
          Allow
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .consent-backdrop {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: blur(4px);
  }

  .consent-modal {
    width: min(420px, calc(100vw - 48px));
    padding: 24px;
    border-radius: var(--radius-xl, 16px);
    background: var(--color-surface, #18181b);
    border: 1px solid var(--glass-border, rgba(255, 255, 255, 0.08));
    box-shadow: 0 24px 60px rgba(0, 0, 0, 0.45);
    text-align: center;
  }

  .consent-icon {
    width: 48px;
    height: 48px;
    margin: 0 auto 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: rgba(245, 158, 11, 0.12);
    color: var(--color-warning, #f59e0b);
  }

  .consent-icon svg {
    width: 24px;
    height: 24px;
  }

  .consent-title {
    margin: 0 0 12px;
    font-size: 18px;
    font-weight: 600;
    color: var(--color-text, #f4f4f5);
  }

  .consent-body {
    margin: 0 0 16px;
    font-size: 14px;
    line-height: 1.6;
    color: var(--color-text-muted, #a1a1aa);
  }

  .consent-body strong {
    color: var(--color-text, #f4f4f5);
    font-weight: 500;
  }

  .consent-meta {
    display: block;
    margin-top: 6px;
    font-size: 12px;
    color: var(--color-text-faint, #71717a);
  }

  .consent-countdown {
    margin-bottom: 20px;
  }

  .consent-countdown-label {
    display: block;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--color-text-faint, #71717a);
    margin-bottom: 6px;
  }

  .consent-countdown-bar {
    height: 4px;
    border-radius: var(--radius-pill, 999px);
    background: rgba(255, 255, 255, 0.08);
    overflow: hidden;
  }

  .consent-countdown-fill {
    height: 100%;
    background: var(--color-warning, #f59e0b);
    border-radius: var(--radius-pill, 999px);
    transition: width 1s linear;
  }

  .consent-error {
    margin: 0 0 16px;
    font-size: 12px;
    color: var(--color-error, #ef4444);
  }

  .consent-actions {
    display: flex;
    gap: 12px;
  }

  .consent-actions button {
    flex: 1;
    padding: 12px 16px;
    border-radius: var(--radius-lg, 10px);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition:
      background var(--transition-fast),
      transform var(--transition-fast),
      border-color var(--transition-fast);
    border: 1px solid transparent;
  }

  .consent-actions button:hover {
    transform: translateY(-1px);
  }

  .consent-deny {
    background: rgba(255, 255, 255, 0.06);
    color: var(--color-text, #f4f4f5);
    border-color: rgba(255, 255, 255, 0.1);
  }

  .consent-deny:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .consent-allow {
    background: var(--color-accent, #6366f1);
    color: #fff;
  }

  .consent-allow:hover {
    background: var(--color-accent-hover, #4f46e5);
  }
</style>
