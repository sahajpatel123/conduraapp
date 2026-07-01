package gatekeeper

import (
	_ "embed"
	"fmt"
	"strings"
	"sync/atomic"

	"gopkg.in/yaml.v3"

	"github.com/sahajpatel123/conduraapp/internal/blastradius"
)

//go:embed defaults.yaml
var defaultPolicyYAML []byte

// PolicySchemaError is returned by LoadPolicy when the user-supplied
// YAML violates a hard invariant the engine cannot honor. It is
// exported so the daemon can distinguish schema violations from
// generic parse errors and surface them to the user.
type PolicySchemaError struct {
	// Reason is a stable, programmatic identifier for the violation
	// (e.g. "destructive_downgraded_to_allow"). Suitable for logs,
	// telemetry, and i18n lookup.
	Reason string
	// RuleIndex is the 1-based index of the offending rule in the
	// parsed Rules slice. 0 if the error is not tied to a specific
	// rule.
	RuleIndex int
	// Rule is a short, human-readable summary of the offending rule.
	Rule string
	// Detail is the full sentence shown to the user / log.
	Detail string
}

// Error implements the error interface.
func (e *PolicySchemaError) Error() string {
	return e.Detail
}

// Is supports errors.Is(target, *PolicySchemaError) so callers can
// branch on the underlying reason even when the error has been wrapped
// with %w.
func (e *PolicySchemaError) Is(target error) bool {
	var t *PolicySchemaError
	if !asSchemaError(target, &t) {
		return false
	}
	return e.Reason == t.Reason
}

func asSchemaError(err error, out **PolicySchemaError) bool {
	if err == nil {
		return false
	}
	if pe, ok := err.(*PolicySchemaError); ok {
		*out = pe
		return true
	}
	return false
}

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

// onUserAbsentQueue is the OnUserAbsent policy value meaning "hold the
// action and re-prompt when the user returns." v0.1.x denies for
// safety (the action does not execute); true auto-re-prompt-on-return
// is v0.2.0.
const onUserAbsentQueue = "queue"

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
	// RequireActive: the action must not execute while the user is
	// absent. Carried from ConsentSpec.require_user_active. Also
	// implicitly true for Decision == RequirePresenceAndConsent (the
	// decision's whole point is presence). N1: previously parsed but
	// never read — the dead knob. Now consulted by evaluateConsent.
	RequireActive bool
	// OnUserAbsent: what to do when RequireActive and the user is
	// absent. Carried from ConsentSpec.on_user_absent ("queue" or
	// "deny"). v0.1.x treats both as "deny the action (held for
	// safety)"; true auto-re-prompt-on-return is v0.2.0.
	OnUserAbsent string
}

// Policy is the pure, stateless rules engine. All methods are
// safe for concurrent use.
type Policy struct {
	rules []Rule
}

// LoadPolicy parses YAML into a Policy. Must succeed at startup.
//
// Schema invariants enforced:
//   - No rule may match class == "destructive" with decide == "allow".
//     CLAUDE.md (and formerly MISSION.md) §2.1 invariant #3 requires
//     DESTRUCTIVE actions to always need fresh user consent — a
//     permissive policy downgrades the entire Survival Rule to a
//     soft contract. Violators return *PolicySchemaError.
//   - No rule may match class == "destructive" with decide == "queue".
//     Silently queueing a destructive action is nearly as bad as
//     allowing it: the action still executes without a confirming
//     native modal. Only "deny" and "require_*" decisions are legal
//     for DESTRUCTIVE matches.
func LoadPolicy(data []byte) (*Policy, error) {
	var raw policyYAML
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if err := validatePolicySchema(raw.Rules); err != nil {
		return nil, err
	}
	return &Policy{rules: raw.Rules}, nil
}

// validatePolicySchema walks the rule list and rejects any rule that
// would let a DESTRUCTIVE-class action execute without a fresh consent
// modal. See CLAUDE.md §2.1 invariant #3.
func validatePolicySchema(rules []Rule) error {
	for i, r := range rules {
		classes := strings.Split(strings.ToLower(r.Match.Class), ",")
		for _, c := range classes {
			c = strings.TrimSpace(c)
			// "any" is a wildcard — treat it as covering all classes
			// including DESTRUCTIVE for the purpose of this invariant.
			if c != "destructive" && c != "any" {
				continue
			}
			switch r.Decide {
			case "allow":
				return &PolicySchemaError{
					Reason:    "destructive_downgraded_to_allow",
					RuleIndex: i + 1,
					Rule:      ruleSummary(r),
					Detail: fmt.Sprintf(
						"policy.yaml: rule %d downgrades DESTRUCTIVE to allow — this is forbidden by spec §2.1 invariant #3 (DESTRUCTIVE always requires fresh consent)",
						i+1),
				}
			case "queue":
				return &PolicySchemaError{
					Reason:    "destructive_silently_queued",
					RuleIndex: i + 1,
					Rule:      ruleSummary(r),
					Detail: fmt.Sprintf(
						"policy.yaml: rule %d silently queues DESTRUCTIVE — forbidden by spec §2.1 invariant #3 (DESTRUCTIVE must require consent, not queue)",
						i+1),
				}
			}
		}
	}
	return nil
}

// ruleSummary renders a one-line description of a rule for error
// messages. Best-effort: we never fail to surface an error just
// because the summary is awkward.
func ruleSummary(r Rule) string {
	class := r.Match.Class
	if class == "" {
		class = "<any>"
	}
	return fmt.Sprintf("match=%s decide=%s", class, r.Decide)
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
				RequireActive: r.Consent.RequireActive,
				OnUserAbsent:  r.Consent.OnUserAbsent,
			}
		case "require_presence_and_consent":
			timeout := r.Consent.TimeoutSeconds
			if timeout <= 0 {
				timeout = 300
			}
			absent := r.Consent.OnUserAbsent
			if absent == "" {
				absent = onUserAbsentQueue
			}
			return Verdict{
				Decision:      RequirePresenceAndConsent,
				Reason:        "requires user presence and consent",
				RuleID:        ruleID,
				RequiresModal: true,
				TimeoutSecs:   timeout,
				OnTimeout:     r.Consent.OnTimeout,
				// The decision itself implies presence; RequireActive
				// is true regardless of the ConsentSpec flag.
				RequireActive: true,
				OnUserAbsent:  absent,
			}
		}
	}

	// Default-deny: no rule matched. Safest posture: require presence
	// and consent, queue-on-absent.
	return Verdict{Decision: RequirePresenceAndConsent, Reason: "default-deny: no policy rule matched", RuleID: "default-deny", RequireActive: true, OnUserAbsent: onUserAbsentQueue}
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
	// Target app match. Supports comma-separated list.
	if len(r.Match.TargetApp) > 0 {
		if a.TargetApp == "" {
			return false
		}
		matched := false
		for _, ta := range r.Match.TargetApp {
			if strings.EqualFold(a.TargetApp, ta) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	// Target URL match. Comma-separated list of substrings.
	if r.Match.TargetURL != "" {
		if a.TargetURL == "" {
			return false
		}
		lower := strings.ToLower(a.TargetURL)
		matched := false
		for _, p := range strings.Split(r.Match.TargetURL, ",") {
			if strings.Contains(lower, strings.TrimSpace(p)) {
				matched = true
				break
			}
		}
		if !matched {
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
