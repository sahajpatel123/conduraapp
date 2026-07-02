<!--
  Icon — a single SVG icon component.

  Used by all v1 surfaces. Per spec §7:
    - 1.25px line stroke (never filled except in active/selected state)
    - Geometric, slightly rounded line joins (1.5-2px)
    - 16px in chrome, 20px in command overlay, 24px in empty states
    - Tabular alignment with text (optically sized)
    - Padding: 1px breathing room on every side

  Usage:
    <Icon name="chat" size="md" />
    <Icon name="settings" size="lg" />

  Props:
    name  — kebab-case identifier from the icon set
    size  — 'xs' (12px) | 'sm' (16px) | 'md' (20px) | 'lg' (24px) | 'xl' (32px)
    stroke — override stroke width (defaults to 1.25)
-->
<script lang="ts">
  /**
   * The complete Synaptic v1 icon set. Geometric, 1.25px stroke, rounded joins.
   * Each icon is drawn in a 24x24 viewBox so it scales cleanly.
   */
  import { ICON_PATHS } from './paths';

  export type IconName =
    | 'chat'
    | 'audit'
    | 'replay'
    | 'hub'
    | 'sync'
    | 'skills'
    | 'channels'
    | 'delegation'
    | 'settings'
    | 'about'
    | 'home'
    | 'send'
    | 'pause'
    | 'undo'
    | 'pin'
    | 'plus'
    | 'mic'
    | 'mic-off'
    | 'search'
    | 'file'
    | 'folder'
    | 'mail'
    | 'calendar'
    | 'check'
    | 'x'
    | 'arrow-up'
    | 'arrow-down'
    | 'arrow-left'
    | 'arrow-right'
    | 'chevron-up'
    | 'chevron-down'
    | 'chevron-left'
    | 'chevron-right'
    | 'more'
    | 'history'
    | 'sparkle'
    | 'command'
    | 'eye'
    | 'eye-off'
    | 'lock'
    | 'power'
    | 'bell'
    | 'star'
    | 'heart'
    | 'trash'
    | 'edit'
    | 'external-link'
    | 'menu'
    | 'close'
    | 'play'
    | 'plus-circle'
    | 'globe'
    | 'moon'
    | 'sun';

  interface Props {
    name: IconName;
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
    stroke?: number;
    'aria-label'?: string;
  }

  let { name, size = 'md', stroke = 1.25, 'aria-label': ariaLabel }: Props = $props();

  const SIZE_PX = { xs: 12, sm: 16, md: 20, lg: 24, xl: 32 } as const;
  let px = $derived(SIZE_PX[size]);
</script>

<svg
  class="icon"
  width={px}
  height={px}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width={stroke}
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden={ariaLabel ? undefined : 'true'}
  aria-label={ariaLabel}
  role={ariaLabel ? 'img' : undefined}
>
  <!-- eslint-disable-next-line svelte/no-at-html-tags -->
  {@html ICON_PATHS[name]}
</svg>

<style>
  .icon {
    display: inline-block;
    flex-shrink: 0;
    /* Optical alignment: line icons need to sit ~1px above text baseline */
    vertical-align: -1px;
    color: inherit;
  }
</style>