"use client";

import { useState } from "react";
import { motion } from "motion/react";
import { PLATFORMS } from "@/lib/site";
import type { PlatformKey } from "@/lib/site";

function detectPlatform(): PlatformKey | null {
  if (typeof window === "undefined") return null;
  const ua = navigator.userAgent;
  const platform = navigator.platform;
  const data = (navigator as unknown as { userAgentData?: { platform: string } }).userAgentData;

  if (data?.platform) {
    const p = data.platform.toLowerCase();
    if (p.includes("mac")) return "mac";
    if (p.includes("win")) return "windows";
    if (p.includes("linux")) return "linux";
  }
  if (/Mac/.test(platform) && !/iPhone|iPad/.test(ua)) return "mac";
  if (/Win/.test(platform)) return "windows";
  if (/Linux/.test(platform)) return "linux";
  return null;
}

const PLATFORM_NAMES: Record<PlatformKey, string> = {
  mac: "macOS",
  windows: "Windows",
  linux: "Linux",
};

export default function CTASection() {
  const detected = typeof window !== "undefined" ? detectPlatform() : null;
  const primary = detected ?? "mac";

  const [state, setState] = useState<"idle" | "loading" | "done">("idle");

  const handleClick = () => {
    setState("loading");
    setTimeout(() => {
      setState("done");
      setTimeout(() => {
        window.location.href = "/download";
      }, 600);
    }, 800);
  };

  return (
    <section className="relative overflow-hidden bg-[#050505] py-32">
      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />

      <div className="relative z-10 mx-auto max-w-3xl px-6 text-center">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.8, ease: [0.22, 1, 0.36, 1] as [number, number, number, number] }}
        >
          <p className="text-[13px] font-medium uppercase tracking-widest text-white/30 mb-3">
            Ready when you are
          </p>

          <motion.button
            
            onClick={handleClick}
            disabled={state !== "idle"}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.97 }}
            className={`group relative inline-flex items-center rounded-full px-10 py-5 text-[18px] font-semibold text-white transition-all duration-300 ${
              state === "idle"
                ? "bg-[#0066cc] hover:bg-[#0055aa] hover:shadow-[0_0_60px_rgba(0,102,204,0.35)]"
                : "bg-[#0066cc]/70"
            }`}
          >
            {state === "idle" && (
              <>
                Download for {PLATFORM_NAMES[primary]}
                <svg
                  className="ml-2 h-5 w-5 transition-transform duration-200 group-hover:translate-x-0.5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={2}
                >
                  <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12h15m0 0l-6.75-6.75M19.5 12l-6.75 6.75" />
                </svg>
              </>
            )}
            {state === "loading" && (
              <span className="flex items-center gap-2">
                <svg className="h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
                Preparing...
              </span>
            )}
            {state === "done" && (
              <span className="flex items-center gap-2">
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                </svg>
                Here we go
              </span>
            )}
          </motion.button>

          <p className="mt-6 text-[14px] leading-relaxed text-white/30">
            Free forever. No account. No tracking. No cloud. One hotkey away.
          </p>

          {PLATFORMS.length > 0 && (
            <div className="mt-6 flex flex-wrap items-center justify-center gap-2">
              {PLATFORMS.map((p) => (
                <a
                  key={p.key}
                  href={`/download#${p.key}`}
                  className={`rounded-full border px-3 py-1 text-[12px] transition-colors ${
                    p.key === primary
                      ? "border-[#0066cc]/30 bg-[#0066cc]/5 text-[#64c8ff]"
                      : "border-white/[0.06] bg-white/[0.02] text-white/30 hover:border-white/[0.12] hover:text-white/50"
                  }`}
                >
                  {p.name}
                </a>
              ))}
            </div>
          )}
        </motion.div>
      </div>

      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />
    </section>
  );
}
