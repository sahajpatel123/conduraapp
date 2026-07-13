<script lang="ts">
  /**
   * Meridian OS permissions instrument — live probe + System Settings grant/revoke.
   * Condura cannot flip TCC itself; this desk only opens the OS pane and rechecks.
   */
  import { onMount } from 'svelte'
  import { trust } from '../../stores/trust.svelte'
  import { ipc } from '../../ipc/client'
  import { openPermissionSettings } from '../../utils/openPermissionSettings'
  import type { PermissionGuide } from '../../ipc/types'

  type PermKind = 'accessibility' | 'screen_recording' | 'microphone' | 'notifications' | 'automation'

  let permBusy = $state<string | null>(null)
  let permActionNote = $state('')
  let guideByKind = $state<Partial<Record<PermKind, PermissionGuide | null>>>({})
  let expandedGuide = $state<PermKind | null>(null)
  let guideLoading = $state<PermKind | null>(null)

  const PERM_KINDS: PermKind[] = [
    'accessibility',
    'screen_recording',
    'microphone',
    'automation',
    'notifications',
  ]

  const PERM_META: Record<
    PermKind,
    { mark: string; name: string; body: string; need: string; weight: 'required' | 'optional' }
  > = {
    accessibility: {
      mark: 'AX',
      name: 'Accessibility',
      body: 'Lets Condura see buttons and windows so it can help on your Mac.',
      need: 'Needed for computer control',
      weight: 'required',
    },
    screen_recording: {
      mark: 'SCR',
      name: 'Screen Recording',
      body: 'Lets Condura look at the screen only when a task needs it.',
      need: 'Optional',
      weight: 'optional',
    },
    microphone: {
      mark: 'MIC',
      name: 'Microphone',
      body: 'For voice input and “hey Condura,” if you turn those on.',
      need: 'Optional',
      weight: 'optional',
    },
    automation: {
      mark: 'AUTO',
      name: 'Automation',
      body: 'Lets Condura control other apps when you allow it.',
      need: 'Needed for cross-app tasks',
      weight: 'required',
    },
    notifications: {
      mark: 'NTF',
      name: 'Notifications',
      body: 'Alerts when a task finishes or needs you.',
      need: 'Optional',
      weight: 'optional',
    },
  }

  const permOffline = $derived(
    !!trust.lastError && /IPC client not started|not connected|Failed to fetch|daemon/i.test(trust.lastError)
  )

  const permissionRows = $derived(
    PERM_KINDS.map((kind) => {
      const status = trust.permissions.find((p) => p.kind === kind)
      const meta = PERM_META[kind]
      const state = (status?.status ?? 'unknown') as 'granted' | 'denied' | 'unknown'
      return {
        kind,
        ...meta,
        state,
        granted: state === 'granted',
        denied: state === 'denied',
        note: status?.note ?? '',
      }
    })
  )

  const grantedCount = $derived(permissionRows.filter((p) => p.granted).length)
  const deniedCount = $derived(permissionRows.filter((p) => p.denied).length)
  const allGranted = $derived(grantedCount === PERM_KINDS.length)
  const requiredMissing = $derived(
    permissionRows.filter((p) => p.weight === 'required' && !p.granted).length
  )

  onMount(() => {
    void trust.refreshPermissions()
    const poll = setInterval(() => void trust.refreshPermissions({ quiet: true }), 2000)
    return () => clearInterval(poll)
  })

  function statusLabel(state: string): string {
    if (state === 'granted') return 'On'
    if (state === 'denied') return 'Off'
    return 'Unknown'
  }

  async function ensureGuide(kind: PermKind): Promise<PermissionGuide | null> {
    if (kind in guideByKind) return guideByKind[kind] ?? null
    guideLoading = kind
    try {
      const g = await trust.loadGuide(kind)
      guideByKind = { ...guideByKind, [kind]: g }
      return g
    } catch {
      guideByKind = { ...guideByKind, [kind]: null }
      return null
    } finally {
      if (guideLoading === kind) guideLoading = null
    }
  }

  async function toggleGuide(kind: PermKind): Promise<void> {
    if (expandedGuide === kind) {
      expandedGuide = null
      return
    }
    expandedGuide = kind
    await ensureGuide(kind)
  }

  async function openOSPermission(kind: PermKind, intent: 'grant' | 'revoke'): Promise<void> {
    if (permBusy) return
    permBusy = kind
    permActionNote = ''
    expandedGuide = kind
    try {
      await ensureGuide(kind)
      const res = await openPermissionSettings(kind, ipc)
      await trust.refreshPermissions({ quiet: true })
      if (res.opened) {
        permActionNote =
          intent === 'grant'
            ? `Opened System Settings for ${PERM_META[kind].name}. Turn it on, then come back — we’ll recheck automatically.`
            : `Opened System Settings for ${PERM_META[kind].name}. Turn it off to revoke, then come back.`
      } else {
        permActionNote =
          res.error || 'Could not open System Settings. Use the steps below, then tap Recheck.'
      }
    } catch (e) {
      permActionNote = String(e)
      await ensureGuide(kind)
    } finally {
      permBusy = null
    }
  }

  async function recheckPermissions(): Promise<void> {
    permActionNote = ''
    await trust.refreshPermissions()
    if (!trust.lastError) {
      const n = trust.permissions.filter((p) => p.status === 'granted').length
      permActionNote = `Checked — ${n} of ${PERM_KINDS.length} on.`
    }
  }
