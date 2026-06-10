"use client";

/*
  The Touch — the site's signature sequence.

  A bulb hangs in the dark. As you scroll, a hand reaches in from the
  right; one finger meets the glass, the filament catches, a bloom of
  light swallows the screen, and the whole site flips to its light
  world. Scrolling back undoes it. The metaphor is the product: one
  touch (one hotkey) wakes every AI on the machine.

  Everything is driven off a single scroll progress value, so the hand,
  spark, glow, wash, captions and theme can never drift apart.
*/
import {
  m,
  useMotionValueEvent,
  useScroll,
  useTransform,
  type MotionValue,
} from "motion/react";
import Link from "next/link";
import { useRef } from "react";
import { usePrefersReducedMotion } from "@/lib/use-reduced-motion";
import { useTheme } from "@/components/chrome/theme";
import { Magnetic } from "@/components/motion/magnetic";

const TOUCH = 0.55; // progress at which finger meets glass

/*
  Motion v12 hands plain `opacity` bindings to native scroll-driven
  animations, whose timeline ranges break inside a sticky container.
  Routing the value through a CSS variable keeps it on the rAF path.
*/
const fade = (v: MotionValue<number> | number) => ({
  ["--o" as string]: v,
  opacity: "var(--o)" as unknown as number,
});

/* Deterministic pseudo-random so SSR and client agree. */
function rnd(i: number, salt: number) {
  const x = Math.sin(i * 127.1 + salt * 311.7) * 43758.5453;
  return x - Math.floor(x);
}

function Motes({ count = 16 }: { count?: number }) {
  return (
    <div aria-hidden className="absolute inset-0 overflow-hidden">
      {Array.from({ length: count }, (_, i) => {
        const size = 1.5 + rnd(i, 1) * 2.5;
        return (
          <span
            key={i}
            className="mote"
            style={{
              left: `${8 + rnd(i, 2) * 84}%`,
              top: `${10 + rnd(i, 3) * 70}%`,
              width: size,
              height: size,
              ["--mote-dur" as string]: `${10 + rnd(i, 4) * 14}s`,
              ["--mote-delay" as string]: `${-rnd(i, 5) * 12}s`,
              ["--mote-x" as string]: `${(rnd(i, 6) - 0.5) * 120}px`,
              ["--mote-y" as string]: `${-30 - rnd(i, 7) * 90}px`,
            }}
          />
        );
      })}
    </div>
  );
}

function Bulb({
  glow,
  filament,
  onClick,
  lit,
}: {
  glow: MotionValue<number> | number;
  filament: MotionValue<number> | number;
  onClick: () => void;
  lit: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={lit ? "Turn the light off" : "Turn the light on"}
      className="pointer-events-auto relative block focus:outline-none focus-visible:outline-2"
    >
      <svg
        viewBox="0 0 240 400"
        className="sway h-[46vh] min-h-[260px] w-auto"
        aria-hidden
      >
        <defs>
          <radialGradient id="bloom" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stopColor="#ffe2ae" stopOpacity="0.95" />
            <stop offset="35%" stopColor="#ffc46b" stopOpacity="0.5" />
            <stop offset="100%" stopColor="#ffc46b" stopOpacity="0" />
          </radialGradient>
          <radialGradient id="glass" cx="50%" cy="42%" r="62%">
            <stop offset="0%" stopColor="#fff3d6" stopOpacity="0.95" />
            <stop offset="60%" stopColor="#ffc46b" stopOpacity="0.55" />
            <stop offset="100%" stopColor="#ffa12e" stopOpacity="0.18" />
          </radialGradient>
        </defs>

        {/* cord */}
        <line x1="120" y1="0" x2="120" y2="118" stroke="currentColor" strokeWidth="2.5" className="text-ivory-dim" />

        {/* halo */}
        <m.circle cx="120" cy="234" r="150" fill="url(#bloom)" style={fade(glow)} />

        {/* glass */}
        <m.path
          d="M100,148 C100,176 82,186 75,212 A62,62 0 1,0 165,212 C158,186 140,176 140,148 Z"
          fill="url(#glass)"
          style={fade(glow)}
        />
        <path
          d="M100,148 C100,176 82,186 75,212 A62,62 0 1,0 165,212 C158,186 140,176 140,148 Z"
          fill="none"
          stroke="currentColor"
          strokeWidth="2.5"
          className="text-ivory-dim"
        />

        {/* filament */}
        <g fill="none" strokeWidth="2.5" strokeLinecap="round">
          <path
            d="M106,152 V198 M134,152 V198 M106,198 Q113,216 120,198 Q127,180 134,198"
            stroke="currentColor"
            className="text-ivory-faint"
          />
          <m.path
            d="M106,152 V198 M134,152 V198 M106,198 Q113,216 120,198 Q127,180 134,198"
            stroke="#ffd9a0"
            style={fade(filament)}
          />
        </g>

        {/* cap */}
        <rect x="98" y="118" width="44" height="32" rx="6" fill="var(--t-bg-3)" stroke="currentColor" strokeWidth="2.5" className="text-ivory-dim" />
        <line x1="100" y1="128" x2="140" y2="128" stroke="currentColor" strokeWidth="1.5" className="text-ivory-faint" />
        <line x1="100" y1="138" x2="140" y2="138" stroke="currentColor" strokeWidth="1.5" className="text-ivory-faint" />
      </svg>
    </button>
  );
}

