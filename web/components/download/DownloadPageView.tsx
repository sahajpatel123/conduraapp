"use client";

import { AnimatePresence, motion, useReducedMotion } from "motion/react";
import { useCallback, useState } from "react";
import PageHeader from "@/components/shell/PageHeader";
import Reveal from "@/components/motion/Reveal";
import { useToast } from "@/context/ToastContext";
import { useIsland } from "@/context/IslandContext";
import { usePlatform } from "@/hooks/usePlatform";
import { DOWNLOADS, RELEASE_TAG } from "@/lib/downloads";
import { PLATFORMS, type PlatformKey } from "@/lib/site";
import { EASE_OUT } from "@/lib/motion";

const INSTALL_STEPS: Record<PlatformKey, { title: string; desc: string }[]> = {
  mac: [
    { title: "Open the disk image", desc: "Open the .dmg, then move Condura into Applications." },
    { title: "Approve system access", desc: "Grant Accessibility and Screen Recording from Privacy & Security." },
    { title: "Choose your hotkey", desc: "Record a shortcut that summons Condura from anywhere." },
    { title: "Run your first task", desc: "Press the hotkey, choose a model, and start orchestrating." },
  ],
  windows: [
    { title: "Run the installer", desc: "Open the setup file and follow the signed installer prompts." },
    { title: "Choose your hotkey", desc: "Record a shortcut that summons Condura from anywhere." },
    { title: "Approve permissions", desc: "Review each requested capability. Every grant is reversible." },
    { title: "Run your first task", desc: "Press the hotkey, choose a model, and start orchestrating." },
  ],
  linux: [
    { title: "Install the package", desc: "Install the .deb package, or choose the CLI tarball below." },
    { title: "Start the daemon", desc: "Confirm the user service is running with condura status." },
    { title: "Choose your hotkey", desc: "Record a shortcut during onboarding for fast access." },
    { title: "Open Condura", desc: "Launch the overlay or run condura-tui from your terminal." },
  ],
};

const VERIFY_COMMANDS: Record<PlatformKey, string> = {
  mac: `shasum -a 256 condura-installer-mac.dmg`,
  windows: `Get-FileHash condura-setup.exe -Algorithm SHA256`,
  linux: `sha256sum condura.deb`,
};

const VERSIONS = [
  { version: "v0.1.1", badge: "Latest", date: "June 24, 2026", desc: "Audit-driven fix bundle: autonomy + perception wiring, encrypted secrets, wake word rename, shell sanitize, audit prune." },
  { version: "v0.1.0", badge: "Stable", date: "June 19, 2026", desc: "First signed release. Local orchestration, deterministic gatekeeper, 8 sub-agents, 12 LLM providers." },
  { version: "v0.0.9", badge: "Preview", date: "May 28, 2026", desc: "Initial beta rollout. Improved model context parsing." },
];

const FAQ_ITEMS = [
  { id: "free", title: "Is Condura really free?", body: "Yes. Condura is free for personal and commercial use under the Condura Freeware EULA. There is no account requirement, no premium tier, and no feature gate. A donate button in the menu bar is the whole business model." },
  { id: "keys", title: "Do I need an API key?", body: "Not for local models. Condura auto-detects a local Ollama installation and fills a sentinel key for you. Add provider credentials only when you choose to use a cloud model." },
  { id: "privacy", title: "Where does my data live?", body: "On your machine. Memory, skills, audit log, and API keys are stored locally, encrypted at rest. Requests leave your device only when the model or service you configure requires a network call." },
  { id: "safety", title: "How are computer actions controlled?", body: "A deterministic Gatekeeper — not a model — evaluates every action before execution. Sensitive or destructive actions require a native dialog that blocks until you click Allow. You can always stop the agent with one hotkey." },
  { id: "uninstall", title: "Can I remove it cleanly?", body: "Yes. Uninstall auto-backs-up your data to ~/Documents/condura-backups/ before removing. No cloud account to cancel, no data on someone else's server." },
  { id: "update", title: "How do updates work?", body: "Condura checks published releases and supports signed update packages (Ed25519) with rollback protection. Release details and checksums are on GitHub." },
];

