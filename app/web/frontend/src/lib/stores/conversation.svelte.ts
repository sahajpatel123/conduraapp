// Conversation store. Tracks the current conversation + its
// messages + the in-flight stream. Backed by the daemon's
// conversation store (SQLite + AES-256-GCM).

import { ipc } from '../ipc/client'
import type { Conversation, ConversationMeta, Message, StreamEvent, ToolCall } from '../ipc/types'

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

  async refreshList(): Promise<void> {
    try {
      this.conversations = await ipc.conversationsList()
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

  async deleteCurrent(): Promise<void> {
    if (!this.currentID) {
      return
    }
    await ipc.conversationsDelete(this.currentID)
    this.conversations = this.conversations.filter((c) => c.id !== this.currentID)
    this.currentID = 0
    this.currentTitle = 'New conversation'
    this.messages = []
  }

  /**
   * deleteById removes the conversation with the given id WITHOUT
   * touching the currently-open conversation. Used by the Sidebar
   * undo-delete flow: the timer must delete the conversation the user
   * actually clicked on, not whatever conversation is current when
   * the timer fires (which could be a different one the user opened
   * in the undo window).
   */
  async deleteById(id: number): Promise<void> {
    if (!id) {
      return
    }
    await ipc.conversationsDelete(id)
    this.conversations = this.conversations.filter((c) => c.id !== id)
  }

  /**
   * Send a user message; start streaming the assistant reply.
   * Subscribes to SSE stream events; on Done, persists the
   * assistant's full reply via conversations.append.
   */
  async send(provider: string, model: string, userText: string): Promise<void> {
    if (!this.currentID) {
      await this.createNew('New conversation')
    }
    const userMsg: Message = { role: 'user', content: userText }
    this.messages = [...this.messages, userMsg]
    await ipc.conversationsAppend({ id: this.currentID, message: userMsg })

    this.streamingDelta = ''
    this.streamingError = ''
    this.streamingToolCalls = []
    this.isStreaming = true
    this.currentRequestId = ''

    try {
      const res = await ipc.llmStream({
        conversation_id: this.currentID,
        provider,
        request: {
          model,
          messages: this.messages,
          stream: true
        }
      })
      this.currentRequestId = res.request_id
    } catch (err) {
      this.isStreaming = false
      this.streamingError = String(err)
    }
  }

  async cancel(): Promise<void> {
    if (!this.currentID) {
      return
    }
    await ipc.llmCancel({ conversation_id: this.currentID })
    this.isStreaming = false
  }

  startListening(): void {
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
          this.streamingError = ev.err
          this.isStreaming = false
          return
        }
        if (ev.done) {
          // Persist the assistant message and reset streaming state.
          const assistantMsg: Message = {
            role: 'assistant',
            content: this.streamingDelta,
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
          this.streamingDelta = ''
          this.streamingToolCalls = []
          this.isStreaming = false
          this.currentRequestId = ''
          // Refresh sidebar so updated_at moves to the top.
          void this.refreshList()
          return
        }
        if (ev.delta) {
          this.streamingDelta += ev.delta
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
        }
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

  private clearStreamingState(): void {
    this.isStreaming = false
    this.streamingDelta = ''
    this.streamingError = ''
    this.streamingToolCalls = []
    this.currentRequestId = ''
  }

  stopListening(): void {
    this.cleanups.forEach((c) => c())
    this.cleanups = []
  }
}

export const conversation = new ConversationStore()
