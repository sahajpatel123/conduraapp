<script lang="ts">
  import { t } from '../i18n'

  let {
    open = $bindable(false),
    title = '',
    message = '',
    confirmLabel = '',
    cancelLabel = '',
    danger = false,
    onconfirm,
    oncancel
  }: {
    open: boolean
    title: string
    message: string
    confirmLabel?: string
    cancelLabel?: string
    danger?: boolean
    onconfirm: () => void
    oncancel?: () => void
  } = $props()

  let dialogEl: HTMLDivElement | null = $state(null)
  let cancelBtn: HTMLButtonElement | null = $state(null)
  let confirmBtn: HTMLButtonElement | null = $state(null)
  // The element that had focus before the dialog opened. We restore
  // focus to it on close so keyboard users don't lose their place.
  let previouslyFocused: HTMLElement | null = null

  // Focus management: when `open` flips to true, capture the
  // currently-focused element and move focus to the cancel button
  // (safer default for destructive confirmations — user must take
  // an explicit action to confirm, not just hit Enter). When `open`
  // flips to false, restore focus.
  $effect(() => {
    if (open) {
      // Capture in a microtask so the dialog element exists in the DOM
      // before we query it.
      if (typeof document !== 'undefined') {
        previouslyFocused = (document.activeElement as HTMLElement | null) ?? null
        // Focus the cancel button (safer default). If for some reason
        // it's not rendered, focus the dialog itself as a fallback.
        queueMicrotask(() => {
          if (cancelBtn) {
            cancelBtn.focus()
          } else if (dialogEl) {
            dialogEl.focus()
          }
        })
      }
    } else {
      // Restore focus to the element that had it before the dialog opened.
      if (previouslyFocused && typeof document !== 'undefined' && document.contains(previouslyFocused)) {
        previouslyFocused.focus()
      }
      previouslyFocused = null
    }
  })

  function focusableElements(): HTMLElement[] {
    if (!dialogEl) return []
    // Standard "tabbable" selector — covers all interactive elements.
    // Use a querySelectorAll inside the dialog so Tab only cycles
    // through the dialog's controls, not the background page.
    const sel = 'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    return Array.from(dialogEl.querySelectorAll<HTMLElement>(sel)).filter(
      (el) => !el.hasAttribute('disabled') && el.tabIndex !== -1
    )
  }

  function handleKeydown(e: KeyboardEvent): void {
    if (e.key === 'Escape') {
      e.preventDefault()
      handleCancel()
      return
    }
    if (e.key === 'Tab') {
      // Focus trap: cycle within the dialog's focusable elements.
      const focusables = focusableElements()
      if (focusables.length === 0) {
        // Nothing focusable inside — keep focus on the dialog.
        e.preventDefault()
        dialogEl?.focus()
        return
      }
      const first = focusables[0]
      const last = focusables[focusables.length - 1]
      const active = document.activeElement as HTMLElement | null
      if (e.shiftKey) {
        if (active === first || !dialogEl?.contains(active)) {
          e.preventDefault()
          last.focus()
        }
      } else {
        if (active === last || !dialogEl?.contains(active)) {
          e.preventDefault()
          first.focus()
        }
      }
    }
    // Enter on the cancel button is handled by the button's default
    // click. Enter on the confirm button is handled the same way. We
    // do NOT add a global Enter handler here because that would also
    // fire when the user is typing into a field (none today, but
    // future-proof).
  }

  function handleConfirm(): void {
    open = false
    onconfirm()
  }

  function handleCancel(): void {
    open = false
    oncancel?.()
  }
</script>

{#if open}
  <div
    class="confirm-backdrop"
    onclick={handleCancel}
    onkeydown={handleKeydown}
    role="presentation"
  >
    <div
      bind:this={dialogEl}
      class="confirm-dialog"
      class:danger
      role="alertdialog"
      tabindex="-1"
      aria-modal="true"
      aria-labelledby="confirm-title"
      aria-describedby="confirm-message"
      onclick={(e) => e.stopPropagation()}
      onkeydown={handleKeydown}
    >
      <h3 id="confirm-title">{title}</h3>
      <p id="confirm-message" class="confirm-message">{message}</p>
      <div class="confirm-actions">
        <button
          bind:this={cancelBtn}
          class="btn btn-ghost"
          type="button"
          onclick={handleCancel}
        >
          {cancelLabel || t('common.cancel')}
        </button>
        <button
          bind:this={confirmBtn}
          class="btn"
          class:btn-danger={danger}
          class:btn-primary={!danger}
          type="button"
          onclick={handleConfirm}
        >
          {confirmLabel || t('common.confirm')}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .confirm-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(20, 17, 11, 0.4);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: var(--z-modal);
    animation: backdrop-in var(--transition-base) ease both;
  }

  @keyframes backdrop-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .confirm-dialog {
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur-heavy);
    -webkit-backdrop-filter: var(--glass-blur-heavy);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-xl);
    padding: var(--space-5);
    max-width: 400px;
    width: calc(100% - 32px);
    box-shadow: var(--shadow-lg);
    animation: modal-in var(--transition-spring-soft) var(--ease-out-expo) both;
  }

  .confirm-dialog.danger {
    border-color: rgba(239, 68, 68, 0.4);
  }

  @keyframes modal-in {
    from {
      opacity: 0;
      transform: scale(0.95) translateY(8px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }

  .confirm-dialog h3 {
    font-size: var(--size-lg);
    font-weight: var(--weight-semibold);
    margin-bottom: var(--space-2);
  }

  .confirm-message {
    font-size: var(--size-sm);
    color: var(--color-text-muted);
    line-height: var(--leading-relaxed);
    margin-bottom: var(--space-4);
  }

  .confirm-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
  }

  /*
   * Sighted-keyboard focus ring: when a button inside the dialog
   * receives focus, show a visible ring. The default browser focus
   * outline is often removed by the global .btn reset; this restores
   * it for keyboard users.
   */
  .confirm-dialog button:focus-visible {
    outline: 2px solid var(--color-synapse, #0b6e4f);
    outline-offset: 2px;
  }
</style>
