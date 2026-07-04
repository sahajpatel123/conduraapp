<!--
  Ink — Condura v2 text primitive.

  The single most-used primitive. All text in v2 routes through
  Ink with a `kind` prop that selects a typographic role.

  Kinds:
    display    Instrument Serif, used for hero text and section heads
    title      Instrument Serif, smaller section heads
    body       Inter, the default
    body-2     Inter, slightly larger, for lede paragraphs
    ui         Inter with tighter tracking and tabular numerals
    ui-small   Inter 12 — sidebar labels and small UI text
    caption    Inter 11 — the lowest level of text, very quiet
    mono       JetBrains Mono, used for timers, IDs, hashes
    mono-cap   JetBrains Mono with all-caps tracking

  `tone` lets you shift to ink-2 / ink-3 / accent / accent-ink /
  paper without changing kind. Never override color in consumers;
  always via tone.

  `as` lets you pick the underlying HTML element (default: 'span'
  for inline kinds, 'p' for body, 'h1/h2/h3' for display/title).
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLAttributes } from 'svelte/elements'

  type Kind =
    | 'display' | 'title' | 'body' | 'body-2' | 'ui'
    | 'ui-small' | 'caption' | 'mono' | 'mono-cap'
  type Tone = 'ink' | 'ink-2' | 'ink-3' | 'accent' | 'accent-ink' | 'paper' | 'signal-go' | 'signal-warn' | 'signal-stop'
  type Weight = 'regular' | 'medium' | 'semibold'

  let {
    kind = 'body' as Kind,
    tone = 'ink' as Tone,
    weight = undefined as Weight | undefined,
    as = undefined as string | undefined,
    italic = false,
    children,
    class: klass = '',
    ...rest
  }: {
    kind?: Kind
    tone?: Tone
    weight?: Weight
    as?: string
    italic?: boolean
    children?: Snippet
    class?: string
  } & HTMLAttributes<HTMLElement> = $props()

  // HTML element selection — semantic where reasonable, no surprises
  const elFor = (k: Kind): string => {
    switch (k) {
      case 'display':   return as ?? 'h1'
      case 'title':     return as ?? 'h2'
      case 'body':
      case 'body-2':    return as ?? 'p'
      case 'mono':
      case 'mono-cap':  return as ?? 'span'
      default:          return as ?? 'span'
    }
  }

  const tag = elFor(kind)

  const colorFor = (t: Tone): string => {
    switch (t) {
      case 'ink':         return 'var(--v2-ink)'
      case 'ink-2':       return 'var(--v2-ink-2)'
      case 'ink-3':       return 'var(--v2-ink-3)'
      case 'accent':      return 'var(--v2-accent)'
      case 'accent-ink':  return 'var(--v2-accent-ink)'
      case 'paper':       return 'var(--v2-paper)'
      case 'signal-go':   return 'var(--v2-signal-go)'
      case 'signal-warn': return 'var(--v2-signal-warn)'
      case 'signal-stop': return 'var(--v2-signal-stop)'
    }
  }

  const weightFor = (w: Weight | undefined): number => {
    if (w === 'medium') return 500
    if (w === 'semibold') return 600
    return 400
  }
</script>

<svelte:element
  this={tag}
  data-v2-ink
  data-kind={kind}
  data-tone={tone}
  class={klass}
  style:color={colorFor(tone)}
  style:font-weight={weightFor(weight)}
  style:font-style={italic ? 'italic' : 'normal'}
  {...rest}
>
  {@render children?.()}
</svelte:element>

<style>
  /* display & title — Instrument Serif, expressive */
  [data-v2-ink][data-kind='display'] {
    font-family: var(--v2-font-display);
    font-size: var(--v2-text-40);
    line-height: var(--v2-leading-tight);
    letter-spacing: var(--v2-tracking-tighter);
  }
  [data-v2-ink][data-kind='title'] {
    font-family: var(--v2-font-display);
    font-size: var(--v2-text-28);
    line-height: var(--v2-leading-snug);
    letter-spacing: var(--v2-tracking-tight);
  }

  /* body — Inter, default for prose */
  [data-v2-ink][data-kind='body'] {
    font-family: var(--v2-font-sans);
    font-size: var(--v2-text-16);
    line-height: var(--v2-leading-default);
    font-feature-settings: var(--v2-font-features);
  }
  [data-v2-ink][data-kind='body-2'] {
    font-family: var(--v2-font-sans);
    font-size: var(--v2-text-20);
    line-height: var(--v2-leading-default);
    letter-spacing: var(--v2-tracking-tight);
    font-feature-settings: var(--v2-font-features);
  }

  /* ui — Inter for chrome, tabular numerals */
  [data-v2-ink][data-kind='ui'] {
    font-family: var(--v2-font-sans);
    font-size: var(--v2-text-14);
    line-height: var(--v2-leading-snug);
    font-feature-settings: var(--v2-numeric-features);
  }
  [data-v2-ink][data-kind='ui-small'] {
    font-family: var(--v2-font-sans);
    font-size: var(--v2-text-12);
    line-height: var(--v2-leading-snug);
    font-feature-settings: var(--v2-numeric-features);
  }
  [data-v2-ink][data-kind='caption'] {
    font-family: var(--v2-font-sans);
    font-size: var(--v2-text-11);
    line-height: var(--v2-leading-snug);
    letter-spacing: var(--v2-tracking-wide);
    text-transform: uppercase;
  }

  /* mono — JetBrains Mono for status, code, IDs */
  [data-v2-ink][data-kind='mono'] {
    font-family: var(--v2-font-mono);
    font-size: var(--v2-text-14);
    line-height: var(--v2-leading-snug);
    font-feature-settings: var(--v2-numeric-features);
  }
  [data-v2-ink][data-kind='mono-cap'] {
    font-family: var(--v2-font-mono);
    font-size: var(--v2-text-12);
    line-height: var(--v2-leading-snug);
    letter-spacing: var(--v2-tracking-wider);
    text-transform: uppercase;
    font-feature-settings: var(--v2-numeric-features);
  }
</style>
