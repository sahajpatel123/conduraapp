package daemon

import (
	"context"
	"encoding/json"

	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

func registerReachMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Reach == nil {
		registerReachNotAvailable(srv)
		return
	}

	srv.Register("channels.list", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return subs.Reach.List(ctx)
	})

	srv.Register("channels.connect", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Channel string `json:"channel"`
			Token   string `json:"token"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if !subs.GatekeeperAllow(ctx, "reach.connect", "Connect "+p.Channel+" messaging channel") {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "denied by safety policy"}
		}
		return subs.Reach.Connect(ctx, p.Channel, p.Token)
	})

	srv.Register("channels.disconnect", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Channel string `json:"channel"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if err := subs.Reach.Disconnect(ctx, p.Channel); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{"ok": true}, nil
	})

	srv.Register("channels.status", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Channel string `json:"channel"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		return subs.Reach.Status(ctx, p.Channel)
	})
}

func registerReachNotAvailable(srv *ipc.Server) {
	na := func(_ context.Context, _ json.RawMessage) (any, error) {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "channels subsystem not available"}
	}
	srv.Register("channels.list", na)
	srv.Register("channels.connect", na)
	srv.Register("channels.disconnect", na)
	srv.Register("channels.status", na)
}
