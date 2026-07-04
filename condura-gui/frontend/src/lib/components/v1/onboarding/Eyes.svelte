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
  import Icon, { type IconName } from '../icons/Icon.svelte';

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
        <div class="panel__diagram panel__diagram--accessibility" aria-hidden="true">
          <!-- A window with named structural elements being recognized.
               The "eye" Pulse is over the window — observing structure. -->
          <div class="diagram-window">
            <div class="diagram-window__chrome">
              <span class="diagram-window__dot"></span>
              <span class="diagram-window__dot"></span>
              <span class="diagram-window__dot"></span>
              <span class="diagram-window__title">Mail · Inbox</span>
            </div>
            <div class="diagram-window__body">
              <div class="diagram-row" data-label="Send">
                <span class="diagram-row__name">Send</span>
                <span class="diagram-row__tag">button</span>
                <span class="diagram-row__check">
                  <Icon name="check" size="xs" />
                </span>
              </div>
              <div class="diagram-row" data-label="To">
                <span class="diagram-row__name">To</span>
                <span class="diagram-row__value">sam@team.co</span>
                <span class="diagram-row__tag">field</span>
                <span class="diagram-row__check">
                  <Icon name="check" size="xs" />
                </span>
              </div>
              <div class="diagram-row" data-label="Subject">
                <span class="diagram-row__name">Subject</span>
                <span class="diagram-row__value">Q3 launch</span>
                <span class="diagram-row__tag">field</span>
                <span class="diagram-row__check">
                  <Icon name="check" size="xs" />
                </span>
              </div>
            </div>
            <div class="diagram-eye">
              <Pulse state="idle" size="sm" label="" />
            </div>
          </div>
        </div>
        <h3 class="panel__name">Accessibility</h3>
        <p class="panel__desc">
          I read <em>structure</em> — named buttons, fields, window titles. Not pixels.
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
          icon={accessibilityGranted ? 'check' : 'eye'}
          onclick={handleGrantAccessibility}
          loading={grantingAccessibility}
          disabled={accessibilityGranted}
        >
          {accessibilityGranted ? 'Granted ✓' : 'Grant on this Mac'}
        </Button>
      </article>

      <!-- Screen Recording panel -->
      <article class="panel">
        <div class="panel__diagram panel__diagram--screen" aria-hidden="true">
          <!-- A screen rectangle with a sparse, sampling pattern.
               The Pulse appears intermittently — sampled, not continuous. -->
          <div class="screen-rect">
            <div class="screen-rect__topbar"></div>
            <div class="screen-rect__content">
              <div class="screen-rect__line"></div>
              <div class="screen-rect__line screen-rect__line--short"></div>
              <div class="screen-rect__line"></div>
              <div class="screen-rect__line screen-rect__line--shorter"></div>
            </div>
            <!-- The sampling dot — appears, samples, disappears -->
            <div class="screen-sampler">
              <span class="screen-sampler__ring"></span>
              <span class="screen-sampler__ring screen-sampler__ring--delay"></span>
              <span class="screen-sampler__core"></span>
            </div>
          </div>
        </div>
        <h3 class="panel__name">Screen Recording</h3>
        <p class="panel__desc">
          I <em>sample</em> the screen occasionally when needed. Not continuously.
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
          icon={screenRecordingGranted ? 'check' : 'eye'}
          onclick={handleGrantScreenRecording}
          loading={grantingScreenRecording}
          disabled={screenRecordingGranted}
        >
          {screenRecordingGranted ? 'Granted' : 'Grant on this Mac'}
        </Button>
      </article>
    </div>

    <footer class="screen__footer">
      <p>You can revoke either of these at any time. I will stop the moment you do.</p>
    </footer>

    <div class="screen__actions">
      <Button variant="tertiary" size="md" icon="arrow-left" onclick={onback}>Back</Button>
      <div class="screen__actions-right">
        {#if !accessibilityGranted && !screenRecordingGranted}
          <Button variant="tertiary" size="md" onclick={onskip}>
            <Pulse state="idle" size="sm" label="" /> Continue in limited mode
          </Button>
        {/if}
        <Button variant="primary" size="md" icon="arrow-right" iconPosition="right" onclick={oncontinue}>
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

  /* ── Accessibility diagram — a window with named elements ─── */
  .panel__diagram--accessibility {
    height: auto;
    padding: var(--space-4);
  }

  .diagram-window {
    background-color: var(--paper-warm-0);
    border: 1.25px solid var(--ink-cool-100);
    border-radius: var(--radius-sm);
    overflow: hidden;
    position: relative;
    box-shadow: 0 1px 2px rgba(14, 16, 20, 0.04);
  }

  .diagram-window__chrome {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1) var(--space-2);
    background-color: var(--paper-warm-50);
    border-bottom: 1px solid var(--ink-cool-100);
  }

  .diagram-window__dot {
    width: 6px;
    height: 6px;
    border-radius: var(--radius-pill);
    background-color: var(--ink-cool-200);
  }

  .diagram-window__title {
    margin-left: auto;
    margin-right: auto;
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--content-tertiary);
    letter-spacing: 0.04em;
  }

  .diagram-window__body {
    display: flex;
    flex-direction: column;
    padding: var(--space-2);
  }

  .diagram-row {
    display: grid;
    grid-template-columns: 40px 1fr auto auto;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-xs);
    font-family: var(--font-mono);
    font-size: 9px;
    color: var(--content-secondary);
  }

  .diagram-row__name {
    color: var(--content-tertiary);
    letter-spacing: 0.02em;
  }

  .diagram-row__value {
    color: var(--content-primary);
    font-family: var(--font-sans);
    font-size: 10px;
  }

  .diagram-row__tag {
    font-family: var(--font-mono);
    font-size: 8px;
    color: var(--plum-700);
    background-color: var(--plum-50);
    padding: 1px 4px;
    border-radius: var(--radius-xs);
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .diagram-row__check {
    color: var(--success-500);
    display: inline-flex;
  }

  /* The "eye" — the agent observing the window */
  .diagram-eye {
    position: absolute;
    top: var(--space-2);
    right: var(--space-2);
    background-color: var(--paper-warm-0);
    border-radius: var(--radius-pill);
    padding: 2px;
    display: inline-flex;
    box-shadow: 0 0 0 2px rgba(110, 58, 255, 0.1);
  }

  /* ── Screen Recording diagram — sampled, not continuous ──── */
  .panel__diagram--screen {
    height: 140px;
    padding: var(--space-4);
    position: relative;
  }

  .screen-rect {
    width: 100%;
    height: 100%;
    border: 1.25px solid var(--ink-cool-100);
    border-radius: var(--radius-sm);
    background-color: var(--paper-warm-0);
    position: relative;
    overflow: hidden;
    box-shadow: 0 1px 2px rgba(14, 16, 20, 0.04);
  }

  .screen-rect__topbar {
    height: 8px;
    background-color: var(--paper-warm-50);
    border-bottom: 1px solid var(--ink-cool-100);
  }

  .screen-rect__content {
    padding: var(--space-3);
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .screen-rect__line {
    height: 4px;
    background-color: var(--ink-cool-100);
    border-radius: var(--radius-xs);
    width: 100%;
  }

  .screen-rect__line--short {
    width: 70%;
  }

  .screen-rect__line--shorter {
    width: 50%;
  }

  /* The sampler — appears, samples, disappears */
  .screen-sampler {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .screen-sampler__core {
    width: 8px;
    height: 8px;
    border-radius: var(--radius-pill);
    background-color: var(--plum-600);
    z-index: 2;
  }

  .screen-sampler__ring {
    position: absolute;
    width: 16px;
    height: 16px;
    border-radius: var(--radius-pill);
    border: 1.5px solid var(--plum-500);
    animation: sample-ring 4s ease-out infinite;
  }

  .screen-sampler__ring--delay {
    animation-delay: 1.5s;
  }

  @keyframes sample-ring {
    0% {
      transform: scale(0.6);
      opacity: 0.8;
    }
    100% {
      transform: scale(2.2);
      opacity: 0;
    }
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

  .panel__desc em {
    font-style: italic;
    color: var(--content-secondary);
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