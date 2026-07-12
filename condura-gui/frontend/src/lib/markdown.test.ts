import { describe, it, expect } from 'vitest'
import { renderSafeMarkdown } from './markdown'

describe('renderSafeMarkdown', () => {
  it('renders basic markdown', () => {
    const html = renderSafeMarkdown('**bold** and `code`')
    expect(html).toContain('<strong>bold</strong>')
    expect(html).toContain('<code>code</code>')
  })

  it('strips script and event handlers', () => {
    const html = renderSafeMarkdown('<script>alert(1)</script>hello<img src=x onerror=alert(1)>')
    expect(html.toLowerCase()).not.toContain('<script')
    expect(html.toLowerCase()).not.toContain('onerror')
    expect(html).toContain('hello')
  })

  it('returns empty for empty input', () => {
    expect(renderSafeMarkdown('')).toBe('')
  })
})
