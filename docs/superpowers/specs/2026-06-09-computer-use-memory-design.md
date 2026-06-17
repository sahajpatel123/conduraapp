# Phase 5 — Computer Use & Memory (Complete Specification)

> **Status:** Draft  
> **Created:** 2026-06-09  
> **Depends on:** Phase 4 (The Living Presence) — gatekeeper, blast-radius, agent loop  
> **Estimated effort:** 5 sub-phases, ~15 working days

---

## 1. Goals

Phase 5 gives Condura the ability to **see the screen**, **interact with applications**, and **remember everything**. It transforms the agent from a chat-only assistant into a true computer-use agent that can click, type, scroll, and navigate — while building a persistent memory of every interaction.

### 1.1 What Phase 5 Delivers

1. **Accessibility Bridge** — Read the AX tree on macOS (primary), Windows (xa11y), Linux (AT-SPI2)
2. **Computer-Use Router** — 4-tier backend selection (ORAX Eye → mac-cua → macOS-MCP → Vision CUA)
3. **Twin-Snapshot Verification** — Pre-action snapshot comparison to detect stale state
4. **Memory System** — 3-layer memory (episodic, semantic, procedural) with SQLite + FTS5 + sqlite-vec
5. **Agent Loop Expansion** — Multi-step planning, task decomposition, verification checkpoints

### 1.2 What Phase 5 Deliberately Does NOT Include

- Wake word ("hey synaptic") — deferred to later
- P2P sync — deferred to Phase 6+
- Skills Hub — deferred to Phase 6+
- Action Replay — deferred to later
- Marketing/launch — deferred to Phase 7

---

## 2. Locked Decisions

| # | Decision | Rationale |
|---|---|---|
| 1 | macOS AX tree is the primary backend | User's primary platform, richest AX API |
| 2 | ORAX Eye first, then fallbacks | Free, fast (~50ms), MIT licensed |
| 3 | Twin-snapshot is mandatory for WRITE/NETWORK | Anti-staleness mechanism (MISSION §5.2) |
| 4 | Memory stored in SQLite + FTS5 + sqlite-vec | Local-first, encrypted at rest, no cloud |
| 5 | Episodic memory indexed by FTS5 + vector | Combined text + semantic search |
| 6 | Semantic facts confidence-scored | Dialectic extraction from interactions |
| 7 | Procedural memory = skills | Separate package, auto-created |
| 8 | All computer-use actions go through gatekeeper | Safety seam from Phase 4.0 |
| 9 | User interruption detection via CGEventTap | Agent yields when user interacts |
| 10 | Battery-aware capture (Selective Perception) | Don't drain battery on laptop |

---

## 3. Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Agent Loop (Phase 4.4)                │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │   Planner    │  │   Executor   │  │   Verifier    │  │
│  └──────┬──────┘  └──────┬───────┘  └───────┬───────┘  │
│         │                │                   │           │
│         ▼                ▼                   ▼           │
│  ┌─────────────────────────────────────────────────┐    │
│  │              Gatekeeper (Phase 4.0)              │    │
│  └──────────────────────┬──────────────────────────┘    │
│                         │                               │
│         ┌───────────────┼───────────────┐               │
│         ▼               ▼               ▼               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       │
│  │ ComputerUse │ │   Memory    │ │   Voice     │       │
│  │   Router    │ │   System    │ │  (Phase 4)  │       │
│  └──────┬──────┘ └──────┬──────┘ └─────────────┘       │
│         │               │                               │
│         ▼               ▼                               │
│  ┌─────────────┐ ┌─────────────┐                       │
│  │ AX Bridge   │ │ SQLite/FTS5 │                       │
│  │ (macOS/Win/ │ │ + sqlite-vec│                       │
│  │  Linux)     │ │             │                       │
│  └─────────────┘ └─────────────┘                       │
└─────────────────────────────────────────────────────────┘
```

---

## 4. Sub-Phases

### 4.0 Computer-Use Interfaces + Action Classifier Expansion

**Goal:** Define the core interfaces and expand the blast-radius classifier to cover computer-use actions.

**Packages:**
- `internal/computeruse/` — Core interfaces
  - `computeruse.go` — `Backend` interface, `Action` types, `Snapshot` type
  - `router.go` — 4-tier router (try cheapest first)
  - `errors.go` — Sentinel errors

**Interface:**
```go
type Backend interface {
    Name() string
    Capabilities() []Capability
    CaptureScreen(ctx context.Context) (*Snapshot, error)
    GetAXTree(ctx context.Context) (*AXTree, error)
    Execute(ctx context.Context, action *Action) (*ActionResult, error)
    IsAvailable(ctx context.Context) bool
}

