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
  .page { max-width: 960px; margin: 0 auto; padding: 24px 24px 112px; min-height: 100%; }
  .head {
    display: flex; align-items: flex-end; justify-content: space-between; gap: 20px;
    margin-bottom: 24px; padding-bottom: 18px; border-bottom: 1px solid var(--md-line); position: relative;
  }
  .head::after {
    content: ''; position: absolute; left: 0; bottom: -1px; width: 40px; height: 1.5px;
    border-radius: 1px;
    background: var(--md-cobalt);
    transform-origin: left;
    animation: md-underline 480ms var(--md-ease) 60ms both;
    box-shadow: none;
  }
  .kicker {
    font-family: var(--md-font-mono); font-size: 10px; letter-spacing: 0.12em;
    text-transform: uppercase; color: var(--md-ink-faint); margin: 0 0 8px;
    display: inline-flex; align-items: center; gap: 7px;
    animation: md-rise 360ms var(--md-ease) both;
    font-weight: 500;
  }
  .kicker::before {
    content: '';
    width: 5px; height: 5px; border-radius: 1px;
    background: var(--md-cobalt);
    box-shadow: none;
  }
  h1 {
    font-family: var(--md-font-display); font-size: clamp(28px, 4.2vw, 40px); font-weight: 650;
    letter-spacing: -0.04em; line-height: 1.05; margin: 0 0 8px;
    animation: md-rise 400ms var(--md-ease) 30ms both;
  }
  .lead {
    margin: 0; max-width: 48ch; font-size: 14px; line-height: 1.5; color: var(--md-ink-mute);
    animation: md-rise 400ms var(--md-ease) 60ms both;
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
