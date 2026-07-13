<script lang="ts">
  /**
   * Ask — the conductor’s desk.
   * Signature: ask atlas (empty) · plan plates with gated steps (thread) · tone-aware composer.
   */
  import { onMount } from 'svelte'
  import { conversation } from '../../stores/conversation.svelte'
  import { settings } from '../../stores/settings.svelte'
  import { halt } from '../../stores/halt.svelte'
  import { daemon } from '../../stores/daemon.svelte'
  import { spend } from '../../stores/spend.svelte'
  import { ipc } from '../../ipc/client'
  import { renderSafeMarkdown } from '../../markdown'
  import type { InstalledSkill, Message, ProviderInfo, ToolCall } from '../../ipc/types'
  import {
    buildSkillSystemPrompt,
    filterSlashSuggestions,
    parseAskSlash,
  } from '../../skill-slash'

  const STARTERS: { id: string; kicker: string; label: string; body: string }[] = [
    {
      id: 'sum',
      kicker: 'See',
      label: 'What’s on my screen?',
      body: 'Describe what’s visible. No clicks until you allow them.',
    },
    {
      id: 'fix',
      kicker: 'Fix',
      label: 'Help with the last error I hit',
      body: 'Diagnose first. Propose a fix. Wait for your OK.',
    },
    {
      id: 'plan',
      kicker: 'Plan',
      label: 'Plan the next hour of work',
      body: 'A short plan you can edit — not a silent sprint.',
    },
    {
      id: 'safe',
      kicker: 'Trust',
      label: 'What can you do without asking?',
      body: 'Name the safe band. Everything else stays locked.',
    },
  ]

  const PIPE = [
    { n: '1', t: 'Ask' },
    { n: '2', t: 'Plan' },
    { n: '3', t: 'You OK' },
    { n: '4', t: 'Act' },
    { n: '5', t: 'Audit' },
  ]

  let draft = $state('')
  let focused = $state(false)
  let ta = $state<HTMLTextAreaElement | null>(null)
  let scrollEl = $state<HTMLDivElement | null>(null)
  let selectedModel = $state('')
  let providers = $state<ProviderInfo[]>([])
  let skills = $state<InstalledSkill[]>([])
  let atlasFocus = $state<string | null>(null)
  let copied = $state(false)
  let reduceMotion = $state(false)
  let slashIndex = $state(0)
  let modelPersistError = $state('')
  let slashNotice = $state('')

  const modelOptions = $derived(
    providers.flatMap((p) =>
      (p.models ?? []).map((m) => ({
        value: `${p.name}:${m.id}`,
        label: `${p.name} · ${m.id}`,
      }))
    )
  )

  const slashSuggestions = $derived(filterSlashSuggestions(draft, skills))
  const showSlash = $derived(
    draft.startsWith('/') && !draft.includes('\n') && slashSuggestions.length > 0 && !conversation.isStreaming
  )

  const hasMessages = $derived(conversation.messages.length > 0 || conversation.isStreaming)
  const halted = $derived(halt.state.halted)
  const connected = $derived(daemon.connected)
  /** Selected provider must appear in the live catalog (keyed / local). */
  const selectedProviderLive = $derived(
    !!selectedModel.includes(':') &&
      providers.some((p) => selectedModel.startsWith(`${p.name}:`))
  )
  /** Connected but no usable provider:model — Settings must enable a key / Ollama. */
  const setupNeeded = $derived(connected && !halted && !selectedProviderLive)
  /** Backend refuses llm.stream when daily cap is hit; gate the composer too. */
  const capReached = $derived(spend.cap > 0 && spend.pct >= 100)
  const spendHot = $derived(spend.cap > 0 && spend.pct >= 80 && !capReached)
  const deskBlocked = $derived(setupNeeded || capReached || halted || !connected)
  const canSend = $derived(
    !!draft.trim() &&
      !conversation.isStreaming &&
      !halted &&
      connected &&
      selectedProviderLive &&
      !capReached
  )
  /** Why Send is disabled when the user already typed — reduces “broken button” frustration. */
  const sendBlocker = $derived.by(() => {
    const hasDraft = !!draft.trim()
    if (!hasDraft || canSend || conversation.isStreaming) return ''
    if (halted) return 'Halt is on — resume before sending.'
    if (!connected) return 'Daemon offline — Condura can’t send until it’s connected.'
    if (setupNeeded) return 'No model ready — open Settings → Models first.'
    if (capReached) return 'Daily spend cap reached — raise it in Settings to send.'
    if (!selectedProviderLive) return 'Choose a live model below, then send.'
    return ''
  })
  const tone = $derived<
    'halted' | 'thinking' | 'ready' | 'offline' | 'setup' | 'capped' | 'idle'
  >(
    halted
      ? 'halted'
      : conversation.isStreaming
        ? 'thinking'
        : !connected
          ? 'offline'
          : setupNeeded
            ? 'setup'
            : capReached
              ? 'capped'
              : canSend
                ? 'ready'
                : 'idle'
  )
  const toneLabel = $derived(
    tone === 'halted'
      ? 'Halted — nothing runs'
      : tone === 'thinking'
        ? 'Working · watch the plan'
        : tone === 'ready'
          ? 'Ready — press Enter'
          : tone === 'offline'
            ? 'Offline'
            : tone === 'setup'
              ? 'Needs a model'
              : tone === 'capped'
                ? 'Spend cap hit'
                : 'Type below to ask'
  )
  const liveNote = $derived(
    halted
      ? 'Halt is on — resume to use Ask again.'
      : !connected
        ? 'Offline · start the Condura daemon to chat'
        : setupNeeded
          ? 'Connected · pick a model in Settings to unlock Ask'
          : capReached
            ? `Spend cap · $${spend.spent.toFixed(2)} / $${spend.cap.toFixed(2)} — raise it in Settings`
            : selectedProviderLive
              ? `Ready · ${selectedModel.replace(':', ' · ')}${spend.cap > 0 ? ` · $${spend.spent.toFixed(2)}/$${spend.cap.toFixed(2)}` : ''}`
              : 'Connected · choose a model below'
  )
  const spendError = $derived(
    !!conversation.streamingError &&
      /spend cap|daily spend/i.test(conversation.streamingError)
  )
  /** Enough for daily multi-thread use without a full sidebar. */
  const recent = $derived((conversation.conversations ?? []).slice(0, 12))
  let confirmDeleteId = $state<number | null>(null)
  let deleting = $state(false)
  let renaming = $state(false)
  let renameDraft = $state('')
  let renameBusy = $state(false)

  onMount(() => {
    reduceMotion =
      typeof matchMedia !== 'undefined' && matchMedia('(prefers-reduced-motion: reduce)').matches
    void (async () => {
      await settings.refresh().catch(() => {})
      await conversation.refreshList().catch(() => {})
      try {
        const list = await ipc.providersList()
        // Normalize models (daemon now embeds catalog; tolerate legacy name-only).
        providers = (list ?? []).map((p) => ({
          name: p.name,
          models: (p.models ?? [])
            .map((m) => ({ id: m?.id ?? '' }))
            .filter((m) => !!m.id),
          available: p.available,
        }))
        // Ask only offers providers that can actually stream (keyed / local).
        providers = providers.filter((p) => p.available !== false && p.models.length > 0)
        // If a provider still has no models, fetch via providers.models.
        await Promise.all(
          providers.map(async (p, i) => {
            if (p.models.length > 0) return
            try {
              const ms = await ipc.providersModels(p.name)
              providers[i] = {
                ...p,
                models: (ms ?? []).map((m) => ({ id: m.id })).filter((m) => !!m.id),
              }
            } catch {
              /* leave empty */
            }
          })
        )
        providers = [...providers]
        const cfgProviders = settings.config?.llm?.providers
        if (cfgProviders) {
          const enabled = Object.entries(cfgProviders).find(
            ([name, pr]) =>
              pr.enabled &&
              pr.default_model &&
              providers.some((p) => p.name === name)
          )
          if (enabled?.[1]?.default_model) {
            selectedModel = `${enabled[0]}:${enabled[1].default_model}`
          }
        }
        if (!selectedModel || !providers.some((p) => selectedModel.startsWith(`${p.name}:`))) {
          const withModels = providers.find((p) => p.models.length > 0)
          selectedModel = withModels
            ? `${withModels.name}:${withModels.models[0]!.id}`
            : ''
        }
      } catch {
        /* preview */
      }
      try {
        skills = (await ipc.skillsList(100)) ?? []
      } catch {
        skills = []
      }
      try {
        const seeded = sessionStorage.getItem('md-ask-starter')
        if (seeded) {
          sessionStorage.removeItem('md-ask-starter')
          draft = seeded
        }
      } catch {
        /* ignore */
      }
      ta?.focus()
      resize()
    })()
  })

  $effect(() => {
    // Keep highlight in range when the suggestion list shrinks.
    draft
    if (slashIndex >= slashSuggestions.length) slashIndex = 0
  })

  $effect(() => {
    conversation.messages.length
    conversation.streamingDelta
    conversation.streamingToolCalls.length
    if (!scrollEl) return
    scrollEl.scrollTop = scrollEl.scrollHeight
  })

  function resize(): void {
    if (!ta) return
    ta.style.height = 'auto'
    ta.style.height = `${Math.min(ta.scrollHeight, 160)}px`
  }

  async function persistSelectedModel(): Promise<void> {
    modelPersistError = ''
    if (!selectedModel.includes(':')) return
    if (!settings.config) {
      modelPersistError = 'Couldn’t save model — Condura is offline.'
      return
    }
    const [provider, model] = selectedModel.split(':')
    if (!provider || !model) return
    const providersCfg = { ...(settings.config.llm?.providers ?? {}) }
    const cur = providersCfg[provider] ?? {
      enabled: false,
      api_key: '',
      base_url: '',
      default_model: '',
    }
    providersCfg[provider] = { ...cur, enabled: true, default_model: model }
    try {
      await settings.save({
        llm: { ...settings.config.llm, providers: providersCfg },
      })
    } catch (e) {
      modelPersistError = `Couldn’t save model: ${String(e)}`
    }
  }

  async function send(text = draft): Promise<void> {
    const trimmed = text.trim()
    if (!trimmed || conversation.isStreaming || halted) return
    slashNotice = ''
    if (capReached) {
      draft = trimmed
      goSettingsSpend()
      return
    }
    if (setupNeeded || !selectedModel.includes(':')) {
      draft = trimmed
      goSettingsModels()
      return
    }
    const [provider, model] = selectedModel.split(':')
    if (!provider || !model) {
      draft = trimmed
      goSettingsModels()
      return
    }

    const slash = parseAskSlash(trimmed, skills)
    if (slash.kind === 'builtin') {
      draft = ''
      queueMicrotask(resize)
      if (slash.token === 'clear') return
      if (slash.token === 'model') {
        goSettingsModels()
        return
      }
      if (slash.token === 'help' || slash.token === 'about') {
        slashNotice =
          'Slash tips: /SkillName runs a skill · /model opens Models · /clear clears the box. Add skills under Skills.'
        return
      }
      if (slash.token === 'compact') {
        slashNotice = '/compact isn’t available in Ask yet.'
        return
      }
      return
    }

    // Unknown /Token — fail loudly instead of sending garbage to the model.
    const leading = trimmed.match(/^\/([A-Za-z0-9._-]+)(?:\s|$)/)
    if (slash.kind === 'none' && leading) {
      slashNotice = `Unknown command /${leading[1]}. Try /help, or pick a skill from the / menu.`
      return
    }

    let userText = trimmed
    let skillSystem: string | undefined
    if (slash.kind === 'skill') {
      skillSystem = buildSkillSystemPrompt(slash.skill)
      userText = slash.rest
        ? `${primaryDisplay(slash.skill)} ${slash.rest}`
        : `${primaryDisplay(slash.skill)} — follow the skill procedure.`
      // Refresh skill steps if list metadata was shallow.
      if (!(slash.skill.steps?.length) && slash.skill.id) {
        try {
          const full = await ipc.skillsGet(slash.skill.id)
          skillSystem = buildSkillSystemPrompt(full)
        } catch {
          /* use list metadata */
        }
      }
    }

    draft = ''
    queueMicrotask(resize)
    await conversation.send(provider, model, userText, { skillSystem })
  }

  function primaryDisplay(s: InstalledSkill): string {
    const t = (s.trigger_pattern || '').trim()
    if (t.startsWith('/')) return t
    return `/${s.name.replace(/\s+/g, '') || s.id}`
  }

  function applySlashSuggestion(insert: string): void {
    draft = insert
    slashIndex = 0
    queueMicrotask(() => {
      resize()
      ta?.focus()
    })
  }

  function onKey(e: KeyboardEvent): void {
    if (showSlash && (e.key === 'ArrowDown' || e.key === 'ArrowUp')) {
      e.preventDefault()
      const n = slashSuggestions.length
      if (!n) return
      slashIndex =
        e.key === 'ArrowDown' ? (slashIndex + 1) % n : (slashIndex - 1 + n) % n
      return
    }
    if (showSlash && e.key === 'Tab') {
      e.preventDefault()
      const hit = slashSuggestions[slashIndex] ?? slashSuggestions[0]
      if (hit) applySlashSuggestion(hit.insert)
      return
    }
    if (e.key === 'Enter' && !e.shiftKey) {
      if (showSlash && slashSuggestions[slashIndex] && !draft.includes(' ')) {
        e.preventDefault()
        applySlashSuggestion(slashSuggestions[slashIndex]!.insert)
        return
      }
      e.preventDefault()
      void send()
    }
    if (e.key === 'Escape' && showSlash) {
      e.preventDefault()
      draft = ''
      slashNotice = ''
      queueMicrotask(resize)
      return
    }
    if (e.key === 'Escape' && renaming) {
      e.preventDefault()
      cancelRename()
      return
    }
    if (e.key === 'Escape' && confirmDeleteId) {
      e.preventDefault()
      cancelDelete()
      return
    }
    if (e.key === 'Escape' && conversation.isStreaming) void conversation.cancel()
  }

  function onInput(): void {
    resize()
  }

  function pickStarter(label: string): void {
    if (setupNeeded || halted || capReached) {
      if (capReached) goSettingsSpend()
      else if (setupNeeded) goSettingsModels()
      return
    }
    draft = label
    void send(label)
  }

  function goSettingsModels(): void {
    // Fragment path: hashToRoute treats any #/settings* as settings; Settings
    // scrolls the providers plate into view.
    window.location.hash = '#/settings/models'
  }

  function goSettingsSpend(): void {
    window.location.hash = '#/settings/spend'
  }

  function toolLabel(tc: ToolCall): string {
    return tc.function?.name || 'gated action'
  }

  function toolArgs(tc: ToolCall): string {
    const raw = tc.function?.arguments?.trim()
    if (!raw) return ''
    try {
      const parsed = JSON.parse(raw) as Record<string, unknown>
      const keys = Object.keys(parsed)
      if (!keys.length) return ''
      return keys
        .slice(0, 3)
        .map((k) => `${k}: ${String(parsed[k]).slice(0, 48)}`)
        .join(' · ')
    } catch {
      return raw.slice(0, 80)
    }
  }

  function stepsFor(msg: Message): ToolCall[] {
    return msg.tool_calls ?? []
  }

  async function clearThread(): Promise<void> {
    if (conversation.isStreaming) await conversation.cancel()
    await conversation.createNew().catch(() => {
      conversation.messages = []
    })
  }

  async function openThread(id: number): Promise<void> {
    confirmDeleteId = null
    renaming = false
    await conversation.open(id).catch(() => {})
  }

  function beginRename(): void {
    if (!conversation.currentID) return
    confirmDeleteId = null
    renameDraft = conversation.currentTitle || ''
    renaming = true
  }

  function cancelRename(): void {
    if (renameBusy) return
    renaming = false
    renameDraft = ''
  }

  async function commitRename(): Promise<void> {
    if (!conversation.currentID || renameBusy) return
    const next = renameDraft.trim()
    if (!next) {
      cancelRename()
      return
    }
    renameBusy = true
    try {
      await conversation.rename(conversation.currentID, next)
      renaming = false
      renameDraft = ''
    } catch {
      /* keep editor open; user can retry */
    } finally {
      renameBusy = false
    }
  }

  function requestDelete(id: number): void {
    confirmDeleteId = id
  }

  function cancelDelete(): void {
    if (deleting) return
    confirmDeleteId = null
  }

  async function confirmDelete(): Promise<void> {
    const id = confirmDeleteId
    if (!id || deleting) return
    deleting = true
    try {
      await conversation.deleteById(id)
      confirmDeleteId = null
      await conversation.refreshList().catch(() => {})
    } finally {
      deleting = false
    }
  }

  function goAudit(): void {
    window.location.hash = '#/audit'
  }

  function threadLabel(c: { id: number; title?: string }): string {
    const t = (c.title || '').trim()
    if (t && t !== 'New conversation') return t
    return `Thread ${c.id}`
  }

  async function copyLast(): Promise<void> {
    const last = [...conversation.messages].reverse().find((m) => m.role === 'assistant' && m.content)
    if (!last?.content) return
    try {
      await navigator.clipboard.writeText(last.content)
      copied = true
      setTimeout(() => (copied = false), 1600)
    } catch {
      /* ignore */
    }
  }
