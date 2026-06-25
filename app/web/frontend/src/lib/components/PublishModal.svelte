<script lang="ts">
  // PublishModal (Phase 14G). Form to publish a skill to the Hub:
  // name, version, description, and an archive (.zip) file picker.
  // Submit funnels through the hub store's publish flow, which
  // tracks uploading/success/error so we can show progress.
  import { hub } from '../stores/hub.svelte'
  import { t } from '../i18n'

  interface Props {
    onClose: () => void
  }
  let { onClose }: Props = $props()

  let name = $state('')
  let version = $state('')
  let description = $state('')
  let author = $state('')
  let license = $state('MIT')
  let tagsInput = $state('')
  let fileName = $state('')
  let archive = $state<Uint8Array | null>(null)
  let fileError = $state('')

  const semver = /^\d+\.\d+\.\d+(?:[-+].+)?$/
  const versionValid = $derived(semver.test(version.trim()))
  const canSubmit = $derived(
    name.trim().length > 0 && versionValid && archive !== null && !hub.isPublishing
  )

  async function onFile(e: Event): Promise<void> {
    fileError = ''
    const input = e.target as HTMLInputElement
    const f = input.files?.[0]
    if (!f) {
      archive = null
      fileName = ''
      return
    }
    // 32 MB cap matches the daemon's hub archive limit.
    if (f.size > 32 * 1024 * 1024) {
      fileError = t('hub.publish.file_too_large')
      archive = null
      fileName = ''
      return
    }
    try {
      const buf = await f.arrayBuffer()
      archive = new Uint8Array(buf)
      fileName = f.name
    } catch (err) {
      fileError = String(err)
      archive = null
      fileName = ''
    }
  }

  async function submit(): Promise<void> {
    if (!archive) return
    const tags = tagsInput
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean)
    const slug = name.trim().toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '')
    await hub.publishWithParams({
      id: `${slug}@${version.trim()}`,
      archive,
      name: name.trim(),
      version: version.trim(),
      description: description.trim(),
      author: author.trim(),
      license: license.trim(),
      tags,
    })
  }

  function close(): void {
    hub.resetPublishState()
    onClose()
  }
</script>

<div class="pub-backdrop" role="presentation" onclick={close}>
  <div
    class="pub-modal glass-card elevated"
    role="dialog"
    aria-modal="true"
    aria-label={t('hub.publish.aria_label')}
    tabindex="-1"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => { if (e.key === 'Escape') close() }}
  >
    <header>
      <h2>{t('hub.publish.title')}</h2>
      <button class="close" aria-label={t('hub.publish.close')} onclick={close}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M18 6L6 18M6 6l12 12" /></svg>
      </button>
    </header>

    {#if hub.publishState.kind === 'success'}
      <div class="result ok">
        <p><strong>{t('hub.publish.published')}</strong> {hub.publishState.result.id} v{hub.publishState.result.version}</p>
        {#if hub.publishState.result.url}
          <a href={hub.publishState.result.url} target="_blank" rel="noreferrer">{t('hub.publish.view_on_hub')}</a>
        {/if}
        <button class="btn btn-primary btn-sm" onclick={close}>{t('hub.publish.done')}</button>
      </div>
    {:else}
      <div class="form">
        <label>
          {t('hub.publish.name')}
          <input class="input" type="text" bind:value={name} placeholder={t('hub.publish.name_placeholder')} />
        </label>
        <label>
          {t('hub.publish.version')}
          <input class="input" type="text" bind:value={version} placeholder="1.0.0" class:invalid={version && !versionValid} />
          {#if version && !versionValid}<span class="field-err">{t('hub.publish.semver_hint')}</span>{/if}
        </label>
        <label>
          {t('hub.publish.description')}
          <textarea class="input" bind:value={description} rows="3" placeholder={t('hub.publish.description_placeholder')}></textarea>
        </label>
        <div class="two">
          <label>
            {t('hub.publish.author')}
            <input class="input" type="text" bind:value={author} placeholder="you" />
          </label>
          <label>
            {t('hub.publish.license')}
            <input class="input" type="text" bind:value={license} placeholder="MIT" />
          </label>
        </div>
        <label>
          {t('hub.publish.tags')}
          <input class="input" type="text" bind:value={tagsInput} placeholder="weather, api, utility" />
        </label>
        <label class="file">
          {t('hub.publish.archive')}
          <input type="file" accept=".zip,application/zip" onchange={onFile} />
          {#if fileName}<span class="file-name">{fileName}</span>{/if}
          {#if fileError}<span class="field-err">{fileError}</span>{/if}
        </label>

        {#if hub.publishState.kind === 'error'}
          <p class="result err">{hub.publishState.message}</p>
        {/if}

        <div class="actions">
          <button class="btn btn-ghost" onclick={close}>{t('hub.publish.cancel')}</button>
          <button class="btn btn-primary" disabled={!canSubmit} onclick={submit}>
            {hub.isPublishing ? t('hub.publish.publishing') : t('hub.publish.publish')}
          </button>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .pub-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(20, 17, 11, 0.45);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    padding: var(--space-4);
    animation: backdrop-in var(--transition-base) ease both;
  }
  .pub-modal {
    width: 100%;
    max-width: 460px;
    max-height: 90vh;
    overflow-y: auto;
    padding: var(--space-5);
    animation: modal-in var(--transition-spring) var(--ease-out-expo) both;
  }
  .pub-modal:hover {
    border-color: var(--glass-border-hover);
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-4);
  }
  h2 {
    font-size: var(--size-lg);
    font-weight: var(--weight-semibold);
  }
  .close {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    background: none;
    border: none;
    color: var(--color-text-faint);
    cursor: pointer;
    border-radius: var(--radius-sm);
    transition: color var(--transition-base), background var(--transition-base);
  }
  .close svg { width: 16px; height: 16px; }
  .close:hover {
    color: var(--color-text);
    background: var(--glass-bg-hover);
  }
  .form {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  label {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    font-size: var(--size-sm);
    color: var(--color-text-muted);
  }
  .two {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
  }
  .pub-modal input.invalid {
    border-color: var(--color-error);
  }
  .field-err {
    color: var(--color-error);
    font-size: var(--size-xs);
  }
  .file-name {
    color: var(--color-text);
    font-size: var(--size-xs);
    font-family: var(--font-mono);
  }
  .file input[type='file'] {
    font-size: var(--size-sm);
    color: var(--color-text-muted);
  }
  .actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
    margin-top: var(--space-2);
  }
  .result {
    font-size: var(--size-sm);
  }
  .result.ok {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    align-items: flex-start;
  }
  .result.ok a {
    color: var(--color-accent);
  }
  .result.err {
    color: var(--color-error);
  }
</style>
