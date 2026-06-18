"use client";

import { motion, AnimatePresence } from "motion/react";
import { useState, useEffect } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";
import TiltCard from "@/components/motion/TiltCard";
import MagneticButton from "@/components/motion/MagneticButton";
import StatefulButton from "@/components/motion/StatefulButton";
import BouncyAccordion from "@/components/motion/BouncyAccordion";
import { useToast } from "@/context/ToastContext";
import { useIsland } from "@/context/IslandContext";
import { usePlatform } from "@/hooks/usePlatform";
import { DOWNLOADS, RELEASE_TAG } from "@/lib/downloads";
import { PLATFORMS, SITE, type PlatformKey } from "@/lib/site";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   DOWNLOAD — The Most Important Page
   This is where the user commits. Every section is designed
   to build confidence: the orb, the platform cards, the
   feature highlights, the install timeline, requirements,
   verification, and FAQ.
   ──────────────────────────────────────────────────────────── */

const PLATFORM_ICONS: Record<PlatformKey, string> = {
  mac: "",
  windows: "⊞",
  linux: "🐧",
};

const INSTALL_STEPS: Record<PlatformKey, { title: string; desc: string }[]> = {
  mac: [
    { title: "Open the .dmg", desc: "Double-click the downloaded disk image. Drag Condura.app into Applications." },
    { title: "Grant permissions", desc: "Open Condura. Grant Accessibility and Screen Recording in System Settings → Privacy & Security." },
    { title: "Set your hotkey", desc: "Onboarding walks you through recording a global hotkey. No default — you pick it." },
    { title: "Summon the agent", desc: "Press your hotkey. The overlay appears. Start orchestrating." },
  ],
  windows: [
    { title: "Run the installer", desc: "Double-click the downloaded .exe. Windows SmartScreen may warn — click More Info → Run Anyway." },
    { title: "Set your hotkey", desc: "Onboarding lets you record a global hotkey. Suggested: Ctrl+Space or Ctrl+Ctrl." },
    { title: "Grant permissions", desc: "Condura requests the access it needs. Each grant is clear and reversible." },
    { title: "Summon the agent", desc: "Press your hotkey. The overlay appears. Start orchestrating." },
  ],
  linux: [
    { title: "Install the package", desc: "Run sudo dpkg -i condura_0.1.0_linux_amd64.deb or extract the AppImage and chmod +x." },
    { title: "Start the daemon", desc: "The systemd user service starts automatically. Verify with condura status." },
    { title: "Set your hotkey", desc: "Onboarding walks you through recording a global hotkey." },
    { title: "Open the TUI", desc: "Run condura-tui in your terminal, or use the overlay from your hotkey." },
  ],
};

const FEATURES = [
  { icon: "⚡", title: "Cold start < 500ms", desc: "The daemon lives in your menu bar. Hotkey to overlay in under 100ms." },
  { icon: "🔒", title: "Local-first by default", desc: "Memory, skills, and audit log on disk. Your API keys never leave your machine." },
  { icon: "🆓", title: "Free forever", desc: "No subscriptions. No premium tier. No nags. A donate button, that's it." },
  { icon: "🌐", title: "12+ LLM providers", desc: "Route through Anthropic, OpenAI, Google, xAI, Mistral, DeepSeek, or 100% local Ollama." },
  { icon: "🛡️", title: "Deterministic safety", desc: "A pure-rules gatekeeper. No model can bypass it. Every action is audited." },
  { icon: "🖥️", title: "Native on every OS", desc: "macOS, Windows, Linux. Signed, notarized, and auto-updated on each." },
];