</script>

<section class="plate perms" aria-labelledby="perms-title">
  <header class="plate-head perms-head">
    <div>
      <h2 id="perms-title">Mac permissions</h2>
      <p class="hint">
        Condura asks; your Mac decides. Grant what you need, then return here.
      </p>
    </div>
    <button
      type="button"
      class="md-btn recheck"
      disabled={trust.loadingPermissions && trust.permissions.length === 0}
      onclick={() => void recheckPermissions()}
    >
      {trust.loadingPermissions && trust.permissions.length === 0 ? 'Checking…' : 'Recheck'}
    </button>
  </header>

  <div class="perm-summary" data-state={allGranted ? 'ok' : deniedCount ? 'warn' : 'mid'}>
    <span class="perm-count">{grantedCount} / {PERM_KINDS.length}</span>
    <span class="perm-sum-copy">
      {#if permOffline}
        Offline — can’t check permissions right now
      {:else if trust.loadingPermissions && trust.permissions.length === 0}
        Checking…
      {:else if allGranted}
        All set
      {:else if requiredMissing > 0}
        {requiredMissing} still needed for full computer control
      {:else}
        Optional ones left — core features can still work
      {/if}
    </span>
  </div>

  {#if trust.lastError && !permOffline}
    <p class="perm-error">{trust.lastError}</p>
  {/if}

  <ul class="perm-list">
    {#each permissionRows as perm (perm.kind)}
      <li class="perm-row" data-state={perm.state}>
        <div class="perm-main">
          <span class="perm-mark" aria-hidden="true">{perm.mark}</span>
          <div class="perm-copy">
            <div class="perm-title-row">
              <strong>{perm.name}</strong>
              <span class="perm-need" data-weight={perm.weight}>{perm.need}</span>
            </div>
            <p>{perm.body}</p>
            {#if perm.note}
              <p class="perm-probe">{perm.note}</p>
            {/if}
          </div>
          <span class="perm-status" data-state={perm.state}>
            <i aria-hidden="true"></i>
            {statusLabel(perm.state)}
          </span>
        </div>

        <div class="perm-actions">
          {#if perm.granted}
            <button
              type="button"
              class="md-btn md-btn-danger"
              disabled={permBusy === perm.kind}
              onclick={() => void openOSPermission(perm.kind, 'revoke')}
            >
              {permBusy === perm.kind ? 'Opening…' : 'Turn off in Settings'}
            </button>
          {:else}
            <button
              type="button"
              class="md-btn md-btn-primary"
              disabled={permBusy === perm.kind}
              onclick={() => void openOSPermission(perm.kind, 'grant')}
            >
              {permBusy === perm.kind ? 'Opening…' : 'Turn on in Settings'}
            </button>
          {/if}
          <button
            type="button"
            class="guide-toggle"
            aria-expanded={expandedGuide === perm.kind}
            onclick={() => void toggleGuide(perm.kind)}
          >
            {expandedGuide === perm.kind ? 'Hide how-to' : 'How to'}
          </button>
        </div>

        {#if expandedGuide === perm.kind}
          {@const g = guideByKind[perm.kind]}
          {#if guideLoading === perm.kind}
            <p class="guide-loading">Loading guide…</p>
          {:else if g}
            <div class="guide">
              <p class="guide-title">{g.title}</p>
              <ol>
                {#each g.steps as step, i (i)}
                  <li>{step}</li>
                {/each}
              </ol>
            </div>
          {:else}
            <p class="guide-loading">Guide unavailable offline.</p>
          {/if}
        {/if}
      </li>
    {/each}
  </ul>

  {#if permActionNote}
    <p class="perm-action-note">{permActionNote}</p>
  {/if}
</section>

<style>
  .plate {
    border-radius: 12px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
    padding: 18px 18px 20px;
    box-shadow: none;
  }
  .plate-head { margin-bottom: 14px; }
  .perms-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
  }
  .recheck { flex: none; margin-top: 2px; }
  h2 {
    font-family: var(--md-font-display);
    font-size: 18px;
    letter-spacing: -0.03em;
    margin: 0 0 4px;
  }
  .hint {
    margin: 0;
    font-size: 13px;
    color: var(--md-ink-faint);
    line-height: 1.45;
    max-width: 48ch;
  }
  .perm-summary {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 9px 12px;
    border-radius: 10px;
    border: 1px solid var(--md-line);
    background: var(--md-stage);
    margin-bottom: 14px;
  }
  .perm-summary[data-state='ok'] {
    border-color: color-mix(in oklab, var(--md-live) 35%, transparent);
    background: color-mix(in oklab, var(--md-live) 7%, var(--md-surface));
  }
  .perm-summary[data-state='warn'] {
    border-color: color-mix(in oklab, var(--md-halt) 30%, transparent);
  }
  .perm-count {
    font-family: var(--md-font-mono);
    font-size: 14px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
  }
  .perm-sum-copy { font-size: 13px; color: var(--md-ink-mute); line-height: 1.4; }
  .perm-error { margin: 0 0 12px; font-size: 13px; color: var(--md-halt); }
  .perm-list { list-style: none; margin: 0; padding: 0; display: grid; gap: 10px; }
  .perm-row {
    border: 1px solid var(--md-line-strong);
    border-radius: 10px;
    background: var(--md-stage);
    padding: 12px 12px 11px;
    border-left: 2px solid var(--md-ink-faint);
  }
  .perm-row[data-state='granted'] { border-left-color: var(--md-live); }
  .perm-row[data-state='denied'] { border-left-color: var(--md-halt); }
  .perm-main {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 12px;
    align-items: start;
  }
  .perm-mark {
    width: 40px;
    height: 40px;
    border-radius: 12px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.04em;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 10%, var(--md-surface));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 20%, var(--md-line));
  }
  .perm-title-row {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 8px 10px;
    margin-bottom: 4px;
  }
  .perm-copy strong {
    font-family: var(--md-font-display);
    font-size: 16px;
    letter-spacing: -0.03em;
  }
  .perm-need {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.04em;
    color: var(--md-ink-faint);
  }
  .perm-need[data-weight='required'] { color: var(--md-cobalt); }
  .perm-copy p {
    margin: 0;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .perm-probe {
    margin-top: 8px !important;
    padding: 6px 8px;
    border-radius: 8px;
    background: color-mix(in oklab, var(--md-surface) 80%, transparent);
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-ink-faint);
    word-break: break-word;
  }
  .perm-status {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 8px;
    border-radius: 6px;
    border: 1px solid var(--md-line);
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    white-space: nowrap;
  }
  .perm-status i {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .perm-status[data-state='granted'] {
    color: var(--md-live);
    border-color: color-mix(in oklab, var(--md-live) 35%, transparent);
    background: color-mix(in oklab, var(--md-live) 8%, transparent);
  }
  .perm-status[data-state='granted'] i {
    background: var(--md-live);
    box-shadow: none;
  }
  .perm-status[data-state='denied'] {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 35%, transparent);
    background: color-mix(in oklab, var(--md-halt) 8%, transparent);
  }
  .perm-status[data-state='denied'] i { background: var(--md-halt); }
  .perm-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    margin-top: 12px;
    padding-left: 52px;
  }
  .guide-toggle {
    appearance: none;
    border: 0;
    background: transparent;
    color: var(--md-cobalt);
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    cursor: pointer;
    padding: 6px 4px;
  }
  .guide-toggle:hover { text-decoration: underline; }
  .guide-toggle:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
    border-radius: 6px;
  }
  .guide,
  .guide-loading {
    margin: 10px 0 0;
    padding: 12px 14px;
    margin-left: 52px;
    border-radius: 12px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
  }
  .guide-loading { font-size: 12px; color: var(--md-ink-faint); }
  .guide-title {
    margin: 0 0 8px;
    font-family: var(--md-font-display);
    font-size: 14px;
    letter-spacing: -0.02em;
  }
  .guide ol { margin: 0; padding-left: 18px; display: grid; gap: 6px; }
  .guide li { font-size: 13px; line-height: 1.45; color: var(--md-ink-mute); }
  .perm-action-note {
    margin: 14px 0 0;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
    padding: 10px 12px;
    border-radius: 12px;
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 22%, var(--md-line));
    background: color-mix(in oklab, var(--md-cobalt) 6%, var(--md-surface));
  }
  @media (max-width: 720px) {
    .perms-head { flex-direction: column; }
    .perm-main { grid-template-columns: auto 1fr; }
    .perm-status { grid-column: 2; justify-self: start; }
    .perm-actions,
    .guide,
    .guide-loading { padding-left: 0; margin-left: 0; }
  }
</style>
