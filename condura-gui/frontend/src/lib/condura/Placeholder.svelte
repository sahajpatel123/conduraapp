<script lang="ts">
  import Thread from './Thread.svelte';

  // Coherent skin for routes not yet rebuilt in this pass. Same paper, type,
  // and thread language as everything else — never a grey box. Fleshed out
  // with real data in the next loop iterations.
  let {
    eyebrow,
    title,
    body,
    note = 'This surface is skinned with the Condura design language and wired up in the next pass.',
  }: { eyebrow: string; title: string; body: string; note?: string } = $props();
</script>

<section class="placeholder">
  <div class="eyebrow">{eyebrow}</div>
  <h1 class="headline">{title}</h1>
  <div class="rule"><Thread orientation="h" /></div>
  <p class="lead">{body}</p>
  <p class="small note">{note}</p>
</section>

<style>
  .placeholder {
    max-width: 640px;
    padding-top: var(--space-7);
  }
  .placeholder .headline {
    font-size: clamp(28px, 3vw, 40px);
    margin: var(--space-3) 0 var(--space-4);
  }
  .rule {
    width: 120px;
    margin-bottom: var(--space-5);
  }
  .note {
    margin-top: var(--space-4);
    color: var(--content-faint);
    font-style: italic;
  }
</style>