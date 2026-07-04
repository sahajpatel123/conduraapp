<!--
  Toaster — toast container with stack management.

  Renders a stack of Toasts in the bottom-right corner. Each toast can
  be dismissed individually; the stack auto-arranges with 8px gap.

  Use a Svelte store or a simple prop array to drive the visible toasts.

  Props:
    toasts — array of {id, variant, title, message, icon?, duration?}
    onclose — handler when a toast is dismissed (passes the id)
-->
<script lang="ts">
  import Toast from './Toast.svelte';
  import type { IconName } from './icons/Icon.svelte';

  interface ToastData {
    id: string;
    variant?: 'info' | 'success' | 'warning' | 'error' | 'agent';
    title: string;
    message?: string;
    icon?: IconName;
    duration?: number;
  }

  interface Props {
    toasts?: ToastData[];
    onclose?: (id: string) => void;
  }

  let { toasts = [], onclose }: Props = $props();
</script>

<div class="toaster" aria-live="polite" aria-label="Notifications">
  {#each toasts as toast (toast.id)}
    <Toast
      variant={toast.variant}
      title={toast.title}
      message={toast.message}
      icon={toast.icon}
      duration={toast.duration}
      onclose={() => onclose?.(toast.id)}
    />
  {/each}
</div>

<style>
  .toaster {
    position: fixed;
    bottom: var(--space-5);
    right: var(--space-5);
    z-index: var(--z-toast);
    display: flex;
    flex-direction: column-reverse;  /* newest at the bottom */
    gap: var(--space-2);
    pointer-events: none;
    max-width: calc(100vw - var(--space-9));
  }

  .toaster :global(.toast) {
    pointer-events: auto;
  }
</style>