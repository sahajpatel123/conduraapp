// Safe markdown for chat bubbles (Meridian Ask + legacy Chat).
// marked → HTML, then DOMPurify so model/tool output cannot XSS the Wails shell.

import { marked } from 'marked'
import DOMPurify from 'dompurify'

marked.setOptions({
  gfm: true,
  breaks: true,
})

/**
 * Wrap every <pre> in a <figure class="md-code"> with a Copy button.
 * DOMPurify keeps the button because it has no inline handlers — the
 * MeridianChat handler uses event delegation on the .messages container.
 * The wrap runs AFTER sanitize so the markup is safe-by-construction.
 */
function wrapCodeBlocks(html: string): string {
  if (!html.includes('<pre>')) return html
  return html.replace(
    /<pre>([\s\S]*?)<\/pre>/g,
    (_match, inner: string) =>
      `<figure class="md-code"><button type="button" class="md-code-copy" aria-label="Copy code">Copy</button><pre>${inner}</pre></figure>`,
  )
}

/**
 * Parse markdown to sanitized HTML for {@html} rendering.
 * Empty input returns empty string.
 */
export function renderSafeMarkdown(text: string): string {
  if (!text) return ''
  const html = marked.parse(text, { async: false }) as string
  const sanitized = DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
  })
  return wrapCodeBlocks(sanitized)
}
