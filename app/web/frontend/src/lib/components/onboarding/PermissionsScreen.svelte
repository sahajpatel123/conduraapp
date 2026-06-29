<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { ipc } from '../../ipc/client'
  import { onboarding } from '../../stores/onboarding.svelte'
  import type { PermissionStatus, PermissionGuide } from '../../ipc/types'
  import Button from '../ui/Button.svelte'
  import Card from '../ui/Card.svelte'
  import Badge from '../ui/Badge.svelte'
  import { t } from '../../i18n'

  // Only the two permissions computer-use actually needs up front.
  // Microphone / automation / notifications live in Settings and
  // are requested lazily when the user enables those features.
  const REQUIRED = ['accessibility', 'screen_recording']

  const LABELS: Record<string, string> = {
    accessibility: 'Accessibility',
    screen_recording: 'Screen Recording'
  }
  const WHY_ACCESSIBILITY = $derived(t('onboarding.permissions.why_accessibility'))
  const WHY_SCREEN = $derived(t('onboarding.permissions.why_screen_recording'))
  function whyFor(kind: string): string {
    if (kind === 'accessibility') return WHY_ACCESSIBILITY
    if (kind === 'screen_recording') return WHY_SCREEN
    return ''
  }

  let statuses = $state<PermissionStatus[]>([])
  let guide = $state<PermissionGuide | null>(null)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const rows = $derived(
    REQUIRED.map((kind) => statuses.find((s) => s.kind === kind) ?? { kind, status: 'unknown' as const })
  )

  async function refresh(): Promise<void> {
    try {
      statuses = await ipc.permissionsStatus()
    } catch {
      // keep last-known; daemon may be briefly busy
    }
  }

  onMount(() => {
    void refresh()
    pollTimer = setInterval(refresh, 2000)
  })

  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer)
  })

  function openExternal(url: string): void {
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } }
    if (w.runtime?.BrowserOpenURL) {
      w.runtime.BrowserOpenURL(url)
    } else {
      window.open(url, '_blank')
    }
  }

  async function openSettings(kind: string): Promise<void> {
    try {
      const g = await ipc.permissionsGuide(kind)
      guide = g
      if (g.deep_link) openExternal(g.deep_link)
    } catch (err) {
      guide = {
        kind,
        platform: '',
        title: t('onboarding.permissions.grant_title', LABELS[kind] ?? kind),
        steps: [String(err)]
      }
    }
  }

  function badgeTone(status: string): 'success' | 'warn' | 'neutral' {
    if (status === 'granted') return 'success'
    if (status === 'denied') return 'warn'
    return 'neutral'
  }

  function badgeLabel(status: string): string {
    if (status === 'granted') return t('onboarding.permissions.status_granted')
    if (status === 'denied') return t('onboarding.permissions.status_denied')
    if (status === 'pending') return t('onboarding.permissions.status_pending')
    return t('onboarding.permissions.status_unknown')
  }

  // The gate lets the user proceed when at least one permission
  // is granted (the spec says "at least one of the two"). The
  // primary CTA still reads "Continue" so the user doesn't feel
  // blocked when both are denied — they can opt out and grant
  // later from Settings.
  const atLeastOneGranted = $derived(rows.some((r) => r.status === 'granted'))
  const canContinue = $derived(atLeastOneGranted || !onboarding.busy)

  async function cont(): Promise<void> {
    await onboarding.completePermissions()
  }
  async function skip(): Promise<void> {
    await onboarding.skipStep('permissions')
  }
  async function back(): Promise<void> {
    await onboarding.back()
  }
</script>

