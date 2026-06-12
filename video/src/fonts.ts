/*
  Type system, mirroring the website:
    Archivo          — display grotesk
    Instrument Serif — the accent voice, italic
    Geist            — UI / body
    Geist Mono       — annotations, terminal, code

  Self-hosted from public/fonts (see scripts/fetch-fonts.mjs) so rendering
  needs no network. Loaded via a React hook (not at module top-level, which
  would call delayRender outside a render and crash compositions-calculation).
*/
import { continueRender, delayRender, staticFile } from "remotion";
import { useEffect, useState } from "react";

let linkInjected = false;

// The faces we actually use, so we can wait on them explicitly.
const FACES = [
  '800 64px Archivo',
  '700 32px Archivo',
  '600 24px Archivo',
  'italic 400 48px "Instrument Serif"',
  '600 24px Geist',
  '500 24px Geist',
  '400 24px Geist',
  '500 18px "Geist Mono"',
  '400 18px "Geist Mono"',
];

export function useSynapticFonts() {
  const [handle] = useState(() => delayRender("Loading Synaptic fonts"));

  useEffect(() => {
    let cancelled = false;
    const finish = () => {
      if (!cancelled) continueRender(handle);
    };

    const afterStylesheet = () => {
      const ready =
        typeof document !== "undefined" && document.fonts
          ? Promise.all(FACES.map((f) => document.fonts.load(f).catch(() => null))).then(
              () => document.fonts.ready,
            )
          : Promise.resolve();
      ready.then(finish).catch(finish);
    };

    if (!linkInjected && typeof document !== "undefined") {
      linkInjected = true;
      const link = document.createElement("link");
      link.rel = "stylesheet";
      link.href = staticFile("fonts.css");
      link.onload = afterStylesheet;
      link.onerror = finish;
      document.head.appendChild(link);
    } else {
      afterStylesheet();
    }

    // Safety: never hang the render.
    const t = setTimeout(finish, 8000);
    return () => {
      cancelled = true;
      clearTimeout(t);
    };
  }, [handle]);
}

export function fontVars(): React.CSSProperties {
  return {
    // @ts-expect-error CSS custom properties
    "--ff-display": "Archivo",
    "--ff-serif": "Instrument Serif",
    "--ff-sans": "Geist",
    "--ff-mono": "Geist Mono",
  };
}
