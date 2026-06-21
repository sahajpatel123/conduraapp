# v0.2.0 Roadmap — What the v0.1.0 Marketing Promises But Doesn't Ship

> **Audience:** the user, the next AI agent, the marketing owner (KIMI K2.6),
> the QA verifier for Phase 15. The goal of this doc is to be a single source
> of truth for "what we said we'd ship vs. what we did" so the v0.2.0 release
> scope is unambiguous.

## How to read this

- **Code status column** — what the v0.1.0 binary actually does today.
- **v0.2.0 plan** — the smallest change that would make the marketing claim
  honest (without rewriting the marketing).
- **Marketing TODO** — the line in the marketing site (Next.js, `web/`) that
  needs updating until v0.2.0 ships.

## Backend work

### 1. Subscription OAuth (the biggest missing piece)

| Provider | Status | v0.2.0 plan | Marketing TODO |
|---|---|---|---|
| ChatGPT Plus / Pro | ❌ Not implemented | `internal/subscription/openai.go` — PKCE flow against `auth.openai.com`, refresh-token rotation, model routing based on Plus vs Pro. | Remove "no new monthly bill" claim for OpenAI models. |
| Claude Pro / Max | ❌ Not implemented | `internal/subscription/anthropic.go` — OAuth via Anthropic's authorization server (they're rolling one out 2026). | Same as above. |
| SuperGrok (xAI) | ❌ Not implemented | `internal/subscription/xai.go`. | Same. |
| Codex via ChatGPT Plus | ❌ Not implemented | The `codex` CLI supports OAuth for Plus users; `internal/delegation/codex.go` needs to invoke it. | Same. |

**Estimated effort:** 2-3 weeks. The PKCE mechanics are known; the
unknowns are rate-limit handling and how each provider's
subscription maps to which models.

### 2. Selective Perception — wire the live event source

| Component | Status | v0.2.0 plan |
|---|---|---|
| `internal/perception` package | ✅ Built in Phase 14I | Done. |
| DirtyTracker event source on macOS | ❌ Platform event source not wired | v0.2.0: add `internal/perception/dirty_darwin.go` that hooks `CGEventTap` + `NSWindowDidUpdateNotification`, calling `tracker.Mark(bundleID)`. |
| DirtyTracker on Windows / Linux | ❌ | v0.2.0: `EVENT_OBJECT_LOCATIONCHANGE` (Windows) and `object:state-changed:defunct` (Linux AT-SPI). |
| Energy budget driving actual choices | ✅ SmartCapturer | Done. The capturer picks the cheapest strategy. |
| Per-frame PII redaction in the live pipeline | ❌ PIIRedactor is a function, not in the path | v0.2.0: every screen-text output that crosses an IPC boundary goes through `PIIRedactor.Redact`. |

### 3. Network kill switch — hard Layer 3

| Component | Status | v0.2.0 plan |
|---|---|---|
| `halt.NetworkGuard` interface | ✅ Built | Done. |
| In-process `InProcessGuard` (Layer 3 "soft") | ✅ Built | Done. |
| Real `pf` / `netsh` separate process (Layer 3 "hard") | ❌ | v0.2.0: ship a small companion binary `condura-guard` that holds the OS firewall rules and is started/stopped independently. v0.1.0's InProcessGuard is the v0.2.0 fallback path. |

### 4. Execution waves / DAG scheduler

| Component | Status | v0.2.0 plan |
|---|---|---|
| `internal/delegation.GatedRunner` | ✅ Spawns individual sub-agents | Done. |
| Wave/DAG decomposition | ❌ Marketing claims "Wave 1 → 3 parallel agents spawned" but the runner doesn't decompose. | v0.2.0: add `internal/delegation/wave.go` with `Wave`, `DAG`, and a wave-scheduler that uses the existing semaphore. |
| CE-MCP (code-execution delegation) | ❌ | v0.2.0+ — research says ~70% token reduction; the work is non-trivial and not in v0.1.0's hot path. |

### 5. MCP UI

| Component | Status | v0.2.0 plan |
|---|---|---|
| `internal/mcp` | ✅ RPCs exist | Done. |
| `Mcp.svelte` GUI | ❌ | v0.2.0: server browser, one-click install, OAuth flow, tool list. |

### 6. Channels — real Signal / WhatsApp / iMessage

| Channel | Status | v0.2.0 plan |
|---|---|---|
| Telegram | ✅ | Done. |
| WhatsApp | ❌ Stub returns "v0.2.0" | v0.2.0: integrate `whatsmeow` (Go) or the official Business API. |
| Signal | ❌ Stub | v0.2.0: integrate `libsignal` via the `signal-cli-rest-api` sidecar. |
| iMessage | ⚠️ `imessage_darwin.go` exists for send; receive not implemented | v0.2.0: AppleScript or `applesimutils` integration. |

### 7. Skills Hub + Web Dashboard

| Component | Status | v0.2.0 plan |
|---|---|---|
| `internal/hub` client/server | ✅ | Done. |
| `hub.condura.app` public Next.js app | ❌ Not in repo | v0.2.0: ship the `hub/` Next.js app (curation, moderation, OAuth, Vercel deploy). |
| `condura.app/dashboard` web dashboard | ❌ Not in repo | v0.2.0: same Next.js workspace, separate route. |

### 8. On-device verification

