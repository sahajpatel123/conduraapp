<script lang="ts">
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { ipc } from '../../ipc/client'
  import type { ChannelInfo } from '../../ipc/types'

  // Meridian Channels — messaging integrations. Telegram is the only
  // one wired in v0.1.0 (via the `reach` subsystem). The others are
  // honestly marked "v0.2.0" rather than faked (per the locked
  // decisions).
  //
  // This is the Meridian-shell counterpart to lib/condura/Channels.svelte.
  // It uses the typed `ipc.channelsConnect(channel, token)` RPC (added
  // in lib/ipc/client.ts) so TypeScript catches missing-argument bugs
  // and the daemon returns structured errors. The earlier version of
  // this file cast `ipc as unknown as {...}` and called `connect(id, '')`
  // with an empty token, which is what produced the runtime error
  // "Channel connect is unavailable in this build." — the typed
  // methods are not optional on the real client; the cast was masking
  // them, and the empty token would have failed Telegram validation
  // even if the cast had not.

  type ChannelRow = {
    id: string
    name: string
    state: 'connected' | 'degraded' | 'off' | 'soon'
    hint?: string
    chatId?: string
  }

  let channels = $state<ChannelRow[]>([
    { id: 'telegram', name: 'Telegram', state: 'off', hint: 'Connect a BotFather token' },
    { id: 'whatsapp', name: 'WhatsApp', state: 'soon' },
    { id: 'slack', name: 'Slack', state: 'soon' },
    { id: 'discord', name: 'Discord', state: 'soon' },
    { id: 'imessage', name: 'iMessage', state: 'soon' },
  ])
  let loading = $state(true)
  let error = $state('')

  // Inline token-entry state. openInput is the channel id whose
  // token form is currently visible; inputToken holds the typed
  // value; submitError surfaces a backend failure inline; busy is
  // the channel id currently being submitted.
  let openInput = $state<string | null>(null)
  let inputToken = $state('')
  let submitError = $state<string | null>(null)
  let busy = $state<string | null>(null)

  onMount(() => {
    void load()
  })

  // Map a ChannelInfo from the daemon into the row shape the UI
  // renders. The daemon returns the full ChannelInfo with a
  // `channel` field that other IPCs (like `channel` here in
  // ChannelInfo) use as the canonical id; we fall back to `id`
  // and `name` for backward-compat with any older builds that
  // don't yet emit the `channel` field.
  function mapChannel(c: ChannelInfo): ChannelRow {
    const id = (c.channel || c.id || c.name || '').toLowerCase() || 'unknown'
    let state: ChannelRow['state'] = 'off'
    if (c.connected || c.status === 'connected') state = 'connected'
    else if (c.status === 'degraded') state = 'degraded'
    else if (c.status === 'soon' || c.status === 'v0.2.0') state = 'soon'
    return {
      id,
      name: c.name || c.channel || id,
      state,
      hint: c.detail,
    }
  }

  async function load(): Promise<void> {
    loading = true
    error = ''
    try {
      const list = await ipc.channelsList()
      if (Array.isArray(list) && list.length) {
        channels = list.map(mapChannel)
      }
    } catch (e) {
      const s = String(e)
      // IPC-not-started / connection-refused errors mean the daemon
      // isn't reachable yet; don't surface those as user-visible
      // failures — keep the defaults rendered.
      error = /IPC client not started|not connected|Failed to fetch/i.test(s) ? '' : s
    } finally {
      loading = false
    }
  }

  // connect is the entry point on the row's Connect button. It
  // opens BotFather so the user can mint a bot, then reveals the
  // inline token-entry form. No daemon call yet — the token is
  // what the daemon needs; we wait for it before calling
  // channels.connect.
  function connect(id: string): void {
    if (id !== 'telegram') return
    openBotFather()
    openInput = id
    inputToken = ''
    submitError = null
  }

  async function submitToken(id: string): Promise<void> {
    if (id !== 'telegram') return
    const token = inputToken.trim()
    if (!token) {
      submitError = 'Token is empty'
      return
    }
    busy = id
    submitError = null
    try {
      const info = await ipc.channelsConnect('telegram', token)
      channels = channels.map((c) =>
        c.id === id
          ? {
              ...c,
              state: 'connected',
              hint: undefined,
              chatId: (info as { chat_id?: string }).chat_id,
            }
          : c
      )
      openInput = null
      inputToken = ''
    } catch (e) {
      submitError = String(e)
    } finally {
      busy = null
    }
  }

  function cancelInput(): void {
    openInput = null
    inputToken = ''
    submitError = null
  }

  // disconnect tears down an existing connection.
  async function disconnect(id: string): Promise<void> {
    busy = id
    try {
      await ipc.channelsDisconnect(id)
      channels = channels.map((c) =>
        c.id === id ? { ...c, state: 'off', hint: undefined, chatId: undefined } : c
      )
    } catch {
      // keep honest state
    } finally {
      busy = null
    }
  }

  function openBotFather(): void {
    const url = 'https://t.me/BotFather'
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } }
    if (w.runtime?.BrowserOpenURL) {
      try {
        w.runtime.BrowserOpenURL(url)
        return
      } catch {
        // fall through to window.open
      }
    }
    window.open(url, '_blank', 'noopener,noreferrer')
  }

  function isConnected(ch: ChannelRow): boolean {
    return ch.state === 'connected'
  }

  function isSoon(ch: ChannelRow): boolean {
    return ch.state === 'soon'
  }
</script>

