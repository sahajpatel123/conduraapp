<!--
  Avatar — NOT a face. Per spec §15: "Synaptic has no avatar. It has a pulse."

  This component represents an identity (user or agent) without anthropomorphizing.
  For the agent, it renders a Pulse. For the user, it renders their initials
  (the first letters of name parts) in a subtle circle. No faces, no robots,
  no glowing orbs with eyes.

  Props:
    name   — display name (initials extracted if no custom)
    variant — 'agent' (Pulse) | 'user' (initials)
    size    — 'sm' | 'md' | 'lg'
-->
<script lang="ts">
  import Pulse from './Pulse.svelte';

  interface Props {
    name?: string;
    variant?: 'agent' | 'user';
    size?: 'sm' | 'md' | 'lg';
  }

  let { name = '', variant = 'agent', size = 'md' }: Props = $props();

  function getInitials(n: string): string {
    if (!n) return '?';
    const parts = n.trim().split(/\s+/).filter(Boolean);
    if (parts.length === 0) return '?';
    if (parts.length === 1) return parts[0][0].toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }

  let initials = $derived(getInitials(name));
</script>

<div class="avatar avatar--{variant} avatar--{size}" aria-label={variant === 'agent' ? 'Synaptic' : name}>
  {#if variant === 'agent'}
    <Pulse state="idle" {size} label="Synaptic" />
  {:else}
    <span class="avatar__initials">{initials}</span>
  {/if}
</div>

<style>
  .avatar {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--radius-pill);
    flex-shrink: 0;
  }

  .avatar--user {
    background-color: var(--plum-100);
    color: var(--plum-700);
    font-family: var(--font-sans);
    font-weight: 500;
    border: 1px solid var(--plum-200);
  }

  .avatar--sm.avatar--user {
    width: 24px;
    height: 24px;
    font-size: var(--text-caption-size);
  }
  .avatar--md.avatar--user {
    width: 32px;
    height: 32px;
    font-size: var(--text-body-sm-size);
  }
  .avatar--lg.avatar--user {
    width: 40px;
    height: 40px;
    font-size: var(--text-body-size);
  }

  .avatar__initials {
    line-height: 1;
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
  }
</style>