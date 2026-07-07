import type {EasingFunction} from 'remotion';
import type {ReactNode} from 'react';
import {AbsoluteFill, Easing, interpolate, useCurrentFrame} from 'remotion';

/**
 * ConduraConcerto — a minimalist film.
 *
 * One warm-paper canvas. One ink line (a conductor's baton). Scattered AI-tool
 * dots that snap into formation as the baton passes — order from chaos — then a
 * single orchestral downbeat. No noise, no neon, no HUD. Just the idea.
 */

const TOTAL = 600;
const LINE_Y = 540;
const LINE_X0 = 240;
const LINE_X1 = 1680;

const PAPER = '#f5efe3';
const INK = '#1b1a17';
const INK_2 = '#514b42';
const INK_3 = '#867d6f';
const POLLEN = '#c18a4a';
const SYNAPSE = '#526f3f';

const ease = Easing.bezier(0.16, 1, 0.3, 1);
const easeSlow = Easing.bezier(0.45, 0, 0.1, 1);

function r(f: number, from: number, to: number, e: EasingFunction = ease): number {
  return interpolate(f, [from, to], [0, 1], {easing: e, extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});
}
function out(f: number, from: number, to: number): number {
  return interpolate(f, [from, to], [1, 0], {easing: Easing.in(Easing.cubic), extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});
}

// Deterministic pseudo-random in [0,1)
function rng(seed: number): number {
  const x = Math.sin(seed * 127.1 + 311.7) * 43758.5453;
  return x - Math.floor(x);
}

const TOOLS = ['Claude', 'GPT', 'Gemini', 'Ollama', 'Codex', 'Antigravity', 'Cursor', 'OpenCode', 'Kilo'];

// Target positions: evenly spaced along the line.
const targets = TOOLS.map((_, i) => ({
  x: LINE_X0 + 80 + (i * (LINE_X1 - LINE_X0 - 160)) / (TOOLS.length - 1),
  y: LINE_Y,
}));

// Scattered start positions: deterministic, off the line, across the canvas.
const scattered = TOOLS.map((_, i) => ({
  x: 180 + rng(i + 1) * 1560,
  y: 220 + rng(i + 9) * 640,
}));

function Paper() {
  return (
    <AbsoluteFill className="cc-paper">
      <div className="cc-paper-wash" />
      <div className="cc-grain" />
    </AbsoluteFill>
  );
}

function BatonTip({progress}: {progress: number}) {
  const x = interpolate(progress, [0, 1], [LINE_X0, LINE_X1]);
  return (
    <div
      className="cc-baton-tip"
      style={{left: x, top: LINE_Y, transform: `translate(-50%, -50%) scale(${0.7 + progress * 0.3})`}}
    />
  );
}

