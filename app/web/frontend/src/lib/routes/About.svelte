<script lang="ts">
  // About — mission, architecture, non-negotiables, tech stack, team, legal.
  // This is the canonical "what is Condura?" page.
  import { Card } from '../components/ui'
  import Badge from '../components/ui/Badge.svelte'

  interface Invariant {
    id: string
    title: string
    body: string
  }

  interface ArmorModule {
    code: string
    name: string
    summary: string
    package: string
  }

  interface StackBadge {
    name: string
    tone: 'neutral' | 'accent' | 'success' | 'warn' | 'info'
  }

  const invariants: Invariant[] = [
    {
      id: 'strategist-gatekeeper',
      title: 'The Strategist and the Gatekeeper are separate systems.',
      body: 'The Strategist is a model. The Gatekeeper is deterministic code. They are never the same.',
    },
    {
      id: 'gatekeeper-only-path',
      title: 'The Gatekeeper is the only path to physical action.',
      body: 'No model output flows to a click, type, or shell exec without passing the Gatekeeper.',
    },
    {
      id: 'human-required',
      title: 'Destructive actions require a real human at the keyboard.',
      body: 'A native modal dialog blocks execution until the human physically clicks Allow. No exceptions.',
    },
    {
      id: 'user-stops',
      title: 'The user can always stop the agent.',
      body: 'Hard hotkey, watchdog timer, network isolation, menu bar kill — four independent mechanisms. The agent cannot disable any of them.',
    },
    {
      id: 'audit-chain',
      title: 'Every action is auditable, in a tamper-resistant log.',
      body: 'HMAC-chained, append-only, never deleted. Forensics-ready when something goes wrong.',
    },
    {
      id: 'guest',
      title: 'The agent is a guest, not an owner.',
      body: 'It requests permission to enter rooms (apps, files, URLs). The user grants or denies. No escalation, no bypass.',
    },
    {
      id: 'os-permissions',
      title: 'OS permissions are granted by the user, on their machine.',
      body: "We don't have access. We ask. The onboarding flow makes this easy and clear.",
    },
  ]

  const armor: ArmorModule[] = [
    {
      code: '10.1',
      name: 'Blast Radius Classifier',
      summary: 'Classifies every action into READ / WRITE / NETWORK / DESTRUCTIVE.',
      package: 'internal/blastradius',
    },
    {
      code: '10.2',
      name: 'The Gatekeeper',
      summary: 'Deterministic rules engine. Pure-rules, no neural net. Cannot be prompt-injected.',
      package: 'internal/gatekeeper',
    },
    {
      code: '10.3',
      name: 'Kill Switch (3 layers)',
      summary: 'Hard hotkey, watchdog timer, network isolation in a separate process.',
      package: 'internal/halt',
    },
    {
      code: '10.4',
      name: 'Behavioral Anomaly Detector',
      summary: 'Fires on stuck loops, machine-speed actions, never-seen network targets.',
      package: 'internal/anomaly',
    },
    {
      code: '10.5',
      name: 'Audit Log (HMAC-chained)',
      summary: 'Append-only, forensically sound, 90-day retention, secret-redacted.',
      package: 'internal/audit',
    },
    {
      code: '10.6',
      name: 'Model Isolation / Sanitizers',
      summary: 'Shell, Python, path, URL, and PII sanitizers between any model and the executor.',
      package: 'internal/sanitize',
    },
    {
      code: '10.7',
      name: 'Sensitive Site Detector',
      summary: 'Heuristic + allowlist detection of banking, health, and credential surfaces.',
      package: 'internal/sensitive',
    },
    {
      code: '10.8',
      name: 'Spend Monitor',
      summary: 'Periodic provider-dashboard checks. Hard limits per provider.',
      package: 'internal/failover',
    },
    {
      code: '10.9',
      name: 'Autonomy Matrix',
      summary: 'Per-task-type and per-app autonomy. Default cautious, user dial.',
      package: 'internal/autonomy',
    },
  ]

  const stack: StackBadge[] = [
    { name: 'Go 1.22+', tone: 'info' },
    { name: 'TypeScript', tone: 'info' },
    { name: 'Svelte 5', tone: 'accent' },
    { name: 'React 18 + Vite', tone: 'accent' },
    { name: 'Wails v2', tone: 'accent' },
    { name: 'SQLite + FTS5', tone: 'neutral' },
    { name: 'sqlite-vec', tone: 'neutral' },
    { name: 'whisper.cpp', tone: 'warn' },
    { name: 'openWakeWord', tone: 'warn' },
    { name: 'Next.js 14', tone: 'info' },
    { name: 'GoReleaser', tone: 'neutral' },
    { name: 'JSON-RPC 2.0', tone: 'neutral' },
  ]
