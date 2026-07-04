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
import { fetchMagicToken } from '@/lib/kv'

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
  // 1. Look up. fetchMagicToken validates TTL and deletes
  // the token so it is single-use.
  let raw: string | null
  try {
    raw = await fetchMagicToken(token)
  } catch (e) {
    const message = e instanceof Error ? e.message : String(e)
    if (message === 'token store not configured') {
      return NextResponse.json(
        { error: 'token store not configured' },
        { status: 503 }
      )
    }
    raw = null
  }
  if (raw === null) {
    return NextResponse.json(
      { error: 'invalid or expired token' },
      { status: 401 }
    )
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
