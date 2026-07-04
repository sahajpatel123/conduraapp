<script lang="ts">
  interface Tab {
    id: string
    label: string
    badge?: string | number
    disabled?: boolean
  }

  interface Props {
    tabs: Tab[]
    value: string
    onchange?: (id: string) => void
    size?: 'sm' | 'md'
  }

  let { tabs, value = $bindable(), onchange, size = 'md' }: Props = $props()

  function select(id: string): void {
    if (tabs.find(t => t.id === id)?.disabled) return
    value = id
    onchange?.(id)
  }

  function onKey(e: KeyboardEvent, idx: number): void {
    if (e.key === 'ArrowRight') {
      e.preventDefault()
      const next = tabs[idx + 1] ?? tabs[0]
      select(next.id)
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault()
      const prev = tabs[idx - 1] ?? tabs[tabs.length - 1]
      select(prev.id)
    }
  }
</script>

<div class="tabs tabs-{size}" role="tablist">
  {#each tabs as tab, i (tab.id)}
    <button
      type="button"
      role="tab"
      aria-selected={value === tab.id}
      aria-disabled={tab.disabled || undefined}
      tabindex={value === tab.id ? 0 : -1}
      class="tab"
      class:active={value === tab.id}
      class:disabled={tab.disabled}
      onclick={() => select(tab.id)}
      onkeydown={(e) => onKey(e, i)}
    >
      <span>{tab.label}</span>
      {#if tab.badge != null}<span class="tab-badge">{tab.badge}</span>{/if}
    </button>
  {/each}
</div>

<style>
  .tabs {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 4px;
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
  }

  .tab {
    appearance: none;
    background: transparent;
    border: none;
    color: var(--text-muted);
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-normal);
    padding: 6px 14px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    transition:
      background-color var(--transition-fast) ease,
      color var(--transition-fast) ease,
      transform var(--transition-fast) var(--ease-spring);
  }
  .tabs-sm .tab { font-size: var(--size-xs); padding: 4px 10px; }

  .tab:hover:not(.disabled):not(.active) {
    color: var(--text);
    background: var(--surface-2);
  }
  .tab.active {
    background: var(--surface-3);
    color: var(--text);
    box-shadow: var(--shadow-xs);
  }
  .tab:active:not(.disabled) {
    transform: scale(0.97);
  }
  .tab.disabled { opacity: 0.4; cursor: not-allowed; }

  .tab-badge {
    background: var(--surface-3);
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 1px 6px;
    border-radius: var(--radius-pill);
    line-height: 1.4;
  }
  .tab.active .tab-badge {
    background: var(--accent-soft);
    color: var(--accent);
  }
</style>