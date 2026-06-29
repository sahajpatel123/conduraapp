<script lang="ts">
  // Channels — connect messaging channels (Telegram, Slack, Discord,
  // iMessage, WhatsApp). Polls channels.status every 10s. Telegram
  // opens a Sheet with a BotFather token field validated as
  // digits:secret.
  import { ipc } from '../ipc/client'
  import { onMount, onDestroy } from 'svelte'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import { Button, Card, Input, Sheet, Badge } from '../components/ui'

  interface ChannelStatus {
    name: string
    connected: boolean
    chat_id?: string
    error?: string
  }

  interface ChannelMeta {
    name: string
    label: string
    comingSoon?: boolean
  }

  // The list of channels the user can connect to. Telegram is the
  // first one with a real backend; the others are visible as
  // "Coming soon" cards so the surface is honest about scope.
  const catalog: ChannelMeta[] = [
    { name: 'telegram', label: 'Telegram' },
    { name: 'slack', label: 'Slack', comingSoon: true },
    { name: 'discord', label: 'Discord', comingSoon: true },
    { name: 'imessage', label: 'iMessage', comingSoon: true },
    { name: 'whatsapp', label: 'WhatsApp', comingSoon: true },
  ]

  let channels = $state<ChannelStatus[]>([])
  let loading = $state(false)
  let error = $state<string | null>(null)
  let token = $state('')
  let connecting = $state(false)
  let sheetOpen = $state(false)
  let pollTimer: ReturnType<typeof setInterval> | null = null
  let confirmOpen = $state(false)
  let confirmAction = $state<(() => void) | null>(null)

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

  // digits:secret, per Telegram bot token format.
  const tokenValid = $derived(/^\d+:[A-Za-z0-9_-]{20,}$/.test(token.trim()))

  async function connectTelegram(): Promise<void> {
    if (!tokenValid) return
    connecting = true
    error = null
    try {
      await ipc.call('channels.connect', { channel: 'telegram', token: token.trim() })
      token = ''
      sheetOpen = false
      await refresh()
    } catch (e) {
      error = String(e)
    } finally {
      connecting = false
    }
  }

  function disconnect(name: string): void {
    confirmAction = async () => {
      error = null
      try {
        await ipc.call('channels.disconnect', { channel: name })
        await refresh()
      } catch (e) {
        error = String(e)
      }
    }
    confirmOpen = true
  }

  function statusFor(name: string): ChannelStatus | null {
    return channels.find((c) => c.name === name) ?? null
  }

  function openConnectSheet(name: string): void {
    if (name !== 'telegram') return
    token = ''
    error = null
    sheetOpen = true
  }

  onMount(() => {
    void refresh()
    pollTimer = setInterval(() => void refresh(), 10000)
  })
  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer)
  })
</script>

