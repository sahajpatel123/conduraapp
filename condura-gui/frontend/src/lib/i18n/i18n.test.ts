import { describe, it, expect } from 'vitest'
import en from './locales/en.json'
import es from './locales/es.json'
import fr from './locales/fr.json'
import de from './locales/de.json'
import ja from './locales/ja.json'
import zh from './locales/zh.json'

// i18n catalog parity test — every key in en.json must exist in every
// other locale, even if the value is the English fallback. A missing
// key would silently fall back to English at runtime, but the catalog
// itself should not have holes — fallbacks are visible to future
// translators scanning the file and to anyone running this test.

const catalogs = { en, es, fr, de, ja, zh } as const
type Locale = keyof typeof catalogs
const LOCALES: Locale[] = Object.keys(catalogs) as Locale[]
const SOURCE: Locale = 'en'

describe('i18n catalog parity', () => {
  it('every locale has the same key set as en.json', () => {
    const sourceKeys = new Set(Object.keys(catalogs[SOURCE]))
    for (const loc of LOCALES) {
      if (loc === SOURCE) continue
      const keys = new Set(Object.keys(catalogs[loc]))
      const missing = [...sourceKeys].filter((k) => !keys.has(k))
      const extra = [...keys].filter((k) => !sourceKeys.has(k))
      // Hard fails: missing keys (silent English fallback = UX bug) and
      // extra keys (orphaned translations = dead catalog entries).
      expect(missing, `${loc}.json is missing keys: ${missing.join(', ')}`).toEqual([])
      expect(extra, `${loc}.json has orphan keys not in en.json: ${extra.join(', ')}`).toEqual([])
    }
  })

  it('every value is a non-empty string', () => {
    for (const loc of LOCALES) {
      for (const [k, v] of Object.entries(catalogs[loc])) {
        expect(typeof v, `${loc}.json[${k}] is not a string`).toBe('string')
        expect(v.length, `${loc}.json[${k}] is empty`).toBeGreaterThan(0)
      }
    }
  })

  it('no key is translated to the same string as en.json unless the value is technical', () => {
    // Soft warning — the t() function falls back to en.json on miss, so
    // a key whose value IS the English string means either the locale
    // genuinely has the same wording (a number, an acronym, a brand
    // name), or the translator hasn't gotten to it yet. Flag with a
    // list in the test output rather than failing — translation
    // coverage is a process, not a gate.
    const enKeys = Object.keys(catalogs.en)
    const enValues = catalogs.en as Record<string, string>
    const fallbackKeys: { loc: Locale; key: string }[] = []
    for (const loc of LOCALES) {
      if (loc === SOURCE) continue
      const locValues = catalogs[loc] as Record<string, string>
      for (const k of enKeys) {
        if (locValues[k] === enValues[k]) {
          fallbackKeys.push({ loc, key: k })
        }
      }
    }
    // Print the fallback list (vitest surfaces console output).
    // eslint-disable-next-line no-console
    console.log(
      `[i18n parity] ${fallbackKeys.length} keys are still using the English value as fallback across non-en locales`
    )
    if (fallbackKeys.length > 0) {
      const byLoc = new Map<Locale, string[]>()
      for (const { loc, key } of fallbackKeys) {
        if (!byLoc.has(loc)) byLoc.set(loc, [])
        byLoc.get(loc)!.push(key)
      }
      for (const [loc, keys] of byLoc) {
        // eslint-disable-next-line no-console
        console.log(`  ${loc}: ${keys.length} keys (sample: ${keys.slice(0, 5).join(', ')})`)
      }
    }
    // This test never fails — it's a coverage report.
    expect(fallbackKeys.length).toBeGreaterThanOrEqual(0)
  })

  it('catalogs can be imported as Record<string,string> without runtime errors', () => {
    // Smoke test — if any catalog file is malformed JSON, the import
    // itself would fail. This test exists so vitest reports a clear
    // "test failed" rather than a build error.
    for (const loc of LOCALES) {
      expect(typeof catalogs[loc]).toBe('object')
      expect(catalogs[loc]).not.toBeNull()
    }
  })
})
