<script lang="ts">
  import { onMount } from 'svelte'

  import { settings } from '../stores/settings.svelte'
  import { apiKeys } from '../stores/apikeys.svelte'
  import { account } from '../stores/account.svelte'
  import { halt } from '../stores/halt.svelte'
  import { ipc } from '../ipc/client'
  import type { APIKeyMeta, ProviderInfo } from '../ipc/types'
  import { t } from '../i18n'

  import Button from '../components/ui/Button.svelte'
  import Switch from '../components/ui/Switch.svelte'
  import Input from '../components/ui/Input.svelte'
  import Select from '../components/ui/Select.svelte'
  import Slider from '../components/ui/Slider.svelte'
  import Card from '../components/ui/Card.svelte'
  import Badge from '../components/ui/Badge.svelte'
  import Divider from '../components/ui/Divider.svelte'
  import Avatar from '../components/ui/Avatar.svelte'
  import Dialog from '../components/ui/Dialog.svelte'
  import Kbd from '../components/ui/Kbd.svelte'
  import IconButton from '../components/ui/IconButton.svelte'

  type Section = 'account' | 'safety' | 'models' | 'hotkey' | 'voice' | 'sync' | 'hub' | 'channels' | 'adaptive' | 'updates' | 'legal'

  let section = $state<Section>('account')
  let showHaltConfirm = $state(false)
  let showReRunSetup = $state(false)
  let hotkeyValue = $state('')
  let providers = $state<ProviderInfo[]>([])
  let apiKeysList = $state<APIKeyMeta[]>([])

  onMount(() => {
    void settings.refresh()
    void refreshData()
    void account.checkStatus()
    void halt.startPolling()
    return () => halt.stopPolling()
  })

  async function refreshData(): Promise<void> {
    try { providers = await ipc.providersList() } catch { providers = [] }
    try { apiKeysList = await ipc.apiKeysList() } catch { apiKeysList = [] }
  }

  const sections: Array<{ id: Section; label: string }> = [
    { id: 'account',  label: 'Account' },
    { id: 'safety',   label: 'Safety' },
    { id: 'models',   label: 'Models' },
    { id: 'hotkey',   label: 'Hotkey' },
    { id: 'voice',    label: 'Voice' },
    { id: 'sync',     label: 'Sync' },
    { id: 'hub',      label: 'Hub' },
    { id: 'channels', label: 'Channels' },
    { id: 'adaptive', label: 'Adaptive' },
    { id: 'updates',  label: 'Updates' },
    { id: 'legal',    label: 'Legal' },
  ]

  const modelOptions = $derived(
    providers.flatMap((p) =>
      p.models.map((m) => ({ value: `${p.name}:${m.id}`, label: `${p.name} · ${m.id}` }))
    )
  )

  async function reRunSetup(): Promise<void> {
    showReRunSetup = false
    window.dispatchEvent(new CustomEvent('synaptic:show-onboarding'))
  }

  async function haltAgent(): Promise<void> {
    showHaltConfirm = false
    await halt.halt('user requested from settings')
  }

  function pickSection(s: Section): void { section = s }

  function setHotkey(): void {
    if (!hotkeyValue) return
    settings.config?.hotkey && (settings.config.hotkey.overlay = hotkeyValue)
    void settings.save({ hotkey: { ...(settings.config?.hotkey ?? { overlay: '' }), overlay: hotkeyValue } })
  }
</script>

