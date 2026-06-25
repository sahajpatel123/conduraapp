"use client";

import GlobalNav from "@/components/shell/GlobalNav";
import HeroSection from "@/components/home/HeroSection";
import ManifestoOpening from "@/components/home/ManifestoOpening";
import TheConductor from "@/components/home/TheConductor";
import TheRoster from "@/components/home/TheRoster";
import TheArmor from "@/components/home/TheArmor";
import DownloadCTA from "@/components/home/DownloadCTA";
import Footer from "@/components/home/Footer";

export default function Home() {
  return (
    <div className="relative min-h-screen w-full overflow-x-hidden">
      <GlobalNav />
      <main id="main" className="relative z-10">
        <HeroSection />
        <ManifestoOpening />
        <TheConductor />
        <TheRoster />
        <TheArmor />
        <DownloadCTA />
        <Footer />
      </main>
    </div>
  );
}
