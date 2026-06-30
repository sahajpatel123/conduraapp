<script lang="ts">
  import { onMount } from 'svelte'
  import { onboarding } from '../stores/onboarding.svelte'
  import { ipc } from '../ipc/client'
  import { setLocale, type Locale } from '../i18n'
  import IrisOrb from './IrisOrb.svelte'
  import HotkeyRecorder from './HotkeyRecorder.svelte'
  import SegmentedControl from './ui/SegmentedControl.svelte'
  import Switch from './ui/Switch.svelte'
  import Button from './ui/Button.svelte'

  // ── The Setup Console ──────────────────────────────────────
  // A mandatory, cinematic first-run floating console. The user
  // *selects* how Condura should behave rather than filling forms.
  // It gates entry until the required choices are made and is
  // re-openable from Settings.
  //
  // The console runs a local step controller so it remains fully
  // usable even before the embedded daemon answers (offline / dev
  // preview). Canonical gating steps (EULA, hotkey, permissions)
  // are persisted to the daemon best-effort; the richer selection
  // steps (provider, autonomy, language) are applied at finish.
  interface Props {
    onComplete?: (route?: string) => void
  }
  let { onComplete }: Props = $props()

  type StepId = 'welcome' | 'mind' | 'summon' | 'autonomy' | 'language' | 'ready'

  const steps: { id: StepId; rail: string; eyebrow: string; hue: string }[] = [
    { id: 'welcome',  rail: 'Welcome',   eyebrow: 'First light',     hue: '#8b7bff' },
    { id: 'mind',     rail: 'Connect',   eyebrow: 'A mind to think', hue: '#5a6bff' },
    { id: 'summon',   rail: 'Summon',    eyebrow: 'Your call',       hue: '#a99cff' },
    { id: 'autonomy', rail: 'Autonomy',  eyebrow: 'How much rope',   hue: '#ff8c6b' },
    { id: 'language', rail: 'Language',  eyebrow: 'In your words',   hue: '#c56be6' },
    { id: 'ready',    rail: 'Ready',     eyebrow: 'Come alive',      hue: '#57cc8b' }
  ]

  let index = $state(0)
  let dir = $state<'fwd' | 'back'>('fwd')
  const step = $derived(steps[index])

  // ── Selections ──
  let eulaAccepted = $state(false)
  let eulaScrolled = $state(false)
  let provider = $state('') // 'ollama' | 'skip' | provider id
  let apiKey = $state('')
  let hotkey = $state('')
  let autonomy = $state('cautious')
  let locale = $state<Locale>('en')
  let permsSkipped = $state(false)
  let ready = $state(false)
  let nudge = $state(false)

  const providers = [
    { id: 'anthropic', name: 'Anthropic', sub: 'Claude' },
    { id: 'openai', name: 'OpenAI', sub: 'GPT' },
    { id: 'google', name: 'Google', sub: 'Gemini' },
    { id: 'openrouter', name: 'OpenRouter', sub: '300+ models' }
  ]

  const languages: { code: Locale; name: string; native: string }[] = [
    { code: 'en', name: 'English', native: 'English' },
    { code: 'es', name: 'Spanish', native: 'Español' },
    { code: 'fr', name: 'French', native: 'Français' },
    { code: 'de', name: 'German', native: 'Deutsch' },
    { code: 'ja', name: 'Japanese', native: '日本語' },
    { code: 'zh', name: 'Mandarin', native: '中文' }
  ]

  const autonomyOptions = [
    { value: 'cautious', label: 'Cautious' },
    { value: 'balanced', label: 'Balanced' },
    { value: 'trusting', label: 'Trusting' }
  ]
  const autonomyBlurb = $derived(
    autonomy === 'cautious' ? "I'll ask before any action that writes, sends, or deletes." :
    autonomy === 'balanced' ? "I'll act on safe tasks and ask before anything risky." :
    "I'll act freely — but destructive actions always need your hand."
  )

  const ollamaReady = $derived(!!onboarding.power?.ollama_reachable)

  const canContinue = $derived(
    step.id === 'welcome' ? eulaAccepted :
    step.id === 'mind' ? provider !== '' :
    step.id === 'summon' ? hotkey !== '' :
    true
  )

  onMount(() => {
    void onboarding.sync()
    void onboarding.loadEula()
    void onboarding.probePower()
  })

  function onEulaScroll(e: Event): void {
    const el = e.currentTarget as HTMLElement
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 24) eulaScrolled = true
  }

  function chooseProvider(id: string): void {
    provider = provider === id ? '' : id
    if (id === 'ollama' || id === 'skip') apiKey = ''
  }

  function pickLanguage(code: Locale): void {
    locale = code
    setLocale(code)
  }

  function persistStep(): void {
    // Best-effort daemon persistence for the canonical gating steps.
    if (step.id === 'welcome') {
      void ipc.onboardingSetStep('eula', 'complete', onboarding.eulaVersion || 'v1').catch(() => {})
    } else if (step.id === 'summon') {
      onboarding.setHotkey(hotkey)
      void ipc.onboardingSetStep('hotkey', 'complete', hotkey).catch(() => {})
    } else if (step.id === 'ready') {
      void ipc.onboardingSetStep('permissions', permsSkipped ? 'skipped' : 'complete').catch(() => {})
    }
  }

  function back(): void {
    if (index === 0) return
    dir = 'back'
    index -= 1
  }

  async function next(): Promise<void> {
    if (!canContinue) {
      nudge = true
      setTimeout(() => (nudge = false), 380)
      return
    }
    persistStep()
    if (index < steps.length - 1) {
      dir = 'fwd'
      index += 1
      if (steps[index].id === 'ready') ready = true
      return
    }
    await finish()
  }

  async function finish(): Promise<void> {
    // Apply the richer selections best-effort, then complete.
    if (provider && provider !== 'skip' && provider !== 'ollama' && apiKey.trim()) {
      void ipc
        .apiKeysSet({ provider, label: 'Added during setup', secret: apiKey.trim() })
        .catch(() => {})
    }
    void ipc
      .configUpdate({ autonomy: { global_default: autonomy } } as unknown as Parameters<typeof ipc.configUpdate>[0])
      .catch(() => {})
    void onboarding
      .finish({ hotkey, eula_version: onboarding.eulaVersion || 'v1', permissions_skipped: permsSkipped })
      .catch(() => {})
    onComplete?.('#/')
  }
