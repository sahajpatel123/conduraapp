# Architecture 01 — The Hybrid-with-Memory Router

> How Condura picks the right model or CLI for each sub-task.

---

## Goal

A user prompt is rarely a single task. "Plan a trip to Tokyo" decomposes into research, summarization, code (for a script), and image generation. A single LLM can't do all of these optimally, and even if it could, **the user has subscriptions they want to use**.

The router picks the cheapest, fastest, highest-quality model for each sub-task, using the user's configured priorities.

---

## The Hybrid

"Hybrid-with-memory" means:

- **Model-class routing** (router decides, not user)
- **Memory-aware** (the user model's preferences inform routing)
- **Fallback chain** (if primary fails, try next)
- **Cost & latency aware**
- **Honest about what it knows and doesn't**

---

## What Gets Routed

A `TaskSpec` flows through the router. It has:

```go
type TaskSpec struct {
    Goal           string        // e.g., "summarize this article"
    Context        []Message     // prior conversation
    SubTaskType    SubTaskType   // research, code, image, vision, voice, long-context, etc.
    Constraints    Constraints   // latency, cost, quality, privacy
    RequiredCaps   []Capability  // function-calling, vision-input, json-mode, ...
    UserOverrides  *ProviderChoice
    Memory         MemoryHints   // the Adaptive Engine's hints
}

type SubTaskType string

const (
    SubResearch        SubTaskType = "research"
    SubCode            SubTaskType = "code"
    SubReasoning       SubTaskType = "reasoning"
    SubLongContext     SubTaskType = "long-context"
    SubVision          SubTaskType = "vision"
    SubImageGen        SubTaskType = "image-generation"
    SubTTS             SubTaskType = "text-to-speech"
    SubSTT             SubTaskType = "speech-to-text"
    SubEmbedding       SubTaskType = "embedding"
    SubChat            SubTaskType = "chat"
    SubToolUse         SubTaskType = "tool-use"
    SubCommand         SubTaskType = "command"
    SubBrowser         SubTaskType = "browser-control"
)
```

The router is **deterministic** given the same inputs. The LLM never gets a vote in the routing decision — only in the answer.

---

## Routing Algorithm

```
function Route(spec: TaskSpec) -> Plan:
    plan = empty
    
    # Step 1: Classify sub-task
    subTask = classifySubTask(spec)
    # (deterministic classifier; not LLM)
    
    # Step 2: Filter candidates by capability
    candidates = allProviders ∪ allCLIs
    candidates = filterByCapability(candidates, spec.RequiredCaps)
    
    # Step 3: Apply user priority order
    candidates = applyUserPriority(candidates, spec.Memory.Preferences)
    
    # Step 4: Apply cost/latency constraints
    candidates = filterByConstraints(candidates, spec.Constraints)
    
    # Step 5: Apply fallback chain
    candidates = buildFallbackChain(candidates, spec.Memory.Reliability)
    
    # Step 6: Build plan
    plan.Primary = candidates[0]
    plan.Fallbacks = candidates[1:]
    plan.EstimatedCost = estimateCost(plan, spec)
    plan.EstimatedLatency = estimateLatency(plan, spec)
    
    return plan
```

---

## What the User Can Configure

### Per-Sub-Task Priority (from settings)

```yaml
router:
  priorities:
    chat:        [claude_code, chatgpt, gemini_cli, ollama, openrouter, openai, anthropic, groq, custom]
    code:        [claude_code, codex, antigravity, ollama, openrouter, anthropic, openai, custom]
    research:    [claude_code, hermes, gemini, openrouter, openai, anthropic, custom]
    reasoning:   [claude_code, chatgpt, antigravity, ollama, openrouter, anthropic, custom]
    long-context:[gemini, ollama, openrouter, anthropic, custom]
    vision:      [claude_code, openai, antigravity, gemini, ollama, openrouter, anthropic, custom]
    image-gen:   [openai, antigravity, openrouter, custom]
    tts:         [openai, elevenlabs, custom]
    stt:         [whisper_local, openai, custom]
    embedding:   [local, openai, ollama, custom]
    tool-use:    [claude_code, codex, antigravity, openrouter, anthropic, openai, custom]
    command:     [claude_code, codex, openrouter, anthropic, openai, custom]
    browser:     [claude_code, codex, antigravity, openrouter, anthropic, openai, custom]
```

The user drags items to reorder. This is one of the most-used settings.

### Hard Overrides Per Request

The user can pin a model for one request:

- "Hey Condura, ask `claude-opus-4-7` for that one"
- "Use local Ollama for this — I'm offline"
- "Try `codex` first for code tasks"
- A UI selector in the composer.

### Trust Tiers

Each provider has a trust tier:

```yaml
router:
  trust:
    anthropic:    high
    openai:       high
    google:       high
    xai:          medium
    mistral:      medium
    openrouter:   medium
    together:     medium
    groq:         medium
    fireworks:    medium
    local:        high
    custom:       low
```

**High-trust providers** can be auto-routed to without prompting.
**Medium-trust** can be auto-routed, but the user sees a small badge in the UI.
**Low-trust** always asks: "Send this to `custom`?" before each request.

---

## The Classifier (Deterministic)

The router uses a **rule + embedding** classifier, NOT an LLM, to determine sub-task type:

1. **Keyword scan**: "code", "function", "implement", "compile" → `code`. "Image", "picture", "draw" → `image-gen`. Etc.
2. **Embedding similarity**: embed the prompt, compare to reference embeddings for each sub-task type.
3. **Tool-call analysis**: if the prompt requires specific tools, that informs the type.
4. **Memory hints**: "Last time you asked for this kind of thing, you used X."

The classifier is a tiny model (or rule-based) — it's <10ms and runs locally. We **never** use an LLM to decide which LLM to use. That's circular and adds latency.

---

## Fallback Logic

When the primary fails:

```
try primary
on network-error, rate-limit, model-error, timeout:
  try fallback 1
  on similar:
    try fallback 2
    ...
  on all-failed:
    surface user-friendly error with diagnosis
```

The user's "fallback chain" determines the order. **Reliability is tracked** in the Adaptive Engine: if `codex` fails 3 times in a row today, the router demotes it for today.

---

## Local-First & Offline

- If the user is offline, the router **only** considers local models (Ollama, llama.cpp, vLLM, LM Studio).
- If no local model is configured, the router surfaces: "You're offline. Install a local model or connect to the internet."
- Local models are checked for availability, RAM, and licensing before routing to them.

---

## Streaming & Cancellation

- The router streams the primary response by default.
- The user can press `Esc` to cancel a generation — the router kills the stream, refunds tokens if the provider supports it, and falls back if appropriate.
- The router surfaces "Generating with X, Y seconds in, $Z so far" in the UI as it streams.

---

## Spend & Rate-Limit Awareness

The router maintains per-provider:

- Spend rate (rolling 1h, 24h, 7d, 30d)
- Rate limit status (RPM, TPM, RPD, TPD)
- Last error
- Average latency over last 100 calls
- Reliability score (success rate over last 100 calls)

This is shown in the **Status** tab of the overlay:

```
Status → Router
├── Anthropic: 3,452 tokens/min, $4.21 today, 99% success
├── OpenAI:    1,200 tokens/min, $1.87 today, 98% success
├── Gemini:    800 tokens/min, $0.10 today, 100% success
├── Local Ollama: N/A, 0% failure, free
└── Last routing decision: "research → claude_code (priority 1, healthy)"
```

---

## What the Router Doesn't Do

- **Doesn't talk to the user.** The router is silent. The Strategist LLM talks to the user.
- **Doesn't decide policy.** Safety is the Gatekeeper's job.
- **Doesn't call tools directly.** The Agent Loop orchestrates tool calls; the router just picks the model.
- **Doesn't remember conversations.** Memory is its own subsystem.

---

## Why Hybrid-and-Memory (Not "Just Use Claude")

- **Cost**: routing a long-context summary to Gemini Pro 1.5 is ~10x cheaper than to Claude Opus.
- **Quality**: the best model for code is not the best for image-gen.
- **User freedom**: users have subscriptions and want to use them.
- **Resilience**: if Anthropic has an outage, the user isn't stuck.
- **Local-first**: with Ollama, the user can route anything to local.
- **Learning**: the Adaptive Engine's reliability scores inform routing.

---

## Related Docs

- [00-overview.md](00-overview.md) — The conductor pattern
- [02-computer-use.md](02-computer-use.md) — 4-tier computer use
- [05-adaptive.md](05-adaptive.md) — How the Adaptive Engine informs routing
- [09-ipc.md](09-ipc.md) — JSON-RPC method: `router.plan`
