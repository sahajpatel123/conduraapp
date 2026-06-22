// Package agent implements the thin agent loop for voice interactions.
//
// The agent loop takes a spoken utterance, routes it through the Gatekeeper,
// streams the response via the existing stream.Manager, optionally speaks
// the answer via TTS, and audits the entire turn.
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/conversation"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/sse"
	"github.com/sahajpatel123/synapticapp/internal/stream"
	"github.com/sahajpatel123/synapticapp/internal/voice"
)

// Loop is the thin agent loop. Dependencies are injected.
type Loop struct {
	Gatekeeper    gatekeeper.Gatekeeper
	Stream        *stream.Manager
	Broker        *sse.Broker
	ProviderName  string
	Model         string
	Speaker       voice.Speaker
	Audit         *audit.Log
	Conversations *conversation.Store
}

// AskRequest is the input to Loop.Ask.
type AskRequest struct {
	ConversationID int64
	Text           string
	RequestID      string // optional caller correlation id; stream id is returned
	Spoken         bool   // if true, speak the answer via TTS
}

// AskResult is the output of Loop.Ask.
type AskResult struct {
	RequestID string
	Reply     string
	Finish    string // "stop", "blocked", etc.
}

// Ask processes a spoken utterance through the agent loop.
func (l *Loop) Ask(ctx context.Context, req AskRequest) (AskResult, error) {
	result := AskResult{RequestID: req.RequestID}

	// Step 1: Gatekeeper check — classify the utterance as a chat action.
	action := blastradius.Action{Kind: "chat", Body: req.Text}
	decision, reason := l.Gatekeeper.Evaluate(ctx, action)

	// Audit the gatekeeper decision.
	if l.Audit != nil {
		_ = l.Audit.Append(ctx, audit.Event{
			Actor:   "user",
			Action:  "utterance",
			App:     "voice",
			Level:   "info",
			Result:  decision.String(),
			Message: fmt.Sprintf("text=%q reason=%q", req.Text, reason),
		})
	}

	if decision != gatekeeper.Allow {
		result.Finish = "blocked"
		// Audit the block.
		if l.Audit != nil {
			_ = l.Audit.Append(ctx, audit.Event{
				Actor:   "agent",
				Action:  "blocked",
				App:     "voice",
				Level:   "warn",
				Result:  "deny",
				Message: fmt.Sprintf("text=%q reason=%q class=%s", req.Text, reason, blastradius.Classify(action)),
			})
		}
		return result, nil
	}

	// Step 2: Append user message to conversation.
	if l.Conversations != nil {
		msg := conversation.Message{
			Role:    "user",
			Content: req.Text,
		}
		if err := l.Conversations.Append(ctx, req.ConversationID, msg); err != nil {
			return result, fmt.Errorf("append user message: %w", err)
		}
	}

	// Step 3: Stream the response via the shared stream manager.
	if l.Stream == nil {
		return result, errors.New("agent: stream manager not configured")
	}
	if l.Broker == nil {
		return result, errors.New("agent: SSE broker not configured")
	}
	if l.ProviderName == "" {
		return result, errors.New("agent: provider not configured")
	}
	if l.Model == "" {
		return result, llm.ErrNoModel
	}

	messages, err := l.buildMessages(ctx, req)
	if err != nil {
		return result, fmt.Errorf("build messages: %w", err)
	}

	streamReq := stream.Request{
		ProviderName:   l.ProviderName,
		ConversationID: req.ConversationID,
		Chat: llm.ChatRequest{
			Model:    l.Model,
			Messages: messages,
			Stream:   true,
		},
	}

	// Subscribe before Start so we do not miss early deltas.
	sub := l.Broker.Subscribe()
	defer l.Broker.Unsubscribe(sub)

	requestID, err := l.Stream.Start(ctx, streamReq)
	if err != nil {
		return result, fmt.Errorf("stream start: %w", err)
	}
	result.RequestID = requestID

	full, err := l.collectStream(ctx, requestID, sub)
	if err != nil {
		return result, err
	}
	result.Reply = full

	// Persist the assistant reply for multi-turn context.
	if full != "" && l.Conversations != nil && req.ConversationID != 0 {
		if persistErr := l.Conversations.Append(ctx, req.ConversationID, conversation.Message{
			Role:    "assistant",
			Content: full,
		}); persistErr != nil {
			return result, fmt.Errorf("append assistant message: %w", persistErr)
		}
	}

	if l.Audit != nil && full != "" {
		_ = l.Audit.Append(ctx, audit.Event{
			Actor:   "agent",
			Action:  "reply",
			App:     "voice",
			Level:   "info",
			Result:  "allow",
			Message: fmt.Sprintf("conversation_id=%d reply_len=%d", req.ConversationID, len(full)),
		})
	}

	// Step 4: Speak the answer if requested.
	if req.Spoken && l.Speaker != nil && full != "" {
		if err := l.Speaker.Speak(ctx, full); err != nil {
			return result, fmt.Errorf("speak response: %w", err)
		}
	}

	result.Finish = "stop"
	return result, nil
}

