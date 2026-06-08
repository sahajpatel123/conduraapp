package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Default()
// -----------------------------------------------------------------------------

func TestDefault_HasSchemaVersion(t *testing.T) {
	cfg := Default()
	assert.Equal(t, ConfigSchemaVersion, cfg.Version)
}

func TestDefault_HasDataDir(t *testing.T) {
	cfg := Default()
	assert.NotEmpty(t, cfg.General.DataDir)
}

func TestDefault_HotkeyEmpty(t *testing.T) {
	cfg := Default()
	// Per CLAUDE.md Decision 8, the user must set the hotkey on first run.
	assert.Empty(t, cfg.Hotkey.Overlay)
}

func TestDefault_AutonomyIsWarn(t *testing.T) {
	cfg := Default()
	// Per CLAUDE.md Decision 24, default autonomy is "warn" (cautious).
	assert.Equal(t, "warn", cfg.Daemon.DefaultAutonomy)
	assert.Equal(t, "warn", cfg.Autonomy.DefaultLevel)
}

func TestDefault_AllProvidersPresent(t *testing.T) {
	cfg := Default()
	// All 12 LLM providers should be present (even if disabled).
	expected := []string{
		"anthropic", "openai", "google", "xai", "mistral", "deepseek",
		"openrouter", "together", "groq", "fireworks", "custom", "ollama",
	}
	for _, name := range expected {
		assert.Contains(t, cfg.LLM.Providers, name, "provider %q should be in defaults", name)
	}
}

func TestDefault_OnlyGoogleHasOAuth(t *testing.T) {
	cfg := Default()
	assert.Contains(t, cfg.LLM.OAuthProviders, "google", "Google should be in default OAuth providers")
	// No other providers should have OAuth by default (per user instruction).
	for name := range cfg.LLM.OAuthProviders {
		assert.Equal(t, "google", name, "only Google should be in default OAuth providers, found %q", name)
	}
}

func TestDefault_PerTaskPriorities(t *testing.T) {
	cfg := Default()
	for task := range cfg.Router.Priorities {
		assert.NotEmpty(t, cfg.Router.Priorities[task], "task %q has empty priority list", task)
	}
}

func TestDefault_FirstRunTrue(t *testing.T) {
	cfg := Default()
	assert.True(t, cfg.General.FirstRun)
}

func TestDefault_TelemetryOff(t *testing.T) {
	cfg := Default()
	assert.False(t, cfg.Telemetry.Enabled)
	assert.False(t, cfg.Telemetry.CrashReports)
}

func TestDefault_EncryptionEnabled(t *testing.T) {
	cfg := Default()
	assert.True(t, cfg.Storage.Encryption.Enabled)
	assert.NotEmpty(t, cfg.Storage.Encryption.Columns)
}

func TestDefault_SpendLimit(t *testing.T) {
	cfg := Default()
	assert.Equal(t, 5.0, cfg.Security.SpendLimitUSDPerDay, "default spend limit is $5/day per CLAUDE.md")
}

func TestDefault_APIServer(t *testing.T) {
	cfg := Default()
	assert.Equal(t, "127.0.0.1", cfg.APIServer.Host)
	assert.Equal(t, 7666, cfg.APIServer.Port)
	assert.False(t, cfg.APIServer.TLSEnabled)
}

// -----------------------------------------------------------------------------
// Validate()
// -----------------------------------------------------------------------------

func TestValidate_OK(t *testing.T) {
	cfg := Default()
	cfg.General.DataDir = t.TempDir()
	assert.NoError(t, cfg.Validate())
}

func TestDefault_StoragePathEmpty(t *testing.T) {
	// Default() should leave Storage.Path empty; it's resolved at load time.
	cfg := Default()
	assert.Empty(t, cfg.Storage.Path)
	assert.Empty(t, cfg.Storage.Backup.Dir)
}

func TestValidate_VersionMismatch(t *testing.T) {
	cfg := Default()
	cfg.General.DataDir = t.TempDir()
	cfg.Version = 999
	err := cfg.Validate()
	require.Error(t, err)
	var verr *ErrInvalidConfig
	require.ErrorAs(t, err, &verr)
	assert.Contains(t, err.Error(), "schema version")
}

func TestValidate_DataDirEmpty(t *testing.T) {
	cfg := Default()
	cfg.General.DataDir = ""
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data_dir")
}