</script>

<section class="chat">
  <div class="feed" bind:this={scrollEl}>
    {#if !hasMessages}
      <div class="hero" class:calm={reduceMotion}>
        <header class="thesis">
          <p class="brand-line">
            <span class="word">Condura</span>
            <span class="slash" aria-hidden="true">/</span>
            <span class="edition">Ask</span>
          </p>
          <h1>What do you need?</h1>
          <p class="sub">
            Condura plans first, then waits for your OK before acting. Local by default.
          </p>
          <p
            class="live"
            class:hot={connected && !halted && !setupNeeded && !capReached}
            class:bad={halted || capReached}
            class:setup={setupNeeded}
          >
            <span class="live-dot" aria-hidden="true"></span>
            {liveNote}
          </p>
        </header>

        {#if setupNeeded}
          <div class="setup-plate" role="status">
            <p class="cite">One step left</p>
            <h2>Choose a model to unlock Ask</h2>
            <p>
              Condura is connected. Turn on Ollama or paste a cloud API key, pick the Ask model,
              then come back here.
            </p>
            <button type="button" class="md-btn md-btn-primary" onclick={goSettingsModels}>
              Open Models settings
            </button>
          </div>
        {:else if capReached}
          <div class="setup-plate cap-plate" role="status">
            <p class="cite">Spend limit</p>
            <h2>Today’s cloud spend hit the cap</h2>
            <p>
              ${spend.spent.toFixed(2)} of ${spend.cap.toFixed(2)} used. Raise the cap in Settings,
              or wait until tomorrow. Gatekeeper still holds every action.
            </p>
            <button type="button" class="md-btn md-btn-primary" onclick={goSettingsSpend}>
              Raise spend cap
            </button>
          </div>
        {:else if !connected}
          <div class="setup-plate" role="status">
            <p class="cite">Connection</p>
            <h2>Daemon is offline</h2>
            <p>Start Condura’s background service, then this desk will light up.</p>
          </div>
        {:else if halted}
          <div class="setup-plate cap-plate" role="status">
            <p class="cite">Halt</p>
            <h2>Everything is stopped</h2>
            <p>Resume from the Halt page (or CLI) when you’re ready to ask again.</p>
            <button type="button" class="md-btn md-btn-primary" onclick={() => (window.location.hash = '#/halt')}>
              Open Halt
            </button>
          </div>
        {:else}
          <ol class="pipe" aria-label="How an ask becomes an act">
            {#each PIPE as p (p.n)}
              <li>
                <span class="n">{p.n}</span>
                <span class="t">{p.t}</span>
              </li>
            {/each}
          </ol>

          <div class="atlas">
            <div class="atlas-head">
              <p class="cite">Quick starts</p>
              <h2>Try an example</h2>
              <p class="atlas-note">
                Tap one to fill the box — edit it, or write your own below.
              </p>
            </div>
            <div class="atlas-grid md-stagger">
              {#each STARTERS as s (s.id)}
                <button
                  type="button"
                  class="door"
                  class:focus={atlasFocus === s.id}
                  disabled={halted}
                  onmouseenter={() => (atlasFocus = s.id)}
                  onfocus={() => (atlasFocus = s.id)}
                  onmouseleave={() => (atlasFocus = null)}
                  onblur={() => (atlasFocus = null)}
                  onclick={() => pickStarter(s.label)}
                >
                  <span class="door-k">{s.kicker}</span>
                  <span class="door-t">{s.label}</span>
                  <span class="door-b">{s.body}</span>
                  <span class="door-a">
                    Use this
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                      <path
                        d="M5 12h14M13 6l6 6-6 6"
                        stroke="currentColor"
                        stroke-width="1.8"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      />
                    </svg>
                  </span>
                </button>
              {/each}
            </div>
          </div>
        {/if}

        {#if recent.length && !deskBlocked}
          <div class="recent">
            <p class="cite">Continue</p>
            <div class="recent-row">
              {#each recent as c (c.id)}
                <button
                  type="button"
                  class="recent-chip"
                  class:on={c.id === conversation.currentID}
                  onclick={() => void openThread(c.id)}
                >
                  {threadLabel(c)}
                </button>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {:else}
      <div class="thread">
        <div class="thread-bar">
          <div class="thread-head">
            <p class="cite">thread · gated</p>
            {#if renaming}
              <form
                class="rename-form"
                onsubmit={(e) => {
                  e.preventDefault()
                  void commitRename()
                }}
              >
                <input
                  class="rename-input"
                  bind:value={renameDraft}
                  maxlength="120"
                  aria-label="Thread title"
                  disabled={renameBusy}
                  onkeydown={(e) => {
                    if (e.key === 'Escape') {
                      e.preventDefault()
                      cancelRename()
                    }
                  }}
                />
                <button type="submit" class="md-btn md-btn-primary tiny" disabled={renameBusy || !renameDraft.trim()}>
                  {renameBusy ? 'Saving…' : 'Save'}
                </button>
                <button type="button" class="md-btn md-btn-ghost tiny" disabled={renameBusy} onclick={cancelRename}>
                  Cancel
                </button>
              </form>
            {:else}
              <button type="button" class="thread-title-btn" onclick={beginRename} title="Rename thread">
                <h2 class="thread-title">{conversation.currentTitle || 'Ask'}</h2>
                <span class="rename-hint">rename</span>
              </button>
            {/if}
          </div>
          <div class="thread-actions">
            <button type="button" class="md-btn md-btn-ghost tiny" onclick={() => void copyLast()} disabled={!conversation.messages.some((m) => m.role === 'assistant' && m.content)}>
              {copied ? 'Copied' : 'Copy last'}
            </button>
            <button type="button" class="md-btn md-btn-ghost tiny" onclick={goAudit}>Open audit</button>
            <button type="button" class="md-btn md-btn-ghost tiny" onclick={() => void clearThread()}>
              New ask
            </button>
            {#if conversation.currentID}
              <button
                type="button"
                class="md-btn md-btn-ghost tiny danger"
                disabled={deleting || renaming}
                onclick={() => requestDelete(conversation.currentID)}
              >
                Remove
              </button>
            {/if}
          </div>
        </div>

        {#if confirmDeleteId === conversation.currentID}
          <div class="delete-plate" role="alertdialog" aria-labelledby="thread-delete-title">
            <p class="cite">remove thread</p>
            <h3 id="thread-delete-title">Delete this conversation?</h3>
            <p class="delete-lead">
              Removes the local thread and cancels any in-flight stream. Audit stays intact.
            </p>
            <div class="delete-actions">
              <button
                type="button"
                class="md-btn md-btn-danger"
                disabled={deleting}
                onclick={() => void confirmDelete()}
              >
                {deleting ? 'Removing…' : 'Delete thread'}
              </button>
              <button type="button" class="md-btn md-btn-ghost" disabled={deleting} onclick={cancelDelete}>
                Keep
              </button>
            </div>
          </div>
        {/if}

        {#if recent.length > 0}
          <div class="rail" aria-label="Recent threads">
            {#each recent as c (c.id)}
              <div class="rail-row" class:on={c.id === conversation.currentID}>
                <button
                  type="button"
                  class="rail-item"
                  class:on={c.id === conversation.currentID}
                  onclick={() => void openThread(c.id)}
                >
                  {threadLabel(c)}
                </button>
                <button
                  type="button"
                  class="rail-x"
                  aria-label={`Delete ${threadLabel(c)}`}
                  disabled={deleting}
                  onclick={() => requestDelete(c.id)}
                >
                  ×
                </button>
              </div>
            {/each}
          </div>
        {/if}

        {#if confirmDeleteId && confirmDeleteId !== conversation.currentID}
          <div class="delete-plate slim" role="alertdialog" aria-label="Confirm delete">
            <p class="delete-lead tight">
              Delete
              <strong>{threadLabel(recent.find((c) => c.id === confirmDeleteId) ?? { id: confirmDeleteId })}</strong>?
            </p>
            <div class="delete-actions">
              <button
                type="button"
                class="md-btn md-btn-danger"
                disabled={deleting}
                onclick={() => void confirmDelete()}
              >
                {deleting ? 'Removing…' : 'Delete'}
              </button>
              <button type="button" class="md-btn md-btn-ghost" disabled={deleting} onclick={cancelDelete}>
                Keep
              </button>
            </div>
          </div>
        {/if}

        <div class="messages">
          {#each conversation.messages as msg, i (i)}
            <article class="msg" data-role={msg.role}>
              <header>
                <span>{msg.role === 'user' ? 'You' : 'Condura'}</span>
                {#if msg.role === 'assistant'}
                  <span class="cite">
                    {stepsFor(msg).length
                      ? `${stepsFor(msg).length} gated step${stepsFor(msg).length === 1 ? '' : 's'}`
                      : 'plan · gated'}
                  </span>
                {/if}
              </header>
              {#if msg.content}
                {#if msg.role === 'assistant'}
                  <div class="bubble md">{@html renderSafeMarkdown(msg.content)}</div>
                {:else}
                  <div class="bubble plain">{msg.content}</div>
                {/if}
              {/if}
              {#if stepsFor(msg).length}
                <ol class="steps">
                  {#each stepsFor(msg) as tc, si (tc.id)}
                    <li>
                      <span class="step-n" aria-hidden="true">{String(si + 1).padStart(2, '0')}</span>
                      <div>
                        <strong>{toolLabel(tc)}</strong>
                        {#if toolArgs(tc)}
                          <span class="args">{toolArgs(tc)}</span>
                        {/if}
                        <span class="gate">awaits consent before action</span>
                      </div>
                    </li>
                  {/each}
                </ol>
              {/if}
            </article>
          {/each}
          {#if conversation.isStreaming}
            <article class="msg" data-role="assistant">
              <header>
                <span>Condura</span>
                <span class="cite live">thinking</span>
              </header>
              {#if conversation.streamingDelta}
                <div class="bubble md streaming">{@html renderSafeMarkdown(conversation.streamingDelta)}</div>
              {:else}
                <div class="bubble streaming muted">Reading the room…</div>
              {/if}
              {#if conversation.streamingToolCalls.length}
                <ol class="steps live">
                  {#each conversation.streamingToolCalls as tc, si (tc.id)}
                    <li>
                      <span class="step-n" aria-hidden="true">{String(si + 1).padStart(2, '0')}</span>
                      <div>
                        <strong>{toolLabel(tc)}</strong>
                        {#if toolArgs(tc)}
                          <span class="args">{toolArgs(tc)}</span>
                        {/if}
                        <span class="gate">proposed · not yet allowed</span>
                      </div>
                    </li>
                  {/each}
                </ol>
              {/if}
            </article>
          {/if}
          {#if conversation.streamingError}
            <div class="err-plate" class:cap={spendError} role="alert">
              <p class="err">
                {#if spendError}
                  Daily spend cap blocked this ask.
                {:else}
                  {conversation.streamingError}
                {/if}
              </p>
              {#if spendError}
                <button type="button" class="md-btn md-btn-ghost" onclick={goSettingsSpend}>
                  Raise cap →
                </button>
              {/if}
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>

  <div class="composer" class:focused class:ready={canSend} data-tone={tone}>
    <div class="tone" aria-live="polite">
      <span class="tone-dot" aria-hidden="true"></span>
      {toneLabel}
    </div>
    {#if showSlash}
      <div class="slash-wrap">
        <ul class="slash-menu" role="listbox" aria-label="Slash commands">
          {#each slashSuggestions as item, i (item.label)}
            <li role="option" aria-selected={i === slashIndex}>
              <button
                type="button"
                class="slash-item"
                class:on={i === slashIndex}
                onmousedown={(e) => {
                  e.preventDefault()
                  applySlashSuggestion(item.insert)
                }}
              >
                <code>{item.label}</code>
                <span>{item.hint}</span>
              </button>
            </li>
          {/each}
        </ul>
        <p class="slash-foot">↑↓ move · Tab or Enter to use · Esc clears</p>
      </div>
    {/if}
    {#if slashNotice}
      <p class="composer-note" role="status">{slashNotice}</p>
    {/if}
    {#if modelPersistError}
      <p class="composer-note warn" role="alert">{modelPersistError}</p>
    {/if}
    {#if sendBlocker}
      <p class="composer-note warn" role="status">{sendBlocker}</p>
    {/if}
    <textarea
      bind:this={ta}
      bind:value={draft}
      class="input"
      rows="1"
      placeholder={halted
        ? 'Halted — resume to ask…'
        : setupNeeded
          ? 'Add a model in Settings first…'
          : capReached
            ? 'Spend cap reached — raise it in Settings…'
            : conversation.isStreaming
              ? 'Working…'
              : 'Ask anything…  type / for skills'}
      disabled={halted || conversation.isStreaming || setupNeeded || capReached}
      onfocus={() => (focused = true)}
      onblur={() => (focused = false)}
      onkeydown={onKey}
      oninput={() => {
        slashNotice = ''
        onInput()
      }}
    ></textarea>
    <div class="bar">
      <div class="meta">
        {#if modelOptions.length > 0}
          <label class="model">
            <span class="sr">Model</span>
            <select
              bind:value={selectedModel}
              aria-label="Model"
              disabled={halted || conversation.isStreaming || capReached}
              onchange={() => void persistSelectedModel()}
            >
              {#each modelOptions as opt (opt.value)}
                <option value={opt.value}>{opt.label}</option>
              {/each}
            </select>
          </label>
        {:else if setupNeeded}
          <button type="button" class="setup-chip" onclick={goSettingsModels}>
            Configure model →
          </button>
        {:else if !connected}
          <span class="offline warn" title="Daemon offline">
            <span class="full">Daemon offline</span>
            <span class="short">Offline</span>
          </span>
        {:else}
          <span
            class="offline"
            class:warn={!selectedModel}
            title={selectedModel || 'No model'}
          >
            <span class="full">{selectedModel || 'No model'}</span>
            <span class="short">{selectedModel ? selectedModel.split(':').pop() : 'No model'}</span>
          </span>
        {/if}
        {#if spend.cap > 0}
          <button
            type="button"
            class="spend-chip"
            class:hot={spendHot}
            class:cap={capReached}
            title="Daily spend"
            onclick={goSettingsSpend}
          >
            ${spend.spent.toFixed(2)} / ${spend.cap.toFixed(2)}
            {#if capReached}
              · cap
            {:else if spendHot}
              · {spend.pct}%
            {/if}
          </button>
        {/if}
        <span class="hint">
          {#if setupNeeded}
            Settings · Models
          {:else if capReached}
            Settings · Spend
          {:else if sendBlocker}
            Fix above to send
          {:else}
            Enter · Esc stops · Shift+Enter for line
          {/if}
        </span>
      </div>
      <div class="actions">
        {#if conversation.isStreaming}
          <button type="button" class="md-btn md-btn-ghost" onclick={() => void conversation.cancel()}>
            Stop
          </button>
        {/if}
        {#if setupNeeded}
          <button type="button" class="md-btn md-btn-primary" onclick={goSettingsModels}>
            Models
          </button>
        {:else if capReached}
          <button type="button" class="md-btn md-btn-primary" onclick={goSettingsSpend}>
            Raise cap
          </button>
        {:else}
          <button
            type="button"
            class="md-btn md-btn-primary"
            disabled={!canSend}
            onclick={() => void send()}
          >
            Send
          </button>
        {/if}
      </div>
    </div>
  </div>
</section>

<style>
  .chat {
    height: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
  .feed {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 28px 28px 16px;
    scroll-behavior: smooth;
  }

  /* —— Empty desk —— */
  .hero {
    position: relative;
    max-width: 760px;
    margin: min(5vh, 40px) auto 0;
  }
  .hero::before {
    content: '';
    position: absolute;
    left: -8%;
    top: -6%;
    width: 48%;
    height: 36%;
    pointer-events: none;
    background: radial-gradient(ellipse at center, color-mix(in oklab, var(--md-cobalt) 7%, transparent), transparent 72%);
    filter: blur(10px);
    z-index: 0;
  }
  .hero > * {
    position: relative;
    z-index: 1;
  }
  .live.setup {
    color: var(--md-cobalt);
  }
  .live.setup .live-dot {
    background: var(--md-cobalt);
  }
  .setup-plate {
    margin: 20px 0 8px;
    padding: 16px 18px;
    border-radius: 12px;
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
    background: color-mix(in oklab, var(--md-cobalt) 3%, var(--md-surface));
    box-shadow: none;
  }
  .setup-plate .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 8px;
  }
  .setup-plate h2 {
    font-family: var(--md-font-display);
    font-size: 18px;
    font-weight: 650;
    letter-spacing: -0.03em;
    margin: 0 0 6px;
  }
  .setup-plate p {
    margin: 0 0 16px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
    max-width: 42ch;
  }
  .atlas.dim {
    opacity: 0.55;
  }
  .setup-chip {
    appearance: none;
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 28%, var(--md-line));
    background: color-mix(in oklab, var(--md-cobalt) 8%, var(--md-surface));
    color: var(--md-cobalt);
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    padding: 5px 9px;
    border-radius: 7px;
    cursor: pointer;
  }
  .setup-chip:hover {
    background: color-mix(in oklab, var(--md-cobalt) 16%, var(--md-surface));
  }
  .setup-chip:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .brand-line {
    display: flex;
    align-items: baseline;
    gap: 8px;
    margin: 0 0 18px;
    animation: md-rise 420ms var(--md-ease) both;
  }
  .word {
    font-family: var(--md-font-display);
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.03em;
    color: var(--md-ink);
  }
  .slash {
    color: var(--md-ink-faint);
    font-weight: 400;
  }
  .edition {
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  h1 {
    font-family: var(--md-font-display);
    font-size: clamp(30px, 5vw, 44px);
    font-weight: 650;
    letter-spacing: -0.04em;
    line-height: 1.08;
    margin: 0 0 10px;
    animation: md-rise 420ms var(--md-ease) 30ms both;
    color: var(--md-ink);
  }
  .hero.calm h1 {
    color: var(--md-ink);
  }
  .sub {
    margin: 0 0 12px;
    max-width: 42ch;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
    animation: md-rise 420ms var(--md-ease) 50ms both;
  }
  .live {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 28px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    color: var(--md-ink-faint);
    animation: md-rise 520ms var(--md-ease) 100ms both;
  }
  .live-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .live.hot {
    color: var(--md-live);
  }
  .live.hot .live-dot {
    background: var(--md-live);
    box-shadow: none;
  }
  .live.bad {
    color: var(--md-halt);
  }
  .live.bad .live-dot {
    background: var(--md-halt);
    box-shadow: none;
  }

  .pipe {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 4px;
    margin: 0 0 32px;
    padding: 0;
    list-style: none;
    animation: md-rise 520ms var(--md-ease) 120ms both;
  }
  .pipe li {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    border-radius: 7px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
    backdrop-filter: none;
    box-shadow: none;
  }
  .pipe .n {
    font-family: var(--md-font-mono);
    font-size: 10px;
    color: var(--md-cobalt);
    letter-spacing: 0.08em;
  }
  .pipe .t {
    font-size: 12px;
    font-weight: 550;
    color: var(--md-ink-soft);
  }

  .atlas-head {
    margin-bottom: 14px;
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 8px;
  }
  .atlas h2 {
    font-family: var(--md-font-display);
    font-size: 22px;
    letter-spacing: -0.04em;
    margin: 0 0 6px;
  }
  .atlas-note {
    margin: 0;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
    max-width: 46ch;
  }
  .atlas-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }
  .door {
    position: relative;
    text-align: left;
    display: grid;
    gap: 6px;
    padding: 14px 14px 12px 16px;
    border-radius: 11px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
    cursor: pointer;
    color: inherit;
    overflow: hidden;
    transition:
      border-color 140ms var(--md-ease),
      background 140ms var(--md-ease);
  }
  .door::before {
    content: '';
    position: absolute;
    left: 0;
    top: 12px;
    bottom: 12px;
    width: 2px;
    border-radius: 0 2px 2px 0;
    background: color-mix(in oklab, var(--md-cobalt) 40%, transparent);
    transition: background 140ms var(--md-ease);
  }
  .door:hover,
  .door.focus {
    border-color: var(--md-line-strong);
    background: color-mix(in oklab, var(--md-stage) 40%, var(--md-surface));
  }
  .door:hover::before,
  .door.focus::before {
    background: var(--md-cobalt);
  }
  .door:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .door:disabled {
    opacity: 0.45;
    cursor: not-allowed;
    transform: none;
  }
  .door-k {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-cobalt);
  }
  .door-t {
    font-family: var(--md-font-display);
    font-size: 15px;
    font-weight: 600;
    letter-spacing: -0.025em;
    line-height: 1.25;
  }
  .door-b {
    font-size: 12.5px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .door-a {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin-top: 6px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    color: var(--md-cobalt);
    font-weight: 600;
  }

  .recent {
    margin-top: 28px;
    padding-top: 20px;
    border-top: 1px solid var(--md-line);
  }
  .recent-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .recent-chip {
    padding: 7px 11px;
    border-radius: 7px;
    border: 1px solid var(--md-line);
    background: var(--md-stage);
    font-size: 12px;
    font-weight: 550;
    color: var(--md-ink-mute);
    cursor: pointer;
    max-width: 220px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .recent-chip.on {
    color: #fff;
    background: var(--md-cobalt);
    border-color: var(--md-cobalt);
  }
  .recent-chip:hover {
    border-color: var(--md-cobalt);
    color: var(--md-ink);
  }
  .recent-chip.on:hover {
    color: #fff;
  }
  .recent-chip:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }

  /* —— Thread —— */
  .thread {
    max-width: 760px;
    margin: 0 auto;
  }
  .thread-bar {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 14px;
    padding-bottom: 14px;
    border-bottom: 1px solid var(--md-line);
  }
  .thread-head {
    min-width: 0;
    flex: 1;
  }
  .thread-title {
    font-family: var(--md-font-display);
    font-size: 22px;
    letter-spacing: -0.04em;
    margin: 0;
  }
  .thread-title-btn {
    display: inline-flex;
    align-items: baseline;
    gap: 10px;
    max-width: 100%;
    text-align: left;
    cursor: pointer;
    color: inherit;
    background: none;
    border: 0;
    padding: 0;
  }
  .thread-title-btn .thread-title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .rename-hint {
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    flex: none;
  }
  .thread-title-btn:hover .rename-hint {
    color: var(--md-cobalt);
  }
  .thread-title-btn:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
    border-radius: 8px;
  }
  .rename-form {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    margin-top: 4px;
  }
  .rename-input {
    min-width: min(280px, 100%);
    flex: 1;
    padding: 8px 11px;
    border-radius: 8px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    font-family: var(--md-font-display);
    font-size: 16px;
    letter-spacing: -0.03em;
    color: var(--md-ink);
  }
  .rename-input:focus {
    outline: none;
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
  }
  .thread-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  :global(.md-btn.tiny) {
    padding: 7px 12px;
    font-size: 12px;
  }
  :global(.md-btn.tiny.danger) {
    color: var(--md-halt);
  }
  .delete-plate {
    margin: 0 0 14px;
    padding: 12px 14px;
    border-radius: 10px;
    border: 1px solid color-mix(in oklab, var(--md-halt) 22%, var(--md-line));
    background: color-mix(in oklab, var(--md-halt) 4%, var(--md-surface));
  }
  .delete-plate.slim {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }
  .delete-plate h3 {
    font-family: var(--md-font-display);
    font-size: 16px;
    letter-spacing: -0.03em;
    margin: 0 0 6px;
  }
  .delete-lead {
    margin: 0 0 12px;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
    max-width: 48ch;
  }
  .delete-lead.tight {
    margin: 0;
  }
  .delete-lead strong {
    color: var(--md-ink);
  }
  .delete-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .rail {
    display: flex;
    gap: 6px;
    overflow-x: auto;
    margin-bottom: 16px;
    padding-bottom: 2px;
  }
  .rail-row {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    flex: none;
    border-radius: 8px;
    border: 1px solid var(--md-line);
    background: transparent;
  }
  .rail-row.on {
    border-color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 10%, transparent);
  }
  .rail-item {
    flex: none;
    padding: 7px 10px 7px 12px;
    border: 0;
    background: transparent;
    font-size: 12px;
    font-weight: 600;
    color: var(--md-ink-faint);
    cursor: pointer;
    max-width: 140px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .rail-item.on {
    color: var(--md-cobalt);
  }
  .rail-x {
    width: 26px;
    height: 26px;
    margin-right: 4px;
    border-radius: 6px;
    font-size: 14px;
    line-height: 1;
    color: var(--md-ink-faint);
    cursor: pointer;
  }
  .rail-x:hover {
    color: var(--md-halt);
    background: color-mix(in oklab, var(--md-halt) 10%, transparent);
  }
  .rail-item:focus-visible,
  .rail-x:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }

  .messages {
    display: flex;
    flex-direction: column;
    gap: 18px;
    padding-bottom: 8px;
  }
  .msg {
    display: flex;
    flex-direction: column;
    gap: 6px;
    max-width: min(100%, 580px);
    animation: md-rise 360ms var(--md-ease) both;
  }
  .msg[data-role='user'] {
    align-self: flex-end;
    align-items: flex-end;
  }
  .msg[data-role='assistant'] {
    align-self: flex-start;
    align-items: flex-start;
  }
  .msg header {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    padding: 0 4px;
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .msg header .cite {
    margin: 0;
    letter-spacing: 0.1em;
    color: var(--md-cobalt);
    opacity: 0.85;
  }
  .msg header .cite.live {
    color: var(--md-live);
  }
  .msg[data-role='user'] header {
    color: var(--md-cobalt);
  }
  .msg[data-role='assistant'] header {
    color: var(--md-live);
  }
  .bubble {
    font-size: 15px;
    line-height: 1.55;
    word-break: break-word;
    padding: 11px 14px;
    border-radius: 12px;
  }
  .bubble.plain {
    white-space: pre-wrap;
  }
  .bubble.md {
    white-space: normal;
  }
  .bubble.md :global(p) {
    margin: 0.45em 0;
  }
  .bubble.md :global(p:first-child) {
    margin-top: 0;
  }
  .bubble.md :global(p:last-child) {
    margin-bottom: 0;
  }
  .bubble.md :global(ul),
  .bubble.md :global(ol) {
    margin: 0.45em 0;
    padding-left: 1.35em;
  }
  .bubble.md :global(li) {
    margin: 0.2em 0;
  }
  .bubble.md :global(code) {
    font-family: var(--md-font-mono);
    font-size: 0.88em;
    background: color-mix(in oklab, var(--md-stage) 88%, transparent);
    border: 1px solid var(--md-line);
    padding: 1px 6px;
    border-radius: 6px;
  }
  .bubble.md :global(pre) {
    margin: 0.65em 0;
    padding: 12px 14px;
    overflow-x: auto;
    border-radius: 12px;
    border: 1px solid var(--md-line);
    background: var(--md-stage);
  }
  .bubble.md :global(pre code) {
    background: none;
    border: none;
    padding: 0;
    font-size: 12.5px;
  }
  .bubble.md :global(blockquote) {
    margin: 0.55em 0;
    padding-left: 12px;
    border-left: 2px solid color-mix(in oklab, var(--md-cobalt) 45%, transparent);
    color: var(--md-ink-mute);
  }
  .bubble.md :global(a) {
    color: var(--md-cobalt);
  }
  .msg[data-role='user'] .bubble {
    background: color-mix(in oklab, var(--md-cobalt) 9%, var(--md-surface));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
    border-bottom-right-radius: 4px;
    box-shadow: none;
  }
  .msg[data-role='assistant'] .bubble {
    background: var(--md-surface);
    border: 1px solid var(--md-line);
    border-bottom-left-radius: 4px;
    box-shadow: none;
  }
  .bubble.muted {
    color: var(--md-ink-mute);
    font-style: italic;
  }
  .bubble.streaming::after {
    content: '';
    display: inline-block;
    width: 8px;
    height: 1.1em;
    margin-left: 2px;
    vertical-align: text-bottom;
    background: var(--md-live);
    border-radius: 2px;
    animation: md-caret 1s steps(1) infinite;
  }
  .steps {
    margin: 0;
    padding: 0;
    list-style: none;
    display: grid;
    gap: 8px;
    width: 100%;
  }
  .steps li {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 12px;
    align-items: start;
    padding: 10px 12px;
    border-radius: 10px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-stage) 65%, var(--md-surface));
  }
  .steps.live li {
    border-color: color-mix(in oklab, var(--md-live) 28%, transparent);
  }
  .step-n {
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    color: var(--md-cobalt);
    margin-top: 2px;
  }
  .steps.live .step-n {
    color: var(--md-live);
  }
  .steps strong {
    display: block;
    font-size: 13px;
    font-weight: 600;
    letter-spacing: -0.02em;
  }
  .steps .args {
    display: block;
    margin-top: 3px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-ink-faint);
    word-break: break-word;
  }
  .gate {
    display: block;
    margin-top: 6px;
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .err {
    color: var(--md-halt);
    font-size: 13px;
    margin: 0;
  }
  .err-plate {
    align-self: flex-start;
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    align-items: center;
    padding: 10px 12px;
    border-radius: 14px;
    border: 1px solid color-mix(in oklab, var(--md-halt) 28%, var(--md-line));
    background: color-mix(in oklab, var(--md-halt) 6%, var(--md-surface));
  }
  .err-plate.cap {
    border-color: color-mix(in oklab, var(--md-halt) 40%, transparent);
  }
  .spend-chip {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.04em;
    padding: 5px 9px;
    border-radius: 6px;
    border: 1px solid var(--md-line);
    background: var(--md-stage);
    color: var(--md-ink-mute);
    cursor: pointer;
  }
  .spend-chip.hot {
    color: #c4892a;
    border-color: color-mix(in oklab, #c4892a 35%, var(--md-line));
  }
  .spend-chip.cap {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 35%, var(--md-line));
  }
  .cap-plate {
    border-color: color-mix(in oklab, var(--md-halt) 30%, var(--md-line));
    background: color-mix(in oklab, var(--md-halt) 5%, var(--md-surface));
  }

  /* —— Composer —— */
  .composer {
    position: relative;
    margin: 0 auto 88px;
    max-width: 760px;
    width: calc(100% - 56px);
    border-radius: 12px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
    backdrop-filter: none;
    -webkit-backdrop-filter: none;
    box-shadow: none;
    padding: 10px 12px 12px;
    transition: border-color 140ms var(--md-ease);
  }
  .slash-wrap {
    margin: 0 0 8px;
  }
  .slash-menu {
    list-style: none;
    margin: 0;
    padding: 4px;
    border-radius: 10px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
    box-shadow: none;
    display: grid;
    gap: 1px;
    max-height: 220px;
    overflow: auto;
  }
  .slash-foot {
    margin: 4px 2px 0;
    padding: 2px 6px 0;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.04em;
    color: var(--md-ink-faint);
  }
  .composer-note {
    margin: 0 2px 8px;
    padding: 7px 9px;
    border-radius: 8px;
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 14%, var(--md-line));
    background: color-mix(in oklab, var(--md-cobalt) 4%, var(--md-surface));
    font-size: 12px;
    line-height: 1.4;
    color: var(--md-ink-mute);
  }
  .composer-note.warn {
    border-color: color-mix(in oklab, var(--md-halt) 22%, var(--md-line));
    background: color-mix(in oklab, var(--md-halt) 5%, var(--md-surface));
    color: var(--md-halt);
  }
  .slash-item {
    width: 100%;
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
    text-align: left;
    padding: 8px 10px;
    border: 0;
    border-radius: 10px;
    background: transparent;
    color: var(--md-ink);
    cursor: pointer;
    font: inherit;
  }
  .slash-item.on,
  .slash-item:hover {
    background: color-mix(in oklab, var(--md-cobalt) 10%, var(--md-stage));
  }
  .slash-item code {
    font-family: var(--md-font-mono);
    font-size: 12px;
    color: var(--md-cobalt);
  }
  .slash-item span {
    font-size: 12px;
    color: var(--md-ink-faint);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .composer .tone {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin: 2px 6px 8px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .tone-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .composer[data-tone='ready'] {
    border-color: color-mix(in oklab, var(--md-cobalt) 28%, var(--md-line-strong));
  }
  .composer[data-tone='ready'] .tone {
    color: var(--md-cobalt);
  }
  .composer[data-tone='ready'] .tone-dot {
    background: var(--md-cobalt);
  }
  .composer[data-tone='thinking'] {
    border-color: color-mix(in oklab, var(--md-live) 28%, var(--md-line-strong));
  }
  .composer[data-tone='thinking'] .tone {
    color: var(--md-live);
  }
  .composer[data-tone='thinking'] .tone-dot {
    background: var(--md-live);
    animation: md-tone-pulse 1.2s var(--md-ease) infinite;
  }
  .composer[data-tone='halted'] {
    border-color: color-mix(in oklab, var(--md-halt) 32%, var(--md-line-strong));
  }
  .composer[data-tone='halted'] .tone {
    color: var(--md-halt);
  }
  .composer[data-tone='halted'] .tone-dot {
    background: var(--md-halt);
  }
  .composer[data-tone='offline'] .tone {
    color: var(--md-ink-mute);
  }
  .composer[data-tone='setup'] {
    border-color: color-mix(in oklab, var(--md-cobalt) 32%, transparent);
  }
  .composer[data-tone='setup'] .tone {
    color: var(--md-cobalt);
  }
  .composer[data-tone='setup'] .tone-dot {
    background: var(--md-cobalt);
  }
  .composer[data-tone='capped'] {
    border-color: color-mix(in oklab, var(--md-halt) 40%, transparent);
  }
  .composer[data-tone='capped'] .tone {
    color: var(--md-halt);
  }
  .composer[data-tone='capped'] .tone-dot {
    background: var(--md-halt);
  }
  .composer.focused,
  .composer.ready.focused {
    transform: none;
    border-color: color-mix(in oklab, var(--md-cobalt) 40%, var(--md-line));
    box-shadow: var(--md-focus);
    background: var(--md-surface);
  }
  .composer[data-tone='ready']:not(.focused) {
    box-shadow: none;
    border-color: color-mix(in oklab, var(--md-cobalt) 22%, var(--md-line));
  }
  .input {
    width: 100%;
    border: 0;
    background: transparent;
    color: var(--md-ink);
    font-size: 16px;
    line-height: 1.5;
    resize: none;
    min-height: 28px;
    max-height: 160px;
    outline: none;
    overflow-y: auto;
  }
  .input::placeholder {
    color: var(--md-ink-faint);
  }
  .bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--md-line);
  }
  .meta {
    display: flex;
    flex-direction: column;
    gap: 3px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.04em;
    color: var(--md-ink-faint);
    overflow: hidden;
    min-width: 0;
  }
  .sr {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
  .model {
    display: inline-flex;
    max-width: 100%;
  }
  .model select {
    appearance: none;
    max-width: min(280px, 52vw);
    border: 1px solid var(--md-line);
    background: var(--md-stage)
      url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%236B7A90' stroke-width='2'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E")
      no-repeat right 10px center;
    padding: 6px 28px 6px 10px;
    border-radius: 8px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-ink-soft);
    outline: none;
  }
  .model select:focus-visible {
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
  }
  .offline.warn {
    color: var(--md-halt);
  }
  .offline .short {
    display: none;
  }
  .hint {
    opacity: 0.85;
  }
  .actions {
    display: flex;
    gap: 8px;
    flex: none;
  }

  @keyframes md-rise {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  @keyframes md-caret {
    50% {
      opacity: 0;
    }
  }
  @keyframes md-tone-pulse {
    0%,
    100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(1.35);
      opacity: 0.65;
    }
  }

  @media (max-width: 720px) {
    .feed {
      padding: 20px 14px 12px;
    }
    .atlas-grid {
      grid-template-columns: 1fr;
    }
    .composer {
      width: calc(100% - 28px);
      margin-bottom: 96px;
    }
    .thread-bar {
      flex-direction: column;
      align-items: stretch;
    }
    .pipe li {
      padding: 6px 10px;
    }
  }
  @media (max-width: 420px) {
    h1 {
      font-size: clamp(32px, 11vw, 40px);
    }
    .offline .full {
      display: none;
    }
    .offline .short {
      display: inline;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .hero *,
    .msg,
    .composer,
    .bubble.streaming::after,
    .tone-dot,
    h1 span {
      animation: none !important;
    }
    .feed {
      scroll-behavior: auto;
    }
  }
</style>
