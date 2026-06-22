<script lang="ts">
	import { onMount } from 'svelte';
	import { locale, t, setLocale, SUPPORTED_LOCALES, type Locale, mergeDaemonCatalog } from '../i18n';
	import { ipc } from '../ipc/client';

	const localeNames: Record<Locale, string> = {
		en: 'English',
		es: 'Español',
		fr: 'Français',
		de: 'Deutsch',
		ja: '日本語',
		zh: '中文'
	};

	// currentLocale is the synchronous mirror of the locale store
	// so we can `bind:value` to it. locale.subscribe keeps it in
	// sync; bind:value writes back through onValueChange.
	let currentLocale: Locale = $state('en');
	$effect(() => {
		return locale.subscribe((loc) => {
			currentLocale = loc;
		});
	});

	function onValueChange(e: Event): void {
		const v = (e.target as HTMLSelectElement).value as Locale;
		setLocale(v);
	}

	onMount(() => {
		void ipc.i18nLocale(currentLocale).then((r) => {
			mergeDaemonCatalog(r.locale as Locale, r.translations);
		}).catch(() => {});
	});
</script>

<select
	bind:value={currentLocale}
	onchange={onValueChange}
	class="locale-select"
	aria-label={t('locale.selector.aria_label')}
>
	{#each SUPPORTED_LOCALES as loc}
		<option value={loc}>{localeNames[loc]}</option>
	{/each}
</select>

<style>
	.locale-select {
		background: var(--bg-elevated);
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 6px 10px;
		color: var(--fg);
		font-size: 13px;
		cursor: pointer;
		transition: border-color 0.15s;
	}
	.locale-select:hover {
		border-color: var(--accent);
	}
	.locale-select:focus {
		outline: none;
		border-color: var(--accent);
		box-shadow: 0 0 0 2px var(--accent-alpha);
	}
</style>
