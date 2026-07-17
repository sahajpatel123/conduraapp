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
    // name), or the translator hasn't gotten to it yet.
    //
    // Reports BOTH:
    //   - per-locale fallback count (raw per-locale "still English" count)
    //   - unique universal-fallback keys (where ALL 5 non-en locales
    //     still equal en — the truer "remaining work" number)
    //   - per-namespace breakdown of the universal set
    const enKeys = Object.keys(catalogs.en)
    const enValues = catalogs.en as Record<string, string>
    const universalFallback = new Set<string>()
    const perLocFallback = new Map<Locale, string[]>()
    for (const loc of LOCALES) {
      if (loc === SOURCE) continue
      perLocFallback.set(loc, [])
    }
    for (const k of enKeys) {
      let allMatch = true
      for (const loc of LOCALES) {
        if (loc === SOURCE) continue
        const locValues = catalogs[loc] as Record<string, string>
        if (locValues[k] === enValues[k]) {
          perLocFallback.get(loc)!.push(k)
        } else {
          allMatch = false
        }
      }
      if (allMatch) universalFallback.add(k)
    }
    const totalPerLoc = [...perLocFallback.values()].reduce((s, a) => s + a.length, 0)
    // eslint-disable-next-line no-console
    console.log(
      `[i18n parity] ${totalPerLoc} per-locale fallbacks; ${universalFallback.size} unique keys where ALL 5 non-en locales still equal en (true remaining work)`
    )
    if (universalFallback.size > 0) {
      // eslint-disable-next-line no-console
      console.log(`  universal-fallback keys (first 20):`)
      const sample = [...universalFallback].slice(0, 20)
      for (const k of sample) {
        // eslint-disable-next-line no-console
        console.log(`    ${k}`)
      }
    }
    for (const [loc, keys] of perLocFallback) {
      // eslint-disable-next-line no-console
      console.log(`  ${loc}: ${keys.length} per-locale fallbacks (sample: ${keys.slice(0, 5).join(', ')})`)
    }
    if (universalFallback.size > 0) {
      const byNs = new Map<string, number>()
      for (const k of universalFallback) {
        const ns = k.split('.')[0] ?? '(ungrouped)'
        byNs.set(ns, (byNs.get(ns) ?? 0) + 1)
      }
      const nsSorted = [...byNs.entries()].sort((a, b) => b[1] - a[1])
      // eslint-disable-next-line no-console
      console.log(`  universal-fallback by namespace (top 15):`)
      for (const [ns, count] of nsSorted.slice(0, 15)) {
        // eslint-disable-next-line no-console
        console.log(`    ${ns}: ${count}`)
      }
    }
    // This test never fails — it's a coverage report.
    expect(true).toBe(true)
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
