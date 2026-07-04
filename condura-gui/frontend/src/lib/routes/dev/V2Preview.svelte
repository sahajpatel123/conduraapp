<!--
  V2Preview — Condura v2 design system showcase.

  This page is the proof that the foundation works. It is also
  intentionally the most carefully composed thing in the v2
  system: every primitive, in its most characteristic setting.
  If this page does not feel premium, nothing else will.

  Navigation: served at #/dev/v2-preview (wire-up is one line of
  import + two-line route branch in App.svelte). The user can
  add that wire-up themselves or ask for it.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import {
    Surface, Ink, Stack, Inline, Rule, Button,
  } from '$lib/v2'

  // Live ticker — the heartbeat demo
  let now = $state(new Date())
  $effect(() => {
    const id = setInterval(() => { now = new Date() }, 1000)
    return () => clearInterval(id)
  })
  const time = $derived(
    now.toLocaleTimeString('en-US', { hour12: false })
  )

  // First-time floating interviewer demo state
  let stepIndex = $state(0)
  const steps = [
    { q: "What should I call you?",         kind: 'text' },
    { q: "How do you want me to answer?",   kind: 'cards', options: ['Local only', 'Cloud when needed', 'Always cloud'] },
    { q: "Anything you want me to never touch?", kind: 'chips', options: ['email', 'money', 'calendar', 'files', 'code'] },
    { q: "Pick a hotkey",                   kind: 'hotkey' },
  ] as const
  const step = $derived(steps[stepIndex])

  let answer = $state('')
  let selectedChips = $state<string[]>([])

  function advance() {
    if (stepIndex < steps.length - 1) {
      stepIndex++
      answer = ''
      selectedChips = []
    }
  }
  function back() {
    if (stepIndex > 0) stepIndex--
  }

  // Palette swatches for the demo grid
  const palette = [
    { name: 'paper',     value: '#F7F4EE', role: 'canvas' },
    { name: 'paper-2',   value: '#EFEAE0', role: 'recessed' },
    { name: 'surface',   value: '#FFFFFF', role: 'elevated' },
    { name: 'ink',       value: '#1B1A17', role: 'primary text' },
    { name: 'ink-2',     value: '#4A463E', role: 'secondary text' },
    { name: 'ink-3',     value: '#8A847A', role: 'tertiary / captions' },
    { name: 'rule',      value: '#D9D2C2', role: 'hairlines' },
    { name: 'accent',    value: '#C18A4A', role: 'the one color' },
    { name: 'accent-ink',value: '#7A4F1E', role: 'accent text' },
    { name: 'signal-go', value: '#5C7F4A', role: 'success' },
    { name: 'signal-warn',value:'#B07A2E', role: 'cautious' },
    { name: 'signal-stop',value:'#A84A3F', role: 'destructive' },
  ]
</script>

<!-- Full-page v2 root -->
<div data-v2 style="
  min-height: 100vh;
  background: var(--v2-paper);
  padding: var(--v2-space-16) var(--v2-space-12);
  box-sizing: border-box;
