<script lang="ts">
  import './living-paper.css'

  export type PulsePhase = 'idle' | 'thinking' | 'acting' | 'listening' | 'consent' | 'error' | 'ok'

  interface Props {
    phase?: PulsePhase
    size?: number
    label?: string
    class?: string
  }

  let {
    phase = 'idle',
    size = 8,
    label,
    class: className = '',
  }: Props = $props()

  const phaseConfig: Record<PulsePhase, { color: string; duration: string; pulseColor: string }> = {
    idle:     { color: 'var(--lp-synapse)',     duration: '5s',    pulseColor: 'var(--lp-synapse-glow)' },
    thinking: { color: 'var(--lp-synapse-glow)', duration: '1.8s', pulseColor: 'var(--lp-synapse-light)' },
    acting:   { color: 'var(--lp-pollen)',       duration: '1s',   pulseColor: 'var(--lp-pollen-glow)' },
    listening:{ color: 'var(--lp-pollen-glow)',  duration: '1.2s', pulseColor: 'var(--lp-pollen-light)' },
    consent:  { color: 'var(--lp-sky)',          duration: '1s',   pulseColor: 'var(--lp-sky-deep)' },
    error:    { color: 'var(--lp-danger)',        duration: '0.8s', pulseColor: '#D45A4A' },
    ok:       { color: 'var(--lp-ok)',           duration: '4s',   pulseColor: 'var(--lp-synapse-glow)' },
  }

  const cfg = $derived(phaseConfig[phase])
</script>

<span
  class="lp {className}"
  role="status"
  aria-label={label || phase}
  style="
    display: inline-flex;
    align-items: center;
    gap: 6px;
    vertical-align: middle;
  "
>
  <span
    class="lp-pulse-dot"
    style="
      width: {size}px;
      height: {size}px;
      border-radius: 50%;
      background: {cfg.color};
      animation: lp-pulse-{phase} {cfg.duration} var(--lp-ease-in-out) infinite;
      box-shadow: 0 0 6px {cfg.pulseColor};
    "
  ></span>
  {#if label}
    <span class="lp-pulse-label" style="
      font-family: var(--lp-font-mono);
      font-size: var(--lp-text-micro);
      letter-spacing: var(--lp-tracking-mono);
      text-transform: uppercase;
      color: var(--lp-ink-mute);
    ">{label}</span>
  {/if}
</span>

<style>
  @keyframes lp-pulse-idle {
    0%, 100% { transform: scale(1); opacity: 0.85; }
    50% { transform: scale(1.04); opacity: 1; }
  }
  @keyframes lp-pulse-thinking {
    0%, 100% { transform: scale(1); opacity: 0.8; }
    50% { transform: scale(1.1); opacity: 1; }
  }
  @keyframes lp-pulse-acting {
    0%, 100% { transform: scale(1); opacity: 1; }
    30% { transform: scale(1.15); opacity: 0.9; }
    60% { transform: scale(0.95); opacity: 0.7; }
  }
  @keyframes lp-pulse-listening {
    0%, 100% { transform: scale(1); opacity: 0.7; }
    50% { transform: scale(1.2); opacity: 1; }
  }
  @keyframes lp-pulse-consent {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.25) scaleX(0.7); }
  }
  @keyframes lp-pulse-error {
    0%, 100% { transform: scale(1); opacity: 1; }
    50% { transform: scale(1.1); opacity: 0.6; }
  }
  @keyframes lp-pulse-ok {
    0%, 100% { transform: scale(1); opacity: 0.85; }
    50% { transform: scale(1.03); opacity: 1; }
  }
</style>
