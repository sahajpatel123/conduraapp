<script lang="ts">
  import { ipc } from '../ipc/client'
  import { onMount } from 'svelte'

  let skills = $state<Array<{ id: string; name: string; version: string; trust: string; source?: string; description?: string }>>([])
  let cursor = $state(0)
  let loading = $state(false)
  let error = $state<string | null>(null)

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

  async function remove(id: string) {
    if (!confirm(`Delete skill ${id}?`)) return
    try {
      await ipc.skillsDelete(id)
      await refresh()
    } catch (e) {
      error = String(e)
    }
  }

  onMount(refresh)
</script>

<div class="skills-page">
  <header>
    <h2>Installed Skills</h2>
    <p class="muted">Skills that the agent can invoke. {skills.length} installed.</p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if skills.length === 0}
    <p class="muted">
      No skills installed yet. Use the <strong>Hub</strong> tab to search and install community skills,
      or use <code>synaptic hub search &lt;query&gt;</code> from the command line.
    </p>
  {:else}
    <ul>
      {#each skills as s, i}
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <li class:selected={i === cursor} onclick={() => cursor = i} onkeydown={() => {}}>
          <div class="row">
            <strong>{s.name}</strong>
            <span class="version">v{s.version}</span>
            <span class="trust">[{s.trust}]</span>
            {#if s.source}<span class="source">from {s.source}</span>{/if}
            <span class="spacer"></span>
            <button class="danger" onclick={(e) => { e.stopPropagation(); remove(s.id) }}>Delete</button>
          </div>
          {#if s.description}
            <p class="desc">{s.description}</p>
          {/if}
          <span class="id mono">{s.id}</span>
        </li>
      {/each}
    </ul>
  {/if}

  <button onclick={refresh} disabled={loading}>{loading ? 'Refreshing…' : 'Refresh'}</button>
</div>

<style>
  .skills-page { padding: 16px; }
  ul { list-style: none; padding: 0; margin: 12px 0; }
  li { padding: 12px; border: 1px solid transparent; border-radius: 6px; cursor: pointer; }
  li.selected { border-color: var(--accent, #4a9eff); background: var(--hover, rgba(74, 158, 255, 0.08)); }
  .row { display: flex; align-items: baseline; gap: 8px; }
  .spacer { flex: 1; }
  .version, .trust, .source { color: var(--muted, #888); font-size: 12px; }
  .desc { margin: 6px 0; }
  .id { font-size: 11px; opacity: 0.5; font-family: ui-monospace, monospace; }
  .error { color: var(--error, #c0392b); }
  .muted { color: var(--muted, #888); }
  .danger { background: var(--danger, #c0392b); color: white; }
  .mono { font-family: ui-monospace, monospace; }
  button { padding: 4px 12px; }
</style>
