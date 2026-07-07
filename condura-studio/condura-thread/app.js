/* ─────────────────────────────────────────────────────────────────
   CONDURA · THREAD · orchestration
   master timeline · web audio · helix · constellation · particles
   ───────────────────────────────────────────────────────────────── */

/* ─── Beats · the 11 visible scenes ─────────────────────────────── */

const BEATS = [
  { id: 'strike',        start:  0.0, end:  2.0 },
  { id: 'bloom',         start:  2.0, end:  5.5 },
  { id: 'weave',         start:  5.5, end:  9.5 },
  { id: 'multiply',      start:  9.5, end: 14.0 },
  { id: 'constellate',   start: 14.0, end: 20.0 },
  { id: 'type',          start: 20.0, end: 26.5 },
  { id: 'receipt',       start: 26.5, end: 32.0 },
  { id: 'surge',         start: 32.0, end: 37.0 },
  { id: 'reveal',        start: 37.0, end: 42.0 },
  { id: 'mark',          start: 42.0, end: 48.0 },
  { id: 'cta',           start: 48.0, end: 60.0 },
];
const TOTAL = 60.0;

/* Thread-count checkpoints for the counter at the top-right. */
const THREAD_CHECKPOINTS = [
  { at:  0.0, n: 1  },
  { at:  5.5, n: 2  },
  { at:  6.5, n: 4  },
  { at: 10.5, n: 8  },
  { at: 14.0, n: 12 },
];

/* Receipt-arm checkpoints for the SVG arms in beat 7. */
const ARM_TIMES = [
  27.4, 28.0, 28.6, 29.2, 29.8, 30.4,
  31.0, 31.4, 31.8, 32.0, 32.2, 32.4,
];

/* ─── State ────────────────────────────────────────────────────── */

const state = {
  playing: false,
  muted: localStorage.getItem('condura:muted') === '1',
  startTime: 0,
  pausedAt: 0,
  elapsed: 0,
  currentBeat: -1,
  audioCtx: null,
  cursor: { x: 0, y: 0 },
  helixAngle: 0,
  helixStarted: false,
  constellationStarted: false,
  receiptArmsDone: new Set(),
  threadNum: 1,
  particleStart: 0,
};

/* ─── Web Audio · synthesized leitmotif ────────────────────────── */

