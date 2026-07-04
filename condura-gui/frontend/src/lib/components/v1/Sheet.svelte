<!--
  Sheet — slide-in panel from the right (or left/bottom).

  Used for: account details, sub-flows, detailed settings. Slides in
  over the current screen without dismissing it. The user can dismiss
  with the scrim, an X button, or Esc.

  Three positions:
    - right (default): slides in from the right edge, full height
    - left:  slides in from the left edge
    - bottom: slides up from the bottom (mobile-friendly)

  Props:
    open       — visibility (controlled)
    onclose    — handler for dismiss
    position   — 'right' | 'left' | 'bottom' (default: 'right')
    width      — sheet width in px (default: 480)
    title      — optional header title
    children   — sheet content
-->
<script lang="ts">
  import IconButton from './IconButton.svelte';

  type Position = 'right' | 'left' | 'bottom';

  interface Props {
    open: boolean;
    onclose?: () => void;
    position?: Position;
    width?: number;
    title?: string;
    children?: import('svelte').Snippet;
  }

  let { open, onclose, position = 'right', width = 480, title, children }: Props = $props();

  function handleKeydown(e: KeyboardEvent) {
    if (!open) return;
    if (e.key === 'Escape') {
      e.preventDefault();
      onclose?.();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="scrim"
    onclick={onclose}
    onkeydown={(e) => { if (e.key === 'Escape') onclose?.(); }}
    aria-hidden="true"
  ></div>

  <aside
    class="sheet sheet--{position}"
    role="dialog"
    aria-modal="true"
    aria-label={title ?? 'Sheet'}
    style="--sheet-width: {width}px;"
  >
    {#if title}
      <header class="sheet__head">
        <h2 class="sheet__title">{title}</h2>
        <IconButton name="x" label="Close" size={32} variant="ghost" onclick={onclose} />
      </header>
    {/if}

    <div class="sheet__body">
      {@render children?.()}
    </div>
  </aside>
{/if}

<style>
  .scrim {
    position: fixed;
    inset: 0;
    background-color: var(--surface-scrim);
    backdrop-filter: blur(2px);
    z-index: var(--z-overlay);
    animation: scrim-in var(--duration-base) var(--ease-accelerate) both;
  }

  @keyframes scrim-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .sheet {
    position: fixed;
    background-color: var(--surface-raised);
    border-color: var(--border-default);
    box-shadow: var(--shadow-4);
    z-index: calc(var(--z-overlay) + 1);
    display: flex;
    flex-direction: column;
  }

  /* Right position (default) */
  .sheet--right {
    top: 0;
    right: 0;
    bottom: 0;
    width: var(--sheet-width, 480px);
    max-width: 100vw;
    border-left: 1px solid var(--border-default);
    animation: sheet-in-right var(--duration-emphasized) var(--ease-decelerate) both;
  }

  @keyframes sheet-in-right {
    from { transform: translateX(100%); }
    to { transform: translateX(0); }
  }

  /* Left position */
  .sheet--left {
    top: 0;
    left: 0;
    bottom: 0;
    width: var(--sheet-width, 480px);
    max-width: 100vw;
    border-right: 1px solid var(--border-default);
    animation: sheet-in-left var(--duration-emphasized) var(--ease-decelerate) both;
  }

  @keyframes sheet-in-left {
    from { transform: translateX(-100%); }
    to { transform: translateX(0); }
  }

  /* Bottom position (mobile-friendly) */
  .sheet--bottom {
    left: 0;
    right: 0;
    bottom: 0;
    max-height: 80vh;
    border-top: 1px solid var(--border-default);
    border-radius: var(--radius-xl) var(--radius-xl) 0 0;
    animation: sheet-in-bottom var(--duration-emphasized) var(--ease-decelerate) both;
  }

  @keyframes sheet-in-bottom {
    from { transform: translateY(100%); }
    to { transform: translateY(0); }
  }

  /* Head */
  .sheet__head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4) var(--space-5);
    border-bottom: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  .sheet__title {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    font-weight: 400;
    color: var(--content-primary);
    margin: 0;
  }

  /* Body */
  .sheet__body {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-5);
  }

  @media (prefers-reduced-motion: reduce) {
    .scrim,
    .sheet--right,
    .sheet--left,
    .sheet--bottom {
      animation: none;
    }
  }
</style>