<script lang="ts">
  // Replay — scrubbable 24h timeline of agent actions.
  // Bottom: horizontal timeline with thumbnails at each event point.
  // Above: the selected event's screenshot (or empty state).
  // Header: app filter, "Verify integrity", "Export .mp4".
  // Draggable playhead snaps to event timestamps.
  import { onMount } from 'svelte'
  import { replay } from '../stores/replay.svelte'
  import { notifications } from '../stores/notifications.svelte'
  import { Badge, Button, Card, SegmentedControl } from '../components/ui'
  import type { ReplayFrame } from '../ipc/types'

  type AppFilter = 'all' | string

  let appFilter = $state<AppFilter>('all')

  // Virtualized thumb strip — when > 80 frames we render only the
  // visible window. (Most users have <100 frames in 24h.)
  const STRIP_THUMB_W = 28
  const STRIP_THUMB_GAP = 4

  let scrollerEl = $state<HTMLDivElement | null>(null)
  let stripScroll = $state(0)
  let stripViewport = $state(0)

  const stripItems = $derived(replay.frames)
  const stripItemWidth = $derived(STRIP_THUMB_W + STRIP_THUMB_GAP)
  const stripStartIdx = $derived(
    Math.max(0, Math.floor(stripScroll / stripItemWidth) - 6)
  )
  const stripEndIdx = $derived(
    Math.min(stripItems.length, Math.ceil((stripScroll + stripViewport) / stripItemWidth) + 6)
  )
  const stripVisible = $derived(stripItems.slice(stripStartIdx, stripEndIdx))
  const stripPadLeft = $derived(stripStartIdx * stripItemWidth)
  const stripPadRight = $derived(
    Math.max(0, (stripItems.length - stripEndIdx) * stripItemWidth)
  )

  // App filter list — derive from frames.
  const appOptions = $derived(() => {
    const apps = new Set<string>()
    for (const f of replay.frames) if (f.app) apps.add(f.app)
    return [
      { value: 'all', label: 'All apps' },
      ...Array.from(apps).sort().map((a) => ({ value: a, label: a })),
    ]
  })

  const filteredFrames = $derived(
    appFilter === 'all'
      ? replay.frames
      : replay.frames.filter((f) => f.app === appFilter)
  )

  // Selected frame is derived from replay.selectedIndex, but only if
  // it's still in the filtered list; otherwise snap to 0.
  const selectedFrame = $derived<ReplayFrame | null>(
    filteredFrames[replay.selectedIndex] ?? filteredFrames[0] ?? null
  )
  const selectedFilteredIndex = $derived(
    selectedFrame
      ? filteredFrames.findIndex((f) => f.id === selectedFrame.id)
      : -1
  )

  function outcomeClass(outcome: string): 'success' | 'error' | 'warn' | 'neutral' {
    if (outcome === 'allowed') return 'success'
    if (outcome === 'denied') return 'error'
    if (outcome === 'errored') return 'warn'
    return 'neutral'
  }

  function outcomeText(o: string): string {
    return o || 'unknown'
  }

  function selectFrame(idx: number): void {
    if (idx < 0 || idx >= filteredFrames.length) return
    replay.selectIndex(idx)
  }

  function scrubToIndex(idx: number): void {
    selectFrame(idx)
  }

  function onStripScroll(): void {
    if (scrollerEl) stripScroll = scrollerEl.scrollLeft
  }

  function fmtTs(ts: string): string {
    try {
      const d = new Date(ts)
      return d.toLocaleString(undefined, {
        month: 'short',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
    } catch { return ts }
  }

  function thumbForFrame(f: ReplayFrame): string {
    // Use the before-screenshot if available; otherwise the after.
    return f.before_screenshot || f.after_screenshot || ''
  }

  function thumbMime(f: ReplayFrame): string {
    return f.before_screenshot
      ? f.before_screenshot_mime || 'image/png'
      : f.after_screenshot_mime || 'image/png'
  }

  async function exportVideo(): Promise<void> {
    try {
      const path = await replay.exportMP4()
      notifications.push({
        kind: 'success',
        title: 'Export complete',
        message: path,
      })
    } catch (e) {
      notifications.push({
        kind: 'error',
        title: replay.lastError || 'Export failed',
        message: String(e),
      })
    }
  }

  async function verify(): Promise<void> {
    await replay.verifyIntegrity()
    if (replay.integrity) {
      notifications.push({
        kind: replay.integrity.valid ? 'success' : 'error',
        title: replay.integrity.valid ? 'Chain verified' : 'Chain broken',
        message: replay.integrity.valid
          ? `${replay.integrity.rows_checked} rows OK`
          : `${replay.integrity.first_break_reason ?? 'break detected'}`,
      })
    }
  }

  $effect(() => {
    if (scrollerEl) {
      stripViewport = scrollerEl.clientWidth
      const ro = new ResizeObserver(() => {
        if (scrollerEl) stripViewport = scrollerEl.clientWidth
      })
      ro.observe(scrollerEl)
      return () => ro.disconnect()
    }
  })

  onMount(() => {
    void replay.refresh()
    void replay.verifyIntegrity()
  })
</script>

<div class="replay-page">
  <header class="page-header">
    <div class="title-row">
      <div>
        <h2 class="display-h2">Action replay</h2>
        <p class="lede">Scrub the last 24 hours of agent activity. Before/after screenshots, outcomes, and gatekeeper decisions.</p>
      </div>
      <div class="header-actions">
        <Button variant="ghost" size="sm" onclick={() => replay.refresh()} loading={replay.loading}>Refresh</Button>
        <Button variant="ghost" size="sm" onclick={verify}>Verify integrity</Button>
        <Button
          variant="primary"
          size="sm"
          onclick={exportVideo}
          loading={replay.exporting}
          disabled={replay.frames.length === 0}
        >
          {replay.exporting ? 'Exporting…' : 'Export .mp4'}
        </Button>
      </div>
    </div>

    <div class="filter-row">
      <SegmentedControl
        value={appFilter}
        options={appOptions()}
        size="sm"
        onchange={(v: string) => {
          appFilter = v
          // snap selection back into bounds
          replay.selectIndex(0)
        }}
      />

      {#if replay.integrity}
        <span
          class="integrity mono"
          class:valid={replay.integrity.valid}
        >
          {replay.integrity.valid ? '✓ Chain verified' : '✗ Chain broken'} · {replay.integrity.rows_checked} rows
        </span>
      {/if}
    </div>
  </header>

  {#if replay.loading && replay.frames.length === 0}
    <div class="loading-grid">
      <div class="shot-skel"></div>
      <div class="meta-skel">
        <div class="bar" style:width="40%"></div>
        <div class="bar" style:width="60%"></div>
        <div class="bar" style:width="80%"></div>
      </div>
    </div>
  {:else if filteredFrames.length === 0}
    <Card elevation="1" padding="lg">
      <div class="empty">
        <div class="empty-icon">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="9" />
            <path d="M12 7v5l3 2" />
          </svg>
        </div>
        <h4>Nothing to replay</h4>
        <p>Once the agent takes its first action, you'll see before/after screenshots and decisions here.</p>
      </div>
    </Card>
  {:else}
    <!-- ── Selected frame ────────────────────────────── -->
    {#if selectedFrame}
      {#key selectedFrame.id}
        <Card elevation="glass" padding="md" class="frame-detail">
          <div class="frame-meta">
            <span class="ts mono">{fmtTs(selectedFrame.timestamp)}</span>
            <Badge tone={outcomeClass(selectedFrame.outcome)} size="sm" dot pulse={selectedFrame.outcome === 'allowed'}>
              {outcomeText(selectedFrame.outcome)}
            </Badge>
            <span class="action mono">{selectedFrame.action}</span>
            <span class="app">{selectedFrame.app}</span>
          </div>

          <p class="frame-message">{selectedFrame.message || selectedFrame.outcome_reason || '—'}</p>

          <div class="shots">
            {#if selectedFrame.before_screenshot}
              <figure>
                <figcaption>Before</figcaption>
                <img
                  src={`data:${thumbMime(selectedFrame)};base64,${selectedFrame.before_screenshot}`}
                  alt="Before action"
                />
              </figure>
            {/if}
            {#if selectedFrame.after_screenshot}
              <figure>
                <figcaption>After</figcaption>
                <img
                  src={`data:${thumbMime(selectedFrame)};base64,${selectedFrame.after_screenshot}`}
                  alt="After action"
                />
              </figure>
            {/if}
            {#if !selectedFrame.before_screenshot && !selectedFrame.after_screenshot}
              <div class="no-shot">No screenshots for this frame.</div>
            {/if}
          </div>
        </Card>
      {/key}
    {/if}

    <!-- ── Horizontal scrubber ───────────────────────── -->
    <div class="scrubber-section">
      <div class="scrubber-meta">
        <span class="count mono">
          {selectedFilteredIndex + 1} / {filteredFrames.length}
        </span>
      </div>
      <div
        class="strip"
        bind:this={scrollerEl}
        onscroll={onStripScroll}
        tabindex="0"
        role="slider"
        aria-label="Scrub timeline"
        aria-valuemin={0}
        aria-valuemax={filteredFrames.length - 1}
        aria-valuenow={selectedFilteredIndex}
      >
        {#if stripPadLeft > 0}<div class="strip-spacer" style:width="{stripPadLeft}px"></div>{/if}
        {#each stripVisible as f, i (f.id)}
          {@const idx = filteredFrames.findIndex((x) => x.id === f.id)}
          <button
            type="button"
            class="thumb"
            class:active={idx === selectedFilteredIndex}
            title={`${fmtTs(f.timestamp)} — ${f.action}`}
            onclick={() => scrubToIndex(idx)}
          >
            {#if thumbForFrame(f)}
              <img
                src={`data:${thumbMime(f)};base64,${thumbForFrame(f)}`}
                alt=""
                loading="lazy"
              />
            {:else}
              <span class="thumb-empty"></span>
            {/if}
            <span class="thumb-dot {outcomeClass(f.outcome)}"></span>
          </button>
        {/each}
        {#if stripPadRight > 0}<div class="strip-spacer" style:width="{stripPadRight}px"></div>{/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .replay-page {
    padding: var(--space-6) var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide);
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .page-header {
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .title-row {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--space-4);
    flex-wrap: wrap;
    margin-bottom: var(--space-3);
  }
  .display-h2 {
    font-family: var(--font-display);
    font-size: var(--size-2xl);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    color: var(--text);
    margin: 0 0 var(--space-2) 0;
    line-height: var(--leading-tight);
  }
  .lede {
    font-size: var(--size-md);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    max-width: 640px;
    margin: 0;
  }

  .header-actions {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }

  .filter-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex-wrap: wrap;
  }
  .integrity {
    font-size: var(--size-sm);
    color: var(--text-muted);
  }
  .integrity.valid { color: var(--success); }
  .integrity:not(.valid) { color: var(--error); }

  /* ── Loading ────────────────────────────────────── */
  .loading-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
    height: 360px;
  }
  .shot-skel {
    background: linear-gradient(90deg, var(--surface-2) 0%, var(--surface-3) 50%, var(--surface-2) 100%);
    background-size: 200% 100%;
    animation: shimmer 1.6s ease-in-out infinite;
    border-radius: var(--radius-lg);
  }
  .meta-skel {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .bar {
    height: 14px;
    border-radius: var(--radius-sm);
    background: linear-gradient(90deg, var(--surface-2) 0%, var(--surface-3) 50%, var(--surface-2) 100%);
    background-size: 200% 100%;
    animation: shimmer 1.6s ease-in-out infinite;
  }

  /* ── Empty state ────────────────────────────────── */
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: var(--space-2);
    padding: var(--space-5);
  }
  .empty-icon {
    width: 48px;
    height: 48px;
    border-radius: var(--radius-lg);
    background: var(--surface-2);
    border: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-faint);
    margin-bottom: var(--space-2);
  }
  .empty h4 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    margin: 0;
  }
  .empty p {
    font-size: var(--size-sm);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    max-width: 360px;
    margin: 0;
  }

  /* ── Frame detail card ──────────────────────────── */
  :global(.frame-detail) {
    animation: fade-in-scale var(--transition-slow) var(--ease-out-expo) both;
  }
  .frame-meta {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
    font-size: var(--size-sm);
  }
  .frame-meta .ts {
    color: var(--text-muted);
    font-size: var(--size-xs);
  }
  .frame-meta .action {
    font-family: var(--font-mono);
    font-weight: var(--weight-semibold);
    color: var(--text);
  }
  .frame-meta .app {
    color: var(--text-muted);
  }
  .frame-message {
    margin: 0 0 var(--space-4) 0;
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    color: var(--text);
  }
  .shots {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: var(--space-4);
  }
  figure { margin: 0; }
  figcaption {
    font-size: var(--size-xs);
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wide);
    margin-bottom: var(--space-2);
  }
  .shots img {
    width: 100%;
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
    transition: border-color var(--transition-fast);
    display: block;
  }
  .shots img:hover { border-color: var(--border-strong); }
  .no-shot {
    grid-column: 1 / -1;
    text-align: center;
    color: var(--text-muted);
    font-size: var(--size-sm);
    padding: var(--space-7) 0;
  }

  /* ── Scrubber ───────────────────────────────────── */
  .scrubber-section {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .scrubber-meta {
    display: flex;
    justify-content: flex-end;
  }
  .count {
    font-size: var(--size-xs);
    color: var(--text-muted);
  }

  .strip {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-3);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow-x: auto;
    overflow-y: hidden;
    height: 84px;
    scroll-snap-type: x proximity;
    scrollbar-width: thin;
  }
  .strip-spacer { flex-shrink: 0; }

  .thumb {
    position: relative;
    appearance: none;
    background: transparent;
    border: 2px solid transparent;
    border-radius: var(--radius-md);
    padding: 0;
    width: 28px;
    height: 52px;
    cursor: pointer;
    overflow: hidden;
    flex-shrink: 0;
    transition:
      border-color var(--transition-fast),
      transform var(--transition-fast) var(--ease-spring),
      box-shadow var(--transition-fast);
    scroll-snap-align: center;
  }
  .thumb:hover {
    transform: translateY(-2px);
    border-color: var(--border-strong);
  }
  .thumb.active {
    border-color: var(--accent);
    box-shadow: var(--shadow-glow);
    transform: translateY(-2px);
  }
  .thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .thumb-empty {
    display: block;
    width: 100%;
    height: 100%;
    background: var(--surface-3);
  }
  .thumb-dot {
    position: absolute;
    bottom: 2px;
    left: 50%;
    transform: translateX(-50%);
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--text-faint);
    box-shadow: 0 0 0 2px var(--surface-1);
  }
  .thumb-dot.success { background: var(--success); }
  .thumb-dot.error   { background: var(--error); }
  .thumb-dot.warn    { background: var(--warn); }
  .thumb-dot.neutral { background: var(--text-faint); }
</style>