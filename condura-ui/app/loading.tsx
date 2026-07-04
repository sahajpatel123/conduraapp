/* Soft skeleton while the next route chunk streams in — paper edition. */
export default function Loading() {
  return (
    <div
      className="surface-paper min-h-screen"
      aria-busy="true"
      aria-label="Loading page"
    >
      <div className="paper-grain absolute inset-0" />
      <div className="mx-auto max-w-[1100px] px-6 pb-24 pt-36 sm:pt-44">
        {/* eyebrow */}
        <div className="mb-6 h-3 w-24 animate-pulse rounded-full bg-[rgba(20,17,11,0.12)]" />
        {/* title two lines */}
        <div className="mb-3 h-12 w-2/3 max-w-md animate-pulse rounded-xl bg-[rgba(20,17,11,0.10)]" />
        <div className="mb-7 h-12 w-1/2 max-w-xs animate-pulse rounded-xl bg-[rgba(11,61,46,0.14)]" />
        {/* thread divider */}
        <div className="mb-10 h-px w-full max-w-[520px] animate-pulse bg-[rgba(20,17,11,0.14)]" />
        {/* body lines */}
        <div className="h-4 w-full max-w-lg animate-pulse rounded-lg bg-[rgba(20,17,11,0.08)]" />
        <div className="mt-3 h-4 w-4/5 max-w-sm animate-pulse rounded-lg bg-[rgba(20,17,11,0.06)]" />
      </div>
    </div>
  );
}
