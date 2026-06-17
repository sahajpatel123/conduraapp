"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { motion, AnimatePresence } from "motion/react";
import { SITE, NAV_LINKS } from "@/lib/site";

interface Item {
  label: string;
  href: string;
  category: string;
}

const ITEMS: Item[] = [
  ...NAV_LINKS.map((l) => ({ ...l, category: "Page" })),
  { label: "GitHub", href: SITE.github, category: "External" },
  { label: "Discord", href: SITE.discord, category: "External" },
];

export default function CommandPalette() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const prevOpen = useRef(open);

  const filtered = query.trim()
    ? ITEMS.filter(
        (item) =>
          item.label.toLowerCase().includes(query.toLowerCase()) ||
          item.category.toLowerCase().includes(query.toLowerCase())
      )
    : ITEMS;

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setOpen((prev) => !prev);
      }
      if (e.key === "Escape" && open) {
        setOpen(false);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open]);

  useEffect(() => {
    if (open && !prevOpen.current) {
      document.body.style.overflow = "hidden";
      requestAnimationFrame(() => {
        inputRef.current?.focus();
        setQuery("");
        setSelected(0);
      });
    } else if (!open) {
      document.body.style.overflow = "";
    }
    prevOpen.current = open;
    return () => { document.body.style.overflow = ""; };
  }, [open]);

  const onInputKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelected((i) => Math.min(i + 1, filtered.length - 1));
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelected((i) => Math.max(i - 1, 0));
      } else if (e.key === "Enter") {
        e.preventDefault();
        const item = filtered[selected];
        if (item) {
          setOpen(false);
          if (item.href.startsWith("http")) {
            window.open(item.href, "_blank", "noopener,noreferrer");
          } else {
            window.location.href = item.href;
          }
        }
      }
    },
    [filtered, selected]
  );

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.12 }}
          className="fixed inset-0 z-[200] flex items-start justify-center bg-black/50 pt-[20vh] backdrop-blur-sm"
          onClick={() => setOpen(false)}
        >
          <motion.div
            initial={{ opacity: 0, scale: 0.97, y: -6 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.97, y: -6 }}
            transition={{ duration: 0.18, ease: "easeOut" }}
            className="w-full max-w-xl overflow-hidden rounded-xl border border-white/[0.08] bg-[#0a0a0b] shadow-2xl"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center border-b border-white/[0.06] px-4 py-3">
              <svg className="mr-3 h-5 w-5 shrink-0 text-white/20" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
              </svg>
              <input
                ref={inputRef}
                type="text"
                placeholder="Search pages..."
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value);
                  setSelected(0);
                }}
                onKeyDown={onInputKeyDown}
                className="flex-1 bg-transparent text-[15px] text-white outline-none placeholder:text-white/20"
              />
              <kbd className="ml-3 hidden rounded-md border border-white/[0.08] bg-white/[0.03] px-1.5 py-0.5 text-[11px] text-white/30 sm:inline-block">ESC</kbd>
            </div>
            <div className="max-h-[50vh] overflow-y-auto py-2">
              {filtered.length === 0 ? (
                <div className="px-4 py-6 text-center text-sm text-white/20">
                  No results for &ldquo;{query}&rdquo;
                </div>
              ) : (
                filtered.map((item, index) => (
                  <button
                    key={item.label + item.href}
                    onClick={() => {
                      setOpen(false);
                      if (item.href.startsWith("http")) {
                        window.open(item.href, "_blank", "noopener,noreferrer");
                      } else {
                        window.location.href = item.href;
                      }
                    }}
                    onMouseEnter={() => setSelected(index)}
                    className={`w-full flex items-center justify-between px-4 py-2.5 text-left text-[14px] transition-colors ${
                      index === selected
                        ? "bg-[#0066cc]/15 text-white"
                        : "text-white/60 hover:bg-white/[0.03]"
                    }`}
                  >
                    <span>{item.label}</span>
                    <span className="text-[11px] text-white/20">{item.category}</span>
                  </button>
                ))
              )}
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
