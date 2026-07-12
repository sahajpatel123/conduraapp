<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Dot from './Dot.svelte';
  import Button from './Button.svelte';
  import Pulse from './Pulse.svelte';
  import { trust } from '../../stores/trust.svelte';
  import { ipc } from '../../ipc/client';
  import type { PermissionStatus } from '../../ipc/types';

  type DotVariant = 'success' | 'warning' | 'neutral' | 'error';

  interface PermMeta {
    name: string;
    desc: string;
    required: string;
    icon: string;
  }

  const PERM_KINDS = ['accessibility', 'screen_recording', 'microphone', 'notifications', 'automation'] as const;

  const PERM_META: Record<string, PermMeta> = {
    accessibility: {
      name: 'Accessibility',
      desc: 'Read structured UI elements — named buttons, fields, window titles.',
      required: 'Required for computer-use.',
      icon: 'a11y',
    },
    screen_recording: {
      name: 'Screen Recording',
      desc: 'Sample the screen occasionally when needed. Never continuously.',
      required: 'Optional — needed only for vision-based actions.',
      icon: 'screen',
    },
    microphone: {
      name: 'Microphone',
      desc: 'Voice input and the "hey condura" wake word.',
      required: 'Optional.',
      icon: 'mic',
    },
    notifications: {
      name: 'Notifications',
      desc: 'Task completion and important alerts.',
      required: 'Optional.',
      icon: 'bell',
    },
    automation: {
      name: 'Automation',
      desc: 'Control other apps via AppleEvents (System Events, etc.).',
      required: 'Required for cross-app automation.',
      icon: 'bot',
    },
  };

  let pollTimer: ReturnType<typeof setInterval> | null = null;
  let openBusy = $state<string | null>(null);

  const permissionRows = $derived(
    PERM_KINDS.map((kind) => {
      const status = trust.permissions.find((p) => p.kind === kind);
      const meta = PERM_META[kind];
      return {
        kind,
        name: meta?.name ?? kind,
        desc: meta?.desc ?? '',
        required: meta?.required ?? '',
        icon: meta?.icon ?? '',
        granted: status?.status === 'granted',
        denied: status?.status === 'denied',
        status: status?.status ?? 'unknown',
        note: status?.note ?? '',
      };
    })
  );

  function statusVariant(status: string): DotVariant {
    if (status === 'granted') return 'success';
    if (status === 'denied') return 'error';
    return 'neutral';
  }

  function statusLabel(status: string): string {
    if (status === 'granted') return 'granted';
    if (status === 'denied') return 'denied';
    return 'unknown';
  }

  function allGranted(): boolean {
    return permissionRows.every((p) => p.granted);
  }

  function grantedCount(): number {
    return permissionRows.filter((p) => p.granted).length;
  }

  async function openPermissionSettings(kind: string): Promise<void> {
    openBusy = kind;
    try {
      const { openPermissionSettings: openOS } = await import('../../utils/openPermissionSettings');
      await openOS(kind, ipc);
      await trust.refreshPermissions();
    } catch {
      try {
        const guide = await trust.loadGuide(kind);
        if (guide.deep_link) {
          const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } };
          if (w.runtime?.BrowserOpenURL) w.runtime.BrowserOpenURL(guide.deep_link);
          else window.open(guide.deep_link, '_blank');
        }
      } catch {
        // ignore
      }
    } finally {
      openBusy = null;
    }
  }

  onMount(() => {
    void trust.refreshPermissions();
    pollTimer = setInterval(() => void trust.refreshPermissions(), 2000);
  });

  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer);
    pollTimer = null;
  });
</script>

