import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// conversation.svelte.ts uses Svelte 5 runes ($state). Vitest loads the
// browser build via resolve.conditions=['browser'], so the class is usable
// as long as we mock the IPC client.

const handlers = new Map<string, Array<(...args: unknown[]) => void>>()

const ipcMock = {
  conversationsList: vi.fn().mockResolvedValue([]),
  conversationsGet: vi.fn(),
  conversationsCreate: vi.fn().mockResolvedValue({
    id: 1,
    title: 'New conversation',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }),
  conversationsDelete: vi.fn().mockResolvedValue(undefined),
  conversationsAppend: vi.fn().mockResolvedValue(undefined),
  llmStream: vi.fn().mockResolvedValue({ request_id: 'req-1', conversation_id: 1 }),
  llmCancel: vi.fn().mockResolvedValue(undefined),
  on: vi.fn((event: string, handler: (...args: unknown[]) => void) => {
    const list = handlers.get(event) ?? []
    list.push(handler)
    handlers.set(event, list)
    return () => {
      const next = (handlers.get(event) ?? []).filter((h) => h !== handler)
      handlers.set(event, next)
    }
  }),
}

vi.mock('../ipc/client', () => ({ ipc: ipcMock }))

// Import after mock
const { conversation, STREAM_IDLE_MS } = await import('./conversation.svelte')

function emitStream(ev: Record<string, unknown>) {
  for (const h of handlers.get('stream') ?? []) h(ev)
}
function emitDisconnected(reason = 'sse-error') {
  for (const h of handlers.get('disconnected') ?? []) h(reason)
}

describe('ConversationStore streaming', () => {
  beforeEach(() => {
    handlers.clear()
    vi.clearAllMocks()
    ipcMock.llmStream.mockResolvedValue({ request_id: 'req-1', conversation_id: 1 })
    conversation.stopListening()
    conversation.startListening()
    // reset store fields
    conversation.currentID = 1
    conversation.currentTitle = 'Test'
    conversation.messages = []
    conversation.isStreaming = false
    conversation.streamingDelta = ''
    conversation.streamingError = ''
    conversation.streamingToolCalls = []
    conversation.currentRequestId = ''
  })

  afterEach(() => {
    conversation.stopListening()
    vi.useRealTimers()
  })

  it('captures request_id from llmStream and filters foreign streams', async () => {
    await conversation.send('ollama', 'llama3', 'hello')
    expect(ipcMock.llmStream).toHaveBeenCalled()
    expect(conversation.currentRequestId).toBe('req-1')
    expect(conversation.isStreaming).toBe(true)

    emitStream({
      conversation_id: 1,
      request_id: 'other',
      delta: 'nope',
      done: false,
    })
    expect(conversation.streamingDelta).toBe('')

    emitStream({
      conversation_id: 1,
      request_id: 'req-1',
      delta: 'hi',
      done: false,
    })
    expect(conversation.streamingDelta).toBe('hi')
  })

  it('buffers deltas without losing content and finalizes on done', async () => {
    await conversation.send('ollama', 'llama3', 'hello')
    emitStream({ conversation_id: 1, request_id: 'req-1', delta: 'a', done: false })
    emitStream({ conversation_id: 1, request_id: 'req-1', delta: 'b', done: false })
    emitStream({ conversation_id: 1, request_id: 'req-1', delta: 'c', done: false })
    expect(conversation.streamingDelta).toBe('abc')

    emitStream({ conversation_id: 1, request_id: 'req-1', delta: '', done: true })
    expect(conversation.isStreaming).toBe(false)
    expect(conversation.messages.some((m) => m.role === 'assistant' && m.content === 'abc')).toBe(
      true,
    )
    expect(conversation.currentRequestId).toBe('')
  })

  it('clears isStreaming on disconnected while streaming', async () => {
    await conversation.send('ollama', 'llama3', 'hello')
    expect(conversation.isStreaming).toBe(true)
    emitDisconnected('sse-error')
    expect(conversation.isStreaming).toBe(false)
    expect(conversation.streamingError).toMatch(/Connection lost/i)
  })

  it('clears isStreaming after idle watchdog timeout', async () => {
    vi.useFakeTimers()
    await conversation.send('ollama', 'llama3', 'hello')
    expect(conversation.isStreaming).toBe(true)
    vi.advanceTimersByTime(STREAM_IDLE_MS + 10)
    expect(conversation.isStreaming).toBe(false)
    expect(conversation.streamingError).toMatch(/stalled/i)
  })
})
