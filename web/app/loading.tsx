/* Instant feedback skeleton shown while a page chunk loads.
   Prevents the "frozen old page" feeling during navigation. */
export default function Loading() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-black">
      <div
        className="h-6 w-6 animate-spin rounded-full border-2 border-white/15 border-t-white/60"
        aria-label="Loading"
        role="status"
      />
    </div>
  );
}