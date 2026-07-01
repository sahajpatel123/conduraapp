// IPC client for the Condura daemon.
//
// Talks JSON-RPC 2.0 over HTTP POST to the daemon's /api endpoint.
// The auth token (if any) is sent as a Bearer header.
//
// Connection management:
//   - The client maintains a single WebSocket-style HTTP/1.1
//     keep-alive to the daemon for low-latency calls.
//   - If the daemon restarts (or the URL changes), the client
//     reconnects with exponential backoff.
//   - Streaming events come over a SEPARATE EventSource on
//     /events (see transport-sse.ts).
//
// Lifecycle:
//   - import { ipc } from '$lib/ipc/client'
//   - ipc.start()  — open the connection (idempotent)
//   - await ipc.call('ping', {})  — typed-ish request/response
//   - ipc.on('event-name', handler)  — push notifications
//   - ipc.stop()  — close everything (called on app shutdown)

import { EventEmitter } from '../utils/eventemitter'
import type {
  IPCRequest,
  IPCResponse,
  PingResult,
  VersionInfo,
  HealthSnapshot,
  APIKeyMeta,
  APIKeySetParams,
  ProviderInfo,
  SpendSummary,
  Conversation,
  ConversationMeta,
  ConversationCreateParams,
  ConversationAppendParams,
  AuditEvent,
  AuditListParams,
  AppConfig,
  DaemonHaltResult,
  DaemonResumeRequestResult,
  DaemonUpdateResult,
  HaltState,
  DaemonCapabilities,
  LLMStreamParams,
  LLMCancelParams,
} from './types'

// Emitter that supports typed event handlers.
type EventMap = {
  // Connection lifecycle.
  connected: []
  disconnected: [reason: string]
  reconnecting: [attempt: number, delayMs: number]
  // Server-pushed events.
  halt: [HaltState]
  spend_warning: [SpendSummary]
  audit: [AuditEvent]
  // Raw stream events from the SSE channel.
  stream: [import('./types').StreamEvent]
}

type EventName = keyof EventMap

class TypedEmitter extends EventEmitter<EventMap> {}

// IPC client singleton — there is only one daemon process per
// machine, so a single client is correct.
class IPCClient {
  private emitter = new TypedEmitter()
  private baseURL = ''
  private authToken = ''
  private nextId = 1
  private connected = false
  private reconnectAttempt = 0
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private sse: EventSource | null = null
  private sseURL = ''

  /**
   * Configure and start the IPC client. Must be called before any
   * call() / on() can succeed.
   */
  async start(opts: { baseURL: string; authToken: string }): Promise<void> {
    this.baseURL = opts.baseURL.replace(/\/$/, '')
    this.authToken = opts.authToken
    await this.openSse()
  }

