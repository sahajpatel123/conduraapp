// Audit log store. Loads events with rich filters (When/What/App/Model);
// supports pagination, live SSE append, and chain-integrity verification.
//
// Phase 15 (SCREEN_AUDIT.md):
//   - 4 filter groups: When (date), What (blast class + verdict),
//     App (checkbox), Model (checkbox).
//   - Search across actor/action/message/app/path/url.
//   - Live entries: the SSE channel pushes new AuditEvents; the route
//     inserts them at the top of the timeline.
//   - Chain integrity: verifyIntegrity walks the HMAC chain and
//     reports broken_at_id. Read-only — never edits the log.
//   - Export: writes JSONL to disk through the Gatekeeper (consent).
//     Returns { path, count }.

import { ipc } from '../ipc/client'
import type {
  AuditEvent,
  AuditFacetCounts,
  AuditIntegrityReport,
  AuditListParams,
  BlastClass,
  Verdict,
} from '../ipc/types'

export type VerdictFilter = 'all' | Verdict
export type WhenPreset = '1h' | '24h' | '7d' | '30d' | 'all'

export interface AuditFilters {
  search: string
  whenPreset: WhenPreset
  whenStart: string | null
  whenEnd: string | null
  blastClasses: Set<BlastClass>
  verdict: VerdictFilter
  apps: Set<string>
  models: Set<string>
}

export const DEFAULT_FILTERS: AuditFilters = {
  search: '',
  whenPreset: '24h',
  whenStart: null,
  whenEnd: null,
  blastClasses: new Set<BlastClass>(['READ', 'WRITE', 'NETWORK', 'DESTRUCTIVE']),
  verdict: 'all',
  apps: new Set<string>(),
  models: new Set<string>(),
}

const BLAST_CLASSES: BlastClass[] = ['READ', 'WRITE', 'NETWORK', 'DESTRUCTIVE']

// Derive a BlastClass from an existing AuditEvent when the server hasn't
// enriched the row with the new `blast_class` field. This is the v0.1.0
// fallback so the route renders a meaningful badge before the server-side
// enrichment lands.
export function deriveBlastClass(ev: AuditEvent): BlastClass {
  if (ev.blast_class) return ev.blast_class
  const action = (ev.action || '').toLowerCase()
  if (action.startsWith('shell.exec') || action.startsWith('file')) return 'WRITE'
  if (action.startsWith('http') || action.startsWith('network') || action.includes('send')) {
    return action.includes('http') ? 'NETWORK' : 'NETWORK'
  }
  if (ev.result === 'block' || ev.level === 'error') return 'DESTRUCTIVE'
  return 'READ'
}

// Same idea for verdict — derive from `result` and `level` when the
// enriched field is absent.
export function deriveVerdict(ev: AuditEvent): Verdict {
  if (ev.verdict) return ev.verdict
  if (ev.result === 'allow') return 'allow'
  if (ev.result === 'block') return 'block'
  if (ev.result === 'prompt') return 'prompt'
  return ev.level === 'error' ? 'error' : 'allow'
}

class AuditStore {
  events = $state<AuditEvent[]>([])
  loading = $state<boolean>(false)
  error = $state<string | null>(null)
  filters = $state<AuditFilters>(structuredClone(DEFAULT_FILTERS))
  facetCounts = $state<AuditFacetCounts | null>(null)
  integrity = $state<AuditIntegrityReport | null>(null)
  integrityLoading = $state<boolean>(false)
  exportInFlight = $state<boolean>(false)
  exportResult = $state<{ path: string; count: number } | null>(null)
  exportError = $state<string | null>(null)
  limit = $state<number>(100)
  hasMore = $state<boolean>(true)

  // The SSE subscription handle. We tear it down on filter-set so
  // re-subscriptions land once.
  private sseOff: (() => void) | null = null

  // ── list & paginate ─────────────────────────────────────────
  async refresh(): Promise<void> {
    this.loading = true
    this.error = null
    try {
      const params = this.buildListParams(0, this.limit)
      const events = await ipc.auditList(params)
      this.events = events
      this.hasMore = events.length >= this.limit
      this.deriveFacetCounts()
      await this.refreshFacets()
    } catch (e) {
      this.error = String(e)
    } finally {
      this.loading = false
    }
  }

  async loadMore(): Promise<void> {
    if (this.loading || !this.hasMore) return
    this.loading = true
    try {
      const params = this.buildListParams(this.events.length, this.limit)
      const more = await ipc.auditList(params)
      this.events = [...this.events, ...more]
      this.hasMore = more.length >= this.limit
    } catch (e) {
      this.error = String(e)
    } finally {
      this.loading = false
    }
  }

  private buildListParams(offset: number, limit: number): AuditListParams {
    const f = this.filters
    const params: AuditListParams = { limit, offset }
    if (f.search) params.search = f.search
    if (f.whenStart) params.since = f.whenStart
    if (f.whenEnd) params.until = f.whenEnd
    if (f.blastClasses.size > 0 && f.blastClasses.size < BLAST_CLASSES.length) {
      params.blast_classes = Array.from(f.blastClasses) as BlastClass[]
    }
    if (f.verdict !== 'all') params.verdict = f.verdict
    if (f.apps.size > 0) params.apps = Array.from(f.apps)
    if (f.models.size > 0) params.models = Array.from(f.models)
    return params
  }

