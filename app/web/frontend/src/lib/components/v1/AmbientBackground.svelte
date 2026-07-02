<!--
  AmbientBackground — the "alive factor" for the entire app shell.

  A fixed full-viewport layer that sits behind every surface. It has
  two subtle motions:
    1. A barely-perceptible luminance drift (the "breath of the room")
    2. A faint radial plum tint that follows the agent's Pulse state

  Per spec §6 (Selective Perception) + motion agent §6.5:
    - Slow, considered, never distracting
    - Reads as "the app is alive" without competing with content
    - Reduces to a static layer in low-energy / reduced-motion modes
    - Plum appears in ≤5% of any screen (this layer is the only place
      where plum is allowed as ambient color, since it IS the brand)

  Used once in App.svelte, mounted behind all other content.

  Props:
    agentState — 'idle' | 'thinking' | 'awaiting' | 'error'
                  controls the radial tint intensity and position
-->
<script lang="ts">
  type AgentState = 'idle' | 'thinking' | 'awaiting' | 'error';

  interface Props {
    agentState?: AgentState;
  }

  let { agentState = 'idle' }: Props = $props();

  // Tint intensity per state — idle is barely there, thinking is subtle,
  // awaiting is more visible (the agent is looking at you), error is red
  let tintOpacity = $derived(
    agentState === 'idle' ? 0.04 :
    agentState === 'thinking' ? 0.08 :
    agentState === 'awaiting' ? 0.14 :
    0.0  // error handled separately
  );

  let tintColor = $derived(
    agentState === 'error' ? 'var(--error-500)' : 'var(--content-accent)'
  );

  // Drift phase — slightly different per "breath cycle" so the
  // background isn't perfectly synced to the Pulse (organic feel)
  let driftPhase = $derived(
    agentState === 'thinking' ? '-8s' :
    agentState === 'awaiting' ? '-3s' :
    '0s'
  );
</script>

<div class="ambient" data-state={agentState} aria-hidden="true">
  <!-- The luminance drift — a near-imperceptible lightness wave -->
  <div class="ambient__drift" style="animation-delay: {driftPhase};"></div>

  <!-- The radial plum tint — concentrated in one corner, follows the agent -->
  <div
    class="ambient__tint"
    style="background: radial-gradient(ellipse 50% 50% at 80% 20%, {tintColor} 0%, transparent 70%); opacity: {tintOpacity};"
  ></div>

  <!-- A second, even fainter tint at the opposite corner — gives the
       ambient a sense of depth without being noticeable as a specific element -->
  <div
    class="ambient__tint-2"
    style="background: radial-gradient(ellipse 40% 40% at 15% 85%, var(--content-accent) 0%, transparent 70%);"
  ></div>

  <!-- A subtle vignette at the edges — frames the content without walls -->
  <div class="ambient__vignette"></div>
</div>

<style>
  .ambient {
    position: fixed;
    inset: 0;
    z-index: 0;
    pointer-events: none;
    overflow: hidden;
    background-color: var(--surface-base);
  }

  /* The luminance drift — barely visible 30s wave */
  .ambient__drift {
    position: absolute;
    inset: 0;
    background-color: var(--surface-base);
    animation: drift 30s ease-in-out infinite;
  }

  @keyframes drift {
    0%, 100% { opacity: 1; }
    25% { opacity: 0.985; }
    50% { opacity: 0.97; }
    75% { opacity: 0.985; }
  }

  /* The plum tints — fixed positions, opacity controlled per state */
  .ambient__tint,
  .ambient__tint-2 {
    position: absolute;
    inset: 0;
    transition: opacity 1.2s var(--ease-decelerate);
    mix-blend-mode: multiply;
  }

  /* The second tint — always at half intensity */
  .ambient__tint-2 {
    opacity: 0.5;
    transition: opacity 1.2s var(--ease-decelerate);
  }

  .ambient[data-state="thinking"] .ambient__tint-2,
  .ambient[data-state="awaiting"] .ambient__tint-2 {
    opacity: 0.8;
  }

  /* The vignette — very faint edge darkening */
  .ambient__vignette {
    position: absolute;
    inset: 0;
    background: radial-gradient(ellipse 100% 80% at center, transparent 50%, rgba(14, 16, 20, 0.06) 100%);
    pointer-events: none;
  }

  /* Dark mode: the drift and tints invert slightly */
  [data-mode="dark"] .ambient {
    background-color: var(--surface-base);
  }

  [data-mode="dark"] .ambient__drift {
    /* Slight luminance lift in dark mode — the background is breathing brighter */
    animation-name: drift-dark;
  }

  @keyframes drift-dark {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.94; }
  }

  /* Reduced motion: no drift, but tints remain (they're static color, not motion) */
  @media (prefers-reduced-motion: reduce) {
    .ambient__drift {
      animation: none;
    }
  }
</style>