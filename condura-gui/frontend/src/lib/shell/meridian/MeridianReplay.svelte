<script lang="ts">
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { replay } from '../../stores/replay.svelte'
  onMount(() => { void replay.refresh(); void replay.verifyIntegrity() })
</script>
<MeridianPage kicker="Timeline" title="Replay" lead="Re-walk the last day of actions. Screenshots and frames, scrubbable.">
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void replay.refresh()}>Refresh</button>
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void replay.verifyIntegrity()}>Verify</button>
  {/snippet}
  {#if replay.integrity && replay.integrity.valid === false}
    <div class="md-panel md-panel-static bad">Integrity check failed. Treat this timeline as untrusted.</div>
  {/if}
  {#if replay.loading && replay.frames.length === 0}<div class="md-empty">Loading timeline…</div>
  {:else if replay.lastError && !/IPC client not started|not connected|Failed to fetch/i.test(String(replay.lastError))}
    <div class="md-empty">{replay.lastError}</div>
  {:else if replay.frames.length === 0}
    <div class="md-empty empty">
      <p class="empty-title">No frames yet</p>
      <p class="empty-lead">
        {#if replay.lastError && /IPC client not started|not connected|Failed to fetch/i.test(String(replay.lastError))}
          Connect the daemon to load the last day of actions.
        {:else}
          Re-walk screenshots and frames from the last 24 hours once Condura has acted.
        {/if}
      </p>
    </div>
  {:else}
    <div class="scrub">
      <input
        type="range"
        min="0"
        max={Math.max(0, replay.frames.length - 1)}
        value={replay.selectedIndex}
        aria-label="Replay frame"
        oninput={(e) => replay.selectIndex(Number((e.currentTarget as HTMLInputElement).value))}
      />
      <span class="meta">{replay.selectedIndex + 1} / {replay.frames.length}</span>
    </div>
    {#if replay.selected}
      <article class="md-panel md-panel-static frame">
        <p class="meta">{replay.selected.timestamp}</p>
        <p class="body">{replay.selected.action} · {replay.selected.app}</p>
        <p class="msg">{replay.selected.message}</p>
      </article>
    {/if}
  {/if}
</MeridianPage>
<style>
  .bad {
    margin-bottom: 16px;
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 40%, transparent);
  }
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
  }
  .empty-title {
    margin: 0;
    font-family: var(--md-font-display);
    font-size: 16px;
    letter-spacing: -0.03em;
    color: var(--md-ink);
  }
  .empty-lead {
    margin: 0;
    max-width: 42ch;
    font-size: 13px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .scrub {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 14px;
    padding: 10px 14px;
    border-radius: 16px;
    background: var(--md-surface);
    border: 1px solid var(--md-line-strong);
    box-shadow: var(--md-shadow);
  }
  .scrub input {
    flex: 1;
    accent-color: var(--md-cobalt);
    height: 24px;
  }
  .scrub input:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
    border-radius: 8px;
  }
  .meta {
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-ink-faint);
    flex: none;
    font-variant-numeric: tabular-nums;
  }
  .frame {
    padding: 16px 18px;
  }
  .frame .meta {
    margin: 0;
  }
  .frame .body {
    margin: 6px 0 0;
    font-size: 15px;
    font-weight: 600;
    letter-spacing: -0.02em;
  }
  .frame .msg {
    margin: 6px 0 0;
    color: var(--md-ink-mute);
    font-size: 13px;
    line-height: 1.5;
  }
</style>
