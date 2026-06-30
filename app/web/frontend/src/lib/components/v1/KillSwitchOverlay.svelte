<!--
  KillSwitchOverlay — the safety contract.

  Per spec §5.3 (Layer 3): the kill switch must work faster than the user
  can blink. Per motion agent §3.10: full-viewport black overlay instant-cuts
  in (no fade — speed matters). A thin red 1px horizontal scan-line draws
  across the screen top-to-bottom in 280ms. All in-flight agent processes
  visibly halt — input desaturates, computer-use cursor rings pulse red.

  This overlay is NEVER reduced, even in Low energy mode or
  prefers-reduced-motion (per spec §6.6).

  Props:
    reason    — 'user' | 'watchdog' | 'anomaly' | 'network'
    detail    — optional explanation string
    onresume  — handler for "Resume"
-->
<script lang="ts">
  import Pulse from './Pulse.svelte';
  import Button from './Button.svelte';

  type Reason = 'user' | 'watchdog' | 'anomaly' | 'network';

  interface Props {
    reason?: Reason;
    detail?: string;
    onresume?: () => void;
  }

  let { reason = 'user', detail, onresume }: Props = $props();

  const REASON_TEXT: Record<Reason, string> = {
    user: 'You pressed the kill switch.',
    watchdog: 'I paused myself — no activity for too long.',
    anomaly: 'I paused myself — that action pattern looked unusual.',
    network: "Network guard tripped — that endpoint wasn't on the allowlist.",
  };
</script>

<div class="kill" role="alertdialog" aria-modal="true" aria-label="Synaptic stopped">
  <!-- The instant black overlay -->
  <div class="kill__black" aria-hidden="true"></div>

  <!-- The red scan-line, drawn top-to-bottom in 280ms -->
  <div class="kill__scan" aria-hidden="true"></div>

  <!-- The reason pill, fades in 120ms after the freeze -->
  <div class="kill__pill">
    <Pulse state="error" size="md" label="Synaptic stopped" />
    <div class="kill__pill-text">
      <div class="kill__pill-headline">Synaptic stopped</div>
      <div class="kill__pill-detail">{detail ?? REASON_TEXT[reason]}</div>
    </div>
  </div>

  <!-- Resume action, fades in last -->
  <div class="kill__resume">
    <Button variant="primary" size="lg" onclick={onresume}>
      Resume agent
    </Button>
    <p class="kill__resume-hint">
      Or press <kbd>⌘⇧Space</kbd> to bring up the command surface and start fresh.
    </p>
  </div>
</div>

<style>
  .kill {
    position: fixed;
    inset: 0;
    z-index: var(--z-max);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-7);
    pointer-events: auto;
  }

  /* Instant black overlay — NO fade (per motion agent §3.10) */
  .kill__black {
    position: absolute;
    inset: 0;
    background-color: #000000;
    /* Intentionally no transition. Instant cut. */
  }

  /* Red scan-line, drawn top-to-bottom in 280ms */
  .kill__scan {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 2px;
    background-color: var(--error-500);
    box-shadow: 0 0 16px var(--error-500);
    animation: scan 280ms linear forwards;
  }

  @keyframes scan {
    from { transform: translateY(0); }
    to { transform: translateY(100vh); }
  }

  /* Reason pill */
  .kill__pill {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-4) var(--space-6);
    background-color: rgba(20, 20, 20, 0.85);
    border: 1px solid var(--error-500);
    border-radius: var(--radius-lg);
    animation: pill-in 240ms var(--ease-decelerate) 120ms both;
    max-width: 480px;
  }

  @keyframes pill-in {
    from {
      opacity: 0;
      transform: translateY(-8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .kill__pill-text {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .kill__pill-headline {
    font-family: var(--font-serif);
    font-size: var(--text-h4-size);
    color: var(--paper-warm-0);
    font-weight: 500;
  }

  .kill__pill-detail {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: rgba(245, 241, 232, 0.7);
    line-height: 1.5;
  }

  /* Resume action */
  .kill__resume {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-3);
    animation: resume-in 240ms var(--ease-decelerate) 240ms both;
  }

  @keyframes resume-in {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .kill__resume-hint {
    margin: 0;
    font-size: var(--text-caption-size);
    color: rgba(245, 241, 232, 0.5);
    font-family: var(--font-mono);
  }

  kbd {
    font-family: var(--font-mono);
    background-color: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.2);
    padding: 1px 6px;
    border-radius: var(--radius-xs);
    color: var(--paper-warm-0);
  }

  /* The kill switch is NEVER reduced. Override reduced-motion. */
  @media (prefers-reduced-motion: reduce) {
    .kill__scan {
      animation-duration: 80ms;
    }
  }
</style>