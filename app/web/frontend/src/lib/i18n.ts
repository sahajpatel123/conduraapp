import { writable } from 'svelte/store';
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

// Built-in English fallback for the always-on-screen shell strings.
// Consulted *last* (after the loaded daemon/static catalogs), so real
// translations always win. This guarantees the navigation, command
// palette, and onboarding chrome read correctly on first paint —
// before the daemon catalog has loaded, and in offline/dev preview.
const BUILTIN_FALLBACK: Catalog = {
	'nav.chat': 'Chat',
	'nav.settings': 'Settings',
	'nav.audit': 'Audit',
	'nav.replay': 'Replay',
	'nav.about': 'About',
	'nav.hub': 'Hub',
	'nav.sync': 'Sync',
	'nav.skills': 'Skills',
	'nav.channels': 'Channels',
	'nav.delegation': 'Delegation',
	'common.search': 'Search or run a command',
	'common.no_results': 'No results',
	'common.back': 'Back',
	'common.continue': 'Continue',
	'common.skip': 'Skip for now',
	'common.cancel': 'Cancel',
	'chat.new_chat': 'New chat',
	'chat.placeholder': 'Message Condura…',
	'chat.send': 'Send',
	'chat.stop': 'Stop',
	'chat.assistant': 'Condura',
	'chat.streaming': 'Streaming',
	'chat.thinking': 'Thinking…',
	'chat.messages': 'messages',
	'chat.no_conversations': 'No conversations yet.',
	'chat.composer_hint': 'Shift+Enter for a newline. / for commands.',
	'app.status.connected': 'Connected',
	'app.status.disconnected': 'Disconnected',
	'composer.model_picker_label': 'Model',
	'composer.voice_idle': 'Listening… speak now'
};

// currentLocale is the source of truth for synchronous lookups.
// locale (the writable store below) is still exported for callers
// that need to subscribe/react. We mirror every locale change into
// currentLocale so the synchronous t() function sees the same value.
let currentLocale: Locale = DEFAULT_LOCALE;

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

// Trigger catalog load + mirror into currentLocale whenever locale changes.
// We don't expose a Readable catalog anymore; t() reads directly from the
// in-memory localeCatalogs map via the currentLocale variable.
locale.subscribe((loc) => {
	currentLocale = loc;
	void loadCatalog(loc);
	if (isBrowser) {
		localStorage.setItem('condura_locale', loc);
		document.documentElement.lang = loc;
	}
});

// t is a plain function so it works directly in Svelte 5 template
// expressions as `t('key')` without needing the store-auto-subscribe
// `t(...)` pattern (which is illegal under Svelte 5 runes). It reads
// from the current locale's in-memory catalog, falling back to the
// English catalog when a key isn't translated, then to the key itself.
//
// Args are positional placeholders: `t('greeting', name)` looks up
// `greeting` in the catalog and substitutes {0} with `name`.
export function t(key: string, ...args: unknown[]): string {
	let template: string | undefined = localeCatalogs[currentLocale]?.[key];
	if (!template && currentLocale !== DEFAULT_LOCALE) {
		template = localeCatalogs[DEFAULT_LOCALE][key];
	}
	if (!template) template = BUILTIN_FALLBACK[key];
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
