<!--
  Bridges the consent store to the v1 ConsentModal presentation.
-->
<script lang="ts">
  import { consent } from '../../stores/consent.svelte'
  import ConsentModal from './ConsentModal.svelte'

  function formatVerb(kind: string | undefined): string {
    switch (kind?.toLowerCase()) {
      case 'read':
        return 'Read'
      case 'write':
        return 'Write'
      case 'network':
        return 'Send'
      case 'destructive':
        return 'Delete'
      default:
        return 'Act on'
    }
  }

  function approve(): void {
    void consent.approve()
  }

  function deny(): void {
    void consent.deny()
  }
</script>

{#if consent.ticket}
  <ConsentModal
    verb={formatVerb(consent.ticket.action_kind)}
    target={consent.ticket.detail || consent.ticket.actor || 'this application'}
    details={consent.ticket.actor ? `Requested by ${consent.ticket.actor}` : undefined}
    onapprove={approve}
    ondeny={deny}
  />
{/if}
