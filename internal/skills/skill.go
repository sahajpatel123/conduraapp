// Package skills implements the agentskills.io-compatible skill system.
// Skills are reusable, versioned automation recipes.
package skills

import (
	"context"
	"time"
)

// TrustLevel indicates how much the user trusts a skill.
type TrustLevel string

// Trust levels for skills per MISSION S15.3.
const (
	TrustOfficial     TrustLevel = "official"
	TrustCommunity    TrustLevel = "community"
	TrustExperimental TrustLevel = "experimental"
)

// Skill represents an automation recipe in agentskills.io format.
type Skill struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Version        string     `json:"version"`
	Trust          TrustLevel `json:"trust"`
	TriggerPattern string     `json:"trigger_pattern"`
	Steps          []string   `json:"steps"`
	Dependencies   []string   `json:"dependencies,omitempty"`
	SuccessCount   int        `json:"success_count"`
	FailureCount   int        `json:"failure_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastUsed       time.Time  `json:"last_used"`

	// Provenance fields (Phase 12C — Skills Hub).
	Author      string     `json:"author,omitempty"`
	AuthorKey   string     `json:"author_key,omitempty"` // hex Ed25519 public key
	License     string     `json:"license,omitempty"`
	Source      string     `json:"source,omitempty"`   // "hub", "local", "url"
	HubID       string     `json:"hub_id,omitempty"`   // ID on hub.synaptic.app
	Checksum    string     `json:"checksum,omitempty"` // SHA-256 of archive
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

// Store is the interface for skill persistence.
type Store interface {
	Create(ctx context.Context, skill *Skill) error
	Get(ctx context.Context, id string) (*Skill, error)
	List(ctx context.Context, limit int) ([]*Skill, error)
	Search(ctx context.Context, query string, limit int) ([]*Skill, error)
	Update(ctx context.Context, skill *Skill) error
	Delete(ctx context.Context, id string) error
	IncrementUsage(ctx context.Context, id string, success bool) error
	Close() error
}
