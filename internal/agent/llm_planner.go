package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sahajpatel123/synapticapp/internal/llm"
)

// LLMPlanner decomposes natural-language tasks into ordered plan steps
// by asking an LLM. It produces agent.Action structs (intent layer)
// which the CUResolver later compiles into executable computeruse.Actions.
type LLMPlanner struct {
	provider PlannerProvider
	model    string
}

// PlannerProvider is the subset of llm.Provider we need for planning.
// Tests implement this with a fake; production passes the real registry.
type PlannerProvider interface {
	Chat(ctx context.Context, req llm.ChatRequest) (llm.ChatResponse, error)
}

// NewLLMPlanner creates a planner backed by an LLM provider.
func NewLLMPlanner(p PlannerProvider, model string) *LLMPlanner {
	return &LLMPlanner{provider: p, model: model}
}

// Decompose sends the task to the LLM and parses the response into a Plan.
func (p *LLMPlanner) Decompose(ctx context.Context, task string, planCtx *Context) (*Plan, error) {
	prompt := p.buildPlanningPrompt(task, planCtx)

	resp, err := p.provider.Chat(ctx, llm.ChatRequest{
		Model: p.model,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: planningSystemPrompt},
			{Role: llm.RoleUser, Content: prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("llm_planner: chat: %w", err)
	}

	content := resp.Message.Content
	steps, goal, err := p.parsePlanningResponse(content)
	if err != nil {
		return nil, fmt.Errorf("llm_planner: parse: %w", err)
	}

	if goal == "" {
		goal = task
	}

	return &Plan{
		Steps:   steps,
		Current: 0,
		Goal:    goal,
		Context: planCtx,
	}, nil
}

// Reprioritize adjusts the plan based on new observations.
func (p *LLMPlanner) Reprioritize(_ context.Context, plan *Plan, _ *Observation) (*Plan, error) {
	return plan, nil
}

func (p *LLMPlanner) buildPlanningPrompt(task string, planCtx *Context) string {
	var b strings.Builder
	b.WriteString("Task: ")
	b.WriteString(task)
	b.WriteString("\n")

	if planCtx != nil {
		if planCtx.UserGoal != "" {
			b.WriteString("Goal: ")
			b.WriteString(planCtx.UserGoal)
			b.WriteString("\n")
		}
		if planCtx.CurrentState != "" {
			b.WriteString("Current screen state: ")
			b.WriteString(planCtx.CurrentState)
			b.WriteString("\n")
		}
		if len(planCtx.History) > 0 {
			b.WriteString("Recent history:\n")
			for _, h := range planCtx.History {
				b.WriteString("- ")
				b.WriteString(h)
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\nReturn a JSON object with:\n")
	b.WriteString("- \"goal\": a one-line summary of what to accomplish\n")
	b.WriteString("- \"steps\": an ordered array of actions, each with:\n")
	b.WriteString("    \"type\": click, type, scroll, key_press, launch, focus, or wait\n")
	b.WriteString("    \"target\": UI element name (e.g. \"Submit button\")\n")
	b.WriteString("    \"value\": text to type, key name, bundle ID, or direction\n")
	b.WriteString("    \"description\": what this step does\n")
	b.WriteString("Return ONLY the JSON object, no markdown.\n")

	return b.String()
}

const planningSystemPrompt = `You are a desktop automation planner. Break user tasks into ordered, atomic computer-use actions.

Rules:
1. Each step is one action: click, type, scroll, key_press, launch, focus, or wait.
2. Use natural language for targets (e.g. "Submit button", "Search field").
3. Specify exact values (text, key names, bundle IDs, directions).
4. Order logically: navigate, interact, confirm.
5. Output only valid JSON. No commentary.`

type planningResponse struct {
	Goal  string         `json:"goal"`
	Steps []planningStep `json:"steps"`
}

type planningStep struct {
	Type        string `json:"type"`
	Target      string `json:"target"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func (p *LLMPlanner) parsePlanningResponse(raw string) ([]*Step, string, error) {
	cleaned := extractJSONFromMarkdown(raw)

	var pr planningResponse
	if err := json.Unmarshal([]byte(cleaned), &pr); err != nil {
		return nil, "", fmt.Errorf("invalid JSON: %w", err)
	}

	if len(pr.Steps) == 0 {
		return nil, pr.Goal, fmt.Errorf("plan contains no steps")
	}

	steps := make([]*Step, len(pr.Steps))
	for i, ps := range pr.Steps {
		if ps.Type == "" {
			return nil, pr.Goal, fmt.Errorf("step %d has empty type", i)
		}
		steps[i] = &Step{
			Description: ps.Description,
			Action: &Action{
				Type:        ps.Type,
				Target:      ps.Target,
				Value:       ps.Value,
				Description: ps.Description,
			},
			Status: StepPending,
		}
		if i > 0 {
			steps[i].DependsOn = []int{i - 1}
		}
	}
	return steps, pr.Goal, nil
}

func extractJSONFromMarkdown(raw string) string {
	s := strings.TrimSpace(raw)
	if strings.HasPrefix(s, "```") {
		end := strings.Index(s[3:], "```")
		if end >= 0 {
			s = s[3 : end+3]
		}
	}
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		s = s[start : end+1]
	}
	return s
}
