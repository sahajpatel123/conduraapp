/**
 * ConsentModal — native-looking Gatekeeper consent dialog.
 *
 * The modal appears when the daemon posts a pending consent ticket.
 * It shows the action, the target app/context, and two buttons:
 * Allow and Deny. A countdown bar shows the 5-minute timeout.
 */
<script lang="ts">
  import { consent } from '../stores/consent.svelte'
  import { t } from '../i18n'

  function formatAction(kind: string): string {
    switch (kind?.toLowerCase()) {
      case 'read':
        return t('consent.action.read')
      case 'write':
        return t('consent.action.write')
      case 'network':
        return t('consent.action.network')
      case 'destructive':
        return t('consent.action.destructive')
      default:
        return t('consent.action.default')
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
      class="consent-modal glass-card elevated"
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

      <h2 id="consent-title" class="consent-title">{t('consent.title')}</h2>

      <p id="consent-body" class="consent-body">
        <strong>Condura</strong> {t('consent.wants_to')}
        <strong>{formatAction(consent.ticket.action_kind)}</strong>
        {#if consent.ticket.detail}
          {consent.ticket.detail}
        {:else}
          {t('consent.an_application')}
        {/if}
        {#if consent.ticket.actor}
          <span class="consent-meta">{t('consent.via', consent.ticket.actor)}</span>
        {/if}
      </p>

      <div class="consent-countdown">
        <span class="consent-countdown-label">{t('consent.expires_in', formatCountdown(consent.timer))}</span>
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
        <button class="btn btn-secondary btn-lg consent-deny" onclick={() => consent.deny()}>
          {t('consent.deny')}
        </button>
        <button class="btn btn-primary btn-lg consent-allow" onclick={() => consent.approve()}>
          {t('consent.allow')}
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
    background: rgba(20, 17, 11, 0.45);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    animation: backdrop-in var(--transition-base) ease both;
  }

  .consent-modal {
    width: min(440px, calc(100vw - 48px));
    padding: var(--space-6);
    text-align: center;
    animation: modal-in var(--transition-spring) var(--ease-out-expo) both;
  }
  .consent-modal:hover {
    border-color: var(--glass-border-hover);
  }

  .consent-icon {
    width: 52px;
    height: 52px;
    margin: 0 auto var(--space-4);
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: var(--color-warn-soft);
    color: var(--color-warn);
    animation: pulse-glow 2.6s ease-in-out infinite;
  }
  .consent-icon svg {
    width: 26px;
    height: 26px;
  }

  .consent-title {
    margin: 0 0 var(--space-3);
    font-size: var(--size-xl);
    font-weight: var(--weight-semibold);
    letter-spacing: var(--tracking-tight);
    color: var(--color-text);
  }

  .consent-body {
    margin: 0 0 var(--space-4);
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    color: var(--color-text-muted);
  }
  .consent-body strong {
    color: var(--color-text);
    font-weight: var(--weight-semibold);
  }
  .consent-meta {
    display: block;
    margin-top: var(--space-2);
    font-size: var(--size-xs);
    color: var(--color-text-faint);
  }

  .consent-countdown {
    margin-bottom: var(--space-5);
  }
  .consent-countdown-label {
    display: block;
    font-size: var(--size-xs);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
    color: var(--color-text-faint);
    margin-bottom: var(--space-2);
  }
  .consent-countdown-bar {
    height: 8px;
    border-radius: var(--radius-pill);
    background: var(--color-bg-active);
    overflow: hidden;
    box-shadow: inset 0 1px 2px rgba(20, 17, 11, 0.18);
  }
  .consent-countdown-fill {
    height: 100%;
    background: var(--color-accent-gradient);
    border-radius: var(--radius-pill);
    transition: width 1s linear;
    box-shadow: 0 0 12px var(--color-glow-strong), 0 0 24px var(--color-glow);
  }

  .consent-error {
    margin: 0 0 var(--space-4);
    font-size: var(--size-sm);
    color: var(--color-error);
  }

  .consent-actions {
    display: flex;
    gap: var(--space-3);
  }
  .consent-actions .btn {
    flex: 1;
  }
</style>
