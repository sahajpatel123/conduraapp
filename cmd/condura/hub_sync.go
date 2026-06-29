package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/sahajpatel123/conduraapp/internal/hub"
)

// cmdHub dispatches the `hub` subcommand.
func cmdHub(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println(`usage: synaptic hub <search|get|install|publish|serve>

  search QUERY          search the Skills Hub
  get ID                fetch metadata for a skill
  install ID            download + safety-scan + install
  publish ID PATH       upload a local skill to the Hub
  serve                 run a local Hub server (offline mode)`)
		return nil
	}
	switch args[0] {
	case "search":
		return cmdHubSearch(gf, args[1:])
	case "get":
		return cmdHubGet(gf, args[1:])
	case "install":
		return cmdHubInstall(gf, args[1:])
	case "publish":
		return cmdHubPublish(gf, args[1:])
	case "serve":
		return cmdHubServe(gf, args[1:])
	default:
		return fmt.Errorf("unknown hub subcommand %q", args[0])
	}
}

func cmdHubSearch(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("hub search", flag.ContinueOnError)
	limit := fs.Int("limit", 20, "max results")
	fs.Usage = func() { fmt.Println("usage: synaptic hub search [--limit N] QUERY") }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		fs.Usage()
		return fmt.Errorf("query required")
	}
	query := rest[0]
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out struct {
		Skills []map[string]any `json:"skills"`
		Total  int              `json:"total"`
		Query  string           `json:"query"`
	}
	if err := c.Call(ctx, "hub.search", map[string]any{"query": query, "limit": *limit}, &out); err != nil {
		return err
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	if out.Total == 0 {
		fmt.Printf("no skills matched %q\n", query)
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "ID\tNAME\tVERSION\tAUTHOR\tTRUST\tDOWNLOADS\n")
	for _, s := range out.Skills {
		fmt.Fprintf(tw, "%v\t%v\t%v\t%v\t%v\t%v\n",
			s["id"], s["name"], s["version"], s["author"], s["trust"], s["downloads"])
	}
	_ = tw.Flush()
	return nil
}

func cmdHubGet(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: synaptic hub get ID")
		return nil
	}
	id := args[0]
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var meta map[string]any
	if err := c.Call(ctx, "hub.get", map[string]any{"id": id}, &meta); err != nil {
		return err
	}
	if gf.jsonOut {
		return printJSON(meta)
	}
	for k, v := range meta {
		fmt.Printf("%-14s %v\n", k+":", v)
	}
	return nil
}

func cmdHubInstall(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: synaptic hub install ID")
		return nil
	}
	id := args[0]
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out map[string]any
	if err := c.Call(ctx, "hub.install", map[string]any{"id": id}, &out); err != nil {
		return err
	}
	fmt.Printf("installed skill %q\n", id)
	return nil
}

// cmdHubPublish uploads a local skill archive to the hub. The
// archive is read from PATH and its contents are sent (the daemon
// does NOT read from a local store by ID — the file is the source
// of truth, so a user can re-publish after editing the archive).
//
// Usage: synaptic hub publish PATH
//
//	PATH is a skill archive file (.zip). The skill's metadata is
//	read from the archive itself; the file is uploaded verbatim.
func cmdHubPublish(gf *globalFlags, args []string) error {
	if len(args) < 1 {
		fmt.Println("usage: synaptic hub publish PATH")
		return nil
	}
	path := args[0]
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	archive, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	// Extract the skill ID from the filename (strip .zip).
	id := strings.TrimSuffix(filepath.Base(path), ".zip")
	var out map[string]any
	if err := c.Call(ctx, "hub.publish", map[string]any{
		"id":      id,
		"archive": archive,
	}, &out); err != nil {
		return err
	}
	fmt.Printf("published skill %q (%d bytes)\n", id, len(archive))
	return nil
}

// cmdHubServe starts a local Skills Hub server rooted at --root.
// Useful for offline use, internal company hubs, and CI testing.
// The server speaks the same /api/v1/* surface as the network
// hub, so the same client code can target it by setting
// hub.base_url to http://localhost:7777.
func cmdHubServe(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("hub serve", flag.ContinueOnError)
	addr := fs.String("addr", "127.0.0.1:7777", "address to listen on")
	root := fs.String("root", "./synaptic-hub", "directory to store skills")
	token := fs.String("token", "", "bearer token (default: open)")
	fs.Usage = func() {
		fmt.Println("usage: synaptic hub serve [--addr HOST:PORT] [--root DIR] [--token TOKEN]")
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	srv, err := hub.NewServer(*root, *token)
	if err != nil {
		return fmt.Errorf("hub server init: %w", err)
	}
	fmt.Printf("local Skills Hub listening on %s (root=%s, auth=%s, count=%d)\n",
		*addr, *root, ternary(*token != "", "token-protected", "open"), srv.Count())
	fmt.Println("press Ctrl+C to stop")
	return srv.ListenAndServe(*addr)
}

// cmdSkills dispatches the `skills` subcommand (Phase 12 local
// skills). Each subcommand is a thin wrapper around an RPC.
func cmdSkills(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println(`usage: synaptic skills <list|get|delete>

  list               list installed skills
  get ID             fetch one skill by ID
  delete ID          remove a skill by ID`)
		return nil
	}
	switch args[0] {
	case "list":
		return cmdSyncCall(gf, "skills.list", map[string]any{"limit": 100})
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("skills get: ID is required")
		}
		return cmdSyncCall(gf, "skills.get", map[string]any{"id": args[1]})
	case "delete":
		if len(args) < 2 {
			return fmt.Errorf("skills delete: ID is required")
		}
		return cmdSyncCall(gf, "skills.delete", map[string]any{"id": args[1]})
	default:
		return fmt.Errorf("unknown skills subcommand %q", args[0])
	}
}

