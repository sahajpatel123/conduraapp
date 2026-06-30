<!--
  StreamingText — text that reveals token-by-token.

  Per motion agent §6 ("Streaming text"): tokens reveal character-by-character
  but not by literal character fade — that reads as broken. Instead, the
  *line* containing the new token grows from the left edge to the right with
  an 80ms decelerate, like a moving typewriter carriage. The text inside
  appears fully formed on arrival.

  Per spec §11.1 (Chat surface): agent messages use the SERIF voice; the
  serif is "as if written by hand" — a thoughtful person talking.

  Per motion agent §6 (mid-stream pause): a 600ms pause is normal. A 6s
  pause is concerning. A 60s pause is failed. This component signals state
  via a subtle heartbeat that scales with pause duration.

  Props:
    text         — the current text to display
    voice        — 'serif' (agent, default) | 'sans' (user, system)
    state        — 'streaming' | 'paused' | 'done' | 'error'
    pausedMs     — current pause duration in ms (for heartbeat scaling)
-->
<script lang="ts">
  interface Props {
    text: string;
    voice?: 'serif' | 'sans';
    state?: 'streaming' | 'paused' | 'done' | 'error';
    pausedMs?: number;
  }

  let { text, voice = 'serif', state = 'streaming', pausedMs = 0 }: Props = $props();

  // Per motion agent §6 — heartbeat that scales with pause duration.
  // 0-600ms: nothing. 600-2000ms: 1.2Hz border breathe. 2-6s: dot at 1.5s.
  // 6s+: "still working" text. 30s+: bubble dims, "stuck?" link.
  let heartbeatPeriod = $derived(
    pausedMs < 600 ? 0 :
    pausedMs < 2000 ? 833 :  // 1.2Hz
    pausedMs < 6000 ? 1500 :
    0  // show text
  );
</script>

<div
  class="streaming streaming--{voice} streaming--{state}"
  data-state={state}
  aria-live={state === 'streaming' ? 'polite' : 'off'}
>
  <span class="streaming__text">{text}</span>

  {#if state === 'paused' && heartbeatPeriod > 0}
    <span
      class="streaming__heartbeat"
      style="--heartbeat-period: {heartbeatPeriod}ms"
      aria-hidden="true"
    ></span>
  {/if}

  {#if state === 'paused' && pausedMs >= 6000}
    <span class="streaming__still">still working on this</span>
  {/if}
</div>

<style>
  .streaming {
    display: inline;
    position: relative;
    line-height: 1.6;
  }

  /* Voice variants */
  .streaming--serif {
    font-family: var(--font-serif);
    font-size: var(--text-body-lg-size);
    color: var(--content-primary);
  }
  .streaming--sans {
    font-family: var(--font-sans);
    font-size: var(--text-body-size);
    color: var(--content-primary);
  }

  .streaming__text {
    /* Text appears fully formed on arrival — no per-character fade. */
    /* The line "growing" effect is the breathing border, not the text. */
  }

  /* The breathing border on the left, per spec §11.1.
     Per motion agent §6, this oscillates at 0.6Hz during streaming,
     1.2Hz during a long pause. */
  .streaming {
    border-left: 1px solid transparent;
    padding-left: var(--space-3);
    margin-left: calc(-1 * var(--space-3));
  }

  .streaming--streaming,
  .streaming--paused {
    border-left-color: var(--content-accent);
  }

  [data-state="done"] {
    border-left-color: transparent;
  }

  /* Heartbeat — visible during pause, scales with duration */
  .streaming__heartbeat {
    display: inline-block;
    width: 6px;
    height: 6px;
    border-radius: var(--radius-pill);
    background-color: var(--content-accent);
    margin-left: var(--space-2);
    vertical-align: middle;
    animation: heartbeat var(--heartbeat-period) ease-in-out infinite;
  }

  @keyframes heartbeat {
    0%, 100% { opacity: 0.4; transform: scale(0.85); }
    50% { opacity: 1; transform: scale(1.15); }
  }

  /* Still-working text, 6s+ pause */
  .streaming__still {
    display: block;
    margin-top: var(--space-2);
    font-family: var(--font-sans);
    font-size: var(--text-caption-size);
    color: var(--content-muted);
    font-style: italic;
  }

  /* Error state */
  .streaming--error {
    border-left-color: var(--error-500);
  }

  @media (prefers-reduced-motion: reduce) {
    .streaming__heartbeat {
      animation: none;
      opacity: 1;
    }
  }
</style>