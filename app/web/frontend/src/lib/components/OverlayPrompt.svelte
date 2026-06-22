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
      ↵
    </button>
    <button
      class="overlay-close"
      type="button"
      onclick={() => overlay.hide()}
      aria-label={t('overlay.close')}
    >
      ×
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
    gap: 12px;
    padding: 24px;
  }

  .overlay-input-row {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    max-width: 600px;
  }

  .overlay-input {
    flex: 1;
    padding: 16px 24px;
    font-size: 18px;
    font-family: var(--font-sans);
    color: var(--color-text);
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    backdrop-filter: var(--glass-blur);
    outline: none;
  }

  .overlay-input:focus {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px rgba(var(--color-accent-rgb), 0.2);
  }

  .overlay-input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .overlay-send {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
    color: white;
    background: var(--color-accent-gradient);
    border: none;
    border-radius: 50%;
    cursor: pointer;
    transition: box-shadow var(--transition-base);
  }

  .overlay-send:hover:not(:disabled) {
    box-shadow: var(--shadow-glow);
  }

  .overlay-send:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .overlay-close {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 22px;
    color: var(--color-text-faint);
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 50%;
    cursor: pointer;
    transition: color var(--transition-base), border-color var(--transition-base);
  }

  .overlay-close:hover {
    color: var(--color-text);
    border-color: var(--color-text-faint);
  }

  .overlay-meta {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--color-text-faint);
    letter-spacing: 0.02em;
  }

  .overlay-meta code {
    font-family: var(--font-mono);
    color: var(--color-text-muted);
  }
</style>
