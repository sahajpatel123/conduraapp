<!--
  KeyCombo — renders a key combo like ⌘⇧Space in mono.

  Used in: hotkey display, status bar hints, keyboard shortcut docs.

  Props:
    combo — string like "⌘⇧Space" or "^Space"
-->
<script lang="ts">
  interface Props {
    combo: string;
    size?: 'sm' | 'md';
  }

  let { combo, size = 'sm' }: Props = $props();
</script>

<kbd class="combo combo--{size}" aria-label="Keyboard shortcut: {combo}">
  {#each combo.split('') as ch, i}
    {#if ch === '⌘' || ch === '⇧' || ch === '⌥' || ch === '^' || ch === '⏎' || ch === '⎋' || ch === '⏹' || ch === '⇥' || ch === '↑' || ch === '↓' || ch === '←' || ch === '→' || ch === '⏸'}
      <span class="combo__mod">{ch}</span>
    {:else}
      <span class="combo__char">{ch}</span>
    {/if}
  {/each}
</kbd>

<style>
  .combo {
    display: inline-flex;
    align-items: center;
    gap: 1px;
    padding: 2px 6px;
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-secondary);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
    vertical-align: middle;
    line-height: 1;
  }

  .combo--md {
    padding: 3px 8px;
    font-size: var(--text-body-sm-size);
  }

  .combo__mod {
    color: var(--content-primary);
    font-weight: 600;
  }

  .combo__char {
    color: var(--content-secondary);
  }
</style>