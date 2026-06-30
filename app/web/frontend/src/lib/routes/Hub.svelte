<script lang="ts">
  // Hub — Skills Hub (remote catalog). Search the hub, install skills,
  // publish your own. The publish flow opens PublishModal.
  import { ipc } from '../ipc/client'
  import { hub } from '../stores/hub.svelte'
  import { notifications } from '../stores/notifications.svelte'
  import PublishModal from '../components/PublishModal.svelte'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import Button from '$components/v1/Button.svelte'
  import Card from '$components/v1/Card.svelte'
  import Input from '$components/v1/Input.svelte'
  import Pill from '$components/v1/Pill.svelte'
  import EmptyState from '$components/v1/EmptyState.svelte'
  import Inline from '$components/v1/Inline.svelte'
  import Hairline from '$components/v1/Hairline.svelte'

  let showPublish = $state(false)
  let query = $state('')
  let cursor = $state(0)
  let confirmOpen = $state(false)
  let confirmAction = $state<(() => void) | null>(null)

  function trustPill(trust: string): 'accent' | 'neutral' | 'warning' | 'info' {
    if (trust === 'official') return 'accent'
    if (trust === 'experimental') return 'warning'
    if (trust === 'community') return 'info'
    return 'neutral'
  }

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
    <Inline gap="4" align="end" justify="between" class="title-row">
      <div>
        <h2 class="page-title">Skills Hub</h2>
        <p class="lede">Browse community skills, or publish your own. Safety-scanned on import.</p>
      </div>
      <Button variant="primary" size="md" onclick={() => (showPublish = true)}>
        {#snippet icon()}
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" aria-hidden="true">
            <path d="M12 5v14M5 12h14" />
          </svg>
        {/snippet}
        Publish a skill
      </Button>
    </Inline>

    <form
      class="search-row"
      onsubmit={(e) => {
        e.preventDefault()
        void search()
      }}
    >
      <div class="search-input-wrap">
        <span class="search-icon" aria-hidden="true">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="7" />
            <path d="M21 21l-4.3-4.3" />
          </svg>
        </span>
        <Input
          bind:value={query}
          size="lg"
          placeholder="Search the hub…"
          ariaLabel="Search skills hub"
        />
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
    <p class="error-banner" role="alert">{hub.error}</p>
  {/if}

  {#if hub.results.length === 0 && !hub.loading}
    <Card variant="raised" padding="6">
      {#snippet children()}
        <EmptyState
          primary="Search the Skills Hub"
          secondary="Find reusable procedures contributed by the community. Installs are safety-scanned before they land on your machine."
          voice="serif"
        />
      {/snippet}
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
          <Inline gap="4" align="start" justify="between" wrap={false} class="row-main">
            <div class="name-block">
              <h3 class="result-name">{r.name}</h3>
              <Inline gap="2" align="center" class="meta-row">
                <Pill variant={trustPill(r.trust)} size="xs" label={r.trust} />
                <span class="version">v{r.version}</span>
                <span class="by">by {r.author}</span>
              </Inline>
            </div>

            <div class="install-cell">
              {#if hub.installed.has(r.id)}
                <Pill variant="success" size="sm" label="Installed" />
              {:else}
                <Button variant="primary" size="sm" onclick={(e) => { e.stopPropagation(); install(r.id) }}>Install</Button>
              {/if}
            </div>
          </Inline>

          {#if r.description}
            <p class="desc">{r.description}</p>
          {/if}

          <Inline gap="2" justify="between" align="center" class="id-row">
            <span class="id">{r.id}</span>
            {#if r.downloads != null}
              <span class="downloads">↓ {r.downloads.toLocaleString()}</span>
            {/if}
          </Inline>
        </li>
      {/each}
    </ul>
  {/if}

  <footer class="page-foot">
    <Hairline />
    <p class="muted">
      Skills are signed and safety-scanned before installation. Trust levels:
      <strong>official</strong>, <strong>community</strong>, <strong>experimental</strong>.
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
    max-width: 56rem;
    margin: 0 auto;
    background-color: var(--surface-base);
  }

  .page-header {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    margin-bottom: var(--space-6);
  }

  .title-row {
    width: 100%;
  }

  .page-title {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    font-weight: var(--text-h2-weight);
    letter-spacing: var(--text-h2-tracking);
    color: var(--content-primary);
    margin: 0 0 var(--space-1) 0;
    line-height: var(--text-h2-leading);
  }

  .lede {
    font-size: var(--text-body-size);
    color: var(--content-secondary);
    line-height: 1.55;
    max-width: 35rem;
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
    position: relative;
    display: flex;
    align-items: center;
  }

  .search-icon {
    position: absolute;
    left: var(--space-3);
    color: var(--content-tertiary);
    pointer-events: none;
    display: flex;
  }

  .hub-page :global(.search-input-wrap .input) {
    border: 1px solid var(--border-default);
    border-radius: var(--radius-sm);
    padding: 0 var(--space-3) 0 calc(var(--space-3) + 24px);
  }

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
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition:
      background-color var(--duration-fast) var(--ease-standard),
      border-color var(--duration-fast) var(--ease-standard);
    animation: hub-stagger var(--duration-base) var(--ease-standard) both;
    animation-delay: calc(var(--stagger-index, 0) * 60ms);
  }

  @keyframes hub-stagger {
    from { opacity: 0; transform: translateY(6px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .result-row:hover,
  .result-row.selected {
    border-color: var(--border-focus);
    background-color: var(--surface-sunken);
  }

  .row-main {
    width: 100%;
  }

  .name-block {
    min-width: 0;
    flex: 1;
  }

  .result-name {
    font-size: var(--text-body-size);
    font-weight: 500;
    color: var(--content-primary);
    margin: 0 0 var(--space-1) 0;
    line-height: var(--text-body-leading, 1.5);
  }

  .meta-row {
    flex-wrap: wrap;
  }

  .version {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
  }

  .by {
    font-size: var(--text-caption-size);
    color: var(--content-secondary);
  }

  .install-cell {
    flex-shrink: 0;
  }

  .desc {
    font-size: var(--text-body-sm-size);
    color: var(--content-secondary);
    line-height: 1.55;
    margin: 0;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .id-row {
    width: 100%;
  }

  .id,
  .downloads {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
  }

  .page-foot {
    margin-top: var(--space-7);
    padding-top: var(--space-4);
  }

  .muted {
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    margin: var(--space-3) 0 0 0;
  }

  .muted strong {
    color: var(--content-primary);
    font-weight: 500;
  }

  .error-banner {
    color: var(--status-error-fg);
    font-size: var(--text-body-sm-size);
    padding: var(--space-2) var(--space-3);
    background-color: var(--status-error-bg);
    border: 1px solid var(--status-error-border);
    border-radius: var(--radius-md);
    margin: 0 0 var(--space-3) 0;
  }
</style>
