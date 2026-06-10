import type { Metadata } from "next";
import { Archivo, Geist, Geist_Mono, Instrument_Serif } from "next/font/google";
import "./globals.css";
import { Providers } from "@/components/chrome/providers";
import { Nav } from "@/components/chrome/nav";
import { Footer } from "@/components/chrome/footer";
import { Baton } from "@/components/chrome/baton";
import { SITE } from "@/lib/site";

const archivo = Archivo({
  variable: "--font-archivo",
  subsets: ["latin"],
  axes: ["wdth"],
});

const instrument = Instrument_Serif({
  variable: "--font-instrument",
  subsets: ["latin"],
  weight: "400",
  style: ["normal", "italic"],
});

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  metadataBase: new URL(SITE.url),
  title: {
    default: `${SITE.name} — ${SITE.tagline}`,
    template: `%s — ${SITE.name}`,
  },
  description: SITE.description,
  openGraph: {
    title: `${SITE.name} — ${SITE.tagline}`,
    description: SITE.description,
    url: SITE.url,
    siteName: SITE.name,
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: `${SITE.name} — ${SITE.tagline}`,
    description: SITE.description,
  },
};

/* Subpages get the light before first paint; home begins in the dark. */
const themeScript = `if(location.pathname!=="/"){document.documentElement.dataset.theme="light"}`;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      data-theme="dark"
      suppressHydrationWarning
      className={`${archivo.variable} ${instrument.variable} ${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="flex min-h-full flex-col">
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
        <a
          href="#main"
          className="annotation sr-only z-[90] bg-brass !text-ink focus:not-sr-only focus:fixed focus:top-3 focus:left-3 focus:px-4 focus:py-2"
        >
          Skip to content
        </a>
        <Providers>
          <Baton />
          <Nav />
          <main id="main" className="flex-1">
            {children}
          </main>
          <Footer />
        </Providers>
        <div className="grain" aria-hidden />
      </body>
    </html>
  );
}
