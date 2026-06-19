"use client";

import { motion, AnimatePresence } from "motion/react";
import { useState, useEffect, useCallback } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";
import TiltCard from "@/components/motion/TiltCard";
import MagneticButton from "@/components/motion/MagneticButton";
import BouncyAccordion from "@/components/motion/BouncyAccordion";
import ActionSwap from "@/components/motion/ActionSwap";
import { Icon, type IconKey } from "@/components/motion/Icon";
import { useToast } from "@/context/ToastContext";
import { useIsland } from "@/context/IslandContext";
import { usePlatform } from "@/hooks/usePlatform";
import { DOWNLOADS, RELEASE_TAG } from "@/lib/downloads";
import { PLATFORMS, SITE, type PlatformKey } from "@/lib/site";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   DOWNLOAD — Premium, focused, one shared platform state.

   Six sections, not ten:
     1. Hero + primary download (one viewport, one button)
     2. Platform cards (selecting updates everything below)
     3. Install guide (shares the selected platform)
     4. Verify (one code block, adapts to platform)
     5. FAQ
     6. Final CTA

   The download action NEVER navigates away. It uses an anchor
   with download attribute, so the page stays put and the browser
   handles the file transfer in the background.
   ──────────────────────────────────────────────────────────── */

const PLATFORM_ICON: Record<PlatformKey, IconKey> = {
  mac: "mac",
  windows: "windows",
  linux: "linux",
};

const INSTALL_STEPS: Record<PlatformKey, { title: string; desc: string }[]> = {
  mac: [
    { title: "Open the disk image", desc: "Double-click the .dmg and drag Condura.app into Applications." },
    { title: "Grant system access", desc: "Allow Accessibility and Screen Recording in System Settings → Privacy & Security." },
    { title: "Record your hotkey", desc: "Onboarding walks you through recording a global hotkey. No default — you pick it." },
    { title: "Summon the agent", desc: "Press your hotkey. The overlay appears. Start orchestrating." },
  ],
  windows: [
    { title: "Run the installer", desc: "Double-click the .exe. If SmartScreen warns, choose More Info → Run Anyway." },
    { title: "Record your hotkey", desc: "Onboarding lets you record a global hotkey. Suggested: Ctrl+Space or Ctrl+Ctrl." },
    { title: "Grant permissions", desc: "Condura requests the access it needs. Each grant is clear and reversible." },
    { title: "Summon the agent", desc: "Press your hotkey. The overlay appears. Start orchestrating." },
  ],
  linux: [
    { title: "Install the package", desc: "sudo dpkg -i condura_0.1.0_linux_amd64.deb — or chmod +x the AppImage." },
    { title: "Start the daemon", desc: "The systemd user service starts automatically. Verify with condura status." },
    { title: "Record your hotkey", desc: "Onboarding walks you through recording a global hotkey." },
    { title: "Open the TUI", desc: "Run condura-tui in your terminal, or use the overlay from your hotkey." },
  ],
};

const VERIFY_COMMANDS: Record<PlatformKey, string> = {
  mac: `# Verify the macOS installer
shasum -a 256 condura-installer-mac.dmg

# Compare with the hash on the GitHub release page
# ${RELEASE_TAG}`,
  windows: `# Verify the Windows installer (PowerShell)
Get-FileHash condura-setup.exe -Algorithm SHA256

# Compare with the hash on the GitHub release page
# ${RELEASE_TAG}`,
  linux: `# Verify the Linux package
sha256sum condura.deb

# Compare with the hash on the GitHub release page
# ${RELEASE_TAG}`,
};

