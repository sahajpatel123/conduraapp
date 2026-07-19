package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/ipc"
)

// captureStdout redirects os.Stdout while fn runs, returning
// what was written. Used by the explain tests to capture the
// human-readable output without polluting the test runner's
// output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()
	_ = w.Close()
	<-done
	_ = r.Close()
	return buf.String()
}

// captureStdoutErr is captureStdout with an additional return
// value for the test to assert on (e.g. an error returned
// from the function under test).
func captureStdoutErr(t *testing.T, fn func() (any, error)) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	_, callErr := fn()
	_ = w.Close()
	<-done
	_ = r.Close()
	return buf.String(), callErr
}

// TestExplain_ListAllCodes pins the "no args" behavior:
// condura explain with no args lists all known codes. The
// list MUST stay sorted by code (descending) so the operator
// can scan it quickly. Operators often run `condura explain |
// grep "32603"` to find a code; the sort makes that grep work.
func TestExplain_ListAllCodes(t *testing.T) {
	out := captureStdout(t, func() {
		_ = cmdExplain(&globalFlags{}, []string{})
	})
	want := []string{"-32700", "-32600", "-32601", "-32602", "-32603", "-32099"}
	for _, c := range want {
		if !strings.Contains(out, c) {
			t.Errorf("list output missing code %s\n---\n%s\n---", c, out)
		}
	}
	idx32099 := strings.Index(out, "-32099")
	idx32603 := strings.Index(out, "-32603")
	if idx32099 < idx32603 {
		t.Errorf("expected sort-by-code-descending; -32099 (%d) appeared before -32603 (%d)", idx32099, idx32603)
	}
}

// TestExplain_LookupKnownCode pins the happy-path lookup:
// given a known code, the output includes the code, the
// human-readable name, the explanation, and the next step.
func TestExplain_LookupKnownCode(t *testing.T) {
	out := captureStdout(t, func() {
		if err := cmdExplain(&globalFlags{}, []string{"--", "-32603"}); err != nil {
			t.Fatal(err)
		}
	})
	want := []string{
		"Code:", "Internal error",
		"What:",
		"Next:",
		"condura logs",
	}
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("explain output missing %q\n---\n%s\n---", w, out)
		}
	}
}

// TestExplain_LookupSafetyDenial pins the safety-policy
// message: -32099 is project-specific (not part of JSON-RPC
// 2.0 standard codes), so it must be in the registry and must
// include the "GUI prompted" guidance — that's the #1 reason
// operators hit this code.
func TestExplain_LookupSafetyDenial(t *testing.T) {
	out := captureStdout(t, func() {
		if err := cmdExplain(&globalFlags{}, []string{"--", "-32099"}); err != nil {
			t.Fatal(err)
		}
	})
	want := []string{"Denied by safety policy", "Gatekeeper", "consent prompt"}
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("explain -32099 output missing %q\n---\n%s\n---", w, out)
		}
	}
}

// TestExplain_AcceptsPositiveForm pins the sign convention:
// "32603" works the same as "-32603". Operators shouldn't have
// to remember that the codes are negative.
func TestExplain_AcceptsPositiveForm(t *testing.T) {
	out := captureStdout(t, func() {
		if err := cmdExplain(&globalFlags{}, []string{"32603"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "Code:    -32603") {
		t.Errorf("positive form should normalize to negative\n---\n%s\n---", out)
	}
}

// TestExplain_AcceptsPlusPrefix pins the "+32603" form too:
// some tools print "+32603" instead of "-32603" when formatting
// errors. Operators may copy-paste that.
func TestExplain_AcceptsPlusPrefix(t *testing.T) {
	out := captureStdout(t, func() {
		if err := cmdExplain(&globalFlags{}, []string{"+32603"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "Code:    -32603") {
		t.Errorf("+32603 should normalize to -32603\n---\n%s\n---", out)
	}
}

// TestExplain_UnknownCodeErrorsCleanly pins the error path:
// an unknown code returns a clear error, NOT a panic.
func TestExplain_UnknownCodeErrorsCleanly(t *testing.T) {
	_, err := captureStdoutErr(t, func() (any, error) {
		return nil, cmdExplain(&globalFlags{}, []string{"99999"})
	})
	if err == nil {
		t.Fatal("condura explain 99999: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "99999") {
		t.Errorf("error should include the bad code; got: %v", err)
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should describe the failure; got: %v", err)
	}
}

// TestExplain_InvalidInputErrorsCleanly pins the parse-error
// path: a non-number returns a clear error, NOT a panic.
func TestExplain_InvalidInputErrorsCleanly(t *testing.T) {
	_, err := captureStdoutErr(t, func() (any, error) {
		return nil, cmdExplain(&globalFlags{}, []string{"abc"})
	})
	if err == nil {
		t.Fatal("condura explain abc: expected error, got nil")
	}
	if !strings.Contains(err.Error(), `"abc"`) {
		t.Errorf("error should quote the bad input; got: %v", err)
	}
}

// TestExplain_KnowsAllIPCCodes is the regression-prevention
// test: the registry MUST include every code defined in the
// ipc package. If a future contributor adds ipc.CodeFooBar but
// forgets to add it to knownErrors, this test fails — the
// operator gets a clean "code is defined but undocumented"
// message instead of a mysterious "unknown code".
func TestExplain_KnowsAllIPCCodes(t *testing.T) {
	codes := []int{
		ipc.CodeParseError,
		ipc.CodeInvalidRequest,
		ipc.CodeMethodNotFound,
		ipc.CodeInvalidParams,
		ipc.CodeInternalError,
	}
	for _, c := range codes {
		if _, ok := knownErrorsByCode[c]; !ok {
			t.Errorf("knownErrors is missing ipc.Code* %d; add it to explain.go's knownErrors table", c)
		}
	}
}