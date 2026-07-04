<script lang="ts">
  interface Props {
    value?: number  /* 0..100 — required when not indeterminate */
    indeterminate?: boolean
    label?: string
    tone?: 'accent' | 'success' | 'warn' | 'error'
  }

  let { value = 0, indeterminate = false, label, tone = 'accent' }: Props = $props()

  const clamped = $derived(indeterminate ? 0 : Math.max(0, Math.min(100, value)))
</script>

<div class="progress-wrap" role="progressbar" aria-valuenow={indeterminate ? undefined : clamped} aria-valuemin="0" aria-valuemax="100">
  {#if label}<span class="progress-label">{label}</span>{/if}
  <span class="progress-track">
    {#if indeterminate}
      <span class="progress-fill progress-indeterminate progress-{tone}"></span>
    {:else}
      <span
        class="progress-fill progress-{tone}"
        style:width="{clamped}%"
      ></span>
    {/if}
  </span>
</div>

<style>
  .progress-wrap {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .progress-label {
    font-family: var(--font-mono);
    font-size: var(--size-2xs);
    color: var(--text-muted);
    letter-spacing: var(--tracking-wider);
    text-transform: uppercase;
  }

  .progress-track {
    position: relative;
    height: 4px;
    background: var(--surface-3);
    border-radius: var(--radius-pill);
    overflow: hidden;
  }

  .progress-fill {
    position: absolute;
    top: 0; bottom: 0; left: 0;
    border-radius: inherit;
    transition: width var(--transition-base) var(--ease-out-quart);
  }
  .progress-accent  { background: var(--accent-gradient); box-shadow: 0 0 12px var(--accent-glow); }
  .progress-success { background: var(--success); }
  .progress-warn    { background: var(--warn); }
  .progress-error   { background: var(--error); }

  .progress-indeterminate {
    width: 40% !important;
    animation: progress-slide 1.6s var(--ease-in-out-quart) infinite;
  }
  @keyframes progress-slide {
    0%   { left: -40%; }
    100% { left: 100%; }
  }
</style>