<!--
  Switch — Condura v2 hardware-honest toggle.

  Per the spec: "Toggles are hardware-honest — a 28px switch with a
  1px groove, not a flat rectangle."

  The switch is a 28×16 trough with a 12px thumb. The thumb drags
  between two slots and snaps. There IS a 1px groove painted into
  the trough — not faked glow, not a flat rectangle. The thumb
  has a real shadow that lifts as it animates.

  Props:
    on:        boolean — current state
    onclick:   () => void
    label?:    string  — accessible label (aria-label)
    disabled?: boolean
-->
<script lang="ts">
  let {
    on = false as boolean,
    onclick = undefined as (() => void) | undefined,
    label = undefined as string | undefined,
    disabled = false as boolean,
  }: {
    on: boolean
    onclick?: () => void
    label?: string
    disabled?: boolean
  } = $props()

  let toggling = $state(false)
  function handle() {
    if (disabled || !onclick) return
    toggling = true
    onclick()
    // unlock after the animation window
    setTimeout(() => { toggling = false }, 200)
  }
</script>

<button
  data-v2-switch
  data-on={on}
  disabled={disabled}
  aria-label={label}
  aria-pressed={on}
  type="button"
  onclick={handle}
>
  <!-- The 1px groove is real, painted via box-shadow inset -->
  <span class="trough">
    <span class="thumb"></span>
  </span>
</button>

<style>
  [data-v2-switch] {
    /* Hit area is larger than visual; click anywhere in this rect */
    width: 32px;
    height: 20px;
    background: transparent;
    border: none;
    padding: 0;
    margin: 0;
    cursor: pointer;
    display: inline-grid;
    place-items: center;
    flex-shrink: 0;
  }
  [data-v2-switch]:disabled {
    cursor: not-allowed;
    opacity: 0.5;
  }

  [data-v2-switch] .trough {
    width: 28px;
    height: 16px;
    border-radius: var(--v2-radius-pill);
    background: var(--v2-paper-2);
    position: relative;
    transition: background-color var(--v2-dur-fast) var(--v2-ease-out-soft);
    /* The 1px groove — inset, real, hairline-sharp */
    box-shadow: inset 0 0 0 1px var(--v2-rule);
  }
  [data-v2-switch][data-on='true'] .trough {
    background: var(--v2-accent);
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--v2-accent) 80%, var(--v2-ink));
  }

  [data-v2-switch] .thumb {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 12px;
    height: 12px;
    border-radius: var(--v2-radius-pill);
    background: var(--v2-paper);
    box-shadow: var(--v2-shadow-1);
    transition:
      transform var(--v2-dur-mid) var(--v2-ease-spring),
      box-shadow   var(--v2-dur-mid) var(--v2-ease-settle);
  }
  [data-v2-switch][data-on='true'] .thumb {
    transform: translateX(12px);
    box-shadow: var(--v2-shadow-2);
  }
</style>