</script>

<div class="about-page">
  <!-- ── Mission ────────────────────────────────────── -->
  <section class="section">
    <header class="section-head">
      <h2 class="display-h2">Mission</h2>
      <p class="lede">
        Make AI useful to every ordinary person, on every computer, for free. No lock-in.
        No tracking. No compromise on speed or safety.
      </p>
    </header>
    <Card elevation="glass" padding="lg">
      <p class="mission-body">
        Build a free, downloadable, OS-native AI agent that lives on a user's computer
        and acts as the conductor of every other AI tool installed there. It opens with a
        custom global hotkey, listens for the wake word, clicks and scrolls through any
        app, and runs sub-agents across Claude Code, Codex, Antigravity, OpenCode, Kilo,
        Hermes, Ollama, and any ChatGPT Plus / Claude Pro / Gemini AI Pro / SuperGrok
        subscription the user already has — all while costing the user nothing.
      </p>
    </Card>
  </section>

  <!-- ── Architecture — Seven Invariants ─────────────── -->
  <section class="section">
    <header class="section-head">
      <h2 class="display-h2">Architecture</h2>
      <p class="lede">
        The seven invariants. If a feature conflicts with any of these, the feature is
        wrong — remove it.
      </p>
    </header>

    <div class="invariant-grid">
      {#each invariants as inv, i (inv.id)}
        <Card elevation="glass" padding="md">
          <div class="inv-card">
            <span class="inv-num mono">0{i + 1}</span>
            <h3 class="inv-title">{inv.title}</h3>
            <p class="inv-body">{inv.body}</p>
          </div>
        </Card>
      {/each}
    </div>
  </section>

  <!-- ── Non-Negotiables — The Armor ─────────────────── -->
  <section class="section">
    <header class="section-head">
      <h2 class="display-h2">Non-Negotiables</h2>
      <p class="lede">
        The Safety Layer modules. Built before any agent capability, completed before
        the public binary ships.
      </p>
    </header>

    <div class="armor-list">
      {#each armor as mod (mod.code)}
        <div class="armor-row">
          <span class="armor-code mono">{mod.code}</span>
          <div class="armor-text">
            <h4>{mod.name}</h4>
            <p>{mod.summary}</p>
          </div>
          <code class="armor-pkg mono">{mod.package}</code>
        </div>
      {/each}
    </div>
  </section>

  <!-- ── Tech Stack ──────────────────────────────────── -->
  <section class="section">
    <header class="section-head">
      <h2 class="display-h2">Tech Stack</h2>
      <p class="lede">Locked. Local-first. Single-binary daemon, web UI, native overlay.</p>
    </header>

    <div class="badges">
      {#each stack as s (s.name)}
        <Badge tone={s.tone} size="md">{s.name}</Badge>
      {/each}
    </div>
  </section>

  <!-- ── Team ────────────────────────────────────────── -->
  <section class="section">
    <header class="section-head">
      <h2 class="display-h2">Team</h2>
      <p class="lede">A human + AI partnership. Architect and product lead paired with an implementer and reviewer.</p>
    </header>
    <Card elevation="glass" padding="lg">
      <div class="team-grid">
        <div class="team-card">
          <div class="team-avatar">SP</div>
          <h4>Human</h4>
          <p class="team-role">Architect & product lead</p>
          <p class="team-bio">
            Owns direction, the locked decisions, and the partner commitment to ship the
            best version of what we imagined.
          </p>
        </div>
        <div class="team-card">
          <div class="team-avatar ai">AI</div>
          <h4>AI Implementer</h4>
          <p class="team-role">Engineering & review</p>
          <p class="team-bio">
            Builds the components, runs the audits, writes the tests, and keeps the
            survival invariants honest.
          </p>
        </div>
      </div>
    </Card>
  </section>

  <!-- ── Legal ───────────────────────────────────────── -->
  <section class="section">
    <header class="section-head">
      <h2 class="display-h2">Legal</h2>
      <p class="lede">Synaptic Freeware EULA v1 — free personal and commercial, no redistribution, revocable for abuse.</p>
    </header>

    <div class="legal-links">
      <a class="legal-link" href="https://github.com/sahajpatel123/conduraapp/blob/main/EULA.md" target="_blank" rel="noreferrer">
        <span class="legal-name">EULA.md</span>
        <span class="legal-sub">End-User License Agreement</span>
      </a>
      <a class="legal-link" href="https://github.com/sahajpatel123/conduraapp/blob/main/PRIVACY.md" target="_blank" rel="noreferrer">
        <span class="legal-name">PRIVACY.md</span>
        <span class="legal-sub">Local-first. The agent is a guest.</span>
      </a>
      <a class="legal-link" href="https://github.com/sahajpatel123/conduraapp/blob/main/SECURITY.md" target="_blank" rel="noreferrer">
        <span class="legal-name">SECURITY.md</span>
        <span class="legal-sub">Disclosure policy</span>
      </a>
    </div>
  </section>
</div>

<style>
  .about-page {
    padding: var(--space-7) var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide);
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-10);
  }

  /* ── Section heads ──────────────────────────────── */
  .section {
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .section:nth-of-type(1) { animation-delay: 40ms; }
  .section:nth-of-type(2) { animation-delay: 120ms; }
  .section:nth-of-type(3) { animation-delay: 200ms; }
  .section:nth-of-type(4) { animation-delay: 280ms; }
  .section:nth-of-type(5) { animation-delay: 360ms; }
  .section:nth-of-type(6) { animation-delay: 440ms; }

  .section-head {
    margin-bottom: var(--space-5);
  }
  .display-h2 {
    font-family: var(--font-display);
    font-size: var(--size-2xl);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    color: var(--text);
    margin: 0 0 var(--space-2) 0;
    line-height: var(--leading-tight);
  }
  .lede {
    font-size: var(--size-md);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    max-width: 640px;
    margin: 0;
  }

  /* ── Mission body ────────────────────────────────── */
  .mission-body {
    font-family: var(--font-display);
    font-size: var(--size-lg);
    line-height: var(--leading-snug);
    letter-spacing: var(--tracking-normal);
    color: var(--text);
    margin: 0;
  }

  /* ── Invariant grid ──────────────────────────────── */
  .invariant-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: var(--space-4);
  }
  .inv-card {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    height: 100%;
  }
  .inv-num {
    font-size: var(--size-xs);
    color: var(--accent);
    letter-spacing: var(--tracking-widest);
    text-transform: uppercase;
    opacity: 0.85;
  }
  .inv-title {
    font-family: var(--font-sans);
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    line-height: var(--leading-snug);
    margin: 0;
  }
  .inv-body {
    font-size: var(--size-sm);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    margin: 0;
  }

  /* ── Armor modules ──────────────────────────────── */
  .armor-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-2);
  }
  .armor-row {
    display: grid;
    grid-template-columns: 56px 1fr auto;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    transition: background var(--transition-fast);
  }
  .armor-row:hover { background: var(--surface-2); }
  .armor-code {
    font-size: var(--size-xs);
    color: var(--accent);
    letter-spacing: var(--tracking-wide);
    opacity: 0.85;
  }
  .armor-text h4 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    margin: 0 0 2px 0;
  }
  .armor-text p {
    font-size: var(--size-sm);
    color: var(--text-muted);
    line-height: var(--leading-snug);
    margin: 0;
  }
  .armor-pkg {
    font-size: var(--size-xs);
    color: var(--text-faint);
    background: var(--surface-3);
    border: 1px solid var(--border);
    padding: 3px 8px;
    border-radius: var(--radius-sm);
  }

  /* ── Stack badges ───────────────────────────────── */
  .badges {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }

  /* ── Team ───────────────────────────────────────── */
  .team-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: var(--space-6);
  }
  .team-card {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .team-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: var(--accent-gradient);
    color: var(--text-inverse);
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-mono);
    font-size: var(--size-sm);
    font-weight: var(--weight-semibold);
    letter-spacing: var(--tracking-wide);
    box-shadow: var(--shadow-glow);
  }
  .team-avatar.ai {
    background: linear-gradient(135deg, var(--info) 0%, #6b8cff 100%);
  }
  .team-card h4 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    margin: 0;
  }
  .team-role {
    font-size: var(--size-xs);
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
    margin: 0;
  }
  .team-bio {
    font-size: var(--size-sm);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    margin: 0;
  }

  /* ── Legal links ────────────────────────────────── */
  .legal-links {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: var(--space-3);
  }
  .legal-link {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: var(--space-4);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    text-decoration: none;
    transition:
      background var(--transition-fast),
      border-color var(--transition-fast),
      transform var(--transition-fast) var(--ease-spring);
  }
  .legal-link:hover {
    background: var(--surface-2);
    border-color: var(--border-focus);
    transform: translateY(-2px);
  }
  .legal-name {
    font-family: var(--font-mono);
    font-size: var(--size-sm);
    color: var(--accent);
  }
  .legal-sub {
    font-size: var(--size-xs);
    color: var(--text-muted);
  }
</style>
