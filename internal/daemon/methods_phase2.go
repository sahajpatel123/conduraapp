// Package daemon JSON-RPC method registration for conversations, LLM
// streaming, audit, and kill-switch.
package daemon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/conversation"
	"github.com/sahajpatel123/synapticapp/internal/halt"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

// registerConversationMethods wires conversations.* + llm.stream +
// llm.cancel.
func registerConversationMethods(srv *ipc.Server, store *conversation.Store, auditLog *audit.Log, haltFlag *halt.Flag) {
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
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorDaemon, Action: "conversations.create", App: appSynapticd,
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
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorDaemon, Action: "conversations.delete", App: appSynapticd,
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
		return auditOK(), nil
	})

	// llm.stream: the GUI calls this to start a stream. The actual
	// tokens are pushed over SSE; this method just kicks off the
	// call and returns once the first event is published (or an
	// error). The GUI listens to the 'stream' event on /events.
	//
	// Phase 2: the streaming pipeline through the LLM registry is
	// deferred to Phase 3. The GUI currently uses llm.chat (which
	// drains streams server-side) rather than a true streaming
	// protocol. We still expose llm.stream as a placeholder so the
	// TS side can wire it up; it returns an error pointing the
	// caller to llm.chat.
	srv.Register("llm.stream", func(ctx context.Context, params json.RawMessage) (any, error) {
		if haltFlag.IsHalted() {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "daemon is halted"}
		}
		_ = params
		return nil, &ipc.Error{
			Code:    ipc.CodeMethodNotFound,
			Message: "llm.stream not yet implemented; use llm.chat (which drains the stream server-side)",
		}
	})
	srv.Register("llm.cancel", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ConversationID int64 `json:"conversation_id"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		_ = p
		return auditOK(), nil
	})
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
func registerHaltMethods(srv *ipc.Server, haltFlag *halt.Flag, auditLog *audit.Log) {
	srv.Register("daemon.halt", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Reason string `json:"reason"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		_, _ = haltFlag.Halt(ctx, p.Reason)
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorGUI, Action: "daemon.halt", App: appSynapticG,
			Level: auditLevelWarn, Result: auditResultAllow,
			Message: p.Reason,
		})
		return map[string]any{
			"halted":                  true,
			"active_streams_canceled": 0,
			"timestamp":               time.Now().UTC().Format(time.RFC3339),
		}, nil
	})
	srv.Register("daemon.resume", func(ctx context.Context, _ json.RawMessage) (any, error) {
		_, _ = haltFlag.Resume(ctx)
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorGUI, Action: "daemon.resume", App: appSynapticG,
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
