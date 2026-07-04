<!--
  Tip — small contextual hint with icon + text.

  Per spec: the agent is a guest, not an owner. It should suggest, not nag.
  Tips are how it suggests: brief, optional, dismissible, always at the
  right moment.

  Use cases:
    - "Press ⌘K to summon me" under the empty chat
    - "Last 24 hours" hint above the action replay
    - "Tap to expand" hints on collapsed rows
    - "3 of 7 permissions granted" contextual status

  Props:
    icon      — IconName from the library
    children  — the tip text (rendered as markdown if it has **bold** etc.)
    tone      — 'neutral' (default) | 'accent' | 'muted'
    onclose   — optional close handler; if provided, shows an X button
-->
<script lang="ts">
  import Icon, { type IconName } from './icons/Icon.svelte';
  import IconButton from './IconButton.svelte';

  interface Props {
    icon?: IconName;
    children?: import('svelte').Snippet;
    tone?: 'neutral' | 'accent' | 'muted';
    onclose?: () => void;
  }

  let { icon = 'sparkle', children, tone = 'neutral', onclose }: Props = $props();
</script>

<div class="tip tip--{tone}">
  <span class="tip__icon" aria-hidden="true">
    <Icon name={icon} size="xs" />
  </span>
  <span class="tip__text">{@render children?.()}</span>
  {#if onclose}
    <span class="tip__action">
      <IconButton name="x" label="Dismiss" size={20} variant="ghost" onclick={onclose} />
    </span>
  {/if}
</div>

<style>
  .tip {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-pill);
    background-color: transparent;
    border: 1px solid transparent;
    font-family: var(--font-sans);
    font-size: var(--text-caption-size);
    line-height: 1.4;
    color: var(--content-tertiary);
    /* Subtle entrance — slides up and fades in */
    animation: tip-in var(--duration-base) var(--ease-decelerate) both;
  }

  @keyframes tip-in {
    from {
      opacity: 0;
      transform: translateY(4px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* Neutral tone — most common. Barely-there. */
  .tip--neutral {
    color: var(--content-tertiary);
  }

  /* Accent — uses the plum for invitation */
  .tip--accent {
    color: var(--content-secondary);
    background-color: var(--plum-50);
    border-color: var(--plum-200);
  }

  /* Muted — even quieter */
  .tip--muted {
    color: var(--content-muted);
  }

  .tip__icon {
    display: inline-flex;
    align-items: center;
    color: inherit;
    opacity: 0.8;
  }

  .tip__text {
    /* Allow the text to be selectable for the curious user */
    user-select: text;
  }

  /* Strong emphasis inline — the **word** pattern */
  .tip__text :global(strong) {
    color: var(--content-primary);
    font-weight: 500;
  }

  .tip__action {
    margin-left: var(--space-1);
    margin-right: -6px;
  }

  @media (prefers-reduced-motion: reduce) {
    .tip {
      animation: none;
    }
  }
</style>