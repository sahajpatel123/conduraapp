package failover

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// CircuitBreaker
// -----------------------------------------------------------------------------

func TestCB_StartsClosed(t *testing.T) {
	b := NewCircuitBreaker(2, time.Minute)
	assert.Equal(t, CircuitClosed, b.State())
	assert.True(t, b.Allow())
}

func TestCB_OpensAfterThreshold(t *testing.T) {
	b := NewCircuitBreaker(2, time.Minute)
	b.RecordFailure()
	assert.Equal(t, CircuitClosed, b.State())
	b.RecordFailure()
	assert.Equal(t, CircuitOpen, b.State())
	assert.False(t, b.Allow())
}

func TestCB_HalfOpenAfterCoolDown(t *testing.T) {
	b := NewCircuitBreaker(1, 10*time.Millisecond)
	b.RecordFailure()
	assert.Equal(t, CircuitOpen, b.State())
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, CircuitHalfOpen, b.State())
	assert.True(t, b.Allow(), "first half-open call should be allowed")
	assert.False(t, b.Allow(), "second concurrent half-open call should be blocked")
}

func TestCB_HalfOpenSuccessClosesBreaker(t *testing.T) {
	b := NewCircuitBreaker(1, 10*time.Millisecond)
	b.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	assert.True(t, b.Allow())
	b.RecordSuccess()
	assert.Equal(t, CircuitClosed, b.State())
}

func TestCB_HalfOpenFailureReopens(t *testing.T) {
	b := NewCircuitBreaker(1, 10*time.Millisecond)
	b.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	assert.True(t, b.Allow())
	b.RecordFailure()
	assert.Equal(t, CircuitOpen, b.State())
}

func TestCB_RecordSuccessResets(t *testing.T) {
	b := NewCircuitBreaker(3, time.Minute)
	b.RecordFailure()
	b.RecordFailure()
	b.RecordSuccess()
	b.RecordFailure()
	assert.Equal(t, CircuitClosed, b.State(), "success should reset failure count")
}

func TestCB_Reset(t *testing.T) {
	b := NewCircuitBreaker(1, time.Hour)
	b.RecordFailure()
	assert.Equal(t, CircuitOpen, b.State())
	b.Reset()
	assert.Equal(t, CircuitClosed, b.State())
	assert.True(t, b.Allow())
}

func TestCB_Defaults(t *testing.T) {
	b := NewCircuitBreaker(0, 0)
	assert.Equal(t, 3, b.failureThreshold)
	assert.Equal(t, 30*time.Second, b.coolDown)
}

func TestCB_Concurrent(t *testing.T) {
	b := NewCircuitBreaker(100, time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.RecordFailure()
		}()
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Allow()
		}()
	}
	wg.Wait()
}

// -----------------------------------------------------------------------------
// Run (chain)
// -----------------------------------------------------------------------------

func TestRun_SuccessOnFirst(t *testing.T) {
	var calls atomic.Int32
	res, err := Run(context.Background(),
		[]Candidate{{Provider: "p1", Model: "m1"}},
		func(ctx context.Context, p, m string) error {
			calls.Add(1)
			return nil
		})
	require.NoError(t, err)
	assert.Equal(t, int32(1), calls.Load())
	assert.Equal(t, "p1", res.Provider)
	assert.Equal(t, 1, res.Attempts)
}

func TestRun_Failover(t *testing.T) {
	var calls atomic.Int32
	res, err := Run(context.Background(),
		[]Candidate{
			{Provider: "p1", Model: "m1"},
			{Provider: "p2", Model: "m2"},
			{Provider: "p3", Model: "m3"},
		},
		func(ctx context.Context, p, m string) error {
			calls.Add(1)
			if p == "p2" {
				return nil
			}
			return errors.New("fail")
		})
	require.NoError(t, err)
	assert.Equal(t, int32(2), calls.Load())
	assert.Equal(t, "p2", res.Provider)
	assert.Equal(t, 2, res.Attempts)
}

func TestRun_AllExhausted(t *testing.T) {
	_, err := Run(context.Background(),
		[]Candidate{{Provider: "p1"}, {Provider: "p2"}},
		func(ctx context.Context, p, m string) error {
			return errors.New("nope")
		})
	assert.ErrorIs(t, err, ErrAllExhausted)
}

func TestRun_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Run(ctx,
		[]Candidate{{Provider: "p1"}},
		func(ctx context.Context, p, m string) error {
			return errors.New("x")
		})
	assert.Error(t, err)
}

func TestRun_EmptyCandidates(t *testing.T) {
	_, err := Run(context.Background(), nil, nil)
	assert.ErrorIs(t, err, ErrAllExhausted)
}

