import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import Pulse from './Pulse.svelte';

// Pulse.test.ts — vitest smoke test (audit SB-09).
//
// This is the first frontend test in the repo. It exists to:
//  1. Verify the vitest config + svelte plugin + jsdom are wired.
//  2. Provide a working example for future Svelte test authors.
//  3. Catch obvious regressions in Pulse (the breathing animation
//     that gates every loading state in the production shell).
//
// Pulse is a small, presentational component that takes props for
// `phase`, `size`, and an optional `class`. The component renders
// a <span class="pulse-dot"> with an inline `animation` style.
//
// NOTE: earlier draft asserted a `label` prop + `.pulse` class, but
// the real Pulse API doesn't have either. This test pins the actual
// API contract so any future drift is caught here.

describe('Pulse', () => {
  it('renders a span.pulse-dot for every valid phase', () => {
    const phases = ['idle', 'thinking', 'awaiting', 'acting', 'consent', 'error', 'ok'] as const;
    for (const phase of phases) {
      const { container } = render(Pulse, { props: { phase } });
      const dot = container.querySelector('.pulse-dot');
      expect(dot, `phase=${phase}`).toBeInTheDocument();
      // Inline animation style is the contract — each phase
      // maps to a different animation duration in MAP.
      const style = (dot as HTMLElement).getAttribute('style') ?? '';
      expect(style, `phase=${phase} style`).toMatch(/animation:\s*breathe/);
    }
  });

  it('honors the size prop in the inline style', () => {
    const { container } = render(Pulse, {
      props: { phase: 'working' as any, size: 24 }, // 'working' isn't a real phase — should fall back to idle
    });
    const dot = container.querySelector('.pulse-dot') as HTMLElement;
    expect(dot).toBeInTheDocument();
    const style = dot.getAttribute('style') ?? '';
    expect(style).toContain('width: 24px');
    expect(style).toContain('height: 24px');
  });

  it('falls back to idle phase when given an unknown phase', () => {
    const { container } = render(Pulse, {
      props: { phase: 'unknown-phase' as any },
    });
    const dot = container.querySelector('.pulse-dot') as HTMLElement;
    expect(dot).toBeInTheDocument();
    // Idle phase uses synapse-glow per the MAP constant.
    const style = dot.getAttribute('style') ?? '';
    expect(style).toContain('var(--synapse-glow)');
    expect(style).toMatch(/animation:\s*breathe 5s/);
  });

  it('renders the error phase with danger color and 0.8s duration', () => {
    // This is the regression check for the wired KillSwitchConductor
    // (DRIFT-009). When the kill switch fires, the overlay shows a
    // red, fast-breathing Pulse — this test pins that contract.
    const { container } = render(Pulse, { props: { phase: 'error' } });
    const dot = container.querySelector('.pulse-dot') as HTMLElement;
    expect(dot).toBeInTheDocument();
    const style = dot.getAttribute('style') ?? '';
    expect(style).toContain('var(--danger)');
    expect(style).toMatch(/animation:\s*breathe 0\.8s/);
  });
});