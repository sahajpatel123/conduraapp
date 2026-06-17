# Code Style Guide

> Conventions for Go, TypeScript, and Python in the Condura codebase.

---

## Overview

Condura is a multi-language project:

- **Go** for the core daemon (`synapticd`).
- **TypeScript / React** for the overlay, TUI, and web dashboard.
- **Python** for the 3 computer-use bridges.
- **Swift / Kotlin / Rust** may appear later for native components.

This document covers the **enforced** style for the three primary languages. It is enforced by:

- **Go**: `gofmt`, `goimports`, `golangci-lint` (with `revive`, `gocritic`, `gosec`).
- **TypeScript**: `eslint`, `prettier`, `tsc --strict`.
- **Python**: `ruff`, `black`, `mypy`.

---

## General Rules (All Languages)

1. **No commented-out code.** If it's not needed, delete it. Git remembers.
2. **No "TODO" comments as a substitute for doing the work.** If something is in scope, do it. If it's not in scope, don't add it.
3. **No magic numbers.** All constants get a name and a doc comment.
4. **No global state.** Dependency injection everywhere.
5. **Errors are values.** Handle them. Wrap them. Surface them.
6. **No panics in production code.** Panics are for "this should never happen" only.
7. **No `fmt.Println` for logging.** Use the structured logger.
8. **No secrets in code or logs.** API keys, tokens, etc. are secrets — never log them, never commit them.
9. **File headers are mandatory.** Every file starts with a header doc comment.
10. **Public APIs are documented.** Every exported function, type, method has a doc comment.

---

## Go Style

### File Header

```go
// Package router implements the hybrid-with-memory router.
//
// The router picks the right model or CLI for each sub-task,
// using the user's configured priorities, the task's capabilities,
// the Adaptive Engine's hints, and the delegate's health.
//
// See docs/architecture/01-router.md for the full design.
package router
```

For non-package files (e.g., `internal/foo/bar.go`):

```go
// File: bar.go
// Package foo provides the X functionality.
// This file implements Y.
package foo
```

### Imports

- Standard library first, then third-party, then internal.
- Groups separated by blank lines.
- Sorted within groups.

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/libp2p/go-libp2p"
    "github.com/stretchr/testify/assert"

    "github.com/sahajpatel123/conduraapp/internal/agent"
    "github.com/sahajpatel123/conduraapp/internal/safety"
)
```

### Naming

- **Packages**: short, lowercase, single word: `router`, `safety`, `agent`.
- **Types**: PascalCase, no abbreviations: `Router`, not `Rtr`.
- **Functions**: PascalCase for exported, camelCase for unexported.
- **Variables**: camelCase. Acronyms are all-caps: `ID`, `URL`, `HTTP`, but as types: `type UserID string`.
- **Constants**: PascalCase or SCREAMING_SNAKE for true constants.

```go
type Router struct { ... }
func (r *Router) Plan(ctx context.Context, task Task) (Plan, error) { ... }
func (r *Router) delegates() []*Delegate { ... }  // unexported

const MaxConcurrentRuns = 8
const defaultTimeout = 30 * time.Second
```

### Errors

- Always wrap: `fmt.Errorf("doing X: %w", err)`.
- Use `errors.Is` and `errors.As` for checking.
- Define sentinel errors for known cases: `var ErrNotFound = errors.New("not found")`.
- Don't use `panic` except for "this should never happen."

```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doing something: %w", err)
}
```

### Context

- `context.Context` is always the first parameter.
- Don't store context in a struct.
- Don't pass `nil` context; use `context.Background()` or `context.TODO()`.

```go
func (r *Router) Plan(ctx context.Context, task Task) (Plan, error) {
    select {
    case <-ctx.Done():
        return Plan{}, ctx.Err()
    default:
    }
    // ...
}
```

### Concurrency

- Use goroutines and channels idiomatically.
- Use `sync.WaitGroup` for waiting on multiple goroutines.
- Use `errgroup.Group` for goroutines that can fail.
- Never share mutable state without a lock.
- Document the concurrency model in the package doc.

```go
func (b *Bus) RunAll(ctx context.Context, tasks []Task) ([]Result, error) {
    g, ctx := errgroup.WithContext(ctx)
    results := make([]Result, len(tasks))
    for i, t := range tasks {
        i, t := i, t
        g.Go(func() error {
            r, err := b.Run(ctx, t)
            if err != nil {
                return err
            }
            results[i] = r
            return nil
        })
    }
    if err := g.Wait(); err != nil {
        return nil, err
    }
    return results, nil
}
```

### Testing

- Test files are `*_test.go` in the same package.
- Use table-driven tests where applicable.
- Use `testify/assert` and `testify/require`.
- Coverage must be >80% for safety/perception/agent/llm/ipc packages.

```go
func TestRouter_Plan(t *testing.T) {
    tests := []struct {
        name    string
        task    Task
        want    string
        wantErr bool
    }{
        {"research task", Task{Type: "research"}, "claude_code", false},
        {"code task", Task{Type: "code"}, "claude_code", false},
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r := NewRouter(/* ... */)
            got, err := r.Plan(context.Background(), tt.task)
            if (err != nil) != tt.wantErr {
                t.Errorf("got err = %v, wantErr = %v", err, tt.wantErr)
            }
            if got.Primary != tt.want {
                t.Errorf("got primary = %v, want %v", got.Primary, tt.want)
            }
        })
    }
}
```

### Tools

- `gofmt` and `goimports` must run clean.
- `golangci-lint` with config in `.golangci.yml`.
- `go vet` clean.
- `staticcheck` clean.
- `gosec` for security issues (fail build on any warning).

---

## TypeScript / React Style

### File Header

```typescript
/**
 * Router — picks the right model or CLI for each sub-task.
 *
 * This module wraps the Go daemon's `router.plan` JSON-RPC method
 * and exposes a typed, streaming-friendly API to the React components.
 *
 * See docs/architecture/01-router.md for the full design.
 */