const SOUND = {
  ensure() {
    if (!state.audioCtx) state.audioCtx = new (window.AudioContext || window.webkitAudioContext)();
    if (state.audioCtx.state === 'suspended') state.audioCtx.resume();
    return state.audioCtx;
  },
  tone(freq, duration = 0.4, type = 'sine', volume = 0.06, when = 0, attack = 0.012) {
    if (state.muted) return;
    const ctx = SOUND.ensure();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = type;
    osc.frequency.value = freq;
    const t = ctx.currentTime + when;
    gain.gain.setValueAtTime(0, t);
    gain.gain.linearRampToValueAtTime(volume, t + attack);
    gain.gain.exponentialRampToValueAtTime(0.0008, t + duration);
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start(t);
    osc.stop(t + duration + 0.02);
  },
  /* sharp percussive kick — used on STRIKE */
  kick(volume = 0.4, when = 0) {
    if (state.muted) return;
    const ctx = SOUND.ensure();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = 'sine';
    const t = ctx.currentTime + when;
    osc.frequency.setValueAtTime(140, t);
    osc.frequency.exponentialRampToValueAtTime(40, t + 0.15);
    gain.gain.setValueAtTime(volume, t);
    gain.gain.exponentialRampToValueAtTime(0.001, t + 0.4);
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start(t);
    osc.stop(t + 0.45);
  },
  /* sustained harmonic pad — plays for the duration of the beat */
  pad(freq, duration = 2.5, volume = 0.025, when = 0) {
    if (state.muted) return;
    const ctx = SOUND.ensure();
    const t = ctx.currentTime + when;
    [1, 2, 3].forEach((h, i) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = i === 0 ? 'sine' : 'triangle';
      osc.frequency.value = freq * h;
      gain.gain.setValueAtTime(0, t);
      gain.gain.linearRampToValueAtTime(volume * (1 - i * 0.3), t + 0.2);
      gain.gain.linearRampToValueAtTime(volume * 0.4, t + duration * 0.7);
      gain.gain.exponentialRampToValueAtTime(0.001, t + duration);
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.start(t);
      osc.stop(t + duration + 0.05);
    });
  },
  click(freq = 1800, volume = 0.025) {
    if (state.muted) return;
    SOUND.tone(freq, 0.05, 'square', volume);
  },
  /* per-beat signature */
  beatCue(beatId) {
    if (state.muted) return;
    switch (beatId) {
      case 'strike':
        SOUND.kick(0.5, 0);
        SOUND.kick(0.4, 0.4);
        SOUND.pad(65, 2.0, 0.05);             /* C2 */
        break;
      case 'bloom':
        SOUND.pad(98, 3.0, 0.04);             /* G2 added */
        SOUND.pad(65, 3.0, 0.04);
        break;
      case 'weave':
        SOUND.pad(82, 4.0, 0.04);             /* E2 added */
        SOUND.pad(98, 4.0, 0.035);
        SOUND.pad(131, 4.0, 0.035);            /* C3 */
        SOUND.tone(880, 0.2, 'triangle', 0.04, 0.3);
        SOUND.tone(1100, 0.2, 'triangle', 0.04, 0.9);
        SOUND.tone(1320, 0.2, 'triangle', 0.04, 1.5);
        break;
      case 'multiply':
        SOUND.pad(165, 4.5, 0.04);             /* E3 */
        SOUND.pad(196, 4.5, 0.035);            /* G3 */
        SOUND.pad(262, 4.5, 0.035);            /* C4 */
        /* rhythmic tick */
        for (let i = 0; i < 8; i++) SOUND.tone(1500 + i * 50, 0.05, 'square', 0.02, i * 0.18);
        break;
      case 'constellate':
        /* 12-tone cluster, dense */
        [262, 294, 330, 349, 392, 440, 494, 523, 587, 659, 698, 784].forEach((f, i) =>
          SOUND.tone(f, 0.4, 'triangle', 0.025, i * 0.08));
        SOUND.pad(130, 5.0, 0.03);
        break;
      case 'type':
        /* percussive click per major word */
        [0, 0.7, 1.4, 3.0, 3.6].forEach(t => SOUND.kick(0.2, t));
        SOUND.pad(196, 6.0, 0.03);
        break;
      case 'receipt':
        /* low sustained + per-arm clicks */
        SOUND.pad(98, 5.5, 0.04);
        ARM_TIMES.slice(0, 8).forEach((t, i) => SOUND.tone(660 + i * 40, 0.08, 'triangle', 0.04, t - 26.5));
        break;
      case 'surge':
        /* tempo climax — descending arpeggio that resolves up */
        SOUND.tone(523, 0.3, 'sine', 0.05, 0);
        SOUND.tone(659, 0.3, 'sine', 0.05, 0.3);
        SOUND.tone(784, 0.3, 'sine', 0.05, 0.6);
        SOUND.tone(1046, 0.6, 'sine', 0.07, 0.9);
        SOUND.tone(1318, 1.5, 'sine', 0.05, 1.5);
        SOUND.pad(196, 5.0, 0.04);
        break;
      case 'reveal':
        /* big chord hit + sustained */
        [262, 330, 392, 523].forEach((f, i) => SOUND.tone(f, 4.0, 'sine', 0.06, i * 0.01));
        SOUND.kick(0.3, 0);
        break;
      case 'mark':
        /* held warm chord */
        SOUND.pad(262, 5.0, 0.04);
        SOUND.pad(392, 5.0, 0.03);
        break;
      case 'cta':
        /* chord resolves and fades — final gesture */
        SOUND.pad(330, 6.0, 0.04);
        SOUND.pad(523, 6.0, 0.03);
        SOUND.tone(1046, 4.0, 'sine', 0.04, 1.5);
        SOUND.tone(1318, 4.0, 'sine', 0.03, 2.0);
        SOUND.tone(1568, 4.0, 'sine', 0.025, 2.5);
        break;
    }
  },
};

