/** Meridian routing */
export type RouteId =
  | 'chat' | 'hub' | 'skills' | 'sync' | 'audit' | 'replay'
  | 'channels' | 'delegation' | 'account' | 'settings' | 'about'

export const ROUTE_HASH: Record<RouteId, string> = {
  chat: '#/', hub: '#/hub', skills: '#/skills', sync: '#/sync', audit: '#/audit',
  replay: '#/replay', channels: '#/channels', delegation: '#/delegation',
  account: '#/account', settings: '#/settings', about: '#/about',
}

export function hashToRoute(hash: string): RouteId {
  if (hash.startsWith('#/settings')) return 'settings'
  if (hash.startsWith('#/audit')) return 'audit'
  if (hash.startsWith('#/replay')) return 'replay'
  if (hash.startsWith('#/about')) return 'about'
  if (hash.startsWith('#/hub')) return 'hub'
  if (hash.startsWith('#/sync')) return 'sync'
  if (hash.startsWith('#/skills')) return 'skills'
  if (hash.startsWith('#/channels')) return 'channels'
  if (hash.startsWith('#/delegation')) return 'delegation'
  if (hash.startsWith('#/account')) return 'account'
  return 'chat'
}
