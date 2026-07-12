/**
 * Slash-token helpers for Meridian Ask.
 *
 * Convention: a local skill named "UI" is invoked as `/UI …`.
 * Tokens also match skill id / slug / trigger_pattern.
 */

import type { InstalledSkill } from './ipc/types'

export type SlashParse =
  | { kind: 'none'; text: string }
  | { kind: 'skill'; token: string; rest: string; skill: InstalledSkill }
  | { kind: 'builtin'; token: string; rest: string }

const BUILTINS = new Set(['help', 'clear', 'model', 'about', 'compact'])

/** `/UI redesign` → token=UI, rest=redesign */
export function parseLeadingSlash(text: string): { token: string; rest: string } | null {
  const m = text.trim().match(/^\/([A-Za-z0-9._-]+)(?:\s+([\s\S]*))?$/)
  if (!m) return null
  return { token: m[1]!, rest: (m[2] ?? '').trim() }
}

export function skillSlashAliases(s: InstalledSkill): string[] {
  const out = new Set<string>()
  const add = (v: string | undefined | null) => {
    const t = (v ?? '').trim()
    if (!t) return
    out.add(t)
    out.add(t.replace(/^\/+/, ''))
  }
  add(s.id)
  add(s.name)
  add(s.name.replace(/\s+/g, ''))
  add(s.name.replace(/\s+/g, '-'))
  add(slugify(s.name))
  add(s.hub_id)
  add(s.trigger_pattern)
  // id may be "weather-lookup@1.0.0"
  if (s.id?.includes('@')) add(s.id.split('@')[0])
  return [...out]
}

export function resolveSkillSlash(
  token: string,
  skills: InstalledSkill[]
): InstalledSkill | null {
  const needle = token.toLowerCase()
  for (const s of skills) {
    for (const a of skillSlashAliases(s)) {
      if (a.toLowerCase() === needle) return s
    }
  }
  return null
}

export function parseAskSlash(
  text: string,
  skills: InstalledSkill[]
): SlashParse {
  const parsed = parseLeadingSlash(text)
  if (!parsed) return { kind: 'none', text }
  const { token, rest } = parsed
  if (BUILTINS.has(token.toLowerCase())) {
    return { kind: 'builtin', token: token.toLowerCase(), rest }
  }
  const skill = resolveSkillSlash(token, skills)
  if (skill) return { kind: 'skill', token, rest, skill }
  return { kind: 'none', text }
}

/** Primary token shown in Skills UI / Use in Ask, e.g. `/UI`. */
export function primarySlashToken(s: InstalledSkill): string {
  const name = (s.name || '').trim()
  if (name && !/\s/.test(name) && /^[A-Za-z0-9._-]+$/.test(name)) {
    return `/${name}`
  }
  const slug = slugify(s.name || s.id || 'skill')
  return `/${slug || 'skill'}`
}

export function buildSkillSystemPrompt(s: InstalledSkill): string {
  const steps = (s.steps ?? []).filter((x) => !!String(x).trim())
  const lines = [
    `You are executing the local Condura skill “${s.name}” (id: ${s.id}).`,
    'Follow this skill as a procedure. Do not invent unrelated steps.',
    'Ask for consent before any gated / destructive action — the Gatekeeper still applies.',
  ]
  if (s.description?.trim()) {
    lines.push(`Skill description: ${s.description.trim()}`)
  }
  if (steps.length) {
    lines.push('Procedure steps:')
    steps.forEach((step, i) => lines.push(`${i + 1}. ${step}`))
  } else {
    lines.push(
      'This skill has no structured steps yet — explain what you can do with it, then wait for the user.'
    )
  }
  return lines.join('\n')
}

export function slugify(s: string): string {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

export function filterSlashSuggestions(
  draft: string,
  skills: InstalledSkill[]
): { label: string; hint: string; insert: string }[] {
  if (!draft.startsWith('/')) return []
  const q = draft.slice(1).split(/\s/)[0]?.toLowerCase() ?? ''
  const builtins = [
    { label: '/help', hint: 'Show slash help', insert: '/help ' },
    { label: '/clear', hint: 'Clear composer', insert: '/clear' },
    { label: '/model', hint: 'Open Models settings', insert: '/model' },
  ]
  const skillItems = skills.map((s) => {
    const token = primarySlashToken(s)
    return {
      label: token,
      hint: s.description?.slice(0, 72) || s.name,
      insert: `${token} `,
    }
  })
  return [...builtins, ...skillItems].filter((item) => {
    if (!q) return true
    return item.label.toLowerCase().includes(q) || item.hint.toLowerCase().includes(q)
  }).slice(0, 12)
}