/* ─── Helix rendering · 8 thread ellipses rotating ──────────────── */

function renderHelix() {
  const root = document.getElementById('helix');
  const N = 8;
  for (let i = 0; i < N; i++) {
    const t = document.createElement('div');
    t.className = 'helix-thread';
    const angle = (i / N) * 360;
    /* Z-depth via scaleY + opacity (suggests 3D helix without WebGL) */
    const depthN = 0.5 + 0.5 * Math.sin((i / N) * Math.PI * 2);
    t.style.setProperty('--helix-angle', `${angle}deg`);
    t.style.setProperty('--helix-scale', (0.6 + depthN * 0.4).toFixed(3));
    t.style.setProperty('--helix-opacity', (0.4 + depthN * 0.6).toFixed(3));
    t.style.animationDelay = `${i * 0.06}s`;
    root.appendChild(t);
  }
}

function spinHelix(dt) {
  if (!state.helixStarted) return;
  state.helixAngle += dt * 30;  /* 30 deg/sec */
  const root = document.getElementById('helix');
  root.style.transform = `perspective(1200px) rotateX(0deg) rotateZ(${state.helixAngle}deg)`;
}

/* ─── Constellation · 12 nodes positioned on a circle ─────────── */

const CONSTELLATION_LABELS = [
  'HOTKEY', 'OVERLAY', 'PLAN', 'AGENTS',
  'VOICE', 'CODE', 'MEMORY', 'AUDIT',
  'SAFETY', 'CHANNELS', 'SYNC', 'HUB',
];

function renderConstellation() {
  const root = document.getElementById('constellation');
  const w = root.offsetWidth, h = root.offsetHeight;
  const cx = w / 2, cy = h / 2;
  const radius = Math.min(w, h) * 0.42;
  const N = 12;

  /* SVG for connecting lines */
  const svgNS = 'http://www.w3.org/2000/svg';
  const svg = document.createElementNS(svgNS, 'svg');
  svg.setAttribute('viewBox', `0 0 ${w} ${h}`);
  svg.style.position = 'absolute';
  svg.style.inset = '0';
  svg.style.width = '100%';
  svg.style.height = '100%';
  svg.classList.add('constellation-svg');

  /* Connecting pattern — connect each node to 2-3 neighbors (skip the immediate next) */
  const connections = [
    [0, 2], [0, 3], [1, 4], [1, 5],
    [2, 6], [3, 7], [4, 8], [5, 9],
    [6, 10], [7, 11], [8, 0], [9, 1],
    [10, 2], [11, 3],
  ];
  connections.forEach(([a, b], i) => {
    const ax = cx + radius * Math.cos((a / N) * Math.PI * 2 - Math.PI / 2);
    const ay = cy + radius * Math.sin((a / N) * Math.PI * 2 - Math.PI / 2);
    const bx = cx + radius * Math.cos((b / N) * Math.PI * 2 - Math.PI / 2);
    const by = cy + radius * Math.sin((b / N) * Math.PI * 2 - Math.PI / 2);
    const line = document.createElementNS(svgNS, 'line');
    line.setAttribute('x1', ax); line.setAttribute('y1', ay);
    line.setAttribute('x2', bx); line.setAttribute('y2', by);
    line.setAttribute('stroke', 'var(--amber-bright)');
    line.setAttribute('stroke-width', '0.8');
    line.setAttribute('stroke-opacity', '0.3');
    line.setAttribute('stroke-dasharray', '4 4');
    line.style.opacity = '0';
    line.style.transition = `opacity 0.6s ease ${0.5 + i * 0.05}s`;
    svg.appendChild(line);
    requestAnimationFrame(() => line.style.opacity = '0.6');
  });
  root.appendChild(svg);

  /* Nodes */
  for (let i = 0; i < N; i++) {
    const angle = (i / N) * Math.PI * 2 - Math.PI / 2;
    const x = cx + radius * Math.cos(angle);
    const y = cy + radius * Math.sin(angle);
    const node = document.createElement('div');
    node.className = 'const-node';
    node.style.left = `${x}px`;
    node.style.top = `${y}px`;
    node.innerHTML = `
      <span class="const-node__core"></span>
      <span class="const-node__label">${CONSTELLATION_LABELS[i]}</span>
    `;
    node.style.animationDelay = `${i * 0.07}s`;
    root.appendChild(node);
  }

  state.constellationStarted = true;
}

