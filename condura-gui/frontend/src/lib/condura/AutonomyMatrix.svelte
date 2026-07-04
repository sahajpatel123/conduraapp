<script lang="ts">
  /**
   * AutonomyMatrix.svelte — the hero of Settings.
   *
   * 11 canonical task types (CLAUDE.md §27) × 3 states (block · warn · autonomous).
   * Each cell is a clickable dot; clicking the lit dot cycles forward.
   *
   * Keyboard:
   *   ← / →        move between Block · Warn · Auto within a row
   *   ↑ / ↓        move between rows
   *   Enter / Space on a focused dot cycles that task forward
   *   Tab leaves the matrix (roving tabindex)
   *
   * The live preview line below the grid reads what the current *coding*
   * choice will mean in plain English. Cross-fades 200ms when the dot
   * for coding is clicked (achieved via the inline `style:color`
   * transition on `.preview-verb`).
   *
   * Self-contained: this component owns the popping animation, the
   * roving-tabindex state, and the preview-line copy. The parent only
   * supplies `perTask`, `defaultLevel`, and an `onChange` callback.
   */
  import { onMount } from 'svelte';

  // ── canonical task order (CLAUDE.md §27) ──
  const TASK_ORDER = [
    'coding',
    'file_operations',
    'web_browsing',
    'email',
    'calendar',
    'messaging',
    'shell_commands',
    'computer_use',
    'research',
    'image_generation',
    'code_review',
  ] as const;

  const TASK_LABELS: Record<string, string> = {
    coding: 'Coding',
    file_operations: 'File operations',
    web_browsing: 'Web browsing',
    email: 'Email',
    calendar: 'Calendar',
    messaging: 'Messaging',
    shell_commands: 'Shell commands',
    computer_use: 'Computer use',
    research: 'Research',
    image_generation: 'Image generation',
    code_review: 'Code review',
  };

  // ── three states (block · warn · autonomous) ──
  // Block sits on the synapse stroke (the spine of the product),
  // warn is the warn token (amber), autonomous is the ok token (green).
  const LEVELS = ['block', 'warn', 'autonomous'] as const;
  type Level = (typeof LEVELS)[number];

  const LEVEL_COLOR: Record<Level, string> = {
    block: 'var(--synapse)',
    warn: 'var(--warn)',
    autonomous: 'var(--ok)',
  };

  // Full verb phrase for the per-row "Now" column.
  const LEVEL_VERB: Record<Level, string> = {
    block: 'block before acting',
    warn: 'warn before acting',
    autonomous: 'act on its own',
  };

  // Just the verb (two-color preview line: verb gets state color).
  const LEVEL_VERB_ONLY: Record<Level, string> = {
    block: 'block',
    warn: 'warn',
    autonomous: 'act',
  };

  // The tail after the verb.
  const LEVEL_VERB_TAIL: Record<Level, string> = {
    block: 'before acting',
    warn: 'before acting',
    autonomous: 'on its own',
  };

  type Props = {
    perTask: Record<string, string>;
    defaultLevel: string;
    onChange?: (perTask: Record<string, string>) => void;
  };

  let { perTask = {}, defaultLevel = 'warn', onChange }: Props = $props();

  const reducedMotion =
    typeof window !== 'undefined' && matchMedia('(prefers-reduced-motion: reduce)').matches;

  // Roving tabindex — only one dot at a time is in the tab order.
  // Default to the currently-active dot of the coding row (col = level).
  let focusedRow = $state(0);
  let focusedCol = $state(0);

  // Button refs for keyboard focus management.
  let buttonRefs: HTMLButtonElement[] = $state([]);

  // Pop animation: "<task>:<level>" for the dot just clicked. Cleared on
  // a 200ms timeout so re-clicking the same dot re-triggers the animation.
  let popping = $state<string>('');

  function levelOf(task: string): Level {
    const v = perTask[task] ?? defaultLevel ?? 'warn';
    return (LEVELS as readonly string[]).includes(v) ? (v as Level) : 'warn';
  }

  function nextLevel(l: Level): Level {
    const i = LEVELS.indexOf(l);
    return LEVELS[(i + 1) % LEVELS.length] as Level;
  }

  function setTask(task: string, to: Level): void {
    const cur = levelOf(task);
    const next: Level = cur === to ? nextLevel(cur) : to;
    onChange?.({ ...perTask, [task]: next });

    const key = `${task}:${to}`;
    popping = '';
    queueMicrotask(() => {
      popping = key;
    });
    if (typeof window !== 'undefined') {
      window.setTimeout(() => {
        if (popping === key) popping = '';
      }, reducedMotion ? 0 : 200);
    }
  }

  function focusDot(row: number, col: number): void {
    focusedRow = row;
    focusedCol = col;
    const idx = row * LEVELS.length + col;
    queueMicrotask(() => buttonRefs[idx]?.focus());
  }

  function onDotKey(ev: KeyboardEvent, row: number, col: number, task: string): void {
    const numRows = TASK_ORDER.length;
    const numCols = LEVELS.length;
    switch (ev.key) {
      case 'ArrowRight': {
        ev.preventDefault();
        focusDot(row, (col + 1) % numCols);
        return;
      }
      case 'ArrowLeft': {
        ev.preventDefault();
        focusDot(row, (col - 1 + numCols) % numCols);
        return;
      }
      case 'ArrowDown': {
        ev.preventDefault();
        focusDot((row + 1) % numRows, col);
        return;
      }
      case 'ArrowUp': {
        ev.preventDefault();
        focusDot((row - 1 + numRows) % numRows, col);
        return;
      }
      case 'Home': {
        ev.preventDefault();
        focusDot(row, 0);
        return;
      }
      case 'End': {
        ev.preventDefault();
        focusDot(row, numCols - 1);
        return;
      }
      case 'Enter':
      case ' ': {
        ev.preventDefault();
        setTask(task, LEVELS[col]);
        return;
      }
      default:
        return;
    }
  }

  // The preview line anchors on the *coding* row's current level —
  // a sentence is more readable than a configuration.
  let codingLevel = $derived(levelOf('coding'));

  // Position the focused dot on the active state of the coding row on mount,
  // so Tab into the matrix lands somewhere meaningful.
  onMount(() => {
    const cur = levelOf('coding');
    focusedRow = 0;
    focusedCol = LEVELS.indexOf(cur);
  });
