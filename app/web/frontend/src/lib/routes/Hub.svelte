<script lang="ts">
  // Hub — Skills Hub (remote catalog). Search the hub, install skills,
  // publish your own. The publish flow opens PublishModal.
  import { ipc } from '../ipc/client'
  import { hub } from '../stores/hub.svelte'
  import { notifications } from '../stores/notifications.svelte'
  import PublishModal from '../components/PublishModal.svelte'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import { Button, Card, Input, Badge } from '../components/ui'

  let showPublish = $state(false)
  let query = $state('')
  let cursor = $state(0)
  let confirmOpen = $state(false)
  let confirmAction = $state<(() => void) | null>(null)

  async function search(): Promise<void> {
    if (!query.trim()) return
    hub.loading = true
    hub.error = null
    try {
      const r = await ipc.hubSearch(query, 20)
      hub.results = r.skills || []
      hub.lastQuery = query
      cursor = 0
      if (hub.results.length === 0) {
        hub.error = `No results for "${query}".`
      }
    } catch (e) {
      hub.error = String(e)
    } finally {
      hub.loading = false
    }
  }

  function install(id: string): void {
    confirmAction = async () => {
      try {
        await ipc.hubInstall(id)
        hub.installed.add(id)
        hub.installed = new Set(hub.installed)
        notifications.push({
          kind: 'success',
          title: 'Installed',
          message: `Skill ${id} installed.`,
        })
      } catch (e) {
        hub.error = String(e)
        notifications.push({ kind: 'error', title: 'Install failed', message: String(e) })
      }
    }
    confirmOpen = true
  }

  function onKey(e: KeyboardEvent): void {
    if (e.key === 'ArrowDown' && cursor < hub.results.length - 1) cursor++
    if (e.key === 'ArrowUp' && cursor > 0) cursor--
  }
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div class="hub-page" onkeydown={onKey} role="region" aria-label="Skills Hub">
  <header class="page-header">
    <div class="title-row">
      <div>
        <h2 class="display-h2">Skills Hub</h2>
        <p class="lede">Browse community skills, or publish your own. Safety-scanned on import.</p>
      </div>
      <Button variant="primary" size="md" onclick={() => (showPublish = true)}>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" aria-hidden="true">
          <path d="M12 5v14M5 12h14" />
        </svg>
        Publish a skill
      </Button>
    </div>

    <form
      class="search-row"
      onsubmit={(e) => {
        e.preventDefault()
        void search()
      }}
    >
      <div class="search-input-wrap">
        <Input
          bind:value={query}
          fullWidth
          size="lg"
          placeholder="Search the hub…"
          aria-label="Search skills hub"
        >
          {#snippet leading()}
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <circle cx="11" cy="11" r="7" />
              <path d="M21 21l-4.3-4.3" />
            </svg>
          {/snippet}
        </Input>
      </div>
      <Button
        type="submit"
        variant="primary"
        size="lg"
        loading={hub.loading}
        disabled={!query.trim()}
      >
        {hub.loading ? 'Searching…' : 'Search'}
      </Button>
    </form>
  </header>

  {#if hub.error}
    <p class="error" role="alert">{hub.error}</p>
  {/if}

  {#if hub.results.length === 0 && !hub.loading}
    <Card elevation="glass" padding="lg">
      <div class="empty">
        <div class="empty-icon">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="9" />
            <path d="M3.6 9h16.8M3.6 15h16.8M12 3a14 14 0 010 18M12 3a14 14 0 000 18" />
          </svg>
        </div>
        <h3>Search the Skills Hub</h3>
        <p>Find reusable procedures contributed by the community. Installs are safety-scanned before they land on your machine.</p>
      </div>
    </Card>
  {:else}
    <ul class="results" role="listbox" aria-label="Search results">
      {#each hub.results as r, i (r.id)}
        <li
          class="result-row"
          class:selected={i === cursor}
          style:--stagger-index={i}
          role="option"
          aria-selected={i === cursor}
          onclick={() => (cursor = i)}
          onkeydown={() => {}}
        >
          <div class="row-main">
            <div class="name-block">
              <h3 class="result-name">{r.name}</h3>
              <div class="meta-row">
                <Badge tone={r.trust === 'official' ? 'accent' : 'neutral'} size="xs">{r.trust}</Badge>
                <span class="version mono">v{r.version}</span>
                <span class="by">by {r.author}</span>
              </div>
            </div>

            <div class="install-cell">
              {#if hub.installed.has(r.id)}
                <Badge tone="success" size="sm" dot>Installed</Badge>
              {:else}
                <Button variant="primary" size="sm" onclick={(e) => { e.stopPropagation(); install(r.id) }}>Install</Button>
              {/if}
            </div>
          </div>

          {#if r.description}
            <p class="desc">{r.description}</p>
          {/if}

          <div class="id-row">
            <span class="id mono">{r.id}</span>
            {#if r.downloads != null}
              <span class="downloads mono">↓ {r.downloads.toLocaleString()}</span>
            {/if}
          </div>
        </li>
      {/each}
    </ul>
  {/if}

  <footer class="page-foot">
    <p class="muted">
      Skills are signed and safety-scanned before installation. Trust levels: <strong>official</strong>,
      <strong>community</strong>, <strong>experimental</strong>.
    </p>
  </footer>
</div>

{#if showPublish}
  <PublishModal onClose={() => (showPublish = false)} />
{/if}

<ConfirmDialog
  bind:open={confirmOpen}
  title="Install skill"
  description="This will download and safety-scan the skill. A native prompt will appear if any post-install actions are required."
  tone="primary"
  confirmLabel="Install"
  onconfirm={() => confirmAction?.()}
/>

<style>
  .hub-page {
    padding: var(--space-6) var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide);
    margin: 0 auto;
  }

  .page-header {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    margin-bottom: var(--space-6);
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .title-row {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--space-4);
    flex-wrap: wrap;
  }
  .display-h2 {
    font-family: var(--font-display);
    font-size: var(--size-2xl);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    color: var(--text);
    margin: 0 0 var(--space-1) 0;
    line-height: var(--leading-tight);
  }
  .lede {
    font-size: var(--size-md);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    max-width: 560px;
    margin: 0;
  }

  .search-row {
    display: flex;
    gap: var(--space-2);
    align-items: stretch;
  }
  .search-input-wrap {
    flex: 1;
    min-width: 0;
  }

  /* ── Empty state card ─────────────────────────── */
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: var(--space-3);
    padding: var(--space-7) var(--space-5);
  }
  .empty-icon {
    width: 56px;
    height: 56px;
    border-radius: var(--radius-xl);
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-faint);
    margin-bottom: var(--space-2);
  }
  .empty h3 {
    font-family: var(--font-display);
    font-size: var(--size-xl);
    font-weight: var(--weight-medium);
    color: var(--text);
    margin: 0;
    letter-spacing: var(--tracking-tight);
  }
  .empty p {
    color: var(--text-muted);
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    max-width: 480px;
    margin: 0;
  }

  /* ── Results ──────────────────────────────────── */
  .results {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .result-row {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-4);
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    cursor: pointer;
    transition:
      background-color var(--transition-base),
      border-color var(--transition-base),
      box-shadow var(--transition-base);
    animation: stagger-in var(--transition-base) var(--ease-out-expo) both;
    animation-delay: calc(var(--stagger-index, 0) * 60ms);
  }
  .result-row:hover,
  .result-row.selected {
    border-color: var(--border-focus);
    background: var(--glass-bg-hover);
    box-shadow: var(--shadow-glow);
  }

  .row-main {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: var(--space-4);
  }
  .name-block { min-width: 0; }
  .result-name {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    margin: 0 0 var(--space-1) 0;
    line-height: var(--leading-tight);
  }
  .meta-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .version {
    font-size: var(--size-xs);
    color: var(--text-faint);
  }
  .by {
    font-size: var(--size-xs);
    color: var(--text-muted);
  }
  .install-cell {
    flex-shrink: 0;
  }

  .desc {
    font-size: var(--size-sm);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    margin: 0;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .id-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--space-2);
  }
  .id {
    font-size: var(--size-xs);
    color: var(--text-faint);
  }
  .downloads {
    font-size: var(--size-xs);
    color: var(--text-faint);
  }

  .page-foot {
    margin-top: var(--space-7);
    padding-top: var(--space-4);
    border-top: 1px solid var(--border);
  }
  .muted {
    font-size: var(--size-xs);
    color: var(--text-muted);
    margin: 0;
  }
  .muted strong { color: var(--text); font-weight: var(--weight-semibold); }

  .error {
    color: var(--error);
    font-size: var(--size-sm);
    padding: var(--space-2) var(--space-3);
    background: var(--error-soft);
    border: 1px solid var(--border-danger);
    border-radius: var(--radius-md);
    margin: 0 0 var(--space-3) 0;
  }
</style>