<!--
  Surface — Condura v2 paper-surface primitive.

  The single most important primitive in v2. Everything rests on it
  (cards, panels, sheets, modals). It carries:

    - elevation (0 = flat / hairline, 1 = sharp "pressed", 2 = lift, 3 = peak)
    - tone (paper / paper-2 / surface / ink)
    - padding (none / 4 / 6 / 8 / 12)
    - radius (0 / 1 / 2 / 3 / pill)

  Surfaces are paper-honest: they don't glow, they don't blur the
  background, they don't fake translucency. They are pieces of paper,
  stacked in real shadow.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  type Elevation = 0 | 1 | 2 | 3
  type Tone = 'paper' | 'paper-2' | 'surface' | 'ink'
  type Padding = 'none' | '4' | '6' | '8' | '12'
  type Radius = '0' | '1' | '2' | '3' | 'pill'

  let {
    elevation = 0 as Elevation,
    tone = 'paper' as Tone,
    padding = '6' as Padding,
    radius = '2' as Radius,
    bleed = false,                 // remove padding (for surfaces wrapping full-bleed media)
    interactive = false,           // adds hover lift + focus ring
    onclick = undefined as ((e: MouseEvent) => void) | undefined,
    children,
    class: klass = '',
    ...rest
  } = $props()

  // Map elevation to actual shadow token
  const shadowFor = (e: Elevation): string => {
    switch (e) {
      case 1: return 'var(--v2-shadow-1)'
      case 2: return 'var(--v2-shadow-2)'
      case 3: return 'var(--v2-shadow-3)'
      default: return 'none'
    }
  }

  // Map tone to background + text color
  const bgFor = (t: Tone): string => {
    switch (t) {
      case 'paper':   return 'var(--v2-paper)'
      case 'paper-2': return 'var(--v2-paper-2)'
      case 'surface': return 'var(--v2-surface)'
      case 'ink':     return 'var(--v2-ink)'
    }
  }

  const fgFor = (t: Tone): string =>
    t === 'ink' ? 'var(--v2-paper)' : 'var(--v2-ink)'

  let hovered = $state(false)
  let focused = $state(false)
</script>

<div
  data-v2-surface
  data-elevation={elevation}
  data-tone={tone}
  data-interactive={interactive}
  data-radius={radius}
  class={klass}
  class:hovered={interactive && hovered}
  class:focused={interactive && focused}
  onclick={onclick}
  role={onclick ? 'button' : undefined}
  tabindex={onclick ? 0 : undefined}
  onkeydown={onclick ? (e: KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      onclick(new MouseEvent('click'))
    }
  } : undefined}
  onmouseenter={() => { if (interactive) hovered = true }}
  onmouseleave={() => { if (interactive) hovered = false }}
  onfocus={() => { if (interactive) focused = true }}
  onblur={() => { if (interactive) focused = false }}
  style:background={bgFor(tone)}
  style:color={fgFor(tone)}
  style:padding={bleed ? '0' : `var(--v2-space-${padding})`}
  style:border-radius={`var(--v2-radius-${radius})`}
  style:box-shadow={shadowFor(elevation)}
  style:transition="box-shadow var(--v2-dur-mid) var(--v2-ease-settle), transform var(--v2-dur-fast) var(--v2-ease-out-soft)"
  {...rest}
>
  {@render children()}
</div>

<style>
  [data-v2-surface] {
    position: relative;
    /* Hairline border — only on paper/surface tones, not ink */
  }
  [data-v2-surface][data-tone='paper'],
  [data-v2-surface][data-tone='paper-2'],
  [data-v2-surface][data-tone='surface'] {
    border: 1px solid color-mix(in srgb, var(--v2-rule) 50%, transparent);
  }

  /* Interactive surfaces — gentle hover lift, focus ring */
  [data-v2-surface][data-interactive='true'] {
    cursor: pointer;
    user-select: none;
  }
  [data-v2-surface][data-interactive='true'].hovered {
    transform: translateY(-1px);
    box-shadow: var(--v2-shadow-2) !important;
  }
  [data-v2-surface][data-interactive='true'].focused {
    box-shadow: var(--v2-focus-ring) !important;
  }
</style>
