<script lang="ts">
  import './living-paper.css'

  type Kind = 'hero' | 'display' | 'headline' | 'title' | 'body' | 'body-sm' | 'caption' | 'mono' | 'micro' | 'eyebrow'
  type Tone = 'ink' | 'ink-soft' | 'ink-mute' | 'ink-faint' | 'ink-ghost' | 'synapse' | 'pollen' | 'danger' | 'ok'

  interface Props {
    kind?: Kind
    tone?: Tone
    as?: 'h1' | 'h2' | 'h3' | 'h4' | 'p' | 'span' | 'div' | 'label'
    weight?: string
    italic?: boolean
    class?: string
    children?: import('svelte').Snippet
    style?: string
  }

  let {
    kind = 'body',
    tone = 'ink',
    as: tag = 'p',
    weight,
    italic = false,
    class: className = '',
    children,
    style = '',
  }: Props = $props()

  const fontMap: Record<Kind, string> = {
    hero:     'var(--lp-font-display)',
    display:  'var(--lp-font-display)',
    headline: 'var(--lp-font-display)',
    title:    'var(--lp-font-display)',
    body:     'var(--lp-font-sans)',
    'body-sm':'var(--lp-font-sans)',
    caption:  'var(--lp-font-sans)',
    mono:     'var(--lp-font-mono)',
    micro:    'var(--lp-font-mono)',
    eyebrow:  'var(--lp-font-mono)',
  }

  const sizeMap: Record<Kind, string> = {
    hero:     'var(--lp-text-hero)',
    display:  'var(--lp-text-display)',
    headline: 'var(--lp-text-headline)',
    title:    'var(--lp-text-title)',
    body:     'var(--lp-text-body)',
    'body-sm':'var(--lp-text-body-sm)',
    caption:  'var(--lp-text-caption)',
    mono:     'var(--lp-text-mono)',
    micro:    'var(--lp-text-micro)',
    eyebrow:  'var(--lp-text-micro)',
  }

  const leadingMap: Record<Kind, string> = {
    hero:     'var(--lp-leading-tight)',
    display:  'var(--lp-leading-tight)',
    headline: 'var(--lp-leading-snug)',
    title:    'var(--lp-leading-snug)',
    body:     'var(--lp-leading-normal)',
    'body-sm':'var(--lp-leading-normal)',
    caption:  'var(--lp-leading-normal)',
    mono:     'var(--lp-leading-normal)',
    micro:    'var(--lp-leading-snug)',
    eyebrow:  'var(--lp-leading-snug)',
  }

  const trackingMap: Record<Kind, string> = {
    hero:     'var(--lp-tracking-tight)',
    display:  'var(--lp-tracking-tight)',
    headline: 'var(--lp-tracking-tight)',
    title:    'var(--lp-tracking-normal)',
    body:     'var(--lp-tracking-normal)',
    'body-sm':'var(--lp-tracking-normal)',
    caption:  'var(--lp-tracking-normal)',
    mono:     'var(--lp-tracking-mono)',
    micro:    'var(--lp-tracking-mono)',
    eyebrow:  'var(--lp-tracking-wide)',
  }

  const colorMap: Record<Tone, string> = {
    ink:       'var(--lp-ink)',
    'ink-soft':'var(--lp-ink-soft)',
    'ink-mute':'var(--lp-ink-mute)',
    'ink-faint':'var(--lp-ink-faint)',
    'ink-ghost':'var(--lp-ink-ghost)',
    synapse:   'var(--lp-synapse)',
    pollen:    'var(--lp-pollen)',
    danger:    'var(--lp-danger)',
    ok:        'var(--lp-ok)',
  }
</script>

<svelte:element
  this={tag}
  class="lp {className}"
  style="
    font-family: {fontMap[kind]};
    font-size: {sizeMap[kind]};
    line-height: {leadingMap[kind]};
    letter-spacing: {trackingMap[kind]};
    color: {colorMap[tone]};
    font-weight: {weight || (kind === 'eyebrow' ? '500' : kind === 'hero' || kind === 'display' ? '500' : '400')};
    font-style: {italic ? 'italic' : 'normal'};
    text-transform: {kind === 'eyebrow' ? 'uppercase' : 'none'};
    margin: 0;
    font-variation-settings: normal;
    {style}
  "
>
  {#if children}{@render children()}{/if}
</svelte:element>
