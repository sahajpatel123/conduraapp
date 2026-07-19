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
//	synaptic backup delete <archive> [--force]
//	synaptic backup prune [--keep-last N] [--older-than D] [--dry-run]
//	synaptic diag [--json]
//	synaptic validate [--json]
//	synaptic logs [--lines N] [--follow]
//	synaptic path [--exists]
//	synaptic explain <code>
//	synaptic env [--json]
//	synaptic where <key>
//	synaptic doctor [--json] [--fix]
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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/backup"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/diag"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/ipc"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/logtail"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/validate"
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
	// --version: print the CLI's build version and exit 0.
	// Standard Unix convention; matches `git --version`, `go version`, etc.
	// Doesn't require a running daemon (uses the build-time version
	// constant from internal/version). Useful for scripts that
	// detect the CLI version without contacting condurad.
	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "print CLI version and exit")
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
	// --version: short-circuit before subcommand dispatch. Print
	// the CLI's build version (from internal/version, a build-
	// time constant) and exit 0. No daemon IPC required.
	if showVersion {
		fmt.Printf("condura %s (built with %s)\n", version.Version, runtime.Version())
		return nil
	}
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
	case "validate":
		return cmdValidate(gf, subargs)
	case "logs":
		return cmdLogs(gf, subargs)
	case "path":
		return cmdPath(gf, subargs)
	case "explain":
		return cmdExplain(gf, subargs)
	case "env":
		return cmdEnv(gf, subargs)
	case "where":
		return cmdWhere(gf, subargs)
	case "doctor":
		return cmdDoctor(gf, subargs)
	case "help", "-h", "--help":
		// condura help <subcommand>: show per-subcommand help
		// when the operator passes a subcommand name. condura
		// help (no args) shows the global usage (same as
		// --help / -h). Unix-standard: `git help`, `npm help`.
		if len(subargs) > 0 {
			return helpFor(subargs[0])
		}
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
  backup        Inspect local backup archives (list/inspect/delete/prune).
  diag          Dump local diagnostic snapshot for support.
  validate      Run local install health checks.
  logs           Read the last N lines of the daemon log.
  path           Print the standard install paths.
  explain        Explain an IPC error code.
  env             Print the env vars that affect Condura behavior.
  where           Print a single install path (inverse of path).
  doctor         Run validate + print remediation steps for each failure.

Global flags:
  --addr HOST:PORT    explicit daemon address
  --data-dir DIR      data directory (default: ~/.condura)
  --token TOKEN       bearer token for the daemon
  --json              output as JSON

