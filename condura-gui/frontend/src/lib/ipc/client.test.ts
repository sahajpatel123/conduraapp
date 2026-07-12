import { describe, it, expect, vi, beforeEach } from 'vitest';

// client.test.ts — typed wrapper for daemon.capabilities.
//
// The contract we pin: the wrapper must call the JSON-RPC method
// `daemon.capabilities` and forward its return value untouched.
// The shape of the response is what the SettingsPane "Trust &
// safety" panel renders, so a future refactor that changes the
// method name or wraps the result would silently break the GUI's
// read-only disclosure of what the kill switch can and can't do.
//
// We mock the IPC client's `call` so the test does not need a
// running daemon. The mock is replaced per-test.

import { ipc } from './client';

describe('ipc.daemonCapabilities', () => {
  beforeEach(() => {
    // Reset the IPC client's connection state. The wrapper
    // requires start() to have run before call() succeeds, but
    // we mock the underlying `call` directly via a fetch stub
    // so we don't need a live listener.
    (ipc as unknown as { baseURL: string }).baseURL = 'http://127.0.0.1:0';
    (ipc as unknown as { authToken: string }).authToken = '';
  });

  it('calls the daemon.capabilities RPC method', async () => {
    const mockResponse = {
      version: { version: '0.1.0', commit: 'abc', build_date: '', go: '', platform: '' },
      kill_switch: {
        layer1_hotkey: true,
        layer2_watchdog: true,
        layer3_network_isolation: {
          in_process: true,
          os_process: false,
          deferred_to: 'v0.2.0',
          reference: 'CLAUDE.md §33.5.2 row C4.14',
        },
      },
      computer_use: {
        orax: 'stub',
        mac_cua: 'wrapper',
        macos_mcp: 'wrapper',
        vision_cua: 'disabled_default',
      },
      audit: { redaction: true, prune_tombstone: true, hmac_subkey: true },
    };
    const fetchMock = vi.fn(async () =>
      new Response(JSON.stringify({ jsonrpc: '2.0', result: mockResponse, id: 1 }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    );
    vi.stubGlobal('fetch', fetchMock);

    const result = await ipc.daemonCapabilities();

    // The wrapper passed the right method name and an empty params.
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [calledURL, calledInit] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(calledURL).toBe('http://127.0.0.1:0/api');
    const body = JSON.parse(String(calledInit.body));
    expect(body.method).toBe('daemon.capabilities');
    expect(body.params).toEqual({});

    // The result is forwarded unchanged.
    expect(result.kill_switch.layer3_network_isolation.in_process).toBe(true);
    expect(result.kill_switch.layer3_network_isolation.os_process).toBe(false);
    expect(result.kill_switch.layer3_network_isolation.deferred_to).toBe('v0.2.0');
    expect(result.audit.redaction).toBe(true);
    expect(result.audit.prune_tombstone).toBe(true);
    expect(result.audit.hmac_subkey).toBe(true);

    vi.unstubAllGlobals();
  });

  it('surfaces RPC errors as thrown Errors', async () => {
    const fetchMock = vi.fn(async () =>
      new Response(
        JSON.stringify({ jsonrpc: '2.0', error: { code: -32603, message: 'boom' }, id: 1 }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      )
    );
    vi.stubGlobal('fetch', fetchMock);

    await expect(ipc.daemonCapabilities()).rejects.toThrow(/boom/);
    vi.unstubAllGlobals();
  });
});

describe('ipc.permissionsStatus unwrap', () => {
  beforeEach(() => {
    (ipc as unknown as { baseURL: string }).baseURL = 'http://127.0.0.1:0'
    ;(ipc as unknown as { authToken: string }).authToken = ''
  })

  it('unwraps { platform, items } from the daemon into an array', async () => {
    const items = [
      { kind: 'accessibility', status: 'denied' },
      { kind: 'screen_recording', status: 'unknown' },
    ]
    const fetchMock = vi.fn(async () =>
      new Response(
        JSON.stringify({
          jsonrpc: '2.0',
          result: { platform: 'darwin', items },
          id: 1,
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )
    vi.stubGlobal('fetch', fetchMock)

    const res = await ipc.permissionsStatus()
    expect(Array.isArray(res)).toBe(true)
    expect(res).toHaveLength(2)
    expect(res[0].kind).toBe('accessibility')
    vi.unstubAllGlobals()
  })
})

describe('ipc.llmStream contract', () => {
  beforeEach(() => {
    (ipc as unknown as { baseURL: string }).baseURL = 'http://127.0.0.1:0';
    (ipc as unknown as { authToken: string }).authToken = '';
  });

  it('returns request_id from the daemon llm.stream RPC', async () => {
    const fetchMock = vi.fn(async () =>
      new Response(
        JSON.stringify({
          jsonrpc: '2.0',
          result: { request_id: '01HXYZ', conversation_id: 42 },
          id: 1,
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      )
    );
    vi.stubGlobal('fetch', fetchMock);

    const res = await ipc.llmStream({
      conversation_id: 42,
      provider: 'ollama',
      request: { model: 'llama3', messages: [], stream: true },
    });

    const body = JSON.parse(String((fetchMock.mock.calls[0] as [string, RequestInit])[1].body));
    expect(body.method).toBe('llm.stream');
    expect(res.request_id).toBe('01HXYZ');
    expect(res.conversation_id).toBe(42);
    vi.unstubAllGlobals();
  });
});

describe('ipc safety.consent.request remap', () => {
  it('exposes consent on EventMap (compile-time shape)', () => {
    // Runtime EventSource remaps safety.consent.request → consent;
    // ensure the typed emitter accepts the handler shape.
    const off = ipc.on('consent', (t) => {
      expect(t.nonce).toBeDefined()
    })
    off()
  })
})

describe('ipc.presenceState contract', () => {
  beforeEach(() => {
    ;(ipc as unknown as { baseURL: string }).baseURL = 'http://127.0.0.1:0'
    ;(ipc as unknown as { authToken: string }).authToken = ''
  })

  it('calls presence.state and returns { state }', async () => {
    const fetchMock = vi.fn(async () =>
      new Response(
        JSON.stringify({ jsonrpc: '2.0', result: { state: 'hidden' }, id: 1 }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      )
    )
    vi.stubGlobal('fetch', fetchMock)
    const res = await ipc.presenceState()
    const body = JSON.parse(String((fetchMock.mock.calls[0]![1] as RequestInit).body))
    expect(body.method).toBe('presence.state')
    expect(res.state).toBe('hidden')
    vi.unstubAllGlobals()
  })
})

describe('ipc.delegateListSpawns contract', () => {
  beforeEach(() => {
    ;(ipc as unknown as { baseURL: string }).baseURL = 'http://127.0.0.1:0'
    ;(ipc as unknown as { authToken: string }).authToken = ''
  })

  it('calls delegate.list_spawns', async () => {
    const fetchMock = vi.fn(async () =>
      new Response(
        JSON.stringify({ jsonrpc: '2.0', result: { running: [] }, id: 1 }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      )
    )
    vi.stubGlobal('fetch', fetchMock)
    const res = await ipc.delegateListSpawns()
    const body = JSON.parse(String((fetchMock.mock.calls[0]![1] as RequestInit).body))
    expect(body.method).toBe('delegate.list_spawns')
    expect(res.running).toEqual([])
    vi.unstubAllGlobals()
  })
})

describe('ipc.delegatePending* contract', () => {
  beforeEach(() => {
    ;(ipc as unknown as { baseURL: string }).baseURL = 'http://127.0.0.1:0'
    ;(ipc as unknown as { authToken: string }).authToken = ''
  })

  function mockRPC(result: unknown) {
    const fetchMock = vi.fn(async () =>
      new Response(JSON.stringify({ jsonrpc: '2.0', result, id: 1 }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    )
    vi.stubGlobal('fetch', fetchMock)
    return fetchMock
  }

  function lastBody(fetchMock: ReturnType<typeof vi.fn>): { method: string; params: unknown } {
    const init = fetchMock.mock.calls[0]![1] as RequestInit
    return JSON.parse(String(init.body)) as { method: string; params: unknown }
  }

  it('lists via delegate.pending.list', async () => {
    const row = {
      id: 'pa-1',
      spawn_id: 'sp-1',
      agent_name: 'claude',
      kind: 'shell.exec',
      payload: { command: 'echo hi' },
      gate_decision: 'require_consent',
      gate_reason: 'shell',
      status: 'pending',
      created_at: '2026-07-11T00:00:00Z',
      expires_at: '2026-07-11T01:00:00Z',
      exit_code: 0,
      result: '',
      duration_ms: 0,
    }
    const fetchMock = mockRPC({ actions: [row] })
    const res = await ipc.delegatePendingList('pending')
    const body = lastBody(fetchMock)
    expect(body.method).toBe('delegate.pending.list')
    expect(body.params).toEqual({ status: 'pending', limit: 50 })
    expect(res.actions).toHaveLength(1)
    expect(res.actions[0].id).toBe('pa-1')
    vi.unstubAllGlobals()
  })

  it('decides via delegate.pending.decide', async () => {
    const fetchMock = mockRPC({ id: 'pa-1', status: 'approved' })
    const res = await ipc.delegatePendingDecide({
      id: 'pa-1',
      decision: 'approve',
      decided_by: 'user:anonymous',
      note: '',
      auto_run: true,
    })
    const body = lastBody(fetchMock)
    expect(body.method).toBe('delegate.pending.decide')
    expect(body.params).toMatchObject({
      id: 'pa-1',
      decision: 'approve',
      auto_run: true,
    })
    expect(res.status).toBe('approved')
    vi.unstubAllGlobals()
  })

  it('executes via delegate.pending.execute', async () => {
    const fetchMock = mockRPC({ id: 'pa-1', status: 'executed', exit_code: 0 })
    const res = await ipc.delegatePendingExecute({ id: 'pa-1' })
    const body = lastBody(fetchMock)
    expect(body.method).toBe('delegate.pending.execute')
    expect(body.params).toEqual({ id: 'pa-1' })
    expect(res.status).toBe('executed')
    vi.unstubAllGlobals()
  })
})
