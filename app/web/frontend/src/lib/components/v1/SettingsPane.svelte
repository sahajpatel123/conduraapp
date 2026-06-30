<!--
  SettingsPane — "What Synaptic Knows."

  Per spec §11.3: full-window pane. Audit-first. The first row is ALWAYS
  "What I've done in the last 24 hours" — trust is built by visibility,
  not explanation. 7 sections in fixed order.

  Sections in order:
    1. Action replay (24h scrubbable timeline, expandable per action)
    2. Adaptive engine profile (editable, deletable, exportable)
    3. Permission grants (one-click revoke)
    4. Hotkey configuration
    5. Autonomy matrix (per-app + per-task-type dials)
    6. Backup controls
    7. Account, sync, integrations (lower priority)

  Motion: surface expands from center, 260ms ease-out. Pulse moves to
  top-left corner. Settings content fades in after surface settles.

  Props:
    activeSection — optional initial section to show
    onclose       — close handler
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import Pulse from './Pulse.svelte';
  import Hairline from './Hairline.svelte';
  import Switch from './Switch.svelte';
  import Pill from './Pill.svelte';
  import Button from './Button.svelte';
  import Receipt from './Receipt.svelte';
  import Dot from './Dot.svelte';
  import HotkeyRecorder from './HotkeyRecorder.svelte';

  import { replay } from '../../stores/replay.svelte';
  import { trust } from '../../stores/trust.svelte';
  import { settings } from '../../stores/settings.svelte';
  import { account } from '../../stores/account.svelte';
  import { sync } from '../../stores/sync.svelte';
  import { ipc } from '../../ipc/client';
  import type {
    AdaptiveStrength,
    AdaptiveUserModel,
    InferredField,
    ReplayFrame,
  } from '../../ipc/types';

  type Section = 'replay' | 'adaptive' | 'permissions' | 'hotkey' | 'autonomy' | 'backup' | 'account';
  type AutonomyLevel = 'autonomous' | 'warn' | 'block';
  type ReceiptState = 'done' | 'paused' | 'error' | 'pending';

  interface AdaptiveRow {
    claim: string;
    evidence: string;
    confidence: number;
    field: string;
    value: string;
  }

  interface Props {
    activeSection?: Section;
    onclose?: () => void;
  }

  let { activeSection = 'replay', onclose }: Props = $props();

  let current = $state<Section>('replay');
  let strength = $state<AdaptiveStrength>('balanced');
  let adaptiveProfile = $state<AdaptiveUserModel | null>(null);
  let adaptiveLoading = $state(false);
  let adaptiveBusy = $state(false);
  let wakeEnabled = $state(false);
  let hotkeyCombo = $state('');
  let exportBusy = $state(false);
  let backupBusy = $state(false);
  let restoreBusy = $state(false);
  let permPollTimer: ReturnType<typeof setInterval> | null = null;

  $effect(() => {
    current = activeSection;
  });

  const SECTIONS: Array<{ id: Section; label: string; hint: string }> = [
    { id: 'replay',      label: '01', hint: 'Action replay' },
    { id: 'adaptive',    label: '02', hint: 'Adaptive engine' },
    { id: 'permissions', label: '03', hint: 'Permissions' },
    { id: 'hotkey',      label: '04', hint: 'Hotkey' },
    { id: 'autonomy',    label: '05', hint: 'Autonomy matrix' },
    { id: 'backup',      label: '06', hint: 'Backup & restore' },
    { id: 'account',     label: '07', hint: 'Account & sync' },
  ];

  const PERM_KINDS = ['accessibility', 'screen_recording', 'microphone', 'notifications'] as const;

  const PERM_META: Record<string, { name: string; desc: string; required: string }> = {
    accessibility: {
      name: 'Accessibility',
      desc: 'I read structured UI elements (named buttons, fields, window titles).',
      required: 'Required for computer-use.',
    },
    screen_recording: {
      name: 'Screen Recording',
      desc: 'I sample the screen occasionally when needed. Never continuously.',
      required: 'Optional — needed only for vision-based actions.',
    },
    microphone: {
      name: 'Microphone',
      desc: 'For voice input and the "hey condura" wake word.',
      required: 'Optional.',
    },
    notifications: {
      name: 'Notifications',
      desc: 'For task completion and important alerts.',
      required: 'Optional.',
    },
  };

  const AUTONOMY_TASK_TYPES: Array<{ key: string; label: string }> = [
    { key: 'coding', label: 'Coding' },
    { key: 'file_operations', label: 'File operations' },
    { key: 'web_browsing', label: 'Web browsing' },
    { key: 'email', label: 'Email' },
    { key: 'calendar', label: 'Calendar' },
    { key: 'messaging', label: 'Messaging' },
    { key: 'shell_commands', label: 'Shell commands' },
    { key: 'computer_use', label: 'Computer use' },
    { key: 'research', label: 'Research' },
    { key: 'image_generation', label: 'Image generation' },
    { key: 'code_review', label: 'Code review' },
  ];

  const STRENGTH_OPTS: AdaptiveStrength[] = ['off', 'cautious', 'balanced', 'aggressive'];

  const permissionRows = $derived(
    PERM_KINDS.map((kind) => {
      const status = trust.permissions.find((p) => p.kind === kind);
      const meta = PERM_META[kind];
      return {
        kind,
        name: meta?.name ?? kind,
        desc: meta?.desc ?? '',
        required: meta?.required ?? '',
        granted: status?.status === 'granted',
        status: status?.status ?? 'unknown',
      };
    })
  );

  const adaptiveItems = $derived(adaptiveRowsFromProfile(adaptiveProfile));

  const autonomyTasks = $derived(
    AUTONOMY_TASK_TYPES.map((t) => ({
      key: t.key,
      task: t.label,
      level: taskAutonomyLevel(t.key),
    }))
  );

  const backupDir = $derived(
    settings.config?.storage?.backup?.dir || '~/Documents/condura-backups/'
  );

  const hotkeyDisplay = $derived(formatHotkeyDisplay(hotkeyCombo || settings.config?.hotkey?.overlay || ''));

  const replayRows = $derived(replay.frames.map(frameToReceipt));

  onMount(() => {
    void bootstrap();
    permPollTimer = setInterval(() => {
      if (current === 'permissions') void trust.refreshPermissions();
    }, 2000);
    return () => {
      if (permPollTimer) clearInterval(permPollTimer);
    };
  });

  async function bootstrap(): Promise<void> {
    await Promise.allSettled([
      settings.refresh(),
      replay.refresh(),
      trust.refreshPermissions(),
      trust.refreshBackups(),
      account.checkStatus(),
      sync.refresh(),
      loadAdaptive(),
      loadVoiceProbe(),
    ]);
    hotkeyCombo = settings.config?.hotkey?.overlay ?? '';
  }

  async function loadAdaptive(): Promise<void> {
    adaptiveLoading = true;
    try {
      const [profile, strengthRes] = await Promise.all([
        ipc.adaptiveProfile(),
        ipc.adaptiveStrengthGet(),
      ]);
      adaptiveProfile = profile;
      strength = strengthRes.strength;
    } catch {
      adaptiveProfile = null;
    } finally {
      adaptiveLoading = false;
    }
  }

  async function loadVoiceProbe(): Promise<void> {
    try {
      const probe = await ipc.onboardingProbeVoice();
      wakeEnabled = probe.wake_word_enabled;
    } catch {
      // voice probe unavailable
    }
  }

  function adaptiveRowsFromProfile(model: AdaptiveUserModel | null): AdaptiveRow[] {
    if (!model) return [];
    const rows: AdaptiveRow[] = [];

    const pushField = (field: string, item: InferredField) => {
      if (!item?.value) return;
      rows.push({
        claim: item.value,
        evidence: (item.evidence ?? []).join(' · ') || '—',
        confidence: item.confidence ?? 0,
        field,
        value: item.value,
      });
    };

    const pushMany = (field: string, items: InferredField[]) => {
      for (const item of items ?? []) pushField(field, item);
    };

    pushMany('preferences', model.preferences);
    pushField('style', model.style);
    pushMany('expertise', model.expertise);
    pushMany('pet_peeves', model.pet_peeves);
    pushField('communication', model.communication);
    pushField('risk_tolerance', model.risk_tolerance);
    return rows;
  }

  function taskAutonomyLevel(key: string): AutonomyLevel {
    const level = settings.config?.autonomy?.per_task?.[key]
      ?? settings.config?.autonomy?.default_level
      ?? 'warn';
    if (level === 'autonomous' || level === 'block') return level;
    return 'warn';
  }

  function frameToReceipt(f: ReplayFrame): {
    id: number;
    time: string;
    verb: string;
    target: string;
    state: ReceiptState;
    model: string;
    gate: string;
  } {
    return {
      id: f.id,
      time: fmtTime(f.timestamp),
      verb: f.action || 'acted',
      target: f.message || f.app || '—',
      state: receiptState(f),
      model: f.actor || '—',
      gate: `${f.level || '—'} → ${f.outcome || f.result || '—'}`,
    };
  }

  function receiptState(f: ReplayFrame): ReceiptState {
    if (f.outcome === 'denied' || f.result === 'block') return 'error';
    if (f.outcome === 'queued' || f.result === 'prompt') return 'pending';
    if (f.outcome === 'paused') return 'paused';
    return 'done';
  }

  function fmtTime(ts: string): string {
    try {
      return new Date(ts).toLocaleTimeString(undefined, {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false,
      });
    } catch {
      return ts;
    }
  }

  function formatHotkeyDisplay(spec: string): string {
    if (!spec) return '—';
    return spec
      .split('+')
      .map((part) => {
        switch (part) {
          case 'Cmd': return '⌘';
          case 'Ctrl': return '^';
          case 'Option':
          case 'Alt': return '⌥';
          case 'Shift': return '⇧';
          case 'Space': return 'Space';
          default: return part;
        }
      })
      .join('');
  }

  function daemonHotkeyFromDisplay(display: string): string {
    const replacements: Array<[string, string]> = [
      ['⌘', 'Cmd+'],
      ['^', 'Ctrl+'],
      ['⌥', 'Option+'],
      ['⇧', 'Shift+'],
    ];
    let spec = display;
    for (const [from, to] of replacements) spec = spec.split(from).join(to);
    if (spec.endsWith('+')) spec = spec.slice(0, -1);
    return spec;
  }

  function openExternal(url: string): void {
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } };
    if (w.runtime?.BrowserOpenURL) w.runtime.BrowserOpenURL(url);
    else window.open(url, '_blank');
  }

  async function openPermissionSettings(kind: string): Promise<void> {
    try {
      const guide = await trust.loadGuide(kind);
      if (guide.deep_link) openExternal(guide.deep_link);
    } catch {
      // guide unavailable
    }
  }

  async function saveHotkey(combo: string): Promise<void> {
    const spec = combo.includes('+') ? combo : daemonHotkeyFromDisplay(combo);
    if (!spec) return;
    try {
      hotkeyCombo = spec;
      await settings.save({
        hotkey: { ...(settings.config?.hotkey ?? { overlay: '' }), overlay: spec },
      });
    } catch {
      // keep last-known combo
    }
  }

  async function setWakeEnabled(enabled: boolean): Promise<void> {
    wakeEnabled = enabled;
    try {
      await ipc.configUpdate({ voice: { wake: { enabled } } } as Partial<import('../../ipc/types').AppConfig>);
    } catch {
      wakeEnabled = !enabled;
    }
  }

  async function setStrength(next: AdaptiveStrength): Promise<void> {
    const prev = strength;
    strength = next;
    try {
      await ipc.adaptiveStrengthSet(next);
    } catch {
      strength = prev;
    }
  }

  async function forgetAdaptive(field: string, value: string): Promise<void> {
    adaptiveBusy = true;
    try {
      await ipc.adaptiveForget(field, value);
      await loadAdaptive();
    } finally {
      adaptiveBusy = false;
    }
  }

  async function resetAdaptive(): Promise<void> {
    if (!confirm('Delete all learned inferences and start fresh? This cannot be undone.')) return;
    adaptiveBusy = true;
    try {
      await ipc.adaptiveReset();
      await loadAdaptive();
    } finally {
      adaptiveBusy = false;
    }
  }

  function exportAdaptiveProfile(): void {
    if (!adaptiveProfile) return;
    const blob = new Blob([JSON.stringify(adaptiveProfile, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'condura-adaptive-profile.json';
    a.click();
    URL.revokeObjectURL(url);
  }

  async function setTaskAutonomy(key: string, level: AutonomyLevel): Promise<void> {
    const autonomy = settings.config?.autonomy ?? {
      default_level: 'warn',
      per_app: {},
      per_task: {},
    };
    await settings.save({
      autonomy: {
        ...autonomy,
        per_task: { ...autonomy.per_task, [key]: level },
      },
    });
  }

  async function exportReplay(): Promise<void> {
    exportBusy = true;
    try {
      await replay.exportMP4();
    } finally {
      exportBusy = false;
    }
  }

  async function createBackup(): Promise<void> {
    backupBusy = true;
    try {
      await trust.createBackup();
    } finally {
      backupBusy = false;
    }
  }

  async function restoreBackup(): Promise<void> {
    const latest = trust.backups[0];
    const path = latest?.path ?? window.prompt('Path to backup .zip archive');
    if (!path) return;
    if (!confirm(`Restore from ${path}? Current data will be replaced.`)) return;
    restoreBusy = true;
    try {
      await ipc.backupRestore({ path });
      await Promise.allSettled([settings.refresh(), replay.refresh(), loadAdaptive()]);
    } finally {
      restoreBusy = false;
    }
  }

  function openAuditEntry(id: number): void {
    window.location.hash = `#/audit?id=${id}`;
  }

  function signIn(): void {
    window.location.hash = '#/settings?signin=1';
  }

  function openSync(): void {
    window.location.hash = '#/sync';
  }

  function openChannels(): void {
    window.location.hash = '#/channels';
  }
</script>

<div class="pane" role="region" aria-label="Settings">
  <!-- Top bar with pulse + close -->
  <header class="pane__topbar">
    <div class="pane__topbar-left">
      <Pulse state="idle" size="sm" label="Settings" />
      <span class="pane__topbar-title">What Synaptic knows</span>
    </div>
    <Button variant="tertiary" size="sm" onclick={onclose}>Close</Button>
  </header>

  <div class="pane__layout">
    <!-- Section nav (left) -->
    <nav class="pane__nav" aria-label="Settings sections">
      {#each SECTIONS as sec}
        <button
          class="nav-item"
          class:nav-item--active={current === sec.id}
          type="button"
          onclick={() => current = sec.id}
        >
          <span class="nav-item__num">{sec.label}</span>
          <span class="nav-item__hint">{sec.hint}</span>
        </button>
      {/each}
    </nav>

    <!-- Section content (right) -->
    <div class="pane__content">
      {#if current === 'replay'}
        <header class="content__header">
          <h2>What I've done in the last 24 hours</h2>
          <p class="content__lede">Every action I took, with the model and the gatekeeper rule that approved it. Transparency, not theater.</p>
        </header>
        <div class="replay">
          {#if replay.loading}
            <p class="content__lede">Loading action replay…</p>
          {:else if replayRows.length === 0}
            <p class="content__lede">No actions in the last 24 hours.</p>
          {:else}
            {#each replayRows as action (action.id)}
              <details class="replay__row">
                <summary>
                  <Receipt timestamp={action.time} verb={action.verb} target={action.target} state={action.state} />
                </summary>
                <div class="replay__detail">
                  <div class="replay__field"><span class="caption">model</span> <code>{action.model}</code></div>
                  <div class="replay__field"><span class="caption">gate</span> <code>{action.gate}</code></div>
                  <div class="replay__field">
                    <span class="caption">trace</span>
                    <Button variant="tertiary" size="sm" onclick={() => openAuditEntry(action.id)}>View full audit entry</Button>
                  </div>
                </div>
              </details>
            {/each}
          {/if}
        </div>
        <footer class="content__footer">
          <Button variant="secondary" size="sm" loading={exportBusy} disabled={replay.frames.length === 0} onclick={exportReplay}>
            Export 24h audit (.zip)
          </Button>
        </footer>

      {:else if current === 'adaptive'}
        <header class="content__header">
          <h2>What I've learned about you</h2>
          <p class="content__lede">Inferred from your behavior. Each item has evidence; delete anything that doesn't fit.</p>
        </header>
        <div class="adaptive">
          {#if adaptiveLoading}
            <p class="content__lede">Loading learned profile…</p>
          {:else if adaptiveItems.length === 0}
            <p class="content__lede">Nothing learned yet. Use Condura for a while — preferences will appear here.</p>
          {:else}
            {#each adaptiveItems as item (item.field + item.value)}
              <div class="adaptive__row">
                <div class="adaptive__claim">{item.claim}</div>
                <div class="adaptive__evidence">{item.evidence}</div>
                <div class="adaptive__meta">
                  <span class="caption">confidence</span>
                  <code>{(item.confidence * 100).toFixed(0)}%</code>
                  <Button
                    variant="tertiary"
                    size="sm"
                    disabled={adaptiveBusy}
                    onclick={() => forgetAdaptive(item.field, item.value)}
                  >
                    Forget
                  </Button>
                </div>
              </div>
            {/each}
          {/if}
        </div>
        <footer class="content__footer">
          <Button variant="secondary" size="sm" disabled={!adaptiveProfile} onclick={exportAdaptiveProfile}>Export profile</Button>
          <Button variant="destructive" size="sm" loading={adaptiveBusy} onclick={resetAdaptive}>Reset everything</Button>
        </footer>

      {:else if current === 'permissions'}
        <header class="content__header">
          <h2>Permissions</h2>
          <p class="content__lede">Each grant is revocable. I'll stop the moment you do.</p>
        </header>
        <div class="permissions">
          {#each permissionRows as perm (perm.kind)}
            <div class="perm">
              <div class="perm__head">
                <div class="perm__name">{perm.name}</div>
                <div class="perm__status">
                  <Dot variant={perm.granted ? 'success' : 'neutral'} size="sm" />
                  <span class="caption">{perm.granted ? 'granted' : 'not granted'}</span>
                </div>
              </div>
              <p class="perm__desc">{perm.desc}</p>
              <p class="perm__required"><span class="caption">{perm.required}</span></p>
              <div class="perm__actions">
                <Button
                  variant={perm.granted ? 'destructive' : 'secondary'}
                  size="sm"
                  onclick={() => openPermissionSettings(perm.kind)}
                >
                  {perm.granted ? 'Revoke' : 'Grant'}
                </Button>
              </div>
            </div>
          {/each}
        </div>

      {:else if current === 'hotkey'}
        <header class="content__header">
          <h2>Hotkey</h2>
          <p class="content__lede">Press a combo to record. I'll appear at your cursor when you do.</p>
        </header>
        <div class="hotkey-stage">
          <div class="hotkey-display">
            <span class="hotkey-combo">{hotkeyDisplay}</span>
          </div>
          <HotkeyRecorder
            value={hotkeyDisplay}
            label="Record new combo"
            onrecord={(combo) => { void saveHotkey(combo); }}
          />
        </div>
        <Hairline />
        <div class="hotkey-voice">
          <Switch
            label="Also say 'hey condura' to wake me"
            description="Local wake word, runs on your machine. Open-source model."
            checked={wakeEnabled}
            onchange={(enabled) => { void setWakeEnabled(enabled); }}
          />
        </div>

      {:else if current === 'autonomy'}
        <header class="content__header">
          <h2>Autonomy matrix</h2>
          <p class="content__lede">Each task type has a dial. Default: warn. Tune to taste.</p>
        </header>

        <div class="strength">
          <span class="caption">Strength</span>
          <div class="strength__options">
            {#each STRENGTH_OPTS as opt}
              <button
                class="strength__opt"
                class:strength__opt--active={strength === opt}
                type="button"
                onclick={() => { void setStrength(opt); }}
              >
                {opt}
              </button>
            {/each}
          </div>
        </div>

        <div class="autonomy">
          {#each autonomyTasks as t (t.key)}
            <div class="autonomy__row">
              <span class="autonomy__task">{t.task}</span>
              <div class="autonomy__dial">
                <button
                  class="dial dial--autonomous"
                  class:dial--active={t.level === 'autonomous'}
                  type="button"
                  aria-label="Autonomous"
                  aria-pressed={t.level === 'autonomous'}
                  onclick={() => { void setTaskAutonomy(t.key, 'autonomous'); }}
                ></button>
                <button
                  class="dial dial--warn"
                  class:dial--active={t.level === 'warn'}
                  type="button"
                  aria-label="Warn"
                  aria-pressed={t.level === 'warn'}
                  onclick={() => { void setTaskAutonomy(t.key, 'warn'); }}
                ></button>
                <button
                  class="dial dial--block"
                  class:dial--active={t.level === 'block'}
                  type="button"
                  aria-label="Block"
                  aria-pressed={t.level === 'block'}
                  onclick={() => { void setTaskAutonomy(t.key, 'block'); }}
                ></button>
              </div>
              <Pill variant={t.level === 'autonomous' ? 'success' : t.level === 'warn' ? 'warning' : 'error'} label={t.level} />
            </div>
          {/each}
        </div>

      {:else if current === 'backup'}
        <header class="content__header">
          <h2>Backup & restore</h2>
          <p class="content__lede">Your data lives on your machine. Back it up before uninstalling.</p>
        </header>
        <div class="backup">
          <Button variant="primary" size="md" loading={backupBusy} onclick={createBackup}>Export everything (.zip)</Button>
          <Button variant="secondary" size="md" loading={restoreBusy} disabled={trust.backups.length === 0} onclick={restoreBackup}>
            Restore from backup
          </Button>
          <Button variant="tertiary" size="md" onclick={() => { void trust.refreshBackups(); }}>
            Refresh backup list ({trust.backups.length})
          </Button>
        </div>
        <div class="backup-note">
          <p>
            <strong>Uninstalling?</strong> Condura will prompt you to save a backup to
            <code>{backupDir}</code> automatically.
          </p>
        </div>

      {:else if current === 'account'}
        <header class="content__header">
          <h2>Account & sync</h2>
          <p class="content__lede">Optional. The agent works signed-out, offline, with no channels and no account.</p>
        </header>
        <div class="account">
          {#if account.isSignedIn}
            <p class="content__lede">Signed in as {account.email} via {account.provider}.</p>
            <Button variant="secondary" size="md" loading={account.loading} onclick={() => account.signOut()}>Sign out</Button>
          {:else}
            <Button variant="secondary" size="md" onclick={signIn}>Sign in with email</Button>
          {/if}
          <Button variant="secondary" size="md" onclick={openSync}>
            Pair another device (P2P){sync.pairs.length > 0 ? ` · ${sync.pairs.length} paired` : ''}
          </Button>
          <Button variant="tertiary" size="md" onclick={openChannels}>Connect Telegram</Button>
        </div>
        <div class="account-note">
          <p>Sync is device-to-device, E2E encrypted. No central server. No account required.</p>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .pane {
    background-color: var(--surface-base);
    color: var(--content-primary);
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    font-family: var(--font-sans);
  }

  .pane__topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4) var(--space-6);
    border-bottom: 1px solid var(--border-subtle);
  }
  .pane__topbar-left {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }
  .pane__topbar-title {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-tertiary);
  }

  .pane__layout {
    display: grid;
    grid-template-columns: 240px 1fr;
    flex: 1;
    min-height: 0;
  }

  /* Section nav */
  .pane__nav {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: var(--space-4);
    border-right: 1px solid var(--border-subtle);
    background-color: var(--surface-sunken);
  }

  .nav-item {
    display: grid;
    grid-template-columns: 28px 1fr;
    gap: var(--space-2);
    align-items: baseline;
    padding: var(--space-2) var(--space-3);
    background-color: transparent;
    border: none;
    border-radius: var(--radius-sm);
    cursor: pointer;
    text-align: left;
    font-family: var(--font-sans);
    transition: background-color var(--duration-fast) var(--ease-standard);
  }
  .nav-item:hover {
    background-color: var(--paper-warm-50);
  }
  .nav-item:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: -2px;
  }
  .nav-item--active {
    background-color: var(--paper-warm-100);
  }
  .nav-item__num {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
  }
  .nav-item--active .nav-item__num {
    color: var(--content-accent);
  }
  .nav-item__hint {
    font-size: var(--text-body-sm-size);
    color: var(--content-secondary);
  }
  .nav-item--active .nav-item__hint {
    color: var(--content-primary);
    font-weight: 500;
  }

  /* Content */
  .pane__content {
    padding: var(--space-7) var(--space-9);
    overflow-y: auto;
    max-width: 720px;
  }

  .content__header {
    margin-bottom: var(--space-7);
  }
  .content__header h2 {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    line-height: 1.3;
    font-weight: 400;
    color: var(--content-primary);
    margin: 0 0 var(--space-2) 0;
  }
  .content__lede {
    font-size: var(--text-body-size);
    color: var(--content-tertiary);
    line-height: 1.6;
    margin: 0;
  }

  .content__footer {
    margin-top: var(--space-7);
    padding-top: var(--space-5);
    border-top: 1px solid var(--border-subtle);
    display: flex;
    gap: var(--space-3);
  }

  .caption {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    letter-spacing: 0.02em;
  }

  code {
    font-family: var(--font-mono);
    font-size: 0.9em;
    background-color: var(--paper-warm-50);
    border: 1px solid var(--border-subtle);
    padding: 1px 6px;
    border-radius: var(--radius-xs);
    color: var(--content-primary);
  }

  /* Replay */
  .replay {
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    overflow: hidden;
  }
  .replay__row {
    border-bottom: 1px solid var(--border-subtle);
  }
  .replay__row:last-child {
    border-bottom: none;
  }
  .replay__row summary {
    list-style: none;
    cursor: pointer;
    padding: 0 var(--space-4);
    transition: background-color var(--duration-fast) var(--ease-standard);
  }
  .replay__row summary::-webkit-details-marker {
    display: none;
  }
  .replay__row summary:hover {
    background-color: var(--paper-warm-50);
  }
  .replay__detail {
    padding: var(--space-3) var(--space-4) var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    background-color: var(--surface-sunken);
    border-top: 1px solid var(--border-subtle);
  }
  .replay__field {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  /* Adaptive */
  .adaptive {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .adaptive__row {
    padding: var(--space-4);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
  }
  .adaptive__claim {
    font-family: var(--font-serif);
    font-size: var(--text-body-size);
    color: var(--content-primary);
    margin-bottom: var(--space-1);
  }
  .adaptive__evidence {
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    line-height: 1.5;
    margin-bottom: var(--space-3);
  }
  .adaptive__meta {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  /* Permissions */
  .permissions {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .perm {
    padding: var(--space-4);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
  }
  .perm__head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-2);
  }
  .perm__name {
    font-weight: 500;
    color: var(--content-primary);
  }
  .perm__status {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .perm__desc {
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    line-height: 1.5;
    margin: 0 0 var(--space-2) 0;
  }
  .perm__required {
    margin: 0 0 var(--space-3) 0;
  }

  /* Hotkey */
  .hotkey-stage {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-7);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
  }
  .hotkey-display {
    padding: var(--space-4) var(--space-6);
    background-color: var(--surface-base);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
  }
  .hotkey-combo {
    font-family: var(--font-mono);
    font-size: 24px;
    color: var(--content-accent);
    letter-spacing: 0.04em;
  }
  .hotkey-voice {
    margin-top: var(--space-5);
  }

  /* Strength slider */
  .strength {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-4);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    margin-bottom: var(--space-5);
  }
  .strength__options {
    display: flex;
    gap: 2px;
    background-color: var(--paper-warm-50);
    padding: 2px;
    border-radius: var(--radius-sm);
  }
  .strength__opt {
    padding: var(--space-2) var(--space-4);
    background-color: transparent;
    border: none;
    border-radius: var(--radius-xs);
    color: var(--content-secondary);
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    cursor: pointer;
    text-transform: capitalize;
    transition: background-color var(--duration-fast) var(--ease-standard);
  }
  .strength__opt--active {
    background-color: var(--surface-raised);
    color: var(--content-primary);
  }

  /* Autonomy matrix */
  .autonomy {
    display: flex;
    flex-direction: column;
    gap: 1px;
    background-color: var(--border-subtle);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    overflow: hidden;
  }
  .autonomy__row {
    display: grid;
    grid-template-columns: 1fr auto auto;
    gap: var(--space-4);
    align-items: center;
    padding: var(--space-3) var(--space-4);
    background-color: var(--surface-raised);
  }
  .autonomy__task {
    font-size: var(--text-body-size);
    color: var(--content-primary);
  }

  /* Dial (3-state) */
  .autonomy__dial {
    display: flex;
    gap: var(--space-1);
    padding: 2px;
    background-color: var(--paper-warm-50);
    border-radius: var(--radius-sm);
  }
  .dial {
    width: 24px;
    height: 24px;
    border-radius: var(--radius-sm);
    border: 1px solid transparent;
    background-color: transparent;
    cursor: pointer;
    transition: background-color var(--duration-fast) var(--ease-standard);
    position: relative;
  }
  .dial--autonomous.dial--active {
    background-color: var(--success-500);
  }
  .dial--warn.dial--active {
    background-color: var(--warning-500);
  }
  .dial--block.dial--active {
    background-color: var(--error-500);
  }
  .dial:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: 1px;
  }

  /* Backup */
  .backup {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-5);
  }
  .backup-note {
    padding: var(--space-4);
    background-color: var(--surface-sunken);
    border-radius: var(--radius-md);
    border: 1px solid var(--border-subtle);
  }
  .backup-note p {
    margin: 0;
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    line-height: 1.5;
  }

  /* Account */
  .account {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-5);
  }
  .account-note {
    padding: var(--space-4);
    background-color: var(--surface-sunken);
    border-radius: var(--radius-md);
    border: 1px solid var(--border-subtle);
  }
  .account-note p {
    margin: 0;
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    line-height: 1.5;
  }

  @media (max-width: 720px) {
    .pane__layout {
      grid-template-columns: 1fr;
    }
    .pane__nav {
      flex-direction: row;
      overflow-x: auto;
      border-right: none;
      border-bottom: 1px solid var(--border-subtle);
    }
  }
</style>