">
  <div style="max-width: 1200px; margin: 0 auto;">

    <!-- ── Hero ─────────────────────────────────────────── -->
    <Stack gap={6} align="stretch">

      <Eyebrow left="v2 design system" right="foundation preview" tone="ink-3" />

      <Surface elevation={0} padding="8" radius="3" tone="paper">
        <Stack gap={6}>
          <Ink kind="display" as="h1">
            A quiet companion that gets out of the way.
          </Ink>
          <Ink kind="body-2" tone="ink-2" as="p" style:max-width="640px">
            This page is the foundation of the redesigned Condura GUI.
            It demonstrates six primitives, fifteen color tokens,
            three type families, and the motion grammar that ties them
            together. Nothing here is decorative.
          </Ink>
          <Inline gap={3}>
            <Button variant="primary">Begin a session</Button>
            <Button variant="ghost">Read the spec</Button>
          </Inline>
        </Stack>
      </Surface>

      <!-- ── The heartbeat (motion demo) ───────────────── -->
      <Surface elevation={0} padding="6" radius="2" tone="paper-2">
        <Inline gap={6} align="center">
          <div
            style="
              width: 8px; height: 8px;
              border-radius: var(--v2-radius-pill);
              background: var(--v2-accent);
              animation: v2-heartbeat 1s var(--v2-ease-linear) infinite;
            "
          ></div>
          <Stack gap={1}>
            <Ink kind="mono-cap" tone="ink-3">agent heartbeat · 1Hz</Ink>
            <Ink kind="mono" tone="ink">{time}</Ink>
          </Stack>
          <div style="flex: 1"></div>
          <Ink kind="caption" tone="ink-3">motion = acknowledgment</Ink>
        </Inline>
      </Surface>

    </Stack>

    <Rule orientation="horizontal" weight="1" tone="ink-3" inset="var(--v2-space-12) 0" />

    <!-- ── Typography scale ────────────────────────────── -->
    <Stack gap={8}>

      <Stack gap={3}>
        <Ink kind="mono-cap" tone="accent">01 · typography</Ink>
        <Ink kind="title" as="h2">Type is the visual.</Ink>
        <Ink kind="body" tone="ink-2">
          Three families. Eight sizes. Every word is
          doing work — chrome fades, words carry.
        </Ink>
      </Stack>

      <Surface elevation={0} padding="8" radius="2" tone="surface">
        <Stack gap={4}>
          <Stack gap={2}>
            <Ink kind="caption" tone="ink-3">display · instrument serif · 40px</Ink>
            <Ink kind="display">A quiet companion.</Ink>
          </Stack>
          <Rule />
          <Stack gap={2}>
            <Ink kind="caption" tone="ink-3">title · instrument serif · 28px</Ink>
            <Ink kind="title">Welcome to the colophon.</Ink>
          </Stack>
          <Rule />
          <Stack gap={2}>
            <Ink kind="caption" tone="ink-3">body-2 · inter · 20px</Ink>
            <Ink kind="body-2">
              Ledes open a story. They're confident without urgency.
            </Ink>
          </Stack>
          <Rule />
          <Stack gap={2}>
            <Ink kind="caption" tone="ink-3">body · inter · 16px</Ink>
            <Ink kind="body">
              The default for prose. Warm line height (1.55), with
              <Ink kind="ui" tone="ink">tabular numerals</Ink>
              and
              <Ink kind="ui" tone="accent">accent pullouts</Ink>
              built in.
            </Ink>
          </Stack>
          <Rule />
          <Stack gap={2}>
            <Ink kind="caption" tone="ink-3">ui · inter · 14px</Ink>
            <Ink kind="ui">
              Sidebar labels, table cells, navigation.
              Uses tnum and zero for jingle-free counts.
            </Ink>
          </Stack>
          <Rule />
          <Stack gap={2}>
            <Ink kind="caption" tone="ink-3">mono · jetbrains mono · 14px</Ink>
            <Ink kind="mono">
              a4f3-92bc · 01:23:42.7 · 8.430 s · $0.0014
            </Ink>
          </Stack>
          <Rule />
          <Stack gap={2}>
            <Ink kind="caption" tone="ink-3">caption · inter · 11px</Ink>
            <Ink kind="caption">a11y label · status hint · section eyebrow</Ink>
          </Stack>
        </Stack>
      </Surface>

      <!-- ── Palette grid ───────────────────────────────── -->
      <Stack gap={3}>
        <Ink kind="mono-cap" tone="accent">02 · palette</Ink>
        <Ink kind="title" as="h2">Fifteen tokens. That's it.</Ink>
        <Ink kind="body" tone="ink-2">
          Warm earth-amber on paper-white. No glass, no neon,
          no gradient-everywhere.
        </Ink>
      </Stack>

      <div style="
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
        gap: var(--v2-space-3);
      ">
        {#each palette as p}
          <Surface elevation={0} padding="4" radius="2" tone="surface">
            <Stack gap={3}>
              <div style="
                height: 64px;
                border-radius: var(--v2-radius-1);
                background: {p.value};
                border: 1px solid color-mix(in srgb, var(--v2-rule) 50%, transparent);
              "></div>
              <Stack gap={1}>
                <Ink kind="ui" weight="medium">{p.name}</Ink>
                <Ink kind="mono" tone="ink-3">{p.value}</Ink>
                <Ink kind="caption" tone="ink-3">{p.role}</Ink>
              </Stack>
            </Stack>
          </Surface>
        {/each}
      </div>

      <!-- ── Elevation ──────────────────────────────────── -->
      <Stack gap={3}>
        <Ink kind="mono-cap" tone="accent">03 · elevation</Ink>
        <Ink kind="title" as="h2">Paper stacked in real shadow.</Ink>
        <Ink kind="body" tone="ink-2">
          Four elevations, all real. No glass-faking,
          no translucent haze.
        </Ink>
      </Stack>

      <div style="
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
        gap: var(--v2-space-6);
      ">
        {#each [{ elev: 0, label: '0 · flat', note: 'hairline only' },
                { elev: 1, label: '1 · pressed', note: 'sharp, near-zero blur' },
                { elev: 2, label: '2 · lift',    note: 'cards on hover' },
                { elev: 3, label: '3 · peak',    note: 'overlays & consent' }] as e}
          <Surface elevation={e.elev} padding="8" radius="2" tone="surface">
            <Stack gap={2}>
              <Ink kind="caption" tone="ink-3">{e.note}</Ink>
              <Ink kind="title">{e.label}</Ink>
            </Stack>
          </Surface>
        {/each}
      </div>

      <!-- ── Buttons ─────────────────────────────────────── -->
      <Stack gap={3}>
        <Ink kind="mono-cap" tone="accent">04 · buttons</Ink>
        <Ink kind="title" as="h2">Hardware-honest.</Ink>
        <Ink kind="body" tone="ink-2">
          The press has to feel real. Not a flat rectangle.
        </Ink>
      </Stack>

      <Surface elevation={0} padding="8" radius="2" tone="surface">
        <Stack gap={6}>
          <Inline gap={3} align="center">
            <Button variant="primary">Allow once</Button>
            <Button variant="primary">Allow for session</Button>
            <Button variant="ghost">Read details</Button>
            <Button variant="deny">Deny</Button>
            <Button variant="primary" disabled>Disabled</Button>
          </Inline>
          <Rule />
          <Inline gap={3} align="center">
            <Button variant="primary" size="small">Small primary</Button>
            <Button variant="ghost"   size="small">Small ghost</Button>
            <Button variant="deny"    size="small">Small deny</Button>
          </Inline>
        </Stack>
      </Surface>

    </Stack>

    <Rule orientation="horizontal" weight="1" tone="ink-3" inset="var(--v2-space-12) 0" />

    <!-- ── Mandatory first-time floating panel demo ────── -->
    <Stack gap={6}>
      <Stack gap={3}>
        <Ink kind="mono-cap" tone="accent">05 · the first-time interview</Ink>
        <Ink kind="title" as="h2">Not a wizard. A floating panel.</Ink>
        <Ink kind="body" tone="ink-2">
          One question at a time, anchored bottom-right, with the
          real app visible behind. Step through it:
        </Ink>
      </Stack>

      <Surface
        elevation={3}
        padding="8"
        radius="3"
        tone="paper"
        interactive
        style:max-width="640px"
      >
        <Stack gap={6}>

          <Inline gap={2} align="center">
            <Ink kind="mono-cap" tone="accent">interview · step {stepIndex + 1} of {steps.length}</Ink>
            <div style="flex: 1"></div>
            <Inline gap={1}>
              {#each steps as _, i}
                <div style="
                  width: 6px; height: 6px;
                  border-radius: var(--v2-radius-pill);
                  background: {i <= stepIndex ? 'var(--v2-accent)' : 'var(--v2-rule)'};
                "></div>
              {/each}
            </Inline>
          </Inline>

          <Ink kind="title">{step.q}</Ink>

          {#if step.kind === 'text'}
            <Surface elevation={0} padding="4" radius="1" tone="paper-2">
              <input
                type="text"
                placeholder="Type your answer…"
                bind:value={answer}
                style="
                  width: 100%;
                  font-family: var(--v2-font-display);
                  font-size: var(--v2-text-20);
                  background: transparent;
                  color: var(--v2-ink);
                  padding: var(--v2-space-2) 0;
                  font-style: italic;
                "
              />
            </Surface>
          {/if}

          {#if step.kind === 'cards'}
            <Stack gap={3}>
              {#each step.options as opt}
                <Surface
                  elevation={0}
                  padding="6"
                  radius="2"
                  tone="paper-2"
                  interactive
                >
                  <Ink kind="body" weight="medium">{opt}</Ink>
                </Surface>
              {/each}
            </Stack>
          {/if}

          {#if step.kind === 'chips'}
            <Inline gap={2}>
              {#each step.options as opt}
                {@const isOn = selectedChips.includes(opt)}
                <button
                  data-v2
                  onclick={() => {
                    if (isOn) selectedChips = selectedChips.filter(c => c !== opt)
                    else selectedChips = [...selectedChips, opt]
                  }}
                  style="
                    font-family: var(--v2-font-sans);
                    font-size: var(--v2-text-12);
                    font-weight: 500;
                    padding: var(--v2-space-2) var(--v2-space-3);
                    border-radius: var(--v2-radius-pill);
                    border: 1px solid {isOn ? 'var(--v2-accent)' : 'var(--v2-rule)'};
                    background: {isOn ? 'color-mix(in srgb, var(--v2-accent) 14%, transparent)' : 'transparent'};
                    color: var(--v2-ink);
                    cursor: pointer;
                    transition: all var(--v2-dur-fast) var(--v2-ease-out-soft);
                  "
                >{opt}</button>
              {/each}
            </Inline>
          {/if}

          {#if step.kind === 'hotkey'}
            <Surface
              elevation={0}
              padding="8"
              radius="2"
              tone="paper-2"
            >
              <Stack gap={2} align="center">
                <Ink kind="mono" tone="ink-3">press a key combination</Ink>
                <Ink kind="mono" weight="medium" tone="accent" style:font-size="var(--v2-text-20)">
                  ⌘  ⇧  Space
                </Ink>
              </Stack>
            </Surface>
          {/if}

          <Inline gap={2} justify="end">
            {#if stepIndex > 0}
              <Button variant="ghost" onclick={back}>Back</Button>
            {/if}
            {#if stepIndex < steps.length - 1}
              <Button variant="primary" onclick={advance}>Next</Button>
            {:else}
              <Button variant="primary">Finish</Button>
            {/if}
          </Inline>

        </Stack>
      </Surface>
    </Stack>

    <Rule orientation="horizontal" weight="1" tone="ink-3" inset="var(--v2-space-12) 0" />

    <!-- ── Closing ─────────────────────────────────────── -->
    <Stack gap={4}>
      <Ink kind="mono-cap" tone="ink-3">end of foundation preview</Ink>
      <Ink kind="body" tone="ink-2">
        If this feels like a $50M product, the foundation is ready.
        Next iteration builds the chat surface, the overlay arrival,
        and wires this into the rest of the app via
        <Ink kind="mono" tone="ink">$lib/v2</Ink>.
      </Ink>
    </Stack>

  </div>
</div>
