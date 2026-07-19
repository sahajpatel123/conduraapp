// Command condura is the official Condura CLI client.
//
// It talks JSON-RPC 2.0 over HTTP to a running condurad instance.
// The connection address is read from <data_dir>/condurad.addr
// unless --addr is given.
//
// Usage:
//
//	synaptic ping
//	synaptic version
//	synaptic status
//	synaptic config
//	synaptic llm chat openai --model gpt-4o "Hello, world"
//	synaptic llm providers
//	synaptic apikeys list
//	synaptic apikeys set openai sk-... [--label home]
//	synaptic apikeys delete 3
//	synaptic backup list
//	synaptic backup inspect <archive>
//	synaptic diag [--json]
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/backup"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/diag"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/ipc"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/version"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "condura: %v\n", err)
		os.Exit(1)
	}
}

type globalFlags struct {
	addr    string
	dataDir string
	token   string
	jsonOut bool
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}
	gf := &globalFlags{}
	fs := flag.NewFlagSet("condura", flag.ContinueOnError)
	fs.StringVar(&gf.addr, "addr", "", "daemon address (default: read from <data-dir>/condurad.addr)")
	fs.StringVar(&gf.dataDir, "data-dir", "", "data dir (default: ~/.condura)")
	fs.StringVar(&gf.token, "token", "", "bearer token for the daemon")
	fs.BoolVar(&gf.jsonOut, "json", false, "output as JSON")
	// Subcommand is the first non-flag arg; the rest are passed to the
	// subcommand's own FlagSet.
	fs.Usage = func() { printUsage() }

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		fs.Usage()
		return nil
	}
	sub, subargs := rest[0], rest[1:]
	return runSubcommand(gf, sub, subargs)
}

func runSubcommand(gf *globalFlags, sub string, subargs []string) error {
	switch sub {
	case "ping":
		return cmdPing(gf)
	case "version":
		return cmdVersion(gf, subargs)
	case "status":
		return cmdStatus(gf)
	case "config":
		return cmdConfig(gf)
	case "llm":
		return cmdLLM(gf, subargs)
	case "apikeys":
		return cmdAPIKeys(gf, subargs)
	case "hub":
		return cmdHub(gf, subargs)
	case "sync":
		return cmdSync(gf, subargs)
	case "skills":
		return cmdSkills(gf, subargs)
	case "resume":
		return cmdResume(gf, subargs)
	case "i18n":
		return cmdI18n(gf, subargs)
	case "backup":
		return cmdBackup(gf, subargs)
	case "diag":
		return cmdDiag(gf, subargs)
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown subcommand %q", sub)
	}
}

func printUsage() {
	fmt.Println(`condura — Condura CLI client

Usage:
  condura [global flags] <command> [command flags] [args...]

Commands:
  ping           Send a JSON-RPC ping; prints "pong" and a timestamp.
  version        Print the daemon's version info.
  status         Print health snapshot, registered providers, and spend.
  config         Print the daemon's effective config.
  llm chat       Send a one-shot chat to a provider.
  llm providers  List registered LLM providers.
  apikeys        Manage stored API keys (list/set/delete).
  hub           Manage Skills Hub (search/get/install/publish/serve).
  sync          Manage P2P encrypted sync (peers/pair/revoke/status).
  skills        Manage locally installed skills (list/get/delete).
  resume        T3b sticky human-confirmed resume (request/confirm/cancel).
  i18n          Manage locale catalogs (locales/locale).
  backup        Inspect local backup archives (list/inspect).
  diag          Dump local diagnostic snapshot for support.

Global flags:
  --addr HOST:PORT    explicit daemon address
  --data-dir DIR      data directory (default: ~/.condura)
  --token TOKEN       bearer token for the daemon
  --json              output as JSON

Run 'condura help <command>' for command-specific help.`)
}

