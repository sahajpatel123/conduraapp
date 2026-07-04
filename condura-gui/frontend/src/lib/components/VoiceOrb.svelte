<!-- VoiceOrb.svelte
     Animated orb that shows voice state: idle/off, listening,
     speaking, muted. Hand-rolled CSS, tokens only, named presets
     where they fit. Reduced motion: static orb with status text. -->

<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { ipc } from '../ipc/client'
  import { t } from '../i18n'

  type Status = 'off' | 'listening' | 'speaking' | 'muted'
  type Size = 'sm' | 'md' | 'lg'

  interface Props {
    /** Voice status. If omitted, the orb subscribes to IPC events
        and computes its own status from voice/tray activity. */
    status?: Status
    /** Visual size of the orb (60/96/160 px). */
    size?: Size
  }

  let { status, size = 'md' }: Props = $props()

  let internal: Status = $state('off')
  // When the parent passes `status`, prefer it. Otherwise fall
  // back to the IPC-driven `internal` value.
  const effective: Status = $derived(status ?? internal)
  let cleanups: Array<() => void> = []
  let idleTimer: ReturnType<typeof setTimeout> | null = null

  function clearIdleTimer(): void {
    if (idleTimer !== null) {
      clearTimeout(idleTimer)
      idleTimer = null
    }
  }

  // Map internal voice IPC state to the Status vocabulary.
  function applyVoiceEvent(name: 'partial' | 'final' | 'tray', payload?: { status?: string }): void {
    if (name === 'tray' && payload?.status) {
      if (payload.status === 'listening') internal = 'listening'
      else if (payload.status === 'speaking') internal = 'speaking'
      else if (payload.status === 'idle') internal = 'off'
      return
    }
    if (name === 'partial') {
      internal = 'listening'
      return
    }
    if (name === 'final') {
      internal = 'speaking'
      clearIdleTimer()
      idleTimer = setTimeout(() => {
        internal = 'off'
        idleTimer = null
      }, 5000)
    }
  }

  onMount(() => {
    if (status !== undefined) return // parent-driven, skip subscriptions
    try {
      cleanups.push(
        ipc.on('voice.partial' as never, (() => applyVoiceEvent('partial')) as never),
        ipc.on('voice.final' as never, (() => applyVoiceEvent('final')) as never),
        ipc.on('tray.status' as never, ((data: { status?: string }) => applyVoiceEvent('tray', data)) as never)
      )
    } catch {
      // Not running inside Wails (unit tests / static preview).
    }
  })

  onDestroy(() => {
    clearIdleTimer()
    cleanups.forEach((c) => c())
    cleanups = []
  })

  const label = $derived.by(() => {
    switch (effective) {
      case 'listening': return t('voice.orb.listening')
      case 'speaking':  return t('voice.orb.speaking')
      case 'muted':     return t('voice.orb.muted')
      case 'off':
      default:          return t('voice.orb.off')
    }
  })

  // Size token → pixel diameter for the outermost ring.
  const diameter = $derived(size === 'sm' ? 60 : size === 'lg' ? 160 : 96)
</script>

<div
  class="voice-orb orb-size-{size}"
  class:listening={effective === 'listening'}
  class:speaking={effective === 'speaking'}
  class:muted={effective === 'muted'}
  class:off={effective === 'off'}
  style="--orb-diameter: {diameter}px"
  role="status"
  aria-live="polite"
