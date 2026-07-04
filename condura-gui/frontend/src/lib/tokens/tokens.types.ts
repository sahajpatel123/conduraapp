/*
 * Synaptic — Token Type Definitions
 *
 * Hand-maintained TypeScript types for every token. Use as:
 *   import type { SemanticToken, MotionToken } from '$tokens/tokens.types';
 *
 * If you add a token to primitives.css / semantic.css / motion.css, mirror
 * it here. CI asserts coverage (see scripts/check-tokens.ts).
 *
 * Locked: see docs/design-v1-redesign.md.
 */

/* ---------------------------------------------------------------------- *
 * SEMANTIC TOKENS — Layer 2 (consumed by components)
 * ---------------------------------------------------------------------- */

export type SurfaceToken =
  | '--surface-base'
  | '--surface-sunken'
  | '--surface-raised'
  | '--surface-overlay'
  | '--surface-scrim'
  | '--surface-inverted'
  | '--surface-glass';

export type ContentToken =
  | '--content-primary'
  | '--content-secondary'
  | '--content-tertiary'
  | '--content-muted'
  | '--content-disabled'
  | '--content-inverse'
  | '--content-link'
  | '--content-link-hover'
  | '--content-on-accent'
  | '--content-accent';

export type BorderToken =
  | '--border-subtle'
  | '--border-default'
  | '--border-strong'
  | '--border-focus'
  | '--border-inverse';

export type ActionToken =
  | '--action-primary-idle-bg'
  | '--action-primary-idle-fg'
  | '--action-primary-hover-bg'
  | '--action-primary-hover-fg'
  | '--action-primary-active-bg'
  | '--action-primary-active-fg'
  | '--action-primary-disabled-bg'
  | '--action-primary-disabled-fg'
  | '--action-secondary-idle-bg'
  | '--action-secondary-idle-fg'
  | '--action-secondary-idle-border'
  | '--action-secondary-hover-bg'
  | '--action-secondary-hover-fg'
  | '--action-secondary-hover-border'
  | '--action-secondary-active-bg'
  | '--action-secondary-active-fg'
  | '--action-secondary-disabled-bg'
  | '--action-secondary-disabled-fg'
  | '--action-secondary-disabled-border'
  | '--action-tertiary-idle-fg'
  | '--action-tertiary-hover-fg'
  | '--action-tertiary-hover-bg'
  | '--action-tertiary-active-fg'
  | '--action-tertiary-active-bg'
  | '--action-tertiary-disabled-fg'
  | '--action-destructive-idle-bg'
  | '--action-destructive-idle-fg'
  | '--action-destructive-idle-border'
  | '--action-destructive-hover-bg'
  | '--action-destructive-hover-fg'
  | '--action-destructive-hover-border'
  | '--action-destructive-active-bg'
  | '--action-destructive-active-fg';

export type StatusToken =
  | `--status-${'success' | 'warning' | 'error' | 'info' | 'neutral'}-${'bg' | 'fg' | 'border'}`;

export type SemanticToken =
  | SurfaceToken
  | ContentToken
  | BorderToken
  | ActionToken
  | StatusToken
  | '--focus-ring'
  | '--focus-ring-offset'
  | '--selection-bg'
  | '--selection-fg';

/* ---------------------------------------------------------------------- *
 * MOTION TOKENS
 * ---------------------------------------------------------------------- */

export type DurationToken =
  | '--duration-instant'
  | '--duration-fast'
  | '--duration-base'
  | '--duration-slow'
  | '--duration-emphasized'
  | '--duration-epic';

export type EasingToken =
  | '--ease-standard'
  | '--ease-decelerate'
  | '--ease-accelerate'
  | '--ease-emphasized';

export type DistanceToken =
  | '--distance-micro'
  | '--distance-base'
  | '--distance-near'
  | '--distance-far';

export type StaggerToken =
  | '--stagger-fast'
  | '--stagger-base'
  | '--stagger-slow'
  | '--stagger-deliberate';

export type PulseToken =
  | '--pulse-period-idle'
  | '--pulse-period-thinking'
  | '--pulse-period-awaiting';

export type MotionToken =
  | DurationToken
  | EasingToken
  | DistanceToken
  | StaggerToken
  | PulseToken
  | `--transition-${'hover' | 'press' | 'panel' | 'overlay-in' | 'overlay-out' | 'consent' | 'killswitch'}`;

/* ---------------------------------------------------------------------- *
 * SPACING + RADIUS + Z-INDEX
 * ---------------------------------------------------------------------- */

export type SpaceToken = `--space-${0 | '0-5' | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13}`;

export type RadiusToken =
  | '--radius-xs'
  | '--radius-sm'
  | '--radius-md'
  | '--radius-lg'
  | '--radius-xl'
  | '--radius-2xl'
  | '--radius-pill';

export type ZIndexToken =
  | '--z-base'
  | '--z-raised'
  | '--z-sticky'
  | '--z-overlay'
  | '--z-modal'
  | '--z-toast'
  | '--z-tooltip'
  | '--z-max';

export type AllTokens = SemanticToken | MotionToken | SpaceToken | RadiusToken | ZIndexToken;