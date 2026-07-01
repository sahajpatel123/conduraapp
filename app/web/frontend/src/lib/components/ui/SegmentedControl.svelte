<script lang="ts">
  interface Option {
    value: string
    label: string
    icon?: import('svelte').Snippet
  }

  interface Props {
    options: Option[]
    value: string
    size?: 'sm' | 'md'
    onchange?: (value: string) => void
  }

  let { options, value = $bindable(), size = 'md', onchange }: Props = $props()

  let segEl = $state<HTMLDivElement | null>(null)

  function select(v: string): void {
    value = v
    onchange?.(v)
  }

  function onKeydown(e: KeyboardEvent): void {
    if (!segEl) return
    const buttons = segEl.querySelectorAll<HTMLButtonElement>('[role="radio"]')
    if (buttons.length === 0) return
    let current = Array.from(buttons).findIndex((b) => b === document.activeElement)
    // -1 means none focused — treat as first
    if (current < 0) current = 0

    let next: number | null = null

    if (e.key === 'ArrowRight') {
      e.preventDefault()
      next = (current + 1) % buttons.length
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault()
      next = (current - 1 + buttons.length) % buttons.length
    } else if (e.key === 'Home') {
      e.preventDefault()
      next = 0
    } else if (e.key === 'End') {
      e.preventDefault()
      next = buttons.length - 1
    }

    if (next !== null && next >= 0) {
      const btn = buttons[next]
      btn.focus()
      select(btn.getAttribute('data-value') ?? btn.textContent ?? '')
    }
  }
</script>

<div bind:this={segEl} class="seg seg-{size}" role="radiogroup" onkeydown={onKeydown}>
  {#each options as opt (opt.value)}
    <button
      type="button"
      role="radio"
      aria-checked={value === opt.value}
      data-value={opt.value}
      class="seg-btn"
      class:active={value === opt.value}
      onclick={() => select(opt.value)}
    >
      {#if opt.icon}<span class="seg-icon">{@render opt.icon()}</span>{/if}
      <span>{opt.label}</span>
    </button>
  {/each}
</div>

<style>
  .seg {
    display: inline-flex;
    align-items: center;
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: 3px;
    gap: 2px;
  }

  .seg-btn {
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
    display: inline-flex;
    align-items: center;
    gap: 6px;
    transition:
      background-color var(--transition-fast) ease,
      color var(--transition-fast) ease,
      transform var(--transition-fast) var(--ease-spring);
  }
  .seg-sm .seg-btn { font-size: var(--size-xs); padding: 4px 10px; }

  .seg-btn:hover:not(.active) { color: var(--text); }
  .seg-btn.active {
    background: var(--surface-3);
    color: var(--text);
    box-shadow: var(--shadow-xs);
  }
  .seg-btn:active:not(.active) { transform: scale(0.97); }

  .seg-icon { display: inline-flex; }
</style>