// cmdI18n dispatches the `i18n` subcommand (Phase 12 locales).
func cmdI18n(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println(`usage: synaptic i18n <locales|locale>

  locales            list available locales
  locale CODE        print all translations for a locale`)
		return nil
	}
	switch args[0] {
	case "locales":
		return cmdSyncCall(gf, "i18n.locales", nil)
	case "locale":
		if len(args) < 2 {
			return fmt.Errorf("i18n locale: locale code is required (e.g. 'en', 'fr')")
		}
		return cmdSyncCall(gf, "i18n.locale", map[string]any{"locale": args[1]})
	default:
		return fmt.Errorf("unknown i18n subcommand %q", args[0])
	}
}

// cmdSync dispatches the `sync` subcommand.
func cmdSync(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println(`usage: synaptic sync <status|peers|put|get|start|stop|list-pairs|pair|revoke>

  status            show sync engine status
  peers             list discovered peers on the LAN
  put KEY VALUE     store a value in the local CRDT
  get KEY           retrieve a value from the local CRDT
  start             start the sync engine
  stop              stop the sync engine
  list-pairs        list paired (trusted) devices
  pair DEVICE       begin pairing with a discovered peer
  revoke DEVICE     revoke a paired device`)
		return nil
	}
	switch args[0] {
	case "status":
		return cmdSyncCall(gf, "sync.status", nil)
	case "peers":
		return cmdSyncCall(gf, "sync.peers", nil)
	case "put":
		return cmdSyncPut(gf, args[1:])
	case "get":
		return cmdSyncGet(gf, args[1:])
	case "start":
		return cmdSyncCall(gf, "sync.start", nil)
	case "stop":
		return cmdSyncCall(gf, "sync.stop", nil)
	case "list-pairs":
		return cmdSyncListPairs(gf)
	case "pair":
		return cmdSyncPair(gf, args[1:])
	case "revoke":
		return cmdSyncRevoke(gf, args[1:])
	case "sync_with", "sync-with":
		return cmdSyncWith(gf, args[1:])
	default:
		return fmt.Errorf("unknown sync subcommand %q", args[0])
	}
}

func cmdSyncCall(gf *globalFlags, method string, params map[string]any) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out any
	if err := c.Call(ctx, method, params, &out); err != nil {
		return err
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	raw, err := jsonMarshal(out)
	if err != nil {
		return err
	}
	fmt.Println(string(raw))
	return nil
}

func cmdSyncPut(gf *globalFlags, args []string) error {
	if len(args) < 2 {
		fmt.Println("usage: synaptic sync put KEY VALUE")
		return nil
	}
	key, value := args[0], args[1]
	return cmdSyncCall(gf, "sync.put", map[string]any{"key": key, "value": []byte(value)})
}

func cmdSyncGet(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: synaptic sync get KEY")
		return nil
	}
	return cmdSyncCall(gf, "sync.get", map[string]any{"key": args[0]})
}

func cmdSyncListPairs(gf *globalFlags) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out struct {
		Devices []struct {
			DeviceID  string `json:"device_id"`
			Name      string `json:"device_name"`
			PublicKey string `json:"public_key"`
			PairedAt  string `json:"paired_at"`
		} `json:"devices"`
	}
	if err := c.Call(ctx, "sync.list_pairs", nil, &out); err != nil {
		return err
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	if len(out.Devices) == 0 {
		fmt.Println("no paired devices")
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "DEVICE_ID\tNAME\tPUBLIC_KEY\tPAIRED_AT\n")
	for _, d := range out.Devices {
		fmt.Fprintf(tw, "%s\t%s\t%s...\t%s\n",
			d.DeviceID, d.Name, truncateMiddle(d.PublicKey, 16), d.PairedAt)
	}
	_ = tw.Flush()
	return nil
}

func cmdSyncPair(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: synaptic sync pair DEVICE_ID [--pin PIN]")
		return nil
	}
	fs := flag.NewFlagSet("sync pair", flag.ContinueOnError)
	pin := fs.String("pin", "", "PIN to confirm (skip to begin pairing)")
	fs.Usage = func() { fmt.Println("usage: synaptic sync pair DEVICE_ID [--pin PIN]") }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	deviceID := fs.Arg(0)
	if *pin == "" {
		// Begin pairing: ask the daemon to generate a token + PIN.
		return cmdSyncCall(gf, "sync.pair_begin", map[string]any{"device_id": deviceID})
	}
	return cmdSyncCall(gf, "sync.pair_confirm", map[string]any{
		"device_id": deviceID,
		"pin":       *pin,
	})
}

func cmdSyncRevoke(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: synaptic sync revoke DEVICE_ID")
		return nil
	}
	return cmdSyncCall(gf, "sync.revoke", map[string]any{"device_id": args[0]})
}

func cmdSyncWith(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: synaptic sync sync_with DEVICE_ID")
		return nil
	}
	return cmdSyncCall(gf, "sync.sync_with", map[string]any{"device_id": args[0]})
}

func truncateMiddle(s string, n int) string {
	if len(s) <= n {
		return s
	}
	half := n / 2
	return s[:half] + "..." + s[len(s)-half:]
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func jsonMarshal(v any) ([]byte, error) {
	// thin wrapper so we don't import encoding/json in this file
	// (the main.go already imports it).
	return jsonMarshalImpl(v)
}

func jsonMarshalImpl(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// strToInt is a helper used elsewhere.
func strToInt(s string) (int, error) {
	return strconv.Atoi(s)
}
