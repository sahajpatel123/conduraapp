"use client";

import { motion, AnimatePresence } from "motion/react";
import { useState, useEffect } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";
import AnimatedText from "@/components/motion/AnimatedText";
import AnimatedNumber from "@/components/motion/AnimatedNumber";
import TiltCard from "@/components/motion/TiltCard";
import MagneticButton from "@/components/motion/MagneticButton";
import StatefulButton from "@/components/motion/StatefulButton";
import SharedLayoutTabs from "@/components/motion/SharedLayoutTabs";
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
   DOWNLOAD — Premium Motion Architecture
   Every section has a job: orient → select → trust → guide →
   verify → answer → commit. Motion carries hierarchy, focus,
   and progression. No decorative fades. No emojis.
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

const FEATURES: { icon: IconKey; title: string; desc: string }[] = [
  { icon: "bolt", title: "Cold start under 500ms", desc: "The daemon lives in your menu bar. Hotkey to overlay in under 100ms." },
  { icon: "lock", title: "Local-first by default", desc: "Memory, skills, and audit log on disk, encrypted. Keys never leave your machine." },
  { icon: "gift", title: "Free forever", desc: "No subscriptions. No premium tier. No nags. A donate button, that's it." },
  { icon: "globe", title: "Twelve plus providers", desc: "Anthropic, OpenAI, Google, xAI, Mistral, DeepSeek, or fully local Ollama." },
  { icon: "shield", title: "Deterministic safety", desc: "A pure-rules gatekeeper. No model can bypass it. Every action is audited." },
  { icon: "monitor", title: "Native on every OS", desc: "macOS, Windows, Linux. Signed, notarized, and auto-updated on each." },
];

const STATS: { value: number; suffix: string; label: string }[] = [
  { value: 500, suffix: "ms", label: "Cold start" },
  { value: 12, suffix: "+", label: "LLM providers" },
  { value: 8, suffix: "", label: "CLI sub-agents" },
  { value: 0, suffix: "", label: "Telemetry calls" },
];

const FAQ_ITEMS = [
  { id: "free", title: "Is it really free?", body: "Yes. No feature gates, no premium tier, no nags. Free for personal and commercial use under the Condura Freeware EULA. There's a donate button in the menu bar if you want to support development." },
  { id: "keys", title: "Do I need an API key?", body: "Not to start. Condura auto-detects local Ollama and any installed CLI tools. You can use it fully offline with local models. Add API keys in Settings only if you want cloud providers." },
  { id: "privacy", title: "What happens to my data?", body: "Everything stays on your machine. Memory, skills, audit logs, embeddings — all local, encrypted at rest. The only network calls are to the LLM provider(s) you configured. No telemetry. No tracking. Ever." },
  { id: "safety", title: "How does it stay safe?", body: "A deterministic Go rules engine — the Gatekeeper — evaluates every action before it executes. No model output reaches a click, keystroke, or shell command without passing through it. Destructive actions require a native dialog with a real human at the keyboard." },
  { id: "uninstall", title: "What if I want to uninstall?", body: "Condura auto-backs-up your data to ~/Documents/condura-backups/ before uninstalling. No cloud account to cancel. No data sitting on someone else's server." },
  { id: "update", title: "How do updates work?", body: "Condura checks GitHub Releases every 6 hours and on launch. Updates are signed with Ed25519 and applied atomically with rollback on failure. You can choose stable, beta, or dev channels." },
];

const TRUST_SIGNALS = [
  "Code signed", "Notarized", "Ed25519 verified", "SHA-256 published",
  "Open source changelog", "No telemetry", "Local-first", "Free forever",
];

export default function DownloadPageView() {
  return (
    <main className="relative w-full bg-black text-white">
      <DownloadHero />
      <StatsBand />
      <PlatformSelector />
      <WhyDownload />
      <InstallTimeline />
      <SystemRequirements />
      <VerificationSection />
      <TrustMarquee />
      <FAQSection />
      <FinalCTA />
    </main>
  );
}

