<!--
  Sync — Condura v2 P2P device-pairing surface.

  Per the spec: "Two devices meeting in a quiet room. Pairing is a
  'handshake' — your device is on the left, theirs on the right,
  and a curved line animates between them." and "The connecting line
  draws in --dur-cinematic, then settles."

  Composition:
    - Two device cards (mine left, theirs right)
    - A curved SVG line that animates between them on mount
    - A QR code + 6-digit PIN with TTL countdown
    - Once paired, the line settles and the cards reveal paired state

  Pure presentation. Parent owns the QR payload, the PIN, the peer
  device info, and the paired-state flag.

  Props:
    myId:         string — this device's identity (e.g. "alex-mbp")
    myName:       string
    peerName?:    string  — once known
    qrPayload:    string  — base64 QR for the other device to scan
    pin:          string  — 6-digit PIN
    ttlSeconds:   number  — countdown in seconds
    paired?:      boolean
    onPinChange?: (pin: string) => void — fires if user re-rolls
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Eyebrow, Glyph, Button } from '$lib/v2'

  let {
    myId = 'this-device',
    myName = 'This device',
    peerName = undefined as string | undefined,
    qrPayload = '' as string,
    pin = '000000',
    ttlSeconds = 60 as number,
    paired = false as boolean,
    onPinChange = undefined as (() => void) | undefined,
  }: {
    myId?: string
    myName?: string
    peerName?: string
    qrPayload?: string
    pin?: string
    ttlSeconds?: number
    paired?: boolean
    onPinChange?: () => void
  } = $props()

  // TTL countdown — real-time, decays each second.
  let ttl = $state(ttlSeconds)
  $effect(() => {
    if (paired) return
    const id = setInterval(() => {
      ttl = ttl - 1
      if (ttl <= 0) ttl = 0
    }, 1000)
    return () => clearInterval(id)
  })
  const fmtTtl = $derived(`${Math.floor(ttl / 60)}:${String(ttl % 60).padStart(2, '0')}`)
</script>

<div data-v2 style="
  flex: 1;
  background: var(--v2-paper);
  overflow-y: auto;
  padding: var(--v2-space-12);
  box-sizing: border-box;
