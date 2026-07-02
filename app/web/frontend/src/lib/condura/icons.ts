/**
 * Condura Glyph — single source of truth for the icon library.
 *
 * Design contract (per DESIGNLANG.md §7):
 *   - Grid:    24u (viewBox 0 0 24 24)
 *   - Stroke:  1.5 (uniform; overridable per <Glyph stroke={n}>)
 *   - Caps/joins: round
 *   - Single metaphor per icon. No doubles, no flourishes.
 *   - One path (or a small set of paths) per icon. No filled shapes
 *     unless the metaphor demands it (dot, stop, menu kebab).
 *   - Naming: kebab-case (`chevron-right`, `theme-sun`, `kill-switch`).
 *
 * Consumers:
 *   - <Glyph name="..."/> renders PATHS[name] as SVG.
 *   - Settings / docs / palette enumeration: ICONS array, ICON_NAMES,
 *     ICONS_BY_CATEGORY.
 *
 * To add an icon:
 *   1. Add it to DEFINITIONS with the kebab-case name + path string.
 *   2. Pick a category: nav | action | state | theme | media.
 *   3. Keep stroke-only paths unless the metaphor demands fill
 *      (dots, stop, kill-switch).
 */

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type IconCategory = 'nav' | 'action' | 'state' | 'theme' | 'media';

export interface IconMeta {
  name: string;
  path: string;
  category: IconCategory;
}

interface IconDef {
  path: string;
  category: IconCategory;
}

// ---------------------------------------------------------------------------
// Definitions — one entry per canonical icon.
// Each `path` is the inner SVG markup rendered inside the Glyph <svg>.
// Stroke defaults to currentColor at 1.5; fills use currentColor only where
// the metaphor demands it (dot, dot-active, menu kebab, stop, kill-switch).
// ---------------------------------------------------------------------------

