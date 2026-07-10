<script lang="ts">
  /**
   * Replay — day meridian: frame stage + cobalt scrub rail + filmstrip.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { replay } from '../../stores/replay.svelte'

  onMount(() => {
    void replay.refresh()
    void replay.verifyIntegrity()
  })

  function onKey(e: KeyboardEvent): void {
    if (!replay.frames.length) return
    if (e.key === 'ArrowLeft' || e.key === 'j' || e.key === 'J') {
      e.preventDefault()
      replay.selectIndex(Math.max(0, replay.selectedIndex - 1))
    }
    if (e.key === 'ArrowRight' || e.key === 'k' || e.key === 'K') {
      e.preventDefault()
      replay.selectIndex(Math.min(replay.frames.length - 1, replay.selectedIndex + 1))
    }
  }

  function shotSrc(data?: string, mime?: string): string | null {
    if (!data) return null
    if (data.startsWith('data:')) return data
    return `data:${mime || 'image/png'};base64,${data}`
  }

  const before = $derived(shotSrc(replay.selected?.before_screenshot, replay.selected?.before_screenshot_mime))
  const after = $derived(shotSrc(replay.selected?.after_screenshot, replay.selected?.after_screenshot_mime))
</script>

<svelte:window onkeydown={onKey} />

<MeridianPage
  kicker="Day meridian"
  title="Replay"
  lead="Re-walk the last day of actions. Scrub the rail, step with ← → or J K, read each sealed frame."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void replay.refresh()}>Refresh</button>
    <button type="button" class="md-btn md-btn-primary" onclick={() => void replay.verifyIntegrity()}>
      Verify seal
    </button>
  {/snippet}

  {#if replay.integrity}
    <div class="seal" class:bad={replay.integrity.valid === false} class:ok={replay.integrity.valid !== false}>
      <span class="dot" aria-hidden="true"></span>
      Timeline {replay.integrity.valid === false ? 'untrusted' : 'sealed'}
      {#if replay.integrity.first_break_reason}
        <span class="reason">— {replay.integrity.first_break_reason}</span>
      {/if}
    </div>
  {/if}

  {#if replay.loading && replay.frames.length === 0}
    <div class="md-empty">Loading timeline…</div>
  {:else if replay.lastError && !/IPC client not started|not connected|Failed to fetch/i.test(String(replay.lastError))}
    <div class="md-empty">{replay.lastError}</div>
  {:else if replay.frames.length === 0}
    <div class="md-empty empty">
      <p class="empty-title">No frames yet</p>
      <p class="empty-lead">
        {#if replay.lastError && /IPC client not started|not connected|Failed to fetch/i.test(String(replay.lastError))}
          Connect the daemon to load the last day of actions.
        {:else}
          Once Condura acts, screenshots and frames land on this meridian.
        {/if}
      </p>
    </div>
  {:else}
    <div class="theatre md-stagger">
      <div class="stage">
        {#if before || after}
          <div class="shots" class:pair={!!(before && after)}>
            {#if before}
              <figure>
                <img src={before} alt="Before" />
                <figcaption>before</figcaption>
              </figure>
            {/if}
            {#if after}
              <figure>
                <img src={after} alt="After" />
                <figcaption>after</figcaption>
              </figure>
            {/if}
          </div>
        {:else}
          <div class="placeholder">
            <p class="cite">frame · no screenshot</p>
            <h2>{replay.selected?.action || 'Action'}</h2>
            <p>{replay.selected?.message || 'This link has no image plane — the ledger text still holds.'}</p>
          </div>
        {/if}
      </div>

      {#if replay.selected}
        <aside class="plate">
          <p class="cite">
            {replay.selectedIndex + 1} / {replay.frames.length}
            · {replay.selected.result || replay.selected.outcome || 'event'}
          </p>
          <h2>{replay.selected.action}</h2>
          <p class="app">{replay.selected.app || '—'}</p>
          <p class="msg">{replay.selected.message}</p>
          <p class="when">{replay.selected.timestamp}</p>
        </aside>
      {/if}
    </div>

    <div class="rail">
      <input
        type="range"
        min="0"
        max={Math.max(0, replay.frames.length - 1)}
        value={replay.selectedIndex}
        aria-label="Replay frame"
        oninput={(e) => replay.selectIndex(Number((e.currentTarget as HTMLInputElement).value))}
      />
      <span class="meta">{replay.selectedIndex + 1} / {replay.frames.length}</span>
    </div>

    <div class="strip">
      {#each replay.frames as frame, i (frame.id)}
        <button
          type="button"
          class="thumb"
          class:on={i === replay.selectedIndex}
          onclick={() => replay.selectIndex(i)}
          aria-label={`Frame ${i + 1}: ${frame.action}`}
        >
          <span class="tick" data-v={frame.result || frame.outcome} aria-hidden="true"></span>
          <span class="t-label">{frame.action}</span>
        </button>
      {/each}
    </div>
  {/if}
</MeridianPage>

<style>
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 8px;
  }
  .seal {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 16px;
    padding: 12px 16px;
    border-radius: 16px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }
  .seal.ok {
    border-color: color-mix(in oklab, var(--md-live) 30%, transparent);
  }
  .seal.bad {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 40%, transparent);
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--md-live);
  }
  .seal.bad .dot {
    background: var(--md-halt);
  }
  .reason {
    text-transform: none;
    letter-spacing: 0;
    opacity: 0.85;
  }
  .empty-title {
    margin: 0;
    font-family: var(--md-font-display);
    font-size: 16px;
  }
  .empty-lead {
    margin: 6px 0 0;
    max-width: 42ch;
    font-size: 13px;
    color: var(--md-ink-mute);
  }

  .theatre {
    display: grid;
    grid-template-columns: 1.4fr 0.8fr;
    gap: 14px;
    margin-bottom: 14px;
  }
  .stage {
    border-radius: 22px;
    border: 1px solid var(--md-line-strong);
    background:
      radial-gradient(120% 80% at 50% 0%, color-mix(in oklab, var(--md-cobalt) 10%, transparent), transparent 55%),
      var(--md-stage);
    min-height: 280px;
    overflow: hidden;
    box-shadow: inset 0 1px 0 color-mix(in oklab, var(--md-surface) 40%, transparent);
  }
  .shots {
    display: grid;
    height: 100%;
    min-height: 280px;
  }
  .shots.pair {
    grid-template-columns: 1fr 1fr;
  }
  figure {
    margin: 0;
    position: relative;
    min-height: 280px;
    background: #0b1526;
  }
  figure img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
  figcaption {
    position: absolute;
    left: 10px;
    bottom: 10px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: #fff;
    background: color-mix(in oklab, #0b1526 55%, transparent);
    padding: 4px 8px;
    border-radius: 6px;
  }
  .placeholder {
    padding: 36px 28px;
  }
  .placeholder h2 {
    font-family: var(--md-font-display);
    font-size: 28px;
    letter-spacing: -0.04em;
    margin: 0 0 10px;
  }
  .placeholder p:last-child {
    margin: 0;
    color: var(--md-ink-mute);
    max-width: 40ch;
    line-height: 1.5;
  }
  .plate {
    border-radius: 22px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 20px;
    box-shadow: var(--md-shadow);
  }
  .plate h2 {
    font-family: var(--md-font-display);
    font-size: 22px;
    letter-spacing: -0.04em;
    margin: 0 0 6px;
  }
  .app {
    margin: 0 0 12px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-cobalt);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }
  .msg {
    margin: 0 0 14px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .when {
    margin: 0;
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-ink-faint);
  }

  .rail {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 12px;
    padding: 12px 14px;
    border-radius: 16px;
    background: var(--md-surface);
    border: 1px solid var(--md-line-strong);
  }
  .rail input {
    flex: 1;
    accent-color: var(--md-cobalt);
    height: 24px;
  }
  .meta {
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-ink-faint);
    font-variant-numeric: tabular-nums;
  }

  .strip {
    display: flex;
    gap: 8px;
    overflow-x: auto;
    padding-bottom: 4px;
  }
  .thumb {
    flex: none;
    width: 120px;
    text-align: left;
    padding: 10px;
    border-radius: 14px;
    border: 1px solid var(--md-line);
    background: var(--md-stage);
    cursor: pointer;
    color: inherit;
  }
  .thumb.on {
    border-color: color-mix(in oklab, var(--md-cobalt) 45%, transparent);
    background: var(--md-surface);
    box-shadow: var(--md-shadow);
  }
  .thumb:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .tick {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--md-ink-faint);
    margin-bottom: 8px;
  }
  .tick[data-v='allow'],
  .tick[data-v='ok'] {
    background: var(--md-live);
  }
  .tick[data-v='block'],
  .tick[data-v='error'] {
    background: var(--md-halt);
  }
  .tick[data-v='prompt'] {
    background: var(--md-cobalt);
  }
  .t-label {
    display: block;
    font-size: 11px;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  @media (max-width: 800px) {
    .theatre {
      grid-template-columns: 1fr;
    }
    .shots.pair {
      grid-template-columns: 1fr;
    }
  }
</style>
