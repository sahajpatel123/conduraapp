<script lang="ts">
  /**
   * Settings — instrument desk for lighting, spend, and gate defaults.
   * Signature: live contract + autonomy tone plate + spend gauge + shortcut atlas.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { settings } from '../../stores/settings.svelte'
  import {
    getResolvedTheme,
    onThemeChange,
    setResolvedTheme,
    type ResolvedTheme,
  } from '../../theme/condura-theme'

  type Autonomy = 'ask' | 'suggest' | 'auto'

  let theme = $state<ResolvedTheme>(getResolvedTheme())
  let saving = $state(false)
  let note = $state('')
  let modLabel = $state('⌘')
  let noteTimer: ReturnType<typeof setTimeout> | null = null

  const spend = $derived(Number(settings.config?.security?.spend_limit_usd_per_day ?? 0))
  const autonomy = $derived((settings.config?.autonomy?.default_level ?? 'ask') as Autonomy)
  const spendPct = $derived(Math.min(100, Math.round((spend / 50) * 100)))
  const offline = $derived(!settings.config)
  const liveNote = $derived(
    saving
      ? 'Writing defaults to the daemon…'
      : note.includes('offline')
        ? note
        : note === 'Saved'
          ? 'Defaults saved — Gatekeeper still holds the door'
          : offline
            ? 'Daemon offline — lighting still works; spend and autonomy need a connection'
            : `Autonomy · ${autonomy} · spend cap $${spend}/day`
  )

  const AUTONOMY: {
    id: Autonomy
    title: string
    body: string
    tone: string
  }[] = [
    {
      id: 'ask',
      title: 'Ask first',
      body: 'Every gated action waits for you. Safest default.',
      tone: 'cobalt',
    },
    {
      id: 'suggest',
      title: 'Suggest',
      body: 'Condura proposes a plan; you still open the door.',
      tone: 'live',
    },
    {
      id: 'auto',
      title: 'Auto',
      body: 'Routine work may proceed — still Gatekeeper-gated.',
      tone: 'caution',
    },
  ]

  const KEYS = [
    { keys: 'K', label: 'Jump anywhere' },
    { keys: 'Shift+T', label: 'Toggle light / dark' },
    { keys: 'Esc', label: 'Close overlays' },
    { keys: 'Halt', label: 'Cut the line in the dock' },
  ]

  onMount(() => {
    theme = getResolvedTheme()
    modLabel = /Mac|iPhone|iPad/.test(navigator.platform) ? '⌘' : 'Ctrl'
    const off = onThemeChange((resolved) => {
      theme = resolved
    })
    void settings.refresh()
    return () => {
      off()
      if (noteTimer) clearTimeout(noteTimer)
    }
  })

  async function savePatch(patch: Record<string, unknown>): Promise<void> {
    saving = true
    note = ''
    if (noteTimer) clearTimeout(noteTimer)
    try {
      await settings.save(patch as never)
      note = 'Saved'
      noteTimer = setTimeout(() => {
        if (note === 'Saved') note = ''
      }, 2200)
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

  function setAutonomy(level: Autonomy): void {
    if (!settings.config) return
    void savePatch({
      autonomy: {
        ...settings.config.autonomy,
        default_level: level,
      },
    })
  }

  function go(hash: string): void {
    window.location.hash = hash
  }
</script>

<MeridianPage
  kicker="Desk · instruments"
  title="Settings"
  lead="Lighting, spend, and the defaults Condura uses when you ask — instruments, not a form dump."
>
  <div class="desk md-stagger">
    <p class="contract" class:hot={note === 'Saved'} class:off={offline || note.includes('offline')}>
      <span class="live-dot" aria-hidden="true"></span>
      {liveNote}.
    </p>

    <ol class="pipe" aria-label="Instrument map">
      <li><span class="n">01</span><span class="t">Lighting</span></li>
      <li><span class="n">02</span><span class="t">Autonomy</span></li>
      <li><span class="n">03</span><span class="t">Spend</span></li>
      <li><span class="n">04</span><span class="t">Keys</span></li>
    </ol>

    <!-- Appearance -->
    <section class="plate lighting" data-mode={theme}>
      <header class="plate-head">
        <p class="cite">01 · appearance</p>
        <h2>Desk lighting</h2>
        <p class="hint">Shift+T toggles anytime. The mist follows — no daemon required.</p>
      </header>
      <div class="seg" role="group" aria-label="Theme">
        <button type="button" class:on={theme === 'light'} onclick={() => setTheme('light')}>
          Light
        </button>
        <button type="button" class:on={theme === 'dark'} onclick={() => setTheme('dark')}>
          Dark
        </button>
      </div>
      <div class="swatch" aria-hidden="true">
        <span class="sw mist"></span>
        <span class="sw stage"></span>
        <span class="sw cobalt"></span>
        <span class="sw live"></span>
      </div>
    </section>

    <!-- Gate defaults -->
    <section class="plate gate">
      <header class="plate-head">
        <p class="cite">02–03 · gate defaults</p>
        <h2>How Condura asks</h2>
        <p class="hint">Autonomy never bypasses the Gatekeeper. It only changes how often you are asked.</p>
      </header>

      {#if !settings.config}
        <p class="muted">Config unavailable offline. Connect the daemon to edit defaults.</p>
      {:else}
        <div class="autonomy" role="radiogroup" aria-label="Default autonomy">
          {#each AUTONOMY as a (a.id)}
            <button
              type="button"
              class="auto-card"
              class:on={autonomy === a.id}
              data-tone={a.tone}
              role="radio"
              aria-checked={autonomy === a.id}
              disabled={saving}
              onclick={() => setAutonomy(a.id)}
            >
              <span class="auto-dot" aria-hidden="true"></span>
              <strong>{a.title}</strong>
              <span>{a.body}</span>
            </button>
          {/each}
        </div>

        <div class="spend">
          <div class="spend-copy">
            <label for="spend">Daily spend cap</label>
            <p class="hint tight">Hard ceiling in USD. Zero means no spend allowed.</p>
          </div>
          <div class="spend-ctrl">
            <div class="gauge" aria-hidden="true">
              <div class="gauge-fill" style={`width: ${spendPct}%`}></div>
            </div>
            <div class="spend-row">
              <span class="currency">$</span>
              <input
                id="spend"
                type="number"
                min="0"
                step="1"
                value={spend}
                disabled={saving}
                onchange={(e) =>
                  void savePatch({
                    security: {
                      ...settings.config!.security,
                      spend_limit_usd_per_day: Number((e.currentTarget as HTMLInputElement).value),
                    },
                  })}
              />
              <span class="unit">/ day</span>
            </div>
          </div>
        </div>

        {#if note && note !== 'Saved'}
          <p class="note" class:saving class:warn={note.includes('offline') || note.includes('Error')}>
            {saving ? 'Saving…' : note}
          </p>
        {/if}
      {/if}
    </section>

    <!-- Shortcuts -->
    <section class="plate keys">
      <header class="plate-head">
        <p class="cite">04 · atlas · keys</p>
        <h2>Always within reach</h2>
        <p class="hint">The same map About teaches — wired into the shell.</p>
      </header>
      <ul class="key-list">
        {#each KEYS as k (k.label)}
          <li>
            <kbd>{k.keys === 'K' ? `${modLabel}K` : k.keys}</kbd>
            <span>{k.label}</span>
          </li>
        {/each}
      </ul>
    </section>

    <div class="doors">
      <button type="button" class="door" onclick={() => go('#/account')}>
        <span class="door-k">passport</span>
        <strong>Account</strong>
        <span>Optional cloud doors — Hub publish and donate</span>
      </button>
      <button type="button" class="door" onclick={() => go('#/about')}>
        <span class="door-k">colophon</span>
        <strong>About</strong>
        <span>Seven stations and the Gatekeeper contract</span>
      </button>
      <button type="button" class="door" onclick={() => go('#/audit')}>
        <span class="door-k">ledger</span>
        <strong>Audit</strong>
        <span>Read what Condura did — and refused</span>
      </button>
    </div>
  </div>
</MeridianPage>

<style>
  .desk {
    display: grid;
    gap: 16px;
  }
  .contract {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    margin: 0;
    padding: 12px 14px;
    border-radius: 14px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 80%, transparent);
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .contract.hot {
    border-color: color-mix(in oklab, var(--md-live) 28%, transparent);
    background: color-mix(in oklab, var(--md-live) 6%, var(--md-surface));
  }
  .contract.off {
    border-color: color-mix(in oklab, var(--md-halt) 22%, var(--md-line));
  }
  .live-dot {
    width: 8px;
    height: 8px;
    margin-top: 5px;
    flex: none;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .contract.hot .live-dot {
    background: var(--md-live);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-live) 16%, transparent);
  }
  .pipe {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .pipe li {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 7px 11px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
  }
  .pipe .n {
    font-family: var(--md-font-mono);
    font-size: 10px;
    color: var(--md-cobalt);
  }
  .pipe .t {
    font-size: 12px;
    font-weight: 700;
    color: var(--md-ink-soft);
  }
  .plate {
    border-radius: 22px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    padding: 22px 24px 24px;
    box-shadow: inset 0 1px 0 color-mix(in oklab, var(--md-surface) 55%, transparent);
  }
  .plate-head {
    margin-bottom: 18px;
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 8px;
  }
  h2 {
    font-family: var(--md-font-display);
    font-size: 22px;
    letter-spacing: -0.04em;
    margin: 0 0 6px;
  }
  .hint {
    margin: 0;
    font-size: 13px;
    color: var(--md-ink-faint);
    line-height: 1.45;
    max-width: 48ch;
  }
  .hint.tight {
    margin-top: 4px;
  }

  .seg {
    display: inline-flex;
    padding: 4px;
    border-radius: 999px;
    background: var(--md-stage);
    border: 1px solid var(--md-line);
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
      box-shadow 180ms var(--md-ease);
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
  .swatch {
    display: flex;
    gap: 8px;
    margin-top: 16px;
  }
  .sw {
    width: 28px;
    height: 28px;
    border-radius: 10px;
    border: 1px solid var(--md-line);
  }
  .sw.mist { background: var(--md-mist); }
  .sw.stage { background: var(--md-stage); }
  .sw.cobalt { background: var(--md-cobalt); }
  .sw.live { background: var(--md-live); }

  .autonomy {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
    margin-bottom: 22px;
  }
  .auto-card {
    text-align: left;
    padding: 14px 14px 16px;
    border-radius: 16px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-stage);
    cursor: pointer;
    display: grid;
    gap: 6px;
    color: inherit;
    transition:
      border-color 180ms var(--md-ease),
      transform 180ms var(--md-spring),
      box-shadow 180ms var(--md-ease),
      background 180ms var(--md-ease);
  }
  .auto-card:disabled {
    opacity: 0.7;
    cursor: wait;
  }
  .auto-card strong {
    font-family: var(--md-font-display);
    font-size: 15px;
    letter-spacing: -0.03em;
  }
  .auto-card span:last-child {
    font-size: 12px;
    line-height: 1.4;
    color: var(--md-ink-mute);
  }
  .auto-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--md-ink-faint);
    margin-bottom: 4px;
  }
  .auto-card[data-tone='cobalt'].on {
    border-color: color-mix(in oklab, var(--md-cobalt) 55%, transparent);
    background: color-mix(in oklab, var(--md-cobalt) 10%, var(--md-surface));
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-cobalt) 14%, transparent);
  }
  .auto-card[data-tone='cobalt'].on .auto-dot { background: var(--md-cobalt); }
  .auto-card[data-tone='live'].on {
    border-color: color-mix(in oklab, var(--md-live) 55%, transparent);
    background: color-mix(in oklab, var(--md-live) 10%, var(--md-surface));
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-live) 14%, transparent);
  }
  .auto-card[data-tone='live'].on .auto-dot { background: var(--md-live); }
  .auto-card[data-tone='caution'].on {
    border-color: color-mix(in oklab, var(--md-halt) 40%, transparent);
    background: color-mix(in oklab, var(--md-halt) 8%, var(--md-surface));
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-halt) 12%, transparent);
  }
  .auto-card[data-tone='caution'].on .auto-dot { background: var(--md-halt); }
  .auto-card:hover:not(.on):not(:disabled) {
    border-color: var(--md-cobalt);
    transform: translateY(-1px);
  }
  .auto-card:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }

  .spend {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 18px;
    align-items: end;
    padding-top: 4px;
    border-top: 1px solid var(--md-line);
  }
  .spend label {
    font-family: var(--md-font-display);
    font-size: 15px;
    letter-spacing: -0.02em;
    font-weight: 700;
  }
  .gauge {
    height: 6px;
    border-radius: 999px;
    background: var(--md-stage);
    border: 1px solid var(--md-line);
    overflow: hidden;
    margin-bottom: 10px;
  }
  .gauge-fill {
    height: 100%;
    border-radius: inherit;
    background: linear-gradient(90deg, var(--md-live), var(--md-cobalt));
    transition: width 280ms var(--md-ease);
  }
  .spend-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .currency, .unit {
    font-family: var(--md-font-mono);
    font-size: 12px;
    color: var(--md-ink-faint);
  }
  .spend-row input {
    width: 96px;
    padding: 10px 12px;
    border-radius: 12px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    font-family: var(--md-font-mono);
    font-size: 16px;
    font-weight: 600;
  }
  .spend-row input:focus-visible {
    outline: none;
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
  }
  .spend-row input:disabled {
    opacity: 0.6;
  }

  .key-list {
    display: grid;
    gap: 10px;
    margin: 0;
    padding: 0;
  }
  .key-list li {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 10px 0;
    border-bottom: 1px solid var(--md-line);
  }
  .key-list li:last-child { border-bottom: 0; }
  .key-list kbd {
    font-family: var(--md-font-mono);
    font-size: 11px;
    min-width: 88px;
    text-align: center;
    padding: 6px 10px;
    border-radius: 8px;
    background: var(--md-stage);
    border: 1px solid var(--md-line-strong);
    color: var(--md-ink-soft);
  }
  .key-list span {
    font-size: 14px;
    color: var(--md-ink-mute);
  }

  .doors {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
  }
  .door {
    appearance: none;
    text-align: left;
    border: 1px solid var(--md-line-strong);
    background: var(--md-stage);
    border-radius: 18px;
    padding: 14px 14px 16px;
    cursor: pointer;
    display: grid;
    gap: 6px;
    color: inherit;
    transition:
      border-color 180ms var(--md-ease),
      transform 180ms var(--md-spring),
      box-shadow 180ms var(--md-ease);
  }
  .door-k {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .door strong {
    font-family: var(--md-font-display);
    font-size: 16px;
    letter-spacing: -0.03em;
  }
  .door > span:not(.door-k) {
    font-size: 12px;
    line-height: 1.4;
    color: var(--md-ink-mute);
  }
  .door:hover {
    border-color: var(--md-cobalt);
    transform: translateY(-2px);
    box-shadow: var(--md-shadow);
  }
  .door:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
    border-color: var(--md-cobalt);
  }

  .muted {
    color: var(--md-ink-mute);
    font-size: 14px;
    margin: 0;
    line-height: 1.5;
  }
  .note {
    margin: 14px 0 0;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--md-live);
  }
  .note.saving { color: var(--md-ink-faint); }
  .note.warn { color: var(--md-halt); }

  @media (max-width: 720px) {
    .autonomy,
    .doors { grid-template-columns: 1fr; }
    .spend { grid-template-columns: 1fr; gap: 14px; }
  }
  @media (max-width: 420px) {
    .plate { padding: 18px 16px 20px; }
    .seg { width: 100%; }
    .seg button { flex: 1; }
  }
</style>
