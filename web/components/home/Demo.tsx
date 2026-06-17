"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "motion/react";

const COMMANDS = [
  "Summarize this page",
  "Write a haiku about AI",
  "Open Safari",
  "Draft an email to the team",
];

const RESPONSES: Record<string, string> = {
  "Summarize this page": "This page describes Condura — a local-first AI agent that lives on your computer. It orchestrates every AI tool you have installed, operates through a hotkey-triggered overlay, and keeps all data on-device.",
  "Write a haiku about AI": "Ghost in the machine wakes —\nOne hotkey, one thought, one deed.\nDawn breaks in silence.",
  "Open Safari": "Opening Safari...",
  "Draft an email to the team": "Subject: Quick update\n\nHi team,\n\nJust wanted to share that Condura is now running locally. All tests pass. The overlay responds in under 100ms.\n\n— You",
};

export default function Demo() {
  const [isOpen, setIsOpen] = useState(false);
  const [command, setCommand] = useState<string | null>(null);
  const [typingIdx, setTypingIdx] = useState(0);
  const [thinking, setThinking] = useState(false);

  const runDemo = (cmd: string) => {
    setIsOpen(true);
    setCommand(cmd);
    setTypingIdx(0);
    setThinking(true);

    setTimeout(() => {
      setThinking(false);
      const resp = RESPONSES[cmd] ?? "Done.";
      let i = 0;
      const interval = setInterval(() => {
        i++;
        setTypingIdx(i);
        if (i >= resp.length) clearInterval(interval);
      }, 25);
    }, 1200);
  };

  return (
    <section className="relative overflow-hidden bg-[#0a0a0b] py-32">
      <div className="bg-circuit-fine pointer-events-none absolute inset-0 opacity-20" />
      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />

      <div className="relative z-10 mx-auto max-w-4xl px-6 text-center">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.6 }}
        >
          <p className="text-[13px] font-medium uppercase tracking-widest text-white/30 mb-3">
            Try it
          </p>
          <h2 className="gradient-headline text-[32px] font-semibold tracking-tighter sm:text-[44px]">
            See it in action
          </h2>
          <p className="mx-auto mt-4 max-w-lg text-[17px] leading-relaxed text-white/40">
            Click a command. Watch the overlay think, respond, and vanish.
            This is a simulation — but it feels real.
          </p>
        </motion.div>

        {/* Command pills */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.5, delay: 0.1 }}
          className="mt-10 flex flex-wrap justify-center gap-3"
        >
          {COMMANDS.map((cmd) => (
            <button
              key={cmd}
              onClick={() => runDemo(cmd)}
              className="rounded-full border border-white/[0.08] bg-white/[0.03] px-5 py-2.5 text-[14px] text-white/50 transition-all duration-200 hover:border-[#0066cc]/30 hover:bg-[#0066cc]/5 hover:text-white/80"
            >
              {cmd}
            </button>
          ))}
        </motion.div>

        {/* Inline overlay simulation */}
        <div className="mt-12 flex min-h-[280px] items-center justify-center">
          <AnimatePresence mode="wait">
            {isOpen ? (
              <motion.div
                key="overlay"
                initial={{ opacity: 0, y: 40, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: 20, scale: 0.97 }}
                transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] as [number, number, number, number] }}
                className="w-full max-w-[480px]"
              >
                <div className="glass-overlay rounded-2xl p-5">
                  <div className="flex items-center gap-2 border-b border-white/[0.06] pb-3">
                    <div className="h-2 w-2 rounded-full bg-[#ff5f57]" />
                    <div className="h-2 w-2 rounded-full bg-[#febc2e]" />
                    <div className="h-2 w-2 rounded-full bg-[#28c840]" />
                    <span className="ml-auto text-[11px] font-medium tracking-wide text-white/25">Condura</span>
                  </div>

                  <div className="mt-4 space-y-3">
                    <div className="flex items-start gap-2.5">
                      <div className="h-3 w-3 shrink-0 rounded-full bg-white/20" />
                      <div className="rounded-xl bg-white/[0.04] px-3.5 py-2.5 text-[13px] text-white/70">
                        {command}
                      </div>
                    </div>

                    {thinking && (
                      <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        className="flex items-center gap-2 pl-6"
                      >
                        <div className="animate-orb h-3 w-3 rounded-full bg-[#0066cc]" />
                        <span className="text-[13px] text-white/30">Thinking...</span>
                      </motion.div>
                    )}

                    {!thinking && command && (
                      <motion.div
                        initial={{ opacity: 0, y: 4 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: 0.1 }}
                        className="flex items-start gap-2.5 pl-6"
                      >
                        <div className="rounded-xl bg-[#0066cc]/10 px-3.5 py-2.5 text-[13px] text-[#64c8ff] leading-relaxed">
                          {RESPONSES[command].slice(0, typingIdx)}
                          {typingIdx < RESPONSES[command].length && (
                            <span className="animate-cursor ml-0.5 inline-block h-3.5 w-[2px] bg-[#64c8ff] align-middle" />
                          )}
                        </div>
                      </motion.div>
                    )}
                  </div>

                  <button
                    onClick={() => setIsOpen(false)}
                    className="mt-4 text-[12px] text-white/20 transition-colors hover:text-white/40"
                  >
                    Dismiss
                  </button>
                </div>
              </motion.div>
            ) : (
              <motion.div
                key="placeholder"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="text-white/15 text-[14px]"
              >
                Click a command above to begin
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>

      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />
    </section>
  );
}
