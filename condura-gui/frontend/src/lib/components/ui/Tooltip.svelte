<script lang="ts">
  import type { Snippet } from 'svelte'

  interface Props {
    text: string
    side?: 'top' | 'bottom' | 'left' | 'right'
    delay?: number
    children: Snippet
  }

  let { text, side = 'top', delay = 200, children }: Props = $props()

  let open = $state(false)
  let timer: ReturnType<typeof setTimeout> | null = null

  function show(): void {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => { open = true }, delay)
  }
  function hide(): void {
    if (timer) { clearTimeout(timer); timer = null }
    open = false
  }
</script>

<span
  class="tooltip-host"
  onmouseenter={show}
  onmouseleave={hide}
  onfocusin={show}
  onfocusout={hide}
  role="presentation"
>
  {@render children()}
  {#if open}
    <span class="tooltip tooltip-{side}" role="tooltip">{text}</span>
  {/if}
</span>

<style>
  .tooltip-host { position: relative; display: inline-flex; }

  .tooltip {
    position: absolute;
    z-index: var(--z-overlay);
    background: var(--surface-3);
    color: var(--text);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-sm);
    padding: 4px 8px;
    font-size: var(--size-xs);
    white-space: nowrap;
    box-shadow: var(--shadow-md);
    pointer-events: none;
    animation: fade-in var(--transition-fast) ease both;
  }

  .tooltip-top    { bottom: calc(100% + 6px); left: 50%; transform: translateX(-50%); }
  .tooltip-bottom { top:    calc(100% + 6px); left: 50%; transform: translateX(-50%); }
  .tooltip-left   { right:  calc(100% + 6px); top: 50%;  transform: translateY(-50%); }
  .tooltip-right  { left:   calc(100% + 6px); top: 50%;  transform: translateY(-50%); }
</style>