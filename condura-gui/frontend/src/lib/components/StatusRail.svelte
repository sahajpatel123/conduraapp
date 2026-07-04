<script lang="ts">
  import { daemon } from '../stores/daemon.svelte'
  import { halt } from '../stores/halt.svelte'
  import { t } from '../i18n'
</script>

<footer class="status-rail">
  <div class="status-rail-left">
    <span class="rail-item">
      <span class="rail-dot" class:rail-dot-on={daemon.connected}></span>
      <span class="rail-label">
        {daemon.connected ? t('app.status.connected') : t('app.status.disconnected')}
      </span>
    </span>

    {#if halt.state.halted}
      <span class="rail-item rail-halt">
        <svg viewBox="0 0 24 24" width="11" height="11" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <path d="M12 9v4M12 17h.01M10.3 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
        </svg>
        Halted
      </span>
    {/if}
  </div>

  <div class="status-rail-right">
    <span class="rail-item rail-version">v0.1.0</span>
  </div>
</footer>

<style>
  .status-rail {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 var(--space-3);
    height: var(--status-rail-height);
    background: var(--surface-1);
    border-top: 1px solid var(--border);
    flex-shrink: 0;
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-faint);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
  }

  .status-rail-left,
  .status-rail-right {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .rail-item {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .rail-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--text-faint);
    flex-shrink: 0;
  }
  .rail-dot-on {
    background: var(--success);
    box-shadow: 0 0 8px var(--success-glow);
    animation: breathe 2.4s var(--ease-in-out-quart) infinite;
  }

  .rail-label { color: var(--text-muted); }
  .rail-halt { color: var(--error); font-weight: var(--weight-semibold); }
  .rail-version { color: var(--text-faint); }
</style>