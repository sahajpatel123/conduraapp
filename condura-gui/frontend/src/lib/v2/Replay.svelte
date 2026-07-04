<!--
  Replay — Condura v2 action-replay surface.

  Per the spec: "A film strip you can scrub. Scrubber is a real timeline
  with 24h of frames. Frame thumbnails crossfade as you scrub."

  Composition:
    1. Title + scrub position (real-time clock)
    2. The strip — a horizontal row of frame thumbnails representing
       the last 24h, with the current frame highlighted
    3. The screenshot at scrub position — large preview, changeable
    4. Decision / model output / intent stack for the scrubbed moment

  Pure-presentation. Parent owns the frames array.

  Props:
    frames:  { ts: string, hour: number, summary: string, decision?: string, intent?: string }[]
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Eyebrow, Glyph, Button } from '$lib/v2'

  export interface ReplayFrame {
    id: string
    ts: string
    hour: number            // 0..23
    summary: string
    screenshot?: string    // optional url/dataurl
    decision?: string
    intent?: string
  }

  let {
    frames = [] as ReplayFrame[],
  }: {
    frames?: ReplayFrame[]
  } = $props()

  // Initial scrub = most recent frame. The user can scrub freely.
  let scrubIndex = $state(Math.max(0, frames.length - 1))
  $effect(() => {
    // If frames are added/removed, keep scrub on most recent.
    if (scrubIndex >= frames.length) scrubIndex = Math.max(0, frames.length - 1)
  })

  const current = $derived(frames[scrubIndex])

  function setScrub(i: number) {
    scrubIndex = Math.max(0, Math.min(frames.length - 1, i))
  }
  function play() {
    let i = scrubIndex
    const id = setInterval(() => {
      i++
      if (i >= frames.length) { clearInterval(id); return }
      setScrub(i)
    }, 800)
  }

  // The active hour-window for the visible "film". Real timestamps.
  const startTs = $derived(frames.length > 0 ? new Date(frames[frames.length - 1].ts) : new Date())
</script>

<div data-v2 style="
  flex: 1;
  background: var(--v2-paper);
  overflow-y: auto;
  padding: var(--v2-space-12);
  box-sizing: border-box;
