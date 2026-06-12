<script lang="ts">
  import { ipc } from '../ipc/client'
  import { onMount } from 'svelte'

  let version = $state<{ version: string; commit: string; build_date: string; go_version: string; platform: string } | null>(null)
  let health = $state<{ overall: string; checks: Array<{ name: string; state: string; message: string }> } | null>(null)

  onMount(async () => {
    try {
      version = await ipc.version()
    } catch {
      // ignore
    }
    try {
      health = await ipc.healthSnapshot()
    } catch {
      // ignore
    }
  })
</script>

<div class="about-page">
  <header>
    <h2>About Synaptic</h2>
    <p class="muted"><em>A free, on-device AI agent.</em></p>
  </header>

  <div class="divider"></div>

  <section class="card">
    <h3>Version</h3>
    {#if version}
      <div class="kv"><span class="k">Synaptic</span><span class="v">{version.version}</span></div>
      <div class="kv"><span class="k">Commit</span><span class="v mono">{version.commit}</span></div>
      <div class="kv"><span class="k">Built</span><span class="v">{version.build_date}</span></div>
      <div class="kv"><span class="k">Go</span><span class="v mono">{version.go_version}</span></div>
      <div class="kv"><span class="k">Platform</span><span class="v mono">{version.platform}</span></div>
    {:else}
      <p class="muted">Loading…</p>
    {/if}
  </section>

  <section class="card">
    <h3>Daemon health</h3>
    {#if health}
      <p>Overall: <strong class="health-{health.overall}">{health.overall}</strong></p>
      <ul class="check-list">
        {#each health.checks as c}
          <li>
            <span class="badge health-{c.state}">{c.state}</span>
            <span class="name">{c.name}</span>
            <span class="msg">{c.message}</span>
          </li>
        {/each}
      </ul>
    {:else}
      <p class="muted">Loading…</p>
    {/if}
  </section>

  <section class="card">
    <h3>Links</h3>
    <ul class="links">
      <li><a href="https://github.com/sahajpatel123/synapticapp" target="_blank" rel="noreferrer">GitHub repository</a></li>
      <li><a href="https://synaptic.app" target="_blank" rel="noreferrer">synaptic.app</a></li>
      <li><a href="https://hub.synaptic.app" target="_blank" rel="noreferrer">Skills Hub</a></li>
    </ul>
  </section>
</div>

<style>
  .about-page {
    padding: var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: 760px;
    margin: 0 auto;
  }
  .about-page header {
    padding: var(--space-4) 0;
  }
  .about-page header h2 {
    font-size: 32px;
    font-weight: 600;
    margin-bottom: var(--space-2);
    background: var(--color-accent-gradient);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .muted {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
  }
  .divider {
    height: 1px;
    background: linear-gradient(90deg, var(--color-accent), transparent);
    opacity: 0.3;
    margin-bottom: var(--space-4);
  }
  .card {
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-xl);
    padding: var(--space-5);
    margin-top: var(--space-5);
    transition: border-color var(--transition-base);
  }
  .card:hover {
    border-color: rgba(255,255,255,0.12);
  }
  .card h3 {
    font-size: var(--size-lg);
    font-weight: 600;
    margin-bottom: var(--space-3);
  }
  .kv {
    display: flex;
    justify-content: space-between;
    padding: var(--space-2) 0;
    font-size: var(--size-md);
    border-bottom: 1px dotted var(--glass-border);
  }
  .kv:last-child {
    border-bottom: none;
  }
  .kv .k {
    color: var(--color-text-muted);
  }
  .kv .v {
    color: var(--color-text);
  }
  .kv .v.mono {
    font-family: var(--font-mono);
    font-size: var(--size-sm);
  }
  .check-list {
    list-style: none;
    margin-top: var(--space-3);
  }
  .check-list li {
    display: grid;
    grid-template-columns: 90px 1fr 2fr;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--glass-border);
  }
  .check-list li:last-child {
    border-bottom: none;
  }
  .badge {
    display: inline-block;
    padding: 4px 10px;
    border-radius: var(--radius-pill);
    font-size: var(--size-xs);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 500;
    text-align: center;
  }
  .badge.health-ok {
    background: rgba(74, 222, 128, 0.15);
    color: var(--color-success);
  }
  .badge.health-degraded {
    background: rgba(251, 191, 36, 0.15);
    color: var(--color-warn);
  }
  .badge.health-down {
    background: rgba(248, 113, 113, 0.15);
    color: var(--color-error);
  }
  .name {
    font-weight: 600;
  }
  .msg {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
  }
  .health-ok { color: var(--color-success); }
  .health-degraded { color: var(--color-warn); }
  .health-down { color: var(--color-error); }
  .links {
    list-style: none;
    padding: 0;
  }
  .links li {
    padding: var(--space-2) 0;
  }
  .links a {
    transition: all var(--transition-fast);
  }
  .links a:hover {
    text-shadow: var(--shadow-glow);
    text-decoration: underline;
  }
</style>
