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

import (
	"errors"
	"runtime"
	"time"
)

// ConfigSchemaVersion is the current schema version. Bump on breaking changes.
const ConfigSchemaVersion = 4

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

	// Hotkey combos (overlay, kill switch).
	Hotkey HotkeyConfig `yaml:"hotkey"`

	// Persisted GUI window state.
	Window WindowConfig `yaml:"window"`

	// Voice input/output configuration.
	Voice VoiceConfig `yaml:"voice"`

	// Skills Hub configuration.
	Hub HubConfig `yaml:"hub"`

	// P2P Sync configuration.
	Sync SyncConfig `yaml:"sync"`
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
	// RollbackWindow is how far back RevertLastSession looks. 0 = 1h default.
	RollbackWindow time.Duration `yaml:"rollback_window"`
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
	// Endpoint is the URL to POST anonymous events to.
	Endpoint string `yaml:"endpoint"`
}

// HotkeyConfig holds the global hotkey combos.
type HotkeyConfig struct {
	// Overlay is the hotkey for the chat overlay (default: Cmd+Shift+Space).
	Overlay string `yaml:"overlay"`
	// KillSwitch is the hotkey for the kill switch (default: Cmd+Shift+Escape).
	KillSwitch string `yaml:"kill_switch"`
}

// WindowConfig holds the persisted GUI window state.
type WindowConfig struct {
	Width              int   `yaml:"width"`
	Height             int   `yaml:"height"`
	X                  int   `yaml:"x"`
	Y                  int   `yaml:"y"`
	LastConversationID int64 `yaml:"last_conversation_id"`
}

// VoiceConfig controls speech recognition and synthesis.
type VoiceConfig struct {
	// Enabled toggles voice I/O. Default: false.
	Enabled bool `yaml:"enabled"`

	// PushToTalk toggles push-to-talk mode (hold key to speak).
	// When false, voice is always-on (voice-activated).
	PushToTalk bool `yaml:"push_to_talk"`

	// Hotkey is the push-to-talk key combo (default: "Cmd+Shift+V").
	Hotkey string `yaml:"hotkey"`

	// BinaryPath is the path to the whisper-cli binary. Required
	// when Enabled is true. The pipeline's SHA256 verification
	// will refuse to load a binary whose hash doesn't match
	// BinarySHA256.
	BinaryPath string `yaml:"binary_path"`

	// ModelPath is the path to the whisper model file (e.g. ggml-base.bin).
	// Required when Enabled is true. The pipeline's SHA256
	// verification will refuse to load a model whose hash doesn't
	// match ModelSHA256.
	ModelPath string `yaml:"model_path"`

	// BinarySHA256 is the expected hex-encoded SHA256 of the
	// whisper binary at BinaryPath. Empty disables the check
	// (development only). Production deployments must set this.
	BinarySHA256 string `yaml:"binary_sha256"`

	// ModelSHA256 is the expected hex-encoded SHA256 of the
	// whisper model at ModelPath. Empty disables the check
	// (development only). Production deployments must set this.
	ModelSHA256 string `yaml:"model_sha256"`

	// Model is the whisper model variant: "tiny", "base", "small", "medium".
	// Default: "base" (~142 MB, multilingual, good accuracy/speed balance).
	Model string `yaml:"model"`

	// Language is the BCP-47 language code for whisper (default: "auto" for auto-detect).
	Language string `yaml:"language"`

	// SampleRate for audio capture. Default: 16000 (whisper's native rate).
	SampleRate int `yaml:"sample_rate"`

	// Channels for audio capture. Default: 1 (mono).
	Channels int `yaml:"channels"`

	// VADThreshold is the energy threshold for voice activity detection.
	// Range: 0.0 (silence) to 1.0 (loud). Default: 0.015.
	VADThreshold float64 `yaml:"vad_threshold"`

	// SilenceTimeoutMs is how many ms of silence before we stop capture.
	// Default: 1500.
	SilenceTimeoutMs int `yaml:"silence_timeout_ms"`

	// MaxCaptureDurationSec is the max seconds of audio before forced stop.
	// Default: 30.
	MaxCaptureDurationSec int `yaml:"max_capture_duration_sec"`

	// SpeakerEnabled toggles TTS output. Default: false.
	SpeakerEnabled bool `yaml:"speaker_enabled"`

	// SpeakerVoice is the macOS voice name for TTS (default: "Samantha").
	SpeakerVoice string `yaml:"speaker_voice"`

	// SpeakerRate is the TTS speaking rate (default: 200 words per minute).
	SpeakerRate int `yaml:"speaker_rate"`
}