import { useState, useEffect } from 'react';
// ...
```

### Imports

- Absolute imports preferred: `import { foo } from '@/lib/foo'`.
- Group: standard lib, third-party, internal.
- No default exports for shared utilities; named exports only.

### Naming

- **Files**: `kebab-case.ts` for non-component, `PascalCase.tsx` for components.
- **Types**: `PascalCase`, no `I` prefix: `Router`, not `IRouter`.
- **Functions**: `camelCase`.
- **Variables**: `camelCase`. Constants in `SCREAMING_SNAKE_CASE`.
- **React components**: `PascalCase`, one per file, named export.

```typescript
// router.tsx
export function RouterView() { ... }

// types.ts
export type RouterPlan = { ... };
```

### Types

- Use `type` for unions, intersections, and aliases.
- Use `interface` for object shapes that may be extended.
- `strict: true` in `tsconfig.json`. No `any` (use `unknown` and narrow).

```typescript
type TaskSpec = {
  goal: string;
  context: Message[];
  subTaskType: SubTaskType;
};

interface Router {
  plan(task: TaskSpec): Promise<Plan>;
  cancel(runId: string): Promise<void>;
}
```

### Async / Await

- `async/await` everywhere. No `.then()` chains.
- Handle errors with try/catch.
- Use `Promise.all` for parallelism.

```typescript
const results = await Promise.all(tasks.map(t => runTask(t)));
```

### React

- Functional components only.
- Hooks: follow the [rules of hooks](https://react.dev/warn/rules-of-hooks).
- Don't use class components.
- Don't use `useEffect` for data fetching — use TanStack Query.
- Use `useMemo` and `useCallback` only when there's a measured performance problem.

```tsx
export function Composer() {
  const [text, setText] = useState('');
  const { mutate: send, isPending } = useSendMessage();
  return (
    <div className="flex flex-col gap-2">
      <textarea
        value={text}
        onChange={(e) => setText(e.target.value)}
        className="border rounded p-2"
      />
      <button
        onClick={() => send({ text })}
        disabled={isPending}
        className="bg-blue-500 text-white px-4 py-2 rounded"
      >
        Send
      </button>
    </div>
  );
}
```

### State

- **Local state**: `useState`.
- **Global state**: Zustand.
- **Server state**: TanStack Query.
- **Form state**: React Hook Form + Zod.

```typescript
// store.ts
import { create } from 'zustand';

type State = {
  user: User | null;
  setUser: (u: User | null) => void;
};

export const useStore = create<State>((set) => ({
  user: null,
  setUser: (u) => set({ user: u }),
}));
```

### Styling

- Tailwind CSS for everything.
- No inline `style={{}}` except for dynamic values (e.g., `style={{ left: x }}`).
- Component-level styles via Tailwind's `cn` utility for conditional classes.

```typescript
import { cn } from '@/lib/cn';

<button className={cn(
  'px-4 py-2 rounded',
  isPending && 'opacity-50',
  isError && 'bg-red-500'
)} />
```

### Testing

- Vitest for unit tests.
- React Testing Library for component tests.
- Playwright for E2E (web dashboard).

```typescript
import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Composer } from './composer';

describe('Composer', () => {
  it('renders a textarea and a button', () => {
    render(<Composer />);
    expect(screen.getByRole('textbox')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /send/i })).toBeInTheDocument();
  });
});
```

### Tools

- `eslint` with config in `.eslintrc.cjs`.
- `prettier` with config in `.prettierrc`.
- `tsc --strict` clean.
- `vitest` for tests.
- No `console.log` in production code (use a logger).

---

## Python Style (Bridges Only)

### File Header

```python
"""
bridge.orax: macOS-specific computer-use bridge.

This module wraps pyobjc, Quartz, and Apple's AX APIs to expose
a JSON-RPC 2.0 interface over stdio for the Go daemon.

See docs/architecture/02-computer-use.md for the full design.
"""

from __future__ import annotations
import json
import sys
# ...
```

### Type Hints

- Type hints on every function and method.
- Use `from __future__ import annotations` for forward references.

```python
def ax_tree(app: str, max_depth: int = 5) -> dict[str, Any]:
    ...
