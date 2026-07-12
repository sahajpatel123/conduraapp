import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

// Canonical catalogs live next to the i18n module (bundled at build time).
const localesDir = resolve(__dirname, 'locales')

function load(name: string): Record<string, string> {
  return JSON.parse(readFileSync(resolve(localesDir, name), 'utf8'))
}

describe('locale key parity', () => {
  const en = load('en.json')
  const locales = ['es.json', 'fr.json', 'de.json', 'ja.json', 'zh.json']

  for (const loc of locales) {
    it(`${loc} has every en key`, () => {
      const d = load(loc)
      const missing = Object.keys(en).filter((k) => !(k in d))
      expect(missing).toEqual([])
    })
  }

  it('non-EN sync.pair keys are not English placeholders for title', () => {
    const enTitle = en['sync.pair.title']
    for (const loc of locales) {
      const d = load(loc)
      // Must exist and (for non-en) differ from English where we translated
      expect(d['sync.pair.title']).toBeTruthy()
      if (loc !== 'en.json') {
        expect(d['sync.pair.title']).not.toBe(enTitle)
      }
    }
  })
})
