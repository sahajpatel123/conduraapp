import {
  Apple,
  BadgeCheck,
  Bell,
  BookOpen,
  BrainCircuit,
  CheckCircle2,
  CircleDashed,
  Clock3,
  Code2,
  Computer,
  FileClock,
  Fingerprint,
  Gauge,
  Github,
  KeyRound,
  Laptop,
  LayoutDashboard,
  LifeBuoy,
  LockKeyhole,
  Mic2,
  Monitor,
  Network,
  PauseCircle,
  Radio,
  ShieldCheck,
  SquareTerminal,
  TriangleAlert,
  Workflow,
  type LucideIcon,
} from "lucide-react";

export type Platform = {
  name: string;
  label: string;
  status: string;
  detail: string;
  version: string;
  checksum: string;
  signing: string;
  icon: LucideIcon;
};

export const platforms: Platform[] = [
  {
    name: "macOS",
    label: "Download for macOS",
    status: "Build pending",
    detail: "Installer will appear after desktop verification and notarization.",
    version: "v0.1.0 preview",
    checksum: "SHA256 pending",
    signing: "Notarization pending",
    icon: Apple,
  },
  {
    name: "Windows",
    label: "Windows coming soon",
    status: "Packaging pending",
    detail: "Signed installer will follow the first stable desktop build.",
    version: "v0.1.0 preview",
    checksum: "SHA256 pending",
    signing: "Code signing pending",
    icon: Monitor,
  },
  {
    name: "Linux",
    label: "Linux coming soon",
    status: "Packaging pending",
    detail: "AppImage and deb packages are planned for the public release.",
    version: "v0.1.0 preview",
    checksum: "SHA256 pending",
    signing: "Package signing pending",
    icon: Computer,
  },
];

export const trustPillars = [
  {
    title: "Local-first by default",
    body: "Memory, approvals, audit records, and keys are designed to stay on the user's machine unless they choose otherwise.",
    icon: Laptop,
  },
  {
    title: "Gatekeeper before action",
    body: "The model can propose. Deterministic policy decides whether a physical action may proceed.",
    icon: ShieldCheck,
  },
  {
    title: "User-owned models",
    body: "Use local Ollama or configured providers. Synaptic is the conductor, not a lock-in cloud.",
    icon: BrainCircuit,
  },
  {
    title: "Auditable activity",
    body: "Every sensitive decision is designed to leave a readable trail for the person at the keyboard.",
    icon: FileClock,
  },
];

export const agentLoop = [
  {
    title: "Ask",
    body: "Summon Synaptic from the desktop and type or speak the task.",
    icon: Mic2,
  },
  {
    title: "Plan",
    body: "The agent turns the request into visible steps before touching anything.",
    icon: Workflow,
  },
  {
    title: "Approve",
    body: "Gatekeeper pauses network, write, and destructive actions for permission.",
    icon: TriangleAlert,
  },
  {
    title: "Act",
    body: "Approved steps run through the safest available computer-use route.",
    icon: SquareTerminal,
  },
  {
    title: "Audit",
    body: "A durable timeline records what happened and why.",
    icon: FileClock,
  },
];

export const safetyMechanisms = [
  {
    title: "Blast-radius classification",
    body: "Read, write, network, and destructive actions are classified before execution.",
    icon: BadgeCheck,
  },
  {
    title: "Twin-snapshot verification",
    body: "The app verifies the screen state before acting so stale UI cannot silently misfire.",
    icon: Gauge,
  },
  {
    title: "Human approval boundary",
    body: "Sensitive and destructive operations stop at a native approval moment.",
    icon: Fingerprint,
  },
  {
    title: "Hard pause paths",
    body: "The user can pause or stop agent activity through independent controls.",
    icon: PauseCircle,
  },
  {
    title: "Model isolation",
    body: "Outputs are validated before they become instructions for another tool or model.",
    icon: Network,
  },
  {
    title: "Audit trail",
    body: "Security-relevant actions are designed to be explainable after the fact.",
    icon: FileClock,
  },
];

