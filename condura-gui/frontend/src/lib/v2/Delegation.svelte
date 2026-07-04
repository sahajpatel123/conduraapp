<!--
  Delegation — Condura v2 sub-agent control-room surface.

  Per the spec: "A control room. Each sub-agent is a 'station' with a
  real-time waveform of its activity."

  Each station shows:
    - Sub-agent name + adapter (e.g. "claude code · claude-sonnet-4.5")
    - Activity waveform (real-time bar visualization)
    - Last output / current task
    - Status (idle / running / stopped / error)
    - Spawn / cancel actions

  Props:
    agents: Agent[]
    onSpawn?: (template: string) => void
    onCancel?: (id: string) => void
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Eyebrow, Glyph, Button, Avatar } from '$lib/v2'

  type Adapter = 'claude_code' | 'codex' | 'antigravity' | 'opencode' | 'kilo' | 'hermes' | 'gemini' | 'ollama'

  type Status = 'idle' | 'running' | 'stopped' | 'error'

  export interface Agent {
    id: string
    name: string
    adapter: Adapter
    model: string
    status: Status
    task?: string
    startedAt?: number   // ms epoch
    durationMs?: number
    output?: string
  }

  let {
    agents = [] as Agent[],
    onSpawn = undefined as ((template: string) => void) | undefined,
    onCancel = undefined as ((id: string) => void) | undefined,
  }: {
    agents?: Agent[]
    onSpawn?: (template: string) => void
    onCancel?: (id: string) => void
  } = $props()

  // Live ticker so each running agent's waveform animates every frame.
  let tick = $state(0)
  $effect(() => {
    const id = setInterval(() => { tick = (tick + 1) % 1000 }, 100)
    return () => clearInterval(id)
  })

  function adapterLabel(a: Adapter): string {
    switch (a) {
      case 'claude_code':  return 'claude code'
      case 'codex':       return 'codex'
      case 'antigravity': return 'antigravity'
      case 'opencode':     return 'opencode'
      case 'kilo':         return 'kilo'
      case 'hermes':       return 'hermes'
      case 'gemini':       return 'gemini'
      case 'ollama':       return 'ollama'
    }
  }
</script>

<div data-v2 style="
  flex: 1;
  background: var(--v2-paper);
  overflow-y: auto;
  padding: var(--v2-space-12);
  box-sizing: border-box;
