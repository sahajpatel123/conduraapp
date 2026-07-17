import { describe, it, expect } from 'vitest'
import { focusOn } from './autofocus'

// focusOn — used by Sync (PIN input), Chat (composer), Channels
// (Telegram token input). Pins the contract: focuses the element
// after a microtask (so Svelte's pending DOM updates complete
// first), respects the `when` guard, defaults to preventScroll,
// .select()s when the select option is true.

describe('focusOn', () => {
  it('focuses the element when when() is true', async () => {
    const el = document.createElement('input')
    document.body.append(el)
    let called = 0
    el.focus = () => {
      called++
    }
    focusOn(() => el, () => true)
    await Promise.resolve()
    expect(called).toBe(1)
    el.remove()
  })

  it('does not focus when when() is false', async () => {
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

  it('does nothing when ref returns null', () => {
    expect(() => focusOn(() => null, () => true)).not.toThrow()
  })

  it('uses preventScroll: true by default', async () => {
    const el = document.createElement('input')
    document.body.append(el)
    let opts: { preventScroll?: boolean } | undefined
    el.focus = ((o?: { preventScroll?: boolean }) => {
      opts = o
    }) as never
    focusOn(() => el, () => true)
    await Promise.resolve()
    expect(opts?.preventScroll).toBe(true)
    el.remove()
  })

  it('select: false skips the .select() call', async () => {
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

  it('select: true (default) calls .select() on the element', async () => {
    const el = document.createElement('input')
    document.body.append(el)
    let selected = 0
    ;(el as HTMLInputElement).select = () => {
      selected++
    }
    focusOn(() => el, () => true)
    await Promise.resolve()
    expect(selected).toBe(1)
    el.remove()
  })

  it('returns a cancel function that aborts a pending focus', async () => {
    const el = document.createElement('input')
    document.body.append(el)
    let called = 0
    el.focus = () => {
      called++
    }
    const cancel = focusOn(() => el, () => true)
    cancel()
    await Promise.resolve()
    expect(called).toBe(0)
    el.remove()
  })

  it('cancel is a no-op when called after the microtask has fired', async () => {
    const el = document.createElement('input')
    document.body.append(el)
    let called = 0
    el.focus = () => {
      called++
    }
    const cancel = focusOn(() => el, () => true)
    await Promise.resolve()
    cancel()
    expect(called).toBe(1)
    el.remove()
  })

  it('returns a no-op cancel when when() is false', () => {
    const cancel = focusOn(() => null, () => false)
    expect(typeof cancel).toBe('function')
    expect(() => cancel()).not.toThrow()
  })
})