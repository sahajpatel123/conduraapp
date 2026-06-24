<script lang="ts">
  import HotkeyRecorder from '../HotkeyRecorder.svelte'
  import { onboarding } from '../../stores/onboarding.svelte'
  import { t } from '../../i18n'

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
  <h2>{t('onboarding.hotkey.title')}</h2>
  <p class="muted">
    {t('onboarding.hotkey.intro')}
  </p>

  <HotkeyRecorder value={combo} {onRecord} />

  {#if onboarding.error}
    <p class="error">{onboarding.error}</p>
  {/if}

  <div class="actions">
    <button class="btn btn-ghost" onclick={back} disabled={onboarding.busy}>← {t('onboarding.hotkey.back')}</button>
    <button class="btn btn-primary" onclick={cont} disabled={!canContinue}>
      {onboarding.busy ? t('onboarding.hotkey.saving') : t('onboarding.hotkey.continue')}
    </button>
  </div>
</div>

<style>
  .wizard {
    width: 100%;
    max-width: 540px;
    padding: var(--space-6) var(--space-5);
    text-align: center;
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  h2 {
    font-size: var(--size-2xl);
    font-weight: var(--weight-semibold);
    letter-spacing: var(--tracking-tight);
    margin-bottom: var(--space-2);
    background: var(--color-accent-gradient);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .wizard > .muted {
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    margin-bottom: var(--space-5);
  }
  .actions {
    display: flex;
    justify-content: space-between;
    margin-top: var(--space-5);
  }
  .error {
    color: var(--color-error);
    font-size: var(--size-sm);
    margin-top: var(--space-3);
  }
</style>
