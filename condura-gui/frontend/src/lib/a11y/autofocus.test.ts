import { describe, it, expect } from 'vitest'
import { focusOn } from './autofocus'

// focusOn — used by Sync, Chat, and Channels for "the user came here
// to do one thing" auto-focus on mount. Pins the contract: fires the
// focus inside a microtask so it runs after Svelte's pending DOM
// updates; respects the `when` guard; defaults to preventScroll + select.

describe('focusOn', () => {
  it('focuses the element when when() is true', async () => {
    const el = document.createElement('input')
    document.body.append(el)
    let called = 0
    const original = el.focus.bind(el)
    el.focus = () => {
      called++
      original()
    }
    focusOn(() => el, () => true)
    // focusOn uses queueMicrotask — await a microtask cycle.
    await Promise.resolve()
    expect(called).toBe(1)
    el.remove()
  })

  it('does nothing when when() is false', async () => {
    const el = document.createElement('input')
    document.body.append(el)
    let called = 0
    el.focus = () => {
      called++
    }
    focusOn(() => el, () => false)
    await Promise.resolve()
    expect(called).toBe(0)
    el.remove()
  })

  it('respects preventScroll: false', async () => {
    const el = document.createElement('input')
    document.body.append(el)
    let scrollOpts: { preventScroll?: boolean } | undefined
    el.focus = ((opts?: { preventScroll?: boolean }) => {
      scrollOpts = opts
    }) as never
    focusOn(() => el, () => true, { preventScroll: false })
    await Promise.resolve()
    expect(scrollOpts?.preventScroll).toBe(false)
    el.remove()
  })

  it('skips select() when select: false', async () => {
    const el = document.createElement('input')
    document.body.append(el)
    let selected = 0
    ;(el as HTMLInputElement).select = () => {
      selected++
    }
    focusOn(() => el, () => true, { select: false })
    await Promise.resolve()
    expect(selected).toBe(0)
    el.remove()
  })

  it('does not throw when ref returns null', async () => {
    expect(() => {
      focusOn(() => null, () => true)
    }).not.toThrow()
    await Promise.resolve()
  })
})