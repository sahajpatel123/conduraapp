// Package config provides the Synaptic daemon's configuration system.
//
// The configuration is loaded from a YAML file (~/.synaptic/config.yaml by
// default) and merged with built-in defaults and environment variable
// overrides. It is validated on load to fail fast on misconfiguration.
//
// Schema versioning: the YAML file starts with a "version" key. The loader
// will refuse to start on a version mismatch (with a clear migration error),
// preserving the user's data.
//
// Overrides: environment variables prefixed with SYNAPTIC_ override YAML
// values, e.g., SYNAPTIC_LOGGING_LEVEL=debug overrides logging.level.
//
// The Config struct is intended to be read-only after Load. Mutating it
// at runtime should go through the IPC API (not direct mutation).
package config

// ConfigSchemaVersion is the current schema version. Bump on breaking changes.
const ConfigSchemaVersion = 1

// Default config file location per OS.
const (
	DefaultConfigFileName = "config.yaml"
	DefaultEnvPrefix      = "SYNAPTIC_"
)

// Config is the root configuration struct.
type Config struct {
	// Version is the schema version. Always 1 for now.
	Version int `yaml:"version"`

	// General settings.
	General GeneralConfig `yaml:"general"`

	// Daemon behavior.
	Daemon DaemonConfig `yaml:"daemon"`

	// Logging configuration.
	Logging LoggingConfig `yaml:"logging"`

	// Storage configuration.
	Storage StorageConfig `yaml:"storage"`

	// Security / safety configuration.
	Security SecurityConfig `yaml:"security"`

	// Router (LLM delegation) configuration.
	Router RouterConfig `yaml:"router"`

	// LLM provider configuration.
	LLM LLMConfig `yaml:"llm"`

	// API server (IPC) configuration.
	APIServer APIServerConfig `yaml:"api_server"`

	// Autonomy matrix (per-app, per-task).
	Autonomy AutonomyConfig `yaml:"autonomy"`

	// Telemetry / crash reporting.
	Telemetry TelemetryConfig `yaml:"telemetry"`
}

// -----------------------------------------------------------------------------
// Sub-configs
// -----------------------------------------------------------------------------

// GeneralConfig holds general settings.
type GeneralConfig struct {
	// InstallID is a stable UUID per install, generated on first run.
	InstallID string `yaml:"install_id"`
	// DataDir is the root data directory (~/.synaptic by default).
	DataDir string `yaml:"data_dir"`
	// CacheDir is for ephemeral caches.
	CacheDir string `yaml:"cache_dir"`
	// Language is the user's preferred UI language (BCP-47, e.g., "en-US").
	Language string `yaml:"language"`
	// FirstRun indicates whether onboarding has completed.
	FirstRun bool `yaml:"first_run"`
}

// DaemonConfig controls daemon behavior.
type DaemonConfig struct {
	// AutoStart on user login (LaunchAgent on macOS, etc.).
	AutoStart bool `yaml:"auto_start"`
	// Hotkey is the global hotkey combo, e.g., "Cmd+Shift+Space".
	// Empty means the user must set it on first run.
	Hotkey string `yaml:"hotkey"`
	// IdleTimeoutMinutes is how long the daemon waits with no user activity
	// before pausing background perception. 0 disables the timeout.
	IdleTimeoutMinutes int `yaml:"idle_timeout_minutes"`
	// DefaultAutonomy is the default autonomy level (supervised, warn, autonomous).
	DefaultAutonomy string `yaml:"default_autonomy"`
}

// LoggingConfig controls logging.
type LoggingConfig struct {
	// Level: "debug", "info", "warn", "error".
	Level string `yaml:"level"`
	// Format: "text" or "json".
	Format string `yaml:"format"`
	// File is an optional file to log to in addition to stderr.
	File string `yaml:"file"`
	// AddSource adds file:line to each entry.
	AddSource bool `yaml:"add_source"`
}

// StorageConfig controls the SQLite store.
type StorageConfig struct {
	// Path is the SQLite file path. Empty means "use DataDir/synaptic.db".
	Path string `yaml:"path"`
	// Backup settings.
	Backup BackupConfig `yaml:"backup"`
	// Encryption settings.
	Encryption EncryptionConfig `yaml:"encryption"`
}

// BackupConfig controls backups.
type BackupConfig struct {
	// OnUninstall triggers a backup before uninstall.
	OnUninstall bool `yaml:"on_uninstall"`
	// Dir is the backup destination. Empty means "use DataDir/backups".
	Dir string `yaml:"dir"`
	// RetentionDays: backups older than this are auto-pruned. 0 = forever.
	RetentionDays int `yaml:"retention_days"`
}

// EncryptionConfig controls column-level encryption.
type EncryptionConfig struct {
	// Enabled enables column-level AES-256-GCM encryption.
	Enabled bool `yaml:"enabled"`
	// Columns is the list of column names to encrypt. If empty, defaults are used.
	Columns []string `yaml:"columns"`
}