  /**
   * Stop the client and release resources. Idempotent.
   */
  stop(): void {
    this.connected = false
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.sse) {
      this.sse.close()
      this.sse = null
    }
  }

  // ---- Event subscription ----

  on<E extends EventName>(event: E, handler: (...args: EventMap[E]) => void): () => void {
    return this.emitter.on(event, handler)
  }

  off<E extends EventName>(event: E, handler: (...args: EventMap[E]) => void): void {
    this.emitter.off(event, handler)
  }

  // ---- Connection state ----

  isConnected(): boolean {
    return this.connected
  }

  // ---- Core RPC call ----

  async call<T = unknown>(method: string, params?: unknown): Promise<T> {
    if (!this.baseURL) {
      throw new Error('IPC client not started; call start() first')
    }
    const id = this.nextId++
    const req: IPCRequest = {
      jsonrpc: '2.0',
      method,
      params,
      id
    }
    const res = await fetch(`${this.baseURL}/api`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(this.authToken ? { Authorization: `Bearer ${this.authToken}` } : {})
      },
      body: JSON.stringify(req)
    })
    if (!res.ok) {
      throw new Error(`IPC HTTP ${res.status}: ${await res.text()}`)
    }
    const rpc = (await res.json()) as IPCResponse<T>
    if (rpc.error) {
      throw new Error(`IPC ${rpc.error.code}: ${rpc.error.message}`)
    }
    return rpc.result as T
  }

  // ---- Typed convenience methods ----

  ping(): Promise<PingResult> {
    return this.call<PingResult>('ping', {})
  }
  version(): Promise<VersionInfo> {
    return this.call<VersionInfo>('version', {})
  }
  configGet(): Promise<AppConfig> {
    return this.call<AppConfig>('config.get', {})
  }
  configUpdate(patch: Partial<AppConfig>): Promise<void> {
    return this.call<void>('config.update', patch)
  }
  healthSnapshot(): Promise<HealthSnapshot> {
    return this.call<HealthSnapshot>('health.snapshot', {})
  }
  providersList(): Promise<ProviderInfo[]> {
    return this.call<ProviderInfo[]>('providers.list', {})
  }
  providersModels(provider: string): Promise<{ id: string }[]> {
    return this.call<{ id: string }[]>('providers.models', { provider })
  }
  apiKeysList(): Promise<APIKeyMeta[]> {
    return this.call<APIKeyMeta[]>('apikeys.list', {})
  }
  apiKeysSet(p: APIKeySetParams): Promise<{ id: number }> {
    return this.call<{ id: number }>('apikeys.set', p)
  }
  apiKeysDelete(id: number): Promise<void> {
    return this.call<void>('apikeys.delete', { id })
  }

  // ---- Voice ----

  voiceListen(): Promise<{ transcript: string; confidence: number }> {
    return this.call('voice.listen', {})
  }
  spendToday(): Promise<SpendSummary> {
    return this.call<SpendSummary>('spend.today', {})
  }
  llmChat(provider: string, model: string, request: import('./types').ChatRequest): Promise<{ response: import('./types').ChatResponse; cost_usd: number }> {
    return this.call('llm.chat', { provider, model, request })
  }
  llmStream(p: LLMStreamParams): Promise<{ started: true }> {
    return this.call('llm.stream', p)
  }
  llmCancel(p: LLMCancelParams): Promise<void> {
    return this.call('llm.cancel', p)
  }
  conversationsList(): Promise<ConversationMeta[]> {
    return this.call<ConversationMeta[]>('conversations.list', {})
  }
  conversationsGet(id: number): Promise<Conversation> {
    return this.call<Conversation>('conversations.get', { id })
  }
  conversationsCreate(p: ConversationCreateParams): Promise<ConversationMeta> {
    return this.call<ConversationMeta>('conversations.create', p)
  }
  conversationsDelete(id: number): Promise<void> {
    return this.call<void>('conversations.delete', { id })
  }
  conversationsAppend(p: ConversationAppendParams): Promise<void> {
    return this.call<void>('conversations.append', p)
  }
  auditList(p: AuditListParams = {}): Promise<AuditEvent[]> {
    return this.call<AuditEvent[]>('audit.list', p)
  }
  daemonHalt(reason: string): Promise<DaemonHaltResult> {
    return this.call<DaemonHaltResult>('daemon.halt', { reason })
  }
  // T3b sticky resume (P0-1 core). The GUI shows the user a code or
  // an instruction to confirm un-halt via `condura resume confirm
  // --ticket T` from a terminal. The GUI itself only mints the ticket
  // (via daemonResumeRequest); it does NOT hold the resume secret and
  // does NOT call halt.confirm_resume. The human-confirmation path is
  // out of the in-process trust boundary.
  daemonResumeRequest(): Promise<DaemonResumeRequestResult> {
    return this.call<DaemonResumeRequestResult>('daemon.resume_request', {})
  }
  haltState(): Promise<HaltState> {
    return this.call<HaltState>('halt.state', {})
  }
  // daemonCapabilities returns the runtime "Trust & Safety" surface
  // — the GUI's read-only "What this build can and can't do" panel
  // renders from this. The shape reflects reality, not aspirations:
  // the Layer 3 in_process flag is honest about v0.1.0's soft
  // network guard. CLAUDE.md §2.1 invariant #4, §33.5.2 row C4.14.
  daemonCapabilities(): Promise<DaemonCapabilities> {
    return this.call<DaemonCapabilities>('daemon.capabilities', {})
  }
  telemetrySetEnabled(enabled: boolean): Promise<void> {
    return this.call<void>('telemetry.setEnabled', { enabled })
  }
  updateCheck(): Promise<DaemonUpdateResult> {
    return this.call<DaemonUpdateResult>('update.check', {})
  }
  updateApply(): Promise<DaemonUpdateResult> {
    return this.call<DaemonUpdateResult>('update.apply', {})
  }
  windowShow(): Promise<void> {
    return this.call<void>('window.show', {})
  }
  windowHide(): Promise<void> {
    return this.call<void>('window.hide', {})
  }
  overlayShow(): Promise<void> {
    return this.call<void>('overlay.show', {})
  }
  overlayHide(): Promise<void> {
    return this.call<void>('overlay.hide', {})
  }
  trayUpdate(p: { spend?: SpendSummary; active_conversation?: ConversationMeta; status?: 'ok' | 'degraded' | 'down' }): Promise<void> {
    return this.call<void>('tray.update', p)
  }
  firstRunComplete(): Promise<void> {
    return this.call<void>('firstRun.complete', {})
  }
  firstRunStatus(): Promise<{ complete: boolean }> {
    return this.call('firstRun.status', {})
  }

  // Phase 11 — backup & permissions
  backupList(): Promise<import('./types').BackupEntry[]> {
    return this.call('backup.list', {})
  }
  backupCreate(destination?: string): Promise<import('./types').BackupCreateResult> {
    return this.call('backup.create', destination ? { destination } : {})
  }
  // backupRestore reverts the daemon's data directory to the
  // contents of a backup archive. Gated through the Gatekeeper
  // (the daemon's defaults.yaml requires explicit user consent
  // for "restore"); the GUI's confirmation dialog is the first
  // line of defense.
  backupRestore(p: import('./types').BackupRestoreParams): Promise<import('./types').BackupRestoreResult> {
    return this.call('backup.restore', p)
  }
  permissionsStatus(): Promise<import('./types').PermissionStatus[]> {
    return this.call('permissions.status', {})
  }
  permissionsGuide(kind: string): Promise<import('./types').PermissionGuide> {
    return this.call('permissions.request_guide', { kind })
  }
  onboardingState(): Promise<import('./types').OnboardingDaemonState> {
    return this.call('onboarding.state', {})
  }
  onboardingAdvance(): Promise<import('./types').OnboardingDaemonState> {
    return this.call('onboarding.advance', {})
  }
  onboardingComplete(): Promise<import('./types').OnboardingDaemonState> {
    return this.call('onboarding.complete', {})
  }
  // onboardingSetStep records a step's status + optional data.
  // Used by the EULA step to record the accepted EULA
  // version in step metadata, so future EULA bumps force a
  // re-accept.
  onboardingSetStep(
    step: import('./types').OnboardingStep | string,
    status: 'pending' | 'in_progress' | 'complete' | 'skipped' | string,
    data?: string
  ): Promise<import('./types').OnboardingDaemonState> {
    return this.call('onboarding.set_step', { step, status, data: data ?? '' })
  }
  onboardingBack(): Promise<import('./types').OnboardingDaemonState> {
    return this.call('onboarding.back', {})
  }
  onboardingReset(): Promise<import('./types').OnboardingDaemonState> {
    return this.call('onboarding.reset', {})
  }

  // Onboarding (Phase 12/14 wizard). The high-level wrappers
  // accept a step name and an optional payload; the daemon
  // handles state transitions and persistence. The four-step
  // union is: eula | permissions | hotkey | complete.
  onboardingEula(): Promise<import('./types').EULADocument> {
    return this.call('onboarding.eula', {})
  }
  onboardingProbePower(): Promise<import('./types').PowerProbeResult> {
    return this.call('onboarding.probe_power', {})
  }
  onboardingSkip(
    step: import('./types').OnboardingStep
  ): Promise<import('./types').OnboardingDaemonState> {
    return this.call('onboarding.skip', { step })
  }
  onboardingFinish(
    p: import('./types').OnboardingFinishParams
  ): Promise<import('./types').OnboardingFinishResult> {
    return this.call('onboarding.finish', p)
  }
  onboardingIsComplete(): Promise<boolean> {
    return this.call('onboarding.is_complete', {})
  }

  // Phase 12 — i18n
  i18nLocales(): Promise<string[]> {
    return this.call('i18n.locales', {})
  }
  i18nLocale(locale: string): Promise<import('./types').I18nLocaleResult> {
    return this.call('i18n.locale', { locale })
  }

  // Phase 12 — Skills Hub
  hubSearch(query: string, limit = 20): Promise<import('./types').HubSearchResult> {
    return this.call('hub.search', { query, limit })
  }
  hubGet(id: string): Promise<import('./types').HubSkillMeta> {
    return this.call('hub.get', { id })
  }
  hubInstall(id: string): Promise<import('./types').HubInstallResult> {
    return this.call('hub.install', { id })
  }
  hubPublish(id: string, path: string): Promise<{ ok: boolean; id: string }> {
    return this.call('hub.publish', { id, path })
  }

  // Phase 12 — Skills (local)
  skillsList(limit = 100): Promise<import('./types').InstalledSkill[]> {
    return this.call('skills.list', { limit })
  }
  skillsGet(id: string): Promise<import('./types').InstalledSkill> {
    return this.call('skills.get', { id })
  }
  skillsDelete(id: string): Promise<{ ok: boolean }> {
    return this.call('skills.delete', { id })
  }

  // Phase 12 — P2P Sync
  syncStatus(): Promise<import('./types').SyncStatus> {
    return this.call('sync.status', {})
  }
  syncPeers(): Promise<import('./types').SyncPeersResult> {
    return this.call('sync.peers', {})
  }
  syncListPairs(): Promise<import('./types').SyncListPairsResult> {
    return this.call('sync.list_pairs', {})
  }
  syncPairBegin(deviceId: string): Promise<{ ok: boolean; pin: string; peer: string }> {
    return this.call('sync.pair_begin', { device_id: deviceId })
  }
  syncPairConfirm(deviceId: string, pin: string): Promise<{ ok: boolean; device_id: string }> {
    return this.call('sync.pair_confirm', { device_id: deviceId, pin })
  }
  syncRevoke(deviceId: string): Promise<{ ok: boolean; revoked_device_id: string; revoker_device_id: string; revoked_at: string; signature: string }> {
    return this.call('sync.revoke', { device_id: deviceId })
  }
  syncWith(deviceId: string): Promise<{ ok: boolean; merged: number }> {
    return this.call('sync.sync_with', { device_id: deviceId })
  }

  replayTimeline(): Promise<import('./types').ReplayFrame[]> {
    return this.call('replay.timeline', {})
  }
  replayVerifyIntegrity(): Promise<import('./types').ReplayIntegrityReport> {
    return this.call('replay.verify_integrity', {})
  }
  replayExport(destination?: string): Promise<import('./types').ReplayExportResult> {
    return this.call('replay.export', destination ? { destination } : {})
  }

  // ----- Phase 15: Gatekeeper consent -----
  gatekeeperPendingConsent(): Promise<import('./types').ConsentPendingResult> {
    return this.call('gatekeeper.pending_consent', {})
  }
  gatekeeperApprove(nonce: string): Promise<{ ok: boolean }> {
    return this.call('gatekeeper.approve', { nonce })
  }
  gatekeeperDeny(nonce: string): Promise<{ ok: boolean }> {
    return this.call('gatekeeper.deny', { nonce })
  }

  // ----- Adaptive engine -----
  adaptiveProfile(): Promise<import('./types').AdaptiveUserModel> {
    return this.call('adaptive.profile', {})
  }
  adaptiveForget(field: string, value: string): Promise<{ ok: boolean }> {
    return this.call('adaptive.forget', { field, value })
  }
  adaptiveReset(): Promise<{ ok: boolean }> {
    return this.call('adaptive.reset', {})
  }
  adaptiveStrengthGet(): Promise<{ strength: import('./types').AdaptiveStrength }> {
    return this.call('adaptive.strength.get', {})
  }
  adaptiveStrengthSet(
    strength: import('./types').AdaptiveStrength
  ): Promise<{ ok: boolean }> {
    return this.call('adaptive.strength.set', { strength })
  }
  onboardingProbeVoice(): Promise<import('./types').VoiceProbeResult> {
    return this.call('onboarding.probe_voice', {})
  }

  // ----- Phase 14B: Account (auth) -----
  // The account.* RPCs talk to the daemon's user record. The
  // magic-link OAuth flow on web talks to the Next.js
  // /api/auth/magic route (see web/app/api/auth/magic/route.ts);
  // the daemon is bypassed so the Vercel KV can store the
  // one-time token. AccountStore routes to whichever is
  // appropriate.
  accountStatus(): Promise<import('./types').AccountStatus> {
    return this.call('account.status', {})
  }
  accountProviders(): Promise<import('./types').ProvidersResult> {
    return this.call('account.providers', {})
  }
  accountOAuthURL(
    p: import('./types').OAuthURLParams
  ): Promise<import('./types').OAuthURLResult> {
    return this.call('account.oauth_url', p)
  }
  accountOAuthCallback(
    p: import('./types').OAuthCallbackParams
  ): Promise<import('./types').AccountStatus> {
    return this.call('account.oauth_callback', p)
  }
  accountMagicLink(
    p: import('./types').MagicLinkParams
  ): Promise<import('./types').MagicLinkResult> {
    return this.call('account.magic_link', p)
  }
  accountLogout(): Promise<import('./types').LogoutResult> {
    return this.call('account.logout', {})
  }

  // ----- Phase 14F: Sync pairing (typed results) -----
  // The plan asks for typed PairBeginResult / PairConfirmResult
  // wrappers. The pre-existing methods use loose object
  // types; the new typed methods are the canonical API.
  syncPairBeginTyped(
    deviceId: string
  ): Promise<import('./types').PairBeginResult> {
    return this.call<import('./types').PairBeginResult>(
      'sync.pair_begin',
      { device_id: deviceId }
    )
  }
  syncPairConfirmTyped(
    deviceId: string,
    pin: string
  ): Promise<import('./types').PairConfirmResult> {
    return this.call<import('./types').PairConfirmResult>(
      'sync.pair_confirm',
      { device_id: deviceId, pin }
    )
  }
  // syncRevokeTyped is just an alias for the existing
  // syncRevoke with a strict return type.
  syncRevokeTyped(
    deviceId: string
  ): Promise<{
    ok: boolean
    revoked_device_id: string
    revoker_device_id: string
    revoked_at: string
    signature: string
  }> {
    return this.call('sync.revoke', { device_id: deviceId })
  }

  // ----- Phase 14G: Hub publish (archive in body) -----
  // The plan's HubPublishParams passes the archive as a
  // number[] | Uint8Array. The IPC layer JSON-encodes
  // binary as base64 automatically; the GUI just needs to
  // pass the typed array.
  hubPublishTyped(
    p: import('./types').HubPublishParams
  ): Promise<import('./types').HubPublishResult> {
    return this.call('hub.publish', p)
  }

  // ---- SSE transport ----

  private async openSse(): Promise<void> {
    if (this.sse) {
      this.sse.close()
      this.sse = null
    }
    const url = new URL(this.baseURL)
    url.protocol = url.protocol === 'https:' ? 'https:' : 'http:'
    url.pathname = '/events'
    this.sseURL = url.toString()

    // EventSource can't send custom headers, so we exchange the
    // real bearer token for a short-lived one-time ticket via
    // POST /sse-ticket (header auth). The ticket is used as
    // ?ticket= on the EventSource URL. This keeps the real token
    // out of URLs, server logs, and browser history.
    if (this.authToken) {
      try {
        const ticketUrl = new URL(this.baseURL)
        ticketUrl.protocol = ticketUrl.protocol === 'https:' ? 'https:' : 'http:'
        ticketUrl.pathname = '/sse-ticket'
        const resp = await fetch(ticketUrl.toString(), {
          method: 'POST',
          headers: { Authorization: `Bearer ${this.authToken}` },
        })
        if (resp.ok) {
          const data = await resp.json()
          if (data.ticket) {
            url.searchParams.set('ticket', data.ticket)
          }
        }
      } catch {
        // Token exchange failed — proceed without a ticket.
        // The SSE connect will 401 and trigger reconnect.
      }
    }

    const es = new EventSource(url.toString())
    this.sse = es

    es.addEventListener('open', () => {
      this.connected = true
      this.reconnectAttempt = 0
      this.emitter.emit('connected')
    })

    es.addEventListener('error', () => {
      this.connected = false
      this.emitter.emit('disconnected', 'sse-error')
      this.scheduleReconnect()
    })

    // Generic message event — the daemon sends data as the message
    // payload (default 'message' event).
    es.addEventListener('message', (ev: MessageEvent) => {
      try {
        const data = JSON.parse(ev.data)
        this.handleServerEvent(data)
      } catch {
        // ignore malformed events
      }
    })

    // Named event types (the daemon can also use named events).
    const namedEvents = ['halt', 'spend_warning', 'audit', 'stream']
    for (const name of namedEvents) {
      es.addEventListener(name, (ev: MessageEvent) => {
        try {
          const data = JSON.parse(ev.data)
          this.emitter.emit(name as EventName, data)
        } catch {
          // ignore
        }
      })
    }

    // Stream events published under namespaced names by the daemon's
    // stream manager (stream.started, stream.delta, stream.finished,
    // stream.error, stream.cancelled). Remap them to the bare 'stream'
    // emitter so the conversation store receives them uniformly.
    const streamLifecycle = ['stream.started', 'stream.delta', 'stream.finished', 'stream.error', 'stream.cancelled']
    for (const name of streamLifecycle) {
      es.addEventListener(name, (ev: MessageEvent) => {
        try {
          const data = JSON.parse(ev.data)
          const streamEvent: import('./types').StreamEvent = {
            conversation_id: data.conversation_id ?? 0,
            delta: data.delta ?? '',
            role: data.role,
            tool_calls: data.tool_calls,
            finish_reason: data.finish_reason,
            usage: data.input_tokens != null ? {
              input_tokens: data.input_tokens ?? 0,
              output_tokens: data.output_tokens ?? 0,
              total_tokens: data.total_tokens ?? 0,
            } : undefined,
            done: name === 'stream.finished' || name === 'stream.cancelled' || (name === 'stream.error' ? true : false),
            err: name === 'stream.error' ? (data.error ?? 'stream error') : undefined,
          }
          this.emitter.emit('stream', streamEvent)
        } catch {
          // ignore
        }
      })
    }
  }

  private handleServerEvent(data: unknown): void {
    // The daemon may send a method-shaped notification:
    //   { method: 'halt', params: { halted: true, ... } }
    // or a raw params object.
    if (typeof data === 'object' && data !== null && 'method' in data) {
      const evt = data as { method: string; params: unknown }
      const allowed: EventName[] = ['halt', 'spend_warning', 'audit', 'stream']
      if (allowed.includes(evt.method as EventName)) {
        // Pass params through; the consumer decides shape.
        this.emitter.emit(evt.method as EventName, evt.params as never)
      }
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      return
    }
    this.reconnectAttempt++
    const delay = Math.min(30000, 250 * Math.pow(2, this.reconnectAttempt - 1))
    this.emitter.emit('reconnecting', this.reconnectAttempt, delay)
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      void this.openSse()
    }, delay)
  }
}

// Singleton instance.
export const ipc = new IPCClient()

// Wails bindings are loaded lazily through the global window.go
// object that Wails injects at runtime. We never import from
// '../wailsjs/...' directly because that path may not exist
// during Vite dev or in tests.
declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          Ping: (name: string) => Promise<string>
          DaemonStatus: () => Promise<{ ready: boolean; addr: string }>
        }
      }
    }
  }
}

export const wailsBindings = {
  Ping: async (name: string): Promise<string> => {
    try {
      const fn = window?.go?.main?.App?.Ping
      if (fn) {
        return await fn(name)
      }
    } catch {
      // ignore
    }
    return ''
  },
  DaemonStatus: async (): Promise<{ ready: boolean; addr: string }> => {
    try {
      const fn = window?.go?.main?.App?.DaemonStatus
      if (fn) {
        return await fn()
      }
    } catch {
      // ignore
    }
    return { ready: false, addr: '' }
  }
}