/** A single tool dot interpolating from scattered -> aligned as the sweep passes it. */
function ToolDot({
  index,
  frame,
  sweepX,
  aligned,
  ripple,
}: {
  index: number;
  frame: number;
  sweepX: number;
  aligned: number;
  ripple: number;
}) {
  const label = TOOLS[index];
  const t = targets[index];
  const s = scattered[index];

  // Activation: how far past this dot's target the baton has travelled.
  const act = interpolate(sweepX, [t.x - 90, t.x + 50], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: ease,
  });

  // Position: scattered -> target.
  const x = interpolate(act, [0, 1], [s.x, t.x]);
  const y = interpolate(act, [0, 1], [s.y, t.y]);

  // Entrance fade (the scattered dots appearing).
  const enter = r(frame, 90 + index * 6, 130 + index * 6);
  // Settle hold then fade for the contraction beat.
  const exit = out(frame, 430, 470);
  const opacity = Math.min(enter, exit) * aligned;

  // Downbeat ripple: a scale pulse delayed by distance from center.
  const dist = Math.abs(t.x - 960);
  const rippleLocal = interpolate(ripple, [dist / 4, dist / 4 + 60], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const pulse = Math.sin(rippleLocal * Math.PI);
  const dotScale = 1 + pulse * 0.5;

  // Label appears once aligned.
  const labelOp = interpolate(act, [0.6, 1], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'});

  return (
    <div
      className="cc-tool"
      style={{left: x, top: y, opacity, transform: `translate(-50%, -50%)`}}
    >
      <div
        className="cc-dot"
        style={{
          transform: `scale(${dotScale})`,
          background: index % 3 === 0 ? SYNAPSE : INK,
        }}
      />
      <span className="cc-label" style={{opacity: labelOp}}>
        {label}
      </span>
    </div>
  );
}

function TypeWord({text, frame, at, size}: {text: string; frame: number; at: number; size: 'display' | 'line' | 'url'}) {
  const show = r(frame, at, at + 36, easeSlow);
  const y = interpolate(show, [0, 1], [16, 0]);
  return (
    <div className={`cc-word cc-word-${size}`} style={{opacity: show, transform: `translate3d(0, ${y}px, 0)`}}>
      {text}
    </div>
  );
}

function Scene({children}: {children: ReactNode}) {
  return <AbsoluteFill className="cc-scene">{children}</AbsoluteFill>;
}

export const ConduraConcerto = () => {
  const frame = useCurrentFrame();

  // Beat 1: faint hairline + baton tip at rest (left).
  const hairline = r(frame, 0, 40) * 0.25;

  // Beat 3: baton sweep across, drawing the solid ink line.
  const sweep = r(frame, 180, 330, easeSlow);
  const sweepX = interpolate(sweep, [0, 1], [LINE_X0, LINE_X1]);

  // The drawn line: solid ink from LINE_X0 to sweepX.
  const lineW = sweepX - LINE_X0;

  // aligned factor: dots only show once the scatter beat has begun.
  const aligned = r(frame, 90, 150);

  // Downbeat ripple triggers after the sweep completes.
  const ripple = interpolate(frame, [350, 360], [0, 1], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'}) * 600;

  // Contraction: line + dots fade; brand resolves.
  const brandIn = r(frame, 440, 510, easeSlow);
  const taglineIn = r(frame, 500, 545, ease);
  const urlIn = r(frame, 530, 575, ease);
  const finalFade = out(frame, 575, 600);

  return (
    <AbsoluteFill className="cc-root">
      <Paper />
      <Scene>
        {/* faint guide hairline (always there, very subtle) */}
        <div
          className="cc-hairline"
          style={{
            left: LINE_X0,
            top: LINE_Y,
            width: LINE_X1 - LINE_X0,
            opacity: Math.max(hairline, 0.06) * (1 - brandIn * 0.6),
          }}
        />
        {/* the drawn ink line */}
        <div
          className="cc-inkline"
          style={{left: LINE_X0, top: LINE_Y, width: lineW, opacity: 1 - brandIn}}
        />
        {/* baton tip travels only during the sweep */}
        {sweep > 0 && sweep < 1 && <BatonTip progress={sweep} />}

        {/* tool dots */}
        {TOOLS.map((_, i) => (
          <ToolDot key={i} index={i} frame={frame} sweepX={sweepX} aligned={aligned} ripple={ripple} />
        ))}

        {/* headline once aligned */}
        <div className="cc-headline" style={{opacity: r(frame, 350, 390) * (1 - brandIn)}}>
          Your tools, <em>in concert.</em>
        </div>

        {/* resolve to brand */}
        <div className="cc-brand" style={{opacity: brandIn * finalFade}}>
          <TypeWord text="Condura" frame={frame} at={450} size="display" />
        </div>
        <div className="cc-tagline" style={{opacity: taglineIn * finalFade}}>
          The conductor for every AI on your computer.
        </div>
        <div className="cc-url" style={{opacity: urlIn * finalFade}}>
          <TypeWord text="condura.app" frame={frame} at={535} size="url" />
        </div>
      </Scene>
    </AbsoluteFill>
  );
};