">
  <div style="max-width: 1100px; margin: 0 auto;">

    <Stack gap={10}>

      <Inline gap={4} justify="between" align="end">
        <Stack gap={3}>
          <Eyebrow left="delegation" right="sub-agents" tone="accent" />
          <Ink kind="display" as="h1">The control room.</Ink>
          <Ink kind="body-2" tone="ink-2" as="p" style:max-width="640px">
            Spawn sub-agents on your installed CLIs. Each one runs
            as a real subprocess; you can cancel and steer here.
          </Ink>
        </Stack>
        <Inline gap={2}>
          <Button variant="ghost" size="small" onclick={() => onSpawn?.('review')}>+ code review</Button>
          <Button variant="ghost" size="small" onclick={() => onSpawn?.('tests')}>+ run tests</Button>
          <Button variant="primary" size="small" onclick={() => onSpawn?.('research')}>+ research</Button>
        </Inline>
      </Inline>

      <Rule />

      <Stack gap={3}>
        {#if agents.length === 0}
          <Surface elevation={0} padding="12" radius="2" tone="paper">
            <Stack gap={3} align="center">
              <Ink kind="display" as="h3" style:font-size="var(--v2-text-28)">No sub-agents running.</Ink>
              <Ink kind="body" tone="ink-2">Spawn one to delegate work in parallel.</Ink>
            </Stack>
          </Surface>
        {/if}

        {#each agents as a (a.id)}
          <Surface elevation={0} padding="6" radius="2" tone="paper">
            <Inline gap={4} align="start" justify="between">
              <Inline gap={3} align="start">
                <Avatar
                  role={a.status === 'running' ? 'agent' : 'system'}
                  size={32}
                  monogram={a.name.charAt(0).toUpperCase()}
                />
                <Stack gap={1}>
                  <Inline gap={2} align="baseline">
                    <Ink kind="ui" weight="medium">{a.name}</Ink>
                    <span style="
                      font-family: var(--v2-font-mono);
                      font-size: 11px;
                      color: var(--v2-ink-3);
                      padding: 1px 6px;
                      border-radius: var(--v2-radius-1);
                      border: 1px solid var(--v2-rule);
                    ">{adapterLabel(a.adapter)} · {a.model}</span>
                  </Inline>
                  {#if a.task}
                    <Ink kind="ui-small" tone="ink-2">{a.task}</Ink>
                  {/if}
                </Stack>
              </Inline>

              <!-- Live waveform. Real-time bars driven by the tick state. -->
              <div style="
                flex: 1;
                height: 32px;
                display: flex;
                align-items: flex-end;
                gap: 2px;
                overflow: hidden;
              " aria-label="activity waveform">
                {#each Array.from({ length: 60 }) as _, i}
                  {@const phase = (i + tick * 0.6) % 30}
                  {@const live = a.status === 'running'}
                  <span style="
                    width: 3px;
                    height: {live
                      ? `${30 + Math.sin((i + tick) * 0.4) * 8 + Math.cos(i * 0.7) * 10}%`
                      : `${20 + Math.sin(phase * 0.3) * 5}%`};
                    background: {a.status === 'running'
                      ? 'var(--v2-accent)'
                      : a.status === 'error'
                        ? 'var(--v2-signal-stop)'
                        : a.status === 'stopped'
                          ? 'var(--v2-ink-3)'
                          : 'var(--v2-rule)'};
                    border-radius: 1px;
                    opacity: {live ? 1 : 0.6};
                    transition: height 100ms var(--v2-ease-linear);
                  "></span>
                {/each}
              </div>

              <!-- Status + duration -->
              <Stack gap={2} align="end">
                <Inline gap={2} align="center">
                  {#if a.status === 'running'}
                    <span style="
                      width: 8px; height: 8px;
                      border-radius: var(--v2-radius-pill);
                      background: var(--v2-accent);
                      animation: v2-heartbeat 1s linear infinite;
                    "></span>
                    <Ink kind="ui-small" tone="ink-2" weight="medium">running · {Math.floor((a.durationMs ?? 0) / 1000)}s</Ink>
                  {:else if a.status === 'error'}
                    <Ink kind="ui-small" tone="signal-stop" weight="medium">error</Ink>
                  {:else if a.status === 'stopped'}
                    <Ink kind="ui-small" tone="ink-3">stopped</Ink>
                  {:else}
                    <Ink kind="ui-small" tone="ink-3">idle</Ink>
                  {/if}
                </Inline>

                {#if a.status === 'running'}
                  <Button variant="deny" size="small" onclick={() => onCancel?.(a.id)}>Stop</Button>
                {/if}
              </Stack>
            </Inline>

            {#if a.output}
              <div style="
                margin-top: var(--v2-space-4);
                padding-top: var(--v2-space-4);
                border-top: 1px solid var(--v2-rule);
              ">
                <pre style="
                  font-family: var(--v2-font-mono);
                  font-size: var(--v2-text-12);
                  color: var(--v2-ink-2);
                  white-space: pre-wrap;
                  word-break: break-word;
                  max-height: 160px;
                  overflow: auto;
                  margin: 0;
                  padding: var(--v2-space-3);
                  background: var(--v2-paper-2);
                  border-radius: var(--v2-radius-1);
                ">{a.output}</pre>
              </div>
            {/if}
          </Surface>
        {/each}
      </Stack>

    </Stack>

  </div>
</div>
