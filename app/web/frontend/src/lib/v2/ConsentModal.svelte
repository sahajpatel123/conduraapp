<!--
  ConsentModal — Condura v2 Gatekeeper surface.

  The wax-seal-on-a-letter moment. When the Gatekeeper asks the user
  to approve a non-READ action, this modal appears:

    - Centered card, max 480px, elevation 3
    - Top ribbon (4px × full-width) in `accent` — the "this needs
      a human" signal, learned by repetition
    - Body explains the action in plain language
    - Per-action context: which app, what's about to be touched
    - Three actions: Deny (quiet), Allow once, Allow for session
    - On Allow: a circle draws in `accent` over the button —
      the wax-seal moment. The only ceremonial motion in v2.

  Props:
    open:           boolean
    title:          string   — e.g. "Send email to alex@example.com"
    description:    string   — plain-language explanation
    blastRadius:    'read' | 'write' | 'network' | 'destructive'
    target?:        { app: string, detail: string }
    impact?:        string[] — bullet points of what will happen
    sessionTtlLabel?: string — label for the "allow for session" option
    onDeny:         () => void
    onAllowOnce:    () => void
    onAllowSession: () => void
    busy?:          boolean  — disables all buttons while IPC call in flight
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Button } from '$lib/v2'

  type BlastRadius = 'read' | 'write' | 'network' | 'destructive'

  let {
    open = false as boolean,
    title = '' as string,
    description = '' as string,
    blastRadius = 'write' as BlastRadius,
    target = undefined as { app: string, detail: string } | undefined,
    impact = [] as string[],
    sessionTtlLabel = 'Allow for this session',
    onDeny = undefined as (() => void) | undefined,
    onAllowOnce = undefined as (() => void) | undefined,
    onAllowSession = undefined as (() => void) | undefined,
    busy = false as boolean,
  } = $props()

  // Wax-seal stamp animation — runs when the user clicks an allow button.
  // `stampTarget` is which button the seal appears on ('once' or 'session').
  // `stampRunning` is true during the 280ms animation.
  // `pendingTimeout` is tracked so a deny mid-stamp can cancel it
  // (otherwise the allow callback would still fire 280ms later — a
  // double-action data-loss bug).
  let stampTarget = $state<'once' | 'session' | null>(null)
  let stampRunning = $state(false)
  let pendingTimeout: ReturnType<typeof setTimeout> | null = null

  function allowOnce() {
    if (busy || stampRunning) return
    stampTarget = 'once'
    stampRunning = true
    pendingTimeout = setTimeout(() => {
      pendingTimeout = null
      stampRunning = false
      stampTarget = null
      onAllowOnce?.()
    }, 280)
  }

  function allowSession() {
    if (busy || stampRunning) return
    stampTarget = 'session'
    stampRunning = true
    pendingTimeout = setTimeout(() => {
      pendingTimeout = null
      stampRunning = false
      stampTarget = null
      onAllowSession?.()
    }, 280)
  }

  function deny() {
    // CRITICAL: cancel any pending allow callback. Without this the
    // allow fires 280ms after deny, leading to double-action.
    if (pendingTimeout !== null) {
      clearTimeout(pendingTimeout)
      pendingTimeout = null
    }
    stampRunning = false
    stampTarget = null
    onDeny?.()
  }

  // Blast radius → label + tone for the ribbon
  function radiusLabel(r: BlastRadius): string {
    return r === 'read' ? 'read' :
           r === 'write' ? 'edit' :
           r === 'network' ? 'send over network' :
           'destructive'
  }

  // Esc to deny. CRITICAL: the listener is registered ONLY inside the
  // `{#if open}` block (see template below) so it does not swallow Esc
  // for the rest of the page when the modal is closed.
  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault()
      deny()
    }
  }

  // Reset stamp state when modal closes. Critical to clear any
  // pending setTimeout so a deny mid-stamp cannot leak an allow.
  $effect(() => {
    if (!open) {
      if (pendingTimeout !== null) {
        clearTimeout(pendingTimeout)
        pendingTimeout = null
      }
      stampRunning = false
      stampTarget = null
    }
  })
</script>