<div class="channels-page">
  <header class="page-header">
    <h2 class="display-h2">Channels</h2>
    <p class="lede">
      Talk to Condura from your phone. Connect a messaging channel and the agent will
      answer messages the same way it answers anything else.
    </p>
  </header>

  {#if error}
    <p class="error" role="alert">{error}</p>
  {/if}

  <section class="channels-list">
    {#each catalog as ch, i (ch.name)}
      {@const s = statusFor(ch.name)}
      {@const tone = ch.comingSoon ? 'neutral' : s?.error ? 'error' : s?.connected ? 'success' : 'neutral'}
      {@const statusLabel = ch.comingSoon ? 'Coming soon' : s?.error ? 'Error' : s?.connected ? 'Connected' : 'Not connected'}
      <div class="row-wrap" style:--stagger-index={i}>
        <Card elevation="glass" padding="md">
          <div class="row">
            <span class="logo" data-channel={ch.name} aria-hidden="true">
              {#if ch.name === 'telegram'}
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M9.78 18.65l.28-4.23 7.68-6.92c.34-.31-.07-.46-.52-.19l-9.49 5.99-4.1-1.27c-.88-.25-.89-.86.2-1.3l16.04-6.18c.73-.33 1.43.18 1.15 1.3l-2.73 12.86c-.19.91-.74 1.13-1.5.71l-4.13-3.05-1.99 1.93c-.23.23-.42.42-.84.42z"/></svg>
              {:else if ch.name === 'slack'}
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M5 14a2 2 0 11-4 0 2 2 0 014 0zm1 0a2 2 0 114 0v6a2 2 0 11-4 0v-6zm2-9a2 2 0 11-4 0 2 2 0 014 0zm0 1a2 2 0 110 4H2a2 2 0 110-4h6zm9 2a2 2 0 114 0 2 2 0 01-4 0zm-1 0a2 2 0 11-4 0V2a2 2 0 114 0v6zm-2 9a2 2 0 114 0 2 2 0 01-4 0zm0-1a2 2 0 110-4h6a2 2 0 110 4h-6z"/></svg>
              {:else if ch.name === 'discord'}
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M20.317 4.37a19.79 19.79 0 00-4.885-1.515.074.074 0 00-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 00-5.487 0 12.64 12.64 0 00-.617-1.25.077.077 0 00-.079-.037A19.736 19.736 0 003.677 4.37a.07.07 0 00-.032.027C.533 9.046-.32 13.58.099 18.058a.082.082 0 00.031.057 19.9 19.9 0 005.993 3.03.078.078 0 00.084-.028c.462-.63.874-1.295 1.226-1.994a.076.076 0 00-.041-.106 13.107 13.107 0 01-1.872-.892.077.077 0 01-.008-.128 10.2 10.2 0 00.372-.292.074.074 0 01.077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 01.078.01c.12.098.246.198.373.292a.077.077 0 01-.006.127 12.299 12.299 0 01-1.873.892.077.077 0 00-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 00.084.028 19.839 19.839 0 006.002-3.03.077.077 0 00.032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 00-.031-.03zM8.02 15.331c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z"/></svg>
              {:else if ch.name === 'imessage'}
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M12 2C6.477 2 2 6.145 2 11.243c0 2.908 1.434 5.502 3.678 7.21V22l3.453-1.896c.92.254 1.893.392 2.869.392 5.523 0 10-4.145 10-9.243C22 6.145 17.523 2 12 2zm1.6 12.357l-2.557-2.732-5 2.732 5.5-5.857 2.614 2.732 4.943-2.732-5.5 5.857z"/></svg>
              {:else if ch.name === 'whatsapp'}
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor"><path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/></svg>
              {/if}
            </span>

            <div class="row-info">
              <span class="name">{ch.label}</span>
              {#if s?.chat_id}
                <span class="chat-id mono">{s.chat_id}</span>
              {/if}
              {#if s?.error}
                <span class="err-text">{s.error}</span>
              {/if}
            </div>

            <Badge tone={tone} size="sm" dot pulse={tone === 'success'}>
              {statusLabel}
            </Badge>

            <div class="row-actions">
              {#if ch.comingSoon}
                <span class="muted-xs">In v0.2.0</span>
              {:else if s?.connected}
                <Button variant="ghost" size="sm" onclick={() => disconnect(ch.name)}>Disconnect</Button>
              {:else}
                <Button variant="primary" size="sm" onclick={() => openConnectSheet(ch.name)}>Connect</Button>
              {/if}
            </div>
          </div>
        </Card>
      </div>
    {/each}
  </section>

  <footer class="hint">
    <p>
      Polled every 10s. Externally-revoked tokens show as Error until you reconnect.
    </p>
  </footer>
</div>

<Sheet
  bind:open={sheetOpen}
  side="right"
  width="420px"
  title="Connect Telegram"
  onclose={() => (sheetOpen = false)}
>
  <div class="tg-form">
    <p class="hint-text">
      Create a bot with <a href="https://t.me/BotFather" target="_blank" rel="noreferrer">@BotFather</a>
      and paste the token below. Format: <code class="mono">digits:secret</code>.
    </p>

    <Input
      bind:value={token}
      fullWidth
      size="md"
      label="BotFather token"
      placeholder="123456789:ABCdef…"
      autocomplete="off"
      spellcheck="false"
      error={token && !tokenValid ? 'Token must look like 123456789:ABCdefGHIjklMNOpqr' : undefined}
      onkeydown={(e: KeyboardEvent) => { if (e.key === 'Enter' && tokenValid) void connectTelegram() }}
    />

    <div class="tg-actions">
      <Button variant="ghost" onclick={() => (sheetOpen = false)}>Cancel</Button>
      <Button variant="primary" onclick={connectTelegram} disabled={!tokenValid || connecting} loading={connecting}>
        {connecting ? 'Connecting…' : 'Connect'}
      </Button>
    </div>

    {#if error}
      <p class="err-text">{error}</p>
    {/if}
  </div>
</Sheet>

<ConfirmDialog
  bind:open={confirmOpen}
  title="Disconnect channel"
  description="The channel will stop receiving messages. You can reconnect any time."
  tone="danger"
  confirmLabel="Disconnect"
  onconfirm={() => confirmAction?.()}
/>

<style>
  .channels-page {
    padding: var(--space-6) var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width);
    margin: 0 auto;
  }

  .page-header {
    margin-bottom: var(--space-6);
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .display-h2 {
    font-family: var(--font-display);
    font-size: var(--size-2xl);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    color: var(--text);
    margin: 0 0 var(--space-2) 0;
    line-height: var(--leading-tight);
  }
  .lede {
    font-size: var(--size-md);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    max-width: 640px;
    margin: 0;
  }

  /* ── Channel rows ─────────────────────────────── */
  .channels-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .row-wrap {
    animation: stagger-in var(--transition-base) var(--ease-out-expo) both;
    animation-delay: calc(var(--stagger-index, 0) * 60ms);
  }

  .row {
    display: flex;
    align-items: center;
    gap: var(--space-4);
  }
  .logo {
    width: 40px;
    height: 40px;
    border-radius: var(--radius-md);
    background: var(--surface-3);
    border: 1px solid var(--border);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
    flex-shrink: 0;
  }
  .logo[data-channel='telegram'] { color: #229ED9; }
  .logo[data-channel='slack']    { color: #ECB22E; }
  .logo[data-channel='discord']  { color: #5865F2; }
  .logo[data-channel='imessage'] { color: var(--success); }
  .logo[data-channel='whatsapp'] { color: #25D366; }

  .row-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .name {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
  }
  .chat-id {
    font-size: var(--size-xs);
    color: var(--text-faint);
  }
  .err-text {
    font-size: var(--size-xs);
    color: var(--error);
  }

  .row-actions {
    flex-shrink: 0;
  }
  .muted-xs {
    font-size: var(--size-xs);
    color: var(--text-faint);
  }

  .hint {
    margin-top: var(--space-5);
    padding-top: var(--space-4);
    border-top: 1px solid var(--border);
  }
  .hint p {
    font-size: var(--size-xs);
    color: var(--text-muted);
    margin: 0;
  }

  .error {
    color: var(--error);
    font-size: var(--size-sm);
    padding: var(--space-2) var(--space-3);
    background: var(--error-soft);
    border: 1px solid var(--border-danger);
    border-radius: var(--radius-md);
  }

  /* ── Telegram sheet form ──────────────────────── */
  .tg-form {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .hint-text {
    font-size: var(--size-sm);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    margin: 0;
  }
  .hint-text a { color: var(--accent); }
  .hint-text code {
    font-size: var(--size-xs);
    background: var(--surface-3);
    border: 1px solid var(--border);
    padding: 1px 5px;
    border-radius: var(--radius-xs);
  }
  .tg-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-top: var(--space-2);
  }
</style>
