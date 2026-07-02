<script lang="ts">
  /**
   * MagneticButton — Pointer-distance spring pull effect.
   * 
   * The button subtly shifts toward the pointer within a radius
   * threshold, creating a magnetic attraction feel. The pollen
   * glow intensifies as the pointer approaches.
   * 
   * Inspired by the website's MagneticButton component.
   */
  import './living-paper.css'

  type Variant = 'primary' | 'secondary' | 'ghost' | 'danger'

  interface Props {
    variant?: Variant
    size?: 'sm' | 'md' | 'lg'
    /** Magnetic pull strength (0-1). 0 = no magnetism */
    strength?: number
    /** Radius in px within which magnetism activates */
    radius?: number
    disabled?: boolean
    loading?: boolean
    fullWidth?: boolean
    onclick?: () => void
    class?: string
    children?: import('svelte').Snippet
    style?: string
    type?: string
  }

  let {
    variant = 'primary',
    size = 'md',
    strength = 0.32,
    radius = 130,
    disabled = false,
    loading = false,
    fullWidth = false,
    onclick,
    class: className = '',
    children,
    style = '',
    type = 'button',
  }: Props = $props()

  let btnEl = $state<HTMLButtonElement | null>(null)
  let offsetX = $state(0)
  let offsetY = $state(0)
  let isNear = $state(false)
  let rafId = $state(0)

  function onPointerMove(e: PointerEvent) {
    if (!btnEl || disabled || strength === 0) return
    const rect = btnEl.getBoundingClientRect()
    const cx = rect.left + rect.width / 2
    const cy = rect.top + rect.height / 2
    const dx = e.clientX - cx
    const dy = e.clientY - cy
    const dist = Math.sqrt(dx * dx + dy * dy)
    isNear = dist < radius
    if (isNear) {
      const pull = (1 - dist / radius) * strength
      offsetX = dx * pull
      offsetY = dy * pull
    } else {
      offsetX = 0
      offsetY = 0
    }
  }

  function onPointerLeave() {
    offsetX = 0
    offsetY = 0
    isNear = false
  }

  $effect(() => {
    const rm = window.matchMedia('(prefers-reduced-motion: reduce)')
    if (rm.matches) return
    // Track pointer over the button
    return () => { cancelAnimationFrame(rafId) }
  })

  const variantStyles: Record<Variant, string> = {
    primary: `
      background: var(--lp-synapse);
      color: var(--lp-paper);
      border: none;
    `,
    secondary: `
      background: transparent;
      color: var(--lp-ink);
      border: 1px solid var(--lp-ink-ghost);
    `,
    ghost: `
      background: transparent;
      color: var(--lp-ink-mute);
      border: none;
    `,
    danger: `
      background: var(--lp-danger);
      color: var(--lp-paper);
      border: none;
    `,
  }

  const sizeStyles: Record<string, string> = {
    sm: 'padding: 6px 14px; font-size: var(--lp-text-body-sm);',
    md: 'padding: 10px 20px; font-size: var(--lp-text-body);',
    lg: 'padding: 14px 28px; font-size: var(--lp-text-body);',
  }
</script>

<button
  bind:this={btnEl}
  type={type}
  class="lp lp-focus {className}"
  class:lp-magnetic-near={isNear}
  disabled={disabled || loading}
  onclick={onclick}
  onpointermove={onPointerMove}
  onpointerleave={onPointerLeave}
  style="
    position: relative;
    overflow: hidden;
    border-radius: var(--lp-radius-sm);
    cursor: pointer;
    font-family: var(--lp-font-sans);
    font-weight: 500;
    letter-spacing: -0.01em;
    line-height: 1;
    white-space: nowrap;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    isolation: isolate;
    transition: transform var(--lp-dur-normal) var(--lp-ease-spring),
                box-shadow var(--lp-dur-normal) var(--lp-ease-thread),
                opacity var(--lp-dur-fast) ease;
    transform: translate({offsetX}px, {offsetY}px);
    width: {fullWidth ? '100%' : 'auto'};
    opacity: {disabled ? 0.5 : 1};
    {variantStyles[variant]}
    {sizeStyles[size]}
    {style}
  "
>
  {#if loading}
    <span
      class="lp-magnetic-spinner"
      style="
        width: 14px; height: 14px;
        border: 2px solid currentColor;
        border-top-color: transparent;
        border-radius: 50%;
        animation: lp-spin 0.6s linear infinite;
      "
    ></span>
  {/if}
  {#if children}{@render children()}{/if}

  <!-- Pollen glow ring that appears on hover -->
  <span
    class="lp-magnetic-glow"
    style="
      position: absolute;
      inset: -2px;
      border-radius: inherit;
      border: 2px solid var(--lp-pollen-glow);
      opacity: {isNear && !disabled ? 0.3 : 0};
      transition: opacity var(--lp-dur-normal) ease;
      pointer-events: none;
    "
  ></span>
</button>

<style>
  @keyframes lp-spin {
    to { transform: rotate(360deg); }
  }

  .lp-magnetic-near:not([disabled]) {
    box-shadow: var(--lp-shadow-raise);
  }

  .lp-magnetic-near:not([disabled]):active {
    transform: scale(0.97) !important;
  }
</style>
