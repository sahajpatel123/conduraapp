<!--
  Onboarding · Screen 3 · Eyes

  Mood: asking to see.
  Per spec §10.3: two side-by-side panels (Accessibility + Screen Recording).
  Each has a small diagram, a grant button, a live status dot.
  Footer: "You can revoke either of these at any time. I will stop
  the moment you do."

  Skip option: "Limited mode" — agent works for chat, file reading,
  web search only.
-->
<script lang="ts">
  import Button from '../Button.svelte';
  import Dot from '../Dot.svelte';
  import Pulse from '../Pulse.svelte';

  interface Props {
    accessibilityGranted?: boolean;
    screenRecordingGranted?: boolean;
    ongrantAccessibility?: () => Promise<void> | void;
    ongrantScreenRecording?: () => Promise<void> | void;
    oncontinue?: () => void;
    onskip?: () => void;
    onback?: () => void;
  }

  let {
    accessibilityGranted = false,
    screenRecordingGranted = false,
    ongrantAccessibility,
    ongrantScreenRecording,
    oncontinue,
    onskip,
    onback,
  }: Props = $props();

  let grantingAccessibility = $state(false);
  let grantingScreenRecording = $state(false);

  async function handleGrantAccessibility() {
    grantingAccessibility = true;
    try {
      await ongrantAccessibility?.();
    } finally {
      grantingAccessibility = false;
    }
  }

  async function handleGrantScreenRecording() {
    grantingScreenRecording = true;
    try {
      await ongrantScreenRecording?.();
    } finally {
      grantingScreenRecording = false;
    }
  }
</script>

