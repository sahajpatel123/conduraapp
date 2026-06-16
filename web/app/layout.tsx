import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";
import { NAV_LINKS, SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: "Synaptic",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="flex min-h-screen flex-col bg-neutral-950 text-neutral-100 antialiased">
        <header className="sticky top-0 z-50 border-b border-neutral-900 bg-neutral-950/80 backdrop-blur">
          <nav className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
            <Link
              href="/"
              className="text-base font-semibold tracking-tight text-white"
            >
              {SITE.name}
            </Link>
            <div className="flex items-center gap-5 text-sm text-neutral-400">
              <Link href="/" className="transition-colors hover:text-white">
                Home
              </Link>
              {NAV_LINKS.map((l) => (
                <Link
                  key={l.href}
                  href={l.href}
                  className="transition-colors hover:text-white"
                >
                  {l.label}
                </Link>
              ))}
            </div>
          </nav>
        </header>

        <div className="flex-1">{children}</div>

        <footer className="border-t border-neutral-900">
          <div className="mx-auto flex max-w-6xl flex-col gap-4 px-6 py-10 text-sm text-neutral-500 sm:flex-row sm:items-center sm:justify-between">
            <p>© 2026 {SITE.name}</p>
            <div className="flex items-center gap-6">
              <a
                href={SITE.github}
                target="_blank"
                rel="noreferrer"
                className="transition-colors hover:text-white"
              >
                GitHub
              </a>
              <a
                href={SITE.discord}
                target="_blank"
                rel="noreferrer"
                className="transition-colors hover:text-white"
              >
                Discord
              </a>
              <Link href="/legal" className="transition-colors hover:text-white">
                Legal
              </Link>
            </div>
          </div>
        </footer>
      </body>
    </html>
  );
}
