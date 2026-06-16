package onboarding

import (
	"context"
	"testing"
)

func TestProbePower_NoOllama(t *testing.T) {
	pp := ProbePower(context.Background())
	if pp.OllamaReachable {
		t.Skip("ollama is running locally — cannot test unreachable path")
	}
	if pp.Recommended != "none" {
		t.Fatalf("recommended: want none, got %s", pp.Recommended)
	}
}

func TestProbePower_OllamaReachable(t *testing.T) {
	pp := &PowerProbe{
		OllamaReachable: true,
		OllamaModels:    []string{"llama3.2:latest", "mistral:7b"},
	}
	if !pp.OllamaReachable {
		t.Fatal("ollama should be reachable")
	}
	if pp.FirstModel() != "llama3.2:latest" {
		t.Fatalf("FirstModel: want llama3.2:latest, got %s", pp.FirstModel())
	}
	if pp.NoModels() {
		t.Fatal("NoModels should be false when models present")
	}
}

func TestProbePower_NoModels(t *testing.T) {
	pp := &PowerProbe{OllamaReachable: true}
	if !pp.NoModels() {
		t.Fatal("NoModels should be true when zero models")
	}
	if pp.FirstModel() != RecommendedOllamaModel {
		t.Fatalf("FirstModel fallback: want %s, got %s", RecommendedOllamaModel, pp.FirstModel())
	}
}

func TestProbePower_Recommended(t *testing.T) {
	pp := ProbePower(context.Background())
	if pp.Recommended == "" {
		t.Fatal("recommended should not be empty")
	}
	if pp.Recommended != "none" && pp.Recommended != "ollama" {
		t.Fatalf("recommended: unexpected value %s", pp.Recommended)
	}
}

func TestProbePower_CLIProbesExist(t *testing.T) {
	pp := ProbePower(context.Background())
	if len(pp.CLIs) == 0 {
		t.Fatal("CLI probes should include at least 2 agents from delegation config")
	}
	names := map[string]bool{}
	for _, c := range pp.CLIs {
		if c.Name == "" {
			t.Fatal("CLI probe has empty name")
		}
		names[c.Name] = c.Found
	}
	if !names["claude"] && !names["ollama"] {
		t.Fatal("expected at least claude or ollama in CLI probes")
	}
}

func TestProbePowerWithTimeout(t *testing.T) {
	pp := ProbePowerWithTimeout(context.Background())
	if pp == nil {
		t.Fatal("ProbePowerWithTimeout returned nil")
	}
	if pp.Recommended == "" {
		t.Fatal("recommended should not be empty")
	}
}

func TestPowerProbeError(t *testing.T) {
	e := newPowerProbeError("something went %s", "wrong")
	if e.Error() != "something went wrong" {
		t.Fatalf("Error(): got %q", e.Error())
	}
	if e.Message != "something went wrong" {
		t.Fatalf("Message: got %q", e.Message)
	}
}

func TestOllamaInstallURL(t *testing.T) {
	u := OllamaInstallURL()
	if u == "" {
		t.Fatal("OllamaInstallURL should not be empty")
	}
}
