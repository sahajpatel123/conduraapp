<script lang="ts">
  import { onMount } from 'svelte';
  import { consent } from '../stores/consent.svelte';
  import Button from './Button.svelte';
  import Glyph from './Glyph.svelte';

  // The gatekeeper's face. A protective synapse thread draws a rectangle
  // around the action summary (armoring it). On approval, a wax seal stamps.
  // Destructive actions get the rare ink surface; gentler ones get paper.
  // The store polls gatekeeper.pending_consent every 1.2s; we only render
  // when a ticket is non-null.

  let sealed = $state(false);
  let draw = $state(false);

  let ticket = $derived(consent.ticket);
  let hasTicket = $derived(ticket !== null);
  let destructive = $derived(ticket?.action_kind === 'destructive');

  function blastLabel(kind: string | undefined): string {
    switch (kind) {
      case 'read':
        return 'Read';
      case 'write':
        return 'Write';
      case 'network':
        return 'Network';
      case 'destructive':
        return 'Destructive';
      default:
        return 'Action';
    }
  }
  let detail = $derived(ticket?.detail || ticket?.actor || 'this application');
  let actorLine = $derived(ticket?.actor ? `Requested by ${ticket.actor}` : 'Requested by the agent');
  let countdownPct = $derived((consent.timer / 300000) * 100);

  onMount(() => {
    try {
      consent.start();
    } catch (e) {
      console.warn('consent.start failed', e);
    }
    return () => {
      try {
        consent.stop();
      } catch {
        /* ignore */
      }
    };
  });

  // re-arm the armor + seal whenever a new ticket arrives
  let lastNonce = '';
  $effect(() => {
    const n = ticket?.nonce ?? '';
    if (n && n !== lastNonce) {
      lastNonce = n;
      sealed = false;
      draw = false;
      requestAnimationFrame(() => requestAnimationFrame(() => (draw = true)));
    }
  });

  async function approve(): Promise<void> {
    sealed = true;
    try {
      await consent.approve();
    } catch (e) {
      console.error('approve failed', e);
    }
  }
  async function deny(): Promise<void> {
    sealed = false;
    try {
      await consent.deny();
    } catch (e) {
      console.error('deny failed', e);
    }
  }
</script>

