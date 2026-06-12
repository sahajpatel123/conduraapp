import type { Metadata } from "next";
import { LegalPage } from "@/components/legal-page";
import { supportEmail } from "@/lib/site-data";

export const metadata: Metadata = {
  title: "Terms",
  description: "Synaptic terms and EULA foundation for public site.",
};

const sections = [
  {
    id: "license",
    title: "Freeware binary",
    body: [
      "Synaptic is planned as a free downloadable desktop application. The binary is intended to be free for personal and commercial use under the Synaptic Freeware EULA.",
      "The source code remains proprietary unless a separate written agreement says otherwise.",
    ],
  },
  {
    id: "downloads",
    title: "Downloads and releases",
    body: [
      "Installer buttons on the site are currently visual placeholders. They should only be wired once release artifacts are signed, checksummed, and ready for public distribution.",
      "Users should be able to download without creating an account.",
    ],
  },
  {
    id: "responsibility",
    title: "Responsible use",
    body: [
      "Synaptic can operate computer surfaces when users grant permissions. Users are responsible for reviewing approvals and using the agent within applicable laws and policies.",
      "The product design includes Gatekeeper approvals, audit logs, and stop controls to keep user control visible.",
    ],
  },
  {
    id: "contact",
    title: "Contact",
    body: [`Terms and license questions can be sent to ${supportEmail}. This page is a foundation shell and should be reviewed before public release.`],
  },
];

export default function TermsPage() {
  return (
    <LegalPage
      title="Terms"
      lead="These terms establish the public site structure for Synaptic's free binary, proprietary source model, and release flow."
      sections={sections}
    />
  );
}
