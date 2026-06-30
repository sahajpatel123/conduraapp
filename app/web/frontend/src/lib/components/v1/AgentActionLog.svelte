<!--
  AgentActionLog — the dense replay table.

  Per spec §8.4: dense, scrolling, time-ordered stream of what the agent
  is doing. Each row: timestamp (mono), action type (chip), target
  (proportional), decision (muted italic), verification result.

  This is the "air traffic control panel" view. The settings pane has
  a more curated version (action replay); this is the live, dense feed.

  Props:
    actions — array of {id, time, type, target, decision, verified, model}
-->
<script lang="ts">
  import Pill from './Pill.svelte';
  import Dot from './Dot.svelte';
  import Chip from './Chip.svelte';

  interface Action {
    id: string;
    time: string;
    type: 'read' | 'write' | 'network' | 'destructive';
    target: string;
    decision: string;
    verified: boolean;
    model: string;
  }

  interface Props {
    actions?: Action[];
    /** When set, rows become clickable and highlight the selected id. */
    onrowclick?: (id: string) => void;
    selectedId?: string;
  }

  let { actions = [], onrowclick, selectedId }: Props = $props();
</script>

<div class="log" role="table" aria-label="Agent action log">
  <div class="log__head" role="row">
    <div role="columnheader" class="log__col-time">TIME</div>
    <div role="columnheader" class="log__col-type">ACTION</div>
    <div role="columnheader" class="log__col-target">TARGET</div>
    <div role="columnheader" class="log__col-decision">DECISION</div>
    <div role="columnheader" class="log__col-verified">✓</div>
  </div>

  <div class="log__body">
    {#each actions as action (action.id)}
      {#if onrowclick}
        <button
          type="button"
          class="log__row log__row--interactive"
          class:log__row--selected={selectedId === action.id}
          role="row"
          data-type={action.type}
          onclick={() => onrowclick(action.id)}
        >
          <span role="cell" class="log__col-time">{action.time}</span>
          <span role="cell" class="log__col-type">
            <Chip>{action.type}</Chip>
          </span>
          <span role="cell" class="log__col-target">{action.target}</span>
          <span role="cell" class="log__col-decision">{action.decision}</span>
          <span role="cell" class="log__col-verified">
            <Dot variant={action.verified ? 'success' : 'warning'} size="xs" />
          </span>
        </button>
      {:else}
        <div class="log__row" role="row" data-type={action.type}>
          <div role="cell" class="log__col-time">{action.time}</div>
          <div role="cell" class="log__col-type">
            <Chip>{action.type}</Chip>
          </div>
          <div role="cell" class="log__col-target">{action.target}</div>
          <div role="cell" class="log__col-decision">{action.decision}</div>
          <div role="cell" class="log__col-verified">
            <Dot variant={action.verified ? 'success' : 'warning'} size="xs" />
          </div>
        </div>
      {/if}
    {/each}
  </div>
</div>

<style>
  .log {
    width: 100%;
    background-color: var(--surface-base);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    overflow: hidden;
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
  }

  /* Header row */
  .log__head {
    display: grid;
    grid-template-columns: 80px 100px 1fr 1.5fr 32px;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-4);
    background-color: var(--surface-sunken);
    border-bottom: 1px solid var(--border-default);
  }

  .log__head > [role="columnheader"] {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.06em;
    color: var(--content-tertiary);
  }

  /* Body rows */
  .log__row {
    display: grid;
    grid-template-columns: 80px 100px 1fr 1.5fr 32px;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--border-subtle);
    align-items: center;
    transition: background-color var(--duration-fast) var(--ease-standard);
  }

  .log__row:last-child {
    border-bottom: none;
  }

  .log__row:hover {
    background-color: var(--paper-warm-50);
  }

  .log__row--interactive {
    appearance: none;
    width: 100%;
    text-align: left;
    font: inherit;
    color: inherit;
    cursor: pointer;
    border: none;
    border-bottom: 1px solid var(--border-subtle);
  }

  .log__row--interactive.log__row--selected {
    background-color: var(--plum-50);
    box-shadow: inset 2px 0 0 var(--content-accent);
  }

  /* Type-based left border (per spec: blast-radius color cue, not background fill) */
  .log__row[data-type="read"] {
    border-left: 2px solid var(--ink-cool-100);
  }
  .log__row[data-type="write"] {
    border-left: 2px solid var(--info-500);
  }
  .log__row[data-type="network"] {
    border-left: 2px solid var(--warning-500);
  }
  .log__row[data-type="destructive"] {
    border-left: 2px solid var(--error-500);
  }

  .log__col-time {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
  }

  .log__col-target {
    font-family: var(--font-sans);
    color: var(--content-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .log__col-decision {
    font-family: var(--font-serif);
    font-style: italic;
    color: var(--content-tertiary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .log__col-verified {
    display: flex;
    justify-content: center;
  }

  @media (max-width: 720px) {
    .log__head,
    .log__row {
      grid-template-columns: 64px 80px 1fr 32px;
    }
    .log__col-decision {
      display: none;
    }
  }
</style>