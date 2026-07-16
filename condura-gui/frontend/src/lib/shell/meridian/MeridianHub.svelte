<script lang="ts">
  /**
   * Hub — community shelf. Install stays local; nothing runs until Ask.
   * Signature: featured stage plate + denser catalog + trust filters.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { isOfflineError } from '../../ipc/errors'
  import { hub } from '../../stores/hub.svelte'
  import { primarySlashToken } from '../../skill-slash'

  type TrustFilter = 'all' | 'verified' | 'trusted' | 'community'

  let q = $state('')
  let installing = $state<string | null>(null)
  let note = $state('')
  let featuredId = $state<string | null>(null)
  let trustFilter = $state<TrustFilter>('all')

  onMount(() => {
    void Promise.all([hub.refreshInstalled(), hub.search('', 24)])
  })

  const filtered = $derived(
    trustFilter === 'all'
      ? hub.results
      : hub.results.filter((s) => {
          const t = (s.trust || 'community').toLowerCase()
          if (trustFilter === 'community') return t === 'community' || t === 'unmarked' || !s.trust
          return t === trustFilter
        })
  )

  const featured = $derived(
    filtered.find((s) => s.id === featuredId) ?? filtered[0] ?? null
  )
  const rest = $derived(featured ? filtered.filter((s) => s.id !== featured.id) : filtered)
  const installedCount = $derived(hub.installed.size)

  /** Latest debounce handle so a fast typist doesn't fire multiple
   *  in-flight requests. Resets on every keystroke. */
  let searchTimer: ReturnType<typeof setTimeout> | null = null

  function search(): void {
    note = ''
    void hub.search(q.trim(), 24)
  }

  /** Debounced live search — fires 300ms after the user stops typing.
   *  Power users expect search-as-you-type; the daemon prefers not to
   *  receive a request per keystroke. Calls search() directly so the
   *  debounced path uses the same store wiring as Enter. */
  $effect(() => {
    // Track q reactively so this effect re-runs on every keystroke.
    const qNow = q
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(() => {
      search()
    }, 300)
    return () => {
      if (searchTimer) clearTimeout(searchTimer)
    }
  })

  function clearSearch(): void {
    q = ''
    note = ''
    void hub.search('', 24)
  }

  function refreshShelf(): void {
    note = ''
    void Promise.all([hub.refreshInstalled(), hub.search(q.trim(), 24)])
  }

  function initial(name: string): string {
    const t = name.trim()
    return t ? t[0]!.toUpperCase() : '?'
  }

  async function install(id: string): Promise<void> {
    if (hub.installed.has(id) || installing) return
    installing = id
    note = ''
    try {
      await hub.install(id)
      note = 'Installed on this machine — open Skills to inspect, or Ask to use it.'
    } finally {
      installing = null
    }
  }

  function goSkills(): void {
    window.location.hash = '#/skills'
  }

  function useInAsk(name: string, id?: string): void {
    const token = primarySlashToken({
      id: id || '',
      name,
      description: '',
      version: '',
      author: '',
      license: '',
      trust: '',
    })
    try {
      sessionStorage.setItem('md-ask-starter', `${token} `)
    } catch {
      /* ignore */
    }
    window.location.hash = '#/'
  }

  function goAddSkill(): void {
    window.location.hash = '#/skills'
    try {
      sessionStorage.setItem('md-skills-open-create', '1')
    } catch {
      /* ignore */
    }
  }
</script>

