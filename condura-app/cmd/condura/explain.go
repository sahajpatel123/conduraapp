package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/ipc"
)

// errorExplanation pairs an IPC error code with a human-readable
// explanation and the next step the user should take. The
// "next step" is the part that makes this command actually
// useful — knowing what the code means is half the battle;
// knowing what to do about it is the other half.
type errorExplanation struct {
	code      int
	name      string
	explanation string
	nextStep  string
}

// knownErrors is the registry of error codes condura can
// explain. New codes added to the IPC layer should also be
// added here — the test (TestExplain_KnowsAllIPCCodes) pins
// this invariant so the registry can't drift out of sync.
//
// Order roughly mirrors the JSON-RPC 2.0 spec, with our
// project-specific codes (-32099 sentinel for "denied by
// safety policy") at the bottom.
var knownErrors = []errorExplanation{
	{
		code:        -32700,
		name:        "Parse error",
		explanation: "The daemon received malformed JSON. The CLI built the request, but it didn't survive serialization.",
		nextStep:    "Re-run the command. If the error persists, the daemon and CLI are out of sync — upgrade both to the same version.",
	},
	{
		code:        -32600,
		name:        "Invalid request",
		explanation: "The request was structurally valid JSON but didn't conform to the JSON-RPC shape (missing method, wrong type, etc.).",
		nextStep:    "Check the CLI's --version. If it matches the daemon, this is a bug — report it with `condura diag` output.",
	},
	{
		code:        -32601,
		name:        "Method not found",
		explanation: "The CLI called an RPC method the daemon doesn't know. Almost always means CLI and daemon are out of sync.",
		nextStep:    "Run `condura version` on both the CLI and the daemon. If they differ, upgrade the older one.",
	},
	{
		code:        -32602,
		name:        "Invalid params",
		explanation: "The request reached the handler but one or more parameters were wrong (missing required field, wrong type, etc.).",
		nextStep:    "Re-run with the --help flag and check the params. The message after the code usually names the bad field.",
	},
	{
		code:        -32603,
		name:        "Internal error",
		explanation: "The daemon hit an unexpected condition. Could be a bug, a resource issue (disk full, OOM), or a downstream service failure.",
		nextStep:    "Run `condura logs --lines 100` to see the daemon's recent activity. If the issue persists, report with `condura diag`.",
	},
	{
		code:        -32099,
		name:        "Denied by safety policy",
		explanation: "The Gatekeeper blocked this action. Condura refused the request because the requested action was too risky, or the user declined the consent prompt.",
		nextStep:    "If the GUI prompted you and you declined, the action is permanently denied for this request — re-trigger the action and accept the prompt. If no prompt appeared, run `condura validate` to check the install state.",
	},
}

// knownErrorsByCode is the lookup index for the registry.
// Built once at package init; not exported because the only
// caller is cmdExplain.
var knownErrorsByCode = func() map[int]errorExplanation {
	m := make(map[int]errorExplanation, len(knownErrors))
	for _, e := range knownErrors {
		m[e.code] = e
	}
	return m
}()

// cmdExplain translates an IPC error code into a human-readable
// explanation with the next step the user should take.
//
// Usage:
//   condura explain -32603
//   condura explain 32603     (positive form, gets normalized)
//   condura explain            (lists all known codes)
//
// Designed for support: when a CLI call returns
// "rpc error -32603: ...", the user (or the support engineer)
// can run `condura explain -32603` to get the explanation
// without reading the daemon source.
//
// No IPC required — this is a local translation table.
// No new internal package needed: the table lives in the CLI
// itself, alongside the user-facing code that needs it.
func cmdExplain(gf *globalFlags, args []string) error {
	fs := flag.NewFlagSet("explain", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil && !errors.Is(err, flag.ErrHelp) {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		// No code given — list all known codes so the operator
		// can see what's available.
		fmt.Println("IPC error codes condura can explain:")
		for _, e := range knownErrors {
			fmt.Printf("  %6d  %s\n", e.code, e.name)
		}
		fmt.Println()
		fmt.Println("Usage: condura explain <code>")
		fmt.Println("  e.g. condura explain -32603")
		return nil
	}

	// Accept the code in either signed (-32603) or unsigned
	// (32603) form. Strip a leading "-" so the user doesn't
	// have to remember the sign convention.
	raw := rest[0]
	if len(raw) > 0 && raw[0] == '+' {
		raw = raw[1:]
	}
	code, err := strconv.Atoi(raw)
	if err != nil {
		return fmt.Errorf("condura explain: %q is not a number (try `condura explain` to list known codes)", rest[0])
	}
	if code > 0 {
		code = -code // normalize "32603" to "-32603"
	}

	// Try the local registry first. If not found, fall back
	// to the ipc package's Code* constants: if the code matches
	// a defined constant but we forgot to add it to knownErrors,
	// the generic explanation still helps the operator.
	entry, ok := knownErrorsByCode[code]
	if !ok {
		// Last-ditch: match against any ipc.Code* constant.
		// This way "condura explain 9999" still produces
		// SOMETHING (an "unknown code" message), and "condura
		// explain -32050" (a hypothetical new code) still works
		// even before we add it to the registry.
		if isKnownIPCCode(code) {
			fmt.Printf("Code %d is defined in the JSON-RPC layer but not documented in condura explain yet.\n", code)
			fmt.Println("This is a code-registration gap; please open an issue or add an entry to the knownErrors table.")
			return nil
		}
		return fmt.Errorf("condura explain: unknown code %d (try `condura explain` to list known codes)", code)
	}

	fmt.Printf("Code:    %d  (%s)\n", entry.code, entry.name)
	fmt.Printf("What:    %s\n", entry.explanation)
	fmt.Printf("Next:    %s\n", entry.nextStep)
	return nil
}

// isKnownIPCCode returns true if code matches one of the
// ipc.Code* constants. Used to give a "defined but
// undocumented" answer for new codes we haven't added to
// knownErrors yet.
//
// Cheap: a small switch. The code we're checking against
// is small (-32000 to -32700 range for JSON-RPC), so a
// linear search through the registry is O(1) for any
// practical number of entries.
func isKnownIPCCode(code int) bool {
	switch code {
	case ipc.CodeParseError,
		ipc.CodeInvalidRequest,
		ipc.CodeMethodNotFound,
		ipc.CodeInvalidParams,
		ipc.CodeInternalError:
		return true
	default:
		return false
	}
}