const FAQ_ITEMS = [
  {
    id: "free",
    title: "Is it really free?",
    body: "Yes. No feature gates, no premium tier, no nags. Free for personal and commercial use under the Condura Freeware EULA. There's a donate button in the menu bar if you want to support development.",
  },
  {
    id: "keys",
    title: "Do I need an API key?",
    body: "Not to start. Condura auto-detects local Ollama and any installed CLI tools (Claude Code, Codex, etc.). You can use it fully offline with local models. Add API keys in Settings only if you want cloud providers.",
  },
  {
    id: "privacy",
    title: "What happens to my data?",
    body: "Everything stays on your machine. Memory, skills, audit logs, embeddings — all local, encrypted at rest. The only network calls are to the LLM provider(s) you configured. No telemetry. No tracking. Ever.",
  },
  {
    id: "safety",
    title: "How does it stay safe?",
    body: "A deterministic Go rules engine — the Gatekeeper — evaluates every action before it executes. No model output reaches a click, keystroke, or shell command without passing through it. Destructive actions require a native dialog with a real human at the keyboard.",
  },
  {
    id: "uninstall",
    title: "What if I want to uninstall?",
    body: "Condura auto-backs-up your data to ~/Documents/condura-backups/ before uninstalling. No cloud account to cancel. No data sitting on someone else's server.",
  },
  {
    id: "update",
    title: "How do updates work?",
    body: "Condura checks GitHub Releases every 6 hours and on launch. Updates are signed with Ed25519 and applied atomically with rollback on failure. You can choose stable, beta, or dev channels.",
  },
];

export default function DownloadPageView() {
  return (
    <main className="relative w-full bg-black text-white overflow-hidden">
      <DownloadHero />
      <PlatformSelector />
      <WhyDownload />
      <InstallTimeline />
      <SystemRequirements />
      <VerificationSection />
      <FAQSection />
      <FinalCTA />
    </main>
  );
}

/* ════════════════════════════════════════════════════════════
   1. HERO — The Download Orb
   ════════════════════════════════════════════════════════════ */

function DownloadHero() {
  const detected = usePlatform();
  const { push } = useToast();
  const { pulseDownload } = useIsland();
  const [mounted, setMounted] = useState(false);
  useEffect(() => { const t = setTimeout(() => setMounted(true), 200); return () => clearTimeout(t); }, []);

  const platformMeta = PLATFORMS.find((p) => p.key === detected)!;
  const current = DOWNLOADS[detected];

  const startDownload = async () => {
    pulseDownload(platformMeta.name);
    push({
      title: "Download started",
      description: `${current.primary.label} for ${platformMeta.name}.`,
      tone: "success",
    });
    window.location.href = current.primary.href;
    return true;
  };

  return (
    <section className="relative min-h-screen flex flex-col items-center justify-center px-6 overflow-hidden">
      {/* Ambient grid */}
      <div className="absolute inset-0 bg-grid-dark opacity-20" />

      {/* The Download Orb — pulsing concentric rings */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ scale: [1, 1.15, 1], opacity: [0.05, 0.1, 0.05] }}
          transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
          className="w-[500px] h-[500px] rounded-full bg-white blur-[150px]"
        />
        {[300, 400, 500, 600].map((size, i) => (
          <motion.div
            key={size}
            animate={{ scale: [1, 1.08, 1], opacity: [0.15, 0.05, 0.15] }}
            transition={{ duration: 3, repeat: Infinity, delay: i * 0.4, ease: "easeInOut" }}
            className="absolute rounded-full border border-white/[0.06]"
            style={{ width: size, height: size }}
          />
        ))}
      </div>

      <div className="relative z-10 max-w-3xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 30 }}
          transition={{ duration: 1, ease: EASE_OUT }}
        >
          <div className="mb-8 flex justify-center">
            <AnimatedBadge tone="neutral" pulse>v0.1.0 Open Alpha</AnimatedBadge>
          </div>

          <h1 className="font-display text-[clamp(2.5rem,7vw,5rem)] font-semibold leading-[1.05] tracking-[-0.04em]">
            Get Condura.
            <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-white via-white to-white/30">
              Free. Forever.
            </span>
          </h1>

          <p className="mt-8 mx-auto max-w-xl font-lead-airy">
            One download. Three platforms. No account required. Condura runs entirely on your
            machine — your keys, your models, your data.
          </p>

          {/* Detected platform + primary CTA */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 20 }}
            transition={{ delay: 0.4, duration: 0.8 }}
            className="mt-12 flex flex-col items-center gap-6"
          >
            <div className="flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.04] px-4 py-2">
              <span className="flex h-2 w-2 rounded-full bg-green-400/60 shadow-[0_0_8px_rgba(74,222,128,0.4)]" />
              <span className="font-mono text-[12px] text-white/50">
                Detected: {platformMeta.name} · {platformMeta.requirement}
              </span>
            </div>

            <StatefulButton
              className="mature-button text-[16px] px-10 py-5"
              idleLabel={`↓ Download for ${platformMeta.name}`}
              loadingLabel="Starting…"
              successLabel="Download started ✓"
              onAction={startDownload}
            />

            <div className="flex items-center gap-4 font-mono text-[11px] text-white/30">
              <span>{current.primary.label}</span>
              <span className="h-3 w-[1px] bg-white/15" />
              <span>Signed & notarized</span>
              <span className="h-3 w-[1px] bg-white/15" />
              <a href={RELEASE_TAG} target="_blank" rel="noopener noreferrer" className="hover:text-white/60 transition-colors">
                Release notes →
              </a>
            </div>
          </motion.div>
        </motion.div>
      </div>

      {/* Scroll indicator */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: mounted ? 1 : 0 }}
        transition={{ delay: 1.5, duration: 1 }}
        className="absolute bottom-10 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2"
      >
        <span className="font-mono text-[10px] uppercase tracking-widest text-white/25">All platforms</span>
        <div className="w-[1px] h-10 bg-gradient-to-b from-white/25 to-transparent" />
      </motion.div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   2. PLATFORM SELECTOR — Three Visual Cards
   ════════════════════════════════════════════════════════════ */

