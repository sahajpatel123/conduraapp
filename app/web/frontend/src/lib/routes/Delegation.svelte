<script lang="ts">
  import { ipc } from '../ipc/client'
  import { onMount } from 'svelte'
  import PendingActions from '../components/PendingActions.svelte'
  import { t } from '../i18n'

  type Agent = {
    name: string
    description: string
    binary: string
  }

  type SpawnResult = {
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

  let agents = $state<Agent[]>([])
  let loading = $state(false)
  let error = $state<string | null>(null)
  let selectedAgent = $state<string>('')
  let taskInput = $state('')
  let modelInput = $state('')
  let budgetInput = $state(1.0)
  let spawning = $state(false)
  let lastSpawn = $state<SpawnResult | null>(null)

  async function refresh() {
    loading = true
    error = null
    try {
      const resp = await ipc.call<{ agents: Agent[] }>('delegate.list_agents', {})
      agents = resp?.agents ?? []
      if (agents.length > 0 && !selectedAgent) {
        selectedAgent = agents[0].name
      }
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  }

  async function spawn() {
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
    } catch (e) {
      error = String(e)
    } finally {
      spawning = false
    }
  }

  async function cancel(spawnId: string) {
    try {
      await ipc.call('delegate.cancel', { spawn_id: spawnId })
      error = null
    } catch (e) {
      error = String(e)
    }
  }

  onMount(refresh)
</script>

<div class="delegation-page">
  <header>
    <h2>{t('delegation.title')}</h2>
    <p class="muted">
      {t('delegation.intro')}
    </p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <section class="card">
    <h3>{t('delegation.available_title')}</h3>
    <p class="muted">
      {t('delegation.available_intro')}
    </p>
    {#if loading}
      <p class="muted">{t('common.loading')}</p>
    {:else if agents.length === 0}
      <p class="muted">{t('delegation.no_agents')}</p>
    {:else}
      <ul class="agent-list">
        {#each agents as a}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
          <li class:selected={a.name === selectedAgent} onclick={() => (selectedAgent = a.name)}>
            <strong>{a.name}</strong>
            <span class="desc">{a.description}</span>
            <span class="binary">{t('delegation.binary', a.binary)}</span>
          </li>
        {/each}
      </ul>
    {/if}
    <button class="btn btn-ghost" onclick={refresh} disabled={loading}>{t('delegation.refresh')}</button>
  </section>

  <section class="card">
    <h3>{t('delegation.spawn_title')}</h3>
    <form onsubmit={(e) => { e.preventDefault(); void spawn(); }}>
      <label class="field">
        <span>{t('delegation.agent_label')}</span>
        <select bind:value={selectedAgent} disabled={spawning || agents.length === 0}>
          {#each agents as a}
            <option value={a.name}>{a.name}</option>
          {/each}
        </select>
      </label>
      <label class="field">
        <span>{t('delegation.task_label')}</span>
        <textarea
          bind:value={taskInput}
          rows="4"
          placeholder={t('delegation.task_placeholder')}
          disabled={spawning}
        ></textarea>
      </label>
      <label class="field">
        <span>{t('delegation.model_label')}</span>
        <input type="text" bind:value={modelInput} placeholder={t('delegation.model_placeholder')} disabled={spawning} />
      </label>
      <label class="field">
        <span>{t('delegation.budget_label')}</span>
        <input type="number" bind:value={budgetInput} min="0.01" step="0.10" disabled={spawning} />
      </label>
      <button type="submit" class="btn btn-primary" disabled={spawning || !selectedAgent || !taskInput.trim()}>
        {spawning ? t('delegation.spawning') : t('delegation.spawn_button')}
      </button>
    </form>
  </section>

  {#if lastSpawn}
    <section class="card">
      <h3>{t('delegation.last_spawn')}</h3>
      <dl class="result">
        <dt>{t('delegation.spawn_id')}</dt>
        <dd class="mono">{lastSpawn.spawn_id}</dd>
        <dt>{t('delegation.agent')}</dt>
        <dd>{lastSpawn.agent_name}</dd>
        <dt>{t('delegation.state')}</dt>
        <dd class="state-{lastSpawn.state}">{lastSpawn.state}</dd>
        <dt>{t('delegation.cost')}</dt>
        <dd>${lastSpawn.cost?.toFixed(4) ?? '0.0000'}</dd>
        <dt>{t('delegation.tokens')}</dt>
        <dd>{lastSpawn.tokens ?? 0}</dd>
        {#if lastSpawn.started}
          <dt>{t('delegation.started')}</dt>
          <dd>{new Date(lastSpawn.started).toLocaleString()}</dd>
        {/if}
        {#if lastSpawn.finished}
          <dt>{t('delegation.finished')}</dt>
          <dd>{new Date(lastSpawn.finished).toLocaleString()}</dd>
        {/if}
      </dl>
      {#if lastSpawn.output}
        <details>
          <summary>{t('delegation.output')}</summary>
          <pre>{lastSpawn.output}</pre>
        </details>
      {/if}
  {#if lastSpawn && lastSpawn.state === 'running'}
    <button class="btn btn-danger" onclick={() => lastSpawn && cancel(lastSpawn.spawn_id)}>{t('delegation.cancel')}</button>
  {/if}
    </section>
  {/if}

  <section class="card">
    <PendingActions />
  </section>
</div>

<style>
  .delegation-page { padding: var(--space-5); overflow-y: auto; height: 100%; max-width: 900px; margin: 0 auto; }
  .delegation-page header h2 { font-size: var(--size-2xl); font-weight: 600; margin-bottom: var(--space-2); }
  .card { background: var(--glass-bg); border: 1px solid var(--glass-border); border-radius: var(--radius-xl); padding: var(--space-5); margin: var(--space-4) 0; }
  .muted { color: var(--color-text-muted); font-size: var(--size-sm); }
  .error { color: var(--color-error, #f87171); margin: var(--space-2) 0; }
  .agent-list { list-style: none; padding: 0; margin: var(--space-2) 0; display: grid; gap: var(--space-2); }
  .agent-list li { display: flex; flex-direction: column; gap: 2px; padding: var(--space-2) var(--space-3); border: 1px solid transparent; border-radius: var(--radius-md, 6px); cursor: pointer; font-size: var(--size-sm); }
  .agent-list li.selected { border-color: var(--color-accent, #4a9eff); background: var(--color-accent-soft, rgba(74,158,255,0.08)); }
  .agent-list .desc { color: var(--color-text-muted); }
  .agent-list .binary { font-family: var(--font-mono); font-size: var(--size-xs); color: var(--color-text-muted); }
  form { display: flex; flex-direction: column; gap: var(--space-3); }
  .field { display: flex; flex-direction: column; gap: 4px; }
  .field span { color: var(--color-text-muted); font-size: var(--size-sm); }
  .field input, .field select, .field textarea { padding: 6px 10px; background: var(--color-bg-elev, rgba(255,255,255,0.04)); color: var(--color-text); border: 1px solid var(--glass-border); border-radius: var(--radius-md, 6px); font-family: inherit; }
  .field textarea { resize: vertical; min-height: 80px; }
  .btn { padding: 8px 16px; border-radius: var(--radius-md, 6px); font-size: var(--size-md); cursor: pointer; border: none; }
  .btn-primary { background: var(--color-accent-gradient); color: white; }
  .btn-ghost { background: transparent; border: 1px solid var(--glass-border); color: var(--color-text-muted); }
  .btn-danger { background: #c0392b; color: white; }
  .btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .result { display: grid; grid-template-columns: 120px 1fr; gap: 4px 12px; font-size: var(--size-sm); }
  .result dt { color: var(--color-text-muted); }
  .result dd { margin: 0; }
  .state-running { color: #fbbf24; }
  .state-completed { color: #4ade80; }
  .state-failed { color: #f87171; }
  .state-cancelled { color: var(--color-text-muted); }
  pre { background: rgba(0,0,0,0.3); padding: var(--space-3); border-radius: var(--radius-md, 6px); overflow: auto; max-height: 300px; font-size: var(--size-xs); }
  .mono { font-family: var(--font-mono); }
</style>
