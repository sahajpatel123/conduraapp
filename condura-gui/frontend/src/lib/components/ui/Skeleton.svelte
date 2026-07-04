<script lang="ts">
  interface Props {
    width?: string
    height?: string
    rounded?: 'sm' | 'md' | 'lg' | 'pill'
    count?: number
  }

  let { width = '100%', height = '14px', rounded = 'sm', count = 1 }: Props = $props()

  const items = $derived(Array.from({ length: count }, (_, i) => i))
</script>

{#each items as i (i)}
  <span
    class="skeleton skel-{rounded}"
    style:width
    style:height
    style:--anim-delay="{i * 80}ms"
  ></span>
{/each}

<style>
  .skeleton {
    display: inline-block;
    background: linear-gradient(
      90deg,
      var(--surface-2) 0%,
      var(--surface-3) 50%,
      var(--surface-2) 100%
    );
    background-size: 200% 100%;
    animation: shimmer 1.6s ease-in-out infinite;
    animation-delay: var(--anim-delay, 0ms);
    vertical-align: middle;
  }
  .skel-sm   { border-radius: var(--radius-sm); }
  .skel-md   { border-radius: var(--radius-md); }
  .skel-lg   { border-radius: var(--radius-lg); }
  .skel-pill { border-radius: var(--radius-pill); }
</style>