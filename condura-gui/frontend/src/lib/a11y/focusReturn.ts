// Focus return helper — capture an element on open, restore focus
// to it on close. Used by modals whose open/close is driven by a
// user action (palette, cheatsheet). Halt and Consent use a
// different shape (reactive \$effect over an external boolean) and
// keep their inline implementation.

export interface FocusReturn {
  /** Call right before opening the modal. Captures the current
   *  document.activeElement — the element to return focus to. */
  capture(): void
  /** Call when the modal closes. Restores focus to the captured
   *  element via queueMicrotask so the focus call lands AFTER
   *  Svelte has removed the dialog from the DOM (no risk of
   *  focusing a detached element). */
  restore(): void
}

export function createFocusReturn(): FocusReturn {
  let trigger: HTMLElement | null = null
  return {
    capture() {
      trigger = (document.activeElement as HTMLElement) ?? null
    },
    restore() {
      const el = trigger
      trigger = null
      if (!el) return
      queueMicrotask(() => el.focus({ preventScroll: true }))
    },
  }
}