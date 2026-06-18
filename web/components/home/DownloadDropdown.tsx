"use client";

import { useState, useRef, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { usePlatform } from "@/hooks/usePlatform";
import { DOWNLOADS } from "@/lib/downloads";
import { PLATFORMS, type PlatformKey } from "@/lib/site";
import { Icon, type IconKey } from "@/components/motion/Icon";
import { EASE_OUT } from "@/lib/motion";

/**
 * DownloadDropdown — the hero's commit surface.
 *
 * A split button. The primary action auto-detects the visitor's OS and
 * reads "Download for {macOS | Windows | Linux}" — clicking it starts
 * the signed download for that platform immediately, no menu required.
 * The chevron beside it unfurls a floating glass panel of all three
 * platforms for anyone on a different machine than the one they're
 * downloading for.
 *
 * Detection happens client-side via usePlatform (navigator.userAgent).
 * The panel is positioned absolutely so the hero copy never reflows.
 * Outside-click, Escape, or lane selection closes it.
 */

const PLATFORM_ICON: Record<PlatformKey, IconKey> = {
  mac: "mac",
  windows: "windows",
  linux: "linux",
};

const PLATFORM_SUBTITLE: Record<PlatformKey, string> = {
  mac: "macOS 13+ · Apple silicon & Intel",
  windows: "Windows 10+ · x64",
  linux: "glibc 2.31+ · x64",
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
        {/* Primary — direct download for the detected OS */}
        <button
          type="button"
          onClick={() => triggerDownload(detected)}
          className="flex-1 px-7 py-3.5 font-body-mature text-[14px] font-semibold inline-flex items-center justify-center gap-2.5 text-left"
        >
          <Icon name={PLATFORM_ICON[detected]} size={16} className="shrink-0 text-white/85" />
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

      {/* ── Floating panel ── */}
      <AnimatePresence>
        {open && (
          <motion.div
            key="download-panel"
            initial={{ opacity: 0, y: -8, scale: 0.97 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -8, scale: 0.97 }}
            transition={{ duration: 0.22, ease: EASE_OUT }}
            role="menu"
            className="absolute left-0 right-0 sm:left-1/2 sm:right-auto sm:-translate-x-1/2 top-[calc(100%+12px)] z-50 w-full sm:w-[380px] origin-top"
          >
            {/* Glass shell */}
            <div className="relative overflow-hidden rounded-2xl border border-white/10 bg-[#0a0a0a]/95 backdrop-blur-xl shadow-[0_30px_80px_rgba(0,0,0,0.6),0_0_0_1px_rgba(255,255,255,0.04)]">
              {/* Top hairline glow */}
              <div className="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-white/25 to-transparent" />

              {/* Header */}
              <div className="px-5 pt-5 pb-3 border-b border-white/[0.06]">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-[10px] uppercase tracking-[0.2em] text-white/40">
                    Other platforms
                  </span>
                  <span className="h-px flex-1 bg-white/[0.06]" />
                  <span className="font-mono text-[10px] text-white/25">v0.1.0</span>
                </div>
              </div>

              {/* Platform lanes */}
              <div className="p-2">
                {PLATFORMS.map((p, i) => {
                  const isDetected = detected === p.key;
                  const download = DOWNLOADS[p.key];
                  return (
                    <motion.button
                      key={p.key}
                      type="button"
                      role="menuitem"
                      onClick={() => triggerDownload(p.key)}
                      initial={{ opacity: 0, x: -12 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: 0.04 * i + 0.05, duration: 0.3, ease: EASE_OUT }}
                      whileHover={{ backgroundColor: "rgba(255,255,255,0.05)" }}
                      whileTap={{ scale: 0.985 }}
                      className={`group relative w-full flex items-center gap-4 rounded-xl px-4 py-3.5 text-left transition-colors ${
                        isDetected ? "bg-white/[0.04]" : ""
                      }`}
                    >
                      {/* Icon tile */}
                      <div
                        className={`relative flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border transition-colors ${
                          isDetected
                            ? "border-white/20 bg-white/[0.08]"
                            : "border-white/10 bg-white/[0.03] group-hover:border-white/20 group-hover:bg-white/[0.06]"
                        }`}
                      >
                        <Icon
                          name={PLATFORM_ICON[p.key]}
                          size={22}
                          className={isDetected ? "text-white/85" : "text-white/55 group-hover:text-white/80"}
                        />
                        {/* Detected pulse */}
                        {isDetected && (
                          <motion.span
                            animate={{ scale: [1, 1.6], opacity: [0.5, 0] }}
                            transition={{ duration: 2.2, repeat: Infinity, ease: "easeOut" }}
                            className="absolute inset-0 rounded-xl border border-green-400/30"
                          />
                        )}
                      </div>

                      {/* Label */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-body-mature text-[14px] font-semibold text-white">
                            {p.name}
                          </span>
                          {isDetected && (
                            <span className="inline-flex items-center gap-1 rounded-full border border-green-400/20 bg-green-400/10 px-2 py-0.5">
                              <span className="h-1 w-1 rounded-full bg-green-400/80" />
                              <span className="font-mono text-[9px] text-green-400/80 uppercase tracking-wider">
                                Yours
                              </span>
                            </span>
                          )}
                        </div>
                        <div className="mt-0.5 flex items-center gap-2 font-mono text-[10.5px] text-white/35">
                          <span className="truncate">{PLATFORM_SUBTITLE[p.key]}</span>
                          <span className="text-white/15">·</span>
                          <span className="text-white/45">{download.primary.label}</span>
                        </div>
                      </div>

                      {/* Download glyph */}
                      <span
                        className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border border-white/[0.08] text-white/40 transition-all group-hover:border-white/20 group-hover:text-white/85 group-hover:bg-white/[0.06]"
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
                          <path d="M12 3v12" />
                          <path d="M7 10l5 5 5-5" />
                          <path d="M5 21h14" />
                        </svg>
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

            {/* Tail pointer */}
            <div className="pointer-events-none absolute -top-[6px] left-1/2 -translate-x-1/2 hidden sm:block">
              <div className="h-3 w-3 rotate-45 border-l border-t border-white/10 bg-[#0a0a0a]/95" />
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}