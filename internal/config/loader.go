package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Common YAML field names used across multiple config sections.
// Defined as constants so the goconst linter doesn't flag the
// repeated literals.
const (
	fieldEnabled = "enabled"
)

// -----------------------------------------------------------------------------
// Default config
// -----------------------------------------------------------------------------

// Canonical names for providers, autonomy levels, and log levels.
// These appear in the default config, the priority lists, and the
// validation functions, so we keep them as named constants to avoid
// goconst noise and to make the magic values easy to find.
const (
	// Provider names. Order is meaningful: the priority lists in
	// defaultRouter() read like "preferred first".
	ProviderAnthropic   = "anthropic"
	ProviderAntigravity = "antigravity"
	ProviderOpenAI      = "openai"
	ProviderGoogle      = "google"
	ProviderXAI         = "xai"
	ProviderMistral     = "mistral"
	ProviderDeepSeek    = "deepseek"
	ProviderOpenRouter  = "openrouter"
	ProviderGroq        = "groq"
	ProviderTogether    = "together"
	ProviderFireworks   = "fireworks"
	ProviderOllama      = "ollama"
	ProviderLocalAI     = "localai"
	ProviderLMStudio    = "lmstudio"
	ProviderVLLM        = "vllm"
	ProviderCustom      = "custom"
	ProviderClaudeCode  = "claude_code"
	ProviderCodex       = "codex"

	// Google OAuth endpoints. Public URLs, not secrets.
	oauthGoogleScope    = "https://www.googleapis.com/auth/generative-language" //nolint:gosec // G101
	oauthGoogleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"        //nolint:gosec // G101
	oauthGoogleTokenURL = "https://oauth2.googleapis.com/token"                 //nolint:gosec // G101

	// Autonomy levels. Used as values for Autonomy.DefaultLevel,
	// PerApp, and PerTask.
	AutonomySupervised = "supervised"
	AutonomyWarn       = "warn"
	AutonomyAutonomous = "autonomous"
	AutonomyBlock      = "block"

	// Log levels.
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Default returns a Config populated with sensible defaults.
//
// Note: Storage.Path and Storage.Backup.Dir are left empty in the returned
// Config. They are resolved by the loader (or by ResolveStoragePath) at use
// time based on General.DataDir. This keeps Default() deterministic and
// avoids baked-in paths that may not match the user's actual DataDir.
func Default() *Config {
	cfg := &Config{
		Version: ConfigSchemaVersion,
		General: GeneralConfig{
			InstallID: "", // generated on first run
			DataDir:   defaultDataDir(),
			CacheDir:  defaultCacheDir(),
			Language:  "en-US",
			FirstRun:  true,
		},
		Daemon: DaemonConfig{
			AutoStart:          true,
			IdleTimeoutMinutes: 15,
			DefaultAutonomy:    AutonomyWarn,
		},
		Logging: LoggingConfig{
			Level:     LogLevelInfo,
			Format:    "text",
			File:      "",
			AddSource: false,
		},
		Storage: StorageConfig{
			Path: "", // resolved by loader
			Backup: BackupConfig{
				OnUninstall:   true,
				Dir:           "", // resolved by loader
				RetentionDays: 30,
			},
			Encryption: EncryptionConfig{
				Enabled: true,
				Columns: []string{
					"api_keys.encrypted_key",
					"oauth_tokens.encrypted_access_token",
					"oauth_tokens.encrypted_refresh_token",
					"user_model.value",
				},
			},
		},
		Security: SecurityConfig{
			AuditRetentionDays:  90,
			SpendLimitUSDPerDay: 5.0,
			PIIRedaction:        true,
			SensitiveApps:       []string{}, // built-in defaults used when empty
		},
		Router: RouterConfig{
			Strategy: "hybrid",
			Priorities: map[string][]string{
				"chat":         {ProviderClaudeCode, ProviderOpenAI, ProviderAnthropic, ProviderGoogle, ProviderXAI, ProviderMistral, ProviderDeepSeek, ProviderOpenRouter, ProviderGroq, ProviderTogether, ProviderFireworks, ProviderOllama, ProviderCustom},
				"code":         {ProviderClaudeCode, ProviderCodex, ProviderAntigravity, ProviderOllama, ProviderOpenRouter, ProviderAnthropic, ProviderOpenAI, ProviderCustom},
				"research":     {ProviderClaudeCode, "hermes", "gemini", ProviderOpenRouter, ProviderOpenAI, ProviderAnthropic, ProviderCustom},
				"reasoning":    {ProviderClaudeCode, ProviderOpenAI, ProviderAntigravity, ProviderOllama, ProviderOpenRouter, ProviderAnthropic, ProviderCustom},
				"long_context": {ProviderGoogle, ProviderOllama, ProviderOpenRouter, ProviderAnthropic, ProviderCustom},
				"vision":       {ProviderClaudeCode, ProviderOpenAI, ProviderAntigravity, ProviderGoogle, ProviderOllama, ProviderOpenRouter, ProviderAnthropic, ProviderCustom},
				"image_gen":    {ProviderOpenAI, ProviderAntigravity, ProviderOpenRouter, ProviderCustom},
				"tts":          {ProviderOpenAI, "elevenlabs", ProviderCustom},
				"stt":          {"whisper_local", ProviderOpenAI, ProviderCustom},
				"embedding":    {"local", ProviderOpenAI, ProviderOllama, ProviderCustom},
				"tool_use":     {ProviderClaudeCode, ProviderCodex, ProviderAntigravity, ProviderOpenRouter, ProviderAnthropic, ProviderOpenAI, ProviderCustom},
				"command":      {ProviderClaudeCode, ProviderCodex, ProviderOpenRouter, ProviderAnthropic, ProviderOpenAI, ProviderCustom},
				"browser":      {ProviderClaudeCode, ProviderCodex, ProviderAntigravity, ProviderOpenRouter, ProviderAnthropic, ProviderOpenAI, ProviderCustom},
			},
			FallbackChain:     []string{ProviderOllama, ProviderOpenRouter, ProviderGroq},
			MemoryBiasWeight:  0.3,
			MinSamplesForBias: 5,
		},
		LLM: LLMConfig{
			Providers:      defaultProviders(),
			OAuthProviders: defaultOAuthProviders(),
		},
		APIServer: APIServerConfig{
			Host:           "127.0.0.1",
			Port:           7666,
			TLSEnabled:     false,
			AuthToken:      "",
			AllowedOrigins: []string{},
		},
		Autonomy: AutonomyConfig{
			DefaultLevel: AutonomyWarn,
			PerApp: map[string]string{
				"com.apple.Mail":       AutonomyWarn,
				"com.tinyspeck.chatly": AutonomyWarn,
				"com.google.Chrome":    AutonomyAutonomous,
				"com.apple.finder":     AutonomyAutonomous,
				"com.microsoft.VSCode": AutonomyAutonomous,
			},
			PerTask: map[string]string{
				"coding":           AutonomyWarn,
				"file_operations":  AutonomyWarn,
				"web_browsing":     AutonomyWarn,
				"email":            AutonomyWarn,
				"calendar":         AutonomyWarn,
				"messaging":        AutonomyWarn,
				"shell_commands":   AutonomyWarn,
				"computer_use":     AutonomyWarn,
				"research":         AutonomyAutonomous,
				"image_generation": AutonomyAutonomous,
				"code_review":      AutonomyAutonomous,
			},
			ShowWarningsForRead:                   false,
			MaxConsecutiveWarnsBeforeAskingAnyway: 5,
		},
		Telemetry: TelemetryConfig{
			Enabled:      false,
			CrashReports: false,
		},
		Hub: HubConfig{
			Enabled:        true,
			BaseURL:        "https://hub.synaptic.app",
			AutoUpdate:     false,
			Token:          "",
			PublishKeyPath: "",
		},
		Sync: SyncConfig{
			Enabled:       false, // opt-in: requires pairing
			DeviceName:    "",
			DiscoveryPort: 7667,
			AutoAnnounce:  true,
		},
		Update: UpdateConfig{
			Enabled:     true,
			ManifestURL: "", // resolved to updater.DefaultManifestURL at runtime
		},
		Account: AccountConfig{
			Enabled:    true,
			SessionTTL: 720 * time.Hour, // 30 days
		},
		Reach: ReachConfig{
			Enabled: true,
		},
		Voice: VoiceConfig{
			Enabled:               false,
			PushToTalk:            true,
			Hotkey:                "Cmd+Shift+V",
			BinaryPath:            "",
			ModelPath:             "",
			BinarySHA256:          "",
			ModelSHA256:           "",
			Model:                 "base",
			Language:              "auto",
			SampleRate:            16000,
			Channels:              1,
			VADThreshold:          0.015,
			SilenceTimeoutMs:      1500,
			MaxCaptureDurationSec: 30,
			SpeakerEnabled:        false,
			SpeakerVoice:          "Samantha",
			SpeakerRate:           200,
			Wake: WakeConfig{
				Enabled:     false,
				ModelPath:   "",
				Sensitivity: 0.5,
				Hotword:     "hey condura",
			},
		},
	}
	return cfg
}

// resolveEmptyPaths fills in Storage.Path and Storage.Backup.Dir from DataDir.
// Called by Loader.Load() after YAML merge and env overrides.
//
// Behavior:
//   - If Storage.Path is empty, derive from DataDir.
//   - If Storage.Backup.Dir is empty, derive from DataDir.
//   - If the user explicitly set a non-default path in YAML, preserve it.
func resolveEmptyPaths(c *Config) {
	if c.Storage.Path == "" {
		c.Storage.Path = filepath.Join(c.General.DataDir, "synaptic.db")
	}
	if c.Storage.Backup.Dir == "" {
		c.Storage.Backup.Dir = filepath.Join(c.General.DataDir, "backups")
	}
}

// OverrideDataDir updates the data directory and re-derives any
// storage paths that were previously filled in by resolveEmptyPaths.
// Use this when a CLI flag (e.g. --data-dir) overrides the YAML
// value: it guarantees Storage.Path and Storage.Backup.Dir point
// inside the new data dir.
//
// Paths that the user explicitly set in YAML are left untouched
// (we re-derive only if the path is still under the OLD data dir).
func (c *Config) OverrideDataDir(newDir string) {
	oldDir := c.General.DataDir
	c.General.DataDir = newDir
	// Re-derive the SQLite path if it was the auto-computed default.
	if c.Storage.Path == "" || isUnderDir(c.Storage.Path, oldDir) {
		c.Storage.Path = filepath.Join(newDir, "synaptic.db")
	}
	if c.Storage.Backup.Dir == "" || isUnderDir(c.Storage.Backup.Dir, oldDir) {
		c.Storage.Backup.Dir = filepath.Join(newDir, "backups")
	}
	// Same for the cache dir if it was the default.
	if c.General.CacheDir == "" || isUnderDir(c.General.CacheDir, oldDir) {
		c.General.CacheDir = filepath.Join(newDir, "cache")
	}
}

// isUnderDir reports whether path resolves to a file or directory
// that lives under (or equals) dir. Both arguments are cleaned
// first. Returns false on error.
func isUnderDir(path, dir string) bool {
	if path == "" || dir == "" {
		return false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absDir, absPath)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if strings.HasPrefix(rel, "..") {
		return false
	}
	return true
}

func defaultProviders() map[string]ProviderConfig {
	mk := func(model string) ProviderConfig {
		return ProviderConfig{
			Enabled:      false,
			BaseURL:      "",
			DefaultModel: model,
			MaxRetries:   3,
			TimeoutSec:   60,
		}
	}
	return map[string]ProviderConfig{
		ProviderAnthropic: {
			Enabled: false, BaseURL: "https://api.anthropic.com", DefaultModel: "claude-sonnet-4-5",
			MaxRetries: 3, TimeoutSec: 60,
		},
		ProviderOpenAI: {
			Enabled: false, BaseURL: "https://api.openai.com/v1", DefaultModel: "gpt-5.5",
			MaxRetries: 3, TimeoutSec: 60,
		},
		ProviderGoogle: {
			Enabled: false, BaseURL: "https://generativelanguage.googleapis.com", DefaultModel: "gemini-3.1-pro",
			MaxRetries: 3, TimeoutSec: 60,
		},
		ProviderXAI:        mk("grok-4.3"),
		ProviderMistral:    mk("mistral-large-3"),
		ProviderDeepSeek:   mk("deepseek-v4"),
		ProviderOpenRouter: mk("anthropic/claude-sonnet-4-5"),
		ProviderTogether:   mk("meta-llama/Llama-4-70b-chat-hf"),
		ProviderGroq:       mk("llama-4-70b-chat"),
		ProviderFireworks:  mk("accounts/fireworks/models/llama-v4-70b-chat"),
		ProviderCustom:     {Enabled: false, BaseURL: "", DefaultModel: "", MaxRetries: 3, TimeoutSec: 60},
		ProviderOllama:     {Enabled: false, BaseURL: "http://127.0.0.1:11434/v1", DefaultModel: "llama4", MaxRetries: 1, TimeoutSec: 120},
	}
}

func defaultOAuthProviders() map[string]OAuthProviderConfig {
	// Only Google has a clean official OAuth for end users today.
	// Others can be added as they publish official OAuth.
	return map[string]OAuthProviderConfig{
		ProviderGoogle: {
			ClientID: "",                         //nolint:gosec // G101
			Scopes:   []string{oauthGoogleScope}, //nolint:gosec // G101
			AuthURL:  oauthGoogleAuthURL,         //nolint:gosec // G101
			TokenURL: oauthGoogleTokenURL,        //nolint:gosec // G101
		},
	}
}

// -----------------------------------------------------------------------------
// Defaults that depend on OS
// -----------------------------------------------------------------------------

func defaultDataDir() string {
	// Per CLAUDE.md Decision 28 and the architecture docs, we use ~/.condura/
	// for cross-platform consistency, with the OS-conventional location as a
	// secondary option. We do NOT use the OS-conventional location by default
	// because users explicitly want ~/.condura/ in the docs.
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".condura")
	}
	// Fallback to OS-conventional location.
	return fallbackDataDir()
}

func defaultCacheDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".condura", "cache")
	}
	return filepath.Join(fallbackDataDir(), "cache")
}

func fallbackDataDir() string {
	switch runtime.GOOS {
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support", "condura")
		}
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "condura")
		}
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "condura")
		}
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".config", "condura")
		}
	}
	return ".condura"
}

// -----------------------------------------------------------------------------
// Loader
// -----------------------------------------------------------------------------

// Loader loads configuration from a YAML file with defaults and env overrides.
type Loader struct {
	// Path is the path to the YAML file. If empty, the default location is used.
	Path string
	// EnvPrefix is the prefix for environment variable overrides.
	// Default: CONDURA_.
	EnvPrefix string
}

// NewLoader returns a Loader with the given path and default env prefix.
func NewLoader(path string) *Loader {
	return &Loader{Path: path, EnvPrefix: DefaultEnvPrefix}
}

// Load reads the config file (if it exists), merges with defaults, applies
// env overrides, and validates the result. If the file does not exist, the
// defaults are returned and an empty file is written (so the user can edit).
func (l *Loader) Load() (*Config, error) {
	if l.Path == "" {
		l.Path = filepath.Join(defaultDataDir(), DefaultConfigFileName)
	}
	if l.EnvPrefix == "" {
		l.EnvPrefix = DefaultEnvPrefix
	}

	cfg := Default()

	// If the file exists, unmarshal it on top of defaults.
	if _, err := os.Stat(l.Path); err == nil {
		data, err := os.ReadFile(l.Path)
		if err != nil {
			return nil, fmt.Errorf("read config file %s: %w", l.Path, err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config file %s: %w", l.Path, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("stat config file %s: %w", l.Path, err)
	}

	// Apply env overrides.
	if err := applyEnvOverrides(cfg, l.EnvPrefix); err != nil {
		return nil, fmt.Errorf("apply env overrides: %w", err)
	}

	// Resolve empty paths to defaults.
	resolveEmptyPaths(cfg)

	// Validate.
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// Save writes the config back to disk in YAML form.
func (l *Loader) Save(cfg *Config) error {
	if l.Path == "" {
		l.Path = filepath.Join(defaultDataDir(), DefaultConfigFileName)
	}
	if err := os.MkdirAll(filepath.Dir(l.Path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(l.Path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Env overrides
// -----------------------------------------------------------------------------

// applyEnvOverrides walks env vars matching the prefix and updates cfg.
//
// Convention: CONDURA_<SECTION>__<FIELD>=value, where __ (double underscore)
// separates the YAML hierarchy and _ is part of a field name.
//
// Examples:
//
//	CONDURA_LOGGING__LEVEL=debug                       -> logging.level = LogLevelDebug
//	CONDURA_DAEMON__AUTO_START=true                    -> daemon.auto_start = true
//	CONDURA_API_SERVER__PORT=9000                      -> api_server.port = 9000
//	CONDURA_SECURITY__SPEND_LIMIT_USD_PER_DAY=20.5     -> security.spend_limit_usd_per_day = 20.5
//	CONDURA_LLM__PROVIDERS__ANTHROPIC__ENABLED=true    -> (read-only; use config file)
func applyEnvOverrides(cfg *Config, prefix string) error {
	for _, env := range os.Environ() {
		idx := strings.Index(env, "=")
		if idx < 0 {
			continue
		}
		name, val := env[:idx], env[idx+1:]
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		// Strip prefix, lowercase, and convert __ to . (hierarchy separator).
		// Single underscores within a field name are preserved.
		key := strings.ToLower(strings.TrimPrefix(name, prefix))
		key = strings.ReplaceAll(key, "__", ".")
		if err := setByYAMLKey(cfg, key, val); err != nil {
			return fmt.Errorf("env %s: %w", name, err)
		}
	}
	return nil
}

// setByYAMLKey applies a single value at a dot-separated YAML path.
// Supports simple paths like "logging.level" and "llm.providers.anthropic.enabled".
// Does NOT support map-of-array paths. Returns an error for unknown keys.
func setByYAMLKey(cfg *Config, key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("env key %q is not a valid path (expected section.field)", key)
	}
	// We use reflection in a structured way.
	return setReflect(cfg, parts, value)
}

// -----------------------------------------------------------------------------
// Validation
// -----------------------------------------------------------------------------

// ErrInvalidConfig is returned by Validate for any validation error.
type ErrInvalidConfig struct {
	Errors []string
}

func (e *ErrInvalidConfig) Error() string {
	if len(e.Errors) == 1 {
		return "invalid config: " + e.Errors[0]
	}
	return "invalid config: " + strings.Join(e.Errors, "; ")
}

// Validate checks the config for obvious errors.
func (c *Config) Validate() error {
	errs := make([]string, 0, len(c.validateVersion())+len(c.validateGeneral())+len(c.validateDaemon())+len(c.validateLogging())+len(c.validateStorage())+len(c.validateSecurity())+len(c.validateAPIServer())+len(c.validateAutonomy())+len(c.validateVoice()))
	errs = append(errs, c.validateVersion()...)
	errs = append(errs, c.validateGeneral()...)
	errs = append(errs, c.validateDaemon()...)
	errs = append(errs, c.validateLogging()...)
	errs = append(errs, c.validateStorage()...)
	errs = append(errs, c.validateSecurity()...)
	errs = append(errs, c.validateAPIServer()...)
	errs = append(errs, c.validateAutonomy()...)
	errs = append(errs, c.validateVoice()...)

	if len(errs) > 0 {
		return &ErrInvalidConfig{Errors: errs}
	}
	return nil
}

func (c *Config) validateVersion() []string {
	if c.Version != ConfigSchemaVersion {
		return []string{fmt.Sprintf("config schema version %d != expected %d (please migrate)", c.Version, ConfigSchemaVersion)}
	}
	return nil
}

func (c *Config) validateGeneral() []string {
	var errs []string
	if c.General.DataDir == "" {
		errs = append(errs, "general.data_dir must not be empty")
	}
	switch c.General.Language {
	case "en-US", "en-GB", "es-ES", "fr-FR", "de-DE", "ja-JP", "zh-CN":
		// ok — these map 1:1 to locale files in internal/i18n/locales/.
		// If you add a new locale file, also add its base tag here.
	default:
		errs = append(errs, fmt.Sprintf("general.language %q is not a supported language (en-US, en-GB, es-ES, fr-FR, de-DE, ja-JP, zh-CN)", c.General.Language))
	}
	return errs
}

func (c *Config) validateDaemon() []string {
	var errs []string
	if c.Daemon.IdleTimeoutMinutes < 0 {
		errs = append(errs, "daemon.idle_timeout_minutes must be >= 0")
	}
	switch c.Daemon.DefaultAutonomy {
	case AutonomySupervised, AutonomyWarn, AutonomyAutonomous, "":
		// ok (empty -> use autonomy default)
	default:
		errs = append(errs, fmt.Sprintf("daemon.default_autonomy %q is invalid (supervised, warn, autonomous)", c.Daemon.DefaultAutonomy))
	}
	return errs
}

func (c *Config) validateLogging() []string {
	switch ParseLevel(c.Logging.Level) {
	case LogLevelDebug, LogLevelInfo, AutonomyWarn, LogLevelError:
		return nil
	}
	return []string{fmt.Sprintf("logging.level %q is invalid", c.Logging.Level)}
}

func (c *Config) validateStorage() []string {
	var errs []string
	// Storage.Path may be empty in Default(); it is resolved at load time.
	// If the user set it explicitly, it must be non-empty (resolved by Load).
	if c.Storage.Path != "" {
		if !filepath.IsAbs(c.Storage.Path) && !strings.HasPrefix(c.Storage.Path, "~") {
			errs = append(errs, fmt.Sprintf("storage.path %q must be absolute or start with ~", c.Storage.Path))
		}
	}
	if c.Storage.Backup.RetentionDays < 0 {
		errs = append(errs, "storage.backup.retention_days must be >= 0")
	}
	return errs
}

func (c *Config) validateSecurity() []string {
	var errs []string
	if c.Security.AuditRetentionDays < 0 {
		errs = append(errs, "security.audit_retention_days must be >= 0")
	}
	if c.Security.SpendLimitUSDPerDay < 0 {
		errs = append(errs, "security.spend_limit_usd_per_day must be >= 0")
	}
	return errs
}

func (c *Config) validateAPIServer() []string {
	var errs []string
	if c.APIServer.Port < 0 || c.APIServer.Port > 65535 {
		errs = append(errs, fmt.Sprintf("api_server.port %d is out of range 0-65535", c.APIServer.Port))
	}
	// We refuse to bind to 0.0.0.0 without auth token (basic safety).
	if c.APIServer.Host != "" && c.APIServer.Host != "127.0.0.1" && c.APIServer.Host != "localhost" && c.APIServer.AuthToken == "" {
		errs = append(errs, "api_server.host is non-loopback but api_server.auth_token is empty; refusing to bind publicly without auth")
	}
	return errs
}

func (c *Config) validateAutonomy() []string {
	var errs []string
	for app, level := range c.Autonomy.PerApp {
		if !isValidAutonomy(level) {
			errs = append(errs, fmt.Sprintf("autonomy.per_app[%s] = %q is invalid (supervised, warn, autonomous, block)", app, level))
		}
	}
	for task, level := range c.Autonomy.PerTask {
		if !isValidAutonomy(level) {
			errs = append(errs, fmt.Sprintf("autonomy.per_task[%s] = %q is invalid", task, level))
		}
	}
	if !isValidAutonomy(c.Autonomy.DefaultLevel) && c.Autonomy.DefaultLevel != "" {
		errs = append(errs, fmt.Sprintf("autonomy.default_level %q is invalid", c.Autonomy.DefaultLevel))
	}
	return errs
}

func isValidAutonomy(level string) bool {
	switch level {
	case AutonomySupervised, AutonomyWarn, AutonomyAutonomous, AutonomyBlock:
		return true
	default:
		return false
	}
}

func (c *Config) validateVoice() []string {
	errs := make([]string, 0, len(c.validateVoiceBasic())+len(c.validateVoiceEnabled()))
	errs = append(errs, c.validateVoiceBasic()...)
	errs = append(errs, c.validateVoiceEnabled()...)
	return errs
}

// validateVoiceBasic checks the voice fields that are always
// validated (regardless of whether voice is enabled).
func (c *Config) validateVoiceBasic() []string {
	var errs []string
	validModels := map[string]bool{"tiny": true, "base": true, "small": true, "medium": true}
	if c.Voice.Model != "" && !validModels[c.Voice.Model] {
		errs = append(errs, fmt.Sprintf("voice.model %q is not a valid whisper model (tiny, base, small, medium)", c.Voice.Model))
	}
	if c.Voice.SampleRate < 0 {
		errs = append(errs, "voice.sample_rate must be non-negative")
	}
	if c.Voice.Channels < 0 || c.Voice.Channels > 2 {
		errs = append(errs, "voice.channels must be 0, 1, or 2")
	}
	if c.Voice.VADThreshold < 0 || c.Voice.VADThreshold > 1 {
		errs = append(errs, "voice.vad_threshold must be between 0 and 1")
	}
	if c.Voice.SilenceTimeoutMs < 0 {
		errs = append(errs, "voice.silence_timeout_ms must be non-negative")
	}
	if c.Voice.MaxCaptureDurationSec < 0 {
		errs = append(errs, "voice.max_capture_duration_sec must be non-negative")
	}
	if c.Voice.SpeakerRate < 0 {
		errs = append(errs, "voice.speaker_rate must be non-negative")
	}
	return errs
}

// validateVoiceEnabled checks the voice fields that are required
// only when voice is enabled. 6A-3 (Phase 6): binary and model
// paths are required so the pipeline can be built from config.
// SHA256 pins are optional in dev, but a production deployment
// should set them.
func (c *Config) validateVoiceEnabled() []string { //nolint:gocyclo // explicit checks
	if !c.Voice.Enabled {
		return nil
	}
	var errs []string
	if c.Voice.BinaryPath == "" {
		errs = append(errs, "voice.enabled is true but voice.binary_path is empty")
	}
	if c.Voice.ModelPath == "" {
		errs = append(errs, "voice.enabled is true but voice.model_path is empty")
	}
	if c.Voice.SampleRate == 0 {
		errs = append(errs, "voice.enabled is true but voice.sample_rate is 0")
	}
	if c.Voice.Channels == 0 {
		errs = append(errs, "voice.enabled is true but voice.channels is 0")
	}
	return errs
}

// ParseLevel is exposed here for use by the logger package and others.
func ParseLevel(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case LogLevelDebug:
		return LogLevelDebug
	case LogLevelInfo, "":
		return LogLevelInfo
	case AutonomyWarn, "warning":
		return AutonomyWarn
	case LogLevelError, "err":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// -----------------------------------------------------------------------------
// Reflection-based env setter
// -----------------------------------------------------------------------------

// setReflect sets a value at a dot-separated path using reflection.
// It supports: simple fields, *int, *string, *bool, *float64, map[string]X.
// It does NOT support slices at non-leaf positions (use the config file for those).
func setReflect(root any, parts []string, value string) error {
	// We use a small set of hand-written setters for clarity and type safety
	// (reflection-based generic setters are notoriously hard to keep correct).
	switch parts[0] {
	case "general":
		return setGeneral(&root.(*Config).General, parts[1:], value)
	case "daemon":
		return setDaemon(&root.(*Config).Daemon, parts[1:], value)
	case "logging":
		return setLogging(&root.(*Config).Logging, parts[1:], value)
	case "storage":
		return setStorage(&root.(*Config).Storage, parts[1:], value)
	case "security":
		return setSecurity(&root.(*Config).Security, parts[1:], value)
	case "router":
		return setRouter(&root.(*Config).Router, parts[1:], value)
	case "apiserver", "api_server":
		return setAPIServer(&root.(*Config).APIServer, parts[1:], value)
	case "autonomy":
		return setAutonomy(&root.(*Config).Autonomy, parts[1:], value)
	case "telemetry":
		return setTelemetry(&root.(*Config).Telemetry, parts[1:], value)
	case "hotkey":
		return setHotkey(&root.(*Config).Hotkey, parts[1:], value)
	case "window":
		return setWindow(&root.(*Config).Window, parts[1:], value)
	case "voice":
		return setVoice(&root.(*Config).Voice, parts[1:], value)
	case "hub":
		return setHub(&root.(*Config).Hub, parts[1:], value)
	case "sync":
		return setSync(&root.(*Config).Sync, parts[1:], value)
	default:
		return fmt.Errorf("unknown section %q", parts[0])
	}
}

func setGeneral(c *GeneralConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("general", parts)
	}
	switch parts[0] {
	case "data_dir":
		c.DataDir = value
	case "cache_dir":
		c.CacheDir = value
	case "language":
		c.Language = value
	case "install_id":
		c.InstallID = value
	case "first_run":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.FirstRun = b
	default:
		return errUnknownField("general", parts[0])
	}
	return nil
}

func setDaemon(c *DaemonConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("daemon", parts)
	}
	switch parts[0] {
	case "auto_start":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.AutoStart = b
	case "idle_timeout_minutes":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.IdleTimeoutMinutes = n
	case "default_autonomy":
		c.DefaultAutonomy = value
	default:
		return errUnknownField("daemon", parts[0])
	}
	return nil
}

func setLogging(c *LoggingConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("logging", parts)
	}
	switch parts[0] {
	case "level":
		c.Level = value
	case "format":
		c.Format = value
	case "file":
		c.File = value
	case "add_source":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.AddSource = b
	default:
		return errUnknownField("logging", parts[0])
	}
	return nil
}

func setStorage(c *StorageConfig, parts []string, value string) error {
	if len(parts) < 1 {
		return errBadPath("storage", parts)
	}
	switch parts[0] {
	case "path":
		if len(parts) != 1 {
			return errBadPath("storage", parts)
		}
		c.Path = value
	case "backup":
		return setBackup(&c.Backup, parts[1:], value)
	case "encryption":
		return errReadOnly("storage.encryption columns list (edit config file)")
	default:
		return errUnknownField("storage", parts[0])
	}
	return nil
}

func setBackup(c *BackupConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("storage.backup", parts)
	}
	switch parts[0] {
	case "on_uninstall":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.OnUninstall = b
	case "dir":
		c.Dir = value
	case "retention_days":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.RetentionDays = n
	default:
		return errUnknownField("storage.backup", parts[0])
	}
	return nil
}

func setSecurity(c *SecurityConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("security", parts)
	}
	switch parts[0] {
	case "audit_retention_days":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.AuditRetentionDays = n
	case "spend_limit_usd_per_day":
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		c.SpendLimitUSDPerDay = f
	case "pii_redaction":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.PIIRedaction = b
	default:
		return errUnknownField("security", parts[0])
	}
	return nil
}

func setRouter(c *RouterConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("router", parts)
	}
	switch parts[0] {
	case "strategy":
		c.Strategy = value
	case "memory_bias_weight":
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		c.MemoryBiasWeight = f
	case "min_samples_for_bias":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.MinSamplesForBias = n
	default:
		return errUnknownField("router", parts[0])
	}
	return nil
}

func setAPIServer(c *APIServerConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("api_server", parts)
	}
	switch parts[0] {
	case "host":
		c.Host = value
	case "port":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.Port = n
	case "tls_enabled":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.TLSEnabled = b
	case "auth_token":
		c.AuthToken = value
	default:
		return errUnknownField("api_server", parts[0])
	}
	return nil
}

