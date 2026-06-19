"use client";

import PageChrome from "@/components/shell/PageChrome";
import { motion, useScroll, useTransform } from "motion/react";
import MagneticButton from "@/components/motion/MagneticButton";
import { useRef, useEffect, useState } from "react";

const ECOSYSTEM_TOOLS = [
  { name: "React", desc: "AST patching and DOM analysis", size: "lg", color: "from-[#61dafb] to-blue-500" },
  { name: "Next.js", desc: "Routing and SSR management", size: "md", color: "from-white to-gray-500" },
  { name: "TypeScript", desc: "Type-safe deterministic execution", size: "lg", color: "from-[#3178c6] to-blue-600" },
  { name: "Rust", desc: "Core engine performance", size: "xl", color: "from-[#dea584] to-orange-600" },
  { name: "Tailwind", desc: "Utility-first design injection", size: "sm", color: "from-[#38bdf8] to-cyan-600" },
  { name: "Git", desc: "Version control sync", size: "md", color: "from-[#f14e32] to-red-600" },
  { name: "Docker", desc: "Containerized isolation", size: "lg", color: "from-[#2496ed] to-blue-400" },
  { name: "Node.js", desc: "V8 runtime operations", size: "xl", color: "from-[#339933] to-green-600" },
  { name: "PostgreSQL", desc: "Vector state memory", size: "md", color: "from-[#336791] to-blue-700" },
];

