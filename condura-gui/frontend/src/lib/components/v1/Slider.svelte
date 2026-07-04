<!--
  Slider — value selector. Used for sensitivity, strength, etc.

  Per spec §9.5 (Adaptive strength): off / cautious / balanced / aggressive.
  Other sliders: wake-word sensitivity, autonomy dials, font size.

  Props:
    value    — bound number (Svelte 5 via $bindable)
    min      — minimum value
    max      — maximum value
    step     — step size
    label    — visible label
    showValue — show current value on the right
    unit     — unit suffix (e.g., "%", "ms")
-->
<script lang="ts">
  interface Props {
    value?: number;
    min?: number;
    max?: number;
    step?: number;
    label?: string;
    showValue?: boolean;
    unit?: string;
    onchange?: (v: number) => void;
  }

  let {
    value = $bindable(0),
    min = 0,
    max = 100,
    step = 1,
    label,
    showValue = true,
    unit = '',
    onchange,
  }: Props = $props();

  let percent = $derived(((value - min) / (max - min)) * 100);
</script>

<label class="slider">
  {#if label}
    <div class="slider__head">
      <span class="slider__label">{label}</span>
      {#if showValue}
        <span class="slider__value">
          <span class="slider__value-num">{value}</span>
          {#if unit}<span class="slider__value-unit">{unit}</span>{/if}
        </span>
      {/if}
    </div>
  {/if}

  <div class="slider__track">
    <div class="slider__fill" style="--slider-pct: {percent}%"></div>
    <input
      class="slider__input"
      type="range"
      {min}
      {max}
      {step}
      bind:value
      onchange={() => onchange?.(value)}
      aria-label={label}
    />
  </div>
</label>

<style>
  .slider {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    width: 100%;
  }

  .slider__head {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
  }

  .slider__label {
    font-size: var(--text-body-sm-size);
    color: var(--content-primary);
  }

  .slider__value {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-secondary);
    font-variant-numeric: tabular-nums;
  }

  .slider__value-unit {
    color: var(--content-tertiary);
    margin-left: 2px;
  }

  .slider__track {
    position: relative;
    height: 24px;
    display: flex;
    align-items: center;
  }

  .slider__fill {
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    width: var(--slider-pct);
    height: 4px;
    background-color: var(--plum-600);
    border-radius: var(--radius-pill);
    transition: width var(--duration-fast) var(--ease-standard);
    pointer-events: none;
  }

  .slider__track::before {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    top: 50%;
    transform: translateY(-50%);
    height: 4px;
    background-color: var(--ink-cool-100);
    border-radius: var(--radius-pill);
    pointer-events: none;
  }

  .slider__input {
    position: relative;
    width: 100%;
    height: 24px;
    appearance: none;
    background: transparent;
    cursor: pointer;
    margin: 0;
    padding: 0;
  }

  /* WebKit thumb */
  .slider__input::-webkit-slider-thumb {
    appearance: none;
    width: 16px;
    height: 16px;
    background-color: var(--paper-warm-0);
    border: 2px solid var(--plum-600);
    border-radius: var(--radius-pill);
    cursor: grab;
    transition: transform var(--duration-fast) var(--ease-standard);
  }
  .slider__input:hover::-webkit-slider-thumb {
    transform: scale(1.1);
  }
  .slider__input:active::-webkit-slider-thumb {
    cursor: grabbing;
  }

  /* Firefox thumb */
  .slider__input::-moz-range-thumb {
    width: 16px;
    height: 16px;
    background-color: var(--paper-warm-0);
    border: 2px solid var(--plum-600);
    border-radius: var(--radius-pill);
    cursor: grab;
  }

  .slider__input:focus-visible {
    outline: none;
  }
  .slider__input:focus-visible::-webkit-slider-thumb {
    box-shadow: 0 0 0 3px var(--plum-100);
  }
</style>