</script>

<div class="scrim" aria-hidden="true"></div>

<div
  class="console"
  style="--step-hue: {step.hue}"
  role="dialog"
  aria-modal="true"
  aria-label="Set up Condura"
>
  <span class="halo" aria-hidden="true"></span>

  <div class="body">
    <!-- Step rail -->
    <aside class="rail" aria-hidden="true">
      <span class="spine"></span>
      {#each steps as s, i (s.id)}
        <div class="node" class:done={i < index} class:current={i === index}>
          <span class="bead">
            {#if i < index}
              <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 8.5l2.5 2.5L12 5" /></svg>
            {/if}
          </span>
          <span class="rail-label">{s.rail}</span>
        </div>
      {/each}
    </aside>

    <!-- Stage -->
    <section class="stage">
      {#key step.id}
        <div class="frame" class:back={dir === 'back'}>
          <p class="eyebrow">{step.eyebrow}</p>

          {#if step.id === 'welcome'}
            <div class="welcome-head">
              <IrisOrb state="idle" size={40} />
              <h2 class="hero">Meet <span class="wordmark">Condura</span>.</h2>
            </div>
            <p class="lede">A calm, powerful intelligence that lives on your computer — and asks before it touches anything that matters.</p>
            <div class="eula-well" onscroll={onEulaScroll}>
              {#if onboarding.eula?.text}
                <pre class="eula-text">{onboarding.eula.text}</pre>
              {:else}
                <pre class="eula-text">Condura Freeware EULA

Condura is free for personal and commercial use. You may not
redistribute the binary. The license is revocable for abuse.

By accepting, you agree to use Condura responsibly. Condura acts
on your computer only with your consent; a deterministic Gatekeeper
gates every action, and you can stop the agent at any time.

Scroll to the end to continue.</pre>
              {/if}
            </div>
            <div class="accept">
              <Switch
                bind:checked={eulaAccepted}
                disabled={!eulaScrolled}
                label={eulaScrolled ? 'I accept the licence' : 'Scroll the licence to the end'}
              />
            </div>

          {:else if step.id === 'mind'}
            <h2 class="title">Where should I think?</h2>
            <p class="lede">Connect a model provider, or point me at a local Ollama. You can change this anytime.</p>
            <div class="grid">
              <button class="opt" class:sel={provider === 'ollama'} class:hot={ollamaReady} onclick={() => chooseProvider('ollama')}>
                <span class="opt-top"><span class="opt-name">Local Ollama</span>{#if ollamaReady}<span class="tag live">detected</span>{:else}<span class="tag">local</span>{/if}</span>
                <span class="opt-sub">{ollamaReady ? `${onboarding.power?.ollama_models?.length ?? 0} models · private, free` : 'Runs fully on your machine'}</span>
              </button>
              {#each providers as p (p.id)}
                <button class="opt" class:sel={provider === p.id} onclick={() => chooseProvider(p.id)}>
                  <span class="opt-top"><span class="opt-name">{p.name}</span></span>
                  <span class="opt-sub">{p.sub}</span>
                </button>
              {/each}
            </div>
            {#if provider && provider !== 'ollama' && provider !== 'skip'}
              <div class="key-row">
                <input class="key-input" type="password" placeholder="Paste your API key (optional now)" bind:value={apiKey} autocomplete="off" spellcheck="false" />
              </div>
            {/if}
            <button class="later" class:sel={provider === 'skip'} onclick={() => chooseProvider('skip')}>Decide later — use Ollama if it's running</button>

          {:else if step.id === 'summon'}
            <h2 class="title">How will you call me?</h2>
            <p class="lede">Pick a global shortcut. Tap it from anywhere and the overlay appears.</p>
            <div class="summon-wrap">
              <HotkeyRecorder value={hotkey} onRecord={(c) => (hotkey = c)} />
            </div>

          {:else if step.id === 'autonomy'}
            <h2 class="title">How much rope?</h2>
            <p class="lede">Set the brakes. You can fine-tune per app and per task later in Settings.</p>
            <div class="autonomy">
              <SegmentedControl options={autonomyOptions} bind:value={autonomy} />
              <div class="preview">
                <span class="preview-dot"></span>
                <span class="preview-text">{autonomyBlurb}</span>
              </div>
              <p class="safety-note">Destructive actions always require a real human at the keyboard — no exceptions.</p>
            </div>

          {:else if step.id === 'language'}
            <h2 class="title">In your words.</h2>
            <p class="lede">Choose your language. I'll reply in it regardless of the interface.</p>
            <div class="lang-grid">
              {#each languages as l (l.code)}
                <button class="lang" class:sel={locale === l.code} onclick={() => pickLanguage(l.code)}>
                  <span class="lang-native">{l.native}</span>
                  <span class="lang-name">{l.name}</span>
                </button>
              {/each}
            </div>

          {:else}
            <div class="ready-head">
              <IrisOrb state={ready ? 'acting' : 'idle'} size={56} />
            </div>
            <h2 class="title center">You're ready.</h2>
            <p class="lede center">Grant a couple of permissions so I can see and act — or skip and do it later.</p>
            <div class="perms">
              <div class="perm">
                <span class="perm-name">Accessibility</span>
                <span class="perm-status">Granted at runtime</span>
              </div>
              <div class="perm">
                <span class="perm-name">Screen Recording</span>
                <span class="perm-status">Granted at runtime</span>
              </div>
            </div>
            <label class="skip-perms">
              <input type="checkbox" bind:checked={permsSkipped} />
              <span>I'll grant permissions later</span>
            </label>
          {/if}
        </div>
      {/key}
    </section>
  </div>

  <!-- Footer -->
  <footer class="footer">
    <div class="progress" aria-hidden="true">
      {#each steps as _, i}
        <span class="seg" class:on={i <= index}></span>
      {/each}
    </div>
    <div class="actions">
      <Button variant="ghost" size="md" disabled={index === 0} onclick={back}>Back</Button>
      <div class="primary-wrap" class:nudge>
        <Button variant="primary" size="md" disabled={!canContinue} onclick={next}>
          {index === steps.length - 1 ? 'Enter Condura' : 'Continue'}
        </Button>
      </div>
    </div>
  </footer>
</div>

<style>
  .scrim {
    position: fixed;
    inset: 0;
    background: var(--overlay-scrim);
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    z-index: 0;
  }

  .console {
    position: fixed;
    top: 46%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: min(640px, 92vw);
    max-height: min(88vh, 760px);
    display: flex;
    flex-direction: column;
    border-radius: var(--radius-2xl);
    background: var(--surface-glass-strong);
    backdrop-filter: var(--blur-heavy);
    -webkit-backdrop-filter: var(--blur-heavy);
    border: 1px solid var(--border-strong);
    box-shadow: var(--shadow-xl), var(--inset-hair);
    z-index: 1;
    overflow: hidden;
    animation: modal-in var(--transition-spring) var(--ease-soft) both;
  }

  /* per-step glow that shifts hue */
  .halo {
    position: absolute;
    inset: -1px;
    border-radius: inherit;
    pointer-events: none;
    box-shadow: 0 0 60px -10px color-mix(in srgb, var(--step-hue) 60%, transparent);
    transition: box-shadow var(--dur-cinematic) var(--ease-soft);
  }

  .body { display: flex; min-height: 0; flex: 1; }

  /* ── Step rail ── */
  .rail {
    position: relative;
    width: 176px;
    flex-shrink: 0;
    padding: var(--space-7) var(--space-5);
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    border-right: 1px solid var(--border);
    background: rgba(0, 0, 0, 0.16);
  }
  .spine {
    position: absolute;
    left: calc(var(--space-5) + 9px);
    top: var(--space-8);
    bottom: var(--space-7);
    width: 1px;
    background: var(--border-strong);
  }
  .node { position: relative; display: flex; align-items: center; gap: var(--space-3); z-index: 1; }
  .bead {
    display: grid;
    place-items: center;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    color: var(--text-inverse);
    flex-shrink: 0;
    transition: all var(--transition-base);
  }
  .bead svg { width: 11px; height: 11px; }
  .node.done .bead { background: var(--accent); border-color: var(--accent); }
  .node.current .bead {
    background: var(--step-hue);
    border-color: var(--step-hue);
    box-shadow: 0 0 14px -2px var(--step-hue);
    animation: iris-breath 3s var(--ease-glide) infinite;
  }
  .rail-label {
    font-size: var(--size-sm);
    color: var(--text-faint);
    transition: color var(--transition-base);
  }
  .node.current .rail-label { color: var(--text); font-weight: var(--weight-medium); }
  .node.done .rail-label { color: var(--text-muted); }

  /* ── Stage ── */
  .stage { flex: 1; min-width: 0; padding: var(--space-7) var(--space-7) var(--space-5); overflow-y: auto; }
  .frame { animation: stage-in var(--dur-cinematic) var(--ease-soft) both; }
  .frame.back { animation: stage-in-back var(--dur-cinematic) var(--ease-soft) both; }

  .eyebrow {
    font-size: var(--size-2xs);
    font-weight: var(--weight-semibold);
    letter-spacing: var(--tracking-wider);
    text-transform: uppercase;
    color: var(--step-hue);
    margin-bottom: var(--space-3);
  }

  .title, .hero {
    font-family: var(--font-display);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    color: var(--text);
    line-height: 1.05;
  }
  .title { font-size: var(--display-md); margin-bottom: var(--space-3); }
  .title.center { text-align: center; }
  .hero { font-size: var(--display-lg); }
  .wordmark {
    background: var(--accent-gradient-warm);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .welcome-head { display: flex; align-items: center; gap: var(--space-4); margin-bottom: var(--space-4); }

  .lede {
    font-family: var(--font-display);
    font-style: italic;
    font-size: var(--size-lg);
    color: var(--text-muted);
    line-height: var(--leading-snug);
    margin-bottom: var(--space-5);
    max-width: 46ch;
  }
  .lede.center { text-align: center; margin-left: auto; margin-right: auto; }

  /* ── EULA well ── */
  .eula-well {
    height: 168px;
    overflow-y: auto;
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-4);
    margin-bottom: var(--space-4);
  }
  .eula-text {
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    white-space: pre-wrap;
    word-break: break-word;
  }
  .accept {
    padding: var(--space-2) var(--space-4);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
  }

  /* ── Provider grid ── */
  .grid { display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-3); margin-bottom: var(--space-3); }
  .opt {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: var(--space-4);
    text-align: left;
    background: var(--surface-2);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    transition: all var(--transition-fast);
  }
  .opt:hover { border-color: var(--border-strong); transform: translateY(-2px); box-shadow: var(--shadow-sm); }
  .opt.sel { border-color: var(--accent); background: var(--accent-soft); box-shadow: var(--glow-iris); }
  .opt.hot { border-color: var(--border-warm); }
  .opt-top { display: flex; align-items: center; justify-content: space-between; gap: var(--space-2); }
  .opt-name { font-size: var(--size-md); font-weight: var(--weight-medium); color: var(--text); }
  .opt-sub { font-size: var(--size-xs); color: var(--text-faint); }
  .tag {
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: var(--tracking-wide);
    color: var(--text-muted);
    padding: 2px 7px;
    border-radius: var(--radius-pill);
    background: var(--surface-3);
  }
  .tag.live { color: var(--accent-2); background: var(--accent-2-soft); }

  .key-row { margin-bottom: var(--space-3); }
  .key-input {
    width: 100%;
    padding: 10px 14px;
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    color: var(--text);
    font-family: var(--font-mono);
    font-size: var(--size-sm);
  }
  .key-input:focus { outline: none; border-color: var(--border-focus); box-shadow: var(--glow-focus); }
  .later {
    display: block;
    width: 100%;
    padding: 10px;
    text-align: center;
    background: transparent;
    border: 1px dashed var(--border-strong);
    border-radius: var(--radius-md);
    color: var(--text-muted);
    font-size: var(--size-sm);
    transition: all var(--transition-fast);
  }
  .later:hover { color: var(--text); border-color: var(--border-focus); }
  .later.sel { color: var(--accent); border-style: solid; border-color: var(--accent); background: var(--accent-soft); }

  .summon-wrap { margin-top: var(--space-2); }

  /* ── Autonomy ── */
  .autonomy { display: flex; flex-direction: column; gap: var(--space-4); }
  .preview {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-4);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
  }
  .preview-dot {
    width: 8px; height: 8px; border-radius: 50%;
    background: var(--step-hue);
    box-shadow: 0 0 10px var(--step-hue);
    flex-shrink: 0;
  }
  .preview-text { font-size: var(--size-md); color: var(--text); }
  .safety-note { font-size: var(--size-xs); color: var(--accent-2); display: flex; align-items: center; gap: 6px; }

  /* ── Language grid ── */
  .lang-grid { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: var(--space-3); }
  .lang {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: var(--space-4);
    background: var(--surface-2);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    transition: all var(--transition-fast);
  }
  .lang:hover { border-color: var(--border-strong); transform: translateY(-2px); }
  .lang.sel { border-color: var(--accent); background: var(--accent-soft); box-shadow: var(--glow-iris); }
  .lang-native { font-size: var(--size-lg); color: var(--text); }
  .lang-name { font-size: var(--size-xs); color: var(--text-faint); }

  /* ── Ready ── */
  .ready-head { display: flex; justify-content: center; margin-bottom: var(--space-4); }
  .perms { display: flex; flex-direction: column; gap: var(--space-2); margin: 0 auto var(--space-4); max-width: 380px; }
  .perm {
    display: flex; align-items: center; justify-content: space-between;
    padding: var(--space-3) var(--space-4);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
  }
  .perm-name { font-size: var(--size-md); color: var(--text); }
  .perm-status { font-size: var(--size-xs); color: var(--success); font-family: var(--font-mono); }
  .skip-perms {
    display: flex; align-items: center; gap: var(--space-2); justify-content: center;
    font-size: var(--size-sm); color: var(--text-muted); cursor: pointer;
  }
  .skip-perms input { accent-color: var(--accent); }

  /* ── Footer ── */
  .footer {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: var(--space-5);
    padding: var(--space-4) var(--space-7);
    border-top: 1px solid var(--border);
    background: rgba(0, 0, 0, 0.16);
  }
  .progress { display: flex; gap: 6px; flex: 1; }
  .seg {
    flex: 1; height: 4px; border-radius: var(--radius-pill);
    background: var(--surface-3);
    transition: background var(--transition-base);
  }
  .seg.on { background: var(--accent-gradient); }
  .actions { display: flex; align-items: center; gap: var(--space-2); }
  .primary-wrap.nudge { animation: nudge 0.38s var(--ease-soft); }

  @media (max-width: 560px) {
    .rail { display: none; }
    .grid, .lang-grid { grid-template-columns: 1fr 1fr; }
  }
  @media (prefers-reduced-motion: reduce) {
    .console, .frame, .node.current .bead { animation: none !important; }
  }
</style>