function PlatformSelector() {
  const detected = usePlatform();
  const [selected, setSelected] = useState<PlatformKey>(detected);
  const { push } = useToast();
  const { pulseDownload } = useIsland();

  const current = DOWNLOADS[selected];
  const platformMeta = PLATFORMS.find((p) => p.key === selected)!;

  const handleDownload = () => {
    pulseDownload(platformMeta.name);
    push({
      title: "Download started",
      description: `${current.primary.label} for ${platformMeta.name}.`,
      tone: "success",
    });
    window.location.href = current.primary.href;
  };

  return (
    <section className="relative w-full py-[140px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-16 text-center">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Choose your platform</span>
          <h2 className="mt-4 font-display text-[clamp(1.75rem,4vw,3rem)] font-semibold tracking-[-0.03em]">
            Three platforms. One agent.
          </h2>
        </div>

        {/* Platform cards */}
        <div className="grid md:grid-cols-3 gap-6">
          {PLATFORMS.map((p) => {
            const isActive = selected === p.key;
            const isDetected = detected === p.key;
            return (
              <TiltCard key={p.key} maxRotate={6} className="h-full">
                <motion.button
                  onClick={() => setSelected(p.key)}
                  whileHover={{ y: -4 }}
                  className={`relative h-full w-full overflow-hidden rounded-2xl border p-8 text-left transition-colors ${
                    isActive
                      ? "border-white/25 bg-white/[0.06]"
                      : "border-white/[0.08] bg-white/[0.02] hover:bg-white/[0.04]"
                  }`}
                >
                  {/* Active glow */}
                  {isActive && (
                    <motion.div
                      layoutId="platform-glow"
                      className="absolute inset-0 bg-gradient-to-b from-white/[0.05] to-transparent"
                    />
                  )}

                  <div className="relative z-10">
                    {/* Icon */}
                    <div className={`flex h-14 w-14 items-center justify-center rounded-2xl border ${isActive ? "border-white/20 bg-white/[0.08]" : "border-white/10 bg-white/[0.03]"}`}>
                      <span className="text-[24px]">{PLATFORM_ICONS[p.key]}</span>
                    </div>

                    {/* Name */}
                    <h3 className="mt-6 font-body-mature text-[20px] font-semibold text-white">{p.name}</h3>
                    <p className="mt-1 font-body-mature text-[13px] text-white/40">{p.requirement}</p>

                    {/* Detected badge */}
                    {isDetected && (
                      <div className="mt-4 inline-flex items-center gap-1.5 rounded-full border border-green-400/20 bg-green-400/10 px-2.5 py-0.5">
                        <span className="h-1.5 w-1.5 rounded-full bg-green-400/60" />
                        <span className="font-mono text-[10px] text-green-400/70">Detected</span>
                      </div>
                    )}

                    {/* Artifacts */}
                    <div className="mt-6 space-y-2 border-t border-white/[0.06] pt-4">
                      <div className="flex items-center justify-between">
                        <span className="font-mono text-[11px] text-white/30">Primary</span>
                        <span className="font-mono text-[11px] text-white/60">{DOWNLOADS[p.key].primary.label}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="font-mono text-[11px] text-white/30">Secondary</span>
                        <span className="font-mono text-[11px] text-white/50">{DOWNLOADS[p.key].secondary.label}</span>
                      </div>
                    </div>
                  </div>
                </motion.button>
              </TiltCard>
            );
          })}
        </div>

        {/* Download bar for selected platform */}
        <AnimatePresence mode="wait">
          <motion.div
            key={selected}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.3 }}
            className="mt-8 mature-panel rounded-2xl p-6 flex flex-col sm:flex-row items-center justify-between gap-6"
          >
            <div>
              <h3 className="font-body-mature text-[16px] font-semibold text-white">
                {platformMeta.name} · {current.primary.label}
              </h3>
              <p className="mt-1 font-body-mature text-[13px] text-white/40">
                Requires {platformMeta.requirement}
              </p>
            </div>
            <div className="flex items-center gap-3">
              <MagneticButton
                onClick={handleDownload}
                className="mature-button rounded-full px-6 py-3 text-[14px] font-semibold"
              >
                ↓ Download now
              </MagneticButton>
              <a
                href={current.secondary.href}
                className="mature-button-secondary rounded-full px-5 py-3 text-[13px]"
              >
                {current.secondary.label}
              </a>
            </div>
          </motion.div>
        </AnimatePresence>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   3. WHY DOWNLOAD — Feature Highlights
   ════════════════════════════════════════════════════════════ */

