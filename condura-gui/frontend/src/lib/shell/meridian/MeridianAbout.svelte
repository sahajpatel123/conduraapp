<script lang="ts">
  /**
   * Meridian About — living instrument colophon.
   * Signature: a constellation of seven stations along the meridian.
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
    <h1 class="brand">
      <span class="word">Condura</span>
      <span class="slash" aria-hidden="true">/</span>
      <span class="meridian">Meridian</span>
    </h1>
    <p class="thesis-line">
      Free. Local. Consent before action.<br />
      <em>This page is the contract — and the live reading of the machine that keeps it.</em>
    </p>

    <div class="identity" data-link={connected ? 'live' : 'quiet'}>
      <button
        type="button"
        class="id-build"
        onclick={copyBuild}
        aria-label="Copy build identity"
        title="Copy build card to clipboard"
      >
        <span class="id-mark" aria-hidden="true">
          <svg viewBox="0 0 56 56" fill="none">
            <circle cx="28" cy="28" r="26" stroke="currentColor" stroke-opacity="0.18" stroke-width="1.25" />
            <circle
              cx="28"
              cy="28"
              r="26"
              stroke="currentColor"
              stroke-width="1.5"
              stroke-linecap="round"
              stroke-dasharray="48 120"
              transform="rotate(-90 28 28)"
              class="id-mark-arc"
            />
            <path
              d="M28 12c0 10 0 22 0 32M18 22c6 2 12 4 20 0M18 34c6-2 12-4 20 0"
              stroke="currentColor"
              stroke-width="1.6"
              stroke-linecap="round"
              opacity="0.9"
            />
          </svg>
        </span>
        <span class="id-body">
          <span class="id-k">Build identity</span>
          <span class="id-v">{buildLine}</span>
          <span class="id-sub">
            {#if version?.build_date}
              Built {version.build_date}
            {:else}
              GUI {GUI_VERSION}
            {/if}
            {#if version?.go_version}
              <span class="id-dot" aria-hidden="true">·</span>
              {version.go_version}
            {/if}
          </span>
        </span>
        <span class="id-chip" class:done={copied}>
          {#if copied}
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <path d="M5 13l4 4L19 7" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            Copied
          {:else}
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <rect x="8" y="8" width="11" height="13" rx="2" stroke="currentColor" stroke-width="1.8"/>
              <path d="M6 16H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
            </svg>
            Copy
          {/if}
        </span>
      </button>

      <div class="id-rail" aria-label="Daemon vitals">
        <div class="meter" data-tone={connected ? 'on' : 'off'}>
          <span class="meter-led" aria-hidden="true"></span>
          <div class="meter-text">
            <span class="meter-k">Link</span>
            <span class="meter-v">{connected ? 'Live' : 'Quiet'}</span>
          </div>
        </div>

        <div class="meter">
          <div class="meter-text">
            <span class="meter-k">Latency</span>
            <span class="meter-v mono">
              {#if loading && pingMs == null}
                …
              {:else if pingMs != null}
                {pingMs}<span class="meter-unit">ms</span>
              {:else}
                —
              {/if}
            </span>
          </div>
        </div>

        {#if version?.platform}
          <div class="meter">
            <div class="meter-text">
              <span class="meter-k">Host</span>
              <span class="meter-v host">{version.platform}</span>
            </div>
          </div>
        {/if}

        <button
          type="button"
          class="id-refresh"
          class:spin={loading}
          onclick={() => void refresh()}
          disabled={loading}
          aria-label={loading ? 'Reading vitals' : 'Refresh vitals'}
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path
              d="M20 12a8 8 0 1 1-2.2-5.4"
              stroke="currentColor"
              stroke-width="1.9"
              stroke-linecap="round"
            />
            <path d="M20 5v5h-5" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <span>{loading ? 'Reading' : 'Refresh'}</span>
        </button>
      </div>
    </div>
  </header>

  <!-- Signature: constellation of seven stations -->
  <section class="meridian" aria-label="Seven promises">
    <div class="meridian-head">
      <div class="meridian-titles">
        <p class="meridian-k">Seven promises</p>
        <h2>Along the meridian</h2>
      </div>
      <div class="meridian-progress" aria-hidden="true">
        <span class="meridian-frac">
          <em>{String(activeIndex + 1).padStart(2, '0')}</em>
          <span>/ {String(STATIONS.length).padStart(2, '0')}</span>
        </span>
        <span class="meridian-hint">↑↓ · J/K</span>
      </div>
    </div>

    <div class="constellation" role="tablist" aria-label="Stations">
      <div
        class="constellation-beam"
        style="--fill: {((activeIndex + 1) / STATIONS.length) * 100}%"
        aria-hidden="true"
      ></div>
      {#each STATIONS as s, i (s.id)}
        <button
          type="button"
          role="tab"
          class="star"
          class:on={active === s.id}
          class:passed={i < activeIndex}
          aria-selected={active === s.id}
          aria-controls="station-panel"
          id={`station-tab-${s.id}`}
          aria-label={`Station ${s.roman}: ${s.title}`}
          onclick={() => (active = s.id)}
        >
          <span class="star-core">
            <span class="star-roman">{s.roman}</span>
          </span>
          <span class="star-label">{s.title}</span>
        </button>
      {/each}
    </div>

    <div
      class="stage-plate"
      id="station-panel"
      role="tabpanel"
      aria-labelledby={`station-tab-${active}`}
    >
      <div class="stage-wash" aria-hidden="true"></div>
      <p class="stage-watermark" aria-hidden="true">{activeStation.roman}</p>

      {#key active}
        <div class="stage-copy">
          <div class="stage-top">
            <span class="stage-index">Station {activeStation.roman}</span>
            {#if activeStation.live}
              {@const live = activeStation.live()}
              <span class="stage-live" data-ok={live.ok}>
                <i></i>
                {live.note}
              </span>
            {:else}
              <span class="stage-cite">{activeStation.cite}</span>
            {/if}
          </div>
          <h3>{activeStation.title}</h3>
          <p class="stage-body">{activeStation.body}</p>
          {#if activeStation.live}
            <p class="stage-cite-foot">{activeStation.cite}</p>
          {/if}
        </div>
      {/key}

      <div class="stage-controls">
        <button
          type="button"
          class="stage-nav"
          disabled={activeIndex <= 0}
          onclick={() => step(-1)}
          aria-label="Previous station"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M15 6 9 12l6 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>

        <div class="stage-ticks" aria-hidden="true">
          {#each STATIONS as s, i (s.id)}
            <span class="tick" class:on={i === activeIndex} class:passed={i < activeIndex}></span>
          {/each}
        </div>

        <button
          type="button"
          class="stage-nav primary"
          disabled={activeIndex >= STATIONS.length - 1}
          onclick={() => step(1)}
          aria-label="Next station"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="m9 6 6 6-6 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
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
  .meridian,
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

  .identity {
    display: grid;
    grid-template-columns: minmax(0, 1.35fr) minmax(220px, 0.9fr);
    gap: 0;
    align-items: stretch;
    border-radius: 22px;
    border: 1px solid color-mix(in oklab, var(--md-ink) 10%, transparent);
    background:
      linear-gradient(
        135deg,
        color-mix(in oklab, var(--md-surface) 92%, transparent) 0%,
        color-mix(in oklab, var(--md-surface) 72%, transparent) 100%
      );
    box-shadow:
      0 1px 0 color-mix(in oklab, #fff 65%, transparent) inset,
      0 18px 40px -28px color-mix(in oklab, var(--md-ink) 28%, transparent);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    overflow: hidden;
  }
  .identity[data-link='live'] {
    border-color: color-mix(in oklab, var(--md-live) 22%, var(--md-line));
  }
  .identity[data-link='quiet'] {
    border-color: color-mix(in oklab, var(--md-halt) 16%, var(--md-line));
  }

  .id-build {
    appearance: none;
    display: grid;
    grid-template-columns: 52px minmax(0, 1fr) auto;
    align-items: center;
    gap: 16px;
    width: 100%;
    padding: 18px 20px;
    text-align: left;
    cursor: pointer;
    color: inherit;
    background:
      radial-gradient(
        120% 90% at 0% 0%,
        color-mix(in oklab, var(--md-cobalt) 8%, transparent),
        transparent 55%
      );
    border: 0;
    border-right: 1px solid color-mix(in oklab, var(--md-ink) 8%, transparent);
    transition:
      background 220ms var(--about-ease),
      transform 200ms var(--about-spring);
  }
  .id-build:hover {
    background:
      radial-gradient(
        120% 90% at 0% 0%,
        color-mix(in oklab, var(--md-cobalt) 14%, transparent),
        transparent 58%
      );
  }
  .id-build:focus-visible {
    outline: none;
    box-shadow: inset 0 0 0 2px color-mix(in oklab, var(--md-cobalt) 55%, transparent);
  }
  .id-build:active {
    transform: scale(0.995);
  }

  .id-mark {
    width: 52px;
    height: 52px;
    color: var(--md-cobalt);
    display: grid;
    place-items: center;
  }
  .id-mark svg {
    width: 52px;
    height: 52px;
  }
  .id-mark-arc {
    transition: stroke-dasharray 600ms var(--about-ease);
  }
  .identity[data-link='live'] .id-mark-arc {
    stroke: var(--md-live);
    stroke-dasharray: 100 120;
  }
  .identity[data-link='quiet'] .id-mark-arc {
    stroke: var(--md-halt);
    stroke-dasharray: 28 120;
  }

  .id-body {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }
  .id-k {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .id-v {
    font-family: var(--md-font-display);
    font-size: 18px;
    font-weight: 650;
    letter-spacing: -0.03em;
    color: var(--md-ink);
    line-height: 1.15;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .id-sub {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.02em;
    color: var(--md-ink-mute);
  }
  .id-dot {
    opacity: 0.5;
  }

  .id-chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    flex: none;
    min-height: 32px;
    padding: 0 12px;
    border-radius: 999px;
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 28%, transparent);
    background: color-mix(in oklab, var(--md-cobalt) 8%, var(--md-surface));
    color: var(--md-cobalt);
    font-family: var(--md-font-sans);
    font-size: 12px;
    font-weight: 700;
    letter-spacing: -0.01em;
    transition:
      background 180ms var(--about-ease),
      border-color 180ms var(--about-ease),
      color 180ms var(--about-ease),
      transform 180ms var(--about-spring);
  }
  .id-build:hover .id-chip {
    background: var(--md-cobalt);
    border-color: transparent;
    color: #fff;
    transform: translateY(-1px);
  }
  .id-chip.done {
    background: color-mix(in oklab, var(--md-live) 12%, var(--md-surface));
    border-color: color-mix(in oklab, var(--md-live) 32%, transparent);
    color: var(--md-live);
  }

  .id-rail {
    display: flex;
    flex-direction: column;
    gap: 1px;
    background: color-mix(in oklab, var(--md-ink) 6%, transparent);
    min-width: 0;
  }
  .meter {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 16px;
    background: color-mix(in oklab, var(--md-surface) 78%, transparent);
    min-width: 0;
  }
  .meter-led {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex: none;
    background: var(--md-ink-faint);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--md-ink-faint) 16%, transparent);
  }
  .meter[data-tone='on'] .meter-led {
    background: var(--md-live);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--md-live) 18%, transparent);
    animation: about-led 1.8s ease-in-out infinite;
  }
  .meter[data-tone='off'] .meter-led {
    background: var(--md-halt);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--md-halt) 16%, transparent);
  }
  .meter-text {
    display: flex;
    flex-direction: row;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
    flex: 1;
    min-width: 0;
  }
  .meter-k {
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    flex: none;
  }
  .meter-v {
    font-family: var(--md-font-sans);
    font-size: 13px;
    font-weight: 700;
    letter-spacing: -0.02em;
    color: var(--md-ink);
    line-height: 1.2;
    text-align: right;
  }
  .meter-v.mono {
    font-family: var(--md-font-mono);
    font-weight: 500;
    letter-spacing: 0;
    font-variant-numeric: tabular-nums;
  }
  .meter-v.host {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-weight: 600;
    font-size: 12px;
    color: var(--md-ink-soft);
  }
  .meter[data-tone='on'] .meter-v {
    color: var(--md-live);
  }
  .meter[data-tone='off'] .meter-v {
    color: var(--md-halt);
  }
  .meter-unit {
    margin-left: 2px;
    font-size: 10px;
    color: var(--md-ink-faint);
  }

  .id-refresh {
    appearance: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 42px;
    padding: 0 16px;
    border: 0;
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    color: var(--md-ink-soft);
    font-family: var(--md-font-sans);
    font-size: 12px;
    font-weight: 700;
    letter-spacing: -0.01em;
    cursor: pointer;
    transition:
      background 180ms var(--about-ease),
      color 180ms var(--about-ease);
  }
  .id-refresh:hover:not(:disabled) {
    background: color-mix(in oklab, var(--md-cobalt) 8%, var(--md-surface));
    color: var(--md-cobalt);
  }
  .id-refresh:focus-visible {
    outline: none;
    box-shadow: inset 0 0 0 2px color-mix(in oklab, var(--md-cobalt) 45%, transparent);
  }
  .id-refresh:disabled {
    cursor: wait;
    opacity: 0.7;
  }
  .id-refresh.spin svg {
    animation: about-spin 0.9s linear infinite;
  }

  @keyframes about-led {
    0%,
    100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(1.15);
      opacity: 0.75;
    }
  }

  /* —— Meridian constellation —— */
  .meridian {
    margin-bottom: 52px;
    opacity: 0;
    transform: translateY(22px);
    transition:
      opacity 800ms var(--about-ease) 140ms,
      transform 800ms var(--about-ease) 140ms;
  }
  .colophon.in .meridian {
    opacity: 1;
    transform: none;
  }

  .meridian-head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 20px;
    margin-bottom: 28px;
  }
  .meridian-k {
    margin: 0 0 8px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .meridian-titles h2 {
    font-family: var(--md-font-display);
    font-size: clamp(28px, 4.5vw, 40px);
    letter-spacing: -0.05em;
    margin: 0;
    line-height: 1;
  }
  .meridian-progress {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 6px;
    flex: none;
  }
  .meridian-frac {
    font-family: var(--md-font-mono);
    font-size: 13px;
    letter-spacing: 0.04em;
    color: var(--md-ink-faint);
    font-variant-numeric: tabular-nums;
  }
  .meridian-frac em {
    font-style: normal;
    font-size: 22px;
    font-weight: 600;
    letter-spacing: -0.04em;
    color: var(--md-ink);
    margin-right: 4px;
  }
  .meridian-hint {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    padding: 4px 8px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 55%, transparent);
  }

  .constellation {
    position: relative;
    display: grid;
    grid-template-columns: repeat(7, minmax(0, 1fr));
    gap: 8px;
    margin-bottom: 18px;
    padding: 4px 0 8px;
  }
  .constellation-beam {
    position: absolute;
    left: 6%;
    right: 6%;
    top: 22px;
    height: 2px;
    border-radius: 999px;
    background: var(--md-line-strong);
    z-index: 0;
    overflow: hidden;
  }
  .constellation-beam::after {
    content: '';
    position: absolute;
    inset: 0 auto 0 0;
    width: var(--fill, 14%);
    border-radius: inherit;
    background: linear-gradient(90deg, var(--md-live), var(--md-cobalt));
    box-shadow: 0 0 16px color-mix(in oklab, var(--md-cobalt) 45%, transparent);
    transition: width 480ms var(--about-ease);
  }

  .star {
    appearance: none;
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 0;
    border: 0;
    background: transparent;
    color: inherit;
    cursor: pointer;
    min-width: 0;
  }
  .star-core {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    border: 1px solid var(--md-line-strong);
    box-shadow: 0 1px 0 color-mix(in oklab, #fff 55%, transparent) inset;
    transition:
      transform 220ms var(--about-spring),
      background 220ms var(--about-ease),
      border-color 220ms var(--about-ease),
      box-shadow 220ms var(--about-ease),
      color 220ms var(--about-ease);
  }
  .star-roman {
    font-family: var(--md-font-mono);
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.06em;
    color: var(--md-ink-mute);
    transition: color 220ms var(--about-ease);
  }
  .star-label {
    font-family: var(--md-font-sans);
    font-size: 11px;
    font-weight: 600;
    letter-spacing: -0.01em;
    color: var(--md-ink-faint);
    text-align: center;
    line-height: 1.25;
    max-width: 9ch;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    transition: color 220ms var(--about-ease);
  }
  .star.passed .star-core {
    border-color: color-mix(in oklab, var(--md-cobalt) 35%, transparent);
    background: color-mix(in oklab, var(--md-cobalt) 8%, var(--md-surface));
  }
  .star.passed .star-roman {
    color: var(--md-cobalt);
  }
  .star.on .star-core {
    background: var(--md-cobalt);
    border-color: transparent;
    color: #fff;
    transform: scale(1.08);
    box-shadow:
      0 0 0 6px color-mix(in oklab, var(--md-cobalt) 16%, transparent),
      0 12px 28px -12px color-mix(in oklab, var(--md-cobalt) 65%, transparent);
  }
  .star.on .star-roman {
    color: #fff;
  }
  .star.on .star-label {
    color: var(--md-ink);
  }
  .star:hover .star-core {
    transform: translateY(-2px) scale(1.04);
    border-color: color-mix(in oklab, var(--md-cobalt) 45%, transparent);
  }
  .star.on:hover .star-core {
    transform: scale(1.1);
  }
  .star:focus-visible .star-core {
    outline: none;
    box-shadow: var(--md-focus), 0 12px 28px -12px color-mix(in oklab, var(--md-cobalt) 45%, transparent);
  }

  .stage-plate {
    position: relative;
    isolation: isolate;
    min-height: 340px;
    padding: 36px 36px 22px;
    border-radius: 28px;
    border: 1px solid color-mix(in oklab, var(--md-ink) 9%, transparent);
    background: color-mix(in oklab, var(--md-surface) 55%, transparent);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    box-shadow:
      0 1px 0 color-mix(in oklab, #fff 50%, transparent) inset,
      0 28px 60px -36px color-mix(in oklab, var(--md-cobalt) 28%, transparent);
  }
  .stage-wash {
    position: absolute;
    inset: -20% -10% auto -10%;
    height: 70%;
    background:
      radial-gradient(
        ellipse 55% 70% at 8% 20%,
        color-mix(in oklab, var(--md-cobalt) 16%, transparent),
        transparent 70%
      ),
      radial-gradient(
        ellipse 45% 55% at 92% 10%,
        color-mix(in oklab, var(--md-live) 10%, transparent),
        transparent 68%
      );
    filter: blur(28px);
    pointer-events: none;
    z-index: 0;
  }
  .stage-watermark {
    position: absolute;
    right: -2%;
    top: -8%;
    margin: 0;
    font-family: var(--md-font-display);
    font-size: clamp(140px, 28vw, 240px);
    font-weight: 700;
    letter-spacing: -0.08em;
    line-height: 0.8;
    color: var(--md-cobalt);
    opacity: 0.06;
    pointer-events: none;
    z-index: 0;
    user-select: none;
    transition: opacity 400ms var(--about-ease);
  }

  .stage-copy {
    position: relative;
    z-index: 1;
    flex: 1;
    display: flex;
    flex-direction: column;
    max-width: 34rem;
    animation: about-station 480ms var(--about-ease) both;
  }
  @keyframes about-station {
    from {
      opacity: 0;
      transform: translateY(14px);
      filter: blur(4px);
    }
    to {
      opacity: 1;
      transform: none;
      filter: blur(0);
    }
  }
  .stage-top {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px;
    margin-bottom: 18px;
  }
  .stage-index {
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-cobalt);
    font-weight: 600;
  }
  .stage-cite,
  .stage-cite-foot {
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    color: var(--md-ink-faint);
    padding: 5px 10px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
  }
  .stage-cite-foot {
    margin: 20px 0 0;
    width: fit-content;
  }
  .stage-live {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 5px 10px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    color: var(--md-ink-faint);
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
  }
  .stage-live i {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
  }
  .stage-live[data-ok='true'] {
    color: var(--md-live);
    border-color: color-mix(in oklab, var(--md-live) 30%, transparent);
    background: color-mix(in oklab, var(--md-live) 8%, transparent);
  }
  .stage-live[data-ok='false'] {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 26%, transparent);
  }
  .stage-copy h3 {
    font-family: var(--md-font-display);
    font-size: clamp(34px, 6vw, 56px);
    letter-spacing: -0.055em;
    line-height: 0.98;
    margin: 0 0 18px;
    max-width: 12ch;
  }
  .stage-body {
    margin: 0;
    font-size: 17px;
    line-height: 1.65;
    color: var(--md-ink-soft);
    max-width: 40ch;
  }

  .stage-controls {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-top: 28px;
    padding-top: 18px;
    border-top: 1px solid color-mix(in oklab, var(--md-ink) 7%, transparent);
  }
  .stage-nav {
    appearance: none;
    width: 44px;
    height: 44px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    border: 1px solid var(--md-line-strong);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    color: var(--md-ink);
    cursor: pointer;
    box-shadow: 0 8px 22px -16px color-mix(in oklab, var(--md-ink) 40%, transparent);
    transition:
      transform 180ms var(--about-spring),
      background 180ms var(--about-ease),
      border-color 180ms var(--about-ease),
      color 180ms var(--about-ease),
      box-shadow 180ms var(--about-ease);
  }
  .stage-nav:hover:not(:disabled) {
    transform: translateY(-1px);
    border-color: color-mix(in oklab, var(--md-cobalt) 40%, transparent);
    color: var(--md-cobalt);
  }
  .stage-nav.primary {
    background: var(--md-cobalt);
    border-color: transparent;
    color: #fff;
    box-shadow: 0 14px 28px -14px color-mix(in oklab, var(--md-cobalt) 70%, transparent);
  }
  .stage-nav.primary:hover:not(:disabled) {
    background: var(--md-cobalt-deep);
    color: #fff;
    transform: translateY(-1px) scale(1.03);
  }
  .stage-nav:disabled {
    opacity: 0.28;
    cursor: not-allowed;
    box-shadow: none;
  }
  .stage-nav:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .stage-ticks {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 1;
    justify-content: center;
    max-width: 220px;
    margin: 0 auto;
  }
  .tick {
    height: 4px;
    flex: 1;
    border-radius: 999px;
    background: var(--md-line-strong);
    transition:
      background 220ms var(--about-ease),
      transform 220ms var(--about-spring);
  }
  .tick.passed {
    background: color-mix(in oklab, var(--md-cobalt) 45%, var(--md-line-strong));
  }
  .tick.on {
    background: var(--md-cobalt);
    transform: scaleY(1.35);
    box-shadow: 0 0 12px color-mix(in oklab, var(--md-cobalt) 45%, transparent);
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

  @media (max-width: 900px) {
    .star-label {
      opacity: 0;
      height: 0;
      overflow: hidden;
      margin: 0;
    }
    .star.on .star-label {
      opacity: 1;
      height: auto;
      max-width: 11ch;
    }
    .constellation {
      gap: 4px;
    }
  }

  @media (max-width: 720px) {
    .colophon {
      padding: 24px 16px 120px;
    }
    .identity {
      grid-template-columns: 1fr;
    }
    .id-build {
      border-right: 0;
      border-bottom: 1px solid color-mix(in oklab, var(--md-ink) 8%, transparent);
    }
    .id-rail {
      display: grid;
      grid-template-columns: 1fr 1fr;
    }
    .id-refresh {
      grid-column: 1 / -1;
    }
    .meridian-head {
      align-items: flex-start;
      flex-direction: column;
      gap: 14px;
    }
    .meridian-progress {
      align-items: flex-start;
      flex-direction: row;
      gap: 12px;
    }
    .constellation {
      display: flex;
      gap: 4px;
      overflow-x: auto;
      padding-bottom: 12px;
      scroll-snap-type: x proximity;
      -webkit-overflow-scrolling: touch;
      scrollbar-width: none;
    }
    .constellation::-webkit-scrollbar {
      display: none;
    }
    .constellation-beam {
      display: none;
    }
    .star {
      flex: 0 0 auto;
      width: 72px;
      scroll-snap-align: center;
    }
    .star-label {
      max-width: 7ch;
      font-size: 10px;
    }
    .stage-plate {
      padding: 26px 20px 18px;
      min-height: 300px;
      border-radius: 22px;
    }
    .stage-copy h3 {
      max-width: none;
      font-size: clamp(30px, 9vw, 42px);
    }
    .stage-watermark {
      font-size: 160px;
      right: -8%;
      top: -4%;
    }
    .orbit {
      width: 260px;
      height: 260px;
      right: -80px;
      top: 20px;
    }
  }

  @media (max-width: 480px) {
    .identity {
      border-radius: 18px;
    }
    .id-build {
      grid-template-columns: 44px minmax(0, 1fr);
      gap: 12px;
      padding: 16px;
    }
    .id-mark {
      width: 44px;
      height: 44px;
    }
    .id-mark svg {
      width: 44px;
      height: 44px;
    }
    .id-chip {
      grid-column: 1 / -1;
      justify-content: center;
      width: 100%;
      min-height: 36px;
    }
    .id-v {
      font-size: 16px;
      white-space: normal;
    }
    .id-rail {
      grid-template-columns: 1fr;
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
  .colophon.calm .meridian,
  .colophon.calm .readout,
  .colophon.calm .plate,
  .colophon.calm .orbit,
  .colophon.calm .stage-copy,
  .colophon.calm .constellation-beam::after,
  .colophon.calm .meter-led,
  .colophon.calm .id-refresh.spin svg {
    transition: none !important;
    animation: none !important;
  }
  .colophon.calm .orbit::after {
    animation: none !important;
  }
  .colophon.calm .wash,
  .colophon.calm .thesis,
  .colophon.calm .meridian,
  .colophon.calm .readout,
  .colophon.calm .plate,
  .colophon.calm .orbit {
    opacity: 1;
    transform: none;
  }
</style>