| Action | Status | v0.2.0 plan |
|---|---|---|
| `docs/phase15-verification.md` sign-off | ❌ All empty | Phase 15: requires physical macOS / Windows / Linux machines (not VMs with broken TCC) and ~2-3 days of human time per OS. |

### 9. Computer-use action indicators in chat

| Component | Status | v0.2.0 plan |
|---|---|---|
| CU events on SSE | ✅ | Done. |
| Live "agent is clicking X" indicator in chat | ❌ No UI binding | v0.2.0: bind `cu.action` SSE events to a chat-side widget. |

## Marketing copy that needs updating

> **Owner:** KIMI K2.6 (the marketing site rebuild agent).
> **Files in `web/`** (Next.js marketing site). Do not touch from
> the daemon agent — KIMI owns the marketing site.

| Claim | File:line | Action |
|---|---|---|
| "12+ LLM providers, including ChatGPT Plus / Claude Pro / SuperGrok" | `web/app/ecosystem/page.tsx:30-42` | After v0.2.0 OAuth ships, restore. Until then, remove the subscription column or label as "v0.2.0". |
| "A model. Fallible. A gatekeeper. Deterministic." — the security page shows 4 kill layers | `web/app/security/page.tsx` | Layer 3 description should say "in-process network guard" not "separate OS process" until v0.2.0 ships the real guard. |
| Ecosystem detection: hardcoded `/usr/local/bin/claude` etc. | `web/app/ecosystem/page.tsx:45-53` | Replace with `onboarding.probe_power` data once the dashboard is wired in v0.2.0. |
| Hub download CTA at `hub.synaptic.app` / `hub.condura.app` | `web/lib/site.ts` | The 404 should be replaced with "coming v0.2.0 — see on-device Skills" until the Hub ships. **Resolved in Phase 16**: daemon + network guard + config defaults now point to `hub.condura.app`. Web marketing text remains for KIMI K2.6 to update. |
| Demo video | — | Required for v0.1.0 public launch per CLAUDE.md §26. 60-second screen capture, no script. |
| Discord URL `discord.gg/synaptic` | `web/lib/site.ts:13` | Rename to `discord.gg/condura` (or whatever the actual channel is). |
| Open Collective URL `opencollective.com/synaptic` | `README.md` | Same. |
| PLATFORMS list in `site.ts` | `web/lib/site.ts` | Use `condura-gui-*.dmg` / `*.exe` / `*.AppImage`, not the legacy `synaptic.*` names. |
| "Adaptive engine learns you" | `web/app/` (any page) | The backend works; the marketing copy oversells the depth. Tone down to "learns from your interactions, fully editable in Settings." |

## Spec debt tracked in CLAUDE.md §33.5

See CLAUDE.md §33.5.2 for the full list of items deferred to
v0.2.0+. This roadmap doc focuses on the *delivery* plan; the
CLAUDE.md section is the *status* table.

## Sequencing

Suggested v0.2.0 milestone order (each is independently
shippable):

1. **Backend: hard Layer 3** (companion binary + pf/netsh
   rules). 2 weeks. Highest safety value.
2. **Backend: Selective Perception event source** (CGEventTap
   on macOS; AT-SPI on Linux; UI Automation events on
   Windows). 2 weeks.
3. **GUI: live CU indicators in chat** + **GUI: MCP browser**
   + **GUI: Signal/WhatsApp/iMessage status** if backend ships
   in same milestone. 2 weeks.
4. **Backend: subscription OAuth** (3 providers). 3 weeks.
   Marketing can flip the switch once the GUI sees the new
   provider class.
5. **Public Skills Hub** (`hub.condura.app` Vercel deploy).
   2 weeks.
6. **Wave scheduler / DAG executor** (the most invasive change
   to the agent loop). 3 weeks.
7. **On-device verification** (Phase 15 close-out on the
   v0.2.0 binary). 1 week.
8. **Marketing re-pass** to restore the v0.1.0-fictional
   claims. 1 week.

## Open questions for the user

1. **Domain:** `hub.condura.app` is in `internal/halt/network.go`'s
   default allow-list, but the v0.1.0 build points to
   `hub.synaptic.app` (locked decision #18). When the user
   migrates to v0.2.0, which do we keep? Recommend
   `hub.condura.app` as the canonical, `hub.synaptic.app` as
   the legacy redirect (already in the allow-list).
2. **Discord / Open Collective:** the user has chosen the
   product name as `condura` but the social URLs in the spec
   are still `synaptic`. We need the canonical URLs.
3. **v0.1.0 release date:** if the user plans to ship v0.1.0
   before the marketing copy is updated to v0.2.0 reality, the
   public launch will carry fictional claims. Recommend
   delaying public launch until (a) the marketing copy is
   aligned and (b) on-device verification on at least one
   clean machine per OS is signed off.

## What is NOT in v0.2.0 scope

These are intentionally out (in v0.3.0+):

- Cloud sync (P2P only)
- Native mobile (iOS, Android)
- Web-based agent (no headless browser takeover)
- Multi-tenant teams / workspaces
- Enterprise SSO / SCIM
- AI-generated skills
- Autonomous skill creation without user review
- Honcho-style user-model export to other agents
- Public release of CE-MCP (it stays internal)

The line in the sand: **v0.1.0 is "the conductor for your own
machine". v0.2.0 is "the conductor for your stack, with
subscription billing handled for you."** v0.3.0+ is "the
conductor for your team."
