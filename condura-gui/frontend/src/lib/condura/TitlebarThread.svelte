<script lang="ts">
  import { onMount } from 'svelte';
  import { conversation } from '../stores/conversation.svelte';
  import { halt } from '../stores/halt.svelte';
  import { consent } from '../stores/consent.svelte';

  // THE SIGNATURE (MOAT §3, DIRECTION §5). A 1.25px synapse hairline that
  //   1. draws itself in left→right on first paint (clip-path reveal),
  //   2. bends toward the pointer via a low-pass filter (0.06 lerp — locked
  //      in SCREEN_TITLEBAR §8.4.6; 0.12 while consent is pending),
  //   3. hardens to --danger and freezes its node when the agent is halted,
  //   4. grows a streaming progress hairline while tokens arrive.
  //
  // The titlebar is never dead. But the rAF loop is paused on
  // visibilitychange + IntersectionObserver (background tab / scrolled away)
  // and never starts under prefers-reduced-motion — so it burns 0 CPU when
  // there is nothing to perceive. Budget: 4 setAttribute calls/frame
  // (line.d, glow.d, node.cx, node.cy) while active & un-halted; 0 otherwise.
  let line: SVGPathElement;
  let glow: SVGPathElement;
  let node: SVGCircleElement;

  let reduce = $state(false);
  let drawn = $state(false);

  // Live agent state → drives node size, stroke color, and the progress
  // hairline. Read reactively for the template; the rAF loop reads the
  // stores directly (a plain read returns the current value).
  let streaming = $derived(conversation.isStreaming);
  let halted = $derived(halt.state.halted);

  // Node radius follows phase: halted freezes small, streaming amplifies the
  // footprint (SHELL §2.1), reduced-motion hides it entirely (§3.7).
  let nodeR = $derived(reduce ? 1.5 : halted ? 2 : streaming ? 6 : 3);
  let nodeOpacity = $derived(reduce ? 0 : 1);
  let strokeColor = $derived(halted ? 'var(--danger)' : 'var(--synapse)');

  onMount(() => {
    reduce = matchMedia('(prefers-reduced-motion: reduce)').matches;

    // First-paint self-draw. Under reduced-motion the global stylesheet zeroes
    // the transition, so this simply snaps to fully drawn.
    requestAnimationFrame(() => (drawn = true));

    if (reduce) return; // static thread, no bend loop, no node.

    let px = 0.5;
    let py = 0.5;
    let cx = 0.5;
    let cy = 0.5;
    let raf = 0;
    let running = !document.hidden;

    const move = (e: PointerEvent) => {
      px = e.clientX / innerWidth;
      py = e.clientY / innerHeight;
    };
    addEventListener('pointermove', move);

    const bend = () => {
      // Halted: the organism stopped. Freeze the node, keep the loop alive so
      // it resumes cleanly when the halt clears (§3.5 — "not paused").
      if (!halt.state.halted) {
        // Consent pending pulls harder (§3.4): 0.06 → 0.12 lerp.
        const lerp = consent.ticket ? 0.12 : 0.06;
        cx += (px - cx) * lerp;
        cy += (py - cy) * lerp;
        const W = innerWidth;
        const H = 44;
        const mid = 22;
        const x = cx * W;
        const y = Math.max(6, Math.min(H - 6, cy * H));
        const d = `M 0 ${mid} C ${W * 0.25} ${mid - (mid - y) * 0.7}, ${W * 0.42} ${y}, ${W * 0.5} ${y} S ${W * 0.78} ${mid + (y - mid) * 0.4}, ${W} ${mid}`;
        line.setAttribute('d', d);
        glow.setAttribute('d', d);
        node.setAttribute('cx', String(x));
        node.setAttribute('cy', String(y));
      }
      raf = running ? requestAnimationFrame(bend) : 0;
    };

    const onVisibility = () => {
      running = !document.hidden;
      if (running && raf === 0) raf = requestAnimationFrame(bend);
    };
    document.addEventListener('visibilitychange', onVisibility);

    // Pause the loop when the titlebar is scrolled out of view.
    const host = line.closest('.titlebar') ?? line.parentElement;
    let inView = true;
    const io = new IntersectionObserver(
      (entries) => {
        inView = !!entries[0]?.isIntersecting;
        running = !document.hidden && inView;
        if (running && raf === 0) raf = requestAnimationFrame(bend);
      },
      { threshold: 0 }
    );
    if (host) io.observe(host);

    if (running) bend();

    return () => {
      removeEventListener('pointermove', move);
      document.removeEventListener('visibilitychange', onVisibility);
      io.disconnect();
      cancelAnimationFrame(raf);
    };
  });
</script>

<svg
  class="titlebar-thread"
  class:drawn
  style="--tb-stroke:{strokeColor}"
  preserveAspectRatio="none"
  aria-hidden="true"
>
  <path
    bind:this={glow}
    d="M 0 22 L 9999 22"
    fill="none"
    stroke="var(--synapse-glow)"
    stroke-width="3"
    opacity="0.18"
    filter="blur(3px)"
  />
  {#if streaming}
    <!-- Progress hairline (§4.2): the only continuous loop in the titlebar,
         bounded to the streaming state. Reduced-motion collapses it to a
         static fill (no dasharray, full opacity). -->
    <path
      class="progress-hairline"
      d="M 0 22 L 9999 22"
      fill="none"
      stroke="var(--synapse-glow)"
      stroke-width="1"
      opacity={reduce ? 0.55 : 0.4}
      stroke-dasharray={reduce ? 'none' : '12 8'}
    />
  {/if}
  <path
    bind:this={line}
    d="M 0 22 L 9999 22"
    fill="none"
    stroke="var(--tb-stroke)"
    stroke-width="var(--thread-w)"
    stroke-linecap="round"
    stroke-dasharray="6 8"
  />
  <circle
    bind:this={node}
    cx="0"
    cy="22"
    r={nodeR}
    opacity={nodeOpacity}
    fill="var(--pollen)"
    filter="drop-shadow(0 0 6px color-mix(in oklab, var(--pollen) 70%, transparent))"
  />
</svg>

<style>
  .titlebar-thread {
    position: absolute;
    left: 160px;
    right: 200px;
    top: 0;
    height: 100%;
    width: auto;
    overflow: visible;
    pointer-events: none;
    opacity: 0.55;
    /* First-paint self-draw: revealed left→right. */
    clip-path: inset(0 100% 0 0);
  }
  .titlebar-thread.drawn {
    clip-path: inset(0 0 0 0);
    transition: clip-path var(--dur-slow) var(--ease);
  }
  .titlebar-thread circle,
  .titlebar-thread .progress-hairline {
    transition:
      r var(--dur) var(--ease),
      opacity var(--dur) var(--ease);
  }
  .progress-hairline {
    animation: thread-travel 1.6s linear infinite;
  }
  /* Draw the dashes left→right by walking the dashoffset one full period. */
  @keyframes thread-travel {
    from { stroke-dashoffset: 20; }
    to { stroke-dashoffset: 0; }
  }
</style>