export const docsCards = [
  {
    title: "Install Synaptic",
    body: "Platform guides will land with signed release artifacts.",
    icon: Laptop,
  },
  {
    title: "Connect models",
    body: "Use Ollama locally or connect a provider with your own key.",
    icon: KeyRound,
  },
  {
    title: "Grant permissions",
    body: "Understand Accessibility, microphone, screen recording, and what each unlocks.",
    icon: LockKeyhole,
  },
  {
    title: "Use the dashboard",
    body: "Optional browser sign-in will manage devices, support, and Skills Hub later.",
    icon: LayoutDashboard,
  },
];

export const dashboardFeatures = [
  {
    title: "Device management",
    body: "Pair devices and see health state when sync features are ready.",
    icon: Laptop,
  },
  {
    title: "Skills Hub",
    body: "Discover and manage safety-reviewed workflows later.",
    icon: Code2,
  },
  {
    title: "Release notifications",
    body: "Get updates without making download access account-gated.",
    icon: Bell,
  },
  {
    title: "Support history",
    body: "Keep support conversations linked to your browser account.",
    icon: LifeBuoy,
  },
];

export const footerGroups = [
  {
    title: "Product",
    links: [
      { label: "Download", href: "/download" },
      { label: "Safety", href: "/safety" },
      { label: "Docs", href: "/docs" },
      { label: "Changelog", href: "/changelog" },
    ],
  },
  {
    title: "Trust",
    links: [
      { label: "Privacy", href: "/privacy" },
      { label: "Terms", href: "/terms" },
      { label: "Security contact", href: "/safety#security-contact" },
      { label: "GitHub", href: "https://github.com/sahajpatel123/synapticapp", icon: Github },
    ],
  },
  {
    title: "Status",
    links: [
      { label: "Builds pending", href: "/download" },
      { label: "Release notes pending", href: "/changelog" },
      { label: "Dashboard placeholder", href: "/dashboard" },
      { label: "Docs shell", href: "/docs" },
    ],
  },
];

export const releaseStates = [
  { label: "Desktop verification", status: "In progress", icon: CircleDashed },
  { label: "Signed installers", status: "Pending", icon: Clock3 },
  { label: "Checksums", status: "Pending", icon: Clock3 },
  { label: "Public release notes", status: "Pending", icon: BookOpen },
  { label: "Download access", status: "Public, no login", icon: CheckCircle2 },
];

export const commandTrace = [
  { label: "Hotkey", value: "Overlay summoned", tone: "cyan" },
  { label: "Voice", value: "Local transcript ready", tone: "green" },
  { label: "Plan", value: "3 steps proposed", tone: "cyan" },
  { label: "Gatekeeper", value: "Approval required", tone: "amber" },
  { label: "Audit", value: "Event prepared", tone: "green" },
];

export const legalUpdated = "June 9, 2026";

export const supportEmail = "support@synaptic.app";
export const securityEmail = "security@synaptic.app";
export const privacyEmail = "privacy@synaptic.app";
export const appName = "Synaptic";

export const siteCopy = {
  headline: "The AI conductor for your computer.",
  lead: "Synaptic is a free, local-first desktop agent that appears from your OS, listens when summoned, routes work through your own models, and pauses at deterministic safety boundaries before touching anything important.",
};

export const accountPrinciples = [
  "No account required to download.",
  "No account required for local-first desktop use.",
  "Browser sign-in is only for dashboard, devices, Skills Hub, support, and release notifications.",
];

export const permissionRows = [
  ["Accessibility", "Read UI structure and operate approved controls", "Asked during onboarding"],
  ["Microphone", "Push-to-talk voice input", "Optional"],
  ["Screen recording", "Verify visual state when needed", "Purpose-limited"],
  ["Network", "Call configured model providers and update services", "User configured"],
];

export const emptyChangelog = {
  title: "Release notes will appear with the first public build.",
  body: "The download UI is ready as a trust surface, but installers are intentionally not wired until the desktop release is verified, signed, and checksummed.",
};
