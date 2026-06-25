<script lang="ts">
  import { onMount } from 'svelte'
  import { replay } from '../stores/replay.svelte'
  import { t } from '../i18n'

  onMount(() => {
    void replay.refresh()
    void replay.verifyIntegrity()
  })

  function outcomeClass(outcome: string): string {
    if (outcome === 'allowed') return 'outcome-allowed'
    if (outcome === 'denied') return 'outcome-denied'
    if (outcome === 'errored') return 'outcome-errored'
    return 'outcome-unknown'
  }

  async function exportVideo(): Promise<void> {
    try {
      const path = await replay.exportMP4()
      alert(t('replay.exported_alert', path))
    } catch {
      alert(replay.lastError || t('replay.export_failed'))
    }
  }
</script>

<div class="replay-page">
  <header class="page-header">
    <h2>{t('replay.title')}</h2>
    <p class="muted">{t('replay.intro')}</p>
    <div class="header-actions">
      <button class="btn btn-ghost" onclick={() => replay.refresh()} disabled={replay.loading}>{t('replay.refresh')}</button>
      <button class="btn btn-ghost" onclick={() => replay.verifyIntegrity()}>{t('replay.verify')}</button>
      <button class="btn btn-primary" onclick={exportVideo} disabled={replay.exporting || replay.frames.length === 0}>
        {replay.exporting ? t('replay.exporting') : t('replay.export')}
      </button>
    </div>
    {#if replay.integrity}
      <p class="integrity" class:valid={replay.integrity.valid}>
        {t('replay.integrity', replay.integrity.valid ? t('replay.integrity_valid') : t('replay.integrity_invalid'), replay.integrity.rows_checked)}
      </p>
    {/if}
  </header>

  {#if replay.loading}
    <p class="muted">{t('replay.loading')}</p>
  {:else if replay.frames.length === 0}
    <p class="muted">{t('replay.empty')}</p>
  {:else}
    <div class="scrubber">
      <input
        type="range"
        min="0"
        max={replay.frames.length - 1}
        value={replay.selectedIndex}
        oninput={(e) => replay.selectIndex(parseInt((e.target as HTMLInputElement).value, 10))}
        class="slider"
        style={`--fill: ${(replay.selectedIndex / Math.max(replay.frames.length - 1, 1)) * 100}%`}
        aria-label={t('replay.scrubber_aria')}
      />
      <span class="scrub-label">{replay.selectedIndex + 1} / {replay.frames.length}</span>
    </div>

    {#if replay.selected}
      {#key replay.selectedIndex}
        <div class="frame-detail glass-card">
        <div class="meta">
          <span class="ts">{new Date(replay.selected.timestamp).toLocaleString()}</span>
          <span class="badge {outcomeClass(replay.selected.outcome)}">{replay.selected.outcome}</span>
          <span class="action">{replay.selected.action}</span>
          <span class="app">{replay.selected.app}</span>
        </div>
        <p class="message">{replay.selected.message || replay.selected.outcome_reason || '—'}</p>
        <div class="shots">
          {#if replay.selected.before_screenshot}
            <figure>
              <figcaption>{t('replay.before')}</figcaption>
              <img src="data:{replay.selected.before_screenshot_mime || 'image/png'};base64,{replay.selected.before_screenshot}" alt={t('replay.before_alt')} />
            </figure>
          {/if}
          {#if replay.selected.after_screenshot}
            <figure>
              <figcaption>{t('replay.after')}</figcaption>
              <img src="data:{replay.selected.after_screenshot_mime || 'image/png'};base64,{replay.selected.after_screenshot}" alt={t('replay.after_alt')} />
            </figure>
          {/if}
          {#if !replay.selected.before_screenshot && !replay.selected.after_screenshot}
            <p class="muted">{t('replay.no_screenshots')}</p>
          {/if}
        </div>
      </div>
      {/key}

      <div class="frame-list">
        {#each replay.frames as f, i (f.id)}
          <button
            class="frame-row"
            class:active={i === replay.selectedIndex}
            onclick={() => replay.selectIndex(i)}
          >
            <span class="ts">{new Date(f.timestamp).toLocaleTimeString()}</span>
            <span class="action">{f.action}</span>
            <span class="badge {outcomeClass(f.outcome)}">{f.outcome}</span>
          </button>
        {/each}
      </div>
    {/if}
  {/if}
</div>

<style>
  .replay-page {
    padding: var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide);
    margin: 0 auto;
  }
  .replay-page .page-header {
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }

  /* ── Header actions ──────────────────────────────────── */
  .header-actions {
    display: flex;
    gap: var(--space-2);
    margin: var(--space-3) 0;
    flex-wrap: wrap;
  }
  .integrity { font-size: var(--size-sm); margin-top: var(--space-2); }
  .integrity.valid { color: var(--color-success); }
  .integrity:not(.valid) { color: var(--color-error); }

  /* ── Premium scrubber — glass track with accent fill ── */
  .scrubber {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin: var(--space-4) 0;
    padding: var(--space-3) var(--space-4);
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-pill);
  }
  .slider {
    flex: 1;
    -webkit-appearance: none;
    appearance: none;
    height: 6px;
    border-radius: var(--radius-pill);
    background: linear-gradient(
      to right,
      var(--color-accent) 0%,
      var(--color-accent-secondary) var(--fill, 0%),
      var(--color-bg-elevated) var(--fill, 0%),
      var(--color-bg-elevated) 100%
    );
    outline: none;
    cursor: pointer;
  }
  .slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--color-accent);
    cursor: pointer;
    box-shadow: var(--shadow-glow-accent), 0 2px 4px rgba(20, 17, 11, 0.18);
    border: 2px solid var(--color-border-strong);
    transition: transform var(--transition-base), box-shadow var(--transition-base);
  }
  .slider::-webkit-slider-thumb:hover {
    transform: scale(1.25);
    box-shadow: var(--shadow-glow-strong), var(--shadow-glow-accent);
  }
  .slider::-webkit-slider-thumb:active {
    transform: scale(1.15);
    box-shadow: var(--shadow-glow-strong), var(--shadow-glow-accent);
  }
  .slider::-moz-range-thumb {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--color-accent);
    cursor: pointer;
    border: 2px solid var(--color-border-strong);
    box-shadow: var(--shadow-glow-accent), 0 2px 4px rgba(0, 0, 0, 0.3);
    transition: transform var(--transition-base), box-shadow var(--transition-base);
  }
  .slider::-moz-range-thumb:hover {
    transform: scale(1.25);
    box-shadow: var(--shadow-glow-strong), var(--shadow-glow-accent);
  }
  .scrub-label {
    font-family: var(--font-mono);
    font-size: var(--size-sm);
    color: var(--color-text-muted);
    min-width: 64px;
    text-align: right;
  }

  /* ── Frame detail card — fade-in-scale on selection change ── */
  .frame-detail {
    padding: var(--space-5);
    margin-bottom: var(--space-4);
    animation: fade-in-scale var(--transition-slow) var(--ease-out-expo) both;
  }
  .meta {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: center;
    margin-bottom: var(--space-3);
    font-size: var(--size-sm);
  }
  .meta .ts { font-family: var(--font-mono); font-size: var(--size-xs); color: var(--color-text-muted); }
  .meta .action { font-weight: var(--weight-semibold); font-family: var(--font-mono); }
  .meta .app { color: var(--color-text-muted); }
  .message {
    margin-bottom: var(--space-4);
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    color: var(--color-text);
  }

  /* ── Screenshots ─────────────────────────────────────── */
  .shots {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: var(--space-4);
  }
  figure { margin: 0; }
  figcaption {
    font-size: var(--size-xs);
    color: var(--color-text-muted);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wide);
    margin-bottom: var(--space-2);
  }
  .shots img {
    width: 100%;
    border-radius: var(--radius-md);
    border: 1px solid var(--glass-border);
    transition: border-color var(--transition-base);
  }
  .shots img:hover {
    border-color: var(--glass-border-hover);
  }

  /* ── Frame list — clean vertical timeline ────────────── */
  .frame-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    position: relative;
  }
  .frame-list::before {
    content: '';
    position: absolute;
    left: 14px;
    top: var(--space-2);
    bottom: var(--space-2);
    width: 1px;
    background: linear-gradient(180deg, transparent, var(--color-accent), transparent);
    opacity: 0.4;
  }
  .frame-row {
    display: grid;
    grid-template-columns: 28px 100px 1fr auto;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    text-align: left;
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    color: var(--color-text);
    cursor: pointer;
    font-size: var(--size-sm);
    align-items: center;
    transition: all var(--transition-base);
    position: relative;
    animation: stagger-in var(--transition-base) var(--ease-out-expo) both;
  }
  .frame-row:nth-of-type(1) { animation-delay: 40ms; }
  .frame-row:nth-of-type(2) { animation-delay: 80ms; }
  .frame-row:nth-of-type(3) { animation-delay: 120ms; }
  .frame-row:nth-of-type(4) { animation-delay: 160ms; }
  .frame-row:nth-of-type(5) { animation-delay: 200ms; }
  .frame-row:nth-of-type(6) { animation-delay: 240ms; }
  .frame-row:nth-of-type(7) { animation-delay: 280ms; }
  .frame-row:nth-of-type(8) { animation-delay: 320ms; }
  .frame-row:nth-of-type(n + 9) { animation-delay: 360ms; }
  .frame-row::before {
    content: '';
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--color-text-faint);
    transition: all var(--transition-base);
    justify-self: center;
  }
  .frame-row:hover {
    background: var(--color-bg-hover);
  }
  .frame-row.active {
    background: var(--color-accent-soft);
    border-color: var(--color-border-accent);
    box-shadow: var(--shadow-glow);
  }
  .frame-row.active::before {
    background: var(--color-accent);
    box-shadow: var(--shadow-glow-strong);
  }
  .frame-row .ts {
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    color: var(--color-text-muted);
  }
  .frame-row .action {
    font-family: var(--font-mono);
    font-size: var(--size-xs);
  }

  /* ── Outcome badges (component-specific colors, glow-enhanced) ── */
  .badge.outcome-allowed {
    color: var(--color-success);
    border-color: rgba(16, 185, 129, 0.4);
    background: var(--color-success-soft);
    box-shadow: 0 0 14px var(--color-success-glow);
  }
  .badge.outcome-denied {
    color: var(--color-error);
    border-color: rgba(239, 68, 68, 0.4);
    background: var(--color-error-soft);
    box-shadow: 0 0 14px var(--color-error-glow);
  }
  .badge.outcome-errored {
    color: var(--color-warn);
    border-color: rgba(245, 158, 11, 0.4);
    background: var(--color-warn-soft);
    box-shadow: 0 0 14px var(--color-warn-glow);
  }
  .badge.outcome-unknown {
    color: var(--color-text-muted);
    border-color: var(--glass-border);
    background: var(--color-bg-elevated);
  }
</style>
