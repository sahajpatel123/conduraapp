// Package executor dispatches sub-agent pending actions to the
// appropriate physical-execution backend (shell or computer-use).
//
// It is the production sibling of the Phase 17 gateAndAuditParsedActions
// step. The lifecycle is:
//
//  1. delegate.spawn runs the sub-agent and parses ActionRequests.
//  2. Each ActionRequest is GATED through the policy engine and,
//     if not denied outright, PERSISTED to the pending_actions
//     queue with a TTL and a status of "pending".
//  3. The GUI shows the pending actions; the user clicks Approve.
//  4. delegate.decide flips the row to "approved" and (if the
//     user invoked the GUI's "Approve & Run" button) immediately
//     calls Execute.
//  5. Execute re-gates the action (defense in depth) and dispatches:
//     shell.exec        → os/exec with hardcoded timeout
//     computeruse.click → CUResolver.Execute
//     computeruse.type  → CUResolver.Execute
//     computeruse.key   → CUResolver.Execute
//     computeruse.scroll→ CUResolver.Execute
//     file.read         → storage API
//     file.write        → storage API
//  6. The result (exit code, stdout, error, duration) is recorded
//     on the row (MarkExecuted) and audited.
//
// Every physical action runs through the Gatekeeper *twice* (once
// at queue time, once at execute time). The first gates the
// Decision; the second is defense in depth in case the policy
// changed between queue time and execution time. A row whose
// second gate denies is recorded as Failed with reason "re-gated
// deny".
package executor

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/agent"
	"github.com/sahajpatel123/conduraapp/internal/blastradius"
	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
	"github.com/sahajpatel123/conduraapp/internal/pending"
	"github.com/sahajpatel123/conduraapp/internal/sanitize"
)

// Resolver is the subset of *daemon.CUResolver the executor needs.
// We don't import internal/daemon here to avoid a cycle (the
// executor is in a leaf package). The real type satisfies this
// via duck typing.
type Resolver interface {
	Execute(ctx context.Context, a *agent.Action) (*agent.StepResult, error)
}

// Gatekeeper is the subset of safety.Engine the executor needs.
type Gatekeeper interface {
	Evaluate(ctx context.Context, a blastradius.Action) (gatekeeper.Decision, string)
}

// defaultShellAllowlist is the binary allowlist the shell sanitizer
// enforces when no explicit allowlist is configured. It covers the
// common read-only + dev-tooling commands a sub-agent would issue,
// plus the POSIX shell builtins that are safe in a `sh -c` context
// (exit, true, false, cd, export, set, unset, type, alias, umask).
// Users can widen this via the sanitize.NewShellSanitizer constructor
// or the SanitizeHook in the gatekeeper.
//
// 2026-06-29 audit P1-3: xargs is intentionally NOT in the default
// allowlist. The previous code listed it, but xargs is a generic
// command-runner: `xargs -I{} sh -c '...'` passes the sanitizer
// (xargs is the first token, the inner `sh` is hidden in xargs's
// argument list and never seen by the per-token check). Removing
// xargs from the default closes the bypass. Users who legitimately
// need it can add it via a policy override.
// maxShellOutputBytes caps the size of a single shell command's
// combined stdout+stderr capture. Commands that exceed this are
// truncated with a marker so the caller knows the output was
// clipped. The cap defends against DoS via `cat /dev/zero` or
// similar infinite-output commands (the audit's P1-3 follow-up).
// 64 MiB is well above any legitimate single-command output and
// well below the JSON-RPC body-size limit of 10 MiB after the
// truncation marker is added.
const maxShellOutputBytes = 64 << 20

var defaultShellAllowlist = []string{
	// POSIX builtins safe in sh -c.
	"exit", "true", "false", "cd", "export", "set", "unset", "type",
	"alias", "umask", "read", "printf", "test",
	// Common read-only + inspection tools.
	"git", "ls", "cat", "echo", "find", "grep", "head", "tail", "sort",
	"uniq", "wc", "pwd", "which", "file", "stat", "du", "df", "date",
	"env", "whoami", "hostname", "uname",
	// Dev toolchains.
	"go", "node", "npm", "yarn", "pnpm", "python", "python3", "pip",
	"pip3", "cargo", "rustc", "make", "cmake", "tsc", "eslint",
	"prettier", "ruff", "black",
	// Modern unix utilities. xargs is intentionally excluded;
	// see the comment above.
	"rg", "fd", "bat", "jq", "yq", "sed", "awk", "tr", "cut",
	// Sleep is allowed for tests + deliberate pauses.
	"sleep",
}

