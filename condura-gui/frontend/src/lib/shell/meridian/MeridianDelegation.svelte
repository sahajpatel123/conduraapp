<script lang="ts">
  /**
   * Agents — control room: spawn sub-agents + Gatekeeper consent queue.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { ipc } from '../../ipc/client'
  import { humanizeIpcError } from '../../ipc/errors'
  import type { DelegateAgentInfo, DelegateRunningSpawn, DelegateSpawnResult } from '../../ipc/types'
  import {
    pendingActions,
    pendingCount,
    pendingRefreshError,
    refreshPendingActions,
    approvePending,
    denyPending,
    executePending,
    startPolling,
    stopPolling,
  } from '../../stores/pending.svelte'

  let busyId = $state<string | null>(null)
  let busyKind = $state<'approve' | 'run' | 'deny' | null>(null)

  let agents = $state<DelegateAgentInfo[]>([])
  let agentsLoading = $state(false)
  let agentsError = $state('')
  let selectedAgent = $state('')
  let task = $state('')
  let spawning = $state(false)
  let spawnNote = $state('')
  let lastSpawn = $state<DelegateSpawnResult | null>(null)
  let running = $state<DelegateRunningSpawn[]>([])
  let cancelBusy = $state<string | null>(null)

  onMount(() => {
    void refreshPendingActions()
    startPolling(4000)
    void loadAgents()
    void refreshRunning()
    const t = setInterval(() => void refreshRunning(), 5000)
    return () => {
      stopPolling()
      clearInterval(t)
    }
  })

  const rows = $derived($pendingActions)
  const count = $derived($pendingCount)
  const refreshErr = $derived($pendingRefreshError)
  const waiting = $derived(rows.filter((r) => r.status === 'pending'))
  const settled = $derived(rows.filter((r) => r.status !== 'pending').slice(0, 8))
  const canSpawn = $derived(!!selectedAgent && !!task.trim() && !spawning)

  function initial(name: string): string {
    const t = name.trim()
    return t ? t[0]!.toUpperCase() : '?'
  }

  function goAsk(): void {
    window.location.hash = '#/'
  }

  function goAudit(): void {
    window.location.hash = '#/audit'
  }

  function formatWhen(v?: string): string {
    if (!v) return ''
    try {
      const d = new Date(v)
      return Number.isNaN(d.getTime()) ? v : d.toLocaleString()
    } catch {
      return v
    }
  }

  async function loadAgents(): Promise<void> {
    agentsLoading = true
    agentsError = ''
    try {
      const res = await ipc.delegateListAgents()
      agents = res?.agents ?? []
      if (!selectedAgent && agents[0]) selectedAgent = agents[0].name
    } catch (e) {
      agents = []
      agentsError = humanizeIpcError(e, 'Daemon offline — agent list unavailable')
    } finally {
      agentsLoading = false
    }
  }

  async function refreshRunning(): Promise<void> {
    try {
      const res = await ipc.delegateListSpawns()
      running = res?.running ?? []
      if (spawnNote.startsWith('Could not list running')) spawnNote = ''
    } catch (e) {
      // Keep last good list — wiping to [] looks like "quiet" when RPC is down.
      spawnNote = humanizeIpcError(
        e,
        'Could not list running agents — daemon offline'
      )
    }
  }

  async function spawn(): Promise<void> {
    if (!canSpawn) return
    spawning = true
    spawnNote = ''
    lastSpawn = null
    try {
      const result = await ipc.delegateSpawn({
        agent_name: selectedAgent,
        task: task.trim(),
        depth: 0,
        budget: 0,
      })
      lastSpawn = result
      task = ''
      const pendingN = result.pending_action_ids?.length ?? result.pending_actions?.length ?? 0
      spawnNote =
        pendingN > 0
          ? `Spawned ${result.agent_name || selectedAgent} · ${pendingN} action(s) need your consent below.`
          : `Spawned ${result.agent_name || selectedAgent}${result.spawn_id ? ` · ${result.spawn_id.slice(0, 10)}` : ''}.`
      await Promise.all([refreshPendingActions(), refreshRunning()])
    } catch (e) {
      spawnNote = humanizeIpcError(e)
    } finally {
      spawning = false
    }
  }

  async function cancelSpawn(id: string): Promise<void> {
    if (!id || cancelBusy) return
    cancelBusy = id
    try {
      await ipc.delegateCancel(id)
      await refreshRunning()
    } catch (e) {
      spawnNote = humanizeIpcError(e)
    } finally {
      cancelBusy = null
    }
  }

  async function act(
    id: string,
    kind: 'approve' | 'run' | 'deny'
  ): Promise<void> {
    if (busyId) return
    busyId = id
    busyKind = kind
    try {
      if (kind === 'approve') await approvePending(id)
      else if (kind === 'run') await executePending(id)
      else await denyPending(id)
      if (spawnNote.includes('Could not') || spawnNote.includes('Daemon offline')) spawnNote = ''
    } catch (e) {
      // Toasts fire from pending store; also pin failure on the board.
      spawnNote = humanizeIpcError(e)
    } finally {
      busyId = null
      busyKind = null
    }
  }
</script>

<MeridianPage
  kicker="Control room"
  title="Agents"
  lead="Spawn a sub-agent, then approve gated actions — the Gatekeeper never decides alone."
>
  {#snippet actions()}
    <button
      type="button"
      class="md-btn md-btn-ghost"
      onclick={() => {
        void refreshPendingActions()
        void loadAgents()
        void refreshRunning()
      }}
    >
      Refresh
    </button>
    <button type="button" class="md-btn md-btn-ghost" onclick={goAudit}>Open Audit</button>
  {/snippet}

  <div class="board md-stagger">
    <p class="contract" class:hot={count > 0 || running.length > 0}>
      <span class="live-dot" aria-hidden="true"></span>
      {#if count > 0}
        {count} sheet{count === 1 ? '' : 's'} awaiting you — live SSE + light poll.
      {:else if running.length > 0}
        {running.length} spawn{running.length === 1 ? '' : 's'} in flight.
      {:else}
        Queue quiet. Spawn below or wait for Ask to delegate.
      {/if}
    </p>

    <section class="spawn-plate" aria-labelledby="spawn-title">
      <p class="cite">01 · launch</p>
      <h2 id="spawn-title">Spawn a sub-agent</h2>
      <p class="spawn-lead">
        CLI agents on this machine (Claude Code, Codex, Ollama, …). Actions they propose still need your consent.
      </p>

      {#if agentsLoading && agents.length === 0}
        <p class="muted">Discovering agents…</p>
      {:else if agentsError}
        <p class="err">{agentsError}</p>
      {:else if agents.length === 0}
        <p class="muted">
          No agents registered. Install a supported CLI on PATH, then refresh.
        </p>
      {:else}
        <div class="spawn-form">
          <label class="field">
            <span>Agent</span>
            <select bind:value={selectedAgent} aria-label="Sub-agent" disabled={spawning}>
              {#each agents as a (a.name)}
                <option value={a.name}>
                  {a.name}{a.binary ? ` · ${a.binary}` : ''}
                </option>
              {/each}
            </select>
          </label>
          <label class="field grow">
            <span>Task</span>
            <textarea
              bind:value={task}
              rows="3"
              placeholder="What should the sub-agent do?"
              disabled={spawning}
            ></textarea>
          </label>
          <div class="spawn-actions">
            <button
              type="button"
              class="md-btn md-btn-primary"
              disabled={!canSpawn}
              onclick={() => void spawn()}
            >
              {spawning ? 'Spawning…' : 'Spawn'}
            </button>
            <button type="button" class="md-btn md-btn-ghost" onclick={goAsk}>
              Or ask Condura
            </button>
          </div>
        </div>
      {/if}

      {#if spawnNote}
        <p class="spawn-note" class:bad={/error|failed|Could not|IPC|denied|unavailable/i.test(spawnNote)}>
          {spawnNote}
        </p>
      {/if}

      {#if lastSpawn?.output}
        <details class="spawn-out">
          <summary>Last spawn output</summary>
          <pre>{lastSpawn.output.slice(0, 4000)}</pre>
        </details>
      {/if}

      {#if running.length > 0}
        <div class="running">
          <p class="cite">in flight · {running.length}</p>
          <ul>
            {#each running as r (r.spawn_id)}
              <li>
                <span class="run-id">{r.spawn_id}</span>
                <span class="run-st">{r.state || 'running'}</span>
                <button
                  type="button"
                  class="md-btn md-btn-ghost sm"
                  disabled={cancelBusy === r.spawn_id}
                  onclick={() => void cancelSpawn(r.spawn_id)}
                >
                  {cancelBusy === r.spawn_id ? '…' : 'Cancel'}
                </button>
              </li>
            {/each}
          </ul>
        </div>
      {/if}
    </section>

    <section class="pulse" class:hot={count > 0} aria-live="polite">
      <div class="pulse-ring" aria-hidden="true">
        <span class="pulse-core">{count}</span>
      </div>
      <div class="pulse-copy">
        <p class="cite">{refreshErr ? 'offline' : count > 0 ? 'awaiting you' : 'quiet'}</p>
        <h2>
          {#if refreshErr}
            Queue unreachable
          {:else if count > 0}
            {count} pending consent
          {:else}
            Nothing gated
          {/if}
        </h2>
        <p>
          {#if refreshErr}
            Could not refresh the pending queue — showing last known rows. {refreshErr}
          {:else if count > 0}
            Each sheet is a gated action. Approve opens the door; Deny seals it; Run executes an approved row.
          {:else}
            When a sub-agent needs you, this room lights up.
          {/if}
        </p>
        <ol class="legend" aria-label="Decision meanings">
          <li><span class="lg allow">Approve</span> open the door</li>
          <li><span class="lg run">Run</span> execute approved</li>
          <li><span class="lg deny">Deny</span> seal shut</li>
        </ol>
      </div>
    </section>

    {#if waiting.length > 0}
      <section class="queue">
        <p class="cite">consent sheets · {waiting.length}</p>
        <div class="list">
          {#each waiting as row (row.id)}
            <article class="sheet">
              <header>
                <div class="mono" aria-hidden="true">{initial(row.agent_name || row.kind || row.id)}</div>
                <div class="copy">
                  <strong>{row.agent_name || row.kind || row.id}</strong>
                  <span class="meta">
                    {row.kind || 'action'} · {row.status}
                    {#if row.expires_at} · expires {formatWhen(row.expires_at)}{/if}
                  </span>
                </div>
              </header>
              <p class="reason">{row.gate_reason || 'Gatekeeper requires your decision.'}</p>
              {#if row.gate_decision}
                <p class="gate-chip">gate · {row.gate_decision}</p>
              {/if}
              {#if row.payload?.command || row.payload?.path || row.payload?.target}
                <p class="payload">
                  <span>cite</span>
                  {row.payload?.command || row.payload?.path || row.payload?.target}
                </p>
              {/if}
              <footer>
                <button
                  type="button"
                  class="md-btn md-btn-primary"
                  disabled={busyId === row.id}
                  onclick={() => void act(row.id, 'approve')}
                >
                  {busyId === row.id && busyKind === 'approve' ? 'Approving…' : 'Approve'}
                </button>
                <button
                  type="button"
                  class="md-btn md-btn-ghost"
                  disabled={busyId === row.id}
                  onclick={() => void act(row.id, 'run')}
                >
                  {busyId === row.id && busyKind === 'run' ? 'Running…' : 'Run'}
                </button>
                <button
                  type="button"
                  class="md-btn md-btn-danger"
                  disabled={busyId === row.id}
                  onclick={() => void act(row.id, 'deny')}
                >
                  {busyId === row.id && busyKind === 'deny' ? 'Denying…' : 'Deny'}
                </button>
              </footer>
            </article>
          {/each}
        </div>
      </section>
    {/if}

    {#if settled.length > 0}
      <section class="settled">
        <p class="cite">recent decisions</p>
        <ul>
          {#each settled as row (row.id)}
            <li data-status={row.status}>
              <span class="dot" aria-hidden="true"></span>
              <strong>{row.agent_name || row.kind || row.id}</strong>
              <span class="meta">{row.status}</span>
            </li>
          {/each}
        </ul>
      </section>
    {/if}
  </div>
</MeridianPage>

<style>
  .board {
    display: grid;
    gap: 18px;
  }
  .contract {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    margin: 0;
    padding: 12px 14px;
    border-radius: 14px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 80%, transparent);
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .contract.hot {
    border-color: color-mix(in oklab, var(--md-cobalt) 32%, transparent);
    background: color-mix(in oklab, var(--md-cobalt) 6%, var(--md-surface));
  }
  .live-dot {
    width: 8px;
    height: 8px;
    margin-top: 5px;
    flex: none;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .contract.hot .live-dot {
    background: var(--md-cobalt);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-cobalt) 16%, transparent);
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 12px;
  }
  .pulse {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 22px;
    align-items: center;
    padding: 22px 24px;
    border-radius: 22px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
  }
  .pulse.hot {
    border-color: color-mix(in oklab, var(--md-cobalt) 35%, transparent);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--md-cobalt) 10%, transparent);
  }
  .pulse-ring {
    width: 88px;
    height: 88px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    border: 2px solid var(--md-line-strong);
    background: var(--md-stage);
  }
  .pulse.hot .pulse-ring {
    border-color: color-mix(in oklab, var(--md-cobalt) 45%, transparent);
    animation: md-pulse-soft 2.4s var(--md-ease) infinite;
  }
  .pulse-core {
    font-family: var(--md-font-display);
    font-size: 32px;
    font-weight: 700;
    letter-spacing: -0.05em;
    color: var(--md-ink-mute);
  }
  .pulse.hot .pulse-core {
    color: var(--md-cobalt);
  }
  .pulse-copy h2 {
    font-family: var(--md-font-display);
    font-size: 24px;
    letter-spacing: -0.04em;
    margin: 0 0 8px;
  }
  .pulse-copy p {
    margin: 0 0 12px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
    max-width: 48ch;
  }
  .legend {
    display: flex;
    flex-wrap: wrap;
    gap: 10px 14px;
    margin: 0 0 14px;
    padding: 0;
    list-style: none;
    font-size: 12px;
    color: var(--md-ink-faint);
  }
  .lg {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    margin-right: 4px;
  }
  .lg.allow {
    color: var(--md-cobalt);
  }
  .lg.run {
    color: var(--md-live);
  }
  .lg.deny {
    color: var(--md-halt);
  }

  .spawn-plate {
    border-radius: 22px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 90%, transparent);
    padding: 22px 24px 20px;
  }
  .spawn-plate h2 {
    font-family: var(--md-font-display);
    font-size: 22px;
    letter-spacing: -0.04em;
    margin: 0 0 8px;
  }
  .spawn-lead {
    margin: 0 0 16px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
    max-width: 52ch;
  }
  .spawn-form {
    display: grid;
    gap: 14px;
  }
  .field {
    display: grid;
    gap: 6px;
  }
  .field span {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .field select,
  .field textarea {
    width: 100%;
    border-radius: 12px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-stage);
    color: var(--md-ink);
    font: inherit;
    font-size: 14px;
    padding: 10px 12px;
  }
  .field textarea {
    resize: vertical;
    min-height: 72px;
    line-height: 1.45;
  }
  .field select:focus-visible,
  .field textarea:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .spawn-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .muted {
    margin: 0;
    font-size: 13px;
    color: var(--md-ink-faint);
  }
  .err {
    margin: 0;
    font-size: 13px;
    color: var(--md-halt);
  }
  .spawn-note {
    margin: 14px 0 0;
    padding: 10px 12px;
    border-radius: 12px;
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 22%, var(--md-line));
    background: color-mix(in oklab, var(--md-cobalt) 6%, transparent);
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .spawn-note.bad {
    border-color: color-mix(in oklab, var(--md-halt) 30%, transparent);
    background: color-mix(in oklab, var(--md-halt) 6%, transparent);
    color: var(--md-halt);
  }
  .spawn-out {
    margin-top: 12px;
    font-size: 13px;
  }
  .spawn-out summary {
    cursor: pointer;
    color: var(--md-cobalt);
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
  }
  .spawn-out pre {
    margin: 8px 0 0;
    padding: 12px;
    border-radius: 12px;
    background: var(--md-stage);
    border: 1px solid var(--md-line);
    overflow: auto;
    max-height: 200px;
    font-size: 12px;
    white-space: pre-wrap;
    word-break: break-word;
  }
  .running {
    margin-top: 16px;
  }
  .running ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 8px;
  }
  .running li {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    border-radius: 12px;
    border: 1px solid var(--md-line);
    background: var(--md-stage);
  }
  .run-id {
    font-family: var(--md-font-mono);
    font-size: 12px;
    flex: 1;
    min-width: 0;
    word-break: break-all;
  }
  .run-st {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--md-live);
  }
  :global(.md-btn.sm) {
    padding: 6px 10px;
    font-size: 12px;
  }

  .list {
    display: grid;
    gap: 12px;
  }
  .sheet {
    border-radius: 20px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 18px 18px 16px;
    box-shadow: var(--md-shadow);
  }
  .sheet header {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 12px;
  }
  .mono {
    width: 44px;
    height: 44px;
    border-radius: 14px;
    display: grid;
    place-items: center;
    flex: none;
    font-family: var(--md-font-display);
    font-size: 18px;
    font-weight: 700;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
  }
  .copy {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .sheet strong {
    font-family: var(--md-font-display);
    font-size: 18px;
    letter-spacing: -0.03em;
  }
  .meta {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .reason {
    margin: 0 0 8px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-soft);
  }
  .gate-chip {
    margin: 0 0 10px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-cobalt);
  }
  .payload {
    margin: 0 0 16px;
    padding: 10px 12px;
    border-radius: 12px;
    background: var(--md-stage);
    border: 1px solid var(--md-line);
    font-family: var(--md-font-mono);
    font-size: 12px;
    color: var(--md-ink-mute);
    word-break: break-all;
  }
  .payload span {
    display: block;
    font-size: 9px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin-bottom: 4px;
  }
  footer {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .settled ul {
    margin: 0;
    padding: 0;
    display: grid;
    gap: 8px;
  }
  .settled li {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    border-radius: 12px;
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
    border: 1px solid var(--md-line);
  }
  .settled .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .settled li[data-status='approved'] .dot,
  .settled li[data-status='executed'] .dot {
    background: var(--md-live);
  }
  .settled li[data-status='denied'] .dot,
  .settled li[data-status='failed'] .dot {
    background: var(--md-halt);
  }
  .settled strong {
    font-size: 13px;
    font-weight: 600;
  }

  @keyframes md-pulse-soft {
    0%,
    100% {
      box-shadow: 0 0 0 0 color-mix(in oklab, var(--md-cobalt) 0%, transparent);
    }
    50% {
      box-shadow: 0 0 0 10px color-mix(in oklab, var(--md-cobalt) 12%, transparent);
    }
  }

  @media (max-width: 640px) {
    .pulse {
      grid-template-columns: 1fr;
      justify-items: start;
    }
    footer :global(.md-btn) {
      flex: 1;
      justify-content: center;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .pulse.hot .pulse-ring {
      animation: none;
    }
  }
</style>
