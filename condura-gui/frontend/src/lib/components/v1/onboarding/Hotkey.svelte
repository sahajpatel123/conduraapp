<!--
  Onboarding · Screen 5 · Hotkey

  Per spec §10.5: a single recordable field. No skip — without a hotkey
  the agent is unreachable. Three suggested combos as ghosted suggestions.

  Below: voice test toggle (the "hey synaptic" wake word).
-->
<script lang="ts">
  import Button from '../Button.svelte';
  import Switch from '../Switch.svelte';

  interface Props {
    hotkey?: string;
    voiceWakeEnabled?: boolean;
    onrecord?: (combo: string) => void;
    onvoicetoggle?: (enabled: boolean) => void;
    oncontinue?: () => void;
    onback?: () => void;
  }

  let {
    hotkey = '',
    voiceWakeEnabled = false,
    onrecord,
    onvoicetoggle,
    oncontinue,
    onback,
  }: Props = $props();

  let recording = $state(false);
  let displayCombo = $state('');
  let containerEl: HTMLDivElement | undefined = $state();

  const SUGGESTIONS = ['⌥⌥', '⌘⇧Space', '^Space'];

  function normalizeKey(e: KeyboardEvent): string | null {
    // Build a human-readable combo like "⌘⇧Space"
    const parts: string[] = [];
    if (e.metaKey) parts.push('⌘');
    if (e.ctrlKey) parts.push('^');
    if (e.altKey) parts.push('⌥');
    if (e.shiftKey) parts.push('⇧');

    // Map common keys to symbols
    const key = e.key;
    if (key === ' ') parts.push('Space');
    else if (key === 'Escape') return null; // don't record Esc
    else if (key === 'Tab') parts.push('⇥');
    else if (key === 'Enter') parts.push('⏎');
    else if (key === 'ArrowUp') parts.push('↑');
    else if (key === 'ArrowDown') parts.push('↓');
    else if (key === 'ArrowLeft') parts.push('←');
    else if (key === 'ArrowRight') parts.push('→');
    else if (key.length === 1) parts.push(key.toUpperCase());
    else parts.push(key);

    return parts.join('');
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (!recording) return;
    e.preventDefault();
    e.stopPropagation();

    if (e.key === 'Escape') {
      recording = false;
      return;
    }

    const combo = normalizeKey(e);
    if (combo && combo.length > 1) {
      // Require at least one modifier
      if (combo.includes('⌘') || combo.includes('^') || combo.includes('⌥') || combo.includes('⇧')) {
        displayCombo = combo;
        onrecord?.(combo);
        recording = false;
      }
    }
  }

  function startRecording() {
    displayCombo = '';
    recording = true;
    containerEl?.focus();
  }

  function pickSuggestion(combo: string) {
    displayCombo = combo;
    onrecord?.(combo);
    recording = false;
  }
</script>

<svelte:window onkeydown={handleKeyDown} />

<div class="screen">
  <div class="screen__inner">
    <header class="screen__header">
      <h2 class="screen__title">Pick the way you'll wake me.</h2>
      <p class="screen__subtitle">Press a key combo. You can change this later in Settings.</p>
    </header>

    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="field"
      class:field--recording={recording}
      bind:this={containerEl}
      tabindex="0"
      role="button"
      aria-label="Record a hotkey"
      onclick={startRecording}
      onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') startRecording(); }}
    >
      {#if displayCombo}
        <span class="field__combo">{displayCombo}</span>
      {:else if recording}
        <span class="field__recording">
          <span class="field__recording-dot"></span>
          Press your key combo
        </span>
      {:else}
        <span class="field__placeholder">Click to record</span>
      {/if}
    </div>

    {#if !displayCombo && !recording}
      <div class="suggestions">
        <span class="suggestions__label">or pick one:</span>
        {#each SUGGESTIONS as s}
          <button class="suggestion" type="button" onclick={() => pickSuggestion(s)}>
            {s}
          </button>
        {/each}
      </div>
    {/if}

    <p class="hint">Press this from anywhere. Synaptic will appear at your cursor.</p>

    <div class="voice">
      <Switch
        label="Also say 'hey synaptic' to wake me"
        description="Local wake word, runs on your machine. You can change the phrase later."
        checked={voiceWakeEnabled}
        onchange={onvoicetoggle}
      />
    </div>

    <div class="screen__actions">
      <Button variant="tertiary" size="md" onclick={onback}>← Back</Button>
      <Button
        variant="primary"
        size="lg"
        disabled={!displayCombo}
        onclick={oncontinue}
      >
        I'm ready
      </Button>
    </div>
  </div>
</div>

<style>
  .screen {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: var(--space-9);
    background-color: var(--surface-base);
    color: var(--content-primary);
  }

  .screen__inner {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-6);
    width: 100%;
    max-width: 600px;
    text-align: center;
  }

  .screen__header {
    text-align: center;
  }

  .screen__title {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    line-height: 1.3;
    font-weight: 400;
    color: var(--content-primary);
    margin: 0 0 var(--space-2) 0;
  }

  .screen__subtitle {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    margin: 0;
  }

  /* The field — large, centered, draws the eye */
  .field {
    width: 100%;
    height: 120px;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--surface-raised);
    border: 2px solid var(--border-default);
    border-radius: var(--radius-lg);
    cursor: pointer;
    transition:
      border-color var(--duration-fast) var(--ease-standard),
      background-color var(--duration-fast) var(--ease-standard);
    outline: none;
  }

  .field:hover {
    border-color: var(--border-strong);
  }

  .field:focus-visible {
    border-color: var(--border-focus);
    box-shadow: 0 0 0 3px var(--plum-100);
  }

  .field--recording {
    border-color: var(--content-accent);
    border-style: dashed;
  }

  .field__combo {
    font-family: var(--font-serif);
    font-size: 48px;
    font-weight: 400;
    color: var(--content-accent);
    letter-spacing: 0.04em;
  }

  .field__placeholder {
    font-family: var(--font-sans);
    font-size: var(--text-body-size);
    color: var(--content-tertiary);
  }

  .field__recording {
    display: inline-flex;
    align-items: center;
    gap: var(--space-3);
    font-family: var(--font-sans);
    font-size: var(--text-body-size);
    color: var(--content-secondary);
  }

  .field__recording-dot {
    width: 8px;
    height: 8px;
    border-radius: var(--radius-pill);
    background-color: var(--content-accent);
    animation: record-pulse 1s ease-in-out infinite;
  }

  @keyframes record-pulse {
    0%, 100% { opacity: 0.4; transform: scale(0.9); }
    50% { opacity: 1; transform: scale(1.1); }
  }

  /* Ghosted suggestions */
  .suggestions {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex-wrap: wrap;
    justify-content: center;
  }

  .suggestions__label {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-muted);
  }

  .suggestion {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-secondary);
    background-color: transparent;
    border: 1px solid var(--border-default);
    border-radius: var(--radius-pill);
    padding: var(--space-1) var(--space-3);
    cursor: pointer;
    transition: border-color var(--duration-fast) var(--ease-standard);
  }

  .suggestion:hover {
    border-color: var(--border-strong);
  }

  .hint {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    margin: 0;
  }

  .voice {
    width: 100%;
    text-align: left;
    padding: var(--space-4);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
  }

  .screen__actions {
    display: flex;
    justify-content: space-between;
    gap: var(--space-3);
    width: 100%;
  }

  @media (prefers-reduced-motion: reduce) {
    .field__recording-dot {
      animation: none;
      opacity: 1;
    }
  }
</style>