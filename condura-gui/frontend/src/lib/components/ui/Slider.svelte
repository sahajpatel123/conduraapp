<script lang="ts">
  interface Props {
    value?: number
    min?: number
    max?: number
    step?: number
    label?: string
    showValue?: boolean
    formatValue?: (v: number) => string
    onchange?: (value: number) => void
  }

  let { value = $bindable(0), min = 0, max = 100, step = 1,
        label, showValue = true, formatValue, onchange }: Props = $props()

  function handle(e: Event): void {
    const v = Number((e.currentTarget as HTMLInputElement).value)
    value = v
    onchange?.(v)
  }

  const pct = $derived(((value - min) / (max - min)) * 100)
</script>

<label class="slider-wrap">
  {#if label || showValue}
    <span class="slider-row">
      {#if label}<span class="slider-label">{label}</span>{/if}
      {#if showValue}
        <span class="slider-value">{formatValue ? formatValue(value) : value}</span>
      {/if}
    </span>
  {/if}
  <span class="slider-shell">
    <span class="slider-fill" style="width: {pct}%"></span>
    <input
      type="range"
      class="slider-input"
      {min} {max} {step}
      {value}
      oninput={handle}
    />
  </span>
</label>

<style>
  .slider-wrap {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }

  .slider-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .slider-label {
    font-size: var(--size-xs);
    color: var(--text-muted);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-wide);
    text-transform: uppercase;
  }
  .slider-value {
    font-family: var(--font-mono);
    font-size: var(--size-sm);
    color: var(--text);
  }

  .slider-shell {
    position: relative;
    height: 24px;
    display: flex;
    align-items: center;
    background: var(--surface-3);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-pill);
  }

  .slider-fill {
    position: absolute;
    top: 0; bottom: 0; left: 0;
    background: var(--accent-gradient);
    border-radius: var(--radius-pill);
    transition: width var(--transition-fast) ease;
  }

  .slider-input {
    position: relative;
    z-index: 1;
    width: 100%;
    height: 100%;
    appearance: none;
    -webkit-appearance: none;
    background: transparent;
    cursor: pointer;
    margin: 0;
    opacity: 0;
  }
  /* Visible thumb via the fill — the input itself is invisible,
     only the painted track + fill give the feedback. */
</style>