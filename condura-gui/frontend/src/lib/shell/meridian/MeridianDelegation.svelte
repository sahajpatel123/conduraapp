<script lang="ts">
  /**
   * Agents — consent control room for pending sub-agent actions.
   * Signature: live queue pulse + consent sheets (Approve / Run / Deny).
   */
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
  const waiting = $derived(rows.filter((r) => r.status === 'pending'))
  const settled = $derived(rows.filter((r) => r.status !== 'pending').slice(0, 6))

  function initial(name: string): string {
    const t = name.trim()
    return t ? t[0]!.toUpperCase() : '?'
  }

  function goAsk(): void {
    window.location.hash = '#/'
  }
</script>

<MeridianPage
  kicker="Control room"
  title="Agents"
  lead="Sub-agents wait here. You approve, run, or deny — the Gatekeeper never decides alone."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void refreshPendingActions()}>
      Refresh
    </button>
  {/snippet}

  <div class="board md-stagger">
    <section class="pulse" class:hot={count > 0} aria-live="polite">
      <div class="pulse-ring" aria-hidden="true">
        <span class="pulse-core">{count}</span>
      </div>
      <div class="pulse-copy">
        <p class="cite">{count > 0 ? 'awaiting you' : 'quiet'}</p>
        <h2>{count > 0 ? `${count} pending consent` : 'Nothing in flight'}</h2>
        <p>
          {#if count > 0}
            Each sheet below is a gated action. Approve opens the door; Deny seals it; Run executes an approved row.
          {:else}
            When a sub-agent needs you, this room lights up. Until then, the meridian stays calm.
          {/if}
        </p>
        {#if count === 0}
          <button type="button" class="md-btn md-btn-primary" onclick={goAsk}>Ask Condura</button>
        {/if}
      </div>
    </section>

    {#if waiting.length > 0}
      <section class="queue">
        <p class="cite">consent sheets</p>
        <div class="list">
          {#each waiting as row (row.id)}
            <article class="sheet">
              <header>
                <div class="mono" aria-hidden="true">{initial(row.agent_name || row.kind || row.id)}</div>
                <div class="copy">
                  <strong>{row.agent_name || row.kind || row.id}</strong>
                  <span class="meta">{row.kind || 'action'} · {row.status}</span>
                </div>
              </header>
              <p class="reason">{row.gate_reason || 'Gatekeeper requires your decision.'}</p>
              {#if row.payload?.command || row.payload?.path || row.payload?.target}
                <p class="payload">
                  <span>cite</span>
                  {row.payload?.command || row.payload?.path || row.payload?.target}
                </p>
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
      </section>
    {/if}

    {#if settled.length > 0}
      <section class="settled">
        <p class="cite">recent decisions</p>
        <ul>
          {#each settled as row (row.id)}
            <li data-status={row.status}>
              <span class="dot" aria-hidden="true"></span>
              <strong>{row.agent_name || row.kind || row.id}</strong>
              <span class="meta">{row.status}</span>
            </li>
          {/each}
        </ul>
      </section>
    {/if}
  </div>
</MeridianPage>

<style>
  .board {
    display: grid;
    gap: 20px;
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 12px;
  }
  .pulse {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 22px;
    align-items: center;
    padding: 22px 24px;
    border-radius: 22px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
  }
  .pulse.hot {
    border-color: color-mix(in oklab, var(--md-cobalt) 35%, transparent);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--md-cobalt) 10%, transparent);
  }
  .pulse-ring {
    width: 88px;
    height: 88px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    border: 2px solid var(--md-line-strong);
    background: var(--md-stage);
    position: relative;
  }
  .pulse.hot .pulse-ring {
    border-color: color-mix(in oklab, var(--md-cobalt) 45%, transparent);
    animation: md-pulse-soft 2.4s var(--md-ease) infinite;
  }
  .pulse-core {
    font-family: var(--md-font-display);
    font-size: 32px;
    font-weight: 700;
    letter-spacing: -0.05em;
    color: var(--md-ink-mute);
  }
  .pulse.hot .pulse-core {
    color: var(--md-cobalt);
  }
  .pulse-copy h2 {
    font-family: var(--md-font-display);
    font-size: 24px;
    letter-spacing: -0.04em;
    margin: 0 0 8px;
  }
  .pulse-copy p {
    margin: 0 0 14px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
    max-width: 48ch;
  }

  .list {
    display: grid;
    gap: 12px;
  }
  .sheet {
    border-radius: 20px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 18px 18px 16px;
    box-shadow: var(--md-shadow);
  }
  .sheet header {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 12px;
  }
  .mono {
    width: 44px;
    height: 44px;
    border-radius: 14px;
    display: grid;
    place-items: center;
    flex: none;
    font-family: var(--md-font-display);
    font-size: 18px;
    font-weight: 700;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
  }
  .copy {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .sheet strong {
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
  .reason {
    margin: 0 0 10px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-soft);
  }
  .payload {
    margin: 0 0 16px;
    padding: 10px 12px;
    border-radius: 12px;
    background: var(--md-stage);
    border: 1px solid var(--md-line);
    font-family: var(--md-font-mono);
    font-size: 12px;
    color: var(--md-ink-mute);
    word-break: break-all;
  }
  .payload span {
    display: block;
    font-size: 9px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin-bottom: 4px;
  }
  footer {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .settled ul {
    margin: 0;
    padding: 0;
    display: grid;
    gap: 8px;
  }
  .settled li {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    border-radius: 12px;
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
    border: 1px solid var(--md-line);
  }
  .settled .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .settled li[data-status='approved'] .dot,
  .settled li[data-status='executed'] .dot {
    background: var(--md-live);
  }
  .settled li[data-status='denied'] .dot,
  .settled li[data-status='failed'] .dot {
    background: var(--md-halt);
  }
  .settled strong {
    font-size: 13px;
    font-weight: 600;
  }

  @keyframes md-pulse-soft {
    0%,
    100% {
      box-shadow: 0 0 0 0 color-mix(in oklab, var(--md-cobalt) 0%, transparent);
    }
    50% {
      box-shadow: 0 0 0 10px color-mix(in oklab, var(--md-cobalt) 12%, transparent);
    }
  }

  @media (max-width: 640px) {
    .pulse {
      grid-template-columns: 1fr;
      justify-items: start;
    }
    footer :global(.md-btn) {
      flex: 1;
      justify-content: center;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .pulse.hot .pulse-ring {
      animation: none;
    }
  }
</style>
