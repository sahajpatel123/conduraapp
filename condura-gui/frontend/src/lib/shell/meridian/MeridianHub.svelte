<script lang="ts">
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { hub } from '../../stores/hub.svelte'

  let q = $state('')
  let installing = $state<string | null>(null)

  onMount(() => {
    void hub.search('', 24)
  })

  function search(): void {
    void hub.search(q.trim(), 24)
  }

  function clearSearch(): void {
    q = ''
    void hub.search('', 24)
  }

  function initial(name: string): string {
    const t = name.trim()
    return t ? t[0]!.toUpperCase() : '?'
  }

  async function install(id: string): Promise<void> {
    if (hub.installed.has(id) || installing) return
    installing = id
    try {
      await hub.install(id)
    } finally {
      installing = null
    }
  }
</script>

<MeridianPage
  kicker="Catalog"
  title="Hub"
  lead="Browse community skills. Install stays local — nothing runs until you say so."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={search}>Refresh</button>
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
    <div class="grid md-stagger">
      {#each hub.results as skill (skill.id)}
        <article class="card">
          <div class="top">
            <div class="mono" aria-hidden="true">{initial(skill.name)}</div>
            <div class="tags">
              {#if skill.trust}
                <span class="tag" data-trust={skill.trust}>{skill.trust}</span>
              {/if}
              {#if skill.version}
                <span class="tag quiet">v{skill.version}</span>
              {/if}
            </div>
          </div>
          <h2>{skill.name}</h2>
          <p>{skill.description || 'No description.'}</p>
          <footer>
            <span class="meta">
              {skill.author || 'community'}
              {#if skill.downloads != null && skill.downloads > 0}
                · {skill.downloads.toLocaleString()} installs
              {/if}
            </span>
            <button
              type="button"
              class="md-btn md-btn-primary"
              disabled={hub.installed.has(skill.id) || installing === skill.id}
              onclick={() => void install(skill.id)}
            >
              {#if hub.installed.has(skill.id)}
                Installed
              {:else if installing === skill.id}
                Installing…
              {:else}
                Install
              {/if}
            </button>
          </footer>
        </article>
      {/each}
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
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 14px;
  }
  .card {
    background: var(--md-surface);
    border: 1px solid var(--md-line-strong);
    border-radius: 22px;
    padding: 18px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    min-height: 196px;
    box-shadow: var(--md-shadow);
    transition:
      transform 220ms var(--md-spring),
      box-shadow 220ms var(--md-ease),
      border-color 220ms var(--md-ease);
  }
  .card:hover {
    transform: translateY(-4px);
    border-color: color-mix(in oklab, var(--md-cobalt) 32%, var(--md-line-strong));
    box-shadow: var(--md-shadow-lift);
  }
  .card:focus-within {
    border-color: color-mix(in oklab, var(--md-cobalt) 45%, var(--md-line-strong));
    box-shadow: var(--md-focus), var(--md-shadow);
  }
  .clear:focus-visible {
    outline: none;
    color: var(--md-ink);
    box-shadow: var(--md-focus);
  }
  .top {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 10px;
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
  .card h2 {
    font-family: var(--md-font-display);
    font-size: 20px;
    letter-spacing: -0.03em;
    margin: 0;
    line-height: 1.15;
  }
  .card p {
    margin: 0;
    flex: 1;
    font-size: 13px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 4px;
  }
  .meta {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
    .grid {
      grid-template-columns: 1fr;
      gap: 10px;
    }
    .card {
      min-height: 0;
      padding: 16px;
      border-radius: 18px;
    }
    .card:hover {
      transform: none;
    }
  }
  @media (max-width: 420px) {
    .card h2 {
      font-size: 18px;
    }
    footer {
      flex-wrap: wrap;
    }
  }
</style>