const FAQ_ITEMS = [
  { id: "free", title: "Is it really free?", body: "Yes. No feature gates, no premium tier, no nags. Free for personal and commercial use under the Condura Freeware EULA. There's a donate button in the menu bar if you want to support development." },
  { id: "keys", title: "Do I need an API key?", body: "Not to start. Condura auto-detects local Ollama and any installed CLI tools. You can use it fully offline with local models. Add API keys in Settings only if you want cloud providers." },
  { id: "privacy", title: "What happens to my data?", body: "Everything stays on your machine. Memory, skills, audit logs, embeddings — all local, encrypted at rest. The only network calls are to the LLM provider(s) you configured. No telemetry. No tracking. Ever." },
  { id: "safety", title: "How does it stay safe?", body: "A deterministic Go rules engine — the Gatekeeper — evaluates every action before it executes. No model output reaches a click, keystroke, or shell command without passing through it. Destructive actions require a native dialog with a real human at the keyboard." },
  { id: "uninstall", title: "What if I want to uninstall?", body: "Condura auto-backs-up your data to ~/Documents/condura-backups/ before uninstalling. No cloud account to cancel. No data sitting on someone else's server." },
  { id: "update", title: "How do updates work?", body: "Condura checks GitHub Releases every 6 hours and on launch. Updates are signed with Ed25519 and applied atomically with rollback on failure. You can choose stable, beta, or dev channels." },
];

export default function DownloadPageView() {
  const detected = usePlatform();
  const [selected, setSelected] = useState<PlatformKey>(detected);
  const [mounted, setMounted] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const { push } = useToast();
  const { pulseDownload } = useIsland();

  useEffect(() => { const t = setTimeout(() => setMounted(true), 200); return () => clearTimeout(t); }, []);

  const platformMeta = PLATFORMS.find((p) => p.key === selected)!;

  // The download trigger: create a temporary anchor with the download
  // attribute and click it. The browser handles the file transfer in the
  // background — the page never navigates away. If the release doesn't
  // exist, the browser simply shows a failed download, not an error page.
  const triggerDownload = useCallback((href: string, label: string) => {
    const a = document.createElement("a");
    a.href = href;
    a.setAttribute("download", "");
    a.rel = "noopener";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);

    setDownloading(true);
    pulseDownload(platformMeta.name);
    push({ title: "Download started", description: `${label} for ${platformMeta.name}.`, tone: "success" });
    setTimeout(() => setDownloading(false), 2000);
  }, [platformMeta.name, pulseDownload, push]);

  return (
    <main className="relative w-full bg-black text-white">
      <Hero mounted={mounted} selected={selected} detected={detected} downloading={downloading} onDownload={triggerDownload} />
      <PlatformCards selected={selected} detected={detected} onSelect={setSelected} onDownload={triggerDownload} />
      <InstallGuide selected={selected} />
      <VerifySection selected={selected} />
      <FAQSection />
      <FinalCTA selected={selected} onDownload={triggerDownload} />
    </main>
  );
}

/* ════════════════════════════════════════════════════════════
   1. HERO — One headline, one button, one detail row.
   ════════════════════════════════════════════════════════════ */

