<script lang="ts">
  import { ipc } from '../ipc/client'
  import PublishModal from '../components/PublishModal.svelte'
  import { t } from '../i18n'

  let showPublish = $state(false)
  let query = $state('')
  let results = $state<Array<{ id: string; name: string; version: string; author: string; description: string; trust: string }>>([])
  let cursor = $state(0)
  let loading = $state(false)
  let error = $state<string | null>(null)
  let installed = $state<Set<string>>(new Set())

  async function search() {
    if (!query.trim()) return
    loading = true
    error = null
    try {
      const r = await ipc.hubSearch(query, 20)
      results = r.skills || []
      cursor = 0
      if (results.length === 0) {
        error = t('hub.no_results', query)
      }
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  }

  async function install(id: string) {
    if (!confirm(t('hub.install_confirm', id))) return
    try {
      await ipc.hubInstall(id)
      installed.add(id)
      installed = new Set(installed) // trigger reactivity
    } catch (e) {
      error = String(e)
    }
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'ArrowDown' && cursor < results.length - 1) cursor++
    if (e.key === 'ArrowUp' && cursor > 0) cursor--
  }
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div class="hub-page" onkeydown={onKey} role="region" aria-label={t('hub.aria_label')}>
  <header class="hub-header">
    <div class="page-header">
      <h2>{t('hub.title')}</h2>
      <p class="muted">{t('hub.intro')}</p>
    </div>
    <button class="btn btn-primary publish-btn" onclick={() => (showPublish = true)}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 5v14M5 12h14" /></svg>
      {t('hub.publish_button')}
    </button>
  </header>

  <div class="search-pill">
    <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="11" cy="11" r="7" /><path d="M21 21l-4.3-4.3" /></svg>
    <input
      type="text"
      class="search-input"
      placeholder={t('hub.search_placeholder')}
      bind:value={query}
      onkeydown={(e) => { if (e.key === 'Enter') search() }}
    />
    <button class="btn btn-primary btn-sm" onclick={search} disabled={loading || !query.trim()}>
      {loading ? t('hub.searching') : t('hub.search')}
    </button>
  </div>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if results.length > 0}
    <ul class="results">
      {#each results as r, i}
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <li class="result-row" class:selected={i === cursor} onclick={() => cursor = i} onkeydown={() => {}}>
          <div class="row">
            <strong>{r.name}</strong>
            <span class="version">v{r.version}</span>
            <span class="badge trust-badge">{r.trust}</span>
            <span class="spacer"></span>
            {#if installed.has(r.id)}
              <span class="badge badge-success">{t('hub.installed')}</span>
            {:else}
              <button class="btn btn-primary btn-xs" onclick={(e) => { e.stopPropagation(); install(r.id) }}>{t('hub.install')}</button>
            {/if}
          </div>
          <div class="meta">
            <span class="author">{t('hub.by', r.author)}</span>
            <span class="id mono">{r.id}</span>
          </div>
          {#if r.description}
            <p class="desc">{r.description}</p>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}

  <footer>
    <p class="muted">
      {t('hub.footer')}
    </p>
  </footer>
</div>

{#if showPublish}
  <PublishModal onClose={() => (showPublish = false)} />
{/if}

<style>
  .hub-page {
    padding: var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide);
    margin: 0 auto;
  }
  .hub-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-4);
  }
  .publish-btn svg {
    width: 16px;
    height: 16px;
  }
  .search-pill {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin: var(--space-4) 0;
    padding: var(--space-1) var(--space-1) var(--space-1) var(--space-4);
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-pill);
    box-shadow: var(--shadow-inset);
    transition: border-color var(--transition-base), box-shadow var(--transition-base);
  }
  .search-pill:focus-within {
    border-color: var(--color-accent);
    box-shadow: var(--shadow-focus);
  }
  .search-icon {
    width: 18px;
    height: 18px;
    color: var(--color-text-faint);
    flex-shrink: 0;
  }
  .search-input {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    color: var(--color-text);
    font-size: var(--size-md);
    font-family: var(--font-sans);
    padding: var(--space-2) 0;
  }
  .search-input::placeholder {
    color: var(--color-text-faint);
  }
  .results {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .result-row {
    padding: var(--space-3) var(--space-4);
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    cursor: pointer;
    transition: border-color var(--transition-base), background var(--transition-base);
  }
  .result-row:hover,
  .result-row.selected {
    border-color: var(--color-border-accent);
    background: var(--glass-bg-hover);
  }
  .row {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .spacer { flex: 1; }
  .version {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    font-family: var(--font-mono);
  }
  .trust-badge {
    color: var(--color-text-muted);
  }
  .meta {
    display: flex;
    gap: var(--space-3);
    margin-top: var(--space-1);
  }
  .author {
    color: var(--color-text-muted);
    font-size: var(--size-xs);
  }
  .id {
    font-size: var(--size-xs);
    color: var(--color-text-faint);
  }
  .desc {
    margin: var(--space-1) 0 0 0;
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-relaxed);
  }
  .error {
    color: var(--color-error);
    font-size: var(--size-sm);
    padding: var(--space-2) var(--space-3);
    background: var(--color-error-soft);
    border-radius: var(--radius-md);
  }
</style>
