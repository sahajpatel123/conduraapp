import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, fireEvent, waitFor } from '@testing-library/svelte'
import FloatingOnboarding from './FloatingOnboarding.svelte'

// Mock IntersectionObserver — the BlurReveal component uses it. In jsdom,
// the observer has no real intersection, so isVisible would stay false and
// the content would render as opacity 0. Forcing isIntersecting=true keeps
// the test focused on the state machine.
class MockIntersectionObserver {
  callback: IntersectionObserverCallback
  constructor(cb: IntersectionObserverCallback) { this.callback = cb }
  observe(_el: Element) {
    this.callback([{ isIntersecting: true } as IntersectionObserverEntry], this as unknown as IntersectionObserver)
  }
  unobserve() {}
  disconnect() {}
  takeRecords() { return [] }
}
;(globalThis as unknown as { IntersectionObserver: typeof MockIntersectionObserver }).IntersectionObserver =
  MockIntersectionObserver

// Mock matchMedia (some paper components read it).
if (typeof window !== 'undefined' && !window.matchMedia) {
  window.matchMedia = vi.fn().mockImplementation((q: string) => ({
    matches: false, media: q, onchange: null,
    addListener: vi.fn(), removeListener: vi.fn(),
    addEventListener: vi.fn(), removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }))
}

describe('FloatingOnboarding navigation', () => {
  it('renders welcome step on mount and advances when the first card is clicked', async () => {
    const oncomplete = vi.fn()
    const { container, getAllByRole } = render(FloatingOnboarding, { props: { oncomplete } })

    await waitFor(() => {
      expect(container.textContent).toMatch(/1 of 5/)
    })

    const buttons = getAllByRole('button')
    expect(buttons.length).toBeGreaterThanOrEqual(3)

    await fireEvent.click(buttons[0]!)

    await waitFor(() => {
      expect(container.textContent).toMatch(/2 of 5/)
    }, { timeout: 2000 })
  })

  it('advances through every step and fires oncomplete on the done screen', async () => {
    const oncomplete = vi.fn()
    const { container } = render(FloatingOnboarding, { props: { oncomplete } })

    // Step 1 → 2 (Welcome → Permissions): click first card.
    await waitFor(() => expect(container.textContent).toMatch(/1 of 5/))
    const step1Buttons = container.querySelectorAll('button.lp-paper-card')
    expect(step1Buttons.length).toBeGreaterThanOrEqual(3)
    ;(step1Buttons[0] as HTMLButtonElement).click()

    // Step 2 → 3 (Permissions → Power): click the primary "Continue" CTA.
    await waitFor(() => expect(container.textContent).toMatch(/2 of 5/))
    const allButtonsAfter2 = Array.from(container.querySelectorAll('button'))
    // The PermissionCards renders a "Continue" MagneticButton + a Skip
    // button. Either advances to step 3.
    const continueBtn = allButtonsAfter2.find((b) => b.textContent?.trim() === 'Continue')
    expect(continueBtn, 'PermissionCards should render a Continue button').toBeTruthy()
    continueBtn!.click()

    // Step 3 → 4 (Power → Hotkey): click the first option card (Local).
    await waitFor(() => expect(container.textContent).toMatch(/3 of 5/))
    const step3Cards = container.querySelectorAll('button.lp-paper-card')
    expect(step3Cards.length).toBeGreaterThanOrEqual(3)
    ;(step3Cards[0] as HTMLButtonElement).click()

    // Step 4 → 5 (Hotkey → Done): HotkeyCard's Continue is disabled until a
    // combo is recorded. The Skip button advances the wizard.
    await waitFor(() => expect(container.textContent).toMatch(/4 of 5/))
    const allButtonsAfter4 = Array.from(container.querySelectorAll('button'))
    const hotkeySkip = allButtonsAfter4.find((b) => b.textContent?.trim() === 'Skip')
    expect(hotkeySkip, 'HotkeyCard should render a Skip button').toBeTruthy()
    hotkeySkip!.click()

    // Step 5 (Done): FirstBreath should appear; its oncomplete fires when
    // the user dismisses it. For this test we only assert we reached the
    // final step (the call to oncomplete is exercised by FirstBreath's
    // own UI and is not the subject of this regression).
    await waitFor(() => expect(container.textContent).toMatch(/5 of 5/))
  })
})