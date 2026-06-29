package adaptive

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sahajpatel123/conduraapp/internal/llm"
)

// Dialectic runs the proposer-critic-adjudicator pipeline on
// observed sessions. The proposer extracts insights about the user;
// the critic challenges them; the adjudicator decides the action.
//
// Gated by strength: off → no analysis; cautious → observe+learn
// only (no auto-apply). Both LLM calls route through the spend
// monitor when available.
type Dialectic struct {
	primary      llm.Provider
	critic       llm.Provider
	primaryModel string
	criticModel  string
	adjudicator  *Adjudicator
	budget       BudgetChecker
	strength     Strength
}

// NewDialectic creates a dialectic engine. critic may be nil (falls
// back to proposer-only with uncriticized proposals).
func NewDialectic(primary llm.Provider, primaryModel string, critic llm.Provider, criticModel string, adj *Adjudicator, budget BudgetChecker, strength Strength) *Dialectic {
	return &Dialectic{
		primary:      primary,
		critic:       critic,
		primaryModel: primaryModel,
		criticModel:  criticModel,
		adjudicator:  adj,
		budget:       budget,
		strength:     strength,
	}
}

// Analyze runs the dialectic on observations and returns proposals.
// Async and best-effort — errors are logged, not returned.
func (d *Dialectic) Analyze(ctx context.Context, observations []Observation) ([]Proposal, error) {
	if d.strength == StrengthOff || len(observations) == 0 {
		return nil, nil
	}
	if d.budget != nil {
		if err := d.budget.CheckBudget(); err != nil {
			return nil, fmt.Errorf("dialectic: spend limit: %w", err)
		}
	}
	proposed, err := d.propose(ctx, observations)
	if err != nil {
		return nil, err
	}
	validated, err := d.criticize(ctx, observations, proposed)
	if err != nil {
		return proposed, nil //nolint:nilerr
	}
	return validated, nil
}

func (d *Dialectic) propose(ctx context.Context, obs []Observation) ([]Proposal, error) {
	prompt := buildProposerPrompt(obs)
	resp, err := d.primary.Chat(ctx, llm.ChatRequest{
		Model: d.primaryModel,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: proposerSysPrompt},
			{Role: llm.RoleUser, Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}
	return parseProposals(resp.Message.Content), nil
}

func (d *Dialectic) criticize(ctx context.Context, _ []Observation, proposals []Proposal) ([]Proposal, error) { //nolint:unparam
	if d.critic == nil {
		return proposals, nil
	}
	prompt := buildCriticPrompt(nil, proposals)
	resp, err := d.critic.Chat(ctx, llm.ChatRequest{
		Model: d.criticModel,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: criticSysPrompt},
			{Role: llm.RoleUser, Content: prompt},
		},
	})
	if err != nil {
		return proposals, nil //nolint:nilerr
	}
	return parseProposals(resp.Message.Content), nil
}

func buildProposerPrompt(obs []Observation) string {
	s := "Based on these user interactions, what can we infer?\n\n"
	for _, o := range obs { //nolint:gocritic
		s += fmt.Sprintf("Q: %s\nA: %s\n\n", o.UserQuery, truncate(o.AgentReply, 200))
	}
	s += "Return JSON: [{\"category\":\"...\",\"field\":\"...\",\"value\":\"...\",\"confidence\":0.0-1.0,\"reason\":\"...\"}]"
	return s
}

func buildCriticPrompt(_ []Observation, proposals []Proposal) string {
	var s string
	s += "Review these inferences. Lower confidence if over-fitting.\n\n"
	for _, p := range proposals {
		s += fmt.Sprintf("- %s: %s (%.2f, %s)\n", p.Category, p.Value, p.Confidence, p.Reason)
	}
	s += "\nReturn revised JSON array with adjusted confidence."
	return s
}

const proposerSysPrompt = `You are a user-modeling assistant. Infer structured preferences. Categories: verbosity, response_length, default_model, time_patterns, communication_style, risk_tolerance, default_backend, new_skill. Return ONLY valid JSON array.`

const criticSysPrompt = `You are a critical reviewer. Check for over-fitting or weak evidence. Adjust confidence scores. Return ONLY valid JSON array.`

func parseProposals(raw string) []Proposal {
	if raw == "" {
		return nil
	}
	cleaned := extractJSONBlock(raw)
	var proposals []Proposal
	if err := json.Unmarshal([]byte(cleaned), &proposals); err != nil {
		return nil
	}
	for i := range proposals {
		if proposals[i].Confidence <= 0 {
			proposals[i].Confidence = 0.5
		}
	}
	return proposals
}

func extractJSONBlock(raw string) string {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "```"); i >= 0 {
		if j := strings.Index(s[i+3:], "```"); j >= 0 {
			s = s[i+3 : i+3+j]
		}
	}
	if i := strings.Index(s, "["); i >= 0 {
		if j := strings.LastIndex(s, "]"); j > i {
			return s[i : j+1]
		}
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