// SecurityConfig holds security/safety settings.
type SecurityConfig struct {
	// KillSwitchHotkey is the global kill-switch hotkey. Default: "Cmd+Shift+Escape".
	KillSwitchHotkey string `yaml:"kill_switch_hotkey"`
	// AuditRetentionDays is how long audit logs are kept. 0 = use default (90).
	AuditRetentionDays int `yaml:"audit_retention_days"`
	// SpendLimitsUSD is the per-day hard spend cap. 0 = disabled.
	SpendLimitUSDPerDay float64 `yaml:"spend_limit_usd_per_day"`
	// PIIRedaction enables PII redaction in perception. Default: true.
	PIIRedaction bool `yaml:"pii_redaction"`
	// SensitiveApps is a list of app bundle IDs / window classes that
	// the agent must never perceive. Empty uses built-in defaults.
	SensitiveApps []string `yaml:"sensitive_apps"`
}

// RouterConfig controls the LLM router.
type RouterConfig struct {
	// Strategy: "cascade", "pareto", "hybrid", "user".
	Strategy string `yaml:"strategy"`
	// Priorities: per task type, an ordered list of provider names.
	Priorities map[string][]string `yaml:"priorities"`
	// FallbackChain: ordered list of providers to try when the primary is down.
	FallbackChain []string `yaml:"fallback_chain"`
	// MemoryBiasWeight: 0..1, how much to weight historical success.
	MemoryBiasWeight float64 `yaml:"memory_bias_weight"`
	// MinSamplesForBias: minimum samples before memory bias kicks in.
	MinSamplesForBias int `yaml:"min_samples_for_bias"`
}

// LLMConfig holds per-provider LLM settings.
type LLMConfig struct {
	// Providers is keyed by provider name (anthropic, openai, etc.).
	Providers map[string]ProviderConfig `yaml:"providers"`
	// OAuthProviders is keyed by provider name; only providers that
	// officially support user-facing OAuth (Google today) appear here.
	OAuthProviders map[string]OAuthProviderConfig `yaml:"oauth_providers"`
}

// ProviderConfig is a single LLM provider.
type ProviderConfig struct {
	// Enabled toggles this provider.
	Enabled bool `yaml:"enabled"`
	// APIKey is the user's API key. Prefer the api_key manager for runtime.
	// This YAML field is for users who want to keep the key in config.
	APIKey string `yaml:"api_key"`
	// BaseURL overrides the provider's default base URL.
	// (For Custom / Local providers this is required.)
	BaseURL string `yaml:"base_url"`
	// DefaultModel is the model to use when the user doesn't specify one.
	DefaultModel string `yaml:"default_model"`
	// MaxRetries on transient errors. Default: 3.
	MaxRetries int `yaml:"max_retries"`
	// TimeoutSec for a single request. Default: 60.
	TimeoutSec int `yaml:"timeout_sec"`
	// Models is an allowlist of model IDs (empty = any).
	Models []string `yaml:"models"`
	// PricingPerMTokens is a map of model -> [inputUSD, outputUSD] per 1M tokens.
	PricingPerMTokens map[string][2]float64 `yaml:"pricing_per_mtokens"`
}

// OAuthProviderConfig holds OAuth settings for providers that support it.
type OAuthProviderConfig struct {
	// ClientID is the OAuth client ID (issued by us; the user trusts us).
	ClientID string `yaml:"client_id"`
	// Scopes is the list of OAuth scopes to request.
	Scopes []string `yaml:"scopes"`
	// AuthURL and TokenURL override the provider's defaults.
	AuthURL  string `yaml:"auth_url"`
	TokenURL string `yaml:"token_url"`
}

// APIServerConfig controls the IPC API server.
type APIServerConfig struct {
	// Host is the bind address. Default: "127.0.0.1".
	Host string `yaml:"host"`
	// Port is the bind port. Default: 7666.
	Port int `yaml:"port"`
	// TLSEnabled enables HTTPS+TLS. Off by default (localhost HTTP).
	TLSEnabled bool `yaml:"tls_enabled"`
	// AuthToken is a shared secret for remote access. Empty = local-only.
	AuthToken string `yaml:"auth_token"`
	// AllowedOrigins for WebSocket upgrades (CORS). Empty = no CORS.
	AllowedOrigins []string `yaml:"allowed_origins"`
}

// AutonomyConfig is the per-app / per-task autonomy matrix.
type AutonomyConfig struct {
	// DefaultLevel: "supervised", "warn", "autonomous", "block".
	DefaultLevel string `yaml:"default_level"`
	// PerApp is keyed by app bundle ID / window class.
	PerApp map[string]string `yaml:"per_app"`
	// PerTask is keyed by task type (file_operations, web_browsing, etc.).
	PerTask map[string]string `yaml:"per_task"`
	// ShowWarningsForRead enables warnings for low-risk READ actions.
	ShowWarningsForRead bool `yaml:"show_warnings_for_read"`
	// MaxConsecutiveWarnsBeforeAskingAnyway caps repeated warnings.
	MaxConsecutiveWarnsBeforeAskingAnyway int `yaml:"max_consecutive_warns"`
}

// TelemetryConfig controls opt-in telemetry.
type TelemetryConfig struct {
	// Enabled toggles telemetry. Default: false.
	Enabled bool `yaml:"enabled"`
	// CrashReports toggles crash report sending. Default: false.
	CrashReports bool `yaml:"crash_reports"`
}