function spinConstellation(dt) {
  if (!state.constellationStarted) return;
  const root = document.getElementById('constellation');
  state.helixAngle += dt * 5;  /* slow rotation */
  /* subtle rotation only */
  root.style.transform = `rotate(${state.helixAngle * 0.15}deg)`;
}

/* ─── Particles · canvas overlay ───────────────────────────────── */

let particles = [];
function initParticles() {
  const c = document.getElementById('particles');
  c.width = window.innerWidth;
  c.height = window.innerHeight;
  particles = [];
  for (let i = 0; i < 60; i++) {
    particles.push({
      x: Math.random() * c.width,
      y: c.height + Math.random() * 100,
      vx: (Math.random() - 0.5) * 0.4,
      vy: -0.3 - Math.random() * 0.6,
      r: 1 + Math.random() * 2.5,
      a: 0.2 + Math.random() * 0.5,
      life: Math.random(),
    });
  }
}

function drawParticles(dt) {
  const c = document.getElementById('particles');
  if (!c.getContext) return;
  const ctx = c.getContext('2d');
  ctx.clearRect(0, 0, c.width, c.height);

  /* density ramps with time — sparse early, dense by CTA */
  const density = Math.min(1, state.elapsed / 30);

  particles.forEach(p => {
    p.x += p.vx * dt * 60;
    p.y += p.vy * dt * 60;
    p.life += dt * 0.2;
    if (p.y < -20 || p.life > 1) {
      p.x = Math.random() * c.width;
      p.y = c.height + 20;
      p.life = 0;
    }
    ctx.beginPath();
    ctx.fillStyle = `rgba(255, 209, 102, ${p.a * density})`;
    ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
    ctx.fill();
    /* glow */
    ctx.beginPath();
    ctx.fillStyle = `rgba(255, 170, 58, ${p.a * 0.3 * density})`;
    ctx.arc(p.x, p.y, p.r * 4, 0, Math.PI * 2);
    ctx.fill();
  });
}

/* ─── Receipt arms · draw in sequence ───────────────────────────── */

function maybeArm() {
  ARM_TIMES.forEach((t, i) => {
    if (state.elapsed >= t && !state.receiptArmsDone.has(i)) {
      state.receiptArmsDone.add(i);
      const arm = document.querySelector(`.receipt-arm--${String(i + 1).padStart(2, '0')}`);
      if (arm) arm.classList.add('is-done');
      SOUND.click(660 + i * 40, 0.03);
    }
  });
}

/* ─── Thread counter ───────────────────────────────────────────── */

function updateThreadCounter() {
  let n = 1;
  THREAD_CHECKPOINTS.forEach(cp => { if (state.elapsed >= cp.at) n = cp.n; });
  if (n !== state.threadNum) {
    state.threadNum = n;
    document.getElementById('thread-num').textContent = String(n).padStart(2, '0');
  }
}

/* ─── HUD time + trust meter ───────────────────────────────────── */

