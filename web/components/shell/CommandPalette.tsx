"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { motion, AnimatePresence, LayoutGroup } from "motion/react";
import { SITE, NAV_LINKS } from "@/lib/site";
import { springSoft } from "@/lib/motion";

interface Item {
  label: string;
  href: string;
  category: string;
}

const ITEMS: Item[] = [
  { label: "Home", href: "/", category: "Page" },
  ...NAV_LINKS.map((l) => ({ ...l, category: "Page" })),
  { label: "GitHub", href: SITE.github, category: "External" },
  { label: "Discord", href: SITE.discord, category: "External" },
];

export default function CommandPalette() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

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
      if (e.key === "Escape" && open) setOpen(false);
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open]);

  useEffect(() => {
    if (!open) {
      document.body.style.overflow = "";
      return;
    }
    document.body.style.overflow = "hidden";
    requestAnimationFrame(() => {
      inputRef.current?.focus();
      setQuery("");
      setSelected(0);
    });
    return () => {
      document.body.style.overflow = "";
    };
  }, [open]);

  const navigate = useCallback((href: string) => {
    setOpen(false);
    if (href.startsWith("http")) {
      window.open(href, "_blank", "noopener,noreferrer");
    } else {
      window.location.href = href;
    }
  }, []);

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
        if (item) navigate(item.href);
      }
    },
    [filtered, selected, navigate]
  );

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 z-[220] flex items-start justify-center bg-black/55 px-4 pt-[18vh] backdrop-blur-lg"
          onClick={() => setOpen(false)}
        >
          <motion.div
            layoutId="condura-command"
            initial={{ opacity: 0, y: -8, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -6, scale: 0.99 }}
            transition={springSoft}
            className="w-full max-w-xl overflow-hidden rounded-2xl border border-white/[0.08] bg-[#111113]/95 shadow-2xl backdrop-blur-xl"
            onClick={(e) => e.stopPropagation()}
            role="dialog"
            aria-label="Command palette"
          >
            <div className="flex items-center gap-3 border-b border-white/[0.07] px-4 py-3">
              <span className="text-white/25" aria-hidden>
                ⌘K
              </span>
              <input
                ref={inputRef}
                type="search"
                placeholder={`Search ${SITE.name}…`}
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value);
                  setSelected(0);
                }}
                onKeyDown={onInputKeyDown}
                className="min-w-0 flex-1 bg-transparent py-2 text-sm text-white outline-none placeholder:text-white/30"
              />
            </div>
            <LayoutGroup>
              <ul className="max-h-[50vh] overflow-y-auto py-2" role="listbox">
                {filtered.length === 0 ? (
                  <li className="px-4 py-6 text-center text-sm text-white/30">No matches</li>
                ) : (
                  filtered.map((item, index) => (
                    <li key={item.label + item.href}>
                      <button
                        type="button"
                        role="option"
                        aria-selected={index === selected}
                        onMouseEnter={() => setSelected(index)}
                        onClick={() => navigate(item.href)}
                        className={`relative flex w-full items-center justify-between px-4 py-2.5 text-left text-sm ${
                          index === selected ? "text-white" : "text-white/55"
                        }`}
                      >
                        {index === selected && (
                          <motion.span
                            layoutId="palette-highlight"
                            className="absolute inset-x-2 inset-y-0 rounded-lg bg-white/[0.08]"
                            transition={springSoft}
                          />
                        )}
                        <span className="relative z-10">{item.label}</span>
                        <span className="relative z-10 text-[11px] text-white/25">{item.category}</span>
                      </button>
                    </li>
                  ))
                )}
              </ul>
            </LayoutGroup>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
