import type {CSSProperties, ReactNode} from 'react';
import {
  AbsoluteFill,
  Easing,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from 'remotion';

const DURATION = 1800;

const scenes = [
  {label: 'Problem', start: 0, duration: 180},
  {label: 'Hotkey', start: 180, duration: 210},
  {label: 'Route', start: 390, duration: 270},
  {label: 'Demo', start: 660, duration: 450},
  {label: 'Control', start: 1110, duration: 285},
  {label: 'Receipt', start: 1395, duration: 225},
  {label: 'Condura', start: 1620, duration: 180},
] as const;

const ease = Easing.bezier(0.16, 1, 0.3, 1);

function clamp01(value: number): number {
  return Math.min(1, Math.max(0, value));
}

function ramp(frame: number, from: number, to: number): number {
  if (from === to) return frame >= to ? 1 : 0;
  return interpolate(frame, [from, to], [0, 1], {
    easing: ease,
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
}

function typeText(text: string, frame: number, from: number, to: number): string {
  const amount = ramp(frame, from, to);
  return text.slice(0, Math.floor(amount * text.length));
}

function countUp(frame: number, from: number, to: number, value: number): string {
  const amount = ramp(frame, from, to);
  return Math.round(amount * value).toString();
}

function SceneLayer({
  start,
  duration,
  children,
  className = '',
}: {
  start: number;
  duration: number;
  className?: string;
  children: (localFrame: number) => ReactNode;
}) {
  const frame = useCurrentFrame();
  const fade = 18;
  const visible = frame >= start - fade && frame <= start + duration + fade;
  if (!visible) return null;

  const local = frame - start;
  const opacityIn = ramp(frame, start, start + fade);
  const opacityOut = interpolate(frame, [start + duration - fade, start + duration], [1, 0], {
    easing: Easing.in(Easing.cubic),
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const opacity = Math.min(opacityIn, opacityOut);
  const scale = interpolate(opacity, [0, 1], [1.015, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <AbsoluteFill
      className={`scene-layer ${className}`}
      style={{opacity, transform: `scale(${scale})`}}
    >
      {children(local)}
    </AbsoluteFill>
  );
}

function BrandMark({size = 64}: {size?: number}) {
  return (
    <div className="brand-mark" style={{width: size, height: size}}>
      <span>C</span>
      <i />
    </div>
  );
}

function Keycap({children, wide = false}: {children: ReactNode; wide?: boolean}) {
  return <span className={`keycap ${wide ? 'keycap-wide' : ''}`}>{children}</span>;
}

function MacDots() {
  return (
    <div className="mac-dots">
      <i />
      <i />
      <i />
    </div>
  );
}

function WindowShell({
  title,
  meta,
  children,
  className = '',
  style,
}: {
  title: string;
  meta?: string;
  children: ReactNode;
  className?: string;
  style?: CSSProperties;
}) {
  return (
    <div className={`window-shell ${className}`} style={style}>
      <div className="window-bar">
        <MacDots />
        <strong>{title}</strong>
        {meta ? <span>{meta}</span> : null}
      </div>
      {children}
    </div>
  );
}

function PaperField({frame}: {frame: number}) {
  const drift = Math.sin(frame / 70) * 16;
  return (
    <AbsoluteFill className="paper-field">
      <div className="paper-spot paper-spot-a" style={{transform: `translate3d(${drift}px, 0, 0)`}} />
      <div className="paper-spot paper-spot-b" style={{transform: `translate3d(${-drift * 0.7}px, ${drift * 0.35}px, 0)`}} />
      <div className="paper-grain" />
    </AbsoluteFill>
  );
}

function Headline({
  eyebrow,
  title,
  body,
  align = 'left',
}: {
  eyebrow: string;
  title: ReactNode;
  body?: ReactNode;
  align?: 'left' | 'center';
}) {
  return (
    <div className={`headline-block ${align === 'center' ? 'headline-center' : ''}`}>
      <div className="eyebrow">{eyebrow}</div>
      <h1>{title}</h1>
      {body ? <p>{body}</p> : null}
    </div>
  );
}

function DesktopFragments({frame}: {frame: number}) {
  const cards = [
    {title: 'Browser', meta: 'docs', x: 108, y: 206, w: 430, h: 240, rot: -6, lag: 0},
    {title: 'Chat', meta: 'new thread', x: 1280, y: 146, w: 470, h: 260, rot: 4, lag: 12},
    {title: 'Notes', meta: 'call summary', x: 192, y: 680, w: 530, h: 230, rot: 5, lag: 24},
    {title: 'Terminal', meta: 'agent cli', x: 1220, y: 664, w: 470, h: 238, rot: -4, lag: 36},
    {title: 'Mail', meta: 'draft', x: 746, y: 732, w: 410, h: 154, rot: 2, lag: 48},
  ];

  return (
    <>
      {cards.map((card) => {
        const entrance = ramp(frame, 8 + card.lag, 42 + card.lag);
        const jitter = Math.sin((frame + card.lag) / 20) * 4;
        const lift = interpolate(entrance, [0, 1], [80, 0]);
        return (
          <WindowShell
            key={card.title}
            title={card.title}
            meta={card.meta}
            className="fragment-window"
            style={{
              left: card.x,
              top: card.y,
              width: card.w,
              height: card.h,
              opacity: entrance * 0.88,
              transform: `translate3d(${jitter}px, ${lift}px, 0) rotate(${card.rot}deg)`,
            }}
          >
            <div className="fake-lines">
              <i className="wide" />
              <i className="mid" />
              <i className="short" />
              <div className="chip-row">
                <span>copy</span>
                <span>paste</span>
                <span className="warn">lost context</span>
              </div>
            </div>
          </WindowShell>
        );
      })}
    </>
  );
}

function HookScene({frame}: {frame: number}) {
  const titleY = interpolate(ramp(frame, 20, 58), [0, 1], [40, 0]);
  const titleOpacity = ramp(frame, 20, 58);
  const resolve = ramp(frame, 116, 164);

  return (
    <>
      <PaperField frame={frame} />
      <DesktopFragments frame={frame} />
      <div className="hook-copy" style={{opacity: titleOpacity, transform: `translateY(${titleY}px)`}}>
        <Headline
          eyebrow="Condura demo"
          title={
            <>
              One task.
              <br />
              Six places to open.
            </>
          }
          body="Condura turns scattered AI work into one hotkey, one plan, and one visible boundary."
        />
      </div>
      <div className="problem-strip" style={{opacity: ramp(frame, 72, 112)}}>
        {['search', 'paste', 'ask again', 'switch', 'verify', 'send'].map((item, index) => (
          <span key={item} style={{opacity: ramp(frame, 74 + index * 6, 104 + index * 6)}}>
            {item}
          </span>
        ))}
      </div>
      <div className="hook-resolve" style={{opacity: resolve, transform: `translate(-50%, ${interpolate(resolve, [0, 1], [20, 0])}px)`}}>
        <BrandMark size={54} />
        <span>Call the conductor instead.</span>
      </div>
    </>
  );
}

function DesktopBase({frame, dense = false}: {frame: number; dense?: boolean}) {
  const float = Math.sin(frame / 55) * 7;
  return (
    <div className={`desktop-base ${dense ? 'desktop-dense' : ''}`}>
      <div className="desktop-menu">
        <BrandMark size={28} />
        <span>Condura</span>
        <i />
        <b>desktop</b>
        <em>{dense ? 'screen context available' : 'local session'}</em>
      </div>
      <div className="desktop-dock">
        {['C', 'N', 'B', 'T', 'M'].map((d, index) => (
          <span key={d} style={{transform: `translateY(${Math.sin(frame / 16 + index) * 3}px)`}}>
            {d}
          </span>
        ))}
      </div>
      <WindowShell
        title="Roadmap note"
        meta="active"
        className="desktop-note"
        style={{transform: `translate3d(0, ${float}px, 0)`}}
      >
        <div className="note-paper">
          <b>Investor follow-up</b>
          <i />
          <i className="mid" />
          <i className="short" />
          <div className="note-tag">needs summary</div>
        </div>
      </WindowShell>
      <WindowShell title="Browser" meta="call transcript" className="desktop-browser">
        <div className="browser-layout">
          <div />
          <div>
            <i />
            <i className="wide" />
            <i className="mid" />
          </div>
        </div>
      </WindowShell>
    </div>
  );
}

function HotkeyScene({frame}: {frame: number}) {
  const {fps} = useVideoConfig();
  const summon = spring({
    frame: frame - 36,
    fps,
    config: {damping: 18, stiffness: 120, mass: 0.7},
  });
  const typed = typeText(
    'Watch this screen, summarize the spec, draft the follow-up.',
    frame,
    76,
    158,
  );
  const ripple = ramp(frame, 30, 72);
  const promptScale = interpolate(summon, [0, 1], [0.92, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <>
      <PaperField frame={frame + 180} />
      <DesktopBase frame={frame} />
      <div className="hotkey-cluster" style={{opacity: ramp(frame, 10, 34)}}>
        <Keycap>CMD</Keycap>
        <Keycap>SHIFT</Keycap>
        <Keycap wide>SPACE</Keycap>
      </div>
      <div
        className="summon-ripple"
        style={{
          opacity: interpolate(ripple, [0, 0.35, 1], [0, 0.55, 0], {
            extrapolateLeft: 'clamp',
            extrapolateRight: 'clamp',
          }),
          transform: `translate(-50%, -50%) scale(${interpolate(ripple, [0, 1], [0.25, 2.2])})`,
        }}
      />
      <div
        className="quick-prompt"
        style={{
          opacity: summon,
          transform: `translate(-50%, -50%) scale(${promptScale})`,
        }}
      >
        <div className="prompt-top">
          <BrandMark size={34} />
          <strong>Condura</strong>
          <span>local context first</span>
        </div>
        <div className="prompt-input">
          {typed}
          <i style={{opacity: frame % 24 < 12 ? 1 : 0.18}} />
        </div>
        <div className="prompt-footer" style={{opacity: ramp(frame, 164, 190)}}>
          <span>screen</span>
          <span>voice ready</span>
          <span>approval required for action</span>
        </div>
      </div>
      <div className="hotkey-caption" style={{opacity: ramp(frame, 142, 184)}}>
        The user stays where they are. The agent comes to the work.
      </div>
    </>
  );
}

function Node({
  label,
  detail,
  x,
  y,
  active,
  delay,
  frame,
}: {
  label: string;
  detail: string;
  x: number;
  y: number;
  active?: boolean;
  delay: number;
  frame: number;
}) {
  const show = ramp(frame, delay, delay + 28);
  return (
    <div
      className={`route-node ${active ? 'route-node-active' : ''}`}
      style={{
        left: x,
        top: y,
        opacity: show,
        transform: `translate(-50%, -50%) scale(${interpolate(show, [0, 1], [0.72, 1])})`,
      }}
    >
      <strong>{label}</strong>
      <span>{detail}</span>
    </div>
  );
}

function ConductorScene({frame}: {frame: number}) {
  const pulse = clamp01(((frame - 80) % 76) / 76);
  const nodes = [
    {label: 'Ollama', detail: 'local model', x: 690, y: 236, delay: 34},
    {label: 'Files', detail: 'private context', x: 400, y: 550, delay: 44},
    {label: 'Keys', detail: 'configured APIs', x: 650, y: 772, delay: 54},
    {label: 'Agent CLI', detail: 'code tasks', x: 1260, y: 248, delay: 64},
    {label: 'Browser', detail: 'web work', x: 1454, y: 548, delay: 74},
    {label: 'Audit', detail: 'receipt', x: 1212, y: 796, delay: 84},
  ];

  return (
    <>
      <PaperField frame={frame + 390} />
      <div className="routing-copy">
        <Headline
          eyebrow="Routing"
          title={
            <>
              The request becomes
              <br />
              a plan.
            </>
          }
          body="Condura chooses local context first, then uses the providers and tools the user has configured."
        />
      </div>
      <div className="routing-board">
        <svg className="route-lines" viewBox="0 0 1920 1080" preserveAspectRatio="none">
          {nodes.map((node) => (
            <line
              key={node.label}
              x1="960"
              y1="540"
              x2={node.x}
              y2={node.y}
              stroke="rgba(86, 110, 71, .28)"
              strokeWidth="2"
            />
          ))}
        </svg>
        {nodes.map((node, index) => (
          <Node key={node.label} {...node} active={index === Math.floor((frame / 44) % nodes.length)} frame={frame} />
        ))}
        <div className="conductor-core" style={{opacity: ramp(frame, 20, 58)}}>
          <BrandMark size={72} />
          <strong>Conductor</strong>
          <span>{countUp(frame, 92, 152, 4)} step plan</span>
        </div>
        <div
          className="route-pulse"
          style={{
            opacity: frame > 82 ? 1 : 0,
            transform: `translate(${interpolate(pulse, [0, 1], [960, 1260])}px, ${interpolate(pulse, [0, 1], [540, 248])}px)`,
          }}
        />
      </div>
      <div className="plan-stack" style={{opacity: ramp(frame, 112, 152)}}>
        {['Read the active screen', 'Use local memory and files', 'Draft two clean options', 'Ask before sending'].map((step, index) => {
          const done = frame > 146 + index * 24;
          return (
            <div className={`plan-row ${done ? 'done' : ''}`} key={step}>
              <i>{done ? 'ok' : `0${index + 1}`}</i>
              <span>{step}</span>
              <b>{done ? 'done' : 'queued'}</b>
            </div>
          );
        })}
      </div>
    </>
  );
}

function ProductWindow({frame}: {frame: number}) {
  const command = typeText(
    'Prepare me for the 3 PM investor follow-up. Use what is on screen.',
    frame,
    56,
    122,
  );
  const approvePhase = ramp(frame, 282, 330);
  const approved = frame > 356;
  const outputText = typeText(
    'Draft ready: concise recap, decision list, and a follow-up note that waits for your approval before sending.',
    frame,
    312,
    410,
  );
  const messages = [
    ['Screen read', 'I found the roadmap note and transcript in the active workspace.', 132],
    ['Local memory', 'Matched it to the last investor thread and your preferred tone.', 176],
    ['Drafting', 'Built two options: concise and detailed.', 220],
  ];

  return (
    <WindowShell title="Condura" meta="screen context on" className="product-window">
      <div className="product-body">
        <aside className="product-nav">
          {['Chat', 'Hub', 'Skills', 'Audit', 'Settings'].map((item, index) => (
            <div className={`nav-item ${index === 0 ? 'active' : ''}`} key={item}>
              <i />
              <span>{item}</span>
            </div>
          ))}
          <div className="nav-spacer" />
          <div className="kill-switch">Hard stop ready</div>
        </aside>
        <main className="chat-pane">
          <div className="voice-pill" style={{opacity: ramp(frame, 14, 48)}}>
            <div className="wave">
              {Array.from({length: 18}).map((_, index) => (
                <i key={index} style={{height: 12 + Math.abs(Math.sin(frame / 8 + index)) * 28}} />
              ))}
            </div>
            <span>Voice captured</span>
          </div>
          <div className="chat-command" style={{opacity: ramp(frame, 42, 76)}}>
            <b>You</b>
            <span>
              {command}
              <i style={{opacity: frame % 24 < 12 ? 1 : 0}} />
            </span>
          </div>
          <div className="agent-feed">
            {messages.map(([head, body, at], index) => (
              <div
                className="agent-card"
                key={head}
                style={{
                  opacity: ramp(frame, Number(at), Number(at) + 22),
                  transform: `translateY(${interpolate(ramp(frame, Number(at), Number(at) + 22), [0, 1], [18, 0])}px)`,
                }}
              >
                <div className="agent-avatar">{index + 1}</div>
                <div>
                  <strong>{head}</strong>
                  <span>{body}</span>
                </div>
              </div>
            ))}
          </div>
          <div className="result-card" style={{opacity: ramp(frame, 340, 392)}}>
            <strong>Result</strong>
            <p>{outputText}</p>
            <div className="result-actions">
              <span>copy draft</span>
              <span>open receipt</span>
            </div>
          </div>
        </main>
        <aside className="context-rail">
          {[
            ['Screen', 'Roadmap note active', 96],
            ['Files', '2 sources found', 150],
            ['Voice', 'transcript attached', 198],
            ['Audit', approved ? 'receipt saved' : 'pending action', 246],
          ].map(([head, body, at]) => (
            <div className="context-card" key={head} style={{opacity: ramp(frame, Number(at), Number(at) + 20)}}>
              <b>{head}</b>
              <span>{body}</span>
            </div>
          ))}
        </aside>
      </div>
      <div
        className="approval-modal"
        style={{
          opacity: approvePhase * (approved ? interpolate(frame, [356, 382], [1, 0], {extrapolateLeft: 'clamp', extrapolateRight: 'clamp'}) : 1),
          transform: `translate(-50%, -50%) scale(${interpolate(approvePhase, [0, 1], [0.94, 1])})`,
        }}
      >
        <div className="approval-badge">Approval required</div>
        <h3>Open Calendar and draft the hold?</h3>
        <p>Condura can prepare this action, but it will not touch the app until you approve.</p>
        <div className="approval-buttons">
          <span>Cancel</span>
          <strong>Approve</strong>
        </div>
      </div>
    </WindowShell>
  );
}

function LiveDemoScene({frame}: {frame: number}) {
  return (
    <>
      <PaperField frame={frame + 660} />
      <DesktopBase frame={frame} dense />
      <div className="demo-title" style={{opacity: ramp(frame, 0, 34)}}>
        <Headline
          eyebrow="Live shape"
          title="Overlay, voice, task done."
          body="The demo stays paced, but the user can still understand what happened."
          align="center"
        />
      </div>
      <div
        className="product-wrap"
        style={{
          opacity: ramp(frame, 34, 82),
          transform: `translate(-50%, ${interpolate(ramp(frame, 34, 82), [0, 1], [28, 0])}px)`,
        }}
      >
        <ProductWindow frame={frame} />
      </div>
    </>
  );
}

function SafetyScene({frame}: {frame: number}) {
  const matrix = [
    ['Coding', 'allow', 'green'],
    ['Files', 'ask', 'amber'],
    ['Browser', 'ask', 'amber'],
    ['Email', 'ask', 'amber'],
    ['Money', 'block', 'red'],
    ['System', 'ask', 'amber'],
  ];
  return (
    <>
      <PaperField frame={frame + 1110} />
      <div className="safety-copy">
        <Headline
          eyebrow="Control"
          title={
            <>
              The boundary
              <br />
              stays visible.
            </>
          }
          body="Condura treats safety as product surface, not a hidden policy."
        />
      </div>
      <div className="safety-board" style={{opacity: ramp(frame, 24, 62)}}>
        <div className="matrix">
          <div className="matrix-head">Autonomy matrix</div>
          {matrix.map(([name, state, tone], index) => (
            <div className="matrix-row" key={name} style={{opacity: ramp(frame, 52 + index * 10, 76 + index * 10)}}>
              <span>{name}</span>
              <b className={`tone-${tone}`}>{state}</b>
            </div>
          ))}
        </div>
        <div className="audit-thread">
          <div className="thread-title">Audit thread</div>
          {['heard voice command', 'read active screen', 'prepared draft', 'asked for calendar approval', 'saved receipt'].map((item, index) => (
            <div className="thread-row" key={item} style={{opacity: ramp(frame, 108 + index * 16, 130 + index * 16)}}>
              <i />
              <span>{item}</span>
            </div>
          ))}
        </div>
        <div className="stop-card" style={{opacity: ramp(frame, 178, 218)}}>
          <span>Hard stop</span>
          <strong>Cmd + Shift + Esc</strong>
          <p>Stop execution immediately. Network guard and kill switch stay reachable.</p>
        </div>
      </div>
      <div className="safety-note" style={{opacity: ramp(frame, 214, 254)}}>
        Fast does not mean autonomous by default.
      </div>
    </>
  );
}

function ReceiptScene({frame}: {frame: number}) {
  const cards = [
    ['Draft ready', 'Follow-up email prepared, not sent.'],
    ['Sources attached', 'Roadmap note and call transcript summarized.'],
    ['Approval logged', 'Calendar action required explicit consent.'],
    ['Receipt saved', 'Every step stays reviewable in Audit.'],
  ];
  return (
    <>
      <PaperField frame={frame + 1395} />
      <div className="receipt-title" style={{opacity: ramp(frame, 6, 38)}}>
        <Headline
          eyebrow="Result"
          title="The work returns as a receipt."
          body="A demo should end with proof, not just motion."
          align="center"
        />
      </div>
      <div className="receipt-grid">
        {cards.map(([title, body], index) => {
          const show = ramp(frame, 52 + index * 18, 84 + index * 18);
          return (
            <div
              className="receipt-card"
              key={title}
              style={{
                opacity: show,
                transform: `translateY(${interpolate(show, [0, 1], [34, 0])}px)`,
              }}
            >
              <i>{`0${index + 1}`}</i>
              <strong>{title}</strong>
              <span>{body}</span>
            </div>
          );
        })}
      </div>
      <div className="before-after" style={{opacity: ramp(frame, 148, 194)}}>
        <div>
          <b>Before</b>
          <span>8 windows, repeated context, manual verification</span>
        </div>
        <strong>{'->'}</strong>
        <div>
          <b>After</b>
          <span>one command, visible plan, reviewed output</span>
        </div>
      </div>
    </>
  );
}

function FinalScene({frame}: {frame: number}) {
  const {fps} = useVideoConfig();
  const mark = spring({
    frame: frame - 18,
    fps,
    config: {damping: 15, stiffness: 100, mass: 0.8},
  });
  return (
    <>
      <PaperField frame={frame + 1620} />
      <div className="final-lockup">
        <div style={{transform: `scale(${interpolate(mark, [0, 1], [0.82, 1])})`, opacity: mark}}>
          <BrandMark size={96} />
        </div>
        <h2 style={{opacity: ramp(frame, 42, 76)}}>Condura</h2>
        <p style={{opacity: ramp(frame, 70, 106)}}>
          One hotkey for your models, agents, files, and desktop.
        </p>
        <div className="final-rule" style={{transform: `scaleX(${ramp(frame, 90, 124)})`}} />
        <div className="final-cta" style={{opacity: ramp(frame, 106, 136)}}>
          condura.app/download
        </div>
      </div>
    </>
  );
}

function BeatRail() {
  const frame = useCurrentFrame();
  const overall = frame / DURATION;
  return (
    <div className="beat-rail">
      <div className="beat-line">
        <i style={{transform: `scaleX(${overall})`}} />
      </div>
      <div className="beat-labels">
        {scenes.map((scene) => {
          const active = frame >= scene.start && frame < scene.start + scene.duration;
          return (
            <span key={scene.label} className={active ? 'active' : ''}>
              {scene.label}
            </span>
          );
        })}
      </div>
    </div>
  );
}

export const ConduraDemo = () => {
  const frame = useCurrentFrame();

  return (
    <AbsoluteFill className="video-root">
      <SceneLayer start={0} duration={180}>
        {(local) => <HookScene frame={local} />}
      </SceneLayer>
      <SceneLayer start={180} duration={210}>
        {(local) => <HotkeyScene frame={local} />}
      </SceneLayer>
      <SceneLayer start={390} duration={270}>
        {(local) => <ConductorScene frame={local} />}
      </SceneLayer>
      <SceneLayer start={660} duration={450}>
        {(local) => <LiveDemoScene frame={local} />}
      </SceneLayer>
      <SceneLayer start={1110} duration={285}>
        {(local) => <SafetyScene frame={local} />}
      </SceneLayer>
      <SceneLayer start={1395} duration={225}>
        {(local) => <ReceiptScene frame={local} />}
      </SceneLayer>
      <SceneLayer start={1620} duration={180}>
        {(local) => <FinalScene frame={local} />}
      </SceneLayer>
      <div className="film-hud">
        <span>Condura demo film</span>
        <b>{Math.floor(frame / 30).toString().padStart(2, '0')}s</b>
      </div>
      <BeatRail />
    </AbsoluteFill>
  );
};