<MeridianPage
  kicker="Reach · on your terms"
  title="Channels"
  lead="Talk to Condura from elsewhere — Telegram today, more later. Each connection is gated; you can revoke it any time."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void load()}>Refresh</button>
  {/snippet}

  {#if loading}
    <div class="md-empty">Checking channels…</div>
  {:else if error}
    <div class="md-empty">{error}</div>
  {:else}
    <div class="list md-stagger">
      {#each channels as ch (ch.id)}
        <div class="item">
          <div class="mono" aria-hidden="true">{ch.name.trim()[0]?.toUpperCase() || '?'}</div>
          <div class="copy">
            <strong>{ch.name}</strong>
            <span class="meta" class:on={isConnected(ch)}>
              {#if submitError && openInput === ch.id}
                {submitError}
              {:else if isConnected(ch) && ch.chatId}
                connected · chat {ch.chatId}
              {:else if isConnected(ch)}
                connected
              {:else if isSoon(ch)}
                coming in v0.2.0
              {:else if openInput === ch.id}
                paste your BotFather token
              {:else}
                {ch.hint ?? 'not connected'}
              {/if}
            </span>
          </div>
          {#if isSoon(ch)}
            <span class="soon">Soon</span>
          {:else if isConnected(ch)}
            <button
              type="button"
              class="md-btn md-btn-danger"
              disabled={busy === ch.id}
              onclick={() => void disconnect(ch.id)}
            >
              {busy === ch.id ? 'disconnecting…' : 'Disconnect'}
            </button>
          {:else if openInput !== ch.id}
            <button
              type="button"
              class="md-btn md-btn-primary"
              onclick={() => connect(ch.id)}
            >
              Connect
            </button>
          {/if}
        </div>
        {#if openInput === ch.id && !isSoon(ch) && !isConnected(ch)}
          <div class="token-form" role="group" aria-label="Telegram bot token">
            <input
              type="password"
              class="token-input"
              placeholder="bot token from BotFather"
              bind:value={inputToken}
              onkeydown={(e) => {
                if (e.key === 'Enter') void submitToken(ch.id)
                if (e.key === 'Escape') cancelInput()
              }}
              disabled={busy === ch.id}
            />
            <button
              type="button"
              class="md-btn md-btn-primary token-submit"
              disabled={busy === ch.id || !inputToken.trim()}
              onclick={() => void submitToken(ch.id)}
            >
              {busy === ch.id ? 'connecting…' : 'Connect'}
            </button>
            <button
              type="button"
              class="md-btn md-btn-ghost token-cancel"
              disabled={busy === ch.id}
              onclick={cancelInput}
            >
              Cancel
            </button>
          </div>
        {/if}
      {/each}
    </div>
  {/if}
</MeridianPage>

<style>
  .list {
    background: var(--md-surface);
    border: 1px solid var(--md-line-strong);
    border-radius: 22px;
    padding: 8px;
    box-shadow: var(--md-shadow);
  }
  .item {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 14px;
    align-items: center;
    margin: 0;
    padding: 14px 12px;
    border-radius: 16px;
    transition: background 160ms var(--md-ease);
  }
  .item:hover { background: var(--md-stage); }
  .item:focus-within {
    background: var(--md-stage);
    box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--md-cobalt) 35%, transparent);
  }
  .mono {
    width: 42px;
    height: 42px;
    border-radius: 14px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -0.04em;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
  }
  .copy { min-width: 0; }
  strong {
    display: block;
    font-family: var(--md-font-display);
    font-size: 18px;
    letter-spacing: -0.03em;
  }
  .meta {
    display: block;
    margin-top: 3px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .meta.on { color: var(--md-live); }
  .soon {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    padding: 6px 10px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    background: var(--md-stage);
  }
  .token-form {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 14px;
    margin: 4px 0 8px 54px;
    border-left: 2px solid var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 5%, transparent);
    border-radius: 0 12px 12px 0;
    animation: token-form-in 240ms var(--md-ease) both;
  }
  .token-input {
    flex: 1 1 auto;
    min-width: 0;
    padding: 9px 12px;
    font-family: var(--md-font-mono);
    font-size: 12px;
    color: var(--md-ink);
    background: var(--md-surface);
    border: 1px solid var(--md-line);
    border-radius: 10px;
    outline: none;
    transition:
      border-color var(--md-dur) var(--md-ease),
      box-shadow var(--md-dur) var(--md-ease);
  }
  .token-input:focus-visible {
    border-color: var(--md-cobalt);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--md-cobalt) 14%, transparent);
  }
  .token-input:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
  .token-submit,
  .token-cancel {
    flex: none;
    padding: 9px 14px;
    font-size: 12px;
  }
  @keyframes token-form-in {
    from { opacity: 0; transform: translateY(-4px); }
    to   { opacity: 1; transform: translateY(0); }
  }
  @media (max-width: 560px) {
    .list {
      padding: 6px;
      border-radius: 18px;
    }
    .item {
      grid-template-columns: auto 1fr;
      gap: 10px;
      padding: 12px 10px;
    }
    .item .md-btn,
    .item button,
    .item .soon {
      grid-column: 1 / -1;
      justify-self: stretch;
      text-align: center;
    }
    .token-form {
      flex-direction: column;
      align-items: stretch;
      margin-left: 12px;
      padding: 10px;
    }
    .token-submit,
    .token-cancel {
      width: 100%;
    }
    .mono {
      width: 38px;
      height: 38px;
      font-size: 16px;
    }
    strong {
      font-size: 16px;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .token-form { animation: none !important; }
  }
</style>