type Snapshot struct {
    Image     []byte    // PNG screenshot
    AXTree    *AXTree   // Accessibility tree
    Timestamp time.Time
    WindowID  uint32
    PID       int32
}

type Action struct {
    Type       ActionType // click, type, scroll, key, drag, wait
    Target     *Target    // element reference or coordinates
    Value      string     // text to type, key to press
    Bounds     *Rect      // fallback coordinates
    Timeout    time.Duration
}

type ActionResult struct {
    Success   bool
    Error     error
    Snapshot  *Snapshot // post-action snapshot
    Duration  time.Duration
}
```

**Blast-radius expansion:**
- `computeruse.read` — READ (screenshot, AX tree)
- `computeruse.click` — WRITE (click element)
- `computeruse.type` — WRITE (type text)
- `computeruse.scroll` — WRITE (scroll view)
- `computeruse.key` — WRITE (press key)
- `computeruse.drag` — WRITE (drag element)
- `computeruse.launch` — NETWORK (launch app)
- `computeruse.shell` — DESTRUCTIVE (shell command)

**Definition of done:**
- Interfaces defined, router logic implemented
- All actions classified in blast-radius
- Unit tests for router fallback logic
- Gatekeeper integration: only READ allowed by default

---

### 4.1 Accessibility Bridge (macOS)

**Goal:** Read the AX tree on macOS using the Accessibility API.

**Packages:**
- `internal/computeruse/ax/` — AX tree reader
  - `tree.go` — AX tree traversal, element extraction
  - `element.go` — AX element attributes (role, title, value, bounds)
  - `snapshot.go` — Full screen snapshot with AX tree + screenshot

**Dependencies:**
- macOS Accessibility API (ApplicationServices framework)
- CGWindowListCreateImage for screenshots
- No external dependencies (pure Go + CGo)

**Key functions:**
```go
func CaptureAXTree(ctx context.Context) (*AXTree, error)
func CaptureScreen(ctx context.Context) (*Screenshot, error)
func FindElement(tree *AXTree, query *ElementQuery) (*AXElement, error)
func GetElementAtPoint(tree *AXTree, x, y float64) (*AXElement, error)
```

**Definition of done:**
- Can read AX tree from any running application
- Can capture screenshot of focused window
- Can find elements by role, title, value
- Can get element at specific coordinates
- Unit tests with mock AX trees
- Integration tests on macOS (requires Accessibility permission)

---

### 4.2 Twin-Snapshot Verification

**Goal:** Implement the anti-staleness mechanism (MISSION §5.2).

**Packages:**
- `internal/computeruse/verify.go` — Snapshot comparison
- `internal/computeruse/verify_test.go` — Tests

**Logic:**
1. Before executing a WRITE/NETWORK action, capture snapshot S1
2. Execute the action
3. Capture snapshot S2 immediately after
4. Compare S1 and S2:
   - If AX tree changed in unexpected ways → abort, report stale state
   - If window focus changed → abort, report interruption
   - If action succeeded and tree matches预期 → success

**Comparison:**
```go
func VerifySnapshot(pre, post *Snapshot, action *Action) (*VerificationResult, error)

