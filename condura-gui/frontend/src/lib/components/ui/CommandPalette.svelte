<script lang="ts">
  import type { Snippet } from 'svelte'

  interface Item {
    id: string
    label: string
    hint?: string
    group?: string
    shortcut?: string
    icon?: Snippet
    onselect?: () => void
    disabled?: boolean
  }

  interface Props {
    open: boolean
    items: Item[]
    placeholder?: string
    emptyMessage?: string
    onclose?: () => void
  }

  let { open = $bindable(false), items, placeholder = 'Search…',
        emptyMessage = 'No results', onclose }: Props = $props()

  let query = $state('')
  let inputEl = $state<HTMLInputElement | null>(null)
  let activeIdx = $state(0)

  const filtered = $derived.by(() => {
    if (!query.trim()) return items
    const q = query.toLowerCase()
    return items.filter(it =>
      it.label.toLowerCase().includes(q) ||
      (it.hint?.toLowerCase().includes(q)) ||
      (it.group?.toLowerCase().includes(q))
    )
  })

  function close(): void {
    open = false
    query = ''
    activeIdx = 0
    onclose?.()
  }

  function run(item: Item): void {
    if (item.disabled) return
    item.onselect?.()
    close()
  }

  function onKey(e: KeyboardEvent): void {
    if (!open) return
    if (e.key === 'Escape') { close(); return }
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      activeIdx = Math.min(activeIdx + 1, filtered.length - 1)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      activeIdx = Math.max(activeIdx - 1, 0)
    } else if (e.key === 'Enter') {
      e.preventDefault()
      const item = filtered[activeIdx]
      if (item) run(item)
    }
  }

  $effect(() => {
    if (open) {
      queueMicrotask(() => inputEl?.focus())
      activeIdx = 0
    }
  })

  $effect(() => { query; activeIdx = 0 })
</script>

<svelte:window onkeydown={onKey} />

{#if open}
  <div class="cmd-backdrop anim-fade" onclick={close} role="presentation"></div>
  <div class="cmd-shell anim-pop" role="combobox" aria-expanded="true" aria-controls="cmd-results" aria-haspopup="listbox">
    <header class="cmd-input-row">
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" class="cmd-search-icon" aria-hidden="true">
        <circle cx="11" cy="11" r="7" />
        <path d="M21 21l-4.3-4.3" />
      </svg>
      <input
        bind:this={inputEl}
        bind:value={query}
        class="cmd-input"
        type="text"
        {placeholder}
        spellcheck="false"
        autocomplete="off"
      />
    </header>
    <div class="cmd-results">
      {#if filtered.length === 0}
        <div class="cmd-empty">{emptyMessage}</div>
      {:else}
        {#each filtered as item, i (item.id)}
          <button
            type="button"
            class="cmd-item"
            class:active={i === activeIdx}
            class:disabled={item.disabled}
            onmouseenter={() => activeIdx = i}
            onclick={() => run(item)}
          >
            {#if item.icon}<span class="cmd-icon">{@render item.icon()}</span>{/if}
            <span class="cmd-text">
              <span class="cmd-label">{item.label}</span>
              {#if item.hint}<span class="cmd-hint">{item.hint}</span>{/if}
            </span>
            {#if item.shortcut}<span class="cmd-shortcut">{item.shortcut}</span>{/if}
          </button>
        {/each}
      {/if}
    </div>
  </div>
{/if}

<style>
  .cmd-backdrop {
    position: fixed;
    inset: 0;
    background: var(--surface-overlay);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    z-index: var(--z-modal);
  }

  .cmd-shell {
    position: fixed;
    top: 18%;
    left: 50%;
    transform: translateX(-50%);
    width: 100%;
    max-width: 580px;
    max-height: 60vh;
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-2xl);
    z-index: var(--z-modal);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .cmd-input-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-4) var(--space-5);
    border-bottom: 1px solid var(--border);
  }
  .cmd-search-icon { color: var(--text-faint); flex-shrink: 0; }
  .cmd-input {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    color: var(--text);
    font-family: var(--font-sans);
    font-size: var(--size-md);
  }
  .cmd-input::placeholder { color: var(--text-faint); }

  .cmd-results {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-2);
  }

  .cmd-empty {
    padding: var(--space-7) var(--space-5);
    text-align: center;
    color: var(--text-faint);
    font-size: var(--size-sm);
  }

  .cmd-item {
    appearance: none;
    background: transparent;
    border: none;
    color: var(--text);
    width: 100%;
    text-align: left;
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: var(--space-3);
    font-family: var(--font-sans);
    transition: background-color var(--transition-fast) ease;
  }
  .cmd-item.active { background: var(--surface-3); }
  .cmd-item.disabled { opacity: 0.5; cursor: not-allowed; }

  .cmd-icon { color: var(--text-faint); flex-shrink: 0; display: inline-flex; }
  .cmd-text { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
  .cmd-label { font-size: var(--size-sm); color: var(--text); font-weight: var(--weight-medium); }
  .cmd-hint  { font-size: var(--size-xs); color: var(--text-muted); }
  .cmd-shortcut {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-faint);
    padding: 2px 6px;
    background: var(--surface-3);
    border-radius: var(--radius-sm);
    flex-shrink: 0;
  }
</style>