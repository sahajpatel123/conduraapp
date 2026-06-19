"use client";

import { useState } from "react";
import { PLATFORMS, type PlatformKey } from "@/lib/site";
import { DOWNLOADS } from "@/lib/downloads";

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
              {p.key === "mac" && (
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 20.94c1.5 0 2.75 1.06 4 1.06 3 0 6-8 6-12.22A4.91 4.91 0 0 0 17 5c-2.22 0-4 1.44-5 1.44C9.78 6.44 8 5 5.78 5 3 5 0 7.46 0 12.69c0 4.22 3 12.22 6 12.22 1.25 0 2.5-1.06 4-1.06s2.75 1.06 4 1.06Z"/></svg>
              )}
              {p.key === "windows" && (
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
              )}
              {p.key === "linux" && (
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 2C8 2 5 6 5 11c0 6 2 11 7 11s7-5 7-11c0-5-3-9-7-9z"/><circle cx="9" cy="9" r="1"/><circle cx="15" cy="9" r="1"/></svg>
              )}
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
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M12 5v14M19 12l-7 7-7-7"/></svg>
          Download for {PLATFORMS.find(p => p.key === activePlatform)?.name}
        </button>
        
        <a 
          href="https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0"
          target="_blank"
          rel="noreferrer"
          className="px-6 py-3.5 rounded-xl border border-white/10 text-white/70 hover:text-white hover:bg-white/[0.02] transition-colors flex items-center gap-2 text-sm font-medium"
        >
          Release notes
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
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
