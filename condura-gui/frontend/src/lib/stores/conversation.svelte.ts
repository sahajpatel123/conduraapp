// Conversation store. Tracks the current conversation + its
// messages + the in-flight stream. Backed by the daemon's
// conversation store (SQLite + AES-256-GCM).

import { ipc } from '../ipc/client'
import type { Conversation, ConversationMeta, Message, StreamEvent, ToolCall } from '../ipc/types'

/** Idle timeout for in-flight streams when no delta arrives (daemon crash / SSE drop). */
export const STREAM_IDLE_MS = 45_000

class ConversationStore {
  // List of conversations (sidebar).
  conversations = $state<ConversationMeta[]>([])
  // Currently-open conversation.
  currentID = $state<number>(0)
  currentTitle = $state<string>('New conversation')
  messages = $state<Message[]>([])

  // Streaming state.
  isStreaming = $state<boolean>(false)
  streamingDelta = $state<string>('')
  streamingError = $state<string>('')
  // Tool calls surfaced by the assistant during the in-flight
  // stream. Merged into the persisted assistant message on
  // Done. Tool calls and text content arrive in separate
  // SSE events; we accumulate the calls here so the UI can
  // show them alongside the streamed text.
  streamingToolCalls = $state<ToolCall[]>([])

  // request_id of the in-flight stream, captured from llmStream()'s
  // return value. Used to filter SSE stream events so a late event
  // from a previous stream (or a concurrent stream on another
  // conversation) can't contaminate this one. Reset to '' when the
  // stream finishes or is cancelled.
  currentRequestId = $state<string>('')

  private cleanups: Array<() => void> = []
  /** Chunk buffer to avoid O(n²) string growth on every token. */
  private deltaParts: string[] = []
  private streamWatchdog: ReturnType<typeof setTimeout> | null = null

  async refreshList(): Promise<void> {
    try {
      const list = await ipc.conversationsList()
      // Daemon may return null for an empty list — never leave the store null
      // (MeridianChat does conversations.slice on every render).
      this.conversations = Array.isArray(list) ? list : []
    } catch (err) {
      // ignore — daemon might not be up yet
      // eslint-disable-next-line no-console
      console.warn('conversationsList failed', err)
    }
  }

  async open(id: number): Promise<void> {
    // Cancel any active stream on the current conversation before
    // switching. Otherwise the old stream's events are filtered by
    // conversation_id and the assistant reply is lost forever, and
    // isStreaming stays true, locking the UI.
    await this.cancelActive()
    const c = await ipc.conversationsGet(id)
    this.currentID = c.id
    this.currentTitle = c.title
    this.messages = c.messages
    this.clearStreamingState()
  }

  async createNew(title?: string): Promise<ConversationMeta> {
    await this.cancelActive()
    const c = await ipc.conversationsCreate({ title: title || 'New conversation' })
    this.conversations = [c, ...this.conversations]
    this.currentID = c.id
    this.currentTitle = c.title
    this.messages = []
    this.clearStreamingState()
    return c
  }

  /**
   * Deletes the open conversation, cancels its stream, then opens
   * the next recent thread (or clears to a blank desk).
   */
  async deleteCurrent(): Promise<void> {
    if (!this.currentID) {
      return
    }
    await this.cancelActive()
    const id = this.currentID
    await ipc.conversationsDelete(id)
    this.conversations = this.conversations.filter((c) => c.id !== id)
    if (this.conversations.length > 0) {
      await this.open(this.conversations[0]!.id)
      return
    }
    this.currentID = 0
    this.currentTitle = 'New conversation'
    this.messages = []
    this.clearStreamingState()
  }

  /**
   * deleteById removes a conversation by id. If it is the open
   * thread, behaves like deleteCurrent (cancel + open next).
   * Daemon cancel-by-conversation still runs on conversations.delete.
   */
  async deleteById(id: number): Promise<void> {
    if (!id) {
      return
    }
    if (id === this.currentID) {
      await this.deleteCurrent()
      return
    }
    await ipc.conversationsDelete(id)
    this.conversations = this.conversations.filter((c) => c.id !== id)
  }

  /**
   * Renames a conversation. Updates the open title when id is current
   * and reorders the sidebar list by updated_at (server returns meta).
   */
  async rename(id: number, title: string): Promise<void> {
    if (!id) return
    const trimmed = title.trim()
    const meta = await ipc.conversationsRename(id, trimmed)
    if (id === this.currentID) {
      this.currentTitle = meta.title
    }
    const rest = this.conversations.filter((c) => c.id !== id)
    this.conversations = [meta, ...rest]
  }

  /**
   * Send a user message; start streaming the assistant reply.
   * Subscribes to SSE stream events; on Done, persists the
   * assistant's full reply via conversations.append.
   */
  async send(
    provider: string,
    model: string,
    userText: string,
    opts?: { skillSystem?: string }
  ): Promise<void> {
    if (!this.currentID) {
      await this.createNew('New conversation')
    }
    const userMsg: Message = { role: 'user', content: userText }
    this.messages = [...this.messages, userMsg]
    await ipc.conversationsAppend({ id: this.currentID, message: userMsg })

    this.deltaParts = []
    this.streamingDelta = ''
    this.streamingError = ''
    this.streamingToolCalls = []
    this.isStreaming = true
    this.currentRequestId = ''
    this.armStreamWatchdog()

    // Skill context is request-only (not persisted as a chat turn) so the
    // transcript stays human-readable while the model sees the procedure.
    const requestMessages: Message[] = opts?.skillSystem
      ? [
          ...this.messages.slice(0, -1),
          { role: 'system', content: opts.skillSystem },
          userMsg,
        ]
      : this.messages

    try {
      const res = await ipc.llmStream({
        conversation_id: this.currentID,
        provider,
        request: {
          model,
          messages: requestMessages,
          stream: true
        }
      })
      this.currentRequestId = res.request_id ?? ''
      this.armStreamWatchdog()
    } catch (err) {
      this.clearStreamWatchdog()
      this.isStreaming = false
      this.streamingError = String(err)
    }
  }

