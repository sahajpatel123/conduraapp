<script lang="ts">
  /**
   * EULA step — loads license from the daemon when available.
   * Offline / browser-preview: falls back to bundled public/EULA.md
   * (or a short built-in text) so re-run setup never shows a blank key soup.
   */
  import { onMount } from 'svelte'
  import { ipc } from '../../ipc/client'
  import { onboarding } from '../../stores/onboarding.svelte'
  import { daemon } from '../../stores/daemon.svelte'
  import type { EULADocument } from '../../ipc/types'
  import Button from '../ui/Button.svelte'
  import Divider from '../ui/Divider.svelte'
  import { catalogVersion, t } from '../../i18n'

  interface Props {
    onaccepted?: () => void
  }

  let { onaccepted }: Props = $props()

  // Subscribe so labels refresh if a late catalog merge arrives.
  const _catalog = $derived($catalogVersion)

  const EULA_TITLE = $derived.by(() => {
    // Read _catalog to register a dependency so the label refreshes
    // when a late i18n catalog merge arrives. Avoid the comma-operator
    // hack `(x, expr)` — Svelte 5 runes track reads directly.
    _catalog
    return t('onboarding.eula.title')
  })

  let doc = $state<EULADocument | null>(null)
  let loading = $state(true)
  let loadError = $state('')
  let usedFallback = $state(false)
  let scrolledToBottom = $state(false)
  let accepted = $state(false)

  let scrollEl = $state<HTMLDivElement | null>(null)

  const offline = $derived(!daemon.connected)
  const canAccept = $derived(
    scrolledToBottom && accepted && !onboarding.busy && !!doc && !offline
  )

  onMount(() => {
    // Reset may have failed offline (TypeError: Load failed) — don't surface that
    // as an EULA error once we have a bundled fallback.
    onboarding.error = null
    void loadDoc()
  })

  function isOfflineError(err: unknown): boolean {
    const s = String(err)
    return /Load failed|Failed to fetch|NetworkError|IPC client not started|not connected|ECONNREFUSED/i.test(
      s
    )
  }

  async function loadBundledEula(): Promise<EULADocument> {
    try {
      const res = await fetch('/EULA.md', { headers: { Accept: 'text/markdown, text/plain, */*' } })
      if (res.ok) {
        const text = await res.text()
        if (text && !text.trimStart().startsWith('<!')) {
          return { version: 'v1', text, updated_at: '2026-06-06' }
        }
      }
    } catch {
      /* use short fallback */
    }
    // Matches condura-app/internal/onboarding/eula.go ReadEULA missing-file fallback.
    return {
      version: 'v1',
      text:
        'By using Condura, you agree to the Condura Freeware EULA v1.\n\n' +
        'The full terms are available at https://condura.app/legal.\n\n' +
        'Condura is free software that runs on your machine. Planning and permission stay separate: ' +
        'only you can approve gated actions. You may stop Condura at any time via Halt.',
      updated_at: '',
    }
  }

  async function loadDoc(): Promise<void> {
    loading = true
    loadError = ''
    usedFallback = false
    scrolledToBottom = false
    accepted = false
    try {
      const d = await ipc.onboardingEula()
      doc = d
      usedFallback = false
    } catch (err) {
      // Root cause when re-running setup in a Vite browser preview: daemon
      // is not on :7666, so fetch throws TypeError: Load failed (WebKit).
      doc = await loadBundledEula()
      usedFallback = true
      loadError = isOfflineError(err)
        ? ''
        : String(err)
    } finally {
      loading = false
      queueMicrotask(checkScroll)
    }
  }

  function checkScroll(): void {
    if (!scrollEl) return
    const { scrollTop, scrollHeight, clientHeight } = scrollEl
    if (scrollHeight - (scrollTop + clientHeight) <= 8) {
      scrolledToBottom = true
    }
  }

  async function accept(): Promise<void> {
    if (!doc || offline) return
    await onboarding.acceptEula(doc.version)
    if (!onboarding.error) {
      onaccepted?.()
    }
  }
</script>

<div class="wizard eula">
  <header class="head">
    <h1>{t('onboarding.eula.welcome')}</h1>
    <p class="lede">{t('onboarding.eula.intro')}</p>
  </header>

  {#if usedFallback || offline}
    <p class="offline-banner" role="status">{t('onboarding.eula.offline_banner')}</p>
  {/if}

  {#if loading}
    <p class="muted">{t('onboarding.eula.loading')}</p>
  {:else if !doc && loadError}
    <p class="error">{t('onboarding.eula.load_error', loadError)}</p>
    <button type="button" class="retry" onclick={() => void loadDoc()}>{t('onboarding.eula.retry')}</button>
  {:else if doc}
    {#if loadError}
      <p class="warn">{t('onboarding.eula.load_error', loadError)}</p>
    {/if}

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

    <label class="checkbox" class:disabled={!scrolledToBottom || offline}>
      <input
        type="checkbox"
        checked={accepted}
        disabled={!scrolledToBottom || offline}
        onchange={(e) => (accepted = (e.target as HTMLInputElement).checked)}
      />
      <span class="checkbox-stack">
        <span class="checkbox-line">
          {t('onboarding.eula.accept', EULA_TITLE)}
        </span>
        <span class="checkbox-sub">
          {offline ? t('onboarding.eula.accept_needs_daemon') : t('onboarding.eula.accept_subline')}
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
    background: var(--md-surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    backdrop-filter: none;
    -webkit-backdrop-filter: none;
    box-shadow: none;
  }

  .head {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
  }

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

  .offline-banner {
    margin: 0;
    width: 100%;
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    border: 1px solid color-mix(in oklab, var(--accent, #2f5bff) 28%, var(--border));
    background: color-mix(in oklab, var(--accent, #2f5bff) 8%, transparent);
    color: var(--text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-snug);
    text-align: left;
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
    border-radius: var(--radius-md);
    padding: var(--space-4) var(--space-5);
    box-shadow: none;
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
  .eula-body::-webkit-scrollbar {
    width: 6px;
  }
  .eula-body::-webkit-scrollbar-track {
    background: transparent;
  }
  .eula-body::-webkit-scrollbar-thumb {
    background: var(--border-strong);
    border-radius: 4px;
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
    transition:
      border-color var(--transition-base),
      background var(--transition-base);
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
  .checkbox-stack {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
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
  .warn {
    color: var(--text-muted);
    font-size: var(--size-sm);
  }
  .muted {
    color: var(--text-muted);
    font-size: var(--size-md);
  }
  .retry {
    appearance: none;
    border: 1px solid var(--border-strong);
    background: var(--surface-1);
    color: var(--accent, #2f5bff);
    font-weight: 550;
    font-size: var(--size-sm);
    padding: 8px 14px;
    border-radius: 8px;
    cursor: pointer;
  }
</style>
