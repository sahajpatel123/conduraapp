"use client";

import Link from "next/link";
import { NAV_LINKS, TOOL_ROSTER } from "@/lib/site";

// Split the reference links into two readable columns.
const EXPLORE_LINKS = NAV_LINKS.filter((l) =>
  ["/orchestration", "/security", "/manifesto"].includes(l.href)
);
const RESOURCE_LINKS = NAV_LINKS.filter((l) =>
  ["/changelog", "/download", "/legal"].includes(l.href)
);

export default function Footer() {
  return (
    <footer className="w-full bg-[#000000] py-16 px-6 border-t border-white/[0.08] relative z-10 select-none">
      <div className="mx-auto w-full max-w-6xl">
        
        <div className="grid grid-cols-2 gap-x-8 gap-y-12 md:grid-cols-12 md:gap-x-10">
          
          <div className="col-span-2 md:col-span-5 md:pr-12">
            <Link
              href="/"
              aria-label="Condura home"
              className="group inline-flex items-end gap-3 text-white"
            >
              <span className="font-body-mature text-[30px] font-semibold leading-none">
                Condura
              </span>
              <span
                aria-hidden="true"
                className="mb-1.5 size-2 bg-[#D97757] transition-transform duration-300 ease-out group-hover:rotate-45"
              />
            </Link>

            <p className="mt-6 max-w-[410px] font-body-mature text-[22px] font-medium leading-7 text-white">
              Intelligence that answers{" "}
              <span className="text-[#8b8b93]">to you.</span>
            </p>

            <p className="mt-3 max-w-[410px] font-body-mature text-[13px] leading-5 text-[#a1a1aa]">
              A local-first layer for your operating system, built to
              coordinate the models and tools you already trust.
            </p>

            <div className="mt-7 flex flex-wrap items-center gap-x-4 gap-y-2 font-body-mature text-[12px] text-[#8b8b93]">
              <span className="inline-flex items-center gap-2">
                <span aria-hidden="true" className="size-1.5 rounded-full bg-[#D97757]" />
                Local by default
              </span>
              <span aria-hidden="true" className="hidden h-3 w-px bg-white/15 sm:block" />
              <span>Permission before action</span>
            </div>
          </div>

          <div className="md:col-span-3">
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

          <div className="md:col-span-2">
            <h4 className="font-body-mature font-medium text-[#ffffff] text-[14px] mb-4">
              Explore
            </h4>
            <ul className="flex flex-col gap-2 font-body-mature text-[13px] text-[#a1a1aa]">
              {EXPLORE_LINKS.map((link) => (
                <li key={link.href}>
                  <Link href={link.href} className="hover:text-[#ffffff] transition-colors">
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div className="md:col-span-2">
            <h4 className="font-body-mature font-medium text-[#ffffff] text-[14px] mb-4">
              Resources
            </h4>
            <ul className="flex flex-col gap-2 font-body-mature text-[13px] text-[#a1a1aa]">
              {RESOURCE_LINKS.map((link) => (
                <li key={link.href}>
                  <Link href={link.href} className="hover:text-[#ffffff] transition-colors">
                    {link.label}
                  </Link>
                </li>
              ))}
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
