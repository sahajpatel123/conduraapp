// TypeScript mirror of internal/ipc.
//
// The Go daemon speaks JSON-RPC 2.0 over HTTP (POST /api) for
// request/response and over a separate SSE stream (GET /events) for
// streaming. The shapes here match the Go struct tags 1:1 so we
// don't have to do any marshaling tricks on the wire.

export interface IPCRequest {
  jsonrpc: '2.0'
  method: string
  params?: unknown
  id: number | string
}

export interface IPCResponse<T = unknown> {
  jsonrpc: '2.0'
  result?: T
  error?: IPCError
  id: number | string
}

export interface IPCError {
  code: number
  message: string
  data?: unknown
}

export interface IPCNotification<T = unknown> {
  jsonrpc: '2.0'
  method: string
  params?: T
}

// JSON-RPC 2.0 standard error codes. Match internal/ipc/server.go.
export const ErrorCode = {
  ParseError: -32700,
  InvalidRequest: -32600,
  MethodNotFound: -32601,
  InvalidParams: -32602,
  InternalError: -32603,
} as const

export type ErrorCodeValue = (typeof ErrorCode)[keyof typeof ErrorCode]

// ----- Domain types (mirrors internal/llm, internal/config, etc.) -----

export interface Role {
  // string enum: 'system' | 'user' | 'assistant' | 'tool'
  value: 'system' | 'user' | 'assistant' | 'tool'
}

export interface Message {
  role: 'system' | 'user' | 'assistant' | 'tool'
  content: string
  tool_calls?: ToolCall[]
  tool_call_id?: string
}

export interface ToolCall {
  id: string
  type: 'function'
  function: {
    name: string
    arguments: string // JSON-encoded
  }
}

export interface ChatRequest {
  model: string
  messages: Message[]
  max_tokens?: number
  temperature?: number
  stream?: boolean
  tools?: ToolSpec[]
}

export interface ToolSpec {
  type: 'function'
  function: {
    name: string
    description: string
    parameters: Record<string, unknown>
  }
}

export interface Usage {
  input_tokens: number
  output_tokens: number
  total_tokens: number
}

export type FinishReason =
  | 'stop'
  | 'length'
  | 'tool_calls'
  | 'content_filter'
  | 'error'

export interface ChatResponse {
  id: string
  model: string
  message: Message
  finish_reason: FinishReason
  usage: Usage
  raw?: string
}

export interface ModelInfo {
  id: string
}

// ----- Daemon introspection -----

export interface PingResult {
  pong: boolean
  ts: number
}

export interface VersionInfo {
  version: string
  commit: string
  build_date: string
  go_version: string
  platform: string
}

export type HealthState = 'ok' | 'degraded' | 'down'

export interface HealthCheckSnapshot {
  name: string
  state: HealthState
  message: string
  last_check: string
  last_error?: string
}

export interface HealthSnapshot {
  overall: HealthState
  checks: HealthCheckSnapshot[]
  ts: string
}

// ----- API keys (secrets) -----

export interface APIKeyMeta {
  id: number
  provider: string
  label: string
  auth_kind: 'api_key' | 'oauth'
  has_token: boolean
}

export interface APIKeySetParams {
  provider: string
  label: string
  secret: string
}

// ----- Provider registry -----

export interface ProviderInfo {
  name: string
  models: ModelInfo[]
}

// ----- Spend monitoring -----

export interface SpendSummary {
  spent: number
  cap: number
  remaining: number
}

// ----- Conversations (sub-phase 2.5) -----

export interface ConversationMeta {
  id: number
  title: string
  created_at: string
  updated_at: string
  message_count: number
}

export interface Conversation {
  id: number
  title: string
  created_at: string
  updated_at: string
  messages: Message[]
}

export interface ConversationCreateParams {
  title?: string
}

export interface ConversationAppendParams {
  id: number
  message: Message
}

// ----- LLM streaming (sub-phase 2.5) -----

export interface StreamEvent {
  conversation_id: number
  delta: string
  role?: 'assistant'
  tool_calls?: ToolCall[]
  finish_reason?: FinishReason
  usage?: Usage
  done: boolean
  err?: string
}

