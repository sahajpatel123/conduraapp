package llm

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// Registry holds the set of configured LLM providers.
//
// Construction:
//   - r := llm.NewRegistry()
//   - r.Register("openai", llm.NewOpenAI(apiKey, models))
//   - r.Register("anthropic", llm.NewAnthropic(apiKey, models))
//   - ...
//
// The router layer (internal/failover) consults the registry to look up
// providers by name. The CLI and IPC layer use List() to enumerate.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{providers: map[string]Provider{}}
}

// Register adds a provider to the registry. If a provider with the same
// name is already registered, it is replaced.
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

// Get returns the provider with the given name, or false.
func (r *Registry) Get(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// MustGet returns the provider with the given name, or panics.
// Intended for code paths where the name is known good.
func (r *Registry) MustGet(name string) Provider {
	p, ok := r.Get(name)
	if !ok {
		panic(fmt.Sprintf("llm: provider %q not registered", name))
	}
	return p
}

// Delete removes a provider. Returns true if it existed.
func (r *Registry) Delete(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.providers[name]
	delete(r.providers, name)
	return ok
}

// Names returns the list of registered provider names, sorted.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for n := range r.providers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// List returns the registered providers in name-sorted order.
func (r *Registry) List() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// Len returns the number of registered providers.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.providers)
}

// Chat looks up the named provider and calls Chat on it.
func (r *Registry) Chat(ctx context.Context, name string, req ChatRequest) (ChatResponse, error) {
	p, ok := r.Get(name)
	if !ok {
		return ChatResponse{}, fmt.Errorf("%w: %q", ErrNoProvider, name)
	}
	return p.Chat(ctx, req)
}

// Stream looks up the named provider and calls Stream on it.
func (r *Registry) Stream(ctx context.Context, name string, req ChatRequest) (<-chan StreamEvent, func(), error) {
	p, ok := r.Get(name)
	if !ok {
		return nil, nil, fmt.Errorf("%w: %q", ErrNoProvider, name)
	}
	return p.Stream(ctx, req)
}

// -----------------------------------------------------------------------------
// Model aggregation across providers
// -----------------------------------------------------------------------------

// AllModels returns a name → []ModelInfo map for every registered provider.
func (r *Registry) AllModels() map[string][]ModelInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string][]ModelInfo, len(r.providers))
	for name, p := range r.providers {
		out[name] = p.Models()
	}
	return out
}
