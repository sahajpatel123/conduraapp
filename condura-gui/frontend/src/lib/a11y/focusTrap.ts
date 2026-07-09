/** Shared focus-trap helpers for modal overlays (mirrors Dialog.svelte). */

const FOCUSABLE =
  'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'

export function trapTab(root: HTMLElement | null, e: KeyboardEvent): void {
  if (e.key !== 'Tab' || !root) return
  const focusable = root.querySelectorAll<HTMLElement>(FOCUSABLE)
  if (focusable.length === 0) return
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  const active = document.activeElement as HTMLElement
  if (e.shiftKey && active === first) {
    e.preventDefault()
    last.focus()
  } else if (!e.shiftKey && active === last) {
    e.preventDefault()
    first.focus()
  }
}

export function focusFirst(root: HTMLElement | null): void {
  if (!root) return
  const focusable = root.querySelectorAll<HTMLElement>(FOCUSABLE)
  const target = focusable[0] ?? root
  queueMicrotask(() => target.focus())
}

/** Arrow/Home/End navigation for a list of focusable row buttons. */
export function navigateListRows(
  e: KeyboardEvent,
  rows: NodeListOf<HTMLElement> | HTMLElement[],
  focusedIndex: number,
): number {
  const list = Array.from(rows)
  if (list.length === 0) return focusedIndex
  let next = focusedIndex
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    next = focusedIndex < list.length - 1 ? focusedIndex + 1 : 0
    list[next]?.focus()
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    next = focusedIndex > 0 ? focusedIndex - 1 : list.length - 1
    list[next]?.focus()
  } else if (e.key === 'Home') {
    e.preventDefault()
    next = 0
    list[0]?.focus()
  } else if (e.key === 'End') {
    e.preventDefault()
    next = list.length - 1
    list[next]?.focus()
  } else if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault()
    const btn = list[focusedIndex >= 0 ? focusedIndex : 0]
    btn?.click()
  }
  return next
}
