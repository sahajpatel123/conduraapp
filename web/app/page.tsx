"use client";

import GlobalNav from "@/components/shell/GlobalNav";
import HeroSection from "@/components/home/HeroSection";
import ProviderBanner from "@/components/home/ProviderBanner";
import BringYourOwnAI from "@/components/home/BringYourOwnAI";
import SafetyTile from "@/components/home/SafetyTile";
import DownloadTile from "@/components/home/DownloadTile";
import Footer from "@/components/home/Footer";

export default function Home() {
  return (
    <>
      {/* Floating Pill Navigation */}
      <GlobalNav />

      {/* Main stacked sections */}
      <main id="main" className="bg-canvas">
        <HeroSection />
        <ProviderBanner />
        <BringYourOwnAI />
        <SafetyTile />
        <DownloadTile />
      </main>

      {/* Global Apple Footer */}
      <Footer />
    </>
  );
}