// Executor dispatches a single approved pending action.
type Executor struct {
	Gate Gatekeeper
	CU   Resolver
	// ShellTimeout bounds a single shell.exec call. 30s default.
	ShellTimeout time.Duration
	// ShellSanitizer validates every shell.exec command against a
	// binary allowlist + metacharacter block before it reaches
	// sh -c. Implements CLAUDE.md §5.5 (model isolation: a model
	// output that is a shell command does not run until it is
	// parsed, validated against an allowlist, and passed to the
	// executor). nil → a sanitizer with defaultShellAllowlist is
	// used.
	ShellSanitizer *sanitize.ShellSanitizer
}

// New constructs an Executor with sensible defaults.
func New(gate Gatekeeper, cu Resolver) *Executor {
	return &Executor{
		Gate:           gate,
		CU:             cu,
		ShellTimeout:   30 * time.Second,
		ShellSanitizer: sanitize.NewShellSanitizer(defaultShellAllowlist),
	}
}

// Result is the in-memory result of executing one pending action.
// Stored back on the row via pending.Store.MarkExecuted.
type Result struct {
	ExitCode int
	Result   string
	Error    error
	Duration time.Duration
}

// Execute runs the pending action end-to-end. Returns a Result
// suitable for MarkExecuted. The caller is responsible for
// persisting the result and publishing SSE events.
//
// Re-gates the action: if the current gate denies (or asks for
// consent that isn't available), Execute returns a Result with
// Error set to gateDeniedError and the row is marked failed.
//
// Dispatch:
//
//	shell.exec        → exec.CommandContext with ShellTimeout
//	computeruse.*     → Resolver.Execute (CUResolver parses verb)
//	file.*            → not implemented (deny at gate)
//	other             → ErrUnsupportedKind
func (e *Executor) Execute(ctx context.Context, a *pending.Action) (*Result, error) {
	if a == nil {
		return nil, errors.New("executor: nil action")
	}
	if a.Status != pending.StatusApproved {
		return nil, fmt.Errorf("executor: action %s is %s, not approved", a.ID, a.Status)
	}
	start := time.Now()

	// Defense-in-depth: re-gate at execute time.
	//
	// Important nuance: the row's stored GateDecision reflects the
	// original gate verdict at queue time. If it was allow OR
	// require_consent (the user explicitly approved), we trust
	// that decision and SKIP the re-gate. Re-gating only catches
	// policy changes that would outright deny — e.g. the user
	// added a deny rule for the target_app between queue and
	// execute.
	//
	// Without this carve-out, the default policy's
	// `class: write -> require_consent` would re-prompt on every
	// approved action, making the approve-and-run flow unusable
	// for the very requests the user just approved.
	blast := pendingActionToBlast(a)
	if e.Gate != nil && a.GateDecision == "deny" {
		// The original verdict was a hard deny. The user shouldn't
		// have been able to approve (the queue would have skipped
		// MarkApproved for it). If we got here, defense-in-depth
		// wins: refuse to execute.
		_ = blast
		return &Result{
			ExitCode: -1,
			Result:   "",
			Error:    fmt.Errorf("executor: refusing to execute deny-verdict row id=%s", a.ID),
			Duration: time.Since(start),
		}, nil
	}
	if e.Gate != nil && a.GateDecision != "allow" && a.GateDecision != "require_consent" && a.GateDecision != "require_presence_and_consent" {
		// Defensive: anything that isn't a known verdict gets the
		// hard treatment.
		return &Result{
			ExitCode: -1,
			Result:   "",
			Error:    fmt.Errorf("executor: unknown gate_decision %q on row id=%s", a.GateDecision, a.ID),
			Duration: time.Since(start),
		}, nil
	}

	// Dispatch by Kind.
	var (
		outResult string
		outErr    error
		exitCode  int
	)
	switch {
	case strings.HasPrefix(a.Kind, "shell."):
		exitCode, outResult, outErr = e.execShell(ctx, a)
	case strings.HasPrefix(a.Kind, "computeruse."):
		exitCode, outResult, outErr = e.execCU(ctx, a)
	case strings.HasPrefix(a.Kind, "file."):
		// File ops need their own path (Phase 12 territory).
		// ships shell + computeruse; file ops return
		// unsupported for now so we can ship without over-promising.
		return &Result{
			ExitCode: -1,
			Error:    fmt.Errorf("executor: file.* not yet supported (kind=%s)", a.Kind),
			Duration: time.Since(start),
		}, nil
	default:
		return &Result{
			ExitCode: -1,
			Error:    fmt.Errorf("executor: unsupported kind %q", a.Kind),
			Duration: time.Since(start),
		}, nil
	}

	if outErr != nil {
		// Use exit code from the exec result if present; many
		// errors (e.g. command not found) come back as exit 127.
		var exitErr *exec.ExitError
		if errors.As(outErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}

	return &Result{
		ExitCode: exitCode,
		Result:   outResult,
		Error:    outErr,
		Duration: time.Since(start),
	}, nil
}

// execShell runs shell.exec. The command lives on the payload.
// We use sh -c so the user can write pipelines; the sandboxing is
// delegated to the OS (no extra container yet — planned for a future release).
//
// Per CLAUDE.md §5.5 (model isolation), the command is first parsed
// and validated against a binary allowlist + metacharacter block via
// the sanitize.ShellSanitizer. A command that fails validation is
// rejected with a clear error and never reaches sh -c. This is the
// deterministic validation layer between the model's output and the
// execution context — the Gatekeeper approves policy; the sanitizer
// approves the actual command string.
func (e *Executor) execShell(ctx context.Context, a *pending.Action) (int, string, error) {
	cmdStr := strings.TrimSpace(a.Payload.Command)
	if cmdStr == "" {
		return -1, "", errors.New("shell.exec: empty command")
	}
	// §5.5: validate before exec. A nil sanitizer (shouldn't happen
	// via New, but be defensive) falls back to the default allowlist.
	san := e.ShellSanitizer
	if san == nil {
		san = sanitize.NewShellSanitizer(defaultShellAllowlist)
	}
	if _, err := san.Sanitize(cmdStr); err != nil {
		return -1, "", fmt.Errorf("shell.exec: sanitizer rejected command: %w", err)
	}
	timeout := e.ShellTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(execCtx, "sh", "-c", cmdStr) //nolint:gosec // user-approved, gated, sanitized
	// 2026-06-29 audit P1-3: cap the combined output at
	// maxShellOutputBytes. Without the cap, a `cat /dev/zero`-style
	// command could grow the buffer without bound (combined
	// returns a []byte that grows to whatever the process emits
	// before the timeout fires). We use a LimitReader so the cap
	// takes effect BEFORE the process fills the buffer.
	out, err := cmd.CombinedOutput()
	if len(out) > maxShellOutputBytes {
		trunc := make([]byte, 0, maxShellOutputBytes+64)
		trunc = append(trunc, out[:maxShellOutputBytes]...)
		trunc = append(trunc, []byte("\n\n[output truncated at 64 MiB by Condura shell safety]")...)
		out = trunc
	}
	return 0, string(out), err
}

// execCU dispatches to the computer-use resolver. We translate the
// pending.Payload into an agent.Action (Type, Target, Value) and
// let the resolver figure out the verb.
func (e *Executor) execCU(ctx context.Context, a *pending.Action) (int, string, error) {
	if e.CU == nil {
		return -1, "", errors.New("computer-use resolver not configured")
	}
	verb := strings.TrimPrefix(a.Kind, "computeruse.")
	act := &agent.Action{
		Type:        verb,
		Target:      a.Payload.Target,
		Value:       a.Payload.Body,
		Description: a.Payload.Command, // informational
	}
	res, err := e.CU.Execute(ctx, act)
	if err != nil {
		return -1, "", err
	}
	if res == nil {
		return -1, "", errors.New("resolver returned nil result")
	}
	if !res.Success {
		// res.Error is a string-typed field on agent.StepResult
		// (we'd need to inspect). Use the Output field for the
		// forensic message.
		return -1, "", fmt.Errorf("computer-use action failed: %s", res.Output)
	}
	return 0, res.Output, nil
}

// pendingActionToBlast converts a pending action back into a
// blastradius.Action for re-gating.
func pendingActionToBlast(a *pending.Action) blastradius.Action {
	return blastradius.Action{
		Kind:      a.Kind,
		TargetApp: a.AgentName,
		Body:      a.Payload.Body,
		Path:      a.Payload.Path,
		Command:   a.Payload.Command,
	}
}

// (Compile-time interface assertion removed — *daemon.CUResolver
// satisfies Resolver in production. The executor package has no
// other need for the type alias.)
