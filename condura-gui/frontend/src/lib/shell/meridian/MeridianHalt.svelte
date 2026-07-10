<script lang="ts">
  import { onMount } from 'svelte'

  interface Props {
    onresume: () => void
  }
  let { onresume }: Props = $props()

  let resumeBtn = $state<HTMLButtonElement | null>(null)

  onMount(() => {
    queueMicrotask(() => resumeBtn?.focus())
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return
      e.preventDefault()
      resumeBtn?.focus()
    }
    window.addEventListener('keydown', onKey, true)
    return () => window.removeEventListener('keydown', onKey, true)
  })
</script>

<div class="back" role="presentation"></div>
<div class="sheet" role="alertdialog" aria-modal="true" aria-label="Agent halted" aria-describedby="md-halt-detail">
  <span class="pulse" aria-hidden="true"></span>
  <p class="kicker">Halt engaged</p>
  <h2>Condura is stopped.</h2>
  <p class="detail" id="md-halt-detail">Nothing will run until you resume. Your audit trail is intact.</p>
  <button
    type="button"
    class="md-btn md-btn-primary resume"
    bind:this={resumeBtn}
    onclick={onresume}
  >
    Resume
  </button>
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
    width: min(400px, calc(100vw - 32px));
    background: var(--md-surface);
    border: 1px solid color-mix(in oklab, var(--md-halt) 32%, var(--md-line-strong));
    border-radius: 24px;
    padding: 30px 28px;
    z-index: 96;
    text-align: center;
    box-shadow: var(--md-shadow-lift);
    animation: md-pop 380ms var(--md-spring) both;
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
    margin: 0 0 22px;
    color: var(--md-ink-mute);
    font-size: 14px;
    line-height: 1.5;
  }
  .resume {
    min-width: 120px;
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
      border-radius: 20px;
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
