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
