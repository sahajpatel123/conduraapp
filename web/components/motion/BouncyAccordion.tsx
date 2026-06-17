"use client";

import { useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { springBouncy } from "@/lib/motion";

export interface AccordionItem {
  id: string;
  title: string;
  body: string;
}

interface BouncyAccordionProps {
  items: AccordionItem[];
  defaultOpenId?: string;
}

export default function BouncyAccordion({ items, defaultOpenId }: BouncyAccordionProps) {
  const reduced = useReducedMotion();

  return (
    <div className="space-y-2">
      {items.map((item) => (
        <AccordionRow
          key={item.id}
          item={item}
          defaultOpen={item.id === defaultOpenId}
          reduced={reduced}
        />
      ))}
    </div>
  );
}

function AccordionRow({
  item,
  defaultOpen,
  reduced,
}: {
  item: AccordionItem;
  defaultOpen: boolean;
  reduced: boolean;
}) {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <div className="overflow-hidden rounded-2xl border border-white/[0.08] bg-white/[0.03]">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        className="flex w-full items-center justify-between gap-4 px-4 py-3.5 text-left"
      >
        <span className="text-sm font-medium text-white/85">{item.title}</span>
        <motion.span
          animate={{ rotate: open ? 45 : 0 }}
          transition={reduced ? { duration: 0 } : springBouncy}
          className="text-lg text-white/35"
          aria-hidden
        >
          +
        </motion.span>
      </button>
      <AnimatePresence initial={false}>
        {open && (
          <motion.div
            initial={reduced ? false : { height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={reduced ? { duration: 0 } : springBouncy}
          >
            <p className="border-t border-white/[0.05] px-4 py-3 text-sm leading-relaxed text-white/45">
              {item.body}
            </p>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
