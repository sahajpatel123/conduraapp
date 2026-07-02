<script lang="ts">
  /**
   * EmptyState — the shared empty-state primitive.
   *
   * Per MOAT §2.4: every empty state's copy has three lines:
   *   1. what this area is
   *   2. why it might be empty
   *   3. the one action that fills it (optional — the Audit S1 empty
   *      state passes no action because the audit logs what happens;
   *      it doesn't prompt the user to make something happen)
   */
  import type { Snippet } from 'svelte';

  let {
    what,
    why,
    action,
    children,
  }: {
    what: string;
    why: string;
    action?: Snippet;
    children?: Snippet;
  } = $props();
</script>

<div class="empty">
  <p class="what">{what}</p>
  <p class="why">{why}</p>
  {#if action}
    <div class="actions">{@render action()}</div>
  {/if}
  {@render children?.()}
</div>

<style>
  .empty {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-3);
    padding: var(--space-5) 0 var(--space-6);
    max-width: 52ch;
  }
  .what,
  .why {
    font-family: var(--font-display);
    font-style: italic;
    line-height: 1.55;
    color: var(--content);
    margin: 0;
  }
  .what {
    font-size: 17px;
    letter-spacing: -0.01em;
  }
  .why {
    font-size: 15px;
    color: var(--content-faint);
  }
  .actions {
    display: flex;
    gap: var(--space-3);
    margin-top: var(--space-2);
  }
</style>