package adaptive

// Adjudicator evaluates dialectic proposals and decides whether to
// auto-apply, require user confirmation, or discard. It implements
// MISSION S9.6: category lists that gate the consent boundary.
//
// Categories in autoApply (per S9.6): verbosity, response_length,
// default_model, time_patterns — these are applied silently.
//
// Categories in requireConfirm: new_skill, default_backend,
// communication_style, risk_tolerance — these surface to the user
// for explicit confirmation.
type Adjudicator struct {
	autoApply      map[string]bool
	requireConfirm map[string]bool
	minConfidence  float64
}

// Proposal is the output of the Dialectic (proposer + critic).
type Proposal struct {
	Category   string  `json:"category"`
	Field      string  `json:"field"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// Decision is the adjudicator's verdict on a proposal.
type Decision string

const (
	// DecisionAutoApply means the proposal is applied silently.
	// DecisionAutoApply means the proposal is applied silently.
	DecisionAutoApply Decision = "auto_apply"
	// DecisionRequireConfirm surfaces the proposal for user confirmation.
	DecisionRequireConfirm Decision = "require_confirm"
	// DecisionDiscard means the proposal is dropped.
	DecisionDiscard Decision = "discard"
)

// NewAdjudicator creates an adjudicator with the given category lists.
func NewAdjudicator(autoApply, requireConfirm []string, minConfidence float64) *Adjudicator {
	if minConfidence <= 0 {
		minConfidence = 0.6
	}
	aa := make(map[string]bool, len(autoApply))
	for _, c := range autoApply {
		aa[c] = true
	}
	rc := make(map[string]bool, len(requireConfirm))
	for _, c := range requireConfirm {
		rc[c] = true
	}
	return &Adjudicator{
		autoApply:      aa,
		requireConfirm: rc,
		minConfidence:  minConfidence,
	}
}

// Evaluate decides the action for a proposal.
func (a *Adjudicator) Evaluate(p Proposal) Decision {
	// Confidence too low → discard.
	if p.Confidence < a.minConfidence {
		return DecisionDiscard
	}

	// Explicit require-confirm → surface to user.
	if a.requireConfirm[p.Category] {
		return DecisionRequireConfirm
	}

	// Explicit auto-apply → apply silently.
	if a.autoApply[p.Category] {
		return DecisionAutoApply
	}

	// Unknown category → require confirmation (safe default).
	return DecisionRequireConfirm
}

// Filter splits proposals into auto-apply and require-confirm groups.
func (a *Adjudicator) Filter(proposals []Proposal) (auto []Proposal, confirm []Proposal, discarded int) {
	for _, p := range proposals {
		switch a.Evaluate(p) {
		case DecisionAutoApply:
			auto = append(auto, p)
		case DecisionRequireConfirm:
			confirm = append(confirm, p)
		case DecisionDiscard:
			discarded++
		}
	}
	return
}
