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

  it('wraps fenced code blocks in <figure class="md-code"> with a Copy button', () => {
    const html = renderSafeMarkdown('```js\nconst x = 1\n```')
    expect(html).toContain('<figure class="md-code">')
    expect(html).toContain('<button type="button" class="md-code-copy"')
    expect(html).toContain('<pre>')
    expect(html).toContain('<code')
    // The button must come BEFORE the <pre> so absolute-positioning lands
    // it in the top-right of the figure without depending on flex order.
    const btnIdx = html.indexOf('md-code-copy')
    const preIdx = html.indexOf('<pre>')
    expect(btnIdx).toBeGreaterThan(-1)
    expect(preIdx).toBeGreaterThan(-1)
    expect(btnIdx).toBeLessThan(preIdx)
  })

  it('does not wrap inline code in <figure>', () => {
    // Single backtick → <code>, not <pre><code>. The Copy button
    // should only appear for block-level code.
    const html = renderSafeMarkdown('use `foo()` here')
    expect(html).not.toContain('md-code')
  })

  it('wraps multiple code blocks independently', () => {
    const html = renderSafeMarkdown('```py\na = 1\n```\n\ntext\n\n```go\nb := 2\n```')
    const matches = html.match(/md-code-copy/g) ?? []
    expect(matches.length).toBe(2)
  })
})