// -----------------------------------------------------------------------------
// BreakerRegistry
// -----------------------------------------------------------------------------

func TestBreakerRegistry_For(t *testing.T) {
	r := NewBreakerRegistry(1, time.Minute)
	b1 := r.For("openai")
	b2 := r.For("openai")
	assert.Same(t, b1, b2)
	b3 := r.For("anthropic")
	assert.NotSame(t, b1, b3)
}

func TestBreakerRegistry_ResetAll(t *testing.T) {
	r := NewBreakerRegistry(1, time.Hour)
	r.For("a").RecordFailure()
	r.For("b").RecordFailure()
	assert.Equal(t, CircuitOpen, r.For("a").State())
	r.ResetAll()
	assert.Equal(t, CircuitClosed, r.For("a").State())
}

func TestBreakerRegistry_States(t *testing.T) {
	r := NewBreakerRegistry(1, time.Hour)
	r.For("a").RecordFailure()
	r.For("b")
	states := r.States()
	assert.Equal(t, CircuitOpen, states["a"])
	assert.Equal(t, CircuitClosed, states["b"])
}

// -----------------------------------------------------------------------------
// SpendMonitor
// -----------------------------------------------------------------------------

func TestSpendMonitor_Basic(t *testing.T) {
	m := NewSpendMonitor(SpendCap{USDPerDay: 5.0})
	assert.Equal(t, 5.0, m.Remaining())
	m.Record(2.0)
	assert.Equal(t, 3.0, m.Remaining())
	assert.Equal(t, 2.0, m.Spent())
}

func TestSpendMonitor_Allow(t *testing.T) {
	m := NewSpendMonitor(SpendCap{USDPerDay: 5.0})
	m.Record(4.0)
	assert.True(t, m.Allow(0.99))
	assert.False(t, m.Allow(1.01))
}

func TestSpendMonitor_NewDay(t *testing.T) {
	m := NewSpendMonitor(SpendCap{USDPerDay: 5.0})
	day := "2025-01-01"
	m.nowFn = func() time.Time { return mustParse(day) }
	m.Record(5.0)
	assert.Equal(t, 0.0, m.Remaining())

	m.nowFn = func() time.Time { return mustParse("2025-01-02") }
	assert.Equal(t, 5.0, m.Remaining(), "new day should reset spend")
}

func TestSpendMonitor_SetCap(t *testing.T) {
	m := NewSpendMonitor(SpendCap{USDPerDay: 5.0})
	m.SetCap(SpendCap{USDPerDay: 10.0})
	assert.Equal(t, 10.0, m.Cap().USDPerDay)
}

func TestSpendMonitor_RemainingFloorZero(t *testing.T) {
	m := NewSpendMonitor(SpendCap{USDPerDay: 5.0})
	m.Record(10.0)
	assert.Equal(t, 0.0, m.Remaining())
}

