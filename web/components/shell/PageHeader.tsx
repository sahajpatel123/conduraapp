"use client";

import { type ReactNode, useEffect, useRef } from "react";
import GlobalNav from "@/components/shell/GlobalNav";
import { motion, useInView } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * PageHeader — the shared masthead for all sub-pages.
 * A generative thread rises behind the eyebrow + title, and a
 * self-drawing synapse divider sits under the description.
 */
export default function PageHeader({
  eyebrow,
  title,
  titleAccent,
  description,
  children,
}: {
  eyebrow: string;
  title: ReactNode;
  titleAccent?: ReactNode;
  description?: ReactNode;
  children?: ReactNode;
}) {
  const ref = useRef<HTMLDivElement | null>(null);
  const inView = useInView(ref, { once: true, margin: "-10%" });

  return (
    <div className="relative min-h-screen w-full">
      <GlobalNav />
      <main
        id="main"
        className="relative z-10 mx-auto max-w-[1100px] px-6 pb-32 pt-36 sm:pt-44"
      >
        <div ref={ref} className="mb-20">
          <motion.p
            initial={{ opacity: 0, y: 8 }}
            animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 8 }}
            transition={{ duration: 0.7, ease: EASE_OUT }}
            className="text-eyebrow"
          >
            — {eyebrow}
          </motion.p>

          <h1 className="text-hero mt-6 text-[var(--color-ink)] text-balance">
            <span className="block overflow-hidden">
              <motion.span
                className="block"
                initial={{ y: "110%" }}
                animate={inView ? { y: "0%" } : { y: "110%" }}
                transition={{ duration: 0.9, ease: EASE_OUT, delay: 0.1 }}
              >
                {title}
              </motion.span>
            </span>
            {titleAccent && (
              <span className="block overflow-hidden">
                <motion.span
                  className="block italic text-[var(--color-synapse)]"
                  initial={{ y: "110%" }}
                  animate={inView ? { y: "0%" } : { y: "110%" }}
                  transition={{ duration: 0.9, ease: EASE_OUT, delay: 0.26 }}
                >
                  {titleAccent}
                </motion.span>
              </span>
            )}
          </h1>

          {description && (
            <motion.p
              initial={{ opacity: 0, y: 14, filter: "blur(6px)" }}
              animate={
                inView
                  ? { opacity: 1, y: 0, filter: "blur(0px)" }
                  : { opacity: 0, y: 14, filter: "blur(6px)" }
              }
              transition={{ duration: 0.9, ease: EASE_OUT, delay: 0.5 }}
              className="text-lead mt-7 max-w-[58ch] text-[var(--color-ink-soft)] text-pretty"
            >
              {description}
            </motion.p>
          )}

          {/* self-drawing thread divider */}
          <ThreadDivider active={inView} />
        </div>

        {children}
      </main>
    </div>
  );
}

function ThreadDivider({ active }: { active: boolean }) {
  const pathRef = useRef<SVGPathElement | null>(null);
  useEffect(() => {
    if (!active) return;
    const p = pathRef.current;
    if (!p) return;
    const len = p.getTotalLength();
    p.style.strokeDasharray = `${len}`;
    p.style.strokeDashoffset = `${len}`;
    p.getBoundingClientRect();
    p.style.transition = "stroke-dashoffset 1.8s cubic-bezier(0.22,1,0.36,1)";
    p.style.strokeDashoffset = "0";
  }, [active]);
  return (
    <svg className="mt-10 h-8 w-full max-w-[520px]" viewBox="0 0 520 32" aria-hidden>
      <path
        ref={pathRef}
        className="synapse-thread"
        d="M 4 16 C 90 4, 180 28, 260 16 C 340 4, 430 28, 516 16"
        strokeWidth="1.4"
      />
      <circle cx="260" cy="16" r="3" className="synapse-node" />
    </svg>
  );
}
