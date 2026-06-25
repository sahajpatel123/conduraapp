"use client";

import PageChrome from "@/components/shell/PageChrome";
import { motion, AnimatePresence } from "motion/react";
import { useState, useEffect } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";

export default function SecurityPage() {
  const [isHovered, setIsHovered] = useState(false);

  const [auditLogs, setAuditLogs] = useState<string[]>([]);
  useEffect(() => {
    const logs = [
      "INSERT INTO audit_log (actor, action, resource) VALUES ('condura-core', 'READ', '/src/index.ts');",
      "INSERT INTO audit_log (actor, action, resource) VALUES ('strategist', 'PROPOSE', 'npm install express');",
      "INSERT INTO audit_log (actor, action, resource) VALUES ('gatekeeper', 'BLOCK', 'npm install express');",
      "INSERT INTO audit_log (actor, action, resource) VALUES ('gatekeeper', 'PROMPT_USER', 'npm install express');",
      "INSERT INTO audit_log (actor, action, resource) VALUES ('user', 'GRANT', 'npm install express');",
      "INSERT INTO audit_log (actor, action, resource) VALUES ('condura-core', 'EXEC', 'npm install express');",
      "INSERT INTO audit_log (actor, action, resource) VALUES ('condura-core', 'WRITE', 'package.json');",
    ];
    let i = 0;
    const int = setInterval(() => {
      setAuditLogs(prev => [...prev, logs[i]]);
      i++;
      if (i >= logs.length) clearInterval(int);
    }, 1200);
    return () => clearInterval(int);
  }, []);

  return (
    <div className="bg-black text-white min-h-screen">
      <PageChrome
        eyebrow="Security Gatekeeper"
        title="Capability without the blank check."
        description="Models hallucinate, prompts get injected, and some actions cannot be undone. Condura draws a hard line between thinking and acting with a deterministic permission gatekeeper."
        badge="Zero Trust"
      >
        {/* --- SECTION 1: The Vault --- */}
        <div className="mt-24 h-[600px] w-full rounded-[32px] border border-white/10 bg-[#050505] relative overflow-hidden flex items-center justify-center p-8 shadow-2xl">
          <motion.div 
            animate={{ opacity: isHovered ? 0.4 : 0.1, scale: isHovered ? 1.2 : 1 }}
            transition={{ duration: 1 }}
            className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(255,255,255,0.2),transparent_50%)]"
          />
          <div 
            className="relative z-10 w-[400px] h-[400px] rounded-full border-4 border-white/5 flex items-center justify-center cursor-pointer group"
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
          >
            <motion.div animate={{ rotate: isHovered ? 180 : 0 }} transition={{ type: "spring", stiffness: 50, damping: 20 }} className="absolute inset-0 rounded-full border-2 border-dashed border-white/20" />
            <motion.div animate={{ rotate: isHovered ? -90 : 0 }} transition={{ type: "spring", stiffness: 60, damping: 25 }} className="absolute inset-8 rounded-full border border-white/10 flex items-center justify-center">
              {[0, 90, 180, 270].map((deg) => (
                <motion.div key={deg} style={{ rotate: deg }} className="absolute w-full h-[2px] flex justify-between">
                  <motion.div animate={{ scaleX: isHovered ? 0 : 1 }} transition={{ duration: 0.3 }} className="w-4 h-full bg-white/40 origin-left" />
                  <motion.div animate={{ scaleX: isHovered ? 0 : 1 }} transition={{ duration: 0.3 }} className="w-4 h-full bg-white/40 origin-right" />
                </motion.div>
              ))}
            </motion.div>
            <motion.div 
              animate={{ 
                scale: isHovered ? 1.05 : 1,
                boxShadow: isHovered ? "0 0 60px rgba(255,255,255,0.1), inset 0 0 20px rgba(255,255,255,0.05)" : "0 0 0px rgba(255,255,255,0), inset 0 0 0px rgba(255,255,255,0)"
              }}
              transition={{ duration: 0.5 }}
              className="relative w-48 h-48 rounded-full bg-black border border-white/20 flex flex-col items-center justify-center z-20"
            >
              <motion.div animate={{ y: isHovered ? -5 : 0 }} className="font-mono text-[11px] text-white/40 uppercase tracking-widest mb-2 transition-all group-hover:text-white/80">Hover to Unlock</motion.div>
              <motion.div animate={{ color: isHovered ? "rgba(255,255,255,1)" : "rgba(255,255,255,0.4)" }} className="text-xl font-medium tracking-tight">
                {isHovered ? "UNLOCKED" : "LOCKED"}
              </motion.div>
              <AnimatePresence>
                {isHovered && (
                  <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -10 }} className="absolute -bottom-16 flex gap-2">
                    <AnimatedBadge tone="neutral">Read FS</AnimatedBadge>
                    <AnimatedBadge tone="neutral">Port 3000</AnimatedBadge>
                  </motion.div>
                )}
              </AnimatePresence>
            </motion.div>
          </div>
        </div>

        {/* --- SECTION 2: The Core Principles --- */}
        <div className="mt-24 grid md:grid-cols-3 gap-6">
          {[
            { title: "Deterministic Rules", desc: "Security rules are hard-coded in TypeScript, not written in a prompt. A model cannot convince the Gatekeeper to drop its guard using clever linguistics." },
            { title: "Air-gapped Memory", desc: "Your vector store, embeddings, and workspace memory live entirely on your local SSD. No proprietary code is ever indexed in the cloud." },
            { title: "Kill Switches", desc: "Four independent kill switches: a hard hotkey, a watchdog timer, a network guard that blocks all non-allow-listed egress, and a UI menu-bar kill button." }
          ].map((feat, i) => (
            <motion.div key={i} initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.1 }} className="p-8 rounded-[32px] border border-white/10 bg-white/[0.02]">
              <h4 className="text-white text-xl font-medium mb-4">{feat.title}</h4>
              <p className="text-white/50 text-[15px] leading-relaxed">{feat.desc}</p>
            </motion.div>
          ))}
        </div>

        {/* --- SECTION 3: The Immutable Audit Log --- */}
        <div className="mt-40 mb-32 flex flex-col md:flex-row gap-16 items-center">
           <div className="flex-1 w-full relative">
            <div className="absolute -inset-10 bg-gradient-to-br from-green-500/10 to-emerald-500/10 blur-3xl opacity-30 pointer-events-none" />
            <div className="rounded-2xl border border-white/10 bg-[#0a0a0a] shadow-2xl overflow-hidden relative z-10">
               <div className="h-10 border-b border-white/5 bg-white/[0.02] flex items-center px-4 gap-2">
                <span className="font-mono text-[10px] text-white/30 tracking-widest uppercase">condura_audit.sqlite</span>
              </div>
              <div className="p-6 h-[300px] overflow-y-auto">
                <div className="font-mono text-[12px] leading-[1.8] text-white/70">
                  {auditLogs.map((log, idx) => (
                    <motion.div key={idx} initial={{ opacity: 0, x: -10 }} animate={{ opacity: 1, x: 0 }} className="mb-2">
                      <span className="text-white/30 mr-2">{String(idx + 1).padStart(4, '0')}</span>
                      <span className={log.includes('BLOCK') ? 'text-[#ff5f57]' : log.includes('GRANT') ? 'text-[#28c840]' : log.includes('PROMPT') ? 'text-[#febc2e]' : 'text-white/70'}>
                        {log}
                      </span>
                    </motion.div>
                  ))}
                  {auditLogs.length < 7 && (
                    <motion.span animate={{ opacity: [1, 0] }} transition={{ repeat: Infinity, duration: 0.8 }} className="w-2 h-4 bg-white/50 inline-block align-middle" />
                  )}
                </div>
              </div>
            </div>
          </div>

          <div className="flex-1">
            <h2 className="text-3xl md:text-5xl font-semibold tracking-tight mb-6 text-white">Immutable Audit Trail.</h2>
            <p className="text-lg text-white/50 leading-relaxed mb-6">
              Trust requires verification. Every single action Condura takes — from reading a file to spawning a sub-process to hitting an external API — is logged locally in an SQLite database.
              <br/><br/>
              These logs are HMAC-chained. If you ever need to know exactly what the agent did while you were away, the proof is mathematically guaranteed to be accurate and untampered.
            </p>
          </div>
        </div>

      </PageChrome>
    </div>
  );
}
