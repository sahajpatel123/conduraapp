<script lang="ts">
  /**
   * PermissionCards — OS permission toggles.
   * Each permission is a card with status and action.
   */
  import { InkText, WordReveal, BlurReveal, PaperCard, PulseDot, InkReveal } from '$lib/components/living'

  interface Permission {
    id: string
    label: string
    description: string
    icon: string
    granted: boolean
  }

  interface Props {
    onnext: () => void
    onskip: () => void
    permissions?: Permission[]
  }

  let {
    onnext,
    onskip,
    permissions = [
      { id: 'accessibility', label: 'Accessibility', description: 'Allow Condura to observe window focus for context-aware assistance', icon: '👁', granted: false },
      { id: 'screen', label: 'Screen Recording', description: 'Allow Condura to capture screenshots for computer-use features', icon: '🖥', granted: false },
      { id: 'microphone', label: 'Microphone', description: 'Allow voice interaction and wake-word detection', icon: '🎤', granted: false },
      { id: 'notifications', label: 'Notifications', description: 'Show alerts when agents need your attention', icon: '🔔', granted: false },
    ],
  }: Props = $props()
</script>

<div style="max-width: 520px; margin: 0 auto; text-align: center;">
  <InkReveal direction="left" duration={900} delay={200}>
    <InkText kind="display" as="h1" style="margin-bottom: var(--lp-space-3);">
      <WordReveal text="Permissions" stagger={50} delay={300} />
    </InkText>
  </InkReveal>

  <BlurReveal delay={500} distance={16}>
    <InkText kind="body" tone="ink-mute" style="max-width: 400px; margin: 0 auto var(--lp-space-6);">
      Condura needs a few permissions to work fully. You can change these anytime.
    </InkText>
  </BlurReveal>

  <BlurReveal delay={700} distance={16}>
    <div style="display: flex; flex-direction: column; gap: var(--lp-space-3); max-width: 440px; margin: 0 auto;">
      {#each permissions as perm, i}
        <BlurReveal delay={800 + i * 100} distance={12}>
          <PaperCard border={perm.granted ? 'synapse' : 'none'} padding="var(--lp-space-3) var(--lp-space-4)">
            <div style="display: flex; align-items: center; gap: var(--lp-space-3);">
              <span style="font-size: 20px;">{perm.icon}</span>
              <div style="text-align: left; flex: 1;">
                <div style="display: flex; align-items: center; gap: var(--lp-space-2);">
                  <InkText kind="title" as="div">{perm.label}</InkText>
                  {#if perm.granted}
                    <PulseDot phase="ok" size={5} />
                  {/if}
                </div>
                <InkText kind="caption" tone="ink-mute">{perm.description}</InkText>
              </div>
              <button
                type="button"
                class="lp-focus"
                style="
                  padding: 6px 14px;
                  border-radius: var(--lp-radius-sm);
                  border: 1px solid {perm.granted ? 'var(--lp-synapse)' : 'var(--lp-ink-ghost)'};
                  background: {perm.granted ? 'var(--lp-synapse)' : 'transparent'};
                  color: {perm.granted ? 'var(--lp-paper)' : 'var(--lp-ink-mute)'};
                  font-family: var(--lp-font-sans);
                  font-size: var(--lp-text-caption);
                  cursor: pointer;
                  white-space: nowrap;
                  transition: all var(--lp-dur-fast) var(--lp-ease-thread);
                "
              >
                {perm.granted ? 'Granted' : 'Grant'}
              </button>
            </div>
          </PaperCard>
        </BlurReveal>
      {/each}
    </div>
  </BlurReveal>

  <BlurReveal delay={1200} distance={16}>
    <div style="display: flex; align-items: center; justify-content: center; gap: var(--lp-space-4); margin-top: var(--lp-space-8);">
      <button
        type="button"
        class="lp-focus"
        onclick={onskip}
        style="
          padding: 10px 20px;
          border-radius: var(--lp-radius-sm);
          border: 1px solid var(--lp-ink-ghost);
          background: transparent;
          color: var(--lp-ink-mute);
          font-family: var(--lp-font-sans);
          font-size: var(--lp-text-body);
          cursor: pointer;
          transition: all var(--lp-dur-fast) var(--lp-ease-thread);
        "
      >Skip for now</button>
      <MagneticButton variant="primary" size="md" onclick={onnext}>
        Continue
      </MagneticButton>
    </div>
  </BlurReveal>
</div>

<style>
  button:hover {
    opacity: 0.85;
  }
</style>
