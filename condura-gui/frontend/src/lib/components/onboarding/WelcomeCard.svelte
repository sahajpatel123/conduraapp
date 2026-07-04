<script lang="ts">
  /**
   * WelcomeCard — First step of the floating onboarding.
   * Warm welcome with "What brings you here?" card selection.
   */
  import { InkText, WordReveal, BlurReveal, PaperCard, MagneticButton, PaperDivider, PollenSpark, InkReveal } from '$lib/components/living'

  interface Props {
    onnext: (intent: string) => void
  }

  let { onnext }: Props = $props()
</script>

<div style="
  max-width: 520px;
  margin: 0 auto;
  text-align: center;
">
  <InkReveal direction="left" duration={900} delay={200}>
    <InkText kind="display" as="h1" style="margin-bottom: var(--lp-space-4);">
      <WordReveal text="Welcome to Condura" stagger={60} delay={300} />
    </InkText>
  </InkReveal>

  <BlurReveal delay={600} distance={16}>
    <InkText kind="body" tone="ink-mute" style="max-width: 400px; margin: 0 auto var(--lp-space-8);">
      Your AI conductor — one hotkey to orchestrate every tool on your machine.
      Free. Local. Private. Let&rsquo;s get you set up.
    </InkText>
  </BlurReveal>

  <BlurReveal delay={800} distance={16}>
    <div style="display: flex; flex-direction: column; gap: var(--lp-space-3); max-width: 380px; margin: 0 auto;">
      <PaperCard border="pollen" padding="var(--lp-space-4) var(--lp-space-5)" onclick={() => onnext('ollama')}>
        <div style="display: flex; align-items: center; gap: var(--lp-space-3);">
          <span style="font-size: 24px;">🦙</span>
          <div style="text-align: left;">
            <InkText kind="title" as="div">I run Ollama locally</InkText>
            <InkText kind="caption" tone="ink-mute">Already have models on my machine</InkText>
          </div>
        </div>
      </PaperCard>

      <PaperCard border="none" padding="var(--lp-space-4) var(--lp-space-5)" onclick={() => onnext('apikey')}>
        <div style="display: flex; align-items: center; gap: var(--lp-space-3);">
          <PollenSpark size={6} phase="pulse" />
          <div style="text-align: left;">
            <InkText kind="title" as="div">I have API keys</InkText>
            <InkText kind="caption" tone="ink-mute">Claude, GPT, or other providers</InkText>
          </div>
        </div>
      </PaperCard>

      <PaperCard border="none" padding="var(--lp-space-4) var(--lp-space-5)" onclick={() => onnext('explore')}>
        <div style="display: flex; align-items: center; gap: var(--lp-space-3);">
          <span style="font-size: 24px;">🌱</span>
          <div style="text-align: left;">
            <InkText kind="title" as="div">Just exploring</InkText>
            <InkText kind="caption" tone="ink-mute">I&rsquo;ll configure providers later</InkText>
          </div>
        </div>
      </PaperCard>
    </div>
  </BlurReveal>
</div>