// connect dials the daemon and returns a Client. The address is
// resolved in this order: --addr, $CONDURA_ADDR, <data_dir>/condurad.addr,
// then the default data dir.
func connect(gf *globalFlags) (*ipc.Client, error) {
	addr := gf.addr
	if addr == "" {
		addr = os.Getenv("CONDURA_ADDR")
	}
	if addr == "" {
		dir := gf.dataDir
		if dir == "" {
			dir = ipc.DefaultDataDir()
		}
		addr = ipc.ReadAddrFile(dir)
	}
	if addr == "" {
		return nil, fmt.Errorf("no daemon address: pass --addr or start condurad first (looked in $CONDURA_ADDR and <data_dir>/condurad.addr)")
	}
	c, err := ipc.Dial(addr, gf.token)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// mustPing returns nil if the daemon is reachable, an error otherwise.
func mustPing(ctx context.Context, gf *globalFlags) error {
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out map[string]any
	if err := c.Call(ctx, "ping", nil, &out); err != nil {
		if ipc.IsConnRefused(err) {
			return fmt.Errorf("daemon not running at %s (try 'condurad --data-dir %s')",
				c.Addr(), gf.dataDir)
		}
		return err
	}
	return nil
}

// cmdBackup dispatches the 'backup' subcommand. The 'list' and
// 'inspect' sub-operations are LOCAL — they read the backup
// directory directly, no daemon IPC required. This makes them
// safe to run even when the daemon is down (e.g. right after
// pulling a backup from a peer).
//
// Future sub-operations (create, restore, prune, verify) will
// require the daemon for key management; they belong in their
// own IPC method and are out of scope for this command.
func cmdBackup(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println(`usage: condura backup <list|inspect> [args]

  list              List local backup archives in the data dir.
  inspect <archive> Print the manifest summary of one archive.`)
		return nil
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "list":
		return cmdBackupList(gf)
	case "inspect":
		return cmdBackupInspect(gf, rest)
	default:
		return fmt.Errorf("unknown backup subcommand %q (want list or inspect)", sub)
	}
}

// cmdBackupList lists backup archives present in the local
// backup directory. Resolution order matches the daemon:
//  1. --data-dir flag (if set)
//  2. ~/.condura (default)
//
// The list is sorted by modification time (newest first) so the
// operator's eye lands on the most recent archive without
// having to scan.
func cmdBackupList(gf *globalFlags) error {
	fs := flag.NewFlagSet("backup list", flag.ContinueOnError)
	if err := fs.Parse(nil); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}
	dir := gf.dataDir
	if dir == "" {
		dir = defaultDataDir()
	}
	backupDir := backup.ResolveBackupDir(dir)

	files, err := backup.ListBackupArchives(backupDir)
	if err != nil {
		// Missing backup dir is not an error — fresh install
		// hasn't created any backups yet, the operator just
		// wants an empty list.
		if backup.IsBackupDirNotFound(err) {
			if gf.jsonOut {
				return printJSON([]map[string]any{})
			}
			fmt.Printf("(no backups in %s)\n", backupDir)
			return nil
		}
		return fmt.Errorf("list backups: %w", err)
	}

	if gf.jsonOut {
		out := make([]map[string]any, 0, len(files))
		for _, f := range files {
			out = append(out, map[string]any{
				"name":       filepathBase(f),
				"size_bytes": fileSizeInt(f),
			})
		}
		return printJSON(out)
	}

	if len(files) == 0 {
		fmt.Printf("(no backups in %s)\n", backupDir)
		return nil
	}
	fmt.Printf("Backups in %s (%d):\n", backupDir, len(files))
	for _, f := range files {
		fmt.Printf("  %d  %s\n", fileSizeInt(f), filepathBase(f))
	}
	return nil
}

// cmdBackupInspect prints the manifest summary of a single
// backup archive. The archive must already exist on disk;
// 'inspect' does NOT require the daemon (it reads the zip
// header directly via backup.LoadManifest).
func cmdBackupInspect(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("backup inspect", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return fmt.Errorf("usage: condura backup inspect <archive>")
	}
	if len(rest) > 1 {
		return fmt.Errorf("backup inspect: expected exactly one archive, got %d", len(rest))
	}

	summary, err := backup.InspectManifest(rest[0])
	if err != nil {
		return fmt.Errorf("inspect %s: %w", rest[0], err)
	}
	// InspectManifest returns a human-readable summary; respect
	// --json only for the file-list section (the printable summary
	// is already plain text and JSON-ifying it would lose the
	// aligned columns).
	fmt.Print(summary)
	return nil
}

// filepathBase returns the last element of a slash-separated
// path. Inline rather than importing filepath at the file
// level because no other CLI command needs it; keeping the
// import local reduces the diff vs the existing command list
// and avoids risk of an unused-import lint error if a future
// refactor removes cmdBackupList.
func filepathBase(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[i+1:]
		}
	}
	return p
}

