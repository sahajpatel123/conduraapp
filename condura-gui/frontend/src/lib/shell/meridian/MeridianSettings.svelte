<script lang="ts">
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { settings } from '../../stores/settings.svelte'
  import {
    getResolvedTheme,
    onThemeChange,
    setResolvedTheme,
    type ResolvedTheme,
  } from '../../theme/condura-theme'

  let theme = $state<ResolvedTheme>(getResolvedTheme())
  let saving = $state(false)
  let note = $state('')

  onMount(() => {
    theme = getResolvedTheme()
    const off = onThemeChange((resolved) => {
      theme = resolved
    })
    void settings.refresh()
    return off
  })

  async function savePatch(patch: Record<string, unknown>): Promise<void> {
    saving = true
    note = ''
    try {
      await settings.save(patch as never)
      note = 'Saved'
    } catch (e) {
      const s = String(e)
      note = /IPC client not started|not connected|Failed to fetch/i.test(s)
        ? 'Daemon offline — connect to save'
        : s
    } finally {
      saving = false
    }
  }

  function setTheme(t: ResolvedTheme): void {
    theme = setResolvedTheme(t)
  }
</script>

<MeridianPage
  kicker="Defaults"
  title="Settings"
  lead="Power, appearance, and the defaults Condura uses when you ask."
>
  <section class="md-panel md-panel-static block">
    <h2>Appearance</h2>
    <p class="hint">Choose the desk lighting. Shift+T toggles anytime.</p>
    <div class="seg" role="group" aria-label="Theme">
      <button type="button" class:on={theme === 'light'} onclick={() => setTheme('light')}>Light</button>
      <button type="button" class:on={theme === 'dark'} onclick={() => setTheme('dark')}>Dark</button>
    </div>
  </section>
  <section class="md-panel md-panel-static block">
    <h2>Daemon</h2>
    {#if !settings.config}
      <p class="muted">Config unavailable offline. Connect the daemon to edit defaults.</p>
    {:else}
      <div class="md-field">
        <label for="spend">Daily spend cap (USD)</label>
        <input
          id="spend"
          type="number"
          min="0"
          step="1"
          value={settings.config.security?.spend_limit_usd_per_day ?? 0}
          onchange={(e) =>
            void savePatch({
              security: {
                ...settings.config!.security,
                spend_limit_usd_per_day: Number((e.currentTarget as HTMLInputElement).value),
              },
            })}
        />
      </div>
      <div class="md-field">
        <label for="autonomy">Default autonomy</label>
        <select
          id="autonomy"
          value={settings.config.autonomy?.default_level ?? 'ask'}
          onchange={(e) =>
            void savePatch({
              autonomy: {
                ...settings.config!.autonomy,
                default_level: (e.currentTarget as HTMLSelectElement).value,
              },
            })}
        >
          <option value="ask">Ask first</option>
          <option value="suggest">Suggest</option>
          <option value="auto">Auto (still gated)</option>
        </select>
      </div>
      {#if note}
        <p class="note" class:saving class:warn={note.includes('offline') || note.includes('Error')}>
          {saving ? 'Saving…' : note}
        </p>
      {/if}
    {/if}
  </section>
</MeridianPage>

<style>
  .block {
    margin-bottom: 14px;
  }
  h2 {
    font-family: var(--md-font-display);
    font-size: 18px;
    letter-spacing: -0.03em;
    margin: 0 0 6px;
  }
  .hint {
    margin: 0 0 14px;
    font-size: 13px;
    color: var(--md-ink-faint);
    line-height: 1.4;
  }
  .seg {
    display: inline-flex;
    padding: 4px;
    border-radius: 999px;
    background: var(--md-stage);
    border: 1px solid var(--md-line);
    box-shadow: inset 0 1px 0 color-mix(in oklab, var(--md-surface) 40%, transparent);
  }
  .seg button {
    padding: 8px 18px;
    border-radius: 999px;
    font-weight: 700;
    font-size: 13px;
    color: var(--md-ink-mute);
    cursor: pointer;
    transition:
      background 180ms var(--md-ease),
      color 180ms var(--md-ease),
      transform 160ms var(--md-spring),
      box-shadow 180ms var(--md-ease);
  }
  .seg button:hover:not(.on) {
    color: var(--md-ink);
  }
  .seg button.on {
    background: var(--md-cobalt);
    color: #fff;
    box-shadow: 0 8px 18px -10px color-mix(in oklab, var(--md-cobalt) 70%, transparent);
  }
  .seg button:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .seg button.on:focus-visible {
    box-shadow: 0 8px 18px -10px color-mix(in oklab, var(--md-cobalt) 70%, transparent), var(--md-focus);
  }
  .muted {
    color: var(--md-ink-mute);
    font-size: 14px;
    margin: 0;
    line-height: 1.5;
  }
  .note {
    margin: 12px 0 0;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--md-live);
    animation: md-rise 280ms var(--md-ease) both;
  }
  .note.saving {
    color: var(--md-ink-faint);
  }
  .note.warn {
    color: var(--md-halt);
  }
  @media (max-width: 420px) {
    .seg {
      width: 100%;
    }
    .seg button {
      flex: 1;
      text-align: center;
    }
  }
</style>
