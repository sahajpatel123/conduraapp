<script lang="ts">
  import Avatar from '../../components/ui/Avatar.svelte'
  import Badge from '../../components/ui/Badge.svelte'
  import Button from '../../components/ui/Button.svelte'
  import Card from '../../components/ui/Card.svelte'
  import CommandPalette from '../../components/ui/CommandPalette.svelte'
  import Dialog from '../../components/ui/Dialog.svelte'
  import Divider from '../../components/ui/Divider.svelte'
  import EmptyState from '../../components/ui/EmptyState.svelte'
  import IconButton from '../../components/ui/IconButton.svelte'
  import Input from '../../components/ui/Input.svelte'
  import Kbd from '../../components/ui/Kbd.svelte'
  import Progress from '../../components/ui/Progress.svelte'
  import SegmentedControl from '../../components/ui/SegmentedControl.svelte'
  import Select from '../../components/ui/Select.svelte'
  import Sheet from '../../components/ui/Sheet.svelte'
  import Skeleton from '../../components/ui/Skeleton.svelte'
  import Slider from '../../components/ui/Slider.svelte'
  import Switch from '../../components/ui/Switch.svelte'
  import Tabs from '../../components/ui/Tabs.svelte'
  import Textarea from '../../components/ui/Textarea.svelte'
  import Toast from '../../components/ui/Toast.svelte'
  import Tooltip from '../../components/ui/Tooltip.svelte'

  let switchOn = $state(true)
  let switchOff = $state(false)
  let dialogOpen = $state(false)
  let sheetOpen = $state(false)
  let paletteOpen = $state(false)
  let tab = $state('overview')
  let seg = $state('daily')
  let sliderVal = $state(64)
  let progressVal = $state(72)
  let indeterminate = $state(true)
  let inputText = $state('')
  let textLong = $state('')
  let selected = $state('sonnet')

  function fireToast(): void {
    // Toast exports `push` from the module — we'll wire that in the next pass.
    // For now this is a visual placeholder until the import works.
  }
</script>

<svelte:head>
  <title>Component smoke · Condura</title>
</svelte:head>

