"use client";

const ShieldIcon = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>
);

const GUARDS = [
  {
    numeral: "01",
    title: "Twin-Snapshot Verification",
    description: "Captures accessibility tree layouts immediately before and after action planning. If the screen changes, execution aborts instantly. Prevents blind click actions.",
  },
  {
    numeral: "02",
    title: "Deterministic Gatekeeper",
    description: "Policy evaluations are run against a static, local YAML config in pure Go code — never passed through model logic. Bypasses are structurally impossible.",
  },
  {
    numeral: "03",
    title: "Runtime Watchdog",
    description: "Monitors loop detection, speed thresholds, and total duration constraints. Halts execution automatically and alerts the operator.",
  },
];

export default function SafetyTile() {
  return (
    <section id="safety-tile" className="relative w-full bg-[#000000] py-[140px] px-6 text-[#ffffff] overflow-hidden border-t border-white/[0.08]">
      <div className="mx-auto w-full max-w-5xl">
        
        <div className="max-w-3xl mb-16 text-center mx-auto">
          <div className="flex justify-center mb-6 text-white/50">
            <ShieldIcon />
          </div>
          <h2 className="font-hero-display">
            Actions that <br /> cannot be undone.
          </h2>
          <p className="mt-6 font-lead-airy max-w-2xl mx-auto">
            Automating user environments is a survival problem, not an accuracy test. Condura guards your files, networks, and system configuration with sandboxed runtimes and deterministic filters.
          </p>
        </div>

        {/* Abstract Unsplash Glass Container */}
        <div className="mature-panel relative mb-16 aspect-[16/9] w-full overflow-hidden rounded-2xl md:aspect-[21/9]">
          <div 
            className="absolute inset-0 bg-cover bg-center opacity-50 mix-blend-screen"
            style={{ backgroundImage: "url('https://images.unsplash.com/photo-1604871000636-074fa5117945?q=80&w=2000&auto=format&fit=crop')" }}
          />
          <div className="absolute inset-0 bg-gradient-to-t from-black via-black/20 to-transparent" />
        </div>

        {/* Minimal 3-Column List */}
        <div className="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
          {GUARDS.map((guard) => (
            <div key={guard.title} className="flex flex-col rounded-2xl border border-white/[0.08] bg-white/[0.03] p-6">
              <span className="font-mono text-[13px] text-white/50 mb-3">
                {guard.numeral}
              </span>
              <h3 className="font-body-mature text-[16px] font-semibold text-[#ffffff] mb-3">
                {guard.title}
              </h3>
              <p className="font-body-mature text-[#a1a1aa]">
                {guard.description}
              </p>
            </div>
          ))}
        </div>

      </div>
    </section>
  );
}
