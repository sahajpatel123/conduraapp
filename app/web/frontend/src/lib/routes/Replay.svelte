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
  <header>
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
        aria-label={t('replay.scrubber_aria')}
      />
      <span class="scrub-label">{replay.selectedIndex + 1} / {replay.frames.length}</span>
    </div>

    {#if replay.selected}
      <div class="frame-detail card">
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
    max-width: 1100px;
    margin: 0 auto;
  }
  header h2 {
    font-size: var(--size-2xl);
    font-weight: 600;
    margin-bottom: var(--space-2);
    background: var(--color-accent-gradient);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .muted { color: var(--color-text-muted); font-size: var(--size-sm); }
  .header-actions {
    display: flex;
    gap: var(--space-2);
    margin: var(--space-3) 0;
    flex-wrap: wrap;
  }
  .integrity { font-size: var(--size-sm); margin-top: var(--space-2); }
  .integrity.valid { color: var(--color-success); }
  .integrity:not(.valid) { color: #f87171; }
  .scrubber {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin: var(--space-4) 0;
  }
  .slider { flex: 1; }
  .scrub-label { font-family: var(--font-mono); font-size: var(--size-sm); color: var(--color-text-muted); }
  .card {
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-xl);
    padding: var(--space-5);
    margin-bottom: var(--space-4);
  }
  .meta {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: center;
    margin-bottom: var(--space-3);
    font-size: var(--size-sm);
  }
  .action { font-weight: 600; font-family: var(--font-mono); }
  .app { color: var(--color-text-muted); }
  .message { margin-bottom: var(--space-4); font-size: var(--size-md); }
  .shots {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: var(--space-4);
  }
  figure { margin: 0; }
  figcaption { font-size: var(--size-xs); color: var(--color-text-muted); margin-bottom: var(--space-2); }
  .shots img {
    width: 100%;
    border-radius: var(--radius-md);
    border: 1px solid var(--glass-border);
  }
  .frame-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .frame-row {
    display: grid;
    grid-template-columns: 100px 1fr auto;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    text-align: left;
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    color: var(--color-text);
    cursor: pointer;
    font-size: var(--size-sm);
  }
  .frame-row:hover, .frame-row.active {
    background: var(--color-accent-soft);
    border-color: var(--glass-border);
  }
  .badge {
    padding: 2px 8px;
    border-radius: var(--radius-pill);
    font-size: var(--size-xs);
    text-transform: uppercase;
  }
  .outcome-allowed { background: rgba(74, 222, 128, 0.15); color: var(--color-success); }
  .outcome-denied { background: rgba(248, 113, 113, 0.15); color: #f87171; }
  .outcome-errored { background: rgba(251, 191, 36, 0.15); color: #fbbf24; }
  .outcome-unknown { background: rgba(148, 163, 184, 0.15); color: var(--color-text-muted); }
  .btn {
    padding: 8px 16px;
    border-radius: var(--radius-md);
    font-size: var(--size-md);
    cursor: pointer;
    border: none;
  }
  .btn-primary { background: var(--color-accent-gradient); color: white; }
  .btn-ghost { background: transparent; border: 1px solid var(--glass-border); color: var(--color-text-muted); }
</style>
