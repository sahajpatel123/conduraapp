<!--
  EmptyState — Equipment-at-rest composition.

  Per spec §8 (style agent): no greeting, no illustration, no "what can I
  help you with?" placeholder. One muted line. The affordance is implicit.

  Three empty states are defined:
    - Chat:        "Awaiting task."
    - Command:     "What would you like me to do?"
    - Action log:  "No actions yet. The agent is idle."

  Props:
    primary      — the main muted line (mono caps for status, serif for question)
    voice        — 'mono' | 'serif' (controls primary font family)
    secondary    — optional smaller line below
    children     — affordance slots (chips, buttons, etc.)
-->
<script lang="ts">
  interface Props {
    primary: string;
    voice?: 'mono' | 'serif';
    secondary?: string;
    children?: import('svelte').Snippet;
  }

  let { primary, voice = 'mono', secondary, children }: Props = $props();
</script>

<div class="empty">
  <p class="empty__primary empty__primary--{voice}">{primary}</p>
  {#if secondary}
    <p class="empty__secondary">{secondary}</p>
  {/if}
  {#if children}
    <div class="empty__affordance">
      {@render children()}
    </div>
  {/if}
</div>

<style>
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    padding: var(--space-9) var(--space-6);
    text-align: center;
  }

  /* Mono voice: status-line aesthetic (e.g., "Awaiting task.") */
  .empty__primary--mono {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    font-weight: 500;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--content-tertiary);
  }

  /* Serif voice: question (e.g., "What would you like me to do?") */
  .empty__primary--serif {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    line-height: 1.4;
    color: var(--content-secondary);
    font-style: italic;
    font-weight: 400;
  }

  .empty__secondary {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    max-width: 32em;
    line-height: 1.5;
  }

  .empty__affordance {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
    justify-content: center;
    margin-top: var(--space-4);
  }
</style>