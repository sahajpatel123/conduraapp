// Package session — N2: chat tool_use → CU executor dispatch loop.
//
// When the Session has an Executor + Gatekeeper wired (act mode), Run
// uses runToolLoop instead of the streaming talk-only path. The model is
// offered a small set of condura_* tools; each tool_use it emits is gated
// through the Gatekeeper (consent + presence, the SAME modal the
// delegate bus uses) and dispatched via the executor (sanitized shell /
// CU resolver). The tool_result is round-tripped back to the model and
// the loop continues until the model stops calling tools (end_turn) or
// the iteration cap is hit.
//
// SAFETY: every dispatch goes through Gatekeeper.Evaluate. The model
// cannot execute anything the gate denies. shell.exec is DESTRUCTIVE
// (consent + presence required); computeruse.* is WRITE (consent).
// The executor re-gates defense-in-depth with the stored allow decision.
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/sahajpatel123/conduraapp/internal/audit"
	"github.com/sahajpatel123/conduraapp/internal/blastradius"
	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
	"github.com/sahajpatel123/conduraapp/internal/llm"
	"github.com/sahajpatel123/conduraapp/internal/pending"
	"github.com/sahajpatel123/conduraapp/internal/sanitize"
	"github.com/sahajpatel123/conduraapp/internal/status"
)

// maxToolIterations bounds the act tool loop so a model that keeps
// requesting tools cannot run forever. 8 is generous for ordinary
// multi-step tasks; the cap is audited when hit.
const maxToolIterations = 8

// Audit-payload constants for the tool loop's audit events. Lifted
// so goconst doesn't flag the repeated magic strings.
const (
	auditApp   = "session"
	auditWarn  = "warn"
	auditAllow = "allow"
	auditDeny  = "deny"
)

// JSON-schema string literals used in tool definitions. Lifted to
// consts so the linter catches typos and the schema stays consistent.
const (
	jsonSchemaObject      = "object"
	jsonSchemaString      = "string"
	jsonSchemaInteger     = "integer"
	jsonSchemaProperties  = "properties"
	jsonSchemaType        = "type"
	jsonSchemaDescription = "description"
	jsonSchemaRequired    = "required"
)

