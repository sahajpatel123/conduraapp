<script lang="ts">
  /**
   * Skills — local shelf with stage-plate inspect + provenance marks.
   * Signature: select a skill → living plate; Use in Ask / Remove wired.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { ipc } from '../../ipc/client'
  import type { InstalledSkill } from '../../ipc/types'

  let skills = $state<InstalledSkill[]>([])
  let loading = $state(true)
  let error = $state('')
  let removing = $state<string | null>(null)
  let activeId = $state<string | null>(null)

  const active = $derived(skills.find((s) => s.id === activeId) ?? skills[0] ?? null)

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
      if (skills.length && !skills.some((s) => s.id === activeId)) {
        activeId = skills[0]!.id
      }
    } catch (e) {
      skills = []
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
      if (activeId === id) activeId = null
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

  function fromHub(s: InstalledSkill): boolean {
    const src = (s.source || '').toLowerCase()
    return src.includes('hub') || src.includes('remote') || src.includes('http')
  }

  function goHub(): void {
    window.location.hash = '#/hub'
  }

  function useInAsk(s: InstalledSkill): void {
    const starter = `Use the local skill “${s.name}” — explain what it does, then wait for my go-ahead before acting.`
    try {
      sessionStorage.setItem('md-ask-starter', starter)
    } catch {
      /* ignore */
    }
    window.location.hash = '#/'
  }
</script>

<MeridianPage
  kicker="Shelf"
  title="Skills"
  lead="Procedures installed on this machine. Inspect, ask with them, or remove — nothing runs until you say so."
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
      <p class="empty-title">Shelf is empty</p>
      <p class="empty-lead">Install from Hub to keep procedures on this machine — they stay local after install.</p>
      <button type="button" class="md-btn md-btn-primary" onclick={goHub}>Open Hub</button>
    </div>
  {:else}
    <div class="layout md-stagger">
      <aside class="rail" aria-label="Installed skills">
        {#each skills as s (s.id)}
          <button
            type="button"
            class="rail-item"
            class:on={active?.id === s.id}
            onclick={() => (activeId = s.id)}
          >
            <span class="mono" class:hub={fromHub(s)} aria-hidden="true">{initial(s.name)}</span>
            <span class="rail-copy">
              <strong>{s.name}</strong>
              <span class="meta">
                {fromHub(s) ? 'from hub' : 'local'}
                {#if s.trust} · {s.trust}{/if}
              </span>
            </span>
          </button>
        {/each}
      </aside>

      {#if active}
        <section class="stage-plate">
          <header>
            <p class="cite">
              {fromHub(active) ? 'provenance · hub' : 'provenance · local'}
              {#if active.version} · v{active.version}{/if}
            </p>
            <h2>{active.name}</h2>
            <p class="body">{active.description || 'No description on file.'}</p>
          </header>
          <dl class="facts">
            <div>
              <dt>Author</dt>
              <dd>{active.author || '—'}</dd>
            </div>
            <div>
              <dt>Trust</dt>
              <dd data-trust={active.trust || 'unknown'}>{active.trust || 'unmarked'}</dd>
            </div>
            <div>
              <dt>Source</dt>
              <dd>{active.source || 'local shelf'}</dd>
            </div>
          </dl>
          <footer>
            <button type="button" class="md-btn md-btn-primary" onclick={() => useInAsk(active)}>
              Use in Ask
            </button>
            <button
              type="button"
              class="md-btn md-btn-danger"
              disabled={removing === active.id}
              onclick={() => void remove(active.id)}
            >
              {removing === active.id ? 'Removing…' : 'Remove'}
            </button>
          </footer>
        </section>
      {/if}
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
  }
  .empty-lead {
    margin: 0 0 8px;
    max-width: 40ch;
    font-size: 13px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .layout {
    display: grid;
    grid-template-columns: minmax(200px, 280px) 1fr;
    gap: 14px;
    align-items: start;
  }
  .rail {
    display: grid;
    gap: 6px;
    padding: 8px;
    border-radius: 20px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
  }
  .rail-item {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 10px;
    align-items: center;
    text-align: left;
    padding: 10px;
    border-radius: 14px;
    border: 1px solid transparent;
    cursor: pointer;
    color: inherit;
    transition:
      background 160ms var(--md-ease),
      border-color 160ms var(--md-ease);
  }
  .rail-item:hover {
    background: var(--md-stage);
  }
  .rail-item.on {
    background: var(--md-surface);
    border-color: color-mix(in oklab, var(--md-cobalt) 35%, transparent);
    box-shadow: var(--md-shadow);
  }
  .rail-item:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .mono {
    width: 36px;
    height: 36px;
    border-radius: 12px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 15px;
    font-weight: 700;
    color: var(--md-live);
    background: color-mix(in oklab, var(--md-live) 12%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-live) 18%, var(--md-line));
  }
  .mono.hub {
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-stage));
    border-color: color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
  }
  .rail-copy {
    min-width: 0;
    display: grid;
    gap: 2px;
  }
  .rail-copy strong {
    font-size: 13px;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .meta {
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .stage-plate {
    border-radius: 22px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 24px;
    box-shadow: var(--md-shadow);
    min-height: 280px;
    display: flex;
    flex-direction: column;
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 10px;
  }
  .stage-plate h2 {
    font-family: var(--md-font-display);
    font-size: clamp(26px, 4vw, 34px);
    letter-spacing: -0.045em;
    margin: 0 0 10px;
  }
  .body {
    margin: 0;
    font-size: 15px;
    line-height: 1.55;
    color: var(--md-ink-mute);
    max-width: 48ch;
  }
  .facts {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
    margin: 22px 0;
    padding: 16px 0;
    border-top: 1px solid var(--md-line);
    border-bottom: 1px solid var(--md-line);
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
  .facts dd[data-trust='verified'],
  .facts dd[data-trust='trusted'] {
    color: var(--md-live);
  }
  footer {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    margin-top: auto;
  }
  @media (max-width: 720px) {
    .layout {
      grid-template-columns: 1fr;
    }
    .rail {
      grid-auto-flow: column;
      grid-auto-columns: minmax(160px, 200px);
      overflow-x: auto;
    }
    .facts {
      grid-template-columns: 1fr;
    }
  }
</style>
