"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { LayoutGroup, motion } from "motion/react";
import Tooltip from "@/components/motion/Tooltip";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { NAV_LINKS } from "@/lib/site";
import { springSnappy } from "@/lib/motion";

const ICONS: Record<string, string> = {
  "/": "⌂",
  "/manifesto": "◎",
  "/download": "↓",
  "/changelog": "◷",
  "/legal": "§",
};

const DOCK = [{ href: "/", label: "Home" }, ...NAV_LINKS];

export default function SiteDock() {
  const pathname = usePathname();
  const reduced = useReducedMotion();

  return (
    <nav
      aria-label="Primary"
      className="fixed bottom-5 left-1/2 z-[160] -translate-x-1/2"
    >
      <LayoutGroup id="site-dock">
        <div className="flex items-center gap-1 rounded-3xl border border-white/[0.08] bg-[#111113]/92 p-1.5 shadow-[0_12px_40px_rgba(0,0,0,0.45)] backdrop-blur-xl">
          {DOCK.map((item) => {
            const active =
              item.href === "/"
                ? pathname === "/"
                : pathname === item.href || pathname.startsWith(`${item.href}/`);
            const icon = ICONS[item.href] ?? "•";

            return (
              <Tooltip key={item.href} label={item.label} side="top">
                <Link
                  href={item.href}
                  aria-label={item.label}
                  aria-current={active ? "page" : undefined}
                  className="relative flex h-11 w-11 items-center justify-center rounded-2xl text-sm text-white/58 transition-colors hover:text-white"
                >
                  {active && (
                    <motion.span
                      layoutId="dock-active"
                      className="absolute inset-0 rounded-2xl bg-white/[0.10]"
                      transition={reduced ? { duration: 0 } : springSnappy}
                    />
                  )}
                  <span className="relative z-10" aria-hidden>
                    {icon}
                  </span>
                </Link>
              </Tooltip>
            );
          })}
        </div>
      </LayoutGroup>
    </nav>
  );
}
