<script lang="ts">
  import { audit } from '../stores/audit.svelte'

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
  <header>
    <h2>Audit log</h2>
    <p class="muted">Every action the daemon takes is recorded here, with timestamp, actor, and outcome.</p>
  </header>

  <div class="filter-bar">
    <input
      type="text"
      bind:value={filterAction}
      placeholder="Action contains…"
      class="input"
      onchange={applyFilter}
    />
    <select
      bind:value={filterLevel}
      class="input"
      onchange={applyFilter}
    >
      <option value="">All levels</option>
      <option value="info">info</option>
      <option value="warn">warn</option>
      <option value="error">error</option>
    </select>
    <button class="btn btn-ghost" onclick={applyFilter}>Apply</button>
  </div>

  {#if audit.loading}
    <p class="muted">Loading…</p>
  {:else if audit.events.length === 0}
    <p class="muted">No matching events.</p>
  {:else}
    <table class="audit-table">
      <thead>
        <tr>
          <th>Time</th>
          <th>Level</th>
          <th>Actor</th>
          <th>Action</th>
          <th>App</th>
          <th>Result</th>
          <th>Message</th>
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
        ← Previous
      </button>
      <span class="muted">Offset: {audit.offset}</span>
      <button
        class="btn btn-ghost"
        onclick={() => audit.nextPage()}
        disabled={audit.events.length < audit.limit}
      >
        Next →
      </button>
    </div>
  {/if}
</div>

<style>
  .audit-page {
    padding: var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: 1100px;
    margin: 0 auto;
  }
  .audit-page header h2 {
    font-size: var(--size-2xl);
    font-weight: 600;
    margin-bottom: var(--space-2);
  }
  .muted {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
  }
  .filter-bar {
    display: flex;
    gap: var(--space-2);
    margin: var(--space-4) 0;
  }
  .input,
  .select {
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    color: var(--color-text);
    padding: 8px 12px;
    border-radius: var(--radius-md);
    font-size: var(--size-md);
  }
  .input {
    flex: 1;
  }
  .btn {
    padding: 8px 16px;
    border-radius: var(--radius-md);
    font-size: var(--size-md);
  }
  .btn-ghost {
    background: transparent;
    color: var(--color-text-muted);
    border: 1px solid var(--color-border);
  }
  .btn-ghost:hover:not(:disabled) {
    color: var(--color-text);
    border-color: var(--color-border-strong);
  }
  .audit-table {
    width: 100%;
    border-collapse: collapse;
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
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
    background: var(--color-bg-overlay);
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: var(--size-xs);
    font-weight: 600;
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
  .badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: var(--radius-pill);
    font-size: var(--size-xs);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .badge.level-info {
    background: var(--color-accent-soft);
    color: var(--color-accent);
  }
  .badge.level-warn {
    background: rgba(251, 191, 36, 0.15);
    color: var(--color-warn);
  }
  .badge.level-error {
    background: rgba(248, 113, 113, 0.15);
    color: var(--color-error);
  }
  .result-allow {
    color: var(--color-success);
  }
  .result-block {
    color: var(--color-error);
  }
  .result-prompt {
    color: var(--color-warn);
  }
  .pagination {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: var(--space-4);
  }
</style>
