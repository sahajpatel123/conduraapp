<!--
  Navigation command palette — v1-styled fuzzy finder for routes.
-->
<script lang="ts">
  import Input from './Input.svelte'
  import Hairline from './Hairline.svelte'
  import Surface from './Surface.svelte'

  interface Item {
    id: string
    label: string
    hint?: string
    group?: string
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

  let {
    open = $bindable(false),
    items,
    placeholder = 'Search…',
    emptyMessage = 'No results',
    onclose,
  }: Props = $props()

  let query = $state('')
  let activeIdx = $state(0)

  const filtered = $derived.by(() => {
    if (!query.trim()) return items
    const q = query.toLowerCase()
    return items.filter(
      (it) =>
        it.label.toLowerCase().includes(q) ||
        (it.hint?.toLowerCase().includes(q) ?? false) ||
        (it.group?.toLowerCase().includes(q) ?? false)
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
    if (e.key === 'Escape') {
      close()
      return
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      activeIdx = Math.min(activeIdx + 1, Math.max(filtered.length - 1, 0))
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
      activeIdx = 0
    }
  })

  $effect(() => {
    query
    activeIdx = 0
  })
</script>

<svelte:window onkeydown={onKey} />

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="scrim" onclick={close} onkeydown={(e) => { if (e.key === 'Escape') close() }} aria-hidden="true"></div>
  <div class="shell" role="dialog" aria-label="Command palette">
    <Surface variant="overlay" padding="0" radius="xl">
      <div class="search">
        <Input
          variant="serif"
          size="lg"
          {placeholder}
          bind:value={query}
          ariaLabel="Search commands"
          autofocus
        />
      </div>
      <Hairline />
      <ul class="results" role="listbox" id="nav-palette-results">
        {#if filtered.length === 0}
          <li class="empty">{emptyMessage}</li>
        {:else}
          {#each filtered as item, i (item.id)}
            <li>
              <button
                type="button"
                class="row"
                class:row--active={i === activeIdx}
                role="option"
                aria-selected={i === activeIdx}
                disabled={item.disabled}
                onclick={() => run(item)}
              >
                <span class="row__label">{item.label}</span>
                {#if item.hint}
                  <span class="row__hint">{item.hint}</span>
                {/if}
                {#if item.group}
                  <span class="row__group">{item.group}</span>
                {/if}
              </button>
            </li>
          {/each}
        {/if}
      </ul>
    </Surface>
  </div>
{/if}

<style>
  .scrim {
    position: fixed;
    inset: 0;
    background: var(--surface-scrim);
    backdrop-filter: blur(2px);
    z-index: var(--z-overlay);
    animation: scrim-in var(--duration-base) var(--ease-accelerate) both;
  }

  .shell {
    position: fixed;
    top: 18%;
    left: 50%;
    transform: translateX(-50%);
    width: min(560px, calc(100vw - var(--space-8)));
    z-index: calc(var(--z-overlay) + 1);
    animation: pop-in var(--duration-base) var(--ease-decelerate) both;
  }

  .search {
    padding: var(--space-4) var(--space-5);
  }

  .results {
    list-style: none;
    margin: 0;
    padding: var(--space-2);
    max-height: 320px;
    overflow-y: auto;
  }

  .empty {
    padding: var(--space-5);
    text-align: center;
    color: var(--content-tertiary);
    font-size: var(--text-body-sm-size);
  }

  .row {
    width: 100%;
    display: grid;
    grid-template-columns: 1fr auto auto;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-3) var(--space-4);
    border: none;
    border-radius: var(--radius-md);
    background: transparent;
    color: var(--content-primary);
    font-family: var(--font-sans);
    font-size: var(--text-body-size);
    text-align: left;
    cursor: pointer;
    transition: background-color var(--duration-fast) var(--ease-standard);
  }

  .row:hover,
  .row--active {
    background: var(--action-tertiary-hover-bg);
  }

  .row__hint {
    color: var(--content-tertiary);
    font-size: var(--text-body-sm-size);
  }

  .row__group {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-muted);
  }

  @keyframes scrim-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes pop-in {
    from {
      opacity: 0;
      transform: translateX(-50%) scale(0.96);
    }
    to {
      opacity: 1;
      transform: translateX(-50%) scale(1);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .scrim,
    .shell {
      animation: none;
    }
  }
</style>
