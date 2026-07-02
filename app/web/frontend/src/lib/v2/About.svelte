<!--
  About — Condura v2 colophon surface.

  Per the spec: "A colophon. Like the credits of a thoughtful book.
  Subtle, not boastful. A single word cycles in the footer: 'With
  care.' → 'For everyone.' → 'Forever free.'"

  Pure presentation. Subtle type, deliberate typography.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Eyebrow, Glyph } from '$lib/v2'

  // A single word cycles in the footer.
  const words = ['With care.', 'For everyone.', 'Forever free.']
  let wordIndex = $state(0)
  $effect(() => {
    const id = setInterval(() => { wordIndex = (wordIndex + 1) % words.length }, 4000)
    return () => clearInterval(id)
  })
</script>

<div data-v2 style="
  flex: 1;
  background: var(--v2-paper);
  overflow-y: auto;
  padding: var(--v2-space-16) var(--v2-space-12);
  box-sizing: border-box;
">
  <div style="max-width: 540px; margin: 0 auto;">

    <Stack gap={16}>

      <!-- ── Title ─────────────────────────────────── -->
      <Stack gap={6} align="center">
        <Glyph name="sparkle" size={28} />
        <Ink kind="display" as="h1" style:text-align="center">Condura.</Ink>
        <Ink kind="body-2" tone="ink-2" as="p" style:text-align="center">
          A free, OS-native AI agent that lives on your computer and
          acts as the conductor of every other AI tool installed here.
        </Ink>
      </Stack>

      <Rule />

      <!-- ── Colophon body ──────────────────────────── -->
      <Stack gap={6}>

        <Ink kind="body">
          Made by people who think AI should be useful, private, and
          free — and who got tired of asking permission to use their
          own software.
        </Ink>

        <Stack gap={4}>
          <Ink kind="caption" tone="ink-3">team</Ink>
          <Stack gap={2}>
            {#each [
              { role: 'product + architecture', who: 'sahaj patel' },
              { role: 'first-class implementer + reviewer', who: 'a partner agent (any model)' },
            ] as line}
              <Inline gap={3} align="baseline">
                <span style="
                  font-family: var(--v2-font-mono);
                  font-size: 11px;
                  color: var(--v2-ink-3);
                  text-transform: uppercase;
                  letter-spacing: 0.06em;
                  min-width: 200px;
                ">{line.role}</span>
                <Ink kind="body" italic>{line.who}</Ink>
              </Inline>
            {/each}
          </Stack>
        </Stack>

        <Stack gap={4}>
          <Ink kind="caption" tone="ink-3">made with</Ink>
          <Stack gap={2}>
            <Inline gap={3} align="baseline">
              <Ink kind="ui" weight="medium">Go 1.22+</Ink>
              <Ink kind="ui-small" tone="ink-3">the daemon</Ink>
            </Inline>
            <Inline gap={3} align="baseline">
              <Ink kind="ui" weight="medium">Svelte 5 + TypeScript</Ink>
              <Ink kind="ui-small" tone="ink-3">the GUI</Ink>
            </Inline>
            <Inline gap={3} align="baseline">
              <Ink kind="ui" weight="medium">Next.js 14 + Tailwind v4 + motion v12</Ink>
              <Ink kind="ui-small" tone="ink-3">the marketing site</Ink>
            </Inline>
            <Inline gap={3} align="baseline">
              <Ink kind="ui" weight="medium">SQLite + AES-GCM</Ink>
              <Ink kind="ui-small" tone="ink-3">memory on disk</Ink>
            </Inline>
            <Inline gap={3} align="baseline">
              <Ink kind="ui" weight="medium">openWakeWord + whisper.cpp</Ink>
              <Ink kind="ui-small" tone="ink-3">voice</Ink>
            </Inline>
          </Stack>
        </Stack>

        <Stack gap={4}>
          <Ink kind="caption" tone="ink-3">principles</Ink>
          <Stack gap={2}>
            {#each [
              'free forever, no premium tier, no nags',
              'local-first · your data stays on your machine',
              'strategist and gatekeeper are separate systems',
              'every action is auditable in a tamper-resistant log',
              'speed is the product',
            ] as line}
              <Inline gap={2} align="baseline">
                <span style="
                  color: var(--v2-accent);
                  font-family: var(--v2-font-mono);
                  font-size: var(--v2-text-14);
                  line-height: 1;
                ">·</span>
                <Ink kind="body" tone="ink-2" italic>"{line}"</Ink>
              </Inline>
            {/each}
          </Stack>
        </Stack>
      </Stack>

      <Rule />

      <!-- ── Footer ────────────────────────────────── -->
      <Stack gap={3} align="center">
        <span style="
          font-family: var(--v2-font-display);
          font-style: italic;
          font-size: var(--v2-text-28);
          color: var(--v2-accent);
          transition: color var(--v2-dur-slow) var(--v2-ease-settle);
        ">{words[wordIndex]}</span>
        <Ink kind="caption" tone="ink-3">v0.1.0 · build {`${Math.floor(Math.random() * 1e6).toString().padStart(6, '0')}`}</Ink>
      </Stack>

    </Stack>

  </div>
</div>
