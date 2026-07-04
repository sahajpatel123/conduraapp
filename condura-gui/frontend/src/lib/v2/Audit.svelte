<!--
  Audit — Condura v2 audit log surface.

  Per the spec: "An evidence locker. Each entry a paper card with a
  left rule in ink-2, top-right a hash, body in mono. Click a row →
  expands to reveal the full reasoning trace (collapsible, like a
  court transcript). HMAC verification at the top with a green
  checkmark or red exclamation — no confetti either way."

  Pure presentation. Parent route owns the entries array and the
  integrity verification result (typically from a `replay.verify_integrity`
  IPC call).

  Props:
    entries:        AuditEntry[]
    integrity:      'unknown' | 'verified' | 'broken'
    integrityDetail?: string  — human-readable summary under the badge
    onLoad?:        (id: string) => void  — fires when row expand loads more detail
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Eyebrow, Glyph } from '$lib/v2'

  export interface AuditEntry {
    id: string
    timestamp: string
    actor: 'condura' | 'user' | 'system' | 'gatekeeper'
    action: string
    blastRadius: 'read' | 'write' | 'network' | 'destructive'
    hash: string
    detail?: string
  }

  let {
    entries = [] as AuditEntry[],
    integrity = 'unknown' as 'unknown' | 'verified' | 'broken',
    integrityDetail = undefined as string | undefined,
  }: {
    entries?: AuditEntry[]
    integrity?: 'unknown' | 'verified' | 'broken'
    integrityDetail?: string
  } = $props()

  let expanded = $state<Record<string, boolean>>({})
  function toggle(id: string) {
    expanded = { ...expanded, [id]: !expanded[id] }
  }

  // Visual variant per blast radius: the left rule's thickness
  // tells the severity at a glance, even before you read.
  function ruleWidth(r: AuditEntry['blastRadius']): string {
    if (r === 'destructive') return '4px'
    if (r === 'network') return '3px'
    if (r === 'write') return '2px'
    return '1px'
  }

  function actorMonogram(a: AuditEntry['actor']): string {
    switch (a) {
      case 'condura':   return 'C'
      case 'user':      return 'U'
      case 'system':    return 'S'
      case 'gatekeeper':return 'G'
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
  <div style="max-width: 960px; margin: 0 auto;">

    <Stack gap={10}>

      <!-- ── Title + integrity badge ──────────────── -->
      <Stack gap={6}>
        <Stack gap={3}>
          <Eyebrow left="audit" right="evidence locker" tone="ink-3" />
          <Ink kind="display" as="h1">Every action is auditable.</Ink>
          <Ink kind="body-2" tone="ink-2" as="p" style:max-width="640px">
            An HMAC-chained, append-only log of every action the agent has taken.
            Tampering breaks the chain; verification surfaces the result.
          </Ink>
        </Stack>

        <!-- Integrity badge -->
        <Surface
          elevation={0}
          padding="6"
          radius="2"
          tone={integrity === 'broken' ? 'paper' : 'paper'}
          style:border-left="3px solid {integrity === 'broken' ? 'var(--v2-signal-stop)' : integrity === 'verified' ? 'var(--v2-signal-go)' : 'var(--v2-rule)'}"
        >
          <Inline gap={4} align="center">
            {#if integrity === 'verified'}
              <Glyph name="check" size={20} />
              <Stack gap={1}>
                <Ink kind="ui" weight="medium" tone="signal-go">HMAC chain verified.</Ink>
                {#if integrityDetail}<Ink kind="ui-small" tone="ink-3">{integrityDetail}</Ink>{/if}
              </Stack>
            {:else if integrity === 'broken'}
              <Glyph name="x" size={20} />
              <Stack gap={1}>
                <Ink kind="ui" weight="medium" tone="signal-stop">Chain integrity broken.</Ink>
                {#if integrityDetail}<Ink kind="ui-small" tone="ink-3">{integrityDetail}</Ink>{/if}
              </Stack>
            {:else}
              <span style="
                width: 14px; height: 14px;
                border-radius: var(--v2-radius-pill);
                border: 2px solid var(--v2-ink-3);
                border-top-color: transparent;
                animation: v2-spin 800ms linear infinite;
                display: inline-block;
              "></span>
              <Ink kind="ui" tone="ink-3">Verifying chain…</Ink>
            {/if}
          </Inline>
        </Surface>
      </Stack>

      <Rule />

      <!-- ── Entries ───────────────────────────────── -->
      <Stack gap={3}>
        <Eyebrow left="entries" right={`${entries.length} total`} tone="ink-3" />

        {#if entries.length === 0}
          <Surface elevation={0} padding="12" radius="2" tone="paper">
            <Stack gap={3} align="center">
              <Ink kind="display" as="h3" style:font-size="var(--v2-text-28)">Nothing logged yet.</Ink>
              <Ink kind="body" tone="ink-2">Once condura takes any action, every step lands here.</Ink>
            </Stack>
          </Surface>
        {:else}
          {#each entries as e (e.id)}
            <Surface
              elevation={0}
              padding="0"
              radius="2"
              tone="paper"
              style:overflow="hidden"
              style:border-left="{ruleWidth(e.blastRadius)} solid var(--v2-ink-2)"
            >
              <button
                data-v2-audit-row
                onclick={() => toggle(e.id)}
                aria-expanded={!!expanded[e.id]}
                aria-controls={`audit-detail-${e.id}`}
                style="
                  all: unset;
                  cursor: pointer;
                  display: block;
                  width: 100%;
                  padding: var(--v2-space-4) var(--v2-space-6);
                  box-sizing: border-box;
                "
              >
                <Inline gap={4} align="baseline" justify="between">
                  <Inline gap={3} align="baseline">
                    <span style="
                      font-family: var(--v2-font-mono);
                      font-size: var(--v2-text-12);
                      color: var(--v2-ink-3);
                      font-feature-settings: var(--v2-numeric-features);
                      min-width: 96px;
                    ">{e.timestamp}</span>
                    <Avatar role={e.actor === 'user' ? 'user' : e.actor === 'gatekeeper' ? 'system' : 'agent'} size={20} monogram={actorMonogram(e.actor)} />
                    <Ink kind="ui" weight="medium">{e.action}</Ink>
                  </Inline>
                  <Inline gap={2} align="center">
                    <Chip on={false} variant={e.blastRadius === 'destructive' ? 'signal-stop' : e.blastRadius === 'network' ? 'signal-warn' : 'default'} size="small">
                      {e.blastRadius}
                    </Chip>
                    <span style="
                      font-family: var(--v2-font-mono);
                      font-size: 11px;
                      color: var(--v2-ink-3);
                      letter-spacing: 0.04em;
                    ">{e.hash.slice(0, 8)}…</span>
                    <Glyph
                      name={expanded[e.id] ? 'chevron-right' : 'chevron-left'}
                      size={12}
                    />
                  </Inline>
                </Inline>
              </button>

              {#if expanded[e.id]}
                <div
                  id={`audit-detail-${e.id}`}
                  role="region"
                  aria-label={`Details for ${e.action}`}
                  style="
                    border-top: 1px solid var(--v2-rule);
                    padding: var(--v2-space-6);
                    background: var(--v2-paper-2);
                  "
                >
                  <Stack gap={3}>
                    <Ink kind="caption" tone="ink-3">full hash</Ink>
                    <span style="
                      font-family: var(--v2-font-mono);
                      font-size: var(--v2-text-12);
                      color: var(--v2-ink);
                      word-break: break-all;
                      background: var(--v2-paper);
                      padding: var(--v2-space-3);
                      border-radius: var(--v2-radius-1);
                      border: 1px solid var(--v2-rule);
                    ">{e.hash}</span>

                    {#if e.detail}
                      <Ink kind="caption" tone="ink-3">reasoning trace</Ink>
                      <Ink kind="ui-small" tone="ink-2" style:font-family="var(--v2-font-mono)" style:white-space="pre-wrap">
                        {e.detail}
                      </Ink>
                    {/if}
                  </Stack>
                </div>
              {/if}
            </Surface>
          {/each}
        {/if}
      </Stack>

    </Stack>

  </div>
</div>

<style>
  @keyframes v2-spin {
    to { transform: rotate(360deg); }
  }
  /* Focus-visible is provided by reset.css' :focus-visible rule
     for any [data-v2] button. */
  [data-v2-audit-row]:focus-visible {
    outline: none;
    background: color-mix(in srgb, var(--v2-rule) 30%, transparent);
  }
</style>
