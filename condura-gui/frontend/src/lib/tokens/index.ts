/*
 * Synaptic — Tokens Public API
 *
 * Single import point for all token modules. Components should import
 * from here, not from individual files.
 *
 *     import { SPRINGS, type SpringName, token } from '$tokens';
 *
 * Locked: see docs/design-v1-redesign.md.
 */

export * from './motion';
export * from './themes';

// Re-export the typed helper for ergonomic token access from JS.
// Components typically consume tokens via var(--token-name) in CSS,
// but sometimes we need them in JS (e.g., dynamic canvas drawing).

/** Type-safe token accessor. Pass a token name (e.g., 'surface.raised')
 *  and get back the resolved CSS value from the computed style of :root.
 *  In SSR contexts, returns the raw var() string. */
export function token(name: string): string {
  if (typeof document === 'undefined') return `var(--${name.replace(/\./g, '-')})`;
  return getComputedStyle(document.documentElement).getPropertyValue(`--${name.replace(/\./g, '-')}`).trim();
}