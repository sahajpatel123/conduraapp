import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { trapTab, navigateListRows } from './focusTrap'

describe('trapTab', () => {
  it('wraps Tab from last focusable to first', () => {
    const root = document.createElement('div')
    const a = document.createElement('button')
    const b = document.createElement('button')
    a.textContent = 'a'
    b.textContent = 'b'
    root.append(a, b)
    document.body.append(root)
    b.focus()
    const e = new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true })
    Object.defineProperty(e, 'shiftKey', { value: false })
    const prevent = vi.spyOn(e, 'preventDefault')
    // activeElement is b
    trapTab(root, e)
    expect(prevent).toHaveBeenCalled()
    expect(document.activeElement).toBe(a)
    root.remove()
  })
})

describe('navigateListRows', () => {
  it('moves focus with ArrowDown and Home/End', () => {
    const rows = [0, 1, 2].map((i) => {
      const b = document.createElement('button')
      b.textContent = String(i)
      document.body.append(b)
      return b
    })
    const down = new KeyboardEvent('keydown', { key: 'ArrowDown', cancelable: true })
    const next = navigateListRows(down, rows, 0)
    expect(next).toBe(1)
    expect(document.activeElement).toBe(rows[1])

    const end = new KeyboardEvent('keydown', { key: 'End', cancelable: true })
    const last = navigateListRows(end, rows, 1)
    expect(last).toBe(2)
    expect(document.activeElement).toBe(rows[2])

    const home = new KeyboardEvent('keydown', { key: 'Home', cancelable: true })
    const first = navigateListRows(home, rows, 2)
    expect(first).toBe(0)
    rows.forEach((r) => r.remove())
  })
})
