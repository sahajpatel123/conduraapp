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
        error = $t('hub.no_results', query)
      }
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  }

  async function install(id: string) {
    if (!confirm($t('hub.install_confirm', id))) return
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
<div class="hub-page" onkeydown={onKey} role="region" aria-label={$t('hub.aria_label')}>
  <header class="hub-header">
    <div>
      <h2>{$t('hub.title')}</h2>
      <p class="muted">{$t('hub.intro')}</p>
    </div>
    <button class="publish-btn" onclick={() => (showPublish = true)}>+ {$t('hub.publish_button')}</button>
  </header>

  <div class="search-bar">
    <input
      type="text"
      placeholder={$t('hub.search_placeholder')}
      bind:value={query}
      onkeydown={(e) => { if (e.key === 'Enter') search() }}
    />
    <button onclick={search} disabled={loading || !query.trim()}>
      {loading ? $t('hub.searching') : $t('hub.search')}
    </button>
  </div>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if results.length > 0}
    <ul class="results">
      {#each results as r, i}
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <li class:selected={i === cursor} onclick={() => cursor = i} onkeydown={() => {}}>
          <div class="row">
            <strong>{r.name}</strong>
            <span class="version">v{r.version}</span>
            <span class="trust">[{r.trust}]</span>
            <span class="spacer"></span>
            {#if installed.has(r.id)}
              <span class="installed">{$t('hub.installed')}</span>
            {:else}
              <button onclick={(e) => { e.stopPropagation(); install(r.id) }}>{$t('hub.install')}</button>
            {/if}
          </div>
          <div class="meta">
            <span class="author">{$t('hub.by', r.author)}</span>
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
      {$t('hub.footer')}
    </p>
  </footer>
</div>

{#if showPublish}
  <PublishModal onClose={() => (showPublish = false)} />
{/if}

<style>
  .hub-page { padding: 16px; }
  .hub-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; }
  .publish-btn {
    padding: 8px 16px;
    border-radius: var(--radius-md, 8px);
    border: none;
    background: var(--color-accent-gradient, #4a9eff);
    color: white;
    font-weight: 500;
    cursor: pointer;
    white-space: nowrap;
  }
  .publish-btn:hover { box-shadow: var(--shadow-glow, 0 0 12px rgba(74,158,255,0.4)); }
  .search-bar { display: flex; gap: 8px; margin: 12px 0; }
  .search-bar input { flex: 1; padding: 8px; font-size: 14px; }
  .search-bar button { padding: 8px 16px; }
  .results { list-style: none; padding: 0; margin: 0; }
  .results li { padding: 12px; border: 1px solid transparent; border-radius: 6px; cursor: pointer; }
  .results li.selected { border-color: var(--accent, #4a9eff); background: var(--hover, rgba(74, 158, 255, 0.08)); }
  .row { display: flex; align-items: baseline; gap: 8px; }
  .spacer { flex: 1; }
  .version, .trust, .author { color: var(--muted, #888); font-size: 12px; }
  .meta { display: flex; gap: 12px; margin-top: 4px; }
  .id { font-size: 11px; opacity: 0.6; }
  .desc { margin: 6px 0 0 0; color: var(--fg, #333); }
  .error { color: var(--error, #c0392b); padding: 8px; background: var(--error-bg, rgba(192, 57, 43, 0.1)); border-radius: 4px; }
  .installed { color: var(--success, #27ae60); font-weight: 500; }
  .muted { color: var(--muted, #888); }
  .mono { font-family: ui-monospace, monospace; }
  kbd { background: var(--bg-elev, #f0f0f0); border: 1px solid var(--border, #ccc); border-radius: 3px; padding: 1px 4px; font-size: 11px; }
</style>
