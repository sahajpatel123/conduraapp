<!--
  CanvasBackground — the living paper behind every surface.

  Translation of the website's BrandSurface to the desktop app. Three layers:

   1. A slow drift of soft "pollen" motes on canvas, GPU-cheap (≤28 particles)
   2. A global synapse thread — an SVG path that sways continuously AND
      nudges toward the cursor when present
   3. Subtle paper grain (CSS, defined in tokens)

  The intent is *quiet* ambience, not a particle storm. A breathing paper
  that earns its keep. Respects prefers-reduced-motion: returns the static
  base color only.

  Mounted ONCE at the app shell root, behind all content. Sits in a
  pointer-events: none layer so it never blocks clicks.

  Props:
    enabled — whether the canvas is animated (default: true; false = static fallback)
-->
<script lang="ts">
  import { onMount } from 'svelte';

  interface Props {
    enabled?: boolean;
  }

  let { enabled = true }: Props = $props();

  let canvasEl: HTMLCanvasElement | undefined = $state();
  let pathEl: SVGPathElement | undefined = $state();

  interface Mote {
    x: number;
    y: number;
    r: number;
    vx: number;
    vy: number;
    life: number;
    max: number;
    hue: number;
  }

  onMount(() => {
    if (!enabled) return;
    if (typeof window === 'undefined') return;

    const prefersReduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (prefersReduced) return;

    const canvas = canvasEl;
    const path = pathEl;
    if (!canvas) return;
    const ctx = canvas.getContext('2d', { alpha: true });
    if (!ctx) return;

    let raf = 0;
    let w = 0;
    let h = 0;
    const dpr = Math.min(window.devicePixelRatio || 1, 2);

    // Pointer state — thread path responds to this
    const pointer = { x: 0.5, y: 0.5, active: false };
    const onPointer = (e: PointerEvent) => {
      pointer.x = e.clientX / window.innerWidth;
      pointer.y = e.clientY / window.innerHeight;
      pointer.active = true;
    };
    window.addEventListener('pointermove', onPointer, { passive: true });

    let motes: Mote[] = [];

    const spawn = (): Mote => {
      const max = 600 + Math.random() * 800;
      return {
        x: Math.random() * w,
        y: Math.random() * h,
        r: 0.6 + Math.random() * 1.6,
        vx: (Math.random() - 0.5) * 0.06,
        vy: -0.04 - Math.random() * 0.09,
        life: Math.random() * max,
        max,
        // Plam-ish or plum. Plum is the brand accent.
        hue: Math.random() < 0.7 ? 268 : 38,
      };
    };

    const resize = () => {
      w = canvas.clientWidth;
      h = canvas.clientHeight;
      canvas.width = Math.floor(w * dpr);
      canvas.height = Math.floor(h * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      const target = Math.min(28, Math.floor((w * h) / 48000));
      motes = Array.from({ length: target }, () => spawn());
    };

    // ── Pollen tick — drift upward, fade with sin curve, respawn ──
    const tick = () => {
      ctx.clearRect(0, 0, w, h);
      for (let i = 0; i < motes.length; i++) {
        const m = motes[i];
        m.life += 1;
        m.x += m.vx;
        m.y += m.vy;
        if (m.life > m.max || m.y < -20 || m.x < -20 || m.x > w + 20) {
          motes[i] = spawn();
          motes[i].y = h + 10;
          continue;
        }
        const t = m.life / m.max;
        const fade = Math.sin(Math.PI * t);
        const alpha = fade * 0.5;
        ctx.beginPath();
        ctx.arc(m.x, m.y, m.r, 0, Math.PI * 2);
        if (m.hue === 268) {
          // Plum — the brand accent
          ctx.fillStyle = `rgba(110, 58, 255, ${alpha})`;
        } else {
          // Warm pollen
          ctx.fillStyle = `rgba(201, 123, 46, ${alpha})`;
        }
        ctx.fill();
      }
      raf = requestAnimationFrame(tick);
    };

    // ── Thread sway — sine waves + pointer nudge ──────────────
    let threadRaf = 0;
    let t = 0;
    const sway = () => {
      t += 0.008;
      if (path) {
        const W = window.innerWidth || 1200;
        const H = window.innerHeight || 800;
        const px = pointer.x;
        const py = pointer.y;
        const amp = 60 + py * 80;
        const midY = H * (0.35 + py * 0.12);
        const cx1 = W * 0.2 + Math.sin(t) * 40 + (px - 0.5) * 80;
        const cy1 = midY - amp + Math.cos(t * 0.8) * 20;
        const cx2 = W * 0.7 + Math.cos(t * 0.9) * 40 - (px - 0.5) * 80;
        const cy2 = midY + amp + Math.sin(t * 0.7) * 20;
        const d = `M -40 ${midY} C ${cx1} ${cy1}, ${cx2} ${cy2}, ${W + 40} ${midY}`;
        path.setAttribute('d', d);
      }
      threadRaf = requestAnimationFrame(sway);
    };

    resize();
    window.addEventListener('resize', resize);
    raf = requestAnimationFrame(tick);
    threadRaf = requestAnimationFrame(sway);

    return () => {
      cancelAnimationFrame(raf);
      cancelAnimationFrame(threadRaf);
      window.removeEventListener('resize', resize);
      window.removeEventListener('pointermove', onPointer);
    };
  });
</script>

<div class="bg" aria-hidden="true">
  <!-- Pollen canvas — drifts motes upward -->
  <canvas
    bind:this={canvasEl}
    class="bg__canvas"
  ></canvas>

  <!-- Thread SVG — sways continuously + nudges toward pointer -->
  <svg
    class="bg__thread-svg"
    preserveAspectRatio="none"
    viewBox="0 0 100 100"
  >
    <path
      bind:this={pathEl}
      class="bg__thread"
      d="M -5 50 C 25 30, 75 70, 105 50"
      pathLength="1"
    />
  </svg>
</div>

<style>
  .bg {
    position: fixed;
    inset: 0;
    z-index: 0;
    pointer-events: none;
    overflow: hidden;
  }

  .bg__canvas {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }

  .bg__thread-svg {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    opacity: 0.14;
  }

  /* The thread — a thin curved line, plum accent */
  .bg__thread {
    fill: none;
    stroke: var(--content-accent);
    stroke-width: 1;
    vector-effect: non-scaling-stroke;
    /* A subtle pulse — slightly brightens as the agent breathes */
    animation: thread-breath 8s ease-in-out infinite;
  }

  @keyframes thread-breath {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }

  @media (prefers-reduced-motion: reduce) {
    .bg {
      /* Static only — no canvas, no thread sway */
    }
  }
</style>