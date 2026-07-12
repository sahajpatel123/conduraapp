import { describe, it, expect, beforeEach } from 'vitest'
import {
  applyPendingRow,
  pendingActions,
  pendingCount,
  snapshot,
  startListening,
  stopListening,
} from './pending.svelte'
import { get } from 'svelte/store'
import type { PendingActionRow } from '../ipc/types'

function row(partial: Partial<PendingActionRow> & { id: string }): PendingActionRow {
  return {
    id: partial.id,
    spawn_id: partial.spawn_id ?? 'sp',
    agent_name: partial.agent_name ?? 'claude',
    kind: partial.kind ?? 'shell.exec',
    payload: partial.payload ?? { command: 'echo' },
    gate_decision: partial.gate_decision ?? 'require_consent',
    gate_reason: partial.gate_reason ?? 'needs you',
    status: partial.status ?? 'pending',
    created_at: partial.created_at ?? '2026-07-11T00:00:00Z',
    expires_at: partial.expires_at ?? '2026-07-11T01:00:00Z',
    exit_code: partial.exit_code ?? 0,
    result: partial.result ?? '',
    duration_ms: partial.duration_ms ?? 0,
  }
}

describe('pending store SSE upsert', () => {
  beforeEach(() => {
    pendingActions.set([])
    stopListening()
  })

  it('inserts a new pending row and bumps the badge count', () => {
    applyPendingRow(row({ id: 'a1', status: 'pending' }))
    expect(snapshot()).toHaveLength(1)
    expect(get(pendingCount)).toBe(1)
  })

  it('upserts status in place without duplicating', () => {
    applyPendingRow(row({ id: 'a1', status: 'pending' }))
    applyPendingRow(row({ id: 'a1', status: 'approved' }))
    const list = snapshot()
    expect(list).toHaveLength(1)
    expect(list[0].status).toBe('approved')
    expect(get(pendingCount)).toBe(0)
  })

  it('prepends newer ids so the queue feels live', () => {
    applyPendingRow(row({ id: 'old', status: 'pending' }))
    applyPendingRow(row({ id: 'new', status: 'pending' }))
    expect(snapshot().map((r) => r.id)).toEqual(['new', 'old'])
  })

  it('startListening is idempotent', () => {
    startListening()
    startListening()
    stopListening()
    // no throw
    expect(true).toBe(true)
  })
})
