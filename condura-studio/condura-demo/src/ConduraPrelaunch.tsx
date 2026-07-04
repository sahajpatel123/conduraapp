import type {CSSProperties, ReactNode} from 'react';
import {
  AbsoluteFill,
  Easing,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from 'remotion';

const TOTAL = 1620;

const beats = [
  {label: 'Signal', start: 0, duration: 150},
  {label: 'Summon', start: 150, duration: 210},
  {label: 'Collapse', start: 360, duration: 270},
  {label: 'Conductor', start: 630, duration: 255},
  {label: 'Rules', start: 885, duration: 255},
  {label: 'Launch', start: 1140, duration: 300},
  {label: 'Name', start: 1440, duration: 180},
] as const;

const easeOut = Easing.bezier(0.16, 1, 0.3, 1);
const hardEase = Easing.bezier(0.76, 0, 0.24, 1);

function r(frame: number, from: number, to: number, easing = easeOut): number {
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
  tone?: 'dark' | 'light' | 'acid';
}) {
  const frame = useCurrentFrame();
  const fade = 12;
  const live = frame >= start - fade && frame <= start + duration + fade;
  if (!live) return null;

  const enter = r(frame, start, start + fade, Easing.out(Easing.cubic));
  const exit = interpolate(frame, [start + duration - fade, start + duration], [1, 0], {
    easing: Easing.in(Easing.cubic),
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const opacity = Math.min(enter, exit);
  const shove = interpolate(enter, [0, 1], [36, 0]);

  return (
    <AbsoluteFill
      className={`pre-beat pre-beat-${tone} ${className}`}
      style={{opacity, transform: `translate3d(0, ${shove}px, 0)`}}
    >
      {children(frame - start)}
    </AbsoluteFill>
  );
}

function Noise() {
  return (
    <>
      <div className="pre-noise" />
      <div className="pre-scanlines" />
    </>
  );
}

function Flash({at, color = '#f4ff5c'}: {at: number; color?: string}) {
  const frame = useCurrentFrame();
  const amount = interpolate(Math.abs(frame - at), [0, 8], [0.75, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  return <AbsoluteFill style={{background: color, opacity: amount, mixBlendMode: 'screen'}} />;
}

function SplitText({
  children,
  className = '',
  delay = 0,
  local,
  size = 'huge',
}: {
  children: string;
  className?: string;
  delay?: number;
  local: number;
  size?: 'mega' | 'huge' | 'medium';
}) {
  const show = r(local, delay, delay + 24);
  const glitch = Math.sin((local + delay) * 0.9) > 0.72 ? 1 : 0;
  const x = interpolate(show, [0, 1], [-90, 0]);
  return (
    <div
      className={`split-text split-${size} ${className}`}
      style={{
        opacity: show,
        transform: `translate3d(${x}px, 0, 0) skewX(${glitch ? -4 : 0}deg)`,
      }}
    >
      <span>{children}</span>
      <span aria-hidden style={{transform: `translate3d(${glitch ? 14 : 0}px, ${glitch ? -5 : 0}px, 0)`}}>
        {children}
      </span>
      <span aria-hidden style={{transform: `translate3d(${glitch ? -12 : 0}px, ${glitch ? 7 : 0}px, 0)`}}>
        {children}
      </span>
    </div>
  );
}

function OrbitalField({local, intense = false}: {local: number; intense?: boolean}) {
  const count = intense ? 18 : 12;
  return (
    <div className="orbital-field">
      {Array.from({length: count}).map((_, index) => {
        const radius = 130 + index * (intense ? 34 : 42);
        const spin = local * (0.42 + index * 0.018);
        const angle = spin + index * 41;
        const size = 12 + (index % 4) * 8;
        return (
          <i
            key={index}
            style={{
              width: size,
              height: size,
              transform: `translate(-50%, -50%) rotate(${angle}deg) translateX(${radius}px) rotate(${-angle}deg)`,
              background: index % 3 === 0 ? '#f4ff5c' : index % 3 === 1 ? '#00f0ff' : '#ff3df2',
              opacity: 0.5 + (index % 5) * 0.08,
            }}
          />
        );
      })}
    </div>
  );
}

function WordStorm({local}: {local: number}) {
  const words = [
    'CHAT',
    'CODE',
    'FILES',
    'BROWSER',
    'VOICE',
    'AGENTS',
    'LOCAL',
    'APPROVAL',
    'REPLAY',
    'CONTROL',
  ];

  return (
    <div className="word-storm">
      {words.map((word, index) => {
        const show = r(local, index * 8, index * 8 + 18);
        const slide = interpolate(show, [0, 1], [180, 0]);
        const y = 120 + (index % 5) * 150;
        const x = index % 2 === 0 ? 80 + index * 98 : 960 + index * 64;
        const rotate = index % 2 === 0 ? -5 : 4;
        return (
          <strong
            key={word}
            style={{
              left: x,
              top: y,
              opacity: show,
              transform: `translate3d(${index % 2 === 0 ? -slide : slide}px, 0, 0) rotate(${rotate}deg)`,
            }}
          >
            {word}
          </strong>
        );
      })}
    </div>
  );
}

function HotkeyGlyph({local}: {local: number}) {
  const hit = r(local, 24, 62, hardEase);
  const second = r(local, 82, 126, hardEase);
  return (
    <div className="hotkey-glyph">
      <div className="hotkey-caps" style={{transform: `scale(${1 - hit * 0.07 + second * 0.04})`}}>
        <span>CMD</span>
        <span>SHIFT</span>
        <span>SPACE</span>
      </div>
      <div
        className="shock shock-a"
        style={{
          opacity: interpolate(hit, [0, 0.2, 1], [0, 1, 0]),
          transform: `translate(-50%, -50%) scale(${interpolate(hit, [0, 1], [0.25, 2.8])})`,
        }}
      />
      <div
        className="shock shock-b"
        style={{
          opacity: interpolate(second, [0, 0.2, 1], [0, 1, 0]),
          transform: `translate(-50%, -50%) scale(${interpolate(second, [0, 1], [0.15, 4.8])}) rotate(${local * 3}deg)`,
        }}
      />
    </div>
  );
}

function SignalScene({local}: {local: number}) {
  const collapse = r(local, 96, 142, hardEase);
  return (
    <>
      <div className="signal-bg" style={{transform: `scale(${1 + collapse * 0.28}) rotate(${collapse * 2}deg)`}} />
      <OrbitalField local={local} />
      <SplitText local={local} delay={12} size="mega">
        STOP
      </SplitText>
      <SplitText local={local} delay={38} size="huge" className="right-text">
        OPENING
      </SplitText>
      <SplitText local={local} delay={64} size="huge" className="lower-text">
        AI APPS
      </SplitText>
      <div className="tiny-line" style={{opacity: r(local, 104, 132)}}>
        Another tab is not the future.
      </div>
      <div className="countdown-stack" style={{opacity: r(local, 20, 60)}}>
        {['17 windows', '9 contexts', '4 agents', '1 brain'].map((line, index) => (
          <span key={line} style={{transform: `translateX(${pulse(local + index * 8, 8, -10, 10)}px)`}}>
            {line}
          </span>
        ))}
      </div>
    </>
  );
}

function SummonScene({local}: {local: number}) {
  const {fps} = useVideoConfig();
  const boom = spring({
    frame: local - 48,
    fps,
    config: {damping: 9, stiffness: 130, mass: 0.7},
  });
  return (
    <>
      <div className="summon-bg" />
      <HotkeyGlyph local={local} />
      <div
        className="summon-title"
        style={{
          opacity: r(local, 64, 94),
          transform: `translate(-50%, -50%) scale(${0.7 + boom * 0.3})`,
        }}
      >
        <span>ONE KEY</span>
        <strong>THE ROOM CHANGES</strong>
      </div>
      <div className="command-beam" style={{transform: `scaleX(${r(local, 112, 178, hardEase)})`}} />
      <div className="typed-command" style={{opacity: r(local, 130, 164)}}>
        {type('summon every capable system, but keep me in control', local, 132, 190)}
        <i />
      </div>
      <div className="summon-shards">
        {Array.from({length: 22}).map((_, index) => {
          const fly = r(local, 70 + index * 2, 138 + index * 2);
          return (
            <i
              key={index}
              style={{
                left: `${8 + (index * 37) % 86}%`,
                top: `${12 + (index * 53) % 76}%`,
                opacity: fly,
                transform: `translate3d(${interpolate(fly, [0, 1], [0, (index % 2 ? -1 : 1) * 260])}px, ${interpolate(fly, [0, 1], [0, (index % 3 - 1) * 180])}px, 0) rotate(${local * (index % 2 ? -2 : 2)}deg)`,
              }}
            />
          );
        })}
      </div>
    </>
  );
}

function CollapseScene({local}: {local: number}) {
  const tunnel = r(local, 120, 230, hardEase);
  const labels = ['models', 'agents', 'files', 'browser', 'voice', 'memory'];
  return (
    <>
      <div className="collapse-bg" />
      <WordStorm local={local} />
      <div className="black-hole" style={{transform: `translate(-50%, -50%) scale(${0.82 + tunnel * 0.62})`}}>
        <div className="hole-core">C</div>
        {Array.from({length: 8}).map((_, index) => (
          <span
            key={index}
            style={{
              transform: `translate(-50%, -50%) rotate(${local * (2 + index * 0.2) + index * 45}deg) scale(${1 + index * 0.13})`,
            }}
          />
        ))}
      </div>
      <div className="collapse-line" style={{opacity: r(local, 170, 214)}}>
        Everything becomes one command.
      </div>
      <div className="tool-comets">
        {labels.map((label, index) => {
          const fall = r(local, 54 + index * 12, 150 + index * 12, hardEase);
          return (
            <b
              key={label}
              style={{
                opacity: fall,
                transform: `translate3d(${interpolate(fall, [0, 1], [index % 2 ? 760 : -760, 0])}px, ${interpolate(fall, [0, 1], [(index - 2) * 110, 0])}px, 0) rotate(${interpolate(fall, [0, 1], [index % 2 ? 18 : -18, 0])}deg)`,
              }}
            >
              {label}
            </b>
          );
        })}
      </div>
    </>
  );
}

function ConductorScene({local}: {local: number}) {
  const reveal = r(local, 28, 78);
  const rail = r(local, 106, 202, hardEase);
  const verbs = [
    ['sees context', 34],
    ['chooses the path', 70],
    ['asks before action', 106],
    ['leaves a receipt', 142],
  ] as const;
  return (
    <>
      <div className="conductor-bg" />
      <div className="mono-halo" style={{opacity: reveal, transform: `translate(-50%, -50%) scale(${0.6 + reveal * 0.4})`}}>
        <i />
        <i />
        <i />
      </div>
      <div className="conductor-word" style={{opacity: reveal}}>
        <span>Not a chatbot.</span>
        <strong>A conductor.</strong>
      </div>
      <div className="verb-rails">
        {verbs.map(([label, at], index) => {
          const show = r(local, at, at + 22);
          return (
            <div className="verb" key={label} style={{opacity: show, transform: `translateX(${interpolate(show, [0, 1], [index % 2 ? 80 : -80, 0])}px)`}}>
              <b>{`0${index + 1}`}</b>
              <span>{label}</span>
            </div>
          );
        })}
      </div>
      <div className="conductor-rail" style={{transform: `scaleX(${rail})`}} />
      <div className="conductor-caption" style={{opacity: r(local, 184, 232)}}>
        Your stack keeps its power. Condura gives it a nervous system.
      </div>
    </>
  );
}

function RulesScene({local}: {local: number}) {
  const gate = r(local, 96, 160, hardEase);
  return (
    <>
      <div className="rules-bg" />
      <div className="rule-title" style={{opacity: r(local, 10, 48)}}>
        <span>FAST</span>
        <strong>WITHOUT LOSING THE WHEEL</strong>
      </div>
      <div className="rule-cards">
        {[
          ['local when it can', '#f4ff5c'],
          ['cloud when you choose', '#00f0ff'],
          ['asks before touching', '#ff7a1a'],
          ['hard stop always visible', '#ff3d5e'],
        ].map(([label, color], index) => {
          const show = r(local, 48 + index * 18, 86 + index * 18);
          return (
            <div
              key={label}
              style={{
                opacity: show,
                borderColor: color,
                transform: `translateY(${interpolate(show, [0, 1], [80, 0])}px) rotate(${index % 2 ? 2 : -2}deg)`,
              }}
            >
              <i style={{background: color}} />
              <span>{label}</span>
            </div>
          );
        })}
      </div>
      <div className="gate-ring" style={{opacity: gate, transform: `translate(-50%, -50%) scale(${0.45 + gate * 1.1}) rotate(${local * 2}deg)`}}>
        <span>APPROVAL</span>
        <span>RECEIPT</span>
        <span>HALT</span>
      </div>
      <div className="rules-footer" style={{opacity: r(local, 190, 228)}}>
        The machine can move fast. The human stays first.
      </div>
    </>
  );
}

function LaunchScene({local}: {local: number}) {
  const launch = r(local, 56, 170, hardEase);
  const planet = r(local, 142, 232, hardEase);
  return (
    <>
      <div className="launch-bg" />
      <OrbitalField local={local * 1.8} intense />
      <div className="launch-stack">
        <SplitText local={local} delay={18} size="medium">
          PRE-LAUNCH SIGNAL
        </SplitText>
        <div className="launch-main" style={{opacity: r(local, 52, 90), transform: `translateY(${interpolate(r(local, 52, 90), [0, 1], [80, 0])}px)`}}>
          <span>THE COMMAND LAYER</span>
          <strong>FOR YOUR COMPUTER</strong>
          <em>IS ABOUT TO WAKE UP</em>
        </div>
      </div>
      <div className="launch-tunnel" style={{opacity: launch}}>
        {Array.from({length: 11}).map((_, index) => (
          <i
            key={index}
            style={{
              transform: `translate(-50%, -50%) scale(${0.2 + index * 0.17 + launch * 1.4}) rotate(${local * (index % 2 ? 1 : -1)}deg)`,
              opacity: Math.max(0, 0.62 - index * 0.036),
            }}
          />
        ))}
      </div>
      <div className="launch-planet" style={{opacity: planet, transform: `translate(-50%, -50%) scale(${0.6 + planet * 0.4})`}}>
        <b>one key</b>
        <span>many systems</span>
      </div>
    </>
  );
}

function NameScene({local}: {local: number}) {
  const {fps} = useVideoConfig();
  const mark = spring({
    frame: local - 10,
    fps,
    config: {damping: 8, stiffness: 90, mass: 0.75},
  });
  const glow = pulse(local, 18, 0.4, 1);
  return (
    <>
      <div className="name-bg" />
      <div className="final-mark" style={{opacity: mark, transform: `translate(-50%, -50%) scale(${0.65 + mark * 0.35})`}}>
        <div className="final-c" style={{boxShadow: `0 0 ${80 + glow * 80}px rgba(244, 255, 92, .45)`}}>C</div>
        <h1>Condura</h1>
        <p>One command layer above your models, agents, files, and desktop.</p>
        <div className="launch-date" style={{opacity: r(local, 82, 118)}}>
          launching soon
        </div>
        <div className="site" style={{opacity: r(local, 104, 142)}}>
          condura.app
        </div>
      </div>
    </>
  );
}

function TrailerHud() {
  const frame = useCurrentFrame();
  const progress = frame / TOTAL;
  return (
    <div className="trailer-hud">
      <div className="hud-top">
        <span>Condura pre-launch transmission</span>
        <b>{Math.floor(frame / 30).toString().padStart(2, '0')}s</b>
      </div>
      <div className="hud-progress">
        <i style={{transform: `scaleX(${progress})`}} />
      </div>
      <div className="hud-beats">
        {beats.map((beat) => {
          const active = frame >= beat.start && frame < beat.start + beat.duration;
          return (
            <span key={beat.label} className={active ? 'active' : ''}>
              {beat.label}
            </span>
          );
        })}
      </div>
    </div>
  );
}

function TransitionSlashes() {
  const frame = useCurrentFrame();
  const cuts = [150, 360, 630, 885, 1140, 1440];
  return (
    <>
      {cuts.map((cut) => {
        const amount = interpolate(Math.abs(frame - cut), [0, 16], [1, 0], {
          extrapolateLeft: 'clamp',
          extrapolateRight: 'clamp',
        });
        return (
          <div
            className="transition-slash"
            key={cut}
            style={{
              opacity: amount,
              transform: `translateX(${interpolate(amount, [0, 1], [-1200, 0])}px) skewX(-16deg)`,
            }}
          />
        );
      })}
    </>
  );
}

export const ConduraPrelaunch = () => {
  return (
    <AbsoluteFill className="prelaunch-root">
      <Beat start={0} duration={150} tone="dark">
        {(local) => <SignalScene local={local} />}
      </Beat>
      <Beat start={150} duration={210} tone="acid">
        {(local) => <SummonScene local={local} />}
      </Beat>
      <Beat start={360} duration={270} tone="dark">
        {(local) => <CollapseScene local={local} />}
      </Beat>
      <Beat start={630} duration={255} tone="light">
        {(local) => <ConductorScene local={local} />}
      </Beat>
      <Beat start={885} duration={255} tone="dark">
        {(local) => <RulesScene local={local} />}
      </Beat>
      <Beat start={1140} duration={300} tone="acid">
        {(local) => <LaunchScene local={local} />}
      </Beat>
      <Beat start={1440} duration={180} tone="dark">
        {(local) => <NameScene local={local} />}
      </Beat>
      {[150, 360, 630, 885, 1140, 1440].map((at) => (
        <Flash key={at} at={at} />
      ))}
      <TransitionSlashes />
      <Noise />
      <TrailerHud />
    </AbsoluteFill>
  );
};
