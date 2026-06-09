import Link from "next/link";
import { NAV_LINKS, SITE } from "@/lib/site";

/*
  The final bar line. Static by design — the page has finished playing.
*/
export function Footer() {
  return (
    <footer className="border-t border-line">
      <div className="mx-auto max-w-6xl px-5 py-16 md:px-8">
        <div className="flex flex-col gap-12 md:flex-row md:items-end md:justify-between">
          <div>
            <p className="display-italic text-4xl md:text-5xl">fin.</p>
            <p className="mt-4 max-w-sm text-sm leading-relaxed text-ivory-dim">
              {SITE.name} is a free, local-first conductor for every AI on your
              computer. Proprietary source, free binary, no telemetry — ever.
            </p>
          </div>
          <nav aria-label="Footer" className="flex flex-col items-start gap-2 md:items-end">
            {[{ href: "/", label: "Overture" }, ...NAV_LINKS].map((link) => (
              <Link key={link.href} href={link.href} className="prose-link text-sm">
                {link.label}
              </Link>
            ))}
          </nav>
        </div>
        <div className="rule mt-14" />
        <div className="mt-6 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
          <p className="annotation">© 2026 The Synaptic Project</p>
          <p className="annotation">
            “The agent is a guest, not an owner.” — Invariant VI
          </p>
        </div>
      </div>
    </footer>
  );
}
