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

  /** Fixed catalog: live Telegram + honest v0.2 soon rows. Daemon list only updates known ids. */
  const CATALOG: ChannelRow[] = [
    { id: 'telegram', name: 'Telegram', state: 'off', hint: 'Connect a BotFather token' },
    { id: 'whatsapp', name: 'WhatsApp', state: 'soon' },
    { id: 'slack', name: 'Slack', state: 'soon' },
    { id: 'discord', name: 'Discord', state: 'soon' },
    { id: 'imessage', name: 'iMessage', state: 'soon' },
  ]

  /** BotFather shape: digits:secret (legacy Channels validates the same). */
  const BOT_TOKEN_RE = /^\d+:[A-Za-z0-9_-]{20,}$/

  let channels = $state<ChannelRow[]>(CATALOG.map((c) => ({ ...c })))
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
  function mapChannel(c: ChannelInfo): Partial<ChannelRow> & { id: string } {
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
      chatId: c.chat_id || undefined,
    }
  }

  function mergeCatalog(list: ChannelInfo[]): ChannelRow[] {
    const byId = new Map<string, ChannelInfo>()
    for (const raw of list) {
      const id = (raw.channel || raw.id || raw.name || '').toLowerCase()
      if (id) byId.set(id, raw)
    }
    return CATALOG.map((base) => {
      const hit = byId.get(base.id)
      if (!hit) return { ...base }
      // Preserve soon rows — daemon only returns persisted live channels.
      if (base.state === 'soon') return { ...base }
      const mapped = mapChannel(hit)
      return {
        ...base,
        ...mapped,
        id: base.id,
        name: mapped.name || base.name,
      }
    })
  }

  async function load(): Promise<void> {
    loading = true
    error = ''
    try {
      const list = await ipc.channelsList()
      if (Array.isArray(list) && list.length) {
        channels = mergeCatalog(list)
      } else {
        channels = CATALOG.map((c) => ({ ...c }))
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
    if (!BOT_TOKEN_RE.test(token)) {
      submitError = 'Token must look like 123456789:ABCdefGHIjklMNOpqr'
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
              chatId: info.chat_id || undefined,
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
    submitError = null
    try {
      await ipc.channelsDisconnect(id)
      channels = channels.map((c) =>
        c.id === id ? { ...c, state: 'off', hint: undefined, chatId: undefined } : c
      )
    } catch (e) {
      submitError = String(e)
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
  const live = $derived(channels.filter((c) => !isSoon(c)))
  const soon = $derived(channels.filter((c) => isSoon(c)))
  const telegram = $derived(live.find((c) => c.id === 'telegram') ?? live[0] ?? null)
  const liveNote = $derived(
    telegram && isConnected(telegram)
      ? 'Telegram live · messages still Gatekeeper-gated'
      : openInput === 'telegram'
        ? 'Paste the BotFather token to seal Reach'
        : 'Reach is opt-in · revoke anytime'
  )

  function beginToken(id: string): void {
    if (id !== 'telegram') return
    openInput = id
    inputToken = ''
    submitError = null
  }

  function goAudit(): void {
    window.location.hash = '#/audit'
  }
</script>

<MeridianPage
  kicker="Reach · on your terms"
  title="Channels"
  lead="Talk to Condura from elsewhere. Telegram is live today — every connection is gated, local, and revocable."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void load()}>Refresh</button>
  {/snippet}

  {#if loading}
    <div class="md-empty">Checking channels…</div>
  {:else if error}
    <div class="md-empty">{error}</div>
  {:else}
    <div class="reach md-stagger">
      <p class="contract" class:hot={!!telegram && isConnected(telegram)}>
        <span class="live-dot" aria-hidden="true"></span>
        {liveNote}. Remote text never bypasses consent.
      </p>

      <ol class="pipe" aria-label="How Reach connects">
        <li><span class="n">01</span><span class="t">Mint bot</span></li>
        <li><span class="n">02</span><span class="t">Paste token</span></li>
        <li><span class="n">03</span><span class="t">Gatekeeper</span></li>
        <li><span class="n">04</span><span class="t">Revoke anytime</span></li>
      </ol>

      {#if telegram}
        <section class="instrument" class:live={isConnected(telegram)} class:degraded={telegram.state === 'degraded'}>
          <p class="cite">instrument · telegram</p>
          <div class="inst-head">
            <div class="node" class:on={isConnected(telegram)} aria-hidden="true">T</div>
            <div>
              <h2>Telegram</h2>
              <p class="hint">
                {#if isConnected(telegram)}
                  Live node. Messages still pass the Gatekeeper before any action.
                {:else if telegram.state === 'degraded'}
                  Connection degraded — refresh or revoke and reconnect with a fresh token.
                {:else}
                  Mint a bot with BotFather, then paste the token. Condura never sees your personal Telegram password.
                {/if}
              </p>
            </div>
          </div>

          {#if isConnected(telegram)}
            <dl class="facts">
              <div>
                <dt>Status</dt>
                <dd class="ok">connected</dd>
              </div>
              <div>
                <dt>Chat</dt>
                <dd>{telegram.chatId || 'linked'}</dd>
              </div>
              <div class="wide">
                <dt>Contract</dt>
                <dd>remote ask → plan → consent → audit</dd>
              </div>
            </dl>
            <div class="actions-row">
              <button type="button" class="md-btn md-btn-ghost" onclick={goAudit}>Open Audit</button>
              <button
                type="button"
                class="md-btn md-btn-danger"
                disabled={busy === telegram.id}
                onclick={() => void disconnect(telegram.id)}
              >
                {busy === telegram.id ? 'Disconnecting…' : 'Revoke channel'}
              </button>
            </div>
            {#if submitError}
              <p class="err" role="alert">{submitError}</p>
            {/if}
          {:else if openInput === telegram.id}
            <p class="step-note">Step 02 — paste the token BotFather gave you.</p>
            <div class="token-form" role="group" aria-label="Telegram bot token">
              <input
                type="password"
                class="token-input"
                placeholder="123456:ABC… bot token"
                bind:value={inputToken}
                onkeydown={(e) => {
                  if (e.key === 'Enter') void submitToken(telegram.id)
                  if (e.key === 'Escape') cancelInput()
                }}
                disabled={busy === telegram.id}
              />
              <button
                type="button"
                class="md-btn md-btn-primary"
                disabled={busy === telegram.id || !inputToken.trim()}
                onclick={() => void submitToken(telegram.id)}
              >
                {busy === telegram.id ? 'Connecting…' : 'Seal Reach'}
              </button>
              <button type="button" class="md-btn md-btn-ghost" disabled={busy === telegram.id} onclick={cancelInput}>
                Cancel
              </button>
            </div>
            {#if submitError}
              <p class="err">{submitError}</p>
            {/if}
          {:else}
            <div class="actions-row">
              <button type="button" class="md-btn md-btn-primary" onclick={() => connect(telegram.id)}>
                Connect with BotFather
              </button>
              <button type="button" class="md-btn md-btn-ghost" onclick={() => beginToken(telegram.id)}>
                I have a token
              </button>
              <button type="button" class="md-btn md-btn-ghost" onclick={openBotFather}>Open BotFather</button>
            </div>
          {/if}
        </section>
      {/if}

      {#if soon.length}
        <section class="horizon">
          <p class="cite">coming along the meridian · v0.2.0</p>
          <p class="horizon-note">Honestly not wired yet — no fake Connect buttons.</p>
          <div class="soon-strip">
            {#each soon as ch (ch.id)}
              <span>{ch.name}</span>
            {/each}
          </div>
        </section>
      {/if}
    </div>
  {/if}
</MeridianPage>

<style>
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 12px;
  }
  .reach {
    display: grid;
    gap: 16px;
    max-width: 640px;
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
    border-radius: 7px;
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
  .instrument {
    border-radius: 22px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 24px;
    box-shadow: var(--md-shadow);
  }
  .instrument.live {
    border-color: color-mix(in oklab, var(--md-live) 35%, transparent);
  }
  .instrument.degraded {
    border-color: color-mix(in oklab, var(--md-halt) 30%, transparent);
  }
  .inst-head {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 14px;
    align-items: start;
    margin-bottom: 18px;
  }
  .node {
    width: 52px;
    height: 52px;
    border-radius: 16px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 22px;
    font-weight: 700;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 22%, var(--md-line));
  }
  .node.on {
    color: var(--md-live);
    background: color-mix(in oklab, var(--md-live) 12%, var(--md-stage));
    border-color: color-mix(in oklab, var(--md-live) 28%, var(--md-line));
  }
  h2 {
    font-family: var(--md-font-display);
    font-size: 26px;
    letter-spacing: -0.04em;
    margin: 0 0 6px;
  }
  .hint {
    margin: 0;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .step-note {
    margin: 0 0 10px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    color: var(--md-cobalt);
  }
  .facts {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
    margin: 0 0 18px;
    padding: 14px 0;
    border-top: 1px solid var(--md-line);
    border-bottom: 1px solid var(--md-line);
  }
  .facts .wide {
    grid-column: 1 / -1;
  }
  .facts dt {
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin-bottom: 4px;
  }
  .facts dd {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
  }
  .facts dd.ok {
    color: var(--md-live);
  }
  .actions-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .token-form {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }
  .token-input {
    flex: 1 1 200px;
    min-width: 0;
    padding: 10px 12px;
    font-family: var(--md-font-mono);
    font-size: 12px;
    border: 1px solid var(--md-line-strong);
    border-radius: 12px;
    background: var(--md-stage);
    outline: none;
  }
  .token-input:focus-visible {
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
  }
  .err {
    margin: 10px 0 0;
    font-size: 13px;
    color: var(--md-halt);
  }
  .horizon {
    padding: 16px 18px;
    border-radius: 18px;
    border: 1px dashed var(--md-line-strong);
    background: color-mix(in oklab, var(--md-stage) 60%, transparent);
  }
  .horizon-note {
    margin: 0 0 10px;
    font-size: 13px;
    color: var(--md-ink-mute);
  }
  .soon-strip {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .soon-strip span {
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    padding: 6px 10px;
    border-radius: 6px;
    border: 1px solid var(--md-line);
    color: var(--md-ink-faint);
    background: var(--md-surface);
  }
</style>