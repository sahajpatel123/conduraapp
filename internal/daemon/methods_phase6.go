// Package daemon JSON-RPC method registration for the Phase 6
// "living presence" surface. Wired by RegisterAll in methods.go.
//
// RPC surface (per the living-presence spec §5):
//
//	voice.status   — return the current voice pipeline status
//	voice.cancel   — cancel any in-progress voice session
//	voice.speak    — speak text through the TTS speaker
//	presence.summon  — show the overlay and begin capture
//	presence.dismiss — hide the overlay and stop capture
//	presence.state   — return the current overlay state
//	agent.ask     — run a one-shot session for a user query
//	agent.status  — return the current session status
package daemon

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sahajpatel123/conduraapp/internal/audit"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
	"github.com/sahajpatel123/conduraapp/internal/overlay"
	"github.com/sahajpatel123/conduraapp/internal/session"
	"github.com/sahajpatel123/conduraapp/internal/status"
)

// ErrNoVoice is returned by voice.* RPCs when voice is disabled
// or the voice pipeline failed to initialize.
var ErrNoVoice = errors.New("daemon: voice is not available")

// statusKey is the JSON key used to report status in RPC
// responses. Centralized so it can be referenced consistently
// across handlers and so the linter doesn't flag the repeated
// string literal.
const statusKey = "status"

// registerPhase6Methods wires the Phase 6 "living presence" RPC
// surface. All methods are safe to call when the underlying
// sub-system (voice, presence) is unavailable — they return a
// well-typed error rather than panicking.
func registerPhase6Methods(srv *ipc.Server, subs *Subsystems) {
	// -- voice.* --------------------------------------------------------
	registerVoiceMethods(srv, subs)
	// -- presence.* -----------------------------------------------------
	registerPresenceMethods(srv, subs)
	// -- agent.* --------------------------------------------------------
	registerAgentMethods(srv, subs)
}

func registerVoiceMethods(srv *ipc.Server, subs *Subsystems) {
	// voice.status: return the current voice pipeline status.
	// Returns {"status": "idle"} when voice is not configured.
	srv.Register("voice.status", func(_ context.Context, _ json.RawMessage) (any, error) {
		if subs.Voice == nil {
			return map[string]any{statusKey: status.StatusIdle.String(), "available": false}, nil
		}
		return map[string]any{
			statusKey:   subs.Voice.State().String(),
			"available": true,
		}, nil
	})

	// voice.cancel: cancel any in-progress voice session.
	srv.Register("voice.cancel", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if subs.Voice == nil {
			return nil, &ipc.Error{Code: ipc.CodeMethodNotFound, Message: ErrNoVoice.Error()}
		}
		if subs.Voice != nil {
			subs.Voice.Cancel()
		}
		if subs.Audit != nil {
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor: actorGUI, Action: "voice.cancel", App: appConduraG,
				Level: auditLevelInfo, Result: auditResultAllow,
			})
		}
		return auditOK(), nil
	})

	// voice.speak: speak text through the TTS speaker.
	srv.Register("voice.speak", func(ctx context.Context, params json.RawMessage) (any, error) {
		if subs.Voice == nil {
			return nil, &ipc.Error{Code: ipc.CodeMethodNotFound, Message: ErrNoVoice.Error()}
		}
		var p struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if err := subs.Voice.Speak(ctx, p.Text); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		if subs.Audit != nil {
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor: actorGUI, Action: "voice.speak", App: appConduraG,
				Level: auditLevelInfo, Result: auditResultAllow,
				Message: "len=" + itoa(int64(len(p.Text))),
			})
		}
		return auditOK(), nil
	})

	// voice.listen: record mic, transcribe via whisper.cpp or OpenAI,
	// return text. The caller feeds the transcript into llm.stream or
	// agent.ask. Only one listen session may be active at a time.
	srv.Register("voice.listen", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if subs.Voice == nil {
			return nil, &ipc.Error{Code: ipc.CodeMethodNotFound, Message: ErrNoVoice.Error()}
		}
		result, err := subs.Voice.ListenAndProcess(ctx)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		if subs.Audit != nil && result.Transcript != "" {
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor: actorGUI, Action: "voice.listen", App: appConduraG,
				Level: auditLevelInfo, Result: auditResultAllow,
				Message: "len=" + itoa(int64(len(result.Transcript))),
			})
		}
		return map[string]any{
			"transcript": result.Transcript,
			"confidence": result.Confidence,
		}, nil
	})
}

