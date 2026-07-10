// Vitest setup: loads jest-dom matchers so we can use
// `expect(el).toBeInTheDocument()` etc. in Svelte tests.
import '@testing-library/jest-dom/vitest'

// jsdom does not always expose a working localStorage under every
// Node version / vitest mode. Provide a minimal in-memory store so
// i18n and onboarding screens can mount without TypeError.
const memory = new Map<string, string>()
const localStorageMock: Storage = {
  get length() {
    return memory.size
  },
  clear() {
    memory.clear()
  },
  getItem(key: string) {
    return memory.has(key) ? (memory.get(key) as string) : null
  },
  key(index: number) {
    return Array.from(memory.keys())[index] ?? null
  },
  removeItem(key: string) {
    memory.delete(key)
  },
  setItem(key: string, value: string) {
    memory.set(key, String(value))
  },
}

Object.defineProperty(globalThis, 'localStorage', {
  value: localStorageMock,
  configurable: true,
  writable: true,
})
if (typeof window !== 'undefined') {
  Object.defineProperty(window, 'localStorage', {
    value: localStorageMock,
    configurable: true,
    writable: true,
  })
}
