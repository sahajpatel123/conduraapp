# ADR-0002: TypeScript + React for the UI

- **Status**: Accepted
- **Date**: 2026-06-06
- **Deciders**: Synaptic core team
- **Supersedes**: —
- **Superseded by**: —

---

## Context

Synaptic has 3 UIs:

1. **Overlay** — a floating chat/voice box on the desktop.
2. **TUI** — a terminal user interface for power users and SSH.
3. **Web dashboard** — for management on the go (and as a marketing surface).

The overlay and web dashboard are **graphical**. The TUI is **text-based**. All three need to be **fast, beautiful, accessible, and internationalized**.

The overlay and web dashboard run inside a **Wails** app (Go + web view). The TUI runs in a **terminal**.

We need to choose a UI framework for the graphical UIs.

## Decision

**We use TypeScript + React 18 for the graphical UIs (overlay, web dashboard).**

**We use TypeScript + Ink for the TUI.**

## Rationale

### Why TypeScript

- **Type safety**: catches UI bugs at compile time, critical for a product with so many state transitions.
- **Ecosystem**: every UI library, every i18n library, every animation library is TS-first.
- **Team velocity**: most developers know TS or JS.
- **Code reuse**: the TUI and the web dashboard can share types and validation logic.

### Why React

- **Mature**: 10+ years old, stable API, huge ecosystem.
- **Concurrent rendering**: React 18's concurrent features are perfect for streaming LLM tokens and live-updating UI.
- **Component model**: maps cleanly to our UI (Composer, Message, ToolCall, StatusBar, Settings).
- **Dev tools**: React DevTools, Storybook, etc.
- **SSR support**: the web dashboard can be statically generated for fast marketing page loads.

### Why Wails (not Electron, not Tauri)

- **Performance**: Wails uses the native web view (WKWebView on macOS, WebView2 on Windows, WebKitGTK on Linux). Cold start < 500ms, memory < 80MB.
- **Size**: Wails apps are ~10MB. Electron apps are ~150MB.
- **Native feel**: the overlay uses native window chrome, native dialogs, native notifications.
- **No Chromium**: we don't ship a browser. We use the OS's web view.
- **Go + Web**: the Go daemon and the TS UI communicate via JSON-RPC over Unix socket / named pipe. Wails makes this easy.

### Why Ink for the TUI

- **React for the terminal**: same mental model as the web UI.
- **TypeScript**: share types with the web UI.
- **Composition**: components compose the same way.
- **Streaming**: streaming LLM tokens to a TUI is well-understood in Ink.

### Why not Svelte / Vue / Solid

- **Svelte** is excellent and would be a defensible alternative. We chose React for ecosystem maturity and team velocity.
- **Vue** is great for SPAs but less universal in the LLM/agent ecosystem.
- **Solid** is fast but smaller ecosystem.

### Why not Flutter / native

- **Flutter** is heavy and not great for desktop overlays.
- **Native (Swift/C#)** would mean 3 separate UIs. Too slow to build.

## Consequences

### Positive

- One language (TypeScript) for all clients.
- Mature ecosystem.
- Fast iteration.
- Easy hiring.

### Negative

- **Bundle size**: must be disciplined about what we ship in the overlay bundle.
- **React's verbosity**: hooks ceremony, prop drilling. We mitigate with Zustand for state.
- **TypeScript build step**: a bit of ceremony. Vite + SWC makes it fast.

### Neutral

- We commit to React 18+ for the graphical UIs.
- We commit to Ink 4+ for the TUI.
- We commit to **Wails v2** for the desktop shell.

---

## The UI Architecture (Simplified)

```
┌────────────────────────────────────────────────────┐
│  Wails App (Go + WebView)                          │
│                                                    │
│  ┌──────────────────────────────────────────────┐ │
│  │  React Overlay                               │ │
│  │                                              │ │
│  │  <App>                                       │ │
│  │    <Composer />                              │ │
│  │    <MessageList />                           │ │
│  │    <StatusBar />                             │ │
│  │    <SettingsDrawer />                        │ │
│  │  </App>                                      │ │
│  └──────────────────────────────────────────────┘ │
│                                                    │
│  IPC: window.runtime.EventsOn / EventsEmit         │
│  Transport: JSON-RPC 2.0 over Unix socket          │
└────────────────────────────────────────────────────┘
                            ↕
┌────────────────────────────────────────────────────┐
│  Go Daemon                                         │
│  (synapticd)                                       │
└────────────────────────────────────────────────────┘
```

---

## State Management

We use **Zustand** for global state in the React apps:

- Small (~1KB).
- No boilerplate.
- Excellent TypeScript support.
- Easy to test.

We **do not** use Redux. It's overkill for our needs and adds ceremony.

For server state (data from the daemon), we use **TanStack Query** (React Query):

- Caching, invalidation, refetch.
- Streaming-friendly.
- Excellent devtools.

---

## Styling

We use **Tailwind CSS** for the React UIs:

- Utility-first: fast iteration.
- Tree-shakeable: small bundle.
- Consistent design system via `tailwind.config.ts`.

For the Ink TUI, we use **Ink's flexbox layout** and ANSI colors. No CSS needed.

---

## Internationalization

We use **react-i18next** (i18next) for translations:

- 6 languages at v0.1.0 (en, es, fr, de, hi, ja).
- Lazy-loaded per language.
- Right-to-left (RTL) support for future Arabic/Hebrew.

---

## Accessibility

We target **WCAG 2.1 AA**:

- Keyboard navigation throughout.
- Screen reader support (ARIA).
- High-contrast mode.
- Respects `prefers-reduced-motion`.
- Voice-control friendly (every action has a label).

---

## Testing

- **Unit tests**: Vitest.
- **Component tests**: React Testing Library.
- **E2E tests**: Playwright (web dashboard), Spectron-like for the overlay (manual + automated).
- **Visual regression**: Chromatic or Percy (deferred to v0.2).

---

## Related Docs

- [ADR-0001](0001-go-over-python.md) — Why Go for the daemon
- [CLAUDE.md Section 6](../CLAUDE.md) — Stack details
- [Wails docs](https://wails.io)
