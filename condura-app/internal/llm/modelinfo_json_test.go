package llm

import (
	"encoding/json"
	"testing"
)

// Meridian Ask binds model options to JSON field "id".
func TestModelInfo_JSONUsesSnakeID(t *testing.T) {
	b, err := json.Marshal(ModelInfo{ID: "llama3.3", DisplayName: "Llama"})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["id"] != "llama3.3" {
		t.Fatalf("json id = %v, want llama3.3 (got %s)", m["id"], string(b))
	}
	if _, ok := m["ID"]; ok {
		t.Fatal("must not emit PascalCase ID for GUI")
	}
}
