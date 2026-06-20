"use client";

import { useState, useRef, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { usePlatform } from "@/hooks/usePlatform";
import { DOWNLOADS } from "@/lib/downloads";
import { PLATFORMS, type PlatformKey } from "@/lib/site";
import { Icon } from "@/components/motion/Icon";
import { EASE_OUT } from "@/lib/motion";

/**
 * DownloadDropdown — the hero's commit surface.
 *
 * A split button. The primary action auto-detects the visitor's OS and
 * reads "Download for {macOS | Windows | Linux}" in plain text — no
 * platform logos, which read as unpolished at small sizes. Clicking it
 * starts the signed download for that platform immediately.
 *
 * The chevron beside it unfurls a floating glass panel laid out as a
 * single horizontal row of three platform cards — macOS, Windows,
 * Linux side by side. Each card is a direct download. The detected
 * platform's card is tagged "Yours". Outside-click, Escape, or card
 * selection closes the panel.
 *
 * Detection happens client-side via usePlatform (navigator.userAgent).
 * The panel is positioned absolutely so the hero copy never reflows.
 */

const PLATFORM_SUBTITLE: Record<PlatformKey, string> = {
  mac: "macOS 13+",
  windows: "Windows 10+",
  linux: "glibc 2.31+",
};

const PLATFORM_LABEL: Record<PlatformKey, string> = {
  mac: "macOS",
  windows: "Windows",
  linux: "Linux",
};

export default function DownloadDropdown() {
  const detected = usePlatform();
  const [open, setOpen] = useState(false);
  const [mounted, setMounted] = useState(false);
  const wrapRef = useRef<HTMLDivElement>(null);

  // Avoid hydration mismatch — render detected label only after mount
  useEffect(() => { setMounted(true); }, []);

  // Close on outside click
  useEffect(() => {
    if (!open) return;
    const handler = (e: globalThis.MouseEvent) => {
      if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  // Close on Escape
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [open]);

  const triggerDownload = (key: PlatformKey) => {
    const url = DOWNLOADS[key].primary.href;
    window.location.href = url;
    setOpen(false);
  };

  const osLabel = mounted ? PLATFORM_LABEL[detected] : "your platform";

  return (
    <div ref={wrapRef} className="relative w-full sm:w-auto">
      {/* ── Split trigger: primary download + chevron toggle ── */}
      <div className="glass-download relative w-full sm:w-auto p-0 flex items-stretch">
        {/* Primary — direct download for the detected OS, text only */}
        <button
          type="button"
          onClick={() => triggerDownload(detected)}
          className="flex-1 px-7 py-3.5 font-body-mature text-[14px] font-semibold inline-flex items-center justify-center gap-2.5 text-left"
        >
          <span>Download for {osLabel}</span>
        </button>

        {/* Divider */}
        <span className="w-px self-stretch bg-white/15" aria-hidden />

        {/* Chevron — opens the panel for other platforms */}
        <button
          type="button"
          onClick={() => setOpen((p) => !p)}
          aria-haspopup="menu"
          aria-expanded={open}
          aria-label="Choose another platform"
          className="px-4 py-3.5 inline-flex items-center justify-center text-white/70 hover:text-white transition-colors"
        >
          <motion.span
            animate={{ rotate: open ? 180 : 0 }}
            transition={{ duration: 0.3, ease: EASE_OUT }}
            className="inline-flex"
            aria-hidden
          >
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M6 9l6 6 6-6" />
            </svg>
          </motion.span>
        </button>
      </div>

      {/* ── Floating panel — horizontal row of platform cards ── */}
      <AnimatePresence>
        {open && (
          <motion.div
            key="download-panel"
            initial={{ opacity: 0, y: -8, scale: 0.97 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -8, scale: 0.97 }}
            transition={{ duration: 0.22, ease: EASE_OUT }}
            role="menu"
            className="absolute left-0 right-0 sm:left-0 sm:right-auto top-[calc(100%+12px)] z-50 w-full sm:w-[min(620px,calc(100vw-1rem))] origin-top-left"
          >
            {/* Glass shell */}
            <div className="relative overflow-hidden rounded-2xl border border-white/10 bg-[#0a0a0a]/95 backdrop-blur-xl shadow-[0_30px_80px_rgba(0,0,0,0.6),0_0_0_1px_rgba(255,255,255,0.04)]">
              {/* Top hairline glow */}
              <div className="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-white/25 to-transparent" />

              {/* Header */}
              <div className="px-5 pt-5 pb-3 border-b border-white/[0.06]">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-[10px] uppercase tracking-[0.2em] text-white/40">
                    All platforms
                  </span>
                  <span className="h-px flex-1 bg-white/[0.06]" />
                  <span className="font-mono text-[10px] text-white/25">v0.1.0</span>
                </div>
              </div>

              {/* Horizontal row of platform cards */}
              <div className="grid grid-cols-3 gap-2 p-3">
                {PLATFORMS.map((p, i) => {
                  const isDetected = detected === p.key;
                  const download = DOWNLOADS[p.key];
                  return (
                    <motion.button
                      key={p.key}
                      type="button"
                      role="menuitem"
                      onClick={() => triggerDownload(p.key)}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.05 * i + 0.05, duration: 0.3, ease: EASE_OUT }}
                      whileHover={{ y: -3 }}
                      whileTap={{ scale: 0.97 }}
                      className={`group relative flex flex-col items-center justify-center gap-3 rounded-xl px-3 py-5 text-center border transition-colors ${
                        isDetected
                          ? "border-white/20 bg-white/[0.06]"
                          : "border-white/[0.08] bg-white/[0.02] hover:border-white/15 hover:bg-white/[0.04]"
                      }`}
                    >
                      {/* Detected marker */}
                      {isDetected && (
                        <span className="absolute top-2.5 right-2.5 inline-flex items-center gap-1 rounded-full border border-green-400/20 bg-green-400/10 px-2 py-0.5">
                          <span className="h-1 w-1 rounded-full bg-green-400/80" />
                          <span className="font-mono text-[9px] text-green-400/80 uppercase tracking-wider">
                            Yours
                          </span>
                        </span>
                      )}

                      {/* Platform brand mark */}
                      <Icon
                        name={p.key}
                        size={26}
                        className={`transition-colors ${
                          isDetected ? "text-white" : "text-white/70 group-hover:text-white"
                        }`}
                      />

                      {/* Platform name */}
                      <span
                        className={`font-body-mature text-[15px] font-semibold transition-colors ${
                          isDetected ? "text-white" : "text-white/80 group-hover:text-white"
                        }`}
                      >
                        {p.name}
                      </span>

                      {/* OS requirement */}
                      <span className="font-mono text-[10.5px] text-white/35">
                        {PLATFORM_SUBTITLE[p.key]}
                      </span>

                      {/* Divider */}
                      <span className="my-0.5 h-px w-8 bg-white/[0.08]" />

                      {/* Artifact + download glyph */}
                      <span className="flex items-center gap-1.5 font-mono text-[10.5px] text-white/45">
                        <svg
                          width="12"
                          height="12"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          strokeWidth="2"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          className="text-white/50 group-hover:text-white/85 transition-colors"
                        >
                          <path d="M12 3v12" />
                          <path d="M7 10l5 5 5-5" />
                          <path d="M5 21h14" />
                        </svg>
                        {download.primary.label}
                      </span>
                    </motion.button>
                  );
                })}
              </div>

              {/* Footer — signed + release notes */}
              <div className="border-t border-white/[0.06] px-5 py-3.5 flex items-center justify-between">
                <span className="flex items-center gap-1.5 font-mono text-[10px] text-white/35">
                  <Icon name="shield" size={12} className="text-white/45" />
                  Signed &amp; notarized
                </span>
                <a
                  href="https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="font-mono text-[10px] text-white/35 hover:text-white/70 transition-colors"
                >
                  Release notes →
                </a>
              </div>
            </div>

            {/* Tail pointer — anchored under the button, not the panel center */}
            <div className="pointer-events-none absolute -top-[6px] left-[130px] hidden sm:block">
              <div className="h-3 w-3 rotate-45 border-l border-t border-white/10 bg-[#0a0a0a]/95" />
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}