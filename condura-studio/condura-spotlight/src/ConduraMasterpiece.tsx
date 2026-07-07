import type {CSSProperties, ReactNode} from 'react';
import type {EasingFunction} from 'remotion';
import {
  AbsoluteFill,
  Easing,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from 'remotion';
import {z} from 'zod';

export const ConduraMasterpieceSchema = z.object({
  brandMark: z.string().default('C'),
  brandName: z.string().default('Condura'),
  tagline: z.string().default('The conductor for every AI on your computer.'),
  problemLines: z.array(z.string()).default([
    'Your computer is full of AI.',
    'None of it talks to each other.',
  ]),
  chaosStats: z.array(z.string()).default([
    '17 windows',
    '9 contexts',
    '0 coordination',
  ]),
  conductorClaim: z
    .string()
    .default('Condura conducts every AI on your machine.'),
  capabilities: z
    .array(z.object({word: z.string(), sub: z.string()}))
    .default([
      {word: 'MODELS', sub: '12 providers, one key'},
      {word: 'AGENTS', sub: '8 CLIs orchestrated'},
      {word: 'BROWSER', sub: 'clicks, types, reads'},
      {word: 'VOICE', sub: 'say the word'},
      {word: 'LOCAL', sub: 'on your machine first'},
      {word: 'SAFE', sub: 'asks before it acts'},
    ]),
  orbitNodes: z.array(z.string()).default([
    'Claude',
    'GPT',
    'Gemini',
    'Ollama',
    'Codex',
    'Antigravity',
    'Cursor',
    'OpenCode',
  ]),
  cta: z.string().default('Free. Forever.'),
  site: z.string().default('condura.app'),
  launch: z.string().default('launching soon'),
});

export type MasterpieceProps = z.infer<typeof ConduraMasterpieceSchema>;

const TOTAL = 990;
const CX = 960;
const CY = 540;

const acid = '#f4ff5c';
const cyan = '#00f0ff';
const magenta = '#ff3df2';
const orange = '#ff7a1a';
const ink = '#030304';
const bone = '#f8f7ef';

const easeOut = Easing.bezier(0.16, 1, 0.3, 1);
const hardEase = Easing.bezier(0.76, 0, 0.24, 1);
const slamEase = Easing.bezier(0.9, 0, 0.1, 1);
const overEase = Easing.bezier(0.34, 1.56, 0.64, 1);

const beats = [
  {label: 'Void', start: 0, dur: 90},
  {label: 'Mesh', start: 90, dur: 150},
  {label: 'Fracture', start: 240, dur: 120},
  {label: 'Conductor', start: 360, dur: 150},
  {label: 'Symphony', start: 510, dur: 150},
  {label: 'Power', start: 660, dur: 150},
  {label: 'Coda', start: 810, dur: 180},
];

function r(
  frame: number,
  from: number,
  to: number,
  easing: EasingFunction = easeOut,
) {
  return interpolate(frame, [from, to], [0, 1], {
    easing,
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
}

function out(frame: number, from: number, to: number) {
  return interpolate(frame, [from, to], [1, 0], {
    easing: Easing.in(Easing.cubic),
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
}

function pulse(frame: number, speed: number, min = 0, max = 1) {
  return interpolate(Math.sin(frame / speed), [-1, 1], [min, max]);
}

function Scene({
  start,
  dur,
  children,
  className = '',
}: {
  start: number;
  dur: number;
  children: (local: number) => ReactNode;
  className?: string;
}) {
  const frame = useCurrentFrame();
  const f = 14;
  if (frame < start - f || frame > start + dur + f) return null;
  const enter = r(frame, start, start + f, Easing.out(Easing.cubic));
  const exit = out(frame, start + dur - f, start + dur);
  const opacity = Math.min(enter, exit);
  const shove = interpolate(enter, [0, 1], [40, 0]);
  return (
    <AbsoluteFill
      className={`mp-scene ${className}`}
      style={{opacity, transform: `translate3d(0, ${shove}px, 0)`}}
    >
      {children(frame - start)}
    </AbsoluteFill>
  );
}

function Backdrop() {
  const frame = useCurrentFrame();
  const drift = Math.sin(frame / 140) * 30;
  return (
    <AbsoluteFill className="mp-backdrop">
      <div
        className="mp-wash mp-wash--a"
        style={{transform: `translate3d(${drift}px, ${-drift * 0.4}px, 0)`}}
      />
      <div
        className="mp-wash mp-wash--b"
        style={{transform: `translate3d(${-drift * 0.6}px, ${drift * 0.35}px, 0)`}}
      />
      <div className="mp-grid" />
      <div className="mp-vignette" />
    </AbsoluteFill>
  );
}

function Flash({at, color = acid}: {at: number; color?: string}) {
  const frame = useCurrentFrame();
  const amount = interpolate(Math.abs(frame - at), [0, 7], [0.78, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  return <AbsoluteFill style={{background: color, opacity: amount, mixBlendMode: 'screen'}} />;
}

/* RGB-split kinetic word. Three stacked copies blended screen, offset by a
 * sine-driven amount that spikes on glitch beats. */
function GlitchWord({
  text,
  local,
  delay,
  size,
  align = 'left',
  accent,
  glitchAt,
}: {
  text: string;
  local: number;
  delay: number;
  size: 'slam' | 'huge' | 'big' | 'med';
  align?: 'left' | 'center';
  accent?: string;
  glitchAt?: number[];
}) {
  const show = r(local, delay, delay + 16, slamEase);
  const enter = interpolate(show, [0, 1], [80, 0]);
  let g = 0;
  if (glitchAt) {
    for (const at of glitchAt) {
      g = Math.max(g, interpolate(Math.abs(local - at), [0, 6], [1, 0], {
        extrapolateLeft: 'clamp',
        extrapolateRight: 'clamp',
      }));
    }
  }
  const wobble = Math.sin((local + delay) * 1.3) > 0.6 ? 1 : 0;
  const off = 8 + g * 22;
  const color = accent ?? bone;
  return (
    <div
      className={`glitch glitch-${size} glitch-${align}`}
      style={{opacity: show, transform: `translate3d(0, ${enter}px, 0)`}}
    >
      <span style={{color}}>{text}</span>
      <span
        aria-hidden
        style={{
          color: cyan,
          transform: `translate3d(${wobble ? off : off * 0.3}px, ${-off * 0.2}px, 0)`,
        }}
      >
        {text}
      </span>
      <span
        aria-hidden
        style={{
          color: magenta,
          transform: `translate3d(${wobble ? -off : -off * 0.3}px, ${off * 0.2}px, 0)`,
        }}
      >
        {text}
      </span>
    </div>
  );
}

/* Per-word spring reveal for a line of text. */
function KineticLine({
  text,
  local,
  delay,
  className = '',
  wordMs = 4,
  size = 'line',
}: {
  text: string;
  local: number;
  delay: number;
  className?: string;
  wordMs?: number;
  size?: 'line' | 'line-lg';
}) {
  const {fps} = useVideoConfig();
  const words = text.split(' ');
  return (
    <div className={`kinetic kinetic-${size} ${className}`}>
      {words.map((word, i) => {
        const at = delay + i * wordMs;
        const s = spring({
          frame: local - at,
          fps,
          config: {damping: 14, stiffness: 120, mass: 0.8},
        });
        return (
          <span key={word + i} style={{opacity: s, transform: `translate3d(0, ${interpolate(s, [0, 1], [40, 0])}px, 0)`}}>
            {word}
            {i < words.length - 1 ? '\u00A0' : ''}
          </span>
        );
      })}
    </div>
  );
}

// Deterministic pseudo-random
function rng(seed: number) {
  const x = Math.sin(seed * 999.13) * 43758.5453;
  return x - Math.floor(x);
}

/* ----- SCENE: Void ----- */
function VoidScene({local}: {local: number}) {
  const p = pulse(local, 18, 0.3, 1);
  const core = r(local, 0, 30);
  const halo = r(local, 20, 70);
  return (
    <AbsoluteFill className="mp-void">
      <div
        className="void-core"
        style={{
          opacity: core,
          transform: `translate(-50%, -50%) scale(${0.4 + core * (0.6 + p * 0.25)})`,
          boxShadow: `0 0 ${60 + p * 80}px rgba(244,255,92,${0.3 + p * 0.4})`,
        }}
      >
        C
      </div>
      <div className="void-halo" style={{opacity: halo, transform: `translate(-50%, -50%) scale(${0.5 + halo * 0.7})`}} />
      <div className="void-halo void-halo--2" style={{opacity: halo * 0.6, transform: `translate(-50%, -50%) scale(${0.3 + halo * 0.9}) rotate(${local * 0.6}deg)`}} />
    </AbsoluteFill>
  );
}

/* ----- SCENE: Neural Mesh (self-drawing SVG constellation) ----- */
const NODE_COUNT = 34;
const nodes = Array.from({length: NODE_COUNT}).map((_, i) => ({
  x: 120 + rng(i + 1) * 1680,
  y: 110 + rng(i + 7) * 860,
  r: 2 + rng(i + 3) * 3,
  seed: i,
}));

const edges: [number, number][] = [];
for (let i = 0; i < NODE_COUNT; i++) {
  for (let j = i + 1; j < NODE_COUNT; j++) {
    const dx = nodes[i].x - nodes[j].x;
    const dy = nodes[i].y - nodes[j].y;
    if (Math.hypot(dx, dy) < 230) edges.push([i, j]);
  }
}

function MeshScene({local, problemLines}: {local: number; problemLines: string[]}) {
  const draw = r(local, 10, 150, hardEase);
  const nodeShow = r(local, 0, 60);
  return (
    <AbsoluteFill className="mp-mesh">
      <svg className="mesh-svg" viewBox="0 0 1920 1080">
        {edges.map(([a, b], i) => {
          const appear = r(local, 20 + (i % 40), 80 + (i % 40));
          const na = nodes[a];
          const nb = nodes[b];
          return (
            <line
              key={i}
              x1={na.x}
              y1={na.y}
              x2={nb.x}
              y2={nb.y}
              stroke={i % 3 === 0 ? acid : i % 3 === 1 ? cyan : magenta}
              strokeWidth={1}
              opacity={appear * 0.4}
              style={{filter: 'drop-shadow(0 0 4px currentColor)'}}
            />
          );
        })}
        {nodes.map((n, i) => {
          const show = r(local, i * 1.2, i * 1.2 + 14);
          const blip = pulse(local + i * 5, 9 + (i % 5), 0.5, 1);
          return (
            <circle
              key={i}
              cx={n.x}
              cy={n.y}
              r={n.r * (0.6 + blip * 0.4) * show}
              fill={i % 3 === 0 ? acid : i % 3 === 1 ? cyan : magenta}
              opacity={show * 0.85}
              style={{filter: 'drop-shadow(0 0 6px currentColor)'}}
            />
          );
        })}
      </svg>
      <div className="mesh-copy">
        <KineticLine text={problemLines[0]} local={local} delay={70} size="line-lg" />
        <KineticLine text={problemLines[1] ?? ''} local={local} delay={70 + problemLines[0].split(' ').length * 4 + 14} className="muted" />
      </div>
      <div className="mesh-draw" style={{opacity: draw}}>
        <i style={{transform: `scaleX(${draw})`}} />
      </div>
    </AbsoluteFill>
  );
}

/* ----- SCENE: Fracture (chaos) ----- */
function FractureScene({local, chaosStats}: {local: number; chaosStats: string[]}) {
  const break_ = r(local, 0, 70, slamEase);
  const glitchBeats = [10, 26, 42, 58];
  return (
    <AbsoluteFill className="mp-fracture">
      <div className="fract-shards">
        {Array.from({length: 9}).map((_, i) => {
          const fall = r(local, i * 3, i * 3 + 16, slamEase);
          return (
            <i
              key={i}
              style={{
                left: `${10 + (i * 37) % 80}%`,
                top: `${15 + (i * 53) % 70}%`,
                opacity: fall * (1 - break_),
                transform: `translate3d(${interpolate(fall, [0, 1], [0, (i % 2 ? -1 : 1) * (160 + break_ * 200)])}px, ${interpolate(fall, [0, 1], [0, (i % 3 - 1) * 140])}px, 0) rotate(${i * 23}deg)`,
              }}
            />
          );
        })}
      </div>
      <div className="fract-stack">
        {chaosStats.map((stat, i) => (
          <GlitchWord
            key={stat}
            text={stat.toUpperCase()}
            local={local}
            delay={8 + i * 16}
            size={i === 2 ? 'huge' : 'big'}
            align="center"
            accent={i === 2 ? magenta : bone}
            glitchAt={glitchBeats}
          />
        ))}
      </div>
      <div className="fract-bg" style={{transform: `scale(${1 + break_ * 0.2}) rotate(${break_ * 4}deg)`}} />
    </AbsoluteFill>
  );
}

/* ----- FAUX-3D CONDUCTOR RING -----
 * Simulates a torus rotating on Y axis. Generates two ring paths (front/back)
 * plus orbiting nodes positioned by angle; depth sorts via scale + opacity. */
function ConductorRing({
  local,
  orbitNodes,
  appear,
  spin = 1,
}: {
  local: number;
  orbitNodes: string[];
  appear: number;
  spin?: number;
}) {
  const RX = 520;
  const RY = 120; // perspective squash
  const angle = local * 0.018 * spin;
  const ringOpacity = appear;
  // orbiting nodes
  const labeled = orbitNodes;
  const nodePoints = labeled.map((_, i) => {
    const base = (i / labeled.length) * Math.PI * 2 + local * 0.01 * spin;
    // Y rotation: x' = x cosθ + z sinθ, z' = -x sinθ + z cosθ (with z=0 on ring plane)
    const x = Math.cos(base) * RX;
    const z = Math.sin(base) * RX; // depth
    const xr = x * Math.cos(angle);
    const zr = x * Math.sin(angle);
    const screenX = CX + xr;
    const screenY = CY + z * (RY / RX) * 0; // ring is flat in XZ; squash for perspective
    // Apply angle to the depth component to make nodes swing in front/back
    const depth = zr; // -RX..RX, front = + (closer)
    const depthN = (depth + RX) / (RX * 2); // 0..1, 1 = front
    const scale = 0.5 + depthN * 0.8;
    const op = 0.25 + depthN * 0.75;
    return {screenX, screenY, scale, op, base, depthN, label: labeled[i]};
  });
  // sorted so back ones render first
  const sorted = [...nodePoints].sort((a, b) => a.depthN - b.depthN);

  return (
    <AbsoluteFill className="mp-ring-layer" style={{opacity: ringOpacity}}>
      <svg className="ring-svg" viewBox="0 0 1920 1080">
        {/* back half ring */}
        <ellipse
          cx={CX}
          cy={CY}
          rx={RX}
          ry={RY}
          fill="none"
          stroke={cyan}
          strokeWidth={1.5}
          opacity={ringOpacity * 0.35}
          strokeDasharray="2 7"
          style={{filter: 'drop-shadow(0 0 6px rgba(0,240,255,.6))'}}
        />
        {/* front half ring drawn as path arc */}
        <path
          d={`M ${CX - RX} ${CY} A ${RX} ${RY} 0 0 0 ${CX + RX} ${CY}`}
          fill="none"
          stroke={acid}
          strokeWidth={2}
          opacity={ringOpacity * 0.9}
          style={{filter: 'drop-shadow(0 0 8px rgba(244,255,92,.7))'}}
        />
        <path
          d={`M ${CX - RX} ${CY} A ${RX} ${RY} 0 0 1 ${CX + RX} ${CY}`}
          fill="none"
          stroke={magenta}
          strokeWidth={1}
          opacity={ringOpacity * 0.5}
          strokeDasharray="3 6"
        />
        {/* inner concentric ring */}
        <ellipse cx={CX} cy={CY} rx={RX * 0.62} ry={RY * 0.62} fill="none" stroke={orange} strokeWidth={1} opacity={ringOpacity * 0.4} strokeDasharray="4 10" />
        {/* tick marks rotating */}
        {Array.from({length: 24}).map((_, i) => {
          const a = (i / 24) * Math.PI * 2 + angle * 2;
          const x1 = CX + Math.cos(a) * RX;
          const y1 = CY + Math.sin(a) * RY;
          const x2 = CX + Math.cos(a) * (RX + 18);
          const y2 = CY + Math.sin(a) * (RY + 4);
          const front = Math.sin(a) > 0;
          return (
            <line
              key={i}
              x1={x1}
              y1={y1}
              x2={x2}
              y2={y2}
              stroke={front ? acid : cyan}
              strokeWidth={front ? 2 : 1}
              opacity={ringOpacity * (front ? 0.85 : 0.3)}
            />
          );
        })}
      </svg>
      {/* orbiting labeled nodes as HTML for crisp text */}
      {sorted.map((n) => (
        <div
          key={n.label}
          className="orbit-node"
          style={{
            left: n.screenX,
            top: n.screenY,
            opacity: ringOpacity * n.op,
            transform: `translate(-50%, -50%) scale(${n.scale})`,
            zIndex: Math.round(n.depthN * 100),
          }}
        >
          <i />
          <b>{n.label}</b>
        </div>
      ))}
      {/* center core */}
      <div
        className="ring-core"
        style={{
          opacity: ringOpacity,
          transform: `translate(-50%, -50%) scale(${0.7 + appear * 0.3})`,
          boxShadow: `0 0 ${60 + pulse(local, 12, 0.3, 1) * 60}px rgba(244,255,92,.5)`,
        }}
      >
        C
      </div>
      {/* connecting pulses core -> front nodes */}
      <svg className="ring-pulses" viewBox="0 0 1920 1080" style={{opacity: ringOpacity}}>
        {nodePoints.map((n, i) => {
          const a = (pulse(local + i * 7, 10, 0, 1) > 0.7) ? 1 : 0;
          return (
            <line
              key={i}
              x1={CX}
              y1={CY}
              x2={n.screenX}
              y2={n.screenY}
              stroke={i % 2 ? cyan : acid}
              strokeWidth={1}
              opacity={a * n.depthN}
            />
          );
        })}
      </svg>
    </AbsoluteFill>
  );
}

function ConductorScene({local, orbitNodes}: {local: number; orbitNodes: string[]}) {
  const appear = r(local, 20, 90, hardEase);
  return (
    <AbsoluteFill className="mp-conductor">
      <div className="conductor-bg" />
      <ConductorRing local={local} orbitNodes={orbitNodes} appear={appear} />
      <div className="conductor-eyebrow" style={{opacity: r(local, 60, 90)}}>
        <span>THE MISSING LAYER</span>
      </div>
    </AbsoluteFill>
  );
}

/* ----- SCENE: Symphony (claim + ring active) ----- */
function SymphonyScene({local, orbitNodes, conductorClaim}: {local: number; orbitNodes: string[]; conductorClaim: string}) {
  return (
    <AbsoluteFill className="mp-symphony">
      <div className="symphony-bg" />
      <ConductorRing local={520 + local} orbitNodes={orbitNodes} appear={1} spin={1.6} />
      <div className="symphony-claim">
        <KineticLine text={conductorClaim} local={local} delay={16} size="line-lg" />
      </div>
    </AbsoluteFill>
  );
}

/* ----- SCENE: Power (capability slams) ----- */
function PowerScene({local, capabilities}: {local: number; capabilities: {word: string; sub: string}[]}) {
  const step = 22;
  const activeIndex = Math.min(capabilities.length - 1, Math.max(0, Math.floor(local / step)));
  const active = capabilities[activeIndex];
  const hit = r(local, activeIndex * step, activeIndex * step + 8, slamEase);
  const ringPulse = 1 + pulse(local, 8, 0, 0.06);
  return (
    <AbsoluteFill className="mp-power">
      <div className="power-bg" />
      {/* faint ring behind */}
      <div className="power-ring" style={{opacity: r(local, 0, 14), transform: `translate(-50%, -50%) scale(${ringPulse})`}}>
        {Array.from({length: 16}).map((_, i) => (
          <i key={i} style={{transform: `rotate(${(i / 16) * 360 + local * 1.5}deg) translateY(-300px)`, background: i % 3 === 0 ? acid : i % 3 === 1 ? cyan : magenta}} />
        ))}
      </div>
      <div className="power-eyebrow" style={{opacity: r(local, 0, 14)}}>
        <span>WHAT IT DOES</span>
      </div>
      <div className="power-word" style={{opacity: hit, transform: `translate(-50%, -50%) scale(${0.7 + hit * 0.3})`}}>
        <span>{active.word}</span>
      </div>
      <div className="power-sub" style={{opacity: r(local, activeIndex * step + 4, activeIndex * step + 10)}}>
        <span>{active.sub}</span>
      </div>
      <div className="power-dots">
        {capabilities.map((c, i) => (
          <i key={c.word} className={i <= activeIndex ? 'on' : ''} />
        ))}
      </div>
    </AbsoluteFill>
  );
}

/* ----- SCENE: Coda (resolve) ----- */
function CodaScene({local, props}: {local: number; props: MasterpieceProps}) {
  const {fps} = useVideoConfig();
  const mark = spring({frame: local - 10, fps, config: {damping: 9, stiffness: 90, mass: 0.75}});
  const glow = pulse(local, 16, 0.4, 1);
  const line = r(local, 60, 110, hardEase);
  return (
    <AbsoluteFill className="mp-coda">
      <div className="coda-bg" />
      {/* resolving ring */}
      <div className="coda-ring" style={{opacity: r(local, 0, 30)}}>
        {Array.from({length: 3}).map((_, i) => (
          <i key={i} style={{transform: `translate(-50%, -50%) scale(${1 + i * 0.4 + local * 0.002})`, opacity: 0.4 - i * 0.1}} />
        ))}
      </div>
      <div className="coda-mark" style={{opacity: mark, transform: `translate(-50%, -50%) scale(${0.6 + mark * 0.4})`}}>
        <div className="coda-c" style={{boxShadow: `0 0 ${60 + glow * 80}px rgba(244,255,92,.5)`}}>{props.brandMark}</div>
        <h1>{props.brandName}</h1>
      </div>
      <div className="coda-line">
        <i style={{transform: `scaleX(${line})`}} />
      </div>
      <div className="coda-tagline" style={{opacity: r(local, 70, 104)}}>
        <span>{props.tagline}</span>
      </div>
      <div className="coda-cta" style={{opacity: r(local, 104, 130)}}>
        <strong>{props.cta}</strong>
      </div>
      <div className="coda-site" style={{opacity: r(local, 122, 150)}}>
        <b>{props.site}</b>
        <i>{props.launch}</i>
      </div>
    </AbsoluteFill>
  );
}

function Hud() {
  const frame = useCurrentFrame();
  return (
    <div className="mp-hud">
      <div className="hud-left">
        <span>{'CONDURA // MASTERPIECE'}</span>
      </div>
      <div className="hud-right">
        <b>{`${Math.floor(frame / 30).toString().padStart(2, '0')}s / ${Math.floor(TOTAL / 30)}s`}</b>
      </div>
      <div className="hud-bar">
        <i style={{transform: `scaleX(${frame / TOTAL})`}} />
      </div>
      <div className="hud-beats">
        {beats.map((b) => {
          const active = frame >= b.start && frame < b.start + b.dur;
          return (
            <span key={b.label} className={active ? 'on' : ''}>
              {b.label}
            </span>
          );
        })}
      </div>
    </div>
  );
}

function Letterbox() {
  return (
    <>
      <div className="mp-letterbox mp-letterbox--top" />
      <div className="mp-letterbox mp-letterbox--bottom" />
    </>
  );
}

export const ConduraMasterpiece: React.FC<MasterpieceProps> = (props) => {
  return (
    <AbsoluteFill className="mp-root">
      <Backdrop />
      <Scene start={0} dur={90}>
        {(local) => <VoidScene local={local} />}
      </Scene>
      <Scene start={90} dur={150}>
        {(local) => <MeshScene local={local} problemLines={props.problemLines} />}
      </Scene>
      <Scene start={240} dur={120}>
        {(local) => <FractureScene local={local} chaosStats={props.chaosStats} />}
      </Scene>
      <Scene start={360} dur={150}>
        {(local) => <ConductorScene local={local} orbitNodes={props.orbitNodes} />}
      </Scene>
      <Scene start={510} dur={150}>
        {(local) => <SymphonyScene local={local} orbitNodes={props.orbitNodes} conductorClaim={props.conductorClaim} />}
      </Scene>
      <Scene start={660} dur={150}>
        {(local) => <PowerScene local={local} capabilities={props.capabilities} />}
      </Scene>
      <Scene start={810} dur={180}>
        {(local) => <CodaScene local={local} props={props} />}
      </Scene>
      {beats.slice(1).map((b) => (
        <Flash key={b.start} at={b.start} color={b.start === 240 ? magenta : b.start === 510 ? cyan : b.start === 810 ? acid : acid} />
      ))}
      <div className="mp-grain" />
      <div className="mp-scanlines" />
      <Letterbox />
      <Hud />
    </AbsoluteFill>
  );
};

const _u: CSSProperties = {};
void _u;