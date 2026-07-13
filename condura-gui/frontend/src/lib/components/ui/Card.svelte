<script lang="ts">
  import type { Snippet } from 'svelte'

  type Elevation = 1 | 2 | 3 | 'glass' | 'glass-strong' | string

  interface Props {
    elevation?: Elevation
    interactive?: boolean
    padding?: 'none' | 'sm' | 'md' | 'lg'
    onclick?: (e: MouseEvent) => void
    children?: Snippet
    class?: string
  }

  let { elevation = 1, interactive = false, padding = 'md',
        onclick, children, class: extra = '' }: Props = $props()
</script>

<svelte:element
  this={onclick ? 'button' : 'div'}
  type={onclick ? 'button' : undefined}
  role={onclick ? 'button' : undefined}
  class="card card-pad-{padding} elev-{elevation}"
  class:interactive
  class:press={interactive}
  onclick={onclick as ((e: MouseEvent) => void) | undefined}
>
  {#if children}{@render children()}{/if}
</svelte:element>

<style>
  .card {
    display: flex;
    flex-direction: column;
    background: var(--surface-2);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    text-align: left;
    transition:
      background-color var(--transition-base) ease,
      border-color var(--transition-base) ease,
      transform var(--transition-base) var(--ease-spring),
      box-shadow var(--transition-base) ease;
    position: relative;
    isolation: isolate;
  }

  .elev-1 {
    box-shadow: none;
  }
  .elev-2 {
    background: var(--surface-2);
    border-color: var(--border-strong);
    box-shadow: none;
  }
  .elev-3 {
    background: var(--surface-3);
    border-color: var(--border-strong);
    box-shadow: none;
  }
  .elev-glass {
    background: var(--md-surface, var(--surface-2));
    backdrop-filter: none;
    -webkit-backdrop-filter: none;
    border-color: var(--border);
    box-shadow: none;
  }
  .elev-glass-strong {
    background: var(--md-surface, var(--surface-2));
    backdrop-filter: none;
    -webkit-backdrop-filter: none;
    border-color: var(--border-strong);
    box-shadow: none;
  }

  .card-pad-none { padding: 0; }
  .card-pad-sm   { padding: var(--space-3); }
  .card-pad-md   { padding: var(--space-5); }
  .card-pad-lg   { padding: var(--space-7); }

  .card.interactive {
    cursor: pointer;
  }
  .card.interactive:hover {
    background: var(--surface-3);
    border-color: var(--border-focus);
    transform: none;
    box-shadow: none;
  }
  .card.interactive:active {
    transform: translateY(0);
  }

  button.card {
    font-family: inherit;
    color: inherit;
  }
</style>