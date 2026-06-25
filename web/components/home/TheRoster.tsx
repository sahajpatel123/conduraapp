"use client";

import { useRef, useState } from "react";
import { motion } from "motion/react";
import Reveal from "@/components/motion/Reveal";
import { EASE_OUT } from "@/lib/motion";

/**
 * TheRoster — the integrations constellation.
 * Each tool is a node on a thread. Hovering a node draws a synapse line
 * from the central Condura node to it and lights it up. Clicking pins it.
 */
const TOOLS = [
  { name: "Claude Code", kind: "CLI", color: "var(--color-pollen)" },
  { name: "Codex", kind: "CLI", color: "var(--color-sky-deep)" },
  { name: "Antigravity", kind: "CLI", color: "var(--color-synapse)" },
  { name: "OpenCode", kind: "CLI", color: "var(--color-pollen-deep)" },
  { name: "Kilo", kind: "CLI", color: "var(--color-ink-mute)" },
  { name: "Hermes", kind: "CLI", color: "var(--color-synapse-deep)" },
  { name: "Gemini", kind: "CLI", color: "var(--color-sky-deep)" },
  { name: "Ollama", kind: "Local", color: "var(--color-synapse-glow)" },
  { name: "Anthropic", kind: "API", color: "var(--color-pollen)" },
  { name: "OpenAI", kind: "API", color: "var(--color-ink)" },
  { name: "Google", kind: "API", color: "var(--color-sky-deep)" },
  { name: "xAI", kind: "API", color: "var(--color-ink-mute)" },
  { name: "Mistral", kind: "API", color: "var(--color-pollen-deep)" },
  { name: "DeepSeek", kind: "API", color: "var(--color-synapse)" },
  { name: "OpenRouter", kind: "API", color: "var(--color-ink-faint)" },
  { name: "Groq", kind: "API", color: "var(--color-pollen-light)" },
] as const;

export default function TheRoster() {
  const [hover, setHover] = useState<number | null>(null);
  const wrapRef = useRef<HTMLDivElement | null>(null);

  return (
    <section className="relative mx-auto max-w-[1180px] px-6 py-28 sm:py-36">
      <Reveal>
        <p className="text-eyebrow mb-4">— The roster</p>
      </Reveal>
      <div className="grid gap-6 md:grid-cols-[1fr_1.4fr] md:gap-16">
        <Reveal as="h2" className="text-display text-[var(--color-ink)] text-balance">
          It conducts the tools you already have.
        </Reveal>
        <Reveal delay={0.1} as="p" className="text-lead text-[var(--color-ink-soft)] max-w-[52ch] text-pretty md:pt-3">
          Condura auto-detects the AI CLIs in your <span className="font-mono text-[var(--color-ink)]">$PATH</span> and the API keys you give it. The ones you don&apos;t have simply don&apos;t appear. No installs forced. No vendor lock-in. Bring your own everything.
        </Reveal>
      </div>

      <Reveal delay={0.15}>
        <div
          ref={wrapRef}
          className="surface-card relative mt-14 overflow-hidden p-8 sm:p-12"
        >
          <div className="paper-grain absolute inset-0" />
          {/* center node */}
          <div className="relative z-10 grid place-items-center py-10">
            <svg viewBox="0 0 600 420" className="w-full max-w-[760px]" aria-hidden>
              {/* center */}
              <g>
                <circle cx="300" cy="210" r="34" fill="var(--color-ink)" />
                <circle cx="300" cy="210" r="34" fill="none" stroke="var(--color-synapse)" strokeWidth="1.5" opacity="0.6">
                  <animate attributeName="r" values="34;46;34" dur="3s" repeatCount="indefinite" />
                  <animate attributeName="opacity" values="0.6;0;0.6" dur="3s" repeatCount="indefinite" />
                </circle>
                <text
                  x="300"
                  y="214"
                  textAnchor="middle"
                  fill="var(--color-paper)"
                  className="font-display"
                  fontSize="15"
                  fontStyle="italic"
                >
                  Condura
                </text>
              </g>

              {/* tool nodes around the circle */}
              {TOOLS.map((t, i) => {
                const angle = (i / TOOLS.length) * Math.PI * 2 - Math.PI / 2;
                const R = 170;
                const x = 300 + Math.cos(angle) * R;
                const y = 210 + Math.sin(angle) * R;
                const isHover = hover === i;
                return (
                  <g
                    key={t.name}
                    onMouseEnter={() => setHover(i)}
                    onMouseLeave={() => setHover(null)}
                    style={{ cursor: "none" }}
                  >
                    {/* connecting thread */}
                    <motion.path
                      d={`M 300 210 L ${x} ${y}`}
                      stroke="var(--color-synapse)"
                      strokeWidth={isHover ? 1.4 : 0.6}
                      strokeDasharray="180"
                      animate={{ strokeDashoffset: isHover ? 0 : 180, opacity: isHover ? 1 : 0.25 }}
                      transition={{ duration: 0.5, ease: EASE_OUT }}
                      fill="none"
                    />
                    {/* node */}
                    <motion.circle
                      cx={x}
                      cy={y}
                      r={isHover ? 7 : 5}
                      fill={t.color}
                      animate={{ scale: isHover ? 1.3 : 1 }}
                      transition={{ duration: 0.3, ease: EASE_OUT }}
                      style={{ transformBox: "fill-box", transformOrigin: "center" }}
                    />
                    {isHover && (
                      <circle cx={x} cy={y} r="11" fill="none" stroke={t.color} strokeWidth="0.8" opacity="0.5">
                        <animate attributeName="r" values="7;14;7" dur="1.2s" repeatCount="indefinite" />
                        <animate attributeName="opacity" values="0.6;0;0.6" dur="1.2s" repeatCount="indefinite" />
                      </circle>
                    )}
                    {/* label */}
                    <text
                      x={x}
                      y={y + 22}
                      textAnchor="middle"
                      fill={isHover ? "var(--color-ink)" : "var(--color-ink-mute)"}
                      fontSize="11"
                      fontFamily="var(--font-mono)"
                      style={{ transition: "fill 0.3s" }}
                    >
                      {t.name}
                    </text>
                  </g>
                );
              })}
            </svg>
          </div>

          {/* hover detail card */}
          <div className="relative z-10 mx-auto flex max-w-[420px] flex-col items-center text-center">
            <motion.div
              key={hover ?? "none"}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3, ease: EASE_OUT }}
              className="text-small"
            >
              {hover === null ? (
                <span className="text-[var(--color-ink-mute)]">
                  Hover a node to see how Condura reaches it.
                </span>
              ) : (
                <span className="text-[var(--color-ink-soft)]">
                  <span className="font-mono text-[var(--color-ink)]">{TOOLS[hover].name}</span>{" "}
                  · {TOOLS[hover].kind} · auto-detected
                </span>
              )}
            </motion.div>
          </div>
        </div>
      </Reveal>
    </section>
  );
}
