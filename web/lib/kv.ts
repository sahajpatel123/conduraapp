// Shared KV helper for the Next.js web app.
//
// In production we use Vercel KV. In local development (when
// @vercel/kv is not installed / not configured) we fall back to
// an in-process Map. The Map is exported as a singleton so that
// all API routes share the same ephemeral store.

import { randomBytes } from 'node:crypto'

type KV = {
  set: (key: string, value: string, opts?: { ex?: number }) => Promise<unknown>
  get: (key: string) => Promise<{ value: string } | null>
  del: (key: string) => Promise<unknown>
}

let kv: KV | null = null

try {
  const mod = (await import('@vercel/kv')) as { kv: KV }
  kv = mod.kv
} catch {
  // @vercel/kv not installed. Will fall back to the dev store.
}

interface DevRecord {
  value: string
  expiresAt: number
}

const devStore = new Map<string, DevRecord>()

export const IS_DEV = process.env.NODE_ENV !== 'production'

export function getKV(): KV | null {
  return kv
}

export function getDevStore(): Map<string, DevRecord> {
  return devStore
}

export function hasKV(): boolean {
  return kv !== null
}

// generateMagicToken creates a 256-bit hex token suitable for
// magic-link URLs. Exported here so both /api/auth/magic and
// /api/auth/verify agree on token shape.
export function generateMagicToken(): string {
  return randomBytes(32).toString('hex')
}

// storeMagicToken persists a magic-link record. Handles both
// Vercel KV and the dev fallback.
export async function storeMagicToken(
  token: string,
  value: string,
  ttlSeconds: number
): Promise<void> {
  if (kv) {
    await kv.set(`magic:${token}`, value, { ex: ttlSeconds })
  } else if (IS_DEV) {
    devStore.set(`magic:${token}`, {
      value,
      expiresAt: Date.now() + ttlSeconds * 1000,
    })
  } else {
    throw new Error('Token store not configured. Set KV_URL/KV_REST_API_URL or add @vercel/kv.')
  }
}

// fetchMagicToken looks up a token, validates TTL, and deletes
// it so the token is single-use. Returns null if missing/expired.
export async function fetchMagicToken(
  token: string
): Promise<string | null> {
  const key = `magic:${token}`
  let raw: string | null = null

  if (kv) {
    const v = await kv.get(key)
    raw = v?.value ?? null
  } else if (IS_DEV) {
    const v = devStore.get(key)
    if (v && v.expiresAt > Date.now()) {
      raw = v.value
    }
  }

  if (raw === null) return null

  // Delete to enforce single-use.
  if (kv) {
    try {
      await kv.del(key)
    } catch {
      // TTL will expire it eventually.
    }
  } else if (IS_DEV) {
    devStore.delete(key)
  }

  return raw
}
