<!--
  Hub — Condura v2 Skills Hub surface.

  Per the spec: "A library with a card catalog. Skills are real book
  spines; browse = pull one out." and "Hover: spine tilts 4°."

  The composition:
    1. Search + filter at the top
    2. A horizontal row of installed skills (book spines, vertical,
       sorted by last-used)
    3. A vertical shelf of available skills from the public Hub
       (filtered by the search query)

  The spines use a real 3D tilt on hover via `transform: rotateY() + translateZ()`-style transforms:
    - at rest: spines are 64×160, paper-2 background, 1px groove at right
    - on hover: spine rotates 4° around its X-axis, color brightens,
      title emerges from the spine

  Pure-presentation: parent route owns the skill catalog data.

  Props:
    query:        string — current search query
    onQueryChange: (q: string) => void
    installed:    Skill[]
    available:    Skill[]
    onSelect?:    (id: string) => void  — fires when a user clicks a spine
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Eyebrow, Chip, Glyph, Avatar } from '$lib/v2'

  export interface Skill {
    id: string
    title: string
    author: string
    description: string
    version?: string
    tags?: string[]
    loaded?: boolean
    trust?: 'official' | 'community' | 'experimental'
  }

  let {
    query = '' as string,
    onQueryChange = undefined as ((q: string) => void) | undefined,
    installed = [] as Skill[],
    available = [] as Skill[],
    onSelect = undefined as ((id: string) => void) | undefined,
  }: {
    query?: string
    onQueryChange?: (q: string) => void
    installed?: Skill[]
    available?: Skill[]
    onSelect?: (id: string) => void
  } = $props()

  const filteredAvailable = $derived(
    query.trim().length === 0
      ? available
      : available.filter(s =>
          s.title.toLowerCase().includes(query.toLowerCase()) ||
          s.description.toLowerCase().includes(query.toLowerCase()) ||
          (s.tags ?? []).some(t => t.toLowerCase().includes(query.toLowerCase()))
        )
  )
</script>

<div data-v2 style="
  flex: 1;
  background: var(--v2-paper);
  overflow-y: auto;
  padding: var(--v2-space-12);
  box-sizing: border-box;
