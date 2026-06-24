<script lang="ts">
  // Overlay prompt — the hotkey-launched always-on-top composer.
  // v0.1.0 ships this as the primary user entry point: type a
  // message, hit Enter, the assistant reply streams into Chat.
  // Provider + model are auto-picked from the first enabled
  // provider in settings.

  import VoiceOrb from './VoiceOrb.svelte'
  import { conversation } from '../stores/conversation.svelte'
  import { overlay } from '../stores/overlay.svelte'
  import { settings } from '../stores/settings.svelte'
  import { t } from '../i18n'

  let inputText = $state('')
  let sending = $state(false)

  // Pick the first enabled provider. Stable across renders —
  // we don't want a settings refetch to swap providers mid-typing.
  const firstEnabled = $derived.by(() => {
    const providers = settings.config?.llm?.providers ?? {}
    for (const [name, p] of Object.entries(providers)) {
      if (p?.enabled && p?.default_model) {
        return { name, model: p.default_model }
      }
    }
    return null
  })

  async function submit(): Promise<void> {
    const text = inputText.trim()
    if (!text || sending) return
    if (!firstEnabled) return
    sending = true
    const target = firstEnabled
    inputText = ''
    // Route to chat BEFORE sending so the streamed reply lands
    // on a visible page, not behind a dismissed overlay.
    window.location.hash = '#/'
    overlay.hide()
    try {
      await conversation.send(target.name, target.model, text)
    } finally {
      sending = false
    }
  }

  function onKeydown(ev: KeyboardEvent): void {
    if (ev.key === 'Enter' && !ev.shiftKey) {
      ev.preventDefault()
      void submit()
    }
    // Escape is handled globally by the overlay store.
  }
</script>

<div class="overlay-prompt">
  <VoiceOrb />
  <div class="overlay-input-row">
    <!-- svelte-ignore a11y_autofocus -->
    <input
      type="text"
      class="overlay-input"
      placeholder={t('overlay.placeholder')}
      bind:value={inputText}
      onkeydown={onKeydown}
      disabled={sending || !firstEnabled}
      autofocus
    />
    <button
      class="overlay-send"
      type="button"
      onclick={() => void submit()}
      disabled={!inputText.trim() || sending || !firstEnabled}
      aria-label={t('overlay.send')}
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 19V5M5 12l7-7 7 7" /></svg>
    </button>
    <button
      class="overlay-close"
      type="button"
      onclick={() => overlay.hide()}
      aria-label={t('overlay.close')}
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M18 6L6 18M6 6l12 12" /></svg>
    </button>
  </div>
  <div class="overlay-meta">
    {#if firstEnabled}
      {t('overlay.via', firstEnabled.name, firstEnabled.model)}
    {:else}
      {t('overlay.no_provider')}
    {/if}
  </div>
</div>

<style>
  .overlay-prompt {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    gap: var(--space-4);
    padding: var(--space-6);
  }

  .overlay-input-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    max-width: 620px;
  }

  .overlay-input {
    flex: 1;
    padding: 16px 26px;
    font-size: var(--size-xl);
    font-family: var(--font-sans);
    color: var(--color-text);
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-pill);
    box-shadow: var(--shadow-md), var(--shadow-inset);
    outline: none;
    transition: border-color var(--transition-base), box-shadow var(--transition-base);
  }
  .overlay-input::placeholder {
    color: var(--color-text-faint);
  }
  .overlay-input:focus {
    border-color: var(--color-accent);
    box-shadow: var(--shadow-focus), var(--shadow-glow-accent);
  }
  .overlay-input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .overlay-send {
    width: 44px;
    height: 44px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    background: var(--color-accent-gradient);
    border: none;
    border-radius: 50%;
    cursor: pointer;
    box-shadow: var(--shadow-sm);
    transition: box-shadow var(--transition-base), transform var(--transition-base);
  }
  .overlay-send svg {
    width: 20px;
    height: 20px;
  }
  .overlay-send:hover:not(:disabled) {
    box-shadow: var(--shadow-glow-strong);
    transform: translateY(-1px);
  }
  .overlay-send:active:not(:disabled) {
    transform: translateY(0);
  }
  .overlay-send:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .overlay-close {
    width: 36px;
    height: 36px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--color-text-faint);
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 50%;
    cursor: pointer;
    transition: color var(--transition-base), border-color var(--transition-base);
  }
  .overlay-close svg {
    width: 16px;
    height: 16px;
  }
  .overlay-close:hover {
    color: var(--color-text);
    border-color: var(--glass-border-hover);
  }

  .overlay-meta {
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    color: var(--color-text-faint);
    letter-spacing: var(--tracking-wide);
  }
</style>
