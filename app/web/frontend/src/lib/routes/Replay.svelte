<script lang="ts">
  // Replay — scrubbable 24h timeline of agent actions.
  import { onMount } from 'svelte'
  import { replay } from '../stores/replay.svelte'
  import { notifications } from '../stores/notifications.svelte'
  import type { ReplayFrame } from '../ipc/types'
  import Button from '$components/v1/Button.svelte'
  import Card from '$components/v1/Card.svelte'
  import Chip from '$components/v1/Chip.svelte'
  import Pill from '$components/v1/Pill.svelte'
  import Stack from '$components/v1/Stack.svelte'
  import Inline from '$components/v1/Inline.svelte'
  import Surface from '$components/v1/Surface.svelte'
  import EmptyState from '$components/v1/EmptyState.svelte'
  import LoadingState from '$components/v1/LoadingState.svelte'
  import Dot from '$components/v1/Dot.svelte'

  type AppFilter = 'all' | string

  let appFilter = $state<AppFilter>('all')

  const STRIP_THUMB_W = 28
  const STRIP_THUMB_GAP = 4

  let scrollerEl = $state<HTMLDivElement | null>(null)
  let stripScroll = $state(0)
  let stripViewport = $state(0)

  const stripItems = $derived(replay.frames)
  const stripItemWidth = $derived(STRIP_THUMB_W + STRIP_THUMB_GAP)
  const stripStartIdx = $derived(Math.max(0, Math.floor(stripScroll / stripItemWidth) - 6))
  const stripEndIdx = $derived(
    Math.min(stripItems.length, Math.ceil((stripScroll + stripViewport) / stripItemWidth) + 6)
  )
  const stripVisible = $derived(stripItems.slice(stripStartIdx, stripEndIdx))
  const stripPadLeft = $derived(stripStartIdx * stripItemWidth)
  const stripPadRight = $derived(
    Math.max(0, (stripItems.length - stripEndIdx) * stripItemWidth)
  )

  const appOptions = $derived(() => {
    const apps = new Set<string>()
    for (const f of replay.frames) if (f.app) apps.add(f.app)
    return [
      { value: 'all', label: 'All apps' },
      ...Array.from(apps)
        .sort()
        .map((a) => ({ value: a, label: a })),
    ]
  })

  const filteredFrames = $derived(
    appFilter === 'all' ? replay.frames : replay.frames.filter((f) => f.app === appFilter)
  )

  const selectedFrame = $derived<ReplayFrame | null>(
    filteredFrames[replay.selectedIndex] ?? filteredFrames[0] ?? null
  )

  const selectedFilteredIndex = $derived(
    selectedFrame ? filteredFrames.findIndex((f) => f.id === selectedFrame.id) : -1
  )

  function outcomePill(outcome: string): 'success' | 'error' | 'warning' | 'neutral' {
    if (outcome === 'allowed') return 'success'
    if (outcome === 'denied') return 'error'
    if (outcome === 'errored') return 'warning'
    return 'neutral'
  }

  function dotVariant(outcome: string): 'success' | 'error' | 'warning' | 'neutral' {
    if (outcome === 'allowed') return 'success'
    if (outcome === 'denied') return 'error'
    if (outcome === 'errored') return 'warning'
    return 'neutral'
  }

  function outcomeText(o: string): string {
    return o || 'unknown'
  }

  function selectFrame(idx: number): void {
    if (idx < 0 || idx >= filteredFrames.length) return
    replay.selectIndex(idx)
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
    } catch {
      return ts
    }
  }

  function thumbForFrame(f: ReplayFrame): string {
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

<Stack class="replay-page" gap="5" padding="7">
  <header class="page-header">
    <Inline gap="4" align="end" justify="between" class="title-row">
      <div class="title-block">
        <h2 class="page-title">Action replay</h2>
        <p class="lede">
          Scrub the last 24 hours of agent activity. Before/after screenshots, outcomes,
          and gatekeeper decisions.
        </p>
      </div>
      <Inline gap="2" align="center" class="header-actions">
        <Button variant="tertiary" size="sm" loading={replay.loading} onclick={() => replay.refresh()}>
          Refresh
        </Button>
        <Button variant="tertiary" size="sm" onclick={verify}>
          Verify integrity
        </Button>
        <Button
          variant="primary"
          size="sm"
          loading={replay.exporting}
          disabled={replay.frames.length === 0}
          onclick={exportVideo}
        >
          {replay.exporting ? 'Exporting…' : 'Export .mp4'}
        </Button>
      </Inline>
    </Inline>

    <Inline gap="3" align="center" wrap={true} class="filter-row">
      {#each appOptions() as opt (opt.value)}
        <Chip
          selected={appFilter === opt.value}
          onclick={() => {
            appFilter = opt.value
            replay.selectIndex(0)
          }}
        >
          {opt.label}
        </Chip>
      {/each}

      {#if replay.integrity}
        <Pill
          variant={replay.integrity.valid ? 'success' : 'error'}
          size="sm"
          label="{replay.integrity.valid ? 'Chain verified' : 'Chain broken'} · {replay.integrity.rows_checked} rows"
        />
      {/if}
    </Inline>
  </header>

  {#if replay.loading && replay.frames.length === 0}
    <LoadingState kind="cold" />
  {:else if filteredFrames.length === 0}
    <Surface variant="raised" padding="4">
      <EmptyState
        primary="Nothing to replay"
        voice="mono"
        secondary="Once the agent takes its first action, you'll see before/after screenshots and decisions here."
      />
    </Surface>
  {:else}
    {#if selectedFrame}
      {#key selectedFrame.id}
        <div class="frame-card">
        <Card variant="raised" padding="4">
          {#snippet children()}
            <Inline gap="2" align="center" wrap={true} class="frame-meta">
              <span class="ts mono">{fmtTs(selectedFrame.timestamp)}</span>
              <Pill
                variant={outcomePill(selectedFrame.outcome)}
                size="sm"
                label={outcomeText(selectedFrame.outcome)}
              />
              <span class="action mono">{selectedFrame.action}</span>
              <span class="app">{selectedFrame.app}</span>
            </Inline>

            <p class="frame-message">
              {selectedFrame.message || selectedFrame.outcome_reason || '—'}
            </p>

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
                <p class="no-shot">No screenshots for this frame.</p>
              {/if}
            </div>
          {/snippet}
        </Card>
        </div>
      {/key}
    {/if}

    <Stack gap="2" class="scrubber-section">
      <div class="scrubber-meta">
        <span class="count mono">
          {selectedFilteredIndex + 1} / {filteredFrames.length}
        </span>
      </div>

      <Surface variant="sunken" padding="3" radius="md" class="strip-surface">
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
          {#if stripPadLeft > 0}
            <div class="strip-spacer" style:width="{stripPadLeft}px"></div>
          {/if}
          {#each stripVisible as f (f.id)}
            {@const idx = filteredFrames.findIndex((x) => x.id === f.id)}
            <button
              type="button"
              class="thumb"
              class:active={idx === selectedFilteredIndex}
              title={`${fmtTs(f.timestamp)} — ${f.action}`}
              onclick={() => selectFrame(idx)}
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
              <span class="thumb-dot">
                <Dot variant={dotVariant(f.outcome)} size="xs" />
              </span>
            </button>
          {/each}
          {#if stripPadRight > 0}
            <div class="strip-spacer" style:width="{stripPadRight}px"></div>
          {/if}
        </div>
      </Surface>
    </Stack>
  {/if}
</Stack>

<style>
  .replay-page {
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide, 72rem);
    margin: 0 auto;
  }

  .page-header {
    animation: replay-enter var(--duration-slow) var(--ease-standard) both;
  }

  @keyframes replay-enter {
    from {
      opacity: 0;
      transform: translateY(6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .title-row {
    width: 100%;
  }

  .title-block {
    flex: 1;
    min-width: 0;
  }

  .page-title {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    font-weight: 500;
    letter-spacing: -0.02em;
    color: var(--content-primary);
    margin: 0 0 var(--space-2);
    line-height: var(--text-h2-leading);
  }

  .lede {
    font-size: var(--text-body-size);
    color: var(--content-secondary);
    line-height: 1.55;
    max-width: 40rem;
    margin: 0;
  }

  .header-actions {
    flex-shrink: 0;
  }

  .frame-card {
    animation: frame-enter var(--duration-base) var(--ease-standard) both;
  }

  @keyframes frame-enter {
    from {
      opacity: 0;
      transform: scale(0.99);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  .frame-meta {
    margin-bottom: var(--space-3);
    font-size: var(--text-body-sm-size);
  }

  .frame-meta .ts {
    color: var(--content-tertiary);
    font-size: var(--text-caption-size);
  }

  .frame-meta .action {
    font-weight: 500;
    color: var(--content-primary);
  }

  .frame-meta .app {
    color: var(--content-tertiary);
  }

  .frame-message {
    margin: 0 0 var(--space-4);
    font-size: var(--text-body-size);
    line-height: 1.55;
    color: var(--content-primary);
  }

  .shots {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: var(--space-4);
  }

  figure {
    margin: 0;
  }

  figcaption {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.06em;
    margin-bottom: var(--space-2);
  }

  .shots img {
    width: 100%;
    border-radius: var(--radius-md);
    border: 1px solid var(--border-default);
    transition: border-color var(--duration-fast) var(--ease-standard);
    display: block;
  }

  .shots img:hover {
    border-color: var(--border-strong);
  }

  .no-shot {
    grid-column: 1 / -1;
    text-align: center;
    color: var(--content-tertiary);
    font-size: var(--text-body-sm-size);
    padding: var(--space-7) 0;
    margin: 0;
  }

  .scrubber-meta {
    display: flex;
    justify-content: flex-end;
  }

  .count {
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
  }

  .mono {
    font-family: var(--font-mono);
  }

  .strip {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    overflow-x: auto;
    overflow-y: hidden;
    height: 68px;
    scroll-snap-type: x proximity;
    scrollbar-width: thin;
  }

  .strip-spacer {
    flex-shrink: 0;
  }

  .thumb {
    position: relative;
    appearance: none;
    background: transparent;
    border: 2px solid transparent;
    border-radius: var(--radius-sm);
    padding: 0;
    width: 28px;
    height: 52px;
    cursor: pointer;
    overflow: hidden;
    flex-shrink: 0;
    transition:
      border-color var(--duration-fast) var(--ease-standard),
      transform var(--duration-fast) var(--ease-standard);
    scroll-snap-align: center;
  }

  .thumb:hover {
    transform: translateY(-2px);
    border-color: var(--border-strong);
  }

  .thumb.active {
    border-color: var(--content-accent);
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
    background-color: var(--surface-sunken);
  }

  .thumb-dot {
    position: absolute;
    bottom: 2px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    box-shadow: 0 0 0 2px var(--surface-base);
    border-radius: var(--radius-pill);
  }
</style>
