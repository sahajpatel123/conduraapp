"use client";

import { AnimatePresence, motion, useReducedMotion } from "motion/react";
import { useCallback, useState } from "react";
import ActionSwap from "@/components/motion/ActionSwap";
import BouncyAccordion from "@/components/motion/BouncyAccordion";
import { Icon, type IconKey } from "@/components/motion/Icon";
import OverlayPreview from "@/components/home/OverlayPreview";
import { useToast } from "@/context/ToastContext";
import { useIsland } from "@/context/IslandContext";
import { usePlatform } from "@/hooks/usePlatform";
import { DOWNLOADS, RELEASE_TAG } from "@/lib/downloads";
import { PLATFORMS, SITE, type PlatformKey } from "@/lib/site";
import { EASE_OUT } from "@/lib/motion";

const PLATFORM_ICON: Record<PlatformKey, IconKey> = {
  mac: "mac",
  windows: "windows",
  linux: "linux",
};

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

const FAQ_ITEMS = [
  { id: "free", title: "Is Condura really free?", body: "Yes. Condura v0.1.0 is free for personal and commercial use under the Condura Freeware EULA. There is no account requirement or paid feature gate." },
  { id: "keys", title: "Do I need an API key?", body: "Not for local models. Condura can detect a local Ollama installation and supported local tools. Add provider credentials only when you choose to use a cloud model." },
  { id: "privacy", title: "Where does my data live?", body: "Condura is designed around local storage and user-controlled providers. Requests leave your machine only when the model or service you configure requires a network call." },
  { id: "safety", title: "How are computer actions controlled?", body: "A deterministic Gatekeeper evaluates requested actions before execution. Sensitive or destructive actions require explicit approval rather than relying on model judgment alone." },
  { id: "uninstall", title: "Can I remove it cleanly?", body: "Yes. You can uninstall the application and remove its local data without cancelling a cloud account or subscription." },
  { id: "update", title: "How do updates work?", body: "Condura checks published releases and supports signed update packages with rollback protection. Release details and checksums are available on GitHub." },
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
    <main className="relative w-full overflow-hidden bg-[#070707] text-white">
      <Hero
        selected={selected}
        detected={detected}
        downloading={downloading}
        onSelect={setSelected}
        onDownload={triggerDownload}
      />
      <BuildChooser selected={selected} detected={detected} onSelect={setSelected} onDownload={triggerDownload} />
      <SetupSection selected={selected} />
      <FAQSection />
      <FinalCTA selected={selected} onDownload={triggerDownload} />
    </main>
  );
}

