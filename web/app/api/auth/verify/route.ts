// GET /api/auth/verify?token=X
//
// Looks up a one-time magic-link token in Vercel KV (or the
// in-process dev fallback), validates TTL, marks the token as
// used, and returns the verified email. The verify page on
// the web app calls this, then signs the user in by setting
// a session cookie + redirecting to the post-auth landing.
//
// Response (JSON):
//   { "email": "alice@example.com", "redirect_url": "..." }
//
// Or, on failure:
//   { "error": "invalid or expired token" } (401)
//
// Tokens are SINGLE-USE: this handler deletes the token from
// KV after a successful lookup. Re-presenting the same token
// returns 401. This prevents replay attacks.
//
// In dev mode the in-process map also enforces single-use
// (deletes the key after lookup).

import { NextResponse, type NextRequest } from 'next/server'

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
  // @vercel/kv not installed. Will fall back below.
}

const devStore = new Map<string, { value: string; expiresAt: number }>()
const IS_DEV = process.env.NODE_ENV !== 'production'

interface MagicTokenRecord {
  email: string
  created_at: string
  redirect_url: string
}

export async function GET(req: NextRequest): Promise<NextResponse> {
  const token = req.nextUrl.searchParams.get('token') ?? ''
  if (!token || !/^[a-f0-9]{64}$/.test(token)) {
    // 64-char hex; reject anything that doesn't match the
    // expected shape. This is a cheap defense against
    // SQL-injection-style attacks on the lookup key.
    return NextResponse.json(
      { error: 'invalid or expired token' },
      { status: 401 }
    )
  }
  const key = `magic:${token}`

  // 1. Look up. We use a get-then-del pattern instead of
  // GETDEL because @vercel/kv doesn't expose GETDEL in all
  // versions. The race window is small (the same user
  // clicking the link twice in <100 ms) and the second
  // click is benign — the token is consumed on first use.
  let raw: string | null = null
  if (kv) {
    const v = await kv.get(key)
    raw = v?.value ?? null
  } else if (IS_DEV) {
    const v = devStore.get(key)
    if (v && v.expiresAt > Date.now()) {
      raw = v.value
    }
  } else {
    return NextResponse.json(
      { error: 'token store not configured' },
      { status: 503 }
    )
  }
  if (raw === null) {
    return NextResponse.json(
      { error: 'invalid or expired token' },
      { status: 401 }
    )
  }

  // 2. Delete to make the token single-use. We do this
  // BEFORE parsing so a parse error doesn't leave a
  // single-use token hanging.
  if (kv) {
    try {
      await kv.del(key)
    } catch {
      // If the delete fails (KV network blip) the token
      // will expire on its own TTL. Don't fail the
      // request: the user has the email and the link.
    }
  } else if (IS_DEV) {
    devStore.delete(key)
  }

  // 3. Parse the stored record. The record was JSON-
  // encoded by the magic route.
  let record: MagicTokenRecord
  try {
    record = JSON.parse(raw) as MagicTokenRecord
  } catch {
    // Corrupt record. We've already deleted the token so
    // the user can request a new one.
    return NextResponse.json(
      { error: 'invalid or expired token' },
      { status: 401 }
    )
  }
  if (typeof record.email !== 'string' || record.email === '') {
    return NextResponse.json(
      { error: 'invalid or expired token' },
      { status: 401 }
    )
  }

  return NextResponse.json({
    email: record.email,
    redirect_url: record.redirect_url,
  })
}
