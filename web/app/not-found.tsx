import Link from "next/link";

export default function NotFound() {
  return (
    <div className="min-h-screen bg-black flex flex-col items-center justify-center px-6">
      <div className="text-center max-w-md">
        <div className="font-mono text-[72px] font-semibold text-white/10 mb-4">
         404
        </div>
        <h1 className="text-2xl font-medium text-white mb-3">
          Page not found
        </h1>
        <p className="text-[#a1a1aa] text-[15px] mb-10 leading-relaxed">
          The page you are looking for doesn't exist or has been moved.
        </p>
        <Link
          href="/"
          className="inline-flex items-center gap-2 rounded-full border border-white/[0.15] bg-white/[0.06] px-6 py-3 font-body-mature text-[14px] text-white transition-colors hover:bg-white/[0.10]"
        >
          Back to home
        </Link>
      </div>
    </div>
  );
}
