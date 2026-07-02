<script lang="ts">
  // Condura Settings — a flowing document, not a form.
  // One long column of italic Instrument Serif section titles separated by
  // hairline rules. No tabs, no settings-sidebar. The autonomy matrix is the
  // hero: a grid of task-type rows × three states (block / warn / autonomous),
  // each state a dot you cycle on click, with a live preview line.
  //
  // Two kinds of state live here:
  //   1. Daemon config (autonomy matrix + per-provider default model) — real
  //      AppConfig paths. We keep a local working copy; a pollen save bar
  //      springs in from the bottom when it is dirty and springs out when
  //      saved or reverted. `settings.save(patch)` does the write.
  //   2. Local UI preferences (motion, grain, energy, voice) — there
  //      is no AppConfig path for any of these. Rather than guess a path, we
  //      apply them at once to :root / CSS variables and persist them to
  //      localStorage. They never set the dirty flag and never hit the bar.
  //      Theme is delegated to <ThemePicker /> (Phase 2), which owns its
  //      own palette-switch choreography.
  //
  // Sections stagger in 40ms on mount; prefers-reduced-motion skips the
  // stagger entirely (one CSS animation owns it; reset on dirty=false).

  import { onMount } from 'svelte';
  import { fly } from 'svelte/transition';
  import { backOut } from 'svelte/easing';
  import { settings } from '../stores/settings.svelte';
  import { account } from '../stores/account.svelte';
  import { ipc } from '../ipc/client';
  import type { AppConfig, AdaptiveStrength, PermissionStatus, VoiceProbeResult } from '../ipc/types';
  import Thread from './Thread.svelte';
  import Button from './Button.svelte';
  import Pulse from './Pulse.svelte';
  import ThemePicker from './ThemePicker.svelte';
  import AutonomyMatrix from './AutonomyMatrix.svelte';

  // ── adaptive engine (best-effort; no adaptive store exists yet) ──
  const STRENGTHS: AdaptiveStrength[] = ['off', 'cautious', 'balanced', 'aggressive'];
  const STRENGTH_COLOR: Record<AdaptiveStrength, string> = {
    off: 'var(--content-faint)',
    cautious: 'var(--info)',
    balanced: 'var(--pollen)',
    aggressive: 'var(--synapse-glow)',
  };

  type ProviderCfg = AppConfig['llm']['providers'][string];

  // ── working copy (daemon config; dirty-tracked) ──
  let autonomyPerTask = $state<Record<string, string>>({});
  let autonomyDefault = $state<string>('warn');
  let providerModels = $state<Record<string, ProviderCfg>>({});
  let dirty = $state(false);
  // Tracks the last successful save so the bar can fade out over 400ms
  // after the daemon acks (rather than popping out instantly). Cleared
  // on dirty=true or after the fade completes.
  let savedFlash = $state(false);

  // ── local UI prefs (applied at once; never dirty) ──
  // Theme is owned by <ThemePicker />. Motion / grain / energy / voice are
  // applied at once to CSS variables and persisted to localStorage; they
  // never trigger the save bar.
  let energyChoice = $state<'low' | 'balanced' | 'high' | 'auto'>('auto');
  let motionValue = $state(100);
  let grainValue = $state(100);
  let wakeEnabled = $state(false);
  let wakeSensitivity = $state(60);
  const reducedMotion =
    typeof window !== 'undefined' && matchMedia('(prefers-reduced-motion: reduce)').matches;

  // ── read-only / best-effort surface state ──
  let adaptiveStrength = $state<AdaptiveStrength | ''>('');
  let permissions = $state<PermissionStatus[]>([]);
  let voiceProbe = $state<VoiceProbeResult | null>(null);
  let eulaVersion = $state('');
  let eulaUpdated = $state('');
  let eulaOpen = $state(false);
  let eulaText = $state('');
  let eulaLoading = $state(false);

  // ── mount-driven section stagger ──
  // Each section reads `--stagger-index` set inline below and runs a single
  // CSS animation (opacity + translateY) with delay = idx × 40ms. The
  // animation is owned by the .stagger-in class which is added on mount;
  // prefers-reduced-motion disables it via a single CSS rule below.
  let mounted = $state(false);

  // ── derived views ──
  let providerNames = $derived(Object.keys(settings.config?.llm?.providers ?? {}));

  // save-bar spring: pollen-accented, back-out for a settling overshoot.
  // Honors prefers-reduced-motion by zeroing the duration.
  const saveSpring = { y: 64, duration: reducedMotion ? 0 : 380, easing: backOut };

  // ── working-copy mutation handlers ──

  function setAutonomy(next: Record<string, string>): void {
    autonomyPerTask = next;
    dirty = true;
  }

  function setModel(name: string, value: string): void {
    const p = providerModels[name];
    if (!p) return;
    providerModels = { ...providerModels, [name]: { ...p, default_model: value } };
    dirty = true;
  }

  function syncFromConfig(): void {
    const cfg = settings.config;
    if (!cfg) return;
    autonomyPerTask = { ...(cfg.autonomy?.per_task ?? {}) };
    autonomyDefault = cfg.autonomy?.default_level ?? 'warn';
    const next: Record<string, ProviderCfg> = {};
    for (const [name, p] of Object.entries(cfg.llm?.providers ?? {})) {
      next[name] = { ...p };
    }
    providerModels = next;
  }

  // Re-sync from the daemon whenever config loads/changes — but never clobber
  // the user's in-flight edits (skip while dirty).
  $effect(() => {
    const cfg = settings.config;
    if (!cfg || dirty) return;
    syncFromConfig();
  });

  async function save(): Promise<void> {
    const cfg = settings.config;
    if (!cfg || settings.saving) return;
    const patch: Partial<AppConfig> = {};
    patch.autonomy = {
      ...cfg.autonomy,
      per_task: { ...autonomyPerTask },
      default_level: autonomyDefault,
    };
    const newProviders: AppConfig['llm']['providers'] = {};
    for (const [name, p] of Object.entries(cfg.llm?.providers ?? {})) {
      newProviders[name] = {
        ...p,
        default_model: providerModels[name]?.default_model ?? p.default_model,
      };
    }
    if (Object.keys(newProviders).length) {
      patch.llm = { ...cfg.llm, providers: newProviders };
    }
    try {
      await settings.save(patch);
      dirty = false;
      savedFlash = true;
      // Fade the bar out 400ms after a successful save (matches the
      // duration of the fade transition in CSS).
      if (typeof window !== 'undefined') {
        window.setTimeout(() => {
          savedFlash = false;
        }, reducedMotion ? 0 : 400);
      } else {
        savedFlash = false;
      }
    } catch {
      // settings.lastSaveError carries the message; the bar stays (still dirty).
      savedFlash = false;
    }
  }

  function revert(): void {
    dirty = false;
    syncFromConfig();
  }

  // ── empty-state / loading helper ──
  // `loaded` is true once the daemon's config has been read. Until then,
  // every section that reflects config (matrix, providers, eula) shows
  // a subtle "—" in ink-faint, with a 1-px pollen hairline that draws
  // across to signal "fetching".
  let configEmpty = $derived(!settings.config && !settings.loaded);

  // ── local UI prefs: applied at once, persisted to localStorage ──

  function applyEnergy(e: 'low' | 'balanced' | 'high' | 'auto'): void {
    energyChoice = e;
    try {
      if (e === 'low') {
        document.documentElement.dataset.energy = 'low';
        localStorage.setItem('condura-energy', 'low');
        // Low energy means no motion too — keep the two consistent.
        if (motionValue !== 0) applyMotion(0);
      } else {
        document.documentElement.removeAttribute('data-energy');
        if (e === 'auto') localStorage.removeItem('condura-energy');
        else localStorage.setItem('condura-energy', e);
      }
    } catch {
      /* storage unavailable */
    }
  }

  function applyMotion(v: number): void {
    motionValue = v;
    const frac = v / 100;
    try {
      document.documentElement.style.setProperty('--dur-cine', `${Math.round(900 * frac)}ms`);
      document.documentElement.style.setProperty('--dur-slow', `${Math.round(520 * frac)}ms`);
      localStorage.setItem('condura-motion', String(v));
      // Raising motion above zero while energy is low would fight the
      // stylesheet's data-energy='low' rule, so bump energy up to balanced.
      if (v > 0 && energyChoice === 'low') {
        energyChoice = 'balanced';
        document.documentElement.removeAttribute('data-energy');
        localStorage.setItem('condura-energy', 'balanced');
      }
    } catch {
      /* storage unavailable */
    }
  }

  function applyGrain(v: number): void {
    grainValue = v;
    try {
      document.documentElement.style.setProperty('--grain-opacity', String((v / 100) * 0.6));
      localStorage.setItem('condura-grain', String(v));
    } catch {
      /* storage unavailable */
    }
  }

  function setWake(on: boolean): void {
    wakeEnabled = on;
    try {
      localStorage.setItem('condura-wake-enabled', on ? '1' : '0');
    } catch {
      /* storage unavailable */
    }
  }

  function setWakeSens(v: number): void {
    wakeSensitivity = v;
    try {
      localStorage.setItem('condura-wake-sensitivity', String(v));
    } catch {
      /* storage unavailable */
    }
  }

  function setAdaptiveStrength(s: AdaptiveStrength): void {
    if (adaptiveStrength === s) return;
    adaptiveStrength = s;
    try {
      void ipc.adaptiveStrengthSet(s).catch(() => {
        /* adaptive engine offline */
      });
    } catch {
      /* ipc unavailable */
    }
  }

  async function openEula(): Promise<void> {
    if (eulaOpen) {
      eulaOpen = false;
      return;
    }
    eulaOpen = true;
    if (eulaText) return;
    eulaLoading = true;
    try {
      const doc = await ipc.onboardingEula();
      eulaVersion = doc.version;
      eulaUpdated = doc.updated_at;
      eulaText = doc.text;
    } catch {
      eulaText = 'Could not load the EULA. The daemon may be offline.';
    } finally {
      eulaLoading = false;
    }
  }

  // ── keyboard chords owned by Settings ──
  // ⌘S = save, ⌘Z = revert (undo last in-flight edit), Esc = discard.
  // Tab order is the DOM order; the AutonomyMatrix owns its own arrow-key
  // roving tabindex internally.
  function onKeydown(ev: KeyboardEvent): void {
    const mod = ev.metaKey || ev.ctrlKey;
    if (!mod && ev.key !== 'Escape') return;
    if (mod && ev.key.toLowerCase() === 's') {
      ev.preventDefault();
      void save();
      return;
    }
    if (mod && ev.key.toLowerCase() === 'z') {
      ev.preventDefault();
      revert();
      return;
    }
    if (ev.key === 'Escape' && dirty) {
      ev.preventDefault();
      revert();
      return;
    }
  }

  onMount(() => {
    // restore local prefs
    try {
      const se = localStorage.getItem('condura-energy');
      energyChoice = se === 'low' || se === 'balanced' || se === 'high' ? se : 'auto';
      if (energyChoice === 'low') document.documentElement.dataset.energy = 'low';
    } catch {
      /* storage unavailable */
    }
    try {
      const sm = Number(localStorage.getItem('condura-motion') ?? '100');
      motionValue = Number.isNaN(sm) ? 100 : Math.max(0, Math.min(100, sm));
    } catch {
      /* storage unavailable */
    }
    applyMotion(motionValue);
    try {
      const sg = Number(localStorage.getItem('condura-grain') ?? '100');
      grainValue = Number.isNaN(sg) ? 100 : Math.max(0, Math.min(100, sg));
    } catch {
      /* storage unavailable */
    }
    applyGrain(grainValue);
    try {
      wakeEnabled = localStorage.getItem('condura-wake-enabled') === '1';
    } catch {
      /* storage unavailable */
    }
    try {
      const ws = Number(localStorage.getItem('condura-wake-sensitivity') ?? '60');
      wakeSensitivity = Number.isNaN(ws) ? 60 : Math.max(0, Math.min(100, ws));
    } catch {
      /* storage unavailable */
    }

    // daemon config — the $effect above syncs the working copy once it lands.
    try {
      void settings.refresh().catch(() => {
        /* daemon offline; working copy stays at defaults */
      });
    } catch {
      /* ipc unavailable */
    }

    // best-effort adaptive strength (no adaptive store ships yet)
    try {
      void ipc
        .adaptiveStrengthGet()
        .then((r) => {
          adaptiveStrength = r.strength;
        })
        .catch(() => {
          /* adaptive engine offline */
        });
    } catch {
      /* ipc unavailable */
    }

    // account + permissions (read-only surfaces)
    try {
      void account.checkStatus?.().catch(() => {
        /* account offline */
      });
    } catch {
      /* account store unavailable */
    }
    try {
      void ipc
        .permissionsStatus()
        .then((p) => {
          permissions = p;
        })
        .catch(() => {
          /* permissions offline */
        });
    } catch {
      /* ipc unavailable */
    }
    try {
      void ipc
        .onboardingProbeVoice()
        .then((v) => {
          voiceProbe = v;
        })
        .catch(() => {
          /* voice probe offline */
        });
    } catch {
      /* ipc unavailable */
    }

    // Wire keyboard chords for the Settings surface.
    window.addEventListener('keydown', onKeydown);

    // Trigger the section stagger after the first paint so the initial
    // styles commit before the animation begins.
    queueMicrotask(() => {
      mounted = true;
    });

    return () => {
      window.removeEventListener('keydown', onKeydown);
    };
  });
