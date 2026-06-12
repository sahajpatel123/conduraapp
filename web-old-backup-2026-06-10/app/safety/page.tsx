import type { Metadata } from "next";
import { Mail } from "lucide-react";
import { LinkButton } from "@/components/button";
import { Section, SectionHeader } from "@/components/section";
import { TrustGrid } from "@/components/trust-grid";
import { permissionRows, safetyMechanisms, securityEmail } from "@/lib/site-data";

export const metadata: Metadata = {
  title: "Safety",
  description: "Synaptic's Gatekeeper, permission model, kill switch, and audit log design.",
};

export default function SafetyPage() {
  return (
    <main id="main-content">
      <Section className="border-b border-white/10">
        <SectionHeader
          label="Safety model"
          title="The model proposes. Deterministic boundaries decide."
          lead="Synaptic is designed around explicit control because it can operate real computer surfaces. The safety model is visible before download, not buried in a footer."
        />
        <div className="mt-10">
          <TrustGrid items={safetyMechanisms} />
        </div>
      </Section>

      <Section className="border-b border-white/10">
        <div className="grid gap-8 lg:grid-cols-[0.82fr_1.18fr]">
          <SectionHeader
            label="Permissions"
            title="Each permission should map to a user-understandable purpose."
            lead="The public site should prepare users for OS permission prompts before they ever install the app."
          />
          <div className="overflow-hidden rounded-lg border border-white/10">
            <table className="w-full border-collapse text-left text-sm">
              <thead className="bg-white/[0.055] text-white">
                <tr>
                  <th className="p-4 font-semibold">Permission</th>
                  <th className="p-4 font-semibold">Purpose</th>
                  <th className="p-4 font-semibold">Timing</th>
                </tr>
              </thead>
              <tbody>
                {permissionRows.map(([permission, purpose, timing]) => (
                  <tr key={permission} className="border-t border-white/10 text-white/64">
                    <th scope="row" className="p-4 font-semibold text-white/86">
                      {permission}
                    </th>
                    <td className="p-4">{purpose}</td>
                    <td className="p-4">{timing}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </Section>

      <Section id="security-contact">
        <div className="panel p-6 sm:p-8">
          <div className="flex flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="section-label">Security contact</p>
              <h2 className="mt-3 text-2xl font-semibold text-white">Report a vulnerability or unsafe behavior.</h2>
              <p className="mt-3 max-w-2xl text-sm leading-6 text-white/62">
                Security reports should be direct, specific, and handled separately from product support.
              </p>
            </div>
            <LinkButton href={`mailto:${securityEmail}`} variant="danger">
              <Mail aria-hidden="true" size={18} />
              {securityEmail}
            </LinkButton>
          </div>
        </div>
      </Section>
    </main>
  );
}
