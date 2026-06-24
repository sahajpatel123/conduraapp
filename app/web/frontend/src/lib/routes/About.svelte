<script lang="ts">
  import { ipc } from '../ipc/client'
  import { onMount } from 'svelte'
  import { t } from '../i18n'

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
  <header class="page-header">
    <h2>{t('about.title')}</h2>
    <p class="muted"><em>{t('about.tagline')}</em></p>
  </header>

  <div class="divider"></div>

  <section class="glass-card about-section">
    <h3>{t('about.version')}</h3>
    {#if version}
      <div class="kv"><span class="k">Condura</span><span class="v">{version.version}</span></div>
      <div class="kv"><span class="k">{t('about.commit')}</span><span class="v mono">{version.commit}</span></div>
      <div class="kv"><span class="k">{t('about.built')}</span><span class="v">{version.build_date}</span></div>
      <div class="kv"><span class="k">Go</span><span class="v mono">{version.go_version}</span></div>
      <div class="kv"><span class="k">{t('about.platform')}</span><span class="v mono">{version.platform}</span></div>
    {:else}
      <p class="muted">{t('common.loading')}</p>
    {/if}
  </section>

  <section class="glass-card about-section">
    <h3>{t('about.daemon_health')}</h3>
    {#if health}
      <p>{t('about.overall')} <strong class="health-{health.overall}">{health.overall}</strong></p>
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
      <p class="muted">{t('common.loading')}</p>
    {/if}
  </section>

  <section class="glass-card about-section">
    <h3>{t('about.links')}</h3>
    <ul class="links">
      <li><a href="https://github.com/sahajpatel123/conduraapp" target="_blank" rel="noreferrer">{t('about.github')}</a></li>
      <li><a href="https://condura.app" target="_blank" rel="noreferrer">condura.app</a></li>
      <li><a href="https://hub.condura.app" target="_blank" rel="noreferrer">{t('about.skills_hub')}</a></li>
    </ul>
  </section>
</div>

<style>
  .about-page {
    padding: var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width);
    margin: 0 auto;
  }
  .about-page header {
    padding: var(--space-4) 0;
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .about-page .divider {
    animation: fade-in var(--transition-slow) var(--ease-out-expo) both;
    animation-delay: 60ms;
  }

  /* ── Section spacing — staggered entrance ─────────── */
  .about-section {
    padding: var(--space-5);
    margin-top: var(--space-5);
    animation: stagger-in var(--transition-slow) var(--ease-out-expo) both;
  }
  .about-section:nth-of-type(1) { animation-delay: 80ms; }
  .about-section:nth-of-type(2) { animation-delay: 140ms; }
  .about-section:nth-of-type(3) { animation-delay: 200ms; }
  .about-section .kv {
    animation: stagger-in var(--transition-base) var(--ease-out-expo) both;
  }
  .about-section .kv:nth-of-type(1) { animation-delay: 120ms; }
  .about-section .kv:nth-of-type(2) { animation-delay: 160ms; }
  .about-section .kv:nth-of-type(3) { animation-delay: 200ms; }
  .about-section .kv:nth-of-type(4) { animation-delay: 240ms; }
  .about-section .kv:nth-of-type(5) { animation-delay: 280ms; }
  .about-section h3 {
    font-size: var(--size-lg);
    font-weight: var(--weight-semibold);
    margin-bottom: var(--space-3);
  }

  /* ── Health check list ───────────────────────────────── */
  .check-list {
    list-style: none;
    margin: var(--space-3) 0 0 0;
    padding: 0;
  }
  .check-list li {
    display: grid;
    grid-template-columns: 90px 1fr 2fr;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--color-border);
    animation: stagger-in var(--transition-base) var(--ease-out-expo) both;
  }
  .check-list li:nth-of-type(1) { animation-delay: 40ms; }
  .check-list li:nth-of-type(2) { animation-delay: 80ms; }
  .check-list li:nth-of-type(3) { animation-delay: 120ms; }
  .check-list li:nth-of-type(4) { animation-delay: 160ms; }
  .check-list li:nth-of-type(5) { animation-delay: 200ms; }
  .check-list li:nth-of-type(n + 6) { animation-delay: 240ms; }
  .check-list li:last-child {
    border-bottom: none;
  }
  .check-list .name { font-weight: var(--weight-semibold); }
  .check-list .msg {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
  }

  /* ── Health state colors (overall + badges) ──────────── */
  .health-ok { color: var(--color-success); }
  .health-degraded { color: var(--color-warn); }
  .health-down { color: var(--color-error); }

  .badge.health-ok {
    color: var(--color-success);
    border-color: rgba(16, 185, 129, 0.4);
    background: var(--color-success-soft);
    box-shadow: 0 0 14px var(--color-success-glow);
  }
  .badge.health-degraded {
    color: var(--color-warn);
    border-color: rgba(245, 158, 11, 0.4);
    background: var(--color-warn-soft);
    box-shadow: 0 0 14px var(--color-warn-glow);
  }
  .badge.health-down {
    color: var(--color-error);
    border-color: rgba(239, 68, 68, 0.4);
    background: var(--color-error-soft);
    box-shadow: 0 0 14px var(--color-error-glow);
  }

  /* ── Links ───────────────────────────────────────────── */
  .links {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  .links li {
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--color-border);
  }
  .links li:last-child {
    border-bottom: none;
  }
  .links a {
    color: var(--color-text-muted);
    transition: color var(--transition-base), text-shadow var(--transition-base), transform var(--transition-base);
    text-decoration: none;
    display: inline-block;
  }
  .links a:hover {
    color: var(--color-accent);
    text-shadow: 0 0 16px var(--color-glow);
    transform: translateX(2px);
  }
</style>