<div class="permission-section">
  <header class="permission-section__header">
    <h2>Permissions</h2>
    <p class="permission-section__lede">
      Each grant is revocable in your System Settings at any time.
    </p>
    {#if trust.loadingPermissions}
      <div class="permission-section__status-bar">
        <Pulse state="thinking" size="sm" label="Refreshing" />
        <span class="caption">Checking permission state…</span>
      </div>
    {:else}
      <div class="permission-section__summary {allGranted() ? 'summary--all-granted' : 'summary--partial'}">
        <span class="summary__count">{grantedCount()} / {PERM_KINDS.length}</span>
        <span class="caption">
          {allGranted() ? 'All permissions granted' : 'Some permissions missing'}
        </span>
      </div>
    {/if}
    {#if trust.lastError}
      <p class="permission-section__error">Error: {trust.lastError}</p>
    {/if}
  </header>

  <div class="permission-section__list">
    {#each permissionRows as perm (perm.kind)}
      <div class="perm-card" class:perm-card--granted={perm.granted} class:perm-card--denied={perm.denied}>
        <div class="perm-card__head">
          <div class="perm-card__name-row">
            <span class="perm-card__icon" aria-hidden="true">
              {#if perm.icon === 'a11y'}👁
              {:else if perm.icon === 'screen'}🖥
              {:else if perm.icon === 'mic'}🎤
              {:else if perm.icon === 'bell'}🔔
              {:else if perm.icon === 'bot'}🤖
              {/if}
            </span>
            <span class="perm-card__name">{perm.name}</span>
          </div>
          <div class="perm-card__status">
            <Dot variant={statusVariant(perm.status)} size="sm" />
            <span class="perm-card__status-label">{statusLabel(perm.status)}</span>
          </div>
        </div>
        <p class="perm-card__desc">{perm.desc}</p>
        {#if perm.note}
          <p class="perm-card__note">{perm.note}</p>
        {/if}
        <div class="perm-card__footer">
          <span class="perm-card__required">{perm.required}</span>
          <div class="perm-card__actions">
            {#if perm.granted}
              <Button
                variant="destructive"
                size="sm"
                loading={openBusy === perm.kind}
                onclick={() => openPermissionSettings(perm.kind)}
              >
                Open Settings
              </Button>
            {:else}
              <Button
                variant="secondary"
                size="sm"
                loading={openBusy === perm.kind}
                onclick={() => openPermissionSettings(perm.kind)}
              >
                Grant
              </Button>
            {/if}
          </div>
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
  .permission-section {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
  }

  .permission-section__header h2 {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    line-height: 1.3;
    font-weight: 400;
    color: var(--content-primary);
    margin: 0 0 var(--space-2) 0;
  }

  .permission-section__lede {
    font-size: var(--text-body-size);
    color: var(--content-tertiary);
    line-height: 1.6;
    margin: 0 0 var(--space-3) 0;
  }

  .permission-section__status-bar {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) 0;
  }

  .permission-section__summary {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-sm);
    background-color: var(--surface-sunken);
    width: fit-content;
  }

  .summary--all-granted {
    border: 1px solid var(--success-300, #6ee7b7);
  }

  .summary--partial {
    border: 1px solid var(--warning-300, #fcd34d);
  }

  .summary__count {
    font-family: var(--font-mono);
    font-size: var(--text-body-sm-size);
    font-weight: 600;
    color: var(--content-primary);
    font-variant-numeric: tabular-nums;
  }

  .permission-section__error {
    font-size: var(--text-body-sm-size);
    color: var(--error-600, #dc2626);
    margin: var(--space-2) 0 0 0;
  }

  .permission-section__list {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .perm-card {
    padding: var(--space-4);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    transition: border-color var(--duration-fast) var(--ease-standard);
  }

  .perm-card--granted {
    border-left: 3px solid var(--success-500, #10b981);
  }

  .perm-card--denied {
    border-left: 3px solid var(--error-500, #ef4444);
  }

  .perm-card__head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-2);
  }

  .perm-card__name-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .perm-card__icon {
    font-size: 1.1em;
    line-height: 1;
  }

  .perm-card__name {
    font-weight: 500;
    color: var(--content-primary);
  }

  .perm-card__status {
    display: flex;
    align-items: center;
    gap: var(--space-1_5);
  }

  .perm-card__status-label {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    letter-spacing: 0.02em;
    text-transform: capitalize;
  }

  .perm-card__desc {
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    line-height: 1.5;
    margin: 0 0 var(--space-1) 0;
  }

  .perm-card__note {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    margin: 0 0 var(--space-2) 0;
    padding: var(--space-1) var(--space-2);
    background-color: var(--surface-sunken);
    border-radius: var(--radius-xs);
    line-height: 1.4;
    word-break: break-word;
  }

  .perm-card__footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-top: var(--space-3);
  }

  .perm-card__required {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
  }
</style>