export interface LLMStreamParams {
  conversation_id: number
  provider: string
  request: ChatRequest
}

export interface LLMCancelParams {
  conversation_id: number
}

// ----- Audit log (sub-phase 2.6) -----

export interface AuditEvent {
  id: number
  ts: string
  actor: string
  action: string
  app: string
  level: 'info' | 'warn' | 'error'
  result: 'allow' | 'block' | 'prompt'
  message: string
}

export interface AuditListParams {
  limit?: number
  offset?: number
  since?: string
  action?: string
  level?: 'info' | 'warn' | 'error'
}

// ----- Config (sub-phase 2.6) -----

export interface AppConfig {
  version: number
  general: {
    data_dir: string
    language: string
  }
  daemon: {
    idle_timeout_minutes: number
    default_autonomy: string
  }
  logging: {
    level: string
    format: 'json' | 'text'
    file: string
    add_source: boolean
  }
  storage: {
    path: string
    backup: { dir: string; retention_days: number }
  }
  security: {
    audit_log_path: string
    audit_retention_days: number
    spend_limit_usd_per_day: number
  }
  api_server: {
    host: string
    port: number
    auth_token: string
    tls_enabled: boolean
    allowed_origins: string[]
  }
  llm: {
    providers: Record<string, {
      enabled: boolean
      api_key: string
      base_url: string
      default_model: string
    }>
    oauth_providers: Record<string, unknown>
  }
  autonomy: {
    default_level: string
    per_app: Record<string, string>
    per_task: Record<string, string>
  }
  telemetry: {
    enabled: boolean
    endpoint: string
  }
  hotkey: {
    overlay: string
  }
  window: {
    width: number
    height: number
    x: number
    y: number
    last_conversation_id: number
  }
}

// ----- Daemon control (sub-phase 2.6) -----

export interface DaemonHaltResult {
  halted: boolean
  active_streams_canceled: number
  timestamp: string
}

export interface DaemonUpdateResult {
  update_available: boolean
  current_version: string
  latest_version?: string
  download_url?: string
  forced: boolean
}

// ----- Halt flag (sub-phase 2.6) -----

export interface HaltState {
  halted: boolean
  since?: string
  reason?: string
}

// ----- Phase 11: Trust & Recovery -----

export interface BackupEntry {
  name: string
  path: string
  size: number
}

export interface BackupCreateResult {
  path: string
}

export interface PermissionStatus {
  kind: string
  status: 'granted' | 'denied' | 'unknown'
  note?: string
}

export interface PermissionGuide {
  kind: string
  platform: string
  title: string
  steps: string[]
  deep_link?: string
  help_url?: string
}

export interface OnboardingStepProgress {
  status: string
  data?: string
  updated_at: string
}

// Step is the high-level onboarding flow. The plan calls for a
// 4-step union: eula → permissions → hotkey → complete.
// Welcome and the daemon-side power-source step are implicit
// (Welcome is the pre-wizard state, power-source is part of
// permissions).
export type OnboardingStep = 'eula' | 'permissions' | 'hotkey' | 'complete'

export interface OnboardingDaemonState {
  current_step: string
  steps: Record<string, OnboardingStepProgress>
  started_at: string
  completed_at?: string
}

// EULADocument is the End-User License Agreement shown during
// onboarding. The full text is fetched from the daemon so the
// Svelte UI does not embed legal copy. Shape mirrors
// internal/onboarding/eula.go.
export interface EULADocument {
  // version is the EULA revision (e.g. "v1"). The wizard
  // stores the accepted version and forces a re-accept on
  // version bump.
  version: string
  // text is the full EULA markdown. Rendered in a scroll
  // area so long content never overflows.
  text: string
  // updated_at is the last-modified date of the EULA file
  // (or the version's release date for the bundled fallback).
  updated_at: string
}

