<script lang="ts">
  /**
   * Keyboard shortcut cheatsheet — `?` opens it.
   * Reads-modifier from navigator.platform so the same strings work on Mac/Win/Linux.
   * Grouped by surface (Global / Palette / Ask) so the cheatsheet scales with new surfaces.
   */
  import { onMount } from 'svelte'

  interface Props {
    open: boolean
    onclose: () => void
  }
  let { open, onclose }: Props = $props()

  let modLabel = $state('⌘')

  type Group = { kind: string; rows: { keys: string[]; label: string }[] }

  const GROUPS: Group[] = [
    {
      kind: 'Global',
      rows: [
        { keys: ['?'], label: 'Show this help' },
        { keys: ['⌘', 'K'], label: 'Open search (palette)' },
        { keys: ['⌘', ','], label: 'Open Settings' },
        { keys: ['⇧', 'T'], label: 'Switch light / dark' },
        { keys: ['⌘', '⇧', 'Esc'], label: 'Hard halt — stop everything' },
        { keys: ['⌘', '⇧', 'A'], label: 'Open Audit ledger' },
        { keys: ['⌘', '⇧', 'H'], label: 'Open Hub (skills shelf)' },
        { keys: ['⌘', '⇧', 'S'], label: 'Open Sync (device pairing)' },
        { keys: ['⌘', '⇧', 'C'], label: 'Open Channels' },
      ],
    },
    {
      kind: 'Palette',
      rows: [
        { keys: ['↑', '↓'], label: 'Move selection' },
        { keys: ['Enter'], label: 'Run item' },
        { keys: ['Esc'], label: 'Close palette' },
      ],
    },
    {
      kind: 'Ask',
      rows: [
        { keys: ['Enter'], label: 'Send' },
        { keys: ['⇧', 'Enter'], label: 'New line' },
        { keys: ['Esc'], label: 'Stop stream' },
        { keys: ['⌘', '.'], label: 'Stop stream (macOS convention)' },
        { keys: ['/'], label: 'Slash commands & skills' },
        { keys: ['⌥', 'C'], label: 'Copy last assistant response' },
        { keys: ['⌘', '⇧', 'R'], label: 'Regenerate last response' },
        { keys: ['⌘', '⇧', 'N'], label: 'Start a new ask' },
        { keys: ['⌘', '⇧', 'E'], label: 'Export thread as Markdown' },
      ],
    },
    {
      kind: 'Settings',
      rows: [
        { keys: ['←', '→'], label: 'Move between tabs' },
        { keys: ['Home'], label: 'First tab' },
        { keys: ['End'], label: 'Last tab' },
        { keys: ['⌘', '1'], label: 'Jump to General (or Ask outside Settings)' },
        { keys: ['⌘', '2'], label: 'Jump to Permissions (or Hub)' },
        { keys: ['⌘', '3'], label: 'Jump to Control (or Skills)' },
        { keys: ['⌘', '4'], label: 'Jump to Models (or Sync)' },
        { keys: ['⌘', '5'], label: 'Jump to Data (or Audit)' },
        { keys: ['⌘', '6'], label: 'Jump to Replay' },
        { keys: ['⌘', '7'], label: 'Jump to Channels' },
        { keys: ['⌘', '8'], label: 'Jump to Agents' },
      ],
    },
  ]

  onMount(() => {
    modLabel = /Mac|iPhone|iPad/.test(navigator.platform) ? '⌘' : 'Ctrl'
  })

  function onKey(e: KeyboardEvent): void {
    if (!open) return
    if (e.key === 'Escape') {
      e.preventDefault()
      onclose()
    }
  }

  function displayLabel(keys: string[]): string {
    return keys.map((k) => (k === 'Mod' ? modLabel : k)).join(' + ')
  }
</script>