{#if open}
<svelte:window onkeydown={onKeydown} />

  <!-- ── Backdrop ───────────────────────────────────────────── -->
  <div
    data-v2
    style="
      position: fixed;
      inset: 0;
      background: color-mix(in srgb, var(--v2-ink) 12%, transparent);
      z-index: var(--v2-z-modal);
      display: grid;
      place-items: center;
      padding: var(--v2-space-8);
      animation: v2-fade-in var(--v2-dur-mid) var(--v2-ease-out-soft) both;
    "
  >
    <!-- ── Modal card ──────────────────────────────────────── -->
    <div style="
      width: 100%;
      max-width: 480px;
      animation: v2-scale-in var(--v2-dur-slow) var(--v2-ease-settle) both;
    ">
      <Surface
        elevation={3}
        padding="0"
        radius="3"
        tone="paper"
        style:overflow="hidden"
      >
        <Stack gap={0}>

          <!-- The ribbon — the "this needs a human" signal -->
          <div style="
            width: 100%;
            height: 4px;
            background: var(--v2-accent);
          "></div>

          <div style="padding: var(--v2-space-8);">

            <Stack gap={6}>

              <Stack gap={3}>
                <Inline gap={2} align="center">
                  <Ink kind="mono-cap" tone="accent">condura · gatekeeper</Ink>
                  <span style="
                    font-family: var(--v2-font-mono);
                    font-size: var(--v2-text-12);
                    color: var(--v2-ink-3);
                    padding: 2px var(--v2-space-2);
                    border-radius: var(--v2-radius-pill);
                    border: 1px solid var(--v2-rule);
                    text-transform: uppercase;
                    letter-spacing: 0.06em;
                  ">
                    {radiusLabel(blastRadius)}
                  </span>
                </Inline>
                <Ink kind="title" as="h2">{title}</Ink>
              </Stack>

              <Ink kind="body" tone="ink-2">
                {description}
              </Ink>

              {#if target}
                <Surface elevation={0} padding="4" radius="2" tone="paper-2">
                  <Stack gap={2}>
                    <Ink kind="caption" tone="ink-3">target</Ink>
                    <Stack gap={1}>
                      <Ink kind="ui-small" weight="medium" tone="ink">{target.app}</Ink>
                      <Ink kind="ui-small" tone="ink-2">{target.detail}</Ink>
                    </Stack>
                  </Stack>
                </Surface>
              {/if}

              {#if impact.length > 0}
                <Stack gap={2}>
                  <Ink kind="caption" tone="ink-3">this will</Ink>
                  <Stack gap={1}>
                    {#each impact as bullet}
                      <Inline gap={2} align="baseline">
                        <span style="
                          color: var(--v2-accent);
                          font-family: var(--v2-font-mono);
                          font-size: var(--v2-text-14);
                          line-height: 1;
                        ">·</span>
                        <Ink kind="body" tone="ink-2">{bullet}</Ink>
                      </Inline>
                    {/each}
                  </Stack>
                </Stack>
              {/if}

              <Rule />

              <!-- ── Footer with the wax-seal buttons ──────── -->
              <Stack gap={3}>
                <Inline gap={2} justify="end" align="center">
                  <Button variant="deny" size="small" onclick={deny} disabled={busy}>Deny</Button>
                  <Button
                    variant="ghost"
                    size="small"
                    onclick={allowOnce}
                    disabled={busy || stampRunning}
                    style:position="relative"
                    style:overflow="hidden"
                  >
                    {#if stampTarget === 'once' && stampRunning}
                      <!-- The wax seal — a circle stamped INSIDE the
                           button bounds. 32px diameter = a stamp,
                           not a halo. Same size for both once and
                           session so the ceremony reads as one. -->
                      <span style="
                        position: absolute;
                        left: 50%; top: 50%;
                        transform: translate(-50%, -50%);
                        width: 32px; height: 32px;
                        border: 1.5px solid var(--v2-accent);
                        border-radius: var(--v2-radius-pill);
                        animation: v2-stamp 280ms var(--v2-ease-spring) both;
                        pointer-events: none;
                      "></div>
                    {/if}
                    Allow once
                  </Button>
                  <Button
                    variant="primary"
                    size="small"
                    onclick={allowSession}
                    disabled={busy || stampRunning}
                    style:position="relative"
                    style:overflow="hidden"
                  >
                    {#if stampTarget === 'session' && stampRunning}
                      <span style="
                        position: absolute;
                        left: 50%; top: 50%;
                        transform: translate(-50%, -50%);
                        width: 32px; height: 32px;
                        border: 1.5px solid var(--v2-accent);
                        border-radius: var(--v2-radius-pill);
                        animation: v2-stamp 280ms var(--v2-ease-spring) both;
                        pointer-events: none;
                      "></span>
                    {/if}
                    {sessionTtlLabel}
                  </Button>
                </Inline>
                <Ink kind="caption" tone="ink-3" style:text-align="end">
                  Esc to deny
                </Ink>
              </Stack>

            </Stack>

          </div>
        </Stack>
      </Surface>
    </div>
  </div>
{/if}
