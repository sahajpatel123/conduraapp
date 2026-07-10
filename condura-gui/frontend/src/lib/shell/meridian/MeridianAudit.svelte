<script lang="ts">
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { audit } from '../../stores/audit.svelte'

  onMount(() => {
    void audit.refresh()
    audit.startLive()
    return () => audit.stopLive()
  })

  function formatWhen(v: unknown): string {
    if (!v) return '—'
    try {
      const d = new Date(String(v))
      return Number.isNaN(d.getTime()) ? String(v) : d.toLocaleString()
    } catch {
      return String(v)
    }
  }

  function isOffline(err: string | null): boolean {
    return !!err && /IPC client not started|not connected|Failed to fetch|daemon/i.test(err)
  }

  const showError = $derived(!!audit.error && !isOffline(audit.error))
</script>

<MeridianPage
  kicker="Ledger"
  title="Audit"
  lead="Every action Condura took — and every verdict. The chain is the truth."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void audit.refresh()}>Refresh</button>
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void audit.verifyIntegrity()}>Verify</button>
  {/snippet}

  {#if audit.integrity}
    <div class="md-panel md-panel-static banner" class:bad={!audit.integrity.ok}>
      <span class="dot" aria-hidden="true"></span>
      Chain {audit.integrity.ok ? 'intact' : 'broken'}
      {#if audit.integrity.reason}<span class="reason">— {audit.integrity.reason}</span>{/if}
    </div>
  {/if}

  {#if audit.loading && audit.events.length === 0}
    <div class="md-empty">Loading ledger…</div>
  {:else if showError}
    <div class="md-empty">{audit.error}</div>
  {:else if audit.events.length === 0}
    <div class="md-empty empty">
      <p class="empty-title">{isOffline(audit.error) ? 'Ledger offline' : 'No events yet'}</p>
      <p class="empty-lead">
        {#if isOffline(audit.error)}
          Connect the daemon to load the action chain.
        {:else}
          When Condura acts, every verdict lands here.
        {/if}
      </p>
    </div>
  {:else}
    <div class="table" role="table" aria-label="Audit events">
      <div class="head" role="row">
        <span role="columnheader">When</span>
        <span role="columnheader">Action</span>
        <span role="columnheader">Verdict</span>
      </div>
      <div class="body md-stagger">
        {#each audit.events as ev (ev.id)}
          <div class="row" role="row">
            <span class="when" role="cell">{formatWhen(ev.ts)}</span>
            <span class="action" role="cell">{ev.action || ev.message || 'event'}</span>
            <span class="verdict" role="cell" data-v={ev.result}>
              <span class="chip">{ev.result}</span>
            </span>
          </div>
        {/each}
      </div>
    </div>
    <div class="pager">
      <button
        type="button"
        class="md-btn md-btn-ghost"
        disabled={!audit.hasMore || audit.loading}
        onclick={() => void audit.loadMore()}
      >
        Load more
      </button>
    </div>
  {/if}
</MeridianPage>

<style>
  .banner {
    margin-bottom: 16px;
    font-family: var(--md-font-mono);
    font-size: 12px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex: none;
    background: var(--md-live);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-live) 20%, transparent);
  }
  .banner.bad {
    border-color: color-mix(in oklab, var(--md-halt) 40%, transparent);
    color: var(--md-halt);
  }
  .banner.bad .dot {
    background: var(--md-halt);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-halt) 20%, transparent);
  }
  .reason {
    font-weight: 500;
    text-transform: none;
    letter-spacing: 0;
    opacity: 0.85;
  }
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
  }
  .empty-title {
    margin: 0;
    font-family: var(--md-font-display);
    font-size: 16px;
    letter-spacing: -0.03em;
    color: var(--md-ink);
  }
  .empty-lead {
    margin: 0;
    max-width: 40ch;
    font-size: 13px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .table {
    background: var(--md-surface);
    border: 1px solid var(--md-line-strong);
    border-radius: 22px;
    overflow: hidden;
    box-shadow: var(--md-shadow);
  }
  .head {
    display: grid;
    grid-template-columns: 148px 1fr 96px;
    gap: 12px;
    padding: 10px 16px;
    border-bottom: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-stage) 70%, var(--md-surface));
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .row {
    display: grid;
    grid-template-columns: 148px 1fr 96px;
    gap: 12px;
    align-items: center;
    padding: 11px 16px;
    border-bottom: 1px solid var(--md-line);
    font-size: 13px;
    transition: background 160ms var(--md-ease);
  }
  .row:hover {
    background: var(--md-stage);
  }
  .row:last-child {
    border-bottom: 0;
  }
  .when {
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-ink-faint);
    font-variant-numeric: tabular-nums;
  }
  .action {
    font-weight: 600;
    color: var(--md-ink-soft);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .verdict {
    text-align: right;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 64px;
    padding: 4px 8px;
    border-radius: 999px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    border: 1px solid var(--md-line);
    background: var(--md-stage);
    color: var(--md-ink-mute);
  }
  .verdict[data-v='allow'] .chip {
    color: var(--md-live);
    border-color: color-mix(in oklab, var(--md-live) 30%, transparent);
    background: color-mix(in oklab, var(--md-live) 10%, transparent);
  }
  .verdict[data-v='block'] .chip,
  .verdict[data-v='error'] .chip {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 30%, transparent);
    background: color-mix(in oklab, var(--md-halt) 10%, transparent);
  }
  .verdict[data-v='prompt'] .chip {
    color: var(--md-cobalt);
    border-color: color-mix(in oklab, var(--md-cobalt) 30%, transparent);
    background: color-mix(in oklab, var(--md-cobalt) 10%, transparent);
  }
  .pager {
    display: flex;
    gap: 8px;
    margin-top: 16px;
  }
  @media (max-width: 640px) {
    .head {
      display: none;
    }
    .table {
      border-radius: 18px;
      background: transparent;
      border: 0;
      box-shadow: none;
      display: grid;
      gap: 8px;
    }
    .body {
      display: contents;
    }
    .row {
      grid-template-columns: 1fr auto;
      grid-template-areas:
        'action verdict'
        'when when';
      gap: 6px 10px;
      padding: 12px 14px;
      border: 1px solid var(--md-line-strong);
      border-radius: 16px;
      background: var(--md-surface);
      box-shadow: var(--md-shadow);
      border-bottom: 1px solid var(--md-line-strong);
    }
    .row:last-child {
      border-bottom: 1px solid var(--md-line-strong);
    }
    .when {
      grid-area: when;
      font-size: 10px;
    }
    .action {
      grid-area: action;
      white-space: normal;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }
    .verdict {
      grid-area: verdict;
      text-align: right;
      align-self: start;
    }
  }
</style>
