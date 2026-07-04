<!--
  V2InterviewDemo — Condura v2 floating interview preview.

  Mounts the FloatingInterview component over a "fake app background"
  so the user can preview the mandatory first-time panel in isolation.
  When the user finishes, the interview's answers are echoed in the
  top of the screen so they can see the data flow end-to-end.

  Navigation: served at #/dev/v2-interview after a 6-line wire-up
  documented in `app/web/frontend/src/lib/v2/README.md`.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, FloatingInterview, type InterviewAnswers } from '$lib/v2'

  let answers = $state<InterviewAnswers | null>(null)
  let startOverTrigger = $state(0)   // bump this to reset the interview

  function onComplete(a: InterviewAnswers) {
    answers = a
  }
  function skip() {
    answers = {
      name: 'guest',
      power: 'cloud-when-needed',
      never: [],
      dayVision: '',
      hotkey: '',
    }
  }
  function restart() {
    answers = null
    startOverTrigger++
  }
</script>

<div data-v2 style="
  min-height: 100vh;
  background: var(--v2-paper);
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
">
  <!-- ── Fake app "chrome" behind the interview ───────────── -->
  <div style="
    flex: 1;
    display: grid;
    grid-template-columns: 72px 1fr;
  ">
    <!-- Sidebar mock (just shape, not the real v2 Sidebar yet) -->
    <div style="
      background: var(--v2-paper-2);
      border-right: 1px solid color-mix(in srgb, var(--v2-rule) 60%, transparent);
      display: flex; flex-direction: column;
      align-items: center;
      padding: var(--v2-space-4) 0;
      gap: var(--v2-space-3);
    ">
      {#each ['Ch', 'St', 'Au', 'Co', 'De', 'Hu', 'Re', 'Sy', 'Sk', 'Ab'] as monogram}
        <div style="
          width: 36px; height: 36px;
          border-radius: var(--v2-radius-1);
          background: transparent;
          display: grid; place-items: center;
          font-family: var(--v2-font-display);
          font-size: var(--v2-text-12);
          color: var(--v2-ink-3);
          letter-spacing: 0.04em;
        ">{monogram}</div>
      {/each}
    </div>

    <!-- Main area mock -->
    <div style="padding: var(--v2-space-12); overflow: auto;">
      <Stack gap={6}>

        <Inline gap={2} align="center">
          <Ink kind="mono-cap" tone="accent">first-run preview</Ink>
          <Rule orientation="horizontal" weight="1" tone="ink-3" inset="0" />
          <Ink kind="mono-cap" tone="ink-3">condura · v2</Ink>
        </Inline>

        {#if !answers}
          <!-- Pre-interview: explain what the user is about to see -->
          <Surface elevation={1} padding="12" radius="3" tone="paper">
            <Stack gap={6}>
              <Stack gap={3}>
                <Ink kind="display" as="h1">
                  Hello.
                </Ink>
                <Ink kind="body-2" tone="ink-2" as="p" style:max-width="640px">
                  I'm Condura — a free, OS-native agent that lives on your
                  computer and acts as the conductor of every AI tool
                  installed here. Before I start, I'd like to know you.
                </Ink>
              </Stack>

              <Rule />

              <Stack gap={3}>
                <Ink kind="caption" tone="ink-3">what I'll ask</Ink>
                <Stack gap={2}>
                  {#each [
                      ['1', 'What should I call you?'],
                      ['2', 'How do you want me to answer?'],
                      ['3', 'Anything you want me to never touch?'],
                      ['4', 'What does a great day with me look like?'],
                      ['5', 'Pick a hotkey.']
                    ] as [n, q]}
                    <Inline gap={3} align="baseline">
                      <Ink kind="mono-cap" tone="accent" style:min-width="14px">{n}</Ink>
                      <Ink kind="body" tone="ink-2">{q}</Ink>
                    </Inline>
                  {/each}
                </Stack>
              </Stack>
            </Stack>
          </Surface>
        {:else}
          <!-- Post-interview: confirm what was captured -->
          <Surface elevation={2} padding="12" radius="3" tone="paper">
            <Stack gap={6}>
              <Stack gap={3}>
                <Ink kind="mono-cap" tone="accent">the agent has arrived</Ink>
                <Ink kind="display" as="h1">
                  Nice to meet you, {answers.name}.
                </Ink>
                <Ink kind="body-2" tone="ink-2">
                  You're set up the way you asked. Change anything any time
                  from the sidebar.
                </Ink>
              </Stack>

              <Rule />

              <Stack gap={4}>
                <Ink kind="caption" tone="ink-3">what I learned</Ink>

                <Stack gap={3}>
                  <Inline gap={6} align="baseline">
                    <Ink kind="caption" tone="ink-3" style:min-width="100px">name</Ink>
                    <Ink kind="body" weight="medium">{answers.name}</Ink>
                  </Inline>
                  <Rule orientation="horizontal" weight="1" tone="ink-3" inset="0" />
                  <Inline gap={6} align="baseline">
                    <Ink kind="caption" tone="ink-3" style:min-width="100px">power</Ink>
                    <Ink kind="body" weight="medium">
                      {answers.power === 'local' ? 'Local only'
                       : answers.power === 'always-cloud' ? 'Always cloud'
                       : 'Cloud when needed'}
                    </Ink>
                  </Inline>
                  <Rule orientation="horizontal" weight="1" tone="ink-3" inset="0" />
                  <Inline gap={6} align="baseline">
                    <Ink kind="caption" tone="ink-3" style:min-width="100px">never touch</Ink>
                    <Inline gap={2}>
                      {#each answers.never.length ? answers.never : [] as d}
                        <Surface elevation={0} padding="1" radius="pill" tone="paper-2">
                          <Ink kind="ui-small" tone="ink-2">{d}</Ink>
                        </Surface>
                      {/each}
                      {#if answers.never.length === 0}
                        <Ink kind="body" tone="ink-3" italic>nothing protected (allowed by default)</Ink>
                      {/if}
                    </Inline>
                  </Inline>
                  <Rule orientation="horizontal" weight="1" tone="ink-3" inset="0" />
                  <Inline gap={6} align="baseline">
                    <Ink kind="caption" tone="ink-3" style:min-width="100px">day vision</Ink>
                    <Ink kind="body" tone={answers.dayVision ? 'ink' : 'ink-3'} italic={!answers.dayVision}>
                      {answers.dayVision || '— not set'}
                    </Ink>
                  </Inline>
                  <Rule orientation="horizontal" weight="1" tone="ink-3" inset="0" />
                  <Inline gap={6} align="baseline">
                    <Ink kind="caption" tone="ink-3" style:min-width="100px">hotkey</Ink>
                    {#if answers.hotkey}
                      <Ink kind="mono" tone="accent" weight="medium">{answers.hotkey}</Ink>
                    {:else}
                      <Ink kind="body" tone="ink-3" italic>— not set</Ink>
                    {/if}
                  </Inline>
                </Stack>
              </Stack>

              <Rule />

              <Inline gap={3}>
                <button
                  data-v2
                  onclick={restart}
                  style="
                    font-family: var(--v2-font-sans);
                    font-size: var(--v2-text-14);
                    font-weight: 500;
                    color: var(--v2-ink);
                    background: transparent;
                    border: 1px solid var(--v2-rule);
                    padding: var(--v2-space-3) var(--v2-space-4);
                    border-radius: var(--v2-radius-1);
                    cursor: pointer;
                    transition: all var(--v2-dur-fast) var(--v2-ease-out-soft);
                  "
                >Start over</button>
              </Inline>
            </Stack>
          </Surface>
        {/if}

      </Stack>
    </div>
  </div>

  <!-- ── The floating interview panel ──────────────────────── -->
  {#if !answers}
    {#key startOverTrigger}
      <FloatingInterview
        onComplete={onComplete}
        onSkip={skip}
      />
    {/key}
  {/if}
</div>
