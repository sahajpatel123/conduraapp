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
  <header class="page-header">
    <h2>{t('delegation.title')}</h2>
    <p class="muted">{t('delegation.intro')}</p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <section class="glass-card card">
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
        {#each agents as a, i}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
          <li class="agent-card glass-card stagger-item" class:selected={a.name === selectedAgent} style="--stagger-index: {i}" onclick={() => (selectedAgent = a.name)}>
            <strong>{a.name}</strong>
            <span class="desc">{a.description}</span>
            <span class="binary">{t('delegation.binary', a.binary)}</span>
          </li>
        {/each}
      </ul>
    {/if}
    <button class="btn btn-ghost btn-sm" onclick={refresh} disabled={loading}>{t('delegation.refresh')}</button>
  </section>

  <section class="glass-card card">
    <h3>{t('delegation.spawn_title')}</h3>
    <form onsubmit={(e) => { e.preventDefault(); void spawn(); }}>
      <label class="field">
        <span>{t('delegation.agent_label')}</span>
        <select class="input" bind:value={selectedAgent} disabled={spawning || agents.length === 0}>
          {#each agents as a}
            <option value={a.name}>{a.name}</option>
          {/each}
        </select>
      </label>
      <label class="field">
        <span>{t('delegation.task_label')}</span>
        <textarea
          class="input"
          bind:value={taskInput}
          rows="4"
          placeholder={t('delegation.task_placeholder')}
          disabled={spawning}
        ></textarea>
      </label>
      <label class="field">
        <span>{t('delegation.model_label')}</span>
        <input class="input" type="text" bind:value={modelInput} placeholder={t('delegation.model_placeholder')} disabled={spawning} />
      </label>
      <label class="field">
        <span>{t('delegation.budget_label')}</span>
        <input class="input" type="number" bind:value={budgetInput} min="0.01" step="0.10" disabled={spawning} />
      </label>
      <button type="submit" class="btn btn-primary" disabled={spawning || !selectedAgent || !taskInput.trim()}>
        {spawning ? t('delegation.spawning') : t('delegation.spawn_button')}
      </button>
    </form>
  </section>

  {#if lastSpawn}
    <section class="glass-card card">
      <h3>{t('delegation.last_spawn')}</h3>
      <div class="result">
        <div class="kv"><span class="k">{t('delegation.spawn_id')}</span><span class="v mono">{lastSpawn.spawn_id}</span></div>
        <div class="kv"><span class="k">{t('delegation.agent')}</span><span class="v">{lastSpawn.agent_name}</span></div>
        <div class="kv"><span class="k">{t('delegation.state')}</span><span class="v"><span class="badge state-{lastSpawn.state}">{lastSpawn.state}</span></span></div>
        <div class="kv"><span class="k">{t('delegation.cost')}</span><span class="v">${lastSpawn.cost?.toFixed(4) ?? '0.0000'}</span></div>
        <div class="kv"><span class="k">{t('delegation.tokens')}</span><span class="v">{lastSpawn.tokens ?? 0}</span></div>
        {#if lastSpawn.started}
          <div class="kv"><span class="k">{t('delegation.started')}</span><span class="v">{new Date(lastSpawn.started).toLocaleString()}</span></div>
        {/if}
        {#if lastSpawn.finished}
          <div class="kv"><span class="k">{t('delegation.finished')}</span><span class="v">{new Date(lastSpawn.finished).toLocaleString()}</span></div>
        {/if}
      </div>
      {#if lastSpawn.output}
        <details>
          <summary>{t('delegation.output')}</summary>
          <pre>{lastSpawn.output}</pre>
        </details>
      {/if}
      {#if lastSpawn && lastSpawn.state === 'running'}
        <button class="btn btn-danger btn-sm" onclick={() => lastSpawn && cancel(lastSpawn.spawn_id)}>{t('delegation.cancel')}</button>
      {/if}
    </section>
  {/if}

  <section class="pending-section">
    <PendingActions />
  </section>
</div>

<style>
  .delegation-page {
    padding: var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: 900px;
    margin: 0 auto;
  }
  .page-header {
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .card {
    padding: var(--space-5);
    margin: var(--space-4) 0;
  }
  .card h3 {
    font-size: var(--size-lg);
    font-weight: var(--weight-semibold);
    margin-bottom: var(--space-3);
  }
  .agent-list {
    list-style: none;
    padding: 0;
    margin: var(--space-2) 0;
    display: grid;
    gap: var(--space-2);
  }
  .agent-card {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: var(--space-3);
    cursor: pointer;
    font-size: var(--size-sm);
    transition: border-color var(--transition-base), background var(--transition-base);
  }
  .agent-card:hover:not(.selected) {
    border-color: var(--glass-border-hover);
    box-shadow: var(--shadow-glow-accent);
  }
  .agent-card.selected {
    border-color: var(--color-border-accent);
    background: var(--color-accent-gradient-subtle), var(--glass-bg);
    box-shadow: var(--shadow-glow-accent);
  }
  .agent-card .desc { color: var(--color-text-muted); }
  .agent-card .binary {
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    color: var(--color-text-faint);
  }
  form {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .field > span {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
  }
  .kv :global(.state-running) {
    color: var(--color-warn);
    border-color: var(--color-warn);
    background: var(--color-warn-soft);
  }
  .kv :global(.state-completed) {
    color: var(--color-success);
    border-color: var(--color-success);
    background: var(--color-success-soft);
  }
  .kv :global(.state-failed) {
    color: var(--color-error);
    border-color: var(--color-error);
    background: var(--color-error-soft);
  }
  .kv :global(.state-cancelled) {
    color: var(--color-text-muted);
  }
  pre {
    background: var(--glass-bg-active);
    padding: var(--space-3);
    border-radius: var(--radius-md);
    overflow: auto;
    max-height: 300px;
    font-size: var(--size-xs);
    line-height: var(--leading-relaxed);
  }
  .pending-section {
    margin: var(--space-4) 0;
  }
  .error {
    color: var(--color-error);
    margin: var(--space-2) 0;
    font-size: var(--size-sm);
  }
</style>
