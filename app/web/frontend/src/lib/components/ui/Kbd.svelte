<script lang="ts">
  interface Props {
    label: string
  }
  let { label }: Props = $props()
</script>

<kbd class="kbd">{label}</kbd>

<style>
  .kbd {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 22px;
    height: 22px;
    padding: 0 6px;
    background: var(--surface-3);
    border: 1px solid var(--border-strong);
    border-bottom-width: 2px;
    border-radius: var(--radius-sm);
    color: var(--text);
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-wide);
    line-height: 1;
  }
</style>