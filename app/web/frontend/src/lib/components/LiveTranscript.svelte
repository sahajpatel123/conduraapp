<!-- LiveTranscript.svelte
     Rolling transcript display for voice sessions. Each line is a
     Card with elevation=glass, padding=sm. New lines animate in
     with .anim-slide-up. Auto-scrolls to bottom unless the user
     has scrolled up (manual scroll position is preserved).
     Subscribes to voice.partial and voice.final IPC events. -->

<script lang="ts">
  import { onMount, onDestroy, tick } from 'svelte'
  import { marked } from 'marked'
  import DOMPurify from 'dompurify'
  import { ipc } from '../ipc/client'
  import { t } from '../i18n'
  import Card from './ui/Card.svelte'
  import Badge from './ui/Badge.svelte'

  type Speaker = 'user' | 'assistant'
  type EntryKind = 'partial' | 'final'

  interface Entry {
    id: number
    speaker: Speaker
    text: string
    kind: EntryKind
    confidence?: number
  }

  let entries = $state<Entry[]>([])
  let isRecording = $state<boolean>(false)
  let nextId = 1
  let cleanups: Array<() => void> = []
  let hideTimer: ReturnType<typeof setTimeout> | null = null
  let scrollEl: HTMLDivElement | undefined = $state()
  // True when the user has scrolled away from the bottom. We
  // only auto-scroll while this stays false.
  let stuckToBottom = $state(true)

  function clearHideTimer(): void {
    if (hideTimer !== null) {
      clearTimeout(hideTimer)
      hideTimer = null
    }
  }

  function handleScroll(): void {
    if (!scrollEl) return
    const threshold = 24 // px from the bottom considered "stuck"
    const distance = scrollEl.scrollHeight - scrollEl.scrollTop - scrollEl.clientHeight
    stuckToBottom = distance <= threshold
  }

  async function scrollToBottom(): Promise<void> {
    if (!stuckToBottom || !scrollEl) return
    await tick()
    scrollEl.scrollTop = scrollEl.scrollHeight
  }

  function pushEntry(entry: Entry): void {
    // Replace a trailing partial for the same speaker so the live
    // transcript doesn't grow a wall of "listening…" rows.
    const last = entries[entries.length - 1]
    if (last && last.kind === 'partial' && last.speaker === entry.speaker) {
      entries = [...entries.slice(0, -1), entry]
    } else {
      entries = [...entries, entry]
    }
    void scrollToBottom()
  }

  function renderSafeMarkdown(text: string): string {
    if (!text) return ''
    try {
      // Voice transcripts flow through this renderer with full IPC
      // access behind them, so any <script> / event handler / data: URL
      // that survives marked.parse must be stripped before {@html}.
      const html = marked.parse(text, { async: false, breaks: true }) as string
      return DOMPurify.sanitize(html)
    } catch {
      // Fall back to plain text on parse failure.
      return text.replace(/[&<>"']/g, (c) =>
        c === '&' ? '&amp;'
        : c === '<' ? '&lt;'
        : c === '>' ? '&gt;'
        : c === '"' ? '&quot;'
        : '&#39;'
      )
    }
  }

  onMount(() => {
    try {
      cleanups.push(
        ipc.on('voice.partial' as never, ((data: { recording?: boolean; samples?: number }) => {
          isRecording = data.recording ?? false
          if (isRecording) {
            pushEntry({
              id: nextId++,
              speaker: 'user',
              kind: 'partial',
              text: t('voice.transcript.listening')
            })
          }
        }) as never),

        ipc.on('voice.final' as never, ((data: { text?: string; confidence?: number }) => {
          isRecording = false
          const text = (data.text ?? '').trim()
          if (text.length === 0) return
          pushEntry({
            id: nextId++,
            speaker: 'user',
            kind: 'final',
            text,
            confidence: data.confidence
          })
          clearHideTimer()
          hideTimer = setTimeout(() => {
            // Final, persisted transcripts stay until replaced by
            // the next session — the original component cleared
            // them after 5s; we keep the more useful behaviour.
            hideTimer = null
          }, 5000)
        }) as never)
      )
    } catch {
      // Not running inside Wails (unit tests / static preview).
    }
  })

  onDestroy(() => {
    clearHideTimer()
    cleanups.forEach((c) => c())
    cleanups = []
  })
</script>

{#if entries.length > 0 || isRecording}
  <div
    class="live-transcript glass-card"
    class:recording={isRecording}
    bind:this={scrollEl}
    onscroll={handleScroll}
    role="log"
    aria-live="polite"
    aria-label={t('voice.transcript.aria_label')}
  >
    <div class="transcript-list">
      {#each entries as entry (entry.id)}
        <div class="transcript-line anim-slide-up" class:is-partial={entry.kind === 'partial'}>
          <div class="line-meta">
            <Badge tone={entry.speaker === 'user' ? 'accent' : 'info'} size="xs">
              {entry.speaker === 'user' ? t('voice.transcript.you') : t('voice.transcript.condura')}
            </Badge>
            {#if entry.kind === 'partial'}
              <span class="partial-pulse anim-glow-pulse" aria-hidden="true"></span>
            {/if}
          </div>
          <Card elevation="glass" padding="sm">
            <div class="line-text markdown">
              {@html renderSafeMarkdown(entry.text)}
            </div>
          </Card>
        </div>
      {/each}
    </div>
  </div>
{/if}

<style>
  .live-transcript {
    position: fixed;
    bottom: var(--space-7);
    left: 50%;
    transform: translateX(-50%);
    width: min(720px, 90vw);
    max-height: 50vh;
    padding: var(--space-4);
    z-index: var(--z-toast);
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .live-transcript.recording {
    border-color: var(--accent-soft);
    box-shadow: var(--shadow-lg), 0 0 24px var(--accent-faint);
  }

  .transcript-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    overflow-y: auto;
    overscroll-behavior: contain;
    padding-right: var(--space-2);
    scrollbar-width: thin;
    scrollbar-color: var(--border-strong) transparent;
  }
  .transcript-list::-webkit-scrollbar { width: 6px; }
  .transcript-list::-webkit-scrollbar-thumb {
    background: var(--border-strong);
    border-radius: var(--radius-pill);
  }

  .transcript-line {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .transcript-line.is-partial .line-text {
    color: var(--text-muted);
    font-style: italic;
  }

  .line-meta {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .partial-pulse {
    display: inline-block;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent);
    box-shadow: 0 0 6px var(--accent-glow);
  }

  .line-text {
    color: var(--text);
    font-family: var(--font-sans);
    font-size: var(--size-md);
    line-height: var(--leading-normal);
    word-break: break-word;
  }
  .line-text.markdown :global(p) { margin: 0 0 6px; }
  .line-text.markdown :global(p:last-child) { margin-bottom: 0; }
  .line-text.markdown :global(strong) { color: var(--text); font-weight: var(--weight-semibold); }
  .line-text.markdown :global(em) { color: var(--text-muted); }
  .line-text.markdown :global(code) {
    font-family: var(--font-mono);
    font-size: 0.9em;
    padding: 1px 5px;
    border-radius: var(--radius-xs);
    background: var(--surface-3);
    color: var(--accent);
  }
  .line-text.markdown :global(pre) {
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-3);
    overflow-x: auto;
  }
  .line-text.markdown :global(ul),
  .line-text.markdown :global(ol) {
    margin: 4px 0 4px var(--space-5);
    padding: 0;
  }
  .line-text.markdown :global(a) {
    color: var(--accent);
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  @media (prefers-reduced-motion: reduce) {
    .transcript-line { animation: none !important; }
    .partial-pulse { animation: none !important; }
  }
</style>
