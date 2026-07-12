import { describe, expect, it } from 'vitest'
import {
  buildSkillSystemPrompt,
  parseAskSlash,
  parseLeadingSlash,
  primarySlashToken,
  resolveSkillSlash,
  slugify,
} from './skill-slash'
import type { InstalledSkill } from './ipc/types'

const ui: InstalledSkill = {
  id: 'ui',
  name: 'UI',
  description: 'UI craft',
  version: '1.0.0',
  author: '',
  license: '',
  trust: 'experimental',
  source: 'local',
  steps: ['Inspect the surface', 'Propose a change'],
  trigger_pattern: '/UI',
}

const weather: InstalledSkill = {
  id: 'weather-lookup@1.0.0',
  name: 'Weather Lookup',
  description: 'Forecasts',
  version: '1.0.0',
  author: '',
  license: '',
  trust: 'community',
  source: 'hub',
  hub_id: 'weather-lookup',
}

describe('skill-slash', () => {
  it('parses /UI redesign', () => {
    expect(parseLeadingSlash('/UI redesign the toolbar')).toEqual({
      token: 'UI',
      rest: 'redesign the toolbar',
    })
  })

  it('resolves skill by short name', () => {
    expect(resolveSkillSlash('UI', [ui, weather])?.id).toBe('ui')
    expect(resolveSkillSlash('weather-lookup', [ui, weather])?.id).toBe(
      'weather-lookup@1.0.0'
    )
  })

  it('parseAskSlash attaches skill', () => {
    const r = parseAskSlash('/UI make it denser', [ui])
    expect(r.kind).toBe('skill')
    if (r.kind === 'skill') {
      expect(r.skill.id).toBe('ui')
      expect(r.rest).toBe('make it denser')
    }
  })

  it('primarySlashToken prefers clean name', () => {
    expect(primarySlashToken(ui)).toBe('/UI')
    expect(primarySlashToken(weather)).toBe('/weather-lookup')
  })

  it('buildSkillSystemPrompt includes steps', () => {
    const p = buildSkillSystemPrompt(ui)
    expect(p).toContain('Inspect the surface')
    expect(p).toContain('UI')
  })

  it('slugify', () => {
    expect(slugify('Weather Lookup')).toBe('weather-lookup')
  })
})
