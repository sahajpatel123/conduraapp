/* Soft skeleton while the next route chunk streams in — matches PageTransition tone. */
export default function Loading() {
  return (
    <div
      className="min-h-screen bg-black animate-[page-load-pulse_0.9s_ease-in-out_infinite]"
      aria-busy="true"
      aria-label="Loading page"
    >
      <div className="mx-auto max-w-4xl px-6 pt-32 pb-24">
        <div className="mb-8 h-3 w-20 rounded-full bg-white/[0.06]" />
        <div className="mb-4 h-10 w-2/3 max-w-md rounded-xl bg-white/[0.05]" />
        <div className="h-4 w-full max-w-lg rounded-lg bg-white/[0.04]" />
        <div className="mt-3 h-4 w-4/5 max-w-sm rounded-lg bg-white/[0.03]" />
      </div>
    </div>
  );
}
