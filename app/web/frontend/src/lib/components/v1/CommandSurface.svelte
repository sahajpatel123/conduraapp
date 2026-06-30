<!--
  CommandSurface — the heart of Synaptic.

  Per spec §9: this is the agent's ONLY persistent UI during normal use.
  Three stacked layers:
    1. Contextual strip (44px): what the agent noticed on screen
    2. Omni-bar input (64px): single serif text field
    3. Hint row (16px): keyboard shortcuts

  Four states:
    - idle: empty, ready, breathing pulse
    - active: user typing, ranked interpretations appear below
    - processing: input collapsed, progress bar + about-to preview
    - result: receipt + Undo/Pin/Send-to-chat actions

  Glass background — the ONLY place glass is used in the entire app.

  Note: `mode` (not `state`) is used for the surface's state to avoid
  shadowing the Svelte 5 `$state` rune, which the compiler cannot
  distinguish from a regular local variable when prefixed with `$`.
-->
<script lang="ts">
  import Pulse from './Pulse.svelte';
  import Input from './Input.svelte';
  import Suggestion from './Suggestion.svelte';
  import ContextChip from './ContextChip.svelte';
  import ProgressBar from './ProgressBar.svelte';
  import Receipt from './Receipt.svelte';
  import Button from './Button.svelte';

  type Mode = 'idle' | 'active' | 'processing' | 'result';

  interface ContextChipData { label: string; active?: boolean; }
  interface Interpretation { interpretation: string; steps?: string; highlighted?: boolean; }
  interface Progress { elapsedMs: number; state?: 'thinking' | 'tool-call' | 'verifying' | 'executing'; modelName?: string; }
  interface Result { verb: string; target: string; timestamp: string; state?: 'done' | 'paused' | 'error' | 'pending'; }

  interface Props {
    mode?: Mode;
    contextChips?: ContextChipData[];
    interpretations?: Interpretation[];
    progress?: Progress;
    result?: Result;
    onsubmit?: (text: string) => void;
    onselect?: (interpretation: Interpretation, index: number) => void;
    oncontext?: (label: string) => void;
    onpause?: () => void;
    onundo?: () => void;
    onpin?: () => void;
    onsendtochat?: () => void;
  }

  let {
    mode = 'idle',
    contextChips = [],
    interpretations = [],
    progress,
    result,
    onsubmit,
    onselect,
    oncontext,
    onpause,
    onundo,
    onpin,
    onsendtochat,
  }: Props = $props();

  let inputValue = $state('');
  let highlightIndex = $state(0);

  // Reset highlight when interpretations change.
  $effect(() => {
    interpretations;
    highlightIndex = 0;
  });

  // Per spec §9.2 — serif grows slightly after 8+ characters typed.
  let inputVariant: 'sans' | 'serif' = $derived(
    mode === 'processing' || mode === 'result' ? 'sans' : 'serif'
  );
  let inputSize: 'sm' | 'md' | 'lg' = $derived(
    mode === 'processing' || mode === 'result' ? 'md' : 'lg'
  );

  function handleKey(e: KeyboardEvent) {
    if (mode === 'active' && interpretations.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        highlightIndex = (highlightIndex + 1) % interpretations.length;
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        highlightIndex = (highlightIndex - 1 + interpretations.length) % interpretations.length;
      } else if (e.key === 'Enter' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault();
        const picked = interpretations[highlightIndex];
        if (picked) onselect?.(picked, highlightIndex);
      }
    }
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      if (inputValue.trim()) {
        onsubmit?.(inputValue.trim());
        inputValue = '';
      }
    }
    if (e.key === 'Escape') {
      e.preventDefault();
      inputValue = '';
    }
  }

  function handleSubmit(e: Event) {
    e.preventDefault();
    if (inputValue.trim()) {
      onsubmit?.(inputValue.trim());
      inputValue = '';
    }
  }

  // Pulse mirrors agent state.
  let pulseState: 'idle' | 'thinking' = $derived(
    mode === 'processing' ? 'thinking' : 'idle'
  );
</script>

<div
  class="surface surface--{mode}"
  role="dialog"
  aria-label="Synaptic command surface"
  aria-modal="false"
