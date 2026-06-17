"use client";

import Link from "next/link";
import { SITE, NAV_LINKS, TOOL_ROSTER } from "@/lib/site";

export default function Footer() {
  return (
    <footer className="w-full bg-[#000000] py-16 px-6 border-t border-white/[0.08] relative z-10 select-none">
      <div className="mx-auto w-full max-w-5xl">
        
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 md:gap-12">
          
          <div className="col-span-2 md:col-span-1">
            <h4 className="font-body-mature font-medium text-[#ffffff] text-[14px] mb-4">
              Condura
            </h4>
            <p className="font-body-mature text-[#a1a1aa] text-[13px] leading-relaxed pr-8">
              AI on your computer. Yours alone. A local-first desktop agent that orchestrates your intelligence completely offline.
            </p>
          </div>

          <div>
            <h4 className="font-body-mature font-medium text-[#ffffff] text-[14px] mb-4">
              Integrations
            </h4>
            <ul className="flex flex-col gap-2 font-body-mature text-[13px] text-[#a1a1aa]">
              {TOOL_ROSTER.map((tool) => (
                <li key={tool}>
                  <span className="hover:text-[#ffffff] transition-colors cursor-default">
                    {tool} Agent
                  </span>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h4 className="font-body-mature font-medium text-[#ffffff] text-[14px] mb-4">
              Project
            </h4>
            <ul className="flex flex-col gap-2 font-body-mature text-[13px] text-[#a1a1aa]">
              {NAV_LINKS.map((link) => (
                <li key={link.href}>
                  <Link href={link.href} className="hover:text-[#ffffff] transition-colors">
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h4 className="font-body-mature font-medium text-[#ffffff] text-[14px] mb-4">
              Support
            </h4>
            <ul className="flex flex-col gap-2 font-body-mature text-[13px] text-[#a1a1aa]">
              <li>
                <a href={SITE.github} target="_blank" rel="noopener noreferrer" className="hover:text-[#ffffff] transition-colors">
                  GitHub
                </a>
              </li>
              <li>
                <a href={SITE.discord} target="_blank" rel="noopener noreferrer" className="hover:text-[#ffffff] transition-colors">
                  Discord
                </a>
              </li>
              <li>
                <a href="mailto:support@condura.app" className="hover:text-[#ffffff] transition-colors">
                  Contact
                </a>
              </li>
            </ul>
          </div>

        </div>

        <div className="mt-16 pt-8 border-t border-white/[0.08] text-left">
          <p className="font-body-mature text-[13px] text-[#a1a1aa] leading-relaxed">
            All rights reserved. Free for personal and commercial use under the{" "}
            <Link href="/legal" className="text-[#ffffff] hover:underline">
              Condura EULA
            </Link>.
          </p>
          <div className="mt-4 flex flex-wrap items-center gap-4 text-[13px] text-[#a1a1aa] font-body-mature">
            <span>&copy; {new Date().getFullYear()} Condura Inc.</span>
            <span className="hidden md:inline border-l border-white/20 h-3" />
            <a href="https://condura.app" className="hover:text-[#ffffff]">condura.app</a>
          </div>
        </div>
      </div>
    </footer>
  );
}
