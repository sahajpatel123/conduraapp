<script lang="ts">
  import { onMount } from 'svelte'
  import { ipc } from '../../ipc/client'
  import { onboarding } from '../../stores/onboarding.svelte'
  import type { EULADocument } from '../../ipc/types'
  import Button from '../ui/Button.svelte'
  import Divider from '../ui/Divider.svelte'
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
  <header class="head">
    <h1>{t('onboarding.eula.welcome')}</h1>
    <p class="lede">{t('onboarding.eula.intro')}</p>
  </header>

  {#if loading}
    <p class="muted">{t('onboarding.eula.loading')}</p>
  {:else if loadError}
    <p class="error">{t('onboarding.eula.load_error', loadError)}</p>
  {:else if doc}
    <div class="eula-meta">
      <strong>{EULA_TITLE}</strong>
      <span class="version">
        {doc.version}{doc.updated_at ? ` · ${t('onboarding.eula.updated', doc.updated_at)}` : ''}
      </span>
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

    <Divider />

    <label class="checkbox" class:disabled={!scrolledToBottom}>
      <input
        type="checkbox"
        checked={accepted}
        disabled={!scrolledToBottom}
        onchange={(e) => (accepted = (e.target as HTMLInputElement).checked)}
      />
      <span class="checkbox-stack">
        <span class="checkbox-line">
          {t('onboarding.eula.accept', EULA_TITLE)}
        </span>
        <span class="checkbox-sub">
          {t('onboarding.eula.accept_subline')}
        </span>
      </span>
    </label>

    {#if !scrolledToBottom}
      <p class="scroll-cue">
        {t('onboarding.eula.scroll_cue')}
      </p>
    {/if}
  {/if}

  {#if onboarding.error}
    <p class="error">{onboarding.error}</p>
  {/if}

  <div class="actions center">
    <Button
      variant="primary"
      size="lg"
      onclick={accept}
      disabled={!canAccept}
      loading={onboarding.busy}
    >
      {onboarding.busy ? t('onboarding.eula.saving') : t('onboarding.eula.accept_button')}
    </Button>
  </div>
</div>

<style>
  .wizard {
    width: 100%;
    max-width: 640px;
    padding: var(--space-5);
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-3);
    background: var(--surface-glass);
    border: 1px solid var(--border);
    border-radius: var(--radius-2xl);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    box-shadow: var(--shadow-md), var(--shadow-inset);
  }

  .head { display: flex; flex-direction: column; align-items: center; gap: var(--space-2); }

  h1 {
    font-family: var(--font-display);
    font-size: var(--size-3xl);
    font-weight: var(--weight-light);
    letter-spacing: var(--tracking-tighter);
    line-height: var(--leading-tight);
    margin: 0;
  }

  .lede {
    font-size: var(--size-md);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    margin: 0;
    max-width: 44ch;
  }

  .eula-meta {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    width: 100%;
    justify-content: space-between;
    padding: 0 var(--space-1);
  }
  .eula-meta .version {
    font-family: var(--font-mono);
    font-size: var(--size-sm);
    color: var(--text-faint);
  }

  .eula-body {
    width: 100%;
    height: 280px;
    overflow-y: auto;
    text-align: left;
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-4) var(--space-5);
    box-shadow: var(--shadow-inset);
  }
  .eula-body pre {
    white-space: pre-wrap;
    word-wrap: break-word;
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    line-height: var(--leading-relaxed);
    color: var(--text-muted);
    margin: 0;
  }
  .eula-body::-webkit-scrollbar { width: 6px; }
  .eula-body::-webkit-scrollbar-track { background: transparent; }
  .eula-body::-webkit-scrollbar-thumb {
    background: var(--border-strong);
    border-radius: var(--radius-pill);
  }

  .checkbox {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    text-align: left;
    cursor: pointer;
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
    background: var(--surface-1);
    transition: border-color var(--transition-base), background var(--transition-base);
    width: 100%;
  }
  .checkbox:hover:not(.disabled) {
    border-color: var(--border-strong);
  }
  .checkbox.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .checkbox input {
    margin-top: 3px;
    width: 16px;
    height: 16px;
    accent-color: var(--accent);
    flex-shrink: 0;
  }
  .checkbox-stack { display: flex; flex-direction: column; gap: 2px; }
  .checkbox-line {
    color: var(--text);
    font-size: var(--size-sm);
    line-height: var(--leading-snug);
  }
  .checkbox-sub {
    color: var(--text-faint);
    font-size: var(--size-xs);
    line-height: var(--leading-snug);
  }

  .scroll-cue {
    color: var(--text-faint);
    font-size: var(--size-xs);
    margin: 0;
    font-family: var(--font-mono);
    letter-spacing: var(--tracking-wider);
    text-transform: uppercase;
  }

  .actions {
    display: flex;
    justify-content: center;
    margin-top: var(--space-3);
    width: 100%;
  }

  .error {
    color: var(--error);
    font-size: var(--size-sm);
  }

  .muted {
    color: var(--text-muted);
    font-size: var(--size-md);
  }
</style>
