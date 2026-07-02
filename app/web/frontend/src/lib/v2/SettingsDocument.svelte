<!--
  SettingsDocument — Condura v2 settings surface.

  Per the spec: "A document, not a form. Sections are numbered
  chapter headings (`01 · Account`, `02 · Voice`, `03 · Channels`…).
  Toggles are hardware-honest. Save buttons are quiet text in the
  bottom-right ('Save changes' appears when there's something to save)."

  The whole surface is a single scrollable column with chapter
  headings as numbered eyebrows, vertical sub-rows as `Field` and
  `ToggleRow` primitives, and a quiet "Save changes" affordance that
  only appears when state is dirty.

  Consumers pass `chapters` (a list of chapter specs); each chapter
  has a number, label, optional copy, and rows. Each row is either
  a `ToggleRow` (label + copy + Switch) or a `SelectRow` (label +
  copy + select).

  Props:
    chapters: Chapter[]
    dirty?:   boolean — when true, "Save changes" appears bottom-right
    onSave?:  () => void
    onDiscard?: () => void

  Chapter = { id, number, label, copy?, rows: Row[] }
  ToggleRow = { id, kind: 'toggle', label, copy?, on: boolean, onToggle }
  SelectRow = { id, kind: 'select', label, copy?, value, options, onChange }
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Switch, Eyebrow, Button } from '$lib/v2'

  type ToggleRow = { id: string, kind: 'toggle', label: string, copy?: string, on: boolean, onToggle?: () => void, disabled?: boolean }
  type SelectRow = { id: string, kind: 'select', label: string, copy?: string, value: string, options: { value: string, label: string }[], onChange?: (v: string) => void }
  type TextRow   = { id: string, kind: 'text',   label: string, copy?: string, value: string, onInput?: (v: string) => void, placeholder?: string }
  type Row = ToggleRow | SelectRow | TextRow
  type Chapter = { id: string, number: string, label: string, copy?: string, rows: Row[] }

  let {
    chapters = [] as Chapter[],
    dirty = false as boolean,
    onSave = undefined as (() => void) | undefined,
    onDiscard = undefined as (() => void) | undefined,
  }: {
    chapters?: Chapter[]
    dirty?: boolean
    onSave?: () => void
    onDiscard?: () => void
  } = $props()
</script>

<div data-v2 style="
  flex: 1;
  background: var(--v2-paper);
  overflow-y: auto;
  padding: var(--v2-space-12) var(--v2-space-12);
  box-sizing: border-box;