function Hero({
  mounted,
  selected,
  detected,
  downloading,
  onDownload,
}: {
  mounted: boolean;
  selected: PlatformKey;
  detected: PlatformKey;
  downloading: boolean;
  onDownload: (href: string, label: string) => void;
}) {
  const platformMeta = PLATFORMS.find((p) => p.key === selected)!;
  const current = DOWNLOADS[selected];
  const isDetected = selected === detected;

  return (
    <section className="relative min-h-[85vh] flex flex-col items-center justify-center px-6 overflow-hidden">
      <div className="absolute inset-0 bg-grid-dark opacity-15" />
      {/* Soft ambient glow */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ scale: [1, 1.1, 1], opacity: [0.03, 0.07, 0.03] }}
          transition={{ duration: 5, repeat: Infinity, ease: "easeInOut" }}
          className="w-[500px] h-[400px] rounded-full bg-white blur-[160px]"
        />
      </div>

      <div className="relative z-10 max-w-2xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 24 }}
          transition={{ duration: 0.9, ease: EASE_OUT }}
        >
          <div className="mb-6 flex justify-center">
            <AnimatedBadge tone="neutral" pulse>v0.1.0 Open Alpha</AnimatedBadge>
          </div>

          <h1 className="font-display text-[clamp(2.5rem,6vw,4.5rem)] font-semibold leading-[1.05] tracking-[-0.04em]">
            Get Condura.
          </h1>
          <p className="mt-6 font-lead-airy max-w-lg mx-auto">
            One download. Three platforms. No account required. Runs entirely on your machine —
            your keys, your models, your data.
          </p>

          {/* Primary download button */}
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 16 }}
            transition={{ delay: 0.3, duration: 0.7 }}
            className="mt-10 flex flex-col items-center gap-5"
          >
            <button
              onClick={() => onDownload(current.primary.href, current.primary.label)}
              className="mature-button group inline-flex items-center gap-3 px-10 py-5 text-[16px] font-semibold transition-transform"
            >
              <motion.span
                animate={downloading ? { rotate: 360 } : { rotate: 0 }}
                transition={{ duration: 0.8, repeat: downloading ? Infinity : 0, ease: "linear" }}
              >
                <Icon name={downloading ? "command" : "download"} size={20} />
              </motion.span>
              {downloading ? "Starting…" : `Download for ${platformMeta.name}`}
            </button>

            {/* Detail row */}
            <div className="flex flex-wrap items-center justify-center gap-x-4 gap-y-2 font-mono text-[11px] text-white/30">
              {isDetected && (
                <span className="flex items-center gap-1.5 text-green-400/60">
                  <span className="h-1.5 w-1.5 rounded-full bg-green-400/60" />
                  Detected
                </span>
              )}
              <span className="flex items-center gap-1.5">
                <Icon name="shield" size={12} />
                Signed & notarized
              </span>
              <span className="flex items-center gap-1.5">
                <Icon name="check" size={12} />
                {current.primary.label}
              </span>
              <a href={RELEASE_TAG} target="_blank" rel="noopener noreferrer" className="hover:text-white/60 transition-colors">
                Release notes →
              </a>
            </div>
          </motion.div>
        </motion.div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   2. PLATFORM CARDS — Selecting updates everything below.
   Each card is a complete unit: icon, name, requirements,
   download buttons, all in one place.
   ════════════════════════════════════════════════════════════ */

