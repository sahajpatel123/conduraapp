"use client";

import { useState, type ReactNode } from "react";
import { LayoutGroup, motion } from "motion/react";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { springSnappy } from "@/lib/motion";

export interface TabItem {
  id: string;
  label: string;
  content: ReactNode;
}

interface SharedLayoutTabsProps {
  items: TabItem[];
  defaultId?: string;
  value?: string;
  onChange?: (id: string) => void;
  layoutId?: string;
}

export default function SharedLayoutTabs({
  items,
  defaultId,
  value,
  onChange,
  layoutId = "condura-tabs",
}: SharedLayoutTabsProps) {
  const reduced = useReducedMotion();
  const [internal, setInternal] = useState(defaultId ?? items[0]?.id ?? "");
  const active = value ?? internal;
  const setActive = onChange ?? setInternal;
  const current = items.find((i) => i.id === active) ?? items[0];

  return (
    <div>
      <LayoutGroup id={layoutId}>
        <div
          role="tablist"
          aria-label="Sections"
          className="inline-flex rounded-full border border-white/[0.08] bg-white/[0.035] p-1"
        >
          {items.map((item) => {
            const selected = item.id === active;
            return (
              <button
                key={item.id}
                type="button"
                role="tab"
                aria-selected={selected}
                onClick={() => setActive(item.id)}
                className="relative rounded-full px-4 py-2 text-xs font-medium text-white/58 transition-colors hover:text-white/85"
              >
                {selected && (
                  <motion.span
                    layoutId={`${layoutId}-pill`}
                    className="absolute inset-0 rounded-full bg-white/[0.10] shadow-[inset_0_1px_0_rgba(255,255,255,0.10)]"
                    transition={reduced ? { duration: 0 } : springSnappy}
                  />
                )}
                <span className="relative z-10">{item.label}</span>
              </button>
            );
          })}
        </div>
      </LayoutGroup>

      <motion.div
        key={current?.id}
        role="tabpanel"
        initial={reduced ? false : { opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.28 }}
        className="mt-8"
      >
        {current?.content}
      </motion.div>
    </div>
  );
}
