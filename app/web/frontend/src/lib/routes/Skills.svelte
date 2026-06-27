<script lang="ts">
  import { ipc } from '../ipc/client'
  import { onMount } from 'svelte'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import { t } from '../i18n'

  let skills = $state<Array<{ id: string; name: string; version: string; trust: string; source?: string; description?: string }>>([])
  let cursor = $state(0)
  let loading = $state(false)
  let error = $state<string | null>(null)
  let confirmOpen = $state(false)
  let confirmAction = $state<(() => void) | null>(null)

  async function refresh() {
    loading = true
    error = null
    try {
      skills = (await ipc.skillsList(100)) || []
      cursor = 0
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  }

  function remove(id: string) {
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

  onMount(refresh)
</script>

<div class="skills-page">
  <header class="page-header">
    <h2>{t('skills.title')}</h2>
    <p class="muted">{t('skills.subtitle', skills.length)}</p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if skills.length === 0}
    <div class="empty-state">
      <p>{@html t('skills.empty_html')}</p>
    </div>
  {:else}
    <ul class="skill-list">
      {#each skills as s, i}
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <li class="skill-card glass-card stagger-item" class:selected={i === cursor} style="--stagger-index: {i}" onclick={() => cursor = i} onkeydown={() => {}}>
          <div class="row">
            <strong>{s.name}</strong>
            <span class="version">v{s.version}</span>
            <span class="badge trust-badge">{s.trust}</span>
            {#if s.source}<span class="source">{t('skills.from', s.source)}</span>{/if}
            <span class="spacer"></span>
            <button class="btn btn-danger btn-xs" onclick={(e) => { e.stopPropagation(); remove(s.id) }}>{t('skills.delete')}</button>
          </div>
          {#if s.description}
            <p class="desc">{s.description}</p>
          {/if}
          <span class="id mono">{s.id}</span>
        </li>
      {/each}
    </ul>
  {/if}

  <div class="actions">
    <button class="btn btn-ghost" onclick={refresh} disabled={loading}>{loading ? t('skills.refreshing') : t('skills.refresh')}</button>
  </div>
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title={t('skills.delete')}
  message={t('skills.delete_confirm', '')}
  danger={true}
  onconfirm={() => confirmAction?.()}
/>

<style>
  .skills-page {
    padding: var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width);
    margin: 0 auto;
  }
  .page-header {
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .skill-list {
    list-style: none;
    padding: 0;
    margin: var(--space-4) 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .skill-card {
    padding: var(--space-3) var(--space-4);
    cursor: pointer;
    transition: border-color var(--transition-base), background var(--transition-base);
  }
  .skill-card:hover,
  .skill-card.selected {
    border-color: var(--color-border-accent);
    background: var(--glass-bg-hover);
    box-shadow: var(--shadow-glow-accent);
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
  .source {
    color: var(--color-text-faint);
    font-size: var(--size-xs);
  }
  .desc {
    margin: var(--space-1) 0;
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-relaxed);
  }
  .id {
    font-size: var(--size-xs);
    color: var(--color-text-faint);
  }
  .error {
    color: var(--color-error);
    font-size: var(--size-sm);
  }
  .actions {
    margin-top: var(--space-3);
  }
</style>
