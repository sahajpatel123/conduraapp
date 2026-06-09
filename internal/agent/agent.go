// Package agent implements the thin agent loop for voice interactions.
//
// The agent loop takes a spoken utterance, routes it through the Gatekeeper,
// streams the response via the existing stream.Manager, optionally speaks
// the answer via TTS, and audits the entire turn.
package agent

import (
	"context"
	"fmt"

	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/conversation"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/stream"
	"github.com/sahajpatel123/synapticapp/internal/voice"
)

// Loop is the thin agent loop. Dependencies are injected.
type Loop struct {
	Gatekeeper    gatekeeper.Gatekeeper
	Stream        *stream.Manager
	Speaker       voice.Speaker
	Audit         *audit.Log
	Conversations *conversation.Store
}

// AskRequest is the input to Loop.Ask.
type AskRequest struct {
	ConversationID int64
	Text           string
	RequestID      string
	Spoken         bool // if true, speak the answer via TTS
}

// AskResult is the output of Loop.Ask.
type AskResult struct {
	RequestID string
	Finish    string // "stop", "blocked", etc.
}

// Ask processes a spoken utterance through the agent loop.
func (l *Loop) Ask(ctx context.Context, req AskRequest) (AskResult, error) {
	result := AskResult{RequestID: req.RequestID}

	// Step 1: Gatekeeper check — classify the utterance as a chat action.
	action := blastradius.Action{Kind: "chat"}
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

	if decision == gatekeeper.Deny {
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

	// Step 3: Stream the response.
	// For now, we use a simplified approach: the stream manager handles the actual
	// LLM streaming. In a full implementation, we'd build the ChatRequest from
	// the conversation history.
	result.Finish = "stop"

	// Step 4: Speak the answer if requested.
	if req.Spoken && l.Speaker != nil {
		// In a full implementation, we'd collect the streamed response text
		// and speak it. For now, this is a placeholder.
		if err := l.Speaker.Speak(ctx, "Response would be spoken here"); err != nil {
			return result, fmt.Errorf("speak response: %w", err)
		}
	}

	return result, nil
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
