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

/* ── Platform / OS ──
   Real, recognizable brand marks rendered as filled vector glyphs so they
   pick up currentColor and stay crisp at any size. No external image
   assets, no background, no licensing ambiguity — pure SVG paths. */

/**
 * macOS — the Apple logo. The canonical silhouette.
 */
export function IconMac(props: IconProps) {
  const { fill, stroke, ...rest } = props;
  return (
    <svg {...base(rest)} fill={fill ?? "currentColor"} stroke={stroke ?? "none"}>
      <path d="M17.05 12.04c-.02-2.6 2.12-3.85 2.22-3.91-1.21-1.77-3.09-2.01-3.76-2.04-1.6-.16-3.12.94-3.93.94-.81 0-2.06-.92-3.39-.89-1.74.03-3.35 1.01-4.25 2.57-1.81 3.14-.46 7.79 1.3 10.34.86 1.25 1.89 2.65 3.23 2.6 1.3-.05 1.79-.84 3.36-.84 1.57 0 2.02.84 3.41.81 1.41-.02 2.3-1.27 3.16-2.53 1-1.45 1.41-2.86 1.43-2.93-.03-.01-2.74-1.05-2.77-4.17M14.6 4.59c.72-.87 1.21-2.08 1.07-3.29-1.04.04-2.31.7-3.05 1.57-.67.77-1.26 2.01-1.1 3.2 1.16.09 2.35-.59 3.08-1.48" />
    </svg>
  );
}

/**
 * Windows — the four-pane flag mark.
 */
export function IconWindows(props: IconProps) {
  const { fill, stroke, ...rest } = props;
  return (
    <svg {...base(rest)} fill={fill ?? "currentColor"} stroke={stroke ?? "none"}>
      <path d="M3 5.5L10.4 4.5V11.4H3V5.5Z" />
      <path d="M11.4 4.35L21 3V11.4H11.4V4.35Z" />
      <path d="M3 12.6H10.4V19.5L3 18.5V12.6Z" />
      <path d="M11.4 12.6H21V21L11.4 19.65V12.6Z" />
    </svg>
  );
}

/**
 * Linux — Tux, the official penguin mascot. A single filled silhouette.
 */
