<script lang="ts">
  import { onMount } from 'svelte'
  import { ipc } from '../../ipc/client'
  import { onboarding } from '../../stores/onboarding.svelte'
  import type { EULADocument } from '../../ipc/types'
  import { t } from '../../i18n'

  const EULA_TITLE = $derived(t('onboarding.eula.title'))

  let doc = $state<EULADocument | null>(null)
  let loading = $state(true)
  let loadError = $state('')
  let scrolledToBottom = $state(false)
  let accepted = $state(false)

  let scrollEl = $state<HTMLDivElement | null>(null)

  onMount(() => {
    void ipc
      .onboardingEula()
      .then((d) => {
        doc = d
      })
      .catch((err) => {
        loadError = String(err)
      })
      .finally(() => {
        loading = false
        // If the body is short enough that there's nothing to
        // scroll, treat it as already read.
        queueMicrotask(checkScroll)
      })
  })

  function checkScroll(): void {
    if (!scrollEl) return
    const { scrollTop, scrollHeight, clientHeight } = scrollEl
    if (scrollHeight - (scrollTop + clientHeight) <= 8) {
      scrolledToBottom = true
    }
  }

  async function accept(): Promise<void> {
    if (!doc) return
    await onboarding.acceptEula(doc.version)
  }

  const canAccept = $derived(scrolledToBottom && accepted && !onboarding.busy)
</script>

<div class="wizard eula">
  <h1>{t('onboarding.eula.welcome')}</h1>
  <p class="lede">{t('onboarding.eula.intro')}</p>

  {#if loading}
    <p class="muted">{t('onboarding.eula.loading')}</p>
  {:else if loadError}
    <p class="error">{t('onboarding.eula.load_error', loadError)}</p>
  {:else if doc}
    <div class="eula-meta">
      <strong>{EULA_TITLE}</strong>
      <span class="version">{doc.version}{doc.updated_at ? ` · ${t('onboarding.eula.updated', doc.updated_at)}` : ''}</span>
    </div>

    <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
    <div
      class="eula-body"
      bind:this={scrollEl}
      onscroll={checkScroll}
      tabindex="0"
      role="document"
      aria-label={t('onboarding.eula.aria_label')}
    >
      <pre>{doc.text}</pre>
    </div>

    {#if !scrolledToBottom}
      <p class="scroll-cue">
        {t('onboarding.eula.scroll_cue')}
        <svg class="cue-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M6 9l6 6 6-6" /></svg>
      </p>
    {/if}

    <label class="checkbox" class:disabled={!scrolledToBottom}>
      <input
        type="checkbox"
        checked={accepted}
        disabled={!scrolledToBottom}
        onchange={(e) => (accepted = (e.target as HTMLInputElement).checked)}
      />
      <span>{t('onboarding.eula.accept', EULA_TITLE)}</span>
    </label>
  {/if}

  {#if onboarding.error}
    <p class="error">{onboarding.error}</p>
  {/if}

  <div class="actions center">
    <button class="btn btn-primary btn-lg cta" onclick={accept} disabled={!canAccept}>
      {onboarding.busy ? t('onboarding.eula.saving') : t('onboarding.eula.accept_button')}
    </button>
  </div>
</div>

<style>
  .wizard {
    width: 100%;
    max-width: 640px;
    padding: var(--space-6) var(--space-5);
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    animation: screen-in var(--transition-spring-soft) var(--ease-out-expo) both;
  }
  @keyframes screen-in {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: none; }
  }
  h1 {
    font-size: var(--size-3xl);
    font-weight: var(--weight-semibold);
    letter-spacing: var(--tracking-tight);
    margin-bottom: var(--space-2);
    background: var(--color-accent-gradient);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .lede {
    font-size: var(--size-md);
    color: var(--color-text-muted);
    line-height: var(--leading-relaxed);
    margin-bottom: var(--space-5);
  }
  .eula-meta {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    width: 100%;
    justify-content: space-between;
    margin-bottom: var(--space-2);
  }
  .eula-meta .version {
    font-family: var(--font-mono);
    font-size: var(--size-sm);
    color: var(--color-text-faint);
  }
  .eula-body {
    width: 100%;
    height: 320px;
    overflow-y: auto;
    text-align: left;
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
    margin-bottom: var(--space-2);
    box-shadow: var(--shadow-sm), var(--shadow-inset);
    -webkit-mask-image: linear-gradient(to bottom, transparent 0, #000 14px, #000 calc(100% - 14px), transparent 100%);
    mask-image: linear-gradient(to bottom, transparent 0, #000 14px, #000 calc(100% - 14px), transparent 100%);
  }
  .eula-body pre {
    white-space: pre-wrap;
    word-wrap: break-word;
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    line-height: var(--leading-relaxed);
    color: var(--color-text-muted);
    margin: 0;
  }
  .eula-body::-webkit-scrollbar { width: 6px; }
  .eula-body::-webkit-scrollbar-track { background: transparent; }
  .eula-body::-webkit-scrollbar-thumb {
    background: var(--color-border-strong);
    border-radius: var(--radius-pill);
  }
  .scroll-cue {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--color-accent);
    font-size: var(--size-sm);
    margin: var(--space-2) 0 var(--space-3) 0;
  }
  .cue-arrow {
    width: 14px;
    height: 14px;
    animation: bob 1.6s ease-in-out infinite;
  }
  @keyframes bob {
    0%, 100% { transform: translateY(0); opacity: 0.7; }
    50% { transform: translateY(3px); opacity: 1; }
  }
  .wizard .checkbox {
    margin: var(--space-3) 0;
    color: var(--color-text);
    font-size: var(--size-sm);
  }
  .actions {
    display: flex;
    justify-content: center;
    margin-top: var(--space-4);
  }
  .cta {
    border-radius: var(--radius-pill);
  }
  .error {
    color: var(--color-error);
    font-size: var(--size-sm);
  }
</style>
