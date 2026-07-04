<script lang="ts">
  // PendingActions — list of sub-agent actions awaiting user decision.
  //
  // Each row shows: timestamp, action kind, target, gate decision,
  // and pending status badge. Click a row to expand for details.
  // Groups rows into Awaiting (pending), Approved (approved/
  // executed/failed), and History (denied/expired).
  import { onMount, onDestroy } from 'svelte'
  import Card from './ui/Card.svelte'
  import Button from './ui/Button.svelte'
  import Badge from './ui/Badge.svelte'
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
  let expanded = $state<Set<string>>(new Set())

  function setWorking(id: string, v: boolean): void {
    working = { ...working, [id]: v }
  }

  function toggleExpand(id: string): void {
    const next = new Set(expanded)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    expanded = next
  }

  async function onApprove(a: PendingAction, autoRun = true): Promise<void> {
    setWorking(a.id, true)
    error = null
    const r = await approvePending(a.id, '', autoRun)
    setWorking(a.id, false)
    if (!r) {
      error = t('pending.error.approve', a.id)
    }
  }

  async function onDeny(a: PendingAction): Promise<void> {
    setWorking(a.id, true)
    error = null
    const r = await denyPending(a.id, '')
    setWorking(a.id, false)
    if (!r) {
      error = t('pending.error.deny', a.id)
    }
  }

  async function onExecute(a: PendingAction): Promise<void> {
    setWorking(a.id, true)
    error = null
    const r = await executePending(a.id)
    setWorking(a.id, false)
    if (!r) {
      error = t('pending.error.execute', a.id)
    }
  }

  async function onRefresh(): Promise<void> {
    await refreshPendingActions()
  }

  // Group rows by status for the UI.
  const pending = $derived($pendingActions.filter((r) => r.status === 'pending'))
  const approved = $derived(
    $pendingActions.filter(
      (r) => r.status === 'approved' || r.status === 'executed' || r.status === 'failed',
    ),
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

  function gateTone(decision: string): 'success' | 'error' | 'warn' | 'neutral' {
    switch (decision) {
      case 'allow':
        return 'success'
      case 'deny':
        return 'error'
      case 'require_consent':
        return 'warn'
      default:
        return 'neutral'
    }
  }

  function statusTone(status: string): 'success' | 'error' | 'warn' | 'neutral' {
    switch (status) {
      case 'executed':
        return 'success'
      case 'failed':
      case 'denied':
      case 'expired':
        return 'error'
      case 'approved':
        return 'warn'
      default:
        return 'neutral'
    }
  }
</script>

<div class="pending-panel">
  <header>
    <div class="title-row">
      <h3>{t('pending.title')}</h3>
      <Badge tone={$pendingCount > 0 ? 'accent' : 'neutral'} size="sm">
        {$pendingCount}
      </Badge>
    </div>
    <p class="muted">{t('pending.description')}</p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="actions-row">
    <Button variant="ghost" size="sm" onclick={onRefresh}>
      {t('pending.refresh')}
    </Button>
  </div>

  {#if pending.length === 0 && approved.length === 0 && decided.length === 0}
    <p class="muted">{t('pending.empty')}</p>
  {/if}

  {#if pending.length > 0}
    <Card elevation={1} padding="md">
      <h4>{t('pending.awaiting', pending.length)}</h4>
      <ul class="row-list pending-rows">
        {#each pending as a (a.id)}
          <li class:expanded={expanded.has(a.id)}>
            <button
              type="button"
              class="row-head press"
              onclick={() => toggleExpand(a.id)}
              aria-expanded={expanded.has(a.id)}
            >
              <span class="kind">{a.kind}</span>
              <span class="agent muted">{a.agent_name}</span>
              <Badge tone={gateTone(a.gate_decision)} size="xs">
                {a.gate_decision}
              </Badge>
              <span class="expires muted">
                {t('pending.expires', formatTime(a.expires_at))}
              </span>
              <span class="chev" aria-hidden="true">
                {expanded.has(a.id) ? '▾' : '▸'}
              </span>
            </button>

            {#if expanded.has(a.id)}
              <div class="row-detail">
                <div class="row-payload">{describePayload(a)}</div>
                {#if a.gate_reason}
                  <div class="row-reason muted">{a.gate_reason}</div>
                {/if}
              </div>
            {/if}

            <div class="row-actions">
              <Button
                variant="primary"
                size="sm"
                disabled={working[a.id]}
                loading={working[a.id]}
                onclick={() => onApprove(a, true)}
              >
                {t('pending.approve_run')}
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={working[a.id]}
                onclick={() => onApprove(a, false)}
              >
                {t('pending.approve_only')}
              </Button>
              <Button
                variant="danger"
                size="sm"
                disabled={working[a.id]}
                onclick={() => onDeny(a)}
              >
                {t('pending.deny')}
              </Button>
            </div>
          </li>
        {/each}
      </ul>
    </Card>
  {/if}

  {#if approved.length > 0}
    <Card elevation={1} padding="md">
      <h4>{t('pending.approved', approved.length)}</h4>
      <ul class="row-list compact">
        {#each approved as a (a.id)}
          <li class:expanded={expanded.has(a.id)}>
            <button
              type="button"
              class="row-head press"
              onclick={() => toggleExpand(a.id)}
              aria-expanded={expanded.has(a.id)}
            >
              <span class="kind">{a.kind}</span>
              <span class="agent muted">{a.agent_name}</span>
              <Badge tone={statusTone(a.status)} size="xs">
                {a.status}
              </Badge>
              {#if a.duration_ms > 0}
                <span class="muted">{a.duration_ms}ms</span>
              {/if}
              <span class="chev" aria-hidden="true">
                {expanded.has(a.id) ? '▾' : '▸'}
              </span>
            </button>

            {#if expanded.has(a.id)}
              <div class="row-detail">
                {#if a.result}
                  <pre class="result">{a.result}</pre>
                {/if}
                {#if a.execution_error}
                  <pre class="error-output">{a.execution_error}</pre>
                {/if}
              </div>
            {/if}

            {#if a.status === 'approved'}
              <div class="row-actions">
                <Button
                  variant="primary"
                  size="sm"
                  disabled={working[a.id]}
                  loading={working[a.id]}
                  onclick={() => onExecute(a)}
                >
                  {t('pending.run_now')}
                </Button>
              </div>
            {/if}
          </li>
        {/each}
      </ul>
    </Card>
  {/if}

  {#if decided.length > 0}
    <Card elevation={1} padding="md">
      <h4>{t('pending.history', decided.length)}</h4>
      <ul class="row-list compact">
        {#each decided as a (a.id)}
          <li class:expanded={expanded.has(a.id)}>
            <button
              type="button"
              class="row-head press"
              onclick={() => toggleExpand(a.id)}
              aria-expanded={expanded.has(a.id)}
            >
              <span class="kind">{a.kind}</span>
              <span class="agent muted">{a.agent_name}</span>
              <Badge tone={statusTone(a.status)} size="xs">
                {a.status}
              </Badge>
              <span class="muted">{formatTime(a.decided_at ?? a.created_at)}</span>
              <span class="chev" aria-hidden="true">
                {expanded.has(a.id) ? '▾' : '▸'}
              </span>
            </button>
          </li>
        {/each}
      </ul>
    </Card>
  {/if}
</div>

<style>
  .pending-panel {
    padding: var(--space-4);
    overflow-y: auto;
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  header h3 {
    font-size: var(--size-xl);
    font-weight: var(--weight-semibold);
    margin: 0;
    color: var(--text);
  }
  .title-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .muted {
    color: var(--text-muted);
    font-size: var(--size-sm);
    margin: 0;
  }
  .error {
    color: var(--error);
    margin: 0;
    font-size: var(--size-sm);
  }
  .actions-row {
    display: flex;
    justify-content: flex-end;
  }

  h4 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    margin: 0 0 var(--space-3) 0;
    color: var(--text);
  }

  .row-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .row-list li {
    padding: var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface-2);
    transition:
      border-color var(--transition-fast) ease,
      background-color var(--transition-fast) ease;
  }
  .row-list li:hover {
    border-color: var(--border-strong);
    background: var(--surface-3);
  }
  .pending-rows li {
    border-left: 2px solid var(--warn);
  }

  .row-head {
    appearance: none;
    background: transparent;
    border: none;
    width: 100%;
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
    text-align: left;
    cursor: pointer;
    color: var(--text);
    font-family: inherit;
    padding: 0;
    margin-bottom: var(--space-2);
  }
  .row-head:focus-visible {
    outline: 2px solid var(--border-focus);
    outline-offset: 2px;
    border-radius: var(--radius-sm);
  }
  .kind {
    font-family: var(--font-mono);
    font-weight: var(--weight-semibold);
    font-size: var(--size-sm);
    color: var(--text);
  }
  .agent {
    font-size: var(--size-xs);
    color: var(--text-muted);
  }
  .expires {
    font-size: var(--size-xs);
    color: var(--text-muted);
  }
  .chev {
    margin-left: auto;
    color: var(--text-faint);
    font-size: var(--size-xs);
  }

  .row-detail {
    padding: var(--space-2) 0;
    border-top: 1px dashed var(--border);
    margin-top: var(--space-1);
  }
  .row-payload {
    font-family: var(--font-mono);
    font-size: var(--size-sm);
    padding: var(--space-1) 0;
    color: var(--text);
  }
  .row-reason {
    font-size: var(--size-xs);
    margin-top: var(--space-1);
    color: var(--text-muted);
  }

  .row-actions {
    display: flex;
    gap: var(--space-2);
    margin-top: var(--space-3);
  }

  pre {
    background: var(--surface-3);
    padding: var(--space-2);
    border-radius: var(--radius-md);
    overflow: auto;
    max-height: 200px;
    font-size: var(--size-xs);
    line-height: var(--leading-relaxed);
    margin: var(--space-2) 0 0 0;
    color: var(--text);
  }
  pre.error-output {
    color: var(--error);
  }
</style>