function Hand({ x, draw }: { x: MotionValue<string>; draw: MotionValue<number> }) {
  return (
    <m.div
      aria-hidden
      style={{ x }}
      className="pointer-events-none absolute top-[12svh] right-0 w-[min(58vw,560px)]"
    >
      <svg viewBox="0 0 560 360" fill="none" className="w-full">
        {/* silhouette fill hides whatever passes behind the hand */}
        <m.path
          d="M16,138 C7,142 7,158 16,162 L150,170
             C176,172 188,186 186,200 C184,216 166,220 154,210
             C178,222 184,242 170,252 C158,260 144,254 138,244
             C158,260 158,280 142,288 C130,294 116,288 110,278
             C124,300 170,312 230,312 L560,316 L560,84 L360,82
             C296,82 226,96 182,126 C174,116 158,114 150,128 L16,138 Z"
          fill="var(--t-bg-2)"
          stroke="currentColor"
          strokeWidth="3"
          strokeLinejoin="round"
          className="text-ivory"
          style={{ pathLength: draw }}
        />
        {/* thumb resting over the curled fingers */}
        <m.path
          d="M192,152 C232,142 272,152 290,174 C300,188 294,206 276,210 C248,216 212,200 198,180"
          stroke="currentColor"
          strokeWidth="3"
          strokeLinecap="round"
          className="text-ivory"
          style={{ pathLength: draw }}
        />
        {/* index knuckle crease */}
        <m.path
          d="M150,140 C156,148 156,160 150,168"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          className="text-ivory-dim"
          style={{ pathLength: draw }}
        />
      </svg>
    </m.div>
  );
}

