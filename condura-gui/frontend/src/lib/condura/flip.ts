/**
 * FLIP — First, Last, Invert, Play.
 *
 * Helper for animating a single persistent element between two positions
 * without a layout-shifting keyframe. Required by SCREEN_NAVRAIL §5.1
 * for the synapse-Thread segment that moves between active rows.
 *
 * Usage:
 *   const { capture, apply, play, cancel } = createFLIP(el, 320);
 *   capture();           // snapshot OLD rect + sync layout
 *   apply(newTop, newHeight);
 *   play();              // next frame, transition enabled, back to identity
 *
 * Reduced-motion: a single static `apply()` (no transition) is the
 * behaviour under `prefers-reduced-motion: reduce`. Callers can pass
 * the matchMedia result via `respectMotion(false)` to disable.
 */

export interface FLIPController {
  /** Snapshot the element's current rect as the OLD reference. */
  capture(): void;
  /** Apply new geometry synchronously (no transition). Returns NEW rect. */
  apply(top: number, height: number): { top: number; height: number };
  /** Play the FLIP — re-enables transition then clears transform. */
  play(): void;
  /** Cancel any pending transition + clear transforms. */
  cancel(): void;
  /** Read the last captured OLD rect (read-only). */
  readonly oldRect: { top: number; height: number } | null;
}

export interface FLIPOptions {
  durationMs?: number;
  easing?: string;
  respectReducedMotion?: boolean;
}

export function createFLIP(
  el: HTMLElement | null,
  durationMs = 320,
  options: FLIPOptions = {},
): FLIPController {
  const { easing = 'var(--ease)', respectReducedMotion = true } = options;
  let oldRect: { top: number; height: number } | null = null;
  let frameRequested = false;

  const reducedMotion =
    respectReducedMotion &&
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  function clearTransition(): void {
    if (!el) return;
    el.style.transition = 'none';
  }

  function setTransition(): void {
    if (!el) return;
    el.style.transition = `transform ${durationMs}ms ${easing}`;
  }

  return {
    get oldRect() {
      return oldRect;
    },
    capture(): void {
      if (!el) return;
      oldRect = el.getBoundingClientRect();
    },
    apply(top: number, height: number): { top: number; height: number } {
      if (!el) return { top, height };
      clearTransition();
      el.style.top = `${top}px`;
      el.style.height = `${height}px`;
      // Measure synchronously after layout. getBoundingClientRect forces
      // a layout flush.
      const next = el.getBoundingClientRect();
      if (oldRect && !reducedMotion) {
        const dy = oldRect.top - next.top;
        const sy = oldRect.height / Math.max(1, next.height);
        el.style.transformOrigin = 'top center';
        el.style.transform = `translateY(${dy}px) scaleY(${sy})`;
      } else {
        el.style.transform = '';
      }
      return { top: next.top, height: next.height };
    },
    play(): void {
      if (!el || reducedMotion) {
        // Under reduced motion: snap to identity immediately.
        if (el) el.style.transform = '';
        return;
      }
      if (frameRequested) return;
      frameRequested = true;
      requestAnimationFrame(() => {
        frameRequested = false;
        if (!el) return;
        setTransition();
        el.style.transform = '';
      });
    },
    cancel(): void {
      if (!el) return;
      el.style.transition = '';
      el.style.transform = '';
      oldRect = null;
    },
  };
}
