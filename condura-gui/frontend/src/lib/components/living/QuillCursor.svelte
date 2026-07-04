<script lang="ts">
  /**
   * QuillCursor — Decorative trailing halo inspired by the website.
   * 
   * A synapse-green ring with pollen-amber core that follows the
   * pointer with a gentle spring lag, enlarging over interactive
   * elements. Only visible on devices with fine pointers.
   */
  import './living-paper.css'

  interface Props {
    /** How strongly the ring follows (0-1, lower = more lag) */
    lerpFactor?: number
    /** Size of the ring in px at rest */
    ringSize?: number
    /** Size when hovering interactive elements */
    hoverSize?: number
    class?: string
  }

  let {
    lerpFactor = 0.18,
    ringSize = 16,
    hoverSize = 34,
    class: className = '',
  }: Props = $props()

  let ringEl = $state<HTMLDivElement | null>(null)
  let coreEl = $state<HTMLDivElement | null>(null)
  let x = $state(-100)
  let y = $state(-100)
  let targetX = $state(-100)
  let targetY = $state(-100)
  let isHovering = $state(false)
  let isVisible = $state(false)
  let rafId = $state(0)

  function isFinePointer(): boolean {
    return window.matchMedia('(pointer: fine)').matches
  }

  function prefersReducedMotion(): boolean {
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches
  }

  function onPointerMove(e: PointerEvent) {
    targetX = e.clientX
    targetY = e.clientY
    if (!isVisible) {
      x = targetX
      y = targetY
      isVisible = true
    }
  }

  function onPointerLeave() {
    isVisible = false
  }

  function updateHover() {
    // Check if hovering over an interactive element using :hover
    // This runs in the rAF loop — check via elementFromPoint
    if (!ringEl) return
    const el = document.elementFromPoint(targetX, targetY)
    if (!el) { isHovering = false; return }
    const tag = (el as HTMLElement).tagName?.toLowerCase()
    const role = (el as HTMLElement).getAttribute?.('role') || ''
    const isInteractive =
      tag === 'a' || tag === 'button' || tag === 'input' || tag === 'textarea' || tag === 'select' ||
      role === 'button' || role === 'menuitem' || role === 'radio' || role === 'switch' ||
      role === 'option' || role === 'tab' ||
      (el as HTMLElement).dataset?.cursor === 'hover' ||
      el.closest?.('[data-cursor="hover"]') !== null
    isHovering = isInteractive
  }

  function tick() {
    if (!ringEl || !coreEl) { rafId = requestAnimationFrame(tick); return }
    x += (targetX - x) * lerpFactor
    y += (targetY - y) * lerpFactor
    updateHover()
    const size = isHovering ? hoverSize : ringSize
    const half = size / 2
    ringEl.style.transform = `translate(${x - half}px, ${y - half}px)`
    ringEl.style.width = `${size}px`
    ringEl.style.height = `${size}px`
    ringEl.style.opacity = isHovering ? '1' : isVisible ? '0.6' : '0'
    coreEl.style.transform = `translate(${x - 3}px, ${y - 3}px)`
    coreEl.style.opacity = isVisible ? '1' : '0'
    rafId = requestAnimationFrame(tick)
  }

  $effect(() => {
    if (!isFinePointer() || prefersReducedMotion()) return
    window.addEventListener('pointermove', onPointerMove)
    window.addEventListener('pointerleave', onPointerLeave)
    rafId = requestAnimationFrame(tick)
    return () => {
      window.removeEventListener('pointermove', onPointerMove)
      window.removeEventListener('pointerleave', onPointerLeave)
      cancelAnimationFrame(rafId)
    }
  })
</script>

<svelte:window onpointermove={onPointerMove} onpointerleave={onPointerLeave} />

<!-- Ring — synapse green -->
<div
  bind:this={ringEl}
  class="lp {className}"
  style="
    position: fixed;
    top: 0; left: 0;
    border-radius: 50%;
    border: 1.5px solid var(--lp-synapse);
    pointer-events: none;
    z-index: 99999;
    transition: opacity 0.15s ease, width 0.2s var(--lp-ease-thread), height 0.2s var(--lp-ease-thread);
    will-change: transform, width, height;
  "
></div>

<!-- Core — pollen amber -->
<div
  bind:this={coreEl}
  style="
    position: fixed;
    top: 0; left: 0;
    width: 6px; height: 6px;
    border-radius: 50%;
    background: var(--lp-pollen);
    box-shadow: 0 0 8px var(--lp-pollen-glow);
    pointer-events: none;
    z-index: 99999;
    transition: opacity 0.15s ease;
    will-change: transform;
  "
></div>
