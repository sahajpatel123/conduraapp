<!--
  Switch — boolean toggle.

  Per spec §10.5 + §11.3: settings uses switches everywhere, not checkboxes.
  The switch label is sans, the value (on/off) is mono for tabular alignment.

  Props:
    checked     — bound boolean (Svelte 5 via $bindable)
    label       — visible label (placed left of the switch)
    description — secondary explanation (placed below label, smaller)
    disabled    — disables the switch
    onchange    — handler
-->
<script lang="ts">
  interface Props {
    checked?: boolean;
    label?: string;
    description?: string;
    disabled?: boolean;
    onchange?: (checked: boolean) => void;
  }

  let {
    checked = $bindable(false),
    label,
    description,
    disabled = false,
    onchange,
  }: Props = $props();

  function toggle() {
    if (disabled) return;
    checked = !checked;
    onchange?.(checked);
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === ' ' || e.key === 'Enter') {
      e.preventDefault();
      toggle();
    }
  }
</script>

<label class="switch" class:switch--disabled={disabled}>
  <div class="switch__text">
    {#if label}
      <span class="switch__label">{label}</span>
    {/if}
    {#if description}
      <span class="switch__description">{description}</span>
    {/if}
  </div>
  <button
    role="switch"
    type="button"
    aria-checked={checked}
    aria-label={label}
    {disabled}
    class="switch__track"
    class:switch__track--on={checked}
    onclick={toggle}
    onkeydown={onKey}
  >
    <span class="switch__thumb" aria-hidden="true"></span>
  </button>
</label>

<style>
  .switch {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    cursor: pointer;
    padding: var(--space-3) 0;
  }
  .switch--disabled {
    cursor: not-allowed;
    opacity: 0.5;
  }

  .switch__text {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    flex: 1;
    min-width: 0;
  }
  .switch__label {
    color: var(--content-primary);
    font-size: var(--text-body-size);
  }
  .switch__description {
    color: var(--content-tertiary);
    font-size: var(--text-body-sm-size);
    line-height: 1.5;
  }

  .switch__track {
    position: relative;
    width: 36px;
    height: 20px;
    border-radius: var(--radius-pill);
    background-color: var(--ink-cool-200);
    border: 1px solid var(--border-default);
    cursor: pointer;
    transition: background-color var(--duration-fast) var(--ease-standard);
    padding: 0;
    flex-shrink: 0;
  }
  .switch__track--on {
    background-color: var(--plum-600);
    border-color: var(--plum-600);
  }
  .switch__thumb {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 14px;
    height: 14px;
    border-radius: var(--radius-pill);
    background-color: var(--paper-warm-0);
    box-shadow: var(--shadow-1);
    transition: transform var(--duration-fast) var(--ease-standard);
  }
  .switch__track--on .switch__thumb {
    transform: translateX(16px);
  }
  .switch__track:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: 2px;
  }
</style>