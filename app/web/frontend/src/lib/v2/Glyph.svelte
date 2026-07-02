<!--
  Glyph — Condura v2 single-stroke SVG icon vocabulary.

  The spec requires "Single-stroke SVG icons drawn at 1.5px for our
  vocabulary." Every glyph is hand-authored as a 24×24 SVG with a
  1.5px stroke and `stroke-linecap="round"` / `stroke-linejoin="round"`,
  so the whole system speaks one dialect of line.

  Why not font-icons / emoji / unicode: emoji and Unicode glyphs are
  the most-viewed tell of "vibe coded" chrome — `●` and `‹` are
  2003-era system-tray residue. Linear/Vercel ship hand-authored 1.5px
  SVG stroke icons for a reason: the line-weight and terminations
  match the rest of the chrome.

  Add new names by writing one more `case` below. Never inline an
  SVG path anywhere else in the system; always use this.

  Props:
    name:     Identity name (see `glyphPath` below).
    size?:    Pixel size (default 16 — small chrome default).
    stroke?:  CSS color (default `currentColor`).
    ariaLabel?: string — required for non-decorative glyphs.
-->
<script lang="ts">
  let {
    name = '' as 'dot' | 'dot-active' | 'chevron-left' | 'chevron-right' | 'check' | 'x' | 'arrow-right' | 'plus' | 'minus' | 'paperclip' | 'send' | 'mic' | 'book' | 'eye' | 'shield' | 'gear' | 'sparkle',
    size = 16 as number,
    ariaLabel = undefined as string | undefined,
    class: klass = '',
  }: {
    name: 'dot' | 'dot-active' | 'chevron-left' | 'chevron-right' | 'check' | 'x' | 'arrow-right' | 'plus' | 'minus' | 'paperclip' | 'send' | 'mic' | 'book' | 'eye' | 'shield' | 'gear' | 'sparkle'
    size?: number
    ariaLabel?: string
    class?: string
  } = $props()

  // Each glyph is a 24×24 viewBox with stroke="currentColor" so it
  // inherits color from the consumer. Filled shapes use fill.
  function path(): { type: 'stroke' | 'fill', d: string } {
    switch (name) {
      case 'dot':          return { type: 'fill',   d: 'M12 10.5a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3z' }
      case 'dot-active':   return { type: 'fill',   d: 'M12 9.5a2.5 2.5 0 1 1 0 5 2.5 2.5 0 0 1 0-5z' }
      case 'chevron-left': return { type: 'stroke', d: 'M15 5l-7 7 7 7' }
      case 'chevron-right':return { type: 'stroke', d: 'M9 5l7 7-7 7' }
      case 'check':        return { type: 'stroke', d: 'M5 12.5l4.5 4.5L19 7.5' }
      case 'x':            return { type: 'stroke', d: 'M6 6l12 12M18 6L6 18' }
      case 'arrow-right':  return { type: 'stroke', d: 'M5 12h14M13 6l6 6-6 6' }
      case 'plus':         return { type: 'stroke', d: 'M12 5v14M5 12h14' }
      case 'minus':        return { type: 'stroke', d: 'M5 12h14' }
      case 'paperclip':    return { type: 'stroke', d: 'M16 8l-7.5 7.5a3 3 0 0 0 4.2 4.2l9-9a4.5 4.5 0 0 0-6.4-6.4L6.3 13.3' }
      case 'send':         return { type: 'stroke', d: 'M5 12l14-7-4 14-3-6-7-1z' }
      case 'mic':          return { type: 'stroke', d: 'M12 4a3 3 0 0 0-3 3v5a3 3 0 0 0 6 0V7a3 3 0 0 0-3-3zM6 11a6 6 0 0 0 12 0M12 17v3' }
      case 'book':         return { type: 'stroke', d: 'M5 5a2 2 0 0 1 2-2h11v18H7a2 2 0 0 1-2-2V5zM5 16h13' }
      case 'eye':          return { type: 'stroke', d: 'M2.5 12C5 7 8.5 5 12 5s7 2 9.5 7c-2.5 5-6 7-9.5 7S5 17 2.5 12zM12 9.5a2.5 2.5 0 1 1 0 5 2.5 2.5 0 0 1 0-5z' }
      case 'shield':       return { type: 'stroke', d: 'M12 3l8 3v6c0 5-3.5 8-8 9-4.5-1-8-4-8-9V6l8-3z' }
      case 'gear':         return { type: 'stroke', d: 'M12 9.5a2.5 2.5 0 1 1 0 5 2.5 2.5 0 0 1 0-5zM19 12l2-1-2-3-2 1-1.5-1 .5-2-4-1-.5 2-2 0-1-2-3 2 .5 2-1 1.5-2 1L4 14l2 1 1 1.5L6 19l3 2 1-1.5 1.5-1 2 .5 1 4-2 .5L12 19l2 0 1-2 3 2 2-4-1.5-1L19 12z' }
      case 'sparkle':      return { type: 'stroke', d: 'M12 4v4M12 16v4M4 12h4M16 12h4M6 6l3 3M15 15l3 3M6 18l3-3M15 9l3-3' }
    }
  }

  const g = $derived(path())
  const isFilled = $derived(g.type === 'fill')
</script>

<svg
  data-v2-glyph={name}
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill={isFilled ? 'currentColor' : 'none'}
  stroke={isFilled ? 'none' : 'currentColor'}
  stroke-width="1.5"
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-label={ariaLabel}
  aria-hidden={ariaLabel ? undefined : 'true'}
  role={ariaLabel ? 'img' : undefined}
  class={klass}
  style:display="inline-block"
  style:vertical-align="middle"
  style:flex-shrink="0"
>
  <path d={g.d} />
</svg>
