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
    const c = await ipc.conversationsGet(id)
    this.currentID = c.id
    this.currentTitle = c.title
    this.messages = c.messages
  }

  async createNew(title?: string): Promise<ConversationMeta> {
    const c = await ipc.conversationsCreate({ title: title || 'New conversation' })
    this.conversations = [c, ...this.conversations]
    this.currentID = c.id
    this.currentTitle = c.title
    this.messages = []
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

    try {
      await ipc.llmStream({
        conversation_id: this.currentID,
        provider,
        request: {
          model,
          messages: this.messages,
          stream: true
        }
      })
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

  stopListening(): void {
    this.cleanups.forEach((c) => c())
    this.cleanups = []
  }
}

export const conversation = new ConversationStore()
