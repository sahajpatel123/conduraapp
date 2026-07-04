# Architecture 07 — The 3-Layer Memory System

> How Condura remembers what matters, forgets what doesn't, and never loses the audit trail.

---

## The Goal

Condura must:

1. **Remember** facts, preferences, and context that make future interactions better.
2. **Forget** stale, wrong, or sensitive data automatically.
3. **Recall** the right thing at the right time without the user asking.
4. **Audit** every fact and every access.

We use a 3-layer system: **Working**, **Episodic**, **Semantic**. Plus the **Audit Log** as a separate, append-only memory.

---

## Layer 1: Working Memory (Session)

**What it is**: The current conversation. The last N turns, the current goal, the current task state.

**Storage**: In-memory in the Go daemon. Not persisted to disk. Lost when the session ends (or when the user explicitly resets).

**Lifetime**: The session.

**Size limit**: 64K tokens. Older turns are summarized or dropped.

**Use case**: "The user said X 3 turns ago, the agent should remember that for this conversation."

---

## Layer 2: Episodic Memory (Cross-Session)

**What it is**: Time-stamped events from past sessions. "On 2026-05-12, the user asked me to book a flight to Tokyo. I used Kayak. The user said economy, $1500 max. I succeeded."

**Storage**: SQLite table `episodes(id, ts, session_id, summary, tags, importance)`. Each row is a structured event.

**Lifetime**: 90 days by default. User can mark episodes as "keep forever."

**Size**: ~10K episodes per user, indexed by tag, time, and embedding.

**Recall**: keyword + embedding + recency. "What did I do last time I was planning a trip?"

**Distillation**: nightly, a local small model (e.g., Qwen 1.5B via Ollama) summarizes the day's episodes into 1-3 sentence highlights. These go into the user model.

**Use case**: "The user often books flights on Kayak." (Derived from 5 episodes over 3 months.)

---

## Layer 3: Semantic Memory (Knowledge Graph)

**What it is**: A graph of facts, entities, and relationships. "User works at Acme Corp." "Acme Corp is in San Francisco." "User's manager is Alice." "Alice prefers email."

**Storage**: SQLite + vector index. Each fact has a confidence score and source.

**Lifetime**: Forever (until the user deletes).

**Structure**:
```
(user)
  -- works_at --> (acme_corp)
  -- manages --> (project_x)
  -- prefers_style --> (concise, technical)
  -- expert_in --> (go, typescript)
  -- uses_app --> (vscode, terminal, figma)

(acme_corp)
  -- located_in --> (san_francisco)
  -- has_manager --> (alice)

(alice)
  -- email --> [REDACTED:EMAIL]
  -- prefers_channel --> (email)
  -- trust_level --> (high)
```

**Recall**: graph traversal + embedding. "What do I know about Alice?" → traverses the graph from the user node.

**Confidence decay**: facts have a confidence score. If a fact hasn't been re-confirmed in 6 months, the confidence drops. Below 0.3, the fact is "soft deleted" (kept but not surfaced).

**Conflict resolution**: if two facts contradict, the one with higher confidence wins. The user can override.

**Use case**: "When I email Alice, use the formal style the engine has learned."

---

## The Audit Log (Separate, Append-Only)

**What it is**: Every action the agent has taken, ever. Not a "memory" in the user-facing sense, but a system record.

**Storage**: HMAC-chained append-only log. See [04-safety.md](04-safety.md).

**Lifetime**: 90 days by default. Critical events kept forever.

**Use case**: "Show me everything the agent did in the last hour." "What did the agent do before this error?"

---

## What Goes Where

| Data type | Layer | Why |
|---|---|---|
| Current conversation | Working | Ephemeral |
| Task state (DAG) | Working | Ephemeral |
| "User asked X today" | Episodic | Time-stamped event |
| "User did X 5 times" | Episodic → distilled to Semantic | Pattern |
| "User prefers X" | Semantic | Fact |
| "User's email is X" | Semantic (encrypted) | Fact |
| "Agent clicked Y at 14:32" | Audit | System record |
| "User opened 1Password" | (NEVER remembered) | Privacy |

---

## Embedding & Recall

Each fact (episodic or semantic) is embedded with a local embedding model (mxbai-embed-large via Ollama, or any of the configured embedding providers).

When the agent needs to recall:

1. Embed the current query.
2. Search episodic and semantic for the top-K most similar.
3. Apply filters: time range, confidence threshold, tags.
4. Inject the top results into the prompt context.