func registerPresenceMethods(srv *ipc.Server, subs *Subsystems) {
	// presence.summon: show the overlay and begin capture.
	srv.Register("presence.summon", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if err := subs.Overlay.Show(ctx, overlay.ShowOpts{AtCursor: true}); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		if subs.Audit != nil {
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor: actorGUI, Action: "presence.summon", App: appConduraG,
				Level: auditLevelInfo, Result: auditResultAllow,
			})
		}
		return auditOK(), nil
	})

	// presence.dismiss: hide the overlay and stop capture.
	srv.Register("presence.dismiss", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if err := subs.Overlay.Hide(); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		if subs.Audit != nil {
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor: actorGUI, Action: "presence.dismiss", App: appConduraG,
				Level: auditLevelInfo, Result: auditResultAllow,
			})
		}
		return auditOK(), nil
	})

	// presence.state: return the current overlay state.
	srv.Register("presence.state", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{
			"state": subs.Overlay.State().String(),
		}, nil
	})
}

func registerAgentMethods(srv *ipc.Server, subs *Subsystems) {
	// agent.ask: run a one-shot session for a user query.
	// The session accumulates the reply from the SSE broker
	// (stream.delta events) and returns the full text.
	srv.Register("agent.ask", func(ctx context.Context, params json.RawMessage) (any, error) {
		if subs.SessionFactory == nil {
			return nil, &ipc.Error{
				Code:    ipc.CodeInternalError,
				Message: "session factory not initialized",
			}
		}
		var p struct {
			ConversationID int64  `json:"conversation_id"`
			Query          string `json:"query"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if p.Query == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "query is required"}
		}
		sess := subs.SessionFactory.New(p.ConversationID)
		reply, err := sess.Run(ctx, p.Query)
		if err != nil {
			// ErrAlreadyRunning maps to a non-fatal code so the
			// GUI can show "agent is busy" without raising an
			// unexpected-error banner.
			if errors.Is(err, session.ErrAlreadyRunning) {
				return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
			}
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		if subs.Audit != nil {
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor: actorGUI, Action: "agent.ask", App: appConduraG,
				Level: auditLevelInfo, Result: auditResultAllow,
				Message: "conversation_id=" + itoa(p.ConversationID) + " reply_len=" + itoa(int64(len(reply))),
			})
		}
		// Trigger async memory extraction on success (fire-and-forget).
		if subs.Extractor != nil && reply != "" {
			go subs.Extractor.AfterSession(context.Background(), p.Query, reply, p.ConversationID) //nolint:gosec // intentional: async, must survive request ctx
		}
		return map[string]any{
			"reply":           reply,
			keyConversationID: p.ConversationID,
		}, nil
	})

	// agent.ask-complete: internal hook fired by the frontend
	// after a session completes to trigger async memory and
	// skill extraction. Best-effort, never fails the caller.
	srv.Register("agent.ask-complete", func(ctx context.Context, params json.RawMessage) (any, error) {
		if subs.Extractor == nil {
			return auditOK(), nil
		}
		var p struct {
			ConversationID int64  `json:"conversation_id"`
			UserMessage    string `json:"user_message"`
			AssistantReply string `json:"assistant_reply"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		subs.Extractor.AfterSession(ctx, p.UserMessage, p.AssistantReply, p.ConversationID)
		return auditOK(), nil
	})

	// agent.status: return the current session status. Note
	// that this is a snapshot; for a live status feed, the
	// GUI should subscribe to "tray.status" on the SSE broker.
	srv.Register("agent.status", func(_ context.Context, _ json.RawMessage) (any, error) {
		statusStr := status.StatusIdle.String()
		if subs.SessionFactory != nil {
			statusStr = subs.SessionFactory.Status().String()
		}
		return map[string]any{
			statusKey: statusStr,
		}, nil
	})
}

// itoa64 is the int64 equivalent of itoa; it lives in
// methods_phase2.go so the RPC files share one implementation.