func setAutonomy(c *AutonomyConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("autonomy", parts)
	}
	switch parts[0] {
	case "default_level":
		c.DefaultLevel = value
	case "show_warnings_for_read":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.ShowWarningsForRead = b
	case "max_consecutive_warns":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.MaxConsecutiveWarnsBeforeAskingAnyway = n
	default:
		return errUnknownField("autonomy", parts[0])
	}
	return nil
}

func setTelemetry(c *TelemetryConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("telemetry", parts)
	}
	switch parts[0] {
	case fieldEnabled:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.Enabled = b
	case "crash_reports":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.CrashReports = b
	default:
		return errUnknownField("telemetry", parts[0])
	}
	return nil
}

// errBadPath returns a consistent error for malformed env paths.
func errBadPath(section string, parts []string) error {
	return fmt.Errorf("env path under %q has %d segments, expected 1: %v", section, len(parts), parts)
}

// errUnknownField returns a consistent error for unknown env fields.
func errUnknownField(section, field string) error {
	return fmt.Errorf("unknown field %q in section %q", field, section)
}

// errReadOnly returns a consistent error for env-set fields that are not env-overridable.
func errReadOnly(path string) error {
	return fmt.Errorf("field %q is not overridable via env (edit config file)", path)
}

// setHotkey handles writes to the hotkey section.
func setHotkey(c *HotkeyConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("hotkey", parts)
	}
	switch parts[0] {
	case "overlay":
		c.Overlay = value
	case "kill_switch":
		c.KillSwitch = value
	default:
		return errUnknownField("hotkey", parts[0])
	}
	return nil
}

