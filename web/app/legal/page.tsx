"use client";

import PageChrome from "@/components/shell/PageChrome";
import { motion } from "motion/react";

const SECTIONS = [
  {
    id: "acceptance",
    title: "Acceptance of Terms",
    body: `By downloading, installing, or using Condura ("the Software"), you agree to be bound by the terms of this End-User License Agreement ("EULA"). If you do not agree, do not download, install, or use the Software.`,
  },
  {
    id: "license-grant",
    title: "License Grant",
    body: `Condura Inc. grants you a revocable, non-exclusive, non-transferable, limited license to download, install, and use the Software for personal and commercial purposes on devices you own or control, subject to the terms of this Agreement. This license is per-device; you may install the Software on multiple devices you own.`,
  },
  {
    id: "restrictions",
    title: "Restrictions",
    body: `You may not: (a) redistribute, sublicense, sell, rent, or lease the Software; (b) decompile, reverse-engineer, or disassemble the Software except as permitted by applicable law; (c) remove or alter any proprietary notices; (d) use the Software to violate any applicable law or regulation; (e) circumvent any safety gate, kill switch, or audit mechanism in the Software.`,
  },
  {
    id: "local-data",
    title: "Local-First & Privacy",
    body: `The Software processes data locally on your machine. Your API keys, conversation history, vector stores, file system contents, and audit logs remain on your device unless you explicitly configure cloud features. Any cloud or sync feature is end-to-end encrypted. We do not collect telemetry, usage data, or personal information.`,
  },
  {
    id: "no-warranty",
    title: "Disclaimer of Warranty",
    body: `THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT. THE ENTIRE RISK AS TO THE QUALITY AND PERFORMANCE OF THE SOFTWARE IS WITH YOU.`,
  },
  {
    id: "limitation-liability",
    title: "Limitation of Liability",
    body: `IN NO EVENT SHALL CONDURA INC. BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR EXEMPLARY DAMAGES ARISING OUT OF OR IN CONNECTION WITH THE USE OR INABILITY TO USE THE SOFTWARE, INCLUDING BUT NOT LIMITED TO DAMAGES FOR LOSS OF PROFITS, DATA, OR DATA BREACH, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGES. THIS INCLUDES, WITHOUT LIMITATION, DAMAGES RESULTING FROM AUTOMATED ACTIONS PERFORMED BY THE SOFTWARE AT YOUR DIRECTION. You are responsible for reviewing all actions before they execute, especially destructive operations.`,
  },
  {
    id: "termination",
    title: "Termination",
    body: `This EULA is effective until terminated. Your rights terminate automatically without notice if you fail to comply with any term. Upon termination, you must cease all use of the Software and destroy all copies. Sections 4 through 10 survive termination.`,
  },
  {
    id: "changes",
    title: "Changes to This Agreement",
    body: `We may update this EULA from time to time. Material changes will be communicated through the Software. Continued use after changes take effect constitutes acceptance of the updated terms.`,
  },
  {
    id: "governing-law",
    title: "Governing Law",
    body: `This Agreement shall be governed by and construed in accordance with the laws of the State of California, without regard to its conflict of laws principles. Any legal action shall be brought exclusively in the courts of San Francisco County, California.`,
  },
  {
    id: "contact",
    title: "Contact",
    body: `For questions about this EULA, contact legal@condura.app.`,
  },
];

export default function LegalPage() {
  return (
    <div className="bg-black text-white min-h-screen">
      <PageChrome
        eyebrow="Legal"
        title="The Contract."
        description="We believe legal documents shouldn't be hidden in tiny text. Here is exactly what you agree to when you install Condura on your machine."
        badge="EULA v1"
      >
        <div className="mt-24 max-w-4xl mx-auto">
          <div className="space-y-16">
            {SECTIONS.map((section, i) => (
              <motion.section
                key={section.id}
                id={section.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.08, duration: 0.6 }}
              >
                <h2 className="text-2xl font-medium mb-6 text-white">
                  <span className="text-white/30 font-mono text-sm mr-4">
                    {String(i + 1).padStart(2, "0")}
                  </span>
                  {section.title}
                </h2>
                <p className="text-white/70 text-[17px] leading-relaxed pl-12">
                  {section.body}
                </p>
              </motion.section>
            ))}
          </div>

          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 1.2 }}
            className="mt-24 pt-8 border-t border-white/10 text-center"
          >
            <p className="text-sm text-white/40 font-mono mb-6">
              By downloading and using Condura, you accept these terms.
            </p>
            <p className="text-sm text-white/30">
              Condura Freeware EULA v1 · Last updated June 2026
            </p>
          </motion.div>
        </div>
      </PageChrome>
    </div>
  );
}
