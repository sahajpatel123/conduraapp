<script lang="ts">
  // Skills — local installed skills.
  //
  // Header: search input + filter chips (All / Built-in / Community / Experimental).
  // Body: a grid of Skill cards with icon, name, description, version, "Use" button.
  // Empty state: EmptyState with "Browse the Hub" CTA.
  // Loading state: Skeleton rows.
  import { ipc } from '../ipc/client'
  import { onMount } from 'svelte'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import { Button, Card, Input, Badge, EmptyState, Skeleton } from '../components/ui'

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
    // default sparkle
    return 'M12 2v4M12 18v4M2 12h4M18 12h4M5.6 5.6l2.8 2.8M15.6 15.6l2.8 2.8M5.6 18.4l2.8-2.8M15.6 8.4l2.8-2.8'
  }

  function trustTone(t: string): 'neutral' | 'accent' | 'success' | 'warn' | 'error' | 'info' {
    const v = (t || '').toLowerCase()
    if (v === 'official') return 'accent'
    if (v === 'community') return 'info'
    if (v === 'experimental') return 'warn'
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
    <div class="title-row">
      <div>
        <h2 class="display-h2">Skills</h2>
        <p class="lede">Reusable procedures the agent can call by name. Built-in and installed.</p>
      </div>
      <div class="header-actions">
        <Button variant="ghost" size="sm" onclick={() => goto('#/hub')}>Browse the Hub</Button>
        <Button variant="secondary" size="sm" onclick={refresh} loading={loading}>Refresh</Button>
      </div>
    </div>

    <div class="filters">
      <div class="search-wrap">
        <Input
          bind:value={query}
          fullWidth
          size="md"
          placeholder="Search skills…"
          aria-label="Search skills"
        >
          {#snippet leading()}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <circle cx="11" cy="11" r="7" />
              <path d="M21 21l-4.3-4.3" />
            </svg>
          {/snippet}
        </Input>
      </div>

      <div class="chips" role="tablist" aria-label="Filter skills">
        {#each [
          { id: 'all', label: 'All' },
          { id: 'built-in', label: 'Built-in' },
          { id: 'community', label: 'Community' },
          { id: 'experimental', label: 'Experimental' },
        ] as f (f.id)}
          <button
            type="button"
            role="tab"
            aria-selected={filter === f.id}
            class="chip"
            class:active={filter === f.id}
            onclick={() => (filter = f.id as Filter)}
          >
            {f.label}
          </button>
        {/each}
      </div>
    </div>
  </header>

  {#if error}
    <p class="error" role="alert">{error}</p>
  {/if}

  {#if loading && skills.length === 0}
    <div class="grid">
      {#each Array(6) as _, i (i)}
        <Card elevation={1} padding="md">
          <div class="skel-stack">
            <Skeleton width="36px" height="36px" rounded="md" count={1} />
            <Skeleton width="60%" height="14px" rounded="sm" count={1} />
            <Skeleton width="100%" height="12px" rounded="sm" count={1} />
            <Skeleton width="80%" height="12px" rounded="sm" count={1} />
          </div>
        </Card>
      {/each}
    </div>
  {:else if skills.length === 0}
    <EmptyState
      title="No skills installed yet"
      description="Skills are reusable procedures the agent can call by name. Install some from the Hub."
    >
      {#snippet icon()}
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M12 2v4M12 18v4M2 12h4M18 12h4M5.6 5.6l2.8 2.8M15.6 15.6l2.8 2.8M5.6 18.4l2.8-2.8M15.6 8.4l2.8-2.8" />
        </svg>
      {/snippet}
      {#snippet action()}
        <Button variant="primary" size="md" onclick={() => goto('#/hub')}>Browse the Hub</Button>
      {/snippet}
    </EmptyState>
  {:else if visible.length === 0}
    <EmptyState
      title="No skills match that filter"
      description="Try a different search term or switch the chip filter."
    />
  {:else}
    <div class="grid">
      {#each visible as s, i (s.id)}
        <div class="grid-item" style:--stagger-index={i}>
          <Card elevation="glass" interactive padding="md">
            <div class="card-top">
              <span class="icon-tile">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d={iconForCategory(categoryOf(s))} />
                </svg>
              </span>
              <div class="card-meta">
                <h3 class="skill-name">{s.name}</h3>
                <span class="skill-version mono">v{s.version}</span>
              </div>
              <Badge tone={trustTone(s.trust)} size="xs">{s.trust}</Badge>
            </div>

            {#if s.description}
              <p class="skill-desc">{s.description}</p>
            {/if}

            <div class="card-foot">
              <span class="skill-id mono">{s.id}</span>
              <div class="card-actions">
                <Button variant="primary" size="sm" onclick={() => goto(`#/skills`)}>Use</Button>
                <Button variant="ghost" size="sm" onclick={() => remove(s.id)}>Remove</Button>
              </div>
            </div>
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
  .header-actions {
    display: flex;
    gap: var(--space-2);
  }

  .filters {
    display: flex;
    gap: var(--space-3);
    align-items: center;
    flex-wrap: wrap;
  }
  .search-wrap {
    flex: 1 1 280px;
    min-width: 0;
  }

  .chips {
    display: inline-flex;
    gap: 2px;
    padding: 3px;
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
  }
  .chip {
    appearance: none;
    background: transparent;
    border: none;
    color: var(--text-muted);
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    padding: 5px 12px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition:
      background-color var(--transition-fast) ease,
      color var(--transition-fast) ease,
      transform var(--transition-fast) var(--ease-spring);
  }
  .chip:hover:not(.active) { color: var(--text); }
  .chip.active {
    background: var(--surface-3);
    color: var(--text);
    box-shadow: var(--shadow-xs);
  }
  .chip:active:not(.active) { transform: scale(0.97); }

  /* ── Skill grid ────────────────────────────────── */
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: var(--space-4);
  }
  .grid-item {
    animation: stagger-in var(--transition-base) var(--ease-out-expo) both;
    animation-delay: calc(var(--stagger-index, 0) * 50ms);
  }

  .card-top {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
  }
  .icon-tile {
    width: 36px;
    height: 36px;
    border-radius: var(--radius-md);
    background: var(--accent-faint);
    border: 1px solid var(--border);
    color: var(--accent);
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
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    margin: 0;
    line-height: var(--leading-tight);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .skill-version {
    font-size: var(--size-xs);
    color: var(--text-faint);
  }
  .skill-desc {
    font-size: var(--size-sm);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    margin: 0 0 var(--space-3) 0;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .card-foot {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    padding-top: var(--space-3);
    border-top: 1px solid var(--border);
  }
  .skill-id {
    font-size: var(--size-xs);
    color: var(--text-faint);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
  }
  .card-actions {
    display: flex;
    gap: var(--space-1);
    flex-shrink: 0;
  }

  /* ── Skeleton ─────────────────────────────────── */
  .skel-stack {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .error {
    color: var(--error);
    font-size: var(--size-sm);
    padding: var(--space-2) var(--space-3);
    background: var(--error-soft);
    border: 1px solid var(--border-danger);
    border-radius: var(--radius-md);
  }
</style>
