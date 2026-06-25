"use client";

import GlobalNav from "@/components/shell/GlobalNav";
import HeroSection from "@/components/home/HeroSection";

export default function Home() {
  return (
    <div className="h-screen w-screen overflow-hidden bg-black text-white selection:bg-white/30">
      <GlobalNav />
      
      <main id="main" className="relative h-full w-full">
        <HeroSection />
      </main>
    </div>
  );
}