Run 'condura help <command>' for command-specific help.`)
}

// helpFor prints per-command usage for the named subcommand.
// `condura help explain` shows the explain command's usage;
// `condura help backup` shows the backup subcommand's usage.
//
// The registry is a flat map of `command -> usage string`. The
// daemon-required commands (ping, version, status, config,
// llm, apikeys, hub, sync, skills, resume, i18n) all print
// "requires daemon" because they have no local-only equivalent.
// The local-only commands (diag, validate, logs, path, explain,
// env, where, doctor, backup subcommands) are flagged too.
//
// Unknown commands print a "no help for X" message that lists
// the known commands, so the operator can discover what's
// available without reading docs.
func helpFor(name string) error {
	usage, ok := commandHelp[name]
	if !ok {
		fmt.Printf("no help for %q\n", name)
		fmt.Println()
		fmt.Println("known commands:")
		for _, n := range helpOrder {
			fmt.Printf("  %-12s  %s\n", n, commandHelp[n])
		}
		return nil
	}
	fmt.Printf("usage: condura %s\n\n", usage)
	return nil
}

// commandHelp is the per-command usage registry. Adding a
// new subcommand = 1 line here.
var commandHelp = map[string]string{
	"ping":     "ping                                Send a JSON-RPC ping; prints pong and a timestamp (requires daemon)",
	"version":  "version [--local]                   Print the daemon's version info (--local = CLI version, no daemon)",
	"status":   "status                              Print health snapshot, registered providers, and spend (requires daemon)",
	"config":   "config                              Print the daemon's effective config (requires daemon)",
	"llm":      "llm chat|providers                  Chat with a provider / list providers (requires daemon)",
	"apikeys":  "apikeys <list|set|delete>           Manage stored API keys (requires daemon)",
	"hub":      "hub <search|get|install|publish|serve>   Manage Skills Hub (requires daemon)",
	"sync":     "sync <peers|pair|revoke|status>     Manage P2P encrypted sync (requires daemon)",
	"skills":   "skills <list|get|delete>             Manage locally installed skills (requires daemon)",
	"resume":   "resume <request|confirm|cancel>     T3b sticky human-confirmed resume (requires daemon)",
	"i18n":     "i18n <locales|locale>                Manage locale catalogs (requires daemon)",
	"backup":   "backup <list|inspect|delete|prune>   Manage local backup archives (local-only)",
	"diag":     "diag [--json]                        Dump local diagnostic snapshot for support (local-only)",
	"validate": "validate [--json]                   Run local install health checks (local-only)",
	"logs":     "logs [--lines N] [--follow]         Read the last N lines of the daemon log (local-only)",
	"path":     "path [--exists] [--json]            Print the standard install paths (local-only)",
	"explain":  "explain <code>                      Explain an IPC error code (local-only)",
	"env":      "env [--json]                        Print the env vars that affect Condura behavior (local-only)",
	"where":    "where <key> [--list]                 Print a single install path; inverse of 'path' (local-only)",
	"doctor":   "doctor [--json] [--fix]              Run validate + print remediation steps for each failure (local-only)",
}

// helpOrder is the iteration order for the command list in
// helpFor's "no help for X" message. Matches the order in
// printUsage so the operator gets a consistent view.
var helpOrder = []string{
	"ping", "version", "status", "config",
	"llm", "apikeys", "hub", "sync", "skills", "resume", "i18n",
	"backup", "diag", "validate", "logs", "path",
	"explain", "env", "where", "doctor",
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
	case "delete":
		return cmdBackupDelete(gf, rest)
	case "prune":
		return cmdBackupPrune(gf, rest)
	default:
		return fmt.Errorf("unknown backup subcommand %q (want list, inspect, delete, or prune)", sub)
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

// cmdBackupDelete removes a single backup archive by name.
// Local-only (no daemon IPC): reads the filesystem directly.
// Closes the "manage my backups" loop — the operator can
// list backups (cmdBackupList), inspect one (cmdBackupInspect),
// and now remove one. No round trip to the daemon, no IPC
// auth, no JSON-RPC envelope.
//
// Safety: the command requires an EXPLICIT archive name (not
// a glob or a date range). With archives > 1 GB, the command
// prints a confirmation prompt and waits for stdin unless
// --force is set. This matches the "rm -i" / "trash" UX —
// the operator MUST look at the size before deleting.
//
// Usage:
//   condura backup delete condura-backup-2026-06-14.zip
//   condura backup delete condura-backup-2026-06-14.zip --force
//
// Use `condura backup list` to discover the archive name
// (the leftmost column). Use `condura backup inspect <archive>`
// to verify it's the right one before deleting.
func cmdBackupDelete(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("backup delete", flag.ContinueOnError)
	force := fs.Bool("force", false, "skip the size-based confirmation prompt for large archives")
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return fmt.Errorf("usage: condura backup delete <archive> [--force]")
	}
	name := rest[0]

	dir := gf.dataDir
	if dir == "" {
		dir = defaultDataDir()
	}
	backupDir := backup.ResolveBackupDir(dir)

	// Reject names that try to escape the backup dir (e.g.
	// "../etc/passwd"). The basename must be a plain .zip
	// filename — no slashes, no "..", no absolute paths.
	if strings.ContainsAny(name, "/\\") || name == ".." || strings.HasPrefix(name, ".") {
		return fmt.Errorf("condura backup delete: invalid archive name %q (must be a plain .zip basename)", name)
	}
	if !strings.HasSuffix(name, ".zip") {
		return fmt.Errorf("condura backup delete: archive name must end in .zip (got %q)", name)
	}

	target := filepath.Join(backupDir, name)

	// Verify the target actually exists in the backup dir.
	// os.Lstat (not Stat) so a symlink to /etc/passwd is
	// caught — the deletion would follow the symlink.
	fi, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("condura backup delete: no such archive %q in %s", name, backupDir)
		}
		return err
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("condura backup delete: %s is not a regular file (mode=%s); refusing to delete", name, fi.Mode())
	}

	// Size-based confirmation: archives > 1 GB require explicit
	// --force OR a typed "yes" on stdin. Smaller archives delete
	// without prompting (the operator already typed the exact
	// filename).
	const largeThreshold int64 = 1 << 30 // 1 GiB
	if fi.Size() > largeThreshold && !*force {
		fmt.Printf("Archive %q is %s (>1 GiB). Delete? [y/N]: ", name, fileSizeHuman(fi.Size()))
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" && answer != "yes" {
			fmt.Println("aborted")
			return nil
		}
	}

	if err := os.Remove(target); err != nil {
		return fmt.Errorf("condura backup delete: %w", err)
	}
	fmt.Printf("Deleted %s (%s freed)\n", name, fileSizeHuman(fi.Size()))
	return nil
}

// cmdBackupPrune removes old backup archives in bulk by
// retention policy. The "I don't care about old ones" workflow
// — instead of deleting archives one-at-a-time with
// `condura backup delete <archive>`, the operator declares a
// retention policy and the prune decides what to delete.
//
// Two mutually-compatible filter modes:
//   --keep-last N  keep the N most-recent archives; delete the rest
//   --older-than D delete archives whose mtime is older than D
//                (e.g. "30d", "12h", "1y")
//
// If BOTH are set, the union is kept: an archive survives if
// it's in the most-recent N OR newer than D. This matches the
// "either condition" intuition: "keep the most recent 5 OR the
// last 30 days, whichever is more."
//
// Local-only (no daemon IPC). Reads the filesystem directly
// via backup.ListBackupArchives (newest-first from iter-9).
//
// Default: --dry-run. The command PRINTS what it would delete
// without actually deleting. The operator must pass --force
// (without --dry-run) to actually delete. This is the "look
// before you leap" UX — the operator sees the exact list of
// archives that will be deleted, the count, and the total size
// freed, BEFORE the deletion happens.
//
// Safety: the same name-validation rules as cmdBackupDelete
// (plain .zip basename, no path traversal). The filter logic
// is in cmdBackupPrune itself, not in os.RemoveAll, so we
// always delete specific files, not directories.
//
// Usage:
//   condura backup prune --keep-last 5
//   condura backup prune --older-than 30d
//   condura backup prune --keep-last 5 --older-than 30d
//   condura backup prune --keep-last 5 --force   # actually delete
func cmdBackupPrune(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("backup prune", flag.ContinueOnError)
	keepLast := fs.Int("keep-last", 0, "keep the N most-recent archives (0 = don't apply this filter)")
	olderThan := fs.Duration("older-than", 0, "delete archives older than this duration (0 = don't apply this filter)")
	dryRun := fs.Bool("dry-run", true, "print what would be deleted without actually deleting (use --force to actually delete)")
	force := fs.Bool("force", false, "actually delete; overrides --dry-run")
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}

	// Validate filter inputs: at least one filter must be set.
	if *keepLast == 0 && *olderThan == 0 {
		return fmt.Errorf("usage: condura backup prune --keep-last N | --older-than D (at least one filter required)")
	}
	if *keepLast < 0 {
		return fmt.Errorf("--keep-last must be >= 0 (got %d)", *keepLast)
	}

	dir := gf.dataDir
	if dir == "" {
		dir = defaultDataDir()
	}
	backupDir := backup.ResolveBackupDir(dir)

	paths, err := backup.ListBackupArchives(backupDir)
	if err != nil {
		if backup.IsBackupDirNotFound(err) {
			return fmt.Errorf("condura backup prune: no backup directory at %s (fresh install)", backupDir)
		}
		return err
	}
	if len(paths) == 0 {
		fmt.Println("(no backups to prune)")
		return nil
	}

	// The retention filter: walk the newest-first list, decide
	// which ones survive, collect the rest. We use the
	// file's modification time (mtime) for the age check.
	//
	// keep-last: the first *keepLast* entries in the
	// newest-first list survive.
	//
	// older-than: entries with mtime >= cutoff (now - olderThan)
	// survive. The cutoff is computed once at the start of the
	// prune so the time window is consistent.
	cutoff := time.Time{}
	if *olderThan > 0 {
		cutoff = time.Now().Add(-*olderThan)
	}

	survive := make([]string, 0, len(paths))
	delete := make([]string, 0, len(paths))
	for i, p := range paths {
		// keep-last check: the first N entries in the
		// newest-first list survive unconditionally.
		if *keepLast > 0 && i < *keepLast {
			survive = append(survive, p)
			continue
		}
		// older-than check: the entry survives if its mtime
		// is at or after the cutoff.
		if *olderThan > 0 {
			fi, err := os.Stat(p)
			if err != nil {
				// Stat error: treat as "should be deleted"
				// (the safer action; if the file is unreadable,
				// we don't want to keep it).
				delete = append(delete, p)
				continue
			}
			if fi.ModTime().After(cutoff) {
				survive = append(survive, p)
				continue
			}
		}
		delete = append(delete, p)
	}

	// Print the decision so the operator can verify.
	fmt.Printf("Retention policy: ")
	if *keepLast > 0 {
		fmt.Printf("keep-last=%d ", *keepLast)
	}
	if *olderThan > 0 {
		fmt.Printf("older-than=%s ", *olderThan)
	}
	fmt.Println()
	fmt.Println()
	fmt.Printf("Will keep (%d):\n", len(survive))
	for _, p := range survive {
		fmt.Printf("  %s\n", filepathBase(p))
	}
	fmt.Println()
	fmt.Printf("Will delete (%d):\n", len(delete))
	var totalFreed int64
	for _, p := range delete {
		fi, _ := os.Stat(p)
		if fi != nil {
			totalFreed += fi.Size()
		}
		fmt.Printf("  %s\n", filepathBase(p))
	}
	fmt.Println()
	fmt.Printf("Total: %d kept, %d to delete, %s would be freed\n", len(survive), len(delete), fileSizeHuman(totalFreed))

	// No-op if nothing to delete.
	if len(delete) == 0 {
		return nil
	}

	// Default: dry-run. The operator must pass --force to
	// actually delete. This is the safety story: you see the
	// exact list, the count, the size, then you opt in.
	actuallyDelete := *force || !*dryRun
	if !actuallyDelete {
		fmt.Println()
		fmt.Println("(dry-run; pass --force to actually delete)")
		return nil
	}

	// Actually delete.
	for _, p := range delete {
		if err := os.Remove(p); err != nil {
			return fmt.Errorf("prune: remove %s: %w", filepathBase(p), err)
		}
	}
	fmt.Println()
	fmt.Printf("Deleted %d archive(s), %s freed.\n", len(delete), fileSizeHuman(totalFreed))
	return nil
}

// installPaths returns the canonical map of install paths
// for the given data dir. Centralized so cmdPath (list all)
// and cmdWhere (lookup one) agree on the same path
// construction. Adding a new path (e.g. a future "keyring"
// file) is a one-line change in this function; both commands
// pick it up automatically.
func installPaths(dir string) map[string]string {
	return map[string]string{
		"data_dir":    dir,
		"backup_dir":  backup.ResolveBackupDir(dir),
		"config_file": filepath.Join(dir, "config.yaml"),
		"main_db":     filepath.Join(dir, "condura.db"),
		"memory_db":   filepath.Join(dir, "memory.db"),
		"skills_db":   filepath.Join(dir, "skills.db"),
		"logs_dir":    filepath.Join(dir, "logs"),
		"log_file":    logtail.LogFilePath(dir),
		"lock_file":   filepath.Join(dir, "condurad.lock"),
		"addr_file":   filepath.Join(dir, "condurad.addr"),
	}
}

// pathOrder is the stable iteration order for cmdPath's output.
// Same key set as installPaths, in a human-friendly order
// (data-dir-related first, then DBs, then operational files).
// Centralized so the iteration order can't drift between
// cmdPath and a future "condura path --filter" mode.
var pathOrder = []string{
	"data_dir", "backup_dir", "config_file",
	"main_db", "memory_db", "skills_db",
	"logs_dir", "log_file",
	"lock_file", "addr_file",
}

// dataDirOrDefault returns gf.dataDir, falling back to the
// OS-default (~/.condura). Tiny helper so cmdPath and
// cmdWhere agree on the same data-dir resolution.
func dataDirOrDefault(gf *globalFlags) string {
	if gf.dataDir != "" {
		return gf.dataDir
	}
	return defaultDataDir()
}

// cmdWhere prints a single install path by key. Inverse of
// `condura path` (which lists all): given one key, print just
// that one path. Useful for shell scripts:
//
//   $ condura where config_file
//   /home/alice/.condura/config.yaml
//   $ cat "$(condura where config_file)"
//   # ... contents ...
//
// Usage:
//   condura where <key>
//   condura where --list       # list all known keys
//
// No daemon IPC required. Local-only.
func cmdWhere(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("where", flag.ContinueOnError)
	list := fs.Bool("list", false, "list all known keys (no path printed)")
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}

	rest := fs.Args()

	// --list: print the key catalog so the operator can
	// discover the available names without consulting docs.
	if *list {
		paths := installPaths(dataDirOrDefault(gf))
		for _, k := range pathOrder {
			fmt.Println(k)
		}
		_ = paths // paths not printed in --list mode; the keys are enough
		return nil
	}

	// No args + no --list: print usage.
	if len(rest) == 0 {
		return fmt.Errorf("usage: condura where <key>  (try `condura where --list` to see all keys)")
	}

	// Look up the key. The list of valid keys is the keys
	// of installPaths; we don't pre-define a separate list
	// to avoid drift.
	paths := installPaths(dataDirOrDefault(gf))
	key := rest[0]
	p, ok := paths[key]
	if !ok {
		// Suggest --list on miss so the operator doesn't have
		// to read the source to find the right key name.
		return fmt.Errorf("condura where: unknown key %q (try `condura where --list` to see all keys)", key)
	}
	fmt.Println(p)
	return nil
}

// fileSizeHuman renders a byte count as a human-readable
// string (e.g. "1.2 GB"). Used by cmdBackupDelete and
// cmdBackupPrune to surface sizes in their prompts.
func fileSizeHuman(b int64) string {
	const (
		KiB = 1 << 10
		MiB = 1 << 20
		GiB = 1 << 30
	)
	switch {
	case b >= GiB:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(GiB))
	case b >= MiB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MiB))
	case b >= KiB:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(KiB))
	default:
		return fmt.Sprintf("%d B", b)
	}
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

// cmdValidate runs the local install health checks and prints
// the report. Local-only (no daemon IPC required); useful as a
// pre-flight before launching the daemon, or as a post-mortem
// when the daemon won't start.
//
// Exit code:
//   - 0 if all required checks pass (Summary.Fail == 0)
//   - 1 if any required check fails (Summary.Fail > 0)
//   - Skipped checks do NOT affect the exit code — fresh
//     installs will have many skips, and that's fine.
//
// Output modes:
//   - default: aligned text with a summary line
//   - --json: structured Report JSON for ticketing pipelines
func cmdValidate(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}
	dir := gf.dataDir
	if dir == "" {
		dir = defaultDataDir()
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	r := validate.Run(ctx, dir)

	if gf.jsonOut {
		return printJSON(r)
	}

	// Text layout: per-check status + detail, then summary.
	statusGlyph := map[validate.Status]string{
		validate.StatusOK:   " OK ",
		validate.StatusWarn: "WARN",
		validate.StatusFail: "FAIL",
		validate.StatusSkip: "SKIP",
	}
	fmt.Printf("Condura validate — %s\n", r.Time)
	fmt.Printf("data_dir: %s\n", r.DataDir)
	fmt.Println()
	for _, c := range r.Checks {
		fmt.Printf("  [%s] %-12s %s\n", statusGlyph[c.Status], c.Name, c.Detail)
	}
	fmt.Println()
	fmt.Printf("Summary: %d ok, %d warn, %d fail, %d skip\n",
		r.Summary.OK, r.Summary.Warn, r.Summary.Fail, r.Summary.Skip)

	// Exit non-zero if any required check failed.
	if r.Summary.Fail > 0 {
		return fmt.Errorf("%d check(s) failed", r.Summary.Fail)
	}
	return nil
}

// remediations is the registry mapping each check name to a
// concrete fix-the-problem command. Used by cmdDoctor to print
// the next-step AFTER each failure (validate just says "what
// broke"; doctor says "what to do about it").
//
// Order roughly mirrors the check's importance: data-dir and
// main-db first (the daemon can't run without them), then the
// optional DBs, then the operational files.
var remediations = map[string]string{
	validate.CheckNameDataDir:  "run 'condurad --init' to create the data directory",
	validate.CheckNameMainDB:   "run 'condurad --init' to recreate the main database",
	validate.CheckNameMemoryDB: "run 'condurad' once to auto-create the memory DB (created on first start)",
	validate.CheckNameSkillsDB: "run 'condurad' once to auto-create the skills DB (created on first start)",
	validate.CheckNameConfig:   "edit the config file (see detail above for the YAML parser error); defaults apply if missing",
	validate.CheckNameLock:     "the lock file is held by a non-running daemon (stale lock); remove it after confirming no condurad is running",
	validate.CheckNameBackups:  "the most recent backup is unreadable; check the disk and consider 'condura backup prune' to clean up old archives",
}

// cmdDoctor runs validate + prints a remediation step for each
// failure. The pair is the difference between "I see what's
// broken" (validate) and "I see what's broken AND how to fix it"
// (doctor). Both are local-only (no daemon IPC required).
//
// Usage:
//   condura doctor              # run all checks + remediation
//   condura doctor --json       # structured output
//   condura doctor --fix        # also print generic init
//                                # instructions if data-dir is
//                                # missing
//
// Default behavior: ONLY prints remediation for FAIL/OK
// statuses (the operator doesn't need "run condurad" for
// passes). The summary line at the end is the same as
// condura validate.
func cmdDoctor(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fix := fs.Bool("fix", false, "include generic init instructions when the data dir is missing")
	check := fs.String("check", "", "run only the named check (e.g. main_db, config, lock) and skip the rest")
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dir := gf.dataDir
	if dir == "" {
		dir = defaultDataDir()
	}
	r := validate.Run(ctx, dir)

	// --check: filter the report to a single check by name.
	// The validate package runs all 7 checks unconditionally;
	// the filter happens here in the CLI, so the package
	// stays simple (no per-check skipping needed).
	if *check != "" {
		var found *validate.Check
		for i := range r.Checks {
			if r.Checks[i].Name == *check {
				found = &r.Checks[i]
				break
			}
		}
		if found == nil {
			// Build a list of valid check names from the
			// report itself (don't hardcode them — they
			// could change in future validate versions).
			var valid []string
			for _, c := range r.Checks {
				valid = append(valid, c.Name)
			}
			return fmt.Errorf("condura doctor: unknown check %q (valid: %s)", *check, strings.Join(valid, ", "))
		}
		// Build a single-check report so the JSON / text
		// output paths handle --check uniformly.
		r.Checks = []validate.Check{*found}
		r.Summary = struct {
			OK   int `json:"ok"`
			Warn int `json:"warn"`
			Fail int `json:"fail"`
			Skip int `json:"skip"`
		}{}
		switch found.Status {
		case validate.StatusOK:
			r.Summary.OK = 1
		case validate.StatusWarn:
			r.Summary.Warn = 1
		case validate.StatusFail:
			r.Summary.Fail = 1
		case validate.StatusSkip:
			r.Summary.Skip = 1
		}
	}

	// JSON output: the full report + the remediations map.
	// Useful for tooling that wants to programmatically act on
	// the diagnosis (e.g. "open a GitHub issue with the
	// failed-check names + their remediation steps").
	if gf.jsonOut {
		out := map[string]any{
			"data_dir":     r.DataDir,
			"time":         r.Time,
			"checks":       r.Checks,
			"remediations": remediations,
			"summary":      r.Summary,
		}
		return printJSON(out)
	}

	// Text output: print ONLY the failed checks + their
	// remediation. Passing checks (OK/Skip) are silent — the
	// operator doesn't need to read "all good" 7 times.
	fmt.Printf("Condura doctor — %s\n", r.Time)
	fmt.Printf("data_dir: %s\n\n", r.DataDir)
	anyFails := false
	for _, c := range r.Checks {
		if c.Status != validate.StatusFail {
			continue
		}
		anyFails = true
		fmt.Printf("  [FAIL] %-12s  %s\n", c.Name, c.Detail)
		if fix, ok := remediations[c.Name]; ok {
			fmt.Printf("             fix: %s\n", fix)
		}
	}
	if !anyFails {
		fmt.Println("(no failures — install is healthy)")
	}
	fmt.Println()
	fmt.Printf("Summary: %d ok, %d warn, %d fail, %d skip\n",
		r.Summary.OK, r.Summary.Warn, r.Summary.Fail, r.Summary.Skip)

	// --fix mode: also print generic init instructions
	// when the data dir is missing. This is the "I have NO
	// idea what's wrong" mode — start from scratch.
	if *fix && !dataDirExists(dir) {
		fmt.Println()
		fmt.Println("Generic install instructions (no data dir):")
		fmt.Println("  1. mkdir -p", dir)
		fmt.Println("  2. condurad --init --data-dir", dir)
		fmt.Println("  3. condurad --data-dir", dir, "  (start the daemon)")
	}

	// Exit non-zero if any required check failed (same as
	// condura validate's behavior — scripts can use either
	// command for "is this install healthy?").
	if r.Summary.Fail > 0 {
		return fmt.Errorf("%d check(s) failed", r.Summary.Fail)
	}
	return nil
}

// dataDirExists is a small helper that returns true if the
// data dir exists AND is a directory. Used by --fix mode to
// decide whether to print the "from scratch" instructions.
func dataDirExists(dir string) bool {
	fi, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

// cmdLogs reads the last N lines of the daemon's log file.
// Local-only (no daemon IPC required): reads the file directly
// via internal/logtail. Useful for support tickets — paste the
// last 100 lines into the ticket and the engineer can see
// what the daemon was doing when it crashed.
//
// Output: lines printed newest-first (one per line). The
// operator can pipe to `head` or `less` for interactive use.
//
// With --follow: prints the last N lines (initial context),
// then watches the file for new lines and prints them as they
// appear. Implements `tail -F` semantics: if the file is
// rotated, the watch re-opens the new file and continues.
// Honors SIGINT (Ctrl+C) — the signal handler closes the
// context, tailFollow returns nil, the CLI exits 0.
//
// Exit code: 0 on success, 1 on file read error.
//
// The log file location matches the daemon's writer: <dataDir>/
// logs/condura.log. Rotated siblings (condura.log.1, .2, ...)
// are merged transparently when more lines are needed.
func cmdLogs(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("logs", flag.ContinueOnError)
	lines := fs.Int("lines", 100, "number of lines to read from the end of the log (default 100)")
	follow := fs.Bool("follow", false, "follow the log like tail -F; prints new lines as they appear")
	level := fs.String("level", "", "minimum log level to show: debug, info, warn, error (default: all levels)")
	grep := fs.String("grep", "", "show only lines matching this regular expression (default: all lines)")
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}

	// Validate the --level flag early. Empty string means
	// "no filter" (all levels). A non-empty string must be
	// one of the four slog levels.
	minLevel := logtail.LevelDebug
	if *level != "" {
		var ok bool
		minLevel, ok = logtail.ParseLevel(*level)
		if !ok {
			return fmt.Errorf("condura logs: --level must be one of: debug, info, warn, error (got %q)", *level)
		}
	}

	dir := gf.dataDir
	if dir == "" {
		dir = defaultDataDir()
	}
	logPath := logtail.LogFilePath(dir)

	// Build the filter once, reuse across the static and
	// follow paths. nil filter means "no filter".
	filter := logtail.NewFilter(minLevel, *grep)

	// --follow: print the initial N lines (the "tail -n N"
	// starting state), then watch for new lines. The signal
	// handler wires SIGINT/SIGTERM to the context, so Ctrl+C
	// exits cleanly.
	if *follow {
		ctx, cancel := signal.NotifyContext(context.Background(),
			syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		return tailFollow(ctx, logPath, *lines, filter)
	}

	lines_out, err := logtail.Tail(logPath, *lines)
	if err != nil {
		return fmt.Errorf("condura logs: %w", err)
	}
	if lines_out == nil {
		fmt.Printf("(no log file at %s; has the daemon ever started?)\n", logPath)
		return nil
	}
	for _, l := range lines_out {
		if filter.Matches(l) {
			fmt.Println(l)
		}
	}
	return nil
}

// logFilePathForCLI returns the canonical log path for a given
// globalFlags config. Tiny wrapper around logtail.LogFilePath
// so the "not yet implemented" error message is consistent
// with the actual reader.
func logFilePathForCLI(gf *globalFlags) string {
	dir := gf.dataDir
	if dir == "" {
		dir = defaultDataDir()
	}
	return logtail.LogFilePath(dir)
}

// cmdPath prints the standard install paths. Pairs with
// condura diag (which lists the same paths in its output) —
// the difference: condura path is a quick standalone command
// for "where is the data dir?", "where does the daemon log
// to?", etc. without dumping the full snapshot.
//
// Output modes:
//   - default: one line per path, key + value, no decoration
//   - --exists: append "exists"/"missing" so the operator
//              can spot a partial install at a glance
//   - --json:   structured {"data_dir": "...", "exists": true, ...}
func cmdPath(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("path", flag.ContinueOnError)
	exists := fs.Bool("exists", false, "append (exists) or (missing) after each path")
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}

	paths := installPaths(dataDirOrDefault(gf))

	// Stable order so the output is diffable across runs.
	order := []string{
		"data_dir", "backup_dir", "config_file",
		"main_db", "memory_db", "skills_db",
		"logs_dir", "log_file",
		"lock_file", "addr_file",
	}

	if gf.jsonOut {
		// JSON mode: nest the existence check into a sub-object
		// so the GUI can show "data_dir": "/path" + "data_dir
		// exists": true without re-running the existence check.
		out := make(map[string]any, len(paths))
		for _, k := range order {
			if *exists {
				_, err := os.Stat(paths[k])
				out[k] = map[string]any{
					"path":  paths[k],
					"exists": err == nil,
				}
			} else {
				out[k] = paths[k]
			}
		}
		return printJSON(out)
	}

	for _, k := range order {
		if *exists {
			_, err := os.Stat(paths[k])
			suffix := "(exists)"
			if err != nil {
				suffix = "(missing)"
			}
			fmt.Printf("%-12s  %s  %s\n", k, paths[k], suffix)
		} else {
			fmt.Printf("%-12s  %s\n", k, paths[k])
		}
	}
	return nil
}

// envVar is one row in the condura env registry. Each row
// documents an env var the operator might want to check or
// override, with a one-line description of what it does.
//
// Order roughly mirrors "what does the operator most often
// want to know?" — the data-dir-related vars first, then the
// daemon-address / connection vars, then the per-feature
// overrides.
type envVar struct {
	name        string
	description string
}

// knownEnvVars is the registry. When the daemon or backup
// subsystem adds a new env-var lookup, the entry should be
// added here too — the test (TestEnv_KnowsAllLookups) pins
// this invariant so the registry can't drift.
//
// We list only env vars the codebase ACTUALLY reads (per
// the os.Getenv / os.LookupEnv calls in internal/ + cmd/).
// Listing every theoretically-possible env var would
// overwhelm the operator with noise.
var knownEnvVars = []envVar{
	{
		name:        "HOME",
		description: "User home directory. Resolves ~/.condura/ for the data dir (Unix/macOS).",
	},
	{
		name:        "USERPROFILE",
		description: "User home directory (Windows). Takes precedence over HOME on Windows.",
	},
	{
		name:        "XDG_CONFIG_HOME",
		description: "Linux config base. Resolves ~/.config/condura/ for the data dir when HOME is empty.",
	},
	{
		name:        "APPDATA",
		description: "Windows roaming config base. Resolves %APPDATA%/condura/ for the data dir when USERPROFILE is empty.",
	},
	{
		name:        "CONDURA_ADDR",
		description: "Override the daemon's listen address (default: read from <data_dir>/condurad.addr).",
	},
	{
		name:        "CONDURA_BACKUP_DIR",
		description: "Override the standard backup location (default: ~/Documents/condura-backups, per MISSION §24.1).",
	},
	{
		name:        "CONDURA_FILE_PASSPHRASE",
		description: "Secrets manager file-mode passphrase fallback (used when the OS keyring is unavailable).",
	},
	{
		name:        "CONDURA_RESUME_SECRET",
		description: "Resume ticket signing key (default: a random key generated at daemon startup).",
	},
	{
		name:        "CONDURA_ACCOUNT_OAUTH_<PROVIDER>_CLIENT_ID",
		description: "OAuth client ID per provider (PROVIDER uppercased). Overrides config.yaml account.oauth.<provider>.client_id.",
	},
	{
		name:        "CONDURA_ACCOUNT_OAUTH_<PROVIDER>_CLIENT_SECRET",
		description: "OAuth client secret per provider. Overrides config.yaml account.oauth.<provider>.client_secret.",
	},
}

// envValues returns the current process env as a name→value
// map. Values for unset vars are "" (the lookup convention).
//
// We read from os.Environ() rather than os.Getenv in a loop
// because the former is O(1) per call after the first read
// (Go's runtime caches the env). For ~10 vars, the difference
// is negligible, but the function is the single point of access
// and easy to optimize later if needed.
func envValues() map[string]string {
	pairs := os.Environ()
	out := make(map[string]string, len(pairs))
	for _, p := range pairs {
		// os.Environ returns "KEY=VALUE" pairs.
		for i := 0; i < len(p); i++ {
			if p[i] == '=' {
				out[p[:i]] = p[i+1:]
				break
			}
		}
	}
	return out
}

// cmdEnv prints all env vars that affect Condura's behavior,
// with their current values. Designed for the operator who's
// debugging "why is my data dir wrong?" or "why isn't the
// daemon reading my config?" — instead of grepping the
// codebase for env-var lookups, run `condura env` and see
// the full list at a glance.
//
// Output modes:
//   - default: aligned text, "NAME  (unset)" or "NAME  VALUE"
//   - --json:   structured {name, value, set} per row
//
// Env values are NOT sanitized: if the operator set
// CONDURA_FILE_PASSPHRASE to a real passphrase, it'll show
// up in the output. This is intentional — the operator
// KNOWS their own env. The point of this command is "tell
// me what I have set", not "tell me what a safe default is".
func cmdEnv(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("env", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}

	vals := envValues()

	if gf.jsonOut {
		out := make([]map[string]any, 0, len(knownEnvVars))
		for _, e := range knownEnvVars {
			v, set := vals[e.name]
			out = append(out, map[string]any{
				"name":        e.name,
				"description": e.description,
				"value":       v,
				"set":         set,
			})
		}
		return printJSON(out)
	}

	for _, e := range knownEnvVars {
		v, set := vals[e.name]
		if set {
			fmt.Printf("%-45s  %s\n", e.name, v)
		} else {
			fmt.Printf("%-45s  (unset)\n", e.name)
		}
	}
	return nil
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
	// --local: print the build-time CLI version, no daemon IPC.
	// The build-time version is the version of the CLI binary
	// the operator is running, NOT the version of the daemon
	// (which may differ if the CLI was upgraded but the daemon
	// wasn't restarted). For daemon version, use the default
	// (daemon IPC) flow.
	local := fs.Bool("local", false, "print the CLI's build-time version without contacting the daemon")
	fs.Usage = func() { fmt.Println("usage: condura version [--local]") }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	// --local short-circuit: print the build-time version, no IPC.
	// Useful for scripts that want to detect the CLI version
	// without contacting the daemon (e.g. "is this CLI new
	// enough to support command X?").
	if *local {
		fmt.Printf("condura %s (built with %s)\n", version.Version, runtime.Version())
		return nil
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