</script>

<section class="settings" aria-label="Settings document">
  <header class="doc-head">
    <div class="eyebrow">— Configuration</div>
    <h1 class="doc-title">Settings</h1>
    <div class="rule"><Thread orientation="h" /></div>
    <p class="doc-lead">
      A flowing document. Change what Condura does, how it looks, how freely it may act. Daemon-backed
      choices save when you press save; appearance and motion are yours, applied at once.
    </p>
  </header>

  <!-- ── Appearance ── -->
  <section
    class="sect"
    class:stagger-in={mounted}
    style:--stagger-index={0}
  >
    <h2 class="sect-title">Appearance</h2>

    <div class="row">
      <div class="row-label">
        <span class="row-name">Theme</span>
        <span class="row-hint">Light, dark, or follow your system.</span>
      </div>
      <ThemePicker />
    </div>

    <div class="row">
      <div class="row-label">
        <span class="row-name">Motion strength</span>
        <span class="row-hint">
          {reducedMotion ? 'Reduced by your system preference.' : 'How long animations take to settle.'}
        </span>
      </div>
      <input
        type="range"
        min="0"
        max="100"
        step="1"
        value={motionValue}
        disabled={reducedMotion}
        oninput={(e) => applyMotion(+e.currentTarget.value)}
        class="slider"
        style:--slider-fill={`${motionValue}%`}
        aria-label="Motion strength"
      />
    </div>

    <div class="row">
      <div class="row-label">
        <span class="row-name">Grain intensity</span>
        <span class="row-hint">Paper texture. Zero for flat.</span>
      </div>
      <input
        type="range"
        min="0"
        max="100"
        step="1"
        value={grainValue}
        oninput={(e) => applyGrain(+e.currentTarget.value)}
        class="slider"
        style:--slider-fill={`${grainValue}%`}
        aria-label="Grain intensity"
      />
    </div>
  </section>

  <div class="hair" class:hair-stagger-in={mounted} style:--stagger-index={1}></div>

  <!-- ── Power ── -->
  <section
    class="sect"
    class:stagger-in={mounted}
    style:--stagger-index={2}
  >
    <h2 class="sect-title">Power</h2>

    <div class="row">
      <div class="row-label">
        <span class="row-name">Energy budget</span>
        <span class="row-hint">Low silences ambient motion and grain to save battery.</span>
      </div>
      <div class="seg" role="radiogroup" aria-label="Energy budget">
        {#each ['low', 'balanced', 'high', 'auto'] as opt}
          <button
            class="seg-btn"
            class:active={energyChoice === opt}
            onclick={() => applyEnergy(opt as 'low' | 'balanced' | 'high' | 'auto')}
            role="radio"
            aria-checked={energyChoice === opt}
          >
            {opt}
          </button>
        {/each}
      </div>
    </div>

    <div class="subhead">Default model per provider</div>
    {#if providerNames.length === 0}
      <p class="empty">No providers configured. Add an API key to begin.</p>
    {:else}
      {#each providerNames as name}
        <div class="row">
          <div class="row-label">
            <span class="row-name">{name}</span>
          </div>
          <input
            class="field"
            type="text"
            value={providerModels[name]?.default_model ?? ''}
            placeholder="model id"
            oninput={(e) => setModel(name, e.currentTarget.value)}
            spellcheck="false"
            autocomplete="off"
            aria-label={`Default model for ${name}`}
          />
        </div>
      {/each}
    {/if}
  </section>

  <div class="hair" class:hair-stagger-in={mounted} style:--stagger-index={3}></div>

  <!-- ── Autonomy matrix (the hero) ── -->
  <section
    class="sect hero-sect"
    class:stagger-in={mounted}
    style:--stagger-index={4}
  >
    <h2 class="sect-title">Autonomy matrix</h2>
    <p class="sect-lead">
      For each kind of work, decide whether Condura may act on its own, must warn you first, or may not
      act at all. Click a dot to set it; click the lit dot to cycle.
    </p>

    {#if configEmpty}
      <div class="matrix-loading" aria-busy="true" aria-live="polite">
        <span class="loading-mark" aria-hidden="true">—</span>
        <span class="loading-rule"><Thread orientation="h" /></span>
      </div>
    {:else}
      <AutonomyMatrix
        perTask={autonomyPerTask}
        defaultLevel={autonomyDefault}
        onChange={setAutonomy}
      />
    {/if}
  </section>

  <div class="hair" class:hair-stagger-in={mounted} style:--stagger-index={5}></div>

  <!-- ── Adaptive engine ── -->
  <section
    class="sect"
    class:stagger-in={mounted}
    style:--stagger-index={6}
  >
    <h2 class="sect-title">Adaptive engine</h2>
    <div class="row">
      <div class="row-label">
        <span class="row-name">Learning strength</span>
        <span class="row-hint">
          How eagerly Condura learns your patterns. Off reads nothing; aggressive applies inferences
          faster.
        </span>
      </div>
      <div class="dots4" role="radiogroup" aria-label="Adaptive learning strength">
        {#each STRENGTHS as s}
          <button
            class="auto-dot"
            class:active={adaptiveStrength === s}
            style:--dot-color={STRENGTH_COLOR[s]}
            onclick={() => setAdaptiveStrength(s)}
            aria-label={s}
            aria-pressed={adaptiveStrength === s}
          >
            <span class="dot-fill"></span>
          </button>
        {/each}
      </div>
    </div>
  </section>

  <div class="hair" class:hair-stagger-in={mounted} style:--stagger-index={7}></div>

  <!-- ── Voice ── -->
  <section
    class="sect"
    class:stagger-in={mounted}
    style:--stagger-index={8}
  >
    <h2 class="sect-title">Voice</h2>
    {#if voiceProbe}
      <p class="status-line">
        Mic {voiceProbe.mic_available ? 'available' : 'not granted'} · wake word
        <em>{voiceProbe.wake_word || 'hey condura'}</em>
        {voiceProbe.wake_word_enabled ? 'on' : 'off'} per the daemon.
      </p>
    {/if}

    <div class="row">
      <div class="row-label">
        <span class="row-name">Wake word</span>
        <span class="row-hint">Listen for "hey condura" hands-free.</span>
      </div>
      <button
        class="toggle"
        class:on={wakeEnabled}
        onclick={() => setWake(!wakeEnabled)}
        role="switch"
        aria-checked={wakeEnabled}
        aria-label="Wake word"
      >
        <span class="toggle-knob"></span>
      </button>
    </div>

    <div class="row">
      <div class="row-label">
        <span class="row-name">Sensitivity</span>
        <span class="row-hint">Higher hears quieter phrases.</span>
      </div>
      <input
        type="range"
        min="0"
        max="100"
        step="1"
        value={wakeSensitivity}
        disabled={!wakeEnabled}
        oninput={(e) => setWakeSens(+e.currentTarget.value)}
        class="slider"
        style:--slider-fill={`${wakeSensitivity}%`}
        aria-label="Wake word sensitivity"
      />
    </div>
  </section>

  <div class="hair" class:hair-stagger-in={mounted} style:--stagger-index={9}></div>

  <!-- ── Account ── -->
  <section
    class="sect"
    class:stagger-in={mounted}
    style:--stagger-index={10}
  >
    <h2 class="sect-title">Account</h2>
    {#if account.isSignedIn}
      <div class="account-chip">
        {#if account.avatarURL}
          <img src={account.avatarURL} alt="" class="avatar" />
        {:else}
          <span class="avatar-fallback">
            {(account.displayName || account.email || '?').slice(0, 1).toUpperCase()}
          </span>
        {/if}
        <div class="account-meta">
          <div class="account-name">{account.displayName || 'Signed in'}</div>
          <div class="account-email">{account.email}</div>
        </div>
        <span class="account-provider">{account.provider}</span>
      </div>
    {:else}
      <p class="account-out">
        Not signed in. The agent works without an account —
        <span class="thread-link">sign in</span> to sync skills and hub bookmarks.
      </p>
    {/if}
  </section>

  <div class="hair" class:hair-stagger-in={mounted} style:--stagger-index={11}></div>

  <!-- ── Permissions (read-only) ── -->
  <section
    class="sect"
    class:stagger-in={mounted}
    style:--stagger-index={12}
  >
    <h2 class="sect-title">Permissions</h2>
    <p class="sect-lead">OS grants Condura already has. Read-only here — grant them in System Settings.</p>
    {#if permissions.length === 0}
      <p class="empty">No permissions reported. The daemon may be offline.</p>
    {:else}
      <ul class="perms">
        {#each permissions as p}
          <li class="perm">
            <span class="perm-name">{p.kind}</span>
            <span class="perm-badge" data-status={p.status}>{p.status}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  <div class="hair" class:hair-stagger-in={mounted} style:--stagger-index={13}></div>

  <!-- ── Legal ── -->
  <section
    class="sect"
    class:stagger-in={mounted}
    style:--stagger-index={14}
  >
    <h2 class="sect-title">Legal</h2>
    <button class="eula-row" onclick={openEula} aria-expanded={eulaOpen}>
      <span class="row-name">End-User License Agreement</span>
      <span class="row-meta">{eulaVersion ? `v${eulaVersion}` : 'Open to read'}</span>
    </button>
    {#if eulaOpen}
      <div class="eula-body">
        {#if eulaLoading}
          <div class="eula-loading">
            <Pulse phase="thinking" size={8} />
            <span class="eula-loading-label">READING THE LICENSE…</span>
          </div>
        {/if}
        {#if eulaText && !eulaLoading}<pre class="eula-text">{eulaText}</pre>{/if}
        {#if eulaUpdated && eulaText && !eulaLoading}<p class="small eula-updated">Last updated {eulaUpdated}</p>{/if}
      </div>
    {/if}
  </section>

  {#if dirty || savedFlash}
    <div
      class="save-bar"
      class:save-bar--saved={savedFlash && !dirty}
      class:save-bar--dirty={dirty}
      transition:fly={saveSpring}
    >
      <div class="save-bar-inner">
        <span class="save-status" aria-live="polite">
          {#if settings.saving}
            <Pulse phase="acting" size={8} />
            <span class="save-status-mono">SAVING…</span>
          {:else if settings.lastSaveError}
            <Pulse phase="error" size={8} />
            <span class="save-status-text">Save failed</span>
          {:else if savedFlash && !dirty}
            <Pulse phase="ok" size={8} />
            <span class="save-status-text">Saved</span>
          {:else}
            <Pulse phase="awaiting" size={8} />
            <span class="save-status-text">Unsaved changes</span>
          {/if}
        </span>
        {#if settings.lastSaveError}<span class="save-error">{settings.lastSaveError}</span>{/if}
        <div class="save-actions">
          <Button variant="ghost" size="sm" onclick={revert} disabled={settings.saving}>Revert</Button>
          <Button variant="primary" size="sm" onclick={save} disabled={settings.saving}>
            {settings.saving ? 'Saving' : 'Save'}
          </Button>
        </div>
      </div>
      <div class="save-bar-thread" aria-hidden="true"><Thread orientation="h" /></div>
    </div>
  {/if}
</section>

<style>
  .settings {
    position: relative;
    max-width: 65ch;
    margin: 0 auto;
    padding: var(--space-8) var(--space-6) var(--space-9);
  }

  .doc-head {
    margin-bottom: var(--space-7);
  }
  .doc-title {
    font-family: var(--font-display);
    font-weight: 400;
    font-size: clamp(34px, 4vw, 44px);
    line-height: 1;
    letter-spacing: -0.03em;
    color: var(--content);
    margin: var(--space-3) 0 var(--space-4);
  }
  .doc-head .rule {
    width: 120px;
    margin-bottom: var(--space-5);
  }
  .doc-lead {
    font-size: 16px;
    line-height: 1.55;
    color: var(--content-soft);
    max-width: 54ch;
  }

  .sect {
    padding: var(--space-9) 0 var(--space-7);
  }
  .sect-title {
    font-family: var(--font-display);
    font-style: italic;
    font-weight: 400;
    font-size: 28px;
    line-height: 1.1;
    letter-spacing: -0.015em;
    color: var(--content);
    margin: 0 0 var(--space-6);
  }
  .sect-lead {
    font-size: 14px;
    line-height: 1.55;
    color: var(--content-mute);
    max-width: 52ch;
    margin-bottom: var(--space-5);
  }
  .subhead {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin: var(--space-5) 0 var(--space-2);
  }

  .hair {
    height: 1px;
    border: 0;
    background: linear-gradient(
      90deg,
      transparent,
      var(--hair-strong) 8%,
      var(--hair-strong) 92%,
      transparent
    );
  }

  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-5);
    padding: var(--space-4) 0;
    flex-wrap: wrap;
  }
  .row-label {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .row-name {
    font-size: 15px;
    color: var(--content);
    line-height: 1.3;
  }
  .row-hint {
    font-size: 13px;
    color: var(--content-mute);
    line-height: 1.4;
  }
  .row-meta {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    color: var(--content-faint);
  }

  /* ── section stagger (mount-driven; reduced-motion skipped) ── */
  @keyframes sect-stagger-in {
    from { opacity: 0; transform: translateY(6px); }
    to   { opacity: 1; transform: translateY(0); }
  }
  @keyframes hair-stagger-in {
    from { opacity: 0; }
    to   { opacity: 1; }
  }
  .stagger-in {
    animation: sect-stagger-in var(--dur-slow) var(--ease) both;
    animation-delay: calc(var(--stagger-index, 0) * 40ms);
  }
  .hair-stagger-in {
    animation: hair-stagger-in var(--dur-slow) var(--ease) both;
    animation-delay: calc(var(--stagger-index, 0) * 40ms);
  }
  @media (prefers-reduced-motion: reduce) {
    .stagger-in,
    .hair-stagger-in {
      animation: none;
    }
  }

  /* segmented control (energy) — no <select>, ever */
  .seg {
    display: inline-flex;
    gap: 2px;
    padding: 2px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    background: var(--surface-card);
  }
  .seg-btn {
    padding: 6px 14px;
    font-size: 13px;
    line-height: 1;
    color: var(--content-mute);
    border-radius: var(--r-pill);
    text-transform: capitalize;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .seg-btn:hover:not(.active) {
    color: var(--content);
    background: color-mix(in oklab, var(--synapse) 8%, transparent);
  }
  .seg-btn:active:not(.active) {
    transform: scale(0.97);
  }
  .seg-btn:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .seg-btn.active {
    background: var(--synapse);
    color: var(--paper);
  }

  /* slider — pollen thumb, hairline track that fills with pollen as the
     user drags. We use a CSS variable for the progress fill that the
     <input> updates via inline style, so the track reads as a real
     progress bar instead of a flat hairline. */
  .slider {
    -webkit-appearance: none;
    appearance: none;
    width: 200px;
    max-width: 40vw;
    height: 4px;
    background:
      linear-gradient(
        to right,
        var(--pollen) 0,
        var(--pollen) var(--slider-fill, 50%),
        var(--hair-strong) var(--slider-fill, 50%),
        var(--hair-strong) 100%
      );
    border-radius: 2px;
    outline: none;
    cursor: pointer;
    transition: background var(--dur) var(--ease);
  }
  .slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: var(--paper);
    border: 1px solid var(--pollen);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--pollen) 20%, transparent);
    transition:
      transform var(--dur) var(--ease),
      background var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .slider::-webkit-slider-thumb:hover {
    transform: scale(1.18);
    background: var(--pollen);
    box-shadow: 0 0 0 5px color-mix(in oklab, var(--pollen) 28%, transparent);
  }
  .slider::-moz-range-thumb {
    width: 16px;
    height: 16px;
    background: var(--paper);
    border: 1px solid var(--pollen);
    border-radius: 50%;
    cursor: pointer;
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--pollen) 20%, transparent);
  }
  .slider:disabled {
    opacity: 0.42;
    cursor: not-allowed;
    filter: saturate(0.55);
  }
  .slider:focus-visible {
    outline: none;
  }
  .slider:focus-visible::-webkit-slider-thumb {
    box-shadow:
      0 0 0 3px color-mix(in oklab, var(--pollen) 35%, transparent),
      0 0 0 6px color-mix(in oklab, var(--pollen) 20%, transparent);
    background: var(--pollen);
  }
  .slider:focus-visible::-moz-range-thumb {
    box-shadow:
      0 0 0 3px color-mix(in oklab, var(--pollen) 35%, transparent),
      0 0 0 6px color-mix(in oklab, var(--pollen) 20%, transparent);
    background: var(--pollen);
  }
  .slider:active::-webkit-slider-thumb {
    transform: scale(1.18);
  }

  /* text field — underline only, paper voice */
  .field {
    width: 220px;
    max-width: 40vw;
    padding: 6px 2px;
    background: transparent;
    border: none;
    border-bottom: 1px solid var(--hair-strong);
    color: var(--content);
    font-size: 14px;
    font-family: var(--font-mono);
    transition: border-color var(--dur) var(--ease);
  }
  .field::placeholder {
    color: var(--content-faint);
  }
  .field:focus {
    outline: none;
    border-color: var(--synapse);
  }

  /* loading state for the matrix while the daemon config is still in flight */
  .matrix-loading {
    display: grid;
    grid-template-columns: 1fr;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-5) 0;
  }
  .loading-mark {
    font-family: var(--font-mono);
    font-size: 16px;
    color: var(--content-faint);
    text-align: center;
    letter-spacing: 0.2em;
  }
  .loading-rule {
    display: block;
    width: 100%;
    max-width: 360px;
    margin: 0 auto;
    color: var(--pollen);
  }

  /* adaptive strength — four dots in a row (re-uses the same .auto-dot
     class; the dots are also used by AutonomyMatrix.svelte but that
     component owns its own copy via Svelte's scoped CSS). */
  .dots4 {
    display: inline-flex;
    gap: 10px;
    padding: 2px;
  }
  .dots4 .auto-dot {
    width: 18px;
    height: 18px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid var(--dot-color, var(--hair-strong));
    background: transparent;
    display: grid;
    place-items: center;
    cursor: pointer;
    transition:
      border-color var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .dots4 .auto-dot:hover {
    transform: scale(1.08);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--dot-color, var(--content-mute)) 18%, transparent);
  }
  .dots4 .auto-dot.active {
    border-color: transparent;
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--dot-color, var(--pollen)) 30%, transparent);
  }
  .dots4 .auto-dot:focus-visible {
    outline: none;
    box-shadow: 0 0 0 3px var(--pollen-halo-color);
  }
  .dots4 .dot-fill {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--dot-color, var(--content));
    transform: scale(0);
    transition: transform var(--dur) var(--ease);
    pointer-events: none;
  }
  .dots4 .auto-dot.active .dot-fill {
    transform: scale(1);
  }

  /* voice status line */
  .status-line {
    font-size: 13px;
    color: var(--content-mute);
    margin-bottom: var(--space-3);
  }
  .status-line em {
    font-family: var(--font-display);
    font-style: italic;
    color: var(--content-soft);
  }

  /* toggle switch — 16-px hairline-ringed knob that fills to pollen on */
  .toggle {
    position: relative;
    width: 44px;
    height: 24px;
    flex: none;
    border-radius: var(--r-pill);
    border: 1px solid var(--hair-strong);
    background: var(--surface-card);
    transition:
      background var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease),
      transform var(--dur) var(--ease);
  }
  .toggle:hover {
    border-color: var(--hair-strong);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--synapse) 14%, transparent);
  }
  .toggle:active {
    transform: scale(0.97);
  }
  .toggle:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .toggle-knob {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: var(--content-mute);
    border: 1px solid transparent;
    box-shadow: 0 0 0 0 transparent;
    transition:
      transform var(--dur) var(--ease),
      background var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .toggle.on {
    background: var(--synapse);
    border-color: var(--synapse);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--synapse) 18%, transparent);
  }
  .toggle.on .toggle-knob {
    transform: translateX(20px);
    background: var(--pollen);
    border-color: var(--paper);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--pollen) 28%, transparent);
  }

  /* account */
  .account-chip {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-4);
    background: var(--surface-card);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
  }
  .avatar,
  .avatar-fallback {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    flex: none;
  }
  .avatar-fallback {
    display: grid;
    place-items: center;
    background: var(--synapse);
    color: var(--paper);
    font-family: var(--font-display);
    font-size: 16px;
  }
  .account-meta {
    min-width: 0;
    flex: 1;
  }
  .account-name {
    font-size: 14px;
    color: var(--content);
    line-height: 1.3;
  }
  .account-email {
    font-size: 12px;
    color: var(--content-mute);
    font-family: var(--font-mono);
  }
  .account-provider {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .account-out {
    font-size: 14px;
    color: var(--content-mute);
    line-height: 1.5;
  }
  .account-out :global(.thread-link) {
    padding: 1px 4px;
    border-radius: var(--r-xs);
    transition:
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease);
  }
  .account-out :global(.thread-link:hover) {
    background: color-mix(in oklab, var(--synapse) 8%, transparent);
  }
  .account-out :global(.thread-link:active) {
    transform: scale(0.97);
  }

  /* permissions */
  .perms {
    display: flex;
    flex-direction: column;
  }
  .perm {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-3) 0;
    border-top: 1px solid var(--hair);
  }
  .perm:first-child {
    border-top: none;
  }
  .perm-name {
    font-size: 14px;
    color: var(--content);
  }
  .perm-badge {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }
  .perm-badge[data-status='granted'] {
    color: var(--ok);
  }
  .perm-badge[data-status='denied'] {
    color: var(--danger);
  }
  .perm-badge[data-status='unknown'] {
    color: var(--content-faint);
  }

  /* legal */
  .eula-row {
    display: flex;
    width: 100%;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4) 0;
    text-align: left;
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
    border-radius: var(--r-sm);
  }
  .eula-row:hover {
    color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
    transform: translateX(2px);
  }
  .eula-row:active {
    transform: translateX(2px) scale(0.99);
  }
  .eula-row:focus-visible {
    outline: none;
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .eula-body {
    padding: var(--space-4) 0;
  }
  .eula-text {
    font-family: var(--font-mono);
    font-size: 12px;
    line-height: 1.6;
    color: var(--content-soft);
    white-space: pre-wrap;
    max-height: 320px;
    overflow: auto;
    padding: var(--space-4);
    background: var(--surface-card);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
  }
  .eula-updated {
    margin-top: var(--space-3);
  }
  .eula-loading {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
    padding: var(--space-4) 0;
  }
  .eula-loading-label {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .empty {
    font-size: 14px;
    color: var(--content-mute);
    font-style: italic;
  }

  /* ── sticky save bar (springs in, pollen) ── */
  .save-bar {
    position: sticky;
    bottom: 0;
    z-index: var(--z-sticky);
    margin-top: var(--space-8);
    padding: 0 0 var(--space-5);
    transform-origin: bottom center;
  }
  .save-bar-inner {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-3) var(--space-5);
    background: var(--pollen);
    color: var(--paper);
    border: 1px solid var(--pollen-deep);
    border-radius: var(--r-pill);
    box-shadow: var(--pollen-halo), var(--shadow-float);
    flex-wrap: wrap;
    transition: opacity 400ms var(--ease), transform 400ms var(--ease);
  }
  .save-bar:hover .save-bar-inner {
    box-shadow: 0 0 0 8px color-mix(in oklab, var(--pollen) 14%, transparent), var(--shadow-float);
  }
  .save-bar--saved .save-bar-inner {
    opacity: 0;
    transform: translateY(8px);
    pointer-events: none;
  }
  .save-bar--dirty .save-bar-inner {
    opacity: 1;
    transform: translateY(0);
  }
  .save-bar-thread {
    display: block;
    width: 100%;
    margin-top: var(--space-3);
    color: var(--pollen);
    opacity: 0.55;
  }
  .save-status {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--paper);
  }
  .save-status-text {
    font-size: 13px;
    line-height: 1;
    color: var(--paper);
  }
  .save-status-mono {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--paper);
  }
  .save-error {
    font-size: 11px;
    color: var(--paper);
    font-family: var(--font-mono);
    max-width: 30ch;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    opacity: 0.85;
  }
  .save-actions {
    margin-left: auto;
    display: flex;
    gap: var(--space-2);
  }
  /* The save-actions buttons sit on pollen paper; force paper-on-pollen
     contrast for both ghost (Revert) and primary (Save) variants. */
  .save-actions :global(.btn-ghost) {
    color: var(--paper);
    border-color: color-mix(in oklab, var(--paper) 40%, transparent);
  }
  .save-actions :global(.btn-ghost:hover) {
    background: color-mix(in oklab, var(--paper) 14%, transparent);
    color: var(--paper);
  }
  .save-actions :global(.btn-primary) {
    background: var(--paper);
    color: var(--pollen-deep);
    border-color: var(--paper);
  }
  .save-actions :global(.btn-primary:hover) {
    background: var(--paper-2);
    color: var(--pollen-deep);
  }

  @media (max-width: 640px) {
    .row {
      gap: var(--space-3);
    }
  }
</style>