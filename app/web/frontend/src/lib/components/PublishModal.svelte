<script lang="ts">
  // PublishModal (Phase 14G). Form to publish a skill to the Hub:
  // name, version, description, author, license, tags, archive file.
  // Three columns of form fields at the top, a preview area showing
  // the assembled skill YAML, then the publish button.
  import { Dialog } from './ui'
  import Button from './ui/Button.svelte'
  import Input from './ui/Input.svelte'
  import Textarea from './ui/Textarea.svelte'
  import Select from './ui/Select.svelte'
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

  const previewYaml = $derived(() => {
    const slug = name.trim().toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '')
    const tags = tagsInput
      .split(',')
      .map((tg) => tg.trim())
      .filter(Boolean)
    return [
      `id: ${slug || '<name>'}`,
      `name: ${name.trim() || '<name>'}`,
      `version: ${version.trim() || '0.0.0'}`,
      `description: ${description.trim() || '<description>'}`,
      `author: ${author.trim() || '<author>'}`,
      `license: ${license.trim() || 'MIT'}`,
      `tags: [${tags.length ? tags.map((tg) => `"${tg}"`).join(', ') : ''}]`,
    ].join('\n')
  })

  const licenseOptions = [
    { value: 'MIT', label: 'MIT' },
    { value: 'Apache-2.0', label: 'Apache-2.0' },
    { value: 'BSD-3-Clause', label: 'BSD-3-Clause' },
    { value: 'GPL-3.0', label: 'GPL-3.0' },
    { value: 'ISC', label: 'ISC' },
    { value: 'Proprietary', label: 'Proprietary' },
  ]

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
      .map((tg) => tg.trim())
      .filter(Boolean)
    const slug = name
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '')
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

  const isOpen = $derived(
    hub.publishState.kind !== 'success' || true // keep dialog mounted until close()
  )
</script>

<Dialog
  open={isOpen}
  title={t('hub.publish.title')}
  size="lg"
  onclose={close}
>
  {#snippet children()}
    <div class="publish-body">
      {#if hub.publishState.kind === 'success'}
        <div class="result ok">
          <p>
            <strong>{t('hub.publish.published')}</strong>
            {hub.publishState.result.id}
            v{hub.publishState.result.version}
          </p>
          {#if hub.publishState.result.url}
            <a href={hub.publishState.result.url} target="_blank" rel="noreferrer">
              {t('hub.publish.view_on_hub')}
            </a>
          {/if}
          <Button variant="primary" size="sm" onclick={close}>
            {t('hub.publish.done')}
          </Button>
        </div>
      {:else}
        <div class="grid-three">
          <Input
            fullWidth
            label={t('hub.publish.name')}
            bind:value={name}
            placeholder={t('hub.publish.name_placeholder')}
          />
          <Input
            fullWidth
            label={t('hub.publish.version')}
            bind:value={version}
            placeholder="1.0.0"
            error={version && !versionValid ? t('hub.publish.semver_hint') : undefined}
          />
          <Input
            fullWidth
            label={t('hub.publish.author')}
            bind:value={author}
            placeholder="you"
          />
        </div>

        <Textarea
          fullWidth
          label={t('hub.publish.description')}
          bind:value={description}
          rows={3}
          placeholder={t('hub.publish.description_placeholder')}
        />

        <div class="grid-two">
          <Select
            fullWidth
            label={t('hub.publish.license')}
            options={licenseOptions}
            bind:value={license}
          />
          <Input
            fullWidth
            label={t('hub.publish.tags')}
            bind:value={tagsInput}
            placeholder="weather, api, utility"
          />
        </div>

        <div class="file-field">
          <span class="file-label">{t('hub.publish.archive')}</span>
          <input type="file" accept=".zip,application/zip" onchange={onFile} />
          {#if fileName}<span class="file-name">{fileName}</span>{/if}
          {#if fileError}<span class="field-err">{fileError}</span>{/if}
        </div>

        <div class="preview">
          <span class="preview-label">Preview</span>
          <pre class="preview-yaml">{previewYaml()}</pre>
        </div>

        {#if hub.publishState.kind === 'error'}
          <p class="result err">{hub.publishState.message}</p>
        {/if}
      {/if}
    </div>
  {/snippet}
  {#snippet footer()}
    {#if hub.publishState.kind !== 'success'}
      <Button variant="ghost" onclick={close}>{t('hub.publish.cancel')}</Button>
      <Button
        variant="primary"
        disabled={!canSubmit}
        loading={hub.isPublishing}
        onclick={submit}
      >
        {hub.isPublishing ? t('hub.publish.publishing') : t('hub.publish.publish')}
      </Button>
    {/if}
  {/snippet}
</Dialog>

<style>
  .publish-body {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .grid-three {
    display: grid;
    grid-template-columns: 1.2fr 0.7fr 0.9fr;
    gap: var(--space-3);
  }
  .grid-two {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
  }

  .file-field {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .file-label {
    font-size: var(--size-xs);
    color: var(--text-muted);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-wide);
    text-transform: uppercase;
    padding-left: 2px;
  }
  .file-field input[type='file'] {
    font-size: var(--size-sm);
    color: var(--text-muted);
  }
  .file-name {
    color: var(--text);
    font-size: var(--size-xs);
    font-family: var(--font-mono);
  }
  .field-err {
    color: var(--error);
    font-size: var(--size-xs);
  }

  .preview {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .preview-label {
    font-size: var(--size-xs);
    color: var(--text-muted);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-wide);
    text-transform: uppercase;
    padding-left: 2px;
  }
  .preview-yaml {
    margin: 0;
    padding: var(--space-3);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    line-height: var(--leading-relaxed);
    color: var(--text);
    overflow-x: auto;
    max-height: 160px;
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
    color: var(--accent);
  }
  .result.err {
    color: var(--error);
    margin: 0;
  }
</style>