export function Illumination() {
  const ref = useRef<HTMLDivElement>(null);
  const reduced = usePrefersReducedMotion();
  const { theme, setTheme } = useTheme();
  const lit = theme === "light";

  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start start", "end end"],
  });

  /* Act I headline */
  const titleOpacity = useTransform(scrollYProgress, [0.04, 0.26], [1, 0]);
  const titleY = useTransform(scrollYProgress, [0.04, 0.3], ["0%", "-18%"]);
  const wdth = useTransform(scrollYProgress, [0, 0.3], [112, 70]);
  const hintOpacity = useTransform(scrollYProgress, [0, 0.06], [0.8, 0]);

  /* the hand approaches, touches, withdraws */
  const handX = useTransform(
    scrollYProgress,
    [0.16, TOUCH, 0.66, 0.78],
    ["112%", "-30%", "-24%", "116%"],
  );
  const handDraw = useTransform(scrollYProgress, [0.14, 0.4], [0, 1]);

  /* ignition */
  const filament = useTransform(scrollYProgress, [TOUCH, 0.585], [0, 1]);
  const glow = useTransform(scrollYProgress, [TOUCH, 0.62], [0, 1]);
  const spark = useTransform(
    scrollYProgress,
    [TOUCH - 0.012, TOUCH + 0.012, TOUCH + 0.05],
    [0, 1, 0],
  );
  const washScale = useTransform(scrollYProgress, [TOUCH, 0.7], [0.15, 6]);
  const washOpacity = useTransform(
    scrollYProgress,
    [TOUCH, 0.6, 0.7, 0.82],
    [0, 0.95, 0.6, 0],
  );

  /* the bulb makes room for the headline */
  const bulbY = useTransform(scrollYProgress, [0.6, 0.78], ["0%", "-16%"]);
  const bulbScale = useTransform(scrollYProgress, [0.6, 0.78], [1, 0.72]);

  /* Act II hero content */
  const heroOpacity = useTransform(scrollYProgress, [0.68, 0.8], [0, 1]);
  const heroY = useTransform(scrollYProgress, [0.68, 0.82], ["10%", "0%"]);

  /* captions */
  const cap1 = useTransform(scrollYProgress, [0.02, 0.08, 0.2, 0.26], [0, 1, 1, 0]);
  const cap2 = useTransform(scrollYProgress, [0.3, 0.36, 0.48, 0.53], [0, 1, 1, 0]);
  const cap3 = useTransform(scrollYProgress, [0.57, 0.62, 0.72, 0.78], [0, 1, 1, 0]);

  /* flip the world exactly once per crossing, in either direction */
  useMotionValueEvent(scrollYProgress, "change", (p) => {
    if (reduced) return;
    const next = p > 0.585 ? "light" : "dark";
    if (next !== theme) setTheme(next);
  });

  function toggleByScroll() {
    const el = ref.current;
    if (!el) return;
    if (reduced) {
      setTheme(lit ? "dark" : "light");
      return;
    }
    const range = el.offsetHeight - window.innerHeight;
    const target = lit ? el.offsetTop : el.offsetTop + range * 0.66;
    window.scrollTo({ top: target, behavior: "smooth" });
  }

  /* Reduced motion: a single-screen hero with a real switch. */
  if (reduced) {
    return (
      <section className="staff relative flex min-h-svh flex-col items-center justify-center gap-10 px-5 pt-24 text-center">
        <Bulb glow={lit ? 1 : 0} filament={lit ? 1 : 0} onClick={toggleByScroll} lit={lit} />
        <h1 className="display max-w-4xl text-[clamp(2.6rem,7vw,5.2rem)]">
          One touch wakes every AI on your machine.
        </h1>
        <p className="max-w-xl text-ivory-dim">
          Synaptic is the free, local-first conductor for the AI tools you
          already own. Press the hotkey; the room lights up.
        </p>
        <div className="flex flex-wrap items-center justify-center gap-5">
          <button type="button" onClick={toggleByScroll} className="trace cta">
            {lit ? "Turn the light off" : "Turn on the light"}
          </button>
          <Link href="/download" className="prose-link text-sm">
            Get Synaptic
          </Link>
        </div>
      </section>
    );
  }

  return (
    <section ref={ref} aria-label="The touch" className="relative h-[340vh]">
      <div className="sticky top-0 h-svh overflow-hidden">
        <Motes />

        {/* light shafts appear with the light world */}
        <m.div aria-hidden style={fade(glow)} className="absolute inset-0">
          <div className="shaft top-[-10%] left-[18%] h-[120%] w-40" />
          <div className="shaft top-[-10%] left-[55%] h-[120%] w-64" style={{ animationDelay: "-4s" }} />
        </m.div>

        {/* Act I headline */}
        <m.div
          style={{ ...fade(titleOpacity), y: titleY }}
          className="absolute inset-x-0 top-[46svh] z-10 px-5 text-center"
        >
          <p className="annotation">Act I — in the dark</p>
          <m.h1
            style={{ ["--wdth" as string]: wdth }}
            className="display mx-auto mt-5 max-w-5xl text-[clamp(2.6rem,8vw,6.4rem)] text-balance"
          >
            Your computer is full of genius,
            <span className="display-italic text-brass"> sitting in the dark.</span>
          </m.h1>
        </m.div>

        {/* the bulb */}
        <m.div
          style={{ y: bulbY, scale: bulbScale }}
          className="pointer-events-none absolute inset-x-0 top-0 z-20 flex justify-center"
        >
          <Bulb glow={glow} filament={filament} onClick={toggleByScroll} lit={lit} />
        </m.div>

        {/* contact spark */}
        <m.div
          aria-hidden
          style={{ ...fade(spark), scale: spark }}
          className="absolute top-[22svh] left-[calc(50%+4svh)] z-30"
        >
          <svg viewBox="0 0 60 60" className="h-14 w-14">
            {[0, 45, 90, 135, 180, 225, 270, 315].map((a) => (
              <line
                key={a}
                x1="30"
                y1="30"
                x2={30 + 26 * Math.cos((a * Math.PI) / 180)}
                y2={30 + 26 * Math.sin((a * Math.PI) / 180)}
                stroke="#ffe2ae"
                strokeWidth="2.5"
                strokeLinecap="round"
              />
            ))}
          </svg>
        </m.div>

        {/* the hand */}
        <Hand x={handX} draw={handDraw} />

        {/* the bloom that swallows the screen and masks the flip */}
        <m.div
          aria-hidden
          style={{ ...fade(washOpacity), scale: washScale }}
          className="absolute top-[-7svh] left-1/2 z-40 h-[60svh] w-[60svh] -translate-x-1/2 rounded-full"
        >
          <div
            className="h-full w-full rounded-full"
            style={{
              background:
                "radial-gradient(circle, #fff0d2 0%, #ffc46b 38%, rgba(255,196,107,0) 72%)",
            }}
          />
        </m.div>

        {/* captions */}
        <div className="absolute bottom-[9svh] left-1/2 z-30 w-full max-w-xl -translate-x-1/2 px-5 text-center">
          <m.p style={fade(cap1)} className="annotation absolute inset-x-0">
            Claude Code. Codex. Ollama. All installed. All asleep.
          </m.p>
          <m.p style={fade(cap2)} className="annotation absolute inset-x-0">
            It only takes one finger. One hotkey.
          </m.p>
          <m.p style={fade(cap3)} className="annotation absolute inset-x-0 !text-brass">
            Let there be light.
          </m.p>
        </div>

        {/* scroll hint */}
        <m.p
          style={fade(hintOpacity)}
          className="annotation absolute bottom-[4svh] left-1/2 z-10 -translate-x-1/2"
          aria-hidden
        >
          scroll to reach the bulb ↓
        </m.p>

        {/* Act II hero, revealed by the light */}
        <m.div
          style={{ ...fade(heroOpacity), y: heroY }}
          className="absolute inset-x-0 bottom-0 z-30 px-5 pb-[10svh] text-center"
        >
          <h2 className="display mx-auto max-w-5xl text-[clamp(2.4rem,6.6vw,5.4rem)] text-balance">
            One touch wakes
            <span className="display-italic text-brass"> every AI </span>
            on your machine.
          </h2>
          <p className="mx-auto mt-6 max-w-xl text-ivory-dim">
            Synaptic is the free, local-first conductor for the AI tools you
            already own — summoned by hotkey, governed by a deterministic
            Gatekeeper, telemetry-free forever.
          </p>
          <div className="mt-9 flex flex-wrap items-center justify-center gap-6">
            <Magnetic>
              <Link href="/download" className="trace cta">
                Get Synaptic
              </Link>
            </Magnetic>
            <Link href="/manifesto" className="prose-link text-sm">
              Read the manifesto
            </Link>
          </div>
        </m.div>
      </div>
    </section>
  );
}
