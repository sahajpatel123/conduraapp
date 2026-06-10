<script lang="ts">
  import { notifications } from '../stores/notifications.svelte'
</script>

<div class="toast-container">
  {#each notifications.list as n (n.id)}
    <div class="toast toast-{n.kind}">
      <div class="toast-body">
        <strong>{n.title}</strong>
        <p>{n.message}</p>
      </div>
      <button class="toast-close" onclick={() => notifications.dismiss(n.id)}>×</button>
    </div>
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    bottom: var(--space-5);
    right: var(--space-5);
    display: flex;
    flex-direction: column-reverse;
    gap: var(--space-2);
    z-index: 9999;
    pointer-events: none;
  }

  @keyframes slideIn {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  .toast {
    pointer-events: auto;
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-left: 3px solid var(--color-accent);
    border-radius: var(--radius-xl);
    padding: var(--space-3) var(--space-4);
    min-width: 280px;
    max-width: 420px;
    box-shadow: var(--shadow-md);
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    animation: slideIn 0.4s var(--ease-out-expo);
  }

  .toast-info { 
    border-left-color: var(--color-accent);
  }
  .toast-error {
    border-left-color: #ef4444;
    box-shadow: var(--shadow-md), 0 0 16px rgba(239, 68, 68, 0.12);
  }

  .toast-body {
    flex: 1;
    min-width: 0;
  }
  .toast-body strong {
    font-size: var(--size-md);
    font-weight: 600;
    display: block;
    margin-bottom: var(--space-1);
    color: var(--color-text);
    letter-spacing: -0.01em;
  }
  .toast-body p {
    font-size: var(--size-sm);
    color: var(--color-text-muted);
    line-height: 1.45;
  }

  .toast-close {
    flex-shrink: 0;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    color: var(--color-text-faint);
    font-size: 14px;
    line-height: 1;
    padding: 0;
    cursor: pointer;
    transition: all var(--transition-base, 150ms ease);
  }
  .toast-close:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--color-text);
    border-color: rgba(255, 255, 255, 0.15);
  }
</style>
