<script lang="ts">
  import type { HTMLInputAttributes } from 'svelte/elements'

  interface Props extends Omit<HTMLInputAttributes, 'class' | 'size'> {
    value?: string
    label?: string
    hint?: string
    error?: string
    fullWidth?: boolean
    leading?: import('svelte').Snippet
    trailing?: import('svelte').Snippet
    size?: 'sm' | 'md' | 'lg'
  }

  let {
    value = $bindable(''),
    label,
    hint,
    error,
    fullWidth = false,
    leading,
    trailing,
    size = 'md',
    ...rest
  }: Props = $props()
</script>

<label class="input-wrap input-{size}" class:input-full={fullWidth} class:has-error={!!error}>
  {#if label}<span class="input-label">{label}</span>{/if}
  <span class="input-shell">
    {#if leading}<span class="input-affix input-leading">{@render leading()}</span>{/if}
    <input class="input-control" bind:value {...rest} />
    {#if trailing}<span class="input-affix input-trailing">{@render trailing()}</span>{/if}
  </span>
  {#if error}
    <span class="input-error" role="alert">{error}</span>
  {:else if hint}
    <span class="input-hint">{hint}</span>
  {/if}
</label>

<style>
  .input-wrap {
    display: inline-flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }
  .input-full { width: 100%; }

  .input-label {
    font-size: var(--size-xs);
    color: var(--text-muted);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-wide);
    text-transform: uppercase;
    padding-left: 2px;
  }

  .input-shell {
    display: flex;
    align-items: center;
    background: var(--surface-1);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    transition:
      border-color var(--transition-fast) ease,
      background-color var(--transition-fast) ease,
      box-shadow var(--transition-fast) ease;
    overflow: hidden;
  }
  .input-shell:focus-within {
    border-color: var(--border-focus);
    background: var(--surface-2);
    box-shadow: 0 0 0 3px var(--accent-soft);
  }

  .input-control {
    flex: 1;
    min-width: 0;
    background: transparent;
    border: none;
    outline: none;
    color: var(--text);
    font-family: var(--font-sans);
    font-size: var(--size-md);
    padding: 0 var(--space-3);
    height: 36px;
  }
  .input-control::placeholder { color: var(--text-faint); }

  .input-sm .input-control { font-size: var(--size-sm); height: 30px; }
  .input-lg .input-control { font-size: var(--size-lg); height: 44px; }

  .input-affix {
    display: inline-flex;
    align-items: center;
    color: var(--text-faint);
    padding: 0 var(--space-3);
  }
  .input-leading { border-right: 1px solid var(--border); }
  .input-trailing { border-left: 1px solid var(--border); }

  .input-hint { font-size: var(--size-xs); color: var(--text-faint); padding-left: 2px; }
  .input-error { font-size: var(--size-xs); color: var(--error); padding-left: 2px; }

  .has-error .input-shell {
    border-color: var(--border-danger);
    box-shadow: 0 0 0 3px var(--error-soft);
  }
</style>