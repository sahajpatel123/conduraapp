<script lang="ts">
  /**
   * FirstBreath — The closing moment of onboarding.
   * A beautiful completion animation: the synapse thread draws,
   * words reveal, and Condura comes alive.
   */
  import { InkText, WordReveal, SynapseThread, PulseDot, BlurReveal, InkReveal } from '$lib/components/living'

  interface Props {
    oncomplete: () => void
  }

  let { oncomplete }: Props = $props()
</script>

<div style="
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: var(--lp-space-6);
  min-height: 300px;
  position: relative;
">
  <!-- Decorative thread that draws in -->
  <div style="width: 120px; height: 2px; margin-bottom: var(--lp-space-4);">
    <SynapseThread
      orientation="horizontal"
      length={120}
      animate={true}
      glow={true}
      duration={1500}
      color="var(--lp-synapse-glow)"
      width={1.5}
    />
  </div>

  <InkReveal direction="left" duration={1200} delay={500}>
    <InkText kind="display" as="h1" style="font-size: clamp(32px, 5vw, 52px);">
      <WordReveal text="Condura is here." stagger={80} delay={800} />
    </InkText>
  </InkReveal>

  <BlurReveal delay={1500} distance={16}>
    <div style="display: flex; flex-direction: column; align-items: center; gap: var(--lp-space-2);">
      <InkText kind="body" tone="ink-mute" style="max-width: 360px;">
        Your conductor is ready. Everything you need, one hotkey away.
      </InkText>
      <InkText kind="body-sm" tone="ink-faint" style="margin-top: var(--lp-space-1);">
        Free. Local. Private. Always.
      </InkText>
    </div>
  </BlurReveal>

  <BlurReveal delay={2200} distance={12}>
    <div style="display: flex; align-items: center; gap: var(--lp-space-3);">
      <PulseDot phase="ok" size={8} label="Ready" />
    </div>
  </BlurReveal>

  <BlurReveal delay={2800} distance={12}>
    <button
      type="button"
      onclick={oncomplete}
      class="lp-focus"
      style="
        margin-top: var(--lp-space-4);
        padding: 14px 36px;
        border-radius: var(--lp-radius-sm);
        border: none;
        background: var(--lp-synapse);
        color: var(--lp-paper);
        font-family: var(--lp-font-sans);
        font-size: var(--lp-text-body);
        font-weight: 500;
        cursor: pointer;
        transition: all var(--lp-dur-normal) var(--lp-ease-thread);
        box-shadow: var(--lp-shadow-raise);
      "
    >
      Begin
    </button>
  </BlurReveal>
</div>
