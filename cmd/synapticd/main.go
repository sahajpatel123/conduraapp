// Command synapticd is the Synaptic daemon.
//
// It owns the on-disk database, the OS keyring, and the LLM registry.
// Clients (the Wails GUI, the CLI, the future Skills Hub) talk to it
// over JSON-RPC 2.0 on a local TCP or Unix socket.
//
// Typical lifecycle:
//
//	synapticd --config ~/.synaptic/config.yaml
//	synapticd --print-default-config > ~/.synaptic/config.yaml
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/synapticapp/synaptic/internal/api_key"
	"github.com/synapticapp/synaptic/internal/config"
	"github.com/synapticapp/synaptic/internal/failover"
	"github.com/synapticapp/synaptic/internal/health"
	"github.com/synapticapp/synaptic/internal/ipc"
	"github.com/synapticapp/synaptic/internal/llm"
	"github.com/synapticapp/synaptic/internal/logger"
	"github.com/synapticapp/synaptic/internal/secrets"
	"github.com/synapticapp/synaptic/internal/storage"
	"github.com/synapticapp/synaptic/internal/version"
)

// File mode for the synapticd.addr sidecar. Owner-only because it
// contains a loopback port that the CLI reads to find the daemon;
// leaking it isn't catastrophic, but no other user should need it.
const addrFilePerm = 0o600

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "synapticd: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		cfgPath       = flag.String("config", "", "path to config.yaml (default: ~/.synaptic/config.yaml)")
		dataDir       = flag.String("data-dir", "", "data directory (overrides config)")
		logLevel      = flag.String("log-level", "", "debug | info | warn | error (default: config value)")
		listen        = flag.String("listen", "tcp://127.0.0.1:0", "IPC listen address")
		noIPC         = flag.Bool("no-ipc", false, "disable IPC server (debugging)")
		printDefaults = flag.Bool("print-default-config", false, "print default config YAML to stdout and exit")
		printVersion  = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *printVersion {
		fmt.Println(version.String())
		return nil
	}
	if *printDefaults {
		out, err := yaml.Marshal(config.Default())
		if err != nil {
			return fmt.Errorf("marshal default config: %w", err)
		}
		fmt.Print(string(out))
		return nil
	}

	// Resolve config path.
	if *cfgPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("locate home dir: %w", err)
		}
		*cfgPath = filepath.Join(home, ".synaptic", "config.yaml")
	}

	// Load config (creates default file if missing).
	loader := config.NewLoader(*cfgPath)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if *dataDir != "" {
		cfg.OverrideDataDir(*dataDir)
	}
	// Re-resolve storage path now that data_dir may have changed.
	if sp, err := cfg.ResolveStoragePath(); err == nil {
		cfg.Storage.Path = sp
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Make sure data dir exists (storage + secrets files live here).
	if err := os.MkdirAll(cfg.General.DataDir, 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	// Init logger.
	log := logger.New(logger.Config{
		Level:     logger.ParseLevel(override(*logLevel, cfg.Logging.Level)),
		Format:    logger.ParseFormat(cfg.Logging.Format),
		File:      cfg.Logging.File,
		AddSource: cfg.Logging.AddSource,
	})
	slog.SetDefault(log)
	ver := version.Get()
	log.Info("starting synapticd",
		"version", ver.Version,
		"commit", ver.Commit,
		"build_date", ver.BuildDate,
		"go", ver.GoVersion,
		"platform", ver.Platform,
		"config", *cfgPath,
		"data_dir", cfg.General.DataDir,
		"storage_path", cfg.Storage.Path,
	)

	// Init secrets manager.
	secretsPath := filepath.Join(cfg.General.DataDir, "secrets.json")
	sm, err := secrets.New(secretsPath)
	if err != nil {
		return fmt.Errorf("init secrets: %w", err)
	}
	log.Info("secrets manager ready", "backend", string(sm.Backend()))

	// Init storage.
	db, err := storage.Open(context.Background(), storage.Config{
		Path:    cfg.Storage.Path,
		Secrets: sm,
	})
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}
	defer func() { _ = db.Close() }()
	log.Info("storage ready", "path", cfg.Storage.Path)

	// Init api_key manager.
	akm := api_key.New(db, sm)

	// Init LLM registry from configured providers + saved keys.
	registry := llm.NewRegistry()
	registered := buildProvidersFromConfig(log, registry, cfg, akm)
	log.Info("llm registry ready", "registered_providers", registered)

	// Init failover.
	mon := failover.NewSpendMonitor(failover.SpendCap{
		USDPerDay: cfg.Security.SpendLimitUSDPerDay,
	})
	breakers := failover.NewBreakerRegistry(3, 30*time.Second)
	failoverProviders := buildFailoverProviders(registry, breakers, cfg)
	fo := failover.New(failoverProviders, mon)
	log.Info("failover ready", "providers", len(failoverProviders))

	// Init health.
	hr := health.New()
	hr.Add(health.Check{
		Name: "storage", Required: true, Timeout: 2 * time.Second,
		Check: func(ctx context.Context) error {
			return db.SQL().PingContext(ctx)
		},
	})
	hr.Add(health.Check{
		Name: "secrets", Required: true, Timeout: 2 * time.Second,
		Check: func(_ context.Context) error {
			_, err := sm.Get("__synaptic_health_probe__")
			if err != nil && !errors.Is(err, secrets.ErrNotFound) {
				return err
			}
			return nil
		},
	})

	// Init IPC server.
	ipcSrv := ipc.NewServer()
	registerMethods(ipcSrv, log, cfg, db, sm, akm, registry, fo, mon, hr, ver)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Add an IPC health check now that we have a context.
	hr.Add(health.Check{
		Name: "ipc", Required: false, Timeout: 1 * time.Second,
		Check: func(_ context.Context) error { return nil },
	})

	ipcT := &ipc.ServerTransport{S: ipcSrv, Token: cfg.APIServer.AuthToken}
	var listenAddrs []string
	if !*noIPC {
		if err := ipcT.Listen(rootCtx, *listen); err != nil {
			return fmt.Errorf("ipc listen: %w", err)
		}
		listenAddrs = append(listenAddrs, ipcT.Addr())
		log.Info("ipc listening", "addr", ipcT.Addr())
		// Also bind a Unix socket (macOS/Linux) for fast local access.
		if runtime.GOOS != "windows" {
			unixPath := filepath.Join(cfg.General.DataDir, "synapticd.sock")
			_ = os.Remove(unixPath)
			if err := ipcT.Listen(rootCtx, "unix://"+unixPath); err != nil {
				log.Warn("unix socket bind failed; continuing", "err", err)
			} else {
				listenAddrs = append(listenAddrs, "unix://"+unixPath)
				log.Info("ipc unix socket ready", "path", unixPath)
			}
		}
	}

	// Print the listen addrs to a sidecar file so the CLI can find us
	// without scanning a port range.
	if len(listenAddrs) > 0 {
		path := filepath.Join(cfg.General.DataDir, "synapticd.addr")
		_ = os.WriteFile(path, []byte(listenAddrs[0]+"\n"), addrFilePerm)
		log.Info("address file written", "path", path)
	}

	// Signal handling.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigCh
		log.Info("signal received; shutting down", "signal", s.String())
		cancel()
		_ = ipcT.Close()
	}()

	// Block until cancelled.
	<-rootCtx.Done()
	log.Info("synapticd stopped")
	return nil
}

