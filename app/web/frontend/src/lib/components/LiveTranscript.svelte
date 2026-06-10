<!-- LiveTranscript.svelte
     Displays the live transcript during voice sessions.
     Subscribes to voice.partial and voice.final SSE events
     via the IPC client's event system. -->

<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { ipc } from '../ipc/client'

  let transcript = $state<string>('')
  let isRecording = $state<boolean>(false)
  let isFinal = $state<boolean>(false)
  let cleanups: Array<() => void> = []

  onMount(() => {
    cleanups.push(
      ipc.on('voice.partial' as never, ((data: { recording?: boolean; samples?: number }) => {
        isRecording = data.recording ?? false
        if (isRecording && !isFinal) {
          // Show recording indicator; actual transcript comes from voice.final
          transcript = 'Listening...'
        }
      }) as never),

      ipc.on('voice.final' as never, ((data: { text?: string; confidence?: number }) => {
        isRecording = false
        isFinal = true
        transcript = data.text ?? ''
        // Auto-hide after 5 seconds
        setTimeout(() => {
          transcript = ''
          isFinal = false
        }, 5000)
      }) as never)
    )
  })

  onDestroy(() => {
    cleanups.forEach(c => c())
    cleanups = []
  })
</script>

{#if transcript || isRecording}
  <div class="live-transcript" class:recording={isRecording} class:final={isFinal}>
    <div class="transcript-text">
      {#if isRecording && !isFinal}
        <span class="pulse"></span>
      {/if}
      {transcript}
    </div>
  </div>
{/if}

<style>
  .live-transcript {
    position: fixed;
    bottom: 80px;
    left: 50%;
    transform: translateX(-50%);
    max-width: 600px;
    width: 90%;
    padding: 16px 24px;
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    z-index: 1000;
    transition: opacity var(--transition-base), transform var(--transition-base);
    animation: fadeIn 0.2s ease-out;
  }

  .live-transcript.recording {
    border-color: var(--color-accent);
    box-shadow: var(--shadow-lg), 0 0 16px rgba(var(--color-accent-rgb), 0.2);
  }

  .live-transcript.final {
    border-color: var(--color-success);
    box-shadow: var(--shadow-lg), 0 0 16px rgba(74, 222, 128, 0.2);
  }

  .transcript-text {
    font-family: var(--font-sans);
    font-size: 16px;
    line-height: 1.5;
    color: var(--color-text);
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pulse {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-accent);
    animation: pulse 1.5s ease-in-out infinite;
    flex-shrink: 0;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateX(-50%) translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateX(-50%) translateY(0);
    }
  }

  @keyframes pulse {
    0%, 100% {
      opacity: 1;
      transform: scale(1);
    }
    50% {
      opacity: 0.5;
      transform: scale(1.2);
    }
  }
</style>
