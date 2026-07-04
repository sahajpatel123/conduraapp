<script lang="ts">
  import { onMount } from 'svelte';

  // Condura Cursor — a trailing halo that lags behind the OS pointer and
  // brightens over interactive elements. The real cursor is the pixel
  // quill set in condura.css; this is the decorative trailing dot.
  let halo: HTMLDivElement;

  onMount(() => {
    const reduce = matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (reduce) return;

    let hx = innerWidth / 2;
    let hy = innerHeight / 2;
    let tx = hx;
    let ty = hy;
    let raf = 0;
    // Pause the rAF loop when the tab is hidden or the halo is off-screen —
    // the loop runs at 60fps, so a 30-minute backgrounded tab is ~108k
    // pointless style writes / wasted CPU.
    let running = !document.hidden;

    const move = (e: PointerEvent) => {
      tx = e.clientX;
      ty = e.clientY;
      halo.classList.add('on');
    };
    const over = (e: PointerEvent) => {
      const t = e.target as HTMLElement | null;
      const hov = !!t?.closest?.(
        'button,.choice,.nav-item,.dock-item,.thread-link,input,textarea,a,[data-hoverable]'
      );
      document.body.dataset.hover = hov ? '1' : '0';
      halo.classList.toggle('hover', hov);
    };
    const tick = () => {
      hx += (tx - hx) * 0.16;
      hy += (ty - hy) * 0.16;
      halo.style.transform = `translate(${hx - 7}px, ${hy - 7}px)`;
      raf = running ? requestAnimationFrame(tick) : 0;
    };

    const onVisibility = () => {
      running = !document.hidden;
      if (running && raf === 0) raf = requestAnimationFrame(tick);
    };
    document.addEventListener('visibilitychange', onVisibility);

    // IntersectionObserver pauses the loop when the halo leaves the viewport
    // (e.g. another route is open, the shell is hidden behind an overlay).
    let inView = true;
    const io = new IntersectionObserver(
      (entries) => {
        inView = !!entries[0]?.isIntersecting;
        running = !document.hidden && inView;
        if (running && raf === 0) raf = requestAnimationFrame(tick);
      },
      { threshold: 0 }
    );
    io.observe(halo);

    addEventListener('pointermove', move);
    addEventListener('pointerover', over);
    if (running) tick();

    return () => {
      removeEventListener('pointermove', move);
      removeEventListener('pointerover', over);
      document.removeEventListener('visibilitychange', onVisibility);
      io.disconnect();
      cancelAnimationFrame(raf);
    };
  });
</script>

<div class="cursor-halo" bind:this={halo} aria-hidden="true"></div>

<style>
  /* :global so Svelte doesn't strip the .on/.hover states toggled via classList */
  :global(.cursor-halo) {
    position: fixed;
    top: 0;
    left: 0;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    border: 1px solid var(--synapse-glow);
    background: color-mix(in oklab, var(--pollen) 10%, transparent);
    pointer-events: none;
    z-index: var(--z-max);
    opacity: 0;
    transform: translate(-100px, -100px);
    transition:
      width 0.25s var(--ease),
      height 0.25s var(--ease),
      opacity 0.25s var(--ease),
      background 0.25s var(--ease),
      border-color 0.25s var(--ease);
  }
  :global(.cursor-halo.on) {
    opacity: 0.55;
  }
  :global(.cursor-halo.hover) {
    width: 34px;
    height: 34px;
    background: color-mix(in oklab, var(--pollen) 14%, transparent);
    border-color: var(--pollen);
    opacity: 0.7;
  }
</style>