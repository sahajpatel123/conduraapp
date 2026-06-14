package daemon

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sahajpatel123/synapticapp/internal/ipc"
	"github.com/sahajpatel123/synapticapp/internal/replay"
)

// registerPhase11Methods wires the action-replay RPC methods
// (Phase 11, sub-phase 11A): replay.timeline, replay.frame,
// replay.verify_integrity. These are read-only over the
// HMAC-chained audit log. The user can scrub a 24h window of
// what the agent did, with the screenshots that bracket each
// action and the gatekeeper verdict.
//
// None of these methods are gated through the Gatekeeper —
// they are observability, not action.
func registerPhase11Methods(srv *ipc.Server, subs *Subsystems) {
	srv.Register("replay.timeline", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if subs.Replay == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "replay subsystem not available"}
		}
		frames, err := subs.Replay.Timeline(ctx, zeroTime())
		if err != nil {
			return nil, fmt.Errorf("replay: timeline: %w", err)
		}
		return serializeFrames(frames), nil
	})

	srv.Register("replay.frame", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if subs.Replay == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "replay subsystem not available"}
		}
		frame, err := subs.Replay.FrameByID(ctx, p.ID)
		if err != nil {
			if err == replay.ErrFrameNotFound {
				return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "no frame with that id"}
			}
			return nil, err
		}
		return serializeFrame(*frame), nil
	})

	srv.Register("replay.verify_integrity", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if subs.Replay == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "replay subsystem not available"}
		}
		report, err := subs.Replay.VerifyIntegrity(ctx)
		if err != nil {
			return nil, err
		}
		return report, nil
	})
}

// serializeFrame turns a Frame into a JSON-safe shape. The
// raw screenshot bytes are base64-encoded so the GUI can
// inline them in <img> tags without a separate fetch.
type replayFrameJSON struct {
	ID                   int64  `json:"id"`
	Action               string `json:"action"`
	App                  string `json:"app"`
	Actor                string `json:"actor"`
	Result               string `json:"result"`
	Level                string `json:"level"`
	Message              string `json:"message"`
	Timestamp            string `json:"timestamp"`
	Outcome              string `json:"outcome"`
	OutcomeReason        string `json:"outcome_reason,omitempty"`
	BeforeScreenshot     string `json:"before_screenshot,omitempty"` // base64
	AfterScreenshot      string `json:"after_screenshot,omitempty"`
	BeforeScreenshotMime string `json:"before_screenshot_mime,omitempty"`
	AfterScreenshotMime  string `json:"after_screenshot_mime,omitempty"`
}

func serializeFrame(f replay.Frame) replayFrameJSON {
	out := replayFrameJSON{
		Outcome:       string(f.Outcome),
		OutcomeReason: f.OutcomeReason,
	}
	if f.Event != nil {
		out.ID = f.Event.ID
		out.Action = f.Event.Action
		out.App = f.Event.App
		out.Actor = f.Event.Actor
		out.Result = f.Event.Result
		out.Level = f.Event.Level
		out.Message = f.Event.Message
		out.Timestamp = f.Event.TS.Format("2006-01-02T15:04:05.000Z07:00")
	}
	if len(f.BeforeScreenshot) > 0 {
		out.BeforeScreenshot = base64Encode(f.BeforeScreenshot)
		out.BeforeScreenshotMime = "image/png"
	}
	if len(f.AfterScreenshot) > 0 {
		out.AfterScreenshot = base64Encode(f.AfterScreenshot)
		out.AfterScreenshotMime = "image/png"
	}
	return out
}

func serializeFrames(frames []replay.Frame) []replayFrameJSON {
	out := make([]replayFrameJSON, 0, len(frames))
	for _, f := range frames {
		out = append(out, serializeFrame(f))
	}
	return out
}
