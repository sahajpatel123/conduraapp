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
    <main id="main" className="mx-auto max-w-3xl px-6 pb-32 pt-24">
      <div className="mb-10">
        {badge && (
          <div className="mb-4">
            <AnimatedBadge tone="violet">{badge}</AnimatedBadge>
          </div>
        )}
        <p className="text-[11px] font-medium uppercase tracking-[0.22em] text-white/30">
          {eyebrow}
        </p>
        <h1 className="mt-3 font-display text-[clamp(2rem,5vw,2.75rem)] font-semibold tracking-[-0.03em] text-white">
          {title}
        </h1>
        {description && (
          <p className="mt-4 max-w-2xl text-[17px] leading-relaxed text-white/45">{description}</p>
        )}
      </div>
      {children}
    </main>
  );
}
