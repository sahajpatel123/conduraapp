<script lang="ts">
  interface Option {
    value: string
    label: string
    description?: string
    disabled?: boolean
  }

  interface Props {
    value?: string
    options: Option[]
    label?: string
    hint?: string
    placeholder?: string
    fullWidth?: boolean
    disabled?: boolean
    onchange?: (value: string) => void
  }

  let { value = $bindable(''), options, label, hint, placeholder,
        fullWidth = false, disabled = false, onchange }: Props = $props()

  function handle(e: Event): void {
    const v = (e.currentTarget as HTMLSelectElement).value
    value = v
    onchange?.(v)
  }
</script>

<label class="select-wrap" class:select-full={fullWidth}>
  {#if label}<span class="select-label">{label}</span>{/if}
  <span class="select-shell">
    <select class="select-control" {value} {disabled} onchange={handle}>
      {#if placeholder}<option value="" disabled>{placeholder}</option>{/if}
      {#each options as opt (opt.value)}
        <option value={opt.value} disabled={opt.disabled}>{opt.label}</option>
      {/each}
    </select>
    <svg class="select-chevron" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
      <path d="M6 9l6 6 6-6" />
    </svg>
  </span>
  {#if hint}<span class="select-hint">{hint}</span>{/if}
</label>

<style>
  .select-wrap {
    display: inline-flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }
  .select-full { width: 100%; }

  .select-label {
    font-size: var(--size-xs);
    color: var(--text-muted);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-wide);
    text-transform: uppercase;
    padding-left: 2px;
  }

  .select-shell {
    position: relative;
    display: flex;
    align-items: center;
  }

  .select-control {
    appearance: none;
    -webkit-appearance: none;
    width: 100%;
    background: var(--surface-1);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    color: var(--text);
    font-family: var(--font-sans);
    font-size: var(--size-md);
    padding: 0 36px 0 var(--space-3);
    height: 36px;
    cursor: pointer;
    transition: border-color var(--transition-fast) ease,
                background-color var(--transition-fast) ease,
                box-shadow var(--transition-fast) ease;
  }
  .select-control:focus {
    outline: none;
    border-color: var(--border-focus);
    background: var(--surface-2);
    box-shadow: 0 0 0 3px var(--accent-soft);
  }
  .select-control:disabled { opacity: 0.5; cursor: not-allowed; }

  .select-chevron {
    position: absolute;
    right: 12px;
    color: var(--text-faint);
    pointer-events: none;
  }

  .select-hint { font-size: var(--size-xs); color: var(--text-faint); }
</style>