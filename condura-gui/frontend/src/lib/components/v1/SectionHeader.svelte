<!--
  SectionHeader — section header for settings, panels, anywhere.

  Composes a numbered section label (mono) + icon + title (serif) + subtitle
  (sans tertiary). Used in SettingsPane, action replay sections, anywhere
  a screen has multiple distinct areas.

  Layout:
    [01] [icon] Title
           Subtitle text

  Props:
    number    — '01', '02' etc. (mono caps, plum)
    icon      — IconName (optional)
    title     — section title (serif)
    subtitle  — optional secondary line
    children  — optional slot for trailing actions (buttons, switch)
-->
<script lang="ts">
  import Icon, { type IconName } from './icons/Icon.svelte';

  interface Props {
    number: string;
    icon?: IconName;
    title: string;
    subtitle?: string;
    children?: import('svelte').Snippet;
  }

  let { number, icon, title, subtitle, children }: Props = $props();
</script>

<header class="section-head">
  <div class="section-head__main">
    <div class="section-head__row">
      <span class="section-head__number">{number}</span>
      {#if icon}
        <span class="section-head__icon" aria-hidden="true">
          <Icon name={icon} size="md" />
        </span>
      {/if}
      <h2 class="section-head__title">{title}</h2>
    </div>
    {#if subtitle}
      <p class="section-head__subtitle">{subtitle}</p>
    {/if}
  </div>
  {#if children}
    <div class="section-head__actions">
      {@render children()}
    </div>
  {/if}
</header>

<style>
  .section-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-4);
    padding-bottom: var(--space-3);
    margin-bottom: var(--space-4);
    border-bottom: 1px solid var(--border-subtle);
  }

  .section-head__main {
    flex: 1;
    min-width: 0;
  }

  .section-head__row {
    display: flex;
    align-items: baseline;
    gap: var(--space-3);
    margin-bottom: var(--space-1);
  }

  /* The number — plum, mono, tabular */
  .section-head__number {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    font-weight: 500;
    color: var(--content-accent);
    letter-spacing: 0.06em;
    font-variant-numeric: tabular-nums;
    flex-shrink: 0;
  }

  /* The icon — optional, decorative but informative */
  .section-head__icon {
    color: var(--content-secondary);
    display: inline-flex;
    align-items: center;
    flex-shrink: 0;
  }

  /* The title — serif, the voice of the section */
  .section-head__title {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    line-height: 1.3;
    font-weight: 400;
    color: var(--content-primary);
    margin: 0;
    flex: 1;
    min-width: 0;
  }

  /* The subtitle — sans, small, tertiary */
  .section-head__subtitle {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    line-height: 1.5;
    margin: 0;
    max-width: 56ch;
  }

  /* Trailing actions (buttons, switches) */
  .section-head__actions {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
</style>