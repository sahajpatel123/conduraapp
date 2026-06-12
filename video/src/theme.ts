/*
  The Synaptic palette, lifted verbatim from web/app/globals.css ("The Touch"
  design system) so the film and the website are the same world. Two themes:
  the room starts dark and the bulb turns it on.
*/
import { Easing } from "remotion";

export const DARK = {
  bg: "#08080c",
  bg2: "#0e0e14",
  bg3: "#16161d",
  fg: "#efeae0",
  fgDim: "#97928a",
  fgFaint: "#7c786e",
  accent: "#ffa12e", // filament orange
  accentDeep: "#c87616",
  halt: "#e5484d",
  line: "rgba(239, 234, 224, 0.11)",
  lineStrong: "rgba(239, 234, 224, 0.24)",
  glow: "#ffc46b",
  success: "#5bbf72",
};

export const LIGHT = {
  bg: "#f6f1e5",
  bg2: "#fdfaf1",
  bg3: "#ece4d2",
  fg: "#17140e",
  fgDim: "#5d5749",
  fgFaint: "#756e5d",
  accent: "#a85700",
  accentDeep: "#8a4600",
  halt: "#bb3027",
  line: "rgba(23, 20, 14, 0.13)",
  lineStrong: "rgba(23, 20, 14, 0.3)",
  glow: "#ffb44d",
  success: "#3f9d57",
};

// The one ease curve used site-wide: cubic-bezier(0.16, 1, 0.3, 1).
export const EASE = Easing.bezier(0.16, 1, 0.3, 1);
export const EASE_IN_OUT = Easing.bezier(0.65, 0, 0.35, 1);

export const FPS = 30;

// Font families (loaded in fonts.ts). Fallbacks keep render safe if a
// network fetch is unavailable.
export const FONT = {
  display: 'var(--ff-display), "Archivo", system-ui, sans-serif',
  serif: 'var(--ff-serif), "Instrument Serif", Georgia, serif',
  sans: 'var(--ff-sans), system-ui, -apple-system, sans-serif',
  mono: 'var(--ff-mono), "Geist Mono", ui-monospace, "SF Mono", monospace',
};
