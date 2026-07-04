<!--
  Magnetic — a wrapper that pulls its child gently toward the pointer.

  Like the website's MagneticButton but adapted for the desktop app
  shell. Works with any child element (button, link, etc.). The child
  receives the magnetic motion; the wrapper just provides the tracking.

  Respects reduced motion: no pull, just the child's natural styles.
-->
<script lang="ts">
  import { onMount } from 'svelte';

  interface Props {
    radius?: number;
    strength?: number;
    children?: import('svelte').Snippet;
  }

  let { radius = 80, strength = 0.25, children }: Props = $props();

  let wrapEl: HTMLDivElement | undefined = $state();
  let offsetX = $state(0);
  let offsetY = $state(0);

  onMount(() => {
    if (typeof window === 'undefined') return;
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

    const wrap = wrapEl;
    if (!wrap) return;

    function onMove(e: PointerEvent) {
      if (!wrap) return;
      const r = wrap.getBoundingClientRect();
      const cx = r.left + r.width / 2;
      const cy = r.top + r.height / 2;
      const dx = e.clientX - cx;
      const dy = e.clientY - cy;
      const dist = Math.hypot(dx, dy);
      if (dist > radius) {
        offsetX = 0;
        offsetY = 0;
        return;
      }
      // Pull is strongest at the center, fades to zero at the edge
      const pull = (1 - dist / radius) * strength;
      offsetX = dx * pull;
      offsetY = dy * pull;
    }

    function onLeave() {
      offsetX = 0;
      offsetY = 0;
    }

    window.addEventListener('pointermove', onMove, { passive: true });
    wrap.addEventListener('pointerleave', onLeave);

    return () => {
      window.removeEventListener('pointermove', onMove);
      wrap.removeEventListener('pointerleave', onLeave);
    };
  });
</script>

<div
  bind:this={wrapEl}
  class="magnetic"
  style="--mag-x: {offsetX}px; --mag-y: {offsetY}px;"
>
  {@render children?.()}
</div>

<style>
  .magnetic {
    display: inline-flex;
    transform: translate(var(--mag-x, 0), var(--mag-y, 0));
    /* Spring-like settle via cubic-bezier */
    transition: transform 280ms cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  @media (prefers-reduced-motion: reduce) {
    .magnetic {
      transform: none;
      transition: none;
    }
  }
</style>