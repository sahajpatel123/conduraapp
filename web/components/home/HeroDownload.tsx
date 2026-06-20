"use client";

import { useState } from "react";
import { PLATFORMS, type PlatformKey } from "@/lib/site";
import { DOWNLOADS } from "@/lib/downloads";
import { Icon } from "@/components/motion/Icon";

export default function HeroDownload() {
  const [activePlatform, setActivePlatform] = useState<PlatformKey>("mac");

  const handleDownload = () => {
    const url = DOWNLOADS[activePlatform].primary.href;
    window.location.href = url;
  };

  return (
    <div className="w-full flex flex-col gap-6 max-w-[420px]">
      {/* Platform Selector Tabs */}
      <div className="flex p-1 rounded-xl border border-white/10 bg-[#0a0a0a] w-fit">
        {PLATFORMS.map((p) => {
          const isActive = activePlatform === p.key;
          return (
            <button
              key={p.key}
              onClick={() => setActivePlatform(p.key)}
              className={`flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm transition-colors ${
                isActive 
                  ? "bg-white/[0.08] text-white" 
                  : "text-white/40 hover:text-white/80 hover:bg-white/[0.02]"
              }`}
            >
              {p.key === "mac" && <Icon name="mac" size={14} />}
              {p.key === "windows" && <Icon name="windows" size={14} />}
              {p.key === "linux" && <Icon name="linux" size={14} />}
              {p.name}
            </button>
          );
        })}
      </div>

      {/* Buttons */}
      <div className="flex items-center gap-4">
        <button
          onClick={handleDownload}
          className="flex-1 bg-[#D97757] hover:bg-[#eb8867] text-black font-medium py-3.5 px-6 rounded-xl flex items-center justify-center gap-2 transition-colors shadow-[0_0_40px_rgba(217,119,87,0.2)]"
        >
          <Icon name="download" size={18} />
          Download for {PLATFORMS.find(p => p.key === activePlatform)?.name}
        </button>
        
        <a 
          href="https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0"
          target="_blank"
          rel="noreferrer"
          className="px-6 py-3.5 rounded-xl border border-white/10 text-white/70 hover:text-white hover:bg-white/[0.02] transition-colors flex items-center gap-2 text-sm font-medium"
        >
          Release notes
          <Icon name="arrowRight" size={14} />
        </a>
      </div>

      {/* Info Footer */}
      <div className="flex border-t border-white/10 mt-4 pt-6">
        <div className="flex-1 pr-4 border-r border-white/10">
          <div className="text-[10px] text-white/30 uppercase tracking-widest font-mono mb-1.5">Account</div>
          <div className="text-[13px] text-white/70">Not required</div>
        </div>
        <div className="flex-1 px-4 border-r border-white/10">
          <div className="text-[10px] text-white/30 uppercase tracking-widest font-mono mb-1.5">License</div>
          <div className="text-[13px] text-white/70">Free</div>
        </div>
        <div className="flex-1 pl-4">
          <div className="text-[10px] text-white/30 uppercase tracking-widest font-mono mb-1.5">Build</div>
          <div className="text-[13px] text-white/70">
            {activePlatform === 'mac' ? '.dmg installer' : activePlatform === 'windows' ? '.exe installer' : '.AppImage'}
          </div>
        </div>
      </div>
    </div>
  );
}
