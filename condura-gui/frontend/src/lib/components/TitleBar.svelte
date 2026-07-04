<script lang="ts">
  import IconButton from './ui/IconButton.svelte'
  import { daemon } from '../stores/daemon.svelte'

  interface Props {
    title: string
  }

  let { title }: Props = $props()

  function openPalette(): void {
    window.dispatchEvent(new CustomEvent('synaptic:open-palette'))
  }

  function back(): void {
    if (window.history.length > 1) window.history.back()
    else window.location.hash = '#/'
  }

  // The daemon store may not have connected yet; we just read
  // `.connected` for the indicator color. No subscriptions needed.
  void daemon
</script>

<header class="titlebar">
  <div class="titlebar-drag"></div>

  <button class="titlebar-back" type="button" aria-label="Back" onclick={back}>
    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M19 12H5M12 19l-7-7 7-7" />
    </svg>
  </button>

  <h1 class="titlebar-title">{title}</h1>

  <div class="titlebar-end">
    <IconButton variant="ghost" size="sm" ariaLabel="Search" onclick={openPalette}>
      <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <circle cx="11" cy="11" r="7" />
        <path d="M21 21l-4.3-4.3" />
      </svg>
    </IconButton>
    <IconButton variant="ghost" size="sm" ariaLabel="Settings" onclick={() => { window.location.hash = '#/settings' }}>
      <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 0 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 0 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 0 1 0-4h.1A1.7 1.7 0 0 0 4.6 9a1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 0 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 0 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z" />
      </svg>
    </IconButton>
  </div>
</header>

<style>
  .titlebar {
    position: relative;
    display: flex;
    align-items: center;
    gap: var(--space-2);
    height: var(--titlebar-height);
    padding: 0 var(--space-3) 0 var(--space-3);
    background: var(--surface-1);
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
    -webkit-app-region: drag;
  }

  /* Inert drag region so the OS can move the window — but only the
     empty space; the buttons override the region. */
  .titlebar-drag {
    position: absolute;
    inset: 0;
    -webkit-app-region: drag;
    pointer-events: none;
  }

  .titlebar-back,
  .titlebar-end :global(.icon-btn) {
    -webkit-app-region: no-drag;
    position: relative;
    z-index: 1;
  }

  .titlebar-back {
    appearance: none;
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    width: 28px;
    height: 28px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: background-color var(--transition-fast) ease, color var(--transition-fast) ease;
  }
  .titlebar-back:hover { background: var(--surface-2); color: var(--text); }

  .titlebar-title {
    flex: 1;
    font-family: var(--font-display);
    font-size: var(--size-md);
    font-weight: var(--weight-medium);
    color: var(--text);
    letter-spacing: var(--tracking-tight);
    margin: 0;
    text-align: center;
    pointer-events: none;
  }

  .titlebar-end {
    display: flex;
    align-items: center;
    gap: 4px;
  }
</style>