<div class="settings">
  <aside class="settings-nav">
    <header class="settings-nav-header">
      <h2 class="settings-nav-title">Settings</h2>
      <p class="settings-nav-sub">Everything Condura knows about you.</p>
    </header>
    <nav class="settings-nav-list">
      {#each sections as s (s.id)}
        <button
          type="button"
          class="settings-nav-item"
          class:active={section === s.id}
          onclick={() => pickSection(s.id)}
        >
          <span class="settings-nav-dot"></span>
          <span class="settings-nav-label">{s.label}</span>
        </button>
      {/each}
    </nav>
  </aside>

  <main class="settings-content">
    <div class="settings-content-inner">
      {#if section === 'account'}
        <header class="settings-header">
          <h1>Account</h1>
          <p>Your Condura identity and connected services.</p>
        </header>

        <Card elevation={2} padding="lg">
          {#if account.isSignedIn}
            <div class="settings-account-row">
              <Avatar name={account.email} size="lg" status="online" />
              <div class="settings-account-meta">
                <div class="settings-account-name">{account.email}</div>
                <div class="settings-account-provider">Signed in via {account.provider}</div>
              </div>
              <Button variant="secondary" onclick={() => account.signOut()}>Sign out</Button>
            </div>
          {:else}
            <div class="settings-account-empty">
              <Avatar name="?" size="lg" />
              <div>
                <div class="settings-account-name">You're signed out</div>
                <div class="settings-account-provider">Sign in to sync settings across devices.</div>
              </div>
              <Button variant="primary" onclick={() => { window.location.hash = '#/settings?signin=1' }}>Sign in</Button>
            </div>
          {/if}
        </Card>

        <Divider label="danger zone" />

        <Card elevation={1} padding="md">
          <div class="settings-danger-row">
            <div>
              <div class="settings-row-title">Re-run setup</div>
              <div class="settings-row-sub">Walk through the welcome flow again. Your data is preserved.</div>
            </div>
            <Button variant="secondary" onclick={() => { showReRunSetup = true }}>Re-run</Button>
          </div>
          <Divider />
          <div class="settings-danger-row">
            <div>
              <div class="settings-row-title settings-row-title-danger">Halt the agent</div>
              <div class="settings-row-sub">Cancel every in-flight action. Requires terminal confirmation to resume.</div>
            </div>
            <Button variant="danger" onclick={() => { showHaltConfirm = true }}>Halt</Button>
          </div>
        </Card>

      {:else if section === 'safety'}
        <header class="settings-header">
          <h1>Safety</h1>
          <p>The five modules that decide what the agent can and cannot do.</p>
        </header>

        <div class="settings-grid-2">
          <Card elevation={2} padding="md">
            <Badge tone="success" dot>Active</Badge>
            <h3 class="settings-card-title">Gatekeeper</h3>
            <p class="settings-card-body">Deterministic rules engine. Every WRITE, NETWORK, and DESTRUCTIVE action passes through it.</p>
            <Switch checked={true} label="Enforce gatekeeper" description="Required — do not disable." disabled />
          </Card>
          <Card elevation={2} padding="md">
            <Badge tone="success" dot>Active</Badge>
            <h3 class="settings-card-title">Blast-radius classifier</h3>
            <p class="settings-card-body">Tags every action READ / WRITE / NETWORK / DESTRUCTIVE before it reaches the gatekeeper.</p>
            <Switch checked={true} label="Classify all actions" disabled />
          </Card>
          <Card elevation={2} padding="md">
            <Badge tone="success" dot>Active</Badge>
            <h3 class="settings-card-title">Behavioral anomaly detector</h3>
            <p class="settings-card-body">Hard-pauses the agent on speed, loop, duration, or new-endpoint anomalies.</p>
            <Switch checked={true} label="Anomaly detector" />
          </Card>
          <Card elevation={2} padding="md">
            <Badge tone="success" dot>Active</Badge>
            <h3 class="settings-card-title">HMAC-chained audit log</h3>
            <p class="settings-card-body">Append-only, tamper-evident. View the full chain under Audit.</p>
            <Switch checked={true} label="Write audit events" />
          </Card>
          <Card elevation={2} padding="md">
            <Badge tone="success" dot>Active</Badge>
            <h3 class="settings-card-title">Sensitive site detector</h3>
            <p class="settings-card-body">Prompts before any action on banking, health, or credential surfaces.</p>
            <Switch checked={true} label="Detect sensitive sites" />
          </Card>
          <Card elevation={2} padding="md">
            <Badge tone="warn" dot>Default warn</Badge>
            <h3 class="settings-card-title">Autonomy matrix</h3>
            <p class="settings-card-body">Per-app + per-task-type autonomy. Default is "warn" everywhere; tune per-cell.</p>
            <Button variant="secondary" size="sm">Edit matrix</Button>
          </Card>
        </div>

      {:else if section === 'models'}
        <header class="settings-header">
          <h1>Models</h1>
          <p>Default model and per-provider configuration.</p>
        </header>

        <Card elevation={2} padding="md">
          <Select
            label="Default model"
            value={settings.config?.llm.providers ? Object.entries(settings.config.llm.providers).find(([_, p]) => p.enabled)?.[0] + ':' + Object.entries(settings.config.llm.providers).find(([_, p]) => p.enabled)?.[1].default_model : ''}
            options={modelOptions}
          />
          <Select
            label="Provider"
            value=""
            options={providers.map(p => ({ value: p.name, label: p.name }))}
            placeholder="Choose a provider"
          />
        </Card>

        <Card elevation={1} padding="md">
          <h3 class="settings-card-title">Stored API keys</h3>
          {#if apiKeysList.length === 0}
            <p class="settings-card-body">No API keys yet. Add one to unlock that provider.</p>
          {:else}
            {#each apiKeysList as k (k.id)}
              <div class="settings-key-row">
                <div>
                  <div class="settings-row-title">{k.provider}</div>
                  <div class="settings-row-sub">{k.label}</div>
                </div>
                <Badge tone={k.has_token ? 'success' : 'warn'}>{k.has_token ? 'set' : 'empty'}</Badge>
                <IconButton ariaLabel="Delete key" onclick={() => apiKeys.remove(k.id)}>
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
                    <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                  </svg>
                </IconButton>
              </div>
            {/each}
          {/if}
          <Button variant="primary" size="sm">+ Add API key</Button>
        </Card>

      {:else if section === 'hotkey'}
        <header class="settings-header">
          <h1>Hotkey</h1>
          <p>The combo that opens the overlay. Pick something you won't press by accident.</p>
        </header>

        <Card elevation={2} padding="md">
          <Input label="Capture a combo" placeholder="Press a key combo" bind:value={hotkeyValue} />
          <div class="settings-hotkey-presets">
            <button type="button" class="settings-hotkey-chip" onclick={() => { hotkeyValue = 'Option+Option' }}>
              <Kbd label="⌥" /><span>+</span><Kbd label="⌥" />
            </button>
            <button type="button" class="settings-hotkey-chip" onclick={() => { hotkeyValue = 'Cmd+Shift+Space' }}>
              <Kbd label="⌘" /><span>+</span><Kbd label="⇧" /><span>+</span><Kbd label="Space" />
            </button>
            <button type="button" class="settings-hotkey-chip" onclick={() => { hotkeyValue = 'Ctrl+Space' }}>
              <Kbd label="⌃" /><span>+</span><Kbd label="Space" />
            </button>
          </div>
          <Button variant="primary" disabled={!hotkeyValue} onclick={setHotkey}>Save combo</Button>
        </Card>

      {:else if section === 'voice'}
        <header class="settings-header">
          <h1>Voice</h1>
          <p>Wake word, microphone, and speech-to-text settings.</p>
        </header>

        <Card elevation={2} padding="md">
          <Switch checked={true} label="Wake word" description='"hey condura" — local, runs offline.' />
          <Switch checked={true} label="Push-to-talk" description="Hold the hotkey to talk instead of using the wake word." />
          <Switch checked={true} label="Live transcription" description="Show partial transcripts while the user is speaking." />
          <Select
            label="Speech-to-text backend"
            value="whisper.cpp"
            options={[
              { value: 'whisper.cpp', label: 'whisper.cpp (local)' },
              { value: 'openai',     label: 'OpenAI Whisper (cloud)' },
            ]}
          />
          <Button variant="secondary" size="sm">Test microphone</Button>
        </Card>

      {:else if section === 'sync'}
        <header class="settings-header">
          <h1>Sync</h1>
          <p>End-to-end encrypted peer-to-peer sync. No central server.</p>
        </header>

        <Card elevation={2} padding="md">
          <p class="settings-card-body">Configure paired devices and P2P discovery on the Sync route.</p>
          <Button variant="primary" onclick={() => { window.location.hash = '#/sync' }}>Open sync</Button>
        </Card>

      {:else if section === 'hub'}
        <header class="settings-header">
          <h1>Hub</h1>
          <p>Public Skills Hub at hub.condura.app. Browse and install curated skills.</p>
        </header>

        <Card elevation={2} padding="md">
          <Switch checked={false} label="Auto-update installed skills" description="Receive patches and new versions automatically." />
          <Select
            label="Trust level"
            value="official"
            options={[
              { value: 'official',     label: 'Official only' },
              { value: 'community',    label: 'Official + community' },
              { value: 'experimental', label: 'All (incl. experimental)' },
            ]}
          />
          <Button variant="primary" onclick={() => { window.location.hash = '#/hub' }}>Browse Hub</Button>
        </Card>

      {:else if section === 'channels'}
        <header class="settings-header">
          <h1>Channels</h1>
          <p>Connect Condura to messaging surfaces so the agent can reply on your behalf.</p>
        </header>

        <Card elevation={2} padding="md">
          <p class="settings-card-body">Manage Telegram, Slack, Discord, iMessage, WhatsApp on the Channels route.</p>
          <Button variant="primary" onclick={() => { window.location.hash = '#/channels' }}>Open channels</Button>
        </Card>

      {:else if section === 'adaptive'}
        <header class="settings-header">
          <h1>Adaptive engine</h1>
          <p>How much Condura learns from your behavior.</p>
        </header>

        <Card elevation={2} padding="md">
          <Select
            label="Strength"
            value="balanced"
            options={[
              { value: 'off',       label: 'Off — never apply learned preferences' },
              { value: 'cautious',  label: 'Cautious — apply only when very confident' },
              { value: 'balanced',  label: 'Balanced (default)' },
              { value: 'aggressive',label: 'Aggressive — apply anything above 60%' },
            ]}
          />
          <Switch checked={true} label="Weekly review reminder" description="See what the dialectic has inferred about you." />
          <Button variant="danger" size="sm">Forget everything and start fresh</Button>
        </Card>

      {:else if section === 'updates'}
        <header class="settings-header">
          <h1>Updates</h1>
          <p>Auto-update channel and version information.</p>
        </header>

        <Card elevation={2} padding="md">
          <div class="settings-row">
            <div>
              <div class="settings-row-title">Current version</div>
              <div class="settings-row-sub">v0.1.0</div>
            </div>
            <Badge tone="info">stable</Badge>
          </div>
          <Divider />
          <Select
            label="Channel"
            value="stable"
            options={[
              { value: 'stable', label: 'Stable' },
              { value: 'beta',   label: 'Beta' },
              { value: 'dev',    label: 'Dev' },
            ]}
          />
          <Switch checked={false} label="Auto-apply updates" description="Apply signed updates in the background and restart at next idle." />
          <Button variant="primary">Check for updates</Button>
        </Card>

      {:else if section === 'legal'}
        <header class="settings-header">
          <h1>Legal</h1>
          <p>License and privacy documents.</p>
        </header>

        <Card elevation={2} padding="md">
          <div class="settings-row">
            <div>
              <div class="settings-row-title">Freeware EULA</div>
              <div class="settings-row-sub">Synaptic Freeware License v1</div>
            </div>
            <Button variant="secondary" size="sm" onclick={() => { window.location.hash = '#/about' }}>View</Button>
          </div>
          <Divider />
          <div class="settings-row">
            <div>
              <div class="settings-row-title">Privacy policy</div>
              <div class="settings-row-sub">Local-first. No telemetry. Your data stays on your machine.</div>
            </div>
            <Button variant="secondary" size="sm">View</Button>
          </div>
        </Card>
      {/if}
    </div>
  </main>
</div>

<Dialog bind:open={showHaltConfirm} title="Halt the agent?" description="All in-flight actions will be canceled. Resume requires a terminal confirmation." size="sm">
  <p class="settings-dialog-body">This stops every active stream, every queued action, and every pending consent ticket. The audit log records the halt.</p>
  {#snippet footer()}
    <Button variant="ghost" onclick={() => { showHaltConfirm = false }}>Cancel</Button>
    <Button variant="danger" onclick={haltAgent}>Halt</Button>
  {/snippet}
</Dialog>

<Dialog bind:open={showReRunSetup} title="Re-run setup?" description="Walk through the welcome flow again. Your data is preserved." size="sm">
  {#snippet footer()}
    <Button variant="ghost" onclick={() => { showReRunSetup = false }}>Cancel</Button>
    <Button variant="primary" onclick={reRunSetup}>Re-run</Button>
  {/snippet}
</Dialog>

<style>
  .settings {
    display: grid;
    grid-template-columns: 220px 1fr;
    height: 100%;
    min-height: 0;
  }

  .settings-nav {
    background: var(--surface-1);
    border-right: 1px solid var(--border);
    padding: var(--space-4) var(--space-2);
    overflow-y: auto;
  }
  .settings-nav-header { padding: 0 var(--space-3) var(--space-4); }
  .settings-nav-title {
    font-family: var(--font-display);
    font-size: var(--size-xl);
    font-weight: var(--weight-medium);
    color: var(--text);
    letter-spacing: var(--tracking-tight);
    margin: 0;
  }
  .settings-nav-sub {
    font-size: var(--size-xs);
    color: var(--text-muted);
    margin-top: 4px;
  }

  .settings-nav-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .settings-nav-item {
    appearance: none;
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    padding: var(--space-3);
    color: var(--text-muted);
    cursor: pointer;
    text-align: left;
    display: flex;
    align-items: center;
    gap: var(--space-3);
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    transition: background-color var(--transition-fast) ease, color var(--transition-fast) ease, border-color var(--transition-fast) ease;
  }
  .settings-nav-item:hover { background: var(--surface-2); color: var(--text); }
  .settings-nav-item.active {
    background: var(--surface-2);
    color: var(--text);
    border-color: var(--border-focus);
  }
  .settings-nav-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--border-strong);
    flex-shrink: 0;
  }
  .settings-nav-item.active .settings-nav-dot {
    background: var(--accent);
    box-shadow: 0 0 6px var(--accent-glow);
  }

  .settings-content {
    overflow-y: auto;
    padding: var(--space-7);
    background: var(--bg);
  }
  .settings-content-inner {
    max-width: 760px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  .settings-header h1 {
    font-family: var(--font-display);
    font-size: var(--size-3xl);
    font-weight: var(--weight-medium);
    color: var(--text);
    letter-spacing: var(--tracking-tighter);
    margin: 0;
  }
  .settings-header p {
    color: var(--text-muted);
    font-size: var(--size-md);
    margin-top: var(--space-2);
  }

  .settings-card-title {
    font-family: var(--font-display);
    font-size: var(--size-lg);
    font-weight: var(--weight-medium);
    color: var(--text);
    letter-spacing: var(--tracking-tight);
    margin: var(--space-3) 0 var(--space-2);
  }
  .settings-card-body {
    color: var(--text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-normal);
    margin-bottom: var(--space-3);
  }

  .settings-grid-2 {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: var(--space-4);
  }

  .settings-row,
  .settings-account-row,
  .settings-account-empty,
  .settings-key-row,
  .settings-danger-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-3) 0;
  }
  .settings-account-empty { gap: var(--space-3); }
  .settings-account-meta { flex: 1; min-width: 0; }
  .settings-account-name {
    font-size: var(--size-md);
    font-weight: var(--weight-medium);
    color: var(--text);
  }
  .settings-account-provider {
    font-size: var(--size-xs);
    color: var(--text-muted);
    margin-top: 2px;
  }

  .settings-row-title {
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    color: var(--text);
  }
  .settings-row-title-danger { color: var(--error); }
  .settings-row-sub {
    font-size: var(--size-xs);
    color: var(--text-muted);
    margin-top: 2px;
    line-height: var(--leading-normal);
  }

  .settings-hotkey-presets {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .settings-hotkey-chip {
    appearance: none;
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    padding: var(--space-2) var(--space-3);
    color: var(--text);
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    transition: border-color var(--transition-fast) ease, background-color var(--transition-fast) ease;
  }
  .settings-hotkey-chip:hover {
    border-color: var(--border-focus);
    background: var(--surface-3);
  }

  .settings-dialog-body {
    color: var(--text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-normal);
  }
</style>