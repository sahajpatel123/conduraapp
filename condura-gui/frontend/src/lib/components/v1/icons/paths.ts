/**
 * Synaptic v1 icon path library.
 *
 * Each entry is the SVG inner content (just the shapes) for a 24x24 viewBox.
 * All icons follow the v1 iconography rules:
 *   - 1.25px stroke (handled by the Icon component)
 *   - Line icons, never filled (active state uses a fill variant if needed)
 *   - Geometric proportions, slightly rounded joins
 *   - Same visual weight throughout the set
 *
 * Drawing approach: each path uses M (move), L (line), C (curve), A (arc).
 * Coordinates are in 24x24 space. Padding inside the box: ~2px breathing room.
 */

export const ICON_PATHS = {
  // ── Navigation (sidebar) ────────────────────────────────────────
  chat: `
    <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
  `,
  audit: `
    <circle cx="11" cy="11" r="7"/>
    <path d="m21 21-4.3-4.3"/>
    <path d="M11 8v3"/>
    <path d="M11 14h.01"/>
  `,
  replay: `
    <path d="M3 12a9 9 0 1 0 9-9 9.74 9.74 0 0 0-6.74 2.74L3 8"/>
    <path d="M3 3v5h5"/>
  `,
  hub: `
    <circle cx="12" cy="12" r="3"/>
    <circle cx="4" cy="4" r="2"/>
    <circle cx="20" cy="4" r="2"/>
    <circle cx="4" cy="20" r="2"/>
    <circle cx="20" cy="20" r="2"/>
    <path d="M6 6 9.5 9.5"/>
    <path d="m14.5 14.5 3.5 3.5"/>
    <path d="M18 6 14.5 9.5"/>
    <path d="m9.5 14.5-3.5 3.5"/>
  `,
  sync: `
    <path d="M21 12a9 9 0 0 0-15-6.7L3 8"/>
    <path d="M3 3v5h5"/>
    <path d="M3 12a9 9 0 0 0 15 6.7l3-2.7"/>
    <path d="M21 21v-5h-5"/>
  `,
  skills: `
    <path d="m12 3-1.9 5.8a2 2 0 0 1-1.3 1.3L3 12l5.8 1.9a2 2 0 0 1 1.3 1.3L12 21l1.9-5.8a2 2 0 0 1 1.3-1.3L21 12l-5.8-1.9a2 2 0 0 1-1.3-1.3L12 3Z"/>
    <path d="m5 5 1.5 1.5"/>
    <path d="m18.5 5-1.5 1.5"/>
    <path d="m5 19 1.5-1.5"/>
    <path d="m18.5 19-1.5-1.5"/>
  `,
  channels: `
    <path d="M4 12h16"/>
    <path d="M4 6h16"/>
    <path d="M4 18h16"/>
    <circle cx="7" cy="6" r="1" fill="currentColor"/>
    <circle cx="11" cy="12" r="1" fill="currentColor"/>
    <circle cx="15" cy="18" r="1" fill="currentColor"/>
  `,
  delegation: `
    <path d="M12 2 4 6v6c0 5 3.5 9 8 10 4.5-1 8-5 8-10V6l-8-4Z"/>
    <path d="m9 12 2 2 4-4"/>
  `,
  settings: `
    <circle cx="12" cy="12" r="3"/>
    <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1Z"/>
  `,
  about: `
    <circle cx="12" cy="12" r="9"/>
    <path d="M12 8h.01"/>
    <path d="M11 12h1v4h1"/>
  `,
  home: `
    <path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z"/>
    <path d="M9 22V12h6v10"/>
  `,

  // ── Common actions ──────────────────────────────────────────────
  send: `
    <path d="m22 2-7 20-4-9-9-4 20-7Z"/>
    <path d="M22 2 11 13"/>
  `,
  pause: `
    <rect x="6" y="4" width="4" height="16" rx="1"/>
    <rect x="14" y="4" width="4" height="16" rx="1"/>
  `,
  undo: `
    <path d="M3 7v6h6"/>
    <path d="M21 17a9 9 0 0 0-9-9 9 9 0 0 0-6.7 2.8L3 13"/>
  `,
  pin: `
    <path d="M12 17v5"/>
    <path d="M9 10.76V5a3 3 0 0 1 6 0v5.76l3 1.93V14H6v-1.31l3-1.93Z"/>
  `,
  plus: `
    <path d="M12 5v14"/>
    <path d="M5 12h14"/>
  `,
  mic: `
    <rect x="9" y="2" width="6" height="13" rx="3"/>
    <path d="M19 10a7 7 0 0 1-14 0"/>
    <path d="M12 17v5"/>
  `,
  'mic-off': `
    <path d="m2 2 20 20"/>
    <path d="M18.89 13.23A7 7 0 0 0 19 10v-2"/>
    <path d="M5 10v2a7 7 0 0 0 11.95 4.95"/>
    <path d="M9 5a3 3 0 0 1 6 0v5"/>
    <path d="M12 17v4"/>
  `,
  search: `
    <circle cx="11" cy="11" r="7"/>
    <path d="m21 21-4.3-4.3"/>
  `,
  file: `
    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
    <path d="M14 2v6h6"/>
    <path d="M9 13h6"/>
    <path d="M9 17h4"/>
  `,
  folder: `
    <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
  `,
  mail: `
    <rect x="2" y="4" width="20" height="16" rx="2"/>
    <path d="m22 7-10 5L2 7"/>
  `,
  calendar: `
    <rect x="3" y="4" width="18" height="18" rx="2"/>
    <path d="M16 2v4"/>
    <path d="M8 2v4"/>
    <path d="M3 10h18"/>
  `,
  check: `
    <path d="M20 6 9 17l-5-5"/>
  `,
  x: `
    <path d="M18 6 6 18"/>
    <path d="m6 6 12 12"/>
  `,
  'arrow-up': `
    <path d="m5 12 7-7 7 7"/>
    <path d="M12 19V5"/>
  `,
  'arrow-down': `
    <path d="M12 5v14"/>
    <path d="m19 12-7 7-7-7"/>
  `,
  'arrow-left': `
    <path d="m12 19-7-7 7-7"/>
    <path d="M19 12H5"/>
  `,
  'arrow-right': `
    <path d="M5 12h14"/>
    <path d="m12 5 7 7-7 7"/>
  `,
  'chevron-up': `
    <path d="m18 15-6-6-6 6"/>
  `,
  'chevron-down': `
    <path d="m6 9 6 6 6-6"/>
  `,
  'chevron-left': `
    <path d="m15 18-6-6 6-6"/>
  `,
  'chevron-right': `
    <path d="m9 18 6-6-6-6"/>
  `,
  play: `
    <polygon points="6 4 20 12 6 20 6 4"/>
  `,
  more: `
    <circle cx="5" cy="12" r="1.5" fill="currentColor"/>
    <circle cx="12" cy="12" r="1.5" fill="currentColor"/>
    <circle cx="19" cy="12" r="1.5" fill="currentColor"/>
  `,
  history: `
    <path d="M3 12a9 9 0 1 0 9-9 9.74 9.74 0 0 0-6.74 2.74L3 8"/>
    <path d="M3 3v5h5"/>
    <path d="M12 7v5l4 2"/>
  `,
  sparkle: `
    <path d="M12 3v3"/>
    <path d="M12 18v3"/>
    <path d="M3 12h3"/>
    <path d="M18 12h3"/>
    <path d="m5.6 5.6 2.1 2.1"/>
    <path d="m16.3 16.3 2.1 2.1"/>
    <path d="m5.6 18.4 2.1-2.1"/>
    <path d="m16.3 7.7 2.1-2.1"/>
  `,
  command: `
    <path d="M18 3a3 3 0 0 0-3 3v12a3 3 0 0 0 3 3 3 3 0 0 0 3-3 3 3 0 0 0-3-3H6a3 3 0 0 0-3 3 3 3 0 0 0 3 3 3 3 0 0 0 3-3V6a3 3 0 0 0-3-3 3 3 0 0 0-3 3 3 3 0 0 0 3 3h12a3 3 0 0 0 3-3 3 3 0 0 0-3-3Z"/>
  `,
  eye: `
    <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/>
    <circle cx="12" cy="12" r="3"/>
  `,
  'eye-off': `
    <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24"/>
    <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"/>
    <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"/>
    <path d="m2 2 20 20"/>
  `,
  lock: `
    <rect x="4" y="11" width="16" height="10" rx="2"/>
    <path d="M8 11V7a4 4 0 0 1 8 0v4"/>
  `,
  power: `
    <path d="M12 2v8"/>
    <path d="m4.93 10.93 1.41 1.41"/>
    <path d="M2 18h2"/>
    <path d="M20 18h2"/>
    <path d="m19.07 10.93-1.41 1.41"/>
    <path d="M22 22H2"/>
    <path d="M8 22v-4a4 4 0 0 1 8 0v4"/>
  `,
  bell: `
    <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9"/>
    <path d="M10.3 21a1.94 1.94 0 0 0 3.4 0"/>
  `,
  star: `
    <path d="m12 2 3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2Z"/>
  `,
  heart: `
    <path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z"/>
  `,
  trash: `
    <path d="M3 6h18"/>
    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/>
    <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
    <line x1="10" y1="11" x2="10" y2="17"/>
    <line x1="14" y1="11" x2="14" y2="17"/>
  `,
  edit: `
    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5Z"/>
  `,
  'external-link': `
    <path d="M15 3h6v6"/>
    <path d="M10 14 21 3"/>
    <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
  `,
  menu: `
    <line x1="3" y1="6" x2="21" y2="6"/>
    <line x1="3" y1="12" x2="21" y2="12"/>
    <line x1="3" y1="18" x2="21" y2="18"/>
  `,
  close: `
    <path d="M18 6 6 18"/>
    <path d="m6 6 12 12"/>
  `,
  'plus-circle': `
    <circle cx="12" cy="12" r="9"/>
    <path d="M12 8v8"/>
    <path d="M8 12h8"/>
  `,
  globe: `
    <circle cx="12" cy="12" r="9"/>
    <path d="M3 12h18"/>
    <path d="M12 3a14.5 14.5 0 0 1 0 18 14.5 14.5 0 0 1 0-18"/>
  `,
  moon: `
    <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79Z"/>
  `,
  sun: `
    <circle cx="12" cy="12" r="4"/>
    <path d="M12 2v2"/>
    <path d="M12 20v2"/>
    <path d="m4.93 4.93 1.41 1.41"/>
    <path d="m17.66 17.66 1.41 1.41"/>
    <path d="M2 12h2"/>
    <path d="M20 12h2"/>
    <path d="m6.34 17.66-1.41 1.41"/>
    <path d="m19.07 4.93-1.41 1.41"/>
  `,
} as const;

// Type-safe access — name → path
export type IconPathName = keyof typeof ICON_PATHS;