func override(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// buildProvidersFromConfig reads the LLM config section and registers
// one provider per entry that has a stored API key. Returns the number
// of providers that were actually registered.
func buildProvidersFromConfig(log *slog.Logger, registry *llm.Registry, cfg *config.Config, akm *api_key.Manager) int {
	count := 0
	for name, p := range cfg.LLM.Providers {
		if !p.Enabled {
			continue
		}
		models := modelsForProvider(name)
		keys, err := akm.ListByProvider(context.Background(), name)
		if err != nil {
			log.Warn("list keys failed", "provider", name, "err", err)
			continue
		}
		if len(keys) == 0 {
			log.Info("no api keys for provider; skipping", "provider", name)
			continue
		}
		key := keys[0].Secret
		prov := buildProvider(name, key, models)
		if prov == nil {
			log.Warn("unknown provider", "provider", name)
			continue
		}
		registry.Register(prov)
		log.Info("registered provider", "provider", name, "models", len(prov.Models()))
		count++
	}
	return count
}

func buildProvider(name, key string, models []llm.ModelInfo) llm.Provider {
	switch name {
	case "openai":
		return llm.NewOpenAI(key, models)
	case "openrouter":
		return llm.NewOpenRouter(key, models)
	case "together":
		return llm.NewTogether(key, models)
	case "groq":
		return llm.NewGroq(key, models)
	case "fireworks":
		return llm.NewFireworks(key, models)
	case "deepseek":
		return llm.NewDeepSeek(key, models)
	case "xai":
		return llm.NewXAI(key, models)
	case "mistral":
		return llm.NewMistral(key, models)
	case "ollama":
		return llm.NewOllama("", models)
	case "anthropic":
		return llm.NewAnthropic(key, models)
	case "google":
		return llm.NewGoogle(key, models)
	case "custom":
		return llm.NewCustom("custom", "http://localhost:9999/v1", key, models)
	}
	return nil
}

// modelsForProvider returns the hard-coded model list for a provider.
// In a future phase this can be loaded from the LLM section of the config.
func modelsForProvider(name string) []llm.ModelInfo {
	out := []llm.ModelInfo{}
	for _, m := range allModels {
		if m.Provider == name {
			out = append(out, m.ModelInfo)
		}
	}
	return out
}

// buildFailoverProviders constructs failover.Provider entries with one
// breaker per LLM provider, in the priority order from cfg.Router.Priorities.
func buildFailoverProviders(registry *llm.Registry, br *failover.BreakerRegistry, cfg *config.Config) []failover.Provider {
	priorities := cfg.Router.Priorities["chat"]
	if len(priorities) == 0 {
		// Fall back to the providers list.
		for name, p := range cfg.LLM.Providers {
			if p.Enabled {
				priorities = append(priorities, name)
			}
		}
	}
	var out []failover.Provider
	for _, name := range priorities {
		prov, ok := registry.Get(name)
		if !ok {
			continue
		}
		models := prov.Models()
		if len(models) == 0 {
			continue
		}
		var names []string
		for _, m := range models {
			names = append(names, m.ID)
		}
		out = append(out, failover.Provider{
			Name:    name,
			Breaker: br.For(name),
			Client:  &llmAdapter{prov: prov, defaultModel: models[0].ID},
			Models:  names,
		})
	}
	return out
}

// llmAdapter wraps an llm.Provider to satisfy failover.LLMClient.
// The failover layer is provider-agnostic; it just calls Chat(provider, model).
type llmAdapter struct {
	prov         llm.Provider
	defaultModel string
}

func (a *llmAdapter) Chat(ctx context.Context, _, model string) (failover.Usage, error) {
	modelToUse := model
	if modelToUse == "" {
		modelToUse = a.defaultModel
	}
	resp, err := a.prov.Chat(ctx, llm.ChatRequest{
		Model:    modelToUse,
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "ping"}},
	})
	if err != nil {
		return failover.Usage{}, err
	}
	cost := llm.EstimateCost(modelToUse, resp.Usage)
	return failover.Usage{
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		TotalTokens:  resp.Usage.TotalTokens,
		CostUSD:      cost,
	}, nil
}

