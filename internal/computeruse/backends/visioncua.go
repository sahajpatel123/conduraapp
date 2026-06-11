package backends

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/computeruse"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/llm"
)

// VisionCUAConfig controls the Vision CUA backend settings.
// Enabled defaults to false — Vision CUA requires explicit user
// opt-in because it sends screenshots to a cloud LLM (network/privacy
// boundary per MISSION §2: local-first, no surveillance).
//
// Design: resolve-only (provides coordinates, does not execute).
// The resolved coordinates are returned as Bounds on the ActionResult
// and must still pass through the Gatekeeper before physical execution.
type VisionCUAConfig struct {
	Enabled             bool
	Provider            llm.Provider
	Model               string
	MaxConsecutiveCalls int
	Gatekeeper          gatekeeper.Gatekeeper
}

// VisionCUABackend implements computeruse.Backend using an LLM
// provider's vision capability. This is the fourth (last-resort) tier
// in the 4-tier router.
type VisionCUABackend struct {
	callCount int
	cfg       VisionCUAConfig
}

var _ computeruse.Backend = (*VisionCUABackend)(nil)

// NewVisionCUA creates a Vision CUA backend. Returns nil if disabled.
func NewVisionCUA(cfg VisionCUAConfig) *VisionCUABackend {
	if !cfg.Enabled {
		return nil
	}
	if cfg.MaxConsecutiveCalls <= 0 {
		cfg.MaxConsecutiveCalls = 5
	}
	return &VisionCUABackend{cfg: cfg}
}

// Name returns the backend identifier.
func (b *VisionCUABackend) Name() string { return "vision-cua" }

// IsAvailable checks if the backend is enabled with a provider.
func (b *VisionCUABackend) IsAvailable(_ context.Context) bool {
	return b.cfg.Enabled && b.cfg.Provider != nil
}

// Capabilities returns screenshot and input actions.
func (b *VisionCUABackend) Capabilities() []computeruse.Capability {
	return []computeruse.Capability{
		computeruse.CapScreenshot, computeruse.CapClick,
		computeruse.CapType, computeruse.CapScroll,
		computeruse.CapKeyPress,
	}
}

// CaptureScreen captures via screencapture(1).
func (b *VisionCUABackend) CaptureScreen(_ context.Context) (*computeruse.Screenshot, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "screencapture", "-x", "-t", "png", "-") //nolint:gosec
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("vision-cua: screencapture: %w", err)
	}
	return &computeruse.Screenshot{Image: out, Timestamp: time.Now()}, nil
}

// GetAXTree is unsupported — Vision CUA reads pixels, not AX.
func (b *VisionCUABackend) GetAXTree(_ context.Context) (*computeruse.AXTree, error) {
	return nil, computeruse.ErrUnsupportedAction
}

// Execute runs the vision loop: screenshot → LLM → parse → result.
func (b *VisionCUABackend) Execute(_ context.Context, action *computeruse.Action) (*computeruse.ActionResult, error) {
	return b.doExecute(action)
}

func (b *VisionCUABackend) doExecute(action *computeruse.Action) (*computeruse.ActionResult, error) {
	start := time.Now()

	b.callCount++
	if b.callCount > b.cfg.MaxConsecutiveCalls {
		return &computeruse.ActionResult{Success: false, Duration: time.Since(start), Action: action},
			fmt.Errorf("vision-cua: exceeded max calls (%d)", b.cfg.MaxConsecutiveCalls)
	}

	ctx := context.Background()

	ss, err := b.CaptureScreen(ctx)
	if err != nil {
		return failResult(action, start, err), err
	}

	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(ss.Image)

	if b.cfg.Gatekeeper != nil {
		ba := action.ToBlastRadius()
		decision, _ := b.cfg.Gatekeeper.Evaluate(ctx, ba)
		if decision == gatekeeper.Deny {
			err := fmt.Errorf("vision-cua: gatekeeper denied action")
			return failResult(action, start, err), err
		}
	}

	resp, err := b.cfg.Provider.Chat(ctx, llm.ChatRequest{
		Model: b.cfg.Model,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: visSysPrompt},
			{Role: llm.RoleUser, Content: b.buildPrompt(action) + "\n\nImage: " + dataURI},
		},
	})
	if err != nil {
		return failResult(action, start, err), err
	}

	return b.parseResponse(resp.Message.Content, action), nil
}

func (b *VisionCUABackend) buildPrompt(action *computeruse.Action) string {
	var sb strings.Builder
	sb.WriteString("Perform this action: ")
	sb.WriteString(string(action.Type))
	if action.Target != nil {
		sb.WriteString(" on ")
		sb.WriteString(action.Target.Title)
	}
	if action.Value != "" {
		sb.WriteString(" with value ")
		sb.WriteString(action.Value)
	}
	sb.WriteString("\nReturn JSON: {\"x\":number,\"y\":number}")
	return sb.String()
}

func (b *VisionCUABackend) parseResponse(raw string, orig *computeruse.Action) *computeruse.ActionResult {
	var p struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	cleaned := extractJSON(raw)
	if err := json.Unmarshal([]byte(cleaned), &p); err != nil {
		return failResult(orig, time.Time{}, fmt.Errorf("vision-cua: parse: %w", err))
	}
	return &computeruse.ActionResult{
		Success: true,
		Action: &computeruse.Action{
			Type: orig.Type,
			Bounds: &computeruse.Rect{
				X: p.X, Y: p.Y, Width: 1, Height: 1,
			},
			Value: orig.Value,
		},
	}
}

const visSysPrompt = `You are a computer vision locator. Look at the screenshot and find the pixel coordinates for the requested action. Return ONLY valid JSON: {"x": number, "y": number}`

func failResult(action *computeruse.Action, start time.Time, err error) *computeruse.ActionResult {
	return &computeruse.ActionResult{
		Success:  false,
		Error:    err,
		Duration: time.Since(start),
		Action:   action,
	}
}

func extractJSON(raw string) string {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return s
}