{#if hasTicket}
  <div class="scrim" class:ink={destructive}>
    <div class="consent-card" class:ink={destructive} role="alertdialog" aria-modal="true">
      <div class="c-eyebrow">{blastLabel(ticket?.action_kind)} action · requires your consent</div>
      <h2 class="c-title">Condura wants to act.</h2>
      <p class="c-sub">{destructive ? 'This cannot be undone. Review exactly what will happen before you allow.' : 'Review what will happen before you allow.'}</p>

      <div class="action-summary">
        <div class="as-eyebrow">Action summary</div>
        <div class="as-body">{detail}</div>
        <div class="as-meta">{actorLine} · nonce {ticket?.nonce?.slice(0, 8)}…</div>
        <svg class="armor-rect" preserveAspectRatio="none" aria-hidden="true">
          <rect
            x="2"
            y="2"
            width="96%"
            height="96%"
            rx="12"
            ry="12"
            fill="none"
            stroke="var(--synapse-glow)"
            stroke-width="1.5"
            pathLength="1"
            vector-effect="non-scaling-stroke"
            stroke-dasharray="1"
            stroke-dashoffset={draw ? 0 : 1}
            style="transition: stroke-dashoffset 1.4s var(--ease) 0.15s"
          />
        </svg>
      </div>

      <div class="consent-foot">
        <span class="hint">Esc to deny</span>
        <Button variant={destructive ? 'danger' : 'ghost'} class="deny" onclick={deny}>Deny</Button>
        <Button variant="primary" magnetic class="allow" onclick={approve}>
          {sealed ? 'Allowed' : 'Allow'}
        </Button>
      </div>

      <div class="countdown"><div class="countdown-fill" style:width="{countdownPct}%"></div></div>

      {#if sealed}
        <div class="seal show" aria-hidden="true">
          <div>
            <div class="seal-word">Allowed</div>
            <div class="seal-sub">by you · now</div>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .scrim {
    position: fixed;
    inset: 0;
    z-index: var(--z-modal);
    background: color-mix(in oklab, var(--ink) 32%, transparent);
    backdrop-filter: blur(6px);
    display: grid;
    place-items: center;
    animation: blur-in var(--dur) var(--ease);
  }
  .scrim.ink {
    background: color-mix(in oklab, var(--surface-ink) 50%, transparent);
  }

  .consent-card {
    position: relative;
    width: min(540px, 92vw);
    background: var(--surface);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-lg);
    box-shadow: var(--shadow-float);
    padding: var(--space-8);
    color: var(--content);
  }
  .consent-card.ink {
    background: var(--surface-ink);
    border-color: color-mix(in oklab, var(--paper) 10%, transparent);
    color: var(--paper);
  }

  .c-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--content-mute);
  }
  .consent-card.ink .c-eyebrow {
    color: color-mix(in oklab, var(--paper) 55%, transparent);
  }
  .c-title {
    font-family: var(--font-display);
    font-size: 30px;
    line-height: 1.1;
    letter-spacing: -0.03em;
    margin: var(--space-3) 0 var(--space-2);
  }
  .c-sub {
    font-size: 14px;
    color: var(--content-mute);
    margin-bottom: var(--space-6);
  }
  .consent-card.ink .c-sub {
    color: color-mix(in oklab, var(--paper) 55%, transparent);
  }

  .action-summary {
    position: relative;
    padding: var(--space-5);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
    margin-bottom: var(--space-6);
  }
  .consent-card.ink .action-summary {
    border-color: color-mix(in oklab, var(--paper) 12%, transparent);
    background: color-mix(in oklab, var(--paper) 4%, transparent);
  }
  .as-eyebrow {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--synapse);
    margin-bottom: var(--space-2);
  }
  .consent-card.ink .as-eyebrow {
    color: color-mix(in oklab, var(--synapse-light) 85%, transparent);
  }
  .as-body {
    font-family: var(--font-display);
    font-size: 22px;
    line-height: 1.2;
    letter-spacing: -0.02em;
    color: var(--content);
  }
  .consent-card.ink .as-body {
    color: var(--paper);
  }
  .as-meta {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--content-faint);
    margin-top: var(--space-3);
  }
  .consent-card.ink .as-meta {
    color: color-mix(in oklab, var(--paper) 45%, transparent);
  }
  .armor-rect {
    position: absolute;
    inset: -4px;
    width: calc(100% + 8px);
    height: calc(100% + 8px);
    pointer-events: none;
  }

  .consent-foot {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }
  .hint {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-right: auto;
  }
  .consent-card.ink .hint {
    color: color-mix(in oklab, var(--paper) 40%, transparent);
  }
  .consent-card.ink :global(.btn-ghost.deny) {
    background: transparent;
    color: var(--paper);
    border-color: color-mix(in oklab, var(--paper) 20%, transparent);
  }

  .countdown {
    margin-top: var(--space-5);
    height: 2px;
    background: var(--hair);
    border-radius: 1px;
    overflow: hidden;
  }
  .countdown-fill {
    height: 100%;
    background: var(--warn);
    transition: width 1s linear;
  }

  .seal {
    position: absolute;
    left: 50%;
    top: 50%;
    width: 108px;
    height: 108px;
    border-radius: 50%;
    pointer-events: none;
    z-index: var(--z-tooltip);
    background: radial-gradient(circle at 35% 30%, var(--synapse-glow), var(--synapse-deep) 70%);
    color: var(--paper);
    display: grid;
    place-items: center;
    text-align: center;
    box-shadow:
      0 20px 50px -16px color-mix(in oklab, var(--synapse) 70%, transparent),
      inset 0 0 0 3px color-mix(in oklab, var(--paper) 25%, transparent);
    transform: translate(-50%, -50%) scale(0);
    opacity: 0;
  }
  .seal.show {
    animation: stamp var(--dur-slow) var(--ease) forwards;
  }
  .seal-word {
    font-family: var(--font-display);
    font-size: 20px;
    line-height: 1;
  }
  .seal-sub {
    font-family: var(--font-mono);
    font-size: 8px;
    letter-spacing: 0.18em;
    text-transform: uppercase;
    margin-top: 4px;
    opacity: 0.8;
  }
</style>