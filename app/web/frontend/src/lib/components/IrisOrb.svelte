<script lang="ts">
  // The Iris — Condura's heartbeat. One organism shown in two
  // places (dock + context bar). Its motion encodes agent state:
  //   idle      → slow breath, no glow
  //   listening → thin iris ring pulses outward
  //   thinking  → conic core rotates + iris glow
  //   acting    → coral core + coral ring per action
  //   consent   → core "holds its breath": steady coral ring + warn
  //   offline   → dim, still
  type OrbState = 'idle' | 'listening' | 'thinking' | 'acting' | 'consent' | 'offline'
  interface Props {
    state?: OrbState
    size?: number
    title?: string
  }
  let { state = 'idle', size = 28, title }: Props = $props()
</script>

<span
  class="iris iris-{state}"
  style="--orb: {size}px"
  role="img"
  aria-label={title ?? `Agent ${state}`}
  title={title ?? state}
>
  <span class="core"></span>
  <span class="shimmer"></span>
  <span class="ring"></span>
</span>

<style>
  .iris {
    position: relative;
    display: inline-block;
    width: var(--orb);
    height: var(--orb);
    flex-shrink: 0;
    border-radius: 50%;
  }

  /* shell — dark glassy bead with a hairline ring */
  .iris::before {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: 50%;
    background: radial-gradient(circle at 32% 28%, var(--surface-3), var(--surface-1) 80%);
    border: 1px solid var(--border-strong);
    box-shadow: var(--inset-hair);
  }

  /* core — the living light */
  .core {
    position: absolute;
    inset: 14%;
    border-radius: 50%;
    background: radial-gradient(circle at 38% 34%, var(--aurora-2) 0%, var(--aurora-0) 60%, transparent 100%);
    opacity: 0.85;
    animation: iris-breath 4.2s var(--ease-glide) infinite;
  }

  /* shimmer — conic sweep, only visible while thinking */
  .shimmer {
    position: absolute;
    inset: 10%;
    border-radius: 50%;
    opacity: 0;
    background: conic-gradient(from 0deg, transparent, var(--aurora-1), var(--aurora-3), transparent 70%);
    mix-blend-mode: screen;
  }

  /* ring — expanding pulse for listening/acting/consent */
  .ring {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 1.5px solid transparent;
  }

  /* ── idle ── */
  .iris-idle .core { opacity: 0.78; }

  /* ── listening ── */
  .iris-listening .ring {
    border-color: var(--accent);
    animation: iris-ring 1.8s var(--ease-soft) infinite;
  }

  /* ── thinking ── */
  .iris-thinking .core { animation-duration: 1.6s; }
  .iris-thinking .shimmer { opacity: 0.9; animation: iris-spin 1.6s linear infinite; }
  .iris-thinking::after {
    content: '';
    position: absolute;
    inset: -2px;
    border-radius: 50%;
    box-shadow: var(--glow-iris);
  }

  /* ── acting ── */
  .iris-acting .core {
    background: radial-gradient(circle at 38% 34%, var(--accent-2-hover) 0%, var(--accent-2-press) 70%, transparent 100%);
  }
  .iris-acting .ring {
    border-color: var(--accent-2);
    animation: iris-ring 1.2s var(--ease-soft) infinite;
  }
  .iris-acting::after {
    content: '';
    position: absolute;
    inset: -2px;
    border-radius: 50%;
    box-shadow: var(--glow-coral);
  }

  /* ── consent / Gatekeeper hold — orb holds its breath ── */
  .iris-consent .core {
    animation: none;
    background: radial-gradient(circle at 38% 34%, var(--accent-2) 0%, var(--accent-2-press) 75%, transparent 100%);
  }
  .iris-consent .ring { border-color: var(--accent-2); }

  /* ── offline ── */
  .iris-offline .core { animation: none; opacity: 0.32; filter: grayscale(0.5); }

  @media (prefers-reduced-motion: reduce) {
    .core, .shimmer, .ring { animation: none !important; }
  }
</style>
