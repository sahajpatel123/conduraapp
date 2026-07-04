<script lang="ts">
  import { onMount } from 'svelte';
  import { ipc } from '../ipc/client';
  import Thread from './Thread.svelte';
  import Pulse from './Pulse.svelte';

  // Condura Skills — local installed skills as a card index. Each card is a
  // procedure the bot can run. Auto-created skills carry a green thread (the
  // agent made these); user-authored skills carry a pollen thread (you made these).

  let skills = $state<{ id: string; name: string; description?: string; author?: string; steps?: string[] }[]>([]);
  let loading = $state(true);
  let loadError = $state<string | null>(null);
  let detail = $state<{ id: string; name: string; description?: string; author?: string; steps?: string[] } | null>(null);

  function isYouAuthored(a: string | undefined): boolean {
    if (!a) return false;
    const s = a.toLowerCase();
    return s === 'you' || s.startsWith('user') || s.includes('human');
  }

  onMount(async () => {
    try {
      const list = (await ipc.skillsList(100)) as unknown as { id: string; name: string; description?: string; author?: string; steps?: string[] }[];
      skills = Array.isArray(list) ? list : [];
      loadError = null;
    } catch (e) {
      skills = [];
      loadError = String(e);
    } finally {
      loading = false;
    }
  });

  async function reload(): Promise<void> {
    loading = true;
    loadError = null;
    try {
      const list = (await ipc.skillsList(100)) as unknown as { id: string; name: string; description?: string; author?: string; steps?: string[] }[];
      skills = Array.isArray(list) ? list : [];
    } catch (e) {
      skills = [];
      loadError = String(e);
    } finally {
      loading = false;
    }
  }

  function runSkill(s: { id: string }): void {
    // For now, surface to the chat via the global event. A real "run" wires to conversation.send.
    console.info('run skill', s.id);
  }
</script>

