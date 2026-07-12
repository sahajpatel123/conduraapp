import { writable } from 'svelte/store'
import { ipc } from './ipc/client'

// Bundled catalogs — always available without the daemon or a /locales fetch.
// (Vite was serving SPA HTML for /locales/*.json when public/ was missing,
// so the old fetch fallback silently parsed nothing and left raw i18n keys.)
import en from './i18n/locales/en.json'
import es from './i18n/locales/es.json'
import fr from './i18n/locales/fr.json'
import de from './i18n/locales/de.json'
import ja from './i18n/locales/ja.json'
import zh from './i18n/locales/zh.json'

type Locale = 'en' | 'es' | 'fr' | 'de' | 'ja' | 'zh'
type Catalog = Record<string, string>
type Catalogs = Record<Locale, Catalog>

const DEFAULT_LOCALE: Locale = 'en'
const SUPPORTED_LOCALES: Locale[] = ['en', 'es', 'fr', 'de', 'ja', 'zh']

const localeCatalogs: Catalogs = {
  en: { ...(en as Catalog) },
  es: { ...(es as Catalog) },
  fr: { ...(fr as Catalog) },
  de: { ...(de as Catalog) },
  ja: { ...(ja as Catalog) },
  zh: { ...(zh as Catalog) },
}

const isBrowser = typeof window !== 'undefined' && typeof document !== 'undefined'

// Bumped when catalogs change so Svelte templates that read it re-render.
export const catalogVersion = writable(0)

let currentLocale: Locale = DEFAULT_LOCALE

function bumpCatalog(): void {
  catalogVersion.update((n) => n + 1)
}

function isJsonResponse(res: Response): boolean {
  const ct = res.headers.get('content-type') || ''
  return ct.includes('application/json') || ct.includes('text/json')
}

async function loadCatalog(locale: Locale): Promise<Catalog> {
  // Prefer daemon catalog (may include newer keys), then static public file.
  try {
    const response = await ipc.i18nLocale(locale)
    if (response?.translations && Object.keys(response.translations).length > 0) {
      localeCatalogs[locale] = {
        ...localeCatalogs[locale],
        ...response.translations,
      }
      bumpCatalog()
      return localeCatalogs[locale]
    }
  } catch {
    // Daemon offline / IPC not started — fall through.
  }

  try {
    const response = await fetch('/locales/' + locale + '.json', {
      headers: { Accept: 'application/json' },
    })
    if (!response.ok) throw new Error('Failed to load ' + locale)
    // Vite SPA fallback returns index.html with 200 — reject that.
    if (!isJsonResponse(response)) {
      const preview = (await response.clone().text()).slice(0, 40)
      if (preview.trimStart().startsWith('<!') || preview.trimStart().startsWith('<html')) {
        throw new Error('locales endpoint returned HTML, not JSON')
      }
    }
    const catalog = (await response.json()) as Catalog
    if (catalog && typeof catalog === 'object') {
      localeCatalogs[locale] = { ...localeCatalogs[locale], ...catalog }
      bumpCatalog()
      return localeCatalogs[locale]
    }
  } catch {
    // Bundled catalog already seeded — fine offline.
  }

  return localeCatalogs[locale]
}

function getInitialLocale(): Locale {
  if (!isBrowser) return DEFAULT_LOCALE
  try {
    const saved = localStorage.getItem('condura_locale') as Locale | null
    if (saved && SUPPORTED_LOCALES.includes(saved)) return saved
  } catch {
    /* private mode */
  }
  try {
    const navLang = navigator.language.split('-')[0] as Locale
    if (SUPPORTED_LOCALES.includes(navLang)) return navLang
  } catch {
    /* ignore */
  }
  return DEFAULT_LOCALE
}

export const locale = writable<Locale>(getInitialLocale())

locale.subscribe((loc) => {
  currentLocale = loc
  void loadCatalog(loc)
  if (isBrowser) {
    try {
      localStorage.setItem('condura_locale', loc)
    } catch {
      /* ignore */
    }
    try {
      document.documentElement.lang = loc
    } catch {
      /* ignore */
    }
  }
})

/**
 * Synchronous translator. Catalogs are pre-seeded from bundled JSON so
 * first paint never shows raw keys for known English strings.
 * Pass `catalogVersion` into a template (`{$catalogVersion}`) if a
 * component must refresh after a late daemon merge.
 */
export function t(key: string, ...args: unknown[]): string {
  let template: string | undefined = localeCatalogs[currentLocale]?.[key]
  if (!template && currentLocale !== DEFAULT_LOCALE) {
    template = localeCatalogs[DEFAULT_LOCALE][key]
  }
  if (!template) return key
  try {
    return template.replace(/{(\d+)}/g, (_match: string, i: string) => {
      const idx = parseInt(i, 10)
      return args[idx] !== undefined && args[idx] !== null ? String(args[idx]) : ''
    })
  } catch {
    return template
  }
}

export function setLocale(loc: Locale) {
  if (SUPPORTED_LOCALES.includes(loc)) {
    locale.set(loc)
  }
}

/** Merge daemon-provided translations into the in-memory catalog. */
export function mergeDaemonCatalog(loc: Locale, translations: Record<string, string>) {
  if (!translations || Object.keys(translations).length === 0) return
  localeCatalogs[loc] = { ...localeCatalogs[loc], ...translations }
  bumpCatalog()
}

export { SUPPORTED_LOCALES, DEFAULT_LOCALE }
export type { Locale }