">
  <div style="max-width: 960px; margin: 0 auto;">

    <Stack gap={10}>

      <!-- ── Title ─────────────────────────────────── -->
      <Stack gap={3}>
        <Eyebrow left="sync" right="pairing" tone="accent" />
        <Ink kind="display" as="h1">Two devices, one quiet room.</Ink>
        <Ink kind="body-2" tone="ink-2" as="p" style:max-width="640px">
          Pair a phone or laptop with condura over an encrypted
          peer-to-peer handshake. No central server. No account needed.
        </Ink>
      </Stack>

      <!-- ── The two-device meeting ──────────────────── -->
      <Surface elevation={2} padding="12" radius="3" tone="paper">
        <div style="display: grid; grid-template-columns: 1fr 240px 1fr; align-items: center; gap: var(--v2-space-8);">

          <!-- My device (left) -->
          <Stack gap={3} align="center">
            <div style="
              width: 64px; height: 64px;
              border-radius: var(--v2-radius-3);
              background: var(--v2-paper-2);
              border: 1px solid var(--v2-rule);
              display: grid; place-items: center;
            ">
              <Glyph name="gear" size={28} />
            </div>
            <Stack gap={1} align="center">
              <Ink kind="ui" weight="medium">{myName}</Ink>
              <Ink kind="mono-cap" tone="ink-3">{myId}</Ink>
            </Stack>
          </Stack>

          <!-- The connecting line (drawn as an SVG path that animates) -->
          <div style="
            position: relative;
            height: 80px;
            display: grid; place-items: center;
          ">
            <svg viewBox="0 0 240 80" width="100%" height="80" style="overflow: visible;">
              <path
                d="M 10 40 Q 120 -10 230 40"
                fill="none"
                stroke="var(--v2-accent)"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-dasharray="320"
                stroke-dashoffset={paired ? '0' : '320'}
                style="
                  transition: stroke-dashoffset var(--v2-dur-cinematic) var(--v2-ease-settle);
                "
              />
              <!-- Pulse node — a small dot at each end of the line -->
              <circle cx="10" cy="40" r="3" fill="var(--v2-accent)" />
              <circle cx="230" cy="40" r="3" fill="var(--v2-accent)" />
              {#if paired}
                <!-- The settled state shows a small checkmark at the midpoint -->
                <circle cx="120" cy="14" r="10" fill="var(--v2-paper)" stroke="var(--v2-signal-go)" stroke-width="1.5" />
                <Glyph name="check" size={12} />
              {/if}
            </svg>
          </div>

          <!-- Peer device (right) -->
          <Stack gap={3} align="center">
            <div style="
              width: 64px; height: 64px;
              border-radius: var(--v2-radius-3);
              background: {peerName ? 'var(--v2-accent)' : 'var(--v2-paper-2)'};
              border: 1px solid {peerName ? 'transparent' : 'var(--v2-rule)'};
              color: {peerName ? 'var(--v2-paper)' : 'var(--v2-ink-3)'};
              display: grid; place-items: center;
            ">
              <Glyph name="shield" size={28} />
            </div>
            <Stack gap={1} align="center">
              <Ink kind="ui" weight="medium">{peerName ?? 'Awaiting pair…'}</Ink>
              <Ink kind="mono-cap" tone="ink-3">{peerName ? 'peer' : '—'}</Ink>
            </Stack>
          </Stack>
        </div>
      </Surface>

      {#if !paired}
        <!-- ── QR + PIN ─────────────────────────────────── -->
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: var(--v2-space-8);">

          <!-- QR Code -->
          <Surface elevation={0} padding="6" radius="2" tone="paper" style:display="flex" style:flex-direction="column" style:align-items="center">
            <Ink kind="caption" tone="ink-3" style:align-self="start">scan with the other device</Ink>
            <div style="
              width: 220px; height: 220px;
              margin: var(--v2-space-6) 0;
              padding: var(--v2-space-4);
              border: 1px solid var(--v2-rule);
              border-radius: var(--v2-radius-2);
              background: var(--v2-paper);
              display: grid; place-items: center;
              font-family: var(--v2-font-mono);
              font-size: var(--v2-text-12);
              color: var(--v2-ink-3);
              position: relative;
            ">
              {#if qrPayload}
                <img
                  src={`data:image/svg+xml;utf8,${qrPayload}`}
                  alt={`Pairing QR for ${myName}`}
                  style="max-width: 100%; max-height: 100%; display: block;"
                />
              {:else}
                <Ink kind="mono-cap" tone="ink-3">awaiting qr…</Ink>
              {/if}
            </div>
            <Ink kind="caption" tone="ink-3" style:align-self="start">encryption: ed25519 + noise xx</Ink>
          </Surface>

          <!-- PIN -->
          <Surface elevation={0} padding="6" radius="2" tone="paper">
            <Stack gap={4}>
              <Ink kind="caption" tone="ink-3">or enter this 6-digit pin</Ink>

              <div style="
                display: grid;
                grid-template-columns: repeat(6, 1fr);
                gap: var(--v2-space-2);
              ">
                {#each (pin + '000000').slice(0, 6).split('') as digit, i}
                  <div style="
                    aspect-ratio: 1 / 1.2;
                    border: 1px solid var(--v2-rule);
                    border-radius: var(--v2-radius-1);
                    display: grid; place-items: center;
                    font-family: var(--v2-font-mono);
                    font-size: var(--v2-text-28);
                    color: var(--v2-ink);
                    background: var(--v2-paper-2);
                    font-feature-settings: var(--v2-numeric-features);
                  ">{digit}</div>
                {/each}
              </div>

              <Inline gap={3} justify="between" align="center">
                <Stack gap={1}>
                  <Ink kind="caption" tone="ink-3">expires in</Ink>
                  <Ink kind="mono" weight="medium" tone={ttl <= 10 ? 'signal-stop' : 'ink'}>{fmtTtl}</Ink>
                </Stack>
                <Button variant="ghost" size="small" onclick={onPinChange}>Roll new pin</Button>
              </Inline>
            </Stack>
          </Surface>
        </div>
      {/if}

      {#if paired}
        <Surface elevation={0} padding="6" radius="2" tone="paper" style:border-left="3px solid var(--v2-signal-go)">
          <Inline gap={3} align="center">
            <Glyph name="check" size={16} />
            <Stack gap={1}>
              <Ink kind="ui" weight="medium" tone="signal-go">Paired with {peerName ?? 'peer'}.</Ink>
              <Ink kind="ui-small" tone="ink-3">Encrypted peer-to-peer. No data leaves your machines.</Ink>
            </Stack>
          </Inline>
        </Surface>
      {/if}

    </Stack>

  </div>
</div>
