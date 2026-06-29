<script lang="ts">
  import type { Snippet } from 'svelte'

  interface Props {
    title: string
    description?: string
    icon?: Snippet
    action?: Snippet
  }

  let { title, description, icon, action }: Props = $props()
</script>

<div class="empty">
  {#if icon}<div class="empty-icon">{@render icon()}</div>{/if}
  <h3 class="empty-title">{title}</h3>
  {#if description}<p class="empty-description">{description}</p>{/if}
  {#if action}<div class="empty-action">{@render action()}</div>{/if}
</div>

<style>
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: var(--space-10) var(--space-5);
    gap: var(--space-3);
    animation: fade-in-scale var(--transition-slow) var(--ease-out-expo) both;
  }

  .empty-icon {
    width: 56px;
    height: 56px;
    border-radius: var(--radius-xl);
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-faint);
    margin-bottom: var(--space-3);
  }

  .empty-title {
    font-family: var(--font-display);
    font-size: var(--size-xl);
    font-weight: var(--weight-medium);
    color: var(--text);
    letter-spacing: var(--tracking-tight);
  }

  .empty-description {
    color: var(--text-muted);
    font-size: var(--size-md);
    line-height: var(--leading-normal);
    max-width: 380px;
  }

  .empty-action { margin-top: var(--space-3); }
</style>