">
  <div style="max-width: 1100px; margin: 0 auto;">

    <Stack gap={8}>

      <!-- ── Title ─────────────────────────────────── -->
      <Inline gap={4} justify="between" align="end">
        <Stack gap={3}>
          <Eyebrow left="replay" right="last 24 hours" tone="ink-3" />
          <Ink kind="display" as="h1">Scrub the day.</Ink>
          <Ink kind="body-2" tone="ink-2" as="p" style:max-width="640px">
            Every action condura took in the last 24 hours, with the
            reasoning behind it. Drag the strip — the moment updates.
          </Ink>
        </Stack>
        <Inline gap={2}>
          <Button variant="ghost" size="small" onclick={play}>Play</Button>
          <Button variant="ghost" size="small" onclick={() => setScrub(0)}>↤ Start</Button>
          <Button variant="ghost" size="small" onclick={() => setScrub(frames.length - 1)}>End ↦</Button>
        </Inline>
      </Inline>

      <!-- ── Preview ────────────────────────────────── -->
      {#if current}
        <Surface elevation={2} padding="0" radius="3" tone="paper" style:overflow="hidden">
          <!-- Screenshot-like surface (placeholder via paper-2 gradient + grain) -->
          <div style="
            width: 100%; height: 320px;
            background:
              linear-gradient(135deg, var(--v2-paper-2), var(--v2-paper)),
              repeating-linear-gradient(45deg, transparent 0 4px, rgba(27,26,23,0.02) 4px 5px);
            border-bottom: 1px solid var(--v2-rule);
            display: grid; place-items: center;
            position: relative;
            color: var(--v2-ink-3);
            font-family: var(--v2-font-mono);
            font-size: var(--v2-text-12);
          ">
            {#if current.screenshot}
              <img src={current.screenshot} alt={`screenshot at ${current.ts}`} style="width:100%; height:100%; object-fit: cover; display: block;" />
            {:else}
              <span style="
                position: absolute;
                bottom: var(--v2-space-3);
                right: var(--v2-space-3);
                padding: var(--v2-space-2) var(--v2-space-3);
                background: var(--v2-paper);
                border-radius: var(--v2-radius-1);
                border: 1px solid var(--v2-rule);
              ">screenshot placeholder · {current.ts}</span>
            {/if}
          </div>

          <div style="padding: var(--v2-space-6) var(--v2-space-8);">
            <Stack gap={4}>
              <Inline gap={4} justify="between" align="baseline">
                <Stack gap={1}>
                  <Ink kind="caption" tone="ink-3">moment</Ink>
                  <Ink kind="title" style:font-size="var(--v2-text-20)">{current.ts}</Ink>
                </Stack>
                <Stack gap={1} align="end">
                  <Ink kind="caption" tone="ink-3">hour</Ink>
                  <Ink kind="mono" weight="medium" style:font-size="var(--v2-text-20)">{String(current.hour).padStart(2, '0')}:00</Ink>
                </Stack>
              </Inline>
              <Ink kind="body">{current.summary}</Ink>
              {#if current.decision}
                <Stack gap={1}>
                  <Ink kind="caption" tone="ink-3">decision</Ink>
                  <Ink kind="ui-small" tone="ink-2" style:font-family="var(--v2-font-mono)">{current.decision}</Ink>
                </Stack>
              {/if}
              {#if current.intent}
                <Stack gap={1}>
                  <Ink kind="caption" tone="ink-3">intent</Ink>
                  <Ink kind="ui-small" tone="ink-2" italic>"{current.intent}"</Ink>
                </Stack>
              {/if}
            </Stack>
          </div>
        </Surface>
      {/if}

      <!-- ── The film strip ──────────────────────────── -->
      <Stack gap={3}>
        <Inline gap={3} justify="between" align="baseline">
          <Eyebrow left="strip" right={`${frames.length} frames · 24h`} tone="ink-3" />
          {#if current}
            <Inline gap={2}>
              <Glyph name="book" size={12} />
              <Ink kind="caption" tone="ink-3">click any frame to scrub</Ink>
            </Inline>
          {/if}
        </Inline>

        <div style="
          display: flex;
          align-items: flex-end;
          gap: 4px;
          padding: var(--v2-space-3);
          background: var(--v2-paper-2);
          border-radius: var(--v2-radius-2);
          overflow-x: auto;
        ">
          {#each frames as f, i}
            {@const active = i === scrubIndex}
            <button
              data-v2-strip-frame
              data-active={active}
              onclick={() => setScrub(i)}
              aria-label={`Frame at ${f.ts}`}
              style="
                flex: 0 0 48px;
                height: {active ? '84px' : '64px'};
                border: none;
                background:
                  linear-gradient(180deg, rgba(193,138,74,{active ? '0.18' : '0.06'}), transparent),
                  var(--v2-paper);
                border-radius: var(--v2-radius-1);
                cursor: pointer;
                transition:
                  height var(--v2-dur-fast) var(--v2-ease-out-soft),
                  background-color var(--v2-dur-fast) var(--v2-ease-out-soft),
                  transform var(--v2-dur-fast) var(--v2-ease-out-soft);
                display: grid;
                place-items: center;
                font-family: var(--v2-font-mono);
                font-size: 10px;
                color: var(--v2-ink-3);
                padding: 4px;
                box-sizing: border-box;
                position: relative;
              "
            >
              <span style="transform: rotate(-90deg); white-space: nowrap;">{f.ts.slice(11, 16)}</span>
              {#if active}
                <span style="
                  position: absolute;
                  inset: -2px;
                  border: 1px solid var(--v2-accent);
                  border-radius: var(--v2-radius-1);
                  pointer-events: none;
                "></span>
              {/if}
            </button>
          {/each}
        </div>
      </Stack>

    </Stack>

  </div>
</div>
