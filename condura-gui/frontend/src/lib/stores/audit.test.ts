import { beforeEach, describe, expect, it, vi } from 'vitest'

const { auditVerifyIntegrity, auditList, auditExport } = vi.hoisted(() => ({
  auditVerifyIntegrity: vi.fn(),
  auditList: vi.fn(),
  auditExport: vi.fn(),
}))

vi.mock('../ipc/client', () => ({
  ipc: {
    auditVerifyIntegrity,
    auditList,
    auditFacetCounts: vi.fn().mockRejectedValue(new Error('no facets')),
    auditExport,
    on: vi.fn(() => () => {}),
  },
}))

import { audit } from './audit.svelte'

describe('AuditStore.verifyIntegrity', () => {
  beforeEach(() => {
    auditVerifyIntegrity.mockReset()
    auditList.mockReset()
    audit.integrity = null
    audit.integrityError = null
    audit.integrityLoading = false
  })

  it('stores report on success', async () => {
    auditVerifyIntegrity.mockResolvedValue({
      ok: true,
      broken_at_id: null,
      reason: null,
      rows_verified: 12,
      rows_skipped: 0,
      duration_ms: 4,
    })
    await audit.verifyIntegrity()
    expect(audit.integrity?.ok).toBe(true)
    expect(audit.integrity?.rows_verified).toBe(12)
    expect(audit.integrityError).toBeNull()
    expect(audit.integrityLoading).toBe(false)
  })

  it('surfaces errors instead of silent null', async () => {
    auditVerifyIntegrity.mockRejectedValue(new Error('method not found'))
    await audit.verifyIntegrity()
    expect(audit.integrity).toBeNull()
    expect(audit.integrityError).toMatch(/method not found/i)
    expect(audit.integrityLoading).toBe(false)
  })

  it('exportChain records path and count', async () => {
    auditExport.mockResolvedValue({ path: '/tmp/condura-audit.jsonl', count: 7 })
    audit.exportResult = null
    audit.exportError = null
    await audit.exportChain()
    expect(auditExport).toHaveBeenCalled()
    expect(audit.exportResult).toEqual({ path: '/tmp/condura-audit.jsonl', count: 7 })
    expect(audit.exportError).toBeNull()
    expect(audit.exportInFlight).toBe(false)
  })

  it('appendLiveEvent prepends and stamps lastLiveAt', () => {
    audit.events = []
    audit.lastLiveAt = 0
    audit.appendLiveEvent({
      id: 42,
      ts: new Date().toISOString(),
      actor: 'user',
      action: 'test',
      app: '',
      level: 'info',
      result: 'allow',
      message: 'hi',
      this_hash: 'abc',
      prev_hash: '0'.repeat(64),
    })
    expect(audit.events).toHaveLength(1)
    expect(audit.events[0]?.id).toBe(42)
    expect(audit.events[0]?.this_hash).toBe('abc')
    expect(audit.lastLiveAt).toBeGreaterThan(0)
  })

  it('default filters include a real 24h since/until window', () => {
    expect(audit.filters.whenPreset).toBe('24h')
    expect(audit.filters.whenStart).toBeTruthy()
    expect(audit.filters.whenEnd).toBeTruthy()
    const start = new Date(audit.filters.whenStart!).getTime()
    const end = new Date(audit.filters.whenEnd!).getTime()
    expect(end - start).toBeGreaterThan(23 * 60 * 60 * 1000)
    expect(end - start).toBeLessThanOrEqual(24 * 60 * 60 * 1000 + 1000)
  })

  it('appendLiveEvent skips rows that fail active verdict filter', () => {
    audit.events = []
    audit.filters = {
      ...audit.filters,
      verdict: 'block',
    }
    audit.appendLiveEvent({
      id: 1,
      ts: new Date().toISOString(),
      actor: 'u',
      action: 'ok',
      app: '',
      level: 'info',
      result: 'allow',
      message: 'nope',
    })
    expect(audit.events).toHaveLength(0)
    audit.appendLiveEvent({
      id: 2,
      ts: new Date().toISOString(),
      actor: 'u',
      action: 'deny',
      app: '',
      level: 'info',
      result: 'block',
      message: 'yes',
    })
    expect(audit.events).toHaveLength(1)
    expect(audit.events[0]?.id).toBe(2)
  })
})