">
  <div style="max-width: 720px; margin: 0 auto;">

    <Stack gap={16}>

      <!-- ── Title ─────────────────────────────────── -->
      <Stack gap={3}>
        <Ink kind="mono-cap" tone="accent">document</Ink>
        <Ink kind="display" as="h1">Settings.</Ink>
        <Ink kind="body" tone="ink-2" as="p">
          Change anything. The agent adapts in real time.
        </Ink>
      </Stack>

      <Rule />

      <!-- ── Chapters ────────────────────────────────── -->
      {#each chapters as chapter, ci (chapter.id)}
        <section id={chapter.id} aria-labelledby={`chapter-${chapter.id}`}>
          <Stack gap={8}>

            <!-- Chapter heading -->
            <Stack gap={3}>
              <Eyebrow
                left={chapter.number}
                right={chapter.label.toLowerCase()}
                tone="accent"
              />
              <Ink kind="title" id={`chapter-${chapter.id}`}>{chapter.label}</Ink>
              {#if chapter.copy}
                <Ink kind="body" tone="ink-2">{chapter.copy}</Ink>
              {/if}
            </Stack>

            <!-- The chapter body — a single surface with rows inside.
                 Rows are separated by hairlines, never by full Rule
                 blocks: a setting document, not a form. -->
            <Surface elevation={0} padding="0" radius="2" tone="paper">
              <Stack gap={0}>
                {#each chapter.rows as row, ri (row.id)}
                  {#if ri > 0}<Rule orientation="horizontal" weight="1" tone="rule" inset="0" />{/if}

                  {#if row.kind === 'toggle'}
                    <div style="
                      display: flex; align-items: flex-start; gap: var(--v2-space-6);
                      padding: var(--v2-space-6);
                    ">
                      <div style="flex: 1; min-width: 0;">
                        <Stack gap={1}>
                          <Ink kind="ui" weight="medium">{row.label}</Ink>
                          {#if row.copy}
                            <Ink kind="ui-small" tone="ink-3">{row.copy}</Ink>
                          {/if}
                        </Stack>
                      </div>
                      <Switch
                        on={row.on}
                        label={row.label}
                        disabled={row.disabled}
                        onclick={row.onToggle}
                      />
                    </div>

                  {:else if row.kind === 'select'}
                    <div style="
                      display: flex; align-items: flex-start; gap: var(--v2-space-6);
                      padding: var(--v2-space-6);
                    ">
                      <div style="flex: 1; min-width: 0;">
                        <Stack gap={1}>
                          <Ink kind="ui" weight="medium">{row.label}</Ink>
                          {#if row.copy}
                            <Ink kind="ui-small" tone="ink-3">{row.copy}</Ink>
                          {/if}
                        </Stack>
                      </div>
                      <select
                        data-v2-select
                        value={row.value}
                        onchange={(e) => row.onChange?.((e.currentTarget as HTMLSelectElement).value)}
                        disabled={row.disabled}
                      >
                        {#each row.options as o}
                          <option value={o.value}>{o.label}</option>
                        {/each}
                      </select>
                    </div>

                  {:else if row.kind === 'text'}
                    <div style="
                      display: flex; align-items: flex-start; gap: var(--v2-space-6);
                      padding: var(--v2-space-6);
                    ">
                      <div style="flex: 1; min-width: 0;">
                        <Stack gap={1}>
                          <Ink kind="ui" weight="medium">{row.label}</Ink>
                          {#if row.copy}
                            <Ink kind="ui-small" tone="ink-3">{row.copy}</Ink>
                          {/if}
                        </Stack>
                      </div>
                      <input
                        data-v2-input
                        type="text"
                        value={row.value}
                        placeholder={row.placeholder}
                        oninput={(e) => row.onInput?.((e.currentTarget as HTMLInputElement).value)}
                      />
                    </div>
                  {/if}

                {/each}
              </Stack>
            </Surface>

          </Stack>
        </section>

        {#if ci < chapters.length - 1}<Rule orientation="horizontal" weight="1" tone="rule" inset="var(--v2-space-12) 0" />{/if}
      {/each}

      <!-- ── Bottom save bar — appears only when state is dirty ── -->
      {#if dirty}
        <Surface
          elevation={3}
          padding="4"
          radius="2"
          tone="paper"
          style:position="sticky"
          style:bottom="var(--v2-space-6)"
        >
          <Inline gap={4} justify="between" align="center">
            <Ink kind="ui-small" tone="ink-2" italic>unsaved changes</Ink>
            <Inline gap={2}>
              <Button variant="ghost" size="small" onclick={onDiscard}>Discard</Button>
              <Button variant="primary" size="small" onclick={onSave}>Save changes</Button>
            </Inline>
          </Inline>
        </Surface>
      {/if}

    </Stack>

  </div>
</div>

<style>
  [data-v2-select], [data-v2-input] {
    font-family: var(--v2-font-sans);
    font-size: var(--v2-text-14);
    color: var(--v2-ink);
    background: var(--v2-paper);
    border: 1px solid var(--v2-rule);
    border-radius: var(--v2-radius-1);
    padding: var(--v2-space-2) var(--v2-space-3);
    min-width: 200px;
    transition: border-color var(--v2-dur-fast) var(--v2-ease-out-soft);
  }
  [data-v2-select]:hover, [data-v2-input]:hover {
    border-color: var(--v2-ink-3);
  }
  [data-v2-input]:focus-visible, [data-v2-select]:focus-visible {
    outline: none;
    border-color: var(--v2-accent);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--v2-accent) 14%, transparent);
  }
</style>
