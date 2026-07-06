<script lang="ts">
  /**
   * PermissionCards — OS permission toggles.
   * Each permission is a card with status and action.
   *
   * When running inside the Wails desktop shell, the component
   * polls the daemon's permissions.status RPC every 2s so the
   * badges stay accurate. On the web (design preview), it falls
   * back to static defaults.
   */
  import { onMount, onDestroy } from 'svelte'
  import { InkText, WordReveal, BlurReveal, PaperCard, PulseDot, InkReveal, MagneticButton } from '$lib/components/living'
  import type { PermissionStatus } from '../../ipc/types'

  interface PermItem {
    id: string
    label: string
    description: string
    icon: string
    granted: boolean
    kind: string
  }

  interface Props {
    onnext: () => void
    onskip: () => void
    permissions?: PermItem[]
    busy?: boolean
  }

  const DEFAULTS: PermItem[] = [
    { id: 'accessibility',    label: 'Accessibility',    description: 'Allow Condura to observe window focus for context-aware assistance',     icon: '👁', granted: false, kind: 'accessibility' },
    { id: 'screen_recording', label: 'Screen Recording', description: 'Allow Condura to capture screenshots for computer-use features',          icon: '🖥', granted: false, kind: 'screen_recording' },
    { id: 'microphone',       label: 'Microphone',       description: 'Allow voice interaction and wake-word detection',                         icon: '🎤', granted: false, kind: 'microphone' },
    { id: 'notifications',    label: 'Notifications',    description: 'Show alerts when agents need your attention',                             icon: '🔔', granted: false, kind: 'notifications' },
    { id: 'automation',       label: 'Automation',        description: 'Allow Condura to send AppleEvents to other apps (e.g. \"click Safari\")', icon: '🤖', granted: false, kind: 'automation' },
  ]

  let {
    onnext,
    onskip,
    permissions = DEFAULTS,
    busy = false,
  }: Props = $props()

  let items = $state([...permissions])
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const accessibilityGranted = $derived(
    items.some((p) => p.kind === 'accessibility' && p.granted)
  )
  const screenRecordingGranted = $derived(
    items.some((p) => p.kind === 'screen_recording' && p.granted)
  )
  const canContinue = $derived(
    (accessibilityGranted || screenRecordingGranted) && !busy
  )

  function badgeLabel(granted: boolean): string {
    return granted ? 'Granted' : 'Grant'
  }

  async function grantPerm(idx: number): Promise<void> {
    const perm = items[idx]
    if (!perm || perm.granted) return

    // Open the OS System Settings pane for this permission.
    try {
      const { ipc } = await import('../../ipc/client')
      const guide = await ipc.permissionsGuide(perm.kind)
      if (guide.deep_link) {
        const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } }
        if (w.runtime?.BrowserOpenURL) {
          w.runtime.BrowserOpenURL(guide.deep_link)
        } else {
          window.open(guide.deep_link, '_blank')
        }
      }
    } catch {
      // Best-effort: the deep link may fail in a non-Wails preview.
    }

    // Start polling so the badge updates when the user toggles
    // the permission in System Settings.
    if (!pollTimer) startPolling()
  }

  async function refresh(): Promise<void> {
    try {
      const { ipc } = await import('../../ipc/client')
      const statuses: PermissionStatus[] = await ipc.permissionsStatus()
      const next = [...items]
      for (const s of statuses) {
        const idx = items.findIndex((p) => p.kind === s.kind)
        if (idx >= 0) {
          next[idx] = { ...next[idx], granted: s.status === 'granted' }
        }
      }
      items = next
    } catch {
      // Daemon unavailable; keep last-known statuses.
    }
  }

  function startPolling(): void {
    if (pollTimer) return
    void refresh()
    pollTimer = setInterval(refresh, 2000)
  }

  onMount(() => {
    // Attempt a one-shot probe to pick up any already-granted
    // permissions. If the daemon isn't available (web preview),
    // this silently fails and the static defaults remain.
    void refresh().catch(() => {})
  })

  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer)
  })
</script>

<div style="max-width: 520px; margin: 0 auto; text-align: center;">
  <InkReveal direction="left" duration={900} delay={200}>
    <InkText kind="display" as="h1" style="margin-bottom: var(--lp-space-3);">
      <WordReveal text="Permissions" stagger={50} delay={300} />
    </InkText>
  </InkReveal>

  <BlurReveal delay={500} distance={16}>
    <InkText kind="body" tone="ink-mute" style="max-width: 400px; margin: 0 auto var(--lp-space-6);">
      Condura needs a few permissions to work fully. You can change these anytime.
    </InkText>
  </BlurReveal>

  <BlurReveal delay={700} distance={16}>
    <div style="display: flex; flex-direction: column; gap: var(--lp-space-3); max-width: 440px; margin: 0 auto;">
      {#each items as perm, i}
        <BlurReveal delay={800 + i * 100} distance={12}>
          <PaperCard border={perm.granted ? 'synapse' : 'none'} padding="var(--lp-space-3) var(--lp-space-4)">
            <div style="display: flex; align-items: center; gap: var(--lp-space-3);">
              <span style="font-size: 20px;">{perm.icon}</span>
              <div style="text-align: left; flex: 1;">
                <div style="display: flex; align-items: center; gap: var(--lp-space-2);">
                  <InkText kind="title" as="div">{perm.label}</InkText>
                  {#if perm.granted}
                    <PulseDot phase="ok" size={5} />
                  {/if}
                </div>
                <InkText kind="caption" tone="ink-mute">{perm.description}</InkText>
              </div>
              <button
                type="button"
                class="lp-focus"
                onclick={() => grantPerm(i)}
                style="
                  padding: 6px 14px;
                  border-radius: var(--lp-radius-sm);
                  border: 1px solid {perm.granted ? 'var(--lp-synapse)' : 'var(--lp-ink-ghost)'};
                  background: {perm.granted ? 'var(--lp-synapse)' : 'transparent'};
                  color: {perm.granted ? 'var(--lp-paper)' : 'var(--lp-ink-mute)'};
                  font-family: var(--lp-font-sans);
                  font-size: var(--lp-text-caption);
                  cursor: pointer;
                  white-space: nowrap;
                  transition: all var(--lp-dur-fast) var(--lp-ease-thread);
                "
              >
                {badgeLabel(perm.granted)}
              </button>
            </div>
          </PaperCard>
        </BlurReveal>
      {/each}
    </div>
  </BlurReveal>

  <BlurReveal delay={1200} distance={16}>
    <div style="display: flex; align-items: center; justify-content: center; gap: var(--lp-space-4); margin-top: var(--lp-space-8);">
      <button
        type="button"
        class="lp-focus"
        onclick={onskip}
        disabled={busy}
        style="
          padding: 10px 20px;
          border-radius: var(--lp-radius-sm);
          border: 1px solid var(--lp-ink-ghost);
          background: transparent;
          color: var(--lp-ink-mute);
          font-family: var(--lp-font-sans);
          font-size: var(--lp-text-body);
          cursor: pointer;
          transition: all var(--lp-dur-fast) var(--lp-ease-thread);
        "
      >Skip for now</button>
      <MagneticButton variant="primary" size="md" onclick={onnext} disabled={!canContinue || busy}>
        Continue
      </MagneticButton>
    </div>
  </BlurReveal>
</div>

<style>
  button:hover {
    opacity: 0.85;
  }
</style>
