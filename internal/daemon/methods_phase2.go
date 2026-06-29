// Package daemon JSON-RPC method registration for conversations, LLM
// streaming, audit, and kill-switch.
package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/anomaly"
	"github.com/sahajpatel123/conduraapp/internal/audit"
	"github.com/sahajpatel123/conduraapp/internal/conversation"
	"github.com/sahajpatel123/conduraapp/internal/halt"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
	"github.com/sahajpatel123/conduraapp/internal/llm"
	"github.com/sahajpatel123/conduraapp/internal/stream"
	"github.com/sahajpatel123/conduraapp/internal/watchdog"
)

// registerConversationMethods wires conversations.* + llm.stream +
// llm.cancel. The anomaly detector, if non-nil, is Reset() at the
// start of each new conversation so cross-session loop accumulation
// can't trip false positives (Phase 16, Rec 6).
func registerConversationMethods(
	srv *ipc.Server,
	store *conversation.Store,
	auditLog *audit.Log,
	haltFlag *halt.Flag,
	sm *stream.Manager,
	reg *llm.Registry,
	det *anomaly.Detector,
	wdog *watchdog.Watchdog,
) {
	srv.Register("conversations.list", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return store.List(ctx)
	})
	srv.Register("conversations.get", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		c, err := store.Get(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		return c, nil
	})
	srv.Register("conversations.create", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Title string `json:"title"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		m, err := store.Create(ctx, p.Title)
		if err != nil {
			return nil, err
		}
		// Rec 6: each new conversation is a fresh context for
		// anomaly detection. Reset here so a user who opens a new
		// chat after a long session doesn't immediately trip a
		// stale rate/loop threshold.
		if det != nil {
			det.Reset()
		}
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorDaemon, Action: "conversations.create", App: appCondurad,
			Level: auditLevelInfo, Result: auditResultAllow,
			Message: "id=" + itoa(m.ID),
		})
		return m, nil
	})
	srv.Register("conversations.delete", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if err := store.Delete(ctx, p.ID); err != nil {
			return nil, err
		}
		// Cancel any in-flight streams for this conversation
		// so the GUI doesn't keep receiving tokens for a
		// conversation the user just deleted.
		if sm != nil {
			sm.CancelByConversation(p.ID)
		}
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorDaemon, Action: "conversations.delete", App: appCondurad,
			Level: auditLevelInfo, Result: auditResultAllow,
			Message: "id=" + itoa(p.ID),
		})
		return auditOK(), nil
	})
	srv.Register("conversations.append", func(ctx context.Context, params json.RawMessage) (any, error) {
		if haltFlag.IsHalted() {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "daemon is halted"}
		}
		var p struct {
			ID      int64                `json:"id"`
			Role    string               `json:"role"`
			Content string               `json:"content"`
			Message conversation.Message `json:"message"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if p.Message.Role == "" {
			p.Message.Role = p.Role
		}
		if p.Message.Content == "" {
			p.Message.Content = p.Content
		}
		if err := store.Append(ctx, p.ID, p.Message); err != nil {
			return nil, err
		}
		// Phase 16, Rec 2: every user message counts as a watchdog
		// "verification". Without this the watchdog would only
		// reset on GUI-side calls (which the agent loop running
		// unattended would never make).
		if wdog != nil {
			wdog.Touch()
		}
		return auditOK(), nil
	})

	// llm.stream: the GUI calls this to start a stream. The
	// manager kicks off the provider's Stream and returns
	// immediately with a request_id. Tokens arrive on the SSE
	// broker at /events as "stream.*" events tagged with the
	// request_id.
	srv.Register("llm.stream", func(ctx context.Context, params json.RawMessage) (any, error) {
		return handleLLMStream(ctx, params, haltFlag, sm, reg, auditLog)
	})
	srv.Register("llm.cancel", func(ctx context.Context, params json.RawMessage) (any, error) {
		return handleLLMCancel(ctx, params, sm, auditLog)
	})
}

