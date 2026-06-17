"use client";

import { useState } from "react";

const WavesIcon = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M2 12h4l2-9 5 18 3-9h6"></path></svg>
);
const ServerIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect><rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect><line x1="6" y1="6" x2="6.01" y2="6"></line><line x1="6" y1="18" x2="6.01" y2="18"></line></svg>
);

const WAVES = [
  {
    title: "Wave 1: Research",
    agents: [
      { name: "Ollama (local)", role: "Code analyzer", time: "120ms" },
      { name: "Claude Code", role: "AST Parser", time: "340ms" },
    ],
  },
  {
    title: "Wave 2: Modification",
    agents: [
      { name: "Codex", role: "Write layout.tsx", time: "410ms" },
      { name: "Antigravity", role: "Refactor core.go", time: "running" },
    ],
  },
];

export default function OrchestrationTile() {
  const [activeWave, setActiveWave] = useState(1);

  return (
    <section id="orchestration-tile" className="relative w-full bg-[#000000] py-[140px] px-6 text-white overflow-hidden border-t border-white/[0.08]">
      <div className="mx-auto w-full max-w-5xl">
        
        {/* Unsplash Abstract Image Container */}
        <div className="mature-panel relative mb-16 aspect-[21/9] w-full overflow-hidden rounded-2xl">
          <div 
            className="absolute inset-0 bg-cover bg-center opacity-40 mix-blend-screen"
            style={{ backgroundImage: "url('https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=2000&auto=format&fit=crop')" }}
          />
          <div className="absolute inset-0 bg-gradient-to-t from-black via-black/20 to-transparent" />
          <div className="absolute bottom-6 left-6 right-6 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full border border-white/10 bg-white/[0.05]">
                <WavesIcon />
              </div>
              <span className="font-body-mature text-[16px] text-white">Parallel Processing Matrix</span>
            </div>
          </div>
        </div>

        <div className="grid lg:grid-cols-2 gap-16 items-start">
          
          <div className="max-w-xl">
            <h2 className="font-hero-display text-left">
              A conductor. <br /> Not a wrapper.
            </h2>
            <p className="mt-6 font-lead-airy text-left">
              Summon multiple sub-agents in parallel. Condura decomposes tasks into concurrent waves, maintains strict SQLite file-locks, tracks code tokens, and pipes outcomes to visual execution logs.
            </p>
          </div>

          <div className="mature-panel flex min-h-[380px] flex-col justify-between rounded-2xl p-8">
            <div className="flex gap-4 mb-8">
              {WAVES.map((wave, idx) => (
                <button
                  key={wave.title}
                  onClick={() => setActiveWave(idx)}
                  className={`flex-1 cursor-pointer rounded-xl p-4 text-left transition-colors ${
                    activeWave === idx
                      ? "bg-white/[0.10] text-white"
                      : "border border-white/[0.08] bg-white/[0.03] text-[#a1a1aa] hover:bg-white/[0.05] hover:text-white"
                  }`}
                >
                  <div className="font-body-mature text-[14px] font-medium leading-tight">
                    {wave.title}
                  </div>
                </button>
              ))}
            </div>

            <div className="grid grid-cols-1 gap-4">
              {WAVES[activeWave].agents.map((agent) => {
                const isActive = agent.time === "running";
                return (
                  <div
                    key={agent.name}
                    className={`rounded-xl p-4 ${
                      isActive
                        ? "border border-white/[0.10] bg-white/[0.06]"
                        : "border border-white/[0.08] bg-white/[0.03]"
                    }`}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-body-mature text-[15px] font-medium text-[#ffffff] flex items-center gap-2">
                        <ServerIcon /> {agent.name}
                      </span>
                      {isActive && (
                        <span className="text-[12px] text-white/50 font-mono animate-pulse">Processing...</span>
                      )}
                    </div>
                    <p className="font-body-mature text-[14px] text-[#a1a1aa] mb-2 pl-6">{agent.role}</p>
                  </div>
                );
              })}
            </div>
          </div>

        </div>
      </div>
    </section>
  );
}
