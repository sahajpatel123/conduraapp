<!--
  CursorHalo — a soft trailing halo around the OS cursor.

  Like the website's Cursor component, but adapted for the desktop app.
  The actual OS cursor is preserved (no double-dot, no lag). This adds
  ONE optional layer: a soft plum halo that lerps toward the pointer
  and only becomes visible when hovering interactive elements.

  Disabled on touch devices and when prefers-reduced-motion.

  Mounted ONCE at the app shell root, pointer-events: none.
-->
<script lang="ts">
  import { onMount } from 'svelte';

  let ringEl: HTMLDivElement | undefined = $state();
  let enabled = $state(false);

  onMount(() => {
    if (typeof window === 'undefined') return;

    enabled =
      !window.matchMedia('(prefers-reduced-motion: reduce)').matches &&
      window.matchMedia('(pointer: fine)').matches;

    if (!enabled) return;
    const ring = ringEl;
    if (!ring) return;

    let tx = window.innerWidth / 2;
    let ty = window.innerHeight / 2;
    let cx = tx;
    let cy = ty;
    let raf = 0;

    const onMove = (e: PointerEvent) => {
      tx = e.clientX;
      ty = e.clientY;
    };

    const onOver = (e: PointerEvent) => {
      const target = e.target as Element | null;
      if (!target) return;
      const interactive = target.closest(
        'a, button, [role="button"], input, textarea, select, [data-cursor="hover"]'
      );
      ring.classList.toggle('cursor-halo--hovering', !!interactive);
    };

    const onLeave = () => ring.classList.remove('cursor-halo--visible');
    const onEnter = () => ring.classList.add('cursor-halo--visible');

    const loop = () => {
      if (!ring) return;
      cx += (tx - cx) * 0.18;
      cy += (ty - cy) * 0.18;
      ring.style.transform = `translate(${cx}px, ${cy}px) translate(-50%, -50%)`;
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);

    window.addEventListener('pointermove', onMove, { passive: true });
    window.addEventListener('pointerover', onOver, { passive: true });
    document.addEventListener('pointerleave', onLeave);
    document.addEventListener('pointerenter', onEnter);
    // Show on first mount
    ring.classList.add('cursor-halo--visible');

    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener('pointermove', onMove);
      window.removeEventListener('pointerover', onOver);
      document.removeEventListener('pointerleave', onLeave);
      document.removeEventListener('pointerenter', onEnter);
    };
  });
</script>

{#if enabled}
  <div bind:this={ringEl} class="cursor-halo" aria-hidden="true">
    <div class="cursor-halo__ring"></div>
    <div class="cursor-halo__core"></div>
  </div>
{/if}

<style>
  .cursor-halo {
    position: fixed;
    top: 0;
    left: 0;
    width: 40px;
    height: 40px;
    pointer-events: none;
    z-index: 9999;
    opacity: 0;
    transition: opacity 220ms var(--ease-decelerate);
    will-change: transform;
  }

  .cursor-halo--visible {
    opacity: 1;
  }

  .cursor-halo__ring {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 1px solid var(--content-accent);
    opacity: 0.3;
    transition: transform 220ms var(--ease-decelerate), opacity 220ms var(--ease-decelerate);
  }

  .cursor-halo__core {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 4px;
    height: 4px;
    border-radius: 50%;
    background-color: var(--content-accent);
    transform: translate(-50%, -50%);
    opacity: 0.4;
    transition: transform 220ms var(--ease-decelerate), opacity 220ms var(--ease-decelerate);
  }

  /* On hover, the halo grows and the core intensifies */
  .cursor-halo--hovering .cursor-halo__ring {
    transform: scale(1.6);
    opacity: 0.6;
  }

  .cursor-halo--hovering .cursor-halo__core {
    transform: translate(-50%, -50%) scale(1.5);
    opacity: 0.7;
  }

  @media (prefers-reduced-motion: reduce) {
    .cursor-halo {
      display: none;
    }
  }
</style>