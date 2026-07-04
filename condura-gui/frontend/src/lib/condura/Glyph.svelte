<script lang="ts">
  // Condura Glyph — single-stroke SVG icon set, currentColor, 1.5 stroke.
  // One source of truth for every icon in the app so chrome never uses
  // ambiguous Unicode glyphs (part of the anti-vibe-coded rules).
  //
  // Path data + categorization lives in `./icons.ts`. This file is the
  // renderer; icons.ts is the contract.
  import { PATHS } from './icons';

  let { name, size = 18, stroke = 1.5, class: cls = '' }: {
    name: string;
    size?: number;
    stroke?: number;
    class?: string;
  } = $props();
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width={stroke}
  stroke-linecap="round"
  stroke-linejoin="round"
  class={cls}
  aria-hidden="true"
>
  <!-- eslint-disable-next-line svelte/no-at-html-tags -- static trusted icon paths from ./icons.ts -->
  {@html PATHS[name] ?? ''}
</svg>