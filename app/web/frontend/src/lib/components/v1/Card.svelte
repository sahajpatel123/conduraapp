<!--
  Card — Surface with optional title + actions.

  Used for: agent action rows, autonomy-matrix rows, adaptive profile
  rows, suggestion cards (when not in a list).

  Per spec §11.3 (Settings): each setting has a sentence explaining
  why it exists. The Card primitive enables this pattern.

  Props:
    title       — card title
    description — one-line sentence explaining the card
    children    — main content slot
    actions     — action buttons slot
    variant     — surface variant (passed to Surface)
    padding     — inner padding
-->
<script lang="ts">
  import Surface from './Surface.svelte';

  interface Props {
    title?: string;
    description?: string;
    variant?: 'base' | 'sunken' | 'raised' | 'overlay' | 'inverted';
    padding?: string;
    children?: import('svelte').Snippet;
    actions?: import('svelte').Snippet;
  }

  let {
    title,
    description,
    variant = 'raised',
    padding = '4',
    children,
    actions,
  }: Props = $props();
</script>

<Surface {variant} {padding} radius="md">
  {#if title || description}
    <header class="card__head">
      {#if title}<h3 class="card__title">{title}</h3>{/if}
      {#if description}<p class="card__description">{description}</p>{/if}
    </header>
  {/if}

  {#if children}
    <div class="card__body">
      {@render children()}
    </div>
  {/if}

  {#if actions}
    <footer class="card__actions">
      {@render actions()}
    </footer>
  {/if}
</Surface>

<style>
  .card__head {
    margin-bottom: var(--space-3);
  }

  .card__title {
    font-family: var(--font-sans);
    font-size: var(--text-body-size);
    font-weight: 500;
    color: var(--content-primary);
    margin: 0 0 var(--space-1) 0;
  }

  .card__description {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    line-height: 1.5;
    margin: 0;
  }

  .card__body {
    /* Body slot — components handle their own spacing */
  }

  .card__actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
    margin-top: var(--space-3);
    padding-top: var(--space-3);
    border-top: 1px solid var(--border-subtle);
  }
</style>