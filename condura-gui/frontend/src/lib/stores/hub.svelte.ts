// Hub store (Phase 14G).
//
// Caches the Skills Hub search/install state plus the
// publish state (which is a longer flow: validate name →
// package archive → upload). The store keeps search results
// ephemeral; installed skills live in the daemon's
// `skills.*` namespace (see app/web/frontend/src/lib/stores/
// skills.svelte.ts or the existing skills.list RPC).
//
// State machine (publish flow):
//
//   ┌────────────┐  publish(name,ver,arch) ┌────────────┐
//   │    idle    │ ────────────────────────▶ │ uploading  │
//   │            │                          │            │
//   └────────────┘                          └────────────┘
//        ▲                                        │
//        │                                        ▼
//        │                       ┌────────────────────┐
//        └─────── success/error │    success / error  │
//                               └────────────────────┘

import { ipc } from '../ipc/client'
import type {
  HubSearchResult,
  HubInstallResult,
  HubPublishParams,
  HubPublishResult,
} from '../ipc/types'

// PublishState tracks the publish flow. The GUI uses this
// to show a spinner + result toast. "idle" is the resting
// state; "uploading" is set when publish() is in flight;
// "success" / "error" are set after the RPC returns.
export type PublishState =
  | { kind: 'idle' }
  | { kind: 'uploading'; name: string; version: string }
  | { kind: 'success'; result: HubPublishResult }
  | { kind: 'error'; message: string }

/**
 * HubStore: search results, install state, publish flow.
 *
 * Lifecycle:
 *   1. App mounts → store starts with empty results.
 *   2. Hub.svelte calls hub.search(query) on user input.
 *   3. User clicks "Install" on a result → hub.install(id).
 *   4. User clicks "Publish" → hub.publish(name, version, archive).
 */
export class HubStore {
  /** Most recent search results. */
  results = $state<HubSearchResult['skills']>([])

  /** Last query that returned the current results. */
  lastQuery = $state<string>('')

  /** True while a search or install RPC is in flight. */
  loading = $state<boolean>(false)

  /**
   * The most recent error from a Hub RPC. Cleared on the
   * next call.
   */
  error = $state<string | null>(null)

  /**
   * Tracks the publish flow. "idle" = no publish in
   * progress; "uploading" = RPC in flight; "success" /
   * "error" = terminal states (the GUI renders a toast
   * and resets to idle).
   */
  publishState = $state<PublishState>({ kind: 'idle' })

  /**
   * Set of installed skill IDs. The GUI checks this to
   * render the "installed ✓" badge on a search result.
   * The store is the source of truth; refresh() syncs
   * from the daemon on demand.
   */
  installed = $state<Set<string>>(new Set())

  /**
   * Searches the Hub. Empty query returns the empty list
   * (no RPC). Results overwrite the previous query's
   * results to avoid races.
   */
  async search(query: string, limit: number = 20): Promise<void> {
    if (query.trim() === '') {
      this.results = []
      this.lastQuery = ''
      return
    }
    this.loading = true
    this.error = null
    try {
      const res = await ipc.hubSearch(query, limit)
      // Only commit the result if the query is still
      // current (the user may have typed a new query
      // before this one returned).
      if (query === this.lastQuery || this.lastQuery === '') {
        this.results = res.skills ?? []
        this.lastQuery = query
      }
    } catch (e) {
      this.error = String(e)
    } finally {
      this.loading = false
    }
  }

  /**
   * Downloads + safety-scans + installs a skill from the
   * Hub. The daemon performs the entire flow; the GUI
   * shows a success toast on completion.
   */
  async install(id: string): Promise<HubInstallResult | null> {
    this.loading = true
    this.error = null
    try {
      const result = await ipc.hubInstall(id)
      // Track the installed id so search results can
      // show an "installed" badge without a separate RPC.
      const next = new Set(this.installed)
      next.add(id)
      this.installed = next
      return result
    } catch (e) {
      this.error = String(e)
      return null
    } finally {
      this.loading = false
    }
  }

  /**
   * Uploads a skill to the Hub. The archive bytes are
   * passed in directly (the daemon does NOT reach back
   * to disk). Updates publishState through the flow so
   * the GUI can show a spinner.
   *
   * The plan's signature is publish(name, version, archive).
   * The ID is derived from the convention
   * "<name>@<version>" (lowercased, slugified). If you
   * need a different ID, build the HubPublishParams
   * object yourself and call publishWithParams.
   */
  async publish(
    name: string,
    version: string,
    archive: Uint8Array
  ): Promise<HubPublishResult | null> {
    const id = `${slugify(name)}@${version}`
    return this.publishWithParams({
      id,
      archive,
      name,
      version,
      description: '',
      author: '',
      license: '',
      tags: [],
    })
  }

  /**
   * Lower-level publish that takes the full
   * HubPublishParams. Use this when the caller needs
   * control over the ID, description, author, etc.
   */
  async publishWithParams(
    p: HubPublishParams
  ): Promise<HubPublishResult | null> {
    this.publishState = { kind: 'uploading', name: p.name, version: p.version }
    this.error = null
    try {
      const result = await ipc.hubPublishTyped(p)
      this.publishState = { kind: 'success', result }
      return result
    } catch (e) {
      const message = String(e)
      this.publishState = { kind: 'error', message }
      this.error = message
      return null
    }
  }

  /**
   * Resets the publish state to idle. Called by the GUI
   * after showing a success or error toast.
   */
  resetPublishState(): void {
    this.publishState = { kind: 'idle' }
  }

  /**
   * True when the GUI should render the publish spinner.
   * Convenience getter so components don't have to
   * switch on publishState.kind.
   */
  get isPublishing(): boolean {
    return this.publishState.kind === 'uploading'
  }
}

// slugify converts a human-readable skill name into a
// canonical hub id. e.g. "Weather Lookup" -> "weather-lookup".
// Used to derive the id from the name when the caller
// doesn't supply one explicitly.
function slugify(s: string): string {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

// Singleton instance — only one publish flow at a time.
export const hub = new HubStore()
