<!--
  EmptyState — refined equipment-at-rest composition.

  Per spec §8 (style agent): no greeting, no illustration, no "what can I
  help you with?" placeholder. One muted line. The affordance is implicit.

  But "no illustration" doesn't mean "no atmosphere." A truly considered
  empty state has:
    - a subtle ambient pattern that responds to the Pulse's rhythm
    - refined typography hierarchy (mono caps for the status line,
      serif italic for the question, sans for the description)
    - an affordance that feels like an invitation, not a button
    - perfectly balanced negative space

  Three voice modes:
    - mono:    "Equipment at rest." (status-line aesthetic, e.g., chat idle)
    - serif:   "What would you like me to do?" (a question)
    - sans:    "Nothing here yet." (a quiet notice)

  Props:
    primary      — the main muted line
    voice        — 'mono' | 'serif' | 'sans'
    secondary    — optional smaller line below
    pulse        — optional ambient Pulse in the background (subtle, slow)
    children     — affordance slots (chips, buttons)
-->
<script lang="ts">
  import Pulse from './Pulse.svelte';

  interface Props {
    primary: string;
    voice?: 'mono' | 'serif' | 'sans';
    secondary?: string;
    pulse?: boolean;
    children?: import('svelte').Snippet;
  }

  let { primary, voice = 'mono', secondary, pulse = false, children }: Props = $props();
</script>

<div class="empty empty--{voice}" role="status">
  {#if pulse}
    <div class="empty__ambient" aria-hidden="true">
      <Pulse state="idle" size="lg" label="" />
    </div>
  {/if}

  <div class="empty__pattern" aria-hidden="true">
    <!-- Subtle dot grid — gives the empty state a sense of "ground"
         without being a literal illustration. The grid is barely
         visible — about 4% opacity — and just enough to give the
         screen weight. -->
    <svg class="empty__pattern-svg" width="100%" height="100%" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <pattern id="empty-dotgrid" x="0" y="0" width="24" height="24" patternUnits="userSpaceOnUse">
          <circle cx="1" cy="1" r="0.8" fill="currentColor" />
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#empty-dotgrid)" />
    </svg>
  </div>

  <div class="empty__content">
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
</div>

<style>
  .empty {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-9) var(--space-6);
    text-align: center;
    color: var(--content-tertiary);
    overflow: hidden;
    min-height: 320px;
  }

  /* The ambient Pulse — sits in the background, very faint, slow. */
  .empty__ambient {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    opacity: 0.12;
    pointer-events: none;
  }

  /* The dot-grid pattern — barely visible, gives the screen weight. */
  .empty__pattern {
    position: absolute;
    inset: 0;
    color: var(--content-muted);
    opacity: 0.5;
    pointer-events: none;
    /* The pattern fades to transparent at the edges so it doesn't
       visually compete with the primary text. */
    mask-image: radial-gradient(ellipse 70% 60% at center, black 30%, transparent 80%);
    -webkit-mask-image: radial-gradient(ellipse 70% 60% at center, black 30%, transparent 80%);
  }

  .empty__pattern-svg {
    width: 100%;
    height: 100%;
  }

  /* The content — sits above the pattern, focuses attention. */
  .empty__content {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-3);
    max-width: 32em;
  }

  /* ── Voice variants ───────────────────────────────────────── */

  /* Mono: status-line aesthetic. Used for "Awaiting task." */
  .empty__primary--mono {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    font-weight: 500;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--content-tertiary);
  }

  /* Serif: a question. Used for "What would you like me to do?" */
  .empty__primary--serif {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    line-height: 1.4;
    color: var(--content-secondary);
    font-style: italic;
    font-weight: 400;
  }

  /* Sans: a quiet notice. Used for "Nothing here yet." */
  .empty__primary--sans {
    font-family: var(--font-sans);
    font-size: var(--text-body-lg-size);
    color: var(--content-secondary);
    font-weight: 400;
  }

  /* Secondary — sans, small, tertiary color, comfortable line-height */
  .empty__secondary {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    line-height: 1.5;
    max-width: 32em;
    margin: 0;
  }

  /* Affordance slot */
  .empty__affordance {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
    justify-content: center;
    margin-top: var(--space-4);
  }

  /* Dark mode: the pattern is even fainter (background is darker) */
  [data-mode="dark"] .empty__pattern {
    opacity: 0.3;
  }

  [data-mode="dark"] .empty__ambient {
    opacity: 0.18;
  }

  /* Reduced motion: no ambient pulse motion, no pattern opacity drift */
  @media (prefers-reduced-motion: reduce) {
    .empty__ambient {
      /* The pulse component itself respects reduced-motion, so this is fine */
    }
  }
</style>