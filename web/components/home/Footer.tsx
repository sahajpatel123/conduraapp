"use client";

import Link from "next/link";
import { SITE } from "@/lib/site";

export default function Footer() {
  return (
    <footer className="border-t border-white/[0.06] bg-[#050505]">
      <div className="mx-auto flex max-w-5xl flex-col items-center justify-between gap-6 px-6 py-10 sm:flex-row">
        <div className="flex items-center gap-6 text-[13px] text-white/25">
          <span className="font-semibold text-white/40">{SITE.name}</span>
          <span>© 2026</span>
        </div>
        <div className="flex items-center gap-6 text-[13px] text-white/25">
          <Link href="/download" className="transition-colors hover:text-white/60">
            Download
          </Link>
          <Link href="/manifesto" className="transition-colors hover:text-white/60">
            Manifesto
          </Link>
          <Link href="/legal" className="transition-colors hover:text-white/60">
            Legal
          </Link>
          <a
            href={SITE.github}
            target="_blank"
            rel="noreferrer"
            className="transition-colors hover:text-white/60"
          >
            GitHub
          </a>
          <a
            href={SITE.discord}
            target="_blank"
            rel="noreferrer"
            className="transition-colors hover:text-white/60"
          >
            Discord
          </a>
        </div>
      </div>
    </footer>
  );
}