**The agent does not always recall.** Some queries don't benefit from memory (e.g., "what's 2+2?"). The router decides whether to enable memory recall.

---

## Forgetting (GDPR & User Control)

The user can:

- **Forget a fact**: `Settings → Memory → Forget this`. The fact is deleted from semantic and any episodes that mention it.
- **Forget a session**: `Settings → Memory → Delete session`. All episodes from that session are removed.
- **Forget everything**: `Settings → Memory → Reset`. The semantic and episodic stores are wiped. Audit log is preserved (configurable).
- **Auto-expire**: episodes older than N days are auto-deleted. Default 90. User can set 7, 30, 90, 365, forever.

**Encryption at rest**: the entire store is encrypted with AES-256-GCM, key derived from the install secret.

**Local only**: memory is never sent to the cloud. Never used to train a model.

---

## The Privacy Boundary

Some things are **never** remembered:

- Passwords, API keys, OAuth tokens (in working memory only, for the session).
- 1Password / password manager contents.
- Banking / financial app screens.
- Private/incognito browsing.
- "Pause apps" the user has marked.
- Encrypted file contents (we see the path, not the content).

**Anything tagged "sensitive" by the user is not even embedded.**

---

## The Distillation Pipeline

Nightly, on a schedule the user can configure:

1. Gather the day's episodes.
2. Send to a local small model (Qwen 1.5B or similar via Ollama).
3. Ask: "What are 1-3 facts I can confidently learn about the user from these episodes? Output in user-model format."
4. The dialectic (see [05-adaptive.md](05-adaptive.md)) validates.
5. Accepted facts are added to the semantic memory with a "new" flag.
6. The user is notified in the morning digest: "I learned 2 new things about you. Review?"

**The digest is opt-out-able**, but the learning continues silently. The user can always review at any time.

---

## Memory-Aware Prompting

When the Strategist is reasoning about a request, it can ask the memory system:

- "What does the user model say about this domain?"
- "Have we done this kind of task before?"
- "What did the user say last time about this?"

The response is **injected into the prompt** as a "memory context" block. The model is told: "This is your memory of past interactions. Use it to inform your response."

**The Strategist is told it can be wrong about the user** — if its memory is contradicted by current context, it should defer to the current context.

---

## Memory Schema (Simplified)

```sql
-- Episodes
CREATE TABLE episodes (
    id TEXT PRIMARY KEY,
    ts INTEGER NOT NULL,
    session_id TEXT NOT NULL,
    summary TEXT NOT NULL,
    details TEXT,             -- JSON
    tags TEXT,                -- comma-separated
    importance REAL DEFAULT 0.5,
    embedding BLOB,
    expires_at INTEGER        -- 0 = never
);

CREATE INDEX idx_episodes_ts ON episodes(ts);
CREATE INDEX idx_episodes_session ON episodes(session_id);
CREATE INDEX idx_episodes_importance ON episodes(importance);

-- Semantic facts
CREATE TABLE facts (
    id TEXT PRIMARY KEY,
    subject TEXT NOT NULL,    -- e.g., "user", "acme_corp", "alice"
    predicate TEXT NOT NULL,  -- e.g., "works_at"
    object TEXT NOT NULL,     -- e.g., "Acme Corp"
    confidence REAL DEFAULT 0.5,
    source TEXT,              -- episode ID, user statement, etc.
    created_at INTEGER NOT NULL,
    last_confirmed_at INTEGER,
    last_mentioned_at INTEGER,
    expires_at INTEGER,       -- 0 = never
    embedding BLOB
);

CREATE INDEX idx_facts_subject ON facts(subject);
CREATE INDEX idx_facts_predicate ON facts(predicate);
CREATE INDEX idx_facts_confidence ON facts(confidence);

-- Audit log (separate, append-only)
CREATE TABLE audit (
    seq INTEGER PRIMARY KEY AUTOINCREMENT,
    ts INTEGER NOT NULL,
    session_id TEXT,
    action_type TEXT,
    action_json TEXT,
    blast_radius TEXT,
    decision TEXT,        -- allow, deny, consent
    user_response TEXT,
    result TEXT,
    prev_hash TEXT,
    this_hash TEXT
);
```

---

## Related Docs

- [00-overview.md](00-overview.md) — The conductor pattern
- [05-adaptive.md](05-adaptive.md) — How the Adaptive Engine distills memory into the user model
- [08-sync.md](08-sync.md) — How memory syncs across devices
- [04-safety.md](04-safety.md) — The audit log
- [CLAUDE.md Section 8](../CLAUDE.md) — Memory section