// setWindow handles writes to the window section.
func setWindow(c *WindowConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("window", parts)
	}
	switch parts[0] {
	case "width":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.Width = n
	case "height":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.Height = n
	case "x":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.X = n
	case "y":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.Y = n
	case "last_conversation_id":
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		c.LastConversationID = n
	default:
		return errUnknownField("window", parts[0])
	}
	return nil
}

// setVoice handles writes to the voice section.
func setVoice(c *VoiceConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("voice", parts)
	}
	field := parts[0]
	switch field {
	case "enabled", "push_to_talk", "speaker_enabled":
		return setVoiceBoolField(c, field, value)
	case "hotkey", "model", "language", "speaker_voice",
		"binary_path", "model_path", "binary_sha256", "model_sha256":
		setVoiceStringField(c, field, value)
	case "vad_threshold":
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		c.VADThreshold = f
	default:
		return setVoiceIntField(c, field, value)
	}
	return nil
}

func setVoiceBoolField(c *VoiceConfig, field, value string) error {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	switch field {
	case fieldEnabled:
		c.Enabled = b
	case "push_to_talk":
		c.PushToTalk = b
	case "speaker_enabled":
		c.SpeakerEnabled = b
	}
	return nil
}

func setVoiceStringField(c *VoiceConfig, field, value string) {
	switch field {
	case "hotkey":
		c.Hotkey = value
	case "model":
		c.Model = value
	case "language":
		c.Language = value
	case "speaker_voice":
		c.SpeakerVoice = value
	case "binary_path":
		c.BinaryPath = value
	case "model_path":
		c.ModelPath = value
	case "binary_sha256":
		c.BinarySHA256 = value
	case "model_sha256":
		c.ModelSHA256 = value
	}
}

