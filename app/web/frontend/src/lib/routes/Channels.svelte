<script lang="ts">
  // Channels route (Phase 14C — "Reach"). Connect messaging
  // channels (Telegram first) so you can talk to Condura from
  // your phone. Backend: the daemon's channels.* RPCs (reach
  // subsystem). We use the generic ipc.call so this component
  // owns its own contract.
  import { ipc } from '../ipc/client'
  import { onMount, onDestroy } from 'svelte'
  import { t } from '../i18n'

  // Mirrors reach.ChannelStatus on the Go side.
  interface ChannelStatus {
    name: string
    connected: boolean
    chat_id?: string
    error?: string
  }

  let channels = $state<ChannelStatus[]>([])
  let loading = $state(false)
  let error = $state<string | null>(null)
  let token = $state('')
  let connecting = $state(false)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  async function refresh(): Promise<void> {
    loading = true
    error = null
    try {
      channels = (await ipc.call<ChannelStatus[]>('channels.list', {})) ?? []
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  }

  const tokenValid = $derived(/^\d+:[A-Za-z0-9_-]{20,}$/.test(token.trim()))

  async function connectTelegram(): Promise<void> {
    if (!tokenValid) return
    connecting = true
    error = null
    try {
      await ipc.call('channels.connect', { channel: 'telegram', token: token.trim() })
      token = ''
      await refresh()
    } catch (e) {
      error = String(e)
    } finally {
      connecting = false
    }
  }

  async function disconnect(name: string): Promise<void> {
    if (!confirm(t('channels.disconnect_confirm', name))) return
    error = null
    try {
      await ipc.call('channels.disconnect', { channel: name })
      await refresh()
    } catch (e) {
      error = String(e)
    }
  }

  function prettyName(name: string): string {
    return name.charAt(0).toUpperCase() + name.slice(1)
  }

  onMount(() => {
    void refresh()
    // Light polling so externally-changed status (e.g. a token
    // revoked on Telegram's side) reflects without a manual refresh.
    pollTimer = setInterval(() => void refresh(), 10000)
  })
  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer)
  })
</script>

<div class="channels-page">
  <header class="page-header">
    <h2>{t('channels.title')}</h2>
    <p class="muted">{t('channels.intro')}</p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <section class="glass-card card">
    <h3>{t('channels.connected', channels.length)}</h3>
    {#if loading && channels.length === 0}
      <p class="muted">{t('common.loading')}</p>
    {:else if channels.length === 0}
      <p class="muted">{t('channels.empty')}</p>
    {:else}
      <ul class="channel-list">
        {#each channels as c, i (c.name)}
          <li class="channel-row stagger-item" style="--stagger-index: {i}">
            <span class="dot" class:on={c.connected} class:err={!!c.error}></span>
            <span class="ch-name">{prettyName(c.name)}</span>
            <span class="ch-status">
              {#if c.error}{t('channels.status.error')}{:else if c.connected}{t('channels.status.connected')}{:else}{t('channels.status.disconnected')}{/if}
            </span>
            {#if c.chat_id}<span class="ch-chat mono">{c.chat_id}</span>{/if}
            <span class="spacer"></span>
            <button class="btn btn-ghost btn-xs" onclick={() => disconnect(c.name)}>{t('channels.disconnect')}</button>
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  <section class="glass-card card">
    <h3>{t('channels.telegram_title')}</h3>
    <p class="muted">
      {t('channels.telegram_intro_html')}
    </p>
    <div class="connect-row">
      <input
        type="text"
        class="input token-input"
        bind:value={token}
        placeholder="123456789:ABCdef…"
        autocomplete="off"
        spellcheck="false"
        onkeydown={(e) => { if (e.key === 'Enter') connectTelegram() }}
      />
      <button class="btn btn-primary" onclick={connectTelegram} disabled={!tokenValid || connecting}>
        {connecting ? t('channels.connecting') : t('channels.connect')}
      </button>
    </div>
    {#if token && !tokenValid}
      <p class="hint">{t('channels.invalid_token_html')}</p>
    {/if}
  </section>
</div>

<style>
  .channels-page {
    padding: var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width);
    margin: 0 auto;
  }
  .page-header {
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .card {
    padding: var(--space-5);
    margin-top: var(--space-5);
  }
  .card h3 {
    font-size: var(--size-lg);
    font-weight: var(--weight-semibold);
    margin-bottom: var(--space-3);
  }
  .channel-list {
    list-style: none;
    padding: 0;
    margin: var(--space-2) 0 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .channel-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3);
    background: rgba(0, 0, 0, 0.18);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    font-size: var(--size-sm);
    transition: border-color var(--transition-base);
  }
  .channel-row:hover {
    border-color: var(--glass-border-hover);
    box-shadow: var(--shadow-glow-accent);
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-text-faint);
    flex-shrink: 0;
  }
  .dot.on {
    background: var(--color-success);
    box-shadow: 0 0 10px rgba(74, 222, 128, 0.5);
  }
  .dot.err {
    background: var(--color-error);
  }
  .ch-name { font-weight: var(--weight-semibold); }
  .ch-status { color: var(--color-text-muted); }
  .ch-chat { color: var(--color-text-faint); font-size: var(--size-xs); }
  .spacer { flex: 1; }
  .connect-row {
    display: flex;
    gap: var(--space-2);
    margin-top: var(--space-3);
  }
  .token-input {
    flex: 1;
    font-family: var(--font-mono);
    font-size: var(--size-sm);
  }
  .hint {
    color: var(--color-text-faint);
    font-size: var(--size-xs);
    margin-top: var(--space-2);
  }
  .error {
    color: var(--color-error);
    font-size: var(--size-sm);
  }
</style>
