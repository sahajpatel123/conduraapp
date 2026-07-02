<script lang="ts">
  // PairingModal (Phase 14F). Shows this device's identity as a QR
  // code, a 6-digit PIN with TTL countdown, and a confirmation
  // input. Polls the daemon's sync.pair_confirm status while the
  // modal is open so the peer-typed PIN can be confirmed even
  // before the user clicks the local Connect button.
  import { onMount, onDestroy } from 'svelte'
  import QRCode from 'qrcode'
  import { Sheet } from './ui'
  import Button from './ui/Button.svelte'
  import { ipc } from '../ipc/client'
  import { t } from '../i18n'

  interface Props {
    // This device's identity (shown as a QR for the peer to scan).
    deviceId: string
    deviceName: string
    // The peer we're pairing with + the PIN minted for it.
    peerName: string
    pin: string
    // Seconds remaining before the pairing token expires (optional).
    expiresAt?: string
    onConfirm: (pin: string) => void | Promise<void>
    onCancel: () => void
    busy?: boolean
    error?: string | null
  }
  let {
    deviceId,
    deviceName,
    peerName,
    pin,
    expiresAt,
    onConfirm,
    onCancel,
    busy = false,
    error = null,
  }: Props = $props()

  let open = $state(true)
  let qrDataUrl = $state('')
  let entered = $state('')
  let remaining = $state('')
  let pollTimer: ReturnType<typeof setInterval> | null = null

  // ── TTL ring state ──
  // sync.pair_status returns { pending, device_id, peer, expires_at, created_at }.
  // We track secondsLeft (updated every 1s for smooth animation) and render an
  // SVG ring around the PIN that depletes as the token expires. When pending
  // becomes false (peer confirmed or token revoked), the ring unmounts.
  type PairStatus = {
    pending: boolean
    device_id?: string
    peer?: string
    expires_at?: string
    created_at?: string
  }
  let pairStatus = $state<PairStatus | null>(null)
  let secondsLeft = $state(0)
  // The initial TTL is derived from the prop's expiresAt (the first mint). If
  // the poll later reports a different expires_at, we re-anchor initialTtl so
  // the ring never runs backwards.
  let initialTtl = $state(0)
  let ringTimer: ReturnType<typeof setInterval> | null = null

  const RING_RADIUS = 32
  const RING_CIRC = 2 * Math.PI * RING_RADIUS // ≈ 201.06
  let ringOffset = $derived(
    initialTtl > 0 ? RING_CIRC * (1 - secondsLeft / initialTtl) : RING_CIRC
  )
  let ringDanger = $derived(secondsLeft > 0 && secondsLeft < 30)
  let ringVisible = $derived(
    pairStatus?.pending === true && secondsLeft > 0 && initialTtl > 0
  )

  function recomputeSecondsLeft(exp?: string): void {
    if (!exp) {
      secondsLeft = 0
      return
    }
    const ms = new Date(exp).getTime() - Date.now()
    secondsLeft = ms > 0 ? Math.floor(ms / 1000) : 0
  }

  onMount(() => {
    const payload = JSON.stringify({ v: 1, device_id: deviceId, name: deviceName })
    QRCode.toDataURL(payload, { margin: 1, width: 220 })
      .then((url) => {
        qrDataUrl = url
      })
      .catch(() => {
        qrDataUrl = ''
      })

    // Seed the ring from the prop's expiresAt (the first mint) so the user
    // sees a countdown immediately, before the first poll lands.
    if (expiresAt) {
      const ms = new Date(expiresAt).getTime() - Date.now()
      if (ms > 0) {
        initialTtl = Math.floor(ms / 1000)
        recomputeSecondsLeft(expiresAt)
      }
    }

    // 1s interval for smooth ring depletion + remaining-text refresh.
    ringTimer = setInterval(() => {
      const exp = pairStatus?.expires_at ?? expiresAt
      recomputeSecondsLeft(exp)
      const s = secondsLeft
      if (s <= 0) {
        remaining = t('sync.pair.expired')
      } else {
        remaining = `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`
      }
    }, 1000)
  })

  onDestroy(() => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
    if (ringTimer) {
      clearInterval(ringTimer)
      ringTimer = null
    }
  })

  // Poll the daemon's pair_confirm status. This is a non-mutating
  // status read — the actual confirm happens when the user types
  // the PIN and clicks Connect. The poll exists so we can detect
  // when the peer has already entered the PIN and the daemon is
  // waiting for local confirmation, and to refresh expires_at.
  $effect(() => {
    if (!open) {
      if (pollTimer) {
        clearInterval(pollTimer)
        pollTimer = null
      }
      return
    }
    pollTimer = setInterval(() => {
      void ipc
        .call('sync.pair_status', { device_id: deviceId })
        .then((res) => {
          const st = res as PairStatus
          if (!st || typeof st !== 'object') return
          pairStatus = st
          // Re-anchor initialTtl when a fresh token is minted (created_at
          // changed or expires_at moved past the previous one).
          if (st.expires_at) {
            const nowLeft = Math.floor((new Date(st.expires_at).getTime() - Date.now()) / 1000)
            if (nowLeft > 0) {
              // If we never seeded, or the server reports a longer window than
              // we had left, re-seed initialTtl so the ring stays proportional.
              if (initialTtl === 0 || nowLeft > initialTtl) {
                initialTtl = nowLeft
              }
              recomputeSecondsLeft(st.expires_at)
            }
          }
        })
        .catch(() => {
          // Polling errors are non-fatal; the modal stays open until
          // the user explicitly confirms or cancels.
        })
    }, 5000)
    return () => {
      if (pollTimer) {
        clearInterval(pollTimer)
        pollTimer = null
      }
    }
  })

  onDestroy(() => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  })

  const confirmReady = $derived(/^\d{4,8}$/.test(entered.trim()))

  function close(): void {
    open = false
    onCancel()
  }

  function confirm(): void {
    if (!confirmReady) return
    onConfirm(entered.trim())
  }