export default function DownloadPageView() {
  const detected = usePlatform();
  const [selected, setSelected] = useState<PlatformKey>(detected);
  const [downloading, setDownloading] = useState(false);
  const { push } = useToast();
  const { pulseDownload } = useIsland();

  const platform = PLATFORMS.find((item) => item.key === selected)!;

  const triggerDownload = useCallback((href: string, label: string) => {
    const anchor = document.createElement("a");
    anchor.href = href;
    anchor.setAttribute("download", "");
    anchor.rel = "noopener";
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    setDownloading(true);
    pulseDownload(platform.name);
    push({ title: "Download started", description: `${label} for ${platform.name}.`, tone: "success" });
    window.setTimeout(() => setDownloading(false), 1800);
  }, [platform.name, pulseDownload, push]);

  return (
    <div className="relative w-full">
      <PageHeader
        eyebrow="Download"
        title="Condura,"
        titleAccent="on your machine."
        description="A local-first intelligence layer for your OS. Free forever. Signed and notarized. Current and previous builds for every supported architecture."
      >
        <Hero
          selected={selected}
          downloading={downloading}
          onSelect={setSelected}
          onDownload={triggerDownload}
        />
        <BuildChooser selected={selected} onSelect={setSelected} onDownload={triggerDownload} detected={detected} />
        <SetupSection selected={selected} />
        <FAQSection />
      </PageHeader>
    </div>
  );
}

