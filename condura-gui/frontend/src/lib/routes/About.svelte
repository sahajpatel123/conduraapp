<script lang="ts">
  // About — mission, architecture, non-negotiables, tech stack, team, legal.
  // This is the canonical "what is Condura?" page.
  import Card from '$components/v1/Card.svelte';
  import Surface from '$components/v1/Surface.svelte';
  import Stack from '$components/v1/Stack.svelte';
  import Inline from '$components/v1/Inline.svelte';
  import Hairline from '$components/v1/Hairline.svelte';
  import Pill from '$components/v1/Pill.svelte';

  interface Invariant {
    id: string;
    title: string;
    body: string;
  }

  interface ArmorModule {
    code: string;
    name: string;
    summary: string;
    package: string;
  }

  interface StackBadge {
    name: string;
    tone: 'neutral' | 'accent' | 'success' | 'warn' | 'info';
  }

  type PillVariant = 'success' | 'warning' | 'error' | 'info' | 'neutral' | 'accent';

  function pillVariant(tone: StackBadge['tone']): PillVariant {
    if (tone === 'warn') return 'warning';
    return tone;
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
  ];

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
  ];

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
  ];
</script>

<Stack class="about-page" gap="10" padding="7">
  <!-- ── Mission ────────────────────────────────────── -->
  <Stack as="section" gap="5" class="section">
    <header class="section-head">
      <h2 class="display-h2">Mission</h2>
      <p class="lede">
        Make AI useful to every ordinary person, on every computer, for free. No lock-in.
        No tracking. No compromise on speed or safety.
      </p>
    </header>
    <Card variant="raised" padding="6">
      {#snippet children()}
        <p class="mission-body">
          Build a free, downloadable, OS-native AI agent that lives on a user's computer
          and acts as the conductor of every other AI tool installed there. It opens with a
          custom global hotkey, listens for the wake word, clicks and scrolls through any
          app, and runs sub-agents across Claude Code, Codex, Antigravity, OpenCode, Kilo,
          Hermes, Ollama, and any ChatGPT Plus / Claude Pro / Gemini AI Pro / SuperGrok
          subscription the user already has — all while costing the user nothing.
        </p>
      {/snippet}
    </Card>
  </Stack>

  <!-- ── Architecture — Seven Invariants ─────────────── -->
  <Stack as="section" gap="5" class="section">
    <header class="section-head">
      <h2 class="display-h2">Architecture</h2>
      <p class="lede">
        The seven invariants. If a feature conflicts with any of these, the feature is
        wrong — remove it.
      </p>
    </header>

    <div class="invariant-grid">
      {#each invariants as inv, i (inv.id)}
        <Card variant="raised" padding="4">
          {#snippet children()}
            <Stack gap="2" class="inv-card">
              <span class="inv-num mono">0{i + 1}</span>
              <h3 class="inv-title">{inv.title}</h3>
              <p class="inv-body">{inv.body}</p>
            </Stack>
          {/snippet}
        </Card>
      {/each}
    </div>
  </Stack>

  <!-- ── Non-Negotiables — The Armor ─────────────────── -->
  <Stack as="section" gap="5" class="section">
    <header class="section-head">
      <h2 class="display-h2">Non-Negotiables</h2>
      <p class="lede">
        The Safety Layer modules. Built before any agent capability, completed before
        the public binary ships.
      </p>
    </header>

    <Surface variant="sunken" padding="2" radius="lg">
      {#snippet children()}
        <Stack gap="0">
          {#each armor as mod, i (mod.code)}
            {#if i > 0}
              <Hairline />
            {/if}
            <div class="armor-row">
              <span class="armor-code mono">{mod.code}</span>
              <div class="armor-text">
                <h4>{mod.name}</h4>
                <p>{mod.summary}</p>
              </div>
              <code class="armor-pkg mono">{mod.package}</code>
            </div>
          {/each}
        </Stack>
      {/snippet}
    </Surface>
  </Stack>

  <!-- ── Tech Stack ──────────────────────────────────── -->
  <Stack as="section" gap="5" class="section">
    <header class="section-head">
      <h2 class="display-h2">Tech Stack</h2>
      <p class="lede">Locked. Local-first. Single-binary daemon, web UI, native overlay.</p>
    </header>

    <Inline gap="2" class="badges">
      {#each stack as s (s.name)}
        <Pill variant={pillVariant(s.tone)} size="md" label={s.name} />
      {/each}
    </Inline>
  </Stack>

  <!-- ── Team ────────────────────────────────────────── -->
  <Stack as="section" gap="5" class="section">
    <header class="section-head">
      <h2 class="display-h2">Team</h2>
      <p class="lede">A human + AI partnership. Architect and product lead paired with an implementer and reviewer.</p>
    </header>
    <Card variant="raised" padding="6">
      {#snippet children()}
        <div class="team-grid">
          <Stack gap="2" class="team-card">
            <div class="team-avatar">SP</div>
            <h4 class="team-name">Human</h4>
            <p class="team-role">Architect & product lead</p>
            <p class="team-bio">
              Owns direction, the locked decisions, and the partner commitment to ship the
              best version of what we imagined.
            </p>
          </Stack>
          <Stack gap="2" class="team-card">
            <div class="team-avatar ai">AI</div>
            <h4 class="team-name">AI Implementer</h4>
            <p class="team-role">Engineering & review</p>
            <p class="team-bio">
              Builds the components, runs the audits, writes the tests, and keeps the
              survival invariants honest.
            </p>
          </Stack>
        </div>
      {/snippet}
    </Card>
  </Stack>

  <!-- ── Legal ───────────────────────────────────────── -->
  <Stack as="section" gap="5" class="section">
    <header class="section-head">
      <h2 class="display-h2">Legal</h2>
      <p class="lede">Synaptic Freeware EULA v1 — free personal and commercial, no redistribution, revocable for abuse.</p>
    </header>

    <Inline gap="3" class="legal-links">
      <a
        class="legal-link"
        href="https://github.com/sahajpatel123/conduraapp/blob/main/EULA.md"
        target="_blank"
        rel="noreferrer"
      >
        <Surface variant="raised" padding="4" radius="md" class="legal-surface">
          {#snippet children()}
            <Stack gap="1">
              <span class="legal-name">EULA.md</span>
              <span class="legal-sub">End-User License Agreement</span>
            </Stack>
          {/snippet}
        </Surface>
      </a>
      <a
        class="legal-link"
        href="https://github.com/sahajpatel123/conduraapp/blob/main/PRIVACY.md"
        target="_blank"
        rel="noreferrer"
      >
        <Surface variant="raised" padding="4" radius="md" class="legal-surface">
          {#snippet children()}
            <Stack gap="1">
              <span class="legal-name">PRIVACY.md</span>
              <span class="legal-sub">Local-first. The agent is a guest.</span>
            </Stack>
          {/snippet}
        </Surface>
      </a>
      <a
        class="legal-link"
        href="https://github.com/sahajpatel123/conduraapp/blob/main/SECURITY.md"
        target="_blank"
        rel="noreferrer"
      >
        <Surface variant="raised" padding="4" radius="md" class="legal-surface">
          {#snippet children()}
            <Stack gap="1">
              <span class="legal-name">SECURITY.md</span>
              <span class="legal-sub">Disclosure policy</span>
            </Stack>
          {/snippet}
        </Surface>
      </a>
    </Inline>
  </Stack>
</Stack>

<style>
  :global(.about-page) {
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide);
    margin: 0 auto;
    background-color: var(--surface-base);
  }

  /* ── Section heads ──────────────────────────────── */
  :global(.section) {
    animation: about-fade-in var(--duration-slow) var(--ease-decelerate) both;
  }
  :global(.section:nth-of-type(1)) { animation-delay: 40ms; }
  :global(.section:nth-of-type(2)) { animation-delay: 120ms; }
  :global(.section:nth-of-type(3)) { animation-delay: 200ms; }
  :global(.section:nth-of-type(4)) { animation-delay: 280ms; }
  :global(.section:nth-of-type(5)) { animation-delay: 360ms; }
  :global(.section:nth-of-type(6)) { animation-delay: 440ms; }

  @keyframes about-fade-in {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    :global(.section) {
      animation: none;
    }
  }

  .section-head {
    margin: 0;
  }

  .display-h2 {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    font-weight: var(--text-h2-weight);
    letter-spacing: var(--text-h2-tracking);
    line-height: var(--text-h2-leading);
    color: var(--content-primary);
    margin: 0 0 var(--space-2) 0;
  }

  .lede {
    font-size: var(--text-body-size);
    line-height: var(--text-body-leading);
    color: var(--content-tertiary);
    max-width: 640px;
    margin: 0;
  }

  /* ── Mission body ────────────────────────────────── */
  .mission-body {
    font-family: var(--font-serif);
    font-size: var(--text-body-lg-size);
    line-height: var(--text-body-lg-leading);
    color: var(--content-primary);
    margin: 0;
  }

  /* ── Invariant grid ──────────────────────────────── */
  .invariant-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: var(--space-4);
  }

  :global(.inv-card) {
    height: 100%;
  }

  .inv-num {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-accent);
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .inv-title {
    font-family: var(--font-sans);
    font-size: var(--text-body-size);
    font-weight: 600;
    color: var(--content-primary);
    line-height: var(--text-body-leading);
    margin: 0;
  }

  .inv-body {
    font-size: var(--text-body-sm-size);
    line-height: var(--text-body-sm-leading);
    color: var(--content-tertiary);
    margin: 0;
  }

  /* ── Armor modules ──────────────────────────────── */
  .armor-row {
    display: grid;
    grid-template-columns: 56px 1fr auto;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    transition: background-color var(--duration-fast) var(--ease-standard);
  }

  .armor-row:hover {
    background-color: var(--surface-raised);
  }

  .armor-code {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-accent);
    letter-spacing: 0.04em;
  }

  .armor-text h4 {
    font-size: var(--text-body-size);
    font-weight: 600;
    color: var(--content-primary);
    margin: 0 0 2px 0;
  }

  .armor-text p {
    font-size: var(--text-body-sm-size);
    line-height: var(--text-body-sm-leading);
    color: var(--content-tertiary);
    margin: 0;
  }

  .armor-pkg {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-muted);
    background-color: var(--surface-base);
    border: 1px solid var(--border-subtle);
    padding: 3px var(--space-2);
    border-radius: var(--radius-sm);
  }

  .mono {
    font-family: var(--font-mono);
  }

  /* ── Team ───────────────────────────────────────── */
  .team-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: var(--space-6);
  }

  .team-avatar {
    width: 48px;
    height: 48px;
    border-radius: var(--radius-pill);
    background-color: var(--plum-600);
    color: var(--content-on-accent);
    border: 1px solid var(--border-default);
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-mono);
    font-size: var(--text-body-sm-size);
    font-weight: 600;
    letter-spacing: 0.04em;
  }

  .team-avatar.ai {
    background-color: var(--status-info-bg);
    color: var(--status-info-fg);
    border-color: var(--status-info-border);
  }

  .team-name {
    font-size: var(--text-body-size);
    font-weight: 600;
    color: var(--content-primary);
    margin: 0;
  }

  .team-role {
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.06em;
    margin: 0;
  }

  .team-bio {
    font-size: var(--text-body-sm-size);
    line-height: var(--text-body-sm-leading);
    color: var(--content-tertiary);
    margin: 0;
  }

  /* ── Legal ──────────────────────────────────────── */
  :global(.legal-links) {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  }

  .legal-link {
    display: block;
    text-decoration: none;
    color: inherit;
    transition:
      border-color var(--duration-fast) var(--ease-standard),
      background-color var(--duration-fast) var(--ease-standard);
  }

  .legal-link:hover :global(.legal-surface) {
    background-color: var(--surface-overlay);
    border-color: var(--border-strong);
  }

  .legal-link:focus-visible {
    outline: var(--border-focus-width, 2px) solid var(--border-focus);
    outline-offset: var(--focus-ring-offset);
    border-radius: var(--radius-md);
  }

  :global(.legal-surface) {
    height: 100%;
  }

  .legal-name {
    font-family: var(--font-mono);
    font-size: var(--text-body-sm-size);
    color: var(--content-link);
  }

  .legal-sub {
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
  }
</style>
