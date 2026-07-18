import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import type { Mock } from 'vitest'
import { render, fireEvent, cleanup } from '@testing-library/svelte'
import MeridianDock from './MeridianDock.svelte'
import type { RouteId } from './routes'

// MeridianDock — toolbar with roving tabindex keyboard nav.
//
// Pins the W3C APG "toolbar" pattern: ArrowLeft/Right move focus + selection,
// Home/End jump to extremes, wrap-around at both ends. Each tab carries a
// data-tab-id so the roving focus manager can find its target after route
// change. The Halt button sits OUTSIDE the roving order (tabindex=-1) —
// it's an action, not a destination.

describe('MeridianDock', () => {
  let onnavigate: Mock<(r: RouteId) => void>

  beforeEach(() => {
    onnavigate = vi.fn<(r: RouteId) => void>()
    // matchMedia is referenced by $effect — provide a stub for jsdom.
    Object.defineProperty(window, 'matchMedia', {
      value: (q: string) => ({
        matches: false,
        media: q,
        onchange: null,
        addListener: () => {},
        removeListener: () => {},
        addEventListener: () => {},
        removeEventListener: () => {},
        dispatchEvent: () => false,
      }),
      configurable: true,
    })
    // jsdom doesn't implement Element#scrollIntoView — the dock $effect calls
    // it on route change. Stub globally so the queueMicrotask in $effect
    // doesn't throw an uncaught exception after each render.
    if (!HTMLElement.prototype.scrollIntoView) {
      HTMLElement.prototype.scrollIntoView = function () {} as never
    }
    if (!HTMLElement.prototype.focus) {
      HTMLElement.prototype.focus = function () {} as never
    }
  })

  afterEach(() => {
    cleanup()
  })

  const ids = [
    'chat', 'hub', 'skills', 'sync', 'audit', 'replay', 'channels', 'delegation',
    'account', 'settings', 'about',
  ] as const

  function renderAt(route: RouteId) {
    return render(MeridianDock, { props: { route, onnavigate } })
  }

  function tabFor(container: HTMLElement, id: string): HTMLElement {
    const el = container.querySelector(`[data-tab-id="${id}"]`) as HTMLElement
    if (!el) throw new Error(`No tab for ${id}`)
    return el
  }

  it('renders every primary + more tab with a data-tab-id', () => {
    const { container } = renderAt('chat')
    for (const id of ids) {
      expect(tabFor(container, id), id).toBeInTheDocument()
    }
  })

  it('marks only the active tab with tabindex=0 (roving pattern)', () => {
    const { container } = renderAt('audit')
    for (const id of ids) {
      const tab = tabFor(container, id)
      const expected = id === 'audit' ? '0' : '-1'
      expect(tab.getAttribute('tabindex')).toBe(expected)
    }
  })

  it('sets aria-current=page on the active tab only', () => {
    const { container } = renderAt('settings')
    expect(tabFor(container, 'settings').getAttribute('aria-current')).toBe('page')
    expect(tabFor(container, 'chat').getAttribute('aria-current')).toBeNull()
    expect(tabFor(container, 'audit').getAttribute('aria-current')).toBeNull()
  })

  it('Halt button stays outside the roving order', () => {
    const { container } = renderAt('chat')
    const halt = container.querySelector('.halt') as HTMLElement
    expect(halt.getAttribute('tabindex')).toBe('-1')
    expect(halt.getAttribute('data-tab-id')).toBeNull()
  })

  it('ArrowRight on chat navigates to hub', async () => {
    const { container } = renderAt('chat')
    await fireEvent.keyDown(container.querySelector('.dock') as HTMLElement, { key: 'ArrowRight' })
    expect(onnavigate).toHaveBeenCalledWith('hub')
  })

  it('ArrowLeft on chat wraps to about (last tab)', async () => {
    const { container } = renderAt('chat')
    await fireEvent.keyDown(container.querySelector('.dock') as HTMLElement, { key: 'ArrowLeft' })
    expect(onnavigate).toHaveBeenCalledWith('about')
  })

  it('ArrowRight on about wraps to chat', async () => {
    const { container } = renderAt('about')
    await fireEvent.keyDown(container.querySelector('.dock') as HTMLElement, { key: 'ArrowRight' })
    expect(onnavigate).toHaveBeenCalledWith('chat')
  })

  it('Home jumps to chat (first tab)', async () => {
    const { container } = renderAt('replay')
    await fireEvent.keyDown(container.querySelector('.dock') as HTMLElement, { key: 'Home' })
    expect(onnavigate).toHaveBeenCalledWith('chat')
  })

  it('End jumps to about (last tab)', async () => {
    const { container } = renderAt('replay')
    await fireEvent.keyDown(container.querySelector('.dock') as HTMLElement, { key: 'End' })
    expect(onnavigate).toHaveBeenCalledWith('about')
  })

  it('unhandled keys do not navigate (e.g. ArrowDown is vertical, not used here)', async () => {
    const { container } = renderAt('chat')
    await fireEvent.keyDown(container.querySelector('.dock') as HTMLElement, { key: 'ArrowDown' })
    await fireEvent.keyDown(container.querySelector('.dock') as HTMLElement, { key: 'a' })
    expect(onnavigate).not.toHaveBeenCalled()
  })

  it('ArrowRight on delegation (last primary) advances to account (first more)', async () => {
    const { container } = renderAt('delegation')
    await fireEvent.keyDown(container.querySelector('.dock') as HTMLElement, { key: 'ArrowRight' })
    expect(onnavigate).toHaveBeenCalledWith('account')
  })

  it('ArrowLeft on account (first more) goes back to delegation (last primary)', async () => {
    const { container } = renderAt('account')
    await fireEvent.keyDown(container.querySelector('.dock') as HTMLElement, { key: 'ArrowLeft' })
    expect(onnavigate).toHaveBeenCalledWith('delegation')
  })
})