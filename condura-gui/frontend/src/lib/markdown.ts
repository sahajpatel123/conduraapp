// Safe markdown for chat bubbles (Meridian Ask + legacy Chat).
// marked → HTML, then DOMPurify so model/tool output cannot XSS the Wails shell.

import { marked } from 'marked'
import DOMPurify from 'dompurify'

marked.setOptions({
  gfm: true,
  breaks: true,
})

/**
 * Parse markdown to sanitized HTML for {@html} rendering.
 * Empty input returns empty string.
 */
export function renderSafeMarkdown(text: string): string {
  if (!text) return ''
  const html = marked.parse(text, { async: false }) as string
  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
  })
}