func TestValidate_UnsupportedLanguage(t *testing.T) {
	cfg := Default()
	cfg.General.DataDir = t.TempDir()
	cfg.General.Language = "klingon"
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "language")
}

func TestValidate_AllSupportedLanguages(t *testing.T) {
	for _, lang := range []string{"en-US", "en-GB", "es-ES", "fr-FR", "de-DE", "hi-IN", "ja-JP", "zh-CN"} {
		t.Run(lang, func(t *testing.T) {
			cfg := Default()
			cfg.General.DataDir = t.TempDir()
			cfg.General.Language = lang
			assert.NoError(t, cfg.Validate(), "language %q should be supported", lang)
		})
	}
}

func TestValidate_AutonomyLevels(t *testing.T) {
	for _, level := range []string{"supervised", "warn", "autonomous", "block"} {
		t.Run(level, func(t *testing.T) {
			cfg := Default()
			cfg.General.DataDir = t.TempDir()
			cfg.Autonomy.DefaultLevel = level
			assert.NoError(t, cfg.Validate())
		})
	}
}

func TestValidate_InvalidAutonomy(t *testing.T) {
	cfg := Default()
	cfg.General.DataDir = t.TempDir()
	cfg.Autonomy.DefaultLevel = "medium-rare"
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "default_level")
}

func TestValidate_InvalidPerAppAutonomy(t *testing.T) {
	cfg := Default()
	cfg.General.DataDir = t.TempDir()
	cfg.Autonomy.PerApp["com.bad.app"] = "yolo"
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "per_app")
}

func TestValidate_PortOutOfRange(t *testing.T) {
	cfg := Default()
	cfg.General.DataDir = t.TempDir()
	cfg.APIServer.Port = 100000
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "port")
}

func TestValidate_PublicHostRequiresAuth(t *testing.T) {
	cfg := Default()
	cfg.General.DataDir = t.TempDir()
	cfg.APIServer.Host = "0.0.0.0"
	cfg.APIServer.AuthToken = ""
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth_token")

	// With auth token, it should pass.
	cfg.APIServer.AuthToken = "secret"
	assert.NoError(t, cfg.Validate())
}

func TestValidate_NegativeIdleTimeout(t *testing.T) {
	cfg := Default()
	cfg.General.DataDir = t.TempDir()
	cfg.Daemon.IdleTimeoutMinutes = -1
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "idle_timeout")
}

func TestValidate_NegativeSpendLimit(t *testing.T) {
	cfg := Default()
	cfg.General.DataDir = t.TempDir()
	cfg.Security.SpendLimitUSDPerDay = -1
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "spend_limit")
}

func TestErrInvalidConfig_SingleError(t *testing.T) {
	e := &ErrInvalidConfig{Errors: []string{"one thing"}}
	assert.Equal(t, "invalid config: one thing", e.Error())
}

func TestErrInvalidConfig_MultipleErrors(t *testing.T) {
	e := &ErrInvalidConfig{Errors: []string{"a", "b", "c"}}
	assert.Equal(t, "invalid config: a; b; c", e.Error())
}

// -----------------------------------------------------------------------------
// Loader
// -----------------------------------------------------------------------------

func TestLoader_Load_DefaultsWhenNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(filepath.Join(tmpDir, "nonexistent.yaml"))
	cfg, err := loader.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, ConfigSchemaVersion, cfg.Version)
}

func TestLoader_Load_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	yamlData := `
version: 2
general:
  data_dir: ` + tmpDir + `
  language: fr-FR
hotkey:
  overlay: Cmd+Shift+K
  kill_switch: Ctrl+Alt+\
daemon:
  idle_timeout_minutes: 30
logging:
  level: debug
security:
  spend_limit_usd_per_day: 10.0
`
	require.NoError(t, os.WriteFile(path, []byte(yamlData), 0o600))

	loader := NewLoader(path)
	cfg, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, "fr-FR", cfg.General.Language)
	assert.Equal(t, "Cmd+Shift+K", cfg.Hotkey.Overlay)
	assert.Equal(t, 30, cfg.Daemon.IdleTimeoutMinutes)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, 10.0, cfg.Security.SpendLimitUSDPerDay)
	// Defaults should still apply for unset fields.
	assert.True(t, cfg.Storage.Encryption.Enabled)
}

