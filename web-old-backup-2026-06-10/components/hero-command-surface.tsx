"use client";

import { motion, useReducedMotion } from "motion/react";
import { CheckCircle2, LockKeyhole, Mic2, ShieldAlert, SquareTerminal } from "lucide-react";
import { commandTrace } from "@/lib/site-data";

const ease = [0.22, 1, 0.36, 1] as const;

function toneClass(tone: string) {
  if (tone === "green") return "trace-green";
  if (tone === "amber") return "trace-amber";
  return "trace-blue";
}

export function HeroCommandSurface() {
  const reduceMotion = useReducedMotion();

  return (
    <div className="command-preview" aria-label="Synaptic command state preview">
      <motion.div
        className="command-frame"
        initial={false}
        animate={reduceMotion ? undefined : { opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 0.52, ease }}
      >
        <div className="command-top">
          <div className="window-dots">
            <span className="dot-hot" />
            <em>edge overlay / local session</em>
          </div>
          <motion.div className="status-pill" animate={reduceMotion ? undefined : { opacity: [0.7, 1, 0.7] }} transition={{ duration: 3.6, repeat: Infinity, ease: "easeInOut" }}>
            <span />
            gate active
          </motion.div>
        </div>

        <div className="command-body">
          <div className="prompt-card">
            <div className="prompt-main">
              <span className="icon-box">
                  <Mic2 aria-hidden="true" size={19} />
                </span>
              <div>
                <p className="kicker">voice intent</p>
                <p className="prompt-text">Summarize this folder and draft a reply.</p>
              </div>
            </div>
            <Waveform reduceMotion={Boolean(reduceMotion)} />
          </div>

          <div className="command-grid">
            <div className="plan-card">
              <div className="card-heading">
                <SquareTerminal aria-hidden="true" size={17} />
                Visible plan
              </div>
              <ol className="plan-list">
                {["Read active folder names", "Draft summary locally", "Ask before sending reply"].map((step, index) => (
                  <motion.li
                    key={step}
                    initial={false}
                    animate={reduceMotion ? undefined : { opacity: 1, x: 0 }}
                    transition={{ duration: 0.35, delay: 0.35 + index * 0.16, ease }}
                  >
                    <span className="step-index">
                      {index + 1}
                    </span>
                    {step}
                  </motion.li>
                ))}
              </ol>
            </div>

            <motion.div
              className="gate-card"
              initial={false}
              animate={reduceMotion ? undefined : { opacity: 1, y: 0 }}
              transition={{ duration: 0.42, delay: 0.95, ease }}
            >
              <div className="card-heading">
                <ShieldAlert aria-hidden="true" size={17} />
                Gatekeeper
              </div>
              <p>
                Network action detected. Waiting for human approval before sending.
              </p>
              <div className="approval-box">
                approval required
              </div>
            </motion.div>
          </div>

          <div className="trace-stack">
            {commandTrace.map((item, index) => (
              <motion.div
                key={item.label}
                className="trace-row"
                initial={false}
                animate={reduceMotion ? undefined : { opacity: 1, y: 0 }}
                transition={{ duration: 0.32, delay: 1.08 + index * 0.1, ease }}
              >
                <span className="trace-label">{item.label}</span>
                <span className={`trace-value ${toneClass(item.tone)}`}>
                  {item.value}
                </span>
              </motion.div>
            ))}
          </div>

          <div className="audit-card">
            <CheckCircle2 aria-hidden="true" size={18} />
            Audit event prepared before action leaves the machine.
            <LockKeyhole className="ml-auto" aria-hidden="true" size={17} />
          </div>
        </div>
      </motion.div>
    </div>
  );
}

function Waveform({ reduceMotion }: { reduceMotion: boolean }) {
  const bars = [18, 28, 14, 34, 22, 30, 16];

  return (
    <div className="waveform" aria-hidden="true">
      {bars.map((height, index) => (
        <motion.span
          key={`${height}-${index}`}
          style={{ height }}
          animate={reduceMotion ? undefined : { scaleY: [0.55, 1, 0.68] }}
          transition={{
            duration: 1.4,
            repeat: Infinity,
            delay: index * 0.09,
            ease: "easeInOut",
          }}
        />
      ))}
    </div>
  );
}