// Validate checks the voice config for internal consistency. The
// pipeline construction in Phase 6B uses these errors to fail
// fast at startup.
func (c VoiceConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.BinaryPath == "" {
		return errors.New("voice.enabled is true but voice.binary_path is empty")
	}
	if c.ModelPath == "" {
		return errors.New("voice.enabled is true but voice.model_path is empty")
	}
	if c.SampleRate == 0 {
		return errors.New("voice.sample_rate must be > 0 when voice is enabled")
	}
	if c.Channels == 0 {
		return errors.New("voice.channels must be > 0 when voice is enabled")
	}
	return nil
}

// ApplyDefaults fills in zero-valued fields with their documented
// defaults. Call this after loading config but before validation.
func (c *VoiceConfig) ApplyDefaults() {
	if c.Hotkey == "" {
		c.Hotkey = "Cmd+Shift+V"
	}
	if c.Model == "" {
		c.Model = "base"
	}
	if c.SampleRate == 0 {
		c.SampleRate = 16000
	}
	if c.Channels == 0 {
		c.Channels = 1
	}
	if c.VADThreshold == 0 {
		c.VADThreshold = 0.015
	}
	if c.SilenceTimeoutMs == 0 {
		c.SilenceTimeoutMs = 1500
	}
	if c.MaxCaptureDurationSec == 0 {
		c.MaxCaptureDurationSec = 30
	}
	if c.SpeakerEnabled && c.SpeakerVoice == "" {
		c.SpeakerVoice = "Samantha"
	}
	if c.SpeakerEnabled && c.SpeakerRate == 0 {
		c.SpeakerRate = 200
	}
}

// PlatformIsMac returns true if the daemon is running on macOS. It
// is split out from runtime.GOOS so callers (hotkey defaults, tray
// title, code signing) don't have to import "runtime".
func PlatformIsMac() bool { return runtime.GOOS == "darwin" }

// PlatformIsWindows returns true if the daemon is running on Windows.
func PlatformIsWindows() bool { return runtime.GOOS == "windows" }

// PlatformIsLinux returns true if the daemon is running on Linux.
func PlatformIsLinux() bool { return runtime.GOOS == "linux" }

// HubConfig controls the Skills Hub connection.
type HubConfig struct {
	// Enabled toggles hub connectivity. Default: true.
	Enabled bool `yaml:"enabled"`
	// BaseURL is the hub server URL. Default: "https://hub.synaptic.app".
	BaseURL string `yaml:"base_url"`
	// AutoUpdate enables automatic skill updates. Default: false.
	AutoUpdate bool `yaml:"auto_update"`
	// Token is an optional bearer token for authenticated requests.
	// When empty, anonymous requests are made (the server may reject
	// them depending on the deployment's policy).
	Token string `yaml:"token"`
	// PublishKeyPath is the path to a file containing the Ed25519
	// private key used to sign Publish requests (hex-encoded). When
	// empty, Publish is sent unsigned and the server may reject it.
	PublishKeyPath string `yaml:"publish_key_path"`
}

// SyncConfig controls P2P synchronization.
type SyncConfig struct {
	// Enabled toggles P2P sync. Default: false.
	Enabled bool `yaml:"enabled"`
	// DeviceName is the human-readable name for this device.
	DeviceName string `yaml:"device_name"`
	// DiscoveryPort is the UDP port for LAN discovery. Default: 7667.
	DiscoveryPort int `yaml:"discovery_port"`
	// AutoAnnounce enables periodic LAN broadcast. Default: true.
	AutoAnnounce bool `yaml:"auto_announce"`
}
