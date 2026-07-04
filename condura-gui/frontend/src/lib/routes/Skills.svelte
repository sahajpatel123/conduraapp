<script lang="ts">
  // Skills — local installed skills.
  import { ipc } from '../ipc/client'
  import { onMount } from 'svelte'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import Button from '$components/v1/Button.svelte'
  import Card from '$components/v1/Card.svelte'
  import Input from '$components/v1/Input.svelte'
  import Pill from '$components/v1/Pill.svelte'
  import Chip from '$components/v1/Chip.svelte'
  import EmptyState from '$components/v1/EmptyState.svelte'
  import LoadingState from '$components/v1/LoadingState.svelte'
  import Inline from '$components/v1/Inline.svelte'

  function goto(hash: string): void {
    window.location.hash = hash
  }

  type Skill = {
    id: string
    name: string
    version: string
    trust: string
    source?: string
    description?: string
    category?: string
  }

  type Filter = 'all' | 'built-in' | 'community' | 'experimental'

  let skills = $state<Skill[]>([])
  let loading = $state(false)
  let error = $state<string | null>(null)
  let confirmOpen = $state(false)
  let confirmAction = $state<(() => void) | null>(null)

  let query = $state('')
  let filter = $state<Filter>('all')

  const filters: { id: Filter; label: string }[] = [
    { id: 'all', label: 'All' },
    { id: 'built-in', label: 'Built-in' },
    { id: 'community', label: 'Community' },
    { id: 'experimental', label: 'Experimental' },
  ]

  async function refresh(): Promise<void> {
    loading = true
    error = null
    try {
      skills = (await ipc.skillsList(100)) || []
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  }

  function remove(id: string): void {
    confirmAction = async () => {
      try {
        await ipc.skillsDelete(id)
        await refresh()
      } catch (e) {
        error = String(e)
      }
    }
    confirmOpen = true
  }

  function iconForCategory(cat?: string): string {
    const c = (cat || '').toLowerCase()
    if (c.includes('shell') || c.includes('command')) return 'M4 17l6-6-4-4 2-2 6 6 4-4 2 2-10 10H4v-6z'
    if (c.includes('web') || c.includes('http')) return 'M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 0a9 9 0 009-9H12v9zm0 0v9a9 9 0 009-9h-9z'
    if (c.includes('file') || c.includes('fs')) return 'M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V7z'
    if (c.includes('mail') || c.includes('email')) return 'M3 7l9 6 9-6M3 7v10a2 2 0 002 2h14a2 2 0 002-2V7'
    if (c.includes('time') || c.includes('schedule')) return 'M12 8v4l3 2m6-2a9 9 0 11-18 0 9 9 0 0118 0z'
    if (c.includes('search')) return 'M21 21l-4.3-4.3M11 18a7 7 0 100-14 7 7 0 000 14z'
    return 'M12 2v4M12 18v4M2 12h4M18 12h4M5.6 5.6l2.8 2.8M15.6 15.6l2.8 2.8M5.6 18.4l2.8-2.8M15.6 8.4l2.8-2.8'
  }

  function trustPill(t: string): 'neutral' | 'accent' | 'success' | 'warning' | 'info' {
    const v = (t || '').toLowerCase()
    if (v === 'official') return 'accent'
    if (v === 'community') return 'info'
    if (v === 'experimental') return 'warning'
    return 'neutral'
  }

  function categoryOf(s: Skill): string {
    return (s.source || s.trust || 'utility').toLowerCase()
  }

  function matchesFilter(s: Skill, f: Filter): boolean {
    if (f === 'all') return true
    const src = (s.source || '').toLowerCase()
    const tr = (s.trust || '').toLowerCase()
    if (f === 'built-in') return tr === 'official' || src === 'built-in' || src === 'builtin'
    if (f === 'community') return tr === 'community' || src === 'community'
    if (f === 'experimental') return tr === 'experimental' || src === 'experimental'
    return true
  }

  const visible = $derived(
    skills
      .filter((s) => matchesFilter(s, filter))
      .filter((s) => {
        if (!query.trim()) return true
        const q = query.trim().toLowerCase()
        return (
          s.name.toLowerCase().includes(q) ||
          (s.description || '').toLowerCase().includes(q) ||
          s.id.toLowerCase().includes(q)
        )
      })
  )

  onMount(refresh)
</script>

<div class="skills-page">
  <header class="page-header">
    <Inline gap="4" align="end" justify="between" class="title-row">
      <div>
        <h2 class="page-title">Skills</h2>
        <p class="lede">Reusable procedures the agent can call by name. Built-in and installed.</p>
      </div>
      <Inline gap="2">
        <Button variant="tertiary" size="sm" onclick={() => goto('#/hub')}>Browse the Hub</Button>
        <Button variant="secondary" size="sm" onclick={refresh} loading={loading}>Refresh</Button>
      </Inline>
    </Inline>

    <Inline gap="3" align="center" class="filters">
      <div class="search-wrap">
        <span class="search-icon" aria-hidden="true">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="7" />
            <path d="M21 21l-4.3-4.3" />
          </svg>
        </span>
        <Input
          bind:value={query}
          size="md"
          placeholder="Search skills…"
          ariaLabel="Search skills"
        />
      </div>

      <div class="chips" role="tablist" aria-label="Filter skills">
        {#each filters as f (f.id)}
          <Chip
            selected={filter === f.id}
            onclick={() => (filter = f.id as Filter)}
          >
            {f.label}
          </Chip>
        {/each}
      </div>
    </Inline>
  </header>

  {#if error}
    <p class="error-banner" role="alert">{error}</p>
  {/if}

  {#if loading && skills.length === 0}
    <div class="grid">
      {#each Array(6) as _, i (i)}
        <Card variant="raised" padding="4">
          {#snippet children()}
            <LoadingState kind="cold" />
          {/snippet}
        </Card>
      {/each}
    </div>
  {:else if skills.length === 0}
    <EmptyState
      primary="No skills installed yet"
      secondary="Skills are reusable procedures the agent can call by name. Install some from the Hub."
      voice="mono"
    >
      {#snippet children()}
        <Button variant="primary" size="md" onclick={() => goto('#/hub')}>Browse the Hub</Button>
      {/snippet}
    </EmptyState>
  {:else if visible.length === 0}
    <EmptyState
      primary="No skills match that filter"
      secondary="Try a different search term or switch the chip filter."
      voice="mono"
    />
  {:else}
    <div class="grid">
      {#each visible as s, i (s.id)}
        <div class="grid-item" style:--stagger-index={i}>
          <Card variant="raised" padding="4">
            {#snippet children()}
              <Inline gap="3" align="center" class="card-top">
                <span class="icon-tile">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d={iconForCategory(categoryOf(s))} />
                  </svg>
                </span>
                <div class="card-meta">
                  <h3 class="skill-name">{s.name}</h3>
                  <span class="skill-version">v{s.version}</span>
                </div>
                <Pill variant={trustPill(s.trust)} size="xs" label={s.trust} />
              </Inline>

              {#if s.description}
                <p class="skill-desc">{s.description}</p>
              {/if}

              <Inline gap="2" justify="between" align="center" class="card-foot">
                <span class="skill-id">{s.id}</span>
                <Inline gap="1">
                  <Button variant="primary" size="sm" onclick={() => goto('#/skills')}>Use</Button>
                  <Button variant="tertiary" size="sm" onclick={() => remove(s.id)}>Remove</Button>
                </Inline>
              </Inline>
            {/snippet}
          </Card>
        </div>
      {/each}
    </div>
  {/if}
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title="Remove skill"
  description="This will remove the skill from your local installation. You can reinstall it from the Hub anytime."
  tone="danger"
  confirmLabel="Remove"
  onconfirm={() => confirmAction?.()}
/>

<style>
  .skills-page {
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

  .filters {
    flex-wrap: wrap;
  }

  .search-wrap {
    flex: 1 1 280px;
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

  .skills-page :global(.search-wrap .input) {
    border: 1px solid var(--border-default);
    border-radius: var(--radius-sm);
    padding: 0 var(--space-3) 0 calc(var(--space-3) + 22px);
  }

  .chips {
    display: inline-flex;
    gap: var(--space-1);
    flex-wrap: wrap;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: var(--space-4);
  }

  .grid-item {
    animation: skills-stagger var(--duration-base) var(--ease-standard) both;
    animation-delay: calc(var(--stagger-index, 0) * 50ms);
  }

  @keyframes skills-stagger {
    from { opacity: 0; transform: translateY(6px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .card-top {
    width: 100%;
    margin-bottom: var(--space-3);
  }

  .icon-tile {
    width: 36px;
    height: 36px;
    border-radius: var(--radius-md);
    background-color: var(--plum-50);
    border: 1px solid var(--border-default);
    color: var(--content-accent);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .card-meta {
    flex: 1;
    min-width: 0;
  }

  .skill-name {
    font-size: var(--text-body-size);
    font-weight: 500;
    color: var(--content-primary);
    margin: 0;
    line-height: var(--text-body-leading, 1.5);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .skill-version {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
  }

  .skill-desc {
    font-size: var(--text-body-sm-size);
    color: var(--content-secondary);
    line-height: 1.55;
    margin: 0 0 var(--space-3) 0;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .card-foot {
    width: 100%;
    padding-top: var(--space-3);
    border-top: 1px solid var(--border-subtle);
  }

  .skill-id {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
    flex: 1;
  }

  .error-banner {
    color: var(--status-error-fg);
    font-size: var(--text-body-sm-size);
    padding: var(--space-2) var(--space-3);
    background-color: var(--status-error-bg);
    border: 1px solid var(--status-error-border);
    border-radius: var(--radius-md);
    margin-bottom: var(--space-4);
  }
</style>
