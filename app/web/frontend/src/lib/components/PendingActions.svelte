<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import {
    pendingActions,
    pendingCount,
    refreshPendingActions,
    approvePending,
    denyPending,
    executePending,
    startPolling,
    stopPolling,
    type PendingAction,
  } from '../stores/pending.svelte'
  import { t } from '../i18n'

  // Auto-refresh every 5s while the panel is mounted.
  onMount(() => {
    startPolling(5000)
  })
  onDestroy(() => {
    stopPolling()
  })

  let error = $state<string | null>(null)
  let working = $state<Record<string, boolean>>({})

  function setWorking(id: string, v: boolean) {
    working = { ...working, [id]: v }
  }

  async function onApprove(a: PendingAction, autoRun = true) {
    setWorking(a.id, true)
    error = null
    const r = await approvePending(a.id, '', autoRun)
    setWorking(a.id, false)
    if (!r) {
      error = t('pending.error.approve', a.id)
    }
  }

  async function onDeny(a: PendingAction) {
    setWorking(a.id, true)
    error = null
    const r = await denyPending(a.id, '')
    setWorking(a.id, false)
    if (!r) {
      error = t('pending.error.deny', a.id)
    }
  }

  async function onExecute(a: PendingAction) {
    setWorking(a.id, true)
    error = null
    const r = await executePending(a.id)
    setWorking(a.id, false)
    if (!r) {
      error = t('pending.error.execute', a.id)
    }
  }

  async function onRefresh() {
    await refreshPendingActions()
  }

  // Group rows by status for the UI.
  const pending = $derived($pendingActions.filter((r) => r.status === 'pending'))
  const approved = $derived(
    $pendingActions.filter((r) => r.status === 'approved' || r.status === 'executed' || r.status === 'failed'),
  )
  const decided = $derived(
    $pendingActions.filter((r) => r.status === 'denied' || r.status === 'expired'),
  )

  function formatTime(iso: string): string {
    if (!iso) return ''
    try {
      return new Date(iso).toLocaleString()
    } catch {
      return iso
    }
  }

  function describePayload(a: PendingAction): string {
    if (a.payload.command) return '$ ' + a.payload.command
    if (a.payload.body) return 'type: ' + a.payload.body
    if (a.payload.target) return '→ ' + a.payload.target
    if (a.payload.key) return 'key: ' + a.payload.key
    if (a.payload.path) return 'path: ' + a.payload.path
    return t('pending.no_payload')
  }
</script>

