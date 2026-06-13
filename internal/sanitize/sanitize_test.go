package sanitize

import "testing"

func TestShellSanitizer_AllowsSafeCommand(t *testing.T) {
	s := NewShellSanitizer(nil)
	_, err := s.Sanitize("git status")
	if err != nil {
		t.Fatalf("safe command denied: %v", err)
	}
}

func TestShellSanitizer_RejectsDisallowedBin(t *testing.T) {
	s := NewShellSanitizer(nil)
	_, err := s.Sanitize("rm -rf /")
	if err == nil {
		t.Fatal("disallowed binary should be rejected")
	}
}

func TestShellSanitizer_RejectsPipe(t *testing.T) {
	s := NewShellSanitizer(nil)
	_, err := s.Sanitize("ls | grep secret")
	if err == nil {
		t.Fatal("pipe should be rejected")
	}
}

func TestShellSanitizer_RejectsBacktick(t *testing.T) {
	s := NewShellSanitizer(nil)
	_, err := s.Sanitize("echo `whoami`")
	if err == nil {
		t.Fatal("backtick should be rejected")
	}
}

func TestShellSanitizer_RejectsCommandSub(t *testing.T) {
	s := NewShellSanitizer(nil)
	_, err := s.Sanitize("echo $(whoami)")
	if err == nil {
		t.Fatal("command substitution should be rejected")
	}
}

func TestPathSanitizer_RejectsTraversal(t *testing.T) {
	s := NewPathSanitizer()
	_, err := s.Sanitize("../../../etc/passwd")
	if err == nil {
		t.Fatal("path traversal should be rejected")
	}
}

func TestPathSanitizer_AllowsSafePath(t *testing.T) {
	s := NewPathSanitizer()
	_, err := s.Sanitize("/Users/test/report.pdf")
	if err != nil {
		t.Fatalf("safe path denied: %v", err)
	}
}

func TestPathSanitizer_RejectsSystemPath(t *testing.T) {
	s := NewPathSanitizer()
	_, err := s.Sanitize("/etc/shadow")
	if err == nil {
		t.Fatal("system path should be rejected")
	}
}

func TestURLSanitizer_RejectsHTTP(t *testing.T) {
	s := NewURLSanitizer()
	_, err := s.Sanitize("http://example.com")
	if err == nil {
		t.Fatal("HTTP should be rejected (protocol downgrade)")
	}
}

func TestURLSanitizer_RejectsLocalhost(t *testing.T) {
	s := NewURLSanitizer()
	_, err := s.Sanitize("https://localhost:8080/admin")
	if err == nil {
		t.Fatal("localhost should be rejected (SSRF)")
	}
}

func TestURLSanitizer_RejectsPrivateIP(t *testing.T) {
	tests := []string{
		"https://192.168.1.1/admin",
		"https://10.0.0.1",
		"https://172.16.0.1",
	}
	for _, u := range tests {
		s := NewURLSanitizer()
		_, err := s.Sanitize(u)
		if err == nil {
			t.Errorf("private IP should be rejected: %s", u)
		}
	}
}

func TestURLSanitizer_AllowsHTTPSPublic(t *testing.T) {
	s := NewURLSanitizer()
	_, err := s.Sanitize("https://api.openai.com/v1/chat")
	if err != nil {
		t.Fatalf("public HTTPS should be allowed: %v", err)
	}
}

func TestPIIRegexSanitizer_DetectsCC(t *testing.T) {
	s := NewPIIRegexSanitizer()
	// Valid test card number (passes Luhn).
	_, err := s.Sanitize("my card is 4532015112830366")
	if err == nil {
		t.Fatal("credit card should be detected")
	}
}

func TestPIIRegexSanitizer_DetectsSSN(t *testing.T) {
	s := NewPIIRegexSanitizer()
	_, err := s.Sanitize("SSN: 123-45-6789")
	if err == nil {
		t.Fatal("SSN pattern should be detected")
	}
}

func TestPIIRegexSanitizer_AllowsPlainText(t *testing.T) {
	s := NewPIIRegexSanitizer()
	_, err := s.Sanitize("hello world")
	if err != nil {
		t.Fatal("plain text should pass")
	}
}

func TestPythonImportSanitizer_RejectsBanned(t *testing.T) {
	s := NewPythonImportSanitizer()
	tests := []string{"import os.system", "import subprocess", "eval(", "exec("}
	for _, code := range tests {
		_, err := s.Sanitize(code)
		if err == nil {
			t.Errorf("banned import should be rejected: %s", code)
		}
	}
}

func TestPythonImportSanitizer_AllowsSafeCode(t *testing.T) {
	s := NewPythonImportSanitizer()
	_, err := s.Sanitize("print('hello')")
	if err != nil {
		t.Fatal("safe Python should pass")
	}
}

func TestChain_RunsAll(t *testing.T) {
	chain := DefaultChain()
	out, err := Chain(chain, "echo hello")
	if err != nil {
		t.Fatalf("chain should allow safe echo: %v", err)
	}
	if out != "echo hello" {
		t.Errorf("output = %q", out)
	}
}

func TestChain_FirstErrorStops(t *testing.T) {
	chain := []Sanitizer{
		NewURLSanitizer(),
		NewShellSanitizer(nil), // won't be reached
	}
	_, err := Chain(chain, "http://evil.com")
	if err == nil {
		t.Fatal("chain should stop at first error")
	}
}
