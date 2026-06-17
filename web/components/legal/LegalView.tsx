"use client";

import { motion } from "motion/react";
import { useEffect, useMemo, useRef, useState } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   LEGAL — End-User License Agreement
   The EULA rendered as a readable document with a sticky
   table-of-contents that tracks scroll position. Each h2
   becomes an anchorable section.
   ──────────────────────────────────────────────────────────── */

interface LegalViewProps {
  html: string | null;
}

export default function LegalView({ html }: LegalViewProps) {
  const [mounted, setMounted] = useState(false);
  const [activeId, setActiveId] = useState<string>("");
  const contentRef = useRef<HTMLDivElement>(null);

  useEffect(() => { const t = setTimeout(() => setMounted(true), 100); return () => clearTimeout(t); }, []);

  // Parse headings from the rendered HTML to build the TOC.
  const toc = useMemo(() => {
    if (!html) return [] as { id: string; text: string; level: number }[];
    const matches = [...html.matchAll(/<h([1-3])[^>]*>(.*?)<\/h\1>/gi)];
    return matches.map((m, i) => {
      const level = parseInt(m[1], 10);
      const text = m[2].replace(/<[^>]+>/g, "").trim();
      const id = `sec-${i}-${text.toLowerCase().replace(/[^a-z0-9]+/g, "-").slice(0, 40)}`;
      return { id, text, level };
    });
  }, [html]);

  // Inject ids into the html so anchors work.
  const processedHtml = useMemo(() => {
    if (!html) return "";
    let idx = 0;
    return html.replace(/<h([1-3])[^>]*>(.*?)<\/h\1>/gi, (match, level, content) => {
      const text = content.replace(/<[^>]+>/g, "").trim();
      const id = `sec-${idx}-${text.toLowerCase().replace(/[^a-z0-9]+/g, "-").slice(0, 40)}`;
      idx++;
      return `<h${level} id="${id}">${content}</h${level}>`;
    });
  }, [html]);

  // Track active section via IntersectionObserver.
  useEffect(() => {
    if (!toc.length) return;
    const obs = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) setActiveId(entry.target.id);
        }
      },
      { rootMargin: "-20% 0px -70% 0px" }
    );
    toc.forEach(({ id }) => {
      const el = document.getElementById(id);
      if (el) obs.observe(el);
    });
    return () => obs.disconnect();
  }, [toc]);

  if (!html) {
    return (
      <div className="mx-auto max-w-2xl py-24 text-center">
        <p className="font-body-mature text-white/45">
          The license agreement is not available right now. Contact{" "}
          <a className="underline hover:text-white" href="mailto:legal@condura.app">
            legal@condura.app
          </a>
          .
        </p>
      </div>
    );
  }

  return (
    <main className="relative w-full bg-black text-white overflow-hidden">
      {/* Hero */}
      <section className="relative min-h-[60vh] flex flex-col items-center justify-center px-6 overflow-hidden">
        <div className="absolute inset-0 bg-grid-dark opacity-20" />
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <motion.div
            animate={{ rotate: 360 }}
            transition={{ duration: 80, repeat: Infinity, ease: "linear" }}
            className="w-[450px] h-[450px] rounded-full border border-white/[0.06]"
          />
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 20 }}
          transition={{ duration: 1, ease: EASE_OUT }}
          className="relative z-10 max-w-3xl text-center"
        >
          <div className="mb-8 flex justify-center">
            <AnimatedBadge tone="neutral">Legal</AnimatedBadge>
          </div>
          <h1 className="font-display text-[clamp(2.5rem,6vw,4.5rem)] font-semibold leading-[1.05] tracking-[-0.04em]">
            End-User License Agreement
          </h1>
          <p className="mt-8 mx-auto max-w-xl font-lead-airy">
            The terms governing use of the Condura application. Free for personal and commercial
            use. No redistribution. Revocable for abuse.
          </p>
        </motion.div>
      </section>

      {/* Document body with TOC */}
      <section className="relative w-full py-[120px] px-6 border-t border-white/[0.08]">
        <div className="mx-auto max-w-6xl grid lg:grid-cols-[240px_1fr] gap-12">
          {/* Sticky TOC */}
          {toc.length > 1 && (
            <aside className="hidden lg:block">
              <div className="sticky top-8">
                <span className="font-mono text-[11px] uppercase tracking-widest text-white/30 mb-4 block">
                  Contents
                </span>
                <nav className="space-y-1">
                  {toc.map(({ id, text, level }) => (
                    <a
                      key={id}
                      href={`#${id}`}
                      className={`block rounded-md py-1.5 pr-3 font-body-mature text-[13px] transition-colors ${
                        activeId === id
                          ? "text-white"
                          : "text-white/35 hover:text-white/60"
                      } ${level === 2 ? "pl-3" : level === 3 ? "pl-6" : "pl-0"}`}
                    >
                      {activeId === id && (
                        <motion.span
                          layoutId="toc-active"
                          className="absolute left-0 h-5 w-[2px] rounded-full bg-white/50"
                        />
                      )}
                      <span className="relative">{text}</span>
                    </a>
                  ))}
                </nav>
              </div>
            </aside>
          )}

          {/* Document */}
          <motion.div
            ref={contentRef}
            initial={{ opacity: 0, y: 20 }}
            animate={mounted ? { opacity: 1, y: 0 } : {}}
            transition={{ duration: 0.8, ease: EASE_OUT, delay: 0.2 }}
            className="relative"
          >
            {/* Accent line */}
            <div className="absolute -left-4 top-0 bottom-0 w-[1px] bg-gradient-to-b from-white/0 via-white/15 to-white/0 hidden lg:block" />
            <article
              className="prose-condura max-w-none"
              dangerouslySetInnerHTML={{ __html: processedHtml }}
            />
          </motion.div>
        </div>
      </section>

      {/* Footer note */}
      <section className="relative w-full py-20 px-6 border-t border-white/[0.08]">
        <div className="mx-auto max-w-3xl text-center">
          <p className="font-body-mature text-[14px] text-white/35">
            Questions about this agreement? Email{" "}
            <a href="mailto:legal@condura.app" className="text-white/60 underline decoration-white/20 underline-offset-4 hover:text-white">
              legal@condura.app
            </a>
            .
          </p>
        </div>
      </section>
    </main>
  );
}