>
  <!-- ── Layer 1: Contextual strip ─────────────────────────────── -->
  <div class="surface__context">
    {#if contextChips.length > 0}
      <div class="surface__chips">
        {#each contextChips as chip}
          <ContextChip
            label={chip.label}
            active={chip.active}
            onclick={() => oncontext?.(chip.label)}
          />
        {/each}
      </div>
    {:else}
      <div class="surface__context-empty">
        <Pulse state="idle" size="sm" label="Agent listening" />
        <span class="surface__context-text">Nothing on screen. Listening.</span>
      </div>
    {/if}
  </div>

  <!-- ── Layer 2: Omni-bar input ───────────────────────────────── -->
  <form class="surface__input-row" onsubmit={handleSubmit}>
    <Input
      bind:value={inputValue}
      variant={inputVariant}
      size={inputSize}
      placeholder={mode === 'processing' ? 'Agent working…' : 'What would you like me to do?'}
      ariaLabel="Command input"
      onkeydown={handleKey}
      disabled={mode === 'processing'}
      monospace={mode === 'processing' || mode === 'result'}
    />
  </form>

  <!-- ── State-dependent body ──────────────────────────────────── -->
  {#if mode === 'active' && interpretations.length > 0}
    <div class="surface__interpretations">
      {#each interpretations as interp, i}
        <Suggestion
          interpretation={interp.interpretation}
          steps={interp.steps}
          highlighted={i === highlightIndex}
          onclick={() => onselect?.(interp, i)}
        />
      {/each}
    </div>
  {/if}

  {#if mode === 'processing' && progress}
    <div class="surface__processing">
      <ProgressBar
        elapsedMs={progress.elapsedMs}
        state={progress.state ?? 'thinking'}
        modelName={progress.modelName}
      />
      <div class="surface__processing-actions">
        <Button variant="tertiary" size="sm" onclick={onpause}>⏸ Pause</Button>
      </div>
    </div>
  {/if}

  {#if mode === 'result' && result}
    <div class="surface__result">
      <Receipt
        timestamp={result.timestamp}
        verb={result.verb}
        target={result.target}
        state={result.state ?? 'done'}
      />
      <div class="surface__result-actions">
        <Button variant="secondary" size="sm" onclick={onundo}>↻ Undo</Button>
        <Button variant="secondary" size="sm" onclick={onpin}>📌 Pin</Button>
        <Button variant="tertiary" size="sm" onclick={onsendtochat}>⌘↩ Chat</Button>
      </div>
    </div>
  {/if}

  <!-- ── Layer 3: Hint row ─────────────────────────────────────── -->
  <div class="surface__hints" aria-hidden="true">
    {#if mode === 'processing'}
      <span>esc to interrupt</span>
    {:else if mode === 'result'}
      <span>esc to dismiss</span>
    {:else}
      <span>⌘↵ to send · esc to dismiss · ⌘K for everything</span>
    {/if}
  </div>
</div>

<style>
  /*
   * The Command Surface is the ONLY place in Synaptic where glass is used.
   * Per spec §2: hierarchy via hairline + tone, never shadow. The glass
   * backdrop here is the one place elevation is earned.
   */
  .surface {
    width: 560px;
    max-width: calc(100vw - 32px);
    background-color: var(--surface-glass);
    -webkit-backdrop-filter: blur(var(--blur-2xl)) saturate(180%);
    backdrop-filter: blur(var(--blur-2xl)) saturate(180%);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-3);
    padding: var(--space-3);
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    color: var(--content-primary);
    font-family: var(--font-sans);
    animation: surface-in var(--duration-base) var(--ease-decelerate) both;
    transform-origin: bottom right;
  }

  .surface--idle { min-height: 140px; }
  .surface--active { min-height: 280px; }
  .surface--processing { min-height: 180px; }
  .surface--result { min-height: 240px; }

  /* ── Layer 1: Contextual strip ───────────────────────────── */
  .surface__context {
    min-height: 44px;
    display: flex;
    align-items: center;
    padding: 0 var(--space-2);
  }
  .surface__chips {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .surface__context-empty {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }
  .surface__context-text {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    color: var(--content-tertiary);
    text-transform: uppercase;
  }

  /* ── Layer 2: Input row ──────────────────────────────────── */
  .surface__input-row {
    margin: 0;
  }

  /* Override the input's bottom border inside the surface to feel
     integrated rather than fielded. */
  .surface :global(.input) {
    border-bottom: 1px solid var(--border-subtle);
    padding-left: var(--space-3);
    padding-right: var(--space-3);
  }
  .surface :global(.input:focus-visible) {
    border-bottom-color: var(--content-accent);
    box-shadow: 0 1px 0 0 var(--content-accent);
  }

  /* ── Interpretations list ────────────────────────────────── */
  .surface__interpretations {
    display: flex;
    flex-direction: column;
    gap: 2px;
    margin-top: var(--space-2);
    max-height: 240px;
    overflow-y: auto;
  }

  /* ── Processing body ─────────────────────────────────────── */
  .surface__processing {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-3);
    margin-top: var(--space-2);
    border-top: 1px solid var(--border-subtle);
  }
  .surface__processing-actions {
    display: flex;
    justify-content: flex-end;
  }

  /* ── Result body ─────────────────────────────────────────── */
  .surface__result {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-3);
    margin-top: var(--space-2);
    border-top: 1px solid var(--border-subtle);
  }
  .surface__result-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
  }

  /* ── Layer 3: Hint row ───────────────────────────────────── */
  .surface__hints {
    padding: var(--space-2) var(--space-3) 0;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.04em;
    color: var(--content-muted);
    border-top: 1px solid var(--border-subtle);
    margin-top: var(--space-2);
  }

  /* ── Animations ──────────────────────────────────────────── */
  @keyframes surface-in {
    from {
      opacity: 0;
      transform: scale(0.96) translateY(8px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .surface {
      animation: none;
    }
  }
</style>