>
  <div class="orb-core anim-glow-pulse"></div>
  <div class="orb-ring ring-1"></div>
  <div class="orb-ring ring-2"></div>
  <div class="orb-ring ring-3"></div>
  <div class="orb-ring ring-4"></div>
  <div class="orb-ring ring-5"></div>
  {#if effective !== 'off'}
    <span class="orb-label">{label}</span>
  {/if}
</div>

<style>
  .voice-orb {
    position: relative;
    width: var(--orb-diameter, 96px);
    height: var(--orb-diameter, 96px);
    display: flex;
    align-items: center;
    justify-content: center;
    isolation: isolate;
  }

  .orb-core {
    position: relative;
    width: calc(var(--orb-diameter, 96px) * 0.4);
    height: calc(var(--orb-diameter, 96px) * 0.4);
    border-radius: 50%;
    background: var(--text-faint);
    box-shadow: 0 0 12px rgba(0, 0, 0, 0.35);
    transition: background-color var(--transition-base) ease,
                box-shadow var(--transition-base) ease,
                transform var(--transition-base) var(--ease-spring);
    z-index: 2;
  }

  /* Local "anim-rings" — the named preset lives in animations.css
     for global use; duplicated here so this component is
     self-contained when imported in isolation (tests, previews). */
  @keyframes orb-rings-kf {
    0%   { transform: scale(0.6); opacity: 0; }
    20%  { opacity: 0.9; }
    100% { transform: scale(1.6); opacity: 0; }
  }

  .orb-ring {
    position: absolute;
    border-radius: 50%;
    border: 1px solid transparent;
    width: calc(var(--orb-diameter, 96px) * 0.6);
    height: calc(var(--orb-diameter, 96px) * 0.6);
    opacity: 0;
    pointer-events: none;
    z-index: 1;
  }

  .ring-1 { animation: orb-rings-kf 2.4s var(--ease-out-quart) infinite;            animation-delay: 0.0s; }
  .ring-2 { animation: orb-rings-kf 2.4s var(--ease-out-quart) infinite;            animation-delay: 0.3s; }
  .ring-3 { animation: orb-rings-kf 2.4s var(--ease-out-quart) infinite;            animation-delay: 0.6s; }
  .ring-4 { animation: orb-rings-kf 2.4s var(--ease-out-quart) infinite;            animation-delay: 0.9s; }
  .ring-5 { animation: orb-rings-kf 2.4s var(--ease-out-quart) infinite;            animation-delay: 1.2s; }

  .orb-label {
    position: absolute;
    bottom: -28px;
    left: 50%;
    transform: translateX(-50%);
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
    color: var(--text-faint);
    white-space: nowrap;
    transition: color var(--transition-base) ease;
  }

  /* ── Listening: accent core, accent rings ─────────────── */
  .voice-orb.listening .orb-core {
    background: var(--accent);
    box-shadow: 0 0 28px var(--accent-glow);
  }
  .voice-orb.listening .orb-ring {
    border-color: var(--accent-soft);
  }
  .voice-orb.listening .orb-label { color: var(--accent); }

  /* ── Speaking: stronger glow, slightly faster rings ────── */
  .voice-orb.speaking .orb-core {
    background: var(--accent-hover);
    box-shadow: 0 0 36px var(--accent-glow), 0 0 60px var(--accent-faint);
    transform: scale(1.04);
  }
  .voice-orb.speaking .orb-ring {
    border-color: var(--accent-soft);
  }
  .ring-1, .ring-2, .ring-3, .ring-4, .ring-5 { animation-duration: 1.8s; }
  .voice-orb.speaking .orb-label { color: var(--accent-hover); }

  /* ── Muted: warn-toned core, no rings ─────────────────── */
  .voice-orb.muted .orb-core {
    background: var(--warn);
    box-shadow: 0 0 16px var(--warn-glow);
  }
  .voice-orb.muted .orb-ring { display: none; }
  .voice-orb.muted .orb-label { color: var(--warn); }

  /* ── Off: dim core, no rings, no glow pulse ───────────── */
  .voice-orb.off .orb-core {
    background: var(--text-faint);
    box-shadow: none;
    animation: none;
  }
  .voice-orb.off .orb-ring { display: none; }
  .voice-orb.off .orb-label { color: var(--text-faint); }

  /* Reduced motion: static orb + status text only. */
  @media (prefers-reduced-motion: reduce) {
    .orb-core,
    .orb-ring { animation: none !important; }
    .voice-orb.listening .orb-core,
    .voice-orb.speaking .orb-core { transform: scale(1.03); }
  }
</style>
