<script lang="ts">
  // Condura QuickPromptOverlay — the hero interaction (spec §19.2).
  // Press the global hotkey → this floating paper card summons at top-center,
  // a synapse thread draws across its top edge (the overlay arriving), a
  // breathing Pulse keeps the agent's heartbeat, and a magnetic pollen Send
  // button carries the human spark. Ask anything → it streams a reply.
  //
  // Motion verbs (every motion means something):
  //   thread-draw  = the overlay arriving
  //   pulse        = the agent's heartbeat (idle → thinking)
  //   pollen halo  = the human spark on the CTA
  //   yield line   = Condura is writing (the user yields)
  // No decorative loops.

  import { onMount } from 'svelte';
  import { conversation } from '../stores/conversation.svelte';
  import { settings } from '../stores/settings.svelte';
  import { ipc } from '../ipc/client';
  import type { ProviderInfo } from '../ipc/types';
  import Pulse from './Pulse.svelte';
  import Button from './Button.svelte';
  import Glyph from './Glyph.svelte';
  import Thread from './Thread.svelte';

  let {
    open = false,
    onclose = () => {},
  }: { open: boolean; onclose: () => void } = $props();

  // ── reduced motion (skip slide + thread-draw; just show) ──
  const reduceMotion =
    typeof matchMedia !== 'undefined' && matchMedia('(prefers-reduced-motion: reduce)').matches;

  // ── behavioral constants (not design tokens) ──
  const IDLE_MS = 5000; // spec: auto-dismiss after 5s idle
  const TA_MAX_LINES = 6; // grow cap before internal scroll

  // ── state ──
  let inputText = $state('');
  let providers = $state<ProviderInfo[]>([]);
  let selectedModel = $state('');
  let listening = $state(false); // purely visual mic affordance
  let draw = $state(false); // top-edge thread-draw gesture
  let textareaEl = $state<HTMLTextAreaElement | undefined>(undefined);
  let idleTimer: ReturnType<typeof setTimeout> | null = null;

  // ── provider / model selection (mirrors Chat.svelte's defaultProviderModel) ──
  onMount(async () => {
    try {
      providers = await ipc.providersList();
    } catch {
      providers = [];
    }
    if (settings.config?.llm?.providers) {
      const enabled = Object.entries(settings.config.llm.providers).find(([, p]) => p.enabled);
      if (enabled) selectedModel = `${enabled[0]}:${enabled[1].default_model ?? ''}`;
    }
    if (!selectedModel && providers[0]?.models[0]) {
      selectedModel = `${providers[0].name}:${providers[0].models[0].id}`;
    }
  });

  let modelOptions = $derived(
    providers.flatMap((p) =>
      p.models.map((m) => ({ value: `${p.name}:${m.id}`, label: `${p.name} · ${m.id}` }))
    )
  );

  // wake chip — reflects the configured overlay hotkey (locked decision #8)
  let wakeLabel = $derived.by(() => {
    const hk = settings.config?.hotkey?.overlay;
    return hk && hk.length > 0 ? `${hk} Wake` : '⌥⌥ Wake';
  });

  // ── auto-grow textarea (token-capped via CSS max-height) ──
  $effect(() => {
    void inputText;
    const el = textareaEl;
    if (!el) return;
    el.style.height = 'auto';
    el.style.height = el.scrollHeight + 'px';
  });

  // ── lifecycle: focus, re-arm thread-draw, arm idle, reset draft on each open ──
  $effect(() => {
    if (!open) {
      if (idleTimer) {
        clearTimeout(idleTimer);
        idleTimer = null;
      }
      draw = false;
      listening = false;
      return;
    }
    // summon
    inputText = '';
    requestAnimationFrame(() => textareaEl?.focus());
    if (reduceMotion) {
      draw = true;
    } else {
      draw = false;
      requestAnimationFrame(() => requestAnimationFrame(() => (draw = true)));
    }
    armIdle();
    return () => {
      if (idleTimer) {
        clearTimeout(idleTimer);
        idleTimer = null;
      }
    };
  });

  // ── global Esc dismiss (robust even if focus leaves the card) ──
  $effect(() => {
    if (!open) return;
    const onWinKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onclose();
      }
    };
    window.addEventListener('keydown', onWinKey, true);
    return () => window.removeEventListener('keydown', onWinKey, true);
  });

  // ── idle auto-dismiss (resets on any activity over the card) ──
  function armIdle(): void {
    if (idleTimer) clearTimeout(idleTimer);
    idleTimer = setTimeout(() => {
      // never dismiss mid-stream or with an unsent draft
      if (conversation.isStreaming || inputText.trim().length > 0) {
        armIdle();
        return;
      }
      onclose();
    }, IDLE_MS);
  }

  function markActivity(): void {
    if (open) armIdle();
  }

  // ── send / cancel ──
  async function send(): Promise<void> {
    const text = inputText.trim();
    if (!text || conversation.isStreaming) return;
    inputText = '';
    const [providerName, modelId] = selectedModel.split(':');
    if (!providerName || !modelId) return;
    try {
      await conversation.send(providerName, modelId, text);
    } catch (e) {
      console.error('quick-prompt send failed', e);
    }
  }

  async function cancel(): Promise<void> {
    try {
      await conversation.cancel();
    } catch (e) {
      console.error('quick-prompt cancel failed', e);
    }
  }

  // ── keyboard: ⌘↵ or Enter sends, Shift+Enter newlines, Esc closes ──
  function onKeydown(e: KeyboardEvent): void {
    markActivity();
    if (e.key === 'Escape') {
      e.preventDefault();
      onclose();
      return;
    }
    if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
      e.preventDefault();
      void send();
    }
  }

  // keep the textarea from growing unbounded
  let taMaxHeight = $derived(`calc(var(--space-5) * ${TA_MAX_LINES})`);
