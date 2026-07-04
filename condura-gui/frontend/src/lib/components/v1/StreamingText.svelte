<!--
  StreamingText — text that reveals word-by-word.

  Per spec §11.1 + the v2 vision: the agent's voice should feel like a
  real conversation, not text appearing all at once. When streaming, each
  word rises into view with a refined spring motion — the "drying" pattern
  from the website's WordReveal, adapted for in-place streaming.

  Per motion agent §6: tokens don't character-fade (reads as broken).
  Instead, each word is masked and the inner word rises. The mask is
  a 1.4em-tall clip; the word transforms y: 110% → 0%.

  Per motion agent §6 (mid-stream pause): heartbeat that scales with pause
  duration. 0-600ms: nothing. 600ms-2s: 1.2Hz border breathe. 2-6s: dot
  pulse at 1.5s. 6s+: "still working on this" text.

  Props:
    text         — the current text to display
    voice        — 'serif' (agent, default) | 'sans' (user, system)
    state        — 'streaming' | 'paused' | 'done' | 'error'
    pausedMs     — current pause duration in ms
-->
<script lang="ts">
  interface Props {
    text: string;
    voice?: 'serif' | 'sans';
    state?: 'streaming' | 'paused' | 'done' | 'error';
    pausedMs?: number;
  }

  let { text, voice = 'serif', state = 'streaming', pausedMs = 0 }: Props = $props();

  // Per motion agent §6 — heartbeat that scales with pause duration
  let heartbeatPeriod = $derived(
    pausedMs < 600 ? 0 :
    pausedMs < 2000 ? 833 :
    pausedMs < 6000 ? 1500 :
    0
  );

  // Tokenize the text into words for the word-by-word reveal.
  // Splits on whitespace, preserves the original spacing pattern.
  let tokens = $derived.by(() => {
    if (!text) return [] as Array<{ text: string; space: string }>;
    const result: Array<{ text: string; space: string }> = [];
    const regex = /(\s+|[^\s]+)/g;
    let match;
    let lastEnd = 0;
    while ((match = regex.exec(text)) !== null) {
      if (match[1] === undefined) continue;
      const isWhitespace = /^\s+$/.test(match[1]);
      if (isWhitespace) {
        // Attach the whitespace to the previous token (or push as a free space)
        if (result.length > 0) {
          result[result.length - 1].space = match[1];
        } else {
          result.push({ text: '', space: match[1] });
        }
      } else {
        result.push({ text: match[1], space: '' });
      }
      lastEnd = match.index + match[1].length;
    }
    return result;
  });

  // While streaming, the last word is "new" — it animates in. Previous
  // words are static. After "done", everything is static.
  let newWordIndex = $derived(state === 'done' ? -1 : tokens.length - 1);
</script>

<div
  class="streaming streaming--{voice} streaming--{state}"
  data-state={state}
  aria-live={state === 'streaming' ? 'polite' : 'off'}
>
  <span class="streaming__text">
    {#each tokens as token, i}
      {#if token.text}
        <span class="streaming__word-mask">
          <span
            class="streaming__word"
            class:streaming__word--new={i === newWordIndex && state === 'streaming'}
            class:streaming__word--appearing={i <= newWordIndex}
          >{token.text}</span>
        </span>{token.space}
      {:else}
        {token.space}
      {/if}
    {/each}
  </span>

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
    display: block;
    position: relative;
    line-height: 1.65;
  }

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
    display: inline;
  }

  /* ── Word-by-word reveal ────────────────────────────────── */

  /* The mask — clips the word to its line-height, hides the rise */
  .streaming__word-mask {
    display: inline-block;
    overflow: hidden;
    vertical-align: bottom;
    /* The mask is slightly taller than the line so the descender of the
       word has room to rise without being clipped at the top */
    padding-bottom: 0.1em;
    margin-bottom: -0.1em;
    line-height: 1.65;
  }

  /* The word itself — starts translated down (hidden) and rises */
  .streaming__word {
    display: inline-block;
    transform: translateY(0);
    transition: none;
  }

  /* Words that have NOT yet appeared are hidden below the mask.
     They translate down so they look like they're rising from the line. */
  .streaming__word:not(.streaming__word--appearing) {
    transform: translateY(115%);
  }

  /* When a word becomes "new" (just appeared), it animates up. */
  .streaming__word--new {
    animation: word-rise 520ms cubic-bezier(0.16, 1, 0.3, 1) both;
  }

  @keyframes word-rise {
    from {
      transform: translateY(115%);
      opacity: 0.6;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }

  /* The breathing border on the left, per spec §11.1. While streaming,
     it grows to 2px (vs 1px idle) and oscillates at 0.6Hz. */
  .streaming {
    border-left: 1px solid transparent;
    padding-left: var(--space-3);
    margin-left: calc(-1 * var(--space-3));
    transition: border-left-width var(--duration-base) var(--ease-decelerate);
  }

  .streaming--streaming,
  .streaming--paused {
    border-left: 2px solid var(--content-accent);
    animation: border-breathe 1.8s ease-in-out infinite;
  }

  @keyframes border-breathe {
    0%, 100% {
      border-left-color: var(--plum-500);
    }
    50% {
      border-left-color: var(--plum-700);
    }
  }

  /* The heartbeat dot */
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

  /* "Still working on this" */
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
    .streaming--streaming,
    .streaming--paused {
      animation: none;
      border-left-color: var(--content-accent);
    }
    .streaming__word--new {
      animation: none;
    }
    .streaming__word {
      transform: none;
    }
  }
</style>