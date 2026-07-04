<script lang="ts">
  import type { Snippet } from 'svelte'

  interface Props {
    open: boolean
    title: string
    description?: string
    size?: 'sm' | 'md' | 'lg'
    onclose?: () => void
    children?: Snippet
    footer?: Snippet
  }

  let { open = $bindable(false), title, description, size = 'md',
        onclose, children, footer }: Props = $props()

  let dialogEl = $state<HTMLDivElement | null>(null)
  let previousFocus = $state<HTMLElement | null>(null)

  function close(): void {
    open = false
    onclose?.()
  }

  function trapFocus(e: KeyboardEvent): void {
    if (e.key === 'Escape') { close(); return }
    if (e.key !== 'Tab' || !dialogEl) return

    const focusable = dialogEl.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'
    )
    if (focusable.length === 0) return

    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    const active = document.activeElement as HTMLElement

    if (e.shiftKey && active === first) {
      e.preventDefault()
      last.focus()
    } else if (!e.shiftKey && active === last) {
      e.preventDefault()
      first.focus()
    }
  }

  $effect(() => {
    if (open) {
      previousFocus = document.activeElement as HTMLElement | null
      queueMicrotask(() => dialogEl?.focus())
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
      previousFocus?.focus()
    }
  })
</script>

<svelte:window onkeydown={open ? trapFocus : undefined} />

{#if open}
  <div class="dialog-backdrop anim-fade" onclick={close} role="presentation"></div>
  <div class="dialog-wrap" role="presentation">
    <div
      bind:this={dialogEl}
      class="dialog dialog-{size} anim-pop"
      role="dialog"
      aria-modal="true"
      aria-labelledby="dialog-title"
      aria-describedby={description ? 'dialog-description' : undefined}
      tabindex="-1"
    >
      <header class="dialog-header">
        <div class="dialog-titles">
          <h2 id="dialog-title" class="dialog-title">{title}</h2>
          {#if description}<p id="dialog-description" class="dialog-description">{description}</p>{/if}
        </div>
        <button
          type="button"
          class="dialog-close"
          aria-label="Close dialog"
          onclick={close}
        >
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      </header>
      <div class="dialog-body">
        {#if children}{@render children()}{/if}
      </div>
      {#if footer}
        <footer class="dialog-footer">
          {@render footer()}
        </footer>
      {/if}
    </div>
  </div>
{/if}

<style>
  .dialog-backdrop {
    position: fixed;
    inset: 0;
    background: var(--surface-overlay);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    z-index: var(--z-modal);
  }

  .dialog-wrap {
    position: fixed;
    inset: 0;
    z-index: var(--z-modal);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-5);
    pointer-events: none;
  }

  .dialog {
    pointer-events: auto;
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-2xl);
    display: flex;
    flex-direction: column;
    max-height: calc(100vh - var(--space-9));
    width: 100%;
    overflow: hidden;
  }

  .dialog-sm { max-width: 380px; }
  .dialog-md { max-width: 560px; }
  .dialog-lg { max-width: 760px; }

  .dialog-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-5) var(--space-5) var(--space-3);
  }

  .dialog-titles { display: flex; flex-direction: column; gap: 4px; min-width: 0; }

  .dialog-title {
    font-family: var(--font-display);
    font-size: var(--size-xl);
    font-weight: var(--weight-medium);
    color: var(--text);
    letter-spacing: var(--tracking-tight);
  }
  .dialog-description {
    font-size: var(--size-sm);
    color: var(--text-muted);
    line-height: var(--leading-normal);
  }

  .dialog-close {
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
    flex-shrink: 0;
    transition: background-color var(--transition-fast) ease, color var(--transition-fast) ease;
  }
  .dialog-close:hover { background: var(--surface-3); color: var(--text); }

  .dialog-body {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-3) var(--space-5) var(--space-5);
  }

  .dialog-footer {
    border-top: 1px solid var(--border);
    padding: var(--space-4) var(--space-5);
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: var(--space-2);
    background: var(--surface-1);
  }
</style>