// allModels is a static list mapping model IDs to providers, used to
// populate provider model lists without runtime introspection.
var allModels = []struct {
	Provider string
	llm.ModelInfo
}{
	{Provider: "openai", ModelInfo: llm.ModelInfo{ID: "gpt-4o"}},
	{Provider: "openai", ModelInfo: llm.ModelInfo{ID: "gpt-4o-mini"}},
	{Provider: "anthropic", ModelInfo: llm.ModelInfo{ID: "claude-3-5-sonnet-20241022"}},
	{Provider: "anthropic", ModelInfo: llm.ModelInfo{ID: "claude-3-5-haiku-20241022"}},
	{Provider: "google", ModelInfo: llm.ModelInfo{ID: "gemini-1.5-pro"}},
	{Provider: "google", ModelInfo: llm.ModelInfo{ID: "gemini-1.5-flash"}},
	{Provider: "xai", ModelInfo: llm.ModelInfo{ID: "grok-2"}},
	{Provider: "xai", ModelInfo: llm.ModelInfo{ID: "grok-2-mini"}},
	{Provider: "deepseek", ModelInfo: llm.ModelInfo{ID: "deepseek-chat"}},
	{Provider: "deepseek", ModelInfo: llm.ModelInfo{ID: "deepseek-reasoner"}},
	{Provider: "groq", ModelInfo: llm.ModelInfo{ID: "llama-3.3-70b-versatile"}},
	{Provider: "mistral", ModelInfo: llm.ModelInfo{ID: "mistral-large-latest"}},
	{Provider: "openrouter", ModelInfo: llm.ModelInfo{ID: "openrouter/auto"}},
	{Provider: "together", ModelInfo: llm.ModelInfo{ID: "meta-llama/Llama-3.3-70B-Instruct-Turbo"}},
	{Provider: "fireworks", ModelInfo: llm.ModelInfo{ID: "accounts/fireworks/models/llama-v3p3-70b-instruct"}},
	{Provider: "ollama", ModelInfo: llm.ModelInfo{ID: "llama3.2"}},
	{Provider: "custom", ModelInfo: llm.ModelInfo{ID: "custom-model"}},
}

