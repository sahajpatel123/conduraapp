"use client";

import { useState, useEffect } from "react";
import { motion } from "motion/react";

const HEADLINE_WORDS = ["AI", "on", "your", "computer.", "Free."];
const TYPING_LINES = [
  { text: "Summarize the quarterly report...", delay: 800 },
  { text: "Here's your summary. Revenue up 12%.", delay: 2200, isResponse: true },
];

function useFirstVisit() {
  const [isFirst] = useState(() => {
    if (typeof window === "undefined") return false;
    const key = "condura-hero-seen";
    const alreadyPlayed = sessionStorage.getItem(key);
    if (!alreadyPlayed) {
      sessionStorage.setItem(key, "1");
      return true;
    }
    return false;
  });
  return isFirst;
}

function TypingText({ text, delay, isResponse }: { text: string; delay: number; isResponse?: boolean }) {
  const isFirst = useFirstVisit();
  const [displayed, setDisplayed] = useState(isFirst ? "" : text);

  useEffect(() => {
    if (!isFirst) return;
    let i = 0;
    const start = setTimeout(() => {
      const interval = setInterval(() => {
        i++;
        setDisplayed(text.slice(0, i));
        if (i >= text.length) clearInterval(interval);
      }, 40);
      return () => clearInterval(interval);
    }, delay);
    return () => clearTimeout(start);
  }, [text, delay, isFirst]);

  return (
    <span className={isResponse ? "text-[#64c8ff]" : "text-[#e5e5e5]"}>
      {displayed}
      {displayed.length < text.length && !isResponse && (
        <span className="animate-cursor ml-0.5 inline-block h-4 w-[2px] bg-[#0066cc] align-middle" />
      )}
    </span>
  );
}

function Orb() {
  return (
    <div className="relative flex items-center justify-center">
      <div className="animate-orb h-3 w-3 rounded-full bg-[#0066cc]" />
    </div>
  );
}

export default function HeroSection() {
  const isFirst = useFirstVisit();
  const [showLines] = useState(!isFirst);

  useEffect(() => {
    if (!isFirst) return;
    const t = setTimeout(() => {}, 400);
    return () => clearTimeout(t);
  }, [isFirst]);

  return (
    <section className="relative flex min-h-screen items-center justify-center overflow-hidden bg-[#050505]">
      <div className="bg-circuit pointer-events-none absolute inset-0 opacity-40" />
      <div className="radiant-aura pointer-events-none absolute inset-0" />

      <div className="relative z-10 mx-auto max-w-5xl px-6 text-center">
        <motion.div
          initial={isFirst ? { opacity: 0, scale: 0.9, filter: "blur(12px)" } : false}
          animate={{ opacity: 1, scale: 1, filter: "blur(0px)" }}
          transition={{ duration: 1.2, ease: [0.22, 1, 0.36, 1] as [number, number, number, number] }}
          className="animate-float mx-auto mb-12 max-w-[520px]"
        >
          <div className="glass-overlay relative overflow-hidden rounded-2xl p-5">
            <div className="flex items-center gap-2 border-b border-white/[0.06] pb-3">
              <div className="h-2 w-2 rounded-full bg-[#ff5f57]" />
              <div className="h-2 w-2 rounded-full bg-[#febc2e]" />
              <div className="h-2 w-2 rounded-full bg-[#28c840]" />
              <span className="ml-auto text-[11px] font-medium tracking-wide text-white/30">Condura</span>
            </div>

            <div className="mt-4 space-y-3 text-left">
              {showLines && TYPING_LINES.map((line, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: isFirst ? line.delay / 1000 : 0, duration: 0.4 }}
                  className="flex items-start gap-2.5"
                >
                  {!line.isResponse && <Orb />}
                  <div
                    className={`rounded-xl px-3.5 py-2.5 text-[13px] leading-relaxed ${
                      line.isResponse
                        ? "bg-[#0066cc]/10 border border-[#0066cc]/15 text-[#64c8ff]"
                        : "bg-white/[0.04] text-[#e5e5e5]"
                    }`}
                  >
                    <TypingText text={line.text} delay={line.delay} isResponse={line.isResponse} />
                  </div>
                </motion.div>
              ))}

              {showLines && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  transition={{ delay: isFirst ? 4.5 : 0, duration: 0.3 }}
                  className="flex items-center gap-1.5 pl-6"
                >
                  <span className="h-1 w-1 animate-[typing-dot_1.4s_infinite_-0.32s] rounded-full bg-white/30" />
                  <span className="h-1 w-1 animate-[typing-dot_1.4s_infinite_-0.16s] rounded-full bg-white/30" />
                  <span className="h-1 w-1 animate-[typing-dot_1.4s_infinite_0s] rounded-full bg-white/30" />
                </motion.div>
              )}
            </div>

            <div className="mt-4 flex items-center gap-2 border-t border-white/[0.06] pt-3 text-[12px] text-white/25">
              <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 18.75a6 6 0 006-6v-1.5m-6 7.5a6 6 0 01-6-6v-1.5m6 7.5v3.75m-3.75 0h7.5M12 15.75a3 3 0 01-3-3V4.5a3 3 0 116 0v8.25a3 3 0 01-3 3z" />
              </svg>
              Say &ldquo;hey condura&rdquo; or type...
            </div>
          </div>
        </motion.div>

        <h1 className="text-balance text-[56px] font-semibold leading-[1.05] tracking-tighter text-white sm:text-[72px] md:text-[88px]">
          {HEADLINE_WORDS.map((word, i) => (
            <motion.span
              key={i}
              initial={isFirst ? { opacity: 0, y: 20 } : false}
              animate={{ opacity: 1, y: 0 }}
              transition={{
                duration: 0.6,
                delay: 0.1 + i * 0.08,
                ease: [0.22, 1, 0.36, 1] as [number, number, number, number],
              }}
              className="mr-[0.2em] inline-block"
            >
              {word}
            </motion.span>
          ))}
        </h1>

        <motion.p
          initial={isFirst ? { opacity: 0, y: 12 } : false}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.7, ease: [0.22, 1, 0.36, 1] as [number, number, number, number] }}
          className="mx-auto mt-5 max-w-lg text-[17px] leading-relaxed text-white/40"
        >
          A ghost that lives inside your computer. Press a hotkey. It appears.
          Orchestrates every AI tool you have. Then vanishes.
        </motion.p>

        <motion.div
          initial={isFirst ? { opacity: 0, y: 12 } : false}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.9 }}
          className="mt-10"
        >
          <a
            href="/download"
            className="group inline-flex items-center rounded-full bg-[#0066cc] px-8 py-4 text-[16px] font-semibold text-white transition-all duration-300 hover:bg-[#0055aa] hover:shadow-[0_0_40px_rgba(0,102,204,0.4)] active:scale-[0.95]"
          >
            Download Condura
            <svg className="ml-2 h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12h15m0 0l-6.75-6.75M19.5 12l-6.75 6.75" />
            </svg>
          </a>
          <p className="mt-3 text-[13px] text-white/25">
            Free forever. No account. No tracking. No cloud.
          </p>
        </motion.div>
      </div>

      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />
    </section>
  );
}