function PlatformCards({
  selected,
  detected,
  onSelect,
  onDownload,
}: {
  selected: PlatformKey;
  detected: PlatformKey;
  onSelect: (k: PlatformKey) => void;
  onDownload: (href: string, label: string) => void;
}) {
  return (
    <section className="relative w-full py-24 px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <SectionHeader eyebrow="All platforms" title="Choose your build." />

        <div className="grid md:grid-cols-3 gap-6">
          {PLATFORMS.map((p) => {
            const isActive = selected === p.key;
            const isDetected = detected === p.key;
            const downloads = DOWNLOADS[p.key];

            return (
              <TiltCard key={p.key} maxRotate={4} className="h-full">
                <motion.div
                  whileHover={{ y: -4 }}
                  className={`relative h-full overflow-hidden rounded-2xl border p-6 transition-colors ${
                    isActive ? "border-white/25 bg-white/[0.05]" : "border-white/[0.08] bg-white/[0.02]"
                  }`}
                >
                  {isActive && (
                    <motion.div
                      layoutId="platform-glow"
                      className="absolute inset-0 bg-gradient-to-b from-white/[0.04] to-transparent pointer-events-none"
                    />
                  )}

                  <div className="relative z-10 flex h-full flex-col">
                    {/* Header */}
                    <div className="flex items-center justify-between mb-5">
                      <div className={`flex h-11 w-11 items-center justify-center rounded-xl border ${isActive ? "border-white/20 bg-white/[0.08]" : "border-white/10 bg-white/[0.03]"}`}>
                        <Icon name={PLATFORM_ICON[p.key]} size={22} className={isActive ? "text-white/85" : "text-white/50"} />
                      </div>
                      {isDetected && (
                        <span className="flex items-center gap-1.5 rounded-full border border-green-400/20 bg-green-400/10 px-2.5 py-0.5 font-mono text-[10px] text-green-400/70">
                          <span className="h-1.5 w-1.5 rounded-full bg-green-400/60" />
                          Detected
                        </span>
                      )}
                    </div>

                    {/* Name + requirements */}
                    <h3 className="font-body-mature text-[18px] font-semibold text-white">{p.name}</h3>
                    <p className="mt-1 font-mono text-[12px] text-white/35">{p.requirement}</p>

                    {/* Downloads */}
                    <div className="mt-6 space-y-2.5 border-t border-white/[0.06] pt-5 flex-1">
                      <button
                        onClick={() => {
                          onSelect(p.key);
                          onDownload(downloads.primary.href, downloads.primary.label);
                        }}
                        className={`flex w-full items-center justify-between rounded-xl px-4 py-3 text-left text-[13px] font-medium transition-colors ${
                          isActive
                            ? "bg-white/[0.10] text-white"
                            : "bg-white/[0.03] text-white/60 hover:bg-white/[0.06] hover:text-white"
                        }`}
                      >
                        <span className="flex items-center gap-2">
                          <Icon name="download" size={14} />
                          {downloads.primary.label}
                        </span>
                        <Icon name="arrowRight" size={14} className="opacity-40" />
                      </button>
                      <a
                        href={downloads.secondary.href}
                        download
                        onClick={() => onSelect(p.key)}
                        className="flex w-full items-center justify-between rounded-xl px-4 py-3 text-left text-[13px] text-white/50 transition-colors hover:bg-white/[0.04] hover:text-white"
                      >
                        <span>{downloads.secondary.label}</span>
                        <Icon name="arrowRight" size={14} className="opacity-30" />
                      </a>
                    </div>
                  </div>
                </motion.div>
              </TiltCard>
            );
          })}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   3. INSTALL GUIDE — Shares the platform selection from above.
   No separate platform switcher — it just updates.
   ════════════════════════════════════════════════════════════ */

function InstallGuide({ selected }: { selected: PlatformKey }) {
  const steps = INSTALL_STEPS[selected];

  return (
    <section className="relative w-full py-24 px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-3xl">
        <div className="mb-12 flex items-baseline justify-between">
          <div>
            <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Installation</span>
            <h2 className="mt-3 font-display text-[clamp(1.5rem,3.5vw,2.5rem)] font-semibold tracking-[-0.03em]">
              Up in four steps.
            </h2>
          </div>
          <span className="font-mono text-[12px] text-white/40">
            {PLATFORMS.find((p) => p.key === selected)!.name}
          </span>
        </div>

        <AnimatePresence mode="wait">
          <motion.div
            key={selected}
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: 10 }}
            transition={{ duration: 0.3 }}
            className="relative"
          >
            <div className="absolute left-[20px] top-0 bottom-0 w-[1px] bg-gradient-to-b from-white/20 via-white/10 to-transparent" />
            <div className="space-y-6">
              {steps.map((step, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, x: -16 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: i * 0.08, duration: 0.4, ease: EASE_OUT }}
                  className="relative flex gap-5"
                >
                  <div className="relative z-10 flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-white/15 bg-black">
                    <span className="font-mono text-[12px] text-white/55">{i + 1}</span>
                  </div>
                  <div className="flex-1 rounded-xl border border-white/[0.08] bg-white/[0.02] p-4">
                    <h3 className="font-body-mature text-[15px] font-semibold text-white">{step.title}</h3>
                    <p className="mt-1.5 font-body-mature text-[13.5px] text-white/45 leading-relaxed">{step.desc}</p>
                  </div>
                </motion.div>
              ))}
            </div>
          </motion.div>
        </AnimatePresence>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   4. VERIFY — One code block, adapts to selected platform.
   ════════════════════════════════════════════════════════════ */

