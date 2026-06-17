import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { SITE } from "@/lib/site";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });

export const metadata: Metadata = {
  title: `${SITE.name} — AI on your computer. Free.`,
  description:
    "A ghost that lives inside your computer. Press a hotkey. It appears. Orchestrates every AI tool you have. Then vanishes. Free forever.",
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
    <html lang="en" className={inter.variable}>
      <body className="min-h-screen antialiased bg-[#050505] text-[#e5e5e5]">
        {children}
      </body>
    </html>
  );
}
