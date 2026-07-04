// POST /api/auth/magic
//
// Validates the user's email, generates a one-time magic-link
// token, stores it in Upstash Redis with a 5-minute TTL, and sends
// the link via Resend. The token is the single-use proof that
// the user controls the email address; it expires on first
// use or after 5 minutes, whichever comes first.
//
// Request body (JSON):
//   { "email": "alice@example.com",
//     "redirect_url": "https://condura.app/auth/callback" }
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
// by Upstash Redis, not by us; if Redis is down, the route fails
// closed (5xx) rather than issuing a never-expiring token.

import { NextResponse, type NextRequest } from 'next/server'
import {
  generateMagicToken,
  storeMagicToken,
  IS_DEV,
} from '@/lib/kv'

const TTL_SECONDS = 5 * 60

// Email validation: not perfect, but enough to catch typos.
// Server-side validation matches the client regex.
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

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
  if (typeof redirect_url !== 'string') {
    return NextResponse.json(
      { error: 'redirect_url is required' },
      { status: 400 }
    )
  }
  // Phase 17, Fix #8 (R3): the previous check only verified
  // the scheme (https://). That allowed https://evil.com/...
  // which an attacker can phish with — they put the link in
  // an email "sign in to condura" and the user lands on a
  // convincing clone. We now require the host to be on our
  // allowlist (condura.app or localhost for dev).
  //
  // We parse the URL with the WHATWG URL API. Parsing failure
  // OR a non-https protocol OR a host that's not in the
  // allowlist all reject with 400.
  let parsedHost = ''
  try {
    const u = new URL(redirect_url)
    if (u.protocol !== 'https:') {
      return NextResponse.json(
        { error: 'redirect_url must use https://' },
        { status: 400 }
      )
    }
    parsedHost = u.hostname.toLowerCase()
  } catch {
    return NextResponse.json(
      { error: 'redirect_url is not a valid URL' },
      { status: 400 }
    )
  }
  const allowedHosts = new Set([
    'condura.app',
    'www.condura.app',
    'localhost',
    '127.0.0.1',
  ])
  if (!allowedHosts.has(parsedHost)) {
    return NextResponse.json(
      { error: 'redirect_url host is not allowed' },
      { status: 400 }
    )
  }

  // 2. Generate a 32-byte (256-bit) random token, hex
  // encoded. That's 64 characters; safe for URL use and
  // impossible to brute-force.
  const token = generateMagicToken()

  // 3. Persist the token. We use Vercel KV in production
  // and the in-process map in dev.
  const value = JSON.stringify({
    email,
    created_at: new Date().toISOString(),
    redirect_url,
  })
  try {
    await storeMagicToken(token, value, TTL_SECONDS)
  } catch (e) {
    const message = e instanceof Error ? e.message : String(e)
    if (message === 'token store not configured') {
      // Production with no KV is a configuration error.
      return NextResponse.json(
        { error: 'token store not configured' },
        { status: 503 }
      )
    }
    // KV failure: fail closed. We never issue a token
    // we can't persist.
    return NextResponse.json(
      { error: 'token store unavailable' },
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
        from: 'Condura <noreply@condura.app>',
        to: email,
        subject: 'Sign in to Condura',
        html:
          `<p>Click the link below to sign in to Condura. The link expires in 5 minutes.</p>` +
          `<p><a href="${link}">Sign in</a></p>` +
          `<p>If you didn't request this, you can ignore the email.</p>`,
      })
      sent = true
    } else if (IS_DEV) {
      // Dev mode without Resend: log the link to stdout.
      // The user can copy it from the dev server console.
      console.log(`[magic-link] ${email} → ${link}`)
      sent = true
    } else {
      // Production without Resend: email cannot be sent.
      // Return 503 so the GUI can show a proper error.
      return NextResponse.json(
        { error: 'email provider not configured' },
        { status: 503 }
      )
    }
  } catch {
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