function VerifySection({ selected }: { selected: PlatformKey }) {
  const [copied, setCopied] = useState(false);
  const { push } = useToast();
  const code = VERIFY_COMMANDS[selected];
  const platformName = PLATFORMS.find((p) => p.key === selected)!.name;

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      push({ title: "Copied", description: `${platformName} verify command is on your clipboard.` });
      setTimeout(() => setCopied(false), 1600);
    } catch {
      push({ title: "Copy failed", description: "Browser blocked clipboard access.", tone: "error" });
    }
  };

  return (
    <section className="relative w-full py-24 px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-2xl">
        <SectionHeader eyebrow="Verify" title="Check before you install." />

        <div className="overflow-hidden rounded-2xl border border-white/[0.10] bg-[#0a0a0a]">
          <div className="flex items-center justify-between border-b border-white/[0.06] bg-white/[0.02] px-5 py-3">
            <span className="flex items-center gap-2 font-mono text-[11px] text-white/40">
              <Icon name="terminal" size={14} />
              {platformName}
            </span>
            <button
              onClick={copy}
              className="flex items-center gap-1.5 rounded-md border border-white/[0.06] bg-white/[0.03] px-2.5 py-1 font-mono text-[10px] text-white/50 hover:text-white hover:bg-white/[0.06] transition-colors"
            >
              <Icon name="copy" size={12} />
              <ActionSwap primary="Copy" secondary="Copied" active={copied} />
            </button>
          </div>
          <AnimatePresence mode="wait">
            <motion.pre
              key={selected}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="p-6 font-mono text-[12.5px] leading-relaxed text-white/55 overflow-x-auto"
            >
              {code}
            </motion.pre>
          </AnimatePresence>
        </div>

        {/* Trust badges */}
        <div className="mt-10 flex flex-wrap items-center justify-center gap-8">
          {[
            { icon: "shield" as IconKey, label: "Code signed" },
            { icon: "check" as IconKey, label: "Notarized" },
            { icon: "lock" as IconKey, label: "Ed25519 verified" },
          ].map((badge) => (
            <div key={badge.label} className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-full border border-white/15 bg-white/[0.04]">
                <Icon name={badge.icon} size={15} className="text-white/60" />
              </div>
              <span className="font-body-mature text-[13px] text-white/50">{badge.label}</span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   5. FAQ
   ════════════════════════════════════════════════════════════ */

function FAQSection() {
  return (
    <section className="relative w-full py-24 px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-2xl">
        <SectionHeader eyebrow="FAQ" title="Questions, answered." />
        <BouncyAccordion items={FAQ_ITEMS} defaultOpenId="free" />
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   6. FINAL CTA
   ════════════════════════════════════════════════════════════ */

function FinalCTA({
  selected,
  onDownload,
}: {
  selected: PlatformKey;
  onDownload: (href: string, label: string) => void;
}) {
  const platformMeta = PLATFORMS.find((p) => p.key === selected)!;
  const current = DOWNLOADS[selected];

  return (
    <section className="relative w-full py-32 px-6 border-t border-white/[0.08] overflow-hidden">
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ scale: [1, 1.12, 1], opacity: [0.04, 0.08, 0.04] }}
          transition={{ duration: 5, repeat: Infinity }}
          className="w-[500px] h-[250px] rounded-full bg-white blur-[140px]"
        />
      </div>

      <div className="relative z-10 mx-auto max-w-2xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, ease: EASE_OUT }}
        >
          <h2 className="font-display text-[clamp(1.75rem,5vw,3rem)] font-semibold tracking-[-0.04em] leading-[1.1]">
            Your machine. Your conductor.
          </h2>
          <p className="mt-6 font-lead-airy mx-auto max-w-md">
            Set your hotkey. Let it orchestrate every AI tool you already have.
          </p>
          <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-3">
            <button
              onClick={() => onDownload(current.primary.href, current.primary.label)}
              className="mature-button inline-flex items-center gap-2 px-8 py-4 text-[14px] font-semibold"
            >
              <Icon name="rocket" size={16} />
              Download for {platformMeta.name}
            </button>
            <MagneticButton
              href="/manifesto"
              className="mature-button-secondary rounded-full px-6 py-4 text-[13px]"
            >
              Read the mission
            </MagneticButton>
          </div>
          <p className="mt-6 font-mono text-[11px] text-white/25">
            {SITE.name} v0.1.0 · Free for personal & commercial use ·{" "}
            <a href="/legal" className="underline hover:text-white/50">EULA</a>
          </p>
        </motion.div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   Shared section header
   ════════════════════════════════════════════════════════════ */

function SectionHeader({ eyebrow, title }: { eyebrow: string; title: string }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 14 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.5, ease: EASE_OUT }}
      className="mb-12 text-center"
    >
      <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">{eyebrow}</span>
      <h2 className="mt-3 font-display text-[clamp(1.5rem,3.5vw,2.5rem)] font-semibold tracking-[-0.03em]">{title}</h2>
    </motion.div>
  );
}