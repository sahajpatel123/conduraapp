import type { Metadata } from "next";
import { CriticalStyles } from "@/components/critical-styles";
import { SiteFooter } from "@/components/site-footer";
import { SiteHeader } from "@/components/site-header";
import "./globals.css";

export const metadata: Metadata = {
  title: {
    default: "Synaptic - The AI conductor for your computer",
    template: "%s - Synaptic",
  },
  description:
    "Synaptic is a free, local-first desktop AI agent with Gatekeeper safety, audit logs, voice, and OS-native command surfaces.",
  metadataBase: new URL("https://synaptic.app"),
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <head>
        <CriticalStyles />
      </head>
      <body>
        <SiteHeader />
        {children}
        <SiteFooter />
      </body>
    </html>
  );
}
