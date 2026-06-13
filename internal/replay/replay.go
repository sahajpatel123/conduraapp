// Package replay reconstructs the 24h scrubbable Action Replay timeline
// from the audit log (Phase 11, sub-phase 11A).
//
// Source of truth: the audit log. Replay never duplicates or modifies
// the audit log; it reads it, verifies the HMAC chain, and exposes a
// query API for the GUI's timeline view. Screenshots referenced by
// audit events live in a separate encrypted on-disk ring buffer
// (screenshots.go).
//
// What "Replay" is — and is not (per MISSION §18):
//   - YES: a scrubbable 24h timeline of what the agent did, with the
//     screenshots that bracket each action and the gatekeeper verdict.
//   - NO: a time machine. Irreversible OS actions (sent email, rm -rf)
//     are not undoable from Replay. The user reads Replay to understand
//     what happened, not to reverse it.
package replay

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/audit"
)

// ErrFrameNotFound is returned by FrameByID when no audit event
// exists for the given id.
var ErrFrameNotFound = errors.New("replay: frame not found")

// Replay is the read-only timeline API.
type Replay struct {
	audit *audit.Log
	shots *ScreenshotStore
	ttl   time.Duration
}

// Options configures a Replay.
type Options struct {
	// Audit is the audit log to read from. Required.
	Audit *audit.Log
	// Screenshots is the on-disk screenshot store. If nil, the
	// Frame.ReferencedScreenshots are always nil (the timeline still
	// works, just without image refs).
	Screenshots *ScreenshotStore
	// TTL is the maximum age of frames returned by Timeline. The
	// default is 24h (MISSION §18.1).
	TTL time.Duration
}

// New returns a Replay.
func New(opts Options) (*Replay, error) {
	if opts.Audit == nil {
		return nil, fmt.Errorf("replay: Audit is required")
	}
	ttl := opts.TTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &Replay{
		audit: opts.Audit,
		shots: opts.Screenshots,
		ttl:   ttl,
	}, nil
}

// Frame is one reconstructed action in the timeline. It pairs an
// audit Event with the screenshots it referenced and a derived
// outcome (success, denied, or unknown).
type Frame struct {
	Event            *audit.Event
	BeforeScreenshot []byte // nil if unavailable
	AfterScreenshot  []byte // nil if unavailable
	Outcome          Outcome
	OutcomeReason    string
}

// Outcome classifies what the audit event recorded.
type Outcome string

// Outcome enum values. classifyOutcome maps the audit log's
// Result field vocabulary into one of these.
const (
	// OutcomeAllowed is the success path: the action ran without error.
	OutcomeAllowed Outcome = "allowed"
	// OutcomeDenied is the gatekeeper-deny or consent-deny path.
	OutcomeDenied Outcome = "denied"
	// OutcomeErrored is the action-attempted-but-failed path.
	OutcomeErrored Outcome = "errored"
	// OutcomeUnknown is the catch-all for unrecognized Result values.
	OutcomeUnknown Outcome = "unknown"
)

// Timeline returns the frames in the window [now-ttl, now], ordered
// oldest-first (chronological). Frames outside the window are pruned.
// If a frame references a screenshot ref that the ScreenshotStore
// cannot resolve, the screenshot bytes are nil but the frame is still
// returned.
func (r *Replay) Timeline(ctx context.Context, now time.Time) ([]Frame, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	since := now.Add(-r.ttl)
	evs, err := r.audit.List(ctx, audit.Query{
		Since: since,
		Limit: 1000, // 24h is unlikely to exceed 1000 audit events in v0.1.0
	})
	if err != nil {
		return nil, fmt.Errorf("replay: list audit: %w", err)
	}
	// The audit log returns ts DESC; we want chronological.
	frames := make([]Frame, 0, len(evs))
	for i := len(evs) - 1; i >= 0; i-- {
		ev := evs[i]
		frames = append(frames, r.buildFrame(ctx, &ev))
	}
	return frames, nil
}

// FrameByID returns a single frame by audit event id.
func (r *Replay) FrameByID(ctx context.Context, id int64) (*Frame, error) {
	ev, err := r.audit.GetByID(ctx, id)
	if err != nil {
		// audit.ErrEventNotFound is the source-of-truth sentinel
		// for "this id is not in the log". We translate to
		// replay.ErrFrameNotFound so callers can compare with a
		// single error regardless of the underlying layer.
		if errors.Is(err, audit.ErrEventNotFound) {
			return nil, ErrFrameNotFound
		}
		return nil, fmt.Errorf("replay: get event: %w", err)
	}
	if ev == nil {
		return nil, ErrFrameNotFound
	}
	f := r.buildFrame(ctx, ev)
	return &f, nil
}

// VerifyIntegrity runs the audit log's HMAC chain verifier and returns
// the report. Replay surfaces tampering to the GUI; it does not hide
// it (MISSION §5.4: "every action is auditable, in a tamper-resistant
// log").
func (r *Replay) VerifyIntegrity(ctx context.Context) (*audit.ChainReport, error) {
	return r.audit.VerifyChain(ctx, 0, 0)
}

func (r *Replay) buildFrame(ctx context.Context, ev *audit.Event) Frame {
	f := Frame{
		Event:         ev,
		Outcome:       classifyOutcome(ev),
		OutcomeReason: ev.Message,
	}
	if r.shots != nil {
		if ev.SSBeforeRef != "" {
			if b, err := r.shots.Get(ctx, ev.SSBeforeRef); err == nil {
				f.BeforeScreenshot = b
			}
		}
		if ev.SSAfterRef != "" {
			if b, err := r.shots.Get(ctx, ev.SSAfterRef); err == nil {
				f.AfterScreenshot = b
			}
		}
	}
	return f
}

// classifyOutcome turns an audit event's Result field into an
// outcome enum. We accept the audit log's vocabulary (allow, deny,
// block, error) and map it to the replay vocabulary.
func classifyOutcome(ev *audit.Event) Outcome {
	switch ev.Result {
	case "allow", "approved", "success":
		return OutcomeAllowed
	case "deny", "denied", "block", "blocked":
		return OutcomeDenied
	case "error", "errored", "fail", "failed":
		return OutcomeErrored
	default:
		return OutcomeUnknown
	}
}
