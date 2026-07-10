<script lang="ts">
  import { onMount } from 'svelte';
  import { ipc } from '../ipc/client';
  import Pulse from './Pulse.svelte';
  import Thread from './Thread.svelte';
  import Button from './Button.svelte';
  import { ROUTE_HASH } from './NavRail.svelte';

  // Condura Channels — messaging integrations. Telegram is the only one wired
  // in v0.1.0 (via the `reach` subsystem). The others are honestly marked
  // "v0.2.0" rather than faked (per the locked decisions).
  //
  // Signal-quality is the visual metaphor: each row carries five cellular-bar
  // dots whose heights step up like real radio bars. When connected they
  // breathe with a staggered cascade using the shared `breathe` keyframe.
  // When degraded the lit dots hold steady in --warn. The "Connect" action
  // opens BotFather (via the Wails runtime, falling back to window.open) and
  // reveals an inline token-input field for the user to paste the bot token.

  type ChannelRow = {
    id: string;
    name: string;
    state: 'connected' | 'degraded' | 'off' | 'soon';
    hint?: string;
    chatId?: string;
  };

  let rows = $state<ChannelRow[]>([
    { id: 'telegram', name: 'Telegram', state: 'off', hint: 'Connect a BotFather token' },
    { id: 'whatsapp', name: 'WhatsApp', state: 'soon' },
    { id: 'slack', name: 'Slack', state: 'soon' },
    { id: 'discord', name: 'Discord', state: 'soon' },
    { id: 'imessage', name: 'iMessage', state: 'soon' },
  ]);
  let busy = $state<string | null>(null);
  let noteIn = $state(false);
  let hydrating = $state(true);
  let hydrateError = $state<string | null>(null);

  // Inline token-entry state. openInput is the channel id whose
  // token form is currently visible; inputToken holds the typed
  // value; submitError surfaces a backend failure inline.
  let openInput = $state<string | null>(null);
  let inputToken = $state('');
  let submitError = $state<string | null>(null);
  let submitting = $state<string | null>(null);

  // try to hydrate the list from the daemon (best-effort) and draw the note
  onMount(async () => {
    try {
      const list = await ipc.channelsList();
      if (Array.isArray(list) && list.length) {
        rows = list.map((c) => {
          const id = (c.channel || c.id || c.name || '').toLowerCase() || 'unknown';
          let state: ChannelRow['state'] = 'off';
          if (c.connected || c.status === 'connected') state = 'connected';
          else if (c.status === 'degraded') state = 'degraded';
          else if (c.status === 'soon' || c.status === 'v0.2.0') state = 'soon';
          return {
            id,
            name: c.name || c.channel || id,
            state,
            hint: c.detail,
            chatId: (c as { chat_id?: string }).chat_id,
          };
        });
      }
      hydrateError = null;
    } catch (e) {
      hydrateError = String(e);
    } finally {
      hydrating = false;
    }
    // trigger the under-note thread to draw in
    requestAnimationFrame(() => (noteIn = true));
  });

  function openBotFather(): void {
    const url = 'https://t.me/BotFather';
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } };
    if (w.runtime?.BrowserOpenURL) {
      try {
        w.runtime.BrowserOpenURL(url);
        return;
      } catch {
        // fall through to window.open
      }
    }
    window.open(url, '_blank', 'noopener,noreferrer');
  }

  // connect is the entry point on the row's Connect button. It
  // opens BotFather so the user can mint a bot, then reveals the
  // inline token-entry form. No daemon call yet — the token is
  // what the daemon needs; we wait for it before calling
  // channels.connect.
  function connect(id: string): void {
    if (id !== 'telegram') return;
    openBotFather();
    openInput = id;
    inputToken = '';
    submitError = null;
  }

  // submitToken fires the actual channels.connect IPC. On
  // success, the row flips to 'connected'. On error, the error
  // message is rendered inline (in the hint slot).
  async function submitToken(id: string): Promise<void> {
    if (id !== 'telegram') return;
    const token = inputToken.trim();
    if (!token) {
      submitError = 'Token is empty';
      return;
    }
    submitting = id;
    submitError = null;
    try {
      const info = await ipc.channelsConnect('telegram', token);
      rows = rows.map((r) =>
        r.id === id
          ? { ...r, state: 'connected', hint: undefined, chatId: (info as { chat_id?: string }).chat_id }
          : r
      );
      openInput = null;
      inputToken = '';
    } catch (e) {
      submitError = String(e);
    } finally {
      submitting = null;
    }
  }

  function cancelInput(): void {
    openInput = null;
    inputToken = '';
    submitError = null;
  }

  // disconnect is wired to the future Manage button (post-
  // connected). Currently unused in the row template but kept
  // here so the runtime behavior matches the spec.
  async function disconnect(id: string): Promise<void> {
    busy = id;
    try {
      await ipc.channelsDisconnect(id);
      rows = rows.map((r) => (r.id === id ? { ...r, state: 'off', hint: undefined, chatId: undefined } : r));
    } catch {
      // keep honest state
    } finally {
      busy = null;
    }
  }

  function dotCount(state: ChannelRow['state']): number {
    if (state === 'connected') return 5;
    if (state === 'degraded') return 3;
    return 0; // 'off' and 'soon'
  }

  function stateLabel(r: ChannelRow): string {
    if (r.state === 'connected') return 'connected';
    if (r.state === 'degraded') return 'degraded';
    if (r.state === 'soon') return 'unbuilt — coming in v0.2.0';
    return 'not connected';
  }
