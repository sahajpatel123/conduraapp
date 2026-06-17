"use client";

import { TOOL_ROSTER } from "@/lib/site";

export default function MarqueeTile() {
  const list = [...TOOL_ROSTER, ...TOOL_ROSTER, ...TOOL_ROSTER];

  return (
    <section id="marquee-tile" className="relative w-full bg-[#000000] py-[140px] px-6 flex flex-col items-center overflow-hidden border-t border-white/[0.08]">
      <div className="mx-auto w-full max-w-5xl text-center">
        <h2 className="font-hero-display text-[#ffffff]">
          Ecosystem. <br /> Orchestrated.
        </h2>
        <p className="mt-6 font-lead-airy text-[#a1a1aa] max-w-3xl mx-auto">
          Condura auto-detects installed coding CLIs and API platforms in your system paths. No config files. It hooks standard outputs, parses return codes, and displays logs natively.
        </p>

        {/* Minimal Marquee Track */}
        <div className="mt-20 relative w-full overflow-hidden py-4">
          <div className="absolute left-0 top-0 bottom-0 w-32 bg-gradient-to-r from-black to-transparent z-10 pointer-events-none" />
          <div className="absolute right-0 top-0 bottom-0 w-32 bg-gradient-to-l from-black to-transparent z-10 pointer-events-none" />

          <div className="flex gap-4 w-max animate-[marquee_25s_linear_infinite] hover:[animation-play-state:paused] py-2">
            {list.map((tool, idx) => (
              <div
                key={`${tool}-${idx}`}
                className="flex items-center gap-3 rounded-xl border border-white/[0.08] bg-white/[0.03] px-6 py-4"
              >
                <span className="flex h-1.5 w-1.5 rounded-full bg-white/40 shadow-[0_0_12px_rgba(255,255,255,0.25)]" />
                <span className="font-body-mature font-medium text-[#ffffff] text-[15px]">
                  {tool}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>

      <style jsx global>{`
        @keyframes marquee {
          0% {
            transform: translateX(0);
          }
          100% {
            transform: translateX(calc(-33.33% - 8px));
          }
        }
      `}</style>
    </section>
  );
}
