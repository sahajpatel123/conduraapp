<script lang="ts">
  import { onMount } from 'svelte';
  import { hub } from '../stores/hub.svelte';
  import Thread from './Thread.svelte';
  import Pulse from './Pulse.svelte';

  // Condura Hub — the public Skills Hub as a 3D bookshelf. Each skill is a
  // slim vertical "spine"; hover tilts it forward; install draws a thread.
  // Honors the contract given to the sub-agent: read the real HubSkillMeta
  // shape, never guess.

  let query = $state('');
  let debouncedQuery = $state('');
  let debounceTimer = 0;
  let detail = $state<typeof hub.results[number] | null>(null);

  function setQuery(v: string): void {
    query = v;
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = window.setTimeout(() => {
      debouncedQuery = v;
      hub.search(v);
    }, 250);
  }

  function trustDot(t: string | undefined): string {
    if (t === 'official') return 'var(--synapse)';
    if (t === 'experimental') return 'var(--hair-strong)';
    return 'var(--pollen)';
  }

  function tagsFor(s: typeof hub.results[number]): string[] {
    // HubSkillMeta tags are usually string[]; fall back to [] if missing.
    const t = (s as unknown as { tags?: string[] }).tags;
    return Array.isArray(t) ? t : [];
  }

  function install(s: typeof hub.results[number]): void {
    hub.install(s.id);
  }

  onMount(() => {
    // The store treats an empty query as "return empty without RPC".
    // The Hub is the curated library; landing on an empty shelf isn't
    // useful. Kick off a broad query so the spines show on first paint.
    hub.search('skill');
  });
</script>

