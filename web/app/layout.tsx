import type { Metadata } from "next";
import { Inter, Instrument_Serif, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import Providers from "@/components/shell/Providers";
import SiteShell from "@/components/shell/SiteShell";
import BrandSurface from "@/components/shell/BrandSurface";
import { SITE } from "@/lib/site";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  display: "swap",
});
const display = Instrument_Serif({
  subsets: ["latin"],
  weight: "400",
  style: ["normal", "italic"],
  variable: "--font-display",
  display: "swap",
});
const mono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
  display: "swap",
});

export const metadata: Metadata = {
  title: `${SITE.name} — One hotkey. Every AI you own. Free.`,
  description:
    "A free desktop app that summons every AI tool on your computer with one hotkey. No account, no data leaves your machine. The conductor for your own machine.",
  metadataBase: new URL(SITE.url),
  openGraph: {
    title: `${SITE.name} — One hotkey. Every AI you own. Free.`,
    description: "Your AI tools, one hotkey. Free. Local. Private.",
    url: SITE.url,
    siteName: SITE.name,
    locale: "en_US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: SITE.name,
    description: "Your AI tools, one hotkey. Free. Local. Private.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html
      lang="en"
      className={`${inter.variable} ${display.variable} ${mono.variable}`}
    >
      <body className="surface-paper min-h-screen antialiased">
        <Providers>
          <BrandSurface />
          <SiteShell>{children}</SiteShell>
        </Providers>
      </body>
    </html>
  );
}
