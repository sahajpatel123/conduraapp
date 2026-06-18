/* ────────────────────────────────────────────────────────────
   Icon — Mature SVG icon set
   Stroke-based, minimal, consistent with the dark glass theme.
   No emojis. Every icon is 24×24, currentColor, 1.5 stroke.
   ──────────────────────────────────────────────────────────── */

import { type SVGProps } from "react";

type IconProps = Omit<SVGProps<SVGSVGElement>, "strokeWidth"> & {
  size?: number;
  strokeWidth?: number;
};

function base({ size = 20, strokeWidth = 1.5, ...rest }: IconProps) {
  return {
    width: size,
    height: size,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    ...rest,
  };
}

/* ── Platform / OS ── */

export function IconMac(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M16.5 3c-.3 1.1-1.1 2-2 2.6-.9.5-2 .7-3 .5.3-1.1 1.1-2 2-2.6.9-.5 2-.7 3-.5z" />
      <path d="M18.5 17c-.5 1.1-1 2.1-1.8 2.9-1.1 1.2-2.2 1.3-3.5.8-1.3-.5-2.4-.5-3.7 0-1.3.5-2.4.4-3.5-.8C4.4 18 3 15 3 12c0-2.5 1.8-3.6 3.5-3.6 1.3 0 2.4.8 3.6.8 1.2 0 2-.8 3.6-.8 1.2 0 2.5.5 3.5 1.8-2.3 1.4-2 4.6.3 5.8z" />
    </svg>
  );
}

export function IconWindows(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M3 5.5l7-1v6.5H3V5.5z" />
      <path d="M11 4.2l10-1.5V11H11V4.2z" />
      <path d="M3 13h7v6.5l-7-1V13z" />
      <path d="M11 13h10v7.3l-10-1.5V13z" />
    </svg>
  );
}

export function IconLinux(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M12 3a3 3 0 0 0-3 3c0 1 .4 1.6.4 2.5 0 .8-.6 1.4-1.2 2.2C7.2 12.4 6 14 6 16.5c0 1.8.8 3 2 3.5 1 .4 2.5.5 4 .5s3-.1 4-.5c1.2-.5 2-1.7 2-3.5 0-2.5-1.2-4.1-2.2-5.3-.6-.8-1.2-1.4-1.2-2.2 0-.9.4-1.5.4-2.5a3 3 0 0 0-3-3z" />
      <path d="M10.5 8.5h.01M13.5 8.5h.01" />
    </svg>
  );
}

/* ── Features ── */

export function IconBolt(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M13 2L4.5 13.5H11l-1 8.5 8.5-11.5H12l1-8.5z" />
    </svg>
  );
}

export function IconLock(props: IconProps) {
  return (
    <svg {...base(props)}>
      <rect x="4" y="11" width="16" height="10" rx="2" />
      <path d="M8 11V7a4 4 0 0 1 8 0v4" />
      <circle cx="12" cy="16" r="1" />
    </svg>
  );
}

export function IconShield(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      <path d="M9 12l2 2 4-4" />
    </svg>
  );
}

export function IconGlobe(props: IconProps) {
  return (
    <svg {...base(props)}>
      <circle cx="12" cy="12" r="9" />
      <path d="M3 12h18M12 3c2.5 2.5 2.5 15 0 18M12 3c-2.5 2.5-2.5 15 0 18" />
    </svg>
  );
}

export function IconMonitor(props: IconProps) {
  return (
    <svg {...base(props)}>
      <rect x="3" y="4" width="18" height="12" rx="2" />
      <path d="M8 20h8M12 16v4" />
    </svg>
  );
}

export function IconGift(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M20 12v9H4v-9M2 7h20v5H2zM12 22V7M12 7S11 3 8.5 3 6 5 6 5s.5 2 2.5 2h7c2 0 2.5-2 2.5-2s-.5-2-2.5-2S12 7 12 7z" />
    </svg>
  );
}

export function IconList(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01" />
    </svg>
  );
}

export function IconCheck(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M20 6L9 17l-5-5" />
    </svg>
  );
}

export function IconArrowRight(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M5 12h14M13 6l6 6-6 6" />
    </svg>
  );
}

export function IconArrowDown(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M12 5v14M6 13l6 6 6-6" />
    </svg>
  );
}

