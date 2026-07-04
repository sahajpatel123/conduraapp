<script lang="ts">
  /**
   * PollenSpark — An amber spark/glow accent element.
   * 
   * A small pollen-colored dot with a subtle breathing animation.
   * Used as a decorative accent near CTAs, active states, and
   * living indicators.
   */
  import './living-paper.css'

  interface Props {
    size?: number
    /** Animation phase */
    phase?: 'idle' | 'pulse' | 'float'
    class?: string
    style?: string
  }

  let {
    size = 4,
    phase = 'idle',
    class: className = '',
    style = '',
  }: Props = $props()

  const phaseDurations: Record<string, string> = {
    idle: '3s',
    pulse: '1.5s',
    float: '4s',
  }
</script>

<span
  class="lp lp-pollen-spark lp-pollen-spark-{phase} {className}"
  style="
    display: inline-block;
    width: {size}px;
    height: {size}px;
    border-radius: 50%;
    background: var(--lp-pollen);
    box-shadow: 0 0 {size * 2}px var(--lp-pollen-glow);
    animation-duration: {phaseDurations[phase]};
    {style}
  "
></span>

<style>
  .lp-pollen-spark-idle {
    animation: lp-pollen-idle var(--lp-dur-slow) var(--lp-ease-in-out) infinite;
  }
  .lp-pollen-spark-pulse {
    animation: lp-pollen-pulse 1.5s var(--lp-ease-in-out) infinite;
  }
  .lp-pollen-spark-float {
    animation: lp-pollen-float 4s var(--lp-ease-in-out) infinite;
  }

  @keyframes lp-pollen-idle {
    0%, 100% { opacity: 0.7; transform: scale(1); }
    50% { opacity: 1; transform: scale(1.08); }
  }

  @keyframes lp-pollen-pulse {
    0%, 100% { opacity: 0.6; transform: scale(1); }
    30% { opacity: 1; transform: scale(1.25); box-shadow: 0 0 12px var(--lp-pollen-glow); }
    60% { opacity: 0.8; transform: scale(0.9); }
  }

  @keyframes lp-pollen-float {
    0%, 100% { transform: translateY(0) rotate(0deg); opacity: 0.3; }
    25% { opacity: 0.8; }
    50% { transform: translateY(-8px) rotate(180deg); opacity: 0.5; }
    75% { opacity: 0.8; }
  }
</style>