function updateHud() {
  const t = state.elapsed;
  const mm = String(Math.floor(t / 60)).padStart(2, '0');
  const ss = String(Math.floor(t % 60)).padStart(2, '0');
  document.getElementById('time').textContent = `${mm}:${ss} / 01:00`;
  const trustPct = Math.max(0, Math.min(100, (state.elapsed / 42) * 100));
  document.getElementById('trust-fill').style.height = `${trustPct}%`;
  document.getElementById('trust-value').textContent = String(Math.round(trustPct)).padStart(3, '0');
}

/* ─── Beat enter ───────────────────────────────────────────────── */

function enterBeat(i) {
  const beat = BEATS[i];
  if (!beat) return;
  document.querySelectorAll('.beat').forEach(el => el.classList.remove('is-active', 'is-leaving'));
  const el = document.querySelector(`.beat--${beat.id}`);
  if (el) el.classList.add('is-active');
  state.currentBeat = i;
  SOUND.beatCue(beat.id);
  /* Per-beat side effects */
  if (beat.id === 'multiply') state.helixStarted = true;
  if (beat.id === 'constellate' && !state.constellationStarted) renderConstellation();
  if (beat.id === 'surge') {
    document.getElementById('particles').classList.add('is-on');
    state.particleStart = state.elapsed;
  }
}

/* ─── Master tick ──────────────────────────────────────────────── */

let lastFrame = 0;
function tick(now) {
  if (!state.playing) return;
  const dt = lastFrame ? (now - lastFrame) / 1000 : 0.016;
  lastFrame = now;
  state.elapsed = (now - state.startTime) / 1000;
  if (state.elapsed >= TOTAL) {
    state.elapsed = TOTAL;
    stop();
  }
  /* Find current beat */
  let nextBeat = -1;
  for (let i = 0; i < BEATS.length; i++) {
    if (state.elapsed >= BEATS[i].start && state.elapsed < BEATS[i].end) { nextBeat = i; break; }
  }
  if (nextBeat !== state.currentBeat) enterBeat(nextBeat);
  /* Helix spin */
  spinHelix(dt);
  /* Constellation rotation */
  spinConstellation(dt);
  /* Receipt arms */
  if (state.elapsed >= 27) maybeArm();
  /* Thread counter */
  updateThreadCounter();
  /* HUD */
  updateHud();
  /* Particles */
  if (state.elapsed >= 32) drawParticles(dt);
  requestAnimationFrame(tick);
}

function play() {
  if (state.playing) return;
  state.playing = true;
  state.startTime = performance.now() - state.pausedAt * 1000;
  lastFrame = 0;
  SOUND.ensure();
  document.getElementById('gate').classList.add('is-gone');
  localStorage.setItem('condura:played', '1');
  requestAnimationFrame(tick);
}

function stop() {
  state.playing = false;
  state.pausedAt = state.elapsed;
}

/* ─── Click flash · visual confirmation that a click registered ─── */

let flashTimer = null;
function flashClick() {
  const glow = document.getElementById('cursor-glow');
  glow.classList.add('is-active');
  clearTimeout(flashTimer);
  flashTimer = setTimeout(() => glow.classList.remove('is-active'), 220);
  const cur = document.getElementById('cursor');
  cur.classList.add('is-press');
  setTimeout(() => cur.classList.remove('is-press'), 120);
}

/* ─── Cursor ─────────────────────────────────────────────────── */

function bindCursor() {
  const cur = document.getElementById('cursor');
  const trail = document.getElementById('cursor-trail');
  const glow = document.getElementById('cursor-glow');
  let trailTimer = null;

  document.addEventListener('mousemove', (e) => {
    state.cursor.x = e.clientX;
    state.cursor.y = e.clientY;
    cur.style.transform = `translate3d(${e.clientX}px, ${e.clientY}px, 0)`;
    trail.style.transform = `translate3d(${e.clientX}px, ${e.clientY}px, 0)`;
    glow.style.transform = `translate3d(${e.clientX}px, ${e.clientY}px, 0)`;
    trail.classList.add('is-active');
    glow.classList.add('is-active');
    clearTimeout(trailTimer);
    trailTimer = setTimeout(() => {
      trail.classList.remove('is-active');
      glow.classList.remove('is-active');
    }, 100);
  });
  document.addEventListener('mousedown', () => cur.classList.add('is-press'));
  document.addEventListener('mouseup', () => cur.classList.remove('is-press'));
  document.addEventListener('mouseover', (e) => {
    if (e.target.closest('button, a')) cur.classList.add('is-hover');
    else cur.classList.remove('is-hover');
  });
}

