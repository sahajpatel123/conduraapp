<!--
  SynapseField — the living illustration of the desktop app's empty space.

  This is the desktop equivalent of the website's SynapseGarden. A
  hand-crafted SVG scene that lives behind the chat surface when there
  are no messages. Every element is alive:

    1. A curving horizon line that draws itself on mount
    2. A small tree on the right crest — branches sway in the wind
    3. A breathing sun upper-left
    4. 12 drifting pollen motes (canvas, sin-curve fade)
    5. A light thread that follows the cursor
    6. The agent's Pulse at the center (the presence)

  Total file is intentionally large — this is the centerpiece. Treat it
  as a real composition, not a tech demo.

  Disabled when prefers-reduced-motion.
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import Pulse from './Pulse.svelte';

  let canvasEl: HTMLCanvasElement | undefined = $state();
  let pathEl: SVGPathElement | undefined = $state();
  let horizonEl: SVGPathElement | undefined = $state();
  let treeGroup: SVGGElement | undefined = $state();
  let sunGroup: SVGGElement | undefined = $state();
  let fieldEl: SVGSVGElement | undefined = $state();

  // 12 deterministic pollen motes (no Math.random during render)
  function seededMotes(count: number) {
    const rand = (seed: number) => {
      const x = Math.sin(seed * 12.9898) * 43758.5453;
      return x - Math.floor(x);
    };
    return Array.from({ length: count }, (_, i) => ({
      id: i,
      delay: rand(i * 4 + 1) * 6,
      dur: 11 + rand(i * 4 + 2) * 8,
      size: 2 + rand(i * 4 + 3) * 2.5,
      dx: (rand(i * 4 + 4) - 0.5) * 70,
      dy: -100 - rand(i * 4 + 5) * 100,
      left: 4 + rand(i * 4 + 6) * 92,
      top: 55 + rand(i * 4 + 7) * 35,
    }));
  }

  onMount(() => {
    if (typeof window === 'undefined') return;
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

    const canvas = canvasEl;
    const path = pathEl;
    const horizon = horizonEl;
    const tree = treeGroup;
    const sun = sunGroup;
    const field = fieldEl;
    if (!canvas) return;
    const ctx = canvas.getContext('2d', { alpha: true });
    if (!ctx) return;

    // Pointer state
    const pointer = { x: 0.5, y: 0.5, active: false };
    const onPointer = (e: PointerEvent) => {
      pointer.x = e.clientX / window.innerWidth;
      pointer.y = e.clientY / window.innerHeight;
      pointer.active = true;
    };
    window.addEventListener('pointermove', onPointer, { passive: true });

    let raf = 0;
    let w = 0;
    let h = 0;
    const dpr = Math.min(window.devicePixelRatio || 1, 2);

    interface Mote {
      x: number;
      y: number;
      r: number;
      vx: number;
      vy: number;
      life: number;
      max: number;
    }
    let motes: Mote[] = [];

    const spawn = (): Mote => {
      const max = 800 + Math.random() * 600;
      return {
        x: Math.random() * w,
        y: Math.random() * h,
        r: 0.8 + Math.random() * 1.6,
        vx: (Math.random() - 0.5) * 0.05,
        vy: -0.04 - Math.random() * 0.08,
        life: Math.random() * max,
        max,
      };
    };

    const resize = () => {
      w = canvas.clientWidth;
      h = canvas.clientHeight;
      canvas.width = Math.floor(w * dpr);
      canvas.height = Math.floor(h * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      const target = Math.min(12, Math.floor((w * h) / 60000));
      motes = Array.from({ length: target }, () => spawn());
    };

    // ── Pollen tick ───────────────────────────────────────
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
        const alpha = fade * 0.55;
        ctx.beginPath();
        ctx.arc(m.x, m.y, m.r, 0, Math.PI * 2);
        // Plum-tinted motes
        ctx.fillStyle = `rgba(110, 58, 255, ${alpha})`;
        ctx.fill();
      }
      raf = requestAnimationFrame(tick);
    };

    // ── Thread sway + cursor nudge ───────────────────────────
    let threadRaf = 0;
    let t = 0;
    const sway = () => {
      t += 0.008;
      if (path) {
        const W = window.innerWidth || 1200;
        const H = window.innerHeight || 800;
        const px = pointer.x;
        const py = pointer.y;
        const amp = 80 + py * 90;
        const midY = H * (0.65 + py * 0.1);
        const cx1 = W * 0.2 + Math.sin(t) * 50 + (px - 0.5) * 100;
        const cy1 = midY - amp + Math.cos(t * 0.8) * 25;
        const cx2 = W * 0.7 + Math.cos(t * 0.9) * 50 - (px - 0.5) * 100;
        const cy2 = midY + amp + Math.sin(t * 0.7) * 25;
        const d = `M -40 ${midY} C ${cx1} ${cy1}, ${cx2} ${cy2}, ${W + 40} ${midY}`;
        path.setAttribute('d', d);
      }
      threadRaf = requestAnimationFrame(sway);
    };

    // ── Tree sway + sun breath ─────────────────────────────
    let natureRaf = 0;
    const nature = () => {
      const tt = performance.now() / 1000;
      if (tree) {
        const r = Math.sin(tt * 0.7) * 0.5;
        tree.style.transform = `rotate(${r}deg)`;
      }
      if (sun) {
        const s = 1 + Math.sin(tt * 0.5) * 0.04;
        sun.style.transform = `scale(${s})`;
      }
      natureRaf = requestAnimationFrame(nature);
    };

    // ── Horizon draw-in on mount ───────────────────────────
    if (horizon) {
      const len = horizon.getTotalLength();
      horizon.style.strokeDasharray = `${len}`;
      horizon.style.strokeDashoffset = `${len}`;
      horizon.getBoundingClientRect();
      horizon.style.transition = 'stroke-dashoffset 2.4s cubic-bezier(0.22, 1, 0.36, 1)';
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          if (horizon) horizon.style.strokeDashoffset = '0';
        });
      });
    }

    resize();
    window.addEventListener('resize', resize);
    raf = requestAnimationFrame(tick);
    threadRaf = requestAnimationFrame(sway);
    natureRaf = requestAnimationFrame(nature);

    return () => {
      cancelAnimationFrame(raf);
      cancelAnimationFrame(threadRaf);
      cancelAnimationFrame(natureRaf);
      window.removeEventListener('resize', resize);
      window.removeEventListener('pointermove', onPointer);
    };
  });

  const motes = seededMotes(12);