<div class="smoke">
  <header class="smoke-header">
    <h1>Component smoke</h1>
    <p class="smoke-sub">
      Every primitive renders here so we can iterate the design system
      without breaking the app. <Kbd label="⌘K" /> opens the command palette.
    </p>
  </header>

  <Toast />

  <!-- ── Buttons ──────────────────────────────────────── -->
  <section class="smoke-section">
    <h2>Buttons</h2>
    <div class="row">
      <Button variant="primary">Primary</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="accent-ghost">Accent ghost</Button>
      <Button variant="danger">Danger</Button>
    </div>
    <div class="row">
      <Button size="xs">Extra small</Button>
      <Button size="sm">Small</Button>
      <Button size="md">Medium</Button>
      <Button size="lg">Large</Button>
      <Button loading>Saving…</Button>
      <Button disabled>Disabled</Button>
    </div>
    <div class="row">
      <Button variant="primary" iconOnly ariaLabel="Add">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <path d="M12 5v14M5 12h14" />
        </svg>
      </Button>
      <IconButton variant="ghost" ariaLabel="Settings">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <circle cx="12" cy="12" r="3" />
          <path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 0 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 0 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 0 1 0-4h.1A1.7 1.7 0 0 0 4.6 9a1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 0 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 0 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z" />
        </svg>
      </IconButton>
    </div>
  </section>

  <Divider />

  <!-- ── Badges + Kbd + Avatars ──────────────────────── -->
  <section class="smoke-section">
    <h2>Status & identity</h2>
    <div class="row">
      <Badge>Neutral</Badge>
      <Badge tone="accent" dot pulse>Live</Badge>
      <Badge tone="success" dot>Connected</Badge>
      <Badge tone="warn" dot>Awaiting</Badge>
      <Badge tone="error" dot>Halted</Badge>
      <Badge tone="info" dot>Beta</Badge>
    </div>
    <div class="row align-center">
      <Kbd label="⌘" />
      <Kbd label="K" />
      <span class="muted-pair">Press <Kbd label="⌘K" /> to open command palette</span>
    </div>
    <div class="row">
      <Avatar name="Ada Lovelace" size="sm" status="online" />
      <Avatar name="Claude Sonnet" size="md" status="online" />
      <Avatar name="Hermes" size="md" status="busy" />
      <Avatar name="Codex" size="md" status="away" />
      <Avatar name="Ollama" size="lg" status="offline" />
    </div>
  </section>

  <Divider />

  <!-- ── Inputs ──────────────────────────────────────── -->
  <section class="smoke-section">
    <h2>Inputs</h2>
    <div class="grid-2">
      <Input label="Model" placeholder="Pick a model" bind:value={inputText} hint="Default for new sessions" />
      <Select
        label="Variant"
        bind:value={selected}
        options={[
          { value: 'opus',   label: 'Claude Opus 4.8' },
          { value: 'sonnet', label: 'Claude Sonnet 4.6' },
          { value: 'haiku',  label: 'Claude Haiku 4.5' },
          { value: 'local',  label: 'Local Ollama', disabled: true },
        ]}
        hint="Higher-tier models cost more per token"
      />
    </div>
    <Textarea
      label="Composer"
      placeholder="Ask Condura anything…"
      bind:value={textLong}
      rows={4}
      hint="Use @ to mention skills, / for slash commands"
    />
    <div class="grid-2">
      <Slider label="Confidence threshold" bind:value={sliderVal} min={0} max={100} step={1} formatValue={(v) => `${v}%`} />
      <Switch bind:checked={switchOn} label="Auto-approve low-risk actions" description="Skip the consent prompt for trusted apps" />
      <Switch bind:checked={switchOff} label="Verbose logging" />
    </div>
  </section>

  <Divider />

  <!-- ── Tabs + SegmentedControl ─────────────────────── -->
  <section class="smoke-section">
    <h2>Navigation</h2>
    <Tabs
      tabs={[
        { id: 'overview', label: 'Overview' },
        { id: 'activity', label: 'Activity', badge: 12 },
        { id: 'replay',   label: 'Replay' },
        { id: 'audit',    label: 'Audit', disabled: true },
      ]}
      bind:value={tab}
    />
    <p class="muted-pair">Active tab: <code>{tab}</code></p>
    <SegmentedControl
      bind:value={seg}
      options={[
        { value: 'hourly', label: 'Hourly' },
        { value: 'daily',  label: 'Daily' },
        { value: 'weekly', label: 'Weekly' },
      ]}
    />
  </section>

  <Divider />

  <!-- ── Progress + Skeleton ─────────────────────────── -->
  <section class="smoke-section">
    <h2>Progress & loading</h2>
    <Progress value={progressVal} label="Syncing skills" />
    <Progress indeterminate tone="accent" />
    <Progress value={88} tone="success" />
    <Progress value={32} tone="warn" />
    <Progress value={6} tone="error" />
    <div class="row">
      <Skeleton width="120px" height="14px" />
      <Skeleton width="60px" height="14px" rounded="pill" />
      <Skeleton width="200px" height="32px" rounded="md" />
      <Skeleton width="100%" height="60px" rounded="lg" />
    </div>
  </section>

  <Divider />

  <!-- ── Card + EmptyState ──────────────────────────── -->
  <section class="smoke-section">
    <h2>Surfaces</h2>
    <div class="grid-3">
      <Card elevation={1} padding="md">
        <h3 class="card-h">Elevation 1</h3>
        <p class="card-p">Standard card on the canvas.</p>
      </Card>
      <Card elevation={2} padding="md">
        <h3 class="card-h">Elevation 2</h3>
        <p class="card-p">Raised card with hairline border.</p>
      </Card>
      <Card elevation={3} padding="md">
        <h3 class="card-h">Elevation 3</h3>
        <p class="card-p">Elevated card with stronger shadow.</p>
      </Card>
      <Card elevation="glass" padding="md">
        <h3 class="card-h">Glass</h3>
        <p class="card-p">Translucent with backdrop blur.</p>
      </Card>
      <Card elevation="glass-strong" padding="md">
        <h3 class="card-h">Glass strong</h3>
        <p class="card-p">Heavier glass, more blur.</p>
      </Card>
      <Card interactive padding="md" onclick={() => alert('clicked')}>
        <h3 class="card-h">Interactive</h3>
        <p class="card-p">Click me — hovers, lifts, accents.</p>
      </Card>
    </div>
  </section>

  <Divider label="empty state" />

  <section class="smoke-section">
    <Card elevation="glass" padding="none">
      <EmptyState
        title="No skills yet"
        description="Skills are reusable playbooks the agent learns from your work. Install one from the Hub or let the agent create its first."
      >
        {#snippet icon()}
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
            <path d="M12 2L3 7l9 5 9-5-9-5zM3 17l9 5 9-5M3 12l9 5 9-5" />
          </svg>
        {/snippet}
        {#snippet action()}
          <Button variant="primary">Browse Hub</Button>
        {/snippet}
      </EmptyState>
    </Card>
  </section>

  <Divider />

  <!-- ── Tooltip + Dialog + Sheet + CommandPalette ──── -->
  <section class="smoke-section">
    <h2>Overlays</h2>
    <div class="row">
      <Tooltip text="Ask anything" side="top">
        <Button>Hover me</Button>
      </Tooltip>
      <Button variant="primary" onclick={() => dialogOpen = true}>Open dialog</Button>
      <Button variant="secondary" onclick={() => sheetOpen = true}>Open sheet</Button>
      <Button variant="accent-ghost" onclick={() => paletteOpen = true}>Open command palette</Button>
      <Button onclick={fireToast}>Fire toast</Button>
    </div>
  </section>

  <Dialog bind:open={dialogOpen} title="Destructive action" description="Condura wants to run a shell command on your machine.">
    <div class="dialog-row">
      <code>rm -rf ~/Downloads/test-folder</code>
    </div>
    {#snippet footer()}
      <Button variant="ghost" onclick={() => dialogOpen = false}>Deny</Button>
      <Button variant="primary" onclick={() => dialogOpen = false}>Approve once</Button>
    {/snippet}
  </Dialog>

  <Sheet bind:open={sheetOpen} title="Pair a new device" side="right">
    <p class="card-p">Scan the QR on the other device, or type its 6-digit PIN below.</p>
    <Input label="Pairing PIN" placeholder="123 456" />
    <div class="row" style="margin-top: var(--space-4);">
      <Button variant="primary">Connect</Button>
      <Button variant="ghost" onclick={() => sheetOpen = false}>Cancel</Button>
    </div>
  </Sheet>

  <CommandPalette
    bind:open={paletteOpen}
    items={[
      { id: 'chat',   label: 'Go to Chat',          hint: 'Open the conversation view', group: 'Navigation', onselect: () => {} },
      { id: 'audit',  label: 'Open Audit log',      hint: 'Review HMAC-chained events', group: 'Navigation', onselect: () => {} },
      { id: 'halt',   label: 'Halt the agent',      hint: 'Stop all in-flight actions', group: 'Safety',     onselect: () => {} },
      { id: 'sync',   label: 'Pair a new device',   hint: 'P2P sync',                    group: 'Safety',     onselect: () => {} },
      { id: 'theme',  label: 'Switch to light theme', group: 'Settings',                 onselect: () => {} },
    ]}
  />
</div>

<style>
  .smoke {
    max-width: 980px;
    margin: 0 auto;
    padding: var(--space-9) var(--space-7);
    display: flex;
    flex-direction: column;
    gap: var(--space-7);
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }

  .smoke-header {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .smoke-header h1 {
    font-family: var(--font-display);
    font-size: var(--size-3xl);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tighter);
    color: var(--text);
  }
  .smoke-sub {
    color: var(--text-muted);
    font-size: var(--size-md);
    line-height: var(--leading-normal);
    max-width: 60ch;
  }

  .smoke-section {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .smoke-section h2 {
    font-family: var(--font-mono);
    font-size: var(--size-2xs);
    text-transform: uppercase;
    letter-spacing: var(--tracking-widest);
    color: var(--text-faint);
    font-weight: var(--weight-semibold);
  }

  .row {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
    align-items: center;
  }
  .align-center { align-items: center; }

  .grid-2 {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-4);
  }
  .grid-3 {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: var(--space-4);
  }

  .card-h {
    font-family: var(--font-display);
    font-size: var(--size-lg);
    font-weight: var(--weight-medium);
    color: var(--text);
    margin-bottom: var(--space-2);
    letter-spacing: var(--tracking-tight);
  }
  .card-p {
    color: var(--text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-normal);
  }

  .dialog-row {
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-3) var(--space-4);
    font-family: var(--font-mono);
    font-size: var(--size-sm);
    color: var(--text);
  }

  .muted-pair {
    color: var(--text-muted);
    font-size: var(--size-sm);
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
  }

  code {
    font-family: var(--font-mono);
    font-size: 0.9em;
    background: var(--surface-1);
    border: 1px solid var(--border);
    padding: 2px 6px;
    border-radius: var(--radius-xs);
    color: var(--text);
  }
</style>