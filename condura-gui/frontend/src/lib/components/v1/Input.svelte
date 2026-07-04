<!--
  Input — text field. Two voices:
    - command-surface variant: serif (the user's voice inside the agent surface)
    - settings variant: sans

  Per spec §10.5 (first-run, hotkey field uses serif for "as if written by hand").

  Props:
    value          — bound value (Svelte 5 two-way via $bindable)
    variant        — 'serif' | 'sans' (default sans)
    size           — 'sm' | 'md' | 'lg'
    placeholder    — placeholder text
    type           — 'text' | 'password' | 'email' | 'number' | 'search'
    ariaLabel      — required for icon-only inputs
    disabled       — disables the input
    readonly       — non-editable, still focusable
    autofocus      — focus on mount
    monospace      — force mono (for hotkey fields, IDs, etc.)
-->
<script lang="ts">
  interface Props {
    value?: string;
    variant?: 'serif' | 'sans';
    size?: 'sm' | 'md' | 'lg';
    placeholder?: string;
    type?: 'text' | 'password' | 'email' | 'number' | 'search';
    ariaLabel?: string;
    disabled?: boolean;
    readonly?: boolean;
    autofocus?: boolean;
    monospace?: boolean;
    name?: string;
    id?: string;
    oninput?: (e: Event) => void;
    onkeydown?: (e: KeyboardEvent) => void;
    onfocus?: (e: FocusEvent) => void;
    onblur?: (e: FocusEvent) => void;
  }

  let {
    value = $bindable(''),
    variant = 'sans',
    size = 'md',
    placeholder,
    type = 'text',
    ariaLabel,
    disabled = false,
    readonly = false,
    autofocus = false,
    monospace = false,
    name,
    id,
    oninput,
    onkeydown,
    onfocus,
    onblur,
  }: Props = $props();
</script>

<input
  class="input input--{variant} input--{size}"
  class:input--mono={monospace}
  class:input--disabled={disabled}
  type={type}
  {placeholder}
  {disabled}
  {readonly}
  {name}
  {id}
  aria-label={ariaLabel}
  bind:value
  oninput={(e) => { value = (e.target as HTMLInputElement).value; oninput?.(e); }}
  onkeydown={onkeydown}
  onfocus={onfocus}
  onblur={onblur}
/>

<style>
  .input {
    width: 100%;
    background-color: transparent;
    border: none;
    border-bottom: 1px solid var(--border-default);
    color: var(--content-primary);
    outline: none;
    transition: border-color var(--duration-fast) var(--ease-standard);
    font-feature-settings: 'cv02', 'cv03', 'cv04', 'cv11';
  }
  .input:focus-visible {
    border-bottom-color: var(--border-focus);
    /* subtle weight change on focus */
    box-shadow: 0 1px 0 0 var(--border-focus);
  }
  .input::placeholder {
    color: var(--content-tertiary);
    opacity: 1;
  }
  .input--mono {
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
  }

  /* Sizes */
  .input--sm {
    height: 32px;
    font-size: var(--text-body-sm-size);
  }
  .input--md {
    height: 40px;
    font-size: var(--text-body-size);
  }
  .input--lg {
    height: 56px;
    font-size: var(--text-h3-size);
  }

  /* Voice */
  .input--serif {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    line-height: 1.4;
  }

  /* Disabled */
  .input--disabled {
    color: var(--content-disabled);
    cursor: not-allowed;
  }
</style>