// buildMessages assembles chat history for the LLM call. The user
// utterance is expected to have been persisted in step 2.
func (l *Loop) buildMessages(ctx context.Context, req AskRequest) ([]llm.Message, error) {
	const historyLimit = 20
	if l.Conversations == nil || req.ConversationID == 0 {
		return []llm.Message{
			{Role: llm.RoleUser, Content: req.Text},
		}, nil
	}

	history, err := l.Conversations.GetRecentMessages(ctx, req.ConversationID, historyLimit)
	if err != nil {
		return []llm.Message{{Role: llm.RoleUser, Content: req.Text}}, err
	}

	out := make([]llm.Message, 0, len(history))
	for _, m := range history {
		out = append(out, llm.Message{
			Role:       llm.Role(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		})
	}
	if len(out) == 0 || out[len(out)-1].Content != req.Text || out[len(out)-1].Role != llm.RoleUser {
		out = append(out, llm.Message{Role: llm.RoleUser, Content: req.Text})
	}
	return out, nil
}

// collectStream accumulates stream.delta events for requestID from sub.
func (l *Loop) collectStream(ctx context.Context, requestID string, sub *sse.Subscription) (string, error) {
	const streamBudget = 60 * time.Second
	timer := time.NewTimer(streamBudget)
	defer timer.Stop()

	var full string
	finished := false

	for !finished {
		select {
		case <-ctx.Done():
			return full, ctx.Err()
		case <-timer.C:
			return full, fmt.Errorf("agent: stream timed out after %s", streamBudget)
		case <-sub.Done:
			return full, errors.New("agent: broker channel closed")
		case ev := <-sub.Events:
			if !eventMatchesRequest(ev, requestID) {
				continue
			}
			switch ev.Name {
			case stream.EventDelta:
				if delta, ok := stringField(ev.Data, "delta"); ok {
					full += delta
				}
			case stream.EventFinished:
				finished = true
			case stream.EventError, stream.EventCancelled:
				if msg, ok := stringField(ev.Data, "error"); ok {
					return full, fmt.Errorf("agent: stream %s: %s", ev.Name, msg)
				}
				return full, fmt.Errorf("agent: stream %s", ev.Name)
			}
		}
	}
	return full, nil
}

func eventMatchesRequest(ev sse.Event, requestID string) bool {
	if ev.Data == nil {
		return true
	}
	data, err := json.Marshal(ev.Data)
	if err != nil {
		return true
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		return true
	}
	if id, ok := fields["request_id"].(string); ok {
		return id == requestID
	}
	return true
}

func stringField(data any, key string) (string, bool) {
	raw, err := json.Marshal(data)
	if err != nil {
		return "", false
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		return "", false
	}
	v, ok := fields[key].(string)
	return v, ok
}

// Cancel cancels the underlying stream and stops TTS.
func (l *Loop) Cancel(requestID string) {
	if l.Stream != nil {
		_ = l.Stream.Cancel(requestID)
	}
	if l.Speaker != nil {
		l.Speaker.Stop()
	}
}