func mustParse(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

// -----------------------------------------------------------------------------
// Failover orchestration
// -----------------------------------------------------------------------------

type fakeLLM struct {
	calls   atomic.Int32
	results map[string]error
}

func (f *fakeLLM) Chat(_ context.Context, provider, model string) (Usage, error) {
	f.calls.Add(1)
	if err, ok := f.results[provider+"/"+model]; ok {
		return Usage{}, err
	}
	return Usage{InputTokens: 1, OutputTokens: 1, CostUSD: 0.01}, nil
}

func TestFailover_HappyPath(t *testing.T) {
	fc := &fakeLLM{results: map[string]error{}}
	f := New([]Provider{
		{Name: "p1", Breaker: NewCircuitBreaker(3, time.Minute), Client: fc, Models: []string{"m1"}},
	}, NewSpendMonitor(SpendCap{USDPerDay: 1.0}))
	usage, prov, model, err := f.Chat(context.Background(), 0, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, usage.InputTokens)
	assert.Equal(t, "p1", prov.Name)
	assert.Equal(t, Model("m1"), model)
}

func TestFailover_SkipsOpenBreaker(t *testing.T) {
	fc := &fakeLLM{results: map[string]error{}}
	b1 := NewCircuitBreaker(1, time.Hour)
	b1.RecordFailure() // open
	f := New([]Provider{
		{Name: "p1", Breaker: b1, Client: fc, Models: []string{"m1"}},
		{Name: "p2", Breaker: NewCircuitBreaker(3, time.Minute), Client: fc, Models: []string{"m2"}},
	}, nil)
	_, prov, _, err := f.Chat(context.Background(), 0, nil)
	require.NoError(t, err)
	assert.Equal(t, "p2", prov.Name)
}

func TestFailover_NoProviders(t *testing.T) {
	f := New(nil, nil)
	_, _, _, err := f.Chat(context.Background(), 0, nil)
	assert.Error(t, err)
}

func TestFailover_AllExhausted(t *testing.T) {
	fc := &fakeLLM{results: map[string]error{
		"p1/m1": fmt.Errorf("nope"),
		"p2/m2": fmt.Errorf("nope"),
	}}
	f := New([]Provider{
		{Name: "p1", Breaker: NewCircuitBreaker(3, time.Minute), Client: fc, Models: []string{"m1"}},
		{Name: "p2", Breaker: NewCircuitBreaker(3, time.Minute), Client: fc, Models: []string{"m2"}},
	}, nil)
	_, _, _, err := f.Chat(context.Background(), 0, nil)
	assert.ErrorIs(t, err, ErrAllExhausted)
}

func TestFailover_BreakerOpensOnFailure(t *testing.T) {
	fc := &fakeLLM{results: map[string]error{"p1/m1": fmt.Errorf("nope")}}
	b := NewCircuitBreaker(1, time.Hour)
	f := New([]Provider{
		{Name: "p1", Breaker: b, Client: fc, Models: []string{"m1"}},
	}, nil)
	_, _, _, _ = f.Chat(context.Background(), 0, nil)
	assert.Equal(t, CircuitOpen, b.State())
}

func TestFailover_BreakerClosesOnSuccess(t *testing.T) {
	fc := &fakeLLM{results: map[string]error{}}
	b := NewCircuitBreaker(3, time.Hour)
	b.RecordFailure()
	b.RecordFailure()
	f := New([]Provider{
		{Name: "p1", Breaker: b, Client: fc, Models: []string{"m1"}},
	}, nil)
	_, _, _, err := f.Chat(context.Background(), 0, nil)
	require.NoError(t, err)
	assert.Equal(t, CircuitClosed, b.State())
}

func TestFailover_SpendCap(t *testing.T) {
	fc := &fakeLLM{results: map[string]error{}}
	mon := NewSpendMonitor(SpendCap{USDPerDay: 0.005}) // very low
	est := func(p, m string, _ int) (float64, error) {
		if p == "p1" {
			return 0.01, nil // over cap
		}
		return 0.001, nil
	}
	f := New([]Provider{
		{Name: "p1", Breaker: NewCircuitBreaker(3, time.Minute), Client: fc, Models: []string{"m1"}},
		{Name: "p2", Breaker: NewCircuitBreaker(3, time.Minute), Client: fc, Models: []string{"m2"}},
	}, mon)
	_, prov, _, err := f.Chat(context.Background(), 10, est)
	require.NoError(t, err)
	assert.Equal(t, "p2", prov.Name)
}

func TestFailover_EstimateErrorIsIgnored(t *testing.T) {
	fc := &fakeLLM{results: map[string]error{}}
	mon := NewSpendMonitor(SpendCap{USDPerDay: 1.0})
	est := func(p, m string, _ int) (float64, error) {
		return 0, errors.New("estimate failed")
	}
	f := New([]Provider{
		{Name: "p1", Breaker: NewCircuitBreaker(3, time.Minute), Client: fc, Models: []string{"m1"}},
	}, mon)
	_, _, _, err := f.Chat(context.Background(), 0, est)
	require.NoError(t, err, "estimate error should not block call")
}

func TestFailover_RecordSpendOnSuccess(t *testing.T) {
	fc := &fakeLLM{results: map[string]error{}}
	mon := NewSpendMonitor(SpendCap{USDPerDay: 1.0})
	f := New([]Provider{
		{Name: "p1", Breaker: NewCircuitBreaker(3, time.Minute), Client: fc, Models: []string{"m1"}},
	}, mon)
	_, _, _, err := f.Chat(context.Background(), 0, nil)
	require.NoError(t, err)
	assert.InDelta(t, 0.01, mon.Spent(), 0.0001)
}

func TestFailover_NoMatchingModel(t *testing.T) {
	fc := &fakeLLM{results: map[string]error{}}
	mon := NewSpendMonitor(SpendCap{USDPerDay: 0.0}) // nothing allowed
	est := func(p, m string, _ int) (float64, error) { return 0.01, nil }
	f := New([]Provider{
		{Name: "p1", Breaker: NewCircuitBreaker(3, time.Minute), Client: fc, Models: []string{"m1"}},
		{Name: "p2", Breaker: NewCircuitBreaker(3, time.Minute), Client: fc, Models: []string{"m2"}},
	}, mon)
	_, _, _, err := f.Chat(context.Background(), 0, est)
	assert.ErrorIs(t, err, ErrAllExhausted)
}
