package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/llm"
)

type fakePlannerLLM struct {
	resp llm.ChatResponse
	err  error
}

func (f *fakePlannerLLM) Chat(_ context.Context, _ llm.ChatRequest) (llm.ChatResponse, error) {
	return f.resp, f.err
}

func responseJSON(goal string, steps ...planningStep) llm.ChatResponse {
	b, _ := json.Marshal(planningResponse{Goal: goal, Steps: steps})
	return llm.ChatResponse{
		Message:      llm.Message{Role: llm.RoleAssistant, Content: string(b)},
		FinishReason: "stop",
	}
}

func TestLLMPlanner_Decompose(t *testing.T) {
	fake := &fakePlannerLLM{resp: responseJSON("Search for meeting notes",
		planningStep{Type: "click", Target: "Search field", Description: "Click search"},
		planningStep{Type: "type", Value: "meeting notes", Description: "Type query"},
		planningStep{Type: "click", Target: "Search button", Description: "Submit"},
	)}
	p := NewLLMPlanner(fake, "gpt-4")

	plan, err := p.Decompose(context.Background(), "Search for meeting notes", nil)
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	if plan.Goal != "Search for meeting notes" {
		t.Errorf("goal = %q", plan.Goal)
	}
	if len(plan.Steps) != 3 {
		t.Fatalf("got %d steps", len(plan.Steps))
	}
	if plan.Steps[0].Action.Type != "click" {
		t.Errorf("step[0].type = %q", plan.Steps[0].Action.Type)
	}
	if plan.Steps[1].Action.Value != "meeting notes" {
		t.Errorf("step[1].value = %q", plan.Steps[1].Action.Value)
	}
	if len(plan.Steps[1].DependsOn) != 1 || plan.Steps[1].DependsOn[0] != 0 {
		t.Errorf("step[1].dependsOn = %v", plan.Steps[1].DependsOn)
	}
}

func TestLLMPlanner_Decompose_JSONInCodeBlock(t *testing.T) {
	steps := []planningStep{
		{Type: "launch", Value: "com.apple.Safari", Description: "Open Safari"},
	}
	b, _ := json.Marshal(planningResponse{Goal: "Open Safari", Steps: steps})
	wrapped := "```json\n" + string(b) + "\n```"

	fake := &fakePlannerLLM{resp: llm.ChatResponse{
		Message:      llm.Message{Role: llm.RoleAssistant, Content: wrapped},
		FinishReason: "stop",
	}}
	p := NewLLMPlanner(fake, "llama3")

	plan, err := p.Decompose(context.Background(), "Open Safari", nil)
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	if len(plan.Steps) != 1 {
		t.Fatalf("got %d steps", len(plan.Steps))
	}
	if plan.Steps[0].Action.Type != "launch" {
		t.Errorf("step = %q", plan.Steps[0].Action.Type)
	}
}

func TestLLMPlanner_Decompose_EmptyPlan(t *testing.T) {
	fake := &fakePlannerLLM{resp: responseJSON("nothing")}
	p := NewLLMPlanner(fake, "gpt-4")
	_, err := p.Decompose(context.Background(), "do nothing", nil)
	if err == nil {
		t.Fatal("expected error for empty plan")
	}
}

func TestLLMPlanner_Decompose_LLMError(t *testing.T) {
	fake := &fakePlannerLLM{err: context.Canceled}
	p := NewLLMPlanner(fake, "gpt-4")
	_, err := p.Decompose(context.Background(), "test", nil)
	if err == nil {
		t.Fatal("expected error from LLM")
	}
}

func TestLLMPlanner_Decompose_MissingType(t *testing.T) {
	fake := &fakePlannerLLM{resp: responseJSON("test",
		planningStep{Target: "something", Description: "bad step"},
	)}
	p := NewLLMPlanner(fake, "gpt-4")
	_, err := p.Decompose(context.Background(), "test", nil)
	if err == nil {
		t.Fatal("expected error for missing action type")
	}
}

func TestExtractJSONFromMarkdown(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`{"a":1}`, `{"a":1}`},
		{"```json\n{\"a\":1}\n```", `{"a":1}`},
		{"Here is the plan: {\"a\":1}", `{"a":1}`},
		{"{\"a\":1} extra text", `{"a":1}`},
	}
	for _, tt := range tests {
		got := extractJSONFromMarkdown(tt.input)
		if got != tt.want {
			t.Errorf("extractJSONFromMarkdown(%q) = %q", tt.input, got)
		}
	}
}