// runToolLoop is the act-mode chat loop (N2). It is non-streaming
// (Provider.Chat) because the Anthropic stream emits tool_use input as
// text deltas, not as Delta.ToolCalls, so tool_calls are only available
// on ChatResponse.Message.ToolCalls. The trade-off vs the streaming
// talk-only path is acceptable for v0.1.x act mode.
func (s *Session) runToolLoop(ctx context.Context, messages []llm.Message) (string, error) {
	tools := conduraTools()
	var lastText string
	s.setStatus(status.StatusThinking)

	for iter := 0; iter < maxToolIterations; iter++ {
		resp, err := s.cfg.Provider.Chat(ctx, s.cfg.ProviderName, llm.ChatRequest{
			Model:    s.cfg.Model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			s.setStatus(status.StatusError)
			return lastText, fmt.Errorf("session: act chat failed: %w", err)
		}
		lastText = resp.Message.Content
		if lastText != "" && s.cfg.Speaker != nil {
			_ = s.cfg.Speaker.Speak(ctx, lastText)
		}

		// No tool calls → end of turn. Persist the final answer + idle.
		if len(resp.Message.ToolCalls) == 0 {
			s.persistActAssistant(ctx, lastText)
			s.setStatus(status.StatusIdle)
			return lastText, nil
		}

		// Append the assistant turn (text + tool_calls), then dispatch
		// each tool_call (gated) and append the tool_result messages so
		// the next Chat sees the results.
		messages = append(messages, resp.Message)
		for _, tc := range resp.Message.ToolCalls {
			result := s.dispatchTool(ctx, tc)
			messages = append(messages, llm.Message{Role: llm.RoleTool, ToolCallID: tc.ID, Content: result})
		}
	}

	// Iteration cap: persist the last text + audit, then return.
	s.persistActAssistant(ctx, lastText)
	if s.cfg.Audit != nil {
		_ = s.cfg.Audit.Append(ctx, audit.Event{
			Actor: auditApp, Action: "tool_loop_cap", App: auditApp,
			Level: auditWarn, Result: auditWarn,
			Message: fmt.Sprintf("act tool loop hit %d-iteration cap", maxToolIterations),
		})
	}
	s.setStatus(status.StatusIdle)
	return lastText, nil
}

// persistActAssistant persists the final assistant text (best-effort).
func (s *Session) persistActAssistant(ctx context.Context, text string) {
	if text == "" {
		return
	}
	if err := s.persistAssistantMessage(ctx, text); err != nil {
		slog.Warn("session: persist assistant (act) failed", "err", err)
	}
}

// dispatchTool gates + executes ONE tool_call and returns the tool_result
// string to feed back to the model. The gate is non-negotiable: a denied
// action is NOT executed; the model is told it was denied.
func (s *Session) dispatchTool(ctx context.Context, tc llm.ToolCall) string {
	blast, pa, ok := toolCallToActions(tc)
	if !ok {
		return fmt.Sprintf("error: unsupported tool %q", tc.Function.Name)
	}
	// GATE: every tool dispatch flows through the Gatekeeper. This
	// drives the consent modal (WRITE/NETWORK/DESTRUCTIVE) and the
	// presence gate (N1). Evaluate returns Allow only if consent is
	// granted (or not required). Never bypass.
	decision, reason := s.cfg.Gatekeeper.Evaluate(ctx, blast)
	if s.cfg.Audit != nil {
		level, result := "info", auditAllow
		if decision != gatekeeper.Allow {
			level, result = auditWarn, auditDeny
		}
		_ = s.cfg.Audit.Append(ctx, audit.Event{
			Actor: auditApp, Action: "tool_dispatch", App: auditApp,
			Level: level, Result: result,
			// FIX B: reason can quote user-derived tool inputs.
			// Redact.
			Message: sanitize.RedactSecrets(fmt.Sprintf("tool=%s kind=%s decision=%s reason=%q", tc.Function.Name, pa.Kind, decision, reason)),
		})
	}
	if decision != gatekeeper.Allow {
		return fmt.Sprintf("DENIED by safety: %s", reason)
	}
	// Consent granted (or not required). Dispatch via the executor
	// (sanitized shell for shell.exec; CU resolver for computeruse.*).
	// Mark the action approved with the gate's allow decision; the
	// executor re-gates defense-in-depth (it trusts the stored decision
	// and does not re-prompt, so there is no double consent modal).
	pa.Status = pending.StatusApproved
	pa.GateDecision = "allow"
	res, err := s.cfg.Executor.Execute(ctx, pa)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	if res.Error != nil {
		return fmt.Sprintf("error: %v", res.Error)
	}
	return res.Result
}

// toolCallToActions maps a model tool_call to (a) the blastradius.Action
// the Gatekeeper evaluates and (b) the pending.Action the executor
// dispatches. Returns ok=false for an unrecognized tool name.
func toolCallToActions(tc llm.ToolCall) (blastradius.Action, *pending.Action, bool) {
	args := parseArgs(tc.Function.Arguments)
	blast := blastradius.Action{}
	pa := &pending.Action{}
	switch tc.Function.Name {
	case "condura_bash":
		cmd := stringArg(args, "command")
		blast = blastradius.Action{Kind: "shell.exec", Command: cmd}
		pa.Kind = "shell.exec"
		pa.Payload = pending.Payload{Command: cmd}
	case "condura_click":
		target := stringArg(args, "target")
		coords := fmt.Sprintf("%v,%v", args["x"], args["y"])
		blast = blastradius.Action{Kind: "computeruse.click", TargetApp: target, Body: coords}
		pa.Kind = "computeruse.click"
		pa.Payload = pending.Payload{Target: target, Body: coords}
	case "condura_type":
		text := stringArg(args, "text")
		blast = blastradius.Action{Kind: "computeruse.type", Body: text}
		pa.Kind = "computeruse.type"
		pa.Payload = pending.Payload{Body: text}
	case "condura_scroll":
		coords := fmt.Sprintf("%v,%v", args["dx"], args["dy"])
		blast = blastradius.Action{Kind: "computeruse.scroll", Body: coords}
		pa.Kind = "computeruse.scroll"
		pa.Payload = pending.Payload{Body: coords}
	default:
		return blast, pa, false
	}
	return blast, pa, true
}

// conduraTools returns the tool definitions offered to the model in act
// mode. Each maps cleanly to an executor Kind so dispatch is sanitized +
// gated the same way as delegate-spawned actions.
func conduraTools() []llm.ToolDefinition {
	return []llm.ToolDefinition{
		toolDef("condura_bash",
			"Run a sandboxed shell command on the user's machine. The command is validated against a binary allowlist and the user must approve dangerous operations. Use for read-only inspection, git, and build tools.",
			map[string]any{
				jsonSchemaType: jsonSchemaObject,
				jsonSchemaProperties: map[string]any{
					"command": map[string]any{
						jsonSchemaType:        jsonSchemaString,
						jsonSchemaDescription: "The shell command to run.",
					},
				},
				jsonSchemaRequired: []string{"command"},
			}),
		toolDef("condura_click",
			"Click a UI element on the screen by its accessible name (and optional coordinates).",
			map[string]any{
				jsonSchemaType: jsonSchemaObject,
				jsonSchemaProperties: map[string]any{
					"target": map[string]any{
						jsonSchemaType:        jsonSchemaString,
						jsonSchemaDescription: "Human-readable name of the element to click.",
					},
					"x": map[string]any{
						jsonSchemaType:        jsonSchemaInteger,
						jsonSchemaDescription: "Optional x coordinate.",
					},
					"y": map[string]any{
						jsonSchemaType:        jsonSchemaInteger,
						jsonSchemaDescription: "Optional y coordinate.",
					},
				},
				jsonSchemaRequired: []string{"target"},
			}),
		toolDef("condura_type",
			"Type text into the currently focused element.",
			map[string]any{
				jsonSchemaType: jsonSchemaObject,
				jsonSchemaProperties: map[string]any{
					"text": map[string]any{
						jsonSchemaType:        jsonSchemaString,
						jsonSchemaDescription: "The text to type.",
					},
				},
				jsonSchemaRequired: []string{"text"},
			}),
		toolDef("condura_scroll",
			"Scroll the focused window by dx, dy pixels.",
			map[string]any{
				jsonSchemaType: jsonSchemaObject,
				jsonSchemaProperties: map[string]any{
					"dx": map[string]any{
						jsonSchemaType:        jsonSchemaInteger,
						jsonSchemaDescription: "Horizontal scroll delta in pixels.",
					},
					"dy": map[string]any{
						jsonSchemaType:        jsonSchemaInteger,
						jsonSchemaDescription: "Vertical scroll delta in pixels.",
					},
				},
			}),
	}
}

// toolDef builds a function ToolDefinition. The Function field is an
// anonymous struct with json tags, so we assign its fields by name
// (a struct literal would need the exact tagged type).
func toolDef(name, desc string, params map[string]any) llm.ToolDefinition {
	d := llm.ToolDefinition{Type: "function"}
	d.Function.Name = name
	d.Function.Description = desc
	d.Function.Parameters = params
	return d
}

// parseArgs decodes a tool_call's JSON-encoded arguments into a map.
// Best-effort: an empty or malformed argument string yields an empty map.
func parseArgs(arguments string) map[string]any {
	out := map[string]any{}
	if arguments == "" {
		return out
	}
	_ = json.Unmarshal([]byte(arguments), &out)
	return out
}

// stringArg returns the string value of args[key], or "" if absent.
// Non-string values are stringified so numeric coordinates still flow.
func stringArg(args map[string]any, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
