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
  .toast {
    pointer-events: auto;
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    border-left: 3px solid var(--color-accent);
    border-radius: var(--radius-md);
    padding: var(--space-3) var(--space-4);
    min-width: 280px;
    max-width: 420px;
    box-shadow: var(--shadow-md);
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
  }
  .toast-info { border-left-color: var(--color-info); }
  .toast-success { border-left-color: var(--color-success); }
  .toast-warn { border-left-color: var(--color-warn); }
  .toast-error { border-left-color: var(--color-error); }
  .toast-body {
    flex: 1;
  }
  .toast-body strong {
    font-size: var(--size-md);
    display: block;
    margin-bottom: var(--space-1);
  }
  .toast-body p {
    font-size: var(--size-sm);
    color: var(--color-text-muted);
  }
  .toast-close {
    background: transparent;
    color: var(--color-text-faint);
    font-size: var(--size-xl);
    line-height: 1;
    padding: 0 var(--space-2);
  }
  .toast-close:hover {
    color: var(--color-text);
  }
</style>
