<script lang="ts">
  // Channels — connect messaging channels (Telegram, Slack, Discord,
  // iMessage, WhatsApp). Polls channels.status every 10s. Telegram
  // opens a side panel with a BotFather token field validated as
  // digits:secret.
  import { ipc } from '../ipc/client'
  import { onMount, onDestroy } from 'svelte'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import Button from '$components/v1/Button.svelte'
  import Card from '$components/v1/Card.svelte'
  import Input from '$components/v1/Input.svelte'
  import Pill from '$components/v1/Pill.svelte'
  import Surface from '$components/v1/Surface.svelte'
  import Stack from '$components/v1/Stack.svelte'
  import Inline from '$components/v1/Inline.svelte'
  import Hairline from '$components/v1/Hairline.svelte'

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

  const tokenValid = $derived(/^\d+:[A-Za-z0-9_-]{20,}$/.test(token.trim()))
  const tokenError = $derived(
    token && !tokenValid ? 'Token must look like 123456789:ABCdefGHIjklMNOpqr' : ''
  )

  function pillVariant(
    ch: ChannelMeta,
    s: ChannelStatus | null
  ): 'success' | 'error' | 'neutral' | 'warning' {
    if (ch.comingSoon) return 'neutral'
    if (s?.error) return 'error'
    if (s?.connected) return 'success'
    return 'neutral'
  }

  function statusLabel(ch: ChannelMeta, s: ChannelStatus | null): string {
    if (ch.comingSoon) return 'Coming soon'
    if (s?.error) return 'Error'
    if (s?.connected) return 'Connected'
    return 'Not connected'
  }

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
    <h2 class="page-title">Channels</h2>
    <p class="lede">
      Talk to Condura from your phone. Connect a messaging channel and the agent will
      answer messages the same way it answers anything else.
    </p>
  </header>

  {#if error && !sheetOpen}
    <p class="error-banner" role="alert">{error}</p>
  {/if}

  <Stack gap="3" as="section" class="channels-list">
    {#each catalog as ch, i (ch.name)}
      {@const s = statusFor(ch.name)}
      <div class="row-wrap" style:--stagger-index={i}>
        <Card variant="raised" padding="4">
          {#snippet children()}
            <Inline gap="4" align="center" wrap={false} class="channel-row">
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
                  <span class="chat-id">{s.chat_id}</span>
                {/if}
                {#if s?.error}
                  <span class="err-text">{s.error}</span>
                {/if}
              </div>

              <Pill variant={pillVariant(ch, s)} size="sm" label={statusLabel(ch, s)} />

              <div class="row-actions">
                {#if ch.comingSoon}
                  <span class="muted-xs">In v0.2.0</span>
                {:else if s?.connected}
                  <Button variant="tertiary" size="sm" onclick={() => disconnect(ch.name)}>Disconnect</Button>
                {:else}
                  <Button variant="primary" size="sm" onclick={() => openConnectSheet(ch.name)}>Connect</Button>
                {/if}
              </div>
            </Inline>
          {/snippet}
        </Card>
      </div>
    {/each}
  </Stack>

  <footer class="hint">
    <Hairline />
    <p>Polled every 10s. Externally-revoked tokens show as Error until you reconnect.</p>
  </footer>
</div>

{#if sheetOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
  <div class="sheet-scrim" onclick={() => (sheetOpen = false)} role="presentation"></div>
  <aside class="sheet-panel" role="dialog" aria-labelledby="tg-sheet-title" aria-modal="true">
    <Surface variant="overlay" padding="5" radius="none" class="sheet-surface">
      <Stack gap="4">
        <header>
          <h3 id="tg-sheet-title" class="sheet-title">Connect Telegram</h3>
        </header>

        <p class="hint-text">
          Create a bot with <a href="https://t.me/BotFather" target="_blank" rel="noreferrer">@BotFather</a>
          and paste the token below. Format: <code>digits:secret</code>.
        </p>

        <div class="field">
          <label class="field-label" for="tg-token">BotFather token</label>
          <Input
            id="tg-token"
            bind:value={token}
            size="md"
            placeholder="123456789:ABCdef…"
            monospace
            onkeydown={(e: KeyboardEvent) => { if (e.key === 'Enter' && tokenValid) void connectTelegram() }}
          />
          {#if tokenError}
            <p class="field-error" role="alert">{tokenError}</p>
          {/if}
        </div>

        <Inline gap="2" justify="end">
          <Button variant="tertiary" onclick={() => (sheetOpen = false)}>Cancel</Button>
          <Button
            variant="primary"
            onclick={connectTelegram}
            disabled={!tokenValid || connecting}
            loading={connecting}
          >
            {connecting ? 'Connecting…' : 'Connect'}
          </Button>
        </Inline>

        {#if error}
          <p class="err-text" role="alert">{error}</p>
        {/if}
      </Stack>
    </Surface>
  </aside>
{/if}

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
    max-width: 48rem;
    margin: 0 auto;
    background-color: var(--surface-base);
  }

  .page-header {
    margin-bottom: var(--space-6);
  }

  .page-title {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    font-weight: var(--text-h2-weight);
    letter-spacing: var(--text-h2-tracking);
    color: var(--content-primary);
    margin: 0 0 var(--space-2) 0;
    line-height: var(--text-h2-leading);
  }

  .lede {
    font-size: var(--text-body-size);
    color: var(--content-secondary);
    line-height: 1.55;
    max-width: 40rem;
    margin: 0;
  }

  .row-wrap {
    animation: channels-stagger var(--duration-base) var(--ease-standard) both;
    animation-delay: calc(var(--stagger-index, 0) * 60ms);
  }

  @keyframes channels-stagger {
    from { opacity: 0; transform: translateY(6px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .channel-row {
    width: 100%;
  }

  .logo {
    width: 40px;
    height: 40px;
    border-radius: var(--radius-md);
    background-color: var(--surface-sunken);
    border: 1px solid var(--border-default);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--content-tertiary);
    flex-shrink: 0;
  }
  .logo[data-channel='telegram'] { color: #229ED9; }
  .logo[data-channel='slack']    { color: #ECB22E; }
  .logo[data-channel='discord']  { color: #5865F2; }
  .logo[data-channel='imessage'] { color: var(--status-success-fg); }
  .logo[data-channel='whatsapp'] { color: #25D366; }

  .row-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .name {
    font-size: var(--text-body-size);
    font-weight: 500;
    color: var(--content-primary);
  }

  .chat-id {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
  }

  .err-text {
    font-size: var(--text-caption-size);
    color: var(--status-error-fg);
  }

  .row-actions {
    flex-shrink: 0;
  }

  .muted-xs {
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
  }

  .hint {
    margin-top: var(--space-5);
    padding-top: var(--space-4);
  }

  .hint p {
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    margin: var(--space-3) 0 0 0;
  }

  .error-banner {
    color: var(--status-error-fg);
    font-size: var(--text-body-sm-size);
    padding: var(--space-2) var(--space-3);
    background-color: var(--status-error-bg);
    border: 1px solid var(--status-error-border);
    border-radius: var(--radius-md);
    margin-bottom: var(--space-4);
  }

  .sheet-scrim {
    position: fixed;
    inset: 0;
    background-color: var(--surface-scrim);
    z-index: 200;
  }

  .sheet-panel {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    width: min(420px, 100vw);
    z-index: 201;
    border-left: 1px solid var(--border-default);
    box-shadow: none;
  }

  .sheet-surface {
    height: 100%;
    overflow-y: auto;
  }

  .sheet-title {
    font-family: var(--font-sans);
    font-size: var(--text-h4-size);
    font-weight: var(--text-h4-weight);
    color: var(--content-primary);
    margin: 0;
  }

  .hint-text {
    font-size: var(--text-body-sm-size);
    color: var(--content-secondary);
    line-height: 1.55;
    margin: 0;
  }

  .hint-text a {
    color: var(--content-link);
  }

  .hint-text code {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    background-color: var(--surface-sunken);
    border: 1px solid var(--border-default);
    padding: 1px 5px;
    border-radius: var(--radius-sm);
  }

  .field-label {
    display: block;
    font-size: var(--text-caption-size);
    font-family: var(--font-mono);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-tertiary);
    margin-bottom: var(--space-2);
  }

  .field-error {
    font-size: var(--text-caption-size);
    color: var(--status-error-fg);
    margin: var(--space-1) 0 0 0;
  }

  .channels-page :global(.field .input) {
    border: 1px solid var(--border-default);
    border-radius: var(--radius-sm);
    padding: 0 var(--space-3);
  }
</style>
