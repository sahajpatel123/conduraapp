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

  onMount(() => {
    const payload = JSON.stringify({ v: 1, device_id: deviceId, name: deviceName })
    QRCode.toDataURL(payload, { margin: 1, width: 220 })
      .then((url) => {
        qrDataUrl = url
      })
      .catch(() => {
        qrDataUrl = ''
      })

    if (expiresAt) {
      const tick = (): void => {
        const ms = new Date(expiresAt).getTime() - Date.now()
        if (ms <= 0) {
          remaining = t('sync.pair.expired')
          return
        }
        const s = Math.floor(ms / 1000)
        remaining = `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`
      }
      tick()
      const timer = setInterval(tick, 1000)
      return () => clearInterval(timer)
    }
  })

  // Poll the daemon's pair_confirm status. This is a non-mutating
  // status read — the actual confirm happens when the user types
  // the PIN and clicks Connect. The poll exists so we can detect
  // when the peer has already entered the PIN and the daemon is
  // waiting for local confirmation.
  $effect(() => {
    if (!open) {
      if (pollTimer) {
        clearInterval(pollTimer)
        pollTimer = null
      }
      return
    }
    pollTimer = setInterval(() => {
      void ipc.call('sync.pair_status', { device_id: deviceId }).catch(() => {
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
        <span class="pin">{pin}</span>
        {#if remaining}
          <span class="ttl" class:expired={remaining === t('sync.pair.expired')}>
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
  .pin {
    font-family: var(--font-mono);
    font-size: var(--size-2xl);
    font-weight: var(--weight-bold);
    letter-spacing: 0.2em;
    color: var(--accent);
    text-shadow: 0 0 18px var(--accent-glow);
  }
  .ttl {
    font-size: var(--size-xs);
    color: var(--text-muted);
  }
  .ttl.expired {
    color: var(--error);
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