func TestLoader_Load_ResolvesEmptyPaths(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	yamlData := `
version: 2
general:
  data_dir: ` + tmpDir + `
`
	require.NoError(t, os.WriteFile(path, []byte(yamlData), 0o600))

	loader := NewLoader(path)
	cfg, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "synaptic.db"), cfg.Storage.Path)
	assert.Equal(t, filepath.Join(tmpDir, "backups"), cfg.Storage.Backup.Dir)
}

func TestLoader_Load_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: 2\n  bad indent: oops"), 0o600))
	loader := NewLoader(path)
	_, err := loader.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestLoader_Load_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	yamlData := `
version: 2
general:
  data_dir: ""
`
	require.NoError(t, os.WriteFile(path, []byte(yamlData), 0o600))
	loader := NewLoader(path)
	_, err := loader.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data_dir")
}

func TestLoader_Save_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	loader := NewLoader(path)
	cfg := Default()
	cfg.General.DataDir = tmpDir
	cfg.General.Language = "de-DE"

	require.NoError(t, loader.Save(cfg))

	loaded, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.General.Language, loaded.General.Language)
	assert.Equal(t, tmpDir, loaded.General.DataDir)
}

func TestLoader_Load_EnvOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: 2\ngeneral:\n  data_dir: "+tmpDir+"\n"), 0o600))

	// Convention: __ separates YAML hierarchy; _ is part of a field name.
	t.Setenv("SYNAPTIC_LOGGING__LEVEL", "debug")
	t.Setenv("SYNAPTIC_HOTKEY__OVERLAY", "Ctrl+Space")
	t.Setenv("SYNAPTIC_SECURITY__SPEND_LIMIT_USD_PER_DAY", "20.5")
	t.Setenv("SYNAPTIC_TELEMETRY__ENABLED", "true")
	t.Setenv("SYNAPTIC_API_SERVER__PORT", "9999")
	t.Setenv("SYNAPTIC_GENERAL__LANGUAGE", "ja-JP")

	loader := NewLoader(path)
	cfg, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "Ctrl+Space", cfg.Hotkey.Overlay)
	assert.Equal(t, 20.5, cfg.Security.SpendLimitUSDPerDay)
	assert.True(t, cfg.Telemetry.Enabled)
	assert.Equal(t, 9999, cfg.APIServer.Port)
	assert.Equal(t, "ja-JP", cfg.General.Language)
}

func TestLoader_Load_EnvInvalidBool(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: 2\ngeneral:\n  data_dir: "+tmpDir+"\n"), 0o600))

	t.Setenv("SYNAPTIC_TELEMETRY__ENABLED", "yes-please")
	loader := NewLoader(path)
	_, err := loader.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "env")
}

func TestLoader_Load_EnvUnknownField(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: 2\ngeneral:\n  data_dir: "+tmpDir+"\n"), 0o600))

	t.Setenv("SYNAPTIC_LOGGING__BOGUS", "value")
	loader := NewLoader(path)
	_, err := loader.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown field")
}

func TestLoader_Load_EnvReadOnly(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: 2\ngeneral:\n  data_dir: "+tmpDir+"\n"), 0o600))

	t.Setenv("SYNAPTIC_STORAGE__ENCRYPTION__ENABLED", "false")
	loader := NewLoader(path)
	_, err := loader.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not overridable")
}

// -----------------------------------------------------------------------------
// ResolveStoragePath
// -----------------------------------------------------------------------------

func TestResolveStoragePath_Absolute(t *testing.T) {
	cfg := Default()
	cfg.General.DataDir = t.TempDir()
	cfg.Storage.Path = ""
	p, err := cfg.ResolveStoragePath()
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(p), "expected absolute path, got %q", p)
	assert.True(t, strings.HasSuffix(p, "synaptic.db"))
}

// -----------------------------------------------------------------------------
// defaultDataDir per-OS
// -----------------------------------------------------------------------------

func TestDefaultDataDir(t *testing.T) {
	dir := defaultDataDir()
	// Should be non-empty on all platforms.
	assert.NotEmpty(t, dir)
	if runtime.GOOS != "windows" {
		// On macOS/Linux, default is ~/.synaptic (CLAUDE.md convention).
		assert.Contains(t, dir, ".synaptic")
	}
}

// -----------------------------------------------------------------------------
// isValidAutonomy
// -----------------------------------------------------------------------------

