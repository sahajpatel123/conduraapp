"use client";

import { motion, useReducedMotion } from "motion/react";
import { FileClock, Fingerprint, Gauge, ScanLine, ShieldCheck } from "lucide-react";

const nodes = [
  {
    title: "AX-only",
    body: "Read structure before pixels when the interface is already named.",
    icon: ScanLine,
  },
  {
    title: "Verify",
    body: "Take the second snapshot before write or network actions.",
    icon: Gauge,
  },
  {
    title: "Gate",
    body: "Pause high-risk work at deterministic policy boundaries.",
    icon: ShieldCheck,
  },
  {
    title: "Approve",
    body: "Human approval remains the boundary for irreversible work.",
    icon: Fingerprint,
  },
  {
    title: "Audit",
    body: "Every important decision lands in a traceable timeline.",
    icon: FileClock,
  },
];

export function ControlTheater() {
  const reduceMotion = useReducedMotion();

  return (
    <div className="agent-theater">
      <div className="perception-map" aria-label="Synaptic selective perception stack">
        <motion.div
          className="perception-core"
          animate={
            reduceMotion
              ? undefined
              : {
                  scale: [1, 1.018, 1],
                }
          }
          transition={{ duration: 4.8, repeat: Infinity, ease: "easeInOut" }}
        >
          <div>
            <span />
            <p>Selective perception</p>
            <small>cheapest safe path first</small>
          </div>
        </motion.div>

        {nodes.map((node, index) => {
          const Icon = node.icon;
          return (
            <motion.article
              key={node.title}
              className="perception-node"
              animate={reduceMotion ? undefined : { opacity: [0.82, 1, 0.82] }}
              transition={{
                duration: 4.8 + index * 0.35,
                repeat: Infinity,
                delay: index * 0.28,
                ease: "easeInOut",
              }}
            >
              <Icon aria-hidden="true" size={19} />
              <span>0{index + 1}</span>
              <strong>{node.title}</strong>
              <p>{node.body}</p>
            </motion.article>
          );
        })}
      </div>
    </div>
  );
}
