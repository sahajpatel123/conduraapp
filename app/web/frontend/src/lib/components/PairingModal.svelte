<script lang="ts">
  // PairingModal (Phase 14F). Replaces the old window.prompt() flow.
  // Shows this device's identity as a QR code (so the other device
  // can scan it), the 6-digit PIN minted by sync.pair_begin, and a
  // confirmation input. The QR encodes a JSON identity payload.
  import { onMount } from 'svelte'
  import QRCode from 'qrcode'
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

  let qrDataUrl = $state('')
  let entered = $state('')
  let remaining = $state('')

  onMount(() => {
    const payload = JSON.stringify({ v: 1, device_id: deviceId, name: deviceName })
    QRCode.toDataURL(payload, { margin: 1, width: 220 })
      .then((url) => {
        qrDataUrl = url
      })
      .catch(() => {
        qrDataUrl = ''
      })

    let timer: ReturnType<typeof setInterval> | null = null
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
      timer = setInterval(tick, 1000)
    }
    return () => {
      if (timer) clearInterval(timer)
    }
  })

  const confirmReady = $derived(/^\d{4,8}$/.test(entered.trim()))
</script>

<div class="pair-backdrop" role="presentation" onclick={onCancel}>
  <div
    class="pair-modal glass-card elevated"
    role="dialog"
    aria-modal="true"
    aria-label={t('sync.pair.aria_label')}
    tabindex="-1"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => { if (e.key === 'Escape') onCancel() }}
  >
    <header>
      <h2>{t('sync.pair.title', peerName)}</h2>
      <button class="close" aria-label={t('sync.pair.close')} onclick={onCancel}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M18 6L6 18M6 6l12 12" /></svg>
      </button>
    </header>

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
          {remaining === t('sync.pair.expired') ? t('sync.pair.expired') : t('sync.pair.expires_in', remaining)}
        </span>
      {/if}
    </div>

    <div class="confirm">
      <label for="pair-pin">{t('sync.pair.confirm_label', peerName)}</label>
      <div class="confirm-row">
        <input
          id="pair-pin"
          class="input pin-input"
          type="text"
          inputmode="numeric"
          bind:value={entered}
          placeholder="000000"
          maxlength="8"
          onkeydown={(e) => { if (e.key === 'Enter' && confirmReady) onConfirm(entered.trim()) }}
        />
        <button class="btn btn-primary" disabled={!confirmReady || busy} onclick={() => onConfirm(entered.trim())}>
          {busy ? t('sync.pair.busy') : t('sync.pair.confirm')}
        </button>
      </div>
    </div>

    {#if error}
      <p class="err">{error}</p>
    {/if}
  </div>
</div>

<style>
  .pair-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    padding: var(--space-4);
    animation: backdrop-in var(--transition-base) ease both;
  }
  .pair-modal {
    width: 100%;
    max-width: 380px;
    padding: var(--space-5);
    text-align: center;
    animation: modal-in var(--transition-spring) var(--ease-out-expo) both;
  }
  .pair-modal:hover {
    border-color: var(--glass-border-hover);
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-3);
  }
  h2 {
    font-size: var(--size-lg);
    font-weight: var(--weight-semibold);
  }
  .close {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    background: none;
    border: none;
    color: var(--color-text-faint);
    cursor: pointer;
    border-radius: var(--radius-sm);
    transition: color var(--transition-base), background var(--transition-base);
  }
  .close svg { width: 16px; height: 16px; }
  .close:hover {
    color: var(--color-text);
    background: var(--glass-bg-hover);
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
    color: var(--color-text-muted);
    font-size: var(--size-xs);
    line-height: var(--leading-relaxed);
    max-width: 280px;
  }
  .pin-block {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    margin: var(--space-4) 0;
  }
  .pin-label {
    font-size: var(--size-xs);
    color: var(--color-text-faint);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
  }
  .pin {
    font-family: var(--font-mono);
    font-size: var(--size-2xl);
    font-weight: var(--weight-bold);
    letter-spacing: 0.2em;
    color: var(--color-accent);
    text-shadow: 0 0 18px var(--color-glow-strong);
  }
  .ttl {
    font-size: var(--size-xs);
    color: var(--color-text-muted);
  }
  .ttl.expired { color: var(--color-error); }
  .confirm label {
    display: block;
    font-size: var(--size-sm);
    color: var(--color-text-muted);
    margin-bottom: var(--space-2);
  }
  .confirm-row {
    display: flex;
    gap: var(--space-2);
  }
  .pin-input {
    flex: 1;
    font-family: var(--font-mono);
    font-size: var(--size-lg);
    text-align: center;
    letter-spacing: 0.15em;
  }
  .err {
    color: var(--color-error);
    font-size: var(--size-sm);
    margin-top: var(--space-3);
  }
</style>
