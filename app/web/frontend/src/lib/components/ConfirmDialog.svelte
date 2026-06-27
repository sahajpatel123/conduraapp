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

  function handleConfirm(): void {
    open = false
    onconfirm()
  }

  function handleCancel(): void {
    open = false
    oncancel?.()
  }

  function handleKeydown(e: KeyboardEvent): void {
    if (e.key === 'Escape') handleCancel()
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
        <button class="btn btn-ghost" type="button" onclick={handleCancel}>
          {cancelLabel || t('common.cancel')}
        </button>
        <button class="btn" class:btn-danger={danger} class:btn-primary={!danger} type="button" onclick={handleConfirm}>
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
</style>
