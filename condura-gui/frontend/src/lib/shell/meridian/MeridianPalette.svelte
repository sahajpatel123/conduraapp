<script lang="ts">
  /**
   * ⌘K Jump — routes + daily actions (Halt, theme, summon).
   * Nav-only left Halt buried in the dock; power users need a keyboard surface.
   */
  import type { RouteId } from './routes'
  import { ROUTE_HASH } from './routes'
  import { halt } from '../../stores/halt.svelte'
  import { overlay } from '../../stores/overlay.svelte'
  import { getResolvedTheme, onThemeChange, toggleLightDark, type ResolvedTheme } from '../../theme/condura-theme'

  interface Props {
    open: boolean
    route: RouteId
    onclose: () => void
    onnavigate: (r: RouteId) => void
  }
  let { open, route, onclose, onnavigate }: Props = $props()

  type NavItem = {
    kind: 'nav'
    id: string
    route: RouteId
    label: string
    hint: string
    kbd: string
  }
  type ActionItem = {
    kind: 'action'
    id: string
    label: string
    hint: string
    kbd: string
    run: () => void
    danger?: boolean
  }
  type Item = NavItem | ActionItem

  const NAV: Omit<NavItem, 'kind' | 'id' | 'kbd'>[] = [
    { route: 'chat', label: 'Ask', hint: 'Talk to Condura' },
    { route: 'hub', label: 'Hub', hint: 'Browse skills' },
    { route: 'skills', label: 'Skills', hint: 'Local procedures' },
    { route: 'sync', label: 'Sync', hint: 'Pair a device' },
    { route: 'audit', label: 'Audit', hint: 'Action ledger' },
    { route: 'replay', label: 'Replay', hint: 'Day meridian frames' },
    { route: 'channels', label: 'Channels', hint: 'Telegram & more' },
    { route: 'delegation', label: 'Agents', hint: 'Pending actions' },
    { route: 'account', label: 'Account', hint: 'Sign in' },
    { route: 'settings', label: 'Settings', hint: 'Theme, models, spend limit' },
    { route: 'about', label: 'About', hint: 'Build · promises · safety' },
  ]

  let theme = $state<ResolvedTheme>(getResolvedTheme())
  let q = $state('')
  let idx = $state(0)
  let inputEl = $state<HTMLInputElement | null>(null)
  let listEl = $state<HTMLUListElement | null>(null)

  $effect(() => {
    if (!open) return
    const off = onThemeChange((r) => {
      theme = r
    })
    return off
  })

  const items = $derived.by((): Item[] => {
    const nav: Item[] = NAV.map((n) => ({
      kind: 'nav' as const,
      id: `nav:${n.route}`,
      route: n.route,
      label: n.label,
      hint: n.hint,
      kbd: ROUTE_HASH[n.route].replace('#/', '/') || '/',
    }))
    const actions: Item[] = [
      {
        kind: 'action',
        id: 'action:theme',
        label: theme === 'dark' ? 'Use light mode' : 'Use dark mode',
        hint: 'Lighting',
        kbd: '⇧T',
        run: () => {
          theme = toggleLightDark()
        },
      },
      {
        kind: 'action',
        id: 'action:summon',
        label: 'Summon quick prompt',
        hint: 'Floating overlay',
        kbd: 'Hotkey',
        run: () => {
          overlay.show()
        },
      },
      {
        kind: 'action',
        id: 'action:halt',
        label: 'Stop everything',
        hint: 'Hard halt · cut the line',
        kbd: 'Halt',
        danger: true,
        run: () => {
          // Halt failures toast via halt store — never empty-catch the kill-switch.
          void halt.halt('palette')
          onclose()
        },
      },
    ]
    return [...nav, ...actions]
  })

  const filtered = $derived(
    items.filter((it) =>
      `${it.label} ${it.hint} ${it.id}`.toLowerCase().includes(q.trim().toLowerCase())
    )
  )

  function indexForRoute(list: Item[], id: RouteId): number {
    const i = list.findIndex((it) => it.kind === 'nav' && it.route === id)
    return i >= 0 ? i : 0
  }

  function scrollActiveIntoView(): void {
    queueMicrotask(() => {
      const on = listEl?.querySelector<HTMLElement>('button.on')
      on?.scrollIntoView({ block: 'nearest' })
    })
  }

  $effect(() => {
    if (!open) return
    q = ''
    idx = indexForRoute(items, route)
    queueMicrotask(() => {
      inputEl?.focus()
      scrollActiveIntoView()
    })
  })

  $effect(() => {
    const needle = q.trim().toLowerCase()
    const list = items.filter((it) =>
      `${it.label} ${it.hint} ${it.id}`.toLowerCase().includes(needle)
    )
    idx = needle ? 0 : indexForRoute(list, route)
    if (open) scrollActiveIntoView()
  })

  function run(item: Item): void {
    if (item.kind === 'nav') {
      onnavigate(item.route)
    } else {
      item.run()
    }
    onclose()
  }

  function onKey(e: KeyboardEvent): void {
    if (!open) return
    if (e.key === 'Escape') {
      e.preventDefault()
      onclose()
      return
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      idx = Math.min(idx + 1, Math.max(filtered.length - 1, 0))
      scrollActiveIntoView()
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      idx = Math.max(idx - 1, 0)
      scrollActiveIntoView()
    } else if (e.key === 'Enter') {
      e.preventDefault()
      const hit = filtered[idx]
      if (hit) run(hit)
    }
  }
