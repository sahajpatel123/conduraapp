/*
 * Synaptic v1 Design System — public exports
 *
 * All 35 components of the v1 design system (matches spec §12 exactly).
 * Locked: docs/design-v1-redesign.md.
 */

// ── Tier 1 — Atomic (no dependencies) ─────────────────────────
export { default as Hairline } from './Hairline.svelte';
export { default as Pulse }    from './Pulse.svelte';
export { default as Dot }      from './Dot.svelte';
export { default as Stack }    from './Stack.svelte';
export { default as Inline }   from './Inline.svelte';
export { default as Icon }     from './icons/Icon.svelte';
export { default as Spacer }   from './Spacer.svelte';

// Icon library
export { ICON_PATHS, type IconPathName } from './icons/paths';
export type { IconName } from './icons/Icon.svelte';

// ── Tier 2 — Inputs & controls ─────────────────────────────────
export { default as Button }         from './Button.svelte';
export { default as IconButton }     from './IconButton.svelte';
export { default as Input }          from './Input.svelte';
export { default as Textarea }       from './Textarea.svelte';
export { default as Chip }           from './Chip.svelte';
export { default as Pill }           from './Pill.svelte';
export { default as Switch }         from './Switch.svelte';
export { default as Slider }         from './Slider.svelte';
export { default as KeyCombo }       from './KeyCombo.svelte';
export { default as HotkeyRecorder } from './HotkeyRecorder.svelte';

// ── Tier 3 — Display ──────────────────────────────────────────
export { default as Surface }        from './Surface.svelte';
export { default as Card }           from './Card.svelte';
export { default as Receipt }        from './Receipt.svelte';
export { default as ProgressBar }    from './ProgressBar.svelte';
export { default as EmptyState }     from './EmptyState.svelte';
export { default as LoadingState }   from './LoadingState.svelte';
export { default as Suggestion }     from './Suggestion.svelte';
export { default as ContextChip }    from './ContextChip.svelte';
export { default as Avatar }         from './Avatar.svelte';
export { default as StreamingText }  from './StreamingText.svelte';

// ── Tier 4 — Composite surfaces ────────────────────────────────
export { default as CommandSurface }     from './CommandSurface.svelte';
export { default as ChatSurface }        from './ChatSurface.svelte';
export { default as ConversationDrawer } from './ConversationDrawer.svelte';
export { default as SettingsPane }       from './SettingsPane.svelte';
export { default as Sidebar }            from './Sidebar.svelte';
export { default as StatusBar }          from './StatusBar.svelte';
export { default as ConsentModal }       from './ConsentModal.svelte';
export { default as KillSwitchOverlay }  from './KillSwitchOverlay.svelte';
export { default as AgentActionLog }     from './AgentActionLog.svelte';
export { default as BrandWordmark }      from './BrandWordmark.svelte';
export { default as AmbientBackground }  from './AmbientBackground.svelte';
export { default as Toast }              from './Toast.svelte';
export { default as Toaster }            from './Toaster.svelte';
export { default as Tip }                from './Tip.svelte';
export { default as Tooltip }            from './Tooltip.svelte';
export { default as SectionHeader }      from './SectionHeader.svelte';
export { default as Sheet }              from './Sheet.svelte';
export { default as CanvasBackground }   from './CanvasBackground.svelte';
export { default as CursorHalo }         from './CursorHalo.svelte';
export { default as Magnetic }           from './Magnetic.svelte';
export { default as SynapseField }       from './SynapseField.svelte';

// ── Tier 4 — Onboarding wizard ─────────────────────────────────
export { default as OnboardingWizard }    from './onboarding/OnboardingWizard.svelte';
export { default as Invitation }          from './onboarding/Invitation.svelte';
export { default as Eula }                from './onboarding/Eula.svelte';
export { default as Eyes }                from './onboarding/Eyes.svelte';
export { default as PowerSource }         from './onboarding/PowerSource.svelte';
export { default as Hotkey }              from './onboarding/Hotkey.svelte';
export { default as FirstBreath }         from './onboarding/FirstBreath.svelte';