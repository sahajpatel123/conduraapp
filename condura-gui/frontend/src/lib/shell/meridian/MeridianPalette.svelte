<script lang="ts">
  import type { RouteId } from './routes'
  import { ROUTE_HASH } from './routes'
  interface Props { open: boolean; onclose: () => void; onnavigate: (r: RouteId) => void }
  let { open, onclose, onnavigate }: Props = $props()
  const ITEMS: { id: RouteId; label: string; hint: string }[] = [
    { id: 'chat', label: 'Ask', hint: 'Talk to Condura' },
    { id: 'hub', label: 'Hub', hint: 'Browse skills' },
    { id: 'skills', label: 'Skills', hint: 'Local procedures' },
    { id: 'sync', label: 'Sync', hint: 'Pair a device' },
    { id: 'audit', label: 'Audit', hint: 'Action ledger' },
    { id: 'channels', label: 'Channels', hint: 'Telegram & more' },
    { id: 'delegation', label: 'Agents', hint: 'Pending actions' },
    { id: 'account', label: 'Account', hint: 'Sign in' },
    { id: 'settings', label: 'Settings', hint: 'Defaults' },
    { id: 'about', label: 'About', hint: 'Build · promises · safety' },
  ]
  let q = $state('')
  let idx = $state(0)
  let inputEl = $state<HTMLInputElement | null>(null)
  const filtered = $derived(ITEMS.filter((it) => `${it.label} ${it.hint}`.toLowerCase().includes(q.trim().toLowerCase())))
  $effect(() => { if (!open) return; q = ''; idx = 0; queueMicrotask(() => inputEl?.focus()) })
  $effect(() => { q; idx = 0 })
  function go(id: RouteId): void { onnavigate(id); onclose() }
  function onKey(e: KeyboardEvent): void {
    if (!open) return
    if (e.key === 'Escape') { e.preventDefault(); onclose(); return }
    if (e.key === 'ArrowDown') { e.preventDefault(); idx = Math.min(idx + 1, Math.max(filtered.length - 1, 0)) }
    else if (e.key === 'ArrowUp') { e.preventDefault(); idx = Math.max(idx - 1, 0) }
    else if (e.key === 'Enter') { e.preventDefault(); const hit = filtered[idx]; if (hit) go(hit.id) }
  }
</script>
<svelte:window onkeydown={onKey} />
{#if open}
  <div class="back" onclick={onclose} role="presentation"></div>
  <div class="panel" role="dialog" aria-label="Jump" aria-modal="true">
    <input bind:this={inputEl} bind:value={q} placeholder="Jump anywhere…" class="q" />
    <ul>
      {#each filtered as item, i (item.id)}
        <li>
          <button type="button" class:on={i === idx} onclick={() => go(item.id)}>
            <span>{item.label}</span><span class="hint">{item.hint}</span>
            <kbd>{ROUTE_HASH[item.id].replace('#/', '/') || '/'}</kbd>
          </button>
        </li>
      {:else}<li class="empty">No matches</li>{/each}
    </ul>
  </div>
{/if}
<style>
  .back {
    position: fixed; inset: 0; z-index: 80;
    background: var(--md-scrim);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    animation: md-fade 200ms var(--md-ease) both;
  }
  .panel {
    position: fixed; top: 14vh; left: 50%;
    width: min(520px, calc(100vw - 32px));
    background: var(--md-surface);
    border: 1px solid var(--md-line-strong);
    border-radius: 22px; z-index: 81; overflow: hidden;
    box-shadow: var(--md-shadow-lift);
    transform: translateX(-50%);
    animation: md-palette 320ms var(--md-spring) both;
  }
  .q {
    width: 100%; border: 0; border-bottom: 1px solid var(--md-line);
    padding: 18px 20px; font-size: 16px; background: transparent;
    color: var(--md-ink); outline: none;
  }
  ul { margin: 0; padding: 8px; max-height: 360px; overflow: auto; }
  button {
    width: 100%; display: grid; grid-template-columns: 1fr auto auto; gap: 12px;
    align-items: center; text-align: left; padding: 12px; border-radius: 12px;
    font-weight: 600; cursor: pointer;
    transition: background 140ms var(--md-ease);
  }
  button.on, button:hover { background: color-mix(in oklab, var(--md-cobalt) 10%, transparent); }
  button.on { box-shadow: inset 3px 0 0 var(--md-cobalt); }
  button:focus-visible { outline: none; background: color-mix(in oklab, var(--md-cobalt) 14%, transparent); box-shadow: inset 3px 0 0 var(--md-cobalt), var(--md-focus); }
  .hint { font-size: 12px; font-weight: 500; color: var(--md-ink-faint); }
  kbd { font-family: var(--md-font-mono); font-size: 10px; color: var(--md-ink-faint); }
  .empty { padding: 20px; text-align: center; color: var(--md-ink-faint); font-size: 13px; }
  @keyframes md-palette {
    from { opacity: 0; transform: translateX(-50%) translateY(12px) scale(0.97); }
    to { opacity: 1; transform: translateX(-50%) translateY(0) scale(1); }
  }
  @media (max-width: 560px) {
    .panel {
      top: auto; bottom: 0; left: 0; right: 0;
      width: 100%; max-width: none;
      border-radius: 22px 22px 0 0;
      border-left: 0; border-right: 0; border-bottom: 0;
      transform: none;
      max-height: min(78vh, 640px);
      display: flex; flex-direction: column;
      animation: md-palette-sheet 340ms var(--md-spring) both;
      padding-bottom: env(safe-area-inset-bottom, 0);
    }
    ul { flex: 1; max-height: none; padding: 8px 8px 16px; }
    button { grid-template-columns: 1fr auto; gap: 8px; padding: 14px 12px; }
    button .hint { display: none; }
    kbd { justify-self: end; }
  }
  @keyframes md-palette-sheet {
    from { opacity: 0; transform: translateY(24px); }
    to { opacity: 1; transform: translateY(0); }
  }
  @media (prefers-reduced-motion: reduce) {
    .back, .panel { animation: none !important; }
  }
</style>
