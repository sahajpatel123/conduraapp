<script lang="ts">
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { ipc } from '../../ipc/client'
  import type { InstalledSkill } from '../../ipc/types'

  let skills = $state<InstalledSkill[]>([])
  let loading = $state(true)
  let error = $state('')
  let removing = $state<string | null>(null)

  onMount(() => {
    void load()
  })

  function isOfflineError(e: unknown): boolean {
    const s = String(e)
    return /IPC client not started|not connected|Failed to fetch|daemon/i.test(s)
  }

  async function load(): Promise<void> {
    loading = true
    error = ''
    try {
      skills = await ipc.skillsList(100)
    } catch (e) {
      skills = []
      // Preview / offline: show the empty shelf, not a raw IPC stack string
      error = isOfflineError(e) ? '' : String(e)
    } finally {
      loading = false
    }
  }

  async function remove(id: string): Promise<void> {
    if (removing) return
    removing = id
    try {
      await ipc.skillsDelete(id)
      await load()
    } catch (e) {
      error = String(e)
    } finally {
      removing = null
    }
  }

  function initial(name: string): string {
    const t = name.trim()
    return t ? t[0]!.toUpperCase() : '?'
  }

  function goHub(): void {
    window.location.hash = '#/hub'
  }
</script>

<MeridianPage
  kicker="Local"
  title="Skills"
  lead="Procedures installed on this machine. Yours to run, improve, or remove."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void load()}>Refresh</button>
    <button type="button" class="md-btn md-btn-primary" onclick={goHub}>Browse Hub</button>
  {/snippet}

  {#if loading}
    <div class="md-empty">Indexing local skills…</div>
  {:else if error}
    <div class="md-empty">{error}</div>
  {:else if skills.length === 0}
    <div class="md-empty empty">
      <p class="empty-title">No local skills yet</p>
      <p class="empty-lead">Install from Hub to keep procedures on this machine — nothing runs until you ask.</p>
      <button type="button" class="md-btn md-btn-primary" onclick={goHub}>Open Hub</button>
    </div>
  {:else}
    <div class="list md-stagger">
      {#each skills as s (s.id)}
        <div class="item">
          <div class="mono" aria-hidden="true">{initial(s.name)}</div>
          <div class="copy">
            <div class="title-row">
              <strong>{s.name}</strong>
              {#if s.version}<span class="ver">v{s.version}</span>{/if}
            </div>
            <p>{s.description || 'No description.'}</p>
            <span class="meta">
              {s.author || 'local'}
              {#if s.trust} · {s.trust}{/if}
              {#if s.source} · {s.source}{/if}
            </span>
          </div>
          <button
            type="button"
            class="md-btn md-btn-danger"
            disabled={removing === s.id}
            onclick={() => void remove(s.id)}
          >
            {removing === s.id ? 'Removing…' : 'Remove'}
          </button>
        </div>
      {/each}
    </div>
  {/if}
</MeridianPage>

<style>
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }
  .empty-title {
    margin: 0;
    font-family: var(--md-font-display);
    font-size: 18px;
    letter-spacing: -0.03em;
    color: var(--md-ink);
  }
  .empty-lead {
    margin: 0 0 8px;
    max-width: 40ch;
    font-size: 13px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .list {
    background: var(--md-surface);
    border: 1px solid var(--md-line-strong);
    border-radius: 22px;
    padding: 8px;
    box-shadow: var(--md-shadow);
  }
  .item {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 14px;
    align-items: center;
    margin: 0;
    padding: 14px 12px;
    border-radius: 16px;
    transition: background 160ms var(--md-ease);
  }
  .item:hover {
    background: var(--md-stage);
  }
  .item:focus-within {
    background: var(--md-stage);
    box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--md-cobalt) 35%, transparent);
  }
  .mono {
    width: 42px;
    height: 42px;
    border-radius: 14px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -0.04em;
    color: var(--md-live);
    background: color-mix(in oklab, var(--md-live) 12%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-live) 18%, var(--md-line));
  }
  .copy {
    min-width: 0;
  }
  .title-row {
    display: flex;
    align-items: baseline;
    gap: 8px;
    flex-wrap: wrap;
  }
  .item strong {
    font-family: var(--md-font-display);
    font-size: 18px;
    letter-spacing: -0.03em;
    font-weight: 700;
  }
  .ver {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .item p {
    margin: 4px 0 0;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .meta {
    display: block;
    margin-top: 6px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  @media (max-width: 560px) {
    .empty {
      gap: 6px;
      padding: 36px 16px;
    }
    .empty-title {
      font-size: 16px;
    }
    .empty-lead {
      margin-bottom: 10px;
      font-size: 12.5px;
      max-width: 32ch;
    }
    .empty :global(.md-btn) {
      width: min(100%, 220px);
    }
    .item {
      grid-template-columns: auto 1fr;
      grid-template-rows: auto auto;
      gap: 10px;
      padding: 12px 10px;
    }
    .item .md-btn,
    .item button {
      grid-column: 1 / -1;
      justify-self: stretch;
    }
    .mono {
      width: 38px;
      height: 38px;
      font-size: 16px;
    }
  }
</style>