<MeridianPage
  kicker="Shelf · community"
  title="Hub"
  lead="Browse skills from the community. Install stays local — nothing runs until you ask, and the Gatekeeper still decides."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={refreshShelf}>Refresh</button>
    <button type="button" class="md-btn md-btn-ghost" onclick={goAddSkill}>Add local skill</button>
    <button type="button" class="md-btn md-btn-primary" onclick={goSkills}>
      My Skills{#if installedCount > 0}&nbsp;·&nbsp;{installedCount}{/if}
    </button>
  {/snippet}

  <div class="desk md-stagger">
    <p class="contract">
      <span class="live-dot" aria-hidden="true"></span>
      Install copies a procedure onto this machine. Running it still requires Ask → plan → consent.
    </p>

    <div class="search">
      <div class="field">
        <svg
          class="icon"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          aria-hidden="true"
        >
          <circle cx="11" cy="11" r="7" />
          <path d="M20 20l-3.5-3.5" />
        </svg>
        <input
          type="search"
          placeholder="Search the shelf…"
          bind:value={q}
          onkeydown={(e) => e.key === 'Enter' && search()}
          aria-label="Search skills"
        />
        {#if q}
          <button type="button" class="clear" onclick={clearSearch} aria-label="Clear search">Clear</button>
        {/if}
      </div>
      <button type="button" class="md-btn md-btn-primary" onclick={search}>Search</button>
    </div>

    <div class="filters" role="group" aria-label="Filter by trust">
      {#each (['all', 'verified', 'trusted', 'community'] as const) as f}
        <button type="button" class:on={trustFilter === f} data-f={f} onclick={() => (trustFilter = f)}>
          {f}
        </button>
      {/each}
    </div>

    {#if q.trim()}
      <p class="result-count" aria-live="polite">
        {#if hub.loading}
          Searching…
        {:else}
          {hub.results.length} result{hub.results.length === 1 ? '' : 's'} for “{q.trim()}”
        {/if}
      </p>
    {/if}

    {#if note}
      <p class="install-note">
        {note}
        <button type="button" class="linkish" onclick={goSkills}>Open Skills →</button>
      </p>
    {/if}

    {#if hub.loading}
      <div class="md-empty">Loading the shelf…</div>
    {:else if hub.error && !isOfflineError(hub.error)}
      <div class="md-empty empty">
        <p class="empty-title">Shelf unavailable</p>
        <p class="empty-lead">{hub.error}</p>
        <button type="button" class="md-btn md-btn-ghost" onclick={refreshShelf}>Try again</button>
      </div>
    {:else if hub.results.length === 0}
      <div class="md-empty empty">
        <p class="empty-title">{q ? 'No skills matched' : 'Shelf is quiet'}</p>
        <p class="empty-lead">
          {#if q}
            Try a broader term, or clear the search to browse the shelf.
          {:else if isOfflineError(hub.error ?? '')}
            Connect the daemon to load community skills. Until then, the shelf waits.
          {:else}
            No community skills returned yet. Refresh, or open local Skills for what is already on this machine.
          {/if}
        </p>
        {#if q}
          <button type="button" class="md-btn md-btn-ghost" onclick={clearSearch}>Clear search</button>
        {:else}
          <div class="empty-actions">
            <button type="button" class="md-btn md-btn-ghost" onclick={refreshShelf}>Refresh shelf</button>
            <button type="button" class="md-btn md-btn-ghost" onclick={goSkills}>Open local Skills</button>
          </div>
        {/if}
      </div>
    {:else if filtered.length === 0}
      <div class="md-empty empty">
        <p class="empty-title">No {trustFilter} skills here</p>
        <p class="empty-lead">Try another trust filter, or clear it to see the full shelf.</p>
        <button type="button" class="md-btn md-btn-ghost" onclick={() => (trustFilter = 'all')}>Show all</button>
      </div>
    {:else}
      <div class="shelf">
        {#if featured}
          <article class="feature">
            <p class="cite">featured on the shelf</p>
            <div class="feature-top">
              <div class="mono lg" aria-hidden="true">{initial(featured.name)}</div>
              <div class="tags">
                {#if featured.trust}
                  <span class="tag" data-trust={featured.trust}>{featured.trust}</span>
                {:else}
                  <span class="tag quiet">community</span>
                {/if}
                {#if featured.version}
                  <span class="tag quiet">v{featured.version}</span>
                {/if}
                {#if hub.installed.has(featured.id)}
                  <span class="tag on-machine">on this machine</span>
                {/if}
              </div>
            </div>
            <h2>{featured.name}</h2>
            <p>{featured.description || 'No description on the shelf.'}</p>
            <dl class="facts">
              <div>
                <dt>Author</dt>
                <dd>{featured.author || 'community'}</dd>
              </div>
              <div>
                <dt>Installs</dt>
                <dd>
                  {featured.downloads != null && featured.downloads > 0
                    ? featured.downloads.toLocaleString()
                    : '—'}
                </dd>
              </div>
              <div>
                <dt>After install</dt>
                <dd>local only</dd>
              </div>
            </dl>
            <footer>
              <span class="meta">Nothing runs from Hub itself — Ask is the door.</span>
              <div class="feature-actions">
                {#if hub.installed.has(featured.id)}
                  <button type="button" class="md-btn md-btn-ghost" onclick={() => useInAsk(featured.name, featured.id)}>
                    Use in Ask
                  </button>
                  <button type="button" class="md-btn md-btn-primary" onclick={goSkills}>Open in Skills</button>
                {:else}
                  <button
                    type="button"
                    class="md-btn md-btn-primary"
                    disabled={installing === featured.id}
                    onclick={() => void install(featured.id)}
                  >
                    {installing === featured.id ? 'Installing…' : 'Install locally'}
                  </button>
                {/if}
              </div>
            </footer>
          </article>
        {/if}

        {#if rest.length}
          <section class="catalog">
            <p class="cite">also on the shelf · {rest.length}</p>
            <div class="list">
              {#each rest as skill (skill.id)}
                <div class="row" class:installed={hub.installed.has(skill.id)}>
                  <button type="button" class="row-main" onclick={() => (featuredId = skill.id)}>
                    <div class="mono" aria-hidden="true">{initial(skill.name)}</div>
                    <div class="copy">
                      <strong>{skill.name}</strong>
                      <span class="meta">
                        {skill.author || 'community'}
                        {#if skill.trust} · {skill.trust}{/if}
                        {#if hub.installed.has(skill.id)} · on machine{/if}
                      </span>
                    </div>
                  </button>
                  <button
                    type="button"
                    class="md-btn md-btn-primary mini"
                    disabled={hub.installed.has(skill.id) || installing === skill.id}
                    onclick={() => void install(skill.id)}
                  >
                    {#if hub.installed.has(skill.id)}
                      In
                    {:else if installing === skill.id}
                      …
                    {:else}
                      Install
                    {/if}
                  </button>
                </div>
              {/each}
            </div>
          </section>
        {/if}
      </div>
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
  .search {
    display: flex;
    gap: 10px;
    align-items: center;
  }
  .field {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 10px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
    border-radius: 9px;
    padding: 0 6px 0 14px;
    transition: border-color 140ms var(--md-ease), box-shadow 140ms var(--md-ease);
  }
  .field:focus-within {
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
  }
  .icon {
    flex: none;
    color: var(--md-ink-faint);
  }
  .field input {
    flex: 1;
    border: 0;
    background: transparent;
    padding: 11px 0;
    outline: none;
    color: var(--md-ink);
    min-width: 0;
  }
  .clear {
    flex: none;
    font-size: 11px;
    font-weight: 650;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    padding: 7px 10px;
    border-radius: 7px;
    cursor: pointer;
  }
  .clear:hover {
    color: var(--md-ink);
    background: var(--md-stage);
  }
  .result-count {
    margin: 0 0 12px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.02em;
    color: var(--md-ink-faint);
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
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-ink-mute);
    cursor: pointer;
  }
  .filters button.on {
    background: var(--md-cobalt);
    border-color: var(--md-cobalt);
    color: #fff;
  }
  .filters button.on[data-f='verified'],
  .filters button.on[data-f='trusted'] {
    background: color-mix(in oklab, var(--md-live) 16%, var(--md-surface));
    border-color: color-mix(in oklab, var(--md-live) 40%, transparent);
    color: var(--md-live);
  }
  .filters button:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .install-note {
    margin: 0;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--md-live);
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    align-items: center;
  }
  .linkish {
    color: var(--md-cobalt);
    font-weight: 700;
    cursor: pointer;
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 12px;
  }
  .empty-title {
    margin: 0;
    font-family: var(--md-font-display);
    font-size: 18px;
    letter-spacing: -0.03em;
  }
  .empty-lead {
    margin: 8px 0 12px;
    max-width: 40ch;
    font-size: 13px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .empty-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    justify-content: center;
  }
  .shelf {
    display: grid;
    gap: 16px;
  }
  .feature {
    position: relative;
    overflow: hidden;
    background: var(--md-surface);
    border: 1px solid var(--md-line);
    border-radius: 12px;
    padding: 20px;
    box-shadow: none;
  }
  .feature::after {
    display: none;
  }
  .feature > * {
    position: relative;
    z-index: 1;
  }
  .feature-top {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
  }
  .mono {
    width: 36px;
    height: 36px;
    border-radius: 9px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: -0.04em;
    color: #fff;
    background: var(--md-cobalt);
    border: 0;
    box-shadow: none;
  }
  .mono.lg {
    width: 48px;
    height: 48px;
    border-radius: 11px;
    font-size: 20px;
  }
  .tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    justify-content: flex-end;
  }
  .tag {
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    padding: 3px 7px;
    border-radius: 5px;
    background: var(--md-stage);
    color: var(--md-ink-mute);
    border: 1px solid var(--md-line);
  }
  .tag.quiet {
    color: var(--md-ink-faint);
  }
  .tag[data-trust='verified'],
  .tag[data-trust='trusted'],
  .tag.on-machine {
    color: var(--md-live);
    border-color: color-mix(in oklab, var(--md-live) 28%, transparent);
    background: color-mix(in oklab, var(--md-live) 10%, transparent);
  }
  .feature h2 {
    font-family: var(--md-font-display);
    font-size: clamp(26px, 4vw, 34px);
    letter-spacing: -0.045em;
    margin: 0 0 10px;
  }
  .feature > p {
    margin: 0;
    font-size: 15px;
    line-height: 1.55;
    color: var(--md-ink-mute);
    max-width: 52ch;
  }
  .facts {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
    margin: 18px 0;
    padding: 14px 0;
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
  }
  footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
  }
  .feature-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .meta {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .list {
    display: grid;
    gap: 6px;
    padding: 8px;
    border-radius: 12px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
  }
  .row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 8px;
    align-items: center;
  }
  .row.installed {
    opacity: 0.92;
  }
  .row-main {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 12px;
    align-items: center;
    text-align: left;
    padding: 10px;
    border-radius: 14px;
    border: 0;
    background: transparent;
    cursor: pointer;
    color: inherit;
  }
  .row-main:hover {
    background: var(--md-stage);
  }
  .row-main:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .copy {
    min-width: 0;
  }
  .copy strong {
    display: block;
    font-size: 14px;
    font-weight: 700;
  }
  .mini {
    padding: 8px 12px;
    font-size: 12px;
  }
  @media (max-width: 640px) {
    .search {
      flex-direction: column;
      align-items: stretch;
    }
    .facts {
      grid-template-columns: 1fr;
    }
    .feature {
      padding: 18px;
    }
  }
</style>
