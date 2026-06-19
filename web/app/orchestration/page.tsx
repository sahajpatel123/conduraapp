"use client";

import PageChrome from "@/components/shell/PageChrome";
import { motion, useScroll, useTransform, useSpring } from "motion/react";
import { useRef, useState, useEffect } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";

export default function OrchestrationPage() {
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end end"],
  });

  const smoothProgress = useSpring(scrollYProgress, { damping: 20, stiffness: 100 });
  
  const y1 = useTransform(smoothProgress, [0, 1], [0, -200]);
  const y2 = useTransform(smoothProgress, [0, 1], [0, -400]);
  const y3 = useTransform(smoothProgress, [0, 1], [0, -600]);

  const opacity1 = useTransform(smoothProgress, [0, 0.3, 0.4], [0, 1, 0.2]);
  const opacity2 = useTransform(smoothProgress, [0.3, 0.6, 0.7], [0, 1, 0.2]);
  const opacity3 = useTransform(smoothProgress, [0.6, 0.9, 1], [0, 1, 1]);

  // Simulated Log State
  const [logs, setLogs] = useState<string[]>([]);
  useEffect(() => {
    const messages = [
      "[SYS] Initializing SQLite Event Bus...",
      "[SYS] Memory mapped to /tmp/condura.db",
      "[AGENT:Strategist] Received intent: 'Refactor auth module'",
      "[AGENT:Strategist] Decomposing into 3 subtasks.",
      "[BUS] Spawned agent-01 (AST Parser)",
      "[BUS] Spawned agent-02 (Dependency Graph)",
      "[BUS] Spawned agent-03 (State Machine Analyzer)",
      "[AGENT-01] Parsing /src/auth.ts...",
      "[AGENT-02] Traversing imports...",
      "[AGENT-03] Found 4 unhandled state transitions.",
      "[GATEKEEPER] Requesting FS Write permission...",
      "[USER] Granted.",
      "[BUS] Applying diffs to /src/auth.ts",
      "[SYS] Commit successful. Hash: 8f4a2b1",
    ];
    let i = 0;
    const interval = setInterval(() => {
      if (i < messages.length) {
        setLogs(prev => [...prev, messages[i]]);
        i++;
      } else {
        clearInterval(interval);
      }
    }, 800);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="bg-black text-white min-h-screen">
      <PageChrome
        eyebrow="Engine"
        title="Massive parallel workflows."
        description="Condura doesn't just run agents sequentially. It spins up highly concurrent, local swarms that communicate through a fast SQLite event bus. This is the story of how it thinks."
        badge="Orchestration"
      >
        {/* --- SECTION 1: The 3 Phases (Sticky Scroll) --- */}
        <div ref={containerRef} className="relative mt-24 h-[300vh]">
          <div className="sticky top-32 flex h-[80vh] items-center justify-center overflow-hidden rounded-[32px] border border-white/10 bg-[#050505] shadow-[0_0_80px_rgba(255,255,255,0.03)]">
            <div className="absolute inset-0 bg-grid-dark opacity-30" />
            
            {/* Phase 1: Planning */}
            <motion.div style={{ y: y1, opacity: opacity1 }} className="absolute inset-0 flex flex-col items-center justify-center p-8">
              <div className="mb-8 rounded-2xl border border-white/10 bg-white/[0.02] p-6 backdrop-blur-xl">
                <h3 className="font-mono text-[11px] uppercase tracking-widest text-white/40 mb-2">Phase 1</h3>
                <p className="text-2xl font-medium text-white">Decomposition</p>
                <p className="text-sm text-white/40 mt-2 max-w-sm text-center">A single complex prompt is torn down into discrete, manageable sub-tasks. The Strategist model plans the assault.</p>
              </div>
              <div className="flex gap-4">
                {[...Array(4)].map((_, i) => (
                  <motion.div
                    key={i}
                    initial={{ opacity: 0, scale: 0.8 }}
                    whileInView={{ opacity: 1, scale: 1 }}
                    transition={{ delay: i * 0.1, type: "spring" }}
                    className="w-16 h-16 rounded-xl border border-white/20 bg-white/5 flex items-center justify-center shadow-[0_0_20px_rgba(255,255,255,0.05)]"
                  >
                    <span className="font-mono text-white/50 text-[10px]">T-{i+1}</span>
                  </motion.div>
                ))}
              </div>
            </motion.div>

            {/* Phase 2: Fan Out */}
            <motion.div style={{ y: y2, opacity: opacity2 }} className="absolute inset-0 flex flex-col items-center justify-center p-8">
              <div className="mb-8 rounded-2xl border border-white/10 bg-white/[0.02] p-6 backdrop-blur-xl">
                <h3 className="font-mono text-[11px] uppercase tracking-widest text-white/40 mb-2">Phase 2</h3>
                <p className="text-2xl font-medium text-white">Parallel Fan-Out</p>
                <p className="text-sm text-white/40 mt-2 max-w-sm text-center">Condura spawns lightweight sub-agents for each task, running them concurrently to slash execution time.</p>
              </div>
              <div className="relative w-full max-w-lg h-64 border border-white/10 rounded-3xl bg-black/50 p-6 overflow-hidden">
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="w-16 h-16 rounded-full bg-white/10 blur-xl animate-pulse" />
                </div>
                <div className="flex flex-col gap-3 relative z-10">
                  {[
                    "Spawning react-agent...",
                    "Spawning rust-agent...",
                    "Mounting DOM analyzer...",
                    "Starting headless browser..."
                  ].map((text, i) => (
                    <motion.div key={i} initial={{ x: -20, opacity: 0 }} whileInView={{ x: 0, opacity: 1 }} transition={{ delay: i * 0.15 }} className="font-mono text-[12px] text-white/60 flex items-center gap-3">
                      <span className="w-2 h-2 rounded-full bg-white/30" />
                      {text}
                    </motion.div>
                  ))}
                </div>
              </div>
            </motion.div>

            {/* Phase 3: Resolution */}
            <motion.div style={{ y: y3, opacity: opacity3 }} className="absolute inset-0 flex flex-col items-center justify-center p-8">
               <div className="mb-8 rounded-2xl border border-white/10 bg-white/[0.02] p-6 backdrop-blur-xl text-center">
                <h3 className="font-mono text-[11px] uppercase tracking-widest text-white/40 mb-2">Phase 3</h3>
                <p className="text-2xl font-medium text-white">Deterministic Resolution</p>
                <p className="text-sm text-white/40 mt-2 max-w-sm text-center">Results are collated, verified by strict logic (not a hallucinating LLM), and committed to your workspace.</p>
              </div>
              <motion.div 
                className="w-32 h-32 rounded-[2rem] border-2 border-white/20 bg-white/10 flex items-center justify-center shadow-[0_0_60px_rgba(255,255,255,0.1)]"
                animate={{ rotate: 360 }}
                transition={{ duration: 10, repeat: Infinity, ease: "linear" }}
              >
                <div className="w-16 h-16 rounded-full border border-white/30 bg-white/5" />
              </motion.div>
              <p className="mt-8 font-mono text-[12px] text-white/40">Diffs merged. AST verified. Lockfile updated.</p>
            </motion.div>
          </div>
        </div>

        {/* --- SECTION 2: The Event Bus (Live Terminal) --- */}
        <div className="mt-32 max-w-4xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-3xl md:text-5xl font-semibold tracking-tight mb-4 text-white">The SQLite Event Bus</h2>
            <p className="text-lg text-white/50 max-w-2xl mx-auto">Agents don't just "talk" to each other in a vacuum. Every thought, state change, and action is written to a highly-concurrent local SQLite database. This creates a permanent, auditable, and replayable state.</p>
          </div>

          <div className="relative rounded-[24px] border border-white/10 bg-[#0a0a0a] overflow-hidden shadow-2xl">
            {/* Terminal Header */}
            <div className="h-12 border-b border-white/5 bg-white/[0.02] flex items-center px-6 gap-2">
              <div className="w-3 h-3 rounded-full bg-white/20" />
              <div className="w-3 h-3 rounded-full bg-white/20" />
              <div className="w-3 h-3 rounded-full bg-white/20" />
              <span className="ml-4 font-mono text-xs text-white/30">condura-bus-monitor.db</span>
            </div>
            {/* Terminal Body */}
            <div className="p-6 h-[400px] overflow-y-auto font-mono text-[13px] leading-relaxed relative">
              <div className="absolute inset-0 bg-gradient-to-b from-transparent via-transparent to-[#0a0a0a] pointer-events-none" />
              {logs.map((log, i) => (
                <motion.div 
                  key={i} 
                  initial={{ opacity: 0, x: -10 }} 
                  animate={{ opacity: 1, x: 0 }} 
                  className={`mb-2 ${log.includes('[SYS]') ? 'text-white/40' : log.includes('[AGENT') ? 'text-white/80' : log.includes('[GATEKEEPER]') ? 'text-[#ffb86c]' : log.includes('[USER]') ? 'text-[#50fa7b]' : 'text-white/60'}`}
                >
                  <span className="mr-3 opacity-50">{new Date().toISOString().split('T')[1].split('.')[0]}</span>
                  {log}
                </motion.div>
              ))}
              {logs.length < 14 && (
                <motion.div animate={{ opacity: [1, 0] }} transition={{ repeat: Infinity, duration: 0.8 }} className="w-2 h-4 bg-white/50 inline-block mt-1" />
              )}
            </div>
          </div>
        </div>

        {/* --- SECTION 3: The Architecture Graph --- */}
        <div className="mt-32 pb-32 max-w-5xl mx-auto">
           <div className="text-center mb-16">
            <h2 className="text-3xl md:text-5xl font-semibold tracking-tight mb-4 text-white">Built for speed.</h2>
            <p className="text-lg text-white/50 max-w-2xl mx-auto">No Python dependency hell. Condura is packaged as a single standalone binary. The core engine is incredibly fast, leaning on native OS primitives.</p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {[
              { stat: "< 50ms", label: "Agent Spawn Time", desc: "Using lightweight V8 isolates instead of heavy system processes." },
              { stat: "100k+", label: "Events per second", desc: "The SQLite WAL mode handles massive concurrent write loads effortlessly." },
              { stat: "0 bytes", label: "Cloud Storage", desc: "Every byte of your code stays on your local filesystem." }
            ].map((item, i) => (
              <motion.div 
                key={i}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, margin: "-100px" }}
                transition={{ delay: i * 0.1 }}
                className="p-8 rounded-3xl border border-white/10 bg-white/[0.02] flex flex-col items-center text-center hover:bg-white/[0.04] transition-colors"
              >
                <div className="text-4xl font-display font-semibold text-white mb-2">{item.stat}</div>
                <div className="text-[13px] uppercase tracking-widest text-white/40 font-mono mb-4">{item.label}</div>
                <p className="text-white/50 text-sm leading-relaxed">{item.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>

      </PageChrome>
    </div>
  );
}
