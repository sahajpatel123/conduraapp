import type { Metadata } from "next";
import { Inter, Syne } from "next/font/google";
import "./globals.css";
import Providers from "@/components/shell/Providers";
import SiteShell from "@/components/shell/SiteShell";
import { SITE } from "@/lib/site";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });
const syne = Syne({ subsets: ["latin"], variable: "--font-syne" });

export const metadata: Metadata = {
  title: `${SITE.name} — AI on your computer. Free.`,
  description:
    "A local-first desktop agent that appears from your OS, orchestrates every AI tool you own, and stops at deterministic safety gates.",
  openGraph: {
    title: SITE.name,
    description: "AI on your computer. Free.",
    url: SITE.url,
    siteName: SITE.name,
    locale: "en_US",
    type: "website",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={`${inter.variable} ${syne.variable}`}>
      <body className="min-h-screen bg-void text-fg antialiased">
        <Providers>
          <SiteShell>{children}</SiteShell>
        </Providers>
      </body>
    </html>
  );
}
