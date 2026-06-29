<script lang="ts">
  // ConfirmDialog — generic confirmation dialog built on the Dialog primitive.
  //
  // Two-button confirmation. Tone=primary → confirm button is btn-primary.
  // tone=danger → confirm button is btn-danger and the dialog border picks
  // up a danger accent.
  import { Dialog } from './ui'
  import Button from './ui/Button.svelte'
  import { t } from '../i18n'

  type Tone = 'primary' | 'danger'

  interface Props {
    open: boolean
    title: string
    description: string
    confirmLabel?: string
    cancelLabel?: string
    tone?: Tone
    onconfirm: () => void
    oncancel?: () => void
  }

  let {
    open = $bindable(false),
    title,
    description,
    confirmLabel,
    cancelLabel,
    tone = 'primary',
    onconfirm,
    oncancel,
  }: Props = $props()

  function handleConfirm(): void {
    open = false
    onconfirm()
  }

  function handleCancel(): void {
    open = false
    oncancel?.()
  }
</script>

<Dialog
  bind:open
  {title}
  description={description}
  size="sm"
  onclose={handleCancel}
>
  {#snippet children()}
    <div class="confirm-body" data-tone={tone}>
      <p class="confirm-message">{description}</p>
    </div>
  {/snippet}
  {#snippet footer()}
    <div class="confirm-footer">
      <Button variant="ghost" onclick={handleCancel}>
        {cancelLabel ?? t('common.cancel')}
      </Button>
      <Button
        variant={tone === 'danger' ? 'danger' : 'primary'}
        onclick={handleConfirm}
      >
        {confirmLabel ?? t('common.confirm')}
      </Button>
    </div>
  {/snippet}
</Dialog>

<style>
  .confirm-body {
    padding: 0;
  }
  .confirm-message {
    color: var(--text);
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    margin: 0;
  }
  .confirm-footer {
    display: flex;
    gap: var(--space-2);
    margin-left: auto;
  }
</style>