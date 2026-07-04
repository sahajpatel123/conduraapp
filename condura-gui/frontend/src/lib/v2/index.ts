// ============================================================
// Condura v2 — Design System Root
//
// Every v2 component is exported here. Consumers should import
// from `$lib/v2` — never from individual files — so the design
// system can evolve without touching consumer code.
//
// Components:
//   Surface     — paper card / panel / sheet
//   Ink         — text (display / title / body / ui / mono)
//   Stack       — vertical spacing
//   Inline      — horizontal flow
//   Rule        — hairline divider
//   Button      — primary / ghost / deny
//
// CSS imports:
//   tokens.css  — palette, typography, spacing, radii, shadows, z
//   motion.css  — easings, durations, keyframes
//   reset.css   — scoped element resets
//
// Usage in a route:
//
//   <script>
//     import '$lib/v2/tokens.css'
//     import '$lib/v2/motion.css'
//     import '$lib/v2/reset.css'
//     import { Surface, Ink, Stack } from '$lib/v2'
//   </script>
//
//   <div data-v2>
//     <Surface elevation={1} padding="6" radius="2">
//       <Stack gap={3}>
//         <Ink kind="display">Welcome to Condura.</Ink>
//         <Ink kind="body" tone="ink-2">
//           A quiet companion that gets out of the way.
//         </Ink>
//       </Stack>
//     </Surface>
//   </div>
// ============================================================

export { default as Surface } from './Surface.svelte'
export { default as Ink } from './Ink.svelte'
export { default as Stack } from './Stack.svelte'
export { default as Inline } from './Inline.svelte'
export { default as Rule } from './Rule.svelte'
export { default as Button } from './Button.svelte'
export { default as FloatingInterview, type InterviewAnswers } from './FloatingInterview.svelte'
export { default as ChatSurface, type Turn } from './ChatSurface.svelte'
export { default as Sidebar, type SidebarItem } from './Sidebar.svelte'
export { default as StatusBar } from './StatusBar.svelte'
export { default as ConsentModal } from './ConsentModal.svelte'
export { default as Glyph } from './Glyph.svelte'
export { default as Avatar } from './Avatar.svelte'
export { default as Chip } from './Chip.svelte'
export { default as Eyebrow } from './Eyebrow.svelte'
export { default as Switch } from './Switch.svelte'
export { default as SettingsDocument, type Chapter, type Row } from './SettingsDocument.svelte'
export { default as Hub, type Skill } from './Hub.svelte'
export { default as Audit, type AuditEntry } from './Audit.svelte'
export { default as Sync } from './Sync.svelte'
export { default as Replay, type ReplayFrame } from './Replay.svelte'
export { default as Channels, type Channel } from './Channels.svelte'
export { default as Delegation, type Agent } from './Delegation.svelte'
export { default as Skills, type LocalSkill } from './Skills.svelte'
export { default as About } from './About.svelte'
export { default as V2Shell } from './V2Shell.svelte'
