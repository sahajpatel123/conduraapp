/**
 * ui — primitive component barrel.
 *
 * Every primitive is a pure presentational component (no IPC, no
 * stores). Routes and shared components import from here so we have
 * one place to grep when an audit asks "is there a Button? what does
 * it accept?".
 *
 * Phase A: the primitives exist; the app shell still uses the old
 * ad-hoc styles in components/* and routes/*. Phase B rewrites those
 * consumers. Until then these primitives are unused but available
 * via the dev smoke page at #/dev/components.
 */

export { default as Avatar } from './Avatar.svelte'
export { default as Badge } from './Badge.svelte'
export { default as Button } from './Button.svelte'
export { default as Card } from './Card.svelte'
export { default as CommandPalette } from './CommandPalette.svelte'
export { default as Dialog } from './Dialog.svelte'
export { default as Divider } from './Divider.svelte'
export { default as EmptyState } from './EmptyState.svelte'
export { default as IconButton } from './IconButton.svelte'
export { default as Input } from './Input.svelte'
export { default as Kbd } from './Kbd.svelte'
export { default as Progress } from './Progress.svelte'
export { default as SegmentedControl } from './SegmentedControl.svelte'
export { default as Select } from './Select.svelte'
export { default as Sheet } from './Sheet.svelte'
export { default as Skeleton } from './Skeleton.svelte'
export { default as Slider } from './Slider.svelte'
export { default as Switch } from './Switch.svelte'
export { default as Tabs } from './Tabs.svelte'
export { default as Textarea } from './Textarea.svelte'
export { default as Toast } from './Toast.svelte'
export { default as Tooltip } from './Tooltip.svelte'