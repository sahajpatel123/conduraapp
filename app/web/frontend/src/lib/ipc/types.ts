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

export interface OnboardingDaemonState {
  current_step: string
  steps: Record<string, OnboardingStepProgress>
  started_at: string
  completed_at?: string
}

// ----- Phase 12: i18n -----

export interface I18nLocaleResult {
  locale: string
  translations: Record<string, string>
}

