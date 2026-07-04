<script lang="ts">
  /**
   * HotkeyCard — Key combo capture step.
   * User presses a key combination to set the global hotkey.
   */
  import { InkText, WordReveal, BlurReveal, MagneticButton, InkReveal } from '$lib/components/living'

  interface Props {
    onnext: (hotkey: string) => void
    onskip: () => void
  }

  let { onnext, onskip }: Props = $props()

  let recording = $state(false)
  let combo = $state('')
  let activeKeys = $state<Set<string>>(new Set())

  function startRecording() {
    recording = true
    combo = ''
    activeKeys = new Set()
  }

  function onKeydown(e: KeyboardEvent) {
    if (!recording) return
    e.preventDefault()

    const key = e.key === 'Meta' ? 'Cmd' :
                e.key === 'Control' ? 'Ctrl' :
                e.key === 'Shift' ? 'Shift' :
                e.key === 'Alt' ? 'Alt' : e.key

    if (['Cmd', 'Ctrl', 'Shift', 'Alt'].includes(key)) {
      activeKeys.add(key)
    } else if (key.length === 1 || key.startsWith('F') || ['Space', 'Tab', 'Escape'].includes(key)) {
      const mods = ['Ctrl', 'Cmd', 'Shift', 'Alt'].filter(m => activeKeys.has(m))
      const parts = [...mods, key]
      combo = parts.join(' + ')
      recording = false
    }
  }

  function onKeyup(e: KeyboardEvent) {
    if (!recording) return
    const key = e.key === 'Meta' ? 'Cmd' :
                e.key === 'Control' ? 'Ctrl' :
                e.key === 'Shift' ? 'Shift' :
                e.key === 'Alt' ? 'Alt' : null
    if (key) activeKeys.delete(key)
  }
</script>

<svelte:window onkeydown={onKeydown} onkeyup={onKeyup} />

<div style="max-width: 520px; margin: 0 auto; text-align: center;">
  <InkReveal direction="left" duration={900} delay={200}>
    <InkText kind="display" as="h1" style="margin-bottom: var(--lp-space-3);">
      <WordReveal text="Your Hotkey" stagger={50} delay={300} />
    </InkText>
  </InkReveal>

  <BlurReveal delay={500} distance={16}>
    <InkText kind="body" tone="ink-mute" style="max-width: 380px; margin: 0 auto var(--lp-space-8);">
      Summon Condura from anywhere with one key combination.
      Press your desired shortcut now.
    </InkText>
  </BlurReveal>

  <BlurReveal delay={700} distance={16}>
    <div style="display: flex; flex-direction: column; align-items: center; gap: var(--lp-space-6);">
      <!-- Hotkey display area -->
      <button
        type="button"
        class="lp-focus"
        onclick={startRecording}
        style="
          width: 280px;
          height: 80px;
          border-radius: var(--lp-radius-md);
          background: var(--lp-paper-warm);
          border: 1.5px solid {recording ? 'var(--lp-synapse)' : 'var(--lp-ink-ghost)'};
          display: flex;
          align-items: center;
          justify-content: center;
          font-family: var(--lp-font-display);
          font-size: var(--lp-text-headline);
          color: var(--lp-ink);
          cursor: pointer;
          transition: all var(--lp-dur-normal) var(--lp-ease-thread);
          box-shadow: {recording ? '0 0 0 3px var(--lp-synapse-glow)' : 'none'};
        "
      >
        {#if recording}
          <span style="
            font-family: var(--lp-font-mono);
            font-size: var(--lp-text-body);
            color: var(--lp-synapse);
            letter-spacing: 0.08em;
            text-transform: uppercase;
            animation: lp-pulse-thinking 1.8s ease-in-out infinite;
          ">press a key&hellip;</span>
        {:else if combo}
          {combo}
        {:else}
          <span style="
            font-family: var(--lp-font-mono);
            font-size: var(--lp-text-caption);
            color: var(--lp-ink-mute);
          ">Click to set</span>
        {/if}
      </button>

      <!-- Preset suggestions -->
      <div style="display: flex; gap: var(--lp-space-2); flex-wrap: wrap; justify-content: center;">
        {#each ['Ctrl + Space', 'Cmd + Shift + Space', 'Alt + Space'] as preset}
          <button
            type="button"
            class="lp-focus"
            onclick={() => { combo = preset }}
            style="
              padding: 4px 12px;
              border-radius: var(--lp-radius-pill);
              border: 1px solid var(--lp-ink-ghost);
              background: transparent;
              color: var(--lp-ink-mute);
              font-family: var(--lp-font-mono);
              font-size: var(--lp-text-micro);
              cursor: pointer;
              transition: all var(--lp-dur-fast) var(--lp-ease-thread);
            "
          >{preset}</button>
        {/each}
      </div>

      <!-- Continue / Skip -->
      <div style="display: flex; gap: var(--lp-space-3); margin-top: var(--lp-space-4);">
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
          "
        >Skip</button>
        <MagneticButton
          variant="primary"
          size="md"
          disabled={!combo}
          onclick={() => onnext(combo || 'Ctrl+Space')}
        >
          Continue
        </MagneticButton>
      </div>
    </div>
  </BlurReveal>
</div>