</script>

<svelte:window onkeydown={onKey} />
{#if open}
  <div class="back" onclick={onclose} role="presentation"></div>
  <div class="panel" role="dialog" aria-label="Jump" aria-modal="true">
    <p class="jump-cite">Jump · meridian</p>
    <input bind:this={inputEl} bind:value={q} placeholder="Jump or act…" class="q" />
    <ul bind:this={listEl}>
      {#each filtered as item, i (item.id)}
        <li>
          <button
            type="button"
            class:on={i === idx}
            class:danger={item.kind === 'action' && item.danger}
            onclick={() => run(item)}
          >
            <span class="label">
              {#if item.kind === 'action'}
                <span class="tag" class:halt={item.danger}>act</span>
              {/if}
              {item.label}
            </span>
            <span class="hint">{item.hint}</span>
            <kbd>{item.kbd}</kbd>
          </button>
        </li>
      {:else}
        <li class="empty">No matches</li>
      {/each}
    </ul>
  </div>
{/if}

<style>
  .back {
    position: fixed;
    inset: 0;
    z-index: 80;
    background: var(--md-scrim);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    animation: md-fade 200ms var(--md-ease) both;
  }
  .panel {
    position: fixed;
    top: 14vh;
    left: 50%;
    width: min(520px, calc(100vw - 32px));
    background: color-mix(in oklab, var(--md-surface) 96%, transparent);
    border: 1px solid var(--md-line-strong);
    border-radius: 20px;
    z-index: 81;
    overflow: hidden;
    backdrop-filter: blur(18px) saturate(1.15);
    -webkit-backdrop-filter: blur(18px) saturate(1.15);
    box-shadow:
      var(--md-shadow-lift),
      0 1px 0 color-mix(in oklab, #fff 45%, transparent) inset;
    transform: translateX(-50%);
    animation: md-palette 320ms var(--md-spring) both;
  }
  .jump-cite {
    margin: 0;
    padding: 14px 20px 0;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .q {
    width: 100%;
    border: 0;
    border-bottom: 1px solid var(--md-line);
    padding: 18px 20px;
    font-size: 16px;
    background: transparent;
    color: var(--md-ink);
    outline: none;
  }
  ul {
    margin: 0;
    padding: 8px;
    max-height: 360px;
    overflow: auto;
  }
  button {
    width: 100%;
    display: grid;
    grid-template-columns: 1fr auto auto;
    gap: 12px;
    align-items: center;
    text-align: left;
    padding: 12px;
    border-radius: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: background 140ms var(--md-ease);
  }
  button.on,
  button:hover {
    background: color-mix(in oklab, var(--md-cobalt) 10%, transparent);
  }
  button.on {
    box-shadow: inset 3px 0 0 var(--md-cobalt);
  }
  button.danger.on,
  button.danger:hover {
    background: color-mix(in oklab, var(--md-halt) 10%, transparent);
  }
  button.danger.on {
    box-shadow: inset 3px 0 0 var(--md-halt);
  }
  button:focus-visible {
    outline: none;
    background: color-mix(in oklab, var(--md-cobalt) 14%, transparent);
    box-shadow:
      inset 3px 0 0 var(--md-cobalt),
      var(--md-focus);
  }
  .label {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }
  .tag {
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    padding: 2px 6px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 8%, transparent);
  }
  .tag.halt {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 28%, transparent);
    background: color-mix(in oklab, var(--md-halt) 8%, transparent);
  }
  .hint {
    font-size: 12px;
    font-weight: 500;
    color: var(--md-ink-faint);
  }
  kbd {
    font-family: var(--md-font-mono);
    font-size: 10px;
    color: var(--md-ink-faint);
  }
  .empty {
    padding: 20px;
    text-align: center;
    color: var(--md-ink-faint);
    font-size: 13px;
  }
  @keyframes md-palette {
    from {
      opacity: 0;
      transform: translateX(-50%) translateY(12px) scale(0.97);
    }
    to {
      opacity: 1;
      transform: translateX(-50%) translateY(0) scale(1);
    }
  }
  @media (max-width: 560px) {
    .panel {
      top: auto;
      bottom: 0;
      left: 0;
      right: 0;
      width: 100%;
      max-width: none;
      border-radius: 22px 22px 0 0;
      border-left: 0;
      border-right: 0;
      border-bottom: 0;
      transform: none;
      max-height: min(78vh, 640px);
      display: flex;
      flex-direction: column;
      animation: md-palette-sheet 340ms var(--md-spring) both;
      padding-bottom: env(safe-area-inset-bottom, 0);
    }
    ul {
      flex: 1;
      max-height: none;
      padding: 8px 8px 16px;
    }
    button {
      grid-template-columns: 1fr auto;
      gap: 8px;
      padding: 14px 12px;
    }
    button .hint {
      display: none;
    }
    kbd {
      justify-self: end;
    }
  }
  @keyframes md-palette-sheet {
    from {
      opacity: 0;
      transform: translateY(24px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .back,
    .panel {
      animation: none !important;
    }
  }
</style>
