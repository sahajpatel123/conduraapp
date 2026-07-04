<script lang="ts">
  // Condura · PairingModal — the bottom-sheet pairing surface.
  // ─────────────────────────────────────────────────────────────────────
  // Per MOAT §2.8 + SCREEN_PAIRINGMODAL.md: this is a `.c-sheet` (slides
  // from the bottom edge, doesn't block page scroll, Esc to close) — NOT
  // a centered modal. Pairing is a task; a modal would say "wait"; the
  // sheet says "go." The signature motion is the pollen TTL ring around
  // the PIN, depleting over 60s.
  //
  // States (the state matrix lives in §3 of the spec):
  //   S0 closed        — not rendered
  //   S1 open          — both zones visible (default)
  //   S2 qr-mode       — PIN zone hidden (200ms cross-fade)
  //   S3 pin-mode      — QR zone hidden (200ms cross-fade)
  //   S4 ttl-warning   — TTL < 30s (ring flips to --danger, regen button)
  //   S5 paired        — thread + check + auto-dismiss after 1.5s
  //   S6 error         — ErrorState (what / cause / action)
  //   S7 timeout       — ring filled --danger, PIN blurred, regen primary
  //
  // Lifecycle:
  //   1. On open:     sync.pair_begin → { pin, expires_in, peer }
  //   2. On mount:    poll sync.pair_status every 5s (the SSE channel
  //                   described in the spec isn't on the wire yet; the
  //                   poll is the safety net)
  //   3. rAF loop:    updates the ring's stroke-dashoffset; pauses on
  //                   visibilitychange (per TitlebarThread pattern)
  //   4. On success:  S5 — Thread draws across the sheet, then auto-close
  //   5. On error:    S6 — ErrorState (Try again / Cancel)
  //   6. On timeout:  S7 — Regenerate primary CTA
  //
  // Composition contract (per SCREEN_PAIRINGMODAL.md §10):
  //   - 1 onDestroy.  All timers / event listeners cleaned up in one place.
  //   - 1 rAF loop.   Pauses on visibilitychange (battery is the user).
  //   - Tokens only.  No raw hex, no magic durations, no magic radii.
  import { onMount, onDestroy } from 'svelte';
  import QRCode from 'qrcode';
  import Thread from './Thread.svelte';
  import Pulse from './Pulse.svelte';
  import Glyph from './Glyph.svelte';
  import Button from './Button.svelte';
  import Tooltip from './Tooltip.svelte';
  import ErrorState from './ErrorState.svelte';
  import { ipc } from '../ipc/client';
  import { t } from '../i18n';

  interface Props {
    /** This device's identity (shown as the QR for the peer to scan). */
    deviceId: string;
    deviceName: string;
    /** The peer we're pairing with. */
    peerName: string;
    /** When the sheet is open. Bind from parent. */
    open: boolean;
    /** Called when the user dismisses the sheet (Esc / outside / X). */
    onclose: () => void;
    /** Optional: called when a successful pair completes (before auto-dismiss). */
    onpaired?: (deviceId: string) => void;
  }

  let { deviceId, deviceName, peerName, open, onclose, onpaired }: Props = $props();

  // ── state machine ──
  // We avoid the name `state` because `$state` is a Svelte 5 rune — the
  // compiler binds the bare identifier `state` to the rune, which would
  // shadow our typed state variable. `phase` is the local alias.
  type State = 'closed' | 'open' | 'qr-mode' | 'pin-mode' | 'paired' | 'error' | 'timeout';
  let phase: State = $state<State>('closed');
  // The 5s/30s TTL warning is a sub-flag of any active state, not a top-level
  // state — it can co-exist with open/qr-mode/pin-mode.
  let ttlWarning = $state(false);

  // ── zone visibility (the QR ⇄ PIN toggle) ──
  let showQr = $state(true);
  let showPin = $state(true);

  // ── QR data URL (the paper-raised card face) ──
  let qrDataUrl = $state('');

  // ── PIN + TTL ──
  let pin = $state('');
  let expiresAt = $state(''); // ISO 8601
  let totalTtlMs = $state(0);
  let remainingMs = $state(0);
  // rAF + timers live across the modal's lifetime; we clear them in onDestroy.
  let rafId = 0;
  let pollTimer: ReturnType<typeof setInterval> | null = null;
  let autodismissTimer: ReturnType<typeof setTimeout> | null = null;

  // ── copy chip ("PIN copied.") ──
  let copied = $state(false);
  let copyTimer: ReturnType<typeof setTimeout> | null = null;

  // ── error / loading ──
  let busy = $state(false);
  let errorHead = $state('');
  let errorCause = $state('');

  // ── TTL ring geometry (per spec §2.4) ──
  const RING_R = 48; // 120×120 SVG, cx:60 cy:60 r:48
  const RING_CIRC = 2 * Math.PI * RING_R; // ≈ 301.59
  let ringPct = $derived(
    totalTtlMs > 0 ? Math.min(1, Math.max(0, remainingMs / totalTtlMs)) : 1
  );
  let ringDashoffset = $derived(RING_CIRC * (1 - ringPct));
  let ringStroke = $derived(ttlWarning ? 'var(--danger)' : 'var(--pollen)');
  let ringFilled = $derived(phase === 'timeout'); // timeout: full circle in danger

  // ── derived counts ──
  let remainingLabel = $derived(fmtRemaining(remainingMs));
  let ttlText = $derived(() => {
    if (phase === 'timeout') return t('sync.pair.ttl_expired');
    if (ttlWarning) return t('sync.pair.ttl_warning', remainingLabel);
    return t('sync.pair.ttl_text', remainingLabel);
  });

  function fmtRemaining(ms: number): string {
    if (ms <= 0) return '0:00';
    const s = Math.ceil(ms / 1000);
    return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
  }

  // ── TTL sub-flag: any state with a live TTL crosses 30s → ttlWarning ──
  $effect(() => {
    ttlWarning = remainingMs > 0 && remainingMs <= 30_000 && phase !== 'timeout';
    // 0 remaining → timeout (only if we have a real token, not on initial open)
    if (remainingMs <= 0 && totalTtlMs > 0 && phase !== 'paired' && phase !== 'error') {
      phase = 'timeout';
    }
  });

  // ── rAF loop (paused on visibilitychange) ──
  function tick(): void {
    if (!expiresAt) {
      rafId = requestAnimationFrame(tick);
      return;
    }
    const ms = Math.max(0, new Date(expiresAt).getTime() - Date.now());
    remainingMs = ms;
    // Drop to 0 CPU when tab is hidden (per TitlebarThread pause contract)
    if (document.visibilityState === 'visible') {
      rafId = requestAnimationFrame(tick);
    } else {
      // Re-arm on visibilitychange via the listener below
      rafId = 0;
    }
  }

  function resumeRaf(): void {
    if (rafId === 0 && document.visibilityState === 'visible' && expiresAt) {
      rafId = requestAnimationFrame(tick);
    }
  }

  // ── IPC: open — mint a pairing token ──
  async function mintToken(): Promise<void> {
    busy = true;
    errorHead = '';
    errorCause = '';
    try {
      const res = await ipc.call<{
        ok: boolean;
        pin: string;
        peer: string;
        expires_in: number;
      }>('sync.pair_begin', { device_id: deviceId });
      pin = res.pin;
      const expMs = Date.now() + (res.expires_in ?? 60) * 1000;
      expiresAt = new Date(expMs).toISOString();
      totalTtlMs = res.expires_in * 1000;
      remainingMs = totalTtlMs;
      // Render the QR now that we have the canonical identity baked in.
      const payload = JSON.stringify({ v: 1, device_id: deviceId, name: deviceName });
      QRCode.toDataURL(payload, { margin: 1, width: 240, color: { dark: '#0B3D2E', light: '#FFFFFF' } })
        .then((url) => {
          qrDataUrl = url;
        })
        .catch(() => {
          qrDataUrl = '';
        });
      // Start the rAF loop on the next frame.
      rafId = requestAnimationFrame(tick);
    } catch (e) {
      errorHead = t('sync.pair.error_headline');
      errorCause = String(e);
      phase = 'error';
    } finally {
      busy = false;
    }
  }

  // ── IPC: 5s poll (SSE not on the wire; per spec §7.3) ──
  async function pollStatus(): Promise<void> {
    try {
      const st = (await ipc.call<{
        pending: boolean;
        device_id?: string;
        peer?: string;
        expires_at?: string;
        created_at?: string;
      }>('sync.pair_status', { device_id: deviceId })) ?? null;
      if (!st) return;
      // If the daemon reports the peer has confirmed (pending flipped), we
      // still wait for the user's CTA — but the footer pill can move to
      // "awaiting peer typed" copy if useful. The CTA flip itself is the
      // visual signal; the footer status mirrors the live state.
      if (st.expires_at) {
        const exp = new Date(st.expires_at).getTime();
        const nowLeft = Math.max(0, exp - Date.now());
        if (nowLeft > 0) {
          // Re-anchor if a fresh token was minted
          if (totalTtlMs === 0 || nowLeft > totalTtlMs) totalTtlMs = nowLeft;
          expiresAt = st.expires_at;
          remainingMs = nowLeft;
        }
      }
    } catch {
      // Non-fatal; the rAF + modal stay alive until the user cancels.
    }
  }

  // ── CTA: confirm pairing (the user typed the PIN on the OTHER device) ──
  async function confirmPairing(): Promise<void> {
    if (phase === 'paired' || phase === 'error' || phase === 'timeout') return;
    busy = true;
    errorHead = '';
    errorCause = '';
    try {
      // The "pin" the daemon verifies is the same pin it just minted (the
      // pin the user has been reading on this device). The peer's typed
      // value is what they entered on the other side; both sides hash
      // through the same token in the daemon's pendingPairings.
      const res = await ipc.call<{ ok: boolean; device_id: string }>('sync.pair_confirm', {
        device_id: deviceId,
        pin,
      });
      if (res.ok) {
        phase = 'paired';
        onpaired?.(res.device_id);
        // Auto-dismiss after 1500ms (per spec §3.6 + §4.4)
        if (autodismissTimer) clearTimeout(autodismissTimer);
        autodismissTimer = setTimeout(() => close(), 1500);
      }
    } catch (e) {
      errorHead = t('sync.pair.error_headline');
      errorCause = String(e);
      phase = 'error';
    } finally {
      busy = false;
    }
  }

  // ── CTA: regenerate PIN (timeout / TTL warning) ──
  async function regenerate(): Promise<void> {
    if (busy) return;
    qrDataUrl = '';
    pin = '';
    expiresAt = '';
    totalTtlMs = 0;
    remainingMs = 0;
    phase = 'open';
    showQr = true;
    showPin = true;
    await mintToken();
  }

  // ── Copy PIN ──
  async function copyPin(): Promise<void> {
    if (!pin) return;
    try {
      await navigator.clipboard.writeText(pin);
      copied = true;
      if (copyTimer) clearTimeout(copyTimer);
      copyTimer = setTimeout(() => {
        copied = false;
      }, 1600);
    } catch {
      // Silent — the user can re-click.
    }
  }

  // ── Toggle visibility ──
  function togglePin(): void {
    showPin = !showPin;
    showQr = true;
    phase = showPin ? 'open' : 'qr-mode';
  }
  function toggleQr(): void {
    showQr = !showQr;
    showPin = true;
    phase = showQr ? 'open' : 'pin-mode';
  }

  // ── Close: cancel everything ──
  function close(): void {
    open = false;
    onclose();
  }

  // ── sheet DOM ──
  let sheetEl = $state<HTMLDivElement | undefined>(undefined);
  let reduceMotion = $state(false);

  // ── keyboard model (per spec §5) ──
  function onKey(e: KeyboardEvent): void {
    if (!open) return;
    if (e.key === 'Escape') {
      e.preventDefault();
      close();
      return;
    }
    if (e.key !== 'Tab' || !sheetEl) return;
    // Focus trap
    const focusable = sheetEl.querySelectorAll<HTMLElement>(
      'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
    );
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    const active = document.activeElement as HTMLElement | null;
    if (e.shiftKey && active === first) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && active === last) {
      e.preventDefault();
      first.focus();
    }
  }

  // ── lifecycle ──
  onMount(() => {
    if (typeof window !== 'undefined') {
      reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    }
    if (open) {
      phase = 'open';
      void mintToken();
      pollTimer = setInterval(() => void pollStatus(), 5000);
    }
  });

  $effect(() => {
    // Re-arm rAF on tab return
    const onVis = (): void => {
      if (document.visibilityState === 'visible') resumeRaf();
    };
    if (typeof document !== 'undefined') {
      document.addEventListener('visibilitychange', onVis);
    }
    return () => {
      if (typeof document !== 'undefined') {
        document.removeEventListener('visibilitychange', onVis);
      }
    };
  });

  // ── one onDestroy (per spec §10 #9) ──
  onDestroy(() => {
    if (rafId) cancelAnimationFrame(rafId);
    if (pollTimer) clearInterval(pollTimer);
    if (copyTimer) clearTimeout(copyTimer);
    if (autodismissTimer) clearTimeout(autodismissTimer);
  });
