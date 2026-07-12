<script lang="ts">
  /**
   * Gatekeeper consent sheet — minimal, never blank.
   * Backend fills actor/detail; we still humanize empty/legacy tickets.
   */
  import { consent } from '../../stores/consent.svelte'

  const ticket = $derived(consent.ticket)

  let allowBtn = $state<HTMLButtonElement | null>(null)
  let denyBtn = $state<HTMLButtonElement | null>(null)

  const kindLabel = $derived(humanKind(ticket?.action_kind))
  const title = $derived(
    (ticket?.detail && ticket.detail.trim()) || kindLabel || 'Condura wants to act'
  )
  const actor = $derived(
    (ticket?.actor && ticket.actor.trim()) || 'Condura agent'
  )
  const blast = $derived(blastHint(ticket?.action_kind))

  $effect(() => {
    if (!ticket) return
    queueMicrotask(() => allowBtn?.focus())
  })

  function humanKind(kind?: string): string {
    if (!kind?.trim()) return 'Action'
    const k = kind.trim()
    const map: Record<string, string> = {
      'shell.exec': 'Shell command',
      'file.write': 'Write file',
      'file.read': 'Read file',
      'computeruse.click': 'Click',
      'computeruse.type': 'Type',
      'computeruse.scroll': 'Scroll',
      'computeruse.launch': 'Launch app',
      'delegation.spawn': 'Spawn sub-agent',
      chat: 'Chat',
    }
    return map[k] || k.replace(/\./g, ' · ')
  }

  function blastHint(kind?: string): string {
    if (!kind) return ''
    if (kind.startsWith('shell') || kind.includes('delete')) return 'High impact'
    if (kind.includes('network') || kind.includes('send')) return 'Leaves this machine'
    if (kind.startsWith('file') || kind.startsWith('computeruse')) return 'Changes your system'
    return ''
  }

  function onKey(e: KeyboardEvent): void {
    if (!ticket) return
    if (e.key === 'Escape') {
      e.preventDefault()
      void consent.deny()
      return
    }
    if (e.key !== 'Tab') return
    e.preventDefault()
    const active = document.activeElement
    if (e.shiftKey) {
      if (active === allowBtn) denyBtn?.focus()
      else allowBtn?.focus()
    } else {
      if (active === denyBtn) allowBtn?.focus()
      else denyBtn?.focus()
    }
  }
</script>

<svelte:window onkeydown={onKey} />
{#if ticket}
  <div class="back" role="presentation"></div>
  <div
    class="sheet"
    role="dialog"
    aria-modal="true"
    aria-label="Consent required"
    aria-describedby="md-consent-detail"
  >
    <p class="kicker">
      Consent
      {#if kindLabel}
        <span class="sep" aria-hidden="true">·</span>
        {kindLabel}
      {/if}
      {#if blast}
        <span class="sep" aria-hidden="true">·</span>
        <span class="blast">{blast}</span>
      {/if}
    </p>
    <h2>{title}</h2>
    <p class="detail" id="md-consent-detail">
      <span class="actor">{actor}</span>
      is requesting this. Allow only if you understand what will run.
    </p>
    <div class="actions">
      <button
        type="button"
        class="md-btn md-btn-danger"
        bind:this={denyBtn}
        onclick={() => void consent.deny()}
      >
        Deny
      </button>
      <button
        type="button"
        class="md-btn md-btn-primary"
        bind:this={allowBtn}
        onclick={() => void consent.approve()}
      >
        Allow
      </button>
    </div>
    {#if consent.error}
      <p class="err" role="alert">{consent.error}</p>
    {/if}
    <p class="esc">Esc denies</p>
  </div>
{/if}

<style>
  .back {
    position: fixed;
    inset: 0;
    z-index: 90;
    background: var(--md-scrim);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    animation: md-fade 220ms var(--md-ease) both;
  }
  .sheet {
    position: fixed;
    left: 50%;
    top: 50%;
    width: min(440px, calc(100vw - 32px));
    transform: translate(-50%, -50%);
    background: color-mix(in oklab, var(--md-surface) 97%, transparent);
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 28%, var(--md-line-strong));
    border-radius: 22px;
    padding: 28px;
    z-index: 91;
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    box-shadow:
      var(--md-shadow-lift),
      0 0 0 1px color-mix(in oklab, var(--md-cobalt) 8%, transparent),
      0 1px 0 color-mix(in oklab, #fff 45%, transparent) inset;
    animation: md-pop 360ms var(--md-spring) both;
  }
  .kicker {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-cobalt);
    margin: 0 0 10px;
  }
  .sep {
    margin: 0 0.35em;
    opacity: 0.55;
  }
  .blast {
    color: var(--md-halt);
  }
  h2 {
    font-family: var(--md-font-display);
    font-size: 24px;
    letter-spacing: -0.04em;
    margin: 0 0 12px;
    line-height: 1.2;
    word-break: break-word;
  }
  .detail {
    margin: 0 0 22px;
    font-size: 14px;
    color: var(--md-ink-mute);
    line-height: 1.55;
  }
  .actor {
    font-weight: 600;
    color: var(--md-ink);
  }
  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }
  .esc {
    margin: 14px 0 0;
    text-align: right;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .err {
    margin: 12px 0 0;
    padding: 10px 12px;
    border-radius: 12px;
    border: 1px solid color-mix(in oklab, var(--md-halt) 35%, var(--md-line));
    background: color-mix(in oklab, var(--md-halt) 10%, transparent);
    color: var(--md-halt);
    font-size: 12.5px;
    line-height: 1.4;
  }
  @media (max-width: 420px) {
    .sheet {
      padding: 22px 18px;
      border-radius: 20px;
    }
    h2 {
      font-size: 20px;
    }
    .actions {
      flex-direction: column-reverse;
    }
    .actions :global(.md-btn) {
      width: 100%;
      justify-content: center;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .back,
    .sheet {
      animation: none !important;
    }
  }
</style>