">
  <div style="max-width: 1100px; margin: 0 auto;">

    <Stack gap={12}>

      <!-- ── Title + search ─────────────────────────── -->
      <Stack gap={6}>
        <Stack gap={3}>
          <Eyebrow left="skills" right="hub" tone="accent" />
          <Ink kind="display" as="h1">The library.</Ink>
          <Ink kind="body-2" tone="ink-2" as="p" style:max-width="640px">
            Real book spines. Pull one out to read its description,
            install it, or load it into the active agent.
          </Ink>
        </Stack>

        <Surface elevation={0} padding="3" radius="2" tone="paper-2">
          <input
            type="text"
            value={query}
            oninput={(e) => onQueryChange?.((e.currentTarget as HTMLInputElement).value)}
            placeholder="search by name, description, or tag…"
            style="
              width: 100%;
              font-family: var(--v2-font-display);
              font-style: italic;
              font-size: var(--v2-text-20);
              color: var(--v2-ink);
              background: transparent;
              padding: var(--v2-space-2) var(--v2-space-3);
            "
          />
        </Surface>
      </Stack>

      <Rule />

      <!-- ── Installed skills (the front shelf) ──────── -->
      {#if installed.length > 0}
        <Stack gap={6}>
          <Stack gap={2}>
            <Eyebrow left="on the shelf" right="installed" tone="ink-3" />
            <Ink kind="title" as="h2">Loaded in the agent.</Ink>
          </Stack>

          <div style="
            display: flex;
            align-items: flex-end;
            gap: var(--v2-space-3);
            padding: var(--v2-space-8) var(--v2-space-4) var(--v2-space-6);
            background:
              linear-gradient(to bottom, var(--v2-paper-2), transparent),
              repeating-linear-gradient(to bottom, var(--v2-rule) 0 1px, transparent 1px 24px);
            border-radius: var(--v2-radius-2);
            min-height: 240px;
          ">
            {#each installed as skill (skill.id)}
              <button
                data-v2-spine
                data-loaded
                onclick={() => onSelect?.(skill.id)}
                aria-label={`Loaded skill: ${skill.title}`}
                title={skill.title}
              >
                <Stack gap={3} align="center">
                  <Ink kind="mono-cap" tone="ink-3" weight="medium">{skill.version ?? '1.0.0'}</Ink>
                  <span style="
                    writing-mode: vertical-rl;
                    transform: rotate(180deg);
                    font-family: var(--v2-font-display);
                    font-size: var(--v2-text-16);
                    font-style: italic;
                    color: var(--v2-ink);
                    line-height: 1.2;
                    max-height: 110px;
                    overflow: hidden;
                  ">{skill.title}</span>
                  <Ink kind="mono-cap" tone="ink-3" style:font-size="10px">condura</Ink>
                </Stack>
              </button>
            {/each}
          </div>
          <Ink kind="caption" tone="ink-3">click a spine to pull it out, or to unload.</Ink>
        </Stack>

        <Rule />
      {/if}

      <!-- ── Available skills (the back catalog) ─────── -->
      <Stack gap={6}>
        <Stack gap={2}>
          <Eyebrow left="browse" right="public hub" tone="ink-3" />
          <Ink kind="title" as="h2">
            {#if query.trim().length > 0}
              {filteredAvailable.length} match{filteredAvailable.length === 1 ? '' : 'es'} for "{query}"
            {:else}
              All available skills.
            {/if}
          </Ink>
        </Stack>

        {#if filteredAvailable.length === 0}
          <Surface elevation={0} padding="12" radius="2" tone="paper">
            <Stack gap={3} align="center">
              <Ink kind="display" as="h3" style:font-size="var(--v2-text-28)">No matches.</Ink>
              <Ink kind="body" tone="ink-2">Try a different word, or browse the full catalog.</Ink>
            </Stack>
          </Surface>
        {:else}
          <Stack gap={3}>
            {#each filteredAvailable as skill (skill.id)}
              <Surface
                elevation={0}
                padding="6"
                radius="2"
                tone="paper"
                interactive
                onclick={() => onSelect?.(skill.id)}
              >
                <Inline gap={4} align="start">
                  {#if skill.loaded}
                    <span aria-hidden="true" style="
                      align-self: stretch;
                      width: 3px;
                      background: var(--v2-accent);
                      border-radius: var(--v2-radius-1);
                    "></span>
                  {:else}
                    <span aria-hidden="true" style="
                      align-self: stretch;
                      width: 3px;
                      background: transparent;
                    "></span>
                  {/if}
                  <div style="flex: 1; min-width: 0;">
                    <Stack gap={2}>
                      <Inline gap={3} align="baseline">
                        <Ink kind="title" style:font-size="var(--v2-text-20)">{skill.title}</Ink>
                        <Ink kind="mono-cap" tone="ink-3">v{skill.version ?? '1.0.0'}</Ink>
                        <Chip on={false} size="small">{skill.trust ?? 'community'}</Chip>
                      </Inline>
                      <Ink kind="body" tone="ink-2">
                        {skill.description}
                      </Ink>
                      {#if skill.tags && skill.tags.length > 0}
                        <Inline gap={2}>
                          {#each skill.tags as tag}
                            <span style="
                              font-family: var(--v2-font-mono);
                              font-size: 11px;
                              color: var(--v2-ink-3);
                              padding: 2px var(--v2-space-2);
                              border-radius: var(--v2-radius-1);
                              border: 1px solid var(--v2-rule);
                            ">#{tag}</span>
                          {/each}
                        </Inline>
                      {/if}
                    </Stack>
                  </div>
                  <div style="flex-shrink: 0;">
                    {#if skill.loaded}
                      <Inline gap={2} align="center">
                        <Glyph name="check" size={14} />
                        <Ink kind="ui-small" tone="signal-go" weight="medium">Loaded</Ink>
                      </Inline>
                    {:else}
                      <button
                        data-v2
                        onclick={(e) => { e.stopPropagation(); /* install */ }}
                        style="
                          font-family: var(--v2-font-sans);
                          font-size: var(--v2-text-12);
                          font-weight: 500;
                          padding: var(--v2-space-2) var(--v2-space-3);
                          background: transparent;
                          color: var(--v2-ink);
                          border: 1px solid var(--v2-rule);
                          border-radius: var(--v2-radius-pill);
                          cursor: pointer;
                          transition: all var(--v2-dur-fast) var(--v2-ease-out-soft);
                        "
                      >install</button>
                    {/if}
                  </div>
                </Inline>
              </Surface>
            {/each}
          </Stack>
        {/if}
      </Stack>

    </Stack>

  </div>
</div>

<style>
  /* Book spines — at rest, a 64×160 vertical paper card with a
     1px groove at the right edge. On hover, rotate 4° around the
     X-axis and rise on the Y-axis, lift to v2-shadow-2. This is
     the "alive" detail per spec. */
  [data-v2-spine] {
    width: 64px;
    height: 160px;
    background: var(--v2-paper);
    border: 1px solid var(--v2-rule);
    border-radius: var(--v2-radius-1) var(--v2-radius-1) 0 0;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--v2-space-3) 0;
    transform-origin: bottom center;
    transition:
      transform  var(--v2-dur-mid) var(--v2-ease-spring),
      box-shadow var(--v2-dur-mid) var(--v2-ease-settle);
    box-shadow:
      inset -1px 0 0 var(--v2-rule),                  /* right-edge groove */
      0 1px 0 rgba(27, 26, 23, 0.04);
    flex-shrink: 0;
  }
  [data-v2-spine]:hover {
    transform: perspective(400px) rotateX(-4deg) translateY(-6px);
    box-shadow:
      inset -1px 0 0 var(--v2-rule),
      var(--v2-shadow-2);
  }
  [data-v2-spine][data-loaded] {
    /* A loaded spine has a tiny accent ribbon at the top edge,
       like a tag on a real library book. */
    border-top: 3px solid var(--v2-accent);
  }
</style>