/* ─── Controls ─────────────────────────────────────────────────── */

function bindControls() {
  document.getElementById('gate-play').addEventListener('click', () => play());
  document.getElementById('sound-toggle').addEventListener('click', () => toggleMute());

  document.addEventListener('keydown', (e) => {
    if (e.key === ' ') { e.preventDefault(); state.playing ? stop() : play(); }
    else if (e.key === 'm' || e.key === 'M') toggleMute();
    else if (e.key === 'f' || e.key === 'F') toggleFullscreen();
    else if (e.key === 'ArrowRight') { state.pausedAt = Math.min(TOTAL, state.pausedAt + 2); if (!state.playing) play(); }
    else if (e.key === 'ArrowLeft')  { state.pausedAt = Math.max(0, state.pausedAt - 2); if (!state.playing) play(); }
  });

  /* Skip-to-next on pointerdown. Faster than `click` (fires before release,
   * survives tiny mouse drift between mousedown/mouseup), works for touch
   * + pen + mouse uniformly. Force `currentBeat = -1` so the next rAF tick
   * always re-evaluates and enters the new beat, even if the click landed
   * in the same frame as a prior skip. */
  document.addEventListener('pointerdown', (e) => {
    if (e.target.closest('button, a, input, select, textarea, [data-no-skip]')) return;
    flashClick();
    if (!state.playing) { play(); return; }
    const lastIdx = BEATS.length - 1;
    const currentIdx = state.currentBeat < 0 ? 0 : state.currentBeat;
    if (currentIdx >= lastIdx) {
      /* Past the last beat — restart from the top. */
      state.pausedAt = 0;
      state.startTime = performance.now();
      state.currentBeat = -1;
      SOUND.click();
      return;
    }
    const target = BEATS[currentIdx + 1].start;
    state.startTime = performance.now() - target * 1000;
    state.currentBeat = -1;
    SOUND.click();
  });

  let scrollAccum = 0;
  document.addEventListener('wheel', (e) => {
    if (!state.playing) play();
    e.preventDefault();
    scrollAccum += e.deltaY;
    if (Math.abs(scrollAccum) > 60) {
      const dir = scrollAccum > 0 ? 1 : -1;
      state.pausedAt = Math.max(0, Math.min(TOTAL, state.pausedAt + dir * 2));
      state.startTime = performance.now() - state.pausedAt * 1000;
      scrollAccum = 0;
    }
  }, { passive: false });

  window.addEventListener('resize', () => initParticles());
}

function toggleMute() {
  state.muted = !state.muted;
  localStorage.setItem('condura:muted', state.muted ? '1' : '0');
  const btn = document.getElementById('sound-toggle');
  btn.classList.toggle('is-muted', state.muted);
  document.getElementById('sound-label').textContent = state.muted ? 'sound off' : 'sound on';
}

function toggleFullscreen() {
  if (!document.fullscreenElement) document.documentElement.requestFullscreen();
  else document.exitFullscreen();
}

/* ─── Boot ─────────────────────────────────────────────────────── */

function boot() {
  renderHelix();
  initParticles();
  bindCursor();
  bindControls();
  if (state.muted) {
    document.getElementById('sound-toggle').classList.add('is-muted');
    document.getElementById('sound-label').textContent = 'sound off';
  }
  document.getElementById('gate').classList.add('is-loaded');
  if (localStorage.getItem('condura:played') === '1') setTimeout(play, 600);
}

if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot);
else boot();