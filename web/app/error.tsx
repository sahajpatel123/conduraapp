"use client";

import { useEffect } from "react";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log the error to an error reporting service if needed
    console.error(error);
  }, [error]);

  return (
    <div className="mx-auto flex min-h-[50vh] max-w-lg flex-col items-center justify-center px-6 text-center">
      <h2 className="text-2xl font-semibold text-white">Something went wrong</h2>
      <p className="mt-3 text-sm text-white/50">
        {error.message || "An unexpected error occurred while loading this page."}
      </p>
      <button
        type="button"
        onClick={() => reset()}
        className="mature-button mt-8 rounded-full px-6 py-3 text-sm font-semibold"
      >
        Try again
      </button>
    </div>
  );
}
