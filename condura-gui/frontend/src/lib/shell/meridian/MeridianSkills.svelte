<script lang="ts">
  /**
   * Skills — local procedures shelf.
   * Signature: provenance rail + stage plate + Use in Ask (seeds composer).
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { ipc } from '../../ipc/client'
  import type { InstalledSkill } from '../../ipc/types'

  type Provenance = 'all' | 'local' | 'hub'

  let skills = $state<InstalledSkill[]>([])
  let loading = $state(true)
  let error = $state('')
  let removing = $state<string | null>(null)
  let activeId = $state<string | null>(null)
  let provenance = $state<Provenance>('all')
  let offline = $state(false)

  const filtered = $derived(
    provenance === 'all'
      ? skills
      : skills.filter((s) => (fromHub(s) ? 'hub' : 'local') === provenance)
  )
  const active = $derived(filtered.find((s) => s.id === activeId) ?? filtered[0] ?? null)
  const hubCount = $derived(skills.filter(fromHub).length)
  const localCount = $derived(skills.length - hubCount)

  onMount(() => {
    void load()
  })

  function isOfflineError(e: unknown): boolean {
    const s = String(e)
    return /IPC client not started|not connected|Failed to fetch|daemon/i.test(s)
  }

  function fromHub(s: InstalledSkill): boolean {
    const src = (s.source || '').toLowerCase()
    return !!s.hub_id || src.includes('hub') || src.includes('remote') || src.includes('http')
  }

  async function load(): Promise<void> {
    loading = true
    error = ''
    offline = false
    try {
      skills = await ipc.skillsList(100)
    } catch (e) {
      skills = []
      if (isOfflineError(e)) offline = true
      else error = String(e)
    } finally {
      loading = false
    }
  }

  $effect(() => {
    if (!filtered.length) {
      activeId = null
      return
    }
    if (!filtered.some((s) => s.id === activeId)) {
      activeId = filtered[0]!.id
    }
  })

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

  function onKey(e: KeyboardEvent): void {
    if (!filtered.length) return
    const i = filtered.findIndex((s) => s.id === activeId)
    if (e.key === 'ArrowDown' || e.key === 'j' || e.key === 'J') {
      e.preventDefault()
      const next = filtered[Math.min(filtered.length - 1, Math.max(0, i) + 1)]
      if (next) activeId = next.id
    }
    if (e.key === 'ArrowUp' || e.key === 'k' || e.key === 'K') {
      e.preventDefault()
      const prev = filtered[Math.max(0, (i < 0 ? 0 : i) - 1)]
      if (prev) activeId = prev.id
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<MeridianPage
  kicker="Shelf · this machine"
  title="Skills"
  lead="Procedures installed locally. Inspect provenance, seed Ask, or remove — nothing runs until you open the door."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void load()}>Refresh</button>
    <button type="button" class="md-btn md-btn-primary" onclick={goHub}>Browse Hub</button>
  {/snippet}

  <div class="desk md-stagger">
    <p class="contract">
      <span class="live-dot" aria-hidden="true"></span>
      Skills live on disk. Use in Ask seeds the composer — the Gatekeeper still gates every action.
    </p>

    {#if loading}
      <div class="md-empty">Indexing local skills…</div>
    {:else if error}
      <div class="md-empty">{error}</div>
    {:else if skills.length === 0}
      <div class="empty-atlas">
        <p class="cite">{offline ? 'daemon offline' : 'empty shelf'}</p>
        <h2>{offline ? 'Shelf unread' : 'Nothing installed yet'}</h2>
        <p class="empty-lead">
          {#if offline}
            Connect the daemon to index local skills. You can still browse Hub once Condura is online.
          {:else}
            Install from Hub to keep procedures on this machine — they stay local after install.
          {/if}
        </p>
        <div class="empty-actions">
          <button type="button" class="md-btn md-btn-primary" onclick={goHub}>Open Hub</button>
          <button type="button" class="md-btn md-btn-ghost" onclick={() => void load()}>Try again</button>
        </div>
      </div>
    {:else}
      <div class="filters" role="group" aria-label="Filter by provenance">
        <button type="button" class:on={provenance === 'all'} onclick={() => (provenance = 'all')}>
          All · {skills.length}
        </button>
        <button type="button" class:on={provenance === 'local'} onclick={() => (provenance = 'local')}>
          Local · {localCount}
        </button>
        <button type="button" class:on={provenance === 'hub'} onclick={() => (provenance = 'hub')}>
          From Hub · {hubCount}
        </button>
      </div>

      {#if filtered.length === 0}
        <div class="md-empty empty">
          <p class="empty-title">No {provenance} skills</p>
          <p class="empty-lead">Try another provenance filter.</p>
          <button type="button" class="md-btn md-btn-ghost" onclick={() => (provenance = 'all')}>Show all</button>
        </div>
      {:else}
        <div class="layout">
          <aside class="rail" aria-label="Installed skills">
            <p class="cite">↑↓ / J K to walk</p>
            {#each filtered as s (s.id)}
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
                  <dt>License</dt>
                  <dd>{active.license || '—'}</dd>
                </div>
                <div class="wide">
                  <dt>Source</dt>
                  <dd>{active.source || (fromHub(active) ? 'hub install' : 'local shelf')}</dd>
                </div>
              </dl>
              <p class="how">
                <span class="how-n">01</span> Use in Ask seeds a line ·
                <span class="how-n">02</span> Condura plans ·
                <span class="how-n">03</span> You consent
              </p>
              <footer>
                <button type="button" class="md-btn md-btn-primary" onclick={() => useInAsk(active)}>
                  Use in Ask
                </button>
                {#if fromHub(active)}
                  <button type="button" class="md-btn md-btn-ghost" onclick={goHub}>Back to Hub</button>
                {/if}
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
    {/if}
  </div>
</MeridianPage>

<style>
  .desk {
    display: grid;
    gap: 16px;
  }
  .contract {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    margin: 0;
    padding: 12px 14px;
    border-radius: 14px;
    border: 1px solid color-mix(in oklab, var(--md-live) 22%, var(--md-line));
    background: color-mix(in oklab, var(--md-live) 6%, var(--md-surface));
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .live-dot {
    width: 8px;
    height: 8px;
    margin-top: 5px;
    flex: none;
    border-radius: 50%;
    background: var(--md-live);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-live) 16%, transparent);
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 10px;
  }
  .empty-atlas {
    border-radius: 22px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 28px 24px;
    box-shadow: var(--md-shadow);
    max-width: 520px;
  }
  .empty-atlas h2 {
    font-family: var(--md-font-display);
    font-size: 26px;
    letter-spacing: -0.04em;
    margin: 0 0 8px;
  }
  .empty-lead {
    margin: 0 0 16px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
    max-width: 42ch;
  }
  .empty-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .empty-title {
    margin: 0;
    font-family: var(--md-font-display);
    font-size: 16px;
  }
  .filters {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .filters button {
    padding: 7px 12px;
    border-radius: 999px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-stage);
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--md-ink-mute);
    cursor: pointer;
  }
  .filters button.on {
    background: var(--md-cobalt);
    border-color: var(--md-cobalt);
    color: #fff;
  }
  .filters button:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
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
    padding: 10px 8px 8px;
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
    min-height: 300px;
    display: flex;
    flex-direction: column;
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
    margin: 22px 0 14px;
    padding: 16px 0;
    border-top: 1px solid var(--md-line);
    border-bottom: 1px solid var(--md-line);
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
  .facts dd[data-trust='verified'],
  .facts dd[data-trust='trusted'] {
    color: var(--md-live);
  }
  .how {
    margin: 0 0 18px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--md-ink-faint);
    line-height: 1.5;
  }
  .how-n {
    color: var(--md-cobalt);
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
