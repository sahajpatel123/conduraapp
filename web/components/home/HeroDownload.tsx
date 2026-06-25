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
    <div className="w-full flex flex-col items-center gap-8 max-w-[500px]">
      {/* Platform Selector Tabs - Premium Glass */}
      <div className="flex p-1.5 rounded-full border border-white/10 bg-white/[0.03] backdrop-blur-md shadow-2xl w-fit">
        {PLATFORMS.map((p) => {
          const isActive = activePlatform === p.key;
          return (
            <button
              key={p.key}
              onClick={() => setActivePlatform(p.key)}
              className={`flex items-center justify-center gap-2 px-6 py-2 rounded-full text-sm font-medium transition-all duration-300 ${
                isActive 
                  ? "bg-white text-black shadow-[0_4px_12px_rgba(255,255,255,0.3)]" 
                  : "text-white/40 hover:text-white/90 hover:bg-white/[0.06]"
              }`}
            >
              {p.key === "mac" && <Icon name="mac" size={16} />}
              {p.key === "windows" && <Icon name="windows" size={16} />}
              {p.key === "linux" && <Icon name="linux" size={16} />}
              {p.name}
            </button>
          );
        })}
      </div>

      {/* Buttons */}
      <div className="flex flex-col sm:flex-row items-center gap-4 w-full">
        <button
          onClick={handleDownload}
          className="flex-1 bg-gradient-to-r from-[#00DFD8] to-[#007CF0] hover:from-[#33e8e2] hover:to-[#3399ff] text-white font-semibold py-4 px-8 rounded-full flex items-center justify-center gap-3 transition-all duration-300 shadow-[0_8px_30px_rgba(0,124,240,0.4)] hover:shadow-[0_12px_40px_rgba(0,124,240,0.6)] transform hover:-translate-y-1"
        >
          <Icon name="download" size={20} />
          Download Condura
        </button>
        
        <a 
          href="https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0"
          target="_blank"
          rel="noreferrer"
          className="px-8 py-4 rounded-full border border-white/20 text-white/80 hover:text-white hover:bg-white/[0.05] hover:border-white/40 transition-all duration-300 flex items-center gap-2 text-sm font-medium backdrop-blur-sm"
        >
          View Source
          <Icon name="arrowRight" size={16} />
        </a>
      </div>
    </div>
  );
}