  // ── facet counts ────────────────────────────────────────────
  private async refreshFacets(): Promise<void> {
    try {
      const counts = await ipc.auditFacetCounts(this.buildListParams(0, 1000))
      this.facetCounts = counts
    } catch {
      // RPC may not be implemented yet — the derived fallback below fills in.
      this.deriveFacetCounts()
    }
  }

  // Client-side derivation: cheap, always-available fallback.
  private deriveFacetCounts(): void {
    const apps = new Map<string, number>()
    const models = new Map<string, number>()
    const classes: Record<BlastClass, number> = { READ: 0, WRITE: 0, NETWORK: 0, DESTRUCTIVE: 0 }
    const verdicts: Record<Verdict, number> = { allow: 0, block: 0, prompt: 0, error: 0 }
    for (const ev of this.events) {
      if (ev.app) apps.set(ev.app, (apps.get(ev.app) ?? 0) + 1)
      if (ev.model) models.set(ev.model, (models.get(ev.model) ?? 0) + 1)
      classes[deriveBlastClass(ev)]++
      verdicts[deriveVerdict(ev)]++
    }
    const toList = (m: Map<string, number>) =>
      Array.from(m.entries())
        .map(([name, count]) => ({ name, count }))
        .sort((a, b) => b.count - a.count)
    this.facetCounts = {
      apps: toList(apps),
      models: toList(models),
      blast_classes: classes,
      verdicts,
      total: this.events.length,
    }
  }

  // ── chain integrity ─────────────────────────────────────────
  async verifyIntegrity(): Promise<void> {
    this.integrityLoading = true
    try {
      const params = this.buildListParams(0, 10000)
      const report = await ipc.auditVerifyIntegrity(params)
      this.integrity = report
    } catch {
      // RPC may not be implemented; surface a soft "unknown" status so the
      // chain badge still reads honestly.
      this.integrity = null
    } finally {
      this.integrityLoading = false
    }
  }

  // ── export (consent-gated; the route triggers global ConsentModal) ──
  async exportChain(): Promise<void> {
    if (this.exportInFlight) return
    this.exportInFlight = true
    this.exportError = null
    this.exportResult = null
    try {
      const params = this.buildListParams(0, 100000)
      const res = await ipc.auditExport(params)
      this.exportResult = res
    } catch (e) {
      this.exportError = String(e)
    } finally {
      this.exportInFlight = false
    }
  }

  clearExportResult(): void {
    this.exportResult = null
    this.exportError = null
  }

  // ── filter mutations ────────────────────────────────────────
  setSearch(v: string): void {
    this.filters = { ...this.filters, search: v }
    void this.refresh()
  }
  setWhenPreset(p: WhenPreset): void {
    this.filters = { ...this.filters, whenPreset: p }
    const now = new Date()
    let start: Date | null = null
    switch (p) {
      case '1h':
        start = new Date(now.getTime() - 60 * 60 * 1000)
        break
      case '24h':
        start = new Date(now.getTime() - 24 * 60 * 60 * 1000)
        break
      case '7d':
        start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
        break
      case '30d':
        start = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
        break
      case 'all':
        start = null
        break
    }
    this.filters = {
      ...this.filters,
      whenPreset: p,
      whenStart: start?.toISOString() ?? null,
      whenEnd: p === 'all' ? null : now.toISOString(),
    }
    void this.refresh()
  }
  toggleBlastClass(c: BlastClass): void {
    const next = new Set(this.filters.blastClasses)
    if (next.has(c)) next.delete(c)
    else next.add(c)
    this.filters = { ...this.filters, blastClasses: next }
    void this.refresh()
  }
  setVerdict(v: VerdictFilter): void {
    this.filters = { ...this.filters, verdict: v }
    void this.refresh()
  }
  toggleApp(name: string): void {
    const next = new Set(this.filters.apps)
    if (next.has(name)) next.delete(name)
    else next.add(name)
    this.filters = { ...this.filters, apps: next }
    void this.refresh()
  }
  toggleModel(name: string): void {
    const next = new Set(this.filters.models)
    if (next.has(name)) next.delete(name)
    else next.add(name)
    this.filters = { ...this.filters, models: next }
    void this.refresh()
  }
  resetFilters(): void {
    this.filters = structuredClone(DEFAULT_FILTERS)
    void this.refresh()
  }
  hasNonDefaultFilters(): boolean {
    const f = this.filters
    return (
      f.search !== '' ||
      f.whenPreset !== '24h' ||
      f.verdict !== 'all' ||
      f.blastClasses.size !== BLAST_CLASSES.length ||
      f.apps.size > 0 ||
      f.models.size > 0
    )
  }

  // ── live SSE ────────────────────────────────────────────────
  // Subscribes to the SSE 'audit' event and pushes new rows to the top
  // of the timeline. Re-subscribes are no-ops.
  startLive(): void {
    if (this.sseOff) return
    try {
      this.sseOff = ipc.on('audit', (event) => {
        this.appendLiveEvent(event as AuditEvent)
      })
    } catch {
      this.sseOff = null
    }
  }
  stopLive(): void {
    if (this.sseOff) {
      try {
        this.sseOff()
      } catch {
        /* ignore */
      }
      this.sseOff = null
    }
  }
  appendLiveEvent(ev: AuditEvent): void {
    if (!ev || typeof ev.id !== 'number') return
    if (this.events.some((e) => e.id === ev.id)) return
    this.events = [ev, ...this.events]
  }
}

export const audit = new AuditStore()