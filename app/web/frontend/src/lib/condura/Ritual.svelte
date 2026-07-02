<script lang="ts">
  import { onMount } from 'svelte';
  import { onboarding } from '../stores/onboarding.svelte';
  import { account } from '../stores/account.svelte';
  import { ipc } from '../ipc/client';
  import { FALLBACK_EULA_TEXT, FALLBACK_EULA_VERSION } from './fallbackEula';
  import Pulse from './Pulse.svelte';
  import Glyph from './Glyph.svelte';
  import ConstellationNode, { type NodeId, type NodeState } from './ConstellationNode.svelte';
  import HoverPreview from './HoverPreview.svelte';
  import SidePanel from './SidePanel.svelte';
  import Button from './Button.svelte';

  /**
   * Condura · The Ritual (Constellation-as-Room redesign).
   * ──────────────────────────────────────────────────────────────────────
   * The first-run flow collapses to two screens:
   *
   *   1. **Gate** — EULA + wax seal stamp. Single escape: `not now · quit`.
   *      The seal IS the legal consent gesture (DIRECTION.md §2.2).
   *   2. **Constellation** — the room. Six live nodes on a ring (Perceive,
   *      Power, Summon, Voice, Threads, Account), each opening a side
   *      panel with the same per-step UI from the old wizard — relocated,
   *      not redesigned. The `Enter Condura →` pill at the bottom is
   *      always enabled; the only soft-lock is the Summon (hotkey) node,
   *      which has no silent default (locked decision #8).
   *
   * The 9-step forced sequence dissolves. Configure, don't comply.
   */

  let { onComplete }: { onComplete?: (route?: string) => void } = $props();

  // ── screen machine ───────────────────────────────────────────────────
  type Screen = 'gate' | 'constellation' | 'done';
  let screen = $state<Screen>('gate');
  let dissolving = $state(false);

  // ── Gate state ──────────────────────────────────────────────────────
  let eulaText = $state('');
  let eulaVersion = $state(FALLBACK_EULA_VERSION);
  let eulaScrolled = $state(false);
  let eulaAccepted = $state(false);
  let eulaEl = $state<HTMLDivElement | undefined>(undefined);
  let eulaReadPct = $state(0);
  let stamped = $state(false);

  function effectiveEula(): { text: string; version: string } {
    const live = onboarding.eula;
    if (live && typeof live.text === 'string' && live.text.trim().length > 200) {
      return { text: live.text, version: live.version || eulaVersion };
    }
    return { text: FALLBACK_EULA_TEXT, version: eulaVersion };
  }

  let eulaIsFallback = $derived(
    !(onboarding.eula && typeof onboarding.eula.text === 'string' && onboarding.eula.text.trim().length > 200)
  );

  function recomputeEulaScroll(): void {
    if (!eulaEl) return;
    const max = eulaEl.scrollHeight - eulaEl.clientHeight;
    const ratio = max > 0 ? eulaEl.scrollTop / max : 1;
    eulaReadPct = Math.min(100, Math.max(0, ratio * 100));
    if (max <= 4 || max - eulaEl.scrollTop <= 8) {
      if (ratio > 0.4) eulaScrolled = true;
    }
    if (max > 80 && ratio >= 0.85) eulaScrolled = true;
  }

  $effect(() => {
    void eulaText;
    requestAnimationFrame(() => recomputeEulaScroll());
  });

  let canStamp = $derived(
    !!eulaText && eulaText.trim().length > 100 && eulaAccepted && !onboarding.busy
  );

  async function loadEula(): Promise<void> {
    try {
      await onboarding.loadEula();
    } catch {
      /* fall back to the bundled copy below */
    }
    const eff = effectiveEula();
    eulaText = eff.text;
    eulaVersion = eff.version;
    requestAnimationFrame(() => recomputeEulaScroll());
  }

  function stampSeal(): void {
    if (!canStamp) return;
    stamped = true;
    void onboarding.acceptEula(eulaVersion).catch((e) => console.error(e));
    setTimeout(() => (screen = 'constellation'), 650);
  }

  function quitApp(): void {
    try {
      (window as unknown as { close?: () => void }).close?.();
    } catch {
      /* ignore */
    }
  }

  // ── Constellation: per-node state ───────────────────────────────────
  // Each node has its own state machine + label + glyph. Probes run on
  // mount; writes are persisted via the daemon where applicable.
  type NodeMeta = {
    id: NodeId;
    label: string;
    glyph: string;
    eyebrow: string;
    headline: string;
    skippable: boolean;
  };

  const NODES: NodeMeta[] = [
    { id: 'perceive', label: 'Perceive', glyph: 'bolt',       eyebrow: '— Permission to perceive', headline: 'It reads the screen to act.' },
    { id: 'power',    label: 'Power',    glyph: 'power',      eyebrow: '— A source of power',      headline: 'How will it think?' },
    { id: 'summon',   label: 'Summon',   glyph: 'key',        eyebrow: '— The summoning',          headline: 'How will you call it?' },
    { id: 'voice',    label: 'Voice',    glyph: 'mic',        eyebrow: '— Your voice, its ear',    headline: 'Say "hey condura."' },
    { id: 'threads',  label: 'Threads',  glyph: 'channels',   eyebrow: '— Threads outward',        headline: 'Reach, on your terms.' },
    { id: 'account',  label: 'Account',  glyph: 'account',    eyebrow: '— Your account',           headline: 'Optional. Your dashboard.' },
  ];

  type PermRow = { kind: string; status: 'granted' | 'denied' | 'unknown' };
  const REQUIRED_PERMS = ['accessibility', 'screen_recording'];

  let nodeState = $state<Record<NodeId, NodeState>>({
    perceive: 'probing',
    power: 'probing',
    summon: 'probing',
    voice: 'probing',
    threads: 'skipped',
    account: 'skipped',
  });
  let nodeWired = $state<Record<NodeId, boolean>>({
    perceive: false,
    power: false,
    summon: false,
    voice: false,
    threads: false,
    account: false,
  });

  let permRows = $state<PermRow[]>([]);
  let permError = $state<string | null>(null);
  let powerProbe = $state<{ ollama_reachable: boolean; ollama_models: string[]; recommended: string } | null>(null);
  let voiceProbe = $state<{ mic_available: boolean; ready: boolean; wake_word: string } | null>(null);
  let voiceError = $state<string | null>(null);
  let accountLoaded = $state(false);

  let powerChoice = $state<'local' | 'apikey' | 'sub'>('local');
  let combo = $state<string>(onboarding.daemon?.steps?.hotkey?.data ?? onboarding.hotkeyValue ?? '');
  const PRESETS = ['Option+Option', 'Cmd+Shift+Space', 'Ctrl+Space'];
  let recording = $state(false);
  let voiceEnabled = $state(false);
  let tried = $state(false);

  const CHANNELS = [
    { id: 'telegram', name: 'Telegram', ready: true },
    { id: 'whatsapp', name: 'WhatsApp', ready: false },
    { id: 'slack',    name: 'Slack',    ready: false },
    { id: 'discord',  name: 'Discord',  ready: false },
    { id: 'imsg',     name: 'iMessage', ready: false },
  ];
  let channelPick = $state<Set<string>>(new Set());

  let email = $state('');
  let accountSent = $state(false);
  let accountBusy = $state(false);
  let accountError = $state<string | null>(null);

  // ── Hover preview ───────────────────────────────────────────────────
  let hovered = $state<NodeId | null>(null);
  let selected = $state<NodeId | null>(null);

  function indicatorFor(id: NodeId): string {
    switch (id) {
      case 'perceive': {
        const both = REQUIRED_PERMS.every((k) => permRows.find((r) => r.kind === k)?.status === 'granted');
        if (permError) return '…';
        return both && nodeWired.perceive ? 'wired' : 'pending';
      }
      case 'power': {
        if (!powerProbe) return '…';
        return powerProbe.ollama_reachable ? 'local · detected' : 'local · none';
      }
      case 'summon': return combo ? '⌘ ' + combo.split('+')[0].toLowerCase() : 'unset';
      case 'voice':   return voiceEnabled ? 'on' : voiceProbe?.mic_available ? 'ready' : 'off';
      case 'threads': {
        const n = channelPick.size;
        return n === 0 ? '0 of 5' : `${n} of 5 · picked`;
      }
      case 'account': return accountSent ? 'link sent' : account.isSignedIn ? 'signed in' : 'signed out';
    }
  }

  function stateFor(id: NodeId): NodeState {
    if (id === 'summon' && combo === '') return 'probing';
    if (id === 'threads' && channelPick.size === 0 && !nodeWired.threads) return 'skipped';
    if (id === 'account' && !accountSent && !account.isSignedIn && !accountBusy) return 'skipped';
    return nodeWired[id] ? 'done' : nodeState[id];
  }

  let permPoll = 0;

  async function refreshPerms(): Promise<void> {
    try {
      const rows = (await ipc.permissionsStatus()) as PermRow[];
      permRows = rows;
      permError = null;
      const both = REQUIRED_PERMS.every((k) => rows.find((r) => r.kind === k)?.status === 'granted');
      nodeState.perceive = 'done';
      nodeWired.perceive = both;
    } catch (e) {
      permError = String(e);
      permRows = REQUIRED_PERMS.map((k) => ({ kind: k, status: 'unknown' as const }));
      nodeState.perceive = 'error';
    }
  }

  async function openPermSettings(kind: string): Promise<void> {
    try {
      const g = await ipc.permissionsGuide(kind);
      const url = g.deep_link ?? g.help_url;
      if (!url) return;
      const r = (window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } }).runtime;
      if (r?.BrowserOpenURL) {
        r.BrowserOpenURL(url);
        return;
      }
      try { await navigator.clipboard.writeText(url); } catch { /* ignore */ }
      window.open(url, '_blank');
    } catch (e) {
      console.error('permissionsGuide failed', e);
    }
  }

  async function probePower(): Promise<void> {
    try {
      const p = (await ipc.onboardingProbePower()) as typeof powerProbe;
      powerProbe = p;
      nodeState.power = 'done';
    } catch {
      powerProbe = { ollama_reachable: false, ollama_models: [], recommended: 'none' };
      nodeState.power = 'done';
    }
  }

  async function probeVoice(): Promise<void> {
    try {
      const v = (await ipc.onboardingProbeVoice()) as typeof voiceProbe;
      voiceProbe = v;
      nodeState.voice = 'done';
    } catch (e) {
      voiceError = String(e);
      voiceProbe = { mic_available: false, ready: false, wake_word: 'hey condura' };
      nodeState.voice = 'error';
    }
  }

  function choosePower(c: typeof powerChoice): void {
    powerChoice = c;
    nodeWired.power = true;
    nodeState.power = 'done';
  }

  function startRecording(): void { recording = true; }
  function onRecordKey(e: KeyboardEvent): void {
    if (!recording) return;
    e.preventDefault();
    e.stopPropagation();
    if (e.key === 'Escape') { recording = false; return; }
    const parts: string[] = [];
    if (e.metaKey) parts.push('Cmd');
    if (e.ctrlKey) parts.push('Ctrl');
    if (e.altKey) parts.push('Option');
    if (e.shiftKey) parts.push('Shift');
    let k = e.key;
    if (k === ' ') k = 'Space';
    if (k.length === 1) k = k.toUpperCase();
    if (!['Meta', 'Control', 'Alt', 'Shift'].includes(e.key)) parts.push(k);
    if (parts.length > 0) {
      combo = parts.join('+');
      recording = false;
    }
  }
  function usePreset(p: string): void { combo = p; recording = false; }
  function tryCombo(): void { tried = true; setTimeout(() => (tried = false), 700); }
  function saveHotkey(): void {
    if (!combo) return;
    onboarding.setHotkey(combo);
    void onboarding.saveHotkey().catch((e) => console.error(e));
    nodeWired.summon = true;
    nodeState.summon = 'done';
    selected = null;
  }

  function enableVoice(): void {
    voiceEnabled = true;
    nodeWired.voice = true;
    nodeState.voice = 'done';
    selected = null;
  }

  function toggleChannel(id: string, ready: boolean): void {
    if (!ready) return;
    const next = new Set(channelPick);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    channelPick = next;
    nodeWired.threads = channelPick.size > 0;
    nodeState.threads = channelPick.size > 0 ? 'done' : 'skipped';
  }

  async function sendMagicLink(): Promise<void> {
    if (!email.trim()) return;
    accountBusy = true;
    accountError = null;
    try {
      await account.signInWithEmail(email.trim(), 'en', window.location.origin);
      accountSent = true;
      nodeWired.account = true;
      nodeState.account = 'done';
    } catch (e) {
      accountError = String(e);
    } finally {
      accountBusy = false;
    }
  }

  function skipAccount(): void {
    nodeState.account = 'skipped';
    selected = null;
  }

  function markNodeDone(id: NodeId): void {
    nodeWired[id] = true;
    nodeState[id] = 'done';
  }

  function openNode(id: NodeId): void {
    // Summon is a soft-lock: a clear CTA leads the user back to record.
    selected = id;
  }
  function closePanel(): void { selected = null; }

  // ── Enter Condura — pill soft-lock + handoff ────────────────────────
  let hotkeySet = $derived(!!combo);
  let canEnter = $derived(true);

  function enterCondura(): void {
    dissolving = true;
    void onboarding.finish({
      hotkey: onboarding.daemon?.steps?.hotkey?.data ?? onboarding.hotkeyValue ?? combo,
      eula_version: onboarding.eulaVersion ?? 'v1',
      permissions_skipped: onboarding.daemon?.steps?.permissions?.status === 'skipped',
    }).catch((e) => console.error(e));
    try {
      localStorage.setItem('condura-ritual-seen', '1');
    } catch {
      /* ignore */
    }
    setTimeout(() => {
      const dest = powerChoice === 'apikey' || powerChoice === 'sub' ? '#/settings' : undefined;
      onComplete?.(dest);
      screen = 'done';
    }, 900);
  }

  // ── Cycle ───────────────────────────────────────────────────────────
  let mountTime = 0;
  let ringSettled = $state(false);

  onMount(() => {
    mountTime = Date.now();
    try {
      void onboarding.sync();
    } catch (e) {
      console.warn('onboarding sync failed', e);
    }
    void loadEula();
    void refreshPerms();
    void probePower();
    void probeVoice();
    try {
      void account.checkStatus?.().finally(() => (accountLoaded = true));
    } catch {
      accountLoaded = true;
    }
    permPoll = window.setInterval(refreshPerms, 2000);
    window.addEventListener('keydown', onRecordKey, true);
    // Stagger the ring nodes in; let them settle.
    setTimeout(() => (ringSettled = true), 8 * NODES.length * 80 + 320);

    return () => {
      window.clearInterval(permPoll);
      window.removeEventListener('keydown', onRecordKey, true);
    };
  });

  function nodeDelay(index: number): number {
    return index * 80;
  }
