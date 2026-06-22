import { writable, derived, type Readable } from 'svelte/store';
import { ipc } from './ipc/client';

type Locale = 'en' | 'es' | 'fr' | 'de' | 'ja' | 'zh';
type Catalog = Record<string, string>;
type Catalogs = Record<Locale, Catalog>;

const DEFAULT_LOCALE: Locale = 'en';
const SUPPORTED_LOCALES: Locale[] = ['en', 'es', 'fr', 'de', 'ja', 'zh'];

const localeCatalogs: Catalogs = {
	en: {},
	es: {},
	fr: {},
	de: {},
	ja: {},
	zh: {}
};

const isBrowser = typeof window !== 'undefined' && typeof document !== 'undefined';

async function loadCatalog(locale: Locale): Promise<Catalog> {
	if (Object.keys(localeCatalogs[locale]).length > 0) {
		return localeCatalogs[locale];
	}
	try {
		const response = await ipc.i18nLocale(locale);
		if (response && response.translations) {
			localeCatalogs[locale] = response.translations;
			return response.translations;
		}
	} catch {
		// Fall back to static fetch if IPC fails (e.g. during dev before daemon is ready)
		try {
			const response = await fetch('/locales/' + locale + '.json');
			if (!response.ok) throw new Error('Failed to load ' + locale);
			const catalog = await response.json();
			localeCatalogs[locale] = catalog;
			return catalog;
		} catch {
			// return empty catalog
		}
	}
	return {};
}

function getInitialLocale(): Locale {
	if (!isBrowser) return DEFAULT_LOCALE;
	const saved = localStorage.getItem('condura_locale') as Locale | null;
	if (saved && SUPPORTED_LOCALES.includes(saved)) return saved;
	const navLang = navigator.language.split('-')[0] as Locale;
	if (SUPPORTED_LOCALES.includes(navLang)) return navLang;
	return DEFAULT_LOCALE;
}

export const locale = writable<Locale>(getInitialLocale());

// derived async signature: (store value, set fn, update fn?) => void | Unsubscriber
// Promise<void> is not allowed, so loadCatalog is called inside a synchronous callback.
export const catalog: Readable<Catalog> = derived(
	locale,
	($locale, set) => {
		void loadCatalog($locale).then(set);
	},
	{}
);

export const t: Readable<(key: string, ...args: unknown[]) => string> = derived(
	[locale, catalog],
	([$locale, $catalog]) =>
		(key: string, ...args: unknown[]): string => {
			let template: string | undefined = $catalog[key];
			if (!template && $locale !== DEFAULT_LOCALE) {
				template = localeCatalogs[DEFAULT_LOCALE][key];
			}
			if (!template) return key;
			try {
				return template.replace(/{(\d+)}/g, (_match: string, i: string) => {
					const idx = parseInt(i, 10);
					return args[idx] !== undefined && args[idx] !== null ? String(args[idx]) : '';
				});
			} catch {
				return template;
			}
		}
);

locale.subscribe((loc) => {
	if (isBrowser) {
		localStorage.setItem('condura_locale', loc);
		document.documentElement.lang = loc;
	}
});

export function setLocale(loc: Locale) {
	if (SUPPORTED_LOCALES.includes(loc)) {
		locale.set(loc);
	}
}

/** Merge daemon-provided translations into the in-memory catalog. */
export function mergeDaemonCatalog(loc: Locale, translations: Record<string, string>) {
	if (!translations || Object.keys(translations).length === 0) return;
	localeCatalogs[loc] = { ...localeCatalogs[loc], ...translations };
}

export { SUPPORTED_LOCALES, DEFAULT_LOCALE };
export type { Locale };
