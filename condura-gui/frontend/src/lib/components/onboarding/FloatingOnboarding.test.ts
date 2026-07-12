import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, fireEvent, waitFor } from '@testing-library/svelte'
import FloatingOnboarding from './FloatingOnboarding.svelte'

const { mockOnboarding } = vi.hoisted(() => {
  const mockOnboarding = {
    daemon: { current_step: 'eula', steps: {} },
    busy: false,
    error: null as string | null,
    hotkeyValue: '',
    eulaVersion: 'v1',
    get isComplete() {
      return false
    },
    sync: vi.fn().mockResolvedValue(undefined),
    acceptEula: vi.fn().mockImplementation(async function (this: typeof mockOnboarding) {
      this.daemon = { ...this.daemon, current_step: 'permissions' }
    }),
    completePermissions: vi.fn().mockImplementation(async function (this: typeof mockOnboarding) {
      this.daemon = { ...this.daemon, current_step: 'hotkey' }
    }),
    skipStep: vi.fn().mockResolvedValue(undefined),
    probePower: vi.fn().mockResolvedValue(undefined),
    setHotkey: vi.fn(),
    saveHotkey: vi.fn().mockResolvedValue(undefined),
    finish: vi.fn().mockResolvedValue({ ok: true }),
  }
  return { mockOnboarding }
})

vi.mock('../../stores/onboarding.svelte', () => ({
  onboarding: mockOnboarding,
}))

// EulaScreen gates Accept on daemon.connected — accepting the license is a
// daemon write. Without this mock, tests run "offline" and Accept is a no-op.
vi.mock('../../stores/daemon.svelte', () => ({
  daemon: { connected: true, lastError: '', baseURL: 'http://127.0.0.1:7666' },
}))

vi.mock('../../ipc/client', () => ({
  ipc: {
    onboardingEula: vi.fn().mockResolvedValue({
      version: 'v1',
      text: 'Short EULA for tests.',
    }),
    firstRunComplete: vi.fn().mockResolvedValue(undefined),
    permissionsStatus: vi.fn().mockResolvedValue([]),
    permissionsGuide: vi.fn().mockResolvedValue({ deep_link: '' }),
    permissionsOpenSettings: vi.fn().mockResolvedValue({
      guide: { deep_link: '', steps: [] },
      opened: false,
    }),
  },
}))

// Mock IntersectionObserver — BlurReveal uses it. In jsdom there is no
// real intersection, so isVisible would stay false without this shim.
class MockIntersectionObserver {
  callback: IntersectionObserverCallback
  constructor(cb: IntersectionObserverCallback) {
    this.callback = cb
  }
  observe(_el: Element) {
    this.callback(
      [{ isIntersecting: true } as IntersectionObserverEntry],
      this as unknown as IntersectionObserver
    )
  }
  unobserve() {}
  disconnect() {}
  takeRecords() {
    return []
  }
}
;(globalThis as unknown as { IntersectionObserver: typeof MockIntersectionObserver }).IntersectionObserver =
  MockIntersectionObserver

if (typeof window !== 'undefined' && !window.matchMedia) {
  window.matchMedia = vi.fn().mockImplementation((q: string) => ({
    matches: false,
    media: q,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }))
}

describe('FloatingOnboarding navigation', () => {
  beforeEach(() => {
    mockOnboarding.daemon = { current_step: 'eula', steps: {} }
    mockOnboarding.busy = false
    mockOnboarding.error = null
    mockOnboarding.hotkeyValue = ''
    vi.clearAllMocks()
    if (!SVGElement.prototype.getTotalLength) {
      SVGElement.prototype.getTotalLength = () => 100
    }
  })

  it('renders EULA step on mount and advances after accept', async () => {
    const oncomplete = vi.fn()
    const { container } = render(FloatingOnboarding, { props: { oncomplete } })

    await waitFor(() => {
      expect(container.textContent).toMatch(/1 of 5/)
    })

    await waitFor(() => {
      expect(container.querySelector('input[type="checkbox"]')).toBeTruthy()
    })

    const checkbox = container.querySelector('input[type="checkbox"]') as HTMLInputElement
    await fireEvent.click(checkbox)

    const acceptBtn = await waitFor(() => {
      const btn = container.querySelector('.btn-primary') as HTMLButtonElement | null
      expect(btn).toBeTruthy()
      return btn!
    })
    await fireEvent.click(acceptBtn)

    await waitFor(() => {
      expect(container.textContent).toMatch(/2 of 5/)
    })
    expect(mockOnboarding.acceptEula).toHaveBeenCalled()
  })

  it('advances through every step and fires oncomplete on finish', async () => {
    const oncomplete = vi.fn()
    const { container } = render(FloatingOnboarding, { props: { oncomplete } })

    // Step 1 → 2 (EULA → Permissions)
    await waitFor(() => expect(container.textContent).toMatch(/1 of 5/))
    await waitFor(() => {
      expect(container.querySelector('input[type="checkbox"]')).toBeTruthy()
    })
    const checkbox = container.querySelector('input[type="checkbox"]') as HTMLInputElement
    await fireEvent.click(checkbox)
    const acceptBtn = await waitFor(() => {
      const btn = container.querySelector('.btn-primary') as HTMLButtonElement | null
      expect(btn).toBeTruthy()
      return btn!
    })
    await fireEvent.click(acceptBtn)

    // Step 2 → 3 (Permissions → Power): Skip (no grants in test env)
    await waitFor(() => expect(container.textContent).toMatch(/2 of 5/))
    const permSkip = Array.from(container.querySelectorAll('button')).find(
      (b) => b.textContent?.trim() === 'Skip for now'
    )
    expect(permSkip).toBeTruthy()
    permSkip!.click()

    // Step 3 → 4 (Power → Hotkey): click first option card
    await waitFor(() => expect(container.textContent).toMatch(/3 of 5/))
    const step3Cards = container.querySelectorAll('button.lp-paper-card')
    expect(step3Cards.length).toBeGreaterThanOrEqual(1)
    ;(step3Cards[0] as HTMLButtonElement).click()

    // Step 4 → 5 (Hotkey → Done): Skip
    await waitFor(() => expect(container.textContent).toMatch(/4 of 5/))
    const hotkeySkip = Array.from(container.querySelectorAll('button')).find(
      (b) => b.textContent?.trim() === 'Skip'
    )
    expect(hotkeySkip).toBeTruthy()
    hotkeySkip!.click()

    // Step 5 (Done): Begin finishes onboarding
    await waitFor(() => expect(container.textContent).toMatch(/5 of 5/))
    const beginBtn = Array.from(container.querySelectorAll('button')).find(
      (b) => b.textContent?.trim() === 'Begin'
    )
    expect(beginBtn).toBeTruthy()
    beginBtn!.click()

    await waitFor(() => {
      expect(mockOnboarding.finish).toHaveBeenCalled()
      expect(oncomplete).toHaveBeenCalled()
    })
  })
})
