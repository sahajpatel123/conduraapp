"use client";

import GlobalNav from "@/components/GlobalNav";
import CommandPalette from "@/components/CommandPalette";
import HeroSection from "@/components/home/HeroSection";
import HowItFeels from "@/components/home/HowItFeels";
import Stats from "@/components/home/Stats";
import TrustMarquee from "@/components/home/TrustMarquee";
import Demo from "@/components/home/Demo";
import CTASection from "@/components/home/CTASection";
import Footer from "@/components/home/Footer";

export default function Home() {
  return (
    <>
      <GlobalNav />
      <CommandPalette />
      <main className="bg-[#050505]">
        <HeroSection />
        <HowItFeels />
        <Stats />
        <TrustMarquee />
        <Demo />
        <CTASection />
      </main>
      <Footer />
    </>
  );
}
