"use client";

/*
  The roster in motion — every tool Synaptic conducts, on an endless
  strip. Pauses on hover; stands still under reduced motion.
*/
import { ORCHESTRA } from "@/lib/site";

export function ToolMarquee() {
  const row = [...ORCHESTRA, ...ORCHESTRA];
  return (
    <div
      aria-label={`The orchestra: ${ORCHESTRA.join(", ")}`}
      className="relative overflow-hidden border-y border-line py-5"
    >
      <div className="marquee gap-12 motion-reduce:flex-wrap motion-reduce:justify-center" aria-hidden>
        {row.map((tool, i) => (
          <span key={`${tool}-${i}`} className="flex shrink-0 items-center gap-12">
            <span className="font-display text-2xl font-semibold whitespace-nowrap text-ivory-dim transition-colors duration-200 hover:text-brass">
              {tool}
            </span>
            <span className="size-1.5 rotate-45 bg-brass/60" />
          </span>
        ))}
      </div>
    </div>
  );
}
