<script lang="ts">
  // Delegation — list of CLI sub-agent backends with spawn panel.
  import { ipc } from '../ipc/client'
  import type { AppConfig } from '../ipc/types'
  import { onMount, onDestroy } from 'svelte'
  import { notifications } from '../stores/notifications.svelte'
  import PendingActions from '../components/PendingActions.svelte'
  import Button from '$components/v1/Button.svelte'
  import Card from '$components/v1/Card.svelte'
  import Switch from '$components/v1/Switch.svelte'
  import Input from '$components/v1/Input.svelte'
  import Textarea from '$components/v1/Textarea.svelte'
  import Pill from '$components/v1/Pill.svelte'
  import Inline from '$components/v1/Inline.svelte'

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

  function statePill(s: string): 'success' | 'warning' | 'error' | 'neutral' {
    if (s === 'running' || s === 'pending') return 'warning'
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
    <h2 class="page-title">Delegation</h2>
    <p class="lede">
      Run sub-agents on other AI CLIs installed on your machine. The conductor stays in
      charge — sub-agents execute, the Gatekeeper still decides.
    </p>
  </header>

  {#if error}
    <p class="error-banner" role="alert">{error}</p>
  {/if}

  <Card title="Spawn a sub-agent" description="Task and budget. The agent picks the model unless overridden." variant="raised" padding="4">
    {#snippet children()}
      <form
        class="spawn-form"
        onsubmit={(e) => {
          e.preventDefault()
          void spawn()
        }}
      >
        <div class="form-row">
          <div class="field">
            <label class="field-label" for="delegate-agent">Agent</label>
            <select
              id="delegate-agent"
              class="select"
              bind:value={selectedAgent}
              disabled={spawning || agents.length === 0}
            >
              {#each agents as a (a.name)}
                <option value={a.name}>{a.name}</option>
              {/each}
            </select>
          </div>

          <div class="field">
            <label class="field-label" for="delegate-model">Model (optional)</label>
            <Input
              id="delegate-model"
              bind:value={modelInput}
              placeholder="claude-sonnet-4-5"
              disabled={spawning}
            />
          </div>

          <div class="field">
            <label class="field-label" for="delegate-budget">Budget (USD)</label>
            <Input
              id="delegate-budget"
              type="number"
              bind:value={budgetInput as unknown as string}
              disabled={spawning}
            />
          </div>
        </div>

        <div class="field">
          <label class="field-label" for="delegate-task">Task</label>
          <Textarea
            id="delegate-task"
            bind:value={taskInput}
            rows={4}
            placeholder="What should this sub-agent do?"
            disabled={spawning}
          />
        </div>

        <Inline gap="2" justify="end" class="form-actions">
          <Button
            type="submit"
            variant="primary"
            size="md"
            loading={spawning}
            disabled={!selectedAgent || !taskInput.trim()}
          >
            {spawning ? 'Spawning…' : 'Spawn sub-agent'}
          </Button>
        </Inline>
      </form>
    {/snippet}
  </Card>

  {#if running.length > 0}
    <Card variant="raised" padding="4">
      {#snippet children()}
        <header class="row-head">
          <h3 class="section-title">Running <span class="count">{running.length}</span></h3>
        </header>
        <ul class="running-list">
          {#each running as r (r.spawn_id)}
            <li class="running-row">
              <span class="mono spawn-id">{r.spawn_id.slice(0, 12)}</span>
              <span class="agent-name">{r.agent_name}</span>
              <Pill variant={statePill(r.state)} size="sm" label={r.state} />
              <span class="started">{new Date(r.started).toLocaleTimeString()}</span>
              <Button variant="destructive" size="sm" onclick={() => cancel(r.spawn_id)}>Cancel</Button>
            </li>
          {/each}
        </ul>
      {/snippet}
    </Card>
  {/if}

  <Card variant="raised" padding="4">
    {#snippet actions()}
      <Button variant="tertiary" size="sm" onclick={refresh} loading={loading}>Refresh</Button>
    {/snippet}
    {#snippet children()}
      <header class="row-head">
        <h3 class="section-title">Available backends</h3>
      </header>
      {#if loading && agents.length === 0}
        <p class="muted">Loading…</p>
      {:else if agents.length === 0}
        <p class="muted">
          No backends discovered. Install a CLI (e.g. <code>npm i -g @anthropic-ai/claude-code</code>) and refresh.
        </p>
      {:else}
        <ul class="backend-list">
          {#each agents as a, i (a.name)}
            <li class="backend-row" style:--stagger-index={i}>
              <div class="backend-info">
                <h4 class="backend-name">{a.name}</h4>
                <p class="desc">{a.description}</p>
                <span class="binary mono">binary: {a.binary}</span>
              </div>

              <div class="backend-controls">
                {#if a.available_models && a.available_models.length > 0}
                  <div class="field field--compact">
                    <label class="field-label" for="model-{a.name}">Default model</label>
                    <select
                      id="model-{a.name}"
                      class="select"
                      value={modelMap[a.name] || a.available_models[0]}
                      onchange={(e) => void setDefaultModel(a.name, (e.currentTarget as HTMLSelectElement).value)}
                    >
                      {#each a.available_models as m (m)}
                        <option value={m}>{m}</option>
                      {/each}
                    </select>
                  </div>
                {/if}

                <Switch
                  checked={enabledMap[a.name] !== false}
                  onchange={(v: boolean) => void setEnabled(a.name, v)}
                  label="Enabled"
                  description="Disable to hide from spawn menu."
                />
              </div>

              <div class="backend-meta">
                <Pill
                  variant={a.installed === false ? 'warning' : 'success'}
                  size="xs"
                  label={a.installed === false ? 'Not installed' : 'Installed'}
                />
              </div>
            </li>
          {/each}
        </ul>
      {/if}
    {/snippet}
  </Card>

  <section class="pending-section">
    <PendingActions />
  </section>
</div>

<style>
  .delegation-page {
    padding: var(--space-6) var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: 56rem;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    background-color: var(--surface-base);
  }

  .page-header {
    margin-bottom: var(--space-2);
  }

  .page-title {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    font-weight: var(--text-h2-weight);
    letter-spacing: var(--text-h2-tracking);
    color: var(--content-primary);
    margin: 0 0 var(--space-2) 0;
    line-height: var(--text-h2-leading);
  }

  .lede {
    font-size: var(--text-body-size);
    color: var(--content-secondary);
    line-height: 1.55;
    max-width: 45rem;
    margin: 0;
  }

  .section-title {
    font-size: var(--text-body-size);
    font-weight: 500;
    color: var(--content-primary);
    margin: 0;
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .spawn-form {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .form-row {
    display: grid;
    grid-template-columns: 1.4fr 1fr 0.6fr;
    gap: var(--space-3);
  }

  @media (max-width: 720px) {
    .form-row {
      grid-template-columns: 1fr;
    }
  }

  .form-actions {
    width: 100%;
  }

  .field-label {
    display: block;
    font-size: var(--text-caption-size);
    font-family: var(--font-mono);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-tertiary);
    margin-bottom: var(--space-2);
  }

  .field--compact {
    margin-bottom: var(--space-1);
  }

  .select {
    width: 100%;
    height: 36px;
    padding: 0 var(--space-3);
    background-color: var(--surface-base);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-sm);
    color: var(--content-primary);
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    cursor: pointer;
    transition: border-color var(--duration-fast) var(--ease-standard);
  }

  .select:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: 2px;
    border-color: var(--border-focus);
  }

  .select:disabled {
    color: var(--content-disabled);
    cursor: not-allowed;
  }

  .delegation-page :global(.field .input) {
    border: 1px solid var(--border-default);
    border-radius: var(--radius-sm);
    padding: 0 var(--space-3);
  }

  .row-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-3);
  }

  .count {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-secondary);
    background-color: var(--surface-sunken);
    padding: 2px 6px;
    border-radius: var(--radius-pill);
    font-weight: 400;
  }

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
    background-color: var(--surface-sunken);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    font-size: var(--text-body-sm-size);
  }

  @media (max-width: 720px) {
    .running-row {
      grid-template-columns: 1fr;
    }
  }

  .mono {
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
  }

  .spawn-id {
    color: var(--content-secondary);
  }

  .agent-name {
    color: var(--content-primary);
    font-weight: 500;
  }

  .started {
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
  }

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
    background-color: var(--surface-sunken);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    transition:
      background-color var(--duration-fast) var(--ease-standard),
      border-color var(--duration-fast) var(--ease-standard);
    animation: delegation-stagger var(--duration-base) var(--ease-standard) both;
    animation-delay: calc(var(--stagger-index, 0) * 50ms);
  }

  @keyframes delegation-stagger {
    from { opacity: 0; transform: translateY(6px); }
    to { opacity: 1; transform: translateY(0); }
  }

  @media (max-width: 880px) {
    .backend-row {
      grid-template-columns: 1fr;
    }
  }

  .backend-row:hover {
    background-color: var(--surface-raised);
    border-color: var(--border-strong);
  }

  .backend-name {
    font-size: var(--text-body-size);
    font-weight: 500;
    color: var(--content-primary);
    margin: 0 0 2px 0;
  }

  .desc {
    font-size: var(--text-body-sm-size);
    color: var(--content-secondary);
    margin: 0 0 2px 0;
    line-height: 1.45;
  }

  .binary {
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
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

  .muted {
    color: var(--content-tertiary);
    font-size: var(--text-body-sm-size);
    margin: 0;
  }

  .muted code {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
  }

  .error-banner {
    color: var(--status-error-fg);
    font-size: var(--text-body-sm-size);
    padding: var(--space-2) var(--space-3);
    background-color: var(--status-error-bg);
    border: 1px solid var(--status-error-border);
    border-radius: var(--radius-md);
  }

  .pending-section {
    margin-top: var(--space-3);
  }
</style>
