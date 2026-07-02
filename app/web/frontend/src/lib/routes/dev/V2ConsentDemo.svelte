<!--
  V2ConsentDemo — Condura v2 gatekeeper consent preview.

  Demonstrates the wax-seal-on-a-letter consent modal that fires when
  Gatekeeper asks the user to approve a non-READ action. The demo
  cycle: try each scenario, click an allow button, watch the seal
  stamp, see the modal close.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { ConsentModal, Surface, Ink, Stack, Inline, Rule } from '$lib/v2'

  // The current "pending" action.
  type Scenario = {
    title: string
    description: string
    blastRadius: 'read' | 'write' | 'network' | 'destructive'
    target?: { app: string, detail: string }
    impact?: string[]
    closingCopy: string
  }

  let open = $state(false)
  let current = $state<Scenario | null>(null)

  function openScenario(s: Scenario) {
    current = s
    open = true
  }
  function close() {
    open = false
    setTimeout(() => { current = null }, 280)
  }

  function approve(kind: 'once' | 'session') {
    // The modal's stamp animation runs in the component itself;
    // here we just close after stamp finishes.
    close()
  }

  const scenarios: Scenario[] = [
    {
      title: 'Send email to alex@example.com',
      description: 'Condura drafted the reply to Alex\'s product feedback. Sending will share the contents of your draft with your mail provider.',
      blastRadius: 'network',
      target: { app: 'com.apple.Mail', detail: 'Drafts · Alex Chen — Re: Atlas onboarding v2' },
      impact: [
        'deliver the drafted message to alex@example.com',
        'move the draft to Sent',
        'mark it as read in your inbox when a reply arrives',
      ],
      closingCopy: 'Email sent.',
    },
    {
      title: 'Transfer $128.40 to venmo/@priya',
      description: 'You asked Condura to send Priya her share of the dinner bill. This requires an authorized payment through Venmo.',
      blastRadius: 'destructive',
      target: { app: 'com.venmo.interactive', detail: 'venmo://paycharge?txn=pay&recipients=priya&amount=128.40' },
      impact: [
        'authorize a one-time payment of $128.40 to @priya',
        'deduct from your linked Venmo balance',
        'send a charge-confirmation notification',
      ],
      closingCopy: 'Payment authorized. (Sound: stamp.)',
    },
    {
      title: 'Edit ~/Documents/notes/Atlas.md',
      description: 'Condura is rewriting the section on activation metrics. Saving will overwrite the current version — the previous version is in your Undo history.',
      blastRadius: 'write',
      target: { app: 'com.microsoft.VSCode', detail: '/Users/sahaj/Documents/notes/Atlas.md (lines 42–86)' },
      impact: [
        'rewrite 44 lines in Atlas.md',
        'create a recoverable checkpoint before saving',
      ],
      closingCopy: 'Saved.',
    },
    {
      title: 'Read inbox — first 25 unread',
      description: 'Condura is reading your unread inbox to draft a morning summary. This is a read action — no data leaves your machine.',
      blastRadius: 'read',
      target: { app: 'com.apple.Mail', detail: 'Inbox · 25 unread messages' },
      closingCopy: 'Summary drafted.',
    },
  ]
</script>

<div data-v2 style="
  min-height: 100vh;
  background: var(--v2-paper);
  padding: var(--v2-space-12);
  box-sizing: border-box;
">
  <div style="max-width: 760px; margin: 0 auto;">

    <Stack gap={6} align="stretch">

      <Stack gap={3}>
        <Ink kind="mono-cap" tone="accent">consent modal · the gatekeeper surface</Ink>
        <Ink kind="display" as="h1">The wax seal on a letter.</Ink>
        <Ink kind="body-2" tone="ink-2" as="p">
          When Condura is about to do something that touches the world
          outside the LLM, the Gatekeeper pauses and asks. Trigger
          each scenario below to see the modal:
        </Ink>
      </Stack>

      <Rule />

      <Stack gap={4}>
        {#each scenarios as s, i}
          <button
            data-v2
            onclick={() => openScenario(s)}
            style="
              all: unset;
              cursor: pointer;
              display: block;
              width: 100%;
            "
          >
            <Surface
              elevation={0}
              padding="6"
              radius="2"
              tone="paper"
              interactive
              style:border-left="3px solid var(--v2-accent)"
            >
              <Stack gap={2}>
                <Inline gap={2} align="baseline">
                  <span style="
                    font-family: var(--v2-font-mono);
                    font-size: var(--v2-text-12);
                    color: var(--v2-ink-3);
                    text-transform: uppercase;
                    letter-spacing: 0.06em;
                  ">scenario {i + 1}</span>
                  <span style="
                    font-family: var(--v2-font-mono);
                    font-size: var(--v2-text-12);
                    padding: 2px var(--v2-space-2);
                    border-radius: var(--v2-radius-pill);
                    border: 1px solid {s.blastRadius === 'destructive'
                      ? 'var(--v2-signal-stop)'
                      : 'var(--v2-rule)'};
                    color: {s.blastRadius === 'destructive'
                      ? 'var(--v2-signal-stop)'
                      : 'var(--v2-ink-3)'};
                    text-transform: uppercase;
                    letter-spacing: 0.06em;
                  ">{s.blastRadius}</span>
                </Inline>
                <Ink kind="title" style:font-size="var(--v2-text-20)">{s.title}</Ink>
                <Ink kind="ui-small" tone="ink-3">
                  Click to fire the consent prompt.
                </Ink>
              </Stack>
            </Surface>
          </button>
        {/each}
      </Stack>

      <Rule />

      <Surface elevation={0} padding="6" radius="2" tone="paper-2">
        <Stack gap={3}>
          <Ink kind="caption" tone="ink-3">why a seal?</Ink>
          <Ink kind="body" tone="ink-2">
            Every allow button in the v2 system hides a 280ms wax seal.
            It's the only ceremonial motion — used once, earned every
            time. The brief stamp is the agent's way of saying: <em>I
            see this was yours to decide.</em>
          </Ink>
        </Stack>
      </Surface>

    </Stack>

  </div>
</div>

{#if current}
  <ConsentModal
    open={open}
    title={current.title}
    description={current.description}
    blastRadius={current.blastRadius}
    target={current.target}
    impact={current.impact}
    sessionTtlLabel="Allow for this session"
    onDeny={close}
    onAllowOnce={() => approve('once')}
    onAllowSession={() => approve('session')}
  />
{/if}