// PowerProbeResult reports what the daemon sees on the user's
// machine so the wizard can recommend a "power source"
// (subscription vs API key vs local Ollama). The user's
// choice is independent of the probe; this is informational.
// Shape mirrors internal/onboarding/power.go.
export interface PowerProbeResult {
  // ollama_reachable is true when a local Ollama daemon is
  // reachable on 127.0.0.1:11434.
  ollama_reachable: boolean
  // ollama_models is the list of models Ollama reports. Empty
  // when not reachable.
  ollama_models: string[]
  // clis is the list of CLI tools (claude-code, codex, etc.)
  // found on PATH. The wizard can recommend "use your local
  // Claude Code subscription" when found.
  clis: PowerProbeCLI[]
  // recommended is the daemon's best guess for the user's
  // primary source ("ollama", "claude-code", "codex", or
  // "none"). The user can override.
  recommended: string
}

// PowerProbeCLI describes one CLI tool found on PATH.
export interface PowerProbeCLI {
  name: string
  found: boolean
}

// OnboardingFinishParams is the payload the GUI sends to
// complete the wizard. The daemon persists these as the
// user's first-run preferences. Shape mirrors the params
// the daemon's onboarding.finish RPC expects.
export interface OnboardingFinishParams {
  // hotkey is the user's chosen overlay hotkey (e.g.
  // "Cmd+Shift+Space"). Persisted to config. Required.
  hotkey: string
  // eula_version is the EULA revision the user accepted.
  // Persisted to step metadata so future EULA bumps force
  // re-accept. Required.
  eula_version: string
  // permissions_skipped is true when the user opted to skip
  // the permissions grant step. Per CLAUDE.md the user can
  // grant later; we record the choice so the audit log has it.
  permissions_skipped?: boolean
}

// OnboardingFinishResult tells the GUI whether the daemon
// accepted the finish. On success, the wizard dismisses and
// the main UI mounts. Shape mirrors the Go onboarding.finish
// return value.
export interface OnboardingFinishResult {
  // power is the power probe result, used to render the
  // Ready screen ("found your local Ollama, etc."). The
  // probe is also used internally to auto-enable Ollama in
  // config when reachable.
  power: PowerProbeResult
  // hotkey is the hotkey the daemon actually persisted.
  // May differ from the user's request if the daemon
  // detected a conflict (currently always equal).
  hotkey: string
  // completed_at is the RFC3339 timestamp of the completion.
  completed_at: string
}

// ----- Phase 12: i18n -----

export interface I18nLocaleResult {
  locale: string
  translations: Record<string, string>
}

// ----- Phase 11: Replay -----

export interface ReplayFrame {
  id: number
  action: string
  app: string
  actor: string
  result: string
  level: string
  message: string
  timestamp: string
  outcome: string
  outcome_reason?: string
  before_screenshot?: string
  after_screenshot?: string
  before_screenshot_mime?: string
  after_screenshot_mime?: string
}

export interface ReplayIntegrityReport {
  valid: boolean
  rows_checked: number
  first_break_id?: number
  first_break_reason?: string
}

export interface ReplayExportResult {
  path: string
}

// ----- Phase 12: Skills Hub -----

export interface HubSkillMeta {
  id: string
  name: string
  description: string
  version: string
  author: string
  license: string
  tags: string[]
  trust: string
  checksum: string
  downloads: number
  published_at: string
  updated_at: string
}

export interface HubSearchResult {
  skills: HubSkillMeta[]
  total: number
  query: string
}

export interface HubInstallResult {
  ok: boolean
  id: string
}

// ----- Phase 12: Skills (local) -----

export interface InstalledSkill {
  id: string
  name: string
  description: string
  version: string
  author: string
  license: string
  trust: string
  source?: string
  hub_id?: string
  checksum?: string
}

// ----- Phase 12: P2P Sync -----

export interface SyncStatus {
  device_id: string
  name: string
  peers: number
  entries: number
  running: boolean
  sync_port?: number
  paired_devices?: number
  enabled?: boolean
}

export interface SyncPeer {
  device_id: string
  name: string
  public_key: string
  address: string
  last_seen: string
  fingerprint: string
}

export interface SyncPairedDevice {
  device_id: string
  device_name: string
  public_key: string
  paired_at: string
}

export interface SyncListPairsResult {
  devices: SyncPairedDevice[]
}

export interface SyncPeersResult {
  peers: SyncPeer[]
}

