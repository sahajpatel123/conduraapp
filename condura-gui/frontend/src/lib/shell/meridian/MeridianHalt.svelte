<script lang="ts">
  /**
   * Kill-switch sheet — honest sticky-resume (T3b).
   * Resume is not in-process: mint a ticket, confirm in a terminal.
   */
  import { onMount } from 'svelte'
  import { halt } from '../../stores/halt.svelte'
  import { t } from '../../i18n'

  let primaryBtn = $state<HTMLButtonElement | null>(null)
  let secondaryBtn = $state<HTMLButtonElement | null>(null)

  const reason = $derived(halt.state.reason || 'User or system kill-switch')
  const ticket = $derived(halt.ticket)
  const confirmVia = $derived(
    (() => {
      if (!ticket) return ''
      const via = (halt.confirmVia || '').replace(/<ticket>/g, ticket).trim()
      if (via && !via.includes('<ticket>')) return via
      return `condura resume confirm --ticket ${ticket}`
    })()
  )

  onMount(() => {
    queueMicrotask(() => primaryBtn?.focus())
    // Two-button Tab trap (mirrors MeridianConsent). When the ticket is
    // minted the sheet has Copy command + New ticket — Shift+Tab from
    // either button needs to bounce to the OTHER button, not escape to
    // shell buttons behind the scrim. Capture phase so we win against
    // shell keydown handlers.
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return
      e.preventDefault()
      const active = document.activeElement
      if (e.shiftKey) {
        if (active === primaryBtn && secondaryBtn) secondaryBtn.focus()
        else primaryBtn?.focus()
      } else {
        if (active === secondaryBtn) primaryBtn?.focus()
        else if (secondaryBtn) secondaryBtn.focus()
        else primaryBtn?.focus()
      }
    }
    window.addEventListener('keydown', onKey, true)
    return () => window.removeEventListener('keydown', onKey, true)
  })
</script>

<div class="back" role="presentation"></div>
<div
  class="sheet"
  role="alertdialog"
  aria-modal="true"
  aria-label={t('halt.halted')}
  aria-describedby="md-halt-detail"
>
  <span class="pulse" aria-hidden="true"></span>
  <p class="kicker">Halt engaged</p>
  <h2>Condura is stopped.</h2>
  <p class="detail" id="md-halt-detail">
    Nothing will run until a human confirms resume outside this app. Your audit trail is intact.
  </p>
  <p class="reason" title={reason}>{reason}</p>

  {#if ticket}
    <div class="ticket-block">
      <p class="cite">resume ticket</p>
      <code class="ticket" translate="no" aria-label={`Resume ticket ${ticket}`}>
        {ticket}
      </code>
      <p class="cli">
        In a terminal:
        <code class="cmd" translate="no">{confirmVia}</code>
      </p>
      <div class="row">
        <button type="button" class="md-btn md-btn-primary" bind:this={primaryBtn} onclick={() => void halt.copyTicket()}>
          Copy command
        </button>
        <button
          type="button"
          class="md-btn md-btn-ghost"
          bind:this={secondaryBtn}
          disabled={halt.ticketBusy}
          onclick={() => void halt.resume()}
        >
          {halt.ticketBusy ? 'Minting…' : 'New ticket'}
        </button>
      </div>
      {#if halt.copyNote}
        <p class="note" aria-live="polite">{halt.copyNote}</p>
      {/if}
    </div>
  {:else}
    <button
      type="button"
      class="md-btn md-btn-primary resume"
      bind:this={primaryBtn}
      disabled={halt.ticketBusy}
      onclick={() => void halt.resume()}
    >
      {halt.ticketBusy ? 'Minting…' : 'Request resume ticket'}
    </button>
    <p class="hint">
      Opens a one-time ticket. Confirm with
      <code class="cmd">condura resume confirm --ticket …</code>
      in a terminal — the app cannot un-halt itself.
    </p>
  {/if}
</div>

<style>
  .back {
    position: fixed;
    inset: 0;
    z-index: 95;
    background: var(--md-scrim-strong);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    animation: md-fade 240ms var(--md-ease) both;
  }
  .sheet {
    position: fixed;
    left: 50%;
    top: 50%;
    width: min(420px, calc(100vw - 32px));
    background: var(--md-surface);
    border: 1px solid color-mix(in oklab, var(--md-halt) 28%, var(--md-line));
    border-radius: 12px;
    padding: 24px 22px;
    z-index: 96;
    text-align: center;
    box-shadow: none;
    animation: md-pop 280ms var(--md-ease) both;
    transform: translate(-50%, -50%);
  }
  .pulse {
    display: block;
    width: 12px;
    height: 12px;
    margin: 0 auto 14px;
    border-radius: 50%;
    background: var(--md-halt);
    animation: md-halt-pulse 1.6s var(--md-ease) infinite;
  }
  .kicker {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-halt);
    margin: 0 0 10px;
  }
  h2 {
    font-family: var(--md-font-display);
    font-size: 28px;
    letter-spacing: -0.04em;
    margin: 0 0 10px;
  }
  .detail {
    margin: 0 0 12px;
    color: var(--md-ink-mute);
    font-size: 14px;
    line-height: 1.5;
  }
  .reason {
    margin: 0 0 18px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-ink-faint);
    line-height: 1.4;
    word-break: break-word;
  }
  .ticket-block {
    text-align: left;
    padding: 12px 12px 11px;
    border-radius: 10px;
    border: 1px solid var(--md-line);
    background: var(--md-stage);
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 8px;
  }
  .ticket {
    display: block;
    font-family: var(--md-font-mono);
    font-size: 12px;
    word-break: break-all;
    color: var(--md-ink);
    margin: 0 0 12px;
    line-height: 1.4;
  }
  .cli {
    margin: 0 0 14px;
    font-size: 13px;
    color: var(--md-ink-mute);
    line-height: 1.45;
  }
  .cmd {
    display: block;
    margin-top: 6px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    word-break: break-all;
    color: var(--md-cobalt);
    line-height: 1.4;
  }
  .row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .note {
    margin: 10px 0 0;
    font-size: 12px;
    color: var(--md-live);
  }
  .resume {
    min-width: 160px;
  }
  .hint {
    margin: 12px 0 0;
    font-size: 12px;
    color: var(--md-ink-faint);
    line-height: 1.4;
  }
  @keyframes md-halt-pulse {
    0% {
      box-shadow: 0 0 0 0 color-mix(in oklab, var(--md-halt) 45%, transparent);
    }
    70% {
      box-shadow: 0 0 0 14px transparent;
    }
    100% {
      box-shadow: 0 0 0 0 transparent;
    }
  }
  @media (max-width: 420px) {
    .sheet {
      padding: 24px 20px;
      border-radius: 14px;
    }
    h2 {
      font-size: 24px;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .back,
    .sheet,
    .pulse {
      animation: none !important;
    }
  }
</style>
