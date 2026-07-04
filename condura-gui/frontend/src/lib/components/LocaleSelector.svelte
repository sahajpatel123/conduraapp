<script lang="ts">
  // LocaleSelector — small dropdown to pick the UI language.
  //
  // Shows the six supported locales with their native names.
  // Persists to localStorage via the i18n module's setLocale().
  // On mount, asks the daemon for its locale catalog and merges it
  // into the in-memory catalog.
  import { onMount } from 'svelte'
  import Select from './ui/Select.svelte'
  import {
    locale,
    t,
    setLocale,
    SUPPORTED_LOCALES,
    type Locale,
    mergeDaemonCatalog,
  } from '../i18n'
  import { ipc } from '../ipc/client'

  const localeNames: Record<Locale, string> = {
    en: 'English',
    es: 'Español',
    fr: 'Français',
    de: 'Deutsch',
    ja: '日本語',
    zh: '中文',
  }

  // currentLocale mirrors the locale store reactively so the
  // Select's bound value stays in sync.
  let currentLocale: Locale = $state('en')
  $effect(() => {
    return locale.subscribe((loc) => {
      currentLocale = loc
    })
  })

  const options = SUPPORTED_LOCALES.map((loc) => ({
    value: loc,
    label: localeNames[loc],
  }))

  function onChange(v: string): void {
    setLocale(v as Locale)
  }

  onMount(() => {
    void ipc
      .i18nLocale(currentLocale)
      .then((r) => {
        mergeDaemonCatalog(r.locale as Locale, r.translations)
      })
      .catch(() => {
        // Daemon may be unreachable; the bundled catalog still works.
      })
  })
</script>

<Select
  value={currentLocale}
  options={options}
  fullWidth={false}
  onchange={onChange}
/>

<style>
  :global(.locale-select) {
    min-width: 140px;
  }
</style>