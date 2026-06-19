import type { Metadata } from "next";
import PageChrome from "@/components/shell/PageChrome";
import { readRepoMarkdown } from "@/lib/markdown";
import { SITE } from "@/lib/site";
import FadeInStagger from "@/components/motion/FadeInStagger";

export const metadata: Metadata = {
  title: `Changelog · ${SITE.name}`,
  description: "Notable changes to Condura, release by release.",
};

export default async function ChangelogPage() {
  const html = await readRepoMarkdown("CHANGELOG.md");

  return (
    <div className="bg-black text-white min-h-screen">
      <PageChrome
        eyebrow="Changelog"
        title="Constant velocity."
        description="We ship improvements to Condura every single week. From core engine performance upgrades to new native integrations. Here is the history of what we've built."
        badge="Updates"
      >
        
        {/* Decorative timeline graphic */}
        <div className="relative mt-24">
          <div className="absolute left-[27px] top-0 bottom-0 w-[2px] bg-gradient-to-b from-white/20 via-white/5 to-transparent hidden md:block" />
          
          {html ? (
            <div className="relative z-10 pl-0 md:pl-16">
              <FadeInStagger>
                {/* 
                  Using prose styling to make the injected HTML look like a timeline.
                  We target h2 and h3 elements within prose-md to align them with the timeline.
                */}
                <div className="prose prose-invert max-w-none
                  prose-h2:text-3xl prose-h2:font-semibold prose-h2:tracking-tight prose-h2:mt-16 prose-h2:mb-8 prose-h2:relative
                  prose-h2:before:content-[''] prose-h2:before:absolute prose-h2:before:w-4 prose-h2:before:h-4 prose-h2:before:bg-black prose-h2:before:border-4 prose-h2:before:border-white prose-h2:before:rounded-full prose-h2:before:-left-[71px] prose-h2:before:top-1.5 md:prose-h2:before:block prose-h2:before:hidden
                  
                  prose-p:text-white/60 prose-p:leading-relaxed prose-p:text-lg
                  prose-li:text-white/60 prose-li:text-lg
                  prose-a:text-white prose-a:underline prose-a:underline-offset-4 hover:prose-a:text-white/80
                  prose-code:bg-white/[0.05] prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded-md prose-code:text-white/80 prose-code:before:content-none prose-code:after:content-none
                ">
                  <article dangerouslySetInnerHTML={{ __html: html }} />
                </div>
              </FadeInStagger>
            </div>
          ) : (
            <div className="relative z-10 pl-0 md:pl-16 mt-16 border border-white/10 bg-white/[0.02] rounded-3xl p-12 text-center">
              <div className="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center mx-auto mb-6">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-white/40"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
              </div>
              <h3 className="text-2xl font-medium text-white mb-2">No local changelog found</h3>
              <p className="text-white/45 text-lg mb-8 max-w-md mx-auto">
                We couldn't locate the CHANGELOG.md file in the current build environment.
              </p>
              <a 
                className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.04] px-6 py-3 font-medium text-white hover:bg-white/[0.08] transition-colors" 
                href="https://github.com/sahajpatel123/conduraapp/releases"
                target="_blank"
                rel="noreferrer"
              >
                View on GitHub
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
              </a>
            </div>
          )}
        </div>

      </PageChrome>
    </div>
  );
}
