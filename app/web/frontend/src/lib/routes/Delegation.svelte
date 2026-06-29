<script lang="ts">
  // Delegation — list of CLI sub-agent backends (Claude Code, Codex,
  // Antigravity, OpenCode, Kilo, Hermes, Gemini, Ollama) with install
  // status, default model, and enable toggle. Spawn a sub-agent from
  // the top panel. Cancel running spawns from the running list.
  import { ipc } from '../ipc/client'
  import type { AppConfig } from '../ipc/types'
  import { onMount, onDestroy } from 'svelte'
  import { notifications } from '../stores/notifications.svelte'
  import PendingActions from '../components/PendingActions.svelte'
  import { Button, Card, Switch, Input, Textarea, Select, Badge } from '../components/ui'

  interface Agent {
    name: string
    description: string
    binary: string
    installed?: boolean
    enabled?: boolean
    default_model?: string
    available_models?: string[]
  }

  interface SpawnResult {
    spawn_id: string
    agent_name: string
    state: string
    output: string
    tokens: number
    cost: number
    started: string
    finished: string
    pending_actions?: unknown[]
  }

  interface Running {
    spawn_id: string
    agent_name: string
    state: string
    started: string
  }

  let agents = $state<Agent[]>([])
  let loading = $state(false)
  let error = $state<string | null>(null)
  let selectedAgent = $state('')
  let taskInput = $state('')
  let modelInput = $state('')
  let budgetInput = $state(1.0)
  let spawning = $state(false)
  let lastSpawn = $state<SpawnResult | null>(null)
  let running = $state<Running[]>([])
  let pollTimer: ReturnType<typeof setInterval> | null = null

  // Per-agent enabled / default model state. We cache locally so
  // toggling the switch feels instant; the daemon is updated
  // optimistically via ipc.config.update on each change.
  let enabledMap = $state<Record<string, boolean>>({})
  let modelMap = $state<Record<string, string>>({})

  async function refresh(): Promise<void> {
    loading = true
    error = null
    try {
      const resp = await ipc.call<{ agents: Agent[] }>('delegate.list_agents', {})
      const list = resp?.agents ?? []
      agents = list
      if (list.length > 0 && !selectedAgent) selectedAgent = list[0].name
      // Initialise local cache.
      for (const a of list) {
        if (enabledMap[a.name] === undefined) enabledMap[a.name] = a.enabled !== false
        if (!modelMap[a.name]) modelMap[a.name] = a.default_model || a.available_models?.[0] || ''
      }
      enabledMap = { ...enabledMap }
      modelMap = { ...modelMap }
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  }

  async function refreshRunning(): Promise<void> {
    try {
      const r = await ipc.call<{ running: Running[] }>('delegate.list_spawns', {})
      running = r?.running ?? []
    } catch {
      // non-fatal
    }
  }

  async function spawn(): Promise<void> {
    if (!selectedAgent || !taskInput.trim()) return
    spawning = true
    error = null
    lastSpawn = null
    try {
      const result = await ipc.call<SpawnResult>('delegate.spawn', {
        agent_name: selectedAgent,
        task: taskInput,
        model: modelInput,
        depth: 0,
        budget: budgetInput,
      })
      lastSpawn = result
      taskInput = ''
      notifications.push({
        kind: 'success',
        title: 'Sub-agent spawned',
        message: `${result.agent_name} (${result.spawn_id.slice(0, 8)})`,
      })
      await refreshRunning()
    } catch (e) {
      error = String(e)
      notifications.push({ kind: 'error', title: 'Spawn failed', message: String(e) })
    } finally {
      spawning = false
    }
  }

  async function cancel(spawnId: string): Promise<void> {
    try {
      await ipc.call('delegate.cancel', { spawn_id: spawnId })
      error = null
      notifications.push({ kind: 'info', title: 'Cancelled', message: spawnId.slice(0, 8) })
      await refreshRunning()
    } catch (e) {
      error = String(e)
    }
  }

  async function setEnabled(name: string, on: boolean): Promise<void> {
    enabledMap[name] = on
    enabledMap = { ...enabledMap }
    try {
      await ipc.configUpdate({ delegation: { enabled: { [name]: on } } } as Partial<AppConfig>)
    } catch (e) {
      // revert
      enabledMap[name] = !on
      enabledMap = { ...enabledMap }
      error = String(e)
    }
  }

  async function setDefaultModel(name: string, model: string): Promise<void> {
    modelMap[name] = model
    modelMap = { ...modelMap }
    try {
      await ipc.configUpdate({ delegation: { default_model: { [name]: model } } } as Partial<AppConfig>)
    } catch (e) {
      error = String(e)
    }
  }

  function stateTone(s: string): 'success' | 'warn' | 'error' | 'neutral' {
    if (s === 'running' || s === 'pending') return 'warn'
    if (s === 'completed' || s === 'succeeded') return 'success'
    if (s === 'failed' || s === 'errored') return 'error'
    return 'neutral'
  }

  onMount(() => {
    void refresh()
    void refreshRunning()
    pollTimer = setInterval(() => void refreshRunning(), 5000)
  })
  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer)
  })