```

### Naming

- **Modules**: `snake_case`.
- **Classes**: `PascalCase`.
- **Functions / variables**: `snake_case`.
- **Constants**: `SCREAMING_SNAKE_CASE`.

```python
MAX_DEPTH = 10
DEFAULT_TIMEOUT = 30

def click_button(name: str) -> bool:
    ...
```

### Errors

- Use specific exceptions: `ValueError`, `KeyError`, etc.
- Define custom exceptions for known cases.
- Don't use bare `except:`. Catch specific exceptions.
- Always include a message.

```python
class BridgeError(Exception):
    """Raised when the bridge cannot fulfill a request."""
    pass

try:
    result = do_something()
except FileNotFoundError as e:
    raise BridgeError(f"file not found: {e}") from e
```

### Async

- Use `asyncio` for I/O.
- Use `aiofiles` for file I/O.
- Use `httpx` for HTTP.

```python
import asyncio
import httpx

async def fetch(url: str) -> dict:
    async with httpx.AsyncClient() as client:
        r = await client.get(url)
        return r.json()
```

### Testing

- `pytest` for tests.
- Test files: `test_*.py` in the same directory or `tests/`.
- Use `pytest-asyncio` for async tests.

```python
import pytest
from bridge.orax import ax_tree

def test_ax_tree_returns_dict():
    result = ax_tree("Safari", max_depth=2)
    assert isinstance(result, dict)
    assert "app" in result
    assert "elements" in result
```

### Tools

- `ruff` for linting and formatting.
- `black` for formatting (or use ruff's formatter).
- `mypy --strict` for type checking.
- `pytest` for tests.
- `pip-compile` for dependency management.

---

## Directory Layout

```
/Users/sahajpatel/synaptic/
├── CLAUDE.md
├── LOGBOOK.md
├── EULA.md
├── LICENSE
├── README.md
├── CONTRIBUTING.md
├── SECURITY.md
├── PRIVACY.md
├── docs/
│   ├── README.md
│   ├── architecture/
│   ├── adr/
│   ├── guides/
│   ├── user-guide/
│   ├── recipes/
│   └── api/
├── internal/                  # Go
│   ├── agent/
│   ├── safety/
│   ├── perception/
│   ├── router/
│   ├── llm/
│   ├── memory/
│   ├── adaptive/
│   ├── delegation/
│   ├── sync/
│   ├── ipc/
│   ├── settings/
│   ├── presence/
│   ├── audit/
│   ├── skills/
│   ├── voice/
│   ├── i18n/
│   ├── store/
│   ├── secrets/
│   ├── config/
│   ├── logger/
│   └── bridge/                # Go side of the bridge
├── app/                       # Wails app
│   ├── main.go
│   ├── overlay.go
│   └── frontend/              # React/TS overlay
│       ├── package.json
│       ├── src/
│       └── public/
├── ts/                        # TypeScript packages
│   ├── packages/
│   │   ├── sdk/               # JSON-RPC client SDK
│   │   ├── ui/                # Shared UI components
│   │   ├── tui/               # Ink TUI
│   │   └── web/               # Next.js web dashboard
│   └── package.json
├── web/                       # Marketing site (Next.js)
│   ├── pages/
│   ├── components/
│   └── public/
├── bridge/                    # Python bridges
│   ├── orax/
│   ├── pyautogui/
│   └── mcp/
├── configs/                   # YAML / JSON configs
│   ├── default.yaml
│   └── schemas/               # JSON Schemas for IPC
├── scripts/                   # Build, lint, etc.
├── marketing/                 # Marketing assets
├── .github/                   # GitHub Actions
│   └── workflows/
└── test/                      # Integration & E2E tests
    ├── integration/
    ├── fixtures/
    └── mocks/
```

---

## Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short summary>

<optional body>

<optional footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`, `build`.

Example:

```
feat(router): add fallback chain with circuit breaker

When the primary delegate fails, the router now tries fallbacks
in priority order. After 3 consecutive failures, the delegate
is taken offline for 1 hour.

Closes #42
```

**Note**: The user must explicitly ask for commits. AI does not commit on its own.

---

## Pull Requests

PRs require:

1. A clear title and description.
2. All CI checks passing.
3. Reviewer approval (at least 1 maintainer for the area).
4. No merge conflicts.
5. Linked issue (if applicable).

PR description template:

```markdown
## What

<!-- What does this PR do? -->

## Why

<!-- Why is this change needed? -->

## How

<!-- How did you implement it? -->

## Testing

<!-- How was this tested? -->

## Screenshots / Logs

<!-- If applicable -->

## Checklist

- [ ] Tests added/updated
- [ ] Docs updated
- [ ] LOGBOOK.md updated
- [ ] CLAUDE.md updated (if decision was made)
```

---

## Related Docs

- [CONTRIBUTING.md](../../CONTRIBUTING.md) — Code style, PR process, AI workflow
- [CLAUDE.md](../../CLAUDE.md) — Master thinking
- [LOGBOOK.md](../../LOGBOOK.md) — Session log
