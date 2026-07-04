<script lang="ts">
  // OverlayPrompt — the hotkey-launched always-on-top composer.
  // v0.1.0 ships this as the primary user entry point: type a
  // message, hit Enter, the assistant reply streams into Chat.
  //
  // Two surface modes:
  //   - "compact"  : input + model chip + voice toggle (default)
  //   - "expanded" : + a 200px-tall rolling transcript above the input
  //
  // Glass + hairline border, slide-up + fade entrance, 5s
  // inactivity auto-dismiss (driven by the overlay store), Esc
  // dismisses (also via the store). Submit routes to Chat BEFORE
  // sending so the streamed reply lands on a visible page.

  import VoiceOrb from './VoiceOrb.svelte'
  import Input from './ui/Input.svelte'
  import { conversation } from '../stores/conversation.svelte'
  import { overlay } from '../stores/overlay.svelte'
  import { settings } from '../stores/settings.svelte'
  import { t } from '../i18n'
  import { ipc } from '../ipc/client'

  type Mode = 'compact' | 'expanded'

  interface Props {
    mode?: Mode
    voiceActive?: boolean
    onToggleVoice?: () => void
    transcript?: import('svelte').Snippet
  }

  let {
    mode = 'compact',
    voiceActive = false,
    onToggleVoice,
    transcript,
  }: Props = $props()

  let inputText = $state('')
  let sending = $state(false)
  let recording = $state(false)

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

  async function startVoice(): Promise<void> {
    if (recording || sending || !firstEnabled) return
    recording = true
    try {
      const result = await ipc.voiceListen()
      const text = result?.transcript?.trim()
      if (!text) return
      window.location.hash = '#/'
      overlay.hide()
      await conversation.send(firstEnabled.name, firstEnabled.model, text)
    } catch (err) {
      console.warn('voice.listen failed', err)
    } finally {
      recording = false
    }
  }

  function handleVoiceClick(): void {
    if (onToggleVoice) {
      onToggleVoice()
      return
    }
    void startVoice()
  }

  function onKeydown(ev: KeyboardEvent): void {
    if (ev.key === 'Enter' && !ev.shiftKey) {
      ev.preventDefault()
      void submit()
    }
    // Escape is handled globally by the overlay store.
  }
</script>

<div
  class="overlay-prompt glass-card anim-fade-in-up"
  class:expanded={mode === 'expanded'}
  role="dialog"
  aria-label={t('overlay.title')}
