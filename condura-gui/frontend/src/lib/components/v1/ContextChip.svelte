<!--
  ContextChip — strip item showing what the agent noticed on screen.

  Per spec §9.1: the contextual strip at the top of the command surface
  shows up to 3 chips of detected screen elements. Tapping pre-fills the
  input with an interpretation.

  Props:
    label    — e.g., "this Slack thread", "this file", "this paragraph"
    active   — the chip currently being used to pre-fill
    onclick  — handler when tapped
-->
<script lang="ts">
  interface Props {
    label: string;
    active?: boolean;
    onclick?: (e: MouseEvent) => void;
  }

  let { label, active = false, onclick }: Props = $props();
</script>

<button
  class="context-chip"
  class:context-chip--active={active}
  type="button"
  onclick={onclick}
  aria-pressed={active}
>
  <span class="context-chip__dot" aria-hidden="true"></span>
  <span class="context-chip__label">{label}</span>
</button>

<style>
  .context-chip {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    height: 28px;
    padding: 0 var(--space-3);
    background-color: transparent;
    border: 1px solid var(--border-default);
    border-radius: var(--radius-pill);
    color: var(--content-secondary);
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    cursor: pointer;
    transition:
      background-color var(--duration-fast) var(--ease-standard),
      border-color var(--duration-fast) var(--ease-standard),
      color var(--duration-fast) var(--ease-standard);
    white-space: nowrap;
  }
  .context-chip:hover {
    background-color: var(--paper-warm-50);
    color: var(--content-primary);
  }
  .context-chip:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: 2px;
  }
  .context-chip--active {
    background-color: var(--plum-50);
    border-color: var(--plum-300);
    color: var(--plum-700);
  }
  .context-chip__dot {
    width: 5px;
    height: 5px;
    border-radius: var(--radius-pill);
    background-color: currentColor;
    flex-shrink: 0;
    opacity: 0.6;
  }
</style>