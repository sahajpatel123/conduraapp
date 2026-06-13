package gatekeeper

import (
	_ "embed"
	"strings"
	"sync/atomic"

	"gopkg.in/yaml.v3"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
)

//go:embed defaults.yaml
var defaultPolicyYAML []byte

// Rule defines a single policy rule.
type Rule struct {
	Match   MatchSpec   `yaml:"match"`
	Decide  string      `yaml:"decide"`
	Consent ConsentSpec `yaml:"consent,omitempty"`
}

// MatchSpec is what a rule matches against.
type MatchSpec struct {
	Class      string   `yaml:"class"`
	TargetApp  []string `yaml:"target_app,omitempty"`
	TargetURL  string   `yaml:"target_url,omitempty"`
	ActionKind string   `yaml:"action_kind,omitempty"`
}

// ConsentSpec controls consent behavior.
type ConsentSpec struct {
	Type           string `yaml:"type"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
	OnTimeout      string `yaml:"on_timeout"`
	RequireActive  bool   `yaml:"require_user_active"`
	OnUserAbsent   string `yaml:"on_user_absent"`
}

// policyYAML is the top-level structure.
type policyYAML struct {
	Version string `yaml:"version"`
	Rules   []Rule `yaml:"rules"`
}

// Verdict is the enriched decision from Policy evaluation. Richer than
// the terminal Decision — includes context for the consent runtime.
// Never crosses the gatekeeper.Gatekeeper interface boundary.
type Verdict struct {
	Decision      Decision
	Reason        string
	RuleID        string
	RequiresModal bool
	TimeoutSecs   int
	OnTimeout     string
}

// Policy is the pure, stateless rules engine. All methods are
// safe for concurrent use.
type Policy struct {
	rules []Rule
}

// LoadPolicy parses YAML into a Policy. Must succeed at startup.
func LoadPolicy(data []byte) (*Policy, error) {
	var raw policyYAML
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return &Policy{rules: raw.Rules}, nil
}

// DefaultPolicy returns the embedded defaults.
func DefaultPolicy() *Policy {
	p, _ := LoadPolicy(defaultPolicyYAML)
	return p
}

// Evaluate runs a proposed action through the policy rules.
// Returns the rich Verdict.
func (p *Policy) Evaluate(a blastradius.Action) Verdict {
	class := blastradius.Classify(a)

	// Walk rules in priority order.
	for i := range p.rules {
		r := &p.rules[i]
		if !r.matches(class, a) {
			continue
		}
		ruleID := class.String() + "-" + r.Decide

		switch r.Decide {
		case "allow":
			return Verdict{Decision: Allow, Reason: "policy allow rule", RuleID: ruleID}
		case "deny":
			return Verdict{Decision: Deny, Reason: "policy deny rule", RuleID: ruleID}
		case "require_consent":
			timeout := r.Consent.TimeoutSeconds
			if timeout <= 0 {
				timeout = 300
			}
			return Verdict{
				Decision:      RequireConsent,
				Reason:        "requires user consent",
				RuleID:        ruleID,
				RequiresModal: true,
				TimeoutSecs:   timeout,
				OnTimeout:     r.Consent.OnTimeout,
			}
		case "require_presence_and_consent":
			timeout := r.Consent.TimeoutSeconds
			if timeout <= 0 {
				timeout = 300
			}
			return Verdict{
				Decision:      RequirePresenceAndConsent,
				Reason:        "requires user presence and consent",
				RuleID:        ruleID,
				RequiresModal: true,
				TimeoutSecs:   timeout,
				OnTimeout:     r.Consent.OnTimeout,
			}
		}
	}

	// Default-deny: no rule matched.
	return Verdict{Decision: RequirePresenceAndConsent, Reason: "default-deny: no policy rule matched", RuleID: "default-deny"}
}

func (r *Rule) matches(class blastradius.Class, a blastradius.Action) bool {
	// Class match.
	if r.Match.Class != "" {
		if !matchClass(r.Match.Class, class) {
			return false
		}
	}
	// Kind match.
	if r.Match.ActionKind != "" {
		if !strings.EqualFold(a.Kind, r.Match.ActionKind) {
			return false
		}
	}
	return true
}

func matchClass(pattern string, class blastradius.Class) bool {
	c := strings.ToLower(class.String())
	for _, p := range strings.Split(strings.ToLower(pattern), ",") {
		p = strings.TrimSpace(p)
		if p == c || p == "any" {
			return true
		}
	}
	return false
}

// AtomicPolicy provides atomic reads and swaps of a *Policy.
// Safe for hot-reload without locking.
type AtomicPolicy struct {
	p atomic.Pointer[Policy]
}

// Load returns the current policy atomically.
func (ap *AtomicPolicy) Load() *Policy { return ap.p.Load() }

// Store swaps the policy atomically.
func (ap *AtomicPolicy) Store(p *Policy) { ap.p.Store(p) }
