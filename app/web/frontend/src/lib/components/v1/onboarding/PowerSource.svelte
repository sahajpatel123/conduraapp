<!--
  Onboarding · Screen 4 · Power Source

  Per spec §10.4: THE REAL MANDATORY MOMENT. Without a brain, the agent
  can't act. Four cards in a 2×2 grid (Claude Pro, ChatGPT Plus, API key,
  local model). Skip = "I'll set this up later."

  If skipped: auto-enables Ollama if present, otherwise demo mode.
-->
<script lang="ts">
  import Button from '../Button.svelte';
  import Pulse from '../Pulse.svelte';

  type PowerChoice = 'claude-pro' | 'chatgpt-plus' | 'api-key' | 'ollama';

  interface Props {
    selected?: PowerChoice | null;
    apiKey?: string;
    ollamaDetected?: boolean;
    onselect?: (choice: PowerChoice) => void;
    onapiKeyChange?: (key: string) => void;
    onskip?: () => void;
    onback?: () => void;
  }

  let {
    selected = null,
    apiKey = '',
    ollamaDetected = false,
    onselect,
    onapiKeyChange,
    onskip,
    onback,
  }: Props = $props();

  function handleKeyChange(e: Event) {
    const v = (e.target as HTMLInputElement).value;
    onapiKeyChange?.(v);
  }
</script>

<div class="screen">
  <div class="screen__inner">
    <header class="screen__header">
      <h2 class="screen__title">Pick a brain.</h2>
      <p class="screen__subtitle">
        Synaptic needs at least one language model to think. You can add or change these later.
      </p>
    </header>

    <div class="grid">
      <button
        class="card"
        class:card--selected={selected === 'claude-pro'}
        onclick={() => onselect?.('claude-pro')}
        type="button"
      >
        <div class="card__icon" aria-hidden="true">
          <span class="card__letter">C</span>
        </div>
        <div class="card__name">Use my Claude Pro</div>
        <div class="card__cost">Uses your existing subscription · no extra charge</div>
      </button>

      <button
        class="card"
        class:card--selected={selected === 'chatgpt-plus'}
        onclick={() => onselect?.('chatgpt-plus')}
        type="button"
      >
        <div class="card__icon card__icon--alt" aria-hidden="true">
          <span class="card__letter">G</span>
        </div>
        <div class="card__name">Use my ChatGPT Plus</div>
        <div class="card__cost">Uses your existing subscription · no extra charge</div>
      </button>

      <button
        class="card"
        class:card--selected={selected === 'api-key'}
        onclick={() => onselect?.('api-key')}
        type="button"
      >
        <div class="card__icon card__icon--mono" aria-hidden="true">
          <span class="card__letter">⌘</span>
        </div>
        <div class="card__name">Paste an API key</div>
        <div class="card__cost">Bring your own key · pay-as-you-go</div>
      </button>

      <button
        class="card"
        class:card--selected={selected === 'ollama'}
        onclick={() => onselect?.('ollama')}
        type="button"
      >
        <div class="card__icon card__icon--accent" aria-hidden="true">
          <Pulse state="idle" size="sm" label="" />
        </div>
        <div class="card__name">Use a local model</div>
        <div class="card__cost">
          {ollamaDetected
            ? 'Ollama detected on localhost:11434 · no key needed'
            : 'Install Ollama to enable · free, runs on your machine'}
        </div>
      </button>
    </div>

    {#if selected === 'api-key'}
      <div class="apikey">
        <label for="power-apikey" class="apikey__label">Your API key</label>
        <input
          id="power-apikey"
          class="apikey__input"
          type="password"
          placeholder="sk-..."
          value={apiKey}
          oninput={handleKeyChange}
          autocomplete="off"
          spellcheck="false"
        />
        <p class="apikey__hint">
          Stored locally, encrypted with your machine key. Never sent anywhere except the LLM provider.
        </p>
      </div>
    {/if}

    <div class="screen__actions">
      <Button variant="tertiary" size="md" onclick={onback}>← Back</Button>
      <div class="screen__actions-right">
        <Button variant="tertiary" size="md" onclick={onskip}>
          I'll set this up later in Settings
        </Button>
        <Button
          variant="primary"
          size="md"
          disabled={!selected || (selected === 'api-key' && !apiKey.trim())}
        >
          {selected ? 'Continue' : 'Choose one to continue'}
        </Button>
      </div>
    </div>
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
    max-width: 800px;
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

  /* 2x2 card grid */
  .grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
  }

  .card {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    text-align: left;
    gap: var(--space-2);
    padding: var(--space-5);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition:
      border-color var(--duration-fast) var(--ease-standard),
      background-color var(--duration-fast) var(--ease-standard);
    font-family: var(--font-sans);
  }

  .card:hover {
    border-color: var(--border-strong);
    background-color: var(--paper-warm-50);
  }

  .card:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: 2px;
  }

  .card--selected {
    border-color: var(--content-accent);
    background-color: var(--plum-50);
  }

  .card__icon {
    width: 32px;
    height: 32px;
    border-radius: var(--radius-md);
    background-color: var(--paper-warm-50);
    border: 1px solid var(--border-default);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .card__icon--alt {
    background-color: var(--paper-warm-100);
  }

  .card__icon--mono {
    font-family: var(--font-mono);
  }

  .card__icon--accent {
    background-color: var(--plum-50);
    border-color: var(--plum-200);
  }

  .card__letter {
    font-family: var(--font-serif);
    font-size: 16px;
    font-weight: 600;
    color: var(--content-secondary);
  }

  .card__name {
    font-size: var(--text-body-size);
    font-weight: 500;
    color: var(--content-primary);
    margin-top: var(--space-1);
  }

  .card__cost {
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    line-height: 1.4;
  }

  /* API key input reveal */
  .apikey {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-4);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    animation: apikey-in var(--duration-base) var(--ease-decelerate) both;
  }

  @keyframes apikey-in {
    from {
      opacity: 0;
      transform: translateY(-4px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .apikey__label {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-tertiary);
  }

  .apikey__input {
    width: 100%;
    height: 44px;
    padding: 0 var(--space-3);
    font-family: var(--font-mono);
    font-size: var(--text-body-size);
    color: var(--content-primary);
    background-color: var(--surface-base);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-sm);
    outline: none;
    transition: border-color var(--duration-fast) var(--ease-standard);
  }

  .apikey__input:focus-visible {
    border-color: var(--border-focus);
    box-shadow: 0 0 0 2px var(--border-focus);
  }

  .apikey__hint {
    font-size: var(--text-caption-size);
    color: var(--content-muted);
    margin: 0;
    line-height: 1.5;
  }

  .screen__actions {
    display: flex;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .screen__actions-right {
    display: flex;
    gap: var(--space-3);
    align-items: center;
  }

  @media (max-width: 720px) {
    .grid {
      grid-template-columns: 1fr;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .apikey {
      animation: none;
    }
  }
</style>