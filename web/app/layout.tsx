import type { Metadata } from "next";
import { Inter, Syne } from "next/font/google";
import "./globals.css";
import Providers from "@/components/shell/Providers";
import SiteShell from "@/components/shell/SiteShell";
import { SITE } from "@/lib/site";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });
const syne = Syne({ subsets: ["latin"], variable: "--font-syne" });

export const metadata: Metadata = {
  title: `${SITE.name} — One hotkey. Every AI you own. Free.`,
  description:
    "A free desktop app that summons every AI tool on your computer with one hotkey. No account, no data leaves your machine.",
  openGraph: {
    title: SITE.name,
    description: "Your AI tools, one hotkey. Free. Local. Private.",
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
