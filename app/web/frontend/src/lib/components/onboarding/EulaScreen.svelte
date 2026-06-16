<script lang="ts">
  import { onMount } from 'svelte'
  import { ipc } from '../../ipc/client'
  import { onboarding } from '../../stores/onboarding.svelte'
  import type { EULADocument } from '../../ipc/types'

  const EULA_TITLE = 'Synaptic Freeware License'

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
  <h1>Welcome to Synaptic</h1>
  <p class="lede">A free, on-device AI agent. Before we set things up, please review and accept the license.</p>

  {#if loading}
    <p class="muted">Loading the license…</p>
  {:else if loadError}
    <p class="error">Could not load the EULA: {loadError}</p>
  {:else if doc}
    <div class="eula-meta">
      <strong>{EULA_TITLE}</strong>
      <span class="version">{doc.version}{doc.updated_at ? ` · updated ${doc.updated_at}` : ''}</span>
    </div>

    <div
      class="eula-body"
      bind:this={scrollEl}
      onscroll={checkScroll}
      tabindex="0"
      role="document"
      aria-label="End User License Agreement"
    >
      <pre>{doc.text}</pre>
    </div>

    {#if !scrolledToBottom}
      <p class="scroll-cue">Scroll to the bottom to continue ↓</p>
    {/if}

    <label class="checkbox" class:disabled={!scrolledToBottom}>
      <input
        type="checkbox"
        checked={accepted}
        disabled={!scrolledToBottom}
        onchange={(e) => (accepted = (e.target as HTMLInputElement).checked)}
      />
      <span>I have read and accept the {EULA_TITLE}.</span>
    </label>
  {/if}

  {#if onboarding.error}
    <p class="error">{onboarding.error}</p>
  {/if}

  <div class="actions center">
    <button class="btn btn-primary" onclick={accept} disabled={!canAccept}>
      {onboarding.busy ? 'Saving…' : 'I Accept'}
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
  }
  h1 {
    font-size: var(--size-3xl);
    font-weight: 600;
    margin-bottom: var(--space-2);
    background: var(--color-accent-gradient);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .lede {
    font-size: var(--size-md);
    color: var(--color-text-muted);
    margin-bottom: var(--space-4);
  }
  .muted {
    color: var(--color-text-muted);
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
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
    margin-bottom: var(--space-2);
  }
  .eula-body pre {
    white-space: pre-wrap;
    word-wrap: break-word;
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    line-height: 1.6;
    color: var(--color-text-muted);
    margin: 0;
  }
  .scroll-cue {
    color: var(--color-accent);
    font-size: var(--size-sm);
    margin: 0 0 var(--space-3) 0;
    animation: bob 1.5s ease-in-out infinite;
  }
  @keyframes bob {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(2px); }
  }
  .checkbox {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin: var(--space-3) 0;
    cursor: pointer;
    color: var(--color-text);
  }
  .checkbox.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .actions {
    display: flex;
    justify-content: center;
    margin-top: var(--space-4);
  }
  .btn {
    padding: 12px 28px;
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
  .error {
    color: var(--color-error);
    font-size: var(--size-sm);
  }
</style>
