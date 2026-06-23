<script lang="ts">
  // HotkeyRecorder — captures a global-hotkey combo from the
  // keyboard and emits a spec string the daemon's hotkey parser
  // accepts (e.g. "Cmd+Shift+Space").
  //
  // Parser contract (internal/hotkey/parse.go):
  //   - at least one modifier + exactly one key
  //   - modifiers: Cmd | Ctrl | Alt/Option | Shift | Win/Super
  //   - keys: Space, Esc, Tab, Enter, Delete, arrows, F1-F12,
  //     or a single printable ASCII character (A-Z, 0-9, punctuation)

  import { t } from '../i18n'

  interface Props {
    value?: string
    onRecord?: (combo: string) => void
  }

  let { value = '', onRecord }: Props = $props()

  let recording = $state(false)
  // combo is the recorder's locally-tracked hotkey string.
  // We seed it from the parent's `value` prop but do NOT
  // re-sync on prop changes — the user owns the recording
  // once they start typing, and overwriting their combo
  // when the parent re-renders would be surprising.
  // svelte-ignore state_referenced_locally
  let combo = $state(value)
  let hint = $state('')

  const isMac =
    typeof navigator !== 'undefined' &&
    /mac|iphone|ipad/i.test(navigator.platform || navigator.userAgent)

  const suggestions = isMac
    ? ['Cmd+Shift+Space', 'Ctrl+Space', 'Cmd+Shift+K']
    : ['Ctrl+Shift+Space', 'Ctrl+Space', 'Ctrl+Shift+K']

  // Map a KeyboardEvent's main key to a parser-compatible name.
  // Returns '' when the key is a bare modifier (ignored).
  function keyName(e: KeyboardEvent): string {
    const code = e.code
    const key = e.key

    if (['Meta', 'Control', 'Alt', 'Shift', 'OS'].includes(key)) return ''

    if (code === 'Space' || key === ' ') return 'Space'
    if (key === 'Escape') return 'Esc'
    if (key === 'Tab') return 'Tab'
    if (key === 'Enter') return 'Enter'
    if (key === 'Backspace' || key === 'Delete') return 'Delete'
    if (key === 'ArrowLeft') return 'Left'
    if (key === 'ArrowRight') return 'Right'
    if (key === 'ArrowUp') return 'Up'
    if (key === 'ArrowDown') return 'Down'
    if (/^F([1-9]|1[0-2])$/.test(key)) return key // F1..F12

    // Letters / digits: prefer the physical code so layout +
    // active modifiers don't mangle the character.
    const letter = /^Key([A-Z])$/.exec(code)
    if (letter) return letter[1]
    const digit = /^Digit([0-9])$/.exec(code)
    if (digit) return digit[1]

    // Single printable ASCII fallback.
    if (key.length === 1) {
      const c = key.charCodeAt(0)
      if (c >= 0x20 && c <= 0x7e) return key.toUpperCase()
    }
    return ''
  }

  function modifiers(e: KeyboardEvent): string[] {
    const mods: string[] = []
    if (e.metaKey) mods.push(isMac ? 'Cmd' : 'Win')
    if (e.ctrlKey) mods.push('Ctrl')
    if (e.altKey) mods.push(isMac ? 'Option' : 'Alt')
    if (e.shiftKey) mods.push('Shift')
    return mods
  }

  function onKeydown(e: KeyboardEvent): void {
    if (!recording) return
    e.preventDefault()
    e.stopPropagation()

    const key = keyName(e)
    if (!key) {
      hint = t('hotkey.recorder.hint_modifier')
      return
    }
    const mods = modifiers(e)
    if (mods.length === 0) {
      hint = t('hotkey.recorder.hint_no_modifier')
      return
    }
    const spec = [...mods, key].join('+')
    combo = spec
    hint = ''
    recording = false
    onRecord?.(spec)
  }

  function start(): void {
    recording = true
    hint = t('hotkey.recorder.hint_press')
  }

  function pick(s: string): void {
    combo = s
    hint = ''
    recording = false
    onRecord?.(s)
  }
</script>

<svelte:window onkeydown={onKeydown} />

<div class="recorder">
  <button
    type="button"
    class="capture"
    class:recording
    class:filled={!!combo}
    onclick={start}
    aria-label={t('hotkey.recorder.aria_label')}
  >
    {#if recording}
      <span class="pulse">{t('hotkey.recorder.recording')}</span>
    {:else if combo}
      <kbd>{combo}</kbd>
    {:else}
      <span class="placeholder">{t('hotkey.recorder.placeholder')}</span>
    {/if}
  </button>

  {#if hint}
    <p class="hint">{hint}</p>
  {/if}

  <div class="suggestions">
    <span class="sug-label">{t('hotkey.recorder.suggestions')}</span>
    {#each suggestions as s}
      <button type="button" class="btn btn-ghost btn-xs chip" onclick={() => pick(s)}>{s}</button>
    {/each}
  </div>
</div>

<style>
  .recorder {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    width: 100%;
  }

  .capture {
    width: 100%;
    padding: var(--space-5);
    border-radius: var(--radius-pill);
    border: 1px dashed var(--glass-border);
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    color: var(--color-text);
    font-size: var(--size-lg);
    cursor: pointer;
    transition: border-color var(--transition-base), box-shadow var(--transition-base), background var(--transition-base);
    min-height: 72px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: var(--shadow-inset);
  }
  .capture:hover {
    border-color: var(--color-border-accent);
  }
  .capture.recording {
    border-style: solid;
    border-color: var(--color-accent);
    background: var(--color-accent-faint);
    box-shadow: var(--shadow-glow-strong), var(--shadow-inset);
  }
  .capture.filled {
    border-style: solid;
    border-color: var(--color-border-accent);
  }

  .placeholder {
    color: var(--color-text-muted);
    font-size: var(--size-md);
  }

  .pulse {
    color: var(--color-accent);
    font-weight: var(--weight-medium);
    animation: blink 1.2s ease-in-out infinite;
  }
  @keyframes blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }

  kbd {
    font-family: var(--font-mono);
    font-size: var(--size-xl);
    font-weight: var(--weight-semibold);
    background: var(--color-bg-active);
    border: 1px solid var(--glass-border-hover);
    border-radius: var(--radius-md);
    padding: var(--space-2) var(--space-4);
    letter-spacing: var(--tracking-wide);
    box-shadow: var(--shadow-inset);
  }

  .hint {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    margin: 0;
    text-align: center;
  }

  .suggestions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    justify-content: center;
  }
  .sug-label {
    color: var(--color-text-faint);
    font-size: var(--size-sm);
  }
  .chip {
    font-family: var(--font-mono);
    border-radius: var(--radius-pill);
  }
</style>
