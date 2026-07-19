// Package agent implements the thin agent loop for voice
// interactions.
//
// The agent is a thin coordination layer around three
// subsystems:
//
//   - The LLM planner (internal/llm) — turns a user
//     utterance into a tool-call plan.
//   - The Gatekeeper (internal/gatekeeper) — decides which
//     tool calls are allowed, blocks risky ones.
//   - The stream manager (internal/stream) — publishes
//     events to the SSE broker for the GUI.
//
// The loop itself is small. Most of the value lives in the
// subsystems. The agent's job is the wiring: hand the user
// utterance to the planner, get a plan, gate it, dispatch
// each tool call through the appropriate executor, stream
// the result, audit the turn.
//
// Files in this package:
//
//   - agent.go: Agent — the loop, the public API.
//   - gated_executor.go: the executor wrapper that runs
//     each tool call through the Gatekeeper before doing
//     it. Most of the "what does the agent actually do"
//     logic lives here.
//   - computer_use_executor.go: the computer-use specific
//     executor (screenshots, AX reads, click/type). Built
//     on internal/computeruse; gated by the Gatekeeper.
//   - culool.go: a small helper for the computer-use loop
//     (take-screenshot, ask-the-LLM-what-to-do, dispatch).
//
// The agent itself is local-only. The LLM provider, the
// blast-radius classifier, the Gatekeeper, and the stream
// manager are all separate packages; this file just composes
// them. Adding a new tool type means adding a new executor
// in a sibling file — the loop code itself rarely changes.
package agent