<svelte:window onkeydown={onKey} />
{#if open}
  <div class="back" onclick={onclose} role="presentation"></div>
  <div class="panel" role="dialog" aria-modal="true" aria-label="Keyboard shortcuts">
    <header class="head">
      <p class="kicker">Keyboard · meridian</p>
      <h2>Shortcuts</h2>
      <p class="lead">
        Press
        <kbd class="inline">?</kbd>
        any time to bring this back. Esc closes.
      </p>
    </header>
    <div class="grid">
      {#each GROUPS as g (g.kind)}
        <section class="group">
          <p class="group-k">{g.kind}</p>
          <ul>
            {#each g.rows as r (r.label)}
              <li>
                <span class="label">{r.label}</span>
                <span class="keys" aria-label={displayLabel(r.keys)}>
                  {#each r.keys as k, i (k + i)}
                    {#if i > 0}<span class="plus" aria-hidden="true">+</span>{/if}
                    <kbd>{displayLabel([k])}</kbd>
                  {/each}
                </span>
              </li>
            {/each}
          </ul>
        </section>
      {/each}
    </div>
    <footer class="foot">
      <span>Tip: hold ⇧ and press the first letter of any surface to jump there from the palette.</span>
    </footer>
  </div>
{/if}

<style>
  .back {
    position: fixed;
    inset: 0;
    z-index: 80;
    background: var(--md-scrim);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    animation: md-fade 200ms var(--md-ease) both;
  }
  .panel {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: min(620px, calc(100vw - 32px));
    max-height: min(78vh, 640px);
    overflow: auto;
    background: var(--md-surface);
    border: 1px solid var(--md-line);
    border-radius: 12px;
    z-index: 81;
    box-shadow: none;
    animation: md-keys-pop 240ms var(--md-ease) both;
    padding-bottom: env(safe-area-inset-bottom, 0);
  }
  .head {
    padding: 18px 22px 14px;
    border-bottom: 1px solid var(--md-line);
  }
  .kicker {
    margin: 0 0 6px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  h2 {
    margin: 0 0 6px;
    font-family: var(--md-font-display);
    font-size: 22px;
    font-weight: 650;
    letter-spacing: -0.035em;
  }
  .lead {
    margin: 0;
    font-size: 12.5px;
    line-height: 1.45;
    color: var(--md-ink-mute);
    max-width: 48ch;
  }
  .inline {
    font-family: var(--md-font-mono);
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 5px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-stage) 70%, var(--md-surface));
    color: var(--md-ink-faint);
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 14px;
    padding: 16px 22px 18px;
  }
  .group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .group-k {
    margin: 0;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  ul {
    margin: 0;
    padding: 0;
    list-style: none;
    display: grid;
    gap: 8px;
  }
  li {
    display: grid;
    grid-template-columns: 1fr auto;
    align-items: center;
    gap: 12px;
    padding: 8px 10px;
    border-radius: 8px;
    background: color-mix(in oklab, var(--md-stage) 60%, transparent);
  }
  .label {
    font-size: 12.5px;
    color: var(--md-ink-soft);
    line-height: 1.35;
  }
  .keys {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    flex: none;
  }
  .plus {
    font-family: var(--md-font-mono);
    font-size: 10px;
    color: var(--md-ink-faint);
    opacity: 0.6;
  }
  kbd {
    font-family: var(--md-font-mono);
    font-size: 10.5px;
    line-height: 1;
    padding: 5px 7px;
    min-width: 18px;
    text-align: center;
    border-radius: 6px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    color: var(--md-ink);
    box-shadow: none;
  }
  .foot {
    padding: 12px 22px 16px;
    border-top: 1px solid var(--md-line);
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.04em;
    color: var(--md-ink-faint);
  }
  @keyframes md-keys-pop {
    from {
      opacity: 0;
      transform: translate(-50%, -48%) scale(0.97);
    }
    to {
      opacity: 1;
      transform: translate(-50%, -50%) scale(1);
    }
  }
  @media (max-width: 640px) {
    .panel {
      top: auto;
      bottom: 0;
      left: 0;
      right: 0;
      width: 100%;
      max-width: none;
      max-height: min(82vh, 700px);
      transform: none;
      border-radius: 14px 14px 0 0;
      border-left: 0;
      border-right: 0;
      border-bottom: 0;
      animation: md-keys-sheet 320ms var(--md-spring) both;
    }
    .grid {
      grid-template-columns: 1fr;
      gap: 12px;
      padding: 14px 18px 16px;
    }
    .head,
    .foot {
      padding-left: 18px;
      padding-right: 18px;
    }
  }
  @keyframes md-keys-sheet {
    from {
      opacity: 0;
      transform: translateY(24px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .back,
    .panel {
      animation: none !important;
    }
  }
</style>
