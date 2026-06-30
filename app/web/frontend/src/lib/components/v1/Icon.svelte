<!--
  Icon — wraps SVG content with locked stroke width and optical sizing.

  Per spec §7:
    - 1.25px line stroke, perfectly geometric, slightly rounded joins
    - NEVER filled, NEVER duotone
    - 16px in chrome, 20px in command overlay, 24px in empty states
    - Optical sizing: 16px icons next to 14px text need to render at 16px
      to match perceived weight
    - 1px breathing room on every side (handled by SVG sizing + parent padding)

  Usage:
    <Icon size="md">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.25"
           stroke-linecap="round" stroke-linejoin="round">
        <path d="..." />
      </svg>
    </Icon>

  Props:
    size     — 'xs' (12px) | 'sm' (16px) | 'md' (20px) | 'lg' (24px) | 'xl' (32px)
    label    — accessible label (when omitted, decorative; aria-hidden applied)
-->
<script lang="ts">
  interface Props {
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
    label?: string;
    children?: import('svelte').Snippet;
  }

  let { size = 'md', label, children }: Props = $props();

  const SIZE_PX: Record<NonNullable<Props['size']>, number> = {
    xs: 12,
    sm: 16,
    md: 20,
    lg: 24,
    xl: 32,
  };
</script>

<span
  class="icon icon--{size}"
  style="--icon-size: {SIZE_PX[size]}px;"
  role={label ? 'img' : undefined}
  aria-label={label}
  aria-hidden={label ? undefined : 'true'}
>
  {@render children?.()}
</span>

<style>
  .icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: var(--icon-size);
    height: var(--icon-size);
    flex-shrink: 0;
    color: inherit; /* icons inherit text color from parent */
  }

  /* SVG sizing — locked optical rules */
  .icon :global(svg) {
    width: 100%;
    height: 100%;
    display: block;
    /* Locked stroke-width: 1.25px (set on individual SVG elements in icon
       source files, but this provides a fallback for unstyled SVGs). */
    stroke-width: 1.25;
    vector-effect: non-scaling-stroke;
  }

  /* Filled icons (rare, used only for active/selected state).
     Set class="icon--filled" on the SVG inside if needed. */
  .icon :global(svg.icon--filled) {
    fill: currentColor;
    stroke: none;
  }
</style>