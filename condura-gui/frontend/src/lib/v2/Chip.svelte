<!--
  Chip — Condura v2 selectable pill primitive.

  Used for the "never touch" domain chips in the floating interview,
  for the quick-prompt chips in the chat empty state, and anywhere
  else a single-line, single-token, toggleable label is needed.

  Two modes:
    - default  (a single on/off state, controlled via `on` prop)
    - radio    (also reads `selected` prop — alias)

  Props:
    on:         boolean — selected state
    onclick?:   (e: MouseEvent) => void
    variant?:   'default' | 'accent' | 'signal-go' | 'signal-warn' | 'signal-stop'
    size?:      'small' | 'default'
    disabled?:  boolean
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  let {
    on = false as boolean,
    onclick = undefined as ((e: MouseEvent) => void) | undefined,
    variant = 'default' as 'default' | 'accent' | 'signal-go' | 'signal-warn' | 'signal-stop',
    size = 'default' as 'small' | 'default',
    disabled = false as boolean,
    children,
    class: klass = '',
  }: {
    on?: boolean
    onclick?: (e: MouseEvent) => void
    variant?: 'default' | 'accent' | 'signal-go' | 'signal-warn' | 'signal-stop'
    size?: 'small' | 'default'
    disabled?: boolean
    children?: Snippet
    class?: string
  } = $props()

  function handle(e: MouseEvent) {
    if (disabled) return
    onclick?.(e)
  }
</script>

<button
  data-v2-chip
  data-on={on}
  data-variant={variant}
  data-size={size}
  disabled={disabled}
  onclick={handle}
  class={klass}
  type="button"
>
  {@render children?.()}
</button>

<style>
  [data-v2-chip] {
    font-family: var(--v2-font-sans);
    font-weight: 500;
    line-height: 1;
    letter-spacing: 0;
    border-radius: var(--v2-radius-pill);
    border: 1px solid var(--v2-rule);
    background: transparent;
    color: var(--v2-ink);
    cursor: pointer;
    transition:
      background-color var(--v2-dur-fast) var(--v2-ease-out-soft),
      border-color     var(--v2-dur-fast) var(--v2-ease-out-soft),
      color            var(--v2-dur-fast) var(--v2-ease-out-soft),
      transform        var(--v2-dur-fast) var(--v2-ease-out-soft);
  }
  [data-v2-chip][data-size='default'] {
    padding: var(--v2-space-2) var(--v2-space-3);
    font-size: var(--v2-text-12);
  }
  [data-v2-chip][data-size='small'] {
    padding: 4px var(--v2-space-2);
    font-size: 11px;
  }

  /* ON state — color depends on variant */
  [data-v2-chip][data-on='true'][data-variant='default'] {
    border-color: var(--v2-ink);
    background: var(--v2-ink);
    color: var(--v2-paper);
  }
  [data-v2-chip][data-on='true'][data-variant='accent'] {
    border-color: var(--v2-accent);
    background: color-mix(in srgb, var(--v2-accent) 14%, transparent);
    color: var(--v2-accent-ink);
  }
  [data-v2-chip][data-on='true'][data-variant='signal-go'] {
    border-color: var(--v2-signal-go);
    background: color-mix(in srgb, var(--v2-signal-go) 14%, transparent);
    color: var(--v2-signal-go);
  }
  [data-v2-chip][data-on='true'][data-variant='signal-warn'] {
    border-color: var(--v2-signal-warn);
    background: color-mix(in srgb, var(--v2-signal-warn) 14%, transparent);
    color: var(--v2-signal-warn);
  }
  [data-v2-chip][data-on='true'][data-variant='signal-stop'] {
    border-color: var(--v2-signal-stop);
    background: color-mix(in srgb, var(--v2-signal-stop) 14%, transparent);
    color: var(--v2-signal-stop);
  }

  /* Hover (off-state only): a faint fill hints at interactability */
  [data-v2-chip][data-on='false']:hover {
    background: color-mix(in srgb, var(--v2-rule) 40%, transparent);
  }

  /* Press */
  [data-v2-chip]:active {
    transform: translateY(1px);
  }

  /* Disabled */
  [data-v2-chip]:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Focus-visible is handled by reset.css: scope-wide rule */
</style>