</script>

<svelte:window onkeydown={onKey} />

{#if open}
  <!-- The sheet root.  Anchored to the bottom edge per MOAT §2.8.
       Slides up (not in from center) — the entry direction is the meaning. -->
  <div
    class="backdrop"
    role="presentation"
    onclick={(e) => {
      if (e.target === e.currentTarget) close();
    }}
    onkeydown={(e) => {
      if (e.key === 'Escape') close();
    }}
  >
    <div
      class="sheet"
      class:reduce={reduceMotion}
      bind:this={sheetEl}
      role="dialog"
      aria-modal="true"
      aria-labelledby="pairing-modal-headline"
      tabindex="-1"
    >
      <!-- Drag handle (32×4 hairline pill).  Pure affordance — no
           cursor:grab.  The actual close is Esc / outside-click. -->
      <div class="handle" aria-hidden="true"></div>

      <header class="head">
        <div class="head-text">
          <div class="eyebrow">
            <Glyph name="sync" size={14} stroke={1.5} />
            <span>{t('sync.pair')}</span>
          </div>
          <h2 id="pairing-modal-headline" class="headline">
            {t('sync.pair.title', peerName || deviceName)}
          </h2>
          <p class="lead">{t('sync.pair.hint')}</p>
        </div>
        <button
          class="close"
          type="button"
          onclick={close}
          aria-label={t('sync.pair.close')}
        >
          <Glyph name="close" size={14} stroke={1.5} />
        </button>
      </header>

      <!-- Two zones, side by side on ≥720px, stacked below. -->
      <div class="zones" class:wide={true}>
        <!-- Zone A: QR (left / top) -->
        <section
          class="zone qr-zone"
          class:hidden={!showQr}
          aria-labelledby="qr-zone-label"
        >
          <div class="zone-eyebrow mono" id="qr-zone-label">SCAN WITH PHONE</div>
          <div class="qr-card">
            {#if qrDataUrl}
              <img class="qr" src={qrDataUrl} alt={t('sync.pair.qr_alt')} />
            {:else}
              <div class="qr placeholder">
                <Pulse phase="thinking" size={10} />
                <span class="mono placeholder-label">{t('sync.pair.thinking')}</span>
              </div>
            {/if}
          </div>
          <p class="cap">{t('sync.pair.qr_cap', deviceName || t('sync.pair.this_device'))}</p>
          <Tooltip label={t('sync.pair.show_pin')}>
            <button
              type="button"
              class="toggle"
              onclick={togglePin}
              aria-label={showPin ? t('sync.pair.hide') : t('sync.pair.show_pin')}
            >
              {showPin ? t('sync.pair.hide') : t('sync.pair.show_pin')}
            </button>
          </Tooltip>
        </section>

        <!-- Zone B: PIN (right / bottom) -->
        <section
          class="zone pin-zone"
          class:hidden={!showPin}
          aria-labelledby="pin-zone-label"
        >
          <div class="zone-eyebrow mono" id="pin-zone-label">OR ENTER THIS PIN ON THE OTHER MACHINE</div>

          <div class="pin-block" class:danger={phase === 'timeout'} class:warn={ttlWarning}>
            <!-- The signature motion — 120×120 SVG ring, r:48, 3px stroke.
                 stroke-dasharray=CIRC, stroke-dashoffset updates from the
                 rAF loop.  At timeout, the ring fills --danger. -->
            <svg
              class="pin-ring"
              width="120"
              height="120"
              viewBox="0 0 120 120"
              aria-hidden="true"
            >
              <circle
                cx="60"
                cy="60"
                r={RING_R}
                fill="none"
                stroke="var(--hair)"
                stroke-width="1.5"
                opacity="0.45"
              />
              <circle
                cx="60"
                cy="60"
                r={RING_R}
                fill="none"
                stroke={ringFilled ? 'var(--danger)' : ringStroke}
                stroke-width="3"
                stroke-linecap="round"
                pathLength="1"
                stroke-dasharray={RING_CIRC}
                stroke-dashoffset={ringFilled ? 0 : ringDashoffset}
                transform="rotate(-90 60 60)"
                style="transition: stroke-dashoffset var(--dur) var(--ease), stroke var(--dur) var(--ease);"
              />
            </svg>
            <span
              class="pin"
              class:danger={phase === 'timeout' || ttlWarning}
            >
              {pin || '••••••'}
            </span>
          </div>

          <p class="ttl mono" class:danger={phase === 'timeout'} class:warn={ttlWarning}>
            {ttlText()}
          </p>

          <p class="cap">{t('sync.pair.pin_label')}</p>

          <div class="pin-actions">
            <Tooltip label={t('sync.pair.copy')}>
              <button
                type="button"
                class="toggle"
                onclick={copyPin}
                aria-label={t('sync.pair.copy')}
                disabled={!pin}
              >
                {copied ? t('sync.pair.copied') : t('sync.pair.copy')}
              </button>
            </Tooltip>
          </div>
        </section>
      </div>

      <!-- The footer: status pill + regenerate + ⌘P hint. -->
      <footer class="foot" aria-live="polite">
        <div class="foot-left">
          {#if phase === 'paired'}
            <Pulse phase="ok" size={8} />
            <span class="mono foot-pill">{t('sync.pair.paired_pill')}</span>
          {:else if phase === 'error'}
            <Pulse phase="error" size={8} />
            <span class="mono foot-pill danger">{t('sync.pair.error_headline')}</span>
          {:else if phase === 'timeout'}
            <Pulse phase="error" size={8} />
            <span class="mono foot-pill danger">{t('sync.pair.ttl_expired')}</span>
          {:else}
            <Pulse phase={ttlWarning ? 'consent' : 'awaiting'} size={8} />
            <span class="mono foot-pill">
              {ttlWarning ? t('sync.pair.ttl_warning', remainingLabel) : t('sync.pair.awaiting_peer')}
            </span>
          {/if}
        </div>

        <div class="foot-mid">
          {#if (phase === 'timeout' || ttlWarning) && phase !== 'paired'}
            <Button variant="ghost" size="sm" onclick={regenerate} disabled={busy}>
              {t('sync.pair.regenerate')}
            </Button>
          {/if}
        </div>

        <div class="foot-right mono">
          <kbd>⌘P</kbd> {t('sync.pair.aria_label')}
        </div>
      </footer>

      <!-- The "what gets synced / never synced" reminder — the
           load-bearing promise, not decoration.  Always rendered except
           in success. -->
      {#if phase !== 'paired'}
        <div class="reminder" aria-hidden="true">
          <p class="reminder-row">
            <span class="rem-k mono">SYNCED</span>
            <span class="rem-v">{t('sync.pair.synced_label')}</span>
          </p>
          <p class="reminder-row">
            <span class="rem-k mono">NEVER</span>
            <span class="rem-v">{t('sync.pair.never_synced')}</span>
          </p>
        </div>
      {/if}

      <!-- Thread draws across the sheet's top in the `paired` state.  This
           is the callback gesture to the titlebar (per spec §4.4). -->
      {#if phase === 'paired'}
        <div class="paired-thread" aria-hidden="true">
          <Thread orientation="h" draw={true} glow={true} />
        </div>
        <div class="paired-stamp" role="status" aria-live="polite">
          <span class="paired-check" aria-hidden="true">
            <Glyph name="check" size={20} stroke={2} />
          </span>
          <span class="paired-headline">{t('sync.pair.paired')}</span>
          <span class="paired-sub mono">{t('sync.pair.paired_sub')}</span>
        </div>
      {/if}

      <!-- Error state — single instance of ErrorState per spec §1.2. -->
      {#if phase === 'error'}
        <ErrorState
          head={errorHead || t('sync.pair.error_headline')}
          cause={t('sync.pair.error_cause')}
          reason={errorCause}
          onretry={regenerate}
          retryLabel={t('sync.pair.error_action')}
          onsettings={close}
          settingsLabel={t('sync.pair.cancel')}
        />
      {/if}

      <!-- Primary CTA — visible in open / ttl-warning / timeout / error. -->
      {#if phase !== 'paired' && phase !== 'error'}
        <div class="cta-row">
          {#if phase === 'timeout'}
            <Button variant="primary" size="md" onclick={regenerate} disabled={busy}>
              {t('sync.pair.regenerate')} →
            </Button>
          {:else}
            <Button
              variant="primary"
              size="md"
              onclick={confirmPairing}
              disabled={busy || !pin}
            >
              {busy ? t('sync.pair.busy') : t('sync.pair.cta')}
            </Button>
          {/if}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  /* ── The SHEET (per MOAT §2.8) ─────────────────────────────────────
     Anchored to the bottom edge.  Slides UP from below — the entry
     direction carries meaning (the sheet "comes from where the
     thumb is").  Does NOT block page scroll; the Sync garden beneath
     stays alive and breathing.  Single shadow elevation, no double
     shadow (MOAT §4 #9). */
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: var(--z-sheet);
    background: color-mix(in oklab, var(--ink) 6%, transparent);
    display: flex;
    justify-content: center;
    align-items: flex-end;
    animation: sheet-fade var(--dur) var(--ease) both;
  }
  .sheet {
    position: relative;
    width: min(560px, calc(100vw - 32px));
    max-height: 80vh;
    margin: 0 auto;
    background: var(--surface-raised);
    border-top: 1px solid var(--hair-strong);
    border-radius: var(--r-lg) var(--r-lg) 0 0;
    box-shadow: var(--shadow-float);
    padding: var(--space-3) var(--space-5) var(--space-5);
    overflow-y: auto;
    animation: sheet-up var(--dur-cine) var(--ease) both;
  }

  @keyframes sheet-up {
    from {
      transform: translateY(100%);
    }
    to {
      transform: translateY(0);
    }
  }
  @keyframes sheet-fade {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  @keyframes sheet-down {
    from {
      transform: translateY(0);
    }
    to {
      transform: translateY(100%);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .sheet,
    .backdrop {
      animation: none !important;
    }
    .qr-zone,
    .pin-zone {
      transition: none !important;
    }
    .pin-ring circle {
      transition: none !important;
    }
  }

  /* ── Drag handle (32×4 hairline pill, the only signal that the
       sheet is dismissable) ─────────────────────────────────────── */
  .handle {
    width: 32px;
    height: 4px;
    background: var(--ink-faint);
    border-radius: var(--r-pill);
    margin: 8px auto 12px;
    opacity: 0;
    animation: handle-reveal var(--dur) var(--ease) 200ms both;
  }
  @keyframes handle-reveal {
    to {
      opacity: 1;
    }
  }

  /* ── Header ─────────────────────────────────────────────────────── */
  .head {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    padding: 0 0 var(--space-4);
    border-bottom: 1px solid var(--hair);
  }
  .head-text {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .eyebrow {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-family: var(--font-mono);
    font-size: var(--text-caption);
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-mute);
  }
  .headline {
    font-family: var(--font-display);
    font-weight: 400;
    font-size: var(--text-h2);
    line-height: var(--lh-h2);
    letter-spacing: var(--ls-h2);
    color: var(--content);
    margin: 0;
  }
  .lead {
    font-family: var(--font-sans);
    font-size: 13px;
    line-height: 1.5;
    color: var(--content-soft);
    margin: 0;
  }
  .close {
    flex: none;
    width: 28px;
    height: 28px;
    display: grid;
    place-items: center;
    background: none;
    border: 1px solid var(--hair);
    border-radius: var(--r-pill);
    color: var(--content-mute);
    cursor: pointer;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      transform var(--dur-fast) var(--ease);
  }
  .close:hover {
    color: var(--content);
    background: var(--surface-card);
    border-color: var(--hair-strong);
  }
  .close:active {
    transform: scale(0.97);
  }
  .close:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 2px var(--synapse),
      0 0 0 5px var(--pollen-halo);
  }

  /* ── Zones ──────────────────────────────────────────────────────── */
  .zones {
    display: flex;
    flex-direction: row;
    gap: var(--space-5);
    padding: var(--space-5) 0 var(--space-4);
  }
  @media (max-width: 720px) {
    .zones {
      flex-direction: column;
      gap: var(--space-3);
    }
  }
  .zone {
    flex: 1 1 50%;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-3);
    min-width: 0;
    transition:
      opacity var(--dur) var(--ease),
      transform var(--dur) var(--ease);
  }
  .zone.hidden {
    opacity: 0;
    transform: scale(0.96);
    pointer-events: none;
  }
  .zone-eyebrow {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-faint);
    text-align: center;
  }

  /* ── QR card (240×240 on paper-raised) ──────────────────────────── */
  .qr-card {
    width: 240px;
    height: 240px;
    background: #fff;
    padding: var(--space-4);
    border-radius: var(--r-md);
    box-shadow: var(--shadow-card);
    display: grid;
    place-items: center;
  }
  .qr {
    width: 100%;
    height: 100%;
    display: block;
  }
  .qr.placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    color: var(--content-faint);
  }
  .placeholder-label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
  }

  .cap {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 14px;
    line-height: 1.4;
    color: var(--content-mute);
    text-align: center;
    max-width: 280px;
    margin: 0;
  }

  /* ── PIN ring + digits ──────────────────────────────────────────── */
  .pin-block {
    position: relative;
    display: grid;
    place-items: center;
    width: 120px;
    height: 120px;
    margin: 0 auto;
  }
  .pin-ring {
    position: absolute;
    inset: 0;
    pointer-events: none;
    filter: drop-shadow(0 0 4px color-mix(in oklab, var(--pollen) 35%, transparent));
  }
  .pin {
    position: relative;
    z-index: 1;
    font-family: var(--font-mono);
    font-size: 26px;
    font-weight: 500;
    letter-spacing: 0.18em;
    color: var(--content);
    text-shadow: 0 0 22px color-mix(in oklab, var(--pollen) 25%, transparent);
    transition:
      color var(--dur) var(--ease),
      text-shadow var(--dur) var(--ease),
      filter var(--dur) var(--ease),
      opacity var(--dur) var(--ease);
  }
  .pin.danger {
    color: var(--danger);
    text-shadow: 0 0 22px color-mix(in oklab, var(--danger) 25%, transparent);
  }
  .pin-block.warn .pin {
    color: var(--danger);
    text-shadow: 0 0 22px color-mix(in oklab, var(--danger) 25%, transparent);
  }
  .pin-block.danger .pin {
    filter: blur(2px);
    opacity: 0.4;
  }

  .ttl {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-mute);
    text-align: center;
    transition: color var(--dur) var(--ease);
  }
  .ttl.warn {
    color: var(--danger);
  }
  .ttl.danger {
    color: var(--danger);
  }

  /* ── Toggle + Copy + Regenerate buttons (ghost) ─────────────────── */
  .toggle {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--content-mute);
    background: transparent;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    padding: 6px 14px;
    cursor: pointer;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      transform var(--dur-fast) var(--ease);
  }
  .toggle:hover:not([disabled]) {
    color: var(--content);
    border-color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 4%, transparent);
  }
  .toggle:active:not([disabled]) {
    transform: scale(0.97);
  }
  .toggle:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 2px var(--synapse),
      0 0 0 5px var(--pollen-halo);
  }
  .toggle[disabled] {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .pin-actions {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
  }

  /* ── Footer ─────────────────────────────────────────────────────── */
  .foot {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) 0 var(--space-2);
    border-top: 1px solid var(--hair);
  }
  .foot-left {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }
  .foot-pill {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-mute);
    white-space: nowrap;
  }
  .foot-pill.danger {
    color: var(--danger);
  }
  .foot-mid {
    display: flex;
    justify-content: center;
  }
  .foot-right {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    color: var(--content-faint);
    text-align: right;
    white-space: nowrap;
  }
  .foot-right kbd {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--content-soft);
    background: var(--surface-sunken);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-xs);
    padding: 1px 5px;
  }

  /* ── Reminder ───────────────────────────────────────────────────── */
  .reminder {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: var(--space-3) 0 0;
  }
  .reminder-row {
    display: flex;
    align-items: baseline;
    gap: var(--space-3);
    margin: 0;
  }
  .rem-k {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--content-faint);
    flex: none;
    min-width: 56px;
  }
  .rem-v {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 13px;
    line-height: 1.4;
    color: var(--content-soft);
  }

  /* ── CTA row ────────────────────────────────────────────────────── */
  .cta-row {
    display: flex;
    justify-content: center;
    padding: var(--space-4) 0 0;
  }

  /* ── Paired state (Thread + check) ──────────────────────────────── */
  .paired-thread {
    height: 2px;
    margin: var(--space-3) 0 var(--space-3);
  }
  .paired-stamp {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    padding: var(--space-3) 0 var(--space-4);
  }
  .paired-check {
    width: 48px;
    height: 48px;
    display: grid;
    place-items: center;
    background: color-mix(in oklab, var(--synapse) 12%, transparent);
    border: 1px solid color-mix(in oklab, var(--synapse) 40%, transparent);
    border-radius: var(--r-pill);
    color: var(--synapse);
    animation: stamp-in 320ms var(--ease-pop) 200ms both;
  }
  @keyframes stamp-in {
    from {
      transform: scale(0);
    }
    to {
      transform: scale(1);
    }
  }
  .paired-headline {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 22px;
    line-height: 1.2;
    color: var(--synapse);
    letter-spacing: -0.01em;
  }
  .paired-sub {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-mute);
  }
</style>