<div class="screen">
  <div class="screen__inner">
    <header class="screen__header">
      <h2 class="screen__title">I'd like to see what you see.</h2>
      <p class="screen__subtitle">Two permissions. Each is revocable. I'll stop the moment you do.</p>
    </header>

    <div class="panels">
      <!-- Accessibility panel -->
      <article class="panel">
        <div class="panel__diagram" aria-hidden="true">
          <div class="diagram-row">
            <span class="diagram-name">Submit button</span>
            <span class="diagram-coord">↘</span>
          </div>
          <div class="diagram-row">
            <span class="diagram-name">Email field</span>
            <span class="diagram-coord">▢</span>
          </div>
          <div class="diagram-row">
            <span class="diagram-name">Window title</span>
            <span class="diagram-coord">⌗</span>
          </div>
        </div>
        <h3 class="panel__name">Accessibility</h3>
        <p class="panel__desc">
          Lets me perceive named buttons, window titles, and form fields. I see structure, not pixels.
        </p>
        <div class="panel__status">
          <Dot variant={accessibilityGranted ? 'success' : 'neutral'} size="sm" pulse={!accessibilityGranted} />
          <span class="panel__status-text">
            {accessibilityGranted ? 'Granted' : 'Not yet'}
          </span>
        </div>
        <Button
          variant={accessibilityGranted ? 'secondary' : 'primary'}
          size="md"
          onclick={handleGrantAccessibility}
          loading={grantingAccessibility}
          disabled={accessibilityGranted}
        >
          {accessibilityGranted ? 'Granted ✓' : 'Grant on this Mac'}
        </Button>
      </article>

      <!-- Screen Recording panel -->
      <article class="panel">
        <div class="panel__diagram" aria-hidden="true">
          <div class="screen-rect">
            <span class="screen-dot"></span>
            <span class="screen-dot"></span>
            <span class="screen-dot"></span>
          </div>
        </div>
        <h3 class="panel__name">Screen Recording</h3>
        <p class="panel__desc">
          Lets me sample the screen occasionally when needed. I do not record continuously.
        </p>
        <div class="panel__status">
          <Dot variant={screenRecordingGranted ? 'success' : 'neutral'} size="sm" pulse={!screenRecordingGranted} />
          <span class="panel__status-text">
            {screenRecordingGranted ? 'Granted' : 'Not yet'}
          </span>
        </div>
        <Button
          variant={screenRecordingGranted ? 'secondary' : 'primary'}
          size="md"
          onclick={handleGrantScreenRecording}
          loading={grantingScreenRecording}
          disabled={screenRecordingGranted}
        >
          {screenRecordingGranted ? 'Granted ✓' : 'Grant on this Mac'}
        </Button>
      </article>
    </div>

    <footer class="screen__footer">
      <p>You can revoke either of these at any time. I will stop the moment you do.</p>
    </footer>

    <div class="screen__actions">
      <Button variant="tertiary" size="md" onclick={onback}>← Back</Button>
      <div class="screen__actions-right">
        {#if !accessibilityGranted && !screenRecordingGranted}
          <Button variant="tertiary" size="md" onclick={onskip}>
            <Pulse state="idle" size="sm" label="" /> Continue in limited mode
          </Button>
        {/if}
        <Button variant="primary" size="md" onclick={oncontinue}>
          {accessibilityGranted || screenRecordingGranted ? 'Continue' : 'Continue (skip)'}
        </Button>
      </div>
    </div>
  </div>
</div>

<style>
  .screen {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: var(--space-9);
    background-color: var(--surface-base);
    color: var(--content-primary);
  }

  .screen__inner {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
    width: 100%;
    max-width: 880px;
  }

  .screen__header {
    text-align: left;
  }

  .screen__title {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    line-height: 1.3;
    font-weight: 400;
    color: var(--content-primary);
    margin: 0 0 var(--space-2) 0;
  }

  .screen__subtitle {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    margin: 0;
  }

  /* Two-column panel grid */
  .panels {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-4);
  }

  .panel {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    padding: var(--space-6);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-lg);
    transition: border-color var(--duration-fast) var(--ease-standard);
  }

  .panel:hover {
    border-color: var(--border-strong);
  }

  .panel__diagram {
    height: 96px;
    background-color: var(--surface-sunken);
    border-radius: var(--radius-sm);
    padding: var(--space-4);
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: var(--space-2);
  }

  .diagram-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-secondary);
    letter-spacing: 0.02em;
  }

  .diagram-coord {
    color: var(--content-accent);
    font-size: 14px;
  }

  .screen-rect {
    width: 100%;
    height: 64px;
    border: 1.25px solid var(--ink-cool-300);
    border-radius: var(--radius-xs);
    background-color: var(--paper-warm-0);
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
  }

  .screen-dot {
    width: 8px;
    height: 8px;
    border-radius: var(--radius-pill);
    background-color: var(--ink-cool-300);
  }

  .screen-dot:nth-child(2) {
    background-color: var(--content-accent);
    animation: sample-pulse 4s ease-in-out infinite;
  }

  @keyframes sample-pulse {
    0%, 100% { opacity: 0.4; }
    50% { opacity: 1; }
  }

  .panel__name {
    font-family: var(--font-sans);
    font-size: var(--text-h4-size);
    font-weight: 600;
    color: var(--content-primary);
    margin: 0;
  }

  .panel__desc {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    line-height: 1.5;
    margin: 0;
  }

  .panel__status {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .panel__status-text {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-secondary);
  }

  /* Footer — small print, the philosophy in nine words */
  .screen__footer {
    padding-top: var(--space-3);
  }

  .screen__footer p {
    font-family: var(--font-sans);
    font-size: var(--text-caption-size);
    color: var(--content-muted);
    margin: 0;
    line-height: 1.5;
  }

  .screen__actions {
    display: flex;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .screen__actions-right {
    display: flex;
    gap: var(--space-3);
    align-items: center;
  }

  @media (max-width: 720px) {
    .panels {
      grid-template-columns: 1fr;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .screen-dot:nth-child(2) {
      animation: none;
      opacity: 1;
    }
  }
</style>