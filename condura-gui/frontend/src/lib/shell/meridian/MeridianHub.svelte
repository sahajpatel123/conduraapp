<script lang="ts">
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { hub } from '../../stores/hub.svelte'

  let q = $state('')
  let installing = $state<string | null>(null)
  let note = $state('')
  let featuredId = $state<string | null>(null)

  onMount(() => {
    void hub.search('', 24)
  })

  const featured = $derived(
    hub.results.find((s) => s.id === featuredId) ?? hub.results[0] ?? null
  )
  const rest = $derived(
    featured ? hub.results.filter((s) => s.id !== featured.id) : hub.results
  )

  function search(): void {
    note = ''
    void hub.search(q.trim(), 24)
  }

  function clearSearch(): void {
    q = ''
    note = ''
    void hub.search('', 24)
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
      note = 'Installed locally — open Skills to inspect.'
    } finally {
      installing = null
    }
  }

  function goSkills(): void {
    window.location.hash = '#/skills'
  }
</script>

<MeridianPage
  kicker="Shelf"
  title="Hub"
  lead="Browse community skills. Install stays local — nothing runs until you say so."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={search}>Refresh</button>
    <button type="button" class="md-btn md-btn-ghost" onclick={goSkills}>My Skills</button>
  {/snippet}

  <div class="search">
    <div class="field">
      <svg class="icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
        <circle cx="11" cy="11" r="7" />
        <path d="M20 20l-3.5-3.5" />
      </svg>
      <input
        type="search"
        placeholder="Search skills…"
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

  {#if note}
    <p class="install-note">
      {note}
      <button type="button" class="linkish" onclick={goSkills}>Open Skills →</button>
    </p>
  {/if}

  {#if hub.loading}
    <div class="md-empty">Loading the shelf…</div>
  {:else if hub.error && !/IPC client not started|not connected|Failed to fetch/i.test(hub.error)}
    <div class="md-empty">{hub.error}</div>
  {:else if hub.results.length === 0}
    <div class="md-empty empty">
      <p class="empty-title">{q ? 'No skills matched' : 'Shelf is quiet'}</p>
      <p class="empty-lead">
        {#if q}
          Try a broader term, or clear the search to browse the shelf.
        {:else}
          Connect the daemon to load community skills, or search once you’re online.
        {/if}
      </p>
      {#if q}
        <button type="button" class="md-btn md-btn-ghost" onclick={clearSearch}>Clear search</button>
      {/if}
    </div>
  {:else}
    <div class="shelf md-stagger">
      {#if featured}
        <article class="feature">
          <p class="cite">featured on the shelf</p>
          <div class="feature-top">
            <div class="mono lg" aria-hidden="true">{initial(featured.name)}</div>
            <div class="tags">
              {#if featured.trust}
                <span class="tag" data-trust={featured.trust}>{featured.trust}</span>
              {/if}
              {#if featured.version}
                <span class="tag quiet">v{featured.version}</span>
              {/if}
            </div>
          </div>
          <h2>{featured.name}</h2>
          <p>{featured.description || 'No description.'}</p>
          <footer>
            <span class="meta">
              {featured.author || 'community'}
              {#if featured.downloads != null && featured.downloads > 0}
                · {featured.downloads.toLocaleString()} installs
              {/if}
              · stays local after install
            </span>
            <button
              type="button"
              class="md-btn md-btn-primary"
              disabled={hub.installed.has(featured.id) || installing === featured.id}
              onclick={() => void install(featured.id)}
            >
              {#if hub.installed.has(featured.id)}
                Installed
              {:else if installing === featured.id}
                Installing…
              {:else}
                Install
              {/if}
            </button>
          </footer>
        </article>
      {/if}

      {#if rest.length}
        <div class="list">
          {#each rest as skill (skill.id)}
            <div class="row">
              <button type="button" class="row-main" onclick={() => (featuredId = skill.id)}>
                <div class="mono" aria-hidden="true">{initial(skill.name)}</div>
                <div class="copy">
                  <strong>{skill.name}</strong>
                  <span class="meta">
                    {skill.author || 'community'}
                    {#if skill.trust} · {skill.trust}{/if}
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
      {/if}
    </div>
  {/if}
</MeridianPage>

<style>
  .search {
    display: flex;
    gap: 10px;
    margin-bottom: 24px;
    align-items: center;
  }
  .field {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 10px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    border-radius: 999px;
    padding: 0 6px 0 16px;
    transition:
      border-color var(--md-dur) var(--md-ease),
      box-shadow var(--md-dur) var(--md-ease);
  }
  .field:hover {
    border-color: color-mix(in oklab, var(--md-cobalt) 28%, var(--md-line-strong));
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
    padding: 12px 0;
    outline: none;
    color: var(--md-ink);
    min-width: 0;
  }
  .clear {
    flex: none;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    padding: 8px 12px;
    border-radius: 999px;
    cursor: pointer;
    transition: color 160ms var(--md-ease), background 160ms var(--md-ease);
  }
  .clear:hover {
    color: var(--md-ink);
    background: var(--md-stage);
  }
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
    max-width: 36ch;
    font-size: 13px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .install-note {
    margin: -8px 0 18px;
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
  .shelf {
    display: grid;
    gap: 14px;
  }
  .feature {
    background: var(--md-surface);
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 28%, var(--md-line-strong));
    border-radius: 24px;
    padding: 24px;
    box-shadow: var(--md-shadow-lift);
  }
  .feature-top {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
  }
  .mono {
    width: 40px;
    height: 40px;
    border-radius: 14px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -0.04em;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
  }
  .mono.lg {
    width: 56px;
    height: 56px;
    border-radius: 18px;
    font-size: 24px;
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
    padding: 4px 8px;
    border-radius: 999px;
    background: var(--md-stage);
    color: var(--md-ink-mute);
    border: 1px solid var(--md-line);
  }
  .tag.quiet {
    color: var(--md-ink-faint);
  }
  .tag[data-trust='verified'],
  .tag[data-trust='trusted'] {
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
    margin: 0 0 18px;
    font-size: 15px;
    line-height: 1.55;
    color: var(--md-ink-mute);
    max-width: 52ch;
  }
  footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
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
    border-radius: 20px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
  }
  .row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 8px;
    align-items: center;
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
  @media (max-width: 560px) {
    .search {
      flex-direction: column;
      align-items: stretch;
      gap: 8px;
      margin-bottom: 18px;
    }
    .field {
      padding: 0 6px 0 14px;
      border-radius: 16px;
    }
    .field input {
      padding: 11px 0;
    }
    .search :global(.md-btn-primary) {
      width: 100%;
    }
    .feature {
      padding: 18px;
    }
  }
  @media (max-width: 420px) {
    .feature h2 {
      font-size: 22px;
    }
  }
</style>