/* ════════════════════════════════════════════════════════════
   1. HERO — Orient. The orb focuses the eye; the CTA commits.
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
    push({ title: "Download started", description: `${current.primary.label} for ${platformMeta.name}.`, tone: "success" });
    window.location.href = current.primary.href;
    return true;
  };

  return (
    <section className="relative min-h-screen flex flex-col items-center justify-center px-6 overflow-hidden">
      <div className="absolute inset-0 bg-grid-dark opacity-20" />

      {/* The Orb — concentric pulsing rings, the focal point */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ scale: [1, 1.12, 1], opacity: [0.04, 0.09, 0.04] }}
          transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
          className="w-[460px] h-[460px] rounded-full bg-white blur-[140px]"
        />
        {[260, 360, 460, 560].map((size, i) => (
          <motion.div
            key={size}
            animate={{ scale: [1, 1.06, 1], opacity: [0.18, 0.06, 0.18] }}
            transition={{ duration: 3, repeat: Infinity, delay: i * 0.4, ease: "easeInOut" }}
            className="absolute rounded-full border border-white/[0.06]"
            style={{ width: size, height: size }}
          />
        ))}
      </div>

      <div className="relative z-10 max-w-3xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 24 }}
          transition={{ duration: 1, ease: EASE_OUT }}
        >
          <div className="mb-8 flex justify-center">
            <AnimatedBadge tone="neutral" pulse>v0.1.0 Open Alpha</AnimatedBadge>
          </div>

          <AnimatedText
            as="h1"
            text="Get Condura."
            className="font-display text-[clamp(2.5rem,7vw,5rem)] font-semibold leading-[1.05] tracking-[-0.04em] block"
          />
          <motion.span
            initial={{ opacity: 0, y: 14 }}
            animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 14 }}
            transition={{ delay: 0.5, duration: 0.8, ease: EASE_OUT }}
            className="block font-display text-[clamp(2.5rem,7vw,5rem)] font-semibold leading-[1.05] tracking-[-0.04em] text-transparent bg-clip-text bg-gradient-to-r from-white via-white to-white/30"
          >
            Free. Forever.
          </motion.span>

          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: mounted ? 1 : 0 }}
            transition={{ delay: 0.7, duration: 0.8 }}
            className="mt-8 mx-auto max-w-xl font-lead-airy"
          >
            One download. Three platforms. No account required. Condura runs entirely on your
            machine — your keys, your models, your data.
          </motion.p>

          {/* Detected platform + primary CTA */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 20 }}
            transition={{ delay: 0.9, duration: 0.7 }}
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
              idleLabel={`Download for ${platformMeta.name}`}
              loadingLabel="Starting…"
              successLabel="Download started"
              onAction={startDownload}
            />

            <div className="flex items-center gap-4 font-mono text-[11px] text-white/30">
              <span className="flex items-center gap-1.5">
                <Icon name="shield" size={13} className="text-white/40" />
                {current.primary.label}
              </span>
              <span className="h-3 w-[1px] bg-white/15" />
              <span className="flex items-center gap-1.5">
                <Icon name="check" size={13} className="text-white/40" />
                Signed & notarized
              </span>
              <span className="h-3 w-[1px] bg-white/15" />
              <a href={RELEASE_TAG} target="_blank" rel="noopener noreferrer" className="hover:text-white/60 transition-colors">
                Release notes →
              </a>
            </div>
          </motion.div>
        </motion.div>
      </div>

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
   2. STATS BAND — Number animation, the proof points
   ════════════════════════════════════════════════════════════ */

