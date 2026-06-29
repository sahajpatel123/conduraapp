<script lang="ts">
  interface Props {
    name: string
    src?: string
    size?: 'xs' | 'sm' | 'md' | 'lg'
    status?: 'online' | 'busy' | 'away' | 'offline'
  }

  let { name, src, size = 'md', status }: Props = $props()

  const initials = $derived(
    name
      .split(/\s+/)
      .map(p => p[0]?.toUpperCase() ?? '')
      .slice(0, 2)
      .join('') || '?'
  )

  function hueFor(seed: string): number {
    let h = 0
    for (let i = 0; i < seed.length; i++) h = (h * 31 + seed.charCodeAt(i)) >>> 0
    return h % 360
  }
  const hue = $derived(hueFor(name))
  const fallbackBg = $derived(`hsl(${hue} 38% 22%)`)
  const fallbackFg = $derived(`hsl(${hue} 60% 78%)`)
</script>

<span class="avatar avatar-{size}" style:--avatar-bg={fallbackBg} style:--avatar-fg={fallbackFg}>
  {#if src}
    <img {src} alt={name} class="avatar-img" />
  {:else}
    <span class="avatar-initials">{initials}</span>
  {/if}
  {#if status}<span class="avatar-status status-{status}"></span>{/if}
</span>

<style>
  .avatar {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: var(--avatar-bg, var(--surface-3));
    color: var(--avatar-fg, var(--text));
    font-family: var(--font-sans);
    font-weight: var(--weight-semibold);
    letter-spacing: var(--tracking-wide);
    flex-shrink: 0;
    overflow: hidden;
    border: 1px solid var(--border-strong);
  }
  .avatar-xs { width: 22px; height: 22px; font-size: 9px; }
  .avatar-sm { width: 28px; height: 28px; font-size: 11px; }
  .avatar-md { width: 36px; height: 36px; font-size: 13px; }
  .avatar-lg { width: 48px; height: 48px; font-size: 16px; }

  .avatar-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .avatar-initials {
    text-transform: uppercase;
  }

  .avatar-status {
    position: absolute;
    bottom: 0;
    right: 0;
    width: 30%;
    height: 30%;
    border-radius: 50%;
    border: 2px solid var(--surface-1);
  }
  .status-online  { background: var(--success); }
  .status-busy    { background: var(--error); }
  .status-away    { background: var(--warn); }
  .status-offline { background: var(--text-faint); }
</style>