>
  <button
    class="overlay-close"
    type="button"
    onclick={() => overlay.hide()}
    aria-label={t('overlay.close')}
    title={t('overlay.close')}
  >
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M18 6L6 18M6 6l12 12" />
    </svg>
  </button>

  <div class="overlay-body">
    {#if mode === 'expanded'}
      <div class="overlay-transcript" aria-live="polite">
        <!-- transcript region reserved for LiveTranscript; kept empty
             here so the prompt owns its own surface, parent can
             render <LiveTranscript /> above via slot/layout. -->
        {@render transcript?.()}
      </div>
    {/if}

    <div class="overlay-stage">
      <VoiceOrb status={voiceActive ? 'listening' : 'off'} size="md" />
    </div>

    <div class="overlay-input-row">
      <Input
        size="lg"
        fullWidth
        bind:value={inputText}
        placeholder={t('overlay.placeholder')}
        disabled={sending || !firstEnabled}
        onkeydown={onKeydown}
        leading={modelChip}
        trailing={inputTrailing}
      />
    </div>

    <div class="overlay-meta">
      {#if firstEnabled}
        {t('overlay.via', firstEnabled.name, firstEnabled.model)}
      {:else}
        {t('overlay.no_provider')}
      {/if}
    </div>
  </div>
</div>

{#snippet modelChip()}
  {#if firstEnabled}
    <span class="model-chip" title={t('overlay.model_chip_title')}>
      <span class="model-dot" aria-hidden="true"></span>
      <span class="model-name">{firstEnabled.name}</span>
    </span>
  {/if}
{/snippet}

{#snippet inputTrailing()}
  <button
    class="voice-toggle"
    type="button"
    class:active={voiceActive || recording}
    onclick={handleVoiceClick}
    disabled={!firstEnabled}
    aria-label={t('overlay.voice_toggle')}
    title={t('overlay.voice_toggle')}
  >
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <rect x="9" y="3" width="6" height="12" rx="3" />
      <path d="M5 11a7 7 0 0 0 14 0" />
      <path d="M12 18v3" />
    </svg>
  </button>
  <button
    class="send-btn"
    type="button"
    onclick={() => void submit()}
    disabled={!inputText.trim() || sending || !firstEnabled}
    aria-label={t('overlay.send')}
    title={t('overlay.send')}
  >
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M12 19V5M5 12l7-7 7 7" />
    </svg>
  </button>
{/snippet}

<style>
  .overlay-prompt {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: var(--space-6) var(--space-5) var(--space-5);
    position: relative;
    color: var(--text);
    background: var(--surface-glass);
    backdrop-filter: var(--glass-blur-heavy);
    -webkit-backdrop-filter: var(--glass-blur-heavy);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-2xl);
    box-shadow: var(--shadow-xl), var(--shadow-inset);
    min-width: 520px;
    max-width: 720px;
    width: 100%;
  }

  .overlay-body {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    flex: 1;
    min-height: 0;
  }

  .overlay-stage {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-3) 0;
  }

  .overlay-transcript {
    height: 200px;
    overflow-y: auto;
    overscroll-behavior: contain;
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
    background: var(--surface-1);
    padding: var(--space-3);
    scrollbar-width: thin;
  }

  .overlay-input-row {
    display: flex;
    align-items: center;
    width: 100%;
  }

  .overlay-input-row :global(.input-wrap) {
    width: 100%;
  }

  .overlay-meta {
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    color: var(--text-faint);
    letter-spacing: var(--tracking-wide);
    text-align: center;
    text-transform: uppercase;
  }

  .overlay-close {
    position: absolute;
    top: var(--space-3);
    right: var(--space-3);
    width: 32px;
    height: 32px;
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--text-faint);
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-pill);
    cursor: pointer;
    transition: color var(--transition-fast) ease,
                background-color var(--transition-fast) ease,
                border-color var(--transition-fast) ease;
  }
  .overlay-close svg { width: 14px; height: 14px; }
  .overlay-close:hover {
    color: var(--text);
    background: var(--surface-3);
    border-color: var(--border-strong);
  }

  .model-chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    color: var(--text);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
  }
  .model-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent);
    box-shadow: 0 0 8px var(--accent-glow);
  }

  .voice-toggle,
  .send-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid transparent;
    cursor: pointer;
    transition: background-color var(--transition-fast) ease,
                color var(--transition-fast) ease,
                border-color var(--transition-fast) ease,
                box-shadow var(--transition-fast) ease,
                transform var(--transition-fast) var(--ease-spring);
  }
  .voice-toggle { width: 32px; height: 32px; border-radius: var(--radius-pill); color: var(--text-muted); background: transparent; }
  .voice-toggle svg { width: 16px; height: 16px; }
  .voice-toggle:hover:not(:disabled) {
    color: var(--text);
    background: var(--surface-3);
  }
  .voice-toggle.active {
    color: var(--accent);
    background: var(--accent-soft);
    border-color: var(--accent-soft);
  }

  .send-btn {
    width: 36px;
    height: 36px;
    margin-left: 2px;
    border-radius: var(--radius-pill);
    color: var(--text-inverse);
    background: var(--accent-gradient);
    box-shadow: var(--shadow-inset), 0 1px 2px rgba(0, 0, 0, 0.25);
  }
  .send-btn svg { width: 16px; height: 16px; }
  .send-btn:hover:not(:disabled) {
    box-shadow: var(--shadow-inset), 0 0 20px var(--accent-glow);
    transform: translateY(-1px);
  }
  .send-btn:active:not(:disabled) { transform: translateY(0) scale(0.97); }
  .send-btn:disabled,
  .voice-toggle:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .expanded .overlay-transcript { display: block; }

  @media (prefers-reduced-motion: reduce) {
    .overlay-prompt { animation: none; }
  }
</style>
