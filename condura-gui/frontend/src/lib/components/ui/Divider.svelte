<script lang="ts">
  interface Props {
    label?: string
    orientation?: 'horizontal' | 'vertical'
  }
  let { label, orientation = 'horizontal' }: Props = $props()
</script>

{#if label && orientation === 'horizontal'}
  <div class="divider divider-labeled" role="separator">
    <span class="divider-line"></span>
    <span class="divider-label">{label}</span>
    <span class="divider-line"></span>
  </div>
{:else}
  <span
    class="divider divider-{orientation}"
    role="separator"
    aria-orientation={orientation}
  ></span>
{/if}

<style>
  .divider-horizontal {
    display: block;
    height: 1px;
    width: 100%;
    background: linear-gradient(90deg, transparent, var(--border-strong), transparent);
    margin: var(--space-4) 0;
  }
  .divider-vertical {
    display: inline-block;
    width: 1px;
    align-self: stretch;
    background: linear-gradient(180deg, transparent, var(--border-strong), transparent);
    margin: 0 var(--space-3);
  }

  .divider-labeled {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin: var(--space-4) 0;
  }
  .divider-line {
    flex: 1;
    height: 1px;
    background: var(--border-strong);
  }
  .divider-label {
    font-family: var(--font-mono);
    font-size: var(--size-2xs);
    color: var(--text-faint);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
  }
</style>