</script>

<Sheet
  bind:open
  side="right"
  width="420px"
  title={t('sync.pair.title', peerName)}
  onclose={close}
>
  {#snippet children()}
    <div class="pair-body">
      <div class="qr-area">
        {#if qrDataUrl}
          <img class="qr" src={qrDataUrl} alt={t('sync.pair.qr_alt')} />
        {:else}
          <div class="qr placeholder">QR</div>
        {/if}
        <p class="qr-cap">
          {t('sync.pair.qr_cap', deviceName || t('sync.pair.this_device'))}
        </p>
      </div>

      <div class="pin-block">
        <span class="pin-label">{t('sync.pair.pin_label')}</span>
        <div class="pin-ring-wrap" class:danger={ringDanger}>
          {#if ringVisible}
            <svg
              class="pin-ring"
              width="84"
              height="84"
              viewBox="0 0 84 84"
              aria-hidden="true"
            >
              <circle
                cx="42"
                cy="42"
                r={RING_RADIUS}
                fill="none"
                stroke="var(--pollen-halo, rgba(201,123,46,0.18))"
                stroke-width="4"
                opacity="0.45"
              />
              <circle
                cx="42"
                cy="42"
                r={RING_RADIUS}
                fill="none"
                stroke={ringDanger ? 'var(--danger)' : 'var(--pollen)'}
                stroke-width="4"
                stroke-linecap="round"
                stroke-dasharray={RING_CIRC}
                stroke-dashoffset={ringOffset}
                transform="rotate(-90 42 42)"
                style="transition: stroke-dashoffset 950ms linear, stroke 300ms ease;"
              />
            </svg>
          {/if}
          <span class="pin" class:danger={ringDanger}>{pin}</span>
        </div>
        {#if remaining}
          <span class="ttl" class:expired={remaining === t('sync.pair.expired')} class:danger={ringDanger}>
            {remaining === t('sync.pair.expired')
              ? t('sync.pair.expired')
              : t('sync.pair.expires_in', remaining)}
          </span>
        {/if}
      </div>

      <div class="confirm">
        <label for="pair-pin">{t('sync.pair.confirm_label', peerName)}</label>
        <div class="confirm-row">
          <input
            id="pair-pin"
            class="pin-input"
            type="text"
            inputmode="numeric"
            bind:value={entered}
            placeholder="000000"
            maxlength="8"
            onkeydown={(e) => {
              if (e.key === 'Enter' && confirmReady) confirm()
            }}
          />
          <Button
            variant="primary"
            disabled={!confirmReady || busy}
            loading={busy}
            onclick={confirm}
          >
            {busy ? t('sync.pair.busy') : t('sync.pair.confirm')}
          </Button>
        </div>
      </div>

      {#if error}
        <p class="err">{error}</p>
      {/if}
    </div>
  {/snippet}
</Sheet>

<style>
  .pair-body {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    text-align: center;
  }

  .qr-area {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
  }
  .qr {
    width: 200px;
    height: 200px;
    border-radius: var(--radius-md);
    background: #fff;
    padding: 8px;
    box-shadow: var(--shadow-glow);
  }
  .qr.placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #999;
    font-family: var(--font-mono);
  }
  .qr-cap {
    color: var(--text-muted);
    font-size: var(--size-xs);
    line-height: var(--leading-relaxed);
    max-width: 280px;
  }

  .pin-block {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-3) 0;
    border-top: 1px solid var(--border);
    border-bottom: 1px solid var(--border);
  }
  .pin-label {
    font-size: var(--size-xs);
    color: var(--text-faint);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
  }
  .pin-ring-wrap {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 84px;
    height: 84px;
    margin: var(--space-1) 0;
  }
  .pin-ring {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }
  .pin {
    font-family: var(--font-mono);
    font-size: var(--size-2xl);
    font-weight: var(--weight-bold);
    letter-spacing: 0.2em;
    color: var(--accent);
    text-shadow: 0 0 18px var(--accent-glow);
    transition: color 300ms ease, text-shadow 300ms ease;
  }
  .pin.danger,
  .pin-ring-wrap.danger .pin {
    color: var(--danger);
    text-shadow: 0 0 18px color-mix(in srgb, var(--danger) 35%, transparent);
  }
  .ttl {
    font-size: var(--size-xs);
    color: var(--text-muted);
    transition: color 300ms ease;
  }
  .ttl.expired {
    color: var(--error);
  }
  .ttl.danger {
    color: var(--danger);
  }

  .confirm label {
    display: block;
    font-size: var(--size-sm);
    color: var(--text-muted);
    margin-bottom: var(--space-2);
    text-align: left;
  }
  .confirm-row {
    display: flex;
    gap: var(--space-2);
    align-items: stretch;
  }
  .pin-input {
    flex: 1;
    min-width: 0;
    background: var(--surface-1);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    color: var(--text);
    font-family: var(--font-mono);
    font-size: var(--size-lg);
    text-align: center;
    letter-spacing: 0.15em;
    padding: 0 var(--space-3);
    height: 36px;
    transition:
      border-color var(--transition-fast) ease,
      background-color var(--transition-fast) ease,
      box-shadow var(--transition-fast) ease;
  }
  .pin-input::placeholder {
    color: var(--text-faint);
  }
  .pin-input:focus {
    outline: none;
    border-color: var(--border-focus);
    background: var(--surface-2);
    box-shadow: 0 0 0 3px var(--accent-soft);
  }

  .err {
    color: var(--error);
    font-size: var(--size-sm);
  }
</style>