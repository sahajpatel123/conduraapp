<script lang="ts">
  import { audit } from '../stores/audit.svelte'
  import { t } from '../i18n'

  let filterAction = $state('')
  let filterLevel = $state<'' | 'info' | 'warn' | 'error'>('')

  function applyFilter(): void {
    audit.setFilter(filterAction, filterLevel)
  }

  function levelClass(level: string): string {
    if (level === 'error') return 'level-error'
    if (level === 'warn') return 'level-warn'
    return 'level-info'
  }
</script>

<div class="audit-page">
  <header class="page-header">
    <h2>{t('audit.title')}</h2>
    <p class="muted">{t('audit.intro')}</p>
  </header>

  <div class="filter-bar">
    <div class="filter-pill">
      <input
        type="text"
        bind:value={filterAction}
        placeholder={t('audit.filter_placeholder')}
        class="input"
        onchange={applyFilter}
      />
      <select
        bind:value={filterLevel}
        class="input"
        onchange={applyFilter}
      >
        <option value="">{t('audit.all_levels')}</option>
        <option value="info">info</option>
        <option value="warn">warn</option>
        <option value="error">error</option>
      </select>
    </div>
    <button class="btn btn-ghost" onclick={applyFilter}>{t('audit.apply')}</button>
  </div>

  {#if audit.loading}
    <p class="muted">{t('common.loading')}</p>
  {:else if audit.events.length === 0}
    <p class="muted">{t('audit.empty')}</p>
  {:else}
    <table class="audit-table">
      <thead>
        <tr>
          <th>{t('audit.col.time')}</th>
          <th>{t('audit.col.level')}</th>
          <th>{t('audit.col.actor')}</th>
          <th>{t('audit.col.action')}</th>
          <th>{t('audit.col.app')}</th>
          <th>{t('audit.col.result')}</th>
          <th>{t('audit.col.message')}</th>
        </tr>
      </thead>
      <tbody>
        {#each audit.events as ev (ev.id)}
          <tr>
            <td class="ts">{new Date(ev.ts).toLocaleString()}</td>
            <td><span class="badge {levelClass(ev.level)}">{ev.level}</span></td>
            <td>{ev.actor}</td>
            <td class="action">{ev.action}</td>
            <td>{ev.app}</td>
            <td><span class="result-{ev.result}">{ev.result}</span></td>
            <td class="msg">{ev.message}</td>
          </tr>
        {/each}
      </tbody>
    </table>

    <div class="pagination">
      <button class="btn btn-ghost" onclick={() => audit.prevPage()} disabled={audit.offset === 0}>
        ← {t('audit.previous')}
      </button>
      <span class="muted">{t('audit.offset', audit.offset)}</span>
      <button
        class="btn btn-ghost"
        onclick={() => audit.nextPage()}
        disabled={audit.events.length < audit.limit}
      >
        {t('audit.next')} →
      </button>
    </div>
  {/if}
</div>

<style>
  .audit-page {
    padding: var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide);
    margin: 0 auto;
  }
  .audit-page .page-header {
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .audit-page .filter-bar {
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
    animation-delay: 60ms;
  }

  /* ── Filter bar — premium glass pill ─────────────────── */
  .filter-bar {
    display: flex;
    gap: var(--space-3);
    margin: var(--space-4) 0;
    align-items: center;
  }
  .filter-pill {
    display: flex;
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-pill);
    overflow: hidden;
    flex: 1;
    transition: border-color var(--transition-base), box-shadow var(--transition-base);
  }
  .filter-pill:focus-within {
    border-color: var(--color-accent);
    box-shadow: var(--shadow-focus), var(--shadow-glow-accent);
  }
  .filter-pill .input {
    background: transparent;
    border: none;
    border-right: 1px solid var(--color-border);
    flex: 1;
  }
  .filter-pill .input:focus {
    box-shadow: none;
  }
  .filter-pill select.input {
    border-right: none;
    width: auto;
    min-width: 140px;
  }

  /* ── Audit table ─────────────────────────────────────── */
  .audit-table {
    width: 100%;
    border-collapse: collapse;
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-xl);
    overflow: hidden;
  }
  .audit-table th,
  .audit-table td {
    text-align: left;
    padding: var(--space-3);
    font-size: var(--size-sm);
    border-bottom: 1px solid var(--color-border);
  }
  .audit-table th {
    background: var(--color-bg-elevated);
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wide);
    font-size: var(--size-xs);
    font-weight: var(--weight-semibold);
  }
  .audit-table tbody tr {
    transition: background var(--transition-base), box-shadow var(--transition-base);
    animation: stagger-in var(--transition-base) var(--ease-out-expo) both;
  }
  .audit-table tbody tr:nth-of-type(1) { animation-delay: 20ms; }
  .audit-table tbody tr:nth-of-type(2) { animation-delay: 40ms; }
  .audit-table tbody tr:nth-of-type(3) { animation-delay: 60ms; }
  .audit-table tbody tr:nth-of-type(4) { animation-delay: 80ms; }
  .audit-table tbody tr:nth-of-type(5) { animation-delay: 100ms; }
  .audit-table tbody tr:nth-of-type(6) { animation-delay: 120ms; }
  .audit-table tbody tr:nth-of-type(7) { animation-delay: 140ms; }
  .audit-table tbody tr:nth-of-type(8) { animation-delay: 160ms; }
  .audit-table tbody tr:nth-of-type(n + 9) { animation-delay: 180ms; }
  .audit-table tbody tr:hover {
    background: var(--color-bg-hover);
    box-shadow: inset 2px 0 0 var(--color-accent);
  }
  .audit-table tr:last-child td {
    border-bottom: none;
  }
  .audit-table td.ts,
  .audit-table td.action {
    font-family: var(--font-mono);
    font-size: var(--size-xs);
  }
  .audit-table td.msg {
    color: var(--color-text-muted);
  }

  /* ── Level badges (component-specific colors, glow-enhanced) ── */
  .badge.level-info {
    color: var(--color-info);
    border-color: rgba(59, 130, 246, 0.4);
    background: var(--color-info-soft);
    box-shadow: 0 0 14px rgba(59, 130, 246, 0.15);
  }
  .badge.level-warn {
    color: var(--color-warn);
    border-color: rgba(245, 158, 11, 0.4);
    background: var(--color-warn-soft);
    box-shadow: 0 0 14px var(--color-warn-glow);
  }
  .badge.level-error {
    color: var(--color-error);
    border-color: rgba(239, 68, 68, 0.4);
    background: var(--color-error-soft);
    box-shadow: 0 0 14px var(--color-error-glow);
  }

  /* ── Result indicators ───────────────────────────────── */
  .result-allow { color: var(--color-success); }
  .result-block { color: var(--color-error); }
  .result-prompt { color: var(--color-warn); }

  /* ── Pagination ──────────────────────────────────────── */
  .pagination {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: var(--space-4);
  }
</style>
