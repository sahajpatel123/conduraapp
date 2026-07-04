<script lang="ts">
  import { tick } from 'svelte'

  interface Props {
    value?: string
    label?: string
    hint?: string
    placeholder?: string
    rows?: number
    disabled?: boolean
    fullWidth?: boolean
    autoresize?: boolean
    oninput?: (e: Event) => void
    onkeydown?: (e: KeyboardEvent) => void
  }

  let { value = $bindable(''), label, hint, placeholder, rows = 4,
        disabled = false, fullWidth = false, autoresize = false,
        oninput, onkeydown }: Props = $props()

  let taEl = $state<HTMLTextAreaElement | null>(null)

  function autoGrow(): void {
    if (!taEl) return
    taEl.style.height = 'auto'
    taEl.style.height = taEl.scrollHeight + 'px'
  }

  $effect(() => {
    if (autoresize && value !== undefined) {
      void tick().then(autoGrow)
    }
  })
</script>

<label class="textarea-wrap" class:textarea-full={fullWidth}>
  {#if label}<span class="textarea-label">{label}</span>{/if}
  <textarea
    bind:this={taEl}
    bind:value
    {placeholder}
    {rows}
    {disabled}
    class="textarea-control"
    {oninput}
    {onkeydown}
  ></textarea>
  {#if hint}<span class="textarea-hint">{hint}</span>{/if}
</label>

<style>
  .textarea-wrap {
    display: inline-flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }
  .textarea-full { width: 100%; }

  .textarea-label {
    font-size: var(--size-xs);
    color: var(--text-muted);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-wide);
    text-transform: uppercase;
    padding-left: 2px;
  }

  .textarea-control {
    background: var(--surface-1);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    color: var(--text);
    font-family: var(--font-sans);
    font-size: var(--size-md);
    line-height: var(--leading-normal);
    padding: var(--space-3);
    resize: vertical;
    min-height: 80px;
    transition: border-color var(--transition-fast) ease,
                background-color var(--transition-fast) ease,
                box-shadow var(--transition-fast) ease;
  }
  .textarea-control::placeholder { color: var(--text-faint); }
  .textarea-control:focus {
    outline: none;
    border-color: var(--border-focus);
    background: var(--surface-2);
    box-shadow: 0 0 0 3px var(--accent-soft);
  }
  .textarea-control:disabled { opacity: 0.5; cursor: not-allowed; }

  .textarea-hint { font-size: var(--size-xs); color: var(--text-faint); }
</style>