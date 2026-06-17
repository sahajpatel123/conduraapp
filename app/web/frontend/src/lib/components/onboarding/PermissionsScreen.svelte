<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { ipc } from '../../ipc/client'
  import { onboarding } from '../../stores/onboarding.svelte'
  import type { PermissionStatus, PermissionGuide } from '../../ipc/types'

  // Only the two permissions computer-use actually needs up front.
  // Microphone / automation / notifications live in Settings and
  // are requested lazily when the user enables those features.
  const REQUIRED = ['accessibility', 'screen_recording']

  const LABELS: Record<string, string> = {
    accessibility: 'Accessibility',
    screen_recording: 'Screen Recording'
  }
  const WHY: Record<string, string> = {
    accessibility: 'Lets Condura click, type, and read UI elements in other apps.',
    screen_recording: 'Lets Condura see your screen to understand what to do.'
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
        title: `Grant ${LABELS[kind] ?? kind}`,
        steps: [String(err)]
      }
    }
  }

  function badgeClass(status: string): string {
    if (status === 'granted') return 'granted'
    if (status === 'denied') return 'denied'
    return 'unknown'
  }

  const allGranted = $derived(rows.every((r) => r.status === 'granted'))

  async function cont(): Promise<void> {
    await onboarding.completePermissions()
  }
  async function skip(): Promise<void> {
    await onboarding.skipStep('permissions')
  }
</script>

<div class="wizard perms">
  <h2>Grant access</h2>
  <p class="muted">
    Condura needs two permissions to control your computer. Grant them now for the full experience, or skip and enable
    later in Settings.
  </p>

  <div class="perm-list">
    {#each rows as row (row.kind)}
      <div class="perm-card">
        <div class="perm-head">
          <span class="perm-name">{LABELS[row.kind] ?? row.kind}</span>
          <span class="badge {badgeClass(row.status)}">{row.status}</span>
        </div>
        <p class="perm-why">{WHY[row.kind] ?? ''}</p>
        {#if row.status !== 'granted'}
          <button class="btn btn-secondary" onclick={() => openSettings(row.kind)}>Open System Settings</button>
        {/if}
      </div>
    {/each}
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
        <a class="full-link" href={guide.help_url} target="_blank" rel="noreferrer">More help</a>
      {/if}
      <button class="btn btn-ghost small" onclick={() => (guide = null)}>Close</button>
    </div>
  {/if}

  {#if onboarding.error}
    <p class="error">{onboarding.error}</p>
  {/if}

  <div class="actions">
    <button class="btn btn-ghost" onclick={skip} disabled={onboarding.busy}>Skip for now</button>
    <button class="btn btn-primary" onclick={cont} disabled={onboarding.busy}>
      {allGranted ? 'Continue' : 'Continue anyway'} →
    </button>
  </div>
</div>

<style>
  .wizard {
    width: 100%;
    max-width: 560px;
    padding: var(--space-6) var(--space-5);
    text-align: center;
  }
  h2 {
    font-size: var(--size-2xl);
    font-weight: 600;
    margin-bottom: var(--space-2);
  }
  .muted {
    color: var(--color-text-muted);
    font-size: var(--size-md);
    margin-bottom: var(--space-5);
  }
  .perm-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
  }
  .perm-card {
    text-align: left;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
  }
  .perm-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-2);
  }
  .perm-name {
    font-weight: 600;
    font-size: var(--size-md);
  }
  .perm-why {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    margin: 0 0 var(--space-3) 0;
  }
  .badge {
    font-family: var(--font-mono);
    font-size: var(--size-xs, 11px);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 2px 10px;
    border-radius: var(--radius-pill);
    border: 1px solid var(--glass-border);
  }
  .badge.granted {
    color: var(--color-success);
    border-color: var(--color-success);
  }
  .badge.denied {
    color: var(--color-error);
    border-color: var(--color-error);
  }
  .badge.unknown {
    color: var(--color-text-faint);
  }
  .guide-box {
    text-align: left;
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
    margin-bottom: var(--space-4);
  }
  .guide-box ol {
    padding-left: var(--space-5);
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    line-height: 1.6;
  }
  .full-link {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    text-decoration: underline;
  }
  .actions {
    display: flex;
    justify-content: space-between;
    margin-top: var(--space-4);
  }
  .btn {
    padding: 12px 24px;
    border-radius: var(--radius-pill);
    font-size: var(--size-md);
    font-weight: 500;
    cursor: pointer;
    border: none;
    transition: all var(--transition-spring);
  }
  .btn.small {
    padding: 6px 14px;
    font-size: var(--size-sm);
    margin-top: var(--space-2);
  }
  .btn-primary {
    background: var(--color-accent-gradient);
    color: white;
  }
  .btn-primary:hover:not(:disabled) {
    box-shadow: var(--shadow-glow);
    transform: translateY(-1px);
  }
  .btn-secondary {
    background: var(--glass-bg);
    color: var(--color-text);
    border: 1px solid var(--glass-border);
  }
  .btn-secondary:hover {
    background: rgba(255, 255, 255, 0.08);
  }
  .btn-ghost {
    background: transparent;
    color: var(--color-text-muted);
    border: 1px solid var(--glass-border);
  }
  .btn-ghost:hover {
    color: var(--color-text);
  }
  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .error {
    color: var(--color-error);
    font-size: var(--size-sm);
  }
</style>
