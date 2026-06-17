"use client";

import PageChrome from "@/components/shell/PageChrome";
import { motion } from "motion/react";
import MagneticButton from "@/components/motion/MagneticButton";

const ECOSYSTEM_TOOLS = [
  { name: "React", desc: "AST patching and DOM analysis", size: "lg" },
  { name: "Next.js", desc: "Routing and SSR management", size: "md" },
  { name: "TypeScript", desc: "Type-safe deterministic execution", size: "lg" },
  { name: "Rust", desc: "Core engine performance", size: "xl" },
  { name: "Tailwind", desc: "Utility-first design injection", size: "sm" },
  { name: "Git", desc: "Version control sync", size: "md" },
  { name: "Docker", desc: "Containerized isolation", size: "lg" },
  { name: "Node.js", desc: "V8 runtime operations", size: "xl" },
  { name: "PostgreSQL", desc: "Vector state memory", size: "md" },
];

export default function EcosystemPage() {
  return (
    <PageChrome
      eyebrow="Ecosystem"
      title="Plugs into everything you use."
      description="Condura doesn't ask you to change your stack. It uses native AST parsers, LSP integrations, and raw terminal APIs to work exactly where you already work."
      badge="Integration"
    >
      <div className="mt-24 min-h-[60vh] relative flex flex-col items-center">
        
        {/* Dynamic Tool Grid */}
        <div className="w-full flex flex-wrap justify-center gap-6 p-8">
          {ECOSYSTEM_TOOLS.map((tool, i) => (
            <MagneticButton key={tool.name}>
              <motion.div
                initial={{ opacity: 0, scale: 0.8, y: 20 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                transition={{ 
                  delay: i * 0.1, 
                  type: "spring", 
                  stiffness: 100, 
                  damping: 15 
                }}
                className={`relative overflow-hidden rounded-[24px] border border-white/10 bg-white/[0.02] backdrop-blur-md p-6 group cursor-pointer hover:bg-white/[0.05] transition-colors
                  ${tool.size === "xl" ? "w-[300px] h-[300px]" : 
                    tool.size === "lg" ? "w-[240px] h-[240px]" : 
                    tool.size === "md" ? "w-[200px] h-[200px]" : 
                    "w-[160px] h-[160px]"}`}
              >
                {/* Glow effect on hover */}
                <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
                
                <div className="flex h-full flex-col justify-between relative z-10">
                  <div className="w-12 h-12 rounded-full bg-white/10 flex items-center justify-center border border-white/20">
                    <span className="font-mono text-white text-[14px]">{tool.name[0]}</span>
                  </div>
                  
                  <div>
                    <h3 className="text-white font-medium text-[18px] mb-2">{tool.name}</h3>
                    <p className="text-white/40 text-[13px] leading-relaxed line-clamp-3">{tool.desc}</p>
                  </div>
                </div>
              </motion.div>
            </MagneticButton>
          ))}
        </div>

        {/* Connecting Lines Graphic */}
        <div className="absolute inset-0 pointer-events-none -z-10 opacity-30">
          <svg className="w-full h-full" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
                <path d="M 40 0 L 0 0 0 40" fill="none" stroke="rgba(255,255,255,0.05)" strokeWidth="1"/>
              </pattern>
            </defs>
            <rect width="100%" height="100%" fill="url(#grid)" />
          </svg>
        </div>
      </div>
    </PageChrome>
  );
}
