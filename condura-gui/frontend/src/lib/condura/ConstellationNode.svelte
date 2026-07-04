<script lang="ts">
  import Glyph from './Glyph.svelte';

  /**
   * Condura · ConstellationNode — one of the 6 live nodes on the ring.
   * ──────────────────────────────────────────────────────────────────────
   * A clickable surface on the 340px dashed ring. Wired nodes flip to
   * `--synapse-light` fill + `--synapse` stroke. Skipped nodes render at
   * `--paper` fill + `--content-faint` stroke. Hovered (non-active) nodes
   * lift -2px + a 5px pollen halo.
   *
   * The breath-pulse runs while the node is in an `awaiting` / `probing`
   * state, communicating "the system is talking to this node right now."
   *
   * The thread-draw across the bottom edge renders once when the node
   * becomes `done` — the signature gesture for "this is connected."
   *
   * Props are read-only; the parent owns the wiring & selection state.
   */

  export type NodeId = 'perceive' | 'power' | 'summon' | 'voice' | 'threads' | 'account';
  export type NodeState = 'probing' | 'done' | 'skipped' | 'error' | 'awaiting';

  let {
    id,
    label,
    glyph,
    state,
    indicator,
    selected,
    delayMs = 0,
    onclick,
    onhover,
  }: {
    id: NodeId;
    label: string;
    glyph: string;
    state: NodeState;
    indicator: string;
    selected: boolean;
    delayMs?: number;
    onclick: (id: NodeId) => void;
    onhover: (id: NodeId | null) => void;
  } = $props();

  let hovered = $state(false);

  // The dot color follows state — synapse for done, pollen for probing/awaiting,
  // faint for skipped, danger for error. The label color follows the same.
  let fillColor = $derived(
    state === 'done'
      ? 'var(--synapse-light)'
      : state === 'error'
        ? 'color-mix(in oklab, var(--danger) 30%, var(--paper))'
        : 'var(--paper)'
  );
  let strokeColor = $derived(
    state === 'done'
      ? 'var(--synapse)'
      : state === 'error'
        ? 'var(--danger)'
        : state === 'skipped'
          ? 'var(--content-faint)'
          : 'var(--hair-strong)'
  );

  // Breath-pulse runs while the node is being probed or is awaiting action.
  let pulseActive = $derived(state === 'probing' || state === 'awaiting');

  // The bottom-edge thread-draw signature.
  let threadDrawn = $derived(state === 'done');
</script>

<button
  type="button"
  class="node"
  class:done={state === 'done'}
  class:skipped={state === 'skipped'}
  class:error={state === 'error'}
  class:selected
  class:hovered
  class:pulse-active={pulseActive}
  style="animation-delay: {delayMs}ms"
  aria-label={`${label} — ${state}`}
  aria-pressed={selected}
  onclick={() => onclick(id)}
  onmouseenter={() => {
    hovered = true;
    onhover(id);
  }}
  onmouseleave={() => {
    hovered = false;
    onhover(null);
  }}
  onfocus={() => onhover(id)}
  onblur={() => onhover(null)}
>
  <!-- Thread-draw across the node's bottom edge — the signature gesture
       for "this is connected." Path uses pathLength=1 + stroke-dashoffset
       to draw left-to-right over --dur-slow. -->
  <svg class="thread" viewBox="0 0 24 4" preserveAspectRatio="none" aria-hidden="true">
    <path
      d="M 0 2 L 24 2"
      pathLength="1"
      vector-effect="non-scaling-stroke"
      stroke-dasharray="1"
      stroke-dashoffset={threadDrawn ? 0 : 1}
    />
  </svg>

  <span class="ring" aria-hidden="true">
    <span class="dot" style="background:{fillColor};border-color:{strokeColor}"></span>
  </span>

  <span class="glyph-cell" aria-hidden="true">
    <Glyph name={glyph} size={18} stroke={1.5} />
  </span>

  <span class="label">{label}</span>
  <span class="indicator" data-state={state}>{indicator}</span>
</button>

<style>
  .node {
    position: relative;
    display: grid;
    grid-template-rows: auto auto auto auto;
    justify-items: center;
    gap: 2px;
    padding: var(--space-2) var(--space-1) var(--space-1);
    width: 84px;
    background: none;
    border: none;
    color: inherit;
    cursor: pointer;
    border-radius: var(--r-md);
    /* Stagger entry from top, counterclockwise. */
    opacity: 0;
    transform: scale(0.7);
    animation: node-enter 320ms var(--ease-pop) forwards;
    transition:
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .node:hover,
  .node.hovered {
    transform: translateY(-2px);
    box-shadow: 0 0 0 5px var(--pollen-halo);
  }
  .node:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 2px var(--synapse),
      0 0 0 6px var(--pollen-halo);
  }
  .node.selected {
    transform: translateY(-2px);
    box-shadow: var(--shadow-card);
  }
  .node.done.selected {
    box-shadow: var(--shadow-card), 0 0 0 5px var(--pollen-halo);
  }

  /* Breath-pulse while the node is actively being probed / awaiting action.
     Communicates "the system is talking to this node right now." */
  .node.pulse-active .ring {
    animation: ring-pulse 1.6s var(--ease) infinite;
  }

  .thread {
    position: absolute;
    left: 8px;
    right: 8px;
    bottom: -1px;
    width: calc(100% - 16px);
    height: 4px;
    pointer-events: none;
  }
  .thread path {
    fill: none;
    stroke: var(--synapse);
    stroke-width: 1.5;
    stroke-linecap: round;
    transition: stroke-dashoffset var(--dur-slow) var(--ease);
  }

  .ring {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    border-radius: 50%;
  }
  .dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: 1.5px solid var(--hair-strong);
    transition:
      background var(--dur) var(--ease),
      border-color var(--dur) var(--ease);
  }
  .node.done .dot {
    width: 14px;
    height: 14px;
  }

  .glyph-cell {
    display: grid;
    place-items: center;
    margin-top: 2px;
    color: var(--content-mute);
    transition: color var(--dur) var(--ease);
  }
  .node:hover .glyph-cell,
  .node.hovered .glyph-cell,
  .node.selected .glyph-cell {
    color: var(--synapse);
  }
  .node.done .glyph-cell {
    color: var(--synapse);
  }

  .label {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 13px;
    color: var(--content);
    line-height: 1;
    margin-top: 4px;
  }
  .indicator {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
    line-height: 1.4;
    max-width: 80px;
    text-align: center;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .indicator[data-state='done'] {
    color: var(--synapse);
  }
  .indicator[data-state='awaiting'] {
    color: var(--warn);
  }
  .indicator[data-state='error'] {
    color: var(--danger);
  }
  .indicator[data-state='skipped'] {
    color: var(--content-faint);
  }

  @keyframes node-enter {
    from { opacity: 0; transform: scale(0.7); }
    to   { opacity: 1; transform: scale(1); }
  }
  @keyframes ring-pulse {
    0%, 100% { transform: scale(1); opacity: 0.85; }
    50%      { transform: scale(1.18); opacity: 1; }
  }

  @media (prefers-reduced-motion: reduce) {
    .node {
      animation: none;
      opacity: 1;
      transform: none;
    }
    .node.pulse-active .ring {
      animation: none;
    }
    .node:hover,
    .node.hovered,
    .node.selected {
      transform: none;
    }
  }
</style>