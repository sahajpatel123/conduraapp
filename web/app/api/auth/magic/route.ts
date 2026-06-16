// POST /api/auth/magic
//
// Validates the user's email, generates a one-time magic-link
// token, stores it in Vercel KV with a 5-minute TTL, and sends
// the link via Resend. The token is the single-use proof that
// the user controls the email address; it expires on first
// use or after 5 minutes, whichever comes first.
//
// Request body (JSON):
//   { "email": "alice@example.com",
//     "redirect_url": "https://synaptic.app/auth/callback" }
//
// Response (JSON):
//   { "sent": true, "expires_in": 300 }
//
// In dev mode (when RESEND_API_KEY is unset), the token is
// returned in dev_token so the user can paste it into the
// browser manually. Production builds ALWAYS set RESEND_API_KEY
// and so dev_token is always empty.
//
// KV key format: "magic:<token>". Value: JSON-encoded
// { email, created_at, redirect_url }. The TTL is enforced
// by Vercel KV, not by us; if KV is down, the route fails
// closed (5xx) rather than issuing a never-expiring token.

import { NextResponse, type NextRequest } from 'next/server'
import { randomBytes } from 'node:crypto'

// Vercel KV is a global at runtime in production. We use a
// dynamic import so the dev environment doesn't require
// @vercel/kv to be installed. If @vercel/kv isn't available
// (dev mode without KV), we fall back to an in-process map
// for local testing. The fallback is gated behind NODE_ENV
// !== 'production' so production builds never run with it.
//
// We declare the type as `any` here because we don't want
// this file to fail to typecheck when @vercel/kv is absent.
// The runtime call is gated by try/catch.
type KV = {
  set: (key: string, value: string, opts?: { ex?: number }) => Promise<unknown>
  get: (key: string) => Promise<{ value: string } | null>
  del: (key: string) => Promise<unknown>
}
let kv: KV | null = null
try {
  // Dynamic import so dev environments without @vercel/kv
  // still work.
  const mod = (await import('@vercel/kv')) as { kv: KV }
  kv = mod.kv
} catch {
  // @vercel/kv not installed. Will fall back below.
}

// In-process fallback for dev mode. Maps token → value
// JSON. Expiry is checked at read time.
const devStore = new Map<string, { value: string; expiresAt: number }>()

const TTL_SECONDS = 5 * 60

// Email validation: not perfect, but enough to catch typos.
// Server-side validation matches the client regex.
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

// In dev mode only: return the token in the response so the
// user can paste it into the browser to complete the flow.
const IS_DEV = process.env.NODE_ENV !== 'production'

export async function POST(req: NextRequest): Promise<NextResponse> {
  // 1. Parse + validate the body.
  let body: unknown
  try {
    body = await req.json()
  } catch {
    return NextResponse.json(
      { error: 'invalid JSON body' },
      { status: 400 }
    )
  }
  if (typeof body !== 'object' || body === null) {
    return NextResponse.json(
      { error: 'body must be a JSON object' },
      { status: 400 }
    )
  }
  const { email, redirect_url } = body as {
    email?: unknown
    redirect_url?: unknown
  }
  if (typeof email !== 'string' || !EMAIL_RE.test(email)) {
    return NextResponse.json(
      { error: 'invalid email' },
      { status: 400 }
    )
  }
  if (typeof redirect_url !== 'string' || !redirect_url.startsWith('https://')) {
    // Only https:// redirects are accepted to prevent
    // open-redirect attacks. The desktop app uses
    // synaptic://oauth-callback which is handled
    // separately by the daemon's account.magic_link RPC.
    return NextResponse.json(
      { error: 'redirect_url must be an https:// URL' },
      { status: 400 }
    )
  }

  // 2. Generate a 32-byte (256-bit) random token, hex
  // encoded. That's 64 characters; safe for URL use and
  // impossible to brute-force.
  const token = randomBytes(32).toString('hex')

  // 3. Persist the token. We use Vercel KV in production
  // and the in-process map in dev.
  const value = JSON.stringify({
    email,
    created_at: new Date().toISOString(),
    redirect_url,
  })
  if (kv) {
    try {
      await kv.set(`magic:${token}`, value, { ex: TTL_SECONDS })
    } catch (e) {
      // KV failure: fail closed. We never issue a token
      // we can't persist.
      return NextResponse.json(
        { error: 'token store unavailable' },
        { status: 503 }
      )
    }
  } else if (IS_DEV) {
    devStore.set(`magic:${token}`, {
      value,
      expiresAt: Date.now() + TTL_SECONDS * 1000,
    })
  } else {
    // Production with no KV is a configuration error.
    return NextResponse.json(
      { error: 'token store not configured' },
      { status: 503 }
    )
  }

  // 4. Send the email. The Resend SDK is dynamically
  // imported so dev environments without it still work.
  const link = `${redirect_url}?token=${encodeURIComponent(token)}`
  let sent = false
  try {
    const resendKey = process.env.RESEND_API_KEY
    if (resendKey) {
      const mod = (await import('resend')) as unknown as {
        Resend: new (key: string) => {
          emails: { send: (req: unknown) => Promise<unknown> }
        }
      }
      const resend = new mod.Resend(resendKey)
      await resend.emails.send({
        from: 'Synaptic <noreply@synaptic.app>',
        to: email,
        subject: 'Sign in to Synaptic',
        html:
          `<p>Click the link below to sign in to Synaptic. The link expires in 5 minutes.</p>` +
          `<p><a href="${link}">Sign in</a></p>` +
          `<p>If you didn't request this, you can ignore the email.</p>`,
      })
      sent = true
    } else if (IS_DEV) {
      // Dev mode without Resend: log the link to stdout.
      // The user can copy it from the dev server console.
      console.log(`[magic-link] ${email} → ${link}`)
      sent = true
    }
  } catch (e) {
    // Email dispatch failed: the token is still in KV
    // (or dev map) and will expire on its own. Return
    // 503 so the GUI can retry.
    return NextResponse.json(
      { error: 'email dispatch failed' },
      { status: 503 }
    )
  }

  return NextResponse.json({
    sent,
    expires_in: TTL_SECONDS,
    // Dev mode: include the token so the developer can
    // paste the URL into a browser without setting up
    // Resend. Production: always empty.
    dev_token: IS_DEV && !process.env.RESEND_API_KEY ? token : '',
  })
}
