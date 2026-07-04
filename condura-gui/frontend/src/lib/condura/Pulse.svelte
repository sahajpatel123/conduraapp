<script lang="ts">
  // Condura Pulse — the agent's vital sign. A breathing dot whose rhythm
  // and color follow the agent's phase. Motion is a verb: idle = calm breath,
  // thinking = quicker, consent = a stuttered warn, error = fast alarm.
  let {
    phase = 'idle',
    size = 8,
    class: cls = '',
  }: {
    phase?: 'idle' | 'thinking' | 'awaiting' | 'acting' | 'consent' | 'error' | 'ok';
    size?: number;
    class?: string;
  } = $props();

  const MAP: Record<string, { color: string; dur: string; steps?: string }> = {
    idle: { color: 'var(--synapse-glow)', dur: '5s' },
    thinking: { color: 'var(--pollen)', dur: '1.8s' },
    awaiting: { color: 'var(--pollen)', dur: '1.2s' },
    acting: { color: 'var(--pollen)', dur: '1s' },
    consent: { color: 'var(--warn)', dur: '1s', steps: 'steps(2)' },
    error: { color: 'var(--danger)', dur: '0.8s' },
    ok: { color: 'var(--synapse-glow)', dur: '4s' },
  };

  let cfg = $derived(MAP[phase] ?? MAP.idle);
  // animation is set inline so Svelte does not scope the global `breathe` keyframe name.
  let style = $derived(
    `width:${size}px;height:${size}px;background:${cfg.color};animation:breathe ${cfg.dur} var(--ease) infinite ${cfg.steps ?? ''}`
  );
</script>

<span class="pulse-dot {cls}" {style} aria-hidden="true"></span>

<style>
  .pulse-dot {
    display: inline-block;
    border-radius: 50%;
    flex: none;
  }
</style>