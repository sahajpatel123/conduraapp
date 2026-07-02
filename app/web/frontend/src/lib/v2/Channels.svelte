<!--
  Channels — Condura v2 reach surface.

  Per the spec: "A control panel of radios. Each channel is a 'tuner'
  row with signal-quality dots."

  Each channel row shows:
    - Channel name + signal-quality dots (4 little dots, the more
      filled the stronger the integration health)
    - Last message timestamp
    - Connect / disconnect action
    - Optional "carrier-lock" tone when status flips (future)

  Props:
    channels: Channel[]
    onConnect?: (id: string) => void
    onDisconnect?: (id: string) => void
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Eyebrow, Glyph, Button } from '$lib/v2'

  type Status = 'connected' | 'connecting' | 'disconnected' | 'error'

  export interface Channel {
    id: string
    name: string
    handle?: string
    description: string
    status: Status
    lastSeen?: string
    signalStrength: 0 | 1 | 2 | 3 | 4  // number of filled dots
    unread?: number
  }

  let {
    channels = [] as Channel[],
    onConnect = undefined as ((id: string) => void) | undefined,
    onDisconnect = undefined as ((id: string) => void) | undefined,
  }: {
    channels?: Channel[]
    onConnect?: (id: string) => void
    onDisconnect?: (id: string) => void
  } = $props()

  function dots(n: number, status: Status): { filled: boolean, error: boolean }[] {
    return Array.from({ length: 4 }, (_, i) => ({
      filled: i < n,
      error:  status === 'error' && i === 0,
    }))
  }

  function statusLabel(s: Status): string {
    return s === 'connected' ? 'connected' :
           s === 'connecting' ? 'connecting…' :
           s === 'error' ? 'error' :
           'not connected'
  }
</script>

<div data-v2 style="
  flex: 1;
  background: var(--v2-paper);
  overflow-y: auto;
  padding: var(--v2-space-12);
  box-sizing: border-box;
">
  <div style="max-width: 760px; margin: 0 auto;">

    <Stack gap={10}>

      <Stack gap={3}>
        <Eyebrow left="reach" right="channels" tone="accent" />
        <Ink kind="display" as="h1">Talk to condura from anywhere.</Ink>
        <Ink kind="body-2" tone="ink-2" as="p" style:max-width="640px">
          Connect your messaging apps. Condura replies there too —
          with the same Gatekeeper, the same audit, the same adapter
          for you.
        </Ink>
      </Stack>

      <Rule />

      <Stack gap={3}>
        {#each channels as c (c.id)}
          <Surface
            elevation={0}
            padding="6"
            radius="2"
            tone="paper"
          >
            <Stack gap={4}>
              <Inline gap={4} justify="between" align="start">
                <Stack gap={1} style:flex="1">
                  <Inline gap={2} align="baseline">
                    <Ink kind="title" style:font-size="var(--v2-text-20)">{c.name}</Ink>
                    {#if c.handle}<Ink kind="mono-cap" tone="ink-3">@{c.handle}</Ink>{/if}
                  </Inline>
                  <Ink kind="ui-small" tone="ink-3">{c.description}</Ink>
                </Stack>

                <!-- Signal-quality dots -->
                <Inline gap={2} align="center">
                  <Inline gap={1} align="center" aria-label="signal strength">
                    {#each dots(c.signalStrength, c.status) as d, i}
                      <span style="
                        width: 6px;
                        height: {4 + i * 2}px;
                        border-radius: var(--v2-radius-pill);
                        background: {d.error
                          ? 'var(--v2-signal-stop)'
                          : d.filled
                            ? 'var(--v2-accent)'
                            : 'var(--v2-rule)'};
                      "></span>
                    {/each}
                  </Inline>
                </Inline>
              </Inline>

              <Inline gap={3} justify="between" align="center">
                <Inline gap={2} align="center">
                  {#if c.status === 'connected'}
                    <Glyph name="check" size={12} />
                    <Ink kind="ui-small" tone="signal-go" weight="medium">connected</Ink>
                    {#if c.lastSeen}<span style="
                      font-family: var(--v2-font-mono);
                      font-size: 11px;
                      color: var(--v2-ink-3);
                    ">· last seen {c.lastSeen}</span>{/if}
                    {#if c.unread && c.unread > 0}
                      <span style="
                        margin-left: var(--v2-space-2);
                        font-family: var(--v2-font-mono);
                        font-size: 11px;
                        color: var(--v2-accent);
                        padding: 1px 6px;
                        border-radius: var(--v2-radius-1);
                        border: 1px solid var(--v2-accent);
                      ">{c.unread} unread</span>
                    {/if}
                  {:else if c.status === 'connecting'}
                    <span style="
                      width: 8px; height: 8px;
                      border-radius: var(--v2-radius-pill);
                      border: 1.5px solid var(--v2-accent);
                      border-top-color: transparent;
                      animation: v2-spin 1s linear infinite;
                      display: inline-block;
                    "></span>
                    <Ink kind="ui-small" tone="ink-3">connecting…</Ink>
                  {:else if c.status === 'error'}
                    <Glyph name="x" size={12} />
                    <Ink kind="ui-small" tone="signal-stop" weight="medium">error</Ink>
                  {:else}
                    <Glyph name="circle" size={12} />
                    <Ink kind="ui-small" tone="ink-3">not connected</Ink>
                  {/if}
                </Inline>

                {#if c.status === 'connected'}
                  <Button variant="ghost" size="small" onclick={() => onDisconnect?.(c.id)}>Disconnect</Button>
                {:else if c.status === 'connecting'}
                  <Button variant="ghost" size="small" disabled>Cancel</Button>
                {:else}
                  <Button variant="primary" size="small" onclick={() => onConnect?.(c.id)}>Connect</Button>
                {/if}
              </Inline>
            </Stack>
          </Surface>
        {/each}
      </Stack>

    </Stack>

  </div>
</div>

<style>
  @keyframes v2-spin {
    to { transform: rotate(360deg); }
  }
</style>