func registerMethods(
	srv *ipc.Server,
	log *slog.Logger,
	cfg *config.Config,
	db *storage.DB,
	sm secrets.Manager,
	akm *api_key.Manager,
	registry *llm.Registry,
	fo *failover.Failover,
	mon *failover.SpendMonitor,
	hr *health.Register,
	ver version.Info,
) {
	_ = log
	_ = db
	_ = sm
	_ = fo
	srv.Register("ping", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{"pong": true, "ts": time.Now().Unix()}, nil
	})
	srv.Register("version", func(_ context.Context, _ json.RawMessage) (any, error) {
		return ver, nil
	})
	srv.Register("config.get", func(_ context.Context, _ json.RawMessage) (any, error) {
		return cfg, nil
	})
	srv.Register("health.snapshot", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return hr.Snapshot(ctx), nil
	})
	srv.Register("providers.list", func(_ context.Context, _ json.RawMessage) (any, error) {
		return registry.List(), nil
	})
	srv.Register("providers.models", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Provider string `json:"provider"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		prov, ok := registry.Get(p.Provider)
		if !ok {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "unknown provider: " + p.Provider}
		}
		return prov.Models(), nil
	})
	srv.Register("apikeys.list", func(ctx context.Context, _ json.RawMessage) (any, error) {
		keys, err := akm.List(ctx)
		if err != nil {
			return nil, err
		}
		// Strip secrets before returning.
		type safeKey struct {
			ID       int64  `json:"id"`
			Provider string `json:"provider"`
			Label    string `json:"label"`
			AuthKind string `json:"auth_kind"`
			HasToken bool   `json:"has_token"`
		}
		out := make([]safeKey, 0, len(keys))
		for _, k := range keys {
			out = append(out, safeKey{
				ID: k.ID, Provider: k.Provider, Label: k.Label,
				AuthKind: string(k.AuthKind), HasToken: k.Secret != "",
			})
		}
		return out, nil
	})
	srv.Register("apikeys.set", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Provider string `json:"provider"`
			Label    string `json:"label"`
			Secret   string `json:"secret"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		id, err := akm.Set(ctx, api_key.Key{
			Provider: p.Provider, Label: p.Label, AuthKind: api_key.AuthAPIKey, Secret: p.Secret,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"id": id}, nil
	})
	srv.Register("apikeys.delete", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		return nil, akm.Delete(ctx, p.ID)
	})
	srv.Register("spend.today", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{
			"spent":    mon.Spent(),
			"cap":      mon.Cap().USDPerDay,
			"remaining": mon.Remaining(),
		}, nil
	})
	srv.Register("llm.chat", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Provider        string         `json:"provider"`
			Model           string         `json:"model"`
			Request         llm.ChatRequest `json:"request"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		prov, ok := registry.Get(p.Provider)
		if !ok {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "unknown provider: " + p.Provider}
		}
		if p.Request.Model == "" {
			p.Request.Model = p.Model
		}
		if p.Request.Model == "" {
			p.Request.Model = prov.DefaultModel("chat")
		}
		resp, err := prov.Chat(ctx, p.Request)
		if err != nil {
			return nil, err
		}
		cost := llm.EstimateCost(p.Request.Model, resp.Usage)
		mon.Record(cost)
		return map[string]any{
			"response": resp,
			"cost_usd": cost,
		}, nil
	})
}

// keep net imported for builds.
var _ = net.Listen