<div class="hub">
  <header class="head">
    <div class="eyebrow">— The library · curated · safety-scanned</div>
    <h1 class="title">Skills for the things you ask.</h1>
    <p class="sub">
      A shelf of procedures — auto-created from your complex tasks, or curated by the
      community. Install draws a thread into your machine.
    </p>
  </header>

  <div class="bar">
    <input
      class="search"
      placeholder="search skills…"
      value={query}
      oninput={(e) => setQuery((e.currentTarget as HTMLInputElement).value)}
    />
    <div class="tags">
      {#each Array.from(new Set(hub.results.flatMap((s) => tagsFor(s)))).slice(0, 6) as t (t)}
        <button class="tag" class:active={query.toLowerCase() === t.toLowerCase()} onclick={() => setQuery(t)}>{t}</button>
      {/each}
    </div>
  </div>

  <div class="shelf-stage">
    {#if hub.loading && hub.results.length === 0}
      <div class="state">
        <Pulse phase="thinking" size={8} />
        <span class="state-label">INDEXING THE SHELF…</span>
      </div>
    {:else if hub.error}
      <div class="err-state" role="alert" aria-live="polite">
        <div class="err-row">
          <Pulse phase="error" size={8} />
          <span class="err-head">We couldn't reach the daemon.</span>
        </div>
        <p class="err-sub">{hub.error} Try again, or check that the daemon is running.</p>
        <div class="err-hair"></div>
      </div>
    {:else if hub.results.length === 0}
      <div class="state empty">
        {#if debouncedQuery.trim() !== ''}
          <span class="empty-head">Nothing on this shelf matches.</span>
          <span class="empty-sub">No skills found for "{debouncedQuery}". Try a different word — or come back later.</span>
        {:else}
          <span class="empty-head">The shelf is quiet.</span>
          <span class="empty-sub">The Hub is empty. New skills land here as the community publishes them.</span>
        {/if}
      </div>
    {:else}
      <div class="shelf" style="--count:{hub.results.length}">
        <div class="shelf-rail"></div>
        <div class="spines">
          {#each hub.results as s (s.id)}
            {@const isInstalled = hub.installed.has(s.id)}
            {@const tags = tagsFor(s)}
            <button class="spine" class:installed={isInstalled} onclick={() => (detail = s)} title={s.name}>
              <div class="spine-face">
                <div class="trust" style:background={trustDot(s.trust_level ?? s.trust)}></div>
                <div class="title-v">{s.name}</div>
                <div class="author-v">{(s.author ?? '').split(' ')[0]}</div>
                <div class="tags-v">{#each tags.slice(0, 2) as t (t)}<span>{t}</span>{/each}</div>
              </div>
              <div class="spine-shelf"></div>
            </button>
          {/each}
        </div>
      </div>
    {/if}
  </div>

  {#if detail}
    <div class="detail-overlay" onclick={() => (detail = null)}></div>
    <aside class="detail-sheet" role="dialog" aria-modal="true">
      <button class="d-close" onclick={() => (detail = null)} aria-label="Close detail">×</button>
      <div class="d-eyebrow">— {detail.trust_level ?? detail.trust ?? 'community'}</div>
      <h2 class="d-title">{detail.name}</h2>
      <div class="d-author">by {detail.author ?? 'anonymous'}</div>
      <p class="d-desc">{detail.description ?? 'No description provided.'}</p>
      {#if detail.version}<div class="d-version mono">v{detail.version}</div>{/if}
      <div class="d-tags">
        {#each tagsFor(detail) as t (t)}<span class="d-tag">{t}</span>{/each}
      </div>
      <div class="d-actions">
        <button
          class="install"
          class:installed={hub.installed.has(detail.id)}
          onclick={() => install(detail)}
        >
          {hub.installed.has(detail.id) ? 'Installed ✓' : 'Install →'}
        </button>
      </div>
    </aside>
  {/if}
</div>

<style>
  .hub {
    max-width: 1100px;
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

  .bar {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-5);
  }
  .search {
    flex: 1;
    max-width: 320px;
    padding: 10px 14px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    background: var(--surface);
    color: var(--content);
    font-size: 14px;
    outline: none;
    transition:
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease),
      transform var(--dur) var(--ease);
  }
  .search:hover {
    background: var(--surface-card);
    transform: translateY(-1px);
  }
  .search:focus-visible {
    border-color: var(--synapse);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .search::placeholder {
    color: var(--content-faint);
    font-family: var(--font-display);
    font-style: italic;
  }
  .tags {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }
  .tag {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    padding: 5px 10px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    background: transparent;
    color: var(--content-mute);
    cursor: pointer;
    transition:
      color var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .tag:hover {
    color: var(--content);
    border-color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
    transform: translateY(-1px);
  }
  .tag:active {
    transform: scale(0.97);
  }
  .tag:focus-visible {
    outline: none;
    border-color: var(--synapse);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .tag.active {
    color: var(--content);
    border-color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 10%, transparent);
  }

  .shelf-stage {
    position: relative;
    min-height: 320px;
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
  .empty-head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 22px;
    color: var(--content);
  }
  .empty-sub {
    color: var(--content-mute);
  }

  /* error state — same instrument-serif headline pattern as Chat.svelte */
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

  .shelf {
    position: relative;
    padding: var(--space-7) 0 var(--space-5);
    perspective: 800px;
    perspective-origin: 50% 60%;
  }
  .shelf-rail {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 12px;
    background: linear-gradient(180deg, transparent, color-mix(in oklab, var(--ink) 12%, transparent));
    border-radius: 2px;
  }
  .spines {
    display: flex;
    gap: 14px;
    align-items: flex-end;
    overflow-x: auto;
    padding-bottom: var(--space-2);
  }
  .spine {
    position: relative;
    flex: 0 0 36px;
    height: 220px;
    cursor: pointer;
    background: none;
    padding: 0;
    transition: transform var(--dur) var(--ease);
    transform-style: preserve-3d;
  }
  .spine:hover {
    transform: translateY(-6px);
  }
  .spine:active {
    transform: translateY(-3px) scale(0.97);
  }
  .spine:focus-visible {
    outline: none;
  }
  .spine:focus-visible .spine-face {
    box-shadow:
      0 4px 12px -6px color-mix(in oklab, var(--ink) 25%, transparent),
      0 0 0 3px var(--pollen-halo);
  }
  .spine-face {
    position: relative;
    width: 36px;
    height: 200px;
    background: linear-gradient(90deg, color-mix(in oklab, var(--ink) 14%, transparent), color-mix(in oklab, var(--ink) 6%, transparent) 50%, color-mix(in oklab, var(--ink) 14%, transparent));
    border-left: 1px solid color-mix(in oklab, var(--ink) 18%, transparent);
    border-right: 1px solid color-mix(in oklab, var(--ink) 18%, transparent);
    padding: var(--space-2) 4px;
    color: var(--content-soft);
    transform: rotateY(-22deg);
    transform-origin: left center;
    transition: transform var(--dur-slow) var(--ease);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    font-size: 9px;
    overflow: hidden;
    box-shadow: 0 4px 12px -6px color-mix(in oklab, var(--ink) 25%, transparent);
  }
  :global([data-mode='dark']) .spine-face {
    background: linear-gradient(90deg, color-mix(in oklab, var(--paper) 8%, transparent), color-mix(in oklab, var(--paper) 2%, transparent) 50%, color-mix(in oklab, var(--paper) 8%, transparent));
    border-left-color: color-mix(in oklab, var(--paper) 18%, transparent);
    border-right-color: color-mix(in oklab, var(--paper) 18%, transparent);
  }
  .spine:hover .spine-face {
    transform: rotateY(-30deg);
  }
  .spine.installed .spine-face {
    background: linear-gradient(90deg, color-mix(in oklab, var(--synapse) 18%, transparent), color-mix(in oklab, var(--synapse) 6%, transparent) 50%, color-mix(in oklab, var(--synapse) 18%, transparent));
  }
  .spine-shelf {
    position: absolute;
    left: -6px;
    right: -6px;
    bottom: 0;
    height: 6px;
    background: color-mix(in oklab, var(--ink) 8%, transparent);
    border-radius: 2px;
    transform: rotateY(0deg);
  }
  :global([data-mode='dark']) .spine-shelf {
    background: color-mix(in oklab, var(--paper) 6%, transparent);
  }
  .trust {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    box-shadow: 0 0 4px currentColor;
    flex: none;
  }
  .title-v {
    writing-mode: vertical-rl;
    transform: rotate(180deg);
    font-family: var(--font-display);
    font-size: 13px;
    line-height: 1.1;
    color: var(--content);
    margin-top: var(--space-2);
    white-space: nowrap;
  }
  .author-v {
    writing-mode: vertical-rl;
    transform: rotate(180deg);
    font-family: var(--font-mono);
    font-size: 8px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-top: auto;
  }
  .tags-v {
    writing-mode: vertical-rl;
    transform: rotate(180deg);
    font-family: var(--font-mono);
    font-size: 7px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--synapse);
    display: flex;
    gap: 6px;
  }
  .tags-v span {
    display: block;
  }

  /* detail sheet */
  .detail-overlay {
    position: fixed;
    inset: 0;
    background: color-mix(in oklab, var(--ink) 32%, transparent);
    backdrop-filter: blur(4px);
    z-index: var(--z-modal);
    animation: fade-in var(--dur) var(--ease);
  }
  .detail-sheet {
    position: fixed;
    right: 0;
    top: 0;
    bottom: 0;
    width: min(440px, 96vw);
    background: var(--surface);
    border-left: 1px solid var(--hair-strong);
    box-shadow: var(--shadow-float);
    z-index: var(--z-modal);
    padding: var(--space-7) var(--space-7) var(--space-5);
    overflow-y: auto;
    animation: slide-in-right var(--dur-slow) var(--ease);
  }
  @keyframes slide-in-right {
    from { transform: translateX(20px); opacity: 0; }
    to { transform: translateX(0); opacity: 1; }
  }
  @keyframes fade-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  .d-close {
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
  .d-close:hover {
    color: var(--content);
    background: var(--paper-2);
    border-color: var(--hair-strong);
    transform: scale(1.04);
  }
  .d-close:active {
    transform: scale(0.94);
  }
  .d-close:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .d-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: var(--pollen);
    margin-bottom: var(--space-3);
  }
  .d-title {
    font-family: var(--font-display);
    font-size: 30px;
    line-height: 1.08;
    letter-spacing: -0.03em;
    color: var(--content);
    margin-bottom: var(--space-1);
  }
  .d-author {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-bottom: var(--space-5);
  }
  .d-desc {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 16px;
    line-height: 1.55;
    color: var(--content-soft);
    margin-bottom: var(--space-5);
  }
  .d-version {
    color: var(--content-faint);
    margin-bottom: var(--space-4);
  }
  .d-version.mono {
    font-family: var(--font-mono);
    font-size: 12px;
    letter-spacing: 0.08em;
  }
  .d-tags {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
    margin-bottom: var(--space-6);
  }
  .d-tag {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    padding: 4px 10px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    color: var(--content-mute);
  }
  .d-actions {
    display: flex;
    gap: var(--space-3);
  }
  .install {
    font-family: var(--font-sans);
    font-size: 14px;
    font-weight: 500;
    padding: 11px 20px;
    border-radius: var(--r-pill);
    background: var(--pollen);
    color: var(--paper);
    border: 1px solid transparent;
    cursor: pointer;
    box-shadow: var(--shadow-card);
    transition:
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease),
      background var(--dur) var(--ease);
  }
  :global([data-mode='dark']) .install {
    color: var(--ink);
  }
  .install:hover:not(.installed) {
    box-shadow: 0 1px 0 color-mix(in oklab, var(--paper) 12%, transparent) inset, 0 18px 40px -16px color-mix(in oklab, var(--ink) 60%, transparent), var(--pollen-halo);
    transform: translateY(-1px);
  }
  .install:active:not(.installed) {
    transform: scale(0.97);
  }
  .install:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo), var(--shadow-card);
  }
  .install.installed {
    background: var(--surface-card);
    color: var(--ok);
    border-color: color-mix(in oklab, var(--ok) 40%, transparent);
    box-shadow: none;
    cursor: default;
  }

  @media (prefers-reduced-motion: reduce) {
    .shelf,
    .shelf-face {
      transform: none !important;
      perspective: none;
    }
    .spine:hover {
      transform: none;
    }
    .detail-sheet,
    .detail-overlay {
      animation: none;
    }
  }
</style>