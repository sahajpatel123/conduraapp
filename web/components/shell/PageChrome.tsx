import { type ReactNode } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";

interface PageChromeProps {
  eyebrow: string;
  title: string;
  description?: string;
  children: ReactNode;
  badge?: string;
}

export default function PageChrome({
  eyebrow,
  title,
  description,
  children,
  badge,
}: PageChromeProps) {
  return (
    <div className="relative min-h-screen w-full bg-black">
      {/* Subtle abstract background */}
      <div className="absolute inset-0 bg-grid-dark opacity-20 pointer-events-none" />
      <div className="absolute inset-x-0 top-0 h-[500px] bg-gradient-to-b from-white/[0.03] to-transparent pointer-events-none" />
      
      <main id="main" className="relative z-10 mx-auto max-w-4xl px-6 pb-32 pt-32">
        <div className="mb-16 md:text-center md:flex md:flex-col md:items-center">
          {badge && (
            <div className="mb-6">
              <AnimatedBadge tone="neutral">{badge}</AnimatedBadge>
            </div>
          )}
          <p className="text-[11px] font-medium uppercase tracking-[0.22em] text-white/30">
            {eyebrow}
          </p>
          <h1 className="mt-4 font-display text-[clamp(2.5rem,6vw,4rem)] font-semibold tracking-[-0.04em] text-white leading-[1.1]">
            {title}
          </h1>
          {description && (
            <p className="mt-6 max-w-2xl text-[18px] leading-relaxed text-white/45">{description}</p>
          )}
        </div>
        {children}
      </main>
    </div>
  );
}
