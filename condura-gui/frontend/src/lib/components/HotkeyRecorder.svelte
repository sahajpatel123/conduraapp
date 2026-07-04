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
  //
  // Visual feedback:
  //   - Idle:    pill with placeholder
  //   - Recording: solid accent border, accent-tinted bg, glow
  //   - Captured (valid): green flash on the pill
  //   - Bare modifier:    red shake on the pill
  //
  // Event API preserved: `value` prop + `onRecord(combo)` callback.

  import { t } from '../i18n'
  import Kbd from './ui/Kbd.svelte'

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
  let shake = $state(false)
  let flash = $state(false)
  let shakeTimer: ReturnType<typeof setTimeout> | null = null
  let flashTimer: ReturnType<typeof setTimeout> | null = null

  const isMac =
    typeof navigator !== 'undefined' &&
    /mac|iphone|ipad/i.test(navigator.platform || navigator.userAgent)

  const suggestions = isMac
    ? ['Ctrl+S', 'Cmd+Shift+Space', 'Ctrl+Space', 'Cmd+Shift+K']
    : ['Ctrl+S', 'Ctrl+Shift+Space', 'Ctrl+Space', 'Ctrl+Shift+K']

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

  function triggerShake(): void {
    shake = true
    if (shakeTimer !== null) clearTimeout(shakeTimer)
    shakeTimer = setTimeout(() => {
      shake = false
      shakeTimer = null
    }, 360)
  }

  function triggerFlash(): void {
    flash = true
    if (flashTimer !== null) clearTimeout(flashTimer)
    flashTimer = setTimeout(() => {
      flash = false
      flashTimer = null
    }, 480)
  }

  function onKeydown(e: KeyboardEvent): void {
    if (!recording) return
    e.preventDefault()
    e.stopPropagation()

    const key = keyName(e)
    if (!key) {
      hint = t('hotkey.recorder.hint_modifier')
      triggerShake()
      return
    }
    const mods = modifiers(e)
    if (mods.length === 0) {
      hint = t('hotkey.recorder.hint_no_modifier')
      triggerShake()
      return
    }
    const spec = [...mods, key].join('+')
    combo = spec
    hint = ''
    recording = false
    triggerFlash()
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
    triggerFlash()
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
    class:shake
    class:flash
    onclick={start}
    aria-label={t('hotkey.recorder.aria_label')}
  >
    {#if recording}
      <span class="pulse">{t('hotkey.recorder.recording')}</span>
    {:else if combo}
      <span class="combo-display">
        {#each combo.split('+') as seg, i}
          {#if i > 0}<span class="plus" aria-hidden="true">+</span>{/if}
          <Kbd label={seg} />
        {/each}
      </span>
    {:else}
      <span class="placeholder">{t('hotkey.recorder.placeholder')}</span>
    {/if}
  </button>

  {#if hint}
    <p class="hint" class:hint-error={shake}>{hint}</p>
  {/if}

  <div class="suggestions">
    <span class="sug-label">{t('hotkey.recorder.suggestions')}</span>
    {#each suggestions as s}
      <button
        type="button"
        class="chip"
        onclick={() => pick(s)}
        aria-label={t('hotkey.recorder.suggestion_aria', s)}
      >{s}</button>
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
    color: var(--text);
    font-size: var(--size-lg);
    cursor: pointer;
    transition:
      border-color var(--transition-base) ease,
      background-color var(--transition-base) ease,
      box-shadow var(--transition-base) ease,
      transform var(--transition-base) var(--ease-spring);
    min-height: 72px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: var(--shadow-inset);
  }
  .capture:hover {
    border-color: var(--border-strong);
  }
  .capture.recording {
    border-style: solid;
    border-color: var(--accent);
    background: var(--accent-faint);
    box-shadow: var(--shadow-focus), 0 0 28px var(--accent-glow), var(--shadow-inset);
  }
  .capture.filled {
    border-style: solid;
    border-color: var(--border-focus);
    background: var(--surface-2);
  }
  .capture.flash {
    border-color: var(--success);
    background: var(--success-soft);
    box-shadow: 0 0 0 3px var(--success-glow), var(--shadow-inset);
  }
  .capture.shake {
    border-color: var(--border-danger);
    background: var(--error-soft);
    animation: shake-kf 0.36s var(--ease-in-out-quart);
  }
  @keyframes shake-kf {
    0%, 100% { transform: translateX(0); }
    20%      { transform: translateX(-6px); }
    40%      { transform: translateX(6px); }
    60%      { transform: translateX(-4px); }
    80%      { transform: translateX(4px); }
  }

  .placeholder {
    color: var(--text-muted);
    font-size: var(--size-md);
  }

  .pulse {
    color: var(--accent);
    font-weight: var(--weight-medium);
    font-family: var(--font-mono);
    letter-spacing: var(--tracking-wide);
    animation: blink 1.2s ease-in-out infinite;
  }
  @keyframes blink {
    0%, 100% { opacity: 1; }
    50%      { opacity: 0.4; }
  }

  .combo-display {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
    justify-content: center;
  }
  .plus {
    color: var(--text-faint);
    font-family: var(--font-mono);
    font-size: var(--size-md);
    user-select: none;
  }

  .hint {
    color: var(--text-muted);
    font-size: var(--size-sm);
    margin: 0;
    text-align: center;
    transition: color var(--transition-fast) ease;
  }
  .hint.hint-error { color: var(--error); }

  .suggestions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    justify-content: center;
  }
  .sug-label {
    color: var(--text-faint);
    font-size: var(--size-sm);
  }
  .chip {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 4px 12px;
    background: var(--surface-2);
    color: var(--text-muted);
    border: 1px solid var(--border);
    border-radius: var(--radius-pill);
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    letter-spacing: var(--tracking-wide);
    cursor: pointer;
    transition:
      background-color var(--transition-fast) ease,
      color var(--transition-fast) ease,
      border-color var(--transition-fast) ease,
      transform var(--transition-fast) var(--ease-spring);
  }
  .chip:hover {
    color: var(--text);
    background: var(--surface-3);
    border-color: var(--border-focus);
    transform: translateY(-1px);
  }
  .chip:active {
    transform: translateY(0) scale(0.97);
    transition-duration: var(--transition-instant);
  }

  @media (prefers-reduced-motion: reduce) {
    .capture { animation: none !important; }
    .capture.shake { animation: none !important; transform: none; }
    .pulse { animation: none !important; }
  }
</style>
