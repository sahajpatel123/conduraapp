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
      fileError = $t('hub.publish.file_too_large')
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
    class="pub-modal"
    role="dialog"
    aria-modal="true"
    aria-label={$t('hub.publish.aria_label')}
    tabindex="-1"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => { if (e.key === 'Escape') close() }}
  >
    <header>
      <h2>{$t('hub.publish.title')}</h2>
      <button class="close" aria-label={$t('hub.publish.close')} onclick={close}>&times;</button>
    </header>

    {#if hub.publishState.kind === 'success'}
      <div class="result ok">
        <p><strong>{$t('hub.publish.published')}</strong> {hub.publishState.result.id} v{hub.publishState.result.version}</p>
        {#if hub.publishState.result.url}
          <a href={hub.publishState.result.url} target="_blank" rel="noreferrer">{$t('hub.publish.view_on_hub')}</a>
        {/if}
        <button class="primary" onclick={close}>{$t('hub.publish.done')}</button>
      </div>
    {:else}
      <div class="form">
        <label>
          {$t('hub.publish.name')}
          <input type="text" bind:value={name} placeholder={$t('hub.publish.name_placeholder')} />
        </label>
        <label>
          {$t('hub.publish.version')}
          <input type="text" bind:value={version} placeholder="1.0.0" class:invalid={version && !versionValid} />
          {#if version && !versionValid}<span class="field-err">{$t('hub.publish.semver_hint')}</span>{/if}
        </label>
        <label>
          {$t('hub.publish.description')}
          <textarea bind:value={description} rows="3" placeholder={$t('hub.publish.description_placeholder')}></textarea>
        </label>
        <div class="two">
          <label>
            {$t('hub.publish.author')}
            <input type="text" bind:value={author} placeholder="you" />
          </label>
          <label>
            {$t('hub.publish.license')}
            <input type="text" bind:value={license} placeholder="MIT" />
          </label>
        </div>
        <label>
          {$t('hub.publish.tags')}
          <input type="text" bind:value={tagsInput} placeholder="weather, api, utility" />
        </label>
        <label class="file">
          {$t('hub.publish.archive')}
          <input type="file" accept=".zip,application/zip" onchange={onFile} />
          {#if fileName}<span class="file-name">{fileName}</span>{/if}
          {#if fileError}<span class="field-err">{fileError}</span>{/if}
        </label>

        {#if hub.publishState.kind === 'error'}
          <p class="result err">{hub.publishState.message}</p>
        {/if}

        <div class="actions">
          <button class="ghost" onclick={close}>{$t('hub.publish.cancel')}</button>
          <button class="primary" disabled={!canSubmit} onclick={submit}>
            {hub.isPublishing ? $t('hub.publish.publishing') : $t('hub.publish.publish')}
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
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    padding: var(--space-4);
  }
  .pub-modal {
    width: 100%;
    max-width: 460px;
    max-height: 90vh;
    overflow-y: auto;
    background: var(--color-bg-elevated, var(--color-bg));
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-xl);
    padding: var(--space-5);
    box-shadow: var(--shadow-lg, 0 20px 60px rgba(0, 0, 0, 0.4));
  }
  header { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-4); }
  h2 { font-size: var(--size-lg); font-weight: 600; }
  .close { background: none; border: none; color: var(--color-text-faint); font-size: 24px; cursor: pointer; line-height: 1; }
  .close:hover { color: var(--color-text); }
  .form { display: flex; flex-direction: column; gap: var(--space-3); }
  label { display: flex; flex-direction: column; gap: var(--space-2); font-size: var(--size-sm); color: var(--color-text-muted); }
  .two { display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-3); }
  input, textarea {
    padding: 9px 12px;
    border-radius: var(--radius-md);
    border: 1px solid var(--glass-border);
    background: rgba(0, 0, 0, 0.3);
    color: var(--color-text);
    font-size: var(--size-md);
    font-family: inherit;
  }
  input:focus, textarea:focus { outline: none; border-color: var(--color-accent); }
  input.invalid { border-color: var(--color-error, #f87171); }
  .field-err { color: var(--color-error, #f87171); font-size: var(--size-xs); }
  .file-name { color: var(--color-text); font-size: var(--size-xs); font-family: var(--font-mono); }
  .actions { display: flex; gap: var(--space-2); justify-content: flex-end; margin-top: var(--space-2); }
  .ghost, .primary { padding: 10px 18px; border-radius: var(--radius-md); font-size: var(--size-md); font-weight: 500; cursor: pointer; }
  .ghost { background: transparent; border: 1px solid var(--glass-border); color: var(--color-text-muted); }
  .primary { background: var(--color-accent-gradient); color: white; border: none; }
  .primary:disabled { opacity: 0.5; cursor: not-allowed; }
  .result { font-size: var(--size-sm); }
  .result.ok { display: flex; flex-direction: column; gap: var(--space-3); align-items: flex-start; }
  .result.ok a { color: var(--color-accent); }
  .result.err { color: var(--color-error, #f87171); }
</style>
