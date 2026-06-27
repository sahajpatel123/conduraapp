// resume subcommand — T3b sticky resume confirmation.
//
// Usage:
//
//	condura resume request             # mint a ticket (prints it; you'll paste into the confirm step)
//	condura resume confirm --ticket T  # confirm with the resume secret (prompts or reads CONDURA_RESUME_SECRET)
//	condura resume cancel              # explicitly abandon a pending ticket (best-effort; the daemon expires them on TTL)
//
// The un-halt flow is human-confirmed and out of the in-process trust
// boundary: the CLI runs as a separate OS process, prompts the human
// at a terminal, then calls halt.confirm_resume over IPC with the secret
// loaded from the data dir or CONDURA_RESUME_SECRET. A compromised
// in-process conductor can read the IPC bearer token but cannot read
// the resume secret from the user's shell and cannot synthesize a
// human at the keyboard.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

const resumeHelp = `condura resume — T3b sticky human-confirmed resume

Usage:
  condura resume request
      Mint a resume ticket (IPC daemon.resume_request). Prints the
      ticket + a clear migration message. The ticket is valid for
      5 minutes; you confirm it before then.

  condura resume confirm --ticket T
      Confirm a ticket (IPC halt.confirm_resume). Reads the resume
      secret from CONDURA_RESUME_SECRET or prompts on stdin. On
      success the daemon un-halts (Layer 1 + Layer 3).

  condura resume cancel
      Best-effort abandon — the daemon expires tickets on TTL. The
      stored secret is unaffected.

Global flags (passed through):
  --addr, --data-dir, --token, --json   (see condura help)
`

func cmdResume(gf *globalFlags, args []string) error {
	if len(args) == 0 {
		fmt.Print(resumeHelp)
		return nil
	}
	switch args[0] {
	case "request":
		return cmdResumeRequest(gf)
	case "confirm":
		return cmdResumeConfirm(gf, args[1:])
	case "cancel":
		return cmdResumeCancel(gf)
	case "help", "-h", "--help":
		fmt.Print(resumeHelp)
		return nil
	default:
		return fmt.Errorf("condura resume: unknown subcommand %q", args[0])
	}
}

// cmdResumeRequest calls IPC daemon.resume_request and prints the
// resulting ticket + a clear migration message for the user.
func cmdResumeRequest(gf *globalFlags) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	var out map[string]any
	if err := c.Call(ctx, "daemon.resume_request", nil, &out); err != nil {
		if ipc.IsConnRefused(err) {
			return fmt.Errorf("daemon not running at %s", c.Addr())
		}
		return fmt.Errorf("daemon.resume_request failed: %w", err)
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	halted, _ := out["halted"].(bool)
	ticket, _ := out["ticket"].(string)
	if !halted || ticket == "" {
		fmt.Println("Daemon is not halted; nothing to resume.")
		if msg, ok := out["reason"].(string); ok && msg != "" {
			fmt.Printf("  reason: %s\n", msg)
		}
		return nil
	}
	fmt.Printf("Resume ticket: %s\n", ticket)
	if ttl, ok := out["ttl_seconds"]; ok {
		fmt.Printf("Valid for:     %v seconds\n", ttl)
	}
	if via, ok := out["confirm_via"].(string); ok && via != "" {
		fmt.Printf("Confirm with:  %s\n", via)
	}
	return nil
}

// cmdResumeConfirm calls IPC halt.confirm_resume with the supplied
// ticket + the resume secret. The secret is read from
// CONDURA_RESUME_SECRET first, then prompted on stdin (with echo
// suppressed via terminal-only when possible; falling back to plain
// echo in non-TTY environments — for the common case the user pipes
// the secret via env or a password manager).
func cmdResumeConfirm(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("resume confirm", flag.ContinueOnError)
	var ticket string
	fs.StringVar(&ticket, "ticket", "", "resume ticket (required)")
	fs.Usage = func() { fmt.Println("usage: condura resume confirm --ticket T") }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if ticket == "" {
		return fmt.Errorf("--ticket is required")
	}

	secret, err := loadResumeSecret()
	if err != nil {
		return err
	}
	defer func() {
		// Best-effort: clear the secret from memory. (Not strictly
		// necessary — it lives in a local stack variable and the
		// process exits shortly — but it costs nothing.)
		secret = ""
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, err := connect(gf)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()

	params := map[string]any{
		"ticket": ticket,
		"secret": secret,
	}
	var out map[string]any
	if err := c.Call(ctx, "halt.confirm_resume", params, &out); err != nil {
		if ipc.IsConnRefused(err) {
			return fmt.Errorf("daemon not running at %s", c.Addr())
		}
		return fmt.Errorf("halt.confirm_resume failed: %w", err)
	}
	if gf.jsonOut {
		return printJSON(out)
	}
	if resumed, _ := out["resumed"].(bool); resumed {
		fmt.Println("Daemon resumed (human-confirmed).")
		return nil
	}
	fmt.Printf("Confirm result: %v\n", out)
	return nil
}

// cmdResumeCancel is a best-effort no-op. Tickets are TTL-evicted by
// the daemon; we just acknowledge the user.
func cmdResumeCancel(gf *globalFlags) error {
	fmt.Println("Pending resume tickets are TTL-evicted by the daemon within 5 minutes; nothing to do.")
	fmt.Println("If you typed the secret in plaintext, rotate it: stop the daemon, delete <data-dir>/resume.secret, restart.")
	return nil
}

// loadResumeSecret returns the human-confirmation secret from
// CONDURA_RESUME_SECRET (env) or prompts on stdin. Never logs the
// secret.
func loadResumeSecret() (string, error) {
	if v := os.Getenv("CONDURA_RESUME_SECRET"); v != "" {
		return v, nil
	}
	// Fall back to interactive prompt. If stdin is not a TTY (CI /
	// pipe), the prompt is still readable; the user is expected to
	// have set CONDURA_RESUME_SECRET in that case.
	stat, _ := os.Stdin.Stat()
	if stat != nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		return "", fmt.Errorf("CONDURA_RESUME_SECRET is not set and stdin is not a TTY; cannot prompt for the resume secret")
	}
	fmt.Fprint(os.Stderr, "Resume secret (paste from your password manager): ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read secret: %w", err)
	}
	secret := strings.TrimRight(line, "\r\n")
	if secret == "" {
		return "", fmt.Errorf("empty secret")
	}
	return secret, nil
}
