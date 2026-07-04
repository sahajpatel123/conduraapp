import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent } from '@testing-library/svelte';
import KillSwitchOverlay from './KillSwitchOverlay.svelte';

// KillSwitchOverlay.test.ts — first Svelte test for the kill-switch
// UI surface (tied to the KillSwitchConductor wired in
// internal/conductor/killswitch.go).
//
// Why this test exists:
//  - The KillSwitchConductor (Go) writes the halt reason into the
//    halt.Flag. The condura/ KillSwitchOverlay.svelte reads that
//    reason via a prop and renders it to the user. If the prop
//    contract drifts, the user sees "Reason: undefined" instead
//    of "Reason: hard_hotkey".
//  - This test pins the prop contract so any future refactor that
//    breaks the UI render is caught.
//
// The test is intentionally minimal — we don't test the full visual
// styling (that's an a11y/design review) — only the contract that
// the audit log + UI surface agree on the reason text.
//
// Note on selectors: the rendered DOM contains the word "halted" 3
// times (eyebrow, body, note), so we use scoped queries to target
// the specific element we care about.

describe('KillSwitchOverlay', () => {
  it('renders the default reason when no reason prop is provided', () => {
    const { container } = render(KillSwitchOverlay);
    // The component defaults to reason="user requested" per the
    // <script> block. If this changes, the test surfaces the drift.
    const reasonEl = container.querySelector('.kill-reason');
    expect(reasonEl?.textContent?.trim()).toBe('user requested');
  });

  it('renders the hard_hotkey reason when the prop is set', () => {
    const { container } = render(KillSwitchOverlay, {
      props: { reason: 'hard_hotkey' },
    });
    const reasonEl = container.querySelector('.kill-reason');
    expect(reasonEl?.textContent?.trim()).toBe('hard_hotkey');
  });

  it('renders the Halted eyebrow text', () => {
    const { container } = render(KillSwitchOverlay);
    // Scoped to .kill-eyebrow because the word "halted" appears in
    // three places (eyebrow, body, note). The eyebrow text is the
    // status banner.
    const eyebrow = container.querySelector('.kill-eyebrow');
    expect(eyebrow?.textContent).toMatch(/halted/i);
    expect(eyebrow?.textContent).toMatch(/kill switch engaged/i);
  });

  it('renders the resume button', () => {
    const { container } = render(KillSwitchOverlay);
    const button = container.querySelector('button');
    expect(button).toBeInTheDocument();
    expect(button?.textContent?.trim()).toBe('Mint resume ticket');
  });

  it('fires onresume when the resume button is clicked', async () => {
    const onresume = vi.fn();
    const { container } = render(KillSwitchOverlay, {
      props: { reason: 'hard_hotkey', onresume },
    });
    const button = container.querySelector('button') as HTMLButtonElement;
    expect(button).toBeInTheDocument();
    await fireEvent.click(button);
    expect(onresume).toHaveBeenCalledTimes(1);
  });
});