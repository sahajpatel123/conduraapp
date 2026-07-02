<!--
  FloatingInterview — Condura v2 first-time personalization panel.

  The soul of the new design. Five questions, one floating panel,
  anchored bottom-right. The real app shell is visible behind;
  this isn't a wizard, it's a conversation.

  Steps:
    1. Name           — single text field
    2. Power source   — 3 cards (Local / Cloud when needed / Always cloud)
    3. Never touch    — chips (email, money, calendar, files, code)
    4. Day vision     — freeform, optional
    5. Hotkey         — key combo capture

  On finish: emits `onComplete` with the full answer record. The
  parent route is responsible for persistence (typically to the
  daemon via `onboarding.*` RPCs or to `adaptive.profile`).

  Props:
    onComplete?: (answers: InterviewAnswers) => void
    skipable?: boolean   — allow "Skip for now" footer when true
    onSkip?: () => void
    initialAnswers?: Partial<InterviewAnswers>
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Button, Chip } from '$lib/v2'

  type Power = 'local' | 'cloud-when-needed' | 'always-cloud'
  type Domain = 'email' | 'money' | 'calendar' | 'files' | 'code'

  export interface InterviewAnswers {
    name: string
    power: Power
    never: Domain[]
    dayVision: string
    hotkey: string
  }

  let {
    onComplete = undefined as ((answers: InterviewAnswers) => void) | undefined,
    skipable = true,
    onSkip = undefined as (() => void) | undefined,
    initialAnswers = {} as Partial<InterviewAnswers>,
  }: {
    onComplete?: (a: InterviewAnswers) => void
    skipable?: boolean
    onSkip?: () => void
    initialAnswers?: Partial<InterviewAnswers>
  } = $props()

  // Step + answers — single source of truth
  let stepIndex = $state(0)
  let name = $state(initialAnswers.name ?? '')
  let power = $state<Power | null>(initialAnswers.power ?? null)
  let never = $state<Domain[]>(initialAnswers.never ?? [])
  let dayVision = $state(initialAnswers.dayVision ?? '')
  let hotkey = $state(initialAnswers.hotkey ?? '')

  const DOMAINS: Domain[] = ['email', 'money', 'calendar', 'files', 'code']
  const POWERS: { id: Power; label: string; copy: string }[] = [
    { id: 'local',             label: 'Local only',         copy: 'Runs on your machine. No data leaves unless you ask.' },
    { id: 'cloud-when-needed', label: 'Cloud when needed',  copy: 'Tries local first. Falls back to cloud for hard tasks.' },
    { id: 'always-cloud',      label: 'Always cloud',       copy: 'Best models, faster answers. You pay for what you use.' },
  ]

  // The five step IDs, in order.
  const STEPS = ['name', 'power', 'never', 'day', 'hotkey'] as const
  type StepId = (typeof STEPS)[number]

  // Per-step validation — Next is disabled until satisfied.
  // Hotkey MUST include a modifier (locked decision #8 — "user must
  // set on first run; no default"). A solo "A" would fire in every
  // text field; it is not a usable hotkey.
  function canAdvance(): boolean {
    const s = STEPS[stepIndex]
    if (s === 'name')  return name.trim().length > 0
    if (s === 'power') return power !== null
    if (s === 'never') return true                              // optional
    if (s === 'day')   return true                              // optional
    if (s === 'hotkey') {
      const h = hotkey.trim()
      if (h.length === 0) return false
      // Must contain at least one modifier: ⌘ / ⌃ / ⌥ / ⇧
      return /[⌘⌃⌥⇧]/.test(h)
    }
    return false
  }

  function next() {
    if (!canAdvance()) return
    if (stepIndex < STEPS.length - 1) {
      stepIndex++
    } else {
      finish()
    }
  }

  function back() {
    if (stepIndex > 0) stepIndex--
  }

  function finish() {
    const answers: InterviewAnswers = {
      name: name.trim(),
      power: power ?? 'cloud-when-needed',
      never: [...never],
      dayVision: dayVision.trim(),
      hotkey: hotkey.trim(),
    }
    onComplete?.(answers)
  }

  // Keyboard nav — Enter advances, Esc backs. CRITICAL: bound to the
  // panel root (not <svelte:window>) so it does not intercept keys
  // typed elsewhere on the page. If focus is inside a textarea/input
  // (day-vision step, name field) we let the key do its native thing.
  // Also guards against double-firing while hotkey capture is active.
  function onKeydown(e: KeyboardEvent) {
    if (capturingHotkey) return
    const t = e.target as HTMLElement | null
    const inField = t && (t.tagName === 'TEXTAREA' || t.tagName === 'INPUT')
    if (inField) {
      // Inside the day-vision textarea: newline on Enter is the user's
      // right; we only honor Esc-to-back here as a navigation escape.
      if (e.key === 'Escape') {
        e.preventDefault()
        back()
      }
      return
    }
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      next()
    } else if (e.key === 'Escape') {
      e.preventDefault()
      back()
    }
  }

  // Hotkey capture — listens at the panel root
  let capturingHotkey = $state(false)
  function startCapture() {
    capturingHotkey = true
    hotkey = ''
  }
  function onHotkeyKeydown(e: KeyboardEvent) {
    if (!capturingHotkey) return
    e.preventDefault()
    if (['Control', 'Shift', 'Alt', 'Meta'].includes(e.key)) return
    const parts: string[] = []
    if (e.metaKey)     parts.push('⌘')
    if (e.ctrlKey)     parts.push('⌃')
    if (e.altKey)      parts.push('⌥')
    if (e.shiftKey)    parts.push('⇧')
    parts.push(e.key === ' ' ? 'Space' : e.key.toUpperCase())
    hotkey = parts.join(' ')
    capturingHotkey = false
  }

  // Skip function (available when skipable)
  function skip() {
    onSkip?.()
  }

  // The current step (derived)
  const stepId = $derived<StepId>(STEPS[stepIndex])

  // Display the captured "name" in the corner dot to confirm state
  const displayName = $derived(name.trim().split(/\s+/)[0] || '')
