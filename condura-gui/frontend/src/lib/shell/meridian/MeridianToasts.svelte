<script lang="ts">
  /**
   * Meridian toast stack — mist/cobalt surface for notifications store.
   * Spend warnings, Sync notes, etc. land here (product shell had none).
   */
  import { notifications } from '../../stores/notifications.svelte'
</script>

{#if notifications.list.length}
  <div class="toast-stack" aria-live="polite" aria-atomic="false">
    {#each notifications.list as n (n.id)}
      <div class="toast" data-kind={n.kind} role="status">
        <span class="mark" aria-hidden="true"></span>
        <div class="body">
          <strong>{n.title}</strong>
          {#if n.message}
            <p>{n.message}</p>
          {/if}
        </div>
        <button
          type="button"
          class="x"
          aria-label="Dismiss"
          onclick={() => notifications.dismiss(n.id)}
        >
          ×
        </button>
      </div>
    {/each}
  </div>
{/if}

<style>
  .toast-stack {
    position: fixed;
    right: 16px;
    bottom: calc(72px + env(safe-area-inset-bottom, 0px));
    z-index: 90;
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: min(360px, calc(100vw - 32px));
    pointer-events: none;
  }
  .toast {
    pointer-events: auto;
    display: grid;
    grid-template-columns: 3px 1fr auto;
    gap: 12px;
    align-items: start;
    padding: 10px 10px 10px 0;
    border-radius: 10px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
    box-shadow: none;
    animation: toast-in 220ms var(--md-ease) both;
  }
  .mark {
    width: 4px;
    align-self: stretch;
    border-radius: 0 4px 4px 0;
    background: var(--md-cobalt);
  }
  .toast[data-kind='warn'] .mark {
    background: #c4892a;
  }
  .toast[data-kind='error'] .mark {
    background: var(--md-halt);
  }
  .toast[data-kind='success'] .mark {
    background: var(--md-live);
  }
  .body {
    min-width: 0;
    padding-top: 2px;
  }
  .body strong {
    display: block;
    font-size: 13px;
    font-weight: 600;
    letter-spacing: -0.02em;
    color: var(--md-ink);
  }
  .body p {
    margin: 4px 0 0;
    font-size: 12px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .x {
    width: 26px;
    height: 26px;
    border-radius: 7px;
    font-size: 16px;
    line-height: 1;
    color: var(--md-ink-faint);
    cursor: pointer;
  }
  .x:hover {
    background: var(--md-stage);
    color: var(--md-ink);
  }
  .x:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  @keyframes toast-in {
    from {
      opacity: 0;
      transform: translateY(10px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .toast {
      animation: none;
    }
  }
</style>