function StatsBand() {
  return (
    <section className="relative w-full py-20 px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-4xl grid grid-cols-2 md:grid-cols-4 gap-8">
        {STATS.map((stat, i) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 16 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: i * 0.08, duration: 0.6, ease: EASE_OUT }}
            className="flex flex-col items-center text-center"
          >
            <AnimatedNumber
              value={stat.value}
              suffix={stat.suffix}
              className="font-mono text-[clamp(1.75rem,4vw,2.5rem)] font-medium text-white"
            />
            <span className="mt-2 font-mono text-[10px] uppercase tracking-widest text-white/30">
              {stat.label}
            </span>
          </motion.div>
        ))}
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   3. PLATFORM SELECTOR — Shared-layout tabs glide between
   platforms; the active card lifts and glows.
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
    push({ title: "Download started", description: `${current.primary.label} for ${platformMeta.name}.`, tone: "success" });
    window.location.href = current.primary.href;
  };

  return (
    <section className="relative w-full py-[120px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <SectionHeader eyebrow="Choose your platform" title="Three platforms. One agent." />

        {/* Platform cards with shared active glow */}
        <div className="grid md:grid-cols-3 gap-6">
          {PLATFORMS.map((p) => {
            const isActive = selected === p.key;
            const isDetected = detected === p.key;
            return (
              <TiltCard key={p.key} maxRotate={5} className="h-full">
                <motion.button
                  onClick={() => setSelected(p.key)}
                  whileHover={{ y: -4 }}
                  whileTap={{ scale: 0.98 }}
                  className={`relative h-full w-full overflow-hidden rounded-2xl border p-7 text-left transition-colors ${
                    isActive ? "border-white/25 bg-white/[0.06]" : "border-white/[0.08] bg-white/[0.02] hover:bg-white/[0.04]"
                  }`}
                >
                  {isActive && (
                    <motion.div
                      layoutId="platform-glow"
                      className="absolute inset-0 bg-gradient-to-b from-white/[0.05] to-transparent pointer-events-none"
                    />
                  )}

                  <div className="relative z-10">
                    <div className={`flex h-12 w-12 items-center justify-center rounded-xl border ${isActive ? "border-white/20 bg-white/[0.08]" : "border-white/10 bg-white/[0.03]"}`}>
                      <Icon name={PLATFORM_ICON[p.key]} size={24} className={isActive ? "text-white/85" : "text-white/50"} />
                    </div>

                    <h3 className="mt-5 font-body-mature text-[19px] font-semibold text-white">{p.name}</h3>
                    <p className="mt-1 font-body-mature text-[13px] text-white/40">{p.requirement}</p>

                    {isDetected && (
                      <div className="mt-4 inline-flex items-center gap-1.5 rounded-full border border-green-400/20 bg-green-400/10 px-2.5 py-0.5">
                        <span className="h-1.5 w-1.5 rounded-full bg-green-400/60" />
                        <span className="font-mono text-[10px] text-green-400/70">Detected</span>
                      </div>
                    )}

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

        {/* Download bar — direction-aware content swap */}
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
                className="mature-button rounded-full px-6 py-3 text-[14px] font-semibold inline-flex items-center gap-2"
              >
                <Icon name="key" size={16} />
                Download now
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
   4. WHY DOWNLOAD — Feature grid, staggered, mature icons
   ════════════════════════════════════════════════════════════ */

function WhyDownload() {
  return (
    <section className="relative w-full py-[120px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <SectionHeader eyebrow="Why Condura" title="What you get." />

        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {FEATURES.map((feat, i) => (
            <motion.div
              key={feat.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-40px" }}
              transition={{ delay: i * 0.06, duration: 0.5, ease: EASE_OUT }}
              className="group rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 hover:bg-white/[0.04] hover:border-white/15 transition-colors"
            >
              <div className="flex h-11 w-11 items-center justify-center rounded-xl border border-white/10 bg-white/[0.03] group-hover:border-white/20 group-hover:bg-white/[0.06] transition-colors">
                <Icon name={feat.icon} size={20} className="text-white/60 group-hover:text-white/85 transition-colors" />
              </div>
              <h3 className="mt-5 font-body-mature text-[16px] font-semibold text-white">{feat.title}</h3>
              <p className="mt-2 font-body-mature text-[14px] text-white/45 leading-relaxed">{feat.desc}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   5. INSTALL TIMELINE — Shared-layout tabs for platform;
   direction-aware step reveal.
   ════════════════════════════════════════════════════════════ */

function InstallTimeline() {
  const detected = usePlatform();
  const [selected, setSelected] = useState<PlatformKey>(detected);
  const steps = INSTALL_STEPS[selected];

  const tabs = PLATFORMS.map((p) => ({
    id: p.key,
    label: p.name,
    content: null,
  }));

  return (
    <section className="relative w-full py-[120px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-4xl">
        <SectionHeader eyebrow="Installation" title="Up in four steps." />

        <div className="mb-12 flex justify-center">
          <SharedLayoutTabs
            layoutId="install-platforms"
            value={selected}
            onChange={(id) => setSelected(id as PlatformKey)}
            items={tabs}
          />
        </div>

        <AnimatePresence mode="wait">
          <motion.div
            key={selected}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.3 }}
            className="relative"
          >
            <div className="absolute left-[20px] top-0 bottom-0 w-[1px] bg-gradient-to-b from-white/20 via-white/10 to-transparent" />
            <div className="space-y-8">
              {steps.map((step, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: i * 0.1, duration: 0.5, ease: EASE_OUT }}
                  className="relative flex gap-6"
                >
                  <div className="relative z-10 flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-white/15 bg-black">
                    <span className="font-mono text-[12px] text-white/55">{i + 1}</span>
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
   6. SYSTEM REQUIREMENTS — Calm spec panels
   ════════════════════════════════════════════════════════════ */

function SystemRequirements() {
  return (
    <section className="relative w-full py-[120px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <SectionHeader eyebrow="Requirements" title="Will it run on your machine?" />

        <div className="grid md:grid-cols-3 gap-6">
          {PLATFORMS.map((p, i) => (
            <motion.div
              key={p.key}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.08, duration: 0.5, ease: EASE_OUT }}
              className="mature-panel rounded-2xl p-6"
            >
              <div className="flex items-center gap-3 mb-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-white/10 bg-white/[0.03]">
                  <Icon name={PLATFORM_ICON[p.key]} size={20} className="text-white/60" />
                </div>
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
   7. VERIFICATION — Terminal blocks with copy (ActionSwap)
   ════════════════════════════════════════════════════════════ */

function VerificationSection() {
  return (
    <section className="relative w-full py-[120px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-4xl">
        <SectionHeader
          eyebrow="Verify"
          title="Verify before you install."
          subtitle="Every binary is signed and published from GitHub Releases. Compare the SHA-256 hash before installing. Never install if the hash mismatches."
        />

        <div className="grid md:grid-cols-2 gap-6">
          <CodeBlock title="macOS / Linux" code={`# Download the archive
curl -LO condura.app/api/download/mac

# Verify the hash
shasum -a 256 condura-installer-mac.dmg

# Compare with the GitHub release page
# github.com/sahajpatel123/conduraapp
#   /releases/tag/v0.1.0`} />
          <CodeBlock title="Windows (PowerShell)" code={`# Download the installer
Invoke-WebRequest condura.app/api/download/windows -OutFile condura-setup.exe

# Verify the hash
Get-FileHash condura-setup.exe -Algorithm SHA256

# Compare with the GitHub release page
# github.com/sahajpatel123/conduraapp
#   /releases/tag/v0.1.0`} />
        </div>

        {/* Trust badges */}
        <div className="mt-12 flex flex-wrap items-center justify-center gap-10">
          {[
            { icon: "shield" as IconKey, label: "Code signed", sub: "Apple Developer ID / Authenticode" },
            { icon: "check" as IconKey, label: "Notarized", sub: "Apple notarytool" },
            { icon: "lock" as IconKey, label: "Ed25519 verified", sub: "Auto-update signatures" },
          ].map((badge, i) => (
            <motion.div
              key={badge.label}
              initial={{ opacity: 0, scale: 0.9 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1, type: "spring", stiffness: 200, damping: 18 }}
              className="flex flex-col items-center"
            >
              <div className="flex h-11 w-11 items-center justify-center rounded-full border border-white/15 bg-white/[0.04]">
                <Icon name={badge.icon} size={18} className="text-white/65" />
              </div>
              <span className="mt-3 font-body-mature text-[14px] font-semibold text-white">{badge.label}</span>
              <span className="mt-1 font-mono text-[10px] text-white/30">{badge.sub}</span>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

function CodeBlock({ title, code }: { title: string; code: string }) {
  const [copied, setCopied] = useState(false);
  const { push } = useToast();

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      push({ title: "Copied", description: `${title} snippet is on your clipboard.` });
      window.setTimeout(() => setCopied(false), 1600);
    } catch {
      push({ title: "Copy failed", description: "Browser blocked clipboard access.", tone: "error" });
    }
  };

  return (
    <div className="overflow-hidden rounded-2xl border border-white/[0.10] bg-[#0a0a0a]">
      <div className="flex items-center justify-between border-b border-white/[0.06] bg-white/[0.02] px-5 py-3">
        <span className="flex items-center gap-2 font-mono text-[11px] text-white/40">
          <Icon name="terminal" size={14} className="text-white/40" />
          {title}
        </span>
        <button
          type="button"
          onClick={copy}
          className="flex items-center gap-1.5 rounded-md border border-white/[0.06] bg-white/[0.03] px-2 py-1 font-mono text-[10px] text-white/50 hover:text-white hover:bg-white/[0.06] transition-colors"
        >
          <Icon name="copy" size={12} />
          <ActionSwap primary="Copy" secondary="Copied" active={copied} />
        </button>
      </div>
      <pre className="p-6 font-mono text-[12.5px] leading-relaxed text-white/55 overflow-x-auto">{code}</pre>
    </div>
  );
}

/* ════════════════════════════════════════════════════════════
   8. TRUST MARQUEE — Pause-on-hover signal strip
   ════════════════════════════════════════════════════════════ */

function TrustMarquee() {
  const list = [...TRUST_SIGNALS, ...TRUST_SIGNALS, ...TRUST_SIGNALS];
  return (
    <section className="relative w-full py-16 border-t border-white/[0.08] overflow-hidden">
      <div className="relative w-full overflow-hidden">
        <div className="absolute left-0 top-0 bottom-0 w-24 bg-gradient-to-r from-black to-transparent z-10 pointer-events-none" />
        <div className="absolute right-0 top-0 bottom-0 w-24 bg-gradient-to-l from-black to-transparent z-10 pointer-events-none" />
        <div className="flex gap-4 w-max animate-[trust-marquee_28s_linear_infinite] hover:[animation-play-state:paused] py-2">
          {list.map((signal, idx) => (
            <div
              key={`${signal}-${idx}`}
              className="flex items-center gap-2.5 rounded-full border border-white/[0.08] bg-white/[0.03] px-5 py-2.5"
            >
              <Icon name="check" size={13} className="text-white/40" />
              <span className="font-body-mature text-[13px] font-medium text-white/70">{signal}</span>
            </div>
          ))}
        </div>
      </div>
      <style jsx global>{`
        @keyframes trust-marquee {
          0% { transform: translateX(0); }
          100% { transform: translateX(calc(-33.33% - 8px)); }
        }
      `}</style>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   9. FAQ — Bouncy accordion, single-open
   ════════════════════════════════════════════════════════════ */

function FAQSection() {
  return (
    <section className="relative w-full py-[120px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-3xl">
        <SectionHeader eyebrow="FAQ" title="Questions, answered." />
        <BouncyAccordion items={FAQ_ITEMS} defaultOpenId="free" />
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   10. FINAL CTA — Commit. Glow + magnetic buttons.
   ════════════════════════════════════════════════════════════ */

function FinalCTA() {
  const detected = usePlatform();
  const platformMeta = PLATFORMS.find((p) => p.key === detected)!;

  return (
    <section className="relative w-full py-[180px] px-6 border-t border-white/[0.08] overflow-hidden">
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
            <MagneticButton
              href={DOWNLOADS[detected].primary.href}
              className="mature-button rounded-full px-8 py-4 font-body-mature text-[15px] font-semibold inline-flex items-center gap-2"
            >
              <Icon name="rocket" size={18} />
              Download for {platformMeta.name}
            </MagneticButton>
            <MagneticButton
              href="/manifesto"
              className="mature-button-secondary rounded-full px-6 py-4 font-body-mature text-[14px]"
            >
              Read the manifesto
            </MagneticButton>
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

/* ════════════════════════════════════════════════════════════
   Shared section header
   ════════════════════════════════════════════════════════════ */

function SectionHeader({ eyebrow, title, subtitle }: { eyebrow: string; title: string; subtitle?: string }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.6, ease: EASE_OUT }}
      className="mb-14 text-center"
    >
      <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">{eyebrow}</span>
      <h2 className="mt-4 font-display text-[clamp(1.75rem,4vw,3rem)] font-semibold tracking-[-0.03em]">{title}</h2>
      {subtitle && (
        <p className="mt-6 font-lead-airy mx-auto max-w-xl">{subtitle}</p>
      )}
    </motion.div>
  );
}