</script>

<!--
  Root panel. `tabindex={-1}` makes the panel a key event target
  without putting it in the tab order — combined with
  `onkeydown={onKeydown}`, this scopes Enter/Esc nav to events that
  happen INSIDE the panel (bubble phase). The window-level hotkey
  capture listener is separate; both check `capturingHotkey` so they
  do not double-fire.
-->
<svelte:window onkeydown={onHotkeyKeydown} />

<div
  data-v2
  tabindex={-1}
  onkeydown={onKeydown}
  style="
    position: fixed;
    bottom: var(--v2-space-8);
    right: var(--v2-space-8);
    width: 480px;
    max-width: calc(100vw - var(--v2-space-12));
    z-index: var(--v2-z-overlay);
    animation: v2-slide-up var(--v2-dur-slow) var(--v2-ease-settle) both;
  "
>
  <Surface
    elevation={3}
    padding="8"
    radius="3"
    tone="paper"
  >
    <Stack gap={6}>

      <!-- Step header — eyebrow + progress dots -->
      <Inline gap={3} align="center">
        <Ink kind="mono-cap" tone="accent">interview</Ink>
        <Rule orientation="horizontal" weight="1" tone="ink-3" inset="0" />
        <Ink kind="mono-cap" tone="ink-3">
          step {stepIndex + 1} / {STEPS.length}
        </Ink>
        <div style="flex: 1"></div>
        <Inline gap={1}>
          {#each STEPS as _, i}
            <div style="
              width: 6px; height: 6px;
              border-radius: var(--v2-radius-pill);
              background: {i <= stepIndex ? 'var(--v2-accent)' : 'var(--v2-rule)'};
              transition: background var(--v2-dur-fast) var(--v2-ease-out-soft);
            "></div>
          {/each}
        </Inline>
      </Inline>

      <!-- Committed answers summary — quiet pills visible after first step -->
      {#if stepIndex > 0 && displayName}
        <Inline gap={2}>
          <Surface elevation={0} padding="1" radius="pill" tone="paper-2">
            <Ink kind="ui-small" tone="ink-2" weight="medium">{displayName}</Ink>
          </Surface>
          {#if power}
            <Surface elevation={0} padding="1" radius="pill" tone="paper-2">
              <Ink kind="ui-small" tone="ink-2">
                {POWERS.find(p => p.id === power)?.label}
              </Ink>
            </Surface>
          {/if}
          {#if never.length}
            <Surface elevation={0} padding="1" radius="pill" tone="paper-2">
              <Ink kind="ui-small" tone="ink-2">{never.length} protected</Ink>
            </Surface>
          {/if}
        </Inline>
      {/if}

      <!-- ── Step: name ─────────────────────────────────────── -->
      {#if stepId === 'name'}
        <Stack gap={4}>
          <Ink kind="title" as="h2">What should I call you?</Ink>
          <Ink kind="body" tone="ink-2">
            This is how the agent greets you. You can change it later.
          </Ink>
          <Surface elevation={0} padding="4" radius="1" tone="paper-2">
            <input
              type="text"
              autofocus
              placeholder="Your name"
              bind:value={name}
              style="
                width: 100%;
                font-family: var(--v2-font-display);
                font-size: var(--v2-text-28);
                font-style: italic;
                color: var(--v2-ink);
                padding: var(--v2-space-2) 0;
              "
            />
          </Surface>
        </Stack>

      <!-- ── Step: power ────────────────────────────────────── -->
      {:else if stepId === 'power'}
        <Stack gap={4}>
          <Ink kind="title" as="h2">How should I answer?</Ink>
          <Ink kind="body" tone="ink-2">
            You can mix and match any time. This just sets the starting point.
          </Ink>
          <Stack gap={3}>
            {#each POWERS as p}
              {@const on = power === p.id}
              <button
                data-v2
                onclick={() => { power = p.id }}
                style="
                  all: unset;
                  cursor: pointer;
                  display: block;
                  width: 100%;
                  border-radius: var(--v2-radius-2);
                "
              >
                <Surface
                  elevation={on ? 2 : 0}
                  padding="6"
                  radius="2"
                  tone={on ? 'paper-2' : 'paper-2'}
                  style:border={on
                    ? '1px solid var(--v2-accent)'
                    : '1px solid color-mix(in srgb, var(--v2-rule) 50%, transparent)'}
                >
                  <Inline gap={4} align="start">
                    <div style="
                      width: 18px; height: 18px;
                      border-radius: var(--v2-radius-pill);
                      border: 1px solid {on ? 'var(--v2-accent)' : 'var(--v2-rule)'};
                      background: {on ? 'var(--v2-accent)' : 'transparent'};
                      display: grid; place-items: center;
                      margin-top: 2px;
                      transition: all var(--v2-dur-fast) var(--v2-ease-out-soft);
                    ">
                      {#if on}
                        <div style="
                          width: 6px; height: 6px;
                          border-radius: var(--v2-radius-pill);
                          background: var(--v2-paper);
                        "></div>
                      {/if}
                    </div>
                    <Stack gap={1}>
                      <Ink kind="body" weight="medium">{p.label}</Ink>
                      <Ink kind="ui-small" tone="ink-3">{p.copy}</Ink>
                    </Stack>
                  </Inline>
                </Surface>
              </button>
            {/each}
          </Stack>
        </Stack>

      <!-- ── Step: never touch ──────────────────────────────── -->
      {:else if stepId === 'never'}
        <Stack gap={4}>
          <Ink kind="title" as="h2">Anything you want me to never touch?</Ink>
          <Ink kind="body" tone="ink-2">
            Pick any. The Gatekeeper will block every action in these
            areas until you explicitly allow it.
          </Ink>
          <Inline gap={2}>
            {#each DOMAINS as d}
              {@const on = never.includes(d)}
              <Chip
                {on}
                variant="accent"
                onclick={() => {
                  if (on) never = never.filter(x => x !== d)
                  else never = [...never, d]
                }}
              >{d}</Chip>
            {/each}
          </Inline>
          <Ink kind="caption" tone="ink-3">
            Skip this step to allow everything by default.
          </Ink>
        </Stack>

      <!-- ── Step: day vision ───────────────────────────────── -->
      {:else if stepId === 'day'}
        <Stack gap={4}>
          <Ink kind="title" as="h2">What does a great day with me look like?</Ink>
          <Ink kind="body" tone="ink-2">
            Optional. The agent uses this to shape its suggestions.
          </Ink>
          <Surface elevation={0} padding="4" radius="1" tone="paper-2">
            <textarea
              bind:value={dayVision}
              rows="4"
              placeholder="e.g. mornings are for deep work, afternoons for meetings…"
              style="
                width: 100%;
                font-family: var(--v2-font-sans);
                font-size: var(--v2-text-14);
                line-height: var(--v2-leading-default);
                color: var(--v2-ink);
                background: transparent;
                resize: vertical;
                min-height: 80px;
              "
            ></textarea>
          </Surface>
        </Stack>

      <!-- ── Step: hotkey ───────────────────────────────────── -->
      {:else if stepId === 'hotkey'}
        <Stack gap={4}>
          <Ink kind="title" as="h2">Pick a hotkey.</Ink>
          <Ink kind="body" tone="ink-2">
            This is the combo that summons the overlay. Choose
            something you won't hit by accident.
          </Ink>
          <Surface
            elevation={0}
            padding="8"
            radius="2"
            tone="paper-2"
            interactive
            onclick={startCapture}
            onkeydown={onHotkeyKeydown}
          >
            <Stack gap={2} align="center">
              <Ink kind="mono-cap" tone="ink-3">
                {capturingHotkey ? 'press a key combo…' : 'click to capture'}
              </Ink>
              {#if hotkey}
                <Ink kind="mono" tone="accent" weight="medium" style:font-size="var(--v2-text-28)">
                  {hotkey}
                </Ink>
              {:else if capturingHotkey}
                <div style="
                  width: 12px; height: 12px;
                  border-radius: var(--v2-radius-pill);
                  background: var(--v2-accent);
                  animation: v2-heartbeat 1s var(--v2-ease-linear) infinite;
                "></div>
              {/if}
            </Stack>
          </Surface>
          <Ink kind="caption" tone="ink-3">
            Suggestions: ⌥⌥ · ⌘⇧Space · ⌃Space
          </Ink>
        </Stack>

      {/if}

      <!-- Footer actions -->
      <Inline gap={2} justify="between" align="center">
        <Inline gap={2}>
          {#if stepIndex > 0}
            <Button variant="ghost" onclick={back}>Back</Button>
          {/if}
          {#if skipable && stepIndex < STEPS.length - 1}
            <Button variant="ghost" onclick={skip}>Skip for now</Button>
          {/if}
        </Inline>
        <Button variant="primary" onclick={next} disabled={!canAdvance()}>
          {stepIndex === STEPS.length - 1 ? 'Finish' : 'Next'}
        </Button>
      </Inline>

      <Ink kind="caption" tone="ink-3" style:text-align="center">
        Enter to continue · Esc to go back
      </Ink>

    </Stack>
  </Surface>
</div>
