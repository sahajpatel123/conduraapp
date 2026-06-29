package daemon

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/sahajpatel123/conduraapp/internal/ipc"
	"github.com/sahajpatel123/conduraapp/internal/trust"
)

// registerTrustMethods wires the trust.* method family (Phase 16,
// Rec 5: per-workspace trust).
//
//   - trust.list — list all trust entries (sorted by recency).
//     Drives the Settings → "Trusted folders" view.
//   - trust.grant — add or update a workspace trust. The GUI
//     calls this from the consent dialog's "Always allow in
//     this folder" button (or from Settings).
//   - trust.revoke — remove a workspace trust. The GUI calls
//     this from Settings → "Stop trusting this folder".
//   - trust.workspace_id_for — convert a path to its canonical
//     workspace ID. The GUI calls this to pre-fill the consent
//     dialog with the resolved workspace.
func registerTrustMethods(srv *ipc.Server, subs *Subsystems) {
	store := trustStoreOf(subs)
	if store == nil {
		na := func(_ context.Context, _ json.RawMessage) (any, error) {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "trust store not available"}
		}
		srv.Register("trust.list", na)
		srv.Register("trust.grant", na)
		srv.Register("trust.revoke", na)
		srv.Register("trust.workspace_id_for", na)
		return
	}

	srv.Register("trust.list", func(_ context.Context, _ json.RawMessage) (any, error) {
		entries := store.List()
		out := make([]map[string]any, 0, len(entries))
		for _, e := range entries {
			out = append(out, map[string]any{
				"workspace_id": e.WorkspaceID,
				"label":        e.Label,
				"always_allow": e.AlwaysAllow,
				"created_at":   e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
				"last_used_at": e.LastUsedAt.UTC().Format("2006-01-02T15:04:05Z"),
				"app_scope":    string(e.AppScope),
			})
		}
		return out, nil
	})

	srv.Register("trust.grant", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			WorkspaceID string `json:"workspace_id"`
			Label       string `json:"label"`
			AppScope    string `json:"app_scope"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if p.WorkspaceID == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "workspace_id is required"}
		}
		entry, err := store.Grant(p.WorkspaceID, p.Label, trust.AppScope(p.AppScope))
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{
			"workspace_id": entry.WorkspaceID,
			"label":        entry.Label,
			"created_at":   entry.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}, nil
	})

	srv.Register("trust.revoke", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			WorkspaceID string `json:"workspace_id"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if err := store.Revoke(p.WorkspaceID); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{"ok": true}, nil
	})

	srv.Register("trust.workspace_id_for", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		// Expand ~ to home for the convenience of the GUI.
		if strings.HasPrefix(p.Path, "~/") {
			if home, err := os.UserHomeDir(); err == nil && home != "" {
				p.Path = filepath.Join(home, p.Path[2:])
			}
		}
		return map[string]any{
			"workspace_id": trust.WorkspaceIDFor(p.Path),
		}, nil
	})
}

// trustStoreOf returns the trust store from the safety components,
// or nil if it's not available. Centralizes the lookup so the
// not-available stub above doesn't have to know the layout.
func trustStoreOf(subs *Subsystems) *trust.Store {
	if subs == nil || subs.Safety == nil {
		return nil
	}
	return subs.Safety.Trust
}
