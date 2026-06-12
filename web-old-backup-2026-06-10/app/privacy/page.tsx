import type { Metadata } from "next";
import { LegalPage } from "@/components/legal-page";
import { privacyEmail } from "@/lib/site-data";

export const metadata: Metadata = {
  title: "Privacy",
  description: "Synaptic privacy principles for local-first desktop use.",
};

const sections = [
  {
    id: "local-first",
    title: "Local-first principle",
    body: [
      "Synaptic is designed so core desktop use can happen without a website account. Memory, audit records, API keys, and local configuration should remain on the user's computer unless the user explicitly enables a connected feature.",
      "The public website should explain what leaves the device before users download the app, especially because desktop agents require sensitive OS permissions.",
    ],
  },
  {
    id: "account",
    title: "Optional account",
    body: [
      "Browser sign-in is intended for dashboard, device management, Skills Hub, support, release notifications, and related services. It should not be required for public download or local-first desktop use.",
      "Authentication should happen in the browser. The desktop app should not render password forms or collect unnecessary account data.",
    ],
  },
  {
    id: "keys",
    title: "Model keys and providers",
    body: [
      "Users may configure local models or provider credentials. The intended product boundary is that keys are stored locally and used only for the provider or model the user configured.",
      "The site should not imply that Synaptic stores user model keys in a cloud account.",
    ],
  },
  {
    id: "contact",
    title: "Privacy contact",
    body: [`Questions about privacy can be sent to ${privacyEmail}. This page is a foundation shell and should be finalized before public launch.`],
  },
];

export default function PrivacyPage() {
  return (
    <LegalPage
      title="Privacy"
      lead="Synaptic's privacy posture starts with a simple rule: local desktop use should not require surrendering data to the website."
      sections={sections}
    />
  );
}