<div class="pending-panel">
  <header>
    <div class="title-row">
      <h3>{t('pending.title')}</h3>
      <span class="badge count" class:has-pending={$pendingCount > 0}>{$pendingCount}</span>
    </div>
    <p class="muted">
      {t('pending.description')}
    </p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="actions-row">
    <button class="btn btn-ghost btn-sm" onclick={onRefresh}>{t('pending.refresh')}</button>
  </div>

  {#if pending.length === 0 && approved.length === 0 && decided.length === 0}
    <p class="muted">{t('pending.empty')}</p>
  {/if}

  {#if pending.length > 0}
    <section class="glass-card card urgent">
      <h4>{t('pending.awaiting', pending.length)}</h4>
      <ul class="row-list pending-rows">
        {#each pending as a (a.id)}
          <li>
            <div class="row-head">
              <span class="kind">{a.kind}</span>
              <span class="agent muted">{a.agent_name}</span>
              <span class="badge gate gate-{a.gate_decision}">{a.gate_decision}</span>
              <span class="expires muted">{t('pending.expires', formatTime(a.expires_at))}</span>
            </div>
            <div class="row-payload">{describePayload(a)}</div>
            {#if a.gate_reason}
              <div class="row-reason muted">{a.gate_reason}</div>
            {/if}
            <div class="row-actions">
              <button
                class="btn btn-primary btn-sm"
                disabled={working[a.id]}
                onclick={() => onApprove(a, true)}
              >
                {t('pending.approve_run')}
              </button>
              <button
                class="btn btn-secondary btn-sm"
                disabled={working[a.id]}
                onclick={() => onApprove(a, false)}
              >
                {t('pending.approve_only')}
              </button>
              <button
                class="btn btn-danger btn-sm"
                disabled={working[a.id]}
                onclick={() => onDeny(a)}
              >
                {t('pending.deny')}
              </button>
            </div>
          </li>
        {/each}
      </ul>
    </section>
  {/if}

  {#if approved.length > 0}
    <section class="glass-card card">
      <h4>{t('pending.approved', approved.length)}</h4>
      <ul class="row-list compact">
        {#each approved as a (a.id)}
          <li>
            <div class="row-head">
              <span class="kind">{a.kind}</span>
              <span class="agent muted">{a.agent_name}</span>
              <span class="badge status status-{a.status}">{a.status}</span>
              {#if a.duration_ms > 0}
                <span class="muted">{a.duration_ms}ms</span>
              {/if}
            </div>
            {#if a.result}
              <pre class="result">{a.result}</pre>
            {/if}
            {#if a.execution_error}
              <pre class="error-output">{a.execution_error}</pre>
            {/if}
            {#if a.status === 'approved'}
              <div class="row-actions">
                <button
                  class="btn btn-primary btn-sm"
                  disabled={working[a.id]}
                  onclick={() => onExecute(a)}
                >
                  {t('pending.run_now')}
                </button>
              </div>
            {/if}
          </li>
        {/each}
      </ul>
    </section>
  {/if}

  {#if decided.length > 0}
    <section class="glass-card card">
      <h4>{t('pending.history', decided.length)}</h4>
      <ul class="row-list compact">
        {#each decided as a (a.id)}
          <li>
            <div class="row-head">
              <span class="kind">{a.kind}</span>
              <span class="agent muted">{a.agent_name}</span>
              <span class="badge status status-{a.status}">{a.status}</span>
              <span class="muted">{formatTime(a.decided_at ?? a.created_at)}</span>
            </div>
          </li>
        {/each}
      </ul>
    </section>
  {/if}
</div>

<style>
  .pending-panel {
    padding: var(--space-4);
    overflow-y: auto;
    height: 100%;
  }
  .pending-panel header h3 {
    font-size: var(--size-xl);
    font-weight: var(--weight-semibold);
    margin: 0;
  }
  .title-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .count {
    min-width: 22px;
    height: 22px;
    padding: 0 var(--space-2);
    justify-content: center;
    background: var(--color-bg-active);
    color: var(--color-text-muted);
    text-transform: none;
    letter-spacing: 0;
  }
  .count.has-pending {
    background: var(--color-accent);
    color: #fff;
    border-color: transparent;
  }
  .error {
    color: var(--color-error);
    margin: var(--space-2) 0;
    font-size: var(--size-sm);
  }
  .actions-row {
    display: flex;
    justify-content: flex-end;
    margin: var(--space-3) 0;
  }
  .card {
    padding: var(--space-4);
    margin: var(--space-3) 0;
  }
  .card.urgent {
    animation: pulse-glow 2.6s ease-in-out infinite;
  }
  .card h4 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    margin: 0 0 var(--space-3) 0;
  }
  .row-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .row-list li {
    padding: var(--space-3);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-md);
    background: rgba(0, 0, 0, 0.18);
  }
  .pending-rows li {
    border-left: 2px solid var(--color-warn);
  }
  .row-list.compact li {
    padding: var(--space-2) var(--space-3);
  }
  .row-head {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
    margin-bottom: var(--space-1);
  }
  .kind {
    font-family: var(--font-mono);
    font-weight: var(--weight-semibold);
    font-size: var(--size-sm);
  }
  .agent { font-size: var(--size-xs); }
  .expires { font-size: var(--size-xs); }
  .row-head :global(.gate-allow) {
    color: var(--color-success);
    border-color: var(--color-success);
    background: var(--color-success-soft);
  }
  .row-head :global(.gate-deny) {
    color: var(--color-error);
    border-color: var(--color-error);
    background: var(--color-error-soft);
  }
  .row-head :global(.gate-require_consent) {
    color: var(--color-warn);
    border-color: var(--color-warn);
    background: var(--color-warn-soft);
  }
  .row-head :global(.gate-require_presence_and_consent) {
    color: #fb923c;
    border-color: #fb923c;
    background: rgba(251, 146, 60, 0.12);
  }
  .row-head :global(.status-executed) {
    color: var(--color-success);
    border-color: var(--color-success);
    background: var(--color-success-soft);
  }
  .row-head :global(.status-failed) {
    color: var(--color-error);
    border-color: var(--color-error);
    background: var(--color-error-soft);
  }
  .row-head :global(.status-approved) {
    color: var(--color-warn);
    border-color: var(--color-warn);
    background: var(--color-warn-soft);
  }
  .row-head :global(.status-expired),
  .row-head :global(.status-denied) {
    color: var(--color-text-muted);
  }
  .row-payload {
    font-family: var(--font-mono);
    font-size: var(--size-sm);
    padding: var(--space-1) 0;
    color: var(--color-text);
  }
  .row-reason {
    font-size: var(--size-xs);
    margin-top: var(--space-1);
  }
  .row-actions {
    display: flex;
    gap: var(--space-2);
    margin-top: var(--space-3);
  }
  pre {
    background: rgba(0, 0, 0, 0.3);
    padding: var(--space-2);
    border-radius: var(--radius-md);
    overflow: auto;
    max-height: 200px;
    font-size: var(--size-xs);
    line-height: var(--leading-relaxed);
    margin: var(--space-2) 0 0 0;
  }
  pre.error-output { color: var(--color-error); }
</style>