const DEFINITIONS: Record<string, IconDef> = {
  // -----------------------------------------------------------------------
  // NAV — sidebar / nav rail / command palette routes
  // -----------------------------------------------------------------------
  chat: {
    category: 'nav',
    // Speech bubble, hairline body + tail, no fill.
    path: '<path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>',
  },
  'skills': {
    category: 'nav',
    // 4-point sparkle + 4 small accent ticks (the "skills" metaphor: insight).
    path: '<path d="M12 3l1.8 4.2L18 9l-4.2 1.8L12 15l-1.8-4.2L6 9l4.2-1.8z"/><path d="M18 3v2M21 6h-2M6 17v2M3 20h2"/>',
  },
  'hub': {
    category: 'nav',
    // Three concentric arcs converging on a central node — "nexus".
    path: '<circle cx="12" cy="12" r="2"/><path d="M12 14v8"/><path d="M5 10.5A7 7 0 0 1 12 4a7 7 0 0 1 7 6.5"/><path d="M3 7.5A10 10 0 0 1 12 2a10 10 0 0 1 9 5.5"/>',
  },
  'channels': {
    category: 'nav',
    // Broadcast — two arcs + dot at the source.
    path: '<path d="M5 12a7 7 0 0 1 14 0"/><path d="M8 12a4 4 0 0 1 8 0"/><circle cx="12" cy="12" r="1"/>',
  },
  'audit': {
    category: 'nav',
    // Document with corner fold + check inside — "file ok".
    path: '<path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z"/><path d="M14 3v5h5"/><path d="M9 14l2 2 4-4"/>',
  },
  'about': {
    category: 'nav',
    // Circle + lowercase i — universal "about" mark.
    path: '<circle cx="12" cy="12" r="9"/><path d="M12 11v5"/><path d="M12 8h.01"/>',
  },
  'settings': {
    category: 'nav',
    // Gear: center hole + 8 cardinal teeth on the 24u grid.
    path: '<circle cx="12" cy="12" r="3"/><path d="M12 2v3M12 19v3M2 12h3M19 12h3M4.2 4.2l2.1 2.1M17.7 17.7l2.1 2.1M4.2 19.8l2.1-2.1M17.7 6.3l2.1-2.1"/>',
  },
  'sync': {
    category: 'nav',
    // Two semicircle arcs with arrow tips — the refresh idiom.
    path: '<path d="M3 12a9 9 0 0 1 15.5-6.3L21 8"/><path d="M21 3v5h-5"/><path d="M21 12a9 9 0 0 1-15.5 6.3L3 16"/><path d="M3 21v-5h5"/>',
  },
  'delegation': {
    category: 'nav',
    // Center node + 8 cardinal rays — delegation = fan-out from a point.
    path: '<circle cx="12" cy="12" r="2"/><path d="M12 3v3M12 18v3M3 12h3M18 12h3M5.6 5.6l2.1 2.1M16.3 16.3l2.1 2.1M5.6 18.4l2.1-2.1M16.3 7.7l2.1-2.1"/>',
  },
  'account': {
    category: 'nav',
    // Head + shoulders silhouette.
    path: '<circle cx="12" cy="7" r="4"/><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>',
  },
  'kill-switch': {
    category: 'nav',
    // Shield with an X inside — "stop this NOW". One metaphor: emergency stop.
    path: '<path d="M12 3l8 3v6c0 4.5-3.4 7.8-8 9-4.6-1.2-8-4.5-8-9V6z"/><path d="M9.5 9.5l5 5"/><path d="M14.5 9.5l-5 5"/>',
  },
  'replay': {
    category: 'nav',
    // Stage frame + centered play triangle (hairline).
    path: '<rect x="3" y="6" width="18" height="12" rx="2"/><path d="M10 9.5v5l4-2.5z"/>',
  },

  // -----------------------------------------------------------------------
  // ACTION — generic UI controls
  // -----------------------------------------------------------------------
  'send': {
    category: 'action',
    // Paper plane — fold + outline, no fill.
    path: '<path d="M22 2L11 13"/><path d="M22 2l-7 20-4-9-9-4 20-7z"/>',
  },
  'close': {
    category: 'action',
    path: '<path d="M18 6L6 18"/><path d="M6 6l12 12"/>',
  },
  'back': {
    category: 'action',
    // Left-pointing arrow with shaft.
    path: '<path d="M19 12H5"/><path d="M11 6l-6 6 6 6"/>',
  },
  'check': {
    category: 'action',
    path: '<path d="M20 6L9 17l-5-5"/>',
  },
  'plus': {
    category: 'action',
    path: '<path d="M12 5v14"/><path d="M5 12h14"/>',
  },
  'search': {
    category: 'action',
    // Magnifier: lens circle + handle.
    path: '<circle cx="11" cy="11" r="7"/><path d="M21 21l-4.3-4.3"/>',
  },
  'command': {
    category: 'action',
    // ⌘ symbol: four corner arcs (one per loop). Loops overlap the
    // implicit central body — the outline reads as the ⌘ glyph.
    path: '<path d="M15 6V3a3 3 0 1 1 3 3h-3"/><path d="M9 6V3a3 3 0 1 0-3 3h3"/><path d="M15 18v3a3 3 0 1 0 3-3h-3"/><path d="M9 18v3a3 3 0 1 1-3-3h3"/>',
  },
  'menu': {
    category: 'action',
    // Kebab — 3 vertical dots. Filled at r=1 so they read at small sizes.
    path: '<circle cx="12" cy="5" r="1" fill="currentColor" stroke="none"/><circle cx="12" cy="12" r="1" fill="currentColor" stroke="none"/><circle cx="12" cy="19" r="1" fill="currentColor" stroke="none"/>',
  },
  'trash': {
    category: 'action',
    // Lid bar + handle + body + inner dividers.
    path: '<path d="M4 7h16"/><path d="M9 7V4h6v3"/><path d="M6 7l1 13a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-13"/><path d="M10 11v6"/><path d="M14 11v6"/>',
  },
  'shield': {
    category: 'action',
    // Plain shield outline — "general protection" metaphor.
    path: '<path d="M12 3l8 3v6c0 4.5-3.4 7.8-8 9-4.6-1.2-8-4.5-8-9V6z"/>',
  },
  'bolt': {
    category: 'action',
    // Lightning bolt — closed stroke (the metaphor demands a filled look).
    path: '<path d="M13 2L4 14h7l-1 8 10-12h-7z"/>',
  },
  'spark': {
    category: 'action',
    // 4-point asterisk accent — small inline spark.
    path: '<path d="M12 3v6"/><path d="M12 15v6"/><path d="M3 12h6"/><path d="M15 12h6"/>',
  },
  'chevron-right': {
    category: 'action',
    path: '<path d="M9 6l6 6-6 6"/>',
  },
  'chevron-down': {
    category: 'action',
    path: '<path d="M6 9l6 6 6-6"/>',
  },
  'chevron-left': {
    category: 'action',
    path: '<path d="M15 6l-6 6 6 6"/>',
  },
  'info': {
    category: 'action',
    // Alias visual of `about`. Kept separate for semantic clarity.
    path: '<circle cx="12" cy="12" r="9"/><path d="M12 11v5"/><path d="M12 8h.01"/>',
  },
  'warning': {
    category: 'action',
    // Triangle with rounded corners + exclamation.
    path: '<path d="M10.3 3.86L1.8 16.36A2 2 0 0 0 3.5 19h17a2 2 0 0 0 1.7-2.64L13.7 3.86a2 2 0 0 0-3.4 0z"/><path d="M12 9v4"/><path d="M12 17h.01"/>',
  },
  'stop': {
    category: 'action',
    // Stop square — filled (the metaphor demands it).
    path: '<rect x="6" y="6" width="12" height="12" rx="2" fill="currentColor" stroke="none"/>',
  },
  'book': {
    category: 'action',
    // Open book: spine + covers.
    path: '<path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>',
  },
  'circle': {
    category: 'action',
    // Empty outline circle — neutral bullet / placeholder.
    path: '<circle cx="12" cy="12" r="9"/>',
  },
  'mail': {
    category: 'action',
    // Envelope — rectangle body + flap V.
    path: '<rect x="3" y="5" width="18" height="14" rx="2"/><path d="M3 7l9 6 9-6"/>',
  },
  'heart': {
    category: 'action',
    // Heart — two arcs converging at the bottom point.
    path: '<path d="M20.8 4.6a5.5 5.5 0 0 0-7.8 0L12 5.6l-1-1a5.5 5.5 0 0 0-7.8 7.8l1 1L12 21l7.8-7.8 1-1a5.5 5.5 0 0 0 0-7.6z"/>',
  },
  'lifebuoy': {
    category: 'action',
    // Lifebuoy — outer ring + inner ring + 4 crossbars (N/S/E/W).
    path: '<circle cx="12" cy="12" r="9"/><circle cx="12" cy="12" r="4"/><path d="M5.6 5.6l3.7 3.7"/><path d="M14.7 14.7l3.7 3.7"/><path d="M5.6 18.4l3.7-3.7"/><path d="M14.7 9.3l3.7-3.7"/>',
  },
  'google': {
    category: 'action',
    // Google mark — single-stroke G silhouette (brand-agnostic, no
    // multi-color fill). Per MOAT §8.2 we keep monochrome.
    path: '<path d="M21 12a9 9 0 1 1-3.6-7.2"/><path d="M21 12c0-.7-.1-1.4-.2-2H12v4h5.1c-.2 1.2-.9 2.2-2 2.9v2.4h3.2c1.9-1.7 2.7-4.2 2.7-7.3z"/>',
  },
  'github': {
    category: 'action',
    // GitHub mark — single-stroke octocat silhouette. Per MOAT §8.2
    // monochrome (no brand fill).
    path: '<path d="M12 2a10 10 0 0 0-3.2 19.5c.5.1.7-.2.7-.5v-1.7c-2.8.6-3.4-1.3-3.4-1.3-.5-1.1-1.1-1.5-1.1-1.5-.9-.6.1-.6.1-.6 1 .1 1.5 1 1.5 1 .9 1.5 2.3 1.1 2.9.8.1-.7.4-1.1.7-1.4-2.2-.3-4.6-1.1-4.6-5 0-1.1.4-2 1-2.7-.1-.3-.4-1.3.1-2.7 0 0 .8-.3 2.7 1a9.4 9.4 0 0 1 5 0c1.9-1.3 2.7-1 2.7-1 .5 1.4.2 2.4.1 2.7.6.7 1 1.6 1 2.7 0 3.9-2.4 4.7-4.6 5 .4.3.7.9.7 1.8v2.7c0 .3.2.6.7.5A10 10 0 0 0 12 2z"/>',
  },

  // -----------------------------------------------------------------------
  // STATE — status indicators (dot family)
  // -----------------------------------------------------------------------
  'dot': {
    category: 'state',
    // Small filled dot. Filled because the metaphor demands it.
    path: '<circle cx="12" cy="12" r="2" fill="currentColor" stroke="none"/>',
  },
  'dot-active': {
    category: 'state',
    // Larger filled dot — "on" state, more presence than dot.
    path: '<circle cx="12" cy="12" r="3" fill="currentColor" stroke="none"/>',
  },

  // -----------------------------------------------------------------------
  // THEME — theme picker set (sun / auto / moon).
  // These are the segmented-control icons used in Settings.
  // -----------------------------------------------------------------------
  'theme-sun': {
    category: 'theme',
    // Full sun: center disc + 8 rays on the 24u grid.
    path: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="M4.9 4.9l1.4 1.4"/><path d="M17.7 17.7l1.4 1.4"/><path d="M4.9 19.1l1.4-1.4"/><path d="M17.7 6.3l1.4-1.4"/>',
  },
  'theme-auto': {
    category: 'theme',
    // "Auto" — half-circle (top) + horizontal divider. Reads as "day/dusk".
    path: '<path d="M3 12a9 9 0 0 1 18 0"/><path d="M3 12h18"/>',
  },
  'theme-moon': {
    category: 'theme',
    // Crescent moon (single closed path).
    path: '<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>',
  },

  // -----------------------------------------------------------------------
  // MEDIA — input/output affordances
  // -----------------------------------------------------------------------
  'mic': {
    category: 'media',
    // Capsule + U-cradle + stem.
    path: '<rect x="9" y="3" width="6" height="11" rx="3"/><path d="M5 11a7 7 0 0 0 14 0"/><path d="M12 18v3"/>',
  },
  'key': {
    category: 'media',
    // Keyboard — rounded frame + 4 small keys + spacebar.
    // The "Summon" / hotkey metaphor: a key you press.
    path: '<rect x="2" y="6" width="20" height="12" rx="2"/><path d="M6 10h2"/><path d="M10 10h2"/><path d="M14 10h2"/><path d="M18 10h2"/><path d="M7 14h10"/>',
  },
  'power': {
    category: 'media',
    // Power glyph — open ring + vertical stroke.
    path: '<path d="M18.36 6.64a9 9 0 1 1-12.73 0"/><path d="M12 2v10"/>',
  },
};

