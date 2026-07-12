<script lang="ts">
  import type { Snippet } from 'svelte'
  interface Props {
    kicker?: string
    title: string
    lead?: string
    actions?: Snippet
    children?: Snippet
  }
  let { kicker = 'Meridian', title, lead = '', actions, children }: Props = $props()
</script>

<div class="page md-enter">
  <header class="head">
    <div class="copy">
      <p class="kicker">{kicker}</p>
      <h1>{title}</h1>
      {#if lead}<p class="lead">{lead}</p>{/if}
    </div>
    {#if actions}<div class="actions">{@render actions()}</div>{/if}
  </header>
  <div class="body">{#if children}{@render children()}{/if}</div>
</div>

<style>
  .page { max-width: 980px; margin: 0 auto; padding: 28px 28px 120px; min-height: 100%; }
  .head {
    display: flex; align-items: flex-end; justify-content: space-between; gap: 24px;
    margin-bottom: 32px; padding-bottom: 24px; border-bottom: 1px solid var(--md-line); position: relative;
  }
  .head::after {
    content: ''; position: absolute; left: 0; bottom: -1px; width: 88px; height: 3px;
    border-radius: 3px;
    background: linear-gradient(90deg, var(--md-halt), var(--md-ember, var(--md-halt)));
    transform-origin: left;
    animation: md-underline 560ms var(--md-ease) 80ms both;
    box-shadow: 0 0 16px color-mix(in oklab, var(--md-halt) 35%, transparent);
  }
  .kicker {
    font-family: var(--md-font-mono); font-size: 11px; letter-spacing: 0.16em;
    text-transform: uppercase; color: var(--md-ink-faint); margin: 0 0 10px;
    display: inline-flex; align-items: center; gap: 8px;
    animation: md-rise 420ms var(--md-ease) both;
  }
  .kicker::before {
    content: '';
    width: 6px; height: 6px; border-radius: 2px;
    background: linear-gradient(135deg, var(--md-cobalt), var(--md-live));
    box-shadow: 0 0 10px color-mix(in oklab, var(--md-cobalt) 40%, transparent);
  }
  h1 {
    font-family: var(--md-font-display); font-size: clamp(34px, 5vw, 52px); font-weight: 700;
    letter-spacing: -0.055em; line-height: 0.96; margin: 0 0 10px;
    animation: md-rise 480ms var(--md-ease) 40ms both;
  }
  .lead {
    margin: 0; max-width: 46ch; font-size: 15px; line-height: 1.55; color: var(--md-ink-mute);
    animation: md-rise 480ms var(--md-ease) 90ms both;
  }
  .actions { display: flex; gap: 8px; flex: none; flex-wrap: wrap; animation: md-rise 480ms var(--md-ease) 120ms both; }
  .body { animation: md-rise 520ms var(--md-ease) 140ms both; }
  @keyframes md-underline { from { transform: scaleX(0); } to { transform: scaleX(1); } }
  @media (max-width: 640px) {
    .page { padding: 20px 14px 112px; }
    .head {
      flex-direction: column; align-items: stretch; gap: 14px;
      margin-bottom: 24px; padding-bottom: 18px;
    }
    .head::after { width: 64px; height: 2.5px; }
    h1 { font-size: clamp(30px, 9vw, 40px); margin-bottom: 8px; }
    .lead { font-size: 14px; max-width: none; }
    .actions { width: 100%; }
    .actions :global(.md-btn) { flex: 1; justify-content: center; min-width: 0; }
  }
  @media (max-width: 420px) {
    .page { padding: 16px 12px 104px; }
    .kicker { margin-bottom: 8px; font-size: 10px; }
    .actions { gap: 6px; }
    .actions :global(.md-btn) { padding: 9px 12px; font-size: 12px; }
  }
  @media (prefers-reduced-motion: reduce) {
    .head::after, .kicker, h1, .lead, .actions, .body { animation: none !important; }
  }
</style>
