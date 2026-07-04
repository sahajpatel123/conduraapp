<!--
  ChatSurface — the editorial column.

  This is no longer a chat interface. This is a page from a book that
  happens to have a thoughtful agent reading and writing in it. The
  user's voice is a sans note in the margin. The agent's voice is the
  serif body text. Timestamps are mono footnotes.

  Three voices:
    - User (sans, paper-warm tinted sticky-note background)
    - Agent (serif, generous line-height 1.7, no background)
    - System (italic serif, muted, for notices)

  Per spec §11.1 + the website aesthetic. The chat is a column in a book,
  not a messaging app.
-->
<script lang="ts">
  import SynapseField from './SynapseField.svelte';
  import Inline from './Inline.svelte';
  import Button from './Button.svelte';
  import Icon from './icons/Icon.svelte';

  type Role = 'user' | 'agent' | 'system';

  interface Turn {
    id: string;
    role: Role;
    content: string;
    timestamp?: string;
    thinking?: string;
  }

  interface Props {
    turns?: Turn[];
    showEmptyHero?: boolean;
  }

  let { turns = [], showEmptyHero = false }: Props = $props();
</script>

{#if turns.length === 0 && showEmptyHero}
  <!-- The hero empty state — the SynapseField lives here -->
  <div class="hero">
    <SynapseField />
    <div class="hero__hints">
      <div class="hero__hint-chips">
        <Inline gap="3">
          <Button size="md" variant="secondary" icon="mail" onclick={() => {}}>
            Summarize my last 3 emails
          </Button>
          <Button size="md" variant="secondary" icon="globe" onclick={() => {}}>
            Open Safari and search
          </Button>
          <Button size="md" variant="secondary" icon="edit" onclick={() => {}}>
            Rename a file
          </Button>
        </Inline>
      </div>
      <p class="hero__kbd-hint">
        or press <kbd>⌘K</kbd> for anything
      </p>
    </div>
  </div>
{:else if turns.length === 0}
  <!-- Smaller empty state when hero is disabled -->
  <div class="empty">
    <p class="empty__line">Nothing yet.</p>
    <p class="empty__hint">Press <kbd>⌘K</kbd> to start.</p>
  </div>
{:else}
  <article class="column" role="log" aria-label="Conversation">
    {#each turns as turn (turn.id)}
      <section
        class="turn turn--{turn.role}"
        data-role={turn.role}
        aria-label={turn.role === 'system' ? undefined : `${turn.role} message`}
      >
        <header class="turn__head">
          <span class="turn__role" aria-hidden="true">
            {turn.role === 'agent' ? 'Synaptic' : turn.role === 'user' ? 'You' : 'System'}
          </span>
          {#if turn.timestamp}
            <time class="turn__time">{turn.timestamp}</time>
          {/if}
        </header>

        {#if turn.thinking && turn.role === 'agent'}
          <details class="turn__thinking">
            <summary>
              <Icon name="chevron-right" size="xs" />
              Thinking
            </summary>
            <p class="turn__thinking-body">{turn.thinking}</p>
          </details>
        {/if}

        <div class="turn__body">{turn.content}</div>
      </section>

      <hr class="turn__rule" aria-hidden="true" />
    {/each}
  </article>
{/if}

<style>
  /* ── The empty state — nothing yet ─────────────────────── */

  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 320px;
    padding: var(--space-7);
    gap: var(--space-3);
    background-color: var(--surface-base);
  }

  .empty__line {
    font-family: var(--font-serif);
    font-style: italic;
    font-size: var(--text-h3-size);
    color: var(--content-secondary);
    margin: 0;
  }

  .empty__hint {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    margin: 0;
    letter-spacing: 0.04em;
  }

  .empty__hint kbd {
    font-family: var(--font-mono);
    background-color: var(--paper-warm-50);
    border: 1px solid var(--border-default);
    padding: 1px 6px;
    border-radius: var(--radius-xs);
    color: var(--content-primary);
    margin: 0 var(--space-1);
  }

  /* ── The hero empty state with SynapseField ────────────── */

  .hero {
    position: relative;
    width: 100%;
    min-height: 540px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-9) var(--space-6) var(--space-7);
    background-color: var(--surface-base);
    overflow: hidden;
  }

  .hero__hints {
    position: relative;
    z-index: 3;
    margin-top: var(--space-7);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-4);
  }

  .hero__hint-chips {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--space-3);
  }

  .hero__kbd-hint {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    margin: 0;
    letter-spacing: 0.04em;
  }

  .hero__kbd-hint kbd {
    font-family: var(--font-mono);
    background-color: var(--paper-warm-50);
    border: 1px solid var(--border-default);
    padding: 1px 6px;
    border-radius: var(--radius-xs);
    color: var(--content-primary);
    margin: 0 var(--space-1);
  }

  /* ── The editorial column ──────────────────────────────── */

  .column {
    /* The narrow column — like a magazine page */
    max-width: 68ch;
    margin: 0 auto;
    padding: var(--space-9) var(--space-6) var(--space-13);
    background-color: var(--surface-base);
  }

  /* ── A turn — the unit of conversation ─────────────────── */

  .turn {
    padding: var(--space-5) 0;
  }

  /* The head — role + timestamp in a small mono caps line */
  .turn__head {
    display: flex;
    align-items: baseline;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
  }

  .turn__role {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--content-tertiary);
  }

  .turn__time {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
  }

  /* The body — three voices, three typography profiles */

  .turn--agent .turn__body {
    font-family: var(--font-serif);
    font-size: 17px;       /* — not 15px, not 18px: 17 reads like a book */
    line-height: 1.7;
    color: var(--content-primary);
    /* The agent's serif voice uses real italics for emphasis */
  }

  .turn--agent .turn__body :global(em) {
    font-style: italic;
    color: var(--content-accent);
  }

  .turn--user .turn__body {
    font-family: var(--font-sans);
    font-size: var(--text-body-size);
    line-height: 1.6;
    color: var(--content-primary);
  }

  .turn--system .turn__body {
    font-family: var(--font-serif);
    font-style: italic;
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    text-align: center;
  }

  /* The user's turn gets a soft paper-warm background — like a sticky note
     on the page. Subtle, not a "card", just a tint that says "this is from
     outside the book". */
  .turn--user {
    background-color: var(--paper-warm-50);
    border-radius: var(--radius-md);
    padding-left: var(--space-5);
    padding-right: var(--space-5);
  }

  /* The thinking details — collapsible. Default closed. */
  .turn__thinking {
    margin-bottom: var(--space-3);
    padding: var(--space-2) var(--space-3);
    background-color: var(--surface-sunken);
    border-radius: var(--radius-sm);
    border: 1px dashed var(--border-subtle);
    font-family: var(--font-sans);
  }

  .turn__thinking summary {
    list-style: none;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    color: var(--content-tertiary);
    font-size: var(--text-caption-size);
  }

  .turn__thinking summary::-webkit-details-marker {
    display: none;
  }

  .turn__thinking-body {
    margin: var(--space-2) 0 0;
    font-style: italic;
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    line-height: 1.5;
  }

  /* The hairline rule between turns — a printed-page detail */
  .turn__rule {
    border: 0;
    height: 1px;
    background-color: var(--border-subtle);
    margin: 0;
    width: 100%;
  }

  /* Paper grain texture — barely perceptible, but it changes everything */
  .column::before {
    content: '';
    position: absolute;
    inset: 0;
    pointer-events: none;
    background-image:
      radial-gradient(circle at 20% 30%, rgba(20, 17, 11, 0.012) 0%, transparent 60%),
      radial-gradient(circle at 80% 70%, rgba(20, 17, 11, 0.012) 0%, transparent 60%);
    background-size: 100% 100%;
  }

  /* .column { position: relative } for ::before */
  .column {
    position: relative;
  }
</style>