  async cancel(): Promise<void> {
    if (!this.currentID) {
      return
    }
    this.clearStreamWatchdog()
    await ipc.llmCancel({ conversation_id: this.currentID })
    this.isStreaming = false
    this.currentRequestId = ''
  }

  /**
   * After SSE reconnect: pull sidebar + open thread from the daemon.
   * Does not cancel server-side streams; clears local streaming chrome
   * so a completed reply that finished while we were offline appears.
   * Critical for 24/7 Meridian use across Wi‑Fi blips / daemon restarts.
   */
  async resyncFromDaemon(): Promise<void> {
    await this.refreshList()
    if (!this.currentID) {
      return
    }
    try {
      const c = await ipc.conversationsGet(this.currentID)
      this.currentTitle = c.title
      this.messages = c.messages ?? []
      // Local stream state is stale after reconnect (deltas were lost).
      this.clearStreamingState()
    } catch {
      // Conversation may have been deleted server-side.
    }
  }

  startListening(): void {
    // Idempotent: Meridian init + any legacy route may both call this.
    if (this.cleanups.length > 0) return
    this.cleanups.push(
      ipc.on('stream', (ev: StreamEvent) => {
        if (ev.conversation_id !== this.currentID) {
          return
        }
        // Cross-stream isolation: if both the event and the store
        // carry a request_id and they disagree, the event belongs to
        // a stale or concurrent stream — skip it. This prevents a
        // previous stream's tail from leaking into a new send, and
        // stops a concurrent stream on the same conversation from
        // interleaving deltas.
        if (ev.request_id && this.currentRequestId && ev.request_id !== this.currentRequestId) {
          return
        }
        if (ev.err) {
          this.clearStreamWatchdog()
          this.streamingError = ev.err
          this.isStreaming = false
          this.currentRequestId = ''
          return
        }
        if (ev.done) {
          this.clearStreamWatchdog()
          // Persist the assistant message and reset streaming state.
          const content = this.deltaParts.length > 0
            ? this.deltaParts.join('')
            : this.streamingDelta
          const assistantMsg: Message = {
            role: 'assistant',
            content,
            // Attach tool calls to the persisted message so a
            // page reload shows them in context. Skip the
            // field entirely when no calls were made.
            ...(this.streamingToolCalls.length > 0
              ? { tool_calls: this.streamingToolCalls }
              : {})
          }
          this.messages = [...this.messages, assistantMsg]
          void ipc.conversationsAppend({
            id: this.currentID,
            message: assistantMsg
          })
          this.deltaParts = []
          this.streamingDelta = ''
          this.streamingToolCalls = []
          this.isStreaming = false
          this.currentRequestId = ''
          // Refresh sidebar so updated_at moves to the top.
          void this.refreshList()
          return
        }
        if (ev.delta) {
          this.deltaParts.push(ev.delta)
          this.streamingDelta = this.deltaParts.join('')
          this.armStreamWatchdog()
        }
        if (ev.tool_calls && ev.tool_calls.length > 0) {
          // Merge new tool calls with any we already saw.
          // The daemon streams them as complete entries (not
          // incremental args), so a simple append-by-id is
          // safe — same id won't appear twice in one stream.
          const existing = new Map(
            this.streamingToolCalls.map((tc) => [tc.id, tc])
          )
          for (const tc of ev.tool_calls) {
            existing.set(tc.id, tc)
          }
          this.streamingToolCalls = Array.from(existing.values())
          this.armStreamWatchdog()
        }
      })
    )
    this.cleanups.push(
      ipc.on('disconnected', () => {
        if (!this.isStreaming) return
        this.clearStreamWatchdog()
        this.streamingError = this.streamingError || 'Connection lost during stream'
        this.isStreaming = false
        this.currentRequestId = ''
      })
    )
  }

  /**
   * Cancel any active stream on the current conversation and reset
   * all streaming state. Safe to call when no stream is active
   * (no-op). Called before switching conversations to prevent
   * the orphan-stream bug where isStreaming stays true and the UI
   * locks up.
   */
  private async cancelActive(): Promise<void> {
    if (!this.isStreaming) return
    try {
      await this.cancel()
    } catch {
      // best-effort; the daemon will eventually clean up stale streams
    }
  }

  private armStreamWatchdog(): void {
    this.clearStreamWatchdog()
    this.streamWatchdog = setTimeout(() => {
      if (!this.isStreaming) return
      this.streamingError = this.streamingError || 'Stream stalled — no data received'
      this.isStreaming = false
      this.currentRequestId = ''
      void ipc.llmCancel({ conversation_id: this.currentID }).catch(() => {})
    }, STREAM_IDLE_MS)
  }

  private clearStreamWatchdog(): void {
    if (this.streamWatchdog != null) {
      clearTimeout(this.streamWatchdog)
      this.streamWatchdog = null
    }
  }

  private clearStreamingState(): void {
    this.clearStreamWatchdog()
    this.isStreaming = false
    this.deltaParts = []
    this.streamingDelta = ''
    this.streamingError = ''
    this.streamingToolCalls = []
    this.currentRequestId = ''
  }

  stopListening(): void {
    this.clearStreamWatchdog()
    this.cleanups.forEach((c) => c())
    this.cleanups = []
  }
}

export const conversation = new ConversationStore()
