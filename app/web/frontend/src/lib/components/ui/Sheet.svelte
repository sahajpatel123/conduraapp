<script lang="ts">
  import type { Snippet } from 'svelte'

  interface Props {
    open: boolean
    side?: 'right' | 'left'
    width?: string
    title?: string
    onclose?: () => void
    children?: Snippet
  }

  let { open = $bindable(false), side = 'right', width = '420px',
        title, onclose, children }: Props = $props()

  function close(): void {
    open = false
    onclose?.()
  }

  function onKey(e: KeyboardEvent): void {
    if (e.key === 'Escape' && open) close()
  }
</script>

<svelte:window onkeydown={onKey} />

{#if open}
  <div class="sheet-backdrop anim-fade" onclick={close} role="presentation"></div>
  <aside
    class="sheet sheet-{side} anim-slide-up"
    aria-label={title}
    style:--sheet-width={width}
  >
    {#if title}
      <header class="sheet-header">
        <h3 class="sheet-title">{title}</h3>
        <button type="button" class="sheet-close" aria-label="Close" onclick={close}>
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      </header>
    {/if}
    <div class="sheet-body">
      {#if children}{@render children()}{/if}
    </div>
  </aside>
{/if}

<style>
  .sheet-backdrop {
    position: fixed;
    inset: 0;
    background: var(--surface-overlay);
    backdrop-filter: var(--glass-blur-light);
    -webkit-backdrop-filter: var(--glass-blur-light);
    z-index: var(--z-modal);
  }

  .sheet {
    position: fixed;
    top: 0;
    bottom: 0;
    width: var(--sheet-width, 420px);
    max-width: 92vw;
    background: var(--surface-1);
    border-left: 1px solid var(--border-strong);
    box-shadow: var(--shadow-2xl);
    z-index: var(--z-modal);
    display: flex;
    flex-direction: column;
  }
  .sheet-right { right: 0; animation: sheet-slide-r var(--transition-slow) var(--ease-out-expo) both; }
  .sheet-left  { left:  0; border-left: none; border-right: 1px solid var(--border-strong);
                 animation: sheet-slide-l var(--transition-slow) var(--ease-out-expo) both; }

  @keyframes sheet-slide-r {
    from { transform: translateX(100%); }
    to   { transform: translateX(0); }
  }
  @keyframes sheet-slide-l {
    from { transform: translateX(-100%); }
    to   { transform: translateX(0); }
  }

  .sheet-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4) var(--space-5);
    border-bottom: 1px solid var(--border);
  }
  .sheet-title {
    font-family: var(--font-display);
    font-size: var(--size-lg);
    font-weight: var(--weight-medium);
    color: var(--text);
    letter-spacing: var(--tracking-tight);
  }
  .sheet-close {
    appearance: none;
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    width: 28px;
    height: 28px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }
  .sheet-close:hover { background: var(--surface-2); color: var(--text); }

  .sheet-body {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-5);
  }
</style>