function WhyDownload() {
  return (
    <section className="relative w-full py-[140px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-16 text-center">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Why Condura</span>
          <h2 className="mt-4 font-display text-[clamp(1.75rem,4vw,3rem)] font-semibold tracking-[-0.03em]">
            What you get.
          </h2>
        </div>

        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {FEATURES.map((feat, i) => (
            <motion.div
              key={feat.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.06 }}
              className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 hover:bg-white/[0.04] transition-colors"
            >
              <span className="text-[24px]">{feat.icon}</span>
              <h3 className="mt-4 font-body-mature text-[16px] font-semibold text-white">{feat.title}</h3>
              <p className="mt-2 font-body-mature text-[14px] text-white/45 leading-relaxed">{feat.desc}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   4. INSTALL TIMELINE — Step by Step
   ════════════════════════════════════════════════════════════ */

function InstallTimeline() {
  const detected = usePlatform();
  const [selected, setSelected] = useState<PlatformKey>(detected);
  const steps = INSTALL_STEPS[selected];

  return (
    <section className="relative w-full py-[140px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-4xl">
        <div className="mb-12 text-center">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Installation</span>
          <h2 className="mt-4 font-display text-[clamp(1.75rem,4vw,3rem)] font-semibold tracking-[-0.03em]">
            Up in 4 steps.
          </h2>
        </div>

        {/* Platform toggle */}
        <div className="mb-12 flex justify-center">
          <div className="inline-flex rounded-full border border-white/[0.08] bg-white/[0.03] p-1">
            {PLATFORMS.map((p) => (
              <button
                key={p.key}
                onClick={() => setSelected(p.key)}
                className={`rounded-full px-5 py-2 font-body-mature text-[13px] font-medium transition-colors ${
                  selected === p.key ? "bg-white/[0.10] text-white" : "text-white/50 hover:text-white"
                }`}
              >
                {p.name}
              </button>
            ))}
          </div>
        </div>

        {/* Timeline */}
        <AnimatePresence mode="wait">
          <motion.div
            key={selected}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.3 }}
            className="relative"
          >
            {/* Vertical line */}
            <div className="absolute left-[20px] top-0 bottom-0 w-[1px] bg-gradient-to-b from-white/20 via-white/10 to-transparent" />

            <div className="space-y-8">
              {steps.map((step, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: i * 0.1 }}
                  className="relative flex gap-6"
                >
                  <div className="relative z-10 flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-white/15 bg-black">
                    <span className="font-mono text-[12px] text-white/50">{i + 1}</span>
                    <motion.div
                      animate={{ scale: [1, 1.4], opacity: [0.3, 0] }}
                      transition={{ duration: 2, repeat: Infinity, delay: i * 0.3 }}
                      className="absolute inset-0 rounded-full border border-white/20"
                    />
                  </div>
                  <div className="flex-1 rounded-2xl border border-white/[0.08] bg-white/[0.02] p-5">
                    <h3 className="font-body-mature text-[16px] font-semibold text-white">{step.title}</h3>
                    <p className="mt-2 font-body-mature text-[14px] text-white/45 leading-relaxed">{step.desc}</p>
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
   5. SYSTEM REQUIREMENTS
   ════════════════════════════════════════════════════════════ */

function SystemRequirements() {
  return (
    <section className="relative w-full py-[140px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-16 text-center">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Requirements</span>
          <h2 className="mt-4 font-display text-[clamp(1.75rem,4vw,3rem)] font-semibold tracking-[-0.03em]">
            Will it run on your machine?
          </h2>
        </div>

        <div className="grid md:grid-cols-3 gap-6">
          {PLATFORMS.map((p, i) => (
            <motion.div
              key={p.key}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.08 }}
              className="mature-panel rounded-2xl p-6"
            >
              <div className="flex items-center gap-3 mb-4">
                <span className="text-[20px]">{PLATFORM_ICONS[p.key]}</span>
                <h3 className="font-body-mature text-[17px] font-semibold text-white">{p.name}</h3>
              </div>
              <dl className="space-y-3">
                <div className="flex justify-between">
                  <dt className="font-mono text-[11px] uppercase tracking-wider text-white/30">OS</dt>
                  <dd className="font-body-mature text-[13px] text-white/60">{p.requirement}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="font-mono text-[11px] uppercase tracking-wider text-white/30">RAM</dt>
                  <dd className="font-body-mature text-[13px] text-white/60">4 GB min</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="font-mono text-[11px] uppercase tracking-wider text-white/30">Disk</dt>
                  <dd className="font-body-mature text-[13px] text-white/60">~50 MB</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="font-mono text-[11px] uppercase tracking-wider text-white/30">Network</dt>
                  <dd className="font-body-mature text-[13px] text-white/60">Optional*</dd>
                </div>
              </dl>
            </motion.div>
          ))}
        </div>

        <p className="mt-8 text-center font-body-mature text-[13px] text-white/30">
          * Network only needed for cloud LLM providers. Local Ollama mode works fully offline.
        </p>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   6. VERIFICATION — Checksum Verification
   ════════════════════════════════════════════════════════════ */

function VerificationSection() {
  return (
    <section className="relative w-full py-[140px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-4xl">
        <div className="mb-12 text-center">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Verify</span>
          <h2 className="mt-4 font-display text-[clamp(1.75rem,4vw,3rem)] font-semibold tracking-[-0.03em]">
            Verify before you install.
          </h2>
          <p className="mt-6 font-lead-airy mx-auto max-w-xl">
            Every binary is signed and published from GitHub Releases. Compare the SHA-256 hash
            before installing. Never install if the hash mismatches.
          </p>
        </div>

        {/* Code snippets */}
        <div className="grid md:grid-cols-2 gap-6">
          <div className="overflow-hidden rounded-2xl border border-white/[0.10] bg-[#0a0a0a]">
            <div className="border-b border-white/[0.06] bg-white/[0.02] px-5 py-3 font-mono text-[11px] text-white/30">
              macOS / Linux
            </div>
            <pre className="p-6 font-mono text-[13px] text-white/55 overflow-x-auto">
{`# Download the archive
curl -LO condura.app/dl/mac

# Verify the hash
shasum -a 256 condura.dmg

# Compare with GitHub release page
# https://github.com/sahajpatel123/
#   conduraapp/releases/tag/v0.1.0`}
            </pre>
          </div>

          <div className="overflow-hidden rounded-2xl border border-white/[0.10] bg-[#0a0a0a]">
            <div className="border-b border-white/[0.06] bg-white/[0.02] px-5 py-3 font-mono text-[11px] text-white/30">
              Windows (PowerShell)
            </div>
            <pre className="p-6 font-mono text-[13px] text-white/55 overflow-x-auto">{`# Download the installer
Invoke-WebRequest condura.app/dl/win

# Verify the hash
Get-FileHash condura-setup.exe -Algorithm SHA256

# Compare with GitHub release page
# https://github.com/sahajpatel123/
#   conduraapp/releases/tag/v0.1.0`}</pre>
          </div>
        </div>

        {/* Trust badges */}
        <div className="mt-12 flex flex-wrap items-center justify-center gap-8">
          {[
            { label: "Code signed", sub: "Apple Developer ID / Authenticode" },
            { label: "Notarized", sub: "Apple notarytool" },
            { label: "Ed25519 verified", sub: "Auto-update signatures" },
          ].map((badge) => (
            <div key={badge.label} className="flex flex-col items-center">
              <div className="flex h-10 w-10 items-center justify-center rounded-full border border-white/15 bg-white/[0.04]">
                <span className="text-white/50">✓</span>
              </div>
              <span className="mt-3 font-body-mature text-[14px] font-semibold text-white">{badge.label}</span>
              <span className="mt-1 font-mono text-[10px] text-white/30">{badge.sub}</span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   7. FAQ — Accordion
   ════════════════════════════════════════════════════════════ */

function FAQSection() {
  return (
    <section className="relative w-full py-[140px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-3xl">
        <div className="mb-12 text-center">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">FAQ</span>
          <h2 className="mt-4 font-display text-[clamp(1.75rem,4vw,3rem)] font-semibold tracking-[-0.03em]">
            Questions, answered.
          </h2>
        </div>

        <BouncyAccordion items={FAQ_ITEMS} defaultOpenId="free" />
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   8. FINAL CTA
   ════════════════════════════════════════════════════════════ */

function FinalCTA() {
  const detected = usePlatform();
  const platformMeta = PLATFORMS.find((p) => p.key === detected)!;

  return (
    <section className="relative w-full py-[200px] px-6 border-t border-white/[0.08] overflow-hidden">
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ scale: [1, 1.15, 1], opacity: [0.05, 0.1, 0.05] }}
          transition={{ duration: 5, repeat: Infinity }}
          className="w-[600px] h-[300px] rounded-full bg-white blur-[150px]"
        />
      </div>

      <div className="relative z-10 mx-auto max-w-3xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1, ease: EASE_OUT }}
        >
          <h2 className="font-display text-[clamp(2rem,6vw,4rem)] font-semibold tracking-[-0.04em] leading-[1.05]">
            Your machine.
            <br />
            <span className="text-white/40">Your conductor.</span>
          </h2>
          <p className="mt-8 font-lead-airy mx-auto max-w-lg">
            Download Condura for {platformMeta.name}. Set your hotkey. Let it orchestrate every AI
            tool you already have.
          </p>
          <div className="mt-12 flex flex-col sm:flex-row items-center justify-center gap-4">
            <a
              href={DOWNLOADS[detected].primary.href}
              className="mature-button inline-flex items-center gap-2 px-8 py-4 font-body-mature text-[15px] font-semibold"
            >
              ↓ Download for {platformMeta.name}
            </a>
            <a
              href="/manifesto"
              className="mature-button-secondary inline-flex items-center gap-2 px-6 py-4 font-body-mature text-[14px]"
            >
              Read the manifesto
            </a>
          </div>

          <p className="mt-8 font-mono text-[11px] text-white/25">
            {SITE.name} v0.1.0 · Free for personal & commercial use ·{" "}
            <a href="/legal" className="underline hover:text-white/50">EULA</a>
          </p>
        </motion.div>
      </div>
    </section>
  );
}
