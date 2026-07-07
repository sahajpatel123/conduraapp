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

const TOTAL = 660;

const beats = [
  {label: 'Hook', start: 0, duration: 90, tone: 'dark' as const},
  {label: 'Sting', start: 90, duration: 75, tone: 'blood' as const},
  {label: 'Chaos', start: 165, duration: 90, tone: 'dark' as const},
  {label: 'Reveal', start: 255, duration: 75, tone: 'acid' as const},
  {label: 'Power', start: 330, duration: 120, tone: 'dark' as const},
  {label: 'Fomo', start: 450, duration: 90, tone: 'blood' as const},
  {label: 'Cta', start: 540, duration: 120, tone: 'acid' as const},
];

const acid = '#f4ff5c';
const cyan = '#00f0ff';
const magenta = '#ff3df2';
const orange = '#ff7a1a';
const ink = '#030304';
const bone = '#f8f7ef';

const easeOut = Easing.bezier(0.16, 1, 0.3, 1);
const hardEase = Easing.bezier(0.76, 0, 0.24, 1);
const slamEase = Easing.bezier(0.9, 0, 0.1, 1);

function r(
  frame: number,
  from: number,
  to: number,
  easing: EasingFunction = easeOut,
): number {
  return interpolate(frame, [from, to], [0, 1], {
    easing,
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
}

function pulse(frame: number, speed: number, min = 0, max = 1): number {
  return interpolate(Math.sin(frame / speed), [-1, 1], [min, max]);
}

function type(text: string, frame: number, from: number, to: number): string {
  return text.slice(0, Math.floor(r(frame, from, to) * text.length));
}

function Beat({
  start,
  duration,
  children,
  className = '',
  tone = 'dark',
}: {
  start: number;
  duration: number;
  children: (local: number) => ReactNode;
  className?: string;
  tone?: 'dark' | 'acid' | 'blood';
}) {
  const frame = useCurrentFrame();
  const fade = 10;
  const live = frame >= start - fade && frame <= start + duration + fade;
  if (!live) return null;

  const enter = r(frame, start, start + fade, Easing.out(Easing.cubic));
  const exit = interpolate(frame, [start + duration - fade, start + duration], [1, 0], {
    easing: Easing.in(Easing.cubic),
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const opacity = Math.min(enter, exit);
  const shove = interpolate(enter, [0, 1], [30, 0]);

  return (
    <AbsoluteFill
      className={`hype-beat hype-beat-${tone} ${className}`}
      style={{opacity, transform: `translate3d(0, ${shove}px, 0)`}}
    >
      {children(frame - start)}
    </AbsoluteFill>
  );
}

function Flash({at, color = acid}: {at: number; color?: string}) {
  const frame = useCurrentFrame();
  const amount = interpolate(Math.abs(frame - at), [0, 7], [0.82, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  return <AbsoluteFill style={{background: color, opacity: amount, mixBlendMode: 'screen'}} />;
}

function Shake({local, intensity = 1, children}: {local: number; intensity?: number; children: ReactNode}) {
  const x = Math.sin(local * 1.9) * 6 * intensity;
  const y = Math.cos(local * 2.3) * 5 * intensity;
  return <div style={{transform: `translate3d(${x}px, ${y}px, 0)`}}>{children}</div>;
}

function GlitchWord({
  text,
  local,
  delay,
  size,
  className = '',
}: {
  text: string;
  local: number;
  delay: number;
  size: 'slam' | 'huge' | 'big';
  className?: string;
}) {
  const show = r(local, delay, delay + 14, slamEase);
  const glitch = Math.sin((local + delay) * 1.3) > 0.6 ? 1 : 0;
  const x = interpolate(show, [0, 1], [-120, 0]);
  return (
    <div className={`glitch glitch-${size} ${className}`} style={{opacity: show, transform: `translate3d(${x}px, 0, 0)`}}>
      <span>{text}</span>
      <span aria-hidden style={{transform: `translate3d(${glitch ? 12 : 0}px, ${glitch ? -4 : 0}px, 0)`}}>
        {text}
      </span>
      <span aria-hidden style={{transform: `translate3d(${glitch ? -10 : 0}px, ${glitch ? 6 : 0}px, 0)`}}>
        {text}
      </span>
    </div>
  );
}

function Orbital({local, intense = false}: {local: number; intense?: boolean}) {
  const count = intense ? 16 : 10;
  return (
    <div className="hype-orbital">
      {Array.from({length: count}).map((_, index) => {
        const radius = 110 + index * (intense ? 32 : 40);
        const spin = local * (0.4 + index * 0.02);
        const angle = spin + index * 41;
        const size = 10 + (index % 4) * 7;
        return (
          <i
            key={index}
            style={{
              width: size,
              height: size,
              transform: `translate(-50%, -50%) rotate(${angle}deg) translateX(${radius}px) rotate(${-angle}deg)`,
              background: index % 3 === 0 ? acid : index % 3 === 1 ? cyan : magenta,
              opacity: 0.5 + (index % 5) * 0.08,
            }}
          />
        );
      })}
    </div>
  );
}

function HookScene({local}: {local: number}) {
  return (
    <>
      <div className="hook-bg" style={{transform: `scale(${1 + r(local, 60, 90, slamEase) * 0.3})`}} />
      <Orbital local={local} />
      <div className="hook-stack">
        <div className="hook-tone" style={{opacity: r(local, 6, 22)}}>
          <Shake local={local} intensity={0.6}>
            <span>YOU&rsquo;RE PAYING</span>
          </Shake>
        </div>
        <GlitchWord text="FOR 5" local={local} delay={18} size="slam" />
        <GlitchWord text="AI APPS" local={local} delay={36} size="slam" className="accent-acid" />
      </div>
      <div className="hook-sub" style={{opacity: r(local, 56, 78)}}>
        <b>$20</b>
        <b>$20</b>
        <b>$20</b>
        <b>$20</b>
        <b>$20</b>
      </div>
    </>
  );
}

function StingScene({local}: {local: number}) {
  return (
    <>
      <div className="sting-bg" />
      <div className="sting-stack">
        <GlitchWord text="AND STILL" local={local} delay={4} size="huge" />
        <GlitchWord text="DOING THE" local={local} delay={18} size="huge" />
        <GlitchWord text="WORK" local={local} delay={34} size="slam" className="accent-magenta" />
        <GlitchWord text="YOURSELF" local={local} delay={48} size="slam" className="accent-acid" />
      </div>
      <div className="sting-mark" style={{opacity: r(local, 40, 60)}}>
        <i>YOURSELF</i> <em>.</em>
      </div>
    </>
  );
}

const chaosTools = [
  {label: 'GPT', x: 140, y: 320, rot: -14, color: cyan},
  {label: 'CLAUDE', x: 880, y: 280, rot: 11, color: orange},
  {label: 'CURSOR', x: 220, y: 980, rot: -7, color: acid},
  {label: 'BROWSER', x: 900, y: 1040, rot: 9, color: magenta},
  {label: 'NOTION', x: 540, y: 620, rot: -3, color: cyan},
  {label: 'OBSIDIAN', x: 760, y: 1380, rot: 6, color: acid},
  {label: 'TERMINAL', x: 300, y: 1500, rot: -11, color: orange},
  {label: 'SLACK', x: 820, y: 1680, rot: 8, color: magenta},
  {label: 'FIGMA', x: 160, y: 1760, rot: -5, color: cyan},
];

function ChaosScene({local}: {local: number}) {
  const sink = r(local, 50, 90, slamEase);
  return (
    <>
      <div className="chaos-bg" />
      {chaosTools.map((tool, index) => {
        const show = r(local, index * 4, index * 4 + 12);
        const pull = interpolate(show, [0, 1], [(index % 2 ? 240 : -240), 0]);
        return (
          <div
            key={tool.label}
            className="chaos-tool"
            style={{
              left: tool.x,
              top: tool.y,
              opacity: show * (1 - sink),
              transform: `translate3d(${pull}px, 0, 0) rotate(${tool.rot - sink * tool.rot}deg) scale(${1 - sink * 0.3})`,
              borderColor: `${tool.color}66`,
              color: tool.color,
            }}
          >
            {tool.label}
          </div>
        );
      })}
      <div className="chaos-stop" style={{opacity: sink, transform: `translate(-50%, -50%) scale(${0.5 + sink * 0.5})`}}>
        <Shake local={local} intensity={1.4}>
          <span>STOP</span>
        </Shake>
      </div>
    </>
  );
}

function RevealScene({local}: {local: number}) {
  const {fps} = useVideoConfig();
  const form = spring({
    frame: local - 14,
    fps,
    config: {damping: 11, stiffness: 120, mass: 0.7},
  });
  const line = r(local, 38, 70, hardEase);
  return (
    <>
      <div className="reveal-bg" />
      <Orbital local={local * 1.6} intense />
      <div className="reveal-mark" style={{opacity: form, transform: `translate(-50%, -50%) scale(${0.5 + form * 0.5})`}}>
        <div className="reveal-c">C</div>
      </div>
      <div className="reveal-name" style={{opacity: r(local, 30, 56)}}>
        <strong>CONDURA</strong>
      </div>
      <div className="reveal-line">
        <i style={{transform: `scaleX(${line})`}} />
      </div>
      <div className="reveal-tag" style={{opacity: r(local, 50, 72)}}>
        <span>ONE KEY</span>
        <strong>EVERYTHING MOVES</strong>
      </div>
    </>
  );
}

const powers = [
  {word: 'MODELS', sub: '12 providers', color: acid, at: 0},
  {word: 'AGENTS', sub: '8 CLIs orchestrated', color: cyan, at: 16},
  {word: 'BROWSER', sub: 'clicks & types for you', color: orange, at: 32},
  {word: 'VOICE', sub: '"hey synaptic"', color: magenta, at: 48},
  {word: 'CONTROL', sub: 'asks before it acts', color: acid, at: 64},
  {word: 'FREE', sub: 'no subscription. ever.', color: cyan, at: 80},
];

function PowerScene({local}: {local: number}) {
  const activeIndex = Math.max(0, Math.floor(local / 16));
  const active = powers[Math.min(powers.length - 1, activeIndex)];
  const hit = r(local, active.at, active.at + 8, slamEase);
  return (
    <>
      <div className="power-bg" />
      <div className="power-orbit">
        {Array.from({length: 12}).map((_, index) => {
          const angle = local * 2 + index * 30;
          return (
            <i
              key={index}
              style={{
                transform: `translate(-50%, -50%) rotate(${angle}deg) translateY(-${260 + (index % 3) * 40}px)`,
                background: index % 3 === 0 ? acid : index % 3 === 1 ? cyan : magenta,
              }}
            />
          );
        })}
      </div>
      <div className="power-eyebrow" style={{opacity: r(local, 0, 14)}}>
        <span>WHAT IT DOES</span>
      </div>
      <div className="power-word" style={{opacity: hit, transform: `translate(-50%, -50%) scale(${0.6 + hit * 0.4})`}}>
        <span style={{color: active.color}}>{active.word}</span>
      </div>
      <div className="power-sub" style={{opacity: r(local, active.at + 4, active.at + 10)}}>
        <span style={{color: active.color}}>{active.sub}</span>
      </div>
      <div className="power-dots">
        {powers.map((power, index) => (
          <i key={power.word} className={index <= activeIndex ? 'on' : ''} />
        ))}
      </div>
    </>
  );
}

function FomoScene({local}: {local: number}) {
  const count = Math.floor(r(local, 20, 80) * 8);
  return (
    <>
      <div className="fomo-bg" />
      <div className="fomo-stack">
        <div className="fomo-eyebrow" style={{opacity: r(local, 0, 16)}}>
          <span>RIGHT NOW</span>
        </div>
        <GlitchWord text="WHILE YOU" local={local} delay={10} size="big" />
        <GlitchWord text="WATCH THIS" local={local} delay={26} size="huge" className="accent-cyan" />
        <GlitchWord text="SOMEONE" local={local} delay={44} size="big" />
        <GlitchWord text="FINISHED" local={local} delay={60} size="slam" className="accent-acid" />
      </div>
      <div className="fomo-counter" style={{opacity: r(local, 30, 60)}}>
        <b>{count.toLocaleString()}</b>
        <span>tasks done by Condura users while you read this</span>
      </div>
    </>
  );
}

function CtaScene({local}: {local: number}) {
  const {fps} = useVideoConfig();
  const mark = spring({
    frame: local - 8,
    fps,
    config: {damping: 9, stiffness: 90, mass: 0.75},
  });
  const glow = pulse(local, 16, 0.4, 1);
  return (
    <>
      <div className="cta-bg" />
      <Orbital local={local * 2} intense />
      <div className="cta-mark" style={{opacity: mark, transform: `translate(-50%, -50%) scale(${0.6 + mark * 0.4})`}}>
        <div className="cta-c" style={{boxShadow: `0 0 ${60 + glow * 70}px rgba(244, 255, 92, .45)`}}>C</div>
        <h1>CONDURA</h1>
      </div>
      <div className="cta-promise" style={{opacity: r(local, 34, 62)}}>
        <strong>FREE.</strong>
        <strong>FOREVER.</strong>
      </div>
      <div className="cta-line" style={{transform: `scaleX(${r(local, 56, 92, hardEase)})`}} />
      <div className="cta-url" style={{opacity: r(local, 72, 100)}}>
        <b>condura.app</b>
      </div>
      <div className="cta-download" style={{opacity: r(local, 92, 118)}}>
        <Shake local={local} intensity={0.4}>
          <span>DOWNLOAD BEFORE YOU CLOSE THIS</span>
        </Shake>
      </div>
    </>
  );
}

function Hud() {
  const frame = useCurrentFrame();
  const progress = frame / TOTAL;
  return (
    <div className="hype-hud">
      <div className="hud-rate">
        <span>CONDURA</span>
        <b>{Math.floor(frame / 30).toString().padStart(2, '0')}s</b>
      </div>
      <div className="hud-bar">
        <i style={{transform: `scaleX(${progress})`}} />
      </div>
      <div className="hud-beats">
        {beats.map((beat) => {
          const active = frame >= beat.start && frame < beat.start + beat.duration;
          return (
            <span key={beat.label} className={active ? 'on' : ''}>
              {beat.label}
            </span>
          );
        })}
      </div>
    </div>
  );
}

function Slashes() {
  const frame = useCurrentFrame();
  const cuts = beats.slice(1).map((beat) => beat.start);
  return (
    <>
      {cuts.map((cut) => {
        const amount = interpolate(Math.abs(frame - cut), [0, 12], [1, 0], {
          extrapolateLeft: 'clamp',
          extrapolateRight: 'clamp',
        });
        return (
          <div
            key={cut}
            className="hype-slash"
            style={{
              opacity: amount,
              transform: `translateX(${interpolate(amount, [0, 1], [-1100, 0])}px) skewX(-18deg)`,
            }}
          />
        );
      })}
    </>
  );
}

export const ConduraHypeReel = () => {
  return (
    <AbsoluteFill className="hype-root">
      <Beat start={0} duration={90} tone="dark">
        {(local) => <HookScene local={local} />}
      </Beat>
      <Beat start={90} duration={75} tone="blood">
        {(local) => <StingScene local={local} />}
      </Beat>
      <Beat start={165} duration={90} tone="dark">
        {(local) => <ChaosScene local={local} />}
      </Beat>
      <Beat start={255} duration={75} tone="acid">
        {(local) => <RevealScene local={local} />}
      </Beat>
      <Beat start={330} duration={120} tone="dark">
        {(local) => <PowerScene local={local} />}
      </Beat>
      <Beat start={450} duration={90} tone="blood">
        {(local) => <FomoScene local={local} />}
      </Beat>
      <Beat start={540} duration={120} tone="acid">
        {(local) => <CtaScene local={local} />}
      </Beat>
      {beats.slice(1).map((beat) => (
        <Flash key={beat.start} at={beat.start} color={beat.tone === 'acid' ? acid : beat.tone === 'blood' ? magenta : cyan} />
      ))}
      <Slashes />
      <div className="hype-grain" />
      <div className="hype-scanlines" />
      <Hud />
    </AbsoluteFill>
  );
};

const _unused: CSSProperties = {};
void _unused;