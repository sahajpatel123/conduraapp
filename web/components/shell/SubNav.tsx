"use client";

export default function SubNav() {
  const handleScrollToDownload = () => {
    const el = document.getElementById("download-tile");
    if (el) {
      el.scrollIntoView({ behavior: "smooth" });
    }
  };

  return (
    <div className="sticky left-0 right-0 top-[44px] z-40 h-[52px] border-b border-white/[0.08] bg-black/80 backdrop-blur-xl transition-all duration-300">
      <div className="mx-auto flex h-full max-w-5xl items-center justify-between px-6">
        {/* Left Side: Product Name */}
        <div>
          <span className="font-body-mature text-[15px] font-semibold tracking-[0.011em] text-[#ffffff]">
            Condura
          </span>
        </div>

        {/* Right Side: Price/Version + Button */}
        <div className="flex items-center gap-4">
          <span className="font-body-mature text-[13px] text-[#a1a1aa] hidden sm:inline">
            Free
          </span>
          <button
            onClick={handleScrollToDownload}
            className="mature-button px-3 py-1 text-[12px]"
          >
            Download
          </button>
        </div>
      </div>
    </div>
  );
}
