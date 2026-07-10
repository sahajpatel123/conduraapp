<script lang="ts">
  /**
   * Meridian About — living instrument colophon.
   * Signature: interactive Gatekeeper demo + constellation of seven stations.
   * Logic: capabilities · donate · station navigation.
   */
  import { onMount } from 'svelte'
  import { ipc } from '../../ipc/client'
  import { daemon } from '../../stores/daemon.svelte'
  import type { DaemonCapabilities } from '../../ipc/types'

  const DONATE_URL = 'https://condura.app/donate'
  const SITE_URL = 'https://condura.app'

  type Station = {
    id: string
    roman: string
    title: string
    body: string
    cite: string
    live?: () => { ok: boolean; note: string }
  }

  type GatePhase = 'idle' | 'sending' | 'held' | 'allowed' | 'denied'

  let caps = $state<DaemonCapabilities | null>(null)
  let loading = $state(true)
  let active = $state('i')
  let entered = $state(false)
  let reduceMotion = $state(false)
  let gatePhase = $state<GatePhase>('idle')
  let gateTimers: number[] = []

  const connected = $derived(daemon.connected)

  const STATIONS: Station[] = [
    {
      id: 'i',
      roman: 'I',
      title: 'Separate minds',
      body: 'Any model can plan. Only the Gatekeeper — fixed, deterministic code — can permit action. Planning and permission stay in different systems, so a model can never approve itself.',
      cite: 'gatekeeper',
    },
    {
      id: 'ii',
      roman: 'II',
      title: 'One door to action',
      body: 'Clicks, typing, and shell commands all pass through one door. Model text alone cannot reach your computer. If the Gatekeeper does not open, nothing happens.',
      cite: 'gatekeeper',
    },
    {
      id: 'iii',
      roman: 'III',
      title: 'A human at the keys',
      body: 'Destructive work waits for your Allow or Deny. There is no silent escalate path. Consent is the lock — and only you hold the key.',
      cite: 'consent',
      live: () => ({ ok: true, note: 'Consent sheets live in Meridian' }),
    },
    {
      id: 'iv',
      roman: 'IV',
      title: 'You can always stop',
      body: 'Four independent stops: a hard hotkey, Halt in the dock, a watchdog, and network isolation. Use any one of them to cut the line immediately.',
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
      body: 'Actions are written to an append-only audit ledger, sealed so the past cannot be quietly rewritten. If something goes wrong, you can still see what happened.',
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
      body: 'Condura asks before it enters a new capability. You grant or deny. It never raises its own privileges, and it never assumes it owns the room.',
      cite: 'permissions',
    },
    {
      id: 'vii',
      roman: 'VII',
      title: 'Your machine decides',
      body: 'Screen, accessibility, and input access are granted by your operating system. Condura can only ask — it cannot take those permissions on its own.',
      cite: 'os',
    },
  ]

  const activeStation = $derived(STATIONS.find((s) => s.id === active) ?? STATIONS[0]!)
  const activeIndex = $derived(STATIONS.findIndex((s) => s.id === active))

  function citeLabel(cite: string): string {
    if (cite === 'os') return 'OS'
    return cite.charAt(0).toUpperCase() + cite.slice(1)
  }

  const gateNote = $derived.by(() => {
    switch (gatePhase) {
      case 'sending':
        return 'Intent is moving toward the door…'
      case 'held':
        return 'Gatekeeper holds the line. Nothing proceeds without you.'
      case 'allowed':
        return 'Consent granted — the action may reach your machine.'
      case 'denied':
        return 'Denied. The model never touched your machine.'
      default:
        return 'Model text alone cannot act. Propose an action and feel the lock.'
    }
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
    return () => {
      window.removeEventListener('keydown', onKey)
      clearGateTimers()
    }
  })

  function clearGateTimers(): void {
    for (const t of gateTimers) clearTimeout(t)
    gateTimers = []
  }

  function proposeAction(): void {
    if (gatePhase === 'sending' || gatePhase === 'held') return
    clearGateTimers()
    gatePhase = 'sending'
    const delay = reduceMotion ? 80 : 920
    gateTimers.push(
      window.setTimeout(() => {
        gatePhase = 'held'
      }, delay),
    )
  }

  function decide(allow: boolean): void {
    if (gatePhase !== 'held') return
    clearGateTimers()
    gatePhase = allow ? 'allowed' : 'denied'
    const delay = reduceMotion ? 500 : 1700
    gateTimers.push(
      window.setTimeout(() => {
        gatePhase = 'idle'
      }, delay),
    )
  }

  function step(dir: number): void {
    const next = Math.max(0, Math.min(STATIONS.length - 1, activeIndex + dir))
    active = STATIONS[next]!.id
  }

  async function refresh(): Promise<void> {
    loading = true
    try {
      caps = await ipc.daemonCapabilities().catch(() => null)
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

  function led(on: boolean | undefined): 'on' | 'off' | 'dim' {
    if (on === true) return 'on'
    if (on === false) return 'off'
    return 'dim'
  }
</script>

<article class="colophon" class:in={entered} class:calm={reduceMotion}>
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

    <div class="gate" data-phase={gatePhase} aria-live="polite">
      <div class="gate-head">
        <p class="gate-k">Try the lock</p>
        <p class="gate-note">{gateNote}</p>
      </div>

      <div class="gate-flow" aria-hidden="true">
        <div class="gate-node" data-role="strategist">
          <span class="gate-orb">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
              <path d="M12 3v4M12 17v4M3 12h4M17 12h4" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/>
              <circle cx="12" cy="12" r="3.5" stroke="currentColor" stroke-width="1.7"/>
            </svg>
          </span>
          <span class="gate-role">Strategist</span>
          <span class="gate-sub">any model</span>
        </div>

        <div class="gate-path">
          <span class="gate-track"></span>
          <span class="gate-pulse"></span>
        </div>

        <div class="gate-node" data-role="keeper">
          <span class="gate-orb">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
              <rect x="5" y="11" width="14" height="10" rx="2" stroke="currentColor" stroke-width="1.7"/>
              <path d="M8 11V8a4 4 0 0 1 8 0v3" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/>
            </svg>
          </span>
          <span class="gate-role">Gatekeeper</span>
          <span class="gate-sub">deterministic code</span>
        </div>

        <div class="gate-path">
          <span class="gate-track"></span>
          <span class="gate-pulse late"></span>
        </div>

        <div class="gate-node" data-role="machine">
          <span class="gate-orb">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
              <rect x="3" y="5" width="18" height="12" rx="2" stroke="currentColor" stroke-width="1.7"/>
              <path d="M8 21h8M12 17v4" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/>
            </svg>
          </span>
          <span class="gate-role">Your machine</span>
          <span class="gate-sub">protected</span>
        </div>
      </div>

      <div class="gate-actions">
        {#if gatePhase === 'idle' || gatePhase === 'allowed' || gatePhase === 'denied'}
          <button type="button" class="gate-btn primary" onclick={proposeAction}>
            Propose an action
          </button>
        {:else if gatePhase === 'sending'}
          <button type="button" class="gate-btn" disabled>Traveling…</button>
        {:else}
          <button type="button" class="gate-btn danger" onclick={() => decide(false)}>Deny</button>
          <button type="button" class="gate-btn primary" onclick={() => decide(true)}>Allow</button>
        {/if}
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
              <span class="stage-cite">{citeLabel(activeStation.cite)}</span>
            {/if}
          </div>
          <h3>{activeStation.title}</h3>
          <p class="stage-body">{activeStation.body}</p>
          {#if activeStation.live}
            <p class="stage-cite-foot">{citeLabel(activeStation.cite)}</p>
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

  /* —— Interactive Gatekeeper —— */
  .gate {
    margin-top: 8px;
    padding: 22px 22px 20px;
    border-radius: 26px;
    border: 1px solid color-mix(in oklab, var(--md-ink) 9%, transparent);
    background: var(--md-surface);
    box-shadow: 0 1px 0 color-mix(in oklab, #fff 55%, transparent) inset;
    overflow: hidden;
  }
  .gate-head {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    justify-content: space-between;
    gap: 8px 20px;
    margin-bottom: 22px;
  }
  .gate-k {
    margin: 0;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .gate-note {
    margin: 0;
    flex: 1;
    min-width: 16ch;
    text-align: right;
    font-size: 14px;
    font-weight: 600;
    letter-spacing: -0.02em;
    color: var(--md-ink-soft);
    transition: color 220ms var(--about-ease);
  }
  .gate[data-phase='held'] .gate-note {
    color: var(--md-cobalt);
  }
  .gate[data-phase='allowed'] .gate-note {
    color: var(--md-live);
  }
  .gate[data-phase='denied'] .gate-note {
    color: var(--md-halt);
  }

  .gate-flow {
    display: grid;
    grid-template-columns: 1fr minmax(36px, 1.1fr) 1fr minmax(36px, 1.1fr) 1fr;
    align-items: center;
    gap: 6px;
    margin-bottom: 22px;
  }
  .gate-node {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    text-align: center;
    min-width: 0;
  }
  .gate-orb {
    width: 56px;
    height: 56px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    border: 1px solid var(--md-line-strong);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    color: var(--md-ink-mute);
    box-shadow: 0 10px 24px -18px color-mix(in oklab, var(--md-ink) 40%, transparent);
    transition:
      transform 320ms var(--about-spring),
      border-color 280ms var(--about-ease),
      background 280ms var(--about-ease),
      color 280ms var(--about-ease),
      box-shadow 280ms var(--about-ease);
  }
  .gate-role {
    font-family: var(--md-font-display);
    font-size: 13px;
    font-weight: 700;
    letter-spacing: -0.02em;
    color: var(--md-ink);
  }
  .gate-sub {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.06em;
    color: var(--md-ink-faint);
  }

  .gate-path {
    position: relative;
    height: 28px;
    display: grid;
    place-items: center;
  }
  .gate-track {
    display: block;
    width: 100%;
    height: 2px;
    border-radius: 999px;
    background: linear-gradient(
      90deg,
      color-mix(in oklab, var(--md-line-strong) 40%, transparent),
      var(--md-line-strong),
      color-mix(in oklab, var(--md-line-strong) 40%, transparent)
    );
  }
  .gate-pulse {
    position: absolute;
    left: 0;
    top: 50%;
    width: 10px;
    height: 10px;
    margin-top: -5px;
    margin-left: -5px;
    border-radius: 50%;
    background: var(--md-cobalt);
    box-shadow: 0 0 0 0 color-mix(in oklab, var(--md-cobalt) 40%, transparent);
    opacity: 0;
    transform: translateX(0) scale(0.6);
  }
  .gate-pulse.late {
    background: var(--md-live);
  }

  /* Phase: sending — pulse travels first path */
  .gate[data-phase='sending'] .gate-pulse:not(.late) {
    opacity: 1;
    animation: gate-travel 920ms var(--about-ease) forwards;
  }
  .gate[data-phase='sending'] .gate-node[data-role='strategist'] .gate-orb {
    color: var(--md-cobalt);
    border-color: color-mix(in oklab, var(--md-cobalt) 40%, transparent);
    box-shadow: 0 0 0 6px color-mix(in oklab, var(--md-cobalt) 12%, transparent);
  }

  /* Phase: held — pulse parked at gatekeeper */
  .gate[data-phase='held'] .gate-node[data-role='keeper'] .gate-orb {
    color: #fff;
    background: var(--md-cobalt);
    border-color: transparent;
    transform: scale(1.08);
    box-shadow:
      0 0 0 8px color-mix(in oklab, var(--md-cobalt) 16%, transparent),
      0 16px 32px -14px color-mix(in oklab, var(--md-cobalt) 55%, transparent);
    animation: gate-hold 1.4s ease-in-out infinite;
  }
  .gate[data-phase='held'] .gate-path:nth-child(2) .gate-track {
    background: linear-gradient(90deg, var(--md-cobalt), color-mix(in oklab, var(--md-cobalt) 35%, transparent));
  }

  /* Phase: allowed — second pulse travels to machine */
  .gate[data-phase='allowed'] .gate-pulse.late {
    opacity: 1;
    animation: gate-travel 700ms var(--about-ease) forwards;
  }
  .gate[data-phase='allowed'] .gate-node[data-role='keeper'] .gate-orb,
  .gate[data-phase='allowed'] .gate-node[data-role='machine'] .gate-orb {
    color: var(--md-live);
    border-color: color-mix(in oklab, var(--md-live) 40%, transparent);
    background: color-mix(in oklab, var(--md-live) 10%, var(--md-surface));
  }
  .gate[data-phase='allowed'] .gate-node[data-role='machine'] .gate-orb {
    transform: scale(1.06);
    box-shadow: 0 0 0 7px color-mix(in oklab, var(--md-live) 14%, transparent);
  }
  .gate[data-phase='allowed'] .gate-track {
    background: linear-gradient(90deg, var(--md-live), color-mix(in oklab, var(--md-live) 30%, transparent));
  }

  /* Phase: denied — gate flashes halt, machine stays quiet */
  .gate[data-phase='denied'] .gate-node[data-role='keeper'] .gate-orb {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 45%, transparent);
    background: color-mix(in oklab, var(--md-halt) 10%, var(--md-surface));
    animation: gate-deny 420ms var(--about-spring) both;
  }
  .gate[data-phase='denied'] .gate-path:nth-child(2) .gate-track {
    background: linear-gradient(90deg, color-mix(in oklab, var(--md-halt) 55%, transparent), var(--md-line));
  }

  .gate-actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 10px;
  }
  .gate-btn {
    appearance: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-height: 44px;
    padding: 0 20px;
    border-radius: 999px;
    border: 1px solid var(--md-line-strong);
    background: color-mix(in oklab, var(--md-surface) 90%, transparent);
    color: var(--md-ink);
    font-family: var(--md-font-sans);
    font-size: 13px;
    font-weight: 700;
    letter-spacing: -0.01em;
    cursor: pointer;
    transition:
      transform 180ms var(--about-spring),
      background 180ms var(--about-ease),
      border-color 180ms var(--about-ease),
      color 180ms var(--about-ease),
      box-shadow 180ms var(--about-ease);
  }
  .gate-btn:hover:not(:disabled) {
    transform: translateY(-1px);
    border-color: color-mix(in oklab, var(--md-cobalt) 40%, transparent);
  }
  .gate-btn:disabled {
    opacity: 0.55;
    cursor: wait;
  }
  .gate-btn:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .gate-btn.primary {
    background: var(--md-cobalt);
    border-color: transparent;
    color: #fff;
    box-shadow: 0 14px 28px -14px color-mix(in oklab, var(--md-cobalt) 70%, transparent);
  }
  .gate-btn.primary:hover:not(:disabled) {
    background: var(--md-cobalt-deep);
    border-color: transparent;
  }
  .gate-btn.danger {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 32%, transparent);
    background: color-mix(in oklab, var(--md-halt) 8%, var(--md-surface));
  }
  .gate-btn.danger:hover:not(:disabled) {
    background: color-mix(in oklab, var(--md-halt) 14%, var(--md-surface));
  }

  @keyframes gate-travel {
    0% {
      opacity: 0;
      transform: translateX(0) scale(0.5);
      box-shadow: 0 0 0 0 color-mix(in oklab, var(--md-cobalt) 35%, transparent);
    }
    12% {
      opacity: 1;
      transform: translateX(8%) scale(1);
    }
    70% {
      opacity: 1;
      transform: translateX(88%) scale(1);
      box-shadow: 0 0 0 8px color-mix(in oklab, var(--md-cobalt) 0%, transparent);
    }
    100% {
      opacity: 0;
      transform: translateX(100%) scale(0.7);
    }
  }
  @keyframes gate-hold {
    0%,
    100% {
      transform: scale(1.08);
    }
    50% {
      transform: scale(1.14);
    }
  }
  @keyframes gate-deny {
    0% {
      transform: scale(1);
    }
    35% {
      transform: scale(1.12) rotate(-3deg);
    }
    70% {
      transform: scale(0.96) rotate(2deg);
    }
    100% {
      transform: scale(1);
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
    background: var(--md-surface);
    overflow: hidden;
    display: flex;
    flex-direction: column;
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
    font-family: var(--md-font-sans);
    font-size: 12px;
    letter-spacing: 0.04em;
    text-transform: none;
    color: var(--md-cobalt);
    font-weight: 700;
  }
  .stage-cite,
  .stage-cite-foot {
    font-family: var(--md-font-sans);
    font-size: 12px;
    font-weight: 600;
    letter-spacing: -0.01em;
    text-transform: none;
    color: var(--md-ink-mute);
    padding: 5px 11px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-stage) 55%, var(--md-surface));
  }
  .stage-cite-foot {
    margin: 22px 0 0;
    width: fit-content;
  }
  .stage-live {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    font-family: var(--md-font-sans);
    font-size: 12px;
    font-weight: 600;
    letter-spacing: -0.01em;
    text-transform: none;
    padding: 5px 11px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    color: var(--md-ink-mute);
    background: color-mix(in oklab, var(--md-stage) 55%, var(--md-surface));
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
    font-size: clamp(30px, 5vw, 44px);
    font-weight: 700;
    letter-spacing: -0.045em;
    line-height: 1.08;
    margin: 0 0 14px;
    max-width: 16ch;
    color: var(--md-ink);
    text-wrap: balance;
  }
  .stage-body {
    margin: 0;
    font-family: var(--md-font-sans);
    font-size: 17px;
    font-weight: 450;
    line-height: 1.7;
    letter-spacing: -0.011em;
    text-transform: none;
    color: color-mix(in oklab, var(--md-ink) 72%, var(--md-ink-mute));
    max-width: 48ch;
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
    .gate {
      padding: 18px 16px 16px;
      border-radius: 22px;
    }
    .gate-note {
      text-align: left;
      width: 100%;
      font-size: 13px;
    }
    .gate-flow {
      gap: 2px;
    }
    .gate-orb {
      width: 46px;
      height: 46px;
    }
    .gate-role {
      font-size: 11px;
    }
    .gate-sub {
      display: none;
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
  }

  @media (max-width: 480px) {
    .gate-btn {
      flex: 1;
      min-width: 120px;
    }
    .word {
      font-size: clamp(44px, 14vw, 64px);
    }
    .plate-btn {
      flex: 1;
      justify-content: center;
    }
  }

  .colophon.calm .thesis,
  .colophon.calm .meridian,
  .colophon.calm .readout,
  .colophon.calm .plate,
  .colophon.calm .stage-copy,
  .colophon.calm .constellation-beam::after,
  .colophon.calm .gate-pulse,
  .colophon.calm .gate-orb {
    transition: none !important;
    animation: none !important;
  }
  .colophon.calm .thesis,
  .colophon.calm .meridian,
  .colophon.calm .readout,
  .colophon.calm .plate {
    opacity: 1;
    transform: none;
  }
</style>
