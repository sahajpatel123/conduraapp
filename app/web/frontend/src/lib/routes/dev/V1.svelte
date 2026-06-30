<!--
  /dev/v1 — full app shell preview for the v1 redesign.

  This is the design system's living app. Every surface is wired:
    - Sidebar (compact nav) on the left
    - Main area: Chat (default) | Settings | Design tokens
    - StatusBar (top-right, OS chrome style)
    - CommandSurface overlay (Cmd+K)
    - ConversationDrawer (swipe-right or button)
    - Onboarding wizard launcher (button in main area)

  Keyboard shortcuts:
    Cmd/Ctrl+K — open command surface
    Cmd/Ctrl+\ — toggle sidebar
    Esc — close any overlay
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import Sidebar, { type RouteId } from '$components/v1/Sidebar.svelte';
  import StatusBar from '$components/v1/StatusBar.svelte';
  import CommandSurface from '$components/v1/CommandSurface.svelte';
  import ChatSurface from '$components/v1/ChatSurface.svelte';
  import ConversationDrawer from '$components/v1/ConversationDrawer.svelte';
  import SettingsPane from '$components/v1/SettingsPane.svelte';
  import OnboardingWizard from '$components/v1/onboarding/OnboardingWizard.svelte';
  import Hairline from '$components/v1/Hairline.svelte';
  import Button from '$components/v1/Button.svelte';
  import Pulse from '$components/v1/Pulse.svelte';
  import Stack from '$components/v1/Stack.svelte';
  import Inline from '$components/v1/Inline.svelte';
  import Switch from '$components/v1/Switch.svelte';
  import Card from '$components/v1/Card.svelte';
  import Surface from '$components/v1/Surface.svelte';
  import Pill from '$components/v1/Pill.svelte';

  type Route = RouteId | 'home';

  let activeRoute = $state<Route>('chat');
  let sidebarCollapsed = $state(false);
  let commandOpen = $state(false);
  let drawerOpen = $state(false);
  let onboardingOpen = $state(false);
  let commandMode = $state<'idle' | 'active' | 'processing' | 'result'>('idle');

  // Mock state
  let agentState = $state<'idle' | 'thinking' | 'awaiting' | 'error'>('idle');
  let activeTask = $state<string | null>(null);
  let queuedCount = $state(0);

  // Dark mode toggle (for showing the dark variant)
  let darkMode = $state(false);

  function handleKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      commandOpen = !commandOpen;
    } else if ((e.metaKey || e.ctrlKey) && e.key === '\\') {
      e.preventDefault();
      sidebarCollapsed = !sidebarCollapsed;
    } else if (e.key === 'Escape') {
      commandOpen = false;
      drawerOpen = false;
    }
  }

  onMount(() => {
    document.documentElement.setAttribute('data-mode', darkMode ? 'dark' : 'light');
  });

  $effect(() => {
    document.documentElement.setAttribute('data-mode', darkMode ? 'dark' : 'light');
  });

  // Sample conversation
  const sampleTurns = [
    { id: '1', role: 'user' as const, content: 'Find the typo in my open VS Code file and fix it.', timestamp: '14:22:01', status: 'done' as const },
    { id: '2', role: 'agent' as const, content: "I see you're editing `app/services/billing.ts`. Looking for the typo on line 142 — you wrote `recieve` instead of `receive`. May I correct it?", timestamp: '14:22:07', status: 'done' as const },
    { id: '3', role: 'user' as const, content: 'Yes, fix it.', timestamp: '14:22:18', status: 'done' as const },
    { id: '4', role: 'agent' as const, content: "Done. I changed `recieve` to `receive` and saved the file. Your git diff shows one character changed.", timestamp: '14:22:23', status: 'done' as const },
    { id: '5', role: 'user' as const, content: 'Summarize the last 3 emails from the design team.', timestamp: '14:25:42', status: 'done' as const },
    { id: '6', role: 'agent' as const, content: 'I read the last 3 emails from design@team.co. Two were threads about the Q3 launch (one announcement, one feedback request), and one was a calendar invite for Friday review. Want me to draft a reply to the feedback thread?', timestamp: '14:25:48', status: 'done' as const },
  ];

  const sampleConvos = [
    { id: '1', date: 'Today 14:25', firstSentence: 'Summarize the last 3 emails from the design team.', agentActed: true, active: true },
    { id: '2', date: 'Today 14:22', firstSentence: 'Find the typo in my open VS Code file and fix it.', agentActed: true },
    { id: '3', date: 'Yesterday', firstSentence: "What's on my calendar for tomorrow at 2pm?", agentActed: false },
    { id: '4', date: '2026-06-28', firstSentence: 'Translate the README to Spanish.', agentActed: true },
    { id: '5', date: '2026-06-27', firstSentence: 'Open Safari and search for the latest Linear pricing.', agentActed: true },
  ];

  const sampleInterpretations = [
    { interpretation: 'Open the latest report', steps: 'Open Finder → Documents → Q3-report.pdf' },
    { interpretation: 'Search the web for "Linear pricing"', steps: 'Open Safari → search → return top results' },
    { interpretation: 'Show my calendar for tomorrow', steps: 'Open Calendar → scroll to tomorrow → list events' },
  ];

  function openOnboarding() {
    onboardingOpen = true;
  }

  function closeOnboarding() {
    onboardingOpen = false;
  }

  function simulateAgentWorking() {
    agentState = 'thinking';
    activeTask = 'summarizing Q3-report.pdf';
    queuedCount = 2;
    commandMode = 'processing';
    commandOpen = true;
    setTimeout(() => {
      agentState = 'idle';
      activeTask = null;
      queuedCount = 0;
      commandMode = 'result';
    }, 3000);
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<svelte:head>
  <title>v1 app shell · Synaptic</title>
</svelte:head>

<div class="shell" data-mode={darkMode ? 'dark' : 'light'}>
  {#if onboardingOpen}
    <OnboardingWizard oncomplete={closeOnboarding} />
  {:else}
    <!-- ── Sidebar ───────────────────────────────────────────── -->
    <Sidebar
      active={activeRoute === 'home' ? 'chat' : activeRoute}
      collapsed={sidebarCollapsed}
      onnavigate={(r) => { activeRoute = r; }}
      ontoggle={() => sidebarCollapsed = !sidebarCollapsed}
    />

    <!-- ── Main area ────────────────────────────────────────── -->
    <main class="main">
      <!-- Top bar with title and actions -->
      <header class="main__topbar">
        <div class="main__topbar-left">
          <Pulse state="idle" size="sm" label="Synaptic" />
          <span class="main__topbar-title">
            {activeRoute === 'chat' ? 'Chat' :
             activeRoute === 'settings' ? 'Settings' :
             activeRoute === 'home' ? 'Home' :
             activeRoute === 'skills' ? 'Skills' :
             activeRoute === 'hub' ? 'Hub' :
             'Audit'}
          </span>
        </div>
        <Inline gap="2">
          <Button size="sm" variant="tertiary" onclick={() => drawerOpen = !drawerOpen}>
            History
          </Button>
          <Button size="sm" variant="secondary" onclick={() => { commandOpen = true; commandMode = 'idle'; }}>
            ⌘K Command
          </Button>
          <Switch label="Dark" checked={darkMode} onchange={(v) => darkMode = v} />
        </Inline>
      </header>

      <Hairline />

      <!-- Route content -->
      <div class="main__content">
        {#if activeRoute === 'chat'}
          <ChatSurface turns={sampleTurns} />
        {:else if activeRoute === 'settings'}
          <SettingsPane />
        {:else if activeRoute === 'home'}
          <div class="home">
            <Stack gap="6">
              <header class="home__head">
                <Pulse state="idle" size="xl" label="Synaptic" />
                <h1>Welcome back.</h1>
                <p>What would you like to do today?</p>
              </header>

              <Inline gap="3">
                <Card title="Open chat" description="Pick up where you left off.">
                  {#snippet actions()}
                    <Button size="sm" variant="primary" onclick={() => activeRoute = 'chat'}>Go →</Button>
                  {/snippet}
                </Card>
                <Card title="Re-run onboarding" description="Walk through the 5 screens again. Useful for design review.">
                  {#snippet actions()}
                    <Button size="sm" variant="secondary" onclick={openOnboarding}>Start →</Button>
                  {/snippet}
                </Card>
                <Card title="Simulate agent working" description="See the command surface in processing state.">
                  {#snippet actions()}
                    <Button size="sm" variant="secondary" onclick={simulateAgentWorking}>Run →</Button>
                  {/snippet}
                </Card>
              </Inline>

              <Surface variant="sunken" padding="6" radius="md">
                <Stack gap="3">
                  <div class="home__kbds">
                    <span class="caption">Keyboard shortcuts</span>
                    <Inline gap="4">
                      <span><kbd>⌘K</kbd> command surface</span>
                      <span><kbd>⌘\</kbd> toggle sidebar</span>
                      <span><kbd>esc</kbd> close overlay</span>
                    </Inline>
                  </div>
                </Stack>
              </Surface>
            </Stack>
          </div>
        {:else}
          <div class="placeholder">
            <Pill variant="neutral" label="Coming soon" />
            <p>The {activeRoute} view uses v1 primitives — wire it next.</p>
          </div>
        {/if}
      </div>
    </main>

    <!-- ── StatusBar (top-right, OS chrome style) ──────────── -->
    <StatusBar
      {activeTask}
      {queuedCount}
      {agentState}
      onopen={() => { commandOpen = true; commandMode = 'idle'; }}
      onpause={() => { agentState = 'idle'; activeTask = null; }}
    />

    <!-- ── Command Surface overlay ──────────────────────────── -->
    {#if commandOpen}
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="overlay-scrim" onclick={() => commandOpen = false} onkeydown={(e) => { if (e.key === 'Escape') commandOpen = false; }} aria-hidden="true"></div>
      <div class="overlay-surface" role="dialog" aria-label="Synaptic command">
        <CommandSurface
          mode={commandMode}
          contextChips={commandMode !== 'idle' ? [
            { label: 'this Slack thread' },
            { label: 'this file' },
          ] : []}
          interpretations={commandMode === 'active' ? sampleInterpretations : []}
          progress={commandMode === 'processing' ? {
            elapsedMs: 2400,
            state: 'thinking',
            modelName: 'claude-sonnet-4-6'
          } : undefined}
          result={commandMode === 'result' ? {
            verb: 'summarized',
            target: '3 emails from design@team.co · 0.4s',
            timestamp: 'just now',
            state: 'done'
          } : undefined}
          onsubmit={() => { commandMode = 'processing'; }}
          onselect={() => { commandMode = 'processing'; }}
          onpause={() => { commandMode = 'idle'; agentState = 'idle'; }}
        />
      </div>
    {/if}

    <!-- ── Conversation Drawer ─────────────────────────────── -->
    <ConversationDrawer
      conversations={sampleConvos}
      open={drawerOpen}
      onclose={() => drawerOpen = false}
    />
  {/if}
</div>

<style>
  .shell {
    display: flex;
    height: 100vh;
    background-color: var(--surface-base);
    color: var(--content-primary);
    font-family: var(--font-sans);
    overflow: hidden;
  }

  /* Main area */
  .main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    overflow: hidden;
  }

  .main__topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-3) var(--space-5);
    height: 56px;
    flex-shrink: 0;
    background-color: var(--surface-raised);
  }

  .main__topbar-left {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .main__topbar-title {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-tertiary);
  }

  .main__content {
    flex: 1;
    overflow-y: auto;
    background-color: var(--surface-base);
  }

  /* Home view */
  .home {
    max-width: 800px;
    margin: 0 auto;
    padding: var(--space-9) var(--space-6);
  }

  .home__head {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-3);
    margin-bottom: var(--space-7);
  }

  .home__head h1 {
    font-family: var(--font-serif);
    font-size: var(--text-display-lg-size);
    font-weight: 400;
    color: var(--content-primary);
    margin: 0;
    line-height: 1.15;
  }

  .home__head p {
    font-family: var(--font-sans);
    font-size: var(--text-body-lg-size);
    color: var(--content-tertiary);
    margin: 0;
  }

  .home__kbds {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .home__kbds > span:first-child {
    color: var(--content-tertiary);
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .home__kbds > .inline > span {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-secondary);
  }

  kbd {
    font-family: var(--font-mono);
    font-size: 0.9em;
    background-color: var(--paper-warm-50);
    border: 1px solid var(--border-default);
    padding: 1px 6px;
    border-radius: var(--radius-xs);
    color: var(--content-primary);
    margin-right: var(--space-2);
  }

  /* Placeholder for unimplemented routes */
  .placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    padding: var(--space-13) var(--space-6);
    text-align: center;
  }
  .placeholder p {
    font-family: var(--font-sans);
    color: var(--content-tertiary);
    margin: 0;
    font-size: var(--text-body-size);
  }

  /* Overlays */
  .overlay-scrim {
    position: fixed;
    inset: 0;
    background-color: rgba(14, 16, 20, 0.16);
    backdrop-filter: blur(2px);
    z-index: var(--z-overlay);
    animation: scrim-in var(--duration-base) var(--ease-accelerate) both;
  }

  @keyframes scrim-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .overlay-surface {
    position: fixed;
    top: 18%;
    left: 50%;
    transform: translateX(-50%);
    z-index: calc(var(--z-overlay) + 1);
    animation: surface-pop var(--duration-base) var(--ease-decelerate) both;
    transform-origin: top center;
  }

  @keyframes surface-pop {
    from {
      opacity: 0;
      transform: translateX(-50%) scale(0.96);
    }
    to {
      opacity: 1;
      transform: translateX(-50%) scale(1);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .overlay-scrim,
    .overlay-surface {
      animation: none;
    }
  }
</style>