</script>

<div class="field" aria-hidden="true">
  <!-- Pollen canvas -->
  <canvas bind:this={canvasEl} class="field__canvas"></canvas>

  <!-- The hand-drawn SVG scene -->
  <svg
    bind:this={fieldEl}
    class="field__svg"
    viewBox="0 0 100 100"
    preserveAspectRatio="xMidYMid slice"
  >
    <!-- A warm sky wash -->
    <defs>
      <linearGradient id="sky-wash" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="var(--paper-warm-0)" />
        <stop offset="60%" stop-color="var(--paper-warm-50)" />
        <stop offset="100%" stop-color="var(--paper-warm-100)" />
      </linearGradient>
      <radialGradient id="sun-bloom" cx="0.5" cy="0.5" r="0.5">
        <stop offset="0%" stop-color="rgba(240, 192, 130, 0.4)" />
        <stop offset="100%" stop-color="rgba(240, 192, 130, 0)" />
      </radialGradient>
    </defs>

    <rect width="100" height="100" fill="url(#sky-wash)" />

    <!-- The breathing sun upper-left -->
    <g bind:this={sunGroup} style="transform-origin: 18% 22%;">
      <circle cx="18" cy="22" r="6" fill="rgba(240, 192, 130, 0.5)" />
      <circle cx="18" cy="22" r="9" fill="none" stroke="rgba(240, 192, 130, 0.4)" stroke-width="0.2" />
      <circle cx="18" cy="22" r="13" fill="none" stroke="rgba(240, 192, 130, 0.3)" stroke-width="0.15" />
      <circle cx="18" cy="22" r="18" fill="url(#sun-bloom)" />
    </g>

    <!-- Distant mountains — a single curved ridge -->
    <path
      d="M 0 60 L 18 50 L 28 56 L 42 46 L 56 54 L 70 48 L 84 56 L 100 50 L 100 100 L 0 100 Z"
      fill="rgba(184, 200, 192, 0.4)"
    />

    <!-- A gentle horizon line that draws itself on mount -->
    <path
      bind:this={horizonEl}
      d="M 0 72 C 22 66, 38 70, 50 68 C 62 66, 78 72, 100 70"
      fill="none"
      stroke="var(--content-accent)"
      stroke-width="0.12"
      stroke-linecap="round"
      opacity="0.4"
    />

    <!-- A second, deeper horizon — for visual depth -->
    <path
      d="M 0 80 C 18 76, 32 80, 48 78 C 64 76, 80 82, 100 80 L 100 100 L 0 100 Z"
      fill="var(--paper-warm-100)"
      opacity="0.6"
    />

    <!-- A small tree on the right crest — sways in the wind -->
    <g bind:this={treeGroup} style="transform-origin: 78% 80%;">
      <!-- Trunk -->
      <path
        d="M 78 80 C 78.4 72, 78 64, 79 60"
        stroke="rgba(58, 42, 24, 0.6)"
        stroke-width="0.4"
        fill="none"
        stroke-linecap="round"
      />
      <!-- Canopy — hand-drawn Bézier blobs -->
      <g fill="rgba(46, 106, 62, 0.7)">
        <path d="M 79 60 C 75 54, 77 48, 82 47 C 86 46, 89 50, 88 55 C 87 58, 83 60, 79 60 Z" />
        <path d="M 79 58 C 82 52, 88 51, 91 55 C 93 59, 90 62, 85 62 C 82 62, 79 60, 79 58 Z" />
        <path d="M 78 58 C 76 54, 78 50, 82 50 C 84 50, 84 54, 82 57 C 81 59, 79 59, 78 58 Z" />
      </g>
      <!-- Canopy highlight — gives it dimension -->
      <path
        d="M 80 52 C 82 50, 85 50, 86 52"
        stroke="rgba(94, 154, 106, 0.6)"
        stroke-width="0.3"
        fill="none"
      />
    </g>

    <!-- A few small distant trees -->
    <g fill="rgba(58, 106, 72, 0.5)">
      <line x1="22" y1="80" x2="22" y2="76" stroke="rgba(58, 42, 24, 0.4)" stroke-width="0.2" />
      <circle cx="22" cy="75" r="0.8" />
      <line x1="40" y1="82" x2="40" y2="79" stroke="rgba(58, 42, 24, 0.3)" stroke-width="0.15" />
      <circle cx="40" cy="78" r="0.6" />
      <line x1="64" y1="80" x2="64" y2="77" stroke="rgba(58, 42, 24, 0.3)" stroke-width="0.15" />
      <circle cx="64" cy="76" r="0.5" />
    </g>

    <!-- The light thread — sways continuously + follows cursor -->
    <path
      bind:this={pathEl}
      class="field__thread"
      d="M -5 65 C 25 50, 75 80, 105 65"
      stroke-width="0.4"
    />

    <!-- Foreground grass tufts — hand-drawn detail -->
    <g stroke="rgba(58, 106, 72, 0.5)" stroke-width="0.12" opacity="0.7" stroke-linecap="round">
      {#each Array(20) as _, i}
        <path d={`M ${2 + i * 5} 95 q 0.5 -${1 + (i % 3) * 0.6} 1 -${1.5 + (i % 3) * 0.6}`} fill="none" />
      {/each}
    </g>
  </svg>

  <!-- The agent's presence at the center — the Pulse, large and breathing -->
  <div class="field__presence">
    <Pulse state="idle" size="xl" label="Synaptic" />
  </div>

  <!-- A hand-set serif line, the agent's first word -->
  <p class="field__whisper">I'm here.</p>

  <!-- A subtle paper grain over the scene so it never looks flat -->
  <div class="field__grain" aria-hidden="true"></div>
</div>

<style>
  .field {
    position: relative;
    width: 100%;
    height: 100%;
    min-height: 480px;
    overflow: hidden;
    pointer-events: none;
  }

  .field__canvas {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }

  .field__svg {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }

  /* The thread — a soft line, plum-tinted, sways continuously */
  .field__thread {
    fill: none;
    stroke: var(--content-accent);
    opacity: 0.22;
    vector-effect: non-scaling-stroke;
    animation: thread-breath 9s ease-in-out infinite;
  }

  @keyframes thread-breath {
    0%, 100% { opacity: 0.16; }
    50% { opacity: 0.32; }
  }

  /* The agent's presence — centered, slightly above center */
  .field__presence {
    position: absolute;
    top: 38%;
    left: 50%;
    transform: translate(-50%, -50%);
    opacity: 0.9;
    z-index: 2;
  }

  /* The whisper — italic serif, fades in with the scene */
  .field__whisper {
    position: absolute;
    top: 52%;
    left: 50%;
    transform: translateX(-50%);
    font-family: var(--font-serif);
    font-style: italic;
    font-size: var(--text-h3-size);
    color: var(--content-secondary);
    margin: 0;
    z-index: 2;
    animation: whisper-in 1.2s var(--ease-decelerate) 1.2s both;
    text-align: center;
  }

  @keyframes whisper-in {
    from { opacity: 0; transform: translate(-50%, 6px); }
    to { opacity: 1; transform: translate(-50%, 0); }
  }

  /* A subtle paper grain over the whole scene — adds depth without color */
  .field__grain {
    position: absolute;
    inset: 0;
    background-image:
      radial-gradient(circle at 20% 30%, rgba(14, 16, 20, 0.015) 0%, transparent 60%),
      radial-gradient(circle at 80% 70%, rgba(14, 16, 20, 0.015) 0%, transparent 60%);
    pointer-events: none;
  }

  @media (prefers-reduced-motion: reduce) {
    .field__thread {
      animation: none;
    }
  }
</style>