// ---------------------------------------------------------------------------
// Legacy / alias names — kept so existing component code does not break.
// Each alias points to a canonical name and inherits its category.
// To retire an alias, delete the entry here and update any callers.
// ---------------------------------------------------------------------------

const ALIASES: Record<string, keyof typeof DEFINITIONS> = {
  // Original camelCase form of new kebab-case names.
  chevronRight: 'chevron-right',
  chevronDown: 'chevron-down',
  chevronLeft: 'chevron-left',
  killSwitch: 'kill-switch',
  themeSun: 'theme-sun',
  themeAuto: 'theme-auto',
  themeMoon: 'theme-moon',
  // Existing shorter forms that resolve to canonical icons.
  gear: 'settings',
  sparkle: 'skills',
  sun: 'theme-sun',
  auto: 'theme-auto',
  moon: 'theme-moon',
  x: 'close',
  // "info" and "about" are now distinct canonical entries (both exist).
};

// ---------------------------------------------------------------------------
// Public exports
// ---------------------------------------------------------------------------

/** All canonical icons, ordered for documentation/enumeration. */
export const ICONS: IconMeta[] = Object.entries(DEFINITIONS).map(([name, def]) => ({
  name,
  path: def.path,
  category: def.category,
}));

/** All resolvable names — canonical + aliases — as a flat list. */
export const ICON_NAMES: string[] = [
  ...Object.keys(DEFINITIONS),
  ...Object.keys(ALIASES),
];

/** Lookup by name. Aliases are resolved to their canonical paths. */
export const PATHS: Record<string, string> = (() => {
  const base: Record<string, string> = {};
  for (const [name, def] of Object.entries(DEFINITIONS)) base[name] = def.path;
  for (const [alias, target] of Object.entries(ALIASES)) {
    base[alias] = base[target];
  }
  return base;
})();

/** Group icons by category for documentation / settings listings. */
export const ICONS_BY_CATEGORY: Record<IconCategory, IconMeta[]> = {
  nav: [],
  action: [],
  state: [],
  theme: [],
  media: [],
};
for (const icon of ICONS) ICONS_BY_CATEGORY[icon.category].push(icon);

/** Returns true if the given name resolves to an icon (canonical or alias). */
export function hasIcon(name: string): boolean {
  return Object.prototype.hasOwnProperty.call(PATHS, name);
}

/** Returns the canonical kebab-case name for a given name (alias or not). */
export function canonicalName(name: string): string {
  return ALIASES[name] ?? name;
}