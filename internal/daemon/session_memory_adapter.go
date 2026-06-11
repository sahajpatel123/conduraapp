package daemon

import (
	"context"

	"github.com/sahajpatel123/synapticapp/internal/memory"
	"github.com/sahajpatel123/synapticapp/internal/session"
)

// sessionMemoryAdapter adapts memory.StoreManager → session.MemoryStore.
type sessionMemoryAdapter struct {
	mgr *memory.StoreManager
}

func (a *sessionMemoryAdapter) Recall(ctx context.Context, query string, limit int) ([]string, error) {
	mems, err := a.mgr.Recall(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(mems))
	for i, m := range mems {
		out[i] = m.Content
	}
	return out, nil
}

var _ session.MemoryStore = (*sessionMemoryAdapter)(nil)