<div class="skills">
  <header class="head">
    <div class="eyebrow">— What Condura learned</div>
    <h1 class="title">Skills, fanned on a desk.</h1>
    <p class="sub">
      Each card is a procedure Condura can run. Auto-created skills carry a green thread —
      the agent made these. Ones you authored carry a pollen thread.
    </p>
  </header>

  {#if loading}
    <div class="state">
      <Pulse phase="thinking" size={8} />
      <span class="state-label">INDEXING…</span>
    </div>
  {:else if loadError}
    <div class="err-state" role="alert" aria-live="polite">
      <div class="err-row">
        <Pulse phase="error" size={8} />
        <span class="err-head">We couldn't reach the daemon.</span>
      </div>
      <p class="err-sub">{loadError} Condura couldn't load your skills — try again, or check that the daemon is running.</p>
      <div class="err-actions">
        <button class="retry" onclick={() => void reload()}>Try again</button>
      </div>
      <div class="err-hair"></div>
    </div>
  {:else if skills.length === 0}
    <div class="state empty">
      <span class="eh">No skills yet.</span>
      <span class="es">Run a complex task — Condura will save the procedure as a skill automatically.</span>
    </div>
  {:else}
    <div class="deck">
      {#each skills as s (s.id)}
        <button class="card" onclick={() => (detail = s)}>
          <div class="thread" data-author={isYouAuthored(s.author) ? 'you' : 'agent'}>
            <Thread orientation="v" />
          </div>
          <div class="card-body">
            <div class="c-name">{s.name}</div>
            <div class="c-author">{s.author ?? 'Condura'}</div>
            <div class="c-desc">{(s.description ?? '').slice(0, 140)}{(s.description ?? '').length > 140 ? '…' : ''}</div>
            <div class="c-foot">
              <span class="c-run">Run →</span>
              <span class="c-improve">Improve</span>
            </div>
          </div>
        </button>
      {/each}
    </div>
  {/if}

  {#if detail}
    <div class="overlay" onclick={() => (detail = null)}></div>
    <aside class="sheet" role="dialog" aria-modal="true">
      <button class="s-close" onclick={() => (detail = null)} aria-label="Close">×</button>
      <div class="s-eyebrow">{isYouAuthored(detail.author) ? '— You authored this' : '— Condura authored this'}</div>
      <h2 class="s-title">{detail.name}</h2>
      <p class="s-desc">{detail.description ?? 'No description.'}</p>
      {#if detail.steps?.length}
        <div class="s-steps">
          {#each detail.steps as step, i (i)}
            <div class="s-step"><span class="s-step-n">{i + 1}</span><span>{step}</span></div>
          {/each}
        </div>
      {/if}
      <div class="s-actions">
        <button class="s-run" onclick={() => runSkill(detail)}>Run →</button>
        <button class="s-improve">Improve</button>
      </div>
    </aside>
  {/if}
</div>

<style>
  .skills {
    max-width: 980px;
    padding-top: var(--space-7);
  }
  .head {
    margin-bottom: var(--space-6);
  }
  .eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .title {
    font-family: var(--font-display);
    font-size: clamp(28px, 3vw, 40px);
    line-height: 1.08;
    letter-spacing: -0.03em;
    color: var(--content);
    margin: var(--space-3) 0 var(--space-2);
  }
  .sub {
    font-size: 16px;
    line-height: 1.55;
    color: var(--content-soft);
    max-width: 56ch;
  }

  .state {
    display: flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
    padding: var(--space-7) 0;
  }
  .state-label {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .state.empty {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-2);
    text-transform: none;
    letter-spacing: normal;
    font-size: 14px;
    font-family: var(--font-sans);
  }
  .eh {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 22px;
    color: var(--content);
  }
  .es {
    color: var(--content-mute);
  }

  /* error state — same pattern as Chat / Hub */
  .err-state {
    max-width: 520px;
    margin: var(--space-7) 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .err-row {
    display: inline-flex;
    align-items: center;
    gap: 10px;
  }
  .err-head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 22px;
    line-height: 1.15;
    color: var(--content);
    letter-spacing: -0.01em;
  }
  .err-sub {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    line-height: 1.55;
    color: var(--content-faint);
    max-width: 48ch;
  }
  .err-actions {
    margin-top: var(--space-2);
  }
  .retry {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--synapse);
    background: none;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    padding: 6px 14px;
    cursor: pointer;
    transition: color var(--dur) var(--ease), border-color var(--dur) var(--ease);
  }
  .retry:hover {
    color: var(--content);
    border-color: var(--synapse);
  }
  .err-hair {
    height: 1px;
    width: 100%;
    background: linear-gradient(90deg, var(--hair-strong) 0%, var(--hair-strong) 60%, transparent 100%);
    transform: scaleX(0);
    transform-origin: left;
    animation: err-hair-draw 600ms var(--ease) 120ms forwards;
  }
  @keyframes err-hair-draw {
    to { transform: scaleX(1); }
  }
  @media (prefers-reduced-motion: reduce) {
    .err-hair {
      transform: scaleX(1);
      animation: none;
    }
  }

  .deck {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: var(--space-4);
    perspective: 1000px;
  }
  .card {
    position: relative;
    text-align: left;
    padding: 0;
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
    transform-style: preserve-3d;
    transition: transform var(--dur-slow) var(--ease), border-color var(--dur) var(--ease), box-shadow var(--dur) var(--ease);
    box-shadow: 0 2px 4px -2px color-mix(in oklab, var(--ink) 10%, transparent);
    overflow: hidden;
  }
  .card:hover {
    transform: translateY(-4px) rotateX(2deg);
    border-color: var(--hair-strong);
    box-shadow: 0 12px 28px -12px color-mix(in oklab, var(--ink) 25%, transparent);
  }
  .card:active {
    transform: translateY(-2px) rotateX(1deg) scale(0.98);
  }
  .card:focus-visible {
    outline: none;
    border-color: var(--synapse);
    box-shadow: 0 12px 28px -12px color-mix(in oklab, var(--ink) 25%, transparent), 0 0 0 4px var(--pollen-halo);
  }
  .thread {
    position: absolute;
    left: 12px;
    top: 0;
    bottom: 0;
    width: 2px;
    z-index: 1;
  }
  .thread[data-author='you'] :global(.condura-thread .line) {
    stroke: var(--pollen) !important;
  }
  .thread[data-author='agent'] :global(.condura-thread .line) {
    stroke: var(--synapse) !important;
  }
  .card-body {
    padding: var(--space-4) var(--space-4) var(--space-4) 28px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-height: 160px;
  }
  .c-name {
    font-family: var(--font-display);
    font-size: 18px;
    line-height: 1.15;
    color: var(--content);
  }
  .c-author {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .c-desc {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 13px;
    line-height: 1.5;
    color: var(--content-soft);
    margin: 4px 0 auto;
  }
  .c-foot {
    display: flex;
    gap: var(--space-3);
    margin-top: var(--space-3);
  }
  .c-run,
  .c-improve {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-mute);
  }
  .c-run {
    color: var(--synapse);
  }
  .c-improve {
    color: var(--pollen);
  }

  .overlay {
    position: fixed;
    inset: 0;
    background: color-mix(in oklab, var(--ink) 32%, transparent);
    backdrop-filter: blur(4px);
    z-index: var(--z-modal);
  }
  .sheet {
    position: fixed;
    right: 0;
    top: 0;
    bottom: 0;
    width: min(460px, 96vw);
    background: var(--surface);
    border-left: 1px solid var(--hair-strong);
    box-shadow: var(--shadow-float);
    z-index: var(--z-modal);
    padding: var(--space-7) var(--space-7) var(--space-5);
    overflow-y: auto;
    animation: slide-in var(--dur-slow) var(--ease);
  }
  @keyframes slide-in {
    from { transform: translateX(20px); opacity: 0; }
    to { transform: translateX(0); opacity: 1; }
  }
  .s-close {
    position: absolute;
    top: var(--space-4);
    right: var(--space-4);
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: var(--surface-card);
    border: 1px solid var(--hair);
    color: var(--content-mute);
    font-size: 20px;
    cursor: pointer;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .s-close:hover {
    color: var(--content);
    background: var(--paper-2);
    border-color: var(--hair-strong);
    transform: scale(1.04);
  }
  .s-close:active {
    transform: scale(0.94);
  }
  .s-close:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .s-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--synapse);
    margin-bottom: var(--space-3);
  }
  .s-title {
    font-family: var(--font-display);
    font-size: 28px;
    line-height: 1.08;
    letter-spacing: -0.03em;
    color: var(--content);
    margin-bottom: var(--space-4);
  }
  .s-desc {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    line-height: 1.55;
    color: var(--content-soft);
    margin-bottom: var(--space-5);
  }
  .s-steps {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-6);
  }
  .s-step {
    display: flex;
    gap: var(--space-3);
    font-size: 13px;
    color: var(--content);
  }
  .s-step-n {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--synapse);
    flex: none;
  }
  .s-actions {
    display: flex;
    gap: var(--space-3);
  }
  .s-run,
  .s-improve {
    font-family: var(--font-sans);
    font-size: 14px;
    font-weight: 500;
    padding: 11px 20px;
    border-radius: var(--r-pill);
    border: 1px solid var(--hair-strong);
    cursor: pointer;
    transition:
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      color var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .s-run:hover {
    transform: translateY(-1px);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--synapse-glow) 18%, transparent);
  }
  .s-run:active {
    transform: scale(0.97);
  }
  .s-run:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .s-run {
    background: var(--synapse);
    color: var(--paper);
    border-color: var(--synapse);
  }
  :global([data-mode='dark']) .s-run {
    color: var(--ink);
  }
  .s-improve {
    background: transparent;
    color: var(--pollen);
    border-color: var(--pollen);
  }
  .s-improve:hover {
    background: color-mix(in oklab, var(--pollen) 8%, transparent);
    transform: translateY(-1px);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .s-improve:active {
    transform: scale(0.97);
  }
  .s-improve:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  @media (prefers-reduced-motion: reduce) {
    .card,
    .card:hover {
      transform: none;
    }
    .sheet {
      animation: none;
    }
  }
</style>