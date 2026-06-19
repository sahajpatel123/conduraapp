"use client";

import PageChrome from "@/components/shell/PageChrome";
import { useEffect, useState } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";
import { motion } from "motion/react";

export default function LegalPage() {
  const [html, setHtml] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  // We fetch the markdown on the client here just to allow us to do some fun text generation effects
  useEffect(() => {
    // In a real app we'd fetch the parsed HTML from an API or pass it as initial props
    // For this effect, we'll just simulate loading the classified document
    const timer = setTimeout(() => {
      setHtml(`
        <h2>End-User License Agreement (EULA)</h2>
        <p>By downloading, installing, or using Condura, you agree to be bound by the terms of this EULA.</p>
        <h3>1. License Grant</h3>
        <p>Condura grants you a revocable, non-exclusive, non-transferable, limited license to download, install and use the Application strictly in accordance with the terms of this Agreement.</p>
        <h3>2. Local-First Guarantee</h3>
        <p>You own your data. We do not transmit your local vector stores, chat history, or file system contents to our servers. Any cloud sync is end-to-end encrypted.</p>
        <h3>3. Limitation of Liability</h3>
        <p>Condura is provided "as is". We are not responsible for unintended code execution if you bypass the Gatekeeper or ignore the destructive action modals.</p>
        <h3>4. Termination</h3>
        <p>This EULA is effective until terminated by you or Condura. Your rights under this EULA will terminate automatically without notice if you fail to comply with any of its terms.</p>
      `);
      setLoading(false);
    }, 1500);
    return () => clearTimeout(timer);
  }, []);

  return (
    <div className="bg-black text-white min-h-screen">
      <PageChrome
        eyebrow="Legal"
        title="The Contract."
        description="We believe legal documents shouldn't be hidden in tiny text. Here is exactly what you agree to when you install Condura on your machine."
        badge="EULA"
      >
        <div className="mt-24 max-w-4xl mx-auto relative">
          
          {/* Decorative Security Border */}
          <div className="absolute -inset-8 border border-white/5 bg-[#030303] rounded-[40px] shadow-2xl overflow-hidden -z-10">
            <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-white/20 to-transparent" />
            <div className="absolute inset-0 bg-[url('/images/noise.png')] opacity-[0.02] mix-blend-overlay pointer-events-none" />
          </div>

          <div className="flex items-center justify-between mb-12 pb-6 border-b border-white/10">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 rounded-full border border-white/20 flex items-center justify-center">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
              </div>
              <div>
                <div className="font-mono text-xs text-white/40 uppercase tracking-widest mb-1">Document Classification</div>
                <div className="font-medium text-white">PUBLIC RELEASE</div>
              </div>
            </div>
            <div className="text-right hidden sm:block">
               <div className="font-mono text-xs text-white/40 uppercase tracking-widest mb-1">Last Updated</div>
               <div className="font-mono text-white/80">2026.06.18</div>
            </div>
          </div>

          <div className="min-h-[400px]">
            {loading ? (
              <div className="flex flex-col gap-4">
                {[...Array(6)].map((_, i) => (
                  <motion.div 
                    key={i}
                    animate={{ opacity: [0.1, 0.3, 0.1] }}
                    transition={{ repeat: Infinity, duration: 1.5, delay: i * 0.2 }}
                    className={`h-6 bg-white/10 rounded-md ${i % 2 === 0 ? 'w-3/4' : 'w-full'} ${i === 0 ? 'w-1/3 mb-4 h-8' : ''}`}
                  />
                ))}
                <div className="mt-8 font-mono text-xs text-white/30 animate-pulse">Decrypting local document...</div>
              </div>
            ) : (
              <motion.div 
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.8 }}
                className="prose prose-invert max-w-none 
                  prose-h2:text-2xl prose-h2:font-medium prose-h2:mb-8 prose-h2:border-b prose-h2:border-white/10 prose-h2:pb-4
                  prose-h3:text-lg prose-h3:font-mono prose-h3:uppercase prose-h3:tracking-widest prose-h3:text-white/60 prose-h3:mt-12
                  prose-p:text-white/70 prose-p:leading-relaxed prose-p:text-[17px]
                "
              >
                <div dangerouslySetInnerHTML={{ __html: html! }} />
              </motion.div>
            )}
          </div>

          {/* Interactive Acceptance Footer */}
          {!loading && (
            <motion.div 
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 1 }}
              className="mt-24 pt-8 border-t border-white/10 flex flex-col sm:flex-row items-center justify-between gap-6"
            >
              <p className="text-sm text-white/40 font-mono">By using the software, you accept these terms.</p>
              <button className="rounded-full bg-white text-black px-8 py-3 font-medium text-sm hover:scale-105 transition-transform">
                Acknowledge
              </button>
            </motion.div>
          )}

        </div>
      </PageChrome>
    </div>
  );
}
