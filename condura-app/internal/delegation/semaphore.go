package delegation

import (
	"context"
	"sync"
)

const defaultPerLimit = 4
const defaultGlobalLimit = 5

// SemaphoreManager limits concurrent sub-agent execution per MISSION S13.4.
type SemaphoreManager struct {
	perAgent map[string]chan struct{}
	global   chan struct{}
	perLimit int
	mu       sync.Mutex
}

// NewSemaphoreManager creates a semaphore with the given limits.
func NewSemaphoreManager(perLimit, globalLimit int) *SemaphoreManager {
	if perLimit <= 0 {
		perLimit = defaultPerLimit
	}
	if globalLimit <= 0 {
		globalLimit = defaultGlobalLimit
	}
	return &SemaphoreManager{
		perAgent: make(map[string]chan struct{}),
		global:   make(chan struct{}, globalLimit),
		perLimit: perLimit,
	}
}

// Acquire blocks until a slot is available.
func (m *SemaphoreManager) Acquire(ctx context.Context, agentName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case m.global <- struct{}{}:
	}

	m.mu.Lock()
	ch, ok := m.perAgent[agentName]
	if !ok {
		ch = make(chan struct{}, m.perLimit)
		m.perAgent[agentName] = ch
	}
	m.mu.Unlock()

	select {
	case <-ctx.Done():
		<-m.global
		return ctx.Err()
	case ch <- struct{}{}:
		return nil
	}
}

// Release frees one per-agent and one global slot.
func (m *SemaphoreManager) Release(agentName string) {
	m.mu.Lock()
	ch, ok := m.perAgent[agentName]
	m.mu.Unlock()
	if ok {
		select {
		case <-ch:
		default:
		}
	}
	select {
	case <-m.global:
	default:
	}
}

// Available returns the number of global slots remaining.
func (m *SemaphoreManager) Available() int {
	return cap(m.global) - len(m.global)
}
