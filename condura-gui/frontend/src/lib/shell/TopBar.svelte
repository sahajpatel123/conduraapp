<script lang="ts">
  /**
   * TopBar — Floating minimal top bar for the Living Paper shell.
   * 
   * A thin floating bar at the top of the content area showing:
   * - Synapse pulse dot (agent state)
   * - Page title (Instrument Serif)
   * - Right actions: ⌘K button, theme toggle
   */
  import { PulseDot, InkText, PollenNode } from '$lib/components/living'
  import type { PulsePhase } from '$lib/components/living'

  interface Props {
    title: string
    agentPhase?: PulsePhase
    onPalette?: () => void
    theme?: 'light' | 'dark'
    onThemeToggle?: () => void
  }

  let {
    title,
    agentPhase = 'idle',
    onPalette,
    theme = 'light',
    onThemeToggle,
  }: Props = $props()
</script>

<header
  class="lp-topbar lp"
  style="
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--lp-space-3) var(--lp-space-6);
    background: transparent;
    position: relative;
    z-index: 10;
    min-height: 44px;
  "
>
  <!-- Left: Agent indicator + page title -->
  <div style="display: flex; align-items: center; gap: var(--lp-space-3);">
    <PulseDot phase={agentPhase} size={6} />
    <InkText kind="title" as="h1" style="margin: 0;">{title}</InkText>
  </div>

  <!-- Right: Actions -->
  <div style="display: flex; align-items: center; gap: var(--lp-space-2);">
    <!-- ⌘K button -->
    <button
      type="button"
      class="lp-topbar-btn lp-focus"
      onclick={onPalette}
      aria-label="Open command palette (⌘K)"
      title="Open command palette (⌘K)"
      style="
        display: inline-flex;
        align-items: center;
        gap: 6px;
        padding: 5px 10px;
        border-radius: var(--lp-radius-sm);
        background: transparent;
        border: 1px solid var(--lp-ink-ghost);
        color: var(--lp-ink-mute);
        font-family: var(--lp-font-mono);
        font-size: var(--lp-text-caption);
        cursor: pointer;
        transition: background var(--lp-dur-fast) var(--lp-ease-thread),
                    color var(--lp-dur-fast) var(--lp-ease-thread);
      "
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
        <circle cx="11" cy="11" r="8" />
        <path d="M21 21l-4.35-4.35" />
      </svg>
      <span style="
        font-size: 10px;
        letter-spacing: 0.08em;
        text-transform: uppercase;
      ">⌘K</span>
    </button>

    <!-- Theme toggle -->
    <button
      type="button"
      class="lp-topbar-btn lp-focus"
      onclick={onThemeToggle}
      aria-label="Toggle theme"
      title={theme === 'light' ? 'Switch to dark' : 'Switch to light'}
      style="
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 30px;
        height: 30px;
        border-radius: var(--lp-radius-sm);
        background: transparent;
        border: 1px solid var(--lp-ink-ghost);
        color: var(--lp-ink-mute);
        cursor: pointer;
        transition: background var(--lp-dur-fast) var(--lp-ease-thread),
                    color var(--lp-dur-fast) var(--lp-ease-thread);
      "
    >
      {#if theme === 'light'}
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
        </svg>
      {:else}
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <circle cx="12" cy="12" r="5" />
          <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
        </svg>
      {/if}
    </button>
  </div>
</header>

<style>
  .lp-topbar-btn:hover {
    background: var(--lp-paper-warm);
    color: var(--lp-ink);
  }
  .lp-topbar-btn:active {
    transform: scale(0.97);
  }
</style>
