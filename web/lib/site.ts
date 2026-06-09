/*
  Site-wide constants. Content that appears in more than one place lives here,
  so the pages stay prose and the data stays data.
*/

export const SITE = {
  name: "Synaptic",
  tagline: "Every AI on your machine. One conductor.",
  description:
    "Synaptic is a free, OS-native AI agent that conducts every other AI tool on your computer — summoned by hotkey, governed by a deterministic Gatekeeper, local-first and telemetry-free.",
  url: "https://synaptic.app",
} as const;

export const NAV_LINKS = [
  { href: "/manifesto", label: "Manifesto" },
  { href: "/changelog", label: "Changelog" },
  { href: "/download", label: "Download" },
] as const;

/** The tools Synaptic conducts — the orchestra roster. */
export const ORCHESTRA = [
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
    body: "Synaptic ships with no access at all. It asks; you grant. Onboarding makes each grant clear and reversible.",
  },
] as const;
