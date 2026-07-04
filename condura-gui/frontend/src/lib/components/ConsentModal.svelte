<script lang="ts">
  // ConsentModal — native-looking Gatekeeper consent dialog.
  //
  // Appears when the daemon posts a pending consent ticket. Two buttons:
  // Deny (ghost) and Approve (primary). For DESTRUCTIVE actions we also
  // expose "Approve for this session" (danger variant). A countdown bar
  // shows the timeout.
  import { consent } from '../stores/consent.svelte'
  import { Dialog } from './ui'
  import Button from './ui/Button.svelte'
  import Badge from './ui/Badge.svelte'
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

  const isDestructive = $derived(
    consent.ticket?.action_kind?.toLowerCase() === 'destructive'
  )

  function approve(): void {
    void consent.approve()
  }

  function deny(): void {
    consent.deny()
  }

  function approveSession(): void {
    // For destructive actions: same as approve in this build. The store
    // currently exposes only approve/deny; the "session" variant is a
    // separate future RPC. Calling approve() here keeps the spec intent
    // visible at the call site.
    void consent.approve()
  }
</script>

<Dialog
  open={consent.ticket !== null}
  title={t('consent.title')}
  size="sm"
  onclose={deny}
>
  {#snippet children()}
    <div class="consent-body">
      <div class="consent-icon anim-glow-pulse">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          <path d="M12 8v4M12 16h.01" />
        </svg>
      </div>

      <p class="consent-text">
        <strong>Condura</strong> {t('consent.wants_to')}
        <strong>{formatAction(consent.ticket?.action_kind ?? '')}</strong>
        {#if consent.ticket?.detail}
          {consent.ticket.detail}
        {:else}
          {t('consent.an_application')}
        {/if}
      </p>

      {#if consent.ticket?.actor}
        <span class="consent-meta">{t('consent.via', consent.ticket.actor)}</span>
      {/if}

      <div class="consent-countdown">
        <div class="consent-countdown-row">
          <span class="consent-countdown-label">{t('consent.expires_in', formatCountdown(consent.timer))}</span>
          <Badge tone={isDestructive ? 'error' : 'warn'} size="sm">
            {isDestructive ? t('consent.action.destructive') : t('consent.action.network')}
          </Badge>
        </div>
        <div class="consent-countdown-bar">
          <div
            class="consent-countdown-fill"
            style:width="{(consent.timer / 300000) * 100}%"
          ></div>
        </div>
      </div>

      {#if consent.ticket?.detail}
        <blockquote class="consent-reason">{consent.ticket.detail}</blockquote>
      {/if}

      {#if consent.error}
        <p class="consent-error">{consent.error}</p>
      {/if}
    </div>
  {/snippet}
  {#snippet footer()}
    <div class="consent-actions">
      <Button variant="ghost" onclick={deny} fullWidth>
        {t('consent.deny')}
      </Button>
      {#if isDestructive}
        <Button variant="danger" onclick={approveSession} fullWidth>
          {t('consent.approve_session')}
        </Button>
      {:else}
        <Button variant="primary" onclick={approve} fullWidth>
          {t('consent.allow')}
        </Button>
      {/if}
    </div>
  {/snippet}
</Dialog>

<style>
  .consent-body {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-3);
    text-align: center;
    padding: var(--space-2) 0;
  }
  .consent-icon {
    width: 52px;
    height: 52px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: var(--warn-soft);
    color: var(--warn);
  }
  .consent-icon svg {
    width: 26px;
    height: 26px;
  }

  .consent-text {
    margin: 0;
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    color: var(--text-muted);
  }
  .consent-text strong {
    color: var(--text);
    font-weight: var(--weight-semibold);
  }
  .consent-meta {
    display: block;
    font-size: var(--size-xs);
    color: var(--text-faint);
  }

  .consent-countdown {
    width: 100%;
    margin-top: var(--space-2);
  }
  .consent-countdown-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-2);
  }
  .consent-countdown-label {
    font-size: var(--size-xs);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
    color: var(--text-faint);
    font-family: var(--font-mono);
  }
  .consent-countdown-bar {
    height: 6px;
    border-radius: var(--radius-pill);
    background: var(--surface-3);
    overflow: hidden;
  }
  .consent-countdown-fill {
    height: 100%;
    background: var(--accent-gradient);
    border-radius: var(--radius-pill);
    transition: width 1s linear;
    box-shadow: 0 0 12px var(--accent-glow);
  }

  .consent-reason {
    width: 100%;
    margin: 0;
    padding: var(--space-3);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-left: 3px solid var(--accent);
    border-radius: var(--radius-md);
    font-size: var(--size-sm);
    font-style: italic;
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    text-align: left;
  }

  .consent-error {
    margin: 0;
    font-size: var(--size-sm);
    color: var(--error);
  }

  .consent-actions {
    display: flex;
    gap: var(--space-3);
    width: 100%;
  }
</style>