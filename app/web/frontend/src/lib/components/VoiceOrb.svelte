<!-- VoiceOrb.svelte
     Animated orb that shows voice state: idle (dim), listening
     (pulsing blue), speaking (glowing green). Hand-rolled CSS,
     no Tailwind. Reduced motion: static orb with status text. -->

<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { ipc } from '../ipc/client'
  import { t } from '../i18n'

  type OrbState = 'idle' | 'listening' | 'speaking'

  let state = $state<OrbState>('idle')
  let cleanups: Array<() => void> = []
  let idleTimer: ReturnType<typeof setTimeout> | null = null

  function clearIdleTimer() {
    if (idleTimer !== null) {
      clearTimeout(idleTimer)
      idleTimer = null
    }
  }

  onMount(() => {
    cleanups.push(
      ipc.on('voice.partial' as never, (() => {
        state = 'listening'
      }) as never),

      ipc.on('voice.final' as never, (() => {
        state = 'speaking'
        clearIdleTimer()
        idleTimer = setTimeout(() => {
          state = 'idle'
          idleTimer = null
        }, 5000)
      }) as never),

      ipc.on('tray.status' as never, ((data: { status?: string }) => {
        if (data.status === 'listening') state = 'listening'
        else if (data.status === 'speaking') state = 'speaking'
        else if (data.status === 'idle') state = 'idle'
      }) as never)
    )
  })

  onDestroy(() => {
    clearIdleTimer()
    cleanups.forEach(c => c())
    cleanups = []
  })
</script>

<div class="voice-orb" class:listening={state === 'listening'} class:speaking={state === 'speaking'}>
  <div class="orb-core"></div>
  <div class="orb-ring ring-1"></div>
  <div class="orb-ring ring-2"></div>
  <div class="orb-ring ring-3"></div>
  {#if state !== 'idle'}
    <span class="orb-label">{state === 'listening' ? $t('voice.orb.listening') : $t('voice.orb.speaking')}</span>
  {/if}
</div>

<style>
  .voice-orb {
    position: relative;
    width: 120px;
    height: 120px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .orb-core {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: var(--color-text-faint);
    transition: background 0.3s ease, box-shadow 0.3s ease, transform 0.3s ease;
  }

  .orb-ring {
    position: absolute;
    border-radius: 50%;
    border: 1px solid transparent;
    transition: border-color 0.3s ease, opacity 0.3s ease;
    opacity: 0;
  }

  .ring-1 {
    width: 72px;
    height: 72px;
  }

  .ring-2 {
    width: 96px;
    height: 96px;
  }

  .ring-3 {
    width: 120px;
    height: 120px;
  }

  .orb-label {
    position: absolute;
    bottom: -24px;
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--color-text-faint);
    transition: color 0.3s ease;
  }

  /* Listening state: pulsing blue */
  .voice-orb.listening .orb-core {
    background: #3b82f6;
    box-shadow: 0 0 24px rgba(59, 130, 246, 0.5);
    animation: orbPulse 2s ease-in-out infinite;
  }

  .voice-orb.listening .orb-ring {
    border-color: rgba(59, 130, 246, 0.3);
    opacity: 1;
    animation: ringPulse 2s ease-in-out infinite;
  }

  .voice-orb.listening .ring-1 { animation-delay: 0s; }
  .voice-orb.listening .ring-2 { animation-delay: 0.3s; }
  .voice-orb.listening .ring-3 { animation-delay: 0.6s; }

  .voice-orb.listening .orb-label {
    color: #3b82f6;
  }

  /* Speaking state: glowing green */
  .voice-orb.speaking .orb-core {
    background: #22c55e;
    box-shadow: 0 0 32px rgba(34, 197, 94, 0.6);
    animation: orbGlow 1.5s ease-in-out infinite;
  }

  .voice-orb.speaking .orb-ring {
    border-color: rgba(34, 197, 94, 0.3);
    opacity: 1;
    animation: ringGlow 1.5s ease-in-out infinite;
  }

  .voice-orb.speaking .ring-1 { animation-delay: 0s; }
  .voice-orb.speaking .ring-2 { animation-delay: 0.2s; }
  .voice-orb.speaking .ring-3 { animation-delay: 0.4s; }

  .voice-orb.speaking .orb-label {
    color: #22c55e;
  }

  @keyframes orbPulse {
    0%, 100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(1.1);
      opacity: 0.8;
    }
  }

  @keyframes ringPulse {
    0%, 100% {
      transform: scale(1);
      opacity: 0.3;
    }
    50% {
      transform: scale(1.15);
      opacity: 0.6;
    }
  }

  @keyframes orbGlow {
    0%, 100% {
      transform: scale(1);
      box-shadow: 0 0 32px rgba(34, 197, 94, 0.6);
    }
    50% {
      transform: scale(1.05);
      box-shadow: 0 0 48px rgba(34, 197, 94, 0.8);
    }
  }

  @keyframes ringGlow {
    0%, 100% {
      transform: scale(1);
      opacity: 0.3;
    }
    50% {
      transform: scale(1.1);
      opacity: 0.5;
    }
  }

  /* Reduced motion: static orb */
  @media (prefers-reduced-motion: reduce) {
    .orb-core,
    .orb-ring {
      animation: none !important;
    }

    .voice-orb.listening .orb-core {
      transform: scale(1.05);
    }

    .voice-orb.speaking .orb-core {
      transform: scale(1.05);
    }
  }
</style>
