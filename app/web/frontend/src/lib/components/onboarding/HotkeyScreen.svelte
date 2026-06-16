<script lang="ts">
  import HotkeyRecorder from '../HotkeyRecorder.svelte'
  import { onboarding } from '../../stores/onboarding.svelte'

  // Pre-fill from any previously recorded value (e.g. user went
  // back from the Ready screen). No silent default per locked
  // decision #8 — the user must actively set a hotkey.
  let combo = $state(onboarding.daemon?.steps?.hotkey?.data ?? '')

  function onRecord(c: string): void {
    combo = c
  }

  async function cont(): Promise<void> {
    if (!combo) return
    await onboarding.setHotkey(combo)
  }

  async function back(): Promise<void> {
    await onboarding.back()
  }

  const canContinue = $derived(!!combo && !onboarding.busy)
</script>

<div class="wizard hotkey">
  <h2>Set your hotkey</h2>
  <p class="muted">
    Press the keys you want to use to summon Synaptic from anywhere. You can change this later in Settings.
  </p>

  <HotkeyRecorder value={combo} {onRecord} />

  {#if onboarding.error}
    <p class="error">{onboarding.error}</p>
  {/if}

  <div class="actions">
    <button class="btn btn-ghost" onclick={back} disabled={onboarding.busy}>← Back</button>
    <button class="btn btn-primary" onclick={cont} disabled={!canContinue}>
      {onboarding.busy ? 'Saving…' : 'Continue →'}
    </button>
  </div>
</div>

<style>
  .wizard {
    width: 100%;
    max-width: 540px;
    padding: var(--space-6) var(--space-5);
    text-align: center;
  }
  h2 {
    font-size: var(--size-2xl);
    font-weight: 600;
    margin-bottom: var(--space-2);
  }
  .muted {
    color: var(--color-text-muted);
    font-size: var(--size-md);
    margin-bottom: var(--space-5);
  }
  .actions {
    display: flex;
    justify-content: space-between;
    margin-top: var(--space-5);
  }
  .btn {
    padding: 12px 24px;
    border-radius: var(--radius-pill);
    font-size: var(--size-md);
    font-weight: 500;
    cursor: pointer;
    border: none;
    transition: all var(--transition-spring);
  }
  .btn-primary {
    background: var(--color-accent-gradient);
    color: white;
  }
  .btn-primary:hover:not(:disabled) {
    box-shadow: var(--shadow-glow);
    transform: translateY(-1px);
  }
  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .btn-ghost {
    background: transparent;
    color: var(--color-text-muted);
    border: 1px solid var(--glass-border);
  }
  .btn-ghost:hover {
    color: var(--color-text);
  }
  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .error {
    color: var(--color-error);
    font-size: var(--size-sm);
    margin-top: var(--space-3);
  }
</style>