</script>

<div class="delegation-page">
  <header class="page-header">
    <h2 class="display-h2">Delegation</h2>
    <p class="lede">
      Run sub-agents on other AI CLIs installed on your machine. The conductor stays in
      charge — sub-agents execute, the Gatekeeper still decides.
    </p>
  </header>

  {#if error}
    <p class="error" role="alert">{error}</p>
  {/if}

  <!-- ── Spawn panel ───────────────────────────────── -->
  <Card elevation="glass" padding="md">
    <div class="spawn-panel">
      <div class="spawn-head">
        <h3>Spawn a sub-agent</h3>
        <p class="muted">Task and budget. The agent picks the model unless overridden.</p>
      </div>

      <form
        class="spawn-form"
        onsubmit={(e) => {
          e.preventDefault()
          void spawn()
        }}
      >
        <div class="form-row">
          <Select
            label="Agent"
            bind:value={selectedAgent}
            options={agents.map((a) => ({ value: a.name, label: a.name }))}
            fullWidth
            disabled={spawning || agents.length === 0}
          />

          <Input
            label="Model (optional)"
            bind:value={modelInput}
            fullWidth
            placeholder="claude-sonnet-4-5"
            disabled={spawning}
          />

          <Input
            label="Budget (USD)"
            type="number"
            bind:value={budgetInput as unknown as string}
            fullWidth
            disabled={spawning}
          />
        </div>

        <Textarea
          label="Task"
          bind:value={taskInput}
          fullWidth
          rows={4}
          placeholder="What should this sub-agent do?"
          disabled={spawning}
          autoresize
        />

        <div class="form-actions">
          <Button
            type="submit"
            variant="primary"
            size="md"
            loading={spawning}
            disabled={!selectedAgent || !taskInput.trim()}
          >
            {spawning ? 'Spawning…' : 'Spawn sub-agent'}
          </Button>
        </div>
      </form>
    </div>
  </Card>

  <!-- ── Running ────────────────────────────────────── -->
  {#if running.length > 0}
    <Card elevation="glass" padding="md">
      <header class="row-head">
        <h3>Running <span class="count mono">{running.length}</span></h3>
      </header>
      <ul class="running-list">
        {#each running as r (r.spawn_id)}
          <li class="running-row">
            <span class="mono spawn-id">{r.spawn_id.slice(0, 12)}</span>
            <span class="agent-name">{r.agent_name}</span>
            <Badge tone={stateTone(r.state)} size="sm" dot pulse={r.state === 'running'}>
              {r.state}
            </Badge>
            <span class="started">{new Date(r.started).toLocaleTimeString()}</span>
            <Button variant="danger" size="xs" onclick={() => cancel(r.spawn_id)}>Cancel</Button>
          </li>
        {/each}
      </ul>
    </Card>
  {/if}

  <!-- ── Available backends ─────────────────────────── -->
  <Card elevation="glass" padding="md">
    <header class="row-head">
      <h3>Available backends</h3>
      <Button variant="ghost" size="sm" onclick={refresh} loading={loading}>Refresh</Button>
    </header>
    {#if loading && agents.length === 0}
      <p class="muted">Loading…</p>
    {:else if agents.length === 0}
      <p class="muted">No backends discovered. Install a CLI (e.g. <code>npm i -g @anthropic-ai/claude-code</code>) and refresh.</p>
    {:else}
      <ul class="backend-list">
        {#each agents as a, i (a.name)}
          <li class="backend-row" style:--stagger-index={i}>
            <div class="backend-info">
              <h4>{a.name}</h4>
              <p class="desc">{a.description}</p>
              <span class="binary mono">binary: {a.binary}</span>
            </div>

            <div class="backend-controls">
              {#if a.available_models && a.available_models.length > 0}
                <Select
                  value={modelMap[a.name] || a.available_models[0]}
                  options={a.available_models.map((m) => ({ value: m, label: m }))}
                  onchange={(v: string) => void setDefaultModel(a.name, v)}
                  label="Default model"
                />
              {/if}

              <Switch
                checked={enabledMap[a.name] !== false}
                onchange={(v: boolean) => void setEnabled(a.name, v)}
                label="Enabled"
                description="Disable to hide from spawn menu."
              />
            </div>

            <div class="backend-meta">
              <Badge tone={a.installed === false ? 'warn' : 'success'} size="xs">
                {a.installed === false ? 'Not installed' : 'Installed'}
              </Badge>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </Card>

  <!-- ── Pending actions from the Gatekeeper ───────── -->
  {#if /* keep the same external surface as before */ true}
    <section class="pending-section">
      <PendingActions />
    </section>
  {/if}
</div>

<style>
  .delegation-page {
    padding: var(--space-6) var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide);
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  .page-header {
    margin-bottom: var(--space-2);
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .display-h2 {
    font-family: var(--font-display);
    font-size: var(--size-2xl);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    color: var(--text);
    margin: 0 0 var(--space-2) 0;
    line-height: var(--leading-tight);
  }
  .lede {
    font-size: var(--size-md);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    max-width: 720px;
    margin: 0;
  }

  /* ── Spawn panel ────────────────────────────────── */
  .spawn-panel { display: flex; flex-direction: column; gap: var(--space-4); }
  .spawn-head h3 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    margin: 0 0 var(--space-1) 0;
  }
  .muted {
    color: var(--text-muted);
    font-size: var(--size-sm);
    margin: 0;
  }

  .spawn-form { display: flex; flex-direction: column; gap: var(--space-3); }
  .form-row {
    display: grid;
    grid-template-columns: 1.4fr 1fr 0.6fr;
    gap: var(--space-3);
  }
  @media (max-width: 720px) {
    .form-row { grid-template-columns: 1fr; }
  }
  .form-actions {
    display: flex;
    justify-content: flex-end;
  }

  /* ── Row heads (running, backends) ──────────────── */
  .row-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-3);
  }
  .row-head h3 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    margin: 0;
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .count {
    font-size: var(--size-xs);
    color: var(--text-muted);
    background: var(--surface-2);
    padding: 2px 6px;
    border-radius: var(--radius-pill);
  }

  /* ── Running ────────────────────────────────────── */
  .running-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .running-row {
    display: grid;
    grid-template-columns: 100px 1fr auto auto auto;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-2) var(--space-3);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    font-size: var(--size-sm);
  }
  .spawn-id { color: var(--text-muted); }
  .agent-name { color: var(--text); font-weight: var(--weight-semibold); }
  .started { font-size: var(--size-xs); color: var(--text-muted); }

  /* ── Backend list ───────────────────────────────── */
  .backend-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .backend-row {
    display: grid;
    grid-template-columns: 1.5fr 1.4fr auto;
    gap: var(--space-4);
    align-items: center;
    padding: var(--space-3) var(--space-4);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    transition:
      background var(--transition-fast),
      border-color var(--transition-fast);
    animation: stagger-in var(--transition-base) var(--ease-out-expo) both;
    animation-delay: calc(var(--stagger-index, 0) * 50ms);
  }
  .backend-row:hover {
    background: var(--surface-2);
    border-color: var(--border-strong);
  }
  .backend-info h4 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    margin: 0 0 2px 0;
  }
  .desc {
    font-size: var(--size-sm);
    color: var(--text-muted);
    margin: 0 0 2px 0;
    line-height: var(--leading-snug);
  }
  .binary {
    font-size: var(--size-xs);
    color: var(--text-faint);
  }
  .backend-controls {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    align-items: stretch;
  }
  .backend-meta {
    display: flex;
    justify-content: flex-end;
  }

  .error {
    color: var(--error);
    font-size: var(--size-sm);
    padding: var(--space-2) var(--space-3);
    background: var(--error-soft);
    border: 1px solid var(--border-danger);
    border-radius: var(--radius-md);
  }

  .pending-section {
    margin-top: var(--space-3);
  }
</style>