// handleLLMStream is the body of the llm.stream RPC. Extracted to keep
// registerConversationMethods under the gocognit budget.
func handleLLMStream(
	ctx context.Context,
	params json.RawMessage,
	_ *halt.Flag, // halt is checked inside the stream manager
	sm *stream.Manager,
	reg *llm.Registry,
	auditLog *audit.Log,
) (any, error) {
	var p struct {
		Provider       string          `json:"provider"`
		ConversationID int64           `json:"conversation_id"`
		Request        llm.ChatRequest `json:"request"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
	}
	if p.Provider == "" {
		// Convenience: if exactly one provider is configured,
		// accept calls without an explicit provider name.
		if reg != nil && reg.Len() == 1 {
			for _, name := range reg.Names() {
				p.Provider = name
			}
		}
	}
	if p.Provider == "" {
		return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "provider is required"}
	}
	requestID, err := sm.Start(ctx, stream.Request{
		ConversationID: p.ConversationID,
		ProviderName:   p.Provider,
		Chat:           p.Request,
	})
	if err != nil {
		return nil, mapStreamError(err)
	}
	_ = auditLog.Append(ctx, audit.Event{
		Actor: actorGUI, Action: "llm.stream", App: appConduraG,
		Level: auditLevelInfo, Result: auditResultAllow,
		Message: "provider=" + p.Provider + " model=" + p.Request.Model,
	})
	return map[string]any{
		"request_id":      requestID,
		"conversation_id": p.ConversationID,
	}, nil
}

// handleLLMCancel is the body of the llm.cancel RPC.
//
// Accepts both `request_id` and `conversation_id` for backward
// compatibility. When `request_id` is provided, the specific
// stream is canceled. When only `conversation_id` is provided,
// all streams for that conversation are canceled. When both are
// provided, `request_id` wins.
func handleLLMCancel(
	ctx context.Context,
	params json.RawMessage,
	sm *stream.Manager,
	auditLog *audit.Log,
) (any, error) {
	var p struct {
		RequestID      string `json:"request_id"`
		ConversationID int64  `json:"conversation_id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
	}
	if p.RequestID == "" && p.ConversationID == 0 {
		return nil, &ipc.Error{
			Code:    ipc.CodeInvalidParams,
			Message: "request_id or conversation_id is required",
		}
	}

	if p.RequestID != "" {
		if err := sm.Cancel(p.RequestID); err != nil {
			return nil, mapStreamError(err)
		}
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorGUI, Action: "llm.cancel", App: appConduraG,
			Level: auditLevelInfo, Result: auditResultAllow,
			Message: "request_id=" + p.RequestID,
		})
		return map[string]any{
			"canceled":   true,
			"request_id": p.RequestID,
			"timestamp":  time.Now().UTC().Format(time.RFC3339), //nolint:goconst
		}, nil
	}

	// Cancel all streams for the conversation.
	canceled := sm.CancelByConversation(p.ConversationID)
	_ = auditLog.Append(ctx, audit.Event{
		Actor: actorGUI, Action: "llm.cancel", App: appConduraG,
		Level: auditLevelInfo, Result: auditResultAllow,
		Message: "conversation_id=" + strconv.FormatInt(p.ConversationID, 10),
	})
	return map[string]any{
		"canceled":        true,
		"conversation_id": p.ConversationID,
		"canceled_count":  canceled,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// mapStreamError converts a stream package error into the appropriate
// IPC error code. Unknown errors are returned as-is.
func mapStreamError(err error) error {
	switch {
	case isStreamErr(err, stream.ErrNotFound):
		return &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
	case isStreamErr(err, stream.ErrHalted):
		return &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	case isStreamErr(err, stream.ErrContextFull):
		return &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
	case isStreamErr(err, stream.ErrAlreadyExists):
		return &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	}
	return err
}

// isStreamErr reports whether err is (or wraps) target.
func isStreamErr(err, target error) bool {
	return errors.Is(err, target)
}

// registerAuditMethods wires audit.list.
func registerAuditMethods(srv *ipc.Server, auditLog *audit.Log) {
	srv.Register("audit.list", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Limit  int    `json:"limit"`
			Offset int    `json:"offset"`
			Since  string `json:"since"`
			Action string `json:"action"`
			Level  string `json:"level"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		q := audit.Query{
			Limit:  p.Limit,
			Offset: p.Offset,
			Action: p.Action,
			Level:  p.Level,
		}
		if p.Since != "" {
			if t, err := time.Parse(time.RFC3339, p.Since); err == nil {
				q.Since = t
			}
		}
		return auditLog.List(ctx, q)
	})
}

// registerHaltMethods wires daemon.halt + daemon.resume_request +
// halt.confirm_resume + halt.state.
//
// T3b sticky resume (P0-1 core): the previous `daemon.resume` IPC
// handler was unbounded — any IPC token-holder (including a compromised
// in-process conductor) could silently un-halt the agent. The new
// flow requires a human-confirmed action the in-process conductor
// cannot invoke:
//
//   - `daemon.resume_request` (IPC, this file): mints a one-time ticket
//     with a 5-min TTL + rate limit. Returns the ticket to the caller.
//   - `halt.confirm_resume` (IPC, this file): consumes the ticket + a
//     human-only secret (the CLI prompts the user). Constant-time
//     secret compare. On success: resumes the flag + the net guard.
//   - `condura resume --confirm <ticket>` (CLI, separate OS process):
//     prompts for the secret, opens its own IPC client, calls
//     halt.confirm_resume. Out of the in-process trust boundary.
//
// The OLD `daemon.resume` IPC is kept as a thin deprecation shim that
// returns a clear migration error pointing at the new flow. Frontend
// code calling daemonResume() will see the clear error; the new method
// is daemonResumeRequest() + haltConfirmResume(ticket, secret).
//
// The watchdog auto-halt path (which calls HaltFlag.Halt directly,
// bypassing the RPC handler) is still covered by guardAwareHaltFlag
// from the N3 fix — its Halt call also toggles the network guard.
//
//nolint:goconst // JSON response field keys repeat across the 3 handlers; readability > extracting 4 consts.
func registerHaltMethods(
	srv *ipc.Server,
	haltFlag *halt.Flag,
	auditLog *audit.Log,
	sm *stream.Manager,
	guard halt.NetworkGuard,
	ticketStore *ResumeTicketStore,
	resumeSecret *ResumeSecretManager,
) {
	srv.Register("daemon.halt", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Reason string `json:"reason"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		_, _ = haltFlag.Halt(ctx, p.Reason)
		// N3: toggle the network guard so a halted agent cannot make outbound calls.
		if guard != nil {
			_ = guard.Halt(p.Reason)
		}
		streamsCanceled := 0
		if sm != nil {
			streamsCanceled = sm.CancelAll()
		}
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorIPC, Action: "daemon.halt", App: appConduraG,
			Level: auditLevelWarn, Result: auditResultAllow,
			Message: p.Reason,
		})
		return map[string]any{
			"halted":                  true,
			"active_streams_canceled": streamsCanceled,
			"timestamp":               time.Now().UTC().Format(time.RFC3339),
		}, nil
	})
	srv.Register("daemon.resume_request", func(ctx context.Context, _ json.RawMessage) (any, error) {
		// Must be halted to be resumeable. If not halted, treat as a
		// no-op so the user-facing flow "click resume" doesn't crash.
		if !haltFlag.IsHalted() {
			return map[string]any{
				"halted": false,
				"reason": "daemon is not halted",
				"ticket": "",
			}, nil
		}
		ticket, err := ticketStore.Mint()
		if err != nil {
			_ = auditLog.Append(ctx, audit.Event{
				Actor: actorIPC, Action: "daemon.resume_request", App: appConduraG,
				Level: auditLevelWarn, Result: auditResultError,
				Message: err.Error(),
			})
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorIPC, Action: "daemon.resume_request", App: appConduraG,
			Level: auditLevelInfo, Result: auditResultAllow,
			Message: "ticket minted; awaiting human confirm",
		})
		return map[string]any{
			"halted":      true,
			"ticket":      ticket,
			"ttl_seconds": int(resumeTicketTTL / time.Second),
			"confirm_via": "condura resume --confirm <ticket>   OR   halt.confirm_resume IPC",
		}, nil
	})

	srv.Register("halt.confirm_resume", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Ticket string `json:"ticket"`
			Secret string `json:"secret"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if p.Ticket == "" || p.Secret == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "ticket and secret are required"}
		}
		expected, secErr := resumeSecret.Load()
		if secErr != nil {
			_ = auditLog.Append(ctx, audit.Event{
				Actor: actorIPC, Action: "halt.confirm_resume", App: appConduraG,
				Level: auditLevelError, Result: auditResultError,
				Message: "secret load failed: " + secErr.Error(),
			})
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "resume secret unavailable"}
		}
		if _, err := ticketStore.Consume(p.Ticket, p.Secret, expected); err != nil {
			// Audit denial at actorGUIHuman-failed (i.e. someone
			// attempted the privileged path without a valid secret /
			// ticket). Treat as suspicious; surface the denial reason
			// to the caller but do not leak the expected secret.
			_ = auditLog.Append(ctx, audit.Event{
				Actor: actorIPC, Action: "halt.confirm_resume", App: appConduraG,
				Level: auditLevelWarn, Result: auditResultDeny,
				Message: err.Error(),
			})
			return nil, &ipc.Error{Code: ipc.CodeInvalidRequest, Message: err.Error()}
		}
		// Valid ticket + matching secret → un-halt.
		// Enforce cooldown: if the halt flag refuses to resume (e.g.
		// 5-minute cooldown hasn't elapsed), refuse the entire
		// operation. Never resume the network guard while the halt
		// flag is still active — that would create a split-brain
		// where the flag says "halted" but the guard says "running."
		if _, err := haltFlag.Resume(ctx); err != nil {
			_ = auditLog.Append(ctx, audit.Event{
				Actor: actorGUIHuman, Action: "halt.confirm_resume", App: appConduraG,
				Level: auditLevelWarn, Result: auditResultDeny,
				Message: "resume cooldown: " + err.Error(),
			})
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		if guard != nil {
			_ = guard.Resume()
		}
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorGUIHuman, Action: "halt.confirm_resume", App: appConduraG,
			Level: auditLevelInfo, Result: auditResultAllow,
			Message: "human-confirmed resume via sticky ticket",
		})
		return map[string]any{
			"resumed":   true,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}, nil
	})

	// Deprecation shim for the old IPC name. Returns a clear migration
	// error so any leftover GUI/CLI client gets a usable message
	// instead of silently resuming.
	srv.Register("daemon.resume", func(ctx context.Context, _ json.RawMessage) (any, error) {
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorIPC, Action: "daemon.resume", App: appConduraG,
			Level: auditLevelWarn, Result: auditResultDeny,
			Message: "deprecated: use daemon.resume_request + halt.confirm_resume (or `condura resume --confirm`)",
		})
		return nil, &ipc.Error{
			Code:    ipc.CodeInvalidRequest,
			Message: "daemon.resume is deprecated since T3b sticky-resume; call daemon.resume_request to mint a ticket, then halt.confirm_resume (or `condura resume --confirm <ticket>`) to confirm with the human-confirmation secret",
		}
	})
	srv.Register("halt.state", func(_ context.Context, _ json.RawMessage) (any, error) {
		s := haltFlag.Halted()
		return map[string]any{
			"halted": s.Halted,
			"since":  formatTime(s.Since),
			"reason": s.Reason,
		}, nil
	})
}

// itoa is a tiny int-to-string converter to avoid importing strconv
// in a hot file.
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// formatTime returns the time as RFC3339, or "" if zero.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
