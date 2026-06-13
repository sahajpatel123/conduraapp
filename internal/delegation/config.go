// Package delegation implements the Delegation Bus — the sub-agent
// spawning system. Sub-agents (Claude Code, Codex, OpenCode, etc.) run
// tasks and return structured output. They have zero direct FS/network/
// terminal access; physical actions are structured requests the daemon
// gates through Engine.Evaluate and executes on their behalf.
//
// Architecture: leaves-only (v0.1.0). Sub-agents return output only.
// Peer protocol and capability tokens are deferred to v0.2.0.
//
//nolint:revive,mnd,gocritic // OutputFormat constants; budget values; range copies
package delegation

import (
	"errors"
	"time"
)

// OutputFormat specifies the expected output format from a sub-agent CLI.
type OutputFormat string

const (
	FmtJSON       OutputFormat = "json"
	FmtStreamJSON OutputFormat = "stream-json"
	FmtText       OutputFormat = "text"
)

// AgentConfig describes a sub-agent CLI.
type AgentConfig struct {
	Name         string        `json:"name"`
	Command      string        `json:"command"`
	ArgsTemplate []string      `json:"args_template"`
	OutputFormat OutputFormat  `json:"output_format"`
	ModelFlag    string        `json:"model_flag"`
	BinaryProbe  string        `json:"binary_probe"`
	Description  string        `json:"description"`
	MaxDepth     int           `json:"max_depth"`
	Timeout      time.Duration `json:"timeout"`
	BudgetCap    float64       `json:"budget_cap"`
}

// SpawnRequest is the input for a sub-agent task.
type SpawnRequest struct {
	AgentName string  `json:"agent_name"`
	Task      string  `json:"task"`
	Model     string  `json:"model,omitempty"`
	Depth     int     `json:"depth"`
	Budget    float64 `json:"budget"`
	Timeout   time.Duration
}

// SpawnResult is the output of a sub-agent run.
type SpawnResult struct {
	AgentName  string        `json:"agent_name"`
	Task       string        `json:"task"`
	Output     string        `json:"output"`
	ExitCode   int           `json:"exit_code"`
	Duration   time.Duration `json:"duration"`
	TokenCount int           `json:"token_count,omitempty"`
	SpawnID    string        `json:"spawn_id,omitempty"`
}

// ActionRequest is a structured physical-action request from a sub-agent.
// The daemon gates each one through Engine.Evaluate before execution.
type ActionRequest struct {
	AgentName string `json:"agent_name"`
	Kind      string `json:"kind"`
	Command   string `json:"command,omitempty"`
	Path      string `json:"path,omitempty"`
	Body      string `json:"body,omitempty"`
}

// Config holds all agent configurations.
type Config struct {
	Agents       []AgentConfig
	GlobalBudget float64
	GlobalLimit  int // max concurrent agents across all backends
}

// DefaultConfig returns the built-in config for v0.1.0 agents.
func DefaultConfig() Config {
	return Config{
		GlobalBudget: 10.0,
		GlobalLimit:  5,
		Agents:       DefaultAgents(),
	}
}

// DefaultAgents returns the built-in agent configs.
func DefaultAgents() []AgentConfig {
	return []AgentConfig{
		{
			Name: "claude", Command: "claude",
			ArgsTemplate: []string{"--print", "--output-format", "stream-json", "--model"},
			OutputFormat: FmtStreamJSON, ModelFlag: "--model",
			BinaryProbe: "claude", Description: "Anthropic Claude Code CLI",
			MaxDepth: 3, Timeout: 5 * time.Minute, BudgetCap: 2.0,
		},
		{
			Name: "ollama", Command: "",
			OutputFormat: FmtJSON,
			BinaryProbe:  "ollama", Description: "Local Ollama (HTTP, no subprocess)",
			MaxDepth: 1, Timeout: 5 * time.Minute, BudgetCap: 0,
		},
	}
}

// Sentinel errors.
var (
	ErrAgentNotFound  = errors.New("delegation: agent not found")
	ErrRecursionLimit = errors.New("delegation: recursion limit exceeded")
	ErrBudgetExceeded = errors.New("delegation: budget exceeded")
	ErrTimeout        = errors.New("delegation: timeout")
	ErrGatedDeny      = errors.New("delegation: gatekeeper denied spawn")
)

// FindAgent returns the agent config by name.
func (c Config) FindAgent(name string) (AgentConfig, bool) {
	for _, a := range c.Agents {
		if a.Name == name {
			return a, true
		}
	}
	return AgentConfig{}, false
}
