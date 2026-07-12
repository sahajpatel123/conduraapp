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
  conversationsRename: vi.fn().mockImplementation(async (id: number, title: string) => ({
    id,
    title,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    message_count: 0,
  })),
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

  it('deleteCurrent cancels stream, deletes, and opens next thread', async () => {
    conversation.currentID = 1
    conversation.isStreaming = true
    conversation.currentRequestId = 'req-1'
    conversation.conversations = [
      { id: 1, title: 'A', created_at: '', updated_at: '', message_count: 1 },
      { id: 2, title: 'B', created_at: '', updated_at: '', message_count: 0 },
    ]
    ipcMock.conversationsGet.mockResolvedValueOnce({
      id: 2,
      title: 'B',
      messages: [{ role: 'user', content: 'hi from b' }],
    })
    await conversation.deleteCurrent()
    expect(ipcMock.llmCancel).toHaveBeenCalled()
    expect(ipcMock.conversationsDelete).toHaveBeenCalledWith(1)
    expect(conversation.currentID).toBe(2)
    expect(conversation.messages[0]?.content).toBe('hi from b')
    expect(conversation.isStreaming).toBe(false)
  })

  it('deleteById on non-current only removes from list', async () => {
    conversation.currentID = 1
    conversation.conversations = [
      { id: 1, title: 'A', created_at: '', updated_at: '', message_count: 1 },
      { id: 2, title: 'B', created_at: '', updated_at: '', message_count: 0 },
    ]
    await conversation.deleteById(2)
    expect(ipcMock.conversationsDelete).toHaveBeenCalledWith(2)
    expect(conversation.currentID).toBe(1)
    expect(conversation.conversations.map((c) => c.id)).toEqual([1])
  })

  it('rename updates current title and list head', async () => {
    conversation.currentID = 1
    conversation.currentTitle = 'Old'
    conversation.conversations = [
      { id: 1, title: 'Old', created_at: '', updated_at: '', message_count: 1 },
      { id: 2, title: 'Other', created_at: '', updated_at: '', message_count: 0 },
    ]
    await conversation.rename(1, '  Renamed  ')
    expect(ipcMock.conversationsRename).toHaveBeenCalledWith(1, 'Renamed')
    expect(conversation.currentTitle).toBe('Renamed')
    expect(conversation.conversations[0]?.title).toBe('Renamed')
    expect(conversation.conversations[0]?.id).toBe(1)
  })

  it('resyncFromDaemon reloads messages and clears stale streaming chrome', async () => {
    conversation.isStreaming = true
    conversation.streamingDelta = 'partial…'
    conversation.streamingError = 'Connection lost during stream'
    conversation.currentRequestId = 'req-stale'
    ipcMock.conversationsList.mockResolvedValueOnce([
      { id: 1, title: 'Test', created_at: '', updated_at: '' },
    ])
    ipcMock.conversationsGet.mockResolvedValueOnce({
      id: 1,
      title: 'Test',
      messages: [
        { role: 'user', content: 'hello' },
        { role: 'assistant', content: 'full reply from daemon' },
      ],
    })

    await conversation.resyncFromDaemon()

    expect(ipcMock.conversationsGet).toHaveBeenCalledWith(1)
    expect(conversation.messages).toHaveLength(2)
    expect(conversation.messages[1].content).toBe('full reply from daemon')
    expect(conversation.isStreaming).toBe(false)
    expect(conversation.streamingDelta).toBe('')
    expect(conversation.streamingError).toBe('')
    expect(conversation.currentRequestId).toBe('')
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
