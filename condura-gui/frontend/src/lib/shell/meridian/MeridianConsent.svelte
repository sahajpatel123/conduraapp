<script lang="ts">
  import { consent } from '../../stores/consent.svelte'

  const ticket = $derived(consent.ticket)

  let allowBtn = $state<HTMLButtonElement | null>(null)
  let denyBtn = $state<HTMLButtonElement | null>(null)

  $effect(() => {
    if (!ticket) return
    queueMicrotask(() => allowBtn?.focus())
  })

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
    <p class="kicker">Consent · {ticket.action_kind}</p>
    <h2>{ticket.detail || 'Condura wants to act'}</h2>
    <p class="detail" id="md-consent-detail">
      Actor: {ticket.actor}. Allow only if you understand the action.
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
    background: var(--md-surface);
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 28%, var(--md-line-strong));
    border-radius: 24px;
    padding: 28px;
    z-index: 91;
    box-shadow: var(--md-shadow-lift), 0 0 0 1px color-mix(in oklab, var(--md-cobalt) 8%, transparent);
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
  h2 {
    font-family: var(--md-font-display);
    font-size: 26px;
    letter-spacing: -0.04em;
    margin: 0 0 12px;
    line-height: 1.15;
  }
  .detail {
    margin: 0 0 22px;
    font-size: 14px;
    color: var(--md-ink-mute);
    line-height: 1.55;
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
  @media (max-width: 420px) {
    .sheet {
      padding: 22px 18px;
      border-radius: 20px;
    }
    h2 {
      font-size: 22px;
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
