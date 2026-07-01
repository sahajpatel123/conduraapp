<!--
  ConversationDrawer — the history of conversations.

  Per spec §11.2: left-edge drawer, 320px wide. Slides in from left,
  pushes the chat surface aside (not overlays). Each row: date in sans,
  first sentence of user's request in serif, single small plum dot if
  agent acted (completion indicator).

  Search field at top in serif with faint plum underline. Real-time
  filter, 40ms stagger on results.

  Props:
    conversations — array of {id, date, firstSentence, agentActed, active}
    open          — drawer visibility (controlled)
    onselect      — fired when a row is clicked
    onclose       — fired when user dismisses
-->
<script lang="ts">
  import Pulse from './Pulse.svelte';
  import Dot from './Dot.svelte';

  interface Conversation {
    id: string;
    date: string;       // e.g., "2026-06-30" or "Today"
    firstSentence: string;
    agentActed?: boolean;
    active?: boolean;
  }

  interface Props {
    conversations?: Conversation[];
    open?: boolean;
    onselect?: (id: string) => void;
    onclose?: () => void;
  }

  let { conversations = [], open = false, onselect, onclose }: Props = $props();

  let search = $state('');
  let searchInputEl: HTMLInputElement | undefined = $state();
  let listEl = $state<HTMLDivElement | undefined>();
  let focusedRowIndex = $state(-1);

  $effect(() => {
    if (open && searchInputEl) {
      searchInputEl.focus();
      focusedRowIndex = -1;
    }
  });

  let filtered = $derived(
    search.trim()
      ? conversations.filter((c) =>
          c.firstSentence.toLowerCase().includes(search.toLowerCase())
        )
      : conversations
  );

  function navigateList(e: KeyboardEvent): void {
    if (!listEl) return
    const rows = listEl.querySelectorAll<HTMLButtonElement>('.row')
    if (rows.length === 0) return

    const handleKey = (key: string) => {
      if (key === 'ArrowDown') {
        e.preventDefault()
        focusedRowIndex = focusedRowIndex < rows.length - 1 ? focusedRowIndex + 1 : 0
        rows[focusedRowIndex]?.focus()
      } else if (key === 'ArrowUp') {
        e.preventDefault()
        focusedRowIndex = focusedRowIndex > 0 ? focusedRowIndex - 1 : rows.length - 1
        rows[focusedRowIndex]?.focus()
      } else if (key === 'Home') {
        e.preventDefault()
        focusedRowIndex = 0
        rows[0]?.focus()
      } else if (key === 'End') {
        e.preventDefault()
        focusedRowIndex = rows.length - 1
        rows[rows.length - 1]?.focus()
      } else if (key === 'Enter' || key === ' ') {
        e.preventDefault()
        const btn = rows[focusedRowIndex >= 0 ? focusedRowIndex : 0]
        btn?.click()
      } else if (key === 'Escape') {
        onclose?.()
      }
    }

    handleKey(e.key)
  }
</script>

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="scrim" onclick={onclose} onkeydown={(e) => { if (e.key === 'Escape') onclose?.(); }} aria-hidden="true"></div>
{/if}

<aside
  class="drawer"
  class:drawer--open={open}
  role="complementary"
  aria-label="Conversation history"
  aria-hidden={!open}
>
  <div class="drawer__header">
    <div class="drawer__title">
      <Pulse state="idle" size="sm" label="History" />
      <span>History</span>
    </div>

    <div class="drawer__search">
      <input
        class="drawer__search-input"
        type="search"
        placeholder="Search conversations"
        bind:value={search}
        bind:this={searchInputEl}
        aria-label="Search conversations"
      />
    </div>
  </div>

  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div bind:this={listEl} class="drawer__list" onkeydown={navigateList}>
    {#each filtered as conv, i (conv.id)}
      <button
        class="row"
        class:row--active={conv.active}
        type="button"
        onclick={() => onselect?.(conv.id)}
        style="animation-delay: {Math.min(i, 8) * 40}ms"
      >
        <div class="row__date">{conv.date}</div>
        <div class="row__text">
          <span class="row__sentence">{conv.firstSentence}</span>
        </div>
        {#if conv.agentActed}
          <span class="row__indicator" aria-label="Agent acted on this">
            <Dot variant="accent" size="xs" />
          </span>
        {/if}
      </button>
    {/each}
  </div>
</aside>

<style>
  /* Drawer slides in from the left edge */
  .drawer {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: 320px;
    background-color: var(--surface-raised);
    border-right: 1px solid var(--border-default);
    box-shadow: var(--shadow-3);
    z-index: var(--z-sticky);
    display: flex;
    flex-direction: column;
    transform: translateX(-100%);
    transition: transform var(--duration-base) var(--ease-decelerate);
    pointer-events: none;
  }

  .drawer--open {
    transform: translateX(0);
    pointer-events: auto;
  }

  /* Scrim — invisible but blocks clicks behind */
  .scrim {
    position: fixed;
    inset: 0;
    z-index: calc(var(--z-sticky) - 1);
    background-color: transparent;
  }

  .drawer__header {
    padding: var(--space-5);
    border-bottom: 1px solid var(--border-subtle);
  }

  .drawer__title {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--content-tertiary);
    margin-bottom: var(--space-4);
  }

  .drawer__search {
    position: relative;
  }

  .drawer__search-input {
    width: 100%;
    height: 36px;
    padding: 0 var(--space-3);
    font-family: var(--font-serif);
    font-size: var(--text-body-size);
    color: var(--content-primary);
    background-color: transparent;
    border: none;
    border-bottom: 1px solid var(--border-subtle);
    outline: none;
    transition: border-color var(--duration-fast) var(--ease-standard);
  }

  .drawer__search-input::placeholder {
    color: var(--content-tertiary);
    font-style: italic;
  }

  .drawer__search-input:focus-visible {
    border-bottom-color: var(--content-accent);
    box-shadow: 0 1px 0 0 var(--content-accent);
  }

  .drawer__list {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-2) 0;
  }

  .row {
    display: grid;
    grid-template-columns: 64px 1fr auto;
    gap: var(--space-3);
    align-items: center;
    width: 100%;
    padding: var(--space-3) var(--space-5);
    background-color: transparent;
    border: none;
    border-left: 2px solid transparent;
    cursor: pointer;
    text-align: left;
    font-family: var(--font-sans);
    transition: background-color var(--duration-fast) var(--ease-standard);
    animation: row-in var(--duration-base) var(--ease-decelerate) both;
  }

  @keyframes row-in {
    from {
      opacity: 0;
      transform: translateX(-8px);
    }
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }

  .row:hover {
    background-color: var(--paper-warm-50);
  }

  .row:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: -2px;
  }

  .row--active {
    background-color: var(--paper-warm-100);
    border-left-color: var(--content-accent);
  }

  .row__date {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
  }

  .row__text {
    min-width: 0;
  }

  .row__sentence {
    font-family: var(--font-serif);
    font-size: var(--text-body-sm-size);
    line-height: 1.4;
    color: var(--content-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  .row__indicator {
    flex-shrink: 0;
  }

  @media (prefers-reduced-motion: reduce) {
    .row {
      animation: none;
    }
  }
</style>