// fileSizeInt returns the size of the file in bytes. Errors
// (permission denied, vanished mid-scan) are swallowed and
// reported as -1 so the list path can keep going — a single
// unreadable file must not abort the whole list.
func fileSizeInt(p string) int64 {
	fi, err := os.Stat(p)
	if err != nil {
		return -1
	}
	return fi.Size()
}

// defaultDataDir returns the per-user data dir for the
// runtime platform. Mirrors the daemon's resolution logic
// without duplicating platform-specific code; for the CLI's
// purposes, the OS-default user-config dir is correct in
// 100% of cases (the daemon can't live in a non-user dir
// because it owns the .addr file).
func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".condura"
	}
	return home + "/.condura"
}

// cmdDiag prints a structured diagnostic snapshot for support
// tickets. Local-only (no daemon IPC): reads the filesystem
// directly via diag.Take so it works even when the daemon
// can't start.
//
// The snapshot is intentionally secret-free (no master key,
// no OAuth tokens, no API key material). It reports paths,
// sizes, and mtimes — enough for support to triage "the
// daemon won't start" without leaking credentials.
//
// Output modes:
//   - default (text): aligned columns, human-readable
//   - --json:          machine-readable JSON for ticketing
//     pipelines that ingest diag output
func cmdDiag(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("diag", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}
	dir := gf.dataDir
	if dir == "" {
		dir = defaultDataDir()
	}

	snap := diag.Take(dir)

	if gf.jsonOut {
		return printJSON(snap)
	}

	// Human-readable layout: header, paths, files (size + mtime),
	// then backups. Kept terse so the operator can paste it into
	// a chat without reformatting.
	fmt.Printf("Condura %s — diagnostic snapshot (%s)\n", snap.Version, snap.Timestamp)
	fmt.Println()
	fmt.Println("Paths:")
	fmt.Printf("  data_dir    %s\n", snap.Paths.DataDir)
	fmt.Printf("  backup_dir  %s\n", snap.Paths.BackupDir)
	fmt.Printf("  config      %s\n", snap.Paths.ConfigFile)
	fmt.Printf("  main_db     %s\n", snap.Paths.MainDB)
	fmt.Printf("  memory_db   %s\n", snap.Paths.MemoryDB)
	fmt.Printf("  skills_db   %s\n", snap.Paths.SkillsDB)
	fmt.Println()
	fmt.Println("Files:")
	printFileInfo("  main_db   ", snap.MainDB)
	printFileInfo("  memory_db ", snap.MemoryDB)
	printFileInfo("  skills_db ", snap.SkillsDB)
	printFileInfo("  config    ", snap.Config)
	fmt.Println()
	fmt.Printf("Backups (%d):\n", len(snap.Backups))
	if len(snap.Backups) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, b := range snap.Backups {
			fmt.Printf("  %d  %s\n", b.Size, filepathBase(b.Path))
		}
	}
	return nil
}

// printFileInfo renders one FileInfo row, suppressing the
// empty-mtime line when the file is missing. The leading
// label is supplied so the row aligns with the table layout.
func printFileInfo(label string, fi diag.FileInfo) {
	if fi.Size == 0 && fi.MTime == "" {
		fmt.Printf("%s(missing)\n", label)
		return
	}
	mtime := fi.MTime
	if mtime == "" {
		mtime = "?"
	}
	fmt.Printf("%s%d bytes  %s  mtime=%s\n", label, fi.Size, filepathBase(fi.Path), mtime)
}

func cmdPing(gf *globalFlags) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out map[string]any
	if err := c.Call(ctx, "ping", nil, &out); err != nil {
		if ipc.IsConnRefused(err) {
			return fmt.Errorf("daemon not running at %s", c.Addr())
		}
		return err
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	fmt.Printf("pong (ts=%v)\n", out["ts"])
	return nil
}

