<script lang="ts">
  /**
   * QuillCursor — wires the custom pixel cursor in condura.css.
   *
   * The ink quill + pollen hover ring live in condura.css as SVG data-URIs.
   * This component toggles body[data-hover] so interactive surfaces swap to
   * the pollen target cursor. No trailing DOM dot — Wails/WebKit often fails
   * to deliver continuous pointermove, which made a follow-along halo stick
   * until the next click.
   */
  import { onMount } from 'svelte'
  import './living-paper.css'

  const INTERACTIVE =
    'button:not(.no-tactile), .tactile, [role="button"]:not(.no-tactile), [role="link"]:not(.no-tactile), summary:not(.no-tactile), .choice, .nav-item, .dock-item, .thread-link, .lp-nav-row, .lp-nav-halt, input, textarea, select, a, [data-hoverable], [data-cursor="hover"]'

  onMount(() => {
    if (!window.matchMedia('(pointer: fine)').matches) return

    let lastX = -1
    let lastY = -1

    const setHover = (active: boolean) => {
      document.body.dataset.hover = active ? '1' : '0'
    }

    const updateHoverAt = (x: number, y: number) => {
      if (x === lastX && y === lastY) return
      lastX = x
      lastY = y
      const el = document.elementFromPoint(x, y) as HTMLElement | null
      if (!el) {
        setHover(false)
        return
      }
      const disabled = !!el.closest?.('[disabled], [aria-disabled="true"]')
      if (disabled) {
        setHover(false)
        return
      }
      // Text fields keep the native I-beam — pollen ring is for click targets.
      const isTextField = !!el.closest('input, textarea, [contenteditable="true"]')
      if (isTextField) {
        setHover(false)
        return
      }
      setHover(!!el.closest(INTERACTIVE))
    }

    const onMove = (e: MouseEvent | PointerEvent) => {
      updateHoverAt(e.clientX, e.clientY)
    }

    const onLeave = () => {
      lastX = -1
      lastY = -1
      setHover(false)
    }

    const onWindowLeave = (e: MouseEvent) => {
      if (e.relatedTarget === null) onLeave()
    }

    // mousemove is required for Wails/macOS WebView — pointermove alone often
    // only fires on press/release, which is exactly the "stuck dot" bug.
    const opts: AddEventListenerOptions = { passive: true, capture: true }
    document.addEventListener('mousemove', onMove, opts)
    document.addEventListener('pointermove', onMove, opts)
    document.addEventListener('mouseout', onWindowLeave, opts)

    setHover(false)

    return () => {
      document.removeEventListener('mousemove', onMove, opts)
      document.removeEventListener('pointermove', onMove, opts)
      document.removeEventListener('mouseout', onWindowLeave, opts)
      delete document.body.dataset.hover
    }
  })
</script>