<div class="wizard perms">
  <header class="head">
    <h2>{t('onboarding.permissions.title')}</h2>
    <p class="muted">
      {t('onboarding.permissions.intro')}
    </p>
  </header>

  <div class="perm-grid">
    {#each rows as row, i (row.kind)}
      <div class="stagger-item" style="--stagger-index: {i}">
        <Card elevation="glass" padding="md" class="perm-card">
          <div class="perm-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              {#if row.kind === 'accessibility'}
                <path d="M12 4v16" />
                <path d="M4 12h16" />
                <circle cx="12" cy="12" r="9" />
                <path d="M8 12h.01" />
                <path d="M16 12h.01" />
              {:else}
                <rect x="3" y="4" width="18" height="13" rx="2" />
                <path d="M3 9h18" />
                <path d="M8 21h8" />
                <path d="M12 17v4" />
              {/if}
            </svg>
          </div>
          <div class="perm-head">
            <span class="perm-name">{LABELS[row.kind] ?? row.kind}</span>
            <Badge tone={badgeTone(row.status)} dot>
              {badgeLabel(row.status)}
            </Badge>
          </div>
          <p class="perm-why">{whyFor(row.kind)}</p>
          <div class="perm-actions">
            {#if row.status === 'granted'}
              <span class="perm-granted-note">
                {t('onboarding.permissions.granted_note')}
              </span>
            {:else}
              <Button
                variant="secondary"
                size="sm"
                onclick={() => openSettings(row.kind)}
              >
                {t('onboarding.permissions.open_settings')}
              </Button>
            {/if}
            <button
              class="skip-link"
              type="button"
              onclick={skip}
              disabled={onboarding.busy}
            >
              {t('onboarding.permissions.skip_link')}
            </button>
          </div>
        </Card>
      </div>
    {/each}
  </div>

  <div class="explainer">
    <h4>{t('onboarding.permissions.why_title')}</h4>
    <p>{t('onboarding.permissions.why_body')}</p>
  </div>

  {#if guide}
    <div class="guide-box">
      <h4>{guide.title}</h4>
      <ol>
        {#each guide.steps as step}
          <li>{step}</li>
        {/each}
      </ol>
      {#if guide.help_url}
        <a class="full-link" href={guide.help_url} target="_blank" rel="noreferrer">
          {t('onboarding.permissions.more_help')}
        </a>
      {/if}
      <button class="close-link" type="button" onclick={() => (guide = null)}>
        {t('onboarding.permissions.close')}
      </button>
    </div>
  {/if}

  {#if onboarding.error}
    <p class="error">{onboarding.error}</p>
  {/if}

  <div class="actions">
    <button class="back-link" type="button" onclick={back} disabled={onboarding.busy}>
      ← {t('onboarding.permissions.back')}
    </button>
    <div class="actions-right">
      <button class="skip-link" type="button" onclick={skip} disabled={onboarding.busy}>
        {t('onboarding.permissions.skip')}
      </button>
      <Button variant="primary" size="md" onclick={cont} disabled={!canContinue} loading={onboarding.busy}>
        {t('onboarding.permissions.continue')}
      </Button>
    </div>
  </div>
</div>

<style>
  .wizard {
    width: 100%;
    max-width: 720px;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  .head { display: flex; flex-direction: column; gap: var(--space-2); text-align: center; align-items: center; }

  h2 {
    font-family: var(--font-display);
    font-size: var(--size-2xl);
    font-weight: var(--weight-light);
    letter-spacing: var(--tracking-tight);
    margin: 0;
  }

  .muted {
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    color: var(--text-muted);
    margin: 0;
    max-width: 52ch;
  }

  .perm-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--space-4);
  }
  @media (max-width: 640px) {
    .perm-grid { grid-template-columns: 1fr; }
  }

  .perm-icon {
    width: 36px;
    height: 36px;
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--accent-soft);
    color: var(--accent);
    margin-bottom: var(--space-3);
  }
  .perm-icon svg { width: 20px; height: 20px; }

  .perm-card { height: 100%; }
  .perm-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-2);
  }
  .perm-name {
    font-weight: var(--weight-semibold);
    font-size: var(--size-md);
  }
  .perm-why {
    color: var(--text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-relaxed);
    margin: 0 0 var(--space-3) 0;
    flex: 1;
  }
  .perm-actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
  }
  .perm-granted-note {
    color: var(--success);
    font-size: var(--size-xs);
    font-family: var(--font-mono);
    letter-spacing: var(--tracking-wide);
    text-transform: uppercase;
  }

  .explainer {
    text-align: left;
    padding: var(--space-4);
    border: 1px dashed var(--border);
    border-radius: var(--radius-md);
    background: var(--surface-1);
  }
  .explainer h4 {
    font-size: var(--size-sm);
    font-weight: var(--weight-semibold);
    margin: 0 0 var(--space-1) 0;
    color: var(--text);
  }
  .explainer p {
    color: var(--text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-relaxed);
    margin: 0;
  }

  .guide-box {
    text-align: left;
    padding: var(--space-4);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    background: var(--surface-2);
  }
  .guide-box h4 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    margin: 0 0 var(--space-2) 0;
  }
  .guide-box ol {
    padding-left: var(--space-5);
    color: var(--text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-relaxed);
    margin: 0 0 var(--space-2) 0;
  }
  .full-link {
    color: var(--accent);
    font-size: var(--size-sm);
    text-decoration: underline;
  }

  .actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .actions-right {
    display: flex;
    gap: var(--space-3);
    align-items: center;
  }

  .back-link,
  .skip-link,
  .close-link {
    background: transparent;
    border: 0;
    padding: 0;
    cursor: pointer;
    color: var(--text-muted);
    font-size: var(--size-sm);
    font-family: inherit;
    transition: color var(--transition-fast);
  }
  .back-link:hover:not(:disabled),
  .skip-link:hover:not(:disabled),
  .close-link:hover:not(:disabled) {
    color: var(--text);
  }
  .back-link:disabled,
  .skip-link:disabled,
  .close-link:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .skip-link {
    color: var(--text-faint);
  }

  .error {
    color: var(--error);
    font-size: var(--size-sm);
  }
</style>
