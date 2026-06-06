package health

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_AddRemove(t *testing.T) {
	r := New()
	r.Add(Check{Name: "a", Check: func(_ context.Context) error { return nil }})
	r.Add(Check{Name: "b", Check: func(_ context.Context) error { return nil }})
	assert.Equal(t, []string{"a", "b"}, r.Names())
	r.Remove("a")
	assert.Equal(t, []string{"b"}, r.Names())
}

func TestRegister_DefaultTimeout(t *testing.T) {
	r := New()
	r.Add(Check{Name: "x", Check: func(_ context.Context) error { return nil }})
	r.mu.RLock()
	defer r.mu.RUnlock()
	assert.Equal(t, 2*time.Second, r.checks["x"].Timeout)
}

func TestSnapshot_AllOK(t *testing.T) {
	r := New()
	r.Add(Check{Name: "a", Check: func(_ context.Context) error { return nil }})
	r.Add(Check{Name: "b", Check: func(_ context.Context) error { return nil }})
	snap := r.Snapshot(context.Background())
	assert.Equal(t, StateOK, snap.State)
	assert.Len(t, snap.Results, 2)
}

func TestSnapshot_RequiredDown(t *testing.T) {
	r := New()
	r.Add(Check{Name: "a", Required: true, Check: func(_ context.Context) error { return errors.New("boom") }})
	snap := r.Snapshot(context.Background())
	assert.Equal(t, StateDown, snap.State)
}

func TestSnapshot_OptionalDown_Degrades(t *testing.T) {
	r := New()
	r.Add(Check{Name: "a", Check: func(_ context.Context) error { return nil }})
	r.Add(Check{Name: "b", Required: false, Check: func(_ context.Context) error { return errors.New("x") }})
	snap := r.Snapshot(context.Background())
	assert.Equal(t, StateDegraded, snap.State)
}

func TestSnapshot_TimeoutMarksDown(t *testing.T) {
	r := New()
	r.Add(Check{
		Name:     "slow",
		Required: true,
		Timeout:  50 * time.Millisecond,
		Check: func(ctx context.Context) error {
			select {
			case <-time.After(200 * time.Millisecond):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
	snap := r.Snapshot(context.Background())
	assert.Equal(t, StateDown, snap.State)
}

func TestSnapshot_ConcurrentChecks(t *testing.T) {
	r := New()
	for i := 0; i < 5; i++ {
		r.Add(Check{
			Name: string(rune('a' + i)),
			Check: func(_ context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
		})
	}
	start := time.Now()
	snap := r.Snapshot(context.Background())
	took := time.Since(start)
	// Sequential would be 50ms; concurrent should be ~10-20ms.
	assert.Less(t, took.Milliseconds(), int64(40), "checks should run concurrently")
	assert.Equal(t, StateOK, snap.State)
}

func TestSnapshot_ReplaceByName(t *testing.T) {
	r := New()
	r.Add(Check{Name: "a", Check: func(_ context.Context) error { return nil }})
	r.Add(Check{Name: "a", Required: true, Check: func(_ context.Context) error { return errors.New("boom") }})
	snap := r.Snapshot(context.Background())
	assert.Equal(t, StateDown, snap.State)
}

func TestSnapshot_ResultOrder(t *testing.T) {
	r := New()
	r.Add(Check{Name: "z", Check: func(_ context.Context) error { return nil }})
	r.Add(Check{Name: "a", Check: func(_ context.Context) error { return nil }})
	r.Add(Check{Name: "m", Check: func(_ context.Context) error { return nil }})
	snap := r.Snapshot(context.Background())
	// Order isn't guaranteed (map iteration), but the count must match.
	require.Len(t, snap.Results, 3)
}

func TestSnapshot_Time(t *testing.T) {
	r := New()
	before := time.Now()
	snap := r.Snapshot(context.Background())
	after := time.Now()
	assert.True(t, !snap.Time.Before(before) && !snap.Time.After(after))
}

func TestSnapshot_EmptyRegister(t *testing.T) {
	r := New()
	snap := r.Snapshot(context.Background())
	assert.Equal(t, StateOK, snap.State)
	assert.Empty(t, snap.Results)
}
