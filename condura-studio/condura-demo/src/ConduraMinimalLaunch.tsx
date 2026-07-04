import type {ReactNode} from 'react';
import {
  AbsoluteFill,
  Easing,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from 'remotion';

const TOTAL_FRAMES = 1680;

const scenes = [
  {name: 'open', start: 0, duration: 210},
  {name: 'name', start: 210, duration: 240},
  {name: 'layer', start: 450, duration: 300},
  {name: 'flow', start: 750, duration: 330},
  {name: 'control', start: 1080, duration: 300},
  {name: 'close', start: 1380, duration: 300},
] as const;

const ease = Easing.bezier(0.16, 1, 0.3, 1);
const soft = Easing.bezier(0.22, 1, 0.36, 1);

function fade(frame: number, from: number, to: number): number {
  return interpolate(frame, [from, to], [0, 1], {
    easing: ease,
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
}

function out(frame: number, from: number, to: number): number {
  return interpolate(frame, [from, to], [1, 0], {
    easing: Easing.in(Easing.cubic),
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
}

function textReveal(frame: number, start: number, step = 18): string {
  const marks = ['Context', 'Route', 'Action', 'Receipt'];
  const index = Math.min(marks.length - 1, Math.max(0, Math.floor((frame - start) / step)));
  return marks[index];
}

function Scene({
  start,
  duration,
  children,
  className = '',
}: {
  start: number;
  duration: number;
  className?: string;
  children: (local: number) => ReactNode;
}) {
  const frame = useCurrentFrame();
  const live = frame >= start - 30 && frame <= start + duration + 30;
  if (!live) return null;

  const local = frame - start;
  const opacity = Math.min(fade(frame, start, start + 38), out(frame, start + duration - 42, start + duration));
  const y = interpolate(opacity, [0, 1], [18, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <AbsoluteFill className={`min-scene ${className}`} style={{opacity, transform: `translateY(${y}px)`}}>
      {children(local)}
    </AbsoluteFill>
  );
}

function Background() {
  const frame = useCurrentFrame();
  const drift = Math.sin(frame / 110) * 16;
  return (
    <AbsoluteFill className="min-bg">
      <div className="min-bg__wash" style={{transform: `translate3d(${drift}px, ${-drift * 0.4}px, 0)`}} />
      <div className="min-bg__wash min-bg__wash--second" style={{transform: `translate3d(${-drift * 0.6}px, ${drift * 0.5}px, 0)`}} />
      <div className="min-bg__grain" />
    </AbsoluteFill>
  );
}

function Mark({quiet = false}: {quiet?: boolean}) {
  return (
    <div className={`min-mark ${quiet ? 'min-mark--quiet' : ''}`}>
      <span>C</span>
      <i />
    </div>
  );
}

function Caption({children}: {children: ReactNode}) {
  return <div className="min-caption">{children}</div>;
}

function OpeningScene({local}: {local: number}) {
  const line = fade(local, 78, 142);
  return (
    <div className="min-center min-open">
      <Caption>pre-launch film</Caption>
      <h1>
        The computer can answer.
        <br />
        The work still lives everywhere.
      </h1>
      <p style={{opacity: fade(local, 92, 132)}}>
        Condura is a quiet command layer for the tools already around you.
      </p>
      <div className="min-line" style={{transform: `scaleX(${line})`}} />
    </div>
  );
}

function NameScene({local}: {local: number}) {
  const {fps} = useVideoConfig();
  const rise = spring({
    frame: local - 18,
    fps,
    config: {damping: 18, stiffness: 70, mass: 0.9},
  });
  return (
    <div className="min-center min-name">
      <div style={{opacity: rise, transform: `translateY(${interpolate(rise, [0, 1], [20, 0])}px)`}}>
        <Mark />
      </div>
      <h2 style={{opacity: fade(local, 46, 92)}}>Condura</h2>
      <p style={{opacity: fade(local, 82, 128)}}>
        One place to call your models, agents, files, and desktop.
      </p>
    </div>
  );
}

function LayerScene({local}: {local: number}) {
  const sweep = fade(local, 68, 180);
  const center = {x: 1160, y: 548};
  const nodes = [
    ['models', 46, 850, 318],
    ['agents', 72, 810, 655],
    ['files', 96, 1328, 318],
    ['browser', 120, 1510, 646],
    ['voice', 144, 1092, 794],
  ] as const;
  return (
    <div className="min-layer">
      <div className="min-layer__copy">
        <Caption>not another window</Caption>
        <h2>A layer above the stack.</h2>
        <p>
          Condura does not replace the tools. It gives them a shared moment of attention.
        </p>
      </div>
      <div className="min-map">
        <div
          className="min-map__ring"
          style={{
            left: center.x,
            top: center.y,
            transform: `translate(-50%, -50%) scale(${0.85 + sweep * 0.15})`,
          }}
        />
        <div className="min-map__core" style={{left: center.x, top: center.y}}>
          <Mark quiet />
        </div>
        {nodes.map(([label, at, x, y]) => {
          const show = fade(local, at, at + 32);
          return (
            <div
              key={label}
              className="min-node"
              style={{
                left: x,
                top: y,
                opacity: show,
                transform: `translate(-50%, -50%) scale(${0.94 + show * 0.06})`,
              }}
            >
              {label}
            </div>
          );
        })}
        <svg className="min-map__lines" viewBox="0 0 1920 1080">
          {nodes.map(([label, at, x, y]) => {
            const show = fade(local, at + 12, at + 66);
            return (
              <line
                key={label}
                x1={center.x}
                y1={center.y}
                x2={x}
                y2={y}
                stroke="rgba(44,42,36,.22)"
                strokeWidth="1.5"
                strokeDasharray="1"
                style={{opacity: show}}
              />
            );
          })}
        </svg>
      </div>
    </div>
  );
}

function FlowScene({local}: {local: number}) {
  const progress = fade(local, 70, 246);
  const label = textReveal(local, 84, 42);
  const steps = [
    ['context', 'what is already on your screen', 46],
    ['route', 'the right model or agent', 92],
    ['action', 'prepared, never hidden', 138],
    ['receipt', 'what happened, saved clearly', 184],
  ] as const;

  return (
    <div className="min-flow">
      <Caption>one request becomes a line of work</Caption>
      <h2>{label}</h2>
      <div className="min-flow__rail">
        <i style={{transform: `scaleX(${progress})`}} />
      </div>
      <div className="min-flow__steps">
        {steps.map(([title, body, at], index) => {
          const show = fade(local, at, at + 30);
          return (
            <div key={title} style={{opacity: show, transform: `translateY(${interpolate(show, [0, 1], [18, 0])}px)`}}>
              <span>{`0${index + 1}`}</span>
              <strong>{title}</strong>
              <p>{body}</p>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function ControlScene({local}: {local: number}) {
  const gate = fade(local, 86, 168);
  return (
    <div className="min-control">
      <div className="min-control__copy">
        <Caption>fast, with a boundary</Caption>
        <h2>The human stays in the loop.</h2>
        <p>
          Local when it can. Cloud when you choose. Approval before anything touches the machine.
        </p>
      </div>
      <div className="min-gate" style={{opacity: fade(local, 62, 112)}}>
        <div className="min-gate__circle" style={{transform: `rotate(${local * 0.18}deg) scale(${0.92 + gate * 0.08})`}}>
          <span>approve</span>
          <span>pause</span>
          <span>receipt</span>
        </div>
        <div className="min-gate__center">
          <Mark quiet />
        </div>
      </div>
    </div>
  );
}

function CloseScene({local}: {local: number}) {
  const {fps} = useVideoConfig();
  const settle = spring({
    frame: local - 18,
    fps,
    config: {damping: 20, stiffness: 62, mass: 1},
  });
  return (
    <div className="min-center min-close">
      <div style={{opacity: settle}}>
        <Mark />
      </div>
      <h2 style={{opacity: fade(local, 52, 100)}}>Condura</h2>
      <p style={{opacity: fade(local, 92, 142)}}>
        A command layer for the computer you already use.
      </p>
      <div className="min-launch" style={{opacity: fade(local, 138, 192)}}>
        <span>launching soon</span>
        <strong>condura.app</strong>
      </div>
    </div>
  );
}

function Hud() {
  const frame = useCurrentFrame();
  return (
    <div className="min-hud">
      <span>Condura</span>
      <i style={{transform: `scaleX(${frame / TOTAL_FRAMES})`}} />
      <b>{Math.floor(frame / 30).toString().padStart(2, '0')}s</b>
    </div>
  );
}

export const ConduraMinimalLaunch = () => {
  return (
    <AbsoluteFill className="minimal-root">
      <Background />
      <Scene start={0} duration={210}>
        {(local) => <OpeningScene local={local} />}
      </Scene>
      <Scene start={210} duration={240}>
        {(local) => <NameScene local={local} />}
      </Scene>
      <Scene start={450} duration={300}>
        {(local) => <LayerScene local={local} />}
      </Scene>
      <Scene start={750} duration={330}>
        {(local) => <FlowScene local={local} />}
      </Scene>
      <Scene start={1080} duration={300}>
        {(local) => <ControlScene local={local} />}
      </Scene>
      <Scene start={1380} duration={300}>
        {(local) => <CloseScene local={local} />}
      </Scene>
      <Hud />
    </AbsoluteFill>
  );
};