func TestIsValidAutonomy(t *testing.T) {
	for _, level := range []string{"supervised", "warn", "autonomous", "block"} {
		assert.True(t, isValidAutonomy(level), "level %q should be valid", level)
	}
	for _, level := range []string{"", "maybe", "yes", "off"} {
		assert.False(t, isValidAutonomy(level), "level %q should be invalid", level)
	}
}

// -----------------------------------------------------------------------------
// Env setters (per-section)
// -----------------------------------------------------------------------------

func TestSetReflect_AllSections(t *testing.T) {
	// Cover every setXxx function to drive coverage up.
	cfg := &Config{}

	require.NoError(t, setGeneral(&cfg.General, []string{"data_dir"}, "/tmp/x"))
	assert.Equal(t, "/tmp/x", cfg.General.DataDir)

	require.NoError(t, setGeneral(&cfg.General, []string{"cache_dir"}, "/tmp/y"))
	assert.Equal(t, "/tmp/y", cfg.General.CacheDir)

	require.NoError(t, setGeneral(&cfg.General, []string{"language"}, "fr-FR"))
	assert.Equal(t, "fr-FR", cfg.General.Language)

	require.NoError(t, setGeneral(&cfg.General, []string{"install_id"}, "abc-123"))
	assert.Equal(t, "abc-123", cfg.General.InstallID)

	require.NoError(t, setGeneral(&cfg.General, []string{"first_run"}, "true"))
	assert.True(t, cfg.General.FirstRun)

	require.NoError(t, setHotkey(&cfg.Hotkey, []string{"overlay"}, "Cmd+K"))
	assert.Equal(t, "Cmd+K", cfg.Hotkey.Overlay)

	require.NoError(t, setHotkey(&cfg.Hotkey, []string{"kill_switch"}, "Ctrl+Alt+\\"))
	assert.Equal(t, "Ctrl+Alt+\\", cfg.Hotkey.KillSwitch)

	require.NoError(t, setWindow(&cfg.Window, []string{"width"}, "1280"))
	assert.Equal(t, 1280, cfg.Window.Width)

	require.NoError(t, setWindow(&cfg.Window, []string{"height"}, "800"))
	assert.Equal(t, 800, cfg.Window.Height)

	require.NoError(t, setWindow(&cfg.Window, []string{"x"}, "100"))
	assert.Equal(t, 100, cfg.Window.X)

	require.NoError(t, setWindow(&cfg.Window, []string{"y"}, "50"))
	assert.Equal(t, 50, cfg.Window.Y)

	require.NoError(t, setWindow(&cfg.Window, []string{"last_conversation_id"}, "12345"))
	assert.Equal(t, int64(12345), cfg.Window.LastConversationID)

	require.NoError(t, setDaemon(&cfg.Daemon, []string{"auto_start"}, "false"))
	assert.False(t, cfg.Daemon.AutoStart)

	require.NoError(t, setDaemon(&cfg.Daemon, []string{"idle_timeout_minutes"}, "42"))
	assert.Equal(t, 42, cfg.Daemon.IdleTimeoutMinutes)

	require.NoError(t, setDaemon(&cfg.Daemon, []string{"default_autonomy"}, "supervised"))
	assert.Equal(t, "supervised", cfg.Daemon.DefaultAutonomy)

	require.NoError(t, setLogging(&cfg.Logging, []string{"level"}, "debug"))
	assert.Equal(t, "debug", cfg.Logging.Level)

	require.NoError(t, setLogging(&cfg.Logging, []string{"format"}, "json"))
	assert.Equal(t, "json", cfg.Logging.Format)

	require.NoError(t, setLogging(&cfg.Logging, []string{"file"}, "/var/log/syn.log"))
	assert.Equal(t, "/var/log/syn.log", cfg.Logging.File)

	require.NoError(t, setLogging(&cfg.Logging, []string{"add_source"}, "true"))
	assert.True(t, cfg.Logging.AddSource)

	require.NoError(t, setStorage(&cfg.Storage, []string{"path"}, "/tmp/syn.db"))
	assert.Equal(t, "/tmp/syn.db", cfg.Storage.Path)

	require.NoError(t, setBackup(&cfg.Storage.Backup, []string{"on_uninstall"}, "false"))
	assert.False(t, cfg.Storage.Backup.OnUninstall)

	require.NoError(t, setBackup(&cfg.Storage.Backup, []string{"dir"}, "/tmp/bak"))
	assert.Equal(t, "/tmp/bak", cfg.Storage.Backup.Dir)

	require.NoError(t, setBackup(&cfg.Storage.Backup, []string{"retention_days"}, "7"))
	assert.Equal(t, 7, cfg.Storage.Backup.RetentionDays)

	require.NoError(t, setSecurity(&cfg.Security, []string{"audit_retention_days"}, "30"))
	assert.Equal(t, 30, cfg.Security.AuditRetentionDays)

	require.NoError(t, setSecurity(&cfg.Security, []string{"spend_limit_usd_per_day"}, "1.5"))
	assert.Equal(t, 1.5, cfg.Security.SpendLimitUSDPerDay)

	require.NoError(t, setSecurity(&cfg.Security, []string{"pii_redaction"}, "false"))
	assert.False(t, cfg.Security.PIIRedaction)

	require.NoError(t, setRouter(&cfg.Router, []string{"strategy"}, "user"))
	assert.Equal(t, "user", cfg.Router.Strategy)

	require.NoError(t, setRouter(&cfg.Router, []string{"memory_bias_weight"}, "0.5"))
	assert.Equal(t, 0.5, cfg.Router.MemoryBiasWeight)

	require.NoError(t, setRouter(&cfg.Router, []string{"min_samples_for_bias"}, "10"))
	assert.Equal(t, 10, cfg.Router.MinSamplesForBias)

	require.NoError(t, setAPIServer(&cfg.APIServer, []string{"host"}, "0.0.0.0"))
	assert.Equal(t, "0.0.0.0", cfg.APIServer.Host)

	require.NoError(t, setAPIServer(&cfg.APIServer, []string{"port"}, "8080"))
	assert.Equal(t, 8080, cfg.APIServer.Port)

	require.NoError(t, setAPIServer(&cfg.APIServer, []string{"tls_enabled"}, "true"))
	assert.True(t, cfg.APIServer.TLSEnabled)

	require.NoError(t, setAPIServer(&cfg.APIServer, []string{"auth_token"}, "secret"))
	assert.Equal(t, "secret", cfg.APIServer.AuthToken)

	require.NoError(t, setAutonomy(&cfg.Autonomy, []string{"default_level"}, "block"))
	assert.Equal(t, "block", cfg.Autonomy.DefaultLevel)

	require.NoError(t, setAutonomy(&cfg.Autonomy, []string{"show_warnings_for_read"}, "true"))
	assert.True(t, cfg.Autonomy.ShowWarningsForRead)

	require.NoError(t, setAutonomy(&cfg.Autonomy, []string{"max_consecutive_warns"}, "3"))
	assert.Equal(t, 3, cfg.Autonomy.MaxConsecutiveWarnsBeforeAskingAnyway)

	require.NoError(t, setTelemetry(&cfg.Telemetry, []string{"enabled"}, "true"))
	assert.True(t, cfg.Telemetry.Enabled)

	require.NoError(t, setTelemetry(&cfg.Telemetry, []string{"crash_reports"}, "true"))
	assert.True(t, cfg.Telemetry.CrashReports)
}

