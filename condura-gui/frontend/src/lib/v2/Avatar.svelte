<!--
  Avatar — Condura v2 monogrammed role badge.

  Three roles (user / agent / system), each with its own palette,
  monogram letter, and aria-label. Replaces the inline-styled
  div-with-letter code that previously lived in ChatSurface and
  which the UI-engineer review flagged as duplicate-and-prone-to-
  drift.

  Props:
    role: 'user' | 'agent' | 'system'
    size?: pixel size (default 28)
    monogram?: override letter (default 'U' / 'C' / 'S')
-->
<script lang="ts">
  let {
    role = 'agent' as 'user' | 'agent' | 'system',
    size = 28 as number,
    monogram = undefined as string | undefined,
    class: klass = '',
  }: {
    role?: 'user' | 'agent' | 'system'
    size?: number
    monogram?: string
    class?: string
  } = $props()

  const defaultLetters: Record<string, string> = { user: 'U', agent: 'C', system: 'S' }
  const labels: Record<string, string> = {
    user:   'User message',
    agent:  'Agent message',
    system: 'System message',
  }
  const letter = $derived(monogram ?? defaultLetters[role])
</script>

<div
  data-v2-avatar
  data-role={role}
  role="img"
  aria-label={labels[role]}
  class={klass}
  style:width={`${size}px`}
  style:height={`${size}px`}
  style:font-size={`${Math.round(size * 0.5)}px`}
  style:line-height="1"
>{letter}</div>

<style>
  [data-v2-avatar] {
    border-radius: var(--v2-radius-pill);
    display: grid;
    place-items: center;
    font-family: var(--v2-font-display);
    font-style: italic;
    letter-spacing: 0.02em;
    flex-shrink: 0;
    border: 1px solid color-mix(in srgb, var(--v2-rule) 50%, transparent);
  }
  [data-v2-avatar][data-role='user'] {
    background: var(--v2-paper-2);
    color: var(--v2-ink);
  }
  [data-v2-avatar][data-role='agent'] {
    background: var(--v2-accent);
    color: var(--v2-paper);
    border-color: transparent;
  }
  [data-v2-avatar][data-role='system'] {
    background: transparent;
    color: var(--v2-ink-3);
    border-style: dashed;
  }
</style>
