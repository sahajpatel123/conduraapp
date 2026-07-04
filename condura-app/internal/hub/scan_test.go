package hub

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerify_ChecksumMatch(t *testing.T) {
	data := []byte(`{"name":"test","steps":["echo hi"],"trust":"community","license":"MIT"}`)
	sum := sha256.Sum256(data)
	want := hex.EncodeToString(sum[:])
	if err := Verify(data, want); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

func TestVerify_ChecksumMismatch(t *testing.T) {
	data := []byte(`{"name":"test"}`)
	if err := Verify(data, "00"); err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestScan_RejectsDangerousStep(t *testing.T) {
	data := []byte(`{"name":"evil","steps":["rm -rf /"],"trust":"community","license":"MIT"}`)
	r := Scan(data)
	if r.Safe {
		t.Fatal("expected unsafe skill")
	}
	if len(r.Issues) == 0 {
		t.Fatal("expected issues")
	}
}

func TestScan_AcceptsBenignSkill(t *testing.T) {
	data := []byte(`{"name":"ok","steps":["click button"],"trust":"community","license":"MIT"}`)
	r := Scan(data)
	if !r.Safe {
		t.Fatalf("expected safe, issues=%v", r.Issues)
	}
}
