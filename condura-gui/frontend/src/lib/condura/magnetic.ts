// Condura magnetic action — pointer-distance spring pull on primary CTAs.
// Pull falls off to zero at `radius` and reaches `strength` at the center.
// Honors prefers-reduced-motion (no-op). Used via `use:magnetic`.
export interface MagneticOpts {
  strength?: number;
  radius?: number;
  enabled?: boolean;
}

export function magnetic(node: HTMLElement, opts: MagneticOpts = {}) {
  const reduce =
    typeof matchMedia !== 'undefined' &&
    matchMedia('(prefers-reduced-motion: reduce)').matches;

  let strength = opts.enabled === false ? 0 : (opts.strength ?? 0.32);
  let radius = opts.radius ?? 130;
  let tx = 0;
  let ty = 0;
  let cx = 0;
  let cy = 0;
  let raf = 0;

  function onMove(e: PointerEvent) {
    const r = node.getBoundingClientRect();
    const mx = e.clientX - (r.left + r.width / 2);
    const my = e.clientY - (r.top + r.height / 2);
    const dist = Math.hypot(mx, my);
    if (dist > radius || strength === 0) {
      tx = 0;
      ty = 0;
    } else {
      const pull = (1 - dist / radius) * strength;
      tx = mx * pull;
      ty = my * pull;
    }
  }

  function onLeave() {
    tx = 0;
    ty = 0;
  }

  function tick() {
    cx += (tx - cx) * 0.2;
    cy += (ty - cy) * 0.2;
    node.style.transform = `translate(${cx.toFixed(2)}px, ${cy.toFixed(2)}px)`;
    raf = requestAnimationFrame(tick);
  }

  if (!reduce) {
    node.addEventListener('pointermove', onMove);
    node.addEventListener('pointerleave', onLeave);
    tick();
  }

  return {
    update(o: MagneticOpts) {
      strength = o.enabled === false ? 0 : o.strength ?? 0.32;
      radius = o.radius ?? 130;
    },
    destroy() {
      node.removeEventListener('pointermove', onMove);
      node.removeEventListener('pointerleave', onLeave);
      cancelAnimationFrame(raf);
    },
  };
}