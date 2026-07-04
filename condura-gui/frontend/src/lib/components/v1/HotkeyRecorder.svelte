<!--
  HotkeyRecorder — captures a key combo from the user.

  Used in Settings (changing the wake hotkey) and in the first-run wizard
  (recording initial hotkey). This is the *separate* component for the
  field; the wizard has its own version inline.

  Props:
    value      — current combo string (e.g., "⌘⇧Space")
    label      — visible label
    onrecord   — fired with the new combo string
-->
<script lang="ts">
  interface Props {
    value?: string;
    label?: string;
    onrecord?: (combo: string) => void;
  }

  let { value = '', label = 'Hotkey', onrecord }: Props = $props();

  let recording = $state(false);
  let display = $state(value);

  function normalizeKey(e: KeyboardEvent): string | null {
    const parts: string[] = [];
    if (e.metaKey) parts.push('⌘');
    if (e.ctrlKey) parts.push('^');
    if (e.altKey) parts.push('⌥');
    if (e.shiftKey) parts.push('⇧');

    const key = e.key;
    if (key === ' ') parts.push('Space');
    else if (key === 'Escape') return null;
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

  function handleKey(e: KeyboardEvent) {
    if (!recording) return;
    e.preventDefault();
    e.stopPropagation();

    if (e.key === 'Escape') {
      recording = false;
      return;
    }

    const combo = normalizeKey(e);
    if (combo && combo.length > 1) {
      if (combo.includes('⌘') || combo.includes('^') || combo.includes('⌥') || combo.includes('⇧')) {
        display = combo;
        onrecord?.(combo);
        recording = false;
      }
    }
  }

  function startRecording() {
    recording = true;
  }
</script>

<svelte:window onkeydown={handleKey} />

<div class="recorder" class:recorder--recording={recording}>
  <div class="recorder__head">
    <span class="recorder__label">{label}</span>
    {#if recording}
      <span class="recorder__hint">press a combo · esc to cancel</span>
    {/if}
  </div>

  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <button
    class="recorder__field"
    type="button"
    onclick={startRecording}
    onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') startRecording(); }}
  >
    {#if display}
      <span class="recorder__combo">{display}</span>
    {:else if recording}
      <span class="recorder__recording">
        <span class="recorder__dot"></span>
        Listening…
      </span>
    {:else}
      <span class="recorder__placeholder">Click to record</span>
    {/if}
  </button>
</div>

<style>
  .recorder {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .recorder__head {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
  }

  .recorder__label {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-primary);
  }

  .recorder__hint {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    letter-spacing: 0.02em;
  }

  .recorder__field {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    min-height: 64px;
    padding: var(--space-3);
    background-color: var(--surface-base);
    border: 1.5px solid var(--border-default);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition:
      border-color var(--duration-fast) var(--ease-standard),
      background-color var(--duration-fast) var(--ease-standard);
  }
  .recorder__field:hover {
    border-color: var(--border-strong);
  }
  .recorder__field:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: 2px;
  }

  .recorder--recording .recorder__field {
    border-color: var(--content-accent);
    border-style: dashed;
    background-color: var(--plum-50);
  }

  .recorder__combo {
    font-family: var(--font-serif);
    font-size: var(--text-h4-size);
    color: var(--content-accent);
    letter-spacing: 0.04em;
  }

  .recorder__placeholder {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
  }

  .recorder__recording {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-secondary);
  }

  .recorder__dot {
    width: 8px;
    height: 8px;
    border-radius: var(--radius-pill);
    background-color: var(--content-accent);
    animation: rec 1s ease-in-out infinite;
  }

  @keyframes rec {
    0%, 100% { opacity: 0.4; transform: scale(0.9); }
    50% { opacity: 1; transform: scale(1.1); }
  }

  @media (prefers-reduced-motion: reduce) {
    .recorder__dot {
      animation: none;
      opacity: 1;
    }
  }
</style>