</script>

<div class="matrix-wrap">
  <div class="matrix" role="table" aria-label="Autonomy matrix">
    <div class="matrix-head" role="row">
      <span class="matrix-task-col" role="columnheader">Task</span>
      <span class="matrix-state" role="columnheader">Block</span>
      <span class="matrix-state" role="columnheader">Warn</span>
      <span class="matrix-state" role="columnheader">Auto</span>
      <span class="matrix-now" role="columnheader">Now</span>
    </div>
    {#each TASK_ORDER as task, rowIdx (task)}
      <div class="matrix-row" role="row">
        <span class="matrix-task-col" role="rowheader">{TASK_LABELS[task] ?? task}</span>
        {#each LEVELS as st, colIdx (st)}
          {@const dotKey = `${task}:${st}`}
          {@const isActive = levelOf(task) === st}
          <span class="matrix-state" role="cell">
            <button
              bind:this={buttonRefs[rowIdx * LEVELS.length + colIdx]}
              class="auto-dot"
              class:active={isActive}
              class:popping={popping === dotKey}
              style:--dot-color={LEVEL_COLOR[st]}
              onclick={() => setTask(task, st)}
              onkeydown={(e) => onDotKey(e, rowIdx, colIdx, task)}
              aria-label={`${TASK_LABELS[task] ?? task}: ${st}`}
              aria-pressed={isActive}
              tabindex={focusedRow === rowIdx && focusedCol === colIdx ? 0 : -1}
            >
              <span class="dot-fill"></span>
            </button>
          </span>
        {/each}
        <span class="matrix-now" role="cell">{LEVEL_VERB[levelOf(task)]}</span>
      </div>
    {/each}
  </div>

  <div class="preview" aria-live="polite">
    <span class="preview-label">Preview</span>
    <span class="preview-line">
      Right now, for <em class="preview-task">coding</em>, Condura will
      <span
        class="preview-verb"
        style:color={LEVEL_COLOR[codingLevel]}
      >{LEVEL_VERB_ONLY[codingLevel]}</span>
      <span class="preview-tail">{LEVEL_VERB_TAIL[codingLevel]}.</span>
    </span>
  </div>
</div>

<style>
  .matrix-wrap {
    display: block;
  }

  .matrix {
    display: grid;
    grid-template-columns: minmax(120px, 1fr) 28px 28px 28px minmax(110px, 1.2fr);
    gap: 0 var(--space-4);
    align-items: center;
  }
  /* display:contents lets each row's children lay directly on the grid; the
     row separator hairline is drawn by a pseudo on each row's task cell
     so it spans the full grid row. */
  .matrix-head,
  .matrix-row {
    display: contents;
  }
  .matrix-head > span {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--content-faint);
    padding: var(--space-2) 0;
  }
  .matrix-row > span {
    padding: var(--space-3) 0;
  }
  .matrix-task-col {
    position: relative;
    font-size: 14px;
    color: var(--content);
  }
  .matrix-row .matrix-task-col::before {
    content: '';
    position: absolute;
    left: 0;
    right: calc(-1 * (28px + var(--space-4)) * 3 - var(--space-4) * 2);
    top: 0;
    height: 1px;
    background: var(--hair);
  }
  .matrix-row:first-of-type .matrix-task-col::before {
    display: none;
  }
  .matrix-state {
    display: grid;
    place-items: center;
  }
  .matrix-now {
    font-size: 13px;
    color: var(--content-mute);
    font-style: italic;
  }

  /* The dot — 18×18 hollow ring with a hairline in the state color when
     inactive, a pollen halo + filled inner when active, and a 180ms
     scale-pop on click. --dot-color is set inline per-button. */
  .auto-dot {
    width: 18px;
    height: 18px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid var(--dot-color, var(--hair-strong));
    background: transparent;
    display: grid;
    place-items: center;
    cursor: pointer;
    transition:
      border-color var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .auto-dot:hover {
    transform: scale(1.08);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--dot-color, var(--content-mute)) 18%, transparent);
  }
  .auto-dot:active {
    transform: scale(0.96);
  }
  .auto-dot:focus-visible {
    outline: none;
    box-shadow: 0 0 0 3px var(--pollen-halo-color);
  }
  .auto-dot.active {
    border-color: transparent;
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--dot-color, var(--pollen)) 30%, transparent);
  }
  /* :global(...) so the runtime-toggled class applies across Svelte's
     scoping boundary when the key is built inline. */
  :global(.auto-dot.popping) {
    animation: dot-pop 180ms var(--ease);
  }
  .dot-fill {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--dot-color, var(--content));
    transform: scale(0);
    transition: transform var(--dur) var(--ease);
    pointer-events: none;
  }
  .auto-dot.active .dot-fill {
    transform: scale(1);
  }

  @keyframes dot-pop {
    0%   { transform: scale(1); }
    55%  { transform: scale(1.18); }
    100% { transform: scale(1); }
  }

  /* Preview line — Instrument Serif italic with two-color treatment. */
  .preview {
    display: flex;
    align-items: baseline;
    gap: var(--space-4);
    margin-top: var(--space-7);
    padding: var(--space-5);
    background: var(--surface-card);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
  }
  .preview-label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--synapse);
    flex: none;
  }
  .preview-line {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 18px;
    color: var(--ink);
    line-height: 1.4;
    letter-spacing: -0.005em;
  }
  .preview-task {
    color: var(--synapse);
    font-style: italic;
  }
  .preview-verb {
    font-style: italic;
    font-weight: 500;
    transition: color var(--dur) var(--ease);
  }
  .preview-tail {
    color: var(--content-mute);
    font-style: italic;
  }

  @media (prefers-reduced-motion: reduce) {
    :global(.auto-dot.popping) {
      animation: none;
    }
  }

  @media (max-width: 640px) {
    .matrix {
      grid-template-columns: minmax(110px, 1fr) 24px 24px 24px minmax(90px, 1fr);
      gap: 0 var(--space-3);
    }
    .matrix-row .matrix-task-col::before {
      right: calc(-1 * (24px + var(--space-3)) * 3 - var(--space-3) * 2);
    }
  }
</style>