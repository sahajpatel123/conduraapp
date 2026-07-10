<script lang="ts">
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import {
    pendingActions,
    pendingCount,
    refreshPendingActions,
    approvePending,
    denyPending,
    executePending,
    startPolling,
    stopPolling,
  } from '../../stores/pending.svelte'

  onMount(() => {
    void refreshPendingActions()
    startPolling(4000)
    return () => stopPolling()
  })

  const rows = $derived($pendingActions)
  const count = $derived($pendingCount)

  function initial(name: string): string {
    const t = name.trim()
    return t ? t[0]!.toUpperCase() : '?'
  }
</script>

<MeridianPage
  kicker="Control"
  title="Agents"
  lead="Sub-agents and pending actions. Approve, deny, or execute — you stay in charge."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void refreshPendingActions()}>
      Refresh
    </button>
  {/snippet}

  <p class="count">
    <span class="pill" class:live={count > 0}>{count}</span>
    pending
  </p>

  {#if rows.length === 0}
    <div class="md-empty empty">
      <p class="empty-title">Nothing in flight</p>
      <p class="empty-lead">When a sub-agent waits on you, it shows up here with Approve, Run, or Deny.</p>
    </div>
  {:else}
    <div class="list md-stagger">
      {#each rows as row (row.id)}
        <article class="md-panel md-panel-static card">
          <header>
            <div class="mono" aria-hidden="true">{initial(row.agent_name || row.kind || row.id)}</div>
            <div class="copy">
              <strong>{row.agent_name || row.kind || row.id}</strong>
              <span class="meta">{row.status}</span>
            </div>
          </header>
          {#if row.gate_reason || row.payload?.command}
            <p>{row.gate_reason || row.payload?.command}</p>
          {/if}
          <footer>
            <button type="button" class="md-btn md-btn-primary" onclick={() => void approvePending(row.id)}>
              Approve
            </button>
            <button type="button" class="md-btn md-btn-ghost" onclick={() => void executePending(row.id)}>
              Run
            </button>
            <button type="button" class="md-btn md-btn-danger" onclick={() => void denyPending(row.id)}>
              Deny
            </button>
          </footer>
        </article>
      {/each}
    </div>
  {/if}
</MeridianPage>

<style>
  .count {
    display: flex;
    align-items: center;
    gap: 8px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 16px;
  }
  .pill {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 22px;
    height: 22px;
    padding: 0 7px;
    border-radius: 999px;
    background: var(--md-stage);
    border: 1px solid var(--md-line);
    font-weight: 700;
    color: var(--md-ink-mute);
  }
  .pill.live {
    background: color-mix(in oklab, var(--md-cobalt) 14%, transparent);
    border-color: color-mix(in oklab, var(--md-cobalt) 35%, transparent);
    color: var(--md-cobalt);
    animation: md-rise 280ms var(--md-spring) both;
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
    max-width: 42ch;
    font-size: 13px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .list {
    display: grid;
    gap: 12px;
  }
  .card header {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 12px;
  }
  .mono {
    width: 42px;
    height: 42px;
    border-radius: 14px;
    display: grid;
    place-items: center;
    flex: none;
    font-family: var(--md-font-display);
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -0.04em;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
  }
  .copy {
    min-width: 0;
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .card strong {
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
  .card p {
    margin: 0 0 16px;
    color: var(--md-ink-mute);
    font-size: 14px;
    line-height: 1.5;
  }
  footer {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
  @media (max-width: 560px) {
    .empty {
      padding: 28px 14px;
    }
    .empty-lead {
      font-size: 12.5px;
      max-width: 32ch;
    }
    .card header {
      gap: 10px;
      margin-bottom: 10px;
    }
    .mono {
      width: 38px;
      height: 38px;
      font-size: 16px;
    }
    .card strong {
      font-size: 16px;
    }
    .card p {
      font-size: 13px;
      margin-bottom: 12px;
    }
    footer {
      gap: 6px;
    }
    footer :global(.md-btn) {
      flex: 1;
      min-width: 0;
      justify-content: center;
      padding: 9px 12px;
      font-size: 12px;
    }
  }
</style>