/* ── Hero: platform selector + version table ── */
function Hero({
  selected,
  downloading,
  onSelect,
  onDownload,
}: {
  selected: PlatformKey;
  downloading: boolean;
  onSelect: (key: PlatformKey) => void;
  onDownload: (href: string, label: string) => void;
}) {
  const reduceMotion = useReducedMotion();
  const platform = PLATFORMS.find((item) => item.key === selected)!;

  return (
    <section>
      {/* platform selector */}
      <Reveal>
        <div className="flex justify-center" aria-label="Choose operating system">
          <div className="inline-flex rounded-full border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper-warm)] p-1.5 shadow-[var(--shadow-paper)]">
            {PLATFORMS.map((item) => {
              const active = item.key === selected;
              return (
                <button
                  key={item.key}
                  type="button"
                  onClick={() => onSelect(item.key)}
                  aria-pressed={active}
                  className={`relative flex min-h-10 items-center gap-2.5 rounded-full px-5 text-[13.5px] font-medium transition-colors ${
                    active ? "text-[var(--color-paper)]" : "text-[var(--color-ink-mute)] hover:text-[var(--color-ink)]"
                  }`}
                >
                  {active && (
                    <motion.span
                      layoutId="dl-platform-pill"
                      className="absolute inset-0 rounded-full bg-[var(--color-ink)]"
                      transition={{ type: "spring", stiffness: 420, damping: 34 }}
                    />
                  )}
                  <PlatformIcon platform={item.key} className="relative" />
                  <span className="relative">{item.name}</span>
                </button>
              );
            })}
          </div>
        </div>
      </Reveal>

      {/* version table */}
      <Reveal delay={0.1}>
        <div className="mt-10 w-full overflow-hidden rounded-2xl border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper-warm)] text-left shadow-[var(--shadow-card)]">
          <div className="hidden grid-cols-[1.1fr_1.1fr_2.2fr_auto] gap-4 border-b border-[rgba(20,17,11,0.10)] bg-[var(--color-paper-deep)] px-6 py-4 font-mono text-[10px] uppercase tracking-[0.15em] text-[var(--color-ink-mute)] sm:grid">
            <div>Version</div>
            <div>Release date</div>
            <div>Highlights</div>
            <div className="w-32 text-right">Download</div>
          </div>
          <div className="divide-y divide-[rgba(20,17,11,0.08)]">
            {VERSIONS.map((v, i) => (
              <div key={v.version} className="grid gap-y-3 gap-x-4 p-5 sm:grid-cols-[1.1fr_1.1fr_2.2fr_auto] sm:items-center sm:px-6 sm:py-5 transition-colors hover:bg-[rgba(20,17,11,0.025)]">
                <div className="flex items-center gap-3">
                  <span className="text-[15px] font-semibold text-[var(--color-ink)]">{v.version}</span>
                  {v.badge && (
                    <span className={`rounded-full px-2 py-0.5 font-mono text-[9.5px] uppercase tracking-wider ${
                      v.badge === "Latest"
                        ? "border border-[rgba(201,123,46,0.4)] bg-[rgba(201,123,46,0.12)] text-[var(--color-pollen-deep)]"
                        : "border border-[rgba(20,17,11,0.15)] text-[var(--color-ink-mute)]"
                    }`}>
                      {v.badge}
                    </span>
                  )}
                </div>
                <div className="text-[14px] font-medium text-[var(--color-ink-mute)]">{v.date}</div>
                <div className="pr-4 text-[14px] leading-relaxed text-[var(--color-ink-soft)]">{v.desc}</div>
                <div className="mt-2 flex sm:mt-0 sm:w-32 sm:justify-end">
                  <button
                    onClick={() => onDownload(DOWNLOADS[selected].primary.href, `${v.version} for ${platform.name}`)}
                    className={`flex w-full items-center justify-center gap-2 rounded-full px-5 py-2.5 text-[13px] font-semibold transition-all sm:w-auto ${
                      i === 0
                        ? "bg-[var(--color-pollen)] text-[var(--color-ink)] hover:-translate-y-0.5 hover:bg-[var(--color-pollen-deep)] hover:text-[var(--color-paper)]"
                        : "border border-[rgba(20,17,11,0.18)] bg-transparent text-[var(--color-ink-soft)] hover:border-[rgba(20,17,11,0.35)] hover:bg-[rgba(20,17,11,0.04)]"
                    }`}
                  >
                    <motion.span
                      animate={downloading && i === 0 && !reduceMotion ? { y: [0, 2, 0] } : { y: 0 }}
                      transition={{ duration: 0.55, repeat: downloading ? Infinity : 0 }}
                    >
                      <DownloadGlyph size={15} />
                    </motion.span>
                    {downloading && i === 0 ? "Starting" : "Get"}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </Reveal>
    </section>
  );
}

/* ── Build chooser ── */
function BuildChooser({
  selected,
  onSelect,
  onDownload,
  detected,
}: {
  selected: PlatformKey;
  onSelect: (key: PlatformKey) => void;
  onDownload: (href: string, label: string) => void;
  detected: PlatformKey;
}) {
  return (
    <section className="mt-28">
      <Reveal>
        <p className="text-eyebrow mb-4">— Available builds</p>
        <h2 className="text-display text-[var(--color-ink)] max-w-[16ch] text-balance">Choose the right package.</h2>
        <p className="text-lead mt-5 max-w-[52ch] text-[var(--color-ink-soft)] text-pretty">
          The recommended installer is first. Smaller runtime and portable builds remain available for advanced setups.
        </p>
      </Reveal>

      <Reveal delay={0.1}>
        <div className="mt-10 divide-y divide-[rgba(20,17,11,0.10)] border-y border-[rgba(20,17,11,0.14)]">
          {PLATFORMS.map((platform) => {
            const active = platform.key === selected;
            const isDetected = platform.key === detected;
            const downloads = DOWNLOADS[platform.key];
            return (
              <div
                key={platform.key}
                className={`relative grid gap-5 py-6 transition-colors sm:grid-cols-[1.15fr_1fr_auto] sm:items-center sm:px-5 ${
                  active ? "bg-[rgba(20,17,11,0.035)]" : "hover:bg-[rgba(20,17,11,0.02)]"
                }`}
              >
                {active && (
                  <motion.span layoutId="active-build-rail" className="absolute inset-y-0 left-0 w-0.5 bg-[var(--color-pollen)]" />
                )}
                <button
                  type="button"
                  onClick={() => onSelect(platform.key)}
                  className="flex min-w-0 items-center gap-4 text-left"
                >
                  <span className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border ${
                    active
                      ? "border-[rgba(201,123,46,0.5)] bg-[rgba(201,123,46,0.12)] text-[var(--color-pollen-deep)]"
                      : "border-[rgba(20,17,11,0.14)] text-[var(--color-ink-mute)]"
                  }`}>
                    <PlatformIcon platform={platform.key} className="" size={20} />
                  </span>
                  <span className="min-w-0">
                    <span className="flex flex-wrap items-center gap-2">
                      <span className="text-[16px] font-semibold text-[var(--color-ink)]">{platform.name}</span>
                      {isDetected && (
                        <span className="rounded-sm bg-[rgba(201,123,46,0.16)] px-2 py-0.5 font-mono text-[9px] uppercase tracking-[0.12em] text-[var(--color-pollen-deep)]">
                          Detected
                        </span>
                      )}
                    </span>
                    <span className="mt-1 block text-[12px] text-[var(--color-ink-mute)]">{platform.requirement}</span>
                  </span>
                </button>

                <div className="grid grid-cols-2 gap-5 text-[12px]">
                  <div>
                    <div className="font-mono text-[9px] uppercase tracking-[0.14em] text-[var(--color-ink-faint)]">Recommended</div>
                    <div className="mt-1 text-[var(--color-ink-soft)]">{downloads.primary.label}</div>
                  </div>
                  <div>
                    <div className="font-mono text-[9px] uppercase tracking-[0.14em] text-[var(--color-ink-faint)]">Alternative</div>
                    <a className="mt-1 inline-flex items-center gap-1.5 text-[var(--color-ink-mute)] hover:text-[var(--color-ink)]" href={downloads.secondary.href} download>
                      {downloads.secondary.label}
                      <span aria-hidden>→</span>
                    </a>
                  </div>
                </div>

                <button
                  type="button"
                  onClick={() => {
                    onSelect(platform.key);
                    onDownload(downloads.primary.href, downloads.primary.label);
                  }}
                  aria-label={`Download ${downloads.primary.label} for ${platform.name}`}
                  className="inline-flex min-h-11 items-center justify-center gap-2 rounded-full border border-[rgba(20,17,11,0.18)] bg-[var(--color-paper)] px-5 text-[12.5px] font-semibold text-[var(--color-ink-soft)] transition-all hover:border-[rgba(201,123,46,0.5)] hover:bg-[rgba(201,123,46,0.1)] hover:text-[var(--color-ink)] sm:w-32"
                >
                  <DownloadGlyph size={14} />
                  Download
                </button>
              </div>
            );
          })}
        </div>
      </Reveal>
    </section>
  );
}

/* ── Setup + verify ── */
function SetupSection({ selected }: { selected: PlatformKey }) {
  const [copied, setCopied] = useState(false);
  const { push } = useToast();
  const platform = PLATFORMS.find((item) => item.key === selected)!;
  const steps = INSTALL_STEPS[selected];
  const command = VERIFY_COMMANDS[selected];

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);
      push({ title: "Copied", description: `${platform.name} verification command copied.` });
      window.setTimeout(() => setCopied(false), 1600);
    } catch {
      push({ title: "Copy failed", description: "Browser blocked clipboard access.", tone: "error" });
    }
  };

  return (
    <section className="mt-28 border-y border-[rgba(20,17,11,0.12)] bg-[var(--color-paper-warm)] -mx-6 px-6 py-20 sm:py-24">
      <div className="mx-auto grid max-w-[1100px] gap-14 lg:grid-cols-[1.1fr_0.9fr] lg:gap-24">
        <div>
          <Reveal>
            <p className="text-eyebrow mb-4">— {platform.name} setup</p>
            <h2 className="text-display text-[var(--color-ink)] max-w-[16ch] text-balance">Installed in four clear steps.</h2>
            <p className="text-lead mt-5 max-w-[44ch] text-[var(--color-ink-soft)] text-pretty">
              Condura asks for access only when a feature needs it. Permissions stay visible and reversible in your system settings.
            </p>
          </Reveal>

          <AnimatePresence mode="wait">
            <motion.ol
              key={selected}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.25 }}
              className="mt-10 divide-y divide-[rgba(20,17,11,0.10)] border-t border-[rgba(20,17,11,0.14)]"
            >
              {steps.map((step, index) => (
                <li key={step.title} className="grid grid-cols-[34px_1fr] gap-4 py-5">
                  <span className="font-mono text-[12px] text-[var(--color-synapse)]">0{index + 1}</span>
                  <div>
                    <h3 className="text-[14.5px] font-semibold text-[var(--color-ink)]">{step.title}</h3>
                    <p className="mt-1.5 text-[13px] leading-6 text-[var(--color-ink-mute)]">{step.desc}</p>
                  </div>
                </li>
              ))}
            </motion.ol>
          </AnimatePresence>
        </div>

        <Reveal delay={0.1}>
          <div className="lg:pt-8">
            <div className="text-mono-label">Verify the download</div>
            <h3 className="mt-3 max-w-sm font-display text-[28px] leading-tight tracking-[-0.025em] text-[var(--color-ink)]">
              Trust the artifact, not the page.
            </h3>
            <p className="mt-4 max-w-md text-[13.5px] leading-6 text-[var(--color-ink-mute)]">
              Run the checksum locally, then compare it with the value published alongside the release.
            </p>

            <div className="mt-8 overflow-hidden rounded-2xl border border-[rgba(20,17,11,0.14)] bg-[var(--color-ink)]">
              <div className="flex items-center justify-between border-b border-[rgba(244,239,228,0.1)] px-4 py-3">
                <span className="flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.14em] text-[rgba(244,239,228,0.55)]">
                  <span className="h-2 w-2 rounded-full bg-[var(--color-pollen)]" /> {platform.name}
                </span>
                <button
                  type="button"
                  onClick={copy}
                  className="flex min-h-8 items-center gap-1.5 rounded-md px-2 text-[11px] text-[rgba(244,239,228,0.6)] transition-colors hover:bg-[rgba(244,239,228,0.08)] hover:text-[var(--color-paper)]"
                >
                  {copied ? "Copied" : "Copy"}
                </button>
              </div>
              <AnimatePresence mode="wait">
                <motion.code
                  key={selected}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="block overflow-x-auto p-5 font-mono text-[12px] leading-6 text-[rgba(244,239,228,0.85)]"
                >
                  {command}
                </motion.code>
              </AnimatePresence>
            </div>

            <a href={RELEASE_TAG} target="_blank" rel="noopener noreferrer" className="thread-link mt-5 inline-flex items-center gap-2 text-[12.5px] font-medium text-[var(--color-ink-soft)]">
              Open checksums on GitHub <span aria-hidden>→</span>
            </a>

            <div className="mt-9 grid grid-cols-3 gap-3 border-t border-[rgba(20,17,11,0.12)] pt-6">
              {["Gatekeeper", "Local data", "Signed updates"].map((label) => (
                <div key={label} className="text-[11px] text-[var(--color-ink-mute)]">
                  <div className="mb-2 h-1.5 w-1.5 rounded-full bg-[var(--color-synapse)]" />
                  {label}
                </div>
              ))}
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  );
}

/* ── FAQ ── */
function FAQSection() {
  const [open, setOpen] = useState<string | null>("free");
  return (
    <section className="mt-28">
      <div className="grid gap-12 lg:grid-cols-[0.7fr_1.3fr] lg:gap-24">
        <Reveal>
          <p className="text-eyebrow mb-4">— Before you install</p>
          <h2 className="text-display text-[var(--color-ink)] max-w-[14ch] text-balance">Straight answers.</h2>
          <p className="text-lead mt-5 max-w-[42ch] text-[var(--color-ink-soft)] text-pretty">
            The details that matter before software receives access to your machine.
          </p>
        </Reveal>
        <Reveal delay={0.1}>
          <div className="divide-y divide-[rgba(20,17,11,0.12)] border-y border-[rgba(20,17,11,0.14)]">
            {FAQ_ITEMS.map((item) => {
              const isOpen = open === item.id;
              return (
                <button
                  key={item.id}
                  onClick={() => setOpen(isOpen ? null : item.id)}
                  className="block w-full py-5 text-left"
                >
                  <div className="flex items-center justify-between gap-4">
                    <h3 className={`font-display text-[19px] leading-tight transition-colors ${isOpen ? "text-[var(--color-ink)]" : "text-[var(--color-ink-soft)]"}`}>
                      {item.title}
                    </h3>
                    <span className={`font-mono text-[16px] text-[var(--color-synapse)] transition-transform ${isOpen ? "rotate-45" : "rotate-0"}`}>
                      +
                    </span>
                  </div>
                  <motion.div
                    initial={false}
                    animate={{ height: isOpen ? "auto" : 0, opacity: isOpen ? 1 : 0 }}
                    transition={{ duration: 0.4, ease: EASE_OUT }}
                    style={{ overflow: "hidden" }}
                  >
                    <p className="text-body mt-3 max-w-[58ch] text-[var(--color-ink-mute)]">{item.body}</p>
                  </motion.div>
                </button>
              );
            })}
          </div>
        </Reveal>
      </div>
    </section>
  );
}

/* ── bits ── */
function PlatformIcon({ platform, className = "", size = 16 }: { platform: string; className?: string; size?: number }) {
  if (platform === "mac")
    return (
      <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden>
        <path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.08 1.85-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09l.01-.01zM12 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z" />
      </svg>
    );
  if (platform === "windows")
    return (
      <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden>
        <path d="M3 5.1L10.4 4v7.3H3V5.1zM10.4 12.6v7.3L3 18.9v-6.3h7.4zM11.6 3.8L21 2.5v8.8h-9.4V3.8zM21 12.6v8.8l-9.4-1.3v-7.5H21z" />
      </svg>
    );
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden>
      <path d="M12 2C6.5 2 2 6.5 2 12c0 4.4 2.9 8.2 6.8 9.5.5.1.7-.2.7-.5v-1.7c-2.8.6-3.4-1.4-3.4-1.4-.5-1.1-1.1-1.4-1.1-1.4-.9-.6.1-.6.1-.6 1 .1 1.5 1 1.5 1 .9 1.5 2.3 1.1 2.9.8.1-.6.3-1.1.6-1.4-2.2-.3-4.5-1.1-4.5-5 0-1.1.4-2 1-2.7-.1-.3-.4-1.3.1-2.7 0 0 .8-.3 2.7 1 .8-.2 1.7-.3 2.5-.3s1.7.1 2.5.3c1.9-1.3 2.7-1 2.7-1 .5 1.4.2 2.4.1 2.7.6.7 1 1.6 1 2.7 0 3.9-2.3 4.7-4.5 5 .3.3.6.9.6 1.8v2.6c0 .3.2.6.7.5C19.1 20.2 22 16.4 22 12c0-5.5-4.5-10-10-10z" />
    </svg>
  );
}

function DownloadGlyph({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <path d="M12 3v12m0 0l-4-4m4 4l4-4M5 21h14" />
    </svg>
  );
}