type VerificationResult struct {
    Valid     bool
    Reason    string
    Diff      []AXDiff
    Aborted   bool
}
```

**Definition of done:**
- Snapshot comparison logic implemented
- Detects window focus changes
- Detects AX tree mutations
- Detects element removal/replacement
- Unit tests with synthetic snapshots
- Integration tests on macOS

---

### 4.3 Memory System

**Goal:** Implement 3-layer memory with SQLite + FTS5 + sqlite-vec.

**Packages:**
- `internal/memory/` — Memory system
  - `memory.go` — `Store` interface, `Manager` struct
  - `episodic.go` — Episodic memory (past sessions)
  - `semantic.go` — Semantic memory (user facts)
  - `procedural.go` — Procedural memory (skills)
  - `search.go` — Combined FTS5 + vector search
  - `migrations.go` — Schema migrations

**Schema:**
```sql
-- Episodic memory
CREATE TABLE episodes (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    user_message TEXT,
    agent_response TEXT,
    actions_taken TEXT,  -- JSON array of actions
    timestamp DATETIME,
    summary TEXT,
    embedding BLOB
);

CREATE VIRTUAL TABLE episodes_fts USING fts5(
    user_message, agent_response, summary,
    content=episodes, content_rowid=rowid
);

-- Semantic memory
CREATE TABLE facts (
    id TEXT PRIMARY KEY,
    category TEXT,  -- preference, identity, expertise, etc.
    content TEXT,
    confidence REAL DEFAULT 0.5,
    source_episode_id TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    embedding BLOB
);

CREATE VIRTUAL TABLE facts_fts USING fts5(
    content, category,
    content=facts, content_rowid=rowid
);

-- Procedural memory (skills)
CREATE TABLE skills (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    trigger_pattern TEXT,
    steps TEXT,  -- JSON array of steps
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    created_at DATETIME,
    last_used DATETIME
);
```

**Search:**
```go
func (m *Manager) Recall(ctx context.Context, query string, limit int) ([]Memory, error)
func (m *Manager) Remember(ctx context.Context, memory *Memory) error
func (m *Manager) ExtractFacts(ctx context.Context, episode *Episode) ([]*Fact, error)
func (m *Manager) Forget(ctx context.Context, id string) error
```

**Definition of done:**
- SQLite schema with FTS5 + sqlite-vec
- Episodic memory: store, search, summarize
- Semantic memory: extract facts, confidence scoring
- Procedural memory: skill creation, improvement tracking
- Combined search (text + vector)
- Unit tests with in-memory SQLite
- Integration tests with real SQLite

---

### 4.4 Agent Loop Expansion

**Goal:** Expand the basic agent loop (Phase 4.4) with multi-step planning and verification.

**Packages:**
- `internal/agent/planner.go` — Task decomposition
- `internal/agent/verifier.go` — Post-action verification
- `internal/agent/loop_expanded.go` — Expanded loop logic

**Planner:**
```go
type Planner interface {
    Decompose(ctx context.Context, task string, context *Context) (*Plan, error)
    Reprioritize(ctx context.Context, plan *Plan, newInfo *Observation) (*Plan, error)
}

type Plan struct {
    Steps    []*Step
    Current  int
    Goal     string
    Context  *Context
}