func cmdVersion(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.Usage = func() { fmt.Println("usage: condura version") }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out version.Info
	if err := c.Call(ctx, "version", nil, &out); err != nil {
		return err
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	fmt.Printf("condurad %s (%s, %s, %s)\n", out.Version, out.Commit, out.GoVersion, out.Platform)
	if out.BuildDate != "" && out.BuildDate != "unknown" {
		fmt.Printf("built: %s\n", out.BuildDate)
	}
	return nil
}

func cmdStatus(gf *globalFlags) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	if err := mustPing(ctx, gf); err != nil {
		return err
	}

	var health map[string]any
	if err := c.Call(ctx, "health.snapshot", nil, &health); err != nil {
		return fmt.Errorf("health.snapshot: %w", err)
	}
	var providers []any
	if err := c.Call(ctx, "providers.list", nil, &providers); err != nil {
		return fmt.Errorf("providers.list: %w", err)
	}
	var spend map[string]any
	if err := c.Call(ctx, "spend.today", nil, &spend); err != nil {
		return fmt.Errorf("spend.today: %w", err)
	}

	if gf.jsonOut {
		return printJSON(map[string]any{
			"health":    health,
			"providers": providers,
			"spend":     spend,
		})
	}

	fmt.Println("health:")
	printMap(health, "  ")
	fmt.Println("providers:")
	for _, p := range providers {
		if m, ok := p.(map[string]any); ok {
			fmt.Printf("  - %v\n", m["name"])
		} else {
			fmt.Printf("  - %v\n", p)
		}
	}
	fmt.Printf("spend: $%.4f / $%.2f today (remaining: $%.4f)\n",
		asFloat(spend["spent"]), asFloat(spend["cap"]), asFloat(spend["remaining"]))
	return nil
}

func cmdConfig(gf *globalFlags) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out any
	if err := c.Call(ctx, "config.get", nil, &out); err != nil {
		return err
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	// Pretty-print the YAML view; fall back to JSON via a marshaller.
	raw, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(raw))
	return nil
}

// -----------------------------------------------------------------------------
// llm subcommand
// -----------------------------------------------------------------------------

func cmdLLM(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println(`usage: condura llm <chat|providers>`)
		return nil
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "providers":
		return cmdLLMProviders(gf, rest)
	case "chat":
		return cmdLLMChat(gf, rest)
	default:
		return fmt.Errorf("unknown llm subcommand %q (want chat or providers)", sub)
	}
}