function Hero({
  selected,
  detected,
  downloading,
  onSelect,
  onDownload,
}: {
  selected: PlatformKey;
  detected: PlatformKey;
  downloading: boolean;
  onSelect: (key: PlatformKey) => void;
  onDownload: (href: string, label: string) => void;
}) {
  const reduceMotion = useReducedMotion();
  const platform = PLATFORMS.find((item) => item.key === selected)!;
  const current = DOWNLOADS[selected];

  return (
    <section className="relative border-b border-white/[0.08] px-5 pb-16 pt-28 sm:px-8 sm:pt-32 lg:min-h-[860px] lg:px-10 lg:pb-24 lg:pt-36">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-px bg-[#D97757]/50" />
      <div className="mx-auto grid w-full max-w-[1240px] items-center gap-14 lg:grid-cols-[0.82fr_1.18fr] lg:gap-16">
        <motion.div
          initial={{ opacity: 0, y: reduceMotion ? 0 : 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7, ease: EASE_OUT }}
          className="relative z-10"
        >
          <div className="mb-7 flex items-center gap-3 font-mono text-[10px] uppercase tracking-[0.18em] text-white/45">
            <span className="h-px w-8 bg-[#D97757]" />
            Release 0.1.0 · Open alpha
          </div>

          <h1 className="max-w-[620px] font-display text-[clamp(3rem,6.4vw,5.8rem)] font-medium leading-[0.94] tracking-[-0.045em]">
            Condura,
            <br />
            <span className="text-white/42">on your machine.</span>
          </h1>

          <p className="mt-7 max-w-[520px] text-[16px] leading-7 text-white/55 sm:text-[18px]">
            A local-first intelligence layer for your OS. No account, no subscription, and no new workflow to learn.
          </p>

          <div className="mt-9" aria-label="Choose operating system">
            <div className="inline-flex max-w-full rounded-lg border border-white/[0.10] bg-white/[0.035] p-1">
              {PLATFORMS.map((item) => {
                const active = item.key === selected;
                return (
                  <button
                    key={item.key}
                    type="button"
                    onClick={() => onSelect(item.key)}
                    aria-pressed={active}
                    className={`relative flex min-h-11 items-center gap-2 rounded-md px-3.5 text-[13px] font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#D97757]/70 sm:px-4 ${active ? "text-white" : "text-white/42 hover:text-white/75"}`}
                  >
                    {active && (
                      <motion.span
                        layoutId="download-platform-selector"
                        className="absolute inset-0 rounded-md border border-white/[0.10] bg-white/[0.08] shadow-[inset_0_1px_0_rgba(255,255,255,0.08)]"
                        transition={{ type: "spring", stiffness: 420, damping: 34 }}
                      />
                    )}
                    <Icon name={PLATFORM_ICON[item.key]} size={15} className="relative" />
                    <span className="relative">{item.name}</span>
                  </button>
                );
              })}
            </div>
          </div>

          <div className="mt-5 flex flex-col gap-3 sm:flex-row sm:items-center">
            <button
              type="button"
              onClick={() => onDownload(current.primary.href, current.primary.label)}
              className="group inline-flex min-h-14 items-center justify-center gap-3 rounded-lg bg-[#D97757] px-6 text-[14px] font-semibold text-[#190c08] shadow-[0_16px_44px_rgba(217,119,87,0.18)] transition-[transform,background-color] duration-200 hover:-translate-y-0.5 hover:bg-[#e18465] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-offset-2 focus-visible:ring-offset-black active:translate-y-0"
            >
              <motion.span
                animate={downloading && !reduceMotion ? { y: [0, 2, 0] } : { y: 0 }}
                transition={{ duration: 0.55, repeat: downloading ? Infinity : 0 }}
              >
                <Icon name="download" size={18} />
              </motion.span>
              {downloading ? "Preparing download" : `Download for ${platform.name}`}
            </button>
            <a
              href={RELEASE_TAG}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex min-h-14 items-center justify-center gap-2 rounded-lg border border-white/[0.11] px-5 text-[13px] font-medium text-white/62 transition-colors hover:border-white/20 hover:bg-white/[0.04] hover:text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/60"
            >
              Release notes
              <Icon name="arrowRight" size={14} />
            </a>
          </div>

          <div className="mt-6 grid max-w-[520px] grid-cols-3 border-y border-white/[0.08] py-4">
            {[
              ["Account", "Not required"],
              ["License", "Free"],
              ["Build", current.primary.label],
            ].map(([label, value]) => (
              <div key={label} className="border-l border-white/[0.08] px-3 first:border-l-0 first:pl-0">
                <div className="font-mono text-[9px] uppercase tracking-[0.14em] text-white/28">{label}</div>
                <div className="mt-1 truncate text-[12px] text-white/62">{value}</div>
              </div>
            ))}
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, x: reduceMotion ? 0 : 24 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.12, duration: 0.85, ease: EASE_OUT }}
          className="relative min-w-0"
        >
          <DesktopPreview active />
          <div className="mt-4 flex items-center justify-between px-1 text-[10px] text-white/30">
            <span className="font-mono uppercase tracking-[0.15em]">Product preview</span>
            <span>Press a hotkey. Condura appears.</span>
          </div>
        </motion.div>
      </div>

      <div className="mx-auto mt-14 flex w-full max-w-[1240px] flex-wrap items-center gap-x-7 gap-y-3 border-t border-white/[0.08] pt-6 text-[12px] text-white/42 lg:mt-20">
        <span className="flex items-center gap-2"><Icon name="shield" size={14} /> Permission-gated actions</span>
        <span className="flex items-center gap-2"><Icon name="lock" size={14} /> Local-first storage</span>
        <span className="flex items-center gap-2"><Icon name="check" size={14} /> {selected === detected ? `${platform.name} detected` : `${platform.name} selected`}</span>
      </div>
    </section>
  );
}