type Step struct {
    Action     *computeruse.Action
    DependsOn  []int
    VerifyWith *VerifyRule
    Status     StepStatus
}
```

**Verifier:**
```go
type Verifier interface {
    Verify(ctx context.Context, step *Step, result *ActionResult) (*VerificationResult, error)
    ShouldRetry(ctx context.Context, result *ActionResult, attempt int) bool
}
```

**Expanded loop:**
1. User utterance → Planner decomposes into steps
2. For each step:
   a. Gatekeeper check (Phase 4.0)
   b. Twin-snapshot verification (Phase 4.2)
   c. Execute action via computer-use router (Phase 4.0)
   d. Verify result (Phase 4.4)
   e. If failed → retry or reprioritize
3. After all steps → summarize actions → store in episodic memory (Phase 4.3)
4. Extract semantic facts from interaction (Phase 4.3)

**Definition of done:**
- Planner can decompose tasks into ordered steps
- Verifier checks post-action state
- Retry logic with backoff
- Integration with memory system
- Unit tests for planning logic
- Integration tests with mock backends

---

### 4.5 Polish + Integration

**Goal:** Final integration, error handling, performance verification.

**Tasks:**
- Wire all sub-phases together
- End-to-end tests: voice → plan → execute → verify → remember
- Performance benchmarks:
  - AX tree capture < 100ms
  - Screenshot capture < 200ms
  - Twin-snapshot comparison < 50ms
  - Memory search < 50ms
  - Full action cycle < 500ms
- Error handling: graceful degradation when permissions missing
- Logging: structured logs for all computer-use actions
- Audit trail: all actions logged to HMAC-chained audit log

**Definition of done:**
- All sub-phases integrated
- End-to-end tests passing
- Performance budgets met
- Lint clean, race-free
- LOGBOOK updated
- CI green

---

## 5. Safety Invariants (herited from Phase 4)

1. **Gatekeeper is the only path to physical action.** No computer-use action bypasses the gatekeeper.
2. **Twin-snapshot is mandatory for WRITE/NETWORK.** No exceptions.
3. **User interruption detection.** Agent yields when user interacts.
4. **Battery-aware capture.** Don't drain battery on laptop.
5. **Audit trail.** Every action logged with HMAC chaining.

---

## 6. Performance Budgets

| Metric | Target |
|---|---|
| AX tree capture | < 100ms |
| Screenshot capture | < 200ms |
| Twin-snapshot comparison | < 50ms |
| Memory search (FTS5 + vector) | < 50ms |
| Full action cycle (plan → execute → verify) | < 500ms |
| Memory storage (write + index) | < 100ms |
| Agent loop iteration | < 2s |

---

## 7. Platform Support

| Platform | AX Backend | Status |
|---|---|---|
| macOS | Accessibility API (ApplicationServices) | Primary target |
| Windows | UI Automation (xa11y) | Secondary |
| Linux | AT-SPI2 | Tertiary |

Phase 5 focuses on macOS. Windows and Linux support can be added later.

---

## 8. Dependencies

| Dependency | Purpose | License |
|---|---|---|
| mattn/go-sqlite3 | SQLite driver | MIT |
| wentiln/go-fts5 | FTS5 bindings | MIT |
| asg017/sqlite-vec | Vector similarity | MIT |
| None for AX | Pure Go + CGo (ApplicationServices) | N/A |

All dependencies are local-first, no cloud required.

---

## 9. Definition of Done (Phase 5 as a whole)

Phase 5 is done when:

1. Condura can read the AX tree on macOS
2. Condura can capture screenshots
3. Condura can click, type, scroll, and press keys
4. Twin-snapshot verification prevents stale-state actions
5. 3-layer memory stores and retrieves information
6. Agent loop decomposes tasks into steps
7. All actions go through gatekeeper
8. All performance budgets met
9. Lint clean, race-free, CI green
10. LOGBOOK updated with Phase 5 completion

---

## 10. Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| macOS Accessibility permission denied | Agent can't read AX tree | Graceful degradation, prompt user |
| AX tree unstable across apps | Inconsistent behavior | Standardize element queries |
| sqlite-vec CGo issues | Build failures | Test cross-platform early |
| Performance budget exceeded | Sluggish UX | Profile, optimize hot paths |
| Memory leaks in CGo | Crashes | Careful allocation, finalizers |

---

## 11. What Phase 5 Deliberately Does NOT Include

- Wake word ("hey synaptic") — deferred
- P2P sync — deferred to Phase 6+
- Skills Hub — deferred to Phase 6+
- Action Replay — deferred to later
- Marketing/launch — deferred to Phase 7
- Windows/Linux AX backends — deferred (macOS primary)
- Cloud memory sync — not in scope (local-first)
