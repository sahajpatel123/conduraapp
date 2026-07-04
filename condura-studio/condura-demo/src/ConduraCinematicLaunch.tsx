import type {ReactNode} from 'react';
import {AbsoluteFill, Easing, interpolate, spring, useCurrentFrame, useVideoConfig} from 'remotion';

const TOTAL_FRAMES = 810;

const scenes = [
  {name: 'scatter', start: 0, end: 150},
  {name: 'summon', start: 150, end: 270},
  {name: 'logic', start: 270, end: 610},
  {name: 'launch', start: 610, end: 810},
] as const;

const ease = Easing.bezier(0.16, 1, 0.3, 1);
const snap = Easing.bezier(0.74, 0, 0.2, 1);

function fade(frame: number, from: number, to: number) {
  return interpolate(frame, [from, to], [0, 1], {
    easing: ease,
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
}

function fadeOut(frame: number, from: number, to: number) {
  return interpolate(frame, [from, to], [1, 0], {
    easing: Easing.in(Easing.cubic),
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
}

function travel(frame: number, from: number, to: number) {
  return interpolate(frame, [from, to], [0, 1], {
    easing: snap,
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
}

function Scene({
  start,
  end,
  children,
}: {
  start: number;
  end: number;
  children: (local: number) => ReactNode;
}) {
  const frame = useCurrentFrame();

  if (frame < start - 18 || frame > end + 18) {
    return null;
  }

  const local = frame - start;
  const opacity = Math.min(fade(frame, start, start + 26), fadeOut(frame, end - 30, end + 8));
  const y = interpolate(opacity, [0, 1], [26, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <AbsoluteFill className="logic-scene" style={{opacity, transform: `translate3d(0, ${y}px, 0)`}}>
      {children(local)}
    </AbsoluteFill>
  );
}

function Backdrop() {
  const frame = useCurrentFrame();
  const drift = Math.sin(frame / 120) * 34;

  return (
    <AbsoluteFill className="logic-backdrop">
      <div
        className="logic-backdrop__wash logic-backdrop__wash--one"
        style={{transform: `translate3d(${drift}px, ${-drift * 0.4}px, 0)`}}
      />
      <div
        className="logic-backdrop__wash logic-backdrop__wash--two"
        style={{transform: `translate3d(${-drift * 0.6}px, ${drift * 0.35}px, 0)`}}
      />
      <div className="logic-backdrop__grid" />
      <div className="logic-backdrop__grain" />
    </AbsoluteFill>
  );
}

function Mark({large = false}: {large?: boolean}) {
  return (
    <div className={`logic-mark ${large ? 'logic-mark--large' : ''}`}>
      <span>C</span>
      <i />
    </div>
  );
}

function FloatingTool({
  label,
  title,
  x,
  y,
  rotate,
  delay,
  local,
}: {
  label: string;
  title: string;
  x: number;
  y: number;
  rotate: number;
  delay: number;
  local: number;
}) {
  const show = fade(local, delay, delay + 34);
  const pull = travel(local, 66, 126);
  const drift = Math.sin((local + delay * 2) / 28) * 14;
  const centerX = 960;
  const centerY = 625;
  const left = interpolate(pull, [0, 1], [x + drift, centerX + (x - centerX) * 0.16], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const top = interpolate(pull, [0, 1], [y - drift * 0.3, centerY + (y - centerY) * 0.12], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const blur = interpolate(pull, [0, 1], [2.4, 5.5], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <div
      className="logic-tool"
      style={{
        left,
        top,
        opacity: show * fadeOut(local, 128, 154),
        filter: `blur(${blur}px)`,
        transform: `translate(-50%, -50%) rotate(${rotate - pull * rotate * 0.7}deg) scale(${1 - pull * 0.14})`,
      }}
    >
      <span>{label}</span>
      <strong>{title}</strong>
      <p />
      <p />
      <p />
    </div>
  );
}

function ScatterScene({local}: {local: number}) {
  const words = ['models', 'files', 'agents', 'browser'];
  const word = words[Math.min(words.length - 1, Math.max(0, Math.floor((local - 36) / 20)))];

  return (
    <div className="logic-scatter">
      <FloatingTool local={local} delay={0} x={350} y={270} rotate={-15} label="model" title="summarize this" />
      <FloatingTool local={local} delay={10} x={1390} y={260} rotate={12} label="browser" title="market notes" />
      <FloatingTool local={local} delay={20} x={560} y={884} rotate={-6} label="file" title="launch-plan.md" />
      <FloatingTool local={local} delay={30} x={1370} y={910} rotate={8} label="agent" title="research loop" />
      <div className="logic-scatter__center" style={{opacity: fade(local, 18, 46)}}>
        <h1>
          One request.
          <br />
          Too many <span>{word}</span>.
        </h1>
      </div>
      <div className="logic-pulse" style={{opacity: fade(local, 80, 118)}} />
    </div>
  );
}

function SummonScene({local}: {local: number}) {
  const {fps} = useVideoConfig();
  const settle = spring({
    frame: local - 10,
    fps,
    config: {damping: 18, stiffness: 94, mass: 0.92},
  });
  const bar = travel(local, 58, 114);
  const text = 'plan the launch using everything open';
  const count = Math.floor(interpolate(local, [48, 118], [0, text.length], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  }));

  return (
    <div className="logic-summon">
      <div className="logic-summon__brand" style={{opacity: settle, transform: `scale(${0.9 + settle * 0.1})`}}>
        <Mark large />
        <h1>Condura</h1>
      </div>
      <div
        className="logic-command"
        style={{opacity: fade(local, 40, 82), transform: `translateX(-50%) scaleX(${0.72 + bar * 0.28})`}}
      >
        <span>command</span>
        <strong>
          {text.slice(0, count)}
          <i />
        </strong>
      </div>
      <p style={{opacity: fade(local, 104, 142)}}>A command layer for the tools already around you.</p>
    </div>
  );
}

function LogicNode({
  label,
  title,
  x,
  y,
  active,
  show,
}: {
  label: string;
  title: string;
  x: number;
  y: number;
  active: boolean;
  show: number;
}) {
  return (
    <div
      className={`logic-node ${active ? 'is-active' : ''}`}
      style={{
        left: x,
        top: y,
        opacity: show,
        transform: `translate(-50%, -50%) scale(${0.9 + show * 0.1})`,
      }}
    >
      <span>{label}</span>
      <strong>{title}</strong>
    </div>
  );
}

function LogicScene({local}: {local: number}) {
  const path = travel(local, 34, 246);
  const approval = fade(local, 220, 282);
  const activeIndex = Math.min(4, Math.max(0, Math.floor((local - 42) / 54)));
  const nodes = [
    ['01', 'Context', 354, 520],
    ['02', 'Route', 670, 360],
    ['03', 'Prepare', 982, 540],
    ['04', 'Approve', 1268, 362],
    ['05', 'Receipt', 1518, 584],
  ] as const;

  return (
    <div className="logic-map">
      <div className="logic-map__copy">
        <span>Condura logic</span>
        <h2>It does not replace your tools. It coordinates them.</h2>
      </div>
      <svg className="logic-map__path" viewBox="0 0 1920 1440">
        <path
          d="M354 520 C518 266 760 274 670 360 C828 548 878 612 982 540 C1118 402 1220 248 1268 362 C1332 528 1438 606 1518 584"
          pathLength="1"
          style={{strokeDashoffset: 1 - path}}
        />
      </svg>
      {nodes.map(([label, title, x, y], index) => {
        const show = fade(local, 34 + index * 32, 78 + index * 32);
        return <LogicNode key={label} label={label} title={title} x={x} y={y} active={index === activeIndex} show={show} />;
      })}
      <div className="logic-decision" style={{opacity: fade(local, 122, 180)}}>
        <span>best route</span>
        <strong>{activeIndex < 2 ? 'local model' : activeIndex < 4 ? 'agent plan' : 'saved receipt'}</strong>
      </div>
      <div
        className="logic-approval"
        style={{
          opacity: approval,
          transform: `translate(-50%, -50%) scale(${0.82 + approval * 0.18})`,
        }}
      >
        <b>approve</b>
        <i />
      </div>
      <div className="logic-receipt" style={{opacity: fade(local, 282, 330)}}>
        <span>receipt saved</span>
        <strong>what changed, why it changed, and what to undo</strong>
      </div>
    </div>
  );
}

function LaunchScene({local}: {local: number}) {
  const {fps} = useVideoConfig();
  const settle = spring({
    frame: local - 8,
    fps,
    config: {damping: 20, stiffness: 78, mass: 1},
  });
  const line = travel(local, 76, 132);

  return (
    <div className="logic-launch">
      <div className="logic-launch__mark" style={{opacity: settle, transform: `scale(${0.9 + settle * 0.1})`}}>
        <Mark large />
      </div>
      <h1 style={{opacity: fade(local, 34, 78)}}>Condura</h1>
      <p style={{opacity: fade(local, 78, 118)}}>Your desktop, finally in one command.</p>
      <div className="logic-launch__line">
        <i style={{transform: `scaleX(${line})`}} />
      </div>
      <div className="logic-launch__cta" style={{opacity: fade(local, 126, 170)}}>
        <span>launching soon</span>
        <strong>condura.app</strong>
      </div>
    </div>
  );
}

function TransitionLayer() {
  const frame = useCurrentFrame();
  const cuts = [150, 270, 610] as const;

  return (
    <AbsoluteFill className="logic-transitions">
      {cuts.map((point) => {
        const show = Math.min(fade(frame, point - 16, point), fadeOut(frame, point + 2, point + 32));
        const sweep = travel(frame, point - 18, point + 22);
        return (
          <div key={point} className="logic-wipe" style={{opacity: show}}>
            <i style={{transform: `translateX(${interpolate(sweep, [0, 1], [-116, 116])}%) rotate(-10deg)`}} />
          </div>
        );
      })}
    </AbsoluteFill>
  );
}

function ProgressMarks() {
  const frame = useCurrentFrame();
  const active = scenes.find((scene) => frame >= scene.start && frame < scene.end) ?? scenes[scenes.length - 1];

  return (
    <div className="logic-progress">
      <span>Condura</span>
      <i style={{transform: `scaleX(${frame / TOTAL_FRAMES})`}} />
      <b>{active.name}</b>
    </div>
  );
}

export const ConduraCinematicLaunch = () => {
  return (
    <AbsoluteFill className="logic-root">
      <Backdrop />
      <Scene start={0} end={150}>
        {(local) => <ScatterScene local={local} />}
      </Scene>
      <Scene start={150} end={270}>
        {(local) => <SummonScene local={local} />}
      </Scene>
      <Scene start={270} end={610}>
        {(local) => <LogicScene local={local} />}
      </Scene>
      <Scene start={610} end={810}>
        {(local) => <LaunchScene local={local} />}
      </Scene>
      <TransitionLayer />
      <ProgressMarks />
    </AbsoluteFill>
  );
};
