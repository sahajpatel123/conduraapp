<script lang="ts">
  // Toasts — bottom-right stack bound to the notifications store.
  //
  // Each toast has a tone (info / success / warn / error) and a
  // dismiss button. The store auto-dismisses non-sticky toasts after
  // the configured TTL.
  import { notifications } from '../stores/notifications.svelte'
  import { t } from '../i18n'

  type ToastTone = 'info' | 'success' | 'warn' | 'error'
</script>

<div class="toast-stack" aria-live="polite" aria-atomic="false">
  {#each notifications.list as n (n.id)}
    <div class="toast toast-{n.kind} anim-slide-up" role="status">
      <div class="toast-mark" aria-hidden="true"></div>
      <div class="toast-body">
        <div class="toast-title">{n.title}</div>
        {#if n.message}
          <div class="toast-message">{n.message}</div>
        {/if}
      </div>
      <button
        type="button"
        class="toast-close"
        aria-label={t('common.dismiss', 'Dismiss')}
        onclick={() => notifications.dismiss(n.id)}
      >
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <path d="M18 6L6 18M6 6l12 12" />
        </svg>
      </button>
    </div>
  {/each}
</div>

<style>
  .toast-stack {
    position: fixed;
    bottom: var(--space-5);
    right: var(--space-5);
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    z-index: var(--z-toast);
    width: 380px;
    max-width: calc(100vw - var(--space-7));
    pointer-events: none;
  }

  .toast {
    pointer-events: auto;
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
  }

  .toast-mark {
    flex-shrink: 0;
    width: 3px;
    align-self: stretch;
    border-radius: 2px;
  }

  .toast-info .toast-mark {
    background: var(--info);
  }
  .toast-success .toast-mark {
    background: var(--success);
  }
  .toast-warn .toast-mark {
    background: var(--warn);
  }
  .toast-error .toast-mark {
    background: var(--error);
  }

  .toast-body {
    flex: 1;
    min-width: 0;
  }
  .toast-title {
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    color: var(--text);
    line-height: var(--leading-normal);
  }
  .toast-message {
    margin-top: 2px;
    font-size: var(--size-xs);
    color: var(--text-muted);
    line-height: var(--leading-normal);
  }

  .toast-close {
    appearance: none;
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-sm);
    color: var(--text-faint);
    width: 22px;
    height: 22px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    flex-shrink: 0;
    transition:
      color var(--transition-fast) ease,
      background-color var(--transition-fast) ease;
  }
  .toast-close:hover {
    color: var(--text);
    background: var(--surface-3);
  }
  .toast-close:focus-visible {
    outline: 2px solid var(--border-focus);
    outline-offset: 1px;
  }
</style>