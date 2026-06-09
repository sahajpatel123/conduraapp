"use client";

/*
  The ⌘K palette. One global keybinding, full keyboard operation,
  and the only piece of UI that floats above the score.

  Accessibility notes: the input is a combobox driving a listbox via
  aria-activedescendant; focus is held on the input (the only focusable
  control in the dialog) and restored to the opener on close. The dialog
  state lives in a child that unmounts on close, so every opening starts
  fresh.
*/
import { AnimatePresence, m } from "motion/react";
import { useRouter } from "next/navigation";
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { DUR, EASE } from "@/lib/motion";

type PaletteContextValue = { open: boolean; setOpen: (v: boolean) => void };

const PaletteContext = createContext<PaletteContextValue | null>(null);

export function usePalette() {
  const ctx = useContext(PaletteContext);
  if (!ctx) throw new Error("usePalette must be used inside PaletteProvider");
  return ctx;
}

type Command = {
  id: string;
  group: "Score" | "Act";
  label: string;
  hint: string;
  run: (router: ReturnType<typeof useRouter>) => void;
};

const COMMANDS: Command[] = [
  { id: "home", group: "Score", label: "Overture", hint: "/", run: (r) => r.push("/") },
  { id: "manifesto", group: "Score", label: "Manifesto", hint: "/manifesto", run: (r) => r.push("/manifesto") },
  { id: "changelog", group: "Score", label: "Changelog", hint: "/changelog", run: (r) => r.push("/changelog") },
  { id: "download", group: "Score", label: "Download", hint: "/download", run: (r) => r.push("/download") },
  {
    id: "top",
    group: "Act",
    label: "Return to the podium",
    hint: "scroll to top",
    run: () => window.scrollTo({ top: 0, behavior: "smooth" }),
  },
  {
    id: "copy",
    group: "Act",
    label: "Copy the mission",
    hint: "to clipboard",
    run: () =>
      navigator.clipboard
        ?.writeText(
          "Make AI useful to every ordinary person, on every computer, for free. No lock-in. No tracking. No compromise on speed or safety.",
        )
        .catch(() => {}),
  },
];

export function PaletteProvider({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setOpen((v) => !v);
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  // Hold the page still while the palette is up.
  useEffect(() => {
    if (!open) return;
    const prev = document.documentElement.style.overflow;
    document.documentElement.style.overflow = "hidden";
    return () => {
      document.documentElement.style.overflow = prev;
    };
  }, [open]);

  const value = useMemo(() => ({ open, setOpen }), [open]);

  return (
    <PaletteContext.Provider value={value}>
      {children}
      <AnimatePresence>
        {open && <PaletteDialog onClose={() => setOpen(false)} />}
      </AnimatePresence>
    </PaletteContext.Provider>
  );
}

function PaletteDialog({ onClose }: { onClose: () => void }) {
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [active, setActive] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const results = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return COMMANDS;
    return COMMANDS.filter(
      (c) => c.label.toLowerCase().includes(q) || c.hint.toLowerCase().includes(q),
    );
  }, [query]);

  const activeIdx = Math.min(active, Math.max(0, results.length - 1));

  // Focus the input on open; hand focus back to the opener on close.
  useEffect(() => {
    const opener = document.activeElement as HTMLElement | null;
    inputRef.current?.focus();
    return () => opener?.focus?.();
  }, []);

  // Escape closes no matter where the pointer or focus has wandered.
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  function pick(cmd: Command) {
    cmd.run(router);
    onClose();
  }

  function onKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Tab") {
      // The input is the dialog's only focusable control — keep focus on it.
      e.preventDefault();
      inputRef.current?.focus();
    }
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActive(Math.min(activeIdx + 1, results.length - 1));
    }
    if (e.key === "ArrowUp") {
      e.preventDefault();
      setActive(Math.max(activeIdx - 1, 0));
    }
    if (e.key === "Enter" && results[activeIdx]) {
      e.preventDefault();
      pick(results[activeIdx]);
    }
  }

  return (
    <m.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: DUR.micro }}
      className="fixed inset-0 z-[80] bg-ink/70 backdrop-blur-[2px]"
      onClick={onClose}
    >
      <m.div
        role="dialog"
        aria-modal="true"
        aria-label="Command palette"
        initial={{ opacity: 0, y: 14, scale: 0.985 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        exit={{ opacity: 0, y: 8, scale: 0.99 }}
        transition={{ duration: 0.3, ease: EASE }}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={onKeyDown}
        className="mx-auto mt-[18vh] w-[min(34rem,calc(100vw-2rem))] border border-line-strong bg-ink-2 shadow-[0_40px_120px_rgba(0,0,0,0.6)]"
      >
        <div className="flex items-center gap-3 border-b border-line px-5 transition-colors duration-200 focus-within:border-brass">
          <span aria-hidden className="size-1.5 rotate-45 bg-brass" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setActive(0);
            }}
            placeholder="Where to?"
            role="combobox"
            aria-expanded={results.length > 0}
            aria-controls="palette-listbox"
            aria-activedescendant={
              results[activeIdx] ? `palette-option-${results[activeIdx].id}` : undefined
            }
            aria-autocomplete="list"
            aria-label="Search commands"
            className="w-full bg-transparent py-4 font-mono text-sm text-ivory placeholder:text-ivory-faint focus:outline-none"
          />
          <kbd className="annotation shrink-0">esc</kbd>
        </div>
        {results.length === 0 ? (
          <p className="px-5 py-6 font-mono text-sm text-ivory-faint" role="status">
            Nothing in the score by that name.
          </p>
        ) : (
          <ul
            id="palette-listbox"
            role="listbox"
            aria-label="Commands"
            className="max-h-[40vh] overflow-y-auto py-2"
          >
            {results.map((cmd, i) => (
              <li
                key={cmd.id}
                id={`palette-option-${cmd.id}`}
                role="option"
                aria-selected={i === activeIdx}
                onMouseEnter={() => setActive(i)}
                onClick={() => pick(cmd)}
                className={`flex w-full cursor-pointer items-baseline justify-between gap-4 px-5 py-2.5 text-left transition-colors duration-150 ${
                  i === activeIdx ? "bg-ink-3 text-brass" : "text-ivory"
                }`}
              >
                <span className="text-sm">
                  <span className="annotation mr-3 !text-ivory-faint">{cmd.group}</span>
                  {cmd.label}
                </span>
                <span className="font-mono text-xs text-ivory-faint">{cmd.hint}</span>
              </li>
            ))}
          </ul>
        )}
      </m.div>
    </m.div>
  );
}