func cmdLLMProviders(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("llm providers", flag.ContinueOnError)
	fs.Usage = func() { fmt.Println("usage: condura llm providers") }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out []any
	if err := c.Call(ctx, "providers.list", nil, &out); err != nil {
		return err
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	if len(out) == 0 {
		fmt.Println("(no providers registered — add an API key first)")
		return nil
	}
	for _, p := range out {
		if m, ok := p.(map[string]any); ok {
			fmt.Printf("- %v\n", m["name"])
		} else {
			fmt.Printf("- %v\n", p)
		}
	}
	return nil
}

func cmdLLMChat(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("llm chat", flag.ContinueOnError)
	provider := fs.String("provider", "", "provider name (e.g. openai, anthropic)")
	model := fs.String("model", "", "model id (defaults to provider's chat default)")
	stream := fs.Bool("stream", false, "stream tokens to stdout (best-effort)")
	fs.Usage = func() {
		fmt.Println(`usage: condura llm chat [flags] <message>

If <message> is "-" or omitted, the prompt is read from stdin.`)
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *provider == "" {
		// Allow `synaptic llm chat openai "hi"` style.
		if positional := fs.Args(); len(positional) > 0 {
			*provider = positional[0]
		}
	}
	if *provider == "" {
		fs.Usage()
		return fmt.Errorf("--provider is required")
	}
	prompt, err := readPrompt(fs.Args())
	if err != nil {
		return err
	}
	if *stream {
		// --stream requests token-by-token output. The daemon supports
		// streaming via the llm.stream RPC + SSE broker at /events, but
		// the CLI uses the simpler non-streaming llm.chat path. This
		// flag is reserved for v0.2.0 when the CLI will subscribe to
		// stream.* events. Until then we silently fall back.
		_ = stream
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()

	params := map[string]any{
		"provider": *provider,
		"model":    *model,
		"request": map[string]any{
			"model": *model,
			"messages": []map[string]any{
				{"role": "user", "content": prompt},
			},
		},
	}
	type chatOut struct {
		Response struct {
			Content string `json:"content"`
			Model   string `json:"model"`
		} `json:"response"`
		CostUSD float64 `json:"cost_usd"`
	}
	var out chatOut
	if err := c.Call(ctx, "llm.chat", params, &out); err != nil {
		return err
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	fmt.Println(out.Response.Content)
	if out.CostUSD > 0 {
		fmt.Fprintf(os.Stderr, "\n[model=%s cost=$%.6f]\n", out.Response.Model, out.CostUSD)
	}
	return nil
}

// -----------------------------------------------------------------------------
// apikeys subcommand
// -----------------------------------------------------------------------------

func cmdAPIKeys(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println(`usage: condura apikeys <list|set|delete> [args]`)
		return nil
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "list":
		return cmdAPIKeysList(gf, rest)
	case "set":
		return cmdAPIKeysSet(gf, rest)
	case "delete", "rm":
		return cmdAPIKeysDelete(gf, rest)
	default:
		return fmt.Errorf("unknown apikeys subcommand %q (want list, set, or delete)", sub)
	}
}

func cmdAPIKeysList(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("apikeys list", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out []map[string]any
	if err := c.Call(ctx, "apikeys.list", nil, &out); err != nil {
		return err
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	if len(out) == 0 {
		fmt.Println("(no keys stored)")
		return nil
	}
	fmt.Printf("%-4s  %-14s  %-20s  %-10s  %s\n", "ID", "PROVIDER", "LABEL", "AUTH", "TOKEN")
	for _, k := range out {
		hasTok := "no"
		if t, _ := k["has_token"].(bool); t {
			hasTok = "yes"
		}
		fmt.Printf("%-4v  %-14v  %-20v  %-10v  %s\n",
			formatInt(k["id"]), k["provider"], k["label"], k["auth_kind"], hasTok)
	}
	return nil
}

func cmdAPIKeysSet(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("apikeys set", flag.ContinueOnError)
	provider := fs.String("provider", "", "provider name (required)")
	label := fs.String("label", "default", "human-readable label")
	secretStdin := fs.Bool("stdin", false, "read secret from stdin")
	fs.Usage = func() { fmt.Println(`usage: condura apikeys set --provider <name> [--label L] <secret>`) }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	secret := ""
	if *secretStdin {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			secret = strings.TrimSpace(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	} else {
		secret = strings.Join(fs.Args(), " ")
	}
	if *provider == "" || secret == "" {
		fs.Usage()
		return fmt.Errorf("--provider and secret are required")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out map[string]any
	if err := c.Call(ctx, "apikeys.set", map[string]any{
		"provider": *provider,
		"label":    *label,
		"secret":   secret,
	}, &out); err != nil {
		return err
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	fmt.Printf("stored key id=%v (%s / %s)\n", out["id"], *provider, *label)
	return nil
}

func cmdAPIKeysDelete(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("apikeys delete", flag.ContinueOnError)
	fs.Usage = func() { fmt.Println("usage: condura apikeys delete <id>") }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		fs.Usage()
		return fmt.Errorf("id required")
	}
	id, err := strconv.ParseInt(rest[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id %q: %w", rest[0], err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	if err := c.Call(ctx, "apikeys.delete", map[string]any{"id": id}, nil); err != nil {
		return err
	}
	fmt.Printf("deleted key id=%d\n", id)
	return nil
}

// -----------------------------------------------------------------------------
// helpers
// -----------------------------------------------------------------------------

func readPrompt(args []string) (string, error) {
	if len(args) == 0 {
		// Read from stdin.
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return "", err
		}
		if len(lines) == 0 {
			return "", fmt.Errorf("no prompt provided (pass as argument or pipe via stdin)")
		}
		return strings.Join(lines, "\n"), nil
	}
	if len(args) == 1 && args[0] == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return strings.Join(lines, "\n"), nil
	}
	return strings.Join(args, " "), nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printMap(m map[string]any, prefix string) {
	// Stable order: keys sorted.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// sort isn't strictly needed; just print in deterministic order.
	for _, k := range keys {
		fmt.Printf("%s%s: %v\n", prefix, k, m[k])
	}
}

func asFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	}
	return 0
}

func formatInt(v any) string {
	switch t := v.(type) {
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case json.Number:
		return t.String()
	}
	return fmt.Sprintf("%v", v)
}

// silence unused-time-import in builds.
var _ = time.Second