export function IconLinux(props: IconProps) {
  const { fill, stroke, ...rest } = props;
  return (
    <svg {...base(rest)} fill={fill ?? "currentColor"} stroke={stroke ?? "none"}>
      <path d="M12.504 0c-.155 0-.315.008-.48.021-4.226.333-3.105 4.807-3.17 6.298-.076 1.092-.3 1.953-1.05 3.02-.885 1.051-2.127 2.75-2.716 4.521-.278.832-.41 1.684-.287 2.489a.424.424 0 00-.11.135c-.26.268-.45.6-.663.839-.199.199-.485.267-.797.4-.313.136-.658.269-.864.68-.09.189-.136.394-.132.602 0 .199.027.4.055.536.058.399.116.728.04.97-.249.68-.28 1.145-.106 1.484.174.334.535.47.94.601.81.2 1.91.135 2.774.6.926.466 1.866.67 2.616.47.526-.116.97-.464 1.208-.946.587-.003 1.23-.269 2.26-.334.699-.058 1.574.267 2.577.2.025.134.063.198.114.333l.003.003c.391.778 1.113 1.132 1.884 1.071.771-.06 1.592-.536 2.257-1.306.631-.765 1.683-1.084 2.378-1.503.348-.199.629-.469.649-.853.023-.4-.2-.811-.714-1.376v-.097l-.003-.003c-.17-.2-.25-.535-.338-.926-.085-.401-.182-.786-.492-1.046h-.003c-.059-.054-.123-.067-.188-.135a.357.357 0 00-.19-.064c.431-1.278.264-2.55-.173-3.694-.533-1.41-1.465-2.638-2.175-3.483-.796-1.005-1.576-1.957-1.56-3.368.026-2.152.236-6.133-3.544-6.139z" />
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

export function IconGithub({ size = 20, className }: IconProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="currentColor"
      className={className}
      aria-hidden
    >
      <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.21 11.39.6.11.82-.26.82-.58 0-.29-.01-1.13-.02-2.22-3.34.73-4.04-1.61-4.04-1.61-.55-1.39-1.35-1.76-1.35-1.76-1.1-.75.08-.73.08-.73 1.2.09 1.84 1.24 1.84 1.24 1.08 1.83 2.82 1.3 3.5 1 .11-.78.42-1.31.76-1.61-2.66-.3-5.47-1.33-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.12-.3-.54-1.52.12-3.18 0 0 1.01-.32 3.3 1.23.96-.27 1.98-.4 3-.4 1.02 0 2.04.13 3 .4 2.29-1.55 3.3-1.23 3.3-1.23.66 1.66.24 2.88.12 3.18.77.84 1.24 1.91 1.24 3.22 0 4.61-2.81 5.63-5.48 5.92.43.37.81 1.1.81 2.22 0 1.6-.01 2.89-.01 3.28 0 .32.22.7.83.58C20.56 21.8 24 17.3 24 12 24 5.37 18.63 0 12 0z" />
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

export function IconDiscord({ size = 20, strokeWidth = 2.1, className }: IconProps) {
  /* Thick-outline Clyde — matches dock stroke weight, uses currentColor like sibling icons. */
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={strokeWidth}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden
    >
      <path d="M9.2 10.8a1.2 1.2 0 1 0 0 2.4 1.2 1.2 0 0 0 0-2.4z" />
      <path d="M14.8 10.8a1.2 1.2 0 1 0 0 2.4 1.2 1.2 0 0 0 0-2.4z" />
      <path d="M6.4 8.4c2.8-1.35 8.4-1.35 11.2 0" />
      <path d="M5.6 8.8c-1.45 2.7-1.65 6.1-.15 8.85" />
      <path d="M18.4 8.8c1.45 2.7 1.65 6.1.15 8.85" />
      <path d="M8.6 17c.85.38 2.05.62 3.4.62s2.55-.24 3.4-.62" />
      <path d="M7.4 16.8l-.75 1.35" />
      <path d="M16.6 16.8l.75 1.35" />
    </svg>
  );
}

export function IconHome(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M3 10.5L12 3l9 7.5M5 9.5V20a1 1 0 0 0 1 1h4v-6h4v6h4a1 1 0 0 0 1-1V9.5" />
    </svg>
  );
}

export function IconClock(props: IconProps) {
  return (
    <svg {...base(props)}>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5l3 2" />
    </svg>
  );
}

export function IconScale(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M12 3v18M5 7h14M7 7l-3 6a3 3 0 0 0 6 0L7 7zM17 7l-3 6a3 3 0 0 0 6 0l-3-6zM7 21h10" />
    </svg>
  );
}

export function IconCompass(props: IconProps) {
  return (
    <svg {...base(props)}>
      <circle cx="12" cy="12" r="9" />
      <path d="M15.5 8.5l-2 5-5 2 2-5 5-2z" />
    </svg>
  );
}

export function IconCommand(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M9 6a3 3 0 1 0-3 3h12a3 3 0 1 0-3-3v12a3 3 0 1 0 3-3H6a3 3 0 1 0 3 3V6z" />
    </svg>
  );
}

export function IconArrowUp(props: IconProps) {
  return (
    <svg {...base(props)}>
      <path d="M12 19V5M6 11l6-6 6 6" />
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
  discord: IconDiscord,
  home: IconHome,
  clock: IconClock,
  scale: IconScale,
  compass: IconCompass,
  command: IconCommand,
  arrowUp: IconArrowUp,
  arrowRight: IconArrowRight,
  download: IconDownload,
} as const;

export type IconKey = keyof typeof ICONS;

export function Icon({ name, ...rest }: { name: IconKey } & IconProps) {
  const Cmp = ICONS[name];
  return <Cmp {...rest} />;
}
