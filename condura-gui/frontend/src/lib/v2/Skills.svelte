<!--
  Skills — Condura v2 local-skills library surface.

  Per the spec: "Your local library — same card-catalog model as
  Hub, but with a 'loaded' dot on each. Loaded skills have a
  faint red-warm ribbon at the top."

  This is the local mirror of Hub. Skills appear here after
  they are installed from Hub and live in `~/.synaptic/skills/`.

  Props:
    skills: Skill[]
    onActivate?: (id: string) => void
    onDeactivate?: (id: string) => void
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Eyebrow, Glyph, Button, Chip } from '$lib/v2'

  export interface LocalSkill {
    id: string
    title: string
    author: string
    description: string
    version: string
    tags?: string[]
    active: boolean    // loaded into the active agent vs. just installed
  }

  let {
    skills = [] as LocalSkill[],
    onActivate = undefined as ((id: string) => void) | undefined,
    onDeactivate = undefined as ((id: string) => void) | undefined,
  }: {
    skills?: LocalSkill[]
    onActivate?: (id: string) => void
    onDeactivate?: (id: string) => void
  } = $props()
</script>

<div data-v2 style="
  flex: 1;
  background: var(--v2-paper);
  overflow-y: auto;
  padding: var(--v2-space-12);
  box-sizing: border-box;
">
  <div style="max-width: 880px; margin: 0 auto;">

    <Stack gap={10}>

      <Stack gap={3}>
        <Eyebrow left="skills" right="local library" tone="ink-3" />
        <Ink kind="display" as="h1">Skills installed here.</Ink>
        <Ink kind="body-2" tone="ink-2" as="p" style:max-width="640px">
          Skills that live on this machine — installed from the Hub,
          ready to load into the agent at any time.
        </Ink>
      </Stack>

      <Rule />

      <Stack gap={3}>
        {#if skills.length === 0}
          <Surface elevation={0} padding="12" radius="2" tone="paper">
            <Stack gap={3} align="center">
              <Ink kind="display" as="h3" style:font-size="var(--v2-text-28)">Library empty.</Ink>
              <Ink kind="body" tone="ink-2">Install skills from the public Hub to populate it.</Ink>
            </Stack>
          </Surface>
        {/if}

        {#each skills as s (s.id)}
          <Surface
            elevation={0}
            padding="0"
            radius="2"
            tone="paper"
            style:overflow="hidden"
            style:position="relative"
          >
            <!-- The active ribbon — a faint accent-red at the top edge,
                 reads as "this one is loaded into the agent right now." -->
            {#if s.active}
              <span style="
                position: absolute;
                inset: 0 0 auto 0;
                height: 3px;
                background: linear-gradient(to right, var(--v2-accent), color-mix(in srgb, var(--v2-accent) 40%, transparent));
              "></span>
            {/if}

            <div style="padding: var(--v2-space-6);">
              <Stack gap={3}>
                <Inline gap={4} justify="between" align="start">
                  <Stack gap={2} style:flex="1">
                    <Inline gap={2} align="baseline">
                      <Ink kind="title" style:font-size="var(--v2-text-20)">{s.title}</Ink>
                      <Ink kind="mono-cap" tone="ink-3">v{s.version}</Ink>
                      {#if s.active}
                        <Chip on size="small" variant="signal-go">Active</Chip>
                      {:else}
                        <Chip on={false} size="small">Installed</Chip>
                      {/if}
                    </Inline>
                    <Ink kind="ui-small" tone="ink-2">{s.description}</Ink>
                    {#if s.tags && s.tags.length > 0}
                      <Inline gap={2}>
                        {#each s.tags as t}
                          <span style="
                            font-family: var(--v2-font-mono);
                            font-size: 11px;
                            color: var(--v2-ink-3);
                            padding: 2px 6px;
                            border-radius: var(--v2-radius-1);
                            border: 1px solid var(--v2-rule);
                          ">#{t}</span>
                        {/each}
                      </Inline>
                    {/if}
                  </Stack>

                  {#if s.active}
                    <Button variant="ghost" size="small" onclick={() => onDeactivate?.(s.id)}>Unload</Button>
                  {:else}
                    <Button variant="primary" size="small" onclick={() => onActivate?.(s.id)}>Load</Button>
                  {/if}
                </Inline>
              </Stack>
            </div>
          </Surface>
        {/each}
      </Stack>

    </Stack>

  </div>
</div>
