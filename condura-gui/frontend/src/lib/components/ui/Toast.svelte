<script lang="ts" module>
  // Toast queue — exposed via `push()` so any component can dispatch toasts
  // without prop-drilling.
  type Tone = 'neutral' | 'success' | 'warn' | 'error' | 'info'

  interface ToastEntry {
    id: number
    tone: Tone
    title: string
    description?: string
    duration: number
  }

  const toasts = $state<ToastEntry[]>([])
  let nextId = 1

  export function push(opts: {
    tone?: Tone
    title: string
    description?: string
    duration?: number
  }): number {
    const id = nextId++
    const entry: ToastEntry = {
      id,
      tone: opts.tone ?? 'neutral',
      title: opts.title,
      description: opts.description,
      duration: opts.duration ?? 4000,
    }
    toasts.push(entry)
    if (entry.duration > 0) {
      setTimeout(() => dismiss(id), entry.duration)
    }
    return id
  }

  export function dismiss(id: number): void {
    const i = toasts.findIndex(t => t.id === id)
    if (i >= 0) toasts.splice(i, 1)
  }
</script>

<script lang="ts">
  // Render the queue.
</script>

<div class="toast-stack" role="region" aria-label="Notifications">
  {#each toasts as t (t.id)}
    <div class="toast toast-{t.tone} anim-slide-up" role="status">
      <div class="toast-mark"></div>
      <div class="toast-body">
        <div class="toast-title">{t.title}</div>
        {#if t.description}<div class="toast-description">{t.description}</div>{/if}
      </div>
      <button
        type="button"
        class="toast-close"
        aria-label="Dismiss"
        onclick={() => dismiss(t.id)}
      >
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <path d="M18 6L6 18M6 6l12 12" />
        </svg>
      </button>
    </div>
  {/each}
</div>

<style>
  .toast-stack {
    position: fixed;
    bottom: var(--space-5);
    right: var(--space-5);
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    z-index: var(--z-toast);
    width: 380px;
    max-width: calc(100vw - var(--space-7));
    pointer-events: none;
  }

  .toast {
    pointer-events: auto;
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
  }

  .toast-mark {
    flex-shrink: 0;
    width: 3px;
    align-self: stretch;
    border-radius: 2px;
  }
  .toast-neutral .toast-mark { background: var(--text-muted); }
  .toast-success .toast-mark { background: var(--success); }
  .toast-warn    .toast-mark { background: var(--warn); }
  .toast-error   .toast-mark { background: var(--error); }
  .toast-info    .toast-mark { background: var(--info); }

  .toast-body { flex: 1; min-width: 0; }
  .toast-title {
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    color: var(--text);
  }
  .toast-description {
    margin-top: 2px;
    font-size: var(--size-xs);
    color: var(--text-muted);
    line-height: var(--leading-normal);
  }

  .toast-close {
    appearance: none;
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-sm);
    color: var(--text-faint);
    width: 22px;
    height: 22px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    flex-shrink: 0;
  }
  .toast-close:hover { color: var(--text); background: var(--surface-3); }
</style>