func TestSetReflect_AllErrors(t *testing.T) {
	cfg := &Config{}

	// errBadPath: too many segments.
	assert.Error(t, setGeneral(&cfg.General, []string{"data_dir", "extra"}, "x"))
	assert.Error(t, setDaemon(&cfg.Daemon, []string{"a", "b"}, "x"))
	assert.Error(t, setLogging(&cfg.Logging, []string{"a", "b"}, "x"))
	assert.Error(t, setBackup(&cfg.Storage.Backup, []string{"a", "b"}, "x"))
	assert.Error(t, setSecurity(&cfg.Security, []string{"a", "b"}, "x"))
	assert.Error(t, setRouter(&cfg.Router, []string{"a", "b"}, "x"))
	assert.Error(t, setAPIServer(&cfg.APIServer, []string{"a", "b"}, "x"))
	assert.Error(t, setAutonomy(&cfg.Autonomy, []string{"a", "b"}, "x"))
	assert.Error(t, setTelemetry(&cfg.Telemetry, []string{"a", "b"}, "x"))
	assert.Error(t, setStorage(&cfg.Storage, []string{"a", "b", "c"}, "x"))

	// errUnknownField.
	assert.Error(t, setGeneral(&cfg.General, []string{"unknown"}, "x"))
	assert.Error(t, setDaemon(&cfg.Daemon, []string{"unknown"}, "x"))
	assert.Error(t, setLogging(&cfg.Logging, []string{"unknown"}, "x"))
	assert.Error(t, setBackup(&cfg.Storage.Backup, []string{"unknown"}, "x"))
	assert.Error(t, setSecurity(&cfg.Security, []string{"unknown"}, "x"))
	assert.Error(t, setRouter(&cfg.Router, []string{"unknown"}, "x"))
	assert.Error(t, setAPIServer(&cfg.APIServer, []string{"unknown"}, "x"))
	assert.Error(t, setAutonomy(&cfg.Autonomy, []string{"unknown"}, "x"))
	assert.Error(t, setTelemetry(&cfg.Telemetry, []string{"unknown"}, "x"))
	assert.Error(t, setStorage(&cfg.Storage, []string{"unknown"}, "x"))

	// errReadOnly.
	assert.Error(t, setStorage(&cfg.Storage, []string{"encryption"}, "x"))

	// Unknown section via setReflect.
	assert.Error(t, setReflect(cfg, []string{"unknown_section", "field"}, "x"))

	// Invalid bool / int.
	assert.Error(t, setGeneral(&cfg.General, []string{"first_run"}, "yes"))
	assert.Error(t, setDaemon(&cfg.Daemon, []string{"idle_timeout_minutes"}, "abc"))
	assert.Error(t, setLogging(&cfg.Logging, []string{"add_source"}, "no"))
	assert.Error(t, setBackup(&cfg.Storage.Backup, []string{"on_uninstall"}, "no"))
	assert.Error(t, setBackup(&cfg.Storage.Backup, []string{"retention_days"}, "abc"))
	assert.Error(t, setSecurity(&cfg.Security, []string{"audit_retention_days"}, "x"))
	assert.Error(t, setSecurity(&cfg.Security, []string{"spend_limit_usd_per_day"}, "x"))
	assert.Error(t, setSecurity(&cfg.Security, []string{"pii_redaction"}, "x"))
	assert.Error(t, setRouter(&cfg.Router, []string{"memory_bias_weight"}, "x"))
	assert.Error(t, setRouter(&cfg.Router, []string{"min_samples_for_bias"}, "x"))
	assert.Error(t, setAPIServer(&cfg.APIServer, []string{"port"}, "x"))
	assert.Error(t, setAPIServer(&cfg.APIServer, []string{"tls_enabled"}, "x"))
	assert.Error(t, setAutonomy(&cfg.Autonomy, []string{"show_warnings_for_read"}, "x"))
	assert.Error(t, setAutonomy(&cfg.Autonomy, []string{"max_consecutive_warns"}, "x"))
	assert.Error(t, setTelemetry(&cfg.Telemetry, []string{"enabled"}, "x"))
	assert.Error(t, setTelemetry(&cfg.Telemetry, []string{"crash_reports"}, "x"))
}

func TestSetByYAMLKey(t *testing.T) {
	cfg := &Config{}
	assert.Error(t, setByYAMLKey(cfg, "no_section", "x"))
	assert.NoError(t, setByYAMLKey(cfg, "general.data_dir", "/tmp/y"))
}

func TestApplyEnvOverrides_EmptyPrefix(t *testing.T) {
	cfg := Default()
	require.NoError(t, applyEnvOverrides(cfg, "SYNAPTIC_NONEXISTENT_"))
}

// -----------------------------------------------------------------------------
// ParseLevel
// -----------------------------------------------------------------------------

func TestParseLevel_All(t *testing.T) {
	assert.Equal(t, "debug", ParseLevel("debug"))
	assert.Equal(t, "debug", ParseLevel("DEBUG"))
	assert.Equal(t, "info", ParseLevel("info"))
	assert.Equal(t, "info", ParseLevel(""))
	assert.Equal(t, "warn", ParseLevel("warn"))
	assert.Equal(t, "warn", ParseLevel("warning"))
	assert.Equal(t, "error", ParseLevel("error"))
	assert.Equal(t, "error", ParseLevel("err"))
	assert.Equal(t, "info", ParseLevel("trace"))
}
