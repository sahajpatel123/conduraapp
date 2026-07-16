<script lang="ts">
  /**
   * Skills — local procedures shelf.
   * Signature: provenance rail + stage plate + Use in Ask (seeds composer).
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { ipc } from '../../ipc/client'
  import type { InstalledSkill } from '../../ipc/types'
  import { primarySlashToken } from '../../skill-slash'
  import { focusOn } from '../../a11y/autofocus'

  type Provenance = 'all' | 'local' | 'hub'

  let skills = $state<InstalledSkill[]>([])
  let loading = $state(true)
  let error = $state('')
  let removing = $state<string | null>(null)
  /** Two-step remove — never delete on first click. */
  let confirmRemoveId = $state<string | null>(null)
  let activeId = $state<string | null>(null)
  let provenance = $state<Provenance>('all')
  let offline = $state(false)
  let q = $state('')
  let searchEl = $state<HTMLInputElement | null>(null)

  // Add Skill form
  let showCreate = $state(false)
  let createName = $state('')
  let createDesc = $state('')
  let createSteps = $state('')
  let createBusy = $state(false)
  let createNote = $state('')

  // Filter by provenance, then by name (case-insensitive substring).
  // `q` matches the skill's primary display name + description so a
  // fuzzy "what was that image helper called?" still works.
  const filtered = $derived.by(() => {
    const needle = q.trim().toLowerCase()
    const base = provenance === 'all'
      ? skills
      : skills.filter((s) => (fromHub(s) ? 'hub' : 'local') === provenance)
    if (!needle) return base
    return base.filter((s) => {
      const name = (s.name || '').toLowerCase()
      const desc = (s.description || '').toLowerCase()
      return name.includes(needle) || desc.includes(needle)
    })
  })
  const active = $derived(filtered.find((s) => s.id === activeId) ?? filtered[0] ?? null)
  const hubCount = $derived(skills.filter(fromHub).length)
  const localCount = $derived(skills.length - hubCount)
  const createSlashPreview = $derived(
    createName.trim()
      ? primarySlashToken({
          id: '',
          name: createName.trim(),
          description: '',
          version: '',
          author: '',
          license: '',
          trust: '',
        })
      : '/…'
  )

  onMount(() => {
    try {
      if (sessionStorage.getItem('md-skills-open-create') === '1') {
        sessionStorage.removeItem('md-skills-open-create')
        openCreate()
      }
    } catch {
      /* ignore */
    }
    void load()
    // Drop focus into the search field so users with many skills can
    // start typing immediately — same auto-focus pattern as Sync PIN
    // and the Chat composer. Wrapped in setTimeout so we run AFTER
    // the create-form mount check above completes.
    setTimeout(() => focusOn(() => searchEl, () => !showCreate), 0)
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
      confirmRemoveId = null
      return
    }
    if (!filtered.some((s) => s.id === activeId)) {
      activeId = filtered[0]!.id
    }
    // Drop confirm if the staged skill left the shelf or is no longer selected.
    if (
      confirmRemoveId &&
      (!skills.some((s) => s.id === confirmRemoveId) || confirmRemoveId !== activeId)
    ) {
      confirmRemoveId = null
    }
  })

  function requestRemove(id: string): void {
    confirmRemoveId = id
    error = ''
  }

  function cancelRemove(): void {
    if (removing) return
    confirmRemoveId = null
  }

  async function confirmRemove(): Promise<void> {
    const id = confirmRemoveId
    if (!id || removing) return
    removing = id
    try {
      await ipc.skillsDelete(id)
      confirmRemoveId = null
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
    const token = primarySlashToken(s)
    const starter = `${token} `
    try {
      sessionStorage.setItem('md-ask-starter', starter)
    } catch {
      /* ignore */
    }
    window.location.hash = '#/'
  }

  function openCreate(): void {
    showCreate = true
    createNote = ''
    createName = ''
    createDesc = ''
    createSteps = ''
  }

  function cancelCreate(): void {
    if (createBusy) return
    showCreate = false
    createNote = ''
  }

  async function submitCreate(): Promise<void> {
    const name = createName.trim()
    if (!name || createBusy) return
    createBusy = true
    createNote = ''
    error = ''
    try {
      const steps = createSteps
        .split('\n')
        .map((s) => s.trim())
        .filter(Boolean)
      const sk = await ipc.skillsCreate({
        name,
        description: createDesc.trim(),
        steps,
      })
      showCreate = false
      provenance = 'local'
      await load()
      activeId = sk.id
      createNote = ''
    } catch (e) {
      createNote = String(e)
    } finally {
      createBusy = false
    }
  }

  function onKey(e: KeyboardEvent): void {
    if (!filtered.length) return
    const t = e.target as HTMLElement | null
    if (
      t &&
      (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.isContentEditable)
    ) {
      return
    }
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
    if (e.key === 'Escape' && confirmRemoveId) {
      e.preventDefault()
      cancelRemove()
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<MeridianPage
  kicker="Shelf · this machine"
  title="Skills"
  lead="Author local procedures, install from Hub, then invoke them in Ask with /Name — Gatekeeper still gates every action."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void load()}>Refresh</button>
    <button type="button" class="md-btn md-btn-ghost" onclick={goHub}>Browse Hub</button>
    <button type="button" class="md-btn md-btn-primary" onclick={openCreate}>Add skill</button>
  {/snippet}

  <div class="desk md-stagger">
    <p class="contract">
      <span class="live-dot" aria-hidden="true"></span>
      Skills live on this machine. Use in Ask seeds <code>/Name</code> in the composer — nothing runs until you send.
    </p>

    {#if showCreate}
      <div class="create-plate" role="dialog" aria-label="Add skill">
        <header class="create-head">
          <h3>Add a local skill</h3>
          <p class="hint">
            Name it something callable — e.g. <strong>UI</strong> → invoke with
            <code>{createSlashPreview}</code> in Ask.
          </p>
        </header>
        <label class="field">
          <span>Name</span>
          <input bind:value={createName} placeholder="UI" maxlength="64" />
        </label>
        <label class="field">
          <span>Description</span>
          <input bind:value={createDesc} placeholder="What this skill is for" maxlength="240" />
        </label>
        <label class="field">
          <span>Steps (one per line, optional)</span>
          <textarea
            bind:value={createSteps}
            rows="4"
            placeholder={"Inspect the surface\nPropose a change\nWait for consent"}
          ></textarea>
        </label>
        {#if createNote}
          <p class="create-err">{createNote}</p>
        {/if}
        <div class="create-actions">
          <button type="button" class="md-btn md-btn-ghost" disabled={createBusy} onclick={cancelCreate}>
            Cancel
          </button>
          <button
            type="button"
            class="md-btn md-btn-primary"
            disabled={createBusy || !createName.trim()}
            onclick={() => void submitCreate()}
          >
            {createBusy ? 'Saving…' : 'Create skill'}
          </button>
        </div>
      </div>
    {/if}

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
            Add a local skill or install from Hub — invoke in Ask with /Name.
          {/if}
        </p>
        <div class="empty-actions">
          {#if !offline}
            <button type="button" class="md-btn md-btn-primary" onclick={openCreate}>Add skill</button>
          {/if}
          <button type="button" class="md-btn md-btn-ghost" onclick={goHub}>Open Hub</button>
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

      <label class="search">
        <span class="sr">Search installed skills</span>
        <input
          bind:this={searchEl}
          bind:value={q}
          type="search"
          placeholder="Search skills by name…"
          aria-label="Search skills"
        />
        {#if q}
          <button type="button" class="clear" onclick={() => (q = '')} aria-label="Clear search">Clear</button>
        {/if}
      </label>

      {#if q.trim()}
        <p class="result-count" aria-live="polite">
          {filtered.length} skill{filtered.length === 1 ? '' : 's'} for “{q.trim()}”
        </p>
      {/if}

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
                onclick={() => {
                  activeId = s.id
                  if (confirmRemoveId && confirmRemoveId !== s.id) confirmRemoveId = null
                }}
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
                <p class="slash-call">
                  Ask with <code>{primarySlashToken(active)}</code>
                </p>
                <p class="body">{active.description || 'No description on file.'}</p>
              </header>
              <dl class="facts">
                <div>
                  <dt>Slash</dt>
                  <dd><code>{primarySlashToken(active)}</code></dd>
                </div>
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
                <span class="how-n">01</span> Use in Ask seeds <code>{primarySlashToken(active)}</code> ·
                <span class="how-n">02</span> Condura plans ·
                <span class="how-n">03</span> You consent
              </p>
              {#if confirmRemoveId === active.id}
                <div class="remove-plate" role="alertdialog" aria-labelledby="skill-remove-title">
                  <p class="cite">remove from shelf</p>
                  <h3 id="skill-remove-title">Remove “{active.name}”?</h3>
                  <p class="remove-lead">
                    Uninstalls from this machine. Nothing runs from the shelf alone —
                    reinstall from Hub if you need it again.
                  </p>
                  <div class="remove-actions">
                    <button
                      type="button"
                      class="md-btn md-btn-danger"
                      disabled={!!removing}
                      onclick={() => void confirmRemove()}
                    >
                      {removing === active.id ? 'Removing…' : 'Remove skill'}
                    </button>
                    <button
                      type="button"
                      class="md-btn md-btn-ghost"
                      disabled={!!removing}
                      onclick={cancelRemove}
                    >
                      Keep
                    </button>
                  </div>
                </div>
              {:else}
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
                    disabled={!!removing}
                    onclick={() => requestRemove(active.id)}
                  >
                    Remove
                  </button>
                </footer>
              {/if}
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
    box-shadow: none;
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
    border-radius: 12px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 22px 20px;
    box-shadow: none;
    max-width: 520px;
  }
  .empty-atlas h2 {
    font-family: var(--md-font-display);
    font-size: 22px;
    font-weight: 650;
    letter-spacing: -0.03em;
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
    gap: 6px;
  }
  .filters button {
    padding: 6px 10px;
    border-radius: 7px;
    border: 1px solid var(--md-line);
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
  .search {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    border-radius: 10px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
    margin: 0 0 14px;
    transition: border-color 140ms var(--md-ease);
  }
  .search:focus-within {
    border-color: color-mix(in oklab, var(--md-cobalt) 40%, var(--md-line));
  }
  .search input {
    flex: 1;
    border: 0;
    background: transparent;
    color: var(--md-ink);
    font: inherit;
    outline: none;
    padding: 4px 0;
  }
  .search input::placeholder {
    color: var(--md-ink-faint);
  }
  .result-count {
    margin: 0 0 12px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.02em;
    color: var(--md-ink-faint);
  }
  .search .clear {
    appearance: none;
    border: 0;
    background: transparent;
    color: var(--md-ink-faint);
    font: inherit;
    font-size: 11px;
    letter-spacing: 0.04em;
    cursor: pointer;
    padding: 2px 6px;
    border-radius: 6px;
  }
  .search .clear:hover {
    color: var(--md-ink);
    background: color-mix(in oklab, var(--md-ink) 4%, transparent);
  }
  .sr {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
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
    border-radius: 12px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
  }
  .rail-item {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 10px;
    align-items: center;
    text-align: left;
    padding: 9px 10px;
    border-radius: 8px;
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
    border-color: color-mix(in oklab, var(--md-cobalt) 28%, transparent);
    box-shadow: none;
  }
  .rail-item:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .mono {
    width: 34px;
    height: 34px;
    border-radius: 8px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 14px;
    font-weight: 650;
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
    font-weight: 600;
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
    border-radius: 12px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 18px 18px 20px;
    box-shadow: none;
    min-height: 300px;
    display: flex;
    flex-direction: column;
  }
  .stage-plate h2 {
    font-family: var(--md-font-display);
    font-size: clamp(22px, 3.6vw, 28px);
    font-weight: 650;
    letter-spacing: -0.035em;
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
  .remove-plate {
    margin-top: auto;
    padding: 14px 14px;
    border-radius: 10px;
    border: 1px solid color-mix(in oklab, var(--md-halt) 22%, var(--md-line));
    background: color-mix(in oklab, var(--md-halt) 4%, var(--md-stage));
  }
  .remove-plate h3 {
    font-family: var(--md-font-display);
    font-size: 18px;
    letter-spacing: -0.03em;
    margin: 0 0 8px;
  }
  .remove-lead {
    margin: 0 0 14px;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
    max-width: 42ch;
  }
  .remove-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .create-plate {
    border-radius: 12px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 16px 14px;
    box-shadow: none;
    margin-bottom: 16px;
    max-width: 520px;
    display: grid;
    gap: 12px;
  }
  .create-head h3 {
    font-family: var(--md-font-display);
    font-size: 20px;
    letter-spacing: -0.03em;
    margin: 0 0 6px;
  }
  .create-head .hint {
    margin: 0;
    font-size: 13px;
    color: var(--md-ink-mute);
    line-height: 1.45;
  }
  .field {
    display: grid;
    gap: 4px;
    font-size: 12px;
    font-weight: 600;
    color: var(--md-ink-mute);
  }
  .field input,
  .field textarea {
    padding: 10px 12px;
    border-radius: 12px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-stage);
    font-size: 14px;
    font-weight: 500;
    color: var(--md-ink);
    font-family: inherit;
  }
  .create-err {
    margin: 0;
    color: var(--md-halt);
    font-size: 13px;
  }
  .create-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    justify-content: flex-end;
  }
  .slash-call {
    margin: 4px 0 8px;
    font-family: var(--md-font-mono);
    font-size: 12px;
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
