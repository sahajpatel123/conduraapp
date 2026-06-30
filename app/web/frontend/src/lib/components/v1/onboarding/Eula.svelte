<!--
  Onboarding · Screen 2 · EULA

  Mood: a contract between adults.
  Per spec §10.2: the license scrolls inside the agent surface itself,
  in serif at body size. Accept button is disabled until scrolled to bottom.

  On accept, a 1-second personality moment: "Thank you. Now I can read
  these words aloud if you'd like."
-->
<script lang="ts">
  import Button from '../Button.svelte';
  import { onMount } from 'svelte';

  interface Props {
    onaccept?: () => void;
    onback?: () => void;
  }

  let { onaccept, onback }: Props = $props();

  let scrollerEl: HTMLDivElement | undefined = $state();
  let scrolledToEnd = $state(false);
  let accepted = $state(false);
  let thanksVisible = $state(false);

  // Detect when user has scrolled to bottom of license.
  function onScroll(e: Event) {
    const el = e.target as HTMLDivElement;
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 8) {
      scrolledToEnd = true;
    }
  }

  async function handleAccept() {
    accepted = true;
    thanksVisible = true;
    await new Promise((r) => setTimeout(r, 1000));
    thanksVisible = false;
    onaccept?.();
  }

  onMount(() => {
    // Svelte 5: bind:this on the element handles this; just verify.
    scrollerEl?.focus();
  });
</script>

<div class="screen">
  <div class="screen__inner">
    <header class="screen__header">
      <h2 class="screen__title">A license, briefly.</h2>
      <p class="screen__subtitle">This is shorter than most. Read it.</p>
    </header>

    {#if thanksVisible}
      <p class="screen__thanks">Thank you. Now I can read these words aloud if you'd like.</p>
    {:else}
      <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
      <div
        class="screen__scroller"
        bind:this={scrollerEl}
        onscroll={onScroll}
        tabindex="0"
        role="region"
        aria-label="End-user license agreement"
      >
        <article class="license">
          <h3>Synaptic Freeware License v1</h3>
          <p>This software is free for personal and commercial use. You may install and run it on any computer you own or are authorized to use.</p>
          <p>You may not redistribute the binary, modified or unmodified, without written permission. The source code is proprietary and not redistributable.</p>
          <p>The software is provided "as is", without warranty of any kind. The authors are not liable for any damages arising from use.</p>
          <p>This license may be revoked for abuse, including but not limited to: use of the software to harm others, violation of any applicable law, or attempts to circumvent the safety layer (Gatekeeper, audit log, kill switch).</p>
          <p>The software respects your privacy. No telemetry is collected. No data leaves your computer unless you explicitly configure an LLM provider and the agent takes a network action on your behalf, in which case the network call is logged in the local audit log.</p>
          <p>By accepting, you confirm that you understand this is a tool that performs physical actions on your computer, and that you are responsible for its configuration and use.</p>
          <p style="margin-top: 2em;">— The Synaptic team</p>
        </article>
      </div>

      <div class="screen__actions">
        <Button variant="tertiary" size="md" onclick={onback}>← Back</Button>
        <Button
          variant="primary"
          size="md"
          disabled={!scrolledToEnd || accepted}
          onclick={handleAccept}
        >
          {scrolledToEnd ? 'I accept' : 'Scroll to read'}
        </Button>
      </div>
    {/if}
  </div>
</div>

<style>
  .screen {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: var(--space-9);
    background-color: var(--surface-base);
    color: var(--content-primary);
  }

  .screen__inner {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
    width: 100%;
    max-width: 640px;
  }

  .screen__header {
    text-align: left;
  }

  .screen__title {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    line-height: 1.3;
    font-weight: 400;
    color: var(--content-primary);
    margin: 0 0 var(--space-2) 0;
  }

  .screen__subtitle {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    margin: 0;
  }

  /* The scrollable license region */
  .screen__scroller {
    max-height: 50vh;
    overflow-y: auto;
    padding: var(--space-6);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    outline: none;
  }

  .screen__scroller:focus-visible {
    border-color: var(--border-focus);
    box-shadow: 0 0 0 2px var(--border-focus);
  }

  .license {
    font-family: var(--font-serif);
    font-size: var(--text-body-size);
    line-height: 1.7;
    color: var(--content-secondary);
    max-width: 38em;
  }

  .license h3 {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    color: var(--content-primary);
    margin: 0 0 var(--space-4) 0;
    font-weight: 600;
  }

  .license p {
    margin: 0 0 var(--space-4) 0;
  }

  .screen__actions {
    display: flex;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .screen__thanks {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    color: var(--content-primary);
    text-align: center;
    padding: var(--space-9) 0;
    animation: fade-in var(--duration-base) var(--ease-decelerate) both;
  }

  @keyframes fade-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @media (prefers-reduced-motion: reduce) {
    .screen__thanks {
      animation: none;
    }
  }
</style>