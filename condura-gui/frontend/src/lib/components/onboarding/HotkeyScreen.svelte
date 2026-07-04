<script lang="ts">
  import HotkeyRecorder from '../HotkeyRecorder.svelte'
  import { onboarding } from '../../stores/onboarding.svelte'
  import Button from '../ui/Button.svelte'
  import Kbd from '../ui/Kbd.svelte'
  import { t } from '../../i18n'

  // Pre-fill from any previously recorded value (e.g. user went
  // back from the Ready screen). No silent default per locked
  // decision #8 — the user must actively set a hotkey.
  let combo = $state(onboarding.daemon?.steps?.hotkey?.data ?? '')

  // Three preset chips shown beneath the recorder. They are the
  // same suggestions the wizard surfaces on first run, regardless
  // of platform — the spec is platform-aware inside the recorder.
  const presets = ['Option+Option', 'Cmd+Shift+Space', 'Ctrl+Space']

  function onRecord(c: string): void {
    combo = c
  }

  function pickPreset(c: string): void {
    combo = c
  }

  async function cont(): Promise<void> {
    if (!combo) return
    onboarding.setHotkey(combo)
    await onboarding.saveHotkey()
  }

  async function skip(): Promise<void> {
    onboarding.setHotkey('')
    await onboarding.skipStep('hotkey')
  }

  async function back(): Promise<void> {
    await onboarding.back()
  }

  const canContinue = $derived(!!combo && !onboarding.busy)
</script>

<div class="wizard hotkey">
  <header class="head">
    <h2>{t('onboarding.hotkey.title')}</h2>
    <p class="muted">
      {t('onboarding.hotkey.intro')}
    </p>
  </header>

  <div class="capture-wrap">
    <HotkeyRecorder value={combo} {onRecord} />
  </div>

  <div class="presets">
    <span class="preset-label">{t('onboarding.hotkey.suggest_a_combo')}</span>
    <div class="preset-row">
      {#each presets as p (p)}
        <button
          type="button"
          class="preset-chip"
          class:active={combo === p}
          onclick={() => pickPreset(p)}
        >
          <Kbd label={p} />
        </button>
      {/each}
    </div>
  </div>

  {#if onboarding.error}
    <p class="error">{onboarding.error}</p>
  {/if}

  <div class="actions">
    <button class="back-link" type="button" onclick={back} disabled={onboarding.busy}>
      ← {t('onboarding.hotkey.back')}
    </button>
    <div class="actions-right">
      <button class="skip-link" type="button" onclick={skip} disabled={onboarding.busy}>
        {t('onboarding.hotkey.skip', 'Set later')}
      </button>
      <Button
        variant="primary"
        size="md"
        onclick={cont}
        disabled={!canContinue}
        loading={onboarding.busy}
      >
        {onboarding.busy ? t('onboarding.hotkey.saving') : t('onboarding.hotkey.continue')}
      </Button>
    </div>
  </div>
</div>

<style>
  .wizard {
    width: 100%;
    max-width: 560px;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    align-items: center;
  }

  .head { text-align: center; display: flex; flex-direction: column; gap: var(--space-2); align-items: center; }

  h2 {
    font-family: var(--font-display);
    font-size: var(--size-2xl);
    font-weight: var(--weight-light);
    letter-spacing: var(--tracking-tight);
    margin: 0;
  }
  .muted {
    color: var(--text-muted);
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    margin: 0;
    max-width: 48ch;
  }

  .capture-wrap { width: 100%; }

  .presets {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
  }
  .preset-label {
    color: var(--text-faint);
    font-size: var(--size-xs);
    font-family: var(--font-mono);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
  }
  .preset-row {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--space-2);
  }
  .preset-chip {
    background: var(--surface-2);
    border: 1px solid var(--border);
    border-radius: var(--radius-pill);
    padding: 6px 12px;
    cursor: pointer;
    transition: border-color var(--transition-base), background var(--transition-base);
  }
  .preset-chip:hover {
    border-color: var(--border-strong);
    background: var(--surface-3);
  }
  .preset-chip.active {
    border-color: var(--accent);
    background: var(--accent-soft);
  }

  .actions {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .actions-right {
    display: flex;
    gap: var(--space-3);
    align-items: center;
  }

  .back-link,
  .skip-link {
    background: transparent;
    border: 0;
    padding: 0;
    cursor: pointer;
    color: var(--text-muted);
    font-size: var(--size-sm);
    font-family: inherit;
    transition: color var(--transition-fast);
  }
  .back-link:hover:not(:disabled),
  .skip-link:hover:not(:disabled) {
    color: var(--text);
  }
  .back-link:disabled,
  .skip-link:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .skip-link { color: var(--text-faint); }

  .error {
    color: var(--error);
    font-size: var(--size-sm);
  }
</style>
