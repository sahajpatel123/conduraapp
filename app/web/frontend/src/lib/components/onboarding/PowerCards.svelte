<script lang="ts">
  /**
   * PowerCards — Provider/power source selection.
   * User picks how they want to power the agent.
   */
  import { InkText, WordReveal, BlurReveal, PaperCard, MagneticButton, InkReveal } from '$lib/components/living'

  interface Props {
    onnext: (choice: string) => void
    onskip: () => void
  }

  let { onnext, onskip }: Props = $props()
</script>

<div style="max-width: 520px; margin: 0 auto; text-align: center;">
  <InkReveal direction="left" duration={900} delay={200}>
    <InkText kind="display" as="h1" style="margin-bottom: var(--lp-space-3);">
      <WordReveal text="Power Source" stagger={50} delay={300} />
    </InkText>
  </InkReveal>

  <BlurReveal delay={500} distance={16}>
    <InkText kind="body" tone="ink-mute" style="max-width: 380px; margin: 0 auto var(--lp-space-6);">
      Choose how you want to power Condura. You can change this later.
    </InkText>
  </BlurReveal>

  <BlurReveal delay={700} distance={16}>
    <div style="display: flex; flex-direction: column; gap: var(--lp-space-3); max-width: 380px; margin: 0 auto;">
      <PaperCard border="pollen" padding="var(--lp-space-4) var(--lp-space-5)" onclick={() => onnext('local')}>
        <div style="display: flex; align-items: center; gap: var(--lp-space-3);">
          <span style="font-size: 24px;">🦙</span>
          <div style="text-align: left;">
            <InkText kind="title" as="div">Local (Ollama)</InkText>
            <InkText kind="caption" tone="ink-mute">Free, private, runs on your machine</InkText>
          </div>
        </div>
      </PaperCard>

      <PaperCard border="none" padding="var(--lp-space-4) var(--lp-space-5)" onclick={() => onnext('cloud')}>
        <div style="display: flex; align-items: center; gap: var(--lp-space-3);">
          <span style="font-size: 24px;">☁️</span>
          <div style="text-align: left;">
            <InkText kind="title" as="div">Cloud API</InkText>
            <InkText kind="caption" tone="ink-mute">Claude, GPT, Gemini, or others</InkText>
          </div>
        </div>
      </PaperCard>

      <PaperCard border="none" padding="var(--lp-space-4) var(--lp-space-5)" onclick={() => onnext('later')}>
        <div style="display: flex; align-items: center; gap: var(--lp-space-3);">
          <span style="font-size: 24px;">⏳</span>
          <div style="text-align: left;">
            <InkText kind="title" as="div">Set up later</InkText>
            <InkText kind="caption" tone="ink-mute">I&rsquo;ll configure in Settings</InkText>
          </div>
        </div>
      </PaperCard>
    </div>
  </BlurReveal>

  <BlurReveal delay={1000} distance={16}>
    <div style="margin-top: var(--lp-space-6);">
      <button
        type="button"
        class="lp-focus"
        onclick={onskip}
        style="
          padding: 8px 16px;
          border: none;
          background: transparent;
          color: var(--lp-ink-mute);
          font-family: var(--lp-font-sans);
          font-size: var(--lp-text-body-sm);
          cursor: pointer;
        "
      >Skip — I&rsquo;ll figure it out</button>
    </div>
  </BlurReveal>
</div>