func setVoiceIntField(c *VoiceConfig, field, value string) error {
	n, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	switch field {
	case "sample_rate":
		c.SampleRate = n
	case "channels":
		c.Channels = n
	case "silence_timeout_ms":
		c.SilenceTimeoutMs = n
	case "max_capture_duration_sec":
		c.MaxCaptureDurationSec = n
	case "speaker_rate":
		c.SpeakerRate = n
	default:
		return errUnknownField("voice", field)
	}
	return nil
}

// setHub sets a HubConfig field via env var or direct setter.
func setHub(c *HubConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("hub", parts)
	}
	switch parts[0] {
	case fieldEnabled:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.Enabled = b
	case "base_url":
		c.BaseURL = value
	case "auto_update":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.AutoUpdate = b
	case "token":
		c.Token = value
	case "publish_key_path":
		c.PublishKeyPath = value
	default:
		return errUnknownField("hub", parts[0])
	}
	return nil
}

// setSync sets a SyncConfig field via env var or direct setter.
func setSync(c *SyncConfig, parts []string, value string) error {
	if len(parts) != 1 {
		return errBadPath("sync", parts)
	}
	switch parts[0] {
	case fieldEnabled:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.Enabled = b
	case "device_name":
		c.DeviceName = value
	case "discovery_port":
		n, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		c.DiscoveryPort = n
	case "auto_announce":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		c.AutoAnnounce = b
	default:
		return errUnknownField("sync", parts[0])
	}
	return nil
}

// ResolveStoragePath returns the absolute path for the SQLite DB.
func (c *Config) ResolveStoragePath() (string, error) {
	p := c.Storage.Path
	if p == "" {
		p = filepath.Join(c.General.DataDir, "synaptic.db")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return abs, nil
}
