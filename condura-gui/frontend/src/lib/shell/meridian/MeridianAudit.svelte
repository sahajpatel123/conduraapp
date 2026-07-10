<script lang="ts">
  /**
   * Audit — append-only chain instrument with seal + detail stage.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { audit } from '../../stores/audit.svelte'
  import type { AuditEvent } from '../../ipc/types'

  let selectedId = $state<number | null>(null)
  let filter = $state<'all' | 'allow' | 'block' | 'prompt'>('all')

  onMount(() => {
    void audit.refresh()
    audit.startLive()
    return () => audit.stopLive()
  })

  const filtered = $derived(
    filter === 'all' ? audit.events : audit.events.filter((e) => e.result === filter)
  )
  const selected = $derived(
    filtered.find((e) => e.id === selectedId) ?? filtered[0] ?? null
  )

  $effect(() => {
    if (!filtered.length) {
      selectedId = null
      return
    }
    if (selectedId === null || !filtered.some((e) => e.id === selectedId)) {
      selectedId = filtered[0]!.id
    }
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

  function pick(ev: AuditEvent): void {
    selectedId = ev.id
  }
</script>

<MeridianPage
  kicker="Ledger"
  title="Audit"
  lead="Every gated action is written here. The chain is the truth — verify it, then read each seal."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void audit.refresh()}>Refresh</button>
    <button type="button" class="md-btn md-btn-primary" onclick={() => void audit.verifyIntegrity()}>
      Verify seal
    </button>
  {/snippet}

  {#if audit.integrity}
    <div class="seal" class:bad={!audit.integrity.ok} class:ok={audit.integrity.ok}>
      <span class="seal-mark" aria-hidden="true"></span>
      <div>
        <p class="cite">chain seal</p>
        <strong>Chain {audit.integrity.ok ? 'intact' : 'broken'}</strong>
        {#if audit.integrity.reason}
          <p class="reason">{audit.integrity.reason}</p>
        {/if}
      </div>
    </div>
  {/if}

  <div class="filters" role="group" aria-label="Filter by verdict">
    {#each (['all', 'allow', 'block', 'prompt'] as const) as f}
      <button type="button" class:on={filter === f} data-f={f} onclick={() => (filter = f)}>
        {f}
      </button>
    {/each}
  </div>

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
          When Condura acts, every verdict lands here as a sealed link.
        {/if}
      </p>
    </div>
  {:else if filtered.length === 0}
    <div class="md-empty empty">
      <p class="empty-title">No {filter} events</p>
      <p class="empty-lead">Try another verdict filter.</p>
    </div>
  {:else}
    <div class="layout md-stagger">
      <div class="chain">
        {#each filtered as ev (ev.id)}
          <button
            type="button"
            class="link"
            class:on={selected?.id === ev.id}
            onclick={() => pick(ev)}
          >
            <span class="tick" data-v={ev.result} aria-hidden="true"></span>
            <span class="link-copy">
              <strong>{ev.action || ev.message || 'event'}</strong>
              <span class="when">{formatWhen(ev.ts)}</span>
            </span>
            <span class="chip" data-v={ev.result}>{ev.result}</span>
          </button>
        {/each}
      </div>

      {#if selected}
        <article class="stage-plate">
          <p class="cite">event · {selected.id}</p>
          <h2>{selected.action || 'Event'}</h2>
          <p class="body">{selected.message || 'No message on this link.'}</p>
          <dl class="facts">
            <div>
              <dt>When</dt>
              <dd>{formatWhen(selected.ts)}</dd>
            </div>
            <div>
              <dt>Verdict</dt>
              <dd data-v={selected.result}>{selected.result}</dd>
            </div>
            <div>
              <dt>Actor</dt>
              <dd>{selected.actor || '—'}</dd>
            </div>
            <div>
              <dt>App</dt>
              <dd>{selected.app || selected.target_app || '—'}</dd>
            </div>
            {#if selected.command}
              <div class="wide">
                <dt>Command</dt>
                <dd class="mono">{selected.command}</dd>
              </div>
            {/if}
            {#if selected.this_hash}
              <div class="wide">
                <dt>Hash</dt>
                <dd class="mono">{selected.this_hash}</dd>
              </div>
            {/if}
          </dl>
        </article>
      {/if}
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
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 6px;
  }
  .seal {
    display: flex;
    gap: 14px;
    align-items: center;
    padding: 16px 18px;
    border-radius: 18px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    margin-bottom: 16px;
  }
  .seal.ok {
    border-color: color-mix(in oklab, var(--md-live) 35%, transparent);
  }
  .seal.bad {
    border-color: color-mix(in oklab, var(--md-halt) 40%, transparent);
    color: var(--md-halt);
  }
  .seal-mark {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    background: var(--md-ink-faint);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--md-ink-faint) 18%, transparent);
  }
  .seal.ok .seal-mark {
    background: var(--md-live);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--md-live) 18%, transparent);
  }
  .seal.bad .seal-mark {
    background: var(--md-halt);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--md-halt) 18%, transparent);
  }
  .seal strong {
    font-family: var(--md-font-display);
    font-size: 16px;
    letter-spacing: -0.03em;
  }
  .reason {
    margin: 4px 0 0;
    font-size: 13px;
    color: var(--md-ink-mute);
  }

  .filters {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 16px;
  }
  .filters button {
    padding: 7px 12px;
    border-radius: 999px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-stage);
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-ink-mute);
    cursor: pointer;
  }
  .filters button.on[data-f='all'] {
    background: var(--md-cobalt);
    border-color: var(--md-cobalt);
    color: #fff;
  }
  .filters button.on[data-f='allow'] {
    color: var(--md-live);
    border-color: color-mix(in oklab, var(--md-live) 40%, transparent);
    background: color-mix(in oklab, var(--md-live) 12%, transparent);
  }
  .filters button.on[data-f='block'] {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 40%, transparent);
    background: color-mix(in oklab, var(--md-halt) 12%, transparent);
  }
  .filters button.on[data-f='prompt'] {
    color: var(--md-cobalt);
    border-color: color-mix(in oklab, var(--md-cobalt) 40%, transparent);
    background: color-mix(in oklab, var(--md-cobalt) 12%, transparent);
  }
  .filters button:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }

  .empty-title {
    margin: 0;
    font-family: var(--md-font-display);
    font-size: 16px;
  }
  .empty-lead {
    margin: 6px 0 0;
    font-size: 13px;
    color: var(--md-ink-mute);
    max-width: 40ch;
  }

  .layout {
    display: grid;
    grid-template-columns: minmax(240px, 1fr) minmax(260px, 0.9fr);
    gap: 14px;
    align-items: start;
  }
  .chain {
    display: grid;
    gap: 0;
    position: relative;
    padding-left: 6px;
  }
  .link {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 12px;
    align-items: center;
    width: 100%;
    text-align: left;
    padding: 12px 12px 12px 8px;
    border: 0;
    border-bottom: 1px solid var(--md-line);
    background: transparent;
    cursor: pointer;
    color: inherit;
    position: relative;
  }
  .link:hover {
    background: color-mix(in oklab, var(--md-stage) 70%, transparent);
  }
  .link.on {
    background: var(--md-surface);
    border-radius: 14px;
    border-bottom-color: transparent;
    box-shadow: var(--md-shadow);
  }
  .link:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .tick {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--md-ink-faint);
    box-shadow: 0 0 0 3px var(--md-stage);
  }
  .tick[data-v='allow'] {
    background: var(--md-live);
  }
  .tick[data-v='block'] {
    background: var(--md-halt);
  }
  .tick[data-v='prompt'] {
    background: var(--md-cobalt);
  }
  .link-copy {
    min-width: 0;
    display: grid;
    gap: 3px;
  }
  .link-copy strong {
    font-size: 13px;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .when {
    font-family: var(--md-font-mono);
    font-size: 10px;
    color: var(--md-ink-faint);
  }
  .chip {
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 4px 8px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    color: var(--md-ink-mute);
  }
  .chip[data-v='allow'] {
    color: var(--md-live);
    border-color: color-mix(in oklab, var(--md-live) 30%, transparent);
  }
  .chip[data-v='block'] {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 30%, transparent);
  }
  .chip[data-v='prompt'] {
    color: var(--md-cobalt);
    border-color: color-mix(in oklab, var(--md-cobalt) 30%, transparent);
  }

  .stage-plate {
    border-radius: 22px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 22px;
    box-shadow: var(--md-shadow);
    position: sticky;
    top: 8px;
  }
  .stage-plate h2 {
    font-family: var(--md-font-display);
    font-size: 24px;
    letter-spacing: -0.04em;
    margin: 0 0 8px;
  }
  .body {
    margin: 0;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .facts {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
    margin: 18px 0 0;
    padding-top: 16px;
    border-top: 1px solid var(--md-line);
  }
  .facts .wide {
    grid-column: 1 / -1;
  }
  .facts dt {
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin-bottom: 4px;
  }
  .facts dd {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    color: var(--md-ink-soft);
    word-break: break-word;
  }
  .facts dd[data-v='allow'] {
    color: var(--md-live);
  }
  .facts dd[data-v='block'] {
    color: var(--md-halt);
  }
  .facts dd[data-v='prompt'] {
    color: var(--md-cobalt);
  }
  .facts dd.mono {
    font-family: var(--md-font-mono);
    font-size: 11px;
    font-weight: 500;
  }
  .pager {
    margin-top: 16px;
  }
  @media (max-width: 800px) {
    .layout {
      grid-template-columns: 1fr;
    }
    .stage-plate {
      position: static;
    }
  }
</style>
