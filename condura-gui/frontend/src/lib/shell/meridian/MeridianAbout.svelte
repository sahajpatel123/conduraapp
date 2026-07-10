<script lang="ts">
  /**
   * Meridian About — living instrument colophon.
   * Signature: atlas of ways in + constellation of seven stations.
   * Logic: capabilities · donate · station navigation · deep links.
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

  type AtlasId = 'audit' | 'promises' | 'independence'

  let caps = $state<DaemonCapabilities | null>(null)
  let active = $state('i')
  let entered = $state(false)
  let reduceMotion = $state(false)
  let atlasFocus = $state<AtlasId | null>(null)
  let modLabel = $state('⌘')

  const connected = $derived(daemon.connected)

  const ATLAS: {
    id: AtlasId
    kicker: string
    title: string
    body: string
    action: string
    run: () => void
  }[] = [
    {
      id: 'audit',
      kicker: 'Ledger',
      title: 'Open the audit',
      body: 'Every gated action is written to an append-only chain. Read what Condura did — and what it refused.',
      action: 'Go to Audit',
      run: () => {
        window.location.hash = '#/audit'
      },
    },
    {
      id: 'promises',
      kicker: 'Contract',
      title: 'Walk the meridian',
      body: 'Seven promises, in order. Use the constellation below, or press ↑↓ / J K to move station by station.',
      action: 'Begin at station one',
      run: () => goPromises(),
    },
    {
      id: 'independence',
      kicker: 'Freedom',
      title: 'Keep it independent',
      body: 'Condura is free software. If it earns your trust, help keep the work unbound from ads and lock-in.',
      action: 'Donate',
      run: () => openDonate(),
    },
  ]

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

  onMount(() => {
    reduceMotion = matchMedia('(prefers-reduced-motion: reduce)').matches
    modLabel =
      /Mac|iPhone|iPod|iPad/i.test(navigator.platform) || navigator.userAgent.includes('Mac')
        ? '⌘'
        : 'Ctrl'
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

  function goPromises(): void {
    active = 'i'
    const el = document.getElementById('meridian-section')
    el?.scrollIntoView({ behavior: reduceMotion ? 'auto' : 'smooth', block: 'start' })
  }

  async function refresh(): Promise<void> {
    caps = await ipc.daemonCapabilities().catch(() => null)
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

    <div class="atlas">
      <div class="atlas-head">
        <p class="atlas-k">Ways in</p>
        <h2 class="atlas-title">Where Condura keeps its word</h2>
        <p class="atlas-note">
          Three doors that matter. Each one takes you somewhere real — the ledger, the contract, or the work that keeps Condura free.
        </p>
      </div>

      <div class="atlas-grid">
        {#each ATLAS as door (door.id)}
          <button
            type="button"
            class="atlas-door"
            class:focus={atlasFocus === door.id}
            onmouseenter={() => (atlasFocus = door.id)}
            onfocus={() => (atlasFocus = door.id)}
            onmouseleave={() => (atlasFocus = null)}
            onblur={() => (atlasFocus = null)}
            onclick={door.run}
          >
            <span class="atlas-door-k">{door.kicker}</span>
            <span class="atlas-door-t">{door.title}</span>
            <span class="atlas-door-b">{door.body}</span>
            <span class="atlas-door-a">
              {door.action}
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                <path d="M5 12h14M13 6l6 6-6 6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </span>
          </button>
        {/each}
      </div>

      <ul class="atlas-keys" aria-label="Useful shortcuts">
        <li>
          <kbd>{modLabel}K</kbd>
          <span>Jump anywhere</span>
        </li>
        <li>
          <kbd>↑↓</kbd>
          <span>Walk promises</span>
        </li>
        <li>
          <kbd>{modLabel}D</kbd>
          <span>Donate</span>
        </li>
      </ul>
    </div>
  </header>

  <!-- Signature: constellation of seven stations -->
  <section id="meridian-section" class="meridian" aria-label="Seven promises">
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

  <!-- Closing colophon -->
  <footer class="close">
    <div class="close-mark" aria-hidden="true"></div>
    <p class="close-line">If Condura earns your trust, help keep it independent.</p>
    <nav class="close-index" aria-label="Continue from About">
      <button type="button" class="close-row" onclick={goAudit}>
        <span class="close-name">Audit</span>
        <span class="close-lead" aria-hidden="true"></span>
        <span class="close-meta">the ledger</span>
      </button>
      <button type="button" class="close-row" onclick={openSite}>
        <span class="close-name">Site</span>
        <span class="close-lead" aria-hidden="true"></span>
        <span class="close-meta">condura.app</span>
      </button>
      <button type="button" class="close-row primary" onclick={openDonate}>
        <span class="close-name">Donate</span>
        <span class="close-lead" aria-hidden="true"></span>
        <span class="close-meta">keep it free · {modLabel}D</span>
      </button>
    </nav>
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
  .close {
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

  /* —— Atlas: ways in —— */
  .atlas {
    margin-top: 28px;
  }
  .atlas-head {
    margin-bottom: 22px;
    max-width: 40rem;
  }
  .atlas-k {
    margin: 0 0 10px;
    font-family: var(--md-font-sans);
    font-size: 12px;
    font-weight: 650;
    letter-spacing: 0.02em;
    color: var(--md-ink-faint);
  }
  .atlas-title {
    margin: 0 0 10px;
    font-family: var(--md-font-display);
    font-size: clamp(26px, 4vw, 36px);
    font-weight: 700;
    letter-spacing: -0.045em;
    line-height: 1.08;
    color: var(--md-ink);
    text-wrap: balance;
  }
  .atlas-note {
    margin: 0;
    font-family: var(--md-font-sans);
    font-size: 15px;
    font-weight: 450;
    line-height: 1.65;
    letter-spacing: -0.011em;
    color: color-mix(in oklab, var(--md-ink) 68%, var(--md-ink-mute));
    max-width: 46ch;
  }

  .atlas-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
  }
  .atlas-door {
    appearance: none;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    min-height: 220px;
    padding: 20px 18px 18px;
    text-align: left;
    border-radius: 22px;
    border: 1px solid color-mix(in oklab, var(--md-ink) 9%, transparent);
    background: var(--md-surface);
    color: inherit;
    cursor: pointer;
    transition:
      transform 220ms var(--about-spring),
      border-color 220ms var(--about-ease),
      box-shadow 220ms var(--about-ease),
      background 220ms var(--about-ease);
  }
  .atlas-door:hover,
  .atlas-door.focus {
    transform: translateY(-2px);
    border-color: color-mix(in oklab, var(--md-cobalt) 32%, transparent);
    box-shadow: 0 18px 36px -28px color-mix(in oklab, var(--md-cobalt) 45%, transparent);
  }
  .atlas-door:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .atlas-door-k {
    font-family: var(--md-font-sans);
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
    color: var(--md-cobalt);
  }
  .atlas-door-t {
    font-family: var(--md-font-display);
    font-size: 20px;
    font-weight: 700;
    letter-spacing: -0.035em;
    line-height: 1.15;
    color: var(--md-ink);
  }
  .atlas-door-b {
    flex: 1;
    font-family: var(--md-font-sans);
    font-size: 14px;
    font-weight: 450;
    line-height: 1.55;
    letter-spacing: -0.01em;
    color: color-mix(in oklab, var(--md-ink) 70%, var(--md-ink-mute));
  }
  .atlas-door-a {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin-top: 6px;
    font-family: var(--md-font-sans);
    font-size: 13px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: var(--md-cobalt);
  }
  .atlas-door:hover .atlas-door-a,
  .atlas-door.focus .atlas-door-a {
    gap: 9px;
  }

  .atlas-keys {
    display: flex;
    flex-wrap: wrap;
    gap: 10px 18px;
    list-style: none;
    margin: 18px 0 0;
    padding: 0;
  }
  .atlas-keys li {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: var(--md-font-sans);
    font-size: 13px;
    font-weight: 550;
    color: var(--md-ink-mute);
  }
  .atlas-keys kbd {
    font-family: var(--md-font-mono);
    font-size: 11px;
    font-weight: 500;
    letter-spacing: 0.02em;
    padding: 4px 8px;
    border-radius: 8px;
    border: 1px solid color-mix(in oklab, var(--md-ink) 10%, transparent);
    background: color-mix(in oklab, var(--md-surface) 80%, transparent);
    color: var(--md-ink-soft);
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
    height: 380px;
    min-height: 380px;
    max-height: 380px;
    padding: 28px 36px 22px;
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
    top: 50%;
    transform: translateY(-50%);
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
    min-height: 0;
    width: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    text-align: center;
    animation: about-station 480ms var(--about-ease) both;
  }
  @keyframes about-station {
    from {
      opacity: 0;
      transform: translateY(10px);
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
    justify-content: center;
    gap: 10px;
    margin-bottom: 16px;
  }
  .stage-top:empty {
    display: none;
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
    margin: 18px auto 0;
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
    font-size: clamp(28px, 4.5vw, 40px);
    font-weight: 700;
    letter-spacing: -0.045em;
    line-height: 1.1;
    margin: 0 0 12px;
    max-width: 18ch;
    color: var(--md-ink);
    text-wrap: balance;
  }
  .stage-body {
    margin: 0;
    font-family: var(--md-font-sans);
    font-size: 16px;
    font-weight: 450;
    line-height: 1.65;
    letter-spacing: -0.011em;
    text-transform: none;
    color: color-mix(in oklab, var(--md-ink) 72%, var(--md-ink-mute));
    max-width: 44ch;
  }

  .stage-controls {
    position: relative;
    z-index: 1;
    flex: none;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-top: auto;
    padding-top: 16px;
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

  /* —— Closing colophon —— */
  .close {
    margin-top: 8px;
    padding-top: 28px;
    opacity: 0;
    transition: opacity 700ms var(--about-ease) 280ms;
  }
  .colophon.in .close {
    opacity: 1;
  }
  .close-mark {
    width: 28px;
    height: 2px;
    border-radius: 999px;
    background: var(--md-cobalt);
    margin-bottom: 18px;
    opacity: 0.85;
  }
  .close-line {
    margin: 0 0 28px;
    font-family: var(--md-font-display);
    font-size: clamp(24px, 3.6vw, 32px);
    font-weight: 700;
    letter-spacing: -0.04em;
    line-height: 1.18;
    max-width: 18ch;
    color: var(--md-ink);
    text-wrap: balance;
  }
  .close-index {
    display: flex;
    flex-direction: column;
    gap: 0;
    max-width: 34rem;
  }
  .close-row {
    appearance: none;
    display: grid;
    grid-template-columns: auto minmax(24px, 1fr) auto;
    align-items: baseline;
    gap: 12px;
    width: 100%;
    padding: 14px 0;
    border: 0;
    border-bottom: 1px solid color-mix(in oklab, var(--md-ink) 8%, transparent);
    background: transparent;
    color: inherit;
    cursor: pointer;
    text-align: left;
    transition: color 180ms var(--about-ease), padding 180ms var(--about-ease);
  }
  .close-row:first-child {
    border-top: 1px solid color-mix(in oklab, var(--md-ink) 8%, transparent);
  }
  .close-row:hover,
  .close-row:focus-visible {
    color: var(--md-cobalt);
  }
  .close-row:focus-visible {
    outline: none;
    box-shadow: inset 0 -2px 0 var(--md-cobalt);
  }
  .close-name {
    font-family: var(--md-font-display);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: -0.025em;
    color: var(--md-ink);
    transition: color 180ms var(--about-ease);
  }
  .close-row:hover .close-name,
  .close-row:focus-visible .close-name {
    color: var(--md-cobalt);
  }
  .close-lead {
    height: 1px;
    align-self: center;
    border-bottom: 1px dotted color-mix(in oklab, var(--md-ink) 22%, transparent);
    transform: translateY(-2px);
    transition: border-color 180ms var(--about-ease);
  }
  .close-row:hover .close-lead,
  .close-row:focus-visible .close-lead {
    border-bottom-color: color-mix(in oklab, var(--md-cobalt) 45%, transparent);
  }
  .close-meta {
    font-family: var(--md-font-sans);
    font-size: 13px;
    font-weight: 500;
    letter-spacing: -0.01em;
    color: var(--md-ink-faint);
    transition: color 180ms var(--about-ease);
  }
  .close-row:hover .close-meta,
  .close-row:focus-visible .close-meta {
    color: var(--md-ink-mute);
  }
  .close-row.primary .close-name {
    color: var(--md-cobalt);
  }
  .close-row.primary .close-meta {
    color: color-mix(in oklab, var(--md-cobalt) 70%, var(--md-ink-mute));
  }
  .close-row.primary:hover .close-name,
  .close-row.primary:focus-visible .close-name {
    color: var(--md-cobalt-deep);
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
    .atlas-grid {
      grid-template-columns: 1fr;
    }
    .atlas-door {
      min-height: 0;
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
      height: 360px;
      min-height: 360px;
      max-height: 360px;
      padding: 22px 20px 16px;
      border-radius: 22px;
    }
    .stage-copy h3 {
      max-width: 16ch;
      font-size: clamp(26px, 7vw, 34px);
    }
    .stage-watermark {
      font-size: 160px;
      right: -8%;
      top: -4%;
    }
  }

  @media (max-width: 480px) {
    .atlas-keys {
      gap: 10px 14px;
    }
    .word {
      font-size: clamp(44px, 14vw, 64px);
    }
    .close-line {
      max-width: none;
    }
  }

  .colophon.calm .thesis,
  .colophon.calm .meridian,
    .colophon.calm .close,
  .colophon.calm .stage-copy,
  .colophon.calm .constellation-beam::after,
  .colophon.calm .atlas-door {
    transition: none !important;
  }
  .colophon.calm .thesis,
  .colophon.calm .meridian,
    .colophon.calm .close {
    opacity: 1;
    transform: none;
  }
</style>
