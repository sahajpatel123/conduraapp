<script lang="ts">
  /**
   * Meridian About — living instrument colophon.
   * Signature: a vertical meridian spine with seven lit stations.
   * Logic: version · capabilities · ping · clipboard · donate.
   */
  import { onMount } from 'svelte'
  import { ipc } from '../../ipc/client'
  import { daemon } from '../../stores/daemon.svelte'
  import type { DaemonCapabilities, VersionInfo } from '../../ipc/types'

  const DONATE_URL = 'https://condura.app/donate'
  const SITE_URL = 'https://condura.app'
  const GUI_VERSION = '0.1.0'

  type Station = {
    id: string
    roman: string
    title: string
    body: string
    cite: string
    live?: () => { ok: boolean; note: string }
  }

  let version = $state<VersionInfo | null>(null)
  let caps = $state<DaemonCapabilities | null>(null)
  let loading = $state(true)
  let pingMs = $state<number | null>(null)
  let active = $state('i')
  let copied = $state(false)
  let entered = $state(false)
  let reduceMotion = $state(false)

  const connected = $derived(daemon.connected)

  const STATIONS: Station[] = [
    {
      id: 'i',
      roman: 'I',
      title: 'Separate minds',
      body: 'The Strategist may be any model. The Gatekeeper is deterministic code. They are never the same system — planning and permission stay apart.',
      cite: 'gatekeeper',
    },
    {
      id: 'ii',
      roman: 'II',
      title: 'One door to action',
      body: 'No click, type, or shell leaves Condura without Gatekeeper. Model text alone cannot touch your machine.',
      cite: 'gatekeeper',
    },
    {
      id: 'iii',
      roman: 'III',
      title: 'A human at the keys',
      body: 'Destructive work waits on a real Allow or Deny. There is no silent escalate path — consent is the lock.',
      cite: 'consent',
      live: () => ({ ok: true, note: 'Consent sheets live in Meridian' }),
    },
    {
      id: 'iv',
      roman: 'IV',
      title: 'You can always stop',
      body: 'Hard hotkey. Dock Halt. Watchdog. Network isolation. Four independent ways to cut the line.',
      cite: 'halt',
      live: () =>
        caps
          ? {
              ok: !!(caps.kill_switch.layer1_hotkey && caps.kill_switch.layer2_watchdog),
              note: caps.kill_switch.layer1_hotkey
                ? `Hotkey${caps.kill_switch.layer2_watchdog ? ' · watchdog' : ''} armed`
                : 'Kill layers unread',
            }
          : { ok: false, note: connected ? 'Reading kill-switch…' : 'Offline' },
    },
    {
      id: 'v',
      roman: 'V',
      title: 'Nothing forgotten',
      body: 'HMAC-chained, append-only audit. If something goes wrong, the ledger can prove what happened.',
      cite: 'audit',
      live: () =>
        caps
          ? {
              ok: caps.audit.hmac_subkey,
              note: caps.audit.hmac_subkey
                ? `Chain live${caps.audit.redaction ? ' · redaction' : ''}`
                : 'Audit unread',
            }
          : { ok: false, note: connected ? 'Reading audit…' : 'Offline' },
    },
    {
      id: 'vi',
      roman: 'VI',
      title: 'Guest, not owner',
      body: 'Condura asks to enter rooms. You grant or deny. It never escalates privileges on its own.',
      cite: 'permissions',
    },
    {
      id: 'vii',
      roman: 'VII',
      title: 'Your machine decides',
      body: 'Screen, accessibility, and input access are granted by you in the OS. Condura only asks.',
      cite: 'os',
    },
  ]

  const activeStation = $derived(STATIONS.find((s) => s.id === active) ?? STATIONS[0]!)
  const activeIndex = $derived(STATIONS.findIndex((s) => s.id === active))

  const buildLine = $derived.by(() => {
    if (!version) return `Meridian ${GUI_VERSION}`
    const v = version.version || GUI_VERSION
    const c = version.commit ? version.commit.slice(0, 7) : ''
    return c ? `${v} · ${c}` : v
  })

  onMount(() => {
    reduceMotion = matchMedia('(prefers-reduced-motion: reduce)').matches
    requestAnimationFrame(() => {
      entered = true
    })
    void refresh()

    const onKey = (e: KeyboardEvent) => {
      const mod = e.metaKey || e.ctrlKey
      if (mod && e.key.toLowerCase() === 'd') {
        e.preventDefault()
        openDonate()
        return
      }
      if (e.key === 'ArrowDown' || e.key === 'j') {
        e.preventDefault()
        step(1)
      } else if (e.key === 'ArrowUp' || e.key === 'k') {
        e.preventDefault()
        step(-1)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  })

  function step(dir: number): void {
    const next = Math.max(0, Math.min(STATIONS.length - 1, activeIndex + dir))
    active = STATIONS[next]!.id
  }

  async function refresh(): Promise<void> {
    loading = true
    pingMs = null
    try {
      const t0 = performance.now()
      const [v, c] = await Promise.all([
        ipc.version().catch(() => null),
        ipc.daemonCapabilities().catch(() => null),
      ])
      try {
        await ipc.ping()
        pingMs = Math.round(performance.now() - t0)
      } catch {
        /* offline */
      }
      version = v
      caps = c
    } finally {
      loading = false
    }
  }

  function openDonate(): void {
    window.open(DONATE_URL, '_blank', 'noopener,noreferrer')
  }

  function openSite(): void {
    window.open(SITE_URL, '_blank', 'noopener,noreferrer')
  }

  function goAudit(): void {
    window.location.hash = '#/audit'
  }

  function copyBuild(): void {
    const card = [
      `Condura ${version?.version ?? GUI_VERSION}`,
      version?.commit ? `commit ${version.commit}` : null,
      version?.build_date ? `built ${version.build_date}` : null,
      version?.platform ? `platform ${version.platform}` : null,
      version?.go_version ? `go ${version.go_version}` : null,
      `gui ${GUI_VERSION}`,
      `daemon ${connected ? 'connected' : 'offline'}`,
      pingMs != null ? `ping ${pingMs}ms` : null,
    ]
      .filter(Boolean)
      .join('\n')
    void navigator.clipboard.writeText(card).then(() => {
      copied = true
      setTimeout(() => (copied = false), 1600)
    })
  }

  function led(on: boolean | undefined): 'on' | 'off' | 'dim' {
    if (on === true) return 'on'
    if (on === false) return 'off'
    return 'dim'
  }
</script>

<article class="colophon" class:in={entered} class:calm={reduceMotion}>
  <!-- Atmosphere -->
  <div class="atmosphere" aria-hidden="true">
    <div class="wash"></div>
    <div class="grain"></div>
    <div class="orbit"></div>
  </div>

  <!-- Thesis: brand first -->
  <header class="thesis">
    <p class="edition">Instrument desk · colophon</p>
    <h1 class="brand">
      <span class="word">Condura</span>
      <span class="slash" aria-hidden="true">/</span>
      <span class="meridian">Meridian</span>
    </h1>
    <p class="thesis-line">
      Free. Local. Consent before action.<br />
      <em>This page is the contract — and the live reading of the machine that keeps it.</em>
    </p>

    <div class="seal-row">
      <button
        type="button"
        class="seal"
        onclick={copyBuild}
        aria-label="Copy build identity"
        title="Click to copy build card"
      >
        <span class="seal-ring" aria-hidden="true"></span>
        <span class="seal-core">
          <span class="seal-k">Build</span>
          <span class="seal-v">{buildLine}</span>
          <span class="seal-a">{copied ? 'Copied to clipboard' : 'Press to copy'}</span>
        </span>
      </button>

      <div class="vitals" aria-label="Daemon vitals">
        <span class="vital" data-tone={connected ? 'on' : 'off'}>
          <i></i>
          {connected ? 'Daemon live' : 'Daemon quiet'}
        </span>
        {#if pingMs != null}
          <span class="vital dim">{pingMs} ms</span>
        {/if}
        {#if version?.platform}
          <span class="vital dim">{version.platform}</span>
        {/if}
        <button type="button" class="vital ghost" onclick={() => void refresh()} disabled={loading}>
          {loading ? 'Reading…' : 'Refresh'}
        </button>
      </div>
    </div>
  </header>

  <!-- Signature: meridian spine + station detail -->
  <section class="spine-section" aria-label="Seven promises">
    <div class="spine-head">
      <h2>Along the meridian</h2>
      <p>Seven stations. Order matters. Use ↑↓ or J/K.</p>
    </div>

    <div class="spine-stage">
      <div class="spine" aria-hidden="true">
        <svg class="beam" viewBox="0 0 40 560" preserveAspectRatio="none">
          <defs>
            <linearGradient id="mdAboutBeam" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="var(--md-live)" stop-opacity="0.15" />
              <stop offset="35%" stop-color="var(--md-cobalt)" stop-opacity="0.95" />
              <stop offset="100%" stop-color="var(--md-live)" stop-opacity="0.25" />
            </linearGradient>
          </defs>
          <path
            class="track"
            d="M20 8 C 20 80, 20 480, 20 552"
            fill="none"
            stroke="var(--md-line-strong)"
            stroke-width="2"
          />
          <path
            class="glow"
            d="M20 8 C 20 80, 20 480, 20 552"
            fill="none"
            stroke="url(#mdAboutBeam)"
            stroke-width="3"
            stroke-linecap="round"
            pathLength="100"
            style="--progress: {(activeIndex + 1) * (100 / STATIONS.length)}"
          />
        </svg>

        <ol class="nodes">
          {#each STATIONS as s, i (s.id)}
            <li style="--i:{i}">
              <button
                type="button"
                class="node"
                class:on={active === s.id}
                class:passed={i <= activeIndex}
                aria-current={active === s.id ? 'true' : undefined}
                aria-label={`Station ${s.roman}: ${s.title}`}
                onclick={() => (active = s.id)}
              >
                <span class="node-dot"></span>
                <span class="node-roman">{s.roman}</span>
              </button>
            </li>
          {/each}
        </ol>
      </div>

      <div class="station">
        {#key active}
          <div class="station-inner">
            <p class="station-roman">{activeStation.roman}</p>
            <h3>{activeStation.title}</h3>
            <p class="station-body">{activeStation.body}</p>
            <div class="station-meta">
              <span class="cite">{activeStation.cite}</span>
              {#if activeStation.live}
                {@const live = activeStation.live()}
                <span class="live" data-ok={live.ok}>{live.note}</span>
              {/if}
            </div>
          </div>
        {/key}
        <div class="station-nav">
          <button type="button" class="nav-btn" disabled={activeIndex <= 0} onclick={() => step(-1)}>
            ← Prev
          </button>
          <span class="count">{activeIndex + 1} / {STATIONS.length}</span>
          <button
            type="button"
            class="nav-btn"
            disabled={activeIndex >= STATIONS.length - 1}
            onclick={() => step(1)}
          >
            Next →
          </button>
        </div>
      </div>
    </div>
  </section>

  <!-- Quiet instrument readout -->
  <section class="readout" aria-label="What this build can do">
    <div class="readout-label">
      <span>Instrument readout</span>
      <span class="sub">daemon.capabilities · facts only</span>
    </div>
    {#if !caps && !loading}
      <p class="readout-empty">Connect the daemon to light the board.</p>
    {:else}
      <ul class="leds">
        <li data-led={led(caps?.kill_switch.layer1_hotkey)}>
          <i></i><span>Hotkey</span>
        </li>
        <li data-led={led(caps?.kill_switch.layer2_watchdog)}>
          <i></i><span>Watchdog</span>
        </li>
        <li data-led={led(caps?.kill_switch.layer3_network_isolation.in_process)}>
          <i></i><span>Net isolate</span>
        </li>
        <li data-led={led(caps?.audit.hmac_subkey)}>
          <i></i><span>HMAC chain</span>
        </li>
        <li data-led={led(caps?.audit.redaction)}>
          <i></i><span>Redaction</span>
        </li>
        <li data-led={caps ? 'on' : 'dim'}>
          <i></i>
          <span class="mono">
            {#if caps}
              {caps.computer_use.orax}/{caps.computer_use.mac_cua}
            {:else}
              …
            {/if}
          </span>
        </li>
      </ul>
    {/if}
  </section>

  <!-- Closing plate -->
  <footer class="plate">
    <div class="plate-copy">
      <p class="plate-k">Free means free</p>
      <p class="plate-v">If Condura earns your trust, help keep it independent.</p>
    </div>
    <div class="plate-actions">
      <button type="button" class="plate-btn" onclick={goAudit}>
        <span class="plate-btn-ico" aria-hidden="true">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
            <path d="M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
            <path d="M9 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v1H9V5Z" stroke="currentColor" stroke-width="1.8"/>
            <path d="M9 12h6M9 16h4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
          </svg>
        </span>
        <span>Audit</span>
      </button>
      <button type="button" class="plate-btn" onclick={openSite}>
        <span class="plate-btn-ico" aria-hidden="true">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
            <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.8"/>
            <path d="M3 12h18M12 3c2.5 2.8 3.8 5.8 3.8 9S14.5 18.2 12 21c-2.5-2.8-3.8-5.8-3.8-9S9.5 5.8 12 3Z" stroke="currentColor" stroke-width="1.8"/>
          </svg>
        </span>
        <span>condura.app</span>
      </button>
      <button type="button" class="plate-btn plate-btn-primary" onclick={openDonate}>
        <span class="plate-btn-ico" aria-hidden="true">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
            <path d="M12 21s-7-4.4-7-10a4 4 0 0 1 7-2.5A4 4 0 0 1 19 11c0 5.6-7 10-7 10Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/>
          </svg>
        </span>
        <span>Donate</span>
      </button>
    </div>
    <p class="plate-hint">⌘D opens donate · ↑↓ walks the meridian</p>
  </footer>
</article>

<style>
  .colophon {
    --about-ease: cubic-bezier(0.22, 1, 0.36, 1);
    --about-spring: cubic-bezier(0.34, 1.35, 0.64, 1);
    position: relative;
    max-width: 920px;
    margin: 0 auto;
    padding: 36px 32px 130px;
    min-height: 100%;
    isolation: isolate;
  }

  .atmosphere {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 0;
    overflow: hidden;
    border-radius: 0;
    /* Dissolve atmosphere before the spine so nothing reads as a hard band */
    -webkit-mask-image: linear-gradient(
      180deg,
      #000 0%,
      #000 42%,
      rgb(0 0 0 / 0.55) 68%,
      transparent 100%
    );
    mask-image: linear-gradient(
      180deg,
      #000 0%,
      #000 42%,
      rgb(0 0 0 / 0.55) 68%,
      transparent 100%
    );
  }
  .wash {
    position: absolute;
    inset: -28% -18% -8%;
    background:
      radial-gradient(
        ellipse 78% 62% at 14% 6%,
        color-mix(in oklab, var(--md-cobalt) 22%, transparent) 0%,
        color-mix(in oklab, var(--md-cobalt) 10%, transparent) 28%,
        color-mix(in oklab, var(--md-cobalt) 4%, transparent) 52%,
        transparent 78%
      ),
      radial-gradient(
        ellipse 58% 48% at 88% 4%,
        color-mix(in oklab, var(--md-live) 14%, transparent) 0%,
        color-mix(in oklab, var(--md-live) 5%, transparent) 36%,
        transparent 72%
      ),
      radial-gradient(
        ellipse 42% 34% at 46% 22%,
        color-mix(in oklab, var(--md-cobalt) 7%, transparent) 0%,
        transparent 70%
      );
    filter: blur(36px);
    transform: translateZ(0);
    opacity: 0;
    transition: opacity 1100ms var(--about-ease);
  }
  .colophon.in .wash {
    opacity: 0.9;
  }
  .grain {
    position: absolute;
    inset: 0;
    opacity: 0.028;
    background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
    mix-blend-mode: multiply;
  }
  :global(:root[data-mode='dark']) .grain {
    opacity: 0.07;
    mix-blend-mode: soft-light;
  }
  .orbit {
    position: absolute;
    width: 420px;
    height: 420px;
    right: -120px;
    top: 40px;
    border-radius: 50%;
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, transparent);
    opacity: 0;
    transform: scale(0.92);
    transition:
      opacity 1s var(--about-ease) 120ms,
      transform 1.1s var(--about-spring) 80ms;
  }
  .orbit::after {
    content: '';
    position: absolute;
    inset: 28px;
    border-radius: 50%;
    border: 1px dashed color-mix(in oklab, var(--md-live) 22%, transparent);
    animation: about-spin 48s linear infinite;
  }
  .colophon.in .orbit {
    opacity: 1;
    transform: scale(1);
  }

  .thesis,
  .spine-section,
  .readout,
  .plate {
    position: relative;
    z-index: 1;
  }

  .thesis {
    margin-bottom: 56px;
    opacity: 0;
    transform: translateY(18px);
    transition:
      opacity 700ms var(--about-ease),
      transform 700ms var(--about-ease);
  }
  .colophon.in .thesis {
    opacity: 1;
    transform: none;
  }

  .edition {
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.2em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 18px;
  }

  .brand {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 10px 14px;
    margin: 0 0 22px;
    line-height: 0.92;
  }
  .word {
    font-family: var(--md-font-display);
    font-size: clamp(52px, 11vw, 96px);
    font-weight: 700;
    letter-spacing: -0.065em;
    color: var(--md-ink);
  }
  .slash {
    font-family: var(--md-font-display);
    font-size: clamp(36px, 7vw, 64px);
    font-weight: 500;
    color: color-mix(in oklab, var(--md-cobalt) 55%, var(--md-ink-faint));
    transform: translateY(-0.08em);
  }
  .meridian {
    font-family: var(--md-font-mono);
    font-size: clamp(14px, 2.4vw, 18px);
    font-weight: 500;
    letter-spacing: 0.28em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    align-self: center;
  }

  .thesis-line {
    margin: 0 0 32px;
    max-width: 34ch;
    font-size: clamp(17px, 2.2vw, 20px);
    line-height: 1.45;
    color: var(--md-ink-soft);
    font-weight: 500;
    letter-spacing: -0.02em;
  }
  .thesis-line em {
    font-style: normal;
    color: var(--md-ink-mute);
    font-weight: 450;
    font-size: 0.92em;
  }

  .seal-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 20px 28px;
  }

  .seal {
    position: relative;
    display: grid;
    place-items: center;
    width: min(100%, 280px);
    min-height: 108px;
    padding: 18px 22px;
    text-align: left;
    cursor: pointer;
    border-radius: 28px;
    background: color-mix(in oklab, var(--md-surface) 72%, transparent);
    border: 1px solid var(--md-line-strong);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    box-shadow: var(--md-shadow);
    transition:
      transform 240ms var(--about-spring),
      border-color 220ms var(--about-ease),
      box-shadow 220ms var(--about-ease);
  }
  .seal:hover {
    transform: translateY(-3px) rotate(-0.4deg);
    border-color: color-mix(in oklab, var(--md-cobalt) 40%, var(--md-line-strong));
    box-shadow: var(--md-shadow-lift);
  }
  .seal:focus-visible {
    outline: none;
    box-shadow: var(--md-focus), var(--md-shadow-lift);
  }
  .seal:active {
    transform: scale(0.985);
  }
  .seal-ring {
    position: absolute;
    inset: 8px;
    border-radius: 22px;
    border: 1px dashed color-mix(in oklab, var(--md-cobalt) 28%, transparent);
    opacity: 0.7;
    pointer-events: none;
  }
  .seal-core {
    display: flex;
    flex-direction: column;
    gap: 4px;
    width: 100%;
  }
  .seal-k {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .seal-v {
    font-family: var(--md-font-mono);
    font-size: 15px;
    letter-spacing: 0.02em;
    color: var(--md-ink);
    font-weight: 500;
  }
  .seal-a {
    font-size: 12px;
    color: var(--md-cobalt);
    font-weight: 600;
    margin-top: 4px;
  }

  .vitals {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }
  .vital {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-ink-mute);
    padding: 8px 12px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 55%, transparent);
  }
  .vital i {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: currentColor;
    box-shadow: 0 0 0 3px color-mix(in oklab, currentColor 20%, transparent);
  }
  .vital[data-tone='on'] {
    color: var(--md-live);
    border-color: color-mix(in oklab, var(--md-live) 28%, transparent);
  }
  .vital[data-tone='off'] {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 26%, transparent);
  }
  .vital.dim {
    color: var(--md-ink-faint);
  }
  .vital.ghost {
    cursor: pointer;
    color: var(--md-ink-soft);
    transition: border-color 160ms var(--about-ease), color 160ms var(--about-ease);
  }
  .vital.ghost:hover:not(:disabled) {
    border-color: var(--md-cobalt);
    color: var(--md-ink);
  }
  .vital.ghost:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .vital.ghost:disabled {
    opacity: 0.5;
    cursor: wait;
  }

  /* —— Meridian spine —— */
  .spine-section {
    margin-bottom: 48px;
    opacity: 0;
    transform: translateY(22px);
    transition:
      opacity 800ms var(--about-ease) 140ms,
      transform 800ms var(--about-ease) 140ms;
  }
  .colophon.in .spine-section {
    opacity: 1;
    transform: none;
  }
  .spine-head {
    margin-bottom: 22px;
  }
  .spine-head h2 {
    font-family: var(--md-font-display);
    font-size: clamp(26px, 4vw, 34px);
    letter-spacing: -0.045em;
    margin: 0 0 8px;
  }
  .spine-head p {
    margin: 0;
    font-size: 14px;
    color: var(--md-ink-mute);
  }

  .spine-stage {
    display: grid;
    grid-template-columns: 88px 1fr;
    gap: 28px;
    align-items: stretch;
    min-height: 420px;
  }

  .spine {
    position: relative;
    min-height: 420px;
  }
  .beam {
    position: absolute;
    inset: 0;
    width: 40px;
    height: 100%;
    left: 50%;
    transform: translateX(-50%);
  }
  .glow {
    stroke-dasharray: var(--progress, 14) 100;
    filter: drop-shadow(0 0 6px color-mix(in oklab, var(--md-cobalt) 35%, transparent));
    transition: stroke-dasharray 480ms var(--about-ease);
  }
  .nodes {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    align-items: center;
    padding: 4px 0;
    list-style: none;
    margin: 0;
  }
  .node {
    position: relative;
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    cursor: pointer;
    border-radius: 50%;
    transition: transform 200ms var(--about-spring);
  }
  .node:hover {
    transform: scale(1.08);
  }
  .node:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .node-dot {
    position: absolute;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--md-stage);
    border: 2px solid var(--md-line-strong);
    transition:
      background 220ms var(--about-ease),
      border-color 220ms var(--about-ease),
      box-shadow 220ms var(--about-ease),
      transform 220ms var(--about-spring);
  }
  .node.passed .node-dot {
    background: color-mix(in oklab, var(--md-cobalt) 35%, var(--md-surface));
    border-color: var(--md-cobalt);
  }
  .node.on .node-dot {
    background: var(--md-cobalt);
    border-color: var(--md-cobalt);
    transform: scale(1.35);
    box-shadow:
      0 0 0 6px color-mix(in oklab, var(--md-cobalt) 18%, transparent),
      0 8px 20px -8px color-mix(in oklab, var(--md-cobalt) 55%, transparent);
  }
  .node-roman {
    position: absolute;
    left: calc(100% + 10px);
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    color: var(--md-ink-faint);
    opacity: 0;
    transform: translateX(-4px);
    transition:
      opacity 200ms var(--about-ease),
      transform 200ms var(--about-ease),
      color 200ms var(--about-ease);
  }
  .node.on .node-roman,
  .node:hover .node-roman,
  .node:focus-visible .node-roman {
    opacity: 1;
    transform: none;
    color: var(--md-cobalt);
  }

  .station {
    position: relative;
    padding: 28px 32px 24px;
    border-radius: 28px;
    background: color-mix(in oklab, var(--md-surface) 82%, transparent);
    border: 1px solid var(--md-line-strong);
    box-shadow: var(--md-shadow-lift);
    backdrop-filter: blur(14px);
    -webkit-backdrop-filter: blur(14px);
    display: flex;
    flex-direction: column;
    min-height: 360px;
    overflow: hidden;
  }
  .station::before {
    content: '';
    position: absolute;
    left: 0;
    top: 24px;
    bottom: 24px;
    width: 3px;
    border-radius: 3px;
    background: linear-gradient(180deg, var(--md-cobalt), var(--md-live));
  }
  .station-inner {
    flex: 1;
    display: flex;
    flex-direction: column;
    animation: about-station 420ms var(--about-ease) both;
  }
  @keyframes about-station {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: none;
    }
  }
  .station-roman {
    font-family: var(--md-font-mono);
    font-size: 12px;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: var(--md-cobalt);
    margin: 0 0 12px;
  }
  .station h3 {
    font-family: var(--md-font-display);
    font-size: clamp(28px, 4.5vw, 40px);
    letter-spacing: -0.05em;
    line-height: 1.05;
    margin: 0 0 16px;
    max-width: 14ch;
  }
  .station-body {
    margin: 0;
    font-size: 16px;
    line-height: 1.6;
    color: var(--md-ink-mute);
    max-width: 42ch;
    flex: 1;
  }
  .station-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    align-items: center;
    margin: 22px 0 18px;
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    color: var(--md-ink-faint);
  }
  .live {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 5px 10px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    color: var(--md-ink-faint);
  }
  .live[data-ok='true'] {
    color: var(--md-live);
    border-color: color-mix(in oklab, var(--md-live) 32%, transparent);
    background: color-mix(in oklab, var(--md-live) 10%, transparent);
  }
  .station-nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding-top: 14px;
    border-top: 1px solid var(--md-line);
  }
  .nav-btn {
    font-size: 13px;
    font-weight: 700;
    color: var(--md-ink-soft);
    padding: 8px 4px;
    cursor: pointer;
    transition: color 160ms var(--about-ease);
  }
  .nav-btn:hover:not(:disabled) {
    color: var(--md-cobalt);
  }
  .nav-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }
  .nav-btn:focus-visible {
    outline: none;
    color: var(--md-cobalt);
    box-shadow: var(--md-focus);
    border-radius: 8px;
  }
  .count {
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    color: var(--md-ink-faint);
    font-variant-numeric: tabular-nums;
  }

  /* —— Readout —— */
  .readout {
    margin-bottom: 40px;
    padding: 18px 20px;
    border-radius: 22px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-stage) 65%, transparent);
    opacity: 0;
    transform: translateY(16px);
    transition:
      opacity 700ms var(--about-ease) 220ms,
      transform 700ms var(--about-ease) 220ms;
  }
  .colophon.in .readout {
    opacity: 1;
    transform: none;
  }
  .readout-label {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 14px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .readout-label .sub {
    color: var(--md-ink-mute);
    letter-spacing: 0.06em;
    text-transform: none;
  }
  .readout-empty {
    margin: 0;
    font-size: 13px;
    color: var(--md-ink-mute);
  }
  .leds {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 14px;
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .leds li {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    font-weight: 600;
    color: var(--md-ink-soft);
    padding: 6px 2px;
  }
  .leds i {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--md-ink-faint);
    box-shadow: inset 0 0 0 1px var(--md-line-strong);
  }
  .leds li[data-led='on'] i {
    background: var(--md-live);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-live) 22%, transparent);
  }
  .leds li[data-led='off'] i {
    background: var(--md-halt);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-halt) 18%, transparent);
  }
  .leds li[data-led='dim'] {
    color: var(--md-ink-faint);
  }
  .mono {
    font-family: var(--md-font-mono);
    font-size: 11px;
    font-weight: 500;
    letter-spacing: 0.04em;
  }

  /* —— Plate —— */
  .plate {
    display: grid;
    gap: 16px;
    padding-top: 8px;
    border-top: 1px solid var(--md-line);
    opacity: 0;
    transition: opacity 700ms var(--about-ease) 280ms;
  }
  .colophon.in .plate {
    opacity: 1;
  }
  .plate-k {
    margin: 0 0 6px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .plate-v {
    margin: 0;
    font-family: var(--md-font-display);
    font-size: 22px;
    letter-spacing: -0.035em;
    line-height: 1.2;
    max-width: 28ch;
  }
  .plate-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }
  .plate-btn {
    appearance: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 44px;
    padding: 0 18px;
    border-radius: 999px;
    border: 1px solid color-mix(in oklab, var(--md-ink) 14%, transparent);
    background: color-mix(in oklab, var(--md-surface) 92%, transparent);
    color: var(--md-ink);
    font-family: var(--md-font-sans);
    font-size: 13px;
    font-weight: 700;
    letter-spacing: -0.01em;
    line-height: 1;
    cursor: pointer;
    box-shadow:
      0 1px 0 color-mix(in oklab, #fff 70%, transparent) inset,
      0 8px 22px -16px color-mix(in oklab, var(--md-ink) 35%, transparent);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    transition:
      transform 180ms var(--about-spring),
      background 180ms var(--about-ease),
      border-color 180ms var(--about-ease),
      box-shadow 220ms var(--about-ease),
      color 180ms var(--about-ease);
  }
  .plate-btn:hover {
    transform: translateY(-1px);
    border-color: color-mix(in oklab, var(--md-cobalt) 42%, var(--md-line-strong));
    background: var(--md-surface);
    box-shadow:
      0 1px 0 color-mix(in oklab, #fff 80%, transparent) inset,
      0 14px 28px -16px color-mix(in oklab, var(--md-cobalt) 45%, transparent);
  }
  .plate-btn:active {
    transform: scale(0.97);
  }
  .plate-btn:focus-visible {
    outline: none;
    box-shadow:
      var(--md-focus),
      0 10px 28px -14px color-mix(in oklab, var(--md-cobalt) 55%, transparent);
  }
  .plate-btn-ico {
    display: inline-flex;
    color: var(--md-cobalt);
    opacity: 0.9;
  }
  .plate-btn-primary {
    border-color: transparent;
    background: var(--md-cobalt);
    color: #fff;
    box-shadow: 0 12px 28px -12px color-mix(in oklab, var(--md-cobalt) 72%, transparent);
  }
  .plate-btn-primary .plate-btn-ico {
    color: #fff;
    opacity: 1;
  }
  .plate-btn-primary:hover {
    background: var(--md-cobalt-deep);
    border-color: transparent;
    box-shadow: 0 16px 34px -12px color-mix(in oklab, var(--md-cobalt) 82%, transparent);
  }
  .plate-hint {
    margin: 0;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }

  @keyframes about-spin {
    to {
      transform: rotate(360deg);
    }
  }

  @media (max-width: 720px) {
    .colophon {
      padding: 24px 16px 120px;
    }
    .spine-stage {
      grid-template-columns: 56px 1fr;
      gap: 14px;
      min-height: 0;
    }
    .spine {
      min-height: 320px;
    }
    .node-roman {
      display: none;
    }
    .station {
      padding: 22px 18px 18px;
      min-height: 300px;
      border-radius: 22px;
    }
    .station h3 {
      max-width: none;
    }
    .orbit {
      width: 260px;
      height: 260px;
      right: -80px;
      top: 20px;
    }
  }

  @media (max-width: 480px) {
    .seal-row {
      flex-direction: column;
      align-items: stretch;
    }
    .seal {
      width: 100%;
    }
    .word {
      font-size: clamp(44px, 14vw, 64px);
    }
    .plate-btn {
      flex: 1;
      justify-content: center;
    }
  }

  .colophon.calm .wash,
  .colophon.calm .thesis,
  .colophon.calm .spine-section,
  .colophon.calm .readout,
  .colophon.calm .plate,
  .colophon.calm .orbit,
  .colophon.calm .glow,
  .colophon.calm .node-dot,
  .colophon.calm .station-inner {
    transition: none !important;
    animation: none !important;
  }
  .colophon.calm .orbit::after {
    animation: none !important;
  }
  .colophon.calm .wash,
  .colophon.calm .thesis,
  .colophon.calm .spine-section,
  .colophon.calm .readout,
  .colophon.calm .plate,
  .colophon.calm .orbit {
    opacity: 1;
    transform: none;
  }
</style>
