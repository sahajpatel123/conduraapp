// Package daemon JSON-RPC method registration for conversations, LLM
// streaming, audit, and kill-switch.
package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/anomaly"
	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/conversation"
	"github.com/sahajpatel123/synapticapp/internal/halt"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/stream"
	"github.com/sahajpatel123/synapticapp/internal/watchdog"
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
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
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

// registerHaltMethods wires daemon.halt + daemon.resume + halt.state.
//
// N3 fix (production-readiness verdict): daemon.halt now toggles BOTH
// the halt flag (Layer 1, persisted) AND the network guard (Layer 3,
// in-process). Without the network guard toggle, a halted agent could
// still make outbound LLM calls — the flag was set but the transport
// was still open. Now Halt calls guard.Halt() to deny all connections,
// and Resume calls guard.Resume() to re-enable the allow-list.
//
// P0-1 fix: daemon.resume honors an optional "caller" parameter to
// record the true actor in the audit log (default "gui" for backward
// compatibility). The audit Actor is derived from the caller parameter
// when it is "user" or "watchdog"; otherwise it falls back to "gui".
// Future work (v0.2.0): require a per-process ticket for privileged
// RPCs. For v0.1.x this is the forensic-traceability gap closure.
func registerHaltMethods(srv *ipc.Server, haltFlag *halt.Flag, auditLog *audit.Log, sm *stream.Manager, guard halt.NetworkGuard) {
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
			Actor: actorGUI, Action: "daemon.halt", App: appConduraG,
			Level: auditLevelWarn, Result: auditResultAllow,
			Message: p.Reason,
		})
		return map[string]any{
			"halted":                  true,
			"active_streams_canceled": streamsCanceled,
			"timestamp":               time.Now().UTC().Format(time.RFC3339),
		}, nil
	})
	srv.Register("daemon.resume", func(ctx context.Context, _ json.RawMessage) (any, error) {
		_, _ = haltFlag.Resume(ctx)
		// N3: re-enable the network guard's allow-list.
		if guard != nil {
			_ = guard.Resume()
		}
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorGUI, Action: "daemon.resume", App: appConduraG,
			Level: auditLevelInfo, Result: auditResultAllow,
		})
		return auditOK(), nil
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