function DesktopPreview({ active }: { active: boolean }) {
  return (
    <div className="relative aspect-[16/10] min-h-[410px] overflow-hidden rounded-lg border border-white/[0.14] bg-[#111] shadow-[0_36px_100px_rgba(0,0,0,0.55),0_0_0_1px_rgba(255,255,255,0.03)] sm:min-h-0">
      <div
        className="absolute inset-0 bg-cover bg-center"
        style={{ backgroundImage: "url('/images/macbook-desktop-background.png')" }}
      />
      <div className="absolute inset-0 bg-black/20" />
      <div className="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-black/45 to-transparent" />
      <div className="absolute inset-x-[9%] bottom-[8%] top-[10%] flex items-center justify-center">
        <div className="w-full max-w-[470px] origin-center scale-[0.72] sm:scale-[0.82] lg:scale-[0.72] xl:scale-[0.86]">
          <OverlayPreview active={active} />
        </div>
      </div>
      <div className="pointer-events-none absolute inset-0 shadow-[inset_0_1px_0_rgba(255,255,255,0.15),inset_0_-60px_100px_rgba(0,0,0,0.15)]" />
    </div>
  );
}

function BuildChooser({
  selected,
  detected,
  onSelect,
  onDownload,
}: {
  selected: PlatformKey;
  detected: PlatformKey;
  onSelect: (key: PlatformKey) => void;
  onDownload: (href: string, label: string) => void;
}) {
  return (
    <section className="px-5 py-20 sm:px-8 sm:py-28 lg:px-10">
      <div className="mx-auto max-w-[1120px]">
        <SectionHeading eyebrow="Available builds" title="Choose the right package." description="The recommended installer is first. Smaller runtime and portable builds remain available for advanced setups." />

        <div className="mt-12 divide-y divide-white/[0.08] border-y border-white/[0.10]">
          {PLATFORMS.map((platform) => {
            const active = platform.key === selected;
            const isDetected = platform.key === detected;
            const downloads = DOWNLOADS[platform.key];

            return (
              <motion.div
                key={platform.key}
                layout
                className={`relative grid gap-5 py-6 transition-colors sm:grid-cols-[1.15fr_1fr_auto] sm:items-center sm:px-5 ${active ? "bg-white/[0.035]" : "hover:bg-white/[0.018]"}`}
              >
                {active && <motion.span layoutId="active-build-rail" className="absolute inset-y-0 left-0 w-0.5 bg-[#D97757]" />}
                <button
                  type="button"
                  onClick={() => onSelect(platform.key)}
                  className="flex min-w-0 items-center gap-4 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#D97757]/70"
                >
                  <span className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-md border ${active ? "border-[#D97757]/45 bg-[#D97757]/10 text-[#e39478]" : "border-white/[0.09] text-white/42"}`}>
                    <Icon name={PLATFORM_ICON[platform.key]} size={21} />
                  </span>
                  <span className="min-w-0">
                    <span className="flex flex-wrap items-center gap-2">
                      <span className="text-[16px] font-semibold text-white">{platform.name}</span>
                      {isDetected && <span className="rounded-sm bg-[#D97757]/12 px-2 py-0.5 font-mono text-[9px] uppercase tracking-[0.12em] text-[#e39478]">Detected</span>}
                    </span>
                    <span className="mt-1 block text-[12px] text-white/35">{platform.requirement}</span>
                  </span>
                </button>

                <div className="grid grid-cols-2 gap-5 text-[12px]">
                  <div>
                    <div className="font-mono text-[9px] uppercase tracking-[0.14em] text-white/25">Recommended</div>
                    <div className="mt-1 text-white/58">{downloads.primary.label}</div>
                  </div>
                  <div>
                    <div className="font-mono text-[9px] uppercase tracking-[0.14em] text-white/25">Alternative</div>
                    <a className="mt-1 inline-flex items-center gap-1.5 text-white/45 hover:text-white" href={downloads.secondary.href} download>{downloads.secondary.label}<Icon name="arrowRight" size={11} /></a>
                  </div>
                </div>

                <button
                  type="button"
                  onClick={() => {
                    onSelect(platform.key);
                    onDownload(downloads.primary.href, downloads.primary.label);
                  }}
                  aria-label={`Download ${downloads.primary.label} for ${platform.name}`}
                  className="inline-flex min-h-11 items-center justify-center gap-2 rounded-md border border-white/[0.12] bg-white/[0.04] px-4 text-[12px] font-semibold text-white/72 transition-colors hover:border-[#D97757]/45 hover:bg-[#D97757]/10 hover:text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#D97757]/70 sm:w-32"
                >
                  <Icon name="download" size={14} />
                  Download
                </button>
              </motion.div>
            );
          })}
        </div>
      </div>
    </section>
  );
}

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
    <section className="border-y border-white/[0.08] bg-[#0a0a0a] px-5 py-20 sm:px-8 sm:py-28 lg:px-10">
      <div className="mx-auto grid max-w-[1120px] gap-16 lg:grid-cols-[1.1fr_0.9fr] lg:gap-24">
        <div>
          <SectionHeading eyebrow={`${platform.name} setup`} title="Installed in four clear steps." description="Condura asks for access only when a feature needs it. Permissions remain visible and reversible in your system settings." align="left" />

          <AnimatePresence mode="wait">
            <motion.ol
              key={selected}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.25 }}
              className="mt-10 divide-y divide-white/[0.08] border-t border-white/[0.10]"
            >
              {steps.map((step, index) => (
                <li key={step.title} className="grid grid-cols-[34px_1fr] gap-4 py-5">
                  <span className="font-mono text-[11px] text-[#D97757]">0{index + 1}</span>
                  <div>
                    <h3 className="text-[14px] font-semibold text-white/88">{step.title}</h3>
                    <p className="mt-1.5 text-[13px] leading-6 text-white/42">{step.desc}</p>
                  </div>
                </li>
              ))}
            </motion.ol>
          </AnimatePresence>
        </div>

        <div className="lg:pt-8">
          <div className="font-mono text-[10px] uppercase tracking-[0.16em] text-white/32">Verify the download</div>
          <h3 className="mt-3 max-w-sm font-display text-[28px] font-medium leading-tight tracking-[-0.025em]">Trust the artifact, not the page.</h3>
          <p className="mt-4 max-w-md text-[13px] leading-6 text-white/42">Run the checksum locally, then compare it with the value published alongside the release.</p>

          <div className="mt-8 overflow-hidden rounded-lg border border-white/[0.11] bg-[#050505]">
            <div className="flex items-center justify-between border-b border-white/[0.07] px-4 py-3">
              <span className="flex items-center gap-2 font-mono text-[10px] text-white/38"><Icon name="terminal" size={13} />{platform.name}</span>
              <button type="button" onClick={copy} className="flex min-h-8 items-center gap-1.5 rounded-md px-2 text-[10px] text-white/42 transition-colors hover:bg-white/[0.05] hover:text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/50">
                <Icon name="copy" size={12} />
                <ActionSwap primary="Copy" secondary="Copied" active={copied} />
              </button>
            </div>
            <AnimatePresence mode="wait">
              <motion.code key={selected} initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="block overflow-x-auto p-5 font-mono text-[11px] leading-6 text-white/58">{command}</motion.code>
            </AnimatePresence>
          </div>

          <a href={RELEASE_TAG} target="_blank" rel="noopener noreferrer" className="mt-5 inline-flex items-center gap-2 text-[12px] text-white/42 transition-colors hover:text-white">
            Open checksums on GitHub <Icon name="arrowRight" size={13} />
          </a>

          <div className="mt-9 grid grid-cols-3 gap-3 border-t border-white/[0.08] pt-6">
            {[
              ["shield" as IconKey, "Gatekeeper"],
              ["lock" as IconKey, "Local data"],
              ["check" as IconKey, "Signed updates"],
            ].map(([icon, label]) => (
              <div key={label} className="text-[10px] text-white/36">
                <Icon name={icon} size={15} className="mb-2 text-[#D97757]" />
                {label}
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

function FAQSection() {
  return (
    <section className="px-5 py-20 sm:px-8 sm:py-28 lg:px-10">
      <div className="mx-auto grid max-w-[1120px] gap-12 lg:grid-cols-[0.72fr_1.28fr] lg:gap-24">
        <SectionHeading eyebrow="Before you install" title="Straight answers." description="The details that matter before software receives access to your machine." align="left" />
        <div className="min-w-0"><BouncyAccordion items={FAQ_ITEMS} defaultOpenId="free" /></div>
      </div>
    </section>
  );
}

function FinalCTA({ selected, onDownload }: { selected: PlatformKey; onDownload: (href: string, label: string) => void }) {
  const platform = PLATFORMS.find((item) => item.key === selected)!;
  const current = DOWNLOADS[selected];

  return (
    <section className="border-t border-white/[0.08] px-5 py-20 sm:px-8 sm:py-28 lg:px-10">
      <div className="mx-auto grid max-w-[1120px] items-end gap-10 border-l-2 border-[#D97757] py-3 pl-6 sm:pl-9 lg:grid-cols-[1fr_auto]">
        <div>
          <div className="font-mono text-[10px] uppercase tracking-[0.17em] text-[#D97757]">Ready when you are</div>
          <h2 className="mt-4 max-w-2xl font-display text-[clamp(2.2rem,5vw,4.5rem)] font-medium leading-[0.98] tracking-[-0.04em]">Put your AI tools behind one hotkey.</h2>
          <p className="mt-5 max-w-xl text-[14px] leading-6 text-white/42">Condura v0.1.0 for {platform.name}. Free for personal and commercial use.</p>
        </div>
        <div className="flex flex-col gap-3 sm:flex-row lg:flex-col">
          <button type="button" onClick={() => onDownload(current.primary.href, current.primary.label)} className="inline-flex min-h-13 items-center justify-center gap-2 rounded-lg bg-white px-6 text-[13px] font-semibold text-black transition-transform hover:-translate-y-0.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#D97757] focus-visible:ring-offset-2 focus-visible:ring-offset-black active:translate-y-0">
            <Icon name="download" size={16} /> Download for {platform.name}
          </button>
          <a href="/legal" className="inline-flex min-h-11 items-center justify-center text-[11px] text-white/35 transition-colors hover:text-white">Review the EULA</a>
        </div>
      </div>
      <div className="mx-auto mt-12 max-w-[1120px] border-t border-white/[0.07] pt-5 font-mono text-[9px] uppercase tracking-[0.14em] text-white/20">{SITE.name} · Release 0.1.0 · Local-first desktop intelligence</div>
    </section>
  );
}

function SectionHeading({
  eyebrow,
  title,
  description,
  align = "center",
}: {
  eyebrow: string;
  title: string;
  description: string;
  align?: "left" | "center";
}) {
  return (
    <div className={align === "center" ? "mx-auto max-w-2xl text-center" : "max-w-xl"}>
      <div className="font-mono text-[10px] uppercase tracking-[0.17em] text-[#D97757]">{eyebrow}</div>
      <h2 className="mt-3 font-display text-[clamp(2rem,4vw,3.3rem)] font-medium leading-[1.02] tracking-[-0.035em]">{title}</h2>
      <p className={`mt-5 text-[14px] leading-6 text-white/42 ${align === "center" ? "mx-auto max-w-xl" : "max-w-lg"}`}>{description}</p>
    </div>
  );
}