</script>

<div class="channels">
  <header class="head">
    <div class="eyebrow">— Reach · on your terms</div>
    <h1 class="title">Threads outward.</h1>
    <p class="sub">
      Condura can reach you on Telegram today. WhatsApp, Slack, Discord, and iMessage arrive
      in v0.2.0 — we don't fake them. Each connection is a thread you tie, and you can revoke
      it any time.
    </p>
  </header>

  <div class="grid">
    {#each rows as r (r.id)}
      <div
        class="row-wrap"
      >
        <button
          type="button"
          class="row"
          class:connected={r.state === 'connected'}
          class:degraded={r.state === 'degraded'}
          class:soon={r.state === 'soon'}
          class:open={openInput === r.id}
          aria-label={`${r.name}, ${stateLabel(r)}`}
        >
          <div class="cell">
            <div class="name">{r.name}</div>
            <div class="hint">
              {#if submitError && openInput === r.id}
                {submitError}
              {:else if r.state === 'connected' && r.chatId}
                connected · chat {r.chatId}
              {:else if r.state === 'connected'}
                connected
              {:else if r.state === 'soon'}
                coming in v0.2.0
              {:else if openInput === r.id}
                paste your BotFather token
              {:else}
                {r.hint ?? 'not connected'}
              {/if}
            </div>
          </div>
          <div class="signal" aria-hidden="true">
            {#each Array.from({ length: 5 }) as _, i (i)}
              <span
                class="dot"
                class:on={i < dotCount(r.state)}
                style:--i={i}
              ></span>
            {/each}
          </div>
          <div class="action">
            {#if r.state === 'soon'}
              <span class="pill-soon">v0.2.0</span>
            {:else if r.state === 'connected'}
              <Button
                variant="ghost"
                size="sm"
                magnetic={false}
                disabled={busy === r.id}
                onclick={() => disconnect(r.id)}
              >
                {busy === r.id ? 'disconnecting…' : 'Disconnect'}
              </Button>
            {:else if openInput !== r.id}
              <Button
                variant="primary"
                size="sm"
                magnetic={true}
                onclick={() => connect(r.id)}
              >
                Connect →
              </Button>
            {/if}
          </div>
        </button>
        {#if openInput === r.id && r.state !== 'soon' && r.state !== 'connected'}
          <div class="token-form" role="group" aria-label="Telegram bot token">
            <input
              type="password"
              class="token-input"
              placeholder="bot token from BotFather"
              bind:value={inputToken}
              onkeydown={(e) => {
                if (e.key === 'Enter') void submitToken(r.id);
                if (e.key === 'Escape') cancelInput();
              }}
              disabled={submitting === r.id}
            />
            <Button
              variant="primary"
              size="sm"
              magnetic={true}
              disabled={submitting === r.id || !inputToken.trim()}
              onclick={() => submitToken(r.id)}
            >
              {submitting === r.id ? 'connecting…' : 'Connect'}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              magnetic={false}
              disabled={submitting === r.id}
              onclick={cancelInput}
            >
              Cancel
            </Button>
          </div>
        {/if}
      </div>
    {/each}
  </div>

  {#if hydrating}
    <div class="hydra"><Pulse phase="thinking" size={8} /> <span class="hydra-label">PROBING REACH…</span></div>
  {/if}
  {#if hydrateError}
    <div class="err-state" role="alert" aria-live="polite">
      <div class="err-row">
        <Pulse phase="error" size={8} />
        <span class="err-head">We couldn't reach the daemon.</span>
      </div>
      <p class="err-sub">{hydrateError} The defaults below are honest — Telegram is connectable today, the rest are v0.2.0.</p>
      <div class="err-hair"></div>
    </div>
  {/if}

  <div class="note">
    <Pulse phase="idle" size={6} />
    <span>
      Outbound messages always pass the consent gate. Inbound traffic is logged on the
      <button
        type="button"
        class="threadlink"
        onclick={() => {
          window.location.hash = ROUTE_HASH.audit;
        }}
      >Audit chain</button>.
    </span>
    <Thread orientation="h" draw={noteIn} class="note-thread" />
  </div>
</div>

<style>
  .channels {
    max-width: 880px;
    padding-top: var(--space-7);
  }
  .head {
    margin-bottom: var(--space-6);
  }
  .eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .title {
    font-family: var(--font-display);
    font-size: clamp(28px, 3vw, 40px);
    line-height: 1.08;
    letter-spacing: -0.03em;
    color: var(--content);
    margin: var(--space-3) 0 var(--space-2);
  }
  .sub {
    font-size: 16px;
    line-height: 1.55;
    color: var(--content-soft);
    max-width: 56ch;
  }

  .grid {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-6);
  }
  .row-wrap {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .token-form {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    margin-left: var(--space-5);
    border-left: 2px solid var(--synapse);
    background: color-mix(in oklab, var(--synapse) 4%, transparent);
    border-radius: 0 var(--r-md) var(--r-md) 0;
    animation: token-form-in 240ms var(--ease) both;
  }
  .token-input {
    flex: 1 1 auto;
    min-width: 0;
    padding: var(--space-2) var(--space-3);
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--content);
    background: var(--surface-card);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-sm);
    outline: none;
    transition:
      border-color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .token-input:focus-visible {
    border-color: var(--synapse);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .token-input:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
  .row.open {
    border-color: var(--synapse);
  }
  @keyframes token-form-in {
    from {
      opacity: 0;
      transform: translateY(-4px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  .row {
    display: grid;
    grid-template-columns: 1fr auto auto;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-4) var(--space-5);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
    text-align: left;
    width: 100%;
    cursor: default;
    transition:
      border-color var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      background var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .row:hover {
    border-color: var(--hair-strong);
    transform: translateY(-1px);
    background: var(--paper-2);
  }
  .row:focus-visible {
    outline: none;
    border-color: var(--synapse);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .row.connected {
    border-color: color-mix(in oklab, var(--synapse) 35%, transparent);
  }
  .row.degraded {
    border-color: color-mix(in oklab, var(--warn) 30%, transparent);
  }
  .row.soon {
    opacity: 0.55;
    cursor: not-allowed;
  }
  .row.soon:hover {
    transform: none;
    background: var(--surface-card);
  }
  .name {
    font-family: var(--font-display);
    font-size: 18px;
    color: var(--content);
  }
  .hint {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-top: 2px;
  }

  /* ── cellular-bar signal ──
     Heights step up like real radio bars: 8, 12, 16, 20, 24 px.
     Lit dots inherit the row's color; on connected they breathe in cascade. */
  .signal {
    display: flex;
    align-items: flex-end;
    gap: 3px;
    height: 28px;
  }
  .dot {
    width: 4px;
    background: var(--hair-strong);
    border-radius: 1px;
    transition: background var(--dur) var(--ease);
    height: calc(8px + var(--i, 0) * 4px);
  }
  .dot.on {
    background: var(--synapse);
  }
  .row.degraded .dot.on {
    background: var(--warn);
  }
  .row.connected .dot.on {
    background: var(--synapse);
    animation: breathe 1.6s var(--ease) infinite;
    animation-delay: calc(var(--i, 0) * 0.12s);
  }
  /* honor battery + reduced-motion: never let the cascade thrash */
  :global(:root[data-energy='low']) .row.connected .dot.on {
    animation: none;
  }

  /* v0.2.0 pill — honest, not hype. Mono ink-faint + hairline. */
  .pill-soon {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 6px 12px;
    border-radius: var(--r-pill);
    border: 1px solid var(--hair);
    background: transparent;
    color: var(--content-faint);
    cursor: not-allowed;
  }

  /* ── hydrating indicator + error ── */
  .hydra {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-top: var(--space-3);
  }
  .hydra-label {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .err-state {
    max-width: 560px;
    margin-top: var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .err-row {
    display: inline-flex;
    align-items: center;
    gap: 10px;
  }
  .err-head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 22px;
    line-height: 1.15;
    color: var(--content);
    letter-spacing: -0.01em;
  }
  .err-sub {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    line-height: 1.55;
    color: var(--content-faint);
    max-width: 48ch;
  }
  .err-hair {
    height: 1px;
    width: 100%;
    background: linear-gradient(90deg, var(--hair-strong) 0%, var(--hair-strong) 60%, transparent 100%);
    transform: scaleX(0);
    transform-origin: left;
    animation: err-hair-draw 600ms var(--ease) 120ms forwards;
  }
  @keyframes err-hair-draw {
    to { transform: scaleX(1); }
  }
  @media (prefers-reduced-motion: reduce) {
    .err-hair {
      transform: scaleX(1);
      animation: none;
    }
  }

  /* ── footer note (consent + audit threadlink) ── */
  .note {
    display: grid;
    grid-template-columns: auto 1fr;
    align-items: center;
    column-gap: var(--space-2);
    row-gap: var(--space-2);
    font-family: var(--font-display);
    font-style: italic;
    font-size: 14px;
    color: var(--content-mute);
    padding: var(--space-3) 0;
  }
  .note > span {
    grid-column: 1 / -1;
  }
  .note :global(.note-thread) {
    grid-column: 1 / -1;
    margin-top: var(--space-2);
    opacity: 0.5;
  }
  .threadlink {
    color: var(--synapse);
    margin: 0 2px;
    padding: 2px 4px;
    font: inherit;
    cursor: pointer;
    background: transparent;
    border: 0;
    border-radius: var(--r-xs);
    text-decoration: none;
    position: relative;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .threadlink:hover {
    background: color-mix(in oklab, var(--synapse) 8%, transparent);
  }
  .threadlink:active {
    transform: scale(0.97);
  }
  .threadlink::after {
    content: '';
    position: absolute;
    left: 4px;
    right: 4px;
    bottom: -1px;
    height: 1px;
    background: var(--synapse);
    transform: scaleX(0);
    transform-origin: left;
    transition: transform var(--dur) var(--ease);
  }
  .threadlink:hover::after,
  .threadlink:focus-visible::after {
    transform: scaleX(1);
  }
  .threadlink:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  @media (prefers-reduced-motion: reduce) {
    .row.connected .dot.on {
      animation: none;
    }
    .threadlink::after {
      transition: none;
    }
  }
</style>
