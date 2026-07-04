<!--
  Eyebrow — Condura v2 chapter-marker pattern.

  Renders a horizontal eyebrow row: a small mono-cap label, a
  flexible hairline, and a small mono-cap label, all anchored on
  the same baseline. Used as section markers in long-form content
  (the spec calls these "numbered chapter headings" in Settings).

  Replaces the inline-styled Inline+Rule+Ink pattern that was
  duplicated across four demo files.

  Props:
    left:        string  — left label
    right?:      string  — right label
    tone?:       'ink-3' | 'accent'
-->
<script lang="ts">
  import { Ink, Rule } from '$lib/v2'

  let {
    left = '' as string,
    right = undefined as string | undefined,
    tone = 'ink-3' as 'ink-3' | 'accent',
    class: klass = '',
  }: {
    left: string
    right?: string
    tone?: 'ink-3' | 'accent'
    class?: string
  } = $props()
</script>

<div data-v2-eyebrow data-tone={tone} class={klass}>
  <Ink kind="mono-cap" tone={tone}>{left}</Ink>
  <Rule orientation="horizontal" weight="1" tone={tone === 'accent' ? 'accent' : 'ink-3'} inset="0" />
  {#if right}
    <Ink kind="mono-cap" tone={tone}>{right}</Ink>
  {/if}
</div>

<style>
  [data-v2-eyebrow] {
    display: flex;
    align-items: center;
    gap: var(--v2-space-3);
    flex-wrap: wrap;
  }
  [data-v2-eyebrow] > :global([data-v2-rule]) {
    flex: 1;
    min-width: 24px;
  }
</style>
