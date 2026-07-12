<script lang="ts">
  /**
   * Replay — day meridian theatre.
   * Signature: screenshot stage · cobalt scrub · filmstrip · keyboard scrub.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { replay } from '../../stores/replay.svelte'

  let stripEl = $state<HTMLDivElement | null>(null)
  let exportNote = $state('')

  onMount(() => {
    void replay.refresh()
    void replay.verifyIntegrity()
  })

  function onKey(e: KeyboardEvent): void {
    if (!replay.frames.length) return
    const t = e.target as HTMLElement | null
    if (
      t &&
      (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.isContentEditable)
    ) {
      return
    }
    if (e.key === 'ArrowLeft' || e.key === 'j' || e.key === 'J') {
      e.preventDefault()
      replay.selectIndex(Math.max(0, replay.selectedIndex - 1))
    }
    if (e.key === 'ArrowRight' || e.key === 'k' || e.key === 'K') {
      e.preventDefault()
      replay.selectIndex(Math.min(replay.frames.length - 1, replay.selectedIndex + 1))
    }
    if (e.key === 'Home') {
      e.preventDefault()
      replay.selectIndex(0)
    }
    if (e.key === 'End') {
      e.preventDefault()
      replay.selectIndex(replay.frames.length - 1)
    }
  }

  function shotSrc(data?: string, mime?: string): string | null {
    if (!data) return null
    if (data.startsWith('data:')) return data
    return `data:${mime || 'image/png'};base64,${data}`
  }

  const before = $derived(
    shotSrc(replay.selected?.before_screenshot, replay.selected?.before_screenshot_mime)
  )
  const after = $derived(
    shotSrc(replay.selected?.after_screenshot, replay.selected?.after_screenshot_mime)
  )
  const offline = $derived(
    !!replay.lastError &&
      /IPC client not started|not connected|Failed to fetch/i.test(String(replay.lastError))
  )

  $effect(() => {
    replay.selectedIndex
    if (!stripEl) return
    const thumb = stripEl.querySelector<HTMLElement>('.thumb.on')
    thumb?.scrollIntoView({ inline: 'center', block: 'nearest', behavior: 'smooth' })
  })

  function goAudit(): void {
    window.location.hash = '#/audit'
  }

  function goAsk(): void {
    window.location.hash = '#/'
  }

  async function exportMp4(): Promise<void> {
    exportNote = ''
    try {
      const path = await replay.exportMP4()
      exportNote = path ? `Exported · ${path}` : 'Export finished'
    } catch {
      exportNote = replay.lastError || 'Export failed'
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<MeridianPage
  kicker="Day meridian · 24h"
  title="Replay"
  lead="Re-walk the last day of actions. Scrub the rail, step with ← → or J K, read each sealed frame."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void replay.refresh()}>Refresh</button>
    <button
      type="button"
      class="md-btn md-btn-ghost"
      disabled={replay.exporting || !replay.frames.length}
      onclick={() => void exportMp4()}
    >
      {replay.exporting ? 'Exporting…' : 'Export'}
    </button>
    <button type="button" class="md-btn md-btn-primary" onclick={() => void replay.verifyIntegrity()}>
      Verify seal
    </button>
  {/snippet}

  <div class="desk md-stagger">
    <p class="contract">
      <span class="live-dot" aria-hidden="true"></span>
      Frames are sealed readings of what Condura did — scrub them, don’t rewrite them.
    </p>

    {#if replay.integrity}
      <div
        class="seal"
        class:bad={replay.integrity.valid === false}
        class:ok={replay.integrity.valid !== false}
      >
        <span class="dot" aria-hidden="true"></span>
        <div>
          <p class="cite">timeline seal</p>
          <strong>Timeline {replay.integrity.valid === false ? 'untrusted' : 'sealed'}</strong>
          {#if replay.integrity.first_break_reason}
            <p class="reason">{replay.integrity.first_break_reason}</p>
          {/if}
          {#if replay.integrity.rows_checked}
            <p class="reason">{replay.integrity.rows_checked} rows checked</p>
          {/if}
        </div>
        <button type="button" class="md-btn md-btn-ghost tiny" onclick={goAudit}>Open Audit</button>
      </div>
    {/if}

    {#if exportNote}
      <p class="export-note" class:bad={/fail/i.test(exportNote)}>{exportNote}</p>
    {/if}

    {#if replay.loading && replay.frames.length === 0}
      <div class="md-empty">Loading timeline…</div>
    {:else if replay.lastError && !offline}
      <div class="md-empty">{replay.lastError}</div>
    {:else if replay.frames.length === 0}
      <div class="empty-atlas">
        <p class="cite">{offline ? 'timeline offline' : 'quiet meridian'}</p>
        <h2>{offline ? 'Frames unread' : 'No frames yet'}</h2>
        <p class="empty-lead">
          {#if offline}
            Connect the daemon to load the last day of actions.
          {:else}
            Once Condura acts, screenshots and frames land on this meridian. Ask something to begin.
          {/if}
        </p>
        <div class="empty-actions">
          <button type="button" class="md-btn md-btn-primary" onclick={goAsk}>Go to Ask</button>
          <button type="button" class="md-btn md-btn-ghost" onclick={goAudit}>Open Audit</button>
        </div>
      </div>
    {:else}
      <div class="theatre">
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
              <p>
                {replay.selected?.message ||
                  'This link has no image plane — the ledger text still holds.'}
              </p>
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
            {#if replay.selected.outcome_reason}
              <p class="outcome">{replay.selected.outcome_reason}</p>
            {/if}
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
      <p class="keys">← → or J K to step · Home / End for ends</p>

      <div class="strip" bind:this={stripEl}>
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
            <span class="t-when">{frame.timestamp}</span>
          </button>
        {/each}
      </div>
    {/if}
  </div>
</MeridianPage>

<style>
  .desk {
    display: grid;
    gap: 14px;
  }
  .contract {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    margin: 0;
    padding: 12px 14px;
    border-radius: 14px;
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 22%, var(--md-line));
    background: color-mix(in oklab, var(--md-cobalt) 6%, var(--md-surface));
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .live-dot {
    width: 8px;
    height: 8px;
    margin-top: 5px;
    flex: none;
    border-radius: 50%;
    background: var(--md-cobalt);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-cobalt) 16%, transparent);
  }
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
    gap: 14px;
    align-items: center;
    padding: 14px 16px;
    border-radius: 16px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
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
    flex: none;
    background: var(--md-live);
  }
  .seal.bad .dot {
    background: var(--md-halt);
  }
  .seal strong {
    font-family: var(--md-font-display);
    font-size: 15px;
    letter-spacing: -0.03em;
  }
  .reason {
    margin: 4px 0 0;
    font-size: 12px;
    color: var(--md-ink-mute);
  }
  .seal :global(.md-btn.tiny) {
    margin-left: auto;
    padding: 7px 12px;
    font-size: 12px;
    flex: none;
  }
  .export-note {
    margin: 0;
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-live);
  }
  .export-note.bad {
    color: var(--md-halt);
  }
  .empty-atlas {
    border-radius: 22px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 28px 24px;
    box-shadow: var(--md-shadow);
    max-width: 520px;
  }
  .empty-atlas h2 {
    font-family: var(--md-font-display);
    font-size: 26px;
    letter-spacing: -0.04em;
    margin: 0 0 8px;
  }
  .empty-lead {
    margin: 0 0 16px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
    max-width: 42ch;
  }
  .empty-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .theatre {
    display: grid;
    grid-template-columns: 1.4fr 0.8fr;
    gap: 14px;
  }
  .stage {
    border-radius: 22px;
    border: 1px solid var(--md-line-strong);
    background:
      radial-gradient(120% 80% at 50% 0%, color-mix(in oklab, var(--md-cobalt) 10%, transparent), transparent 55%),
      var(--md-stage);
    min-height: 280px;
    overflow: hidden;
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
  .when,
  .outcome {
    margin: 0;
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-ink-faint);
  }
  .outcome {
    margin-top: 8px;
    color: var(--md-ink-mute);
  }

  .rail {
    display: flex;
    align-items: center;
    gap: 14px;
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
  .keys {
    margin: -4px 0 0;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.06em;
    color: var(--md-ink-faint);
  }

  .strip {
    display: flex;
    gap: 8px;
    overflow-x: auto;
    padding-bottom: 4px;
  }
  .thumb {
    flex: none;
    width: 140px;
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
  .t-when {
    display: block;
    margin-top: 4px;
    font-family: var(--md-font-mono);
    font-size: 9px;
    color: var(--md-ink-faint);
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
    .seal {
      flex-wrap: wrap;
    }
    .seal :global(.md-btn.tiny) {
      margin-left: 0;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .strip :global(*) {
      scroll-behavior: auto;
    }
  }
</style>
