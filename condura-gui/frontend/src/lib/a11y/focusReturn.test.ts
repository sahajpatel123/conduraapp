import { describe, it, expect, vi } from 'vitest'
import { createFocusReturn } from './focusReturn'

// createFocusReturn — used by Palette and Cheatsheet to restore focus
// to the element that opened them. Pins the contract: no-op when
// never captured, restores via queueMicrotask after the dialog is
// removed from the DOM.

describe('createFocusReturn', () => {
  it('restore() is a no-op when capture() was never called', () => {
    const fr = createFocusReturn()
    expect(() => fr.restore()).not.toThrow()
  })

  it('restore() focuses the captured element after a microtask', async () => {
    const el1 = document.createElement('button')
    const el2 = document.createElement('button')
    document.body.append(el1, el2)
    el1.focus()
    expect(document.activeElement).toBe(el1)
    const fr = createFocusReturn()
    fr.capture()
    // Move focus elsewhere — el2 takes focus.
    el2.focus()
    expect(document.activeElement).toBe(el2)
    fr.restore()
    // Restore uses queueMicrotask — wait one tick.
    await Promise.resolve()
    expect(document.activeElement).toBe(el1)
    el1.remove()
    el2.remove()
  })

  it('subsequent restore() calls are no-ops after the first', async () => {
    const el1 = document.createElement('button')
    const el2 = document.createElement('button')
    document.body.append(el1, el2)
    el1.focus()
    const fr = createFocusReturn()
    fr.capture()
    fr.restore()
    await Promise.resolve()
    // After restore, the trigger reference is cleared — second restore is a no-op.
    el2.focus()
    fr.restore()
    await Promise.resolve()
    expect(document.activeElement).toBe(el2)
    el1.remove()
    el2.remove()
  })
})