</script>

{#if open}
  <div
    class="qp-card"
    class:enter={!reduceMotion}
    class:streaming={conversation.isStreaming}
    role="dialog"
    aria-label="Quick prompt"
    aria-modal="false"
    tabindex="-1"
    onkeydown={onKeydown}
    onpointermove={markActivity}
    onpointerdown={markActivity}
  >
    <div class="paper-grain" aria-hidden="true"></div>
    <div class="qp-thread" aria-hidden="true">
      <Thread orientation="h" {draw} />
    </div>

    <div class="qp-body">
      <header class="qp-head">
        <Pulse phase={conversation.isStreaming ? 'thinking' : 'idle'} size={8} />
        <span class="qp-chip qp-wake">{wakeLabel}</span>
        {#if modelOptions.length > 0}
          <label class="qp-chip qp-model" title="Model">
            <select bind:value={selectedModel} aria-label="Model">
              {#each modelOptions as opt (opt.value)}
                <option value={opt.value}>{opt.label}</option>
              {/each}
            </select>
          </label>
        {:else}
          <span class="qp-chip qp-mute">no model</span>
        {/if}
        <button class="qp-close" onclick={onclose} aria-label="Close" title="Esc to close">
          <Glyph name="close" size={14} />
        </button>
      </header>

      {#if conversation.streamingError}
        <div class="qp-error" role="alert" aria-live="polite">
          <div class="qp-err-row">
            <Pulse phase="error" size={8} />
            <span class="qp-err-head">The daemon dropped the thread.</span>
          </div>
          <p class="qp-err-sub">{conversation.streamingError} Press your hotkey again, or check the daemon.</p>
          <div class="qp-err-hair"></div>
        </div>
      {:else if conversation.isStreaming}
        <div class="qp-yield" aria-live="polite">
          <Pulse phase="thinking" size={10} />
          <span class="qp-yield-text">Condura is writing…</span>
          <button
            type="button"
            class="qp-stop"
            onclick={cancel}
            aria-label="Stop generation"
          >
            <Glyph name="stop" size={12} />
            <span>Stop</span>
          </button>
        </div>
      {:else}
        <textarea
          bind:value={inputText}
          bind:this={textareaEl}
          oninput={markActivity}
          rows="1"
          class="qp-input"
          placeholder="Say something…"
          aria-label="Quick prompt input"
          style:--ta-max="{taMaxHeight}"
        ></textarea>
      {/if}

      <footer class="qp-foot">
        <button
          type="button"
          class="qp-mic"
          class:on={listening}
          onclick={() => (listening = !listening)}
          aria-label="Voice input"
          aria-pressed={listening}
          title={listening ? 'Listening (visual)' : 'Voice input'}
        >
          <span class="qp-mic-dot" aria-hidden="true"></span>
          <Glyph name="mic" size={16} class="qp-mic-glyph" />
        </button>
        <span class="qp-cta">
          {#if conversation.isStreaming}
            <Button variant="primary" onclick={cancel}>
              <Glyph name="stop" size={14} /> Stop
            </Button>
          {:else}
            <Button variant="primary" magnetic onclick={send}>
              Send <span class="arrow">↗</span>
            </Button>
          {/if}
        </span>
      </footer>

      {#if !conversation.isStreaming}
        <div class="qp-hint">⌘↵ to send · Esc to close · idle dismiss in 5s</div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .qp-card {
    position: fixed;
    top: var(--space-9);
    left: 50%;
    transform: translateX(-50%);
    width: min(32.5rem, 92vw);
    z-index: var(--z-sheet);
    background: var(--surface);
    border: 1px solid var(--hair);
    border-radius: var(--r-lg);
    box-shadow: var(--shadow-float);
    overflow: hidden;
    color: var(--content);
  }
  /* the warm pollen halo underneath — a quiet, ambient warmth */
  .qp-card::before {
    content: '';
    position: absolute;
    inset: -8px;
    z-index: -1;
    border-radius: calc(var(--r-lg) + 4px);
    box-shadow: 0 0 24px color-mix(in oklab, var(--pollen) 18%, transparent);
    pointer-events: none;
  }
  .qp-card.enter {
    animation: qp-enter 200ms var(--ease) both;
  }
  /* when streaming, the composer collapses: the textarea swaps out and the
     yield bar takes its slot; we drive the collapse with a height transition
     on the body via class:streaming. */
  .qp-card.streaming .qp-body {
    gap: 0;
  }

  @keyframes qp-enter {
    from {
      opacity: 0;
      transform: translateX(-50%) translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateX(-50%) translateY(0);
    }
  }

  /* top-edge synapse thread gesture — the overlay arriving */
  .qp-thread {
    position: absolute;
    top: 0;
    left: var(--space-4);
    right: var(--space-4);
    z-index: 3;
    pointer-events: none;
  }

  /* paper grain sits behind content */
  .qp-body {
    position: relative;
    z-index: 2;
    padding: var(--space-5) var(--space-5) var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  /* ── header: heartbeat + context chips + close ── */
  .qp-head {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  .qp-chip {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-mute);
    border: 1px solid var(--hair);
    border-radius: var(--r-pill);
    padding: var(--space-1) var(--space-3);
    background: var(--surface-card);
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    white-space: nowrap;
  }
  .qp-wake {
    color: var(--synapse);
    border-color: var(--hair-strong);
  }
  .qp-model select {
    font: inherit;
    color: inherit;
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 0;
    text-transform: inherit;
    letter-spacing: inherit;
  }
  .qp-model select:focus-visible {
    outline: 2px solid var(--synapse);
    outline-offset: 3px;
    border-radius: var(--r-pill);
  }
  .qp-mute {
    color: var(--content-faint);
  }
  .qp-close {
    margin-left: auto;
    width: var(--space-7);
    height: var(--space-7);
    border-radius: var(--r-pill);
    border: 1px solid transparent;
    color: var(--content-faint);
    display: grid;
    place-items: center;
    transition:
      color var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .qp-close:hover {
    color: var(--content);
    border-color: var(--hair-strong);
    background: var(--surface-card);
    transform: scale(1.06);
  }
  .qp-close:active {
    transform: scale(0.94);
  }
  .qp-close:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  /* ── textarea ── */
  .qp-input {
    width: 100%;
    min-height: calc(var(--space-6) * 2);
    max-height: var(--ta-max, calc(var(--space-5) * 6));
    resize: none;
    border: none;
    background: transparent;
    color: var(--content);
    font-family: var(--font-sans);
    font-size: 16px;
    line-height: 1.5;
    letter-spacing: -0.008em;
    overflow-y: auto;
    transition: box-shadow 240ms var(--ease);
  }
  .qp-input::placeholder {
    font-family: var(--font-display);
    font-style: italic;
    color: var(--content-faint);
    letter-spacing: -0.01em;
  }
  .qp-input:focus {
    outline: none;
    /* center-out synapse hairline: bottom-inner 1px hairline */
    box-shadow: inset 0 -1px 0 var(--synapse);
  }

  /* ── yield gesture: a thin bar that replaces the textarea via {#if}
     when isStreaming. The collapse is driven by class:streaming on .qp-card
     closing the body gap; here the bar carries the user-yield signal. ── */
  .qp-yield {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    height: 40px;
    max-height: 40px;
    padding: var(--space-2) var(--space-3);
    border-radius: var(--r-sm);
    border: 1px solid color-mix(in srgb, var(--pollen) 32%, transparent);
    background: color-mix(in srgb, var(--pollen) 6%, transparent);
    overflow: hidden;
    animation: yield-in 240ms var(--ease) both;
  }
  .qp-yield-text {
    flex: 1;
    font-family: var(--font-display);
    font-style: italic;
    font-size: 14px;
    color: var(--pollen);
    letter-spacing: -0.005em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .qp-stop {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    height: 24px;
    padding: 0 var(--space-3);
    border-radius: var(--r-pill);
    border: 1px solid var(--pollen);
    background: transparent;
    color: var(--pollen);
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    transition:
      background var(--dur) var(--ease),
      color var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .qp-stop:hover {
    background: var(--pollen);
    color: var(--paper);
    transform: translateY(-1px);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .qp-stop:active {
    transform: scale(0.97);
  }
  .qp-stop:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  @keyframes yield-in {
    from {
      max-height: 0;
      opacity: 0;
      transform: translateY(-2px);
    }
    to {
      max-height: 40px;
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* ── footer: mic + send/stop ── */
  .qp-foot {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }
  /* mic: 32px circle with synapse border; a centered 12px pollen dot
     breathes softly; on hover the circle fills to a faint synapse tint. */
  .qp-mic {
    position: relative;
    width: 32px;
    height: 32px;
    border-radius: var(--r-pill);
    border: 1px solid var(--synapse);
    color: var(--content-mute);
    background: transparent;
    display: grid;
    place-items: center;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .qp-mic :global(.qp-mic-glyph) {
    opacity: 0.6;
    transition: opacity var(--dur) var(--ease);
  }
  .qp-mic-dot {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 12px;
    height: 12px;
    margin-top: -6px;
    margin-left: -6px;
    border-radius: 50%;
    background: var(--pollen);
    animation: breathe 1.6s var(--ease) infinite;
    pointer-events: none;
  }
  .qp-mic:hover {
    color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 8%, transparent);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--synapse) 12%, transparent);
    transform: translateY(-1px);
  }
  .qp-mic:active {
    transform: scale(0.94);
  }
  .qp-mic:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .qp-mic:hover :global(.qp-mic-glyph) {
    opacity: 0;
  }
  .qp-mic.on {
    color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 12%, transparent);
  }
  .qp-mic.on :global(.qp-mic-glyph) {
    opacity: 0;
  }

  /* CTA (Send / Stop) sits at the right edge; the wrapper is scoped so
     layout + the arrow hover apply without relying on Button's root. */
  .qp-cta {
    margin-left: auto;
    display: inline-flex;
  }
  .arrow {
    transition: transform var(--dur) var(--ease);
  }
  .qp-cta:hover .arrow {
    transform: translate(2px, -2px);
  }

  /* ── hint ── */
  .qp-hint {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-faint);
    text-align: center;
  }

  /* ── error state — mid-stream daemon failure inside the overlay ── */
  .qp-error {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-3);
    border-radius: var(--r-sm);
    border: 1px solid color-mix(in srgb, var(--danger) 32%, transparent);
    background: color-mix(in srgb, var(--danger) 6%, transparent);
  }
  .qp-err-row {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }
  .qp-err-head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 14px;
    line-height: 1.2;
    color: var(--content);
    letter-spacing: -0.01em;
  }
  .qp-err-sub {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 13px;
    line-height: 1.5;
    color: var(--content-faint);
    max-width: 48ch;
    word-break: break-word;
  }
  .qp-err-hair {
    height: 1px;
    width: 100%;
    background: linear-gradient(90deg, var(--danger) 0%, var(--danger) 60%, transparent 100%);
    opacity: 0.45;
    transform: scaleX(0);
    transform-origin: left;
    animation: qp-err-hair-draw 600ms var(--ease) 120ms forwards;
  }
  @keyframes qp-err-hair-draw {
    to { transform: scaleX(1); }
  }
  @media (prefers-reduced-motion: reduce) {
    .qp-err-hair {
      transform: scaleX(1);
      animation: none;
    }
  }

  /* reduced motion: global media query in condura.css already collapses
     animation/transition durations to .01ms; the .enter class is simply
     not applied, so the card shows at its resting state. Disable the
     pollen halo, the breathing mic dot, and the yield slide-in here for
     safety, even though the global rule already neutralises them. */
  @media (prefers-reduced-motion: reduce) {
    .qp-card::before { display: none; }
    .qp-mic-dot { animation: none; }
    .qp-input:focus { box-shadow: none; }
    .qp-yield { animation: none; }
  }
</style>