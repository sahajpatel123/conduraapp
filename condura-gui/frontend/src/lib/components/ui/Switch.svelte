<script lang="ts">
  interface Props {
    checked?: boolean
    disabled?: boolean
    label?: string
    description?: string
    onchange?: (checked: boolean) => void
  }

  let { checked = $bindable(false), disabled = false, label, description, onchange }: Props = $props()

  function handle(e: Event): void {
    const v = (e.currentTarget as HTMLInputElement).checked
    checked = v
    onchange?.(v)
  }
</script>

<label class="switch-row" class:disabled>
  <span class="switch-text">
    {#if label}<span class="switch-label">{label}</span>{/if}
    {#if description}<span class="switch-description">{description}</span>{/if}
  </span>
  <span class="switch-track" class:on={checked}>
    <input
      type="checkbox"
      class="switch-input"
      {disabled}
      bind:checked
      onchange={handle}
    />
    <span class="switch-thumb"></span>
  </span>
</label>

<style>
  .switch-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    cursor: pointer;
    padding: var(--space-2) 0;
  }
  .switch-row.disabled { opacity: 0.5; cursor: not-allowed; }

  .switch-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .switch-label {
    font-size: var(--size-md);
    color: var(--text);
    font-weight: var(--weight-medium);
  }
  .switch-description {
    font-size: var(--size-xs);
    color: var(--text-muted);
    line-height: var(--leading-normal);
  }

  .switch-track {
    position: relative;
    flex-shrink: 0;
    width: 36px;
    height: 20px;
    background: var(--surface-3);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-pill);
    transition: background-color var(--transition-fast) ease,
                border-color var(--transition-fast) ease;
  }
  .switch-track.on {
    background: var(--accent);
    border-color: var(--accent);
  }

  .switch-thumb {
    position: absolute;
    top: 50%;
    left: 2px;
    transform: translateY(-50%);
    width: 14px;
    height: 14px;
    background: var(--text);
    border-radius: 50%;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
    transition: left var(--transition-fast) var(--ease-spring),
                background-color var(--transition-fast) ease;
  }
  .switch-track.on .switch-thumb {
    left: 18px;
    background: var(--text-inverse);
  }

  .switch-input {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    opacity: 0;
    cursor: inherit;
    margin: 0;
  }
</style>