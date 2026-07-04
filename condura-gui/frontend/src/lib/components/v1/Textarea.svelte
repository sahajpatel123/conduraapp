<!--
  Textarea — multiline text input. Settings/notes style.

  Per spec §11.3 (Settings): every preference is one row with a label,
  current value, and inline control. Textarea handles multi-row values
  (e.g., API key passphrase, custom instructions).

  Props:
    value          — bound via $bindable
    placeholder    — placeholder text
    rows           — visible row count
    disabled       — disabled state
    readonly       — non-editable
    monospace      — mono font for code/IDs
-->
<script lang="ts">
  interface Props {
    value?: string;
    placeholder?: string;
    rows?: number;
    disabled?: boolean;
    readonly?: boolean;
    monospace?: boolean;
    ariaLabel?: string;
    name?: string;
    id?: string;
    oninput?: (e: Event) => void;
    onkeydown?: (e: KeyboardEvent) => void;
  }

  let {
    value = $bindable(''),
    placeholder,
    rows = 4,
    disabled = false,
    readonly = false,
    monospace = false,
    ariaLabel,
    name,
    id,
    oninput,
    onkeydown,
  }: Props = $props();
</script>

<textarea
  class="textarea"
  class:textarea--mono={monospace}
  class:textarea--disabled={disabled}
  {placeholder}
  {rows}
  {disabled}
  {readonly}
  {name}
  {id}
  aria-label={ariaLabel}
  bind:value
  oninput={(e) => { value = (e.target as HTMLTextAreaElement).value; oninput?.(e); }}
  onkeydown={onkeydown}
></textarea>

<style>
  .textarea {
    width: 100%;
    padding: var(--space-3);
    background-color: var(--surface-base);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-sm);
    color: var(--content-primary);
    font-family: var(--font-sans);
    font-size: var(--text-body-size);
    line-height: 1.55;
    outline: none;
    resize: vertical;
    transition: border-color var(--duration-fast) var(--ease-standard);
    font-feature-settings: 'cv02', 'cv03', 'cv04', 'cv11';
  }

  .textarea--mono {
    font-family: var(--font-mono);
    font-size: var(--text-body-sm-size);
    font-variant-numeric: tabular-nums;
  }

  .textarea:focus-visible {
    border-color: var(--border-focus);
    box-shadow: 0 0 0 2px var(--border-focus);
  }

  .textarea::placeholder {
    color: var(--content-tertiary);
  }

  .textarea--disabled {
    background-color: var(--surface-sunken);
    color: var(--content-disabled);
    cursor: not-allowed;
  }
</style>