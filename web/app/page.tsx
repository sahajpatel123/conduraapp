"use client";

import HeroSection from "@/components/home/HeroSection";
import ProviderBanner from "@/components/home/ProviderBanner";
import BringYourOwnAI from "@/components/home/BringYourOwnAI";
import MarqueeTile from "@/components/home/MarqueeTile";
import SafetyTile from "@/components/home/SafetyTile";
import DownloadTile from "@/components/home/DownloadTile";
import Footer from "@/components/home/Footer";

export default function Home() {
  return (
    <>
      {/* Main stacked sections — navigation is handled globally by SiteDock */}
      <main id="main" className="bg-canvas">
        <HeroSection />
        <ProviderBanner />
        <BringYourOwnAI />
        <MarqueeTile />
        <SafetyTile />
        <DownloadTile />
      </main>

      {/* Global Apple Footer */}
      <Footer />
    </>
  );
}