</script>

<!-- ────────────────────────────────────────────────────────────────────
     Gate · Screen 1 · EULA + wax seal
     ──────────────────────────────────────────────────────────────────── -->
{#if screen === 'gate'}
  <div class="gate surface-paper">
    <div class="pw-bloom"></div>
    <div class="pw-grain"></div>

    <button class="skip-note" onclick={quitApp} aria-label="Quit the app">
      <span>not now · quit</span><span class="arr">→</span>
    </button>

    <div class="gate-content">
      <div class="eyebrow">— The terms</div>
      <h1 class="headline">First, the terms.</h1>
      <p class="sub">
        Free for personal and commercial use, no tracking, no lock-in. Read what
        that means — then stamp the seal.
      </p>

      <div class="eula" bind:this={eulaEl} onscroll={recomputeEulaScroll} tabindex="0" role="region" aria-label="End-user license agreement">
        <div class="eula-read" style:height="{eulaReadPct}%"></div>
        <pre class="eula-text">{eulaText || 'Loading the license…'}</pre>
      </div>

      <label class="check">
        <input
          type="checkbox"
          bind:checked={eulaAccepted}
          disabled={!eulaText}
        />
        <span>I have read and accept the Condura Freeware EULA</span>
      </label>

      {#if eulaIsFallback}
        <p class="eula-offline-note">
          Read offline (daemon unreachable) — your acceptance will be replayed to
          Condura on next boot.
        </p>
      {/if}

      <div class="seal-row">
        <button
          class="seal"
          class:stamped
          disabled={!canStamp}
          onclick={stampSeal}
          aria-label="Stamp to accept"
        >
          <span class="seal-c">C</span>
        </button>
        <div class="seal-text">
          <span class="st-1">
            {#if stamped}
              Accepted · thank you
            {:else if !eulaText}
              Loading the license…
            {:else if !eulaScrolled}
              Scroll to the bottom first
            {:else}
              Stamp to accept
            {/if}
          </span>
          <span class="st-2">a considered act — not a click</span>
        </div>
      </div>
    </div>
  </div>

<!-- ────────────────────────────────────────────────────────────────────
     Constellation · Screen 2 · the room
     ──────────────────────────────────────────────────────────────────── -->
{:else if screen === 'constellation'}
  <div class="room surface-paper">
    <div class="pw-bloom"></div>
    <div class="pw-grain"></div>

    <div class="eyebrow room-eyebrow">— First run</div>

    <div class="room-frame">
      <div class="room-head">
        <div class="room-headline">
          Configure,<br /><span class="alive">don't comply.</span>
        </div>
        <p class="room-sub">
          A quiet, attentive presence. It reads only what it must, acts only
          after showing you what it's about to do. Wire what you like — skip
          the rest.
        </p>
      </div>

      <div class="room-preview">
        <HoverPreview active={hovered} />
      </div>

      <div class="ring" class:settled={ringSettled}>
        {#each NODES as n, i (n.id)}
          <div class="ring-slot" style="--delay:{nodeDelay(i)}ms">
            <ConstellationNode
              id={n.id}
              label={n.label}
              glyph={n.glyph}
              state={stateFor(n.id)}
              indicator={indicatorFor(n.id)}
              selected={selected === n.id}
              delayMs={nodeDelay(i)}
              onclick={openNode}
              onhover={(id: NodeId | null) => (hovered = id)}
            />
          </div>
        {/each}
        <span class="ring-label">CONDURA</span>
      </div>
    </div>

    <!-- ── Per-node side panels ────────────────────────────────────── -->
    <SidePanel
      open={selected === 'perceive'}
      eyebrow={NODES[0].eyebrow}
      headline={NODES[0].headline}
      onclose={closePanel}
    >
      {#if permError}
        <div class="rit-err" role="alert">
          <Pulse phase="error" size={8} />
          <div>
            <span class="rit-err-head">We couldn't read the permission status.</span>
            <span class="rit-err-sub">You can grant them in System Settings and continue — we'll re-check.</span>
          </div>
        </div>
      {/if}
      <div class="perms">
        {#each REQUIRED_PERMS as kind (kind)}
          {@const row = permRows.find((r) => r.kind === kind)}
          <div class="perm" class:granted={row?.status === 'granted'}>
            <Glyph name={kind === 'accessibility' ? 'bolt' : 'audit'} size={18} class="perm-icon" />
            <div class="perm-body">
              <div class="perm-name">{kind === 'accessibility' ? 'Accessibility' : 'Screen Recording'}</div>
              <div class="perm-desc">{kind === 'accessibility' ? 'Read element names and invoke actions' : 'See the screen when it needs to act'}</div>
            </div>
            <span class="perm-status" data-status={row?.status ?? 'unknown'}>{row?.status ?? 'unknown'}</span>
            <button class="perm-open" onclick={() => openPermSettings(kind)}>Open →</button>
          </div>
        {/each}
      </div>
      <Button onclick={() => { void onboarding.completePermissions().catch((e) => console.error(e)); markNodeDone('perceive'); closePanel(); }}>
        Continue →
      </Button>
    </SidePanel>

    <SidePanel
      open={selected === 'power'}
      eyebrow={NODES[1].eyebrow}
      headline={NODES[1].headline}
      onclose={closePanel}
    >
      <div class="choices">
        <button class="choice" class:selected={powerChoice === 'local'} onclick={() => choosePower('local')}>
          <span class="c-radio"></span>
          <span class="c-body">
            <span class="c-title">Local — Ollama</span>
            <span class="c-meta">
              {powerProbe?.ollama_reachable
                ? `${powerProbe.ollama_models.length} models detected · stays on your machine`
                : 'not detected — install Ollama, or pick another'}
            </span>
          </span>
          {#if powerProbe?.ollama_reachable}<span class="c-tag">Recommended</span>{/if}
        </button>
        <button class="choice" class:selected={powerChoice === 'apikey'} onclick={() => choosePower('apikey')}>
          <span class="c-radio"></span>
          <span class="c-body">
            <span class="c-title">Paste an API key</span>
            <span class="c-meta">Anthropic · OpenAI · Google · xAI · more</span>
          </span>
        </button>
        <button class="choice" class:selected={powerChoice === 'sub'} onclick={() => choosePower('sub')}>
          <span class="c-radio"></span>
          <span class="c-body">
            <span class="c-title">Connect a subscription</span>
            <span class="c-meta">Claude Pro · ChatGPT Plus · Gemini · SuperGrok</span>
          </span>
        </button>
      </div>
      <Button onclick={closePanel}>Done →</Button>
    </SidePanel>

    <SidePanel
      open={selected === 'summon'}
      eyebrow={NODES[2].eyebrow}
      headline={NODES[2].headline}
      onclose={closePanel}
    >
      <div
        class="keycaps"
        class:recording
        class:tried
        onclick={startRecording}
        onkeydown={(e: KeyboardEvent) => {
          if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); startRecording(); }
        }}
        tabindex="0"
        role="button"
        aria-label="Record hotkey"
      >
        {#if recording}
          <span class="kc-hint">press a key combo…</span>
        {:else if combo}
          {#each combo.split('+') as k (k)}<span class="keycap">{k}</span>{/each}
        {:else}
          <span class="kc-hint">click and press a key combo</span>
        {/if}
      </div>
      <div class="presets">
        {#each PRESETS as p (p)}
          <button class="preset" class:active={combo === p} onclick={() => usePreset(p)}>{p}</button>
        {/each}
      </div>
      {#if combo}
        <button class="try-it" class:tried onclick={tryCombo}>
          <Pulse phase="awaiting" size={8} /> Try it
        </button>
      {/if}
      <Button onclick={saveHotkey} disabled={!combo}>Continue →</Button>
    </SidePanel>

    <SidePanel
      open={selected === 'voice'}
      eyebrow={NODES[3].eyebrow}
      headline={NODES[3].headline}
      onclose={closePanel}
    >
      {#if voiceError}
        <div class="rit-err" role="alert">
          <Pulse phase="error" size={8} />
          <div>
            <span class="rit-err-head">We couldn't probe the microphone.</span>
            <span class="rit-err-sub">{voiceError} You can still enable voice below.</span>
          </div>
        </div>
      {/if}
      <div class="voice-card">
        <Pulse phase={voiceEnabled ? 'listening' : 'idle'} size={10} />
        <div class="voice-body">
          <div class="voice-name">Microphone</div>
          <div class="voice-meta">
            {voiceProbe?.mic_available ? 'available' : 'not detected'} · wake word "hey condura"
          </div>
        </div>
      </div>
      <Button onclick={enableVoice}>{voiceEnabled ? 'Voice enabled ✓' : 'Enable voice →'}</Button>
    </SidePanel>

    <SidePanel
      open={selected === 'threads'}
      eyebrow={NODES[4].eyebrow}
      headline={NODES[4].headline}
      onclose={closePanel}
    >
      <div class="channels">
        {#each CHANNELS as c (c.id)}
          <button
            class="channel"
            class:selected={channelPick.has(c.id)}
            class:dim={!c.ready}
            onclick={() => toggleChannel(c.id, c.ready)}
          >
            <Glyph name="channels" size={22} class="ch-icon" />
            <span class="ch-name">{c.name}</span>
            <span class="ch-state">
              {c.ready ? (channelPick.has(c.id) ? 'Selected' : 'Connect') : 'v0.2.0'}
            </span>
          </button>
        {/each}
      </div>
      <Button onclick={closePanel}>Done →</Button>
    </SidePanel>

    <SidePanel
      open={selected === 'account'}
      eyebrow={NODES[5].eyebrow}
      headline={NODES[5].headline}
      onclose={closePanel}
    >
      {#if accountSent}
        <div class="account-sent">
          <Pulse phase="ok" size={10} />
          <span>Magic link on its way. You can finish signing in later.</span>
        </div>
        <Button onclick={closePanel}>Done →</Button>
      {:else}
        {#if accountError}
          <div class="rit-err" role="alert">
            <Pulse phase="error" size={8} />
            <div>
              <span class="rit-err-head">The magic link didn't go through.</span>
              <span class="rit-err-sub">{accountError} Try again, or skip.</span>
            </div>
          </div>
        {/if}
        <div class="account-row">
          <input class="email" type="email" bind:value={email} placeholder="you@example.com" disabled={accountBusy} />
          <Button onclick={sendMagicLink} disabled={!email.trim() || accountBusy}>Send magic link →</Button>
        </div>
        <Button variant="ghost" onclick={skipAccount}>skip — I'll do this later</Button>
      {/if}
    </SidePanel>

    <!-- ── Enter Condura pill — always enabled; soft-lock on hotkey ── -->
    <div class="enter">
      {#if hotkeySet}
        <Button variant="primary" magnetic={true} onclick={enterCondura}>
          Enter Condura <span class="arr">→</span>
        </Button>
      {:else}
        <button
          class="enter-pill enter-pill-locked"
          type="button"
          aria-disabled="true"
          aria-describedby="enter-help"
          onclick={() => openNode('summon')}
        >
          <Pulse phase="awaiting" size={8} />
          <span class="enter-label">Set a hotkey to enter</span>
          <span class="arr">→</span>
        </button>
        <span id="enter-help" class="sr-only">
          The hotkey is required. Tap any node to configure.
        </span>
      {/if}
    </div>
  </div>
{/if}

<style>
  /* ── shared surface paper ─────────────────────────────────────────── */
  .surface-paper {
    position: fixed;
    inset: 0;
    z-index: var(--z-modal);
    overflow: hidden;
  }
  .surface-paper,
  .room {
    transition:
      opacity var(--dur-cine) var(--ease),
      filter var(--dur-cine) var(--ease);
  }
  .surface-paper.dissolving,
  .room.dissolving {
    opacity: 0;
    filter: blur(8px);
    pointer-events: none;
  }
  .pw-bloom {
    position: absolute;
    inset: 0;
    z-index: 0;
    background:
      radial-gradient(ellipse at 20% 0%, var(--bloom-1), transparent 50%),
      radial-gradient(ellipse at 92% 8%, var(--bloom-2), transparent 45%),
      radial-gradient(ellipse at 50% 105%, var(--bloom-3), transparent 55%);
    pointer-events: none;
  }
  .pw-grain {
    position: absolute;
    inset: 0;
    z-index: 1;
    pointer-events: none;
    opacity: var(--grain-opacity);
    background-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240'><filter id='n'><feTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='2' stitchTiles='stitch'/><feColorMatrix values='0 0 0 0 0.08  0 0 0 0 0.07  0 0 0 0 0.04  0 0 0 0.06 0'/></filter><rect width='100%25' height='100%25' filter='url(%23n)'/></svg>");
    background-size: 240px 240px;
    mix-blend-mode: multiply;
  }
  :global([data-mode='dark']) .pw-grain { mix-blend-mode: screen; }

  /* ── Gate ─────────────────────────────────────────────────────────── */
  .gate {
    display: grid;
    place-items: center;
  }
  .gate-content {
    position: relative;
    z-index: 3;
    width: 100%;
    max-width: 560px;
    padding: 0 var(--space-5);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    margin-top: -8vh;
  }
  .skip-note {
    position: absolute;
    left: var(--space-5);
    bottom: var(--space-5);
    z-index: 5;
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    color: var(--content-faint);
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    border-radius: var(--r-sm);
    background: none;
    border: none;
    cursor: pointer;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .skip-note:hover {
    color: var(--pollen);
    background: color-mix(in oklab, var(--pollen) 8%, transparent);
    transform: translateX(2px);
  }
  .skip-note:active { transform: scale(0.97); }
  .skip-note:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .skip-note .arr { color: var(--pollen); }

  .eyebrow {
    font-family: var(--font-mono);
    font-size: var(--text-caption);
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .headline {
    font-family: var(--font-display);
    font-weight: 400;
    font-size: var(--text-h1);
    line-height: var(--lh-h1);
    letter-spacing: var(--ls-h1);
    color: var(--content);
    margin: 0;
  }
  .sub {
    font-size: var(--text-lead);
    line-height: var(--lh-lead);
    color: var(--content-soft);
    max-width: 52ch;
    margin: 0;
  }

  .eula {
    position: relative;
    max-height: 280px;
    overflow-y: auto;
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
    padding: var(--space-4) var(--space-5);
  }
  .eula-text {
    margin: 0;
    font-family: var(--font-display);
    font-size: 14px;
    line-height: 1.7;
    color: var(--content-soft);
    white-space: pre-wrap;
  }
  .eula-read {
    position: absolute;
    left: 0;
    top: 0;
    width: 2px;
    background: var(--synapse);
    transition: height 100ms linear;
  }
  .check {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-size: 14px;
    color: var(--content);
    cursor: pointer;
  }
  .check input {
    width: 16px;
    height: 16px;
    accent-color: var(--pollen);
    cursor: pointer;
  }
  .check input[disabled] { opacity: 0.4; }
  .eula-offline-note {
    margin: 2px 0 0 var(--space-5);
    font-family: var(--font-display);
    font-style: italic;
    font-size: 13px;
    line-height: 1.5;
    color: var(--content-faint);
  }

  .seal-row {
    display: flex;
    align-items: center;
    gap: var(--space-4);
  }
  .seal {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: radial-gradient(circle at 35% 30%, var(--synapse-glow), var(--synapse-deep) 70%);
    color: var(--paper);
    display: grid;
    place-items: center;
    cursor: pointer;
    flex: none;
    box-shadow:
      0 8px 20px -8px color-mix(in oklab, var(--synapse) 50%, transparent),
      inset 0 0 0 2px color-mix(in oklab, var(--paper) 20%, transparent);
    transition:
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease),
      opacity var(--dur) var(--ease);
  }
  .seal:hover:not([disabled]) {
    box-shadow:
      0 8px 20px -8px color-mix(in oklab, var(--synapse) 50%, transparent),
      inset 0 0 0 2px color-mix(in oklab, var(--paper) 20%, transparent),
      var(--pollen-halo);
  }
  .seal[disabled] { opacity: 0.35; cursor: not-allowed; }
  .seal.stamped {
    transform: scale(0.94) translateY(2px);
    animation: sealBloom 600ms var(--ease);
  }
  .seal-c {
    font-family: var(--font-display);
    font-size: 28px;
    line-height: 1;
    color: var(--paper);
  }
  :global([data-mode='dark']) .seal-c { color: var(--ink); }
  .seal-text { display: flex; flex-direction: column; gap: 2px; }
  .st-1 { font-size: 15px; color: var(--content); }
  .st-2 {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-faint);
  }

  @keyframes sealBloom {
    0%   { box-shadow: 0 8px 20px -8px rgba(11,61,46,.5), inset 0 0 0 2px rgba(234,246,239,.2), 0 0 0 0 rgba(201,123,46,.5); }
    100% { box-shadow: 0 8px 20px -8px rgba(11,61,46,.5), inset 0 0 0 2px rgba(234,246,239,.2), 0 0 0 28px rgba(201,123,46,0); }
  }

  /* ── Constellation ────────────────────────────────────────────────── */
  .room { display: block; }
  .room-eyebrow {
    position: absolute;
    left: 24px;
    top: 24px;
    z-index: 3;
  }
  .room-frame {
    position: absolute;
    inset: 0;
    display: grid;
    grid-template-rows: auto auto 1fr;
    align-items: center;
    justify-items: center;
    gap: var(--space-3);
    padding: 64px var(--space-5) 96px;
    z-index: 2;
  }
  .room-head {
    text-align: left;
    align-self: end;
    width: 100%;
    max-width: 560px;
    padding: 0 var(--space-3);
  }
  .room-headline {
    font-family: var(--font-display);
    font-weight: 400;
    font-size: clamp(34px, 4.4vw, 48px);
    line-height: 1.04;
    letter-spacing: -0.035em;
    color: var(--content);
    margin: 0 0 var(--space-3);
  }
  .room-headline .alive {
    font-style: italic;
    color: var(--synapse);
  }
  .room-sub {
    font-size: var(--text-lead);
    line-height: var(--lh-lead);
    color: var(--content-soft);
    max-width: 52ch;
    margin: 0;
  }
  .room-preview {
    align-self: center;
    width: 100%;
    max-width: 560px;
    padding: 0 var(--space-3);
  }

  /* ── Ring layout — 6 slots on a dashed circle ───────────────────── */
  .ring {
    position: relative;
    width: 340px;
    height: 340px;
    align-self: start;
    margin-top: var(--space-3);
  }
  .ring::before {
    content: '';
    position: absolute;
    inset: 16px;
    border-radius: 50%;
    border: 1px dashed var(--hair-strong);
    pointer-events: none;
  }
  .ring-slot {
    position: absolute;
    width: 84px;
    /* Counter-clockwise from top: top, top-left, bottom-left, top-right,
       bottom-right, bottom. Anchor positions are ring-radius-based. */
    top: var(--rt, 50%);
    left: var(--rl, 50%);
    transform: translate(-50%, -50%) scale(0.7);
    opacity: 0;
    transition:
      transform var(--dur) var(--ease),
      opacity var(--dur) var(--ease);
    transition-delay: var(--delay);
  }
  .ring.settled .ring-slot {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
  .ring-slot:nth-child(1) { --rl: 50%; --rt: calc(50% - 144px); }     /* Perceive top */
  .ring-slot:nth-child(2) { --rl: calc(50% - 125px); --rt: calc(50% - 75px); } /* Power top-left */
  .ring-slot:nth-child(3) { --rl: calc(50% + 125px); --rt: calc(50% - 75px); } /* Summon top-right */
  .ring-slot:nth-child(4) { --rl: calc(50% - 125px); --rt: calc(50% + 75px); } /* Voice bottom-left */
  .ring-slot:nth-child(5) { --rl: calc(50% + 125px); --rt: calc(50% + 75px); } /* Threads bottom-right */
  .ring-slot:nth-child(6) { --rl: 50%; --rt: calc(50% + 144px); }    /* Account bottom */
  .ring-label {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, calc(-50% + 14px));
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-faint);
  }

  /* ── Enter Condura pill ─────────────────────────────────────────── */
  .enter {
    position: absolute;
    bottom: 28px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 4;
  }
  .enter-pill {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    background: var(--pollen);
    color: var(--paper);
    border: 1px solid color-mix(in oklab, var(--pollen-deep) 50%, transparent);
    border-radius: var(--r-pill);
    padding: 11px 20px;
    font-family: var(--font-sans);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    box-shadow: var(--shadow-card);
    transition:
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  :global([data-mode='dark']) .enter-pill { color: var(--ink); }
  .enter-pill:hover {
    box-shadow:
      0 1px 0 color-mix(in oklab, var(--paper) 12%, transparent) inset,
      0 18px 40px -16px color-mix(in oklab, var(--ink) 60%, transparent),
      var(--pollen-halo);
    transform: translateY(-1px);
  }
  .enter-pill:active:not(.enter-pill-locked) { transform: scale(0.97); }
  .enter-pill:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 4px var(--pollen-halo),
      var(--shadow-card);
  }
  .enter-pill .arr {
    margin-left: 2px;
    transition: transform var(--dur) var(--ease);
  }
  .enter-pill:hover .arr { transform: translate(2px, -2px); }
  .enter-pill-locked {
    background: color-mix(in oklab, var(--synapse) 14%, var(--paper));
    color: var(--content);
    border-color: color-mix(in oklab, var(--synapse) 28%, transparent);
    animation: pill-breathe 1.6s var(--ease) infinite;
  }
  .enter-pill-locked:hover {
    box-shadow:
      0 0 0 4px color-mix(in oklab, var(--synapse) 22%, transparent),
      var(--shadow-card);
  }
  @keyframes pill-breathe {
    0%, 100% { box-shadow: 0 0 0 0 color-mix(in oklab, var(--synapse) 18%, transparent), var(--shadow-card); }
    50%      { box-shadow: 0 0 0 6px color-mix(in oklab, var(--synapse) 18%, transparent), var(--shadow-card); }
  }

  /* ── Side-panel shared content bits ────────────────────────────── */
  .rit-err {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-3) var(--space-4);
    border: 1px solid color-mix(in srgb, var(--danger) 28%, transparent);
    border-radius: var(--r-md);
    background: color-mix(in srgb, var(--danger) 5%, transparent);
  }
  .rit-err > div { display: flex; flex-direction: column; gap: 2px; }
  .rit-err-head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 14px;
    color: var(--content);
  }
  .rit-err-sub {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 12px;
    color: var(--content-faint);
  }

  /* permissions list */
  .perms { display: flex; flex-direction: column; gap: var(--space-3); }
  .perm {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
    transition: border-color var(--dur) var(--ease);
  }
  .perm.granted { border-color: color-mix(in oklab, var(--synapse) 30%, transparent); }
  :global(.perm-icon) { color: var(--content-mute); flex: none; }
  :global(.perm.granted .perm-icon) { color: var(--synapse); }
  .perm-body { display: flex; flex-direction: column; gap: 2px; flex: 1; min-width: 0; }
  .perm-name { font-size: 14px; color: var(--content); }
  .perm-desc { font-size: 12px; color: var(--content-mute); }
  .perm-status {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    padding: 3px 8px;
    border-radius: var(--r-pill);
    border: 1px solid var(--hair-strong);
    color: var(--content-faint);
  }
  .perm-status[data-status='granted'] {
    color: var(--ok);
    border-color: color-mix(in oklab, var(--ok) 40%, transparent);
    background: color-mix(in oklab, var(--ok) 6%, transparent);
  }
  .perm-status[data-status='denied'] {
    color: var(--danger);
    border-color: color-mix(in oklab, var(--danger) 40%, transparent);
    background: color-mix(in oklab, var(--danger) 6%, transparent);
  }
  .perm-open {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--synapse);
    background: none;
    padding: 4px 8px;
    border-radius: var(--r-sm);
  }
  .perm-open:hover { background: color-mix(in oklab, var(--synapse) 8%, transparent); }

  /* power choices */
  .choices { display: flex; flex-direction: column; gap: var(--space-3); }
  .choice {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-4) var(--space-5);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
    text-align: left;
    cursor: pointer;
    transition:
      border-color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      background var(--dur) var(--ease);
  }
  .choice:hover { transform: translateY(-1px); border-color: var(--hair-strong); background: var(--paper-2); }
  .choice:focus-visible { outline: none; border-color: var(--synapse); box-shadow: 0 0 0 4px var(--pollen-halo); }
  .choice.selected { border-color: var(--pollen); box-shadow: var(--pollen-halo); }
  .c-radio {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    border: 2px solid var(--hair-strong);
    flex: none;
    display: grid;
    place-items: center;
    transition: border-color var(--dur) var(--ease);
  }
  .choice.selected .c-radio { border-color: var(--pollen); }
  .choice.selected .c-radio::after {
    content: '';
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--pollen);
  }
  .c-body { display: flex; flex-direction: column; gap: 2px; flex: 1; min-width: 0; }
  .c-title { font-size: 15px; color: var(--content); }
  .c-meta {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .c-tag {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--synapse);
    border: 1px solid color-mix(in oklab, var(--synapse) 30%, transparent);
    padding: 3px 8px;
    border-radius: var(--r-pill);
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
  }

  /* hotkey capture */
  .keycaps {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-4);
    border: 1px dashed var(--hair-strong);
    border-radius: var(--r-md);
    min-height: 64px;
    cursor: pointer;
    transition: border-color var(--dur) var(--ease), box-shadow var(--dur) var(--ease);
  }
  .keycaps.recording {
    border-color: var(--pollen);
    border-style: solid;
    box-shadow: var(--pollen-halo);
  }
  .keycaps.tried {
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--pollen) 25%, transparent);
  }
  .keycap {
    font-family: var(--font-mono);
    font-size: 15px;
    font-weight: 500;
    padding: 8px 14px;
    border: 1px solid var(--hair-strong);
    border-bottom-width: 3px;
    border-radius: var(--r-sm);
    background: var(--surface-card);
    color: var(--content);
    box-shadow: var(--shadow-card);
  }
  .kc-hint {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--content-faint);
  }
  .presets { display: flex; gap: var(--space-2); flex-wrap: wrap; }
  .preset {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 6px 12px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-sm);
    background: var(--surface-card);
    color: var(--content-soft);
    cursor: pointer;
    transition:
      border-color var(--dur) var(--ease),
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease);
  }
  .preset:hover {
    border-color: var(--pollen);
    color: var(--content);
    background: color-mix(in oklab, var(--pollen) 8%, transparent);
    transform: translateY(-1px);
  }
  .preset.active { border-color: var(--pollen); color: var(--content); box-shadow: var(--pollen-halo); }
  .try-it {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    align-self: flex-start;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--synapse);
    background: none;
    cursor: pointer;
    padding: 4px 10px;
    border-radius: var(--r-pill);
    transition:
      transform var(--dur) var(--ease),
      background var(--dur) var(--ease);
  }
  .try-it:hover { background: color-mix(in oklab, var(--synapse) 8%, transparent); }
  .try-it.tried { transform: scale(1.05); }

  /* voice */
  .voice-card {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-4) var(--space-5);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
  }
  .voice-body { display: flex; flex-direction: column; gap: 2px; }
  .voice-name { font-size: 15px; color: var(--content); }
  .voice-meta {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
  }

  /* channels */
  .channels { display: flex; gap: var(--space-3); flex-wrap: wrap; }
  .channel {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-4);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
    cursor: pointer;
    min-width: 96px;
    transition:
      border-color var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      background var(--dur) var(--ease);
  }
  .channel:hover:not(.dim) {
    transform: translateY(-2px);
    border-color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 4%, transparent);
  }
  .channel.dim { opacity: 0.55; cursor: not-allowed; }
  .channel.selected { border-color: color-mix(in oklab, var(--synapse) 30%, transparent); }
  :global(.ch-icon) { color: var(--content-mute); }
  :global(.channel.selected .ch-icon) { color: var(--synapse); }
  .ch-name { font-size: 12px; color: var(--content); }
  .ch-state {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .channel.selected .ch-state { color: var(--synapse); }

  /* account */
  .account-row { display: flex; align-items: center; gap: var(--space-3); }
  .email {
    flex: 1;
    padding: 11px 16px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    background: var(--surface);
    color: var(--content);
    font-size: 14px;
    outline: none;
    transition:
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .email::placeholder { color: var(--content-faint); }
  .email:focus {
    border-color: var(--synapse);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .email:disabled { opacity: 0.42; cursor: not-allowed; }
  .account-sent {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-4) var(--space-5);
    border: 1px solid color-mix(in oklab, var(--ok) 30%, transparent);
    border-radius: var(--r-md);
    background: color-mix(in oklab, var(--ok) 6%, transparent);
    color: var(--content-soft);
  }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    margin: -1px;
    padding: 0;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  @media (prefers-reduced-motion: reduce) {
    .seal.stamped { animation: none; transform: scale(0.94) translateY(2px); }
    .ring-slot {
      transition: none;
      opacity: 1;
      transform: translate(-50%, -50%) scale(1);
    }
    .ring::before { display: none; }
    .enter-pill,
    .enter-pill-locked,
    .pill-breathe { animation: none; }
  }
</style>