export function IconDownload(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M12 3v12M7 10l5 5 5-5M5 21h14" />
    </svg>
  );
}

export function IconKey(props: IconProps) {
  return (
    <svg {...base(props)}>
      <circle cx="8" cy="15" r="4" />
      <path d="M10.8 12.2L21 2M17 6l3 3M14 9l2 2" />
    </svg>
  );
}

export function IconEye(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

export function IconCpu(props: IconProps) {
  return (
    <svg {...base(props)}>
      <rect x="6" y="6" width="12" height="12" rx="2" />
      <rect x="9" y="9" width="6" height="6" />
      <path d="M9 2v3M15 2v3M9 19v3M15 19v3M2 9h3M2 15h3M19 9h3M19 15h3" />
    </svg>
  );
}

export function IconFingerprint(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M12 11a2 2 0 0 0-2 2c0 1.5.5 3 .5 5M12 11a2 2 0 0 1 2 2c0 2.5-1 5-1 6.5M8 8a6 6 0 0 1 8 0M6 13c0-1.5.5-2.5 1.5-3.5M18 13c0 2-1 4-1 5.5M5 18c0-1 .5-2 .5-3M19 18c0 .5-.2 1-.5 1.5" />
    </svg>
  );
}

export function IconTerminal(props: IconProps) {
  return (
    <svg {...base(props)}>
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="M7 9l3 3-3 3M13 15h4" />
    </svg>
  );
}

export function IconCopy(props: IconProps) {
  return (
    <svg {...base(props)}>
      <rect x="9" y="9" width="11" height="11" rx="2" />
      <path d="M5 15V5a2 2 0 0 1 2-2h10" />
    </svg>
  );
}

export function IconSparkle(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M12 3l1.8 5.2L19 10l-5.2 1.8L12 17l-1.8-5.2L5 10l5.2-1.8L12 3z" />
    </svg>
  );
}

export function IconRocket(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M5 15c-1.5 1-2 4-2 4s3-.5 4-2M9 11a3 3 0 1 1 6 0c0 4-3 7-3 7s-3-3-3-7z" />
      <path d="M9 11c-2 0-4 1-4 3l2 1M15 11c2 0 4 1 4 3l-2 1M12 14h.01" />
    </svg>
  );
}

export function IconLayers(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M12 3l9 5-9 5-9-5 9-5z" />
      <path d="M3 13l9 5 9-5M3 17l9 5 9-5" />
    </svg>
  );
}

export function IconGithub(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M9 19c-4 1.5-4-2-6-2.5M15 22v-3.5c0-1 .3-1.8-.5-2.5 2.5-.3 5-1.2 5-5.5 0-1.2-.4-2.2-1-3 .1-.3.5-1.5-.1-3 0 0-1-.3-3.3 1a11 11 0 0 0-6 0C6 3 5 3.3 5 3.3c-.6 1.5-.2 2.7-.1 3-.6.8-1 1.8-1 3 0 4.3 2.5 5.2 5 5.5-.5.5-.8 1.2-.6 2V22" />
    </svg>
  );
}

export function IconHeart(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M12 21s-7-4.5-9-9.5C1.5 7 4 4 7 4c2 0 3 1 5 3 2-2 3-3 5-3 3 0 5.5 3 4 7.5-2 5-9 9.5-9 9.5z" />
    </svg>
  );
}

/* ── Map string keys → components for data-driven sections ── */

export const ICONS = {
  mac: IconMac,
  windows: IconWindows,
  linux: IconLinux,
  bolt: IconBolt,
  lock: IconLock,
  shield: IconShield,
  globe: IconGlobe,
  monitor: IconMonitor,
  gift: IconGift,
  list: IconList,
  check: IconCheck,
  key: IconKey,
  eye: IconEye,
  cpu: IconCpu,
  fingerprint: IconFingerprint,
  terminal: IconTerminal,
  copy: IconCopy,
  sparkle: IconSparkle,
  rocket: IconRocket,
  layers: IconLayers,
  github: IconGithub,
  heart: IconHeart,
} as const;

export type IconKey = keyof typeof ICONS;

export function Icon({ name, ...rest }: { name: IconKey } & IconProps) {
  const Cmp = ICONS[name];
  return <Cmp {...rest} />;
}