export default function EcosystemPage() {
  const scrollRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({ target: scrollRef, offset: ["start end", "end start"] });
  const scale = useTransform(scrollYProgress, [0, 0.5, 1], [0.8, 1, 0.8]);
  const opacity = useTransform(scrollYProgress, [0, 0.5, 1], [0.3, 1, 0.3]);

  const [typedCode, setTypedCode] = useState("");
  const codeSnippet = `// condura.plugin.ts
import { Gatekeeper } from '@condura/core';

export default function MyCustomPlugin() {
  Gatekeeper.registerAction('deploy:staging', async (ctx) => {
    // 1. Analyze diff
    const ast = await ctx.workspace.getAST();
    
    // 2. Request explicit human permission
    await ctx.promptUser({
      title: 'Deploy to Staging?',
      dangerLevel: 'high'
    });
    
    // 3. Execute
    return ctx.shell.exec('npm run deploy');
  });
}`;

  useEffect(() => {
    let i = 0;
    const interval = setInterval(() => {
      setTypedCode(codeSnippet.slice(0, i));
      i++;
      if (i > codeSnippet.length) clearInterval(interval);
    }, 20);
    return () => clearInterval(interval);
  }, [codeSnippet]);

  return (
    <div className="bg-black text-white min-h-screen">
      <PageChrome
        eyebrow="Ecosystem"
        title="Plugs into everything you use."
        description="Condura doesn't ask you to change your stack. It uses native AST parsers, LSP integrations, and raw terminal APIs to work exactly where you already work. It's a layer over your OS, not a walled garden."
        badge="Integration"
      >
        {/* --- SECTION 1: Dynamic Tool Grid --- */}
        <div className="mt-24 min-h-[60vh] relative flex flex-col items-center">
          <div className="w-full flex flex-wrap justify-center gap-6 p-8 relative z-10">
            {ECOSYSTEM_TOOLS.map((tool, i) => (
              <MagneticButton key={tool.name}>
                <motion.div
                  initial={{ opacity: 0, scale: 0.8, y: 20 }}
                  animate={{ opacity: 1, scale: 1, y: 0 }}
                  transition={{ delay: i * 0.1, type: "spring", stiffness: 100, damping: 15 }}
                  className={`relative overflow-hidden rounded-[24px] border border-white/10 bg-[#0a0a0a] shadow-xl p-6 group cursor-pointer hover:bg-white/[0.02] transition-colors
                    ${tool.size === "xl" ? "w-[300px] h-[300px]" : tool.size === "lg" ? "w-[240px] h-[240px]" : tool.size === "md" ? "w-[200px] h-[200px]" : "w-[160px] h-[160px]"}`}
                >
                  <div className={`absolute inset-0 bg-gradient-to-br ${tool.color} opacity-0 group-hover:opacity-10 transition-opacity duration-500`} />
                  <div className="flex h-full flex-col justify-between relative z-10">
                    <div className="w-12 h-12 rounded-full bg-white/5 flex items-center justify-center border border-white/10 group-hover:border-white/30 transition-colors">
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

          <div className="absolute inset-0 pointer-events-none z-0 opacity-20">
            <svg className="w-full h-full" xmlns="http://www.w3.org/2000/svg">
              <defs>
                <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
                  <path d="M 40 0 L 0 0 0 40" fill="none" stroke="rgba(255,255,255,0.1)" strokeWidth="1"/>
                </pattern>
              </defs>
              <rect width="100%" height="100%" fill="url(#grid)" />
            </svg>
          </div>
        </div>

        {/* --- SECTION 2: How It Connects (Story & Node Graph) --- */}
        <motion.div ref={scrollRef} style={{ scale, opacity }} className="mt-40 max-w-5xl mx-auto px-8 relative">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-5xl font-semibold tracking-tight mb-4 text-white">Three layers of native integration.</h2>
            <p className="text-lg text-white/50 max-w-2xl mx-auto">Condura doesn't rely on brittle regex scraping. It hooks directly into the core protocols your development environment already uses.</p>
          </div>
          
          <div className="grid md:grid-cols-3 gap-6">
            <div className="p-8 rounded-[32px] bg-[#050505] border border-white/10">
              <div className="w-10 h-10 mb-6 rounded-full bg-white/10 flex items-center justify-center text-xl">1</div>
              <h3 className="text-xl font-medium text-white mb-3">LSP Protocol</h3>
              <p className="text-white/40 text-sm leading-relaxed">It speaks the Language Server Protocol natively. When it edits TypeScript, it knows immediately if it broke a type definition without running a build.</p>
            </div>
            <div className="p-8 rounded-[32px] bg-[#050505] border border-white/10">
               <div className="w-10 h-10 mb-6 rounded-full bg-white/10 flex items-center justify-center text-xl">2</div>
              <h3 className="text-xl font-medium text-white mb-3">AST Parsing</h3>
              <p className="text-white/40 text-sm leading-relaxed">Instead of text-replace, it modifies the Abstract Syntax Tree. It understands the structure of your React components as logic, not just strings.</p>
            </div>
            <div className="p-8 rounded-[32px] bg-[#050505] border border-white/10">
               <div className="w-10 h-10 mb-6 rounded-full bg-white/10 flex items-center justify-center text-xl">3</div>
              <h3 className="text-xl font-medium text-white mb-3">Terminal PTY</h3>
              <p className="text-white/40 text-sm leading-relaxed">It spawns real pseudo-terminals. It can run tests, start servers, and monitor standard output exactly like a human engineer would.</p>
            </div>
          </div>
        </motion.div>

        {/* --- SECTION 3: Extensibility (Live Code Typing) --- */}
        <div className="mt-40 mb-32 max-w-5xl mx-auto px-8 flex flex-col lg:flex-row gap-16 items-center">
          <div className="flex-1">
            <h2 className="text-3xl md:text-5xl font-semibold tracking-tight mb-6 text-white">Extend it endlessly.</h2>
            <p className="text-lg text-white/50 leading-relaxed mb-6">
              Need Condura to understand your proprietary deployment system? Or talk to your internal API? 
              <br/><br/>
              Write a simple TypeScript plugin. The Gatekeeper guarantees that even custom plugins must still request explicit user permission for destructive actions.
            </p>
            <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.04] px-4 py-2 font-mono text-[13px] text-white/70">
              <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
              Plugin API v1 (Stable)
            </div>
          </div>
          
          <div className="flex-1 w-full relative">
            <div className="absolute -inset-4 bg-gradient-to-r from-blue-500/20 to-purple-500/20 blur-3xl opacity-20 pointer-events-none" />
            <div className="rounded-2xl border border-white/10 bg-[#0d0d0d] shadow-2xl overflow-hidden relative z-10">
               <div className="h-10 border-b border-white/5 bg-white/[0.02] flex items-center px-4 gap-2">
                <div className="w-3 h-3 rounded-full bg-[#ff5f57]" />
                <div className="w-3 h-3 rounded-full bg-[#febc2e]" />
                <div className="w-3 h-3 rounded-full bg-[#28c840]" />
                <span className="ml-auto font-mono text-[10px] text-white/30">condura.plugin.ts</span>
              </div>
              <div className="p-6 overflow-x-auto min-h-[300px]">
                <pre className="font-mono text-[13px] leading-[1.7] text-white/80">
                  <code>
                    {typedCode}
                    <motion.span animate={{ opacity: [1, 0] }} transition={{ repeat: Infinity, duration: 0.8 }} className="w-2 h-4 bg-white/80 inline-block ml-1 align-middle" />
                  </code>
                </pre>
              </div>
            </div>
          </div>
        </div>

      </PageChrome>
    </div>
  );
}
