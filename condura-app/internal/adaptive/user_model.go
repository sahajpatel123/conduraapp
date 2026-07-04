// Package adaptive implements the User-Adaptive Engine — the crown
// jewel per MISSION S9. It observes user interactions, runs a dialectic
// (proposer + critic + adjudicator) to extract insights, builds a user
// model with provenance and decay, and predicts next actions.
//
// All user model data is encrypted at rest via storage.DB (hard-invariant
// per MISSION S2.4: "Memory, skills, audit log — all on disk, encrypted").
package adaptive

import (
	"encoding/json"
	"time"
)

// Strength controls engine aggressiveness (off|cautious|balanced|aggressive). (off|cautious|balanced|aggressive).
type Strength string

const (
	// StrengthOff disables the adaptive engine entirely.
	StrengthOff Strength = "off"
	// StrengthCautious observes and learns but does not auto-apply.
	StrengthCautious Strength = "cautious"
	// StrengthBalanced observes, learns, and applies autoApply categories.
	StrengthBalanced Strength = "balanced"
	// StrengthAggressive applies all categories including suggested ones.
	StrengthAggressive Strength = "aggressive"
)

const defaultMinConfidence = 0.6

// InferredField represents a single inferred property about the user.
// Every field carries provenance (evidence chain) and confidence, so
// the engine can decay stale inferences and the user can audit them.
type InferredField struct {
	Value      string    `json:"value"`
	Confidence float64   `json:"confidence"` // 0.0–1.0
	Evidence   []string  `json:"evidence"`   // session IDs that contributed
	LastSeen   time.Time `json:"last_seen"`
	Source     string    `json:"source"` // "observer", "dialectic", "explicit"
}

// UserModel is the structured profile of a user per MISSION S9.2.
// All fields use InferredField for provenance tracking.
type UserModel struct {
	Identity      InferredField     `json:"identity"`
	Preferences   []InferredField   `json:"preferences"`
	Style         InferredField     `json:"style"`
	Expertise     []InferredField   `json:"expertise"`
	PetPeeves     []InferredField   `json:"pet_peeves"`
	Communication InferredField     `json:"communication"`
	RiskTolerance InferredField     `json:"risk_tolerance"`
	TimePatterns  []TimePattern     `json:"time_patterns"`
	Workflows     []WorkflowPattern `json:"workflows"`
	ToolsHabits   map[string]int    `json:"tools_habits"`
	ModelPrefs    map[string]string `json:"model_prefs"`
	LastUpdated   time.Time         `json:"last_updated"`
	Version       int               `json:"version"`
}

// TimePattern captures a recurring user behavior at a specific time.
type TimePattern struct {
	Hour     int      `json:"hour"`
	Weekday  int      `json:"weekday"` // 0=Sunday
	Action   string   `json:"action"`
	Count    int      `json:"count"`
	Evidence []string `json:"evidence"`
}

// WorkflowPattern captures a learned sequence of actions.
type WorkflowPattern struct {
	Name     string   `json:"name"`
	Steps    []string `json:"steps"`
	Count    int      `json:"count"`
	Evidence []string `json:"evidence"`
}

// Observation is a single session event fed to the engine.
type Observation struct {
	SessionID     string    `json:"session_id"`
	UserQuery     string    `json:"user_query"`
	AgentReply    string    `json:"agent_reply"`
	Duration      float64   `json:"duration_sec"`
	ToolsUsed     []string  `json:"tools_used"`
	FinishReason  string    `json:"finish_reason"`
	UserInitiated bool      `json:"user_initiated"`
	Timestamp     time.Time `json:"timestamp"`
}

// Config holds the engine configuration per MISSION S9.6.
type Config struct {
	Enabled         bool     `json:"enabled"`
	Strength        Strength `json:"strength"`
	AutoApply       []string `json:"auto_apply"`
	RequireConfirm  []string `json:"require_confirm"`
	ForgetAfterDays int      `json:"forget_after_days"`

	DialecticPrimaryModel  string  `json:"primary_model"`
	DialecticCriticModel   string  `json:"critic_model"`
	DialecticMinConfidence float64 `json:"min_confidence"`

	SpendMonitor BudgetChecker `json:"-"`
}

// BudgetChecker is the subset of failover.SpendMonitor we need.
type BudgetChecker interface {
	CheckBudget() error
}

// DefaultConfig returns the safe defaults per MISSION S9.6.
func DefaultConfig() Config {
	return Config{
		Enabled:                true,
		Strength:               StrengthCautious,
		AutoApply:              []string{"verbosity", "response_length", "default_model", "time_patterns"},
		RequireConfirm:         []string{"new_skill", "default_backend", "communication_style", "risk_tolerance"},
		ForgetAfterDays:        30,
		DialecticMinConfidence: defaultMinConfidence,
	}
}

// Store is the interface for adaptive engine persistence.
type Store interface {
	Load() (*UserModel, error)
	Save(model *UserModel) error
	Reset() error
	Close() error
}

// Marshal serializes the user model for encrypted storage.
func (m *UserModel) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalUserModel deserializes from encrypted storage.
func UnmarshalUserModel(data []byte) (*UserModel, error) {
	var m UserModel
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
