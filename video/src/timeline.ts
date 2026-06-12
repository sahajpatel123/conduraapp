/*
  The master timeline. Every scene is [from, duration] in frames @ 30 fps.
  Total ≈ 1830 frames = 61 s, inside the brief's 55–65 s window.
*/
export const SCENES = {
  ignition: { from: 0, duration: 96 },     // 0.0 – 3.2 s
  problem: { from: 96, duration: 132 },    // 3.2 – 7.6 s
  touch: { from: 228, duration: 150 },     // 7.6 – 12.6 s
  hotkey: { from: 378, duration: 174 },    // 12.6 – 18.4 s
  voice: { from: 552, duration: 180 },     // 18.4 – 24.4 s
  perception: { from: 732, duration: 186 },// 24.4 – 30.6 s
  montage: { from: 918, duration: 234 },   // 30.6 – 38.4 s
  audit: { from: 1152, duration: 180 },    // 38.4 – 44.4 s
  features: { from: 1332, duration: 192 }, // 44.4 – 50.8 s
  cta: { from: 1524, duration: 210 },      // 50.8 – 57.8 s
  outro: { from: 1734, duration: 96 },     // 57.8 – 61.0 s
} as const;

export const TOTAL = 1830;
