/*
  Site-wide constants. Content that appears in more than one place lives here,
  so the pages stay prose and the data stays data.
*/

export const SITE = {
  name: "Condura",
  tagline: "A permissioned intelligence layer for your OS.",
  description:
    "Condura is a free, local-first desktop AI agent that appears from the operating system, routes work through your own models, and stops at deterministic safety boundaries before acting.",
  url: "https://condura.app",
  github: "https://github.com/sahajpatel123/conduraapp",
  discord: "https://discord.gg/synaptic",
} as const;

// Reference / informational destinations. These live in the footer,
// not in the dock — they're for browsing, not quick access.
export const NAV_LINKS = [
  { href: "/orchestration", label: "How it works" },
  { href: "/security", label: "Security" },
  { href: "/manifesto", label: "Mission" },
  { href: "/changelog", label: "Changelog" },
  { href: "/download", label: "Download" },
  { href: "/legal", label: "Legal" },
] as const;

/** The local/model delegates Condura can route through. */
export const TOOL_ROSTER = [
  "Claude Code",
  "Codex",
  "Antigravity",
  "OpenCode",
  "Kilo",
  "Hermes",
  "Ollama",
  "Gemini",
] as const;

export const PLATFORMS = [
  {
    key: "mac",
    name: "macOS",
    requirement: "macOS 13+, Apple silicon & Intel",
    artifact: "synaptic.dmg",
  },
  {
    key: "windows",
    name: "Windows",
    requirement: "Windows 10+, x64",
    artifact: "synaptic-setup.exe",
  },
  {
    key: "linux",
    name: "Linux",
    requirement: "glibc 2.31+, x64",
    artifact: "synaptic.AppImage",
  },
] as const;

export type PlatformKey = (typeof PLATFORMS)[number]["key"];

/** The seven non-negotiable invariants, verbatim in spirit from the mission. */
export const INVARIANTS = [
  {
    numeral: "I",
    title: "The Strategist and the Gatekeeper are separate systems.",
    body: "The Strategist is a model. The Gatekeeper is deterministic code. They are never the same, never merged, never shortcut.",
  },
  {
    numeral: "II",
    title: "The Gatekeeper is the only path to physical action.",
    body: "No model output flows to a click, a keystroke, or a shell command without passing through it. There is no side door.",
  },
  {
    numeral: "III",
    title: "Destructive actions require a real human at the keyboard.",
    body: "A native modal dialog that blocks until clicked. No exceptions. No “trust me, the model said it’s safe.”",
  },
  {
    numeral: "IV",
    title: "You can always stop the agent.",
    body: "Hard hotkey, watchdog timer, network isolation, menu-bar kill. Four independent mechanisms. The agent can disable none of them.",
  },
  {
    numeral: "V",
    title: "Every action is auditable, in a tamper-resistant log.",
    body: "HMAC-chained, append-only, never deleted. If something goes wrong, the record proves exactly what happened.",
  },
  {
    numeral: "VI",
    title: "The agent is a guest, not an owner.",
    body: "It asks permission to enter rooms — apps, files, URLs. You grant or deny. It never escalates, never bypasses, never pretends.",
  },
  {
    numeral: "VII",
    title: "OS permissions are granted by you, on your machine.",
    body: "Condura ships with no access at all. It asks; you grant. Onboarding makes each grant clear and reversible.",
  },
] as const;
