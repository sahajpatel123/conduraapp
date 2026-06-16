// Account store (Phase 14B).
//
// The user's authentication state. On the desktop, the
// daemon is the source of truth (it stores the user's
// tokens in the OS keychain). On the web, the AccountStore
// reads from a cookie or the local storage and falls back
// to the daemon.
//
// Sign-in flows:
//   1. Google / GitHub / Apple OAuth:
//      - call accountOAuthURL(provider) → open returned URL
//      - user consents at the provider
//      - provider redirects to synaptic://oauth-callback?code=...&state=...
//      - the GUI's handleCallback(code, state) calls
//        accountOAuthCallback() and stores the result.
//   2. Magic link (no third-party OAuth):
//      - user enters email; call accountMagicLink({ email })
//      - daemon emails a one-time link (or returns dev_token
//        in dev mode)
//      - user clicks the link; lands on a page that calls
//        /api/auth/verify (web) or accountMagicLink again
//        with the token (desktop)
//   3. Sign out: accountLogout clears the daemon's user
//      record and the cached state.
//
// The store is a thin cache of the daemon's response. The
// GUI surfaces isSignedIn / email / provider / avatarURL
// to the topbar (sidebar or HUD).

import { ipc } from '../ipc/client'
import type {
  AccountStatus,
  OAuthURLParams,
  OAuthURLResult,
  OAuthCallbackParams,
  MagicLinkParams,
  MagicLinkResult,
} from '../ipc/types'

// emailLocalPart returns the substring of an email before
// the '@' (or the whole string if no '@' is present).
// Used as a fallback display name when the OAuth provider
// didn't return one.
function emailLocalPart(email: string): string {
  const at = email.lastIndexOf('@')
  if (at < 0) return email
  return email.slice(0, at)
}

/**
 * AccountStore: the signed-in user's state plus auth
 * orchestration methods. Components read $state fields
 * (isSignedIn, email, etc.) and re-render automatically.
 *
 * Lifecycle:
 *   1. App mounts → store starts in the "unknown" state.
 *   2. App calls account.checkStatus() to determine the
 *      current state.
 *   3. On success, status is populated; isSignedIn
 *      returns the right value.
 *   4. Sign-in / sign-out methods mutate status.
 */
export class AccountStore {
  /** Current auth state. Null before checkStatus() runs. */
  status = $state<AccountStatus | null>(null)

  /** True while a check / sign-in / sign-out RPC is in flight. */
  loading = $state<boolean>(false)

  /**
   * The most recent error from an auth RPC. Cleared on
   * the next call. The GUI surfaces this as an inline
   * error message.
   */
  error = $state<string | null>(null)

  /** Cached OAuth state token from a pending sign-in. */
  pendingOAuthState = $state<string | null>(null)

  /** Cached PKCE verifier from a pending sign-in. */
  pendingPKCEVerifier = $state<string | null>(null)

  /** True when the user is signed in. False on first load. */
  get isSignedIn(): boolean {
    return this.status?.signed_in ?? false
  }

  /** The user's email, or empty when signed out. */
  get email(): string {
    return this.status?.email ?? ''
  }

  /** The provider the user signed in with. */
  get provider(): AccountStatus['provider'] {
    return this.status?.provider ?? ''
  }

  /** The user's avatar URL, or empty when not available. */
  get avatarURL(): string {
    return this.status?.avatar_url ?? ''
  }

  /** The user's display name, or the email local-part as fallback. */
  get displayName(): string {
    if (!this.status) return ''
    if (this.status.display_name) return this.status.display_name
    // Fallback: local-part of email
    return emailLocalPart(this.status.email)
  }

  /** The user's subscription tier (free / pro / team / enterprise). */
  get tier(): AccountStatus['tier'] {
    return this.status?.tier ?? ''
  }

  /**
   * Fetches the current account status from the daemon.
   * Called on app start. Idempotent.
   */
  async checkStatus(): Promise<void> {
    this.loading = true
    this.error = null
    try {
      this.status = await ipc.accountStatus()
    } catch (e) {
      // Network/daemon error: assume signed out. The GUI
      // will offer the "Sign in" button.
      this.status = {
        signed_in: false,
        email: '',
        provider: '',
        avatar_url: '',
        display_name: '',
        tier: '',
        created_at: '',
      }
      this.error = String(e)
    } finally {
      this.loading = false
    }
  }

  /**
   * Starts an OAuth sign-in flow with the given provider.
   * Returns the authorization URL the GUI should open in
   * a browser. The state + PKCE verifier are cached so
   * handleCallback() can verify them.
   */
  async signInWithProvider(
    params: OAuthURLParams
  ): Promise<OAuthURLResult | null> {
    this.loading = true
    this.error = null
    try {
      const result = await ipc.accountOAuthURL(params)
      this.pendingOAuthState = result.state
      this.pendingPKCEVerifier = result.code_verifier
      return result
    } catch (e) {
      this.error = String(e)
      return null
    } finally {
      this.loading = false
    }
  }

  /**
   * Convenience: Google sign-in.
   */
  async signInWithGoogle(redirectURI: string): Promise<OAuthURLResult | null> {
    return this.signInWithProvider({
      provider: 'google',
      redirect_uri: redirectURI,
      scopes: ['openid', 'email', 'profile'],
    })
  }

  /**
   * Convenience: GitHub sign-in.
   */
  async signInWithGitHub(redirectURI: string): Promise<OAuthURLResult | null> {
    return this.signInWithProvider({
      provider: 'github',
      redirect_uri: redirectURI,
      scopes: ['read:user', 'user:email'],
    })
  }

  /**
   * Magic-link sign-in. The daemon sends (or in dev mode,
   * returns in dev_token) a one-time link to the user's
   * email. Returns the result so the GUI can show "link
   * sent, check your email".
   */
  async signInWithEmail(
    email: string,
    locale: string = 'en',
    redirectURL: string
  ): Promise<MagicLinkResult | null> {
    this.loading = true
    this.error = null
    try {
      return await ipc.accountMagicLink({
        email,
        locale,
        redirect_url: redirectURL,
      })
    } catch (e) {
      this.error = String(e)
      return null
    } finally {
      this.loading = false
    }
  }

  /**
   * Completes an OAuth flow. The user's browser has
   * returned to the GUI with ?code=...&state=... in the
   * URL. The GUI extracts those and calls this method.
   *
   * Verifies the state matches the one we cached (CSRF
   * defense) and exchanges the code for tokens.
   */
  async handleCallback(
    code: string,
    state: string,
    redirectURI: string,
    provider: 'google' | 'github' | 'apple'
  ): Promise<boolean> {
    this.loading = true
    this.error = null
    if (state !== this.pendingOAuthState) {
      this.error = 'OAuth state mismatch (possible CSRF)'
      this.loading = false
      return false
    }
    const codeVerifier = this.pendingPKCEVerifier
    if (!codeVerifier) {
      this.error = 'Missing PKCE verifier (lost between sign-in and callback?)'
      this.loading = false
      return false
    }
    try {
      const status = await ipc.accountOAuthCallback({
        provider,
        code,
        state,
        code_verifier: codeVerifier,
        redirect_uri: redirectURI,
      })
      this.status = status
      this.pendingOAuthState = null
      this.pendingPKCEVerifier = null
      return status.signed_in
    } catch (e) {
      this.error = String(e)
      return false
    } finally {
      this.loading = false
    }
  }

  /**
   * Signs the user out. Clears the cached status; the
   * daemon removes the local user record.
   */
  async signOut(): Promise<void> {
    this.loading = true
    this.error = null
    try {
      await ipc.accountLogout()
      this.status = {
        signed_in: false,
        email: '',
        provider: '',
        avatar_url: '',
        display_name: '',
        tier: '',
        created_at: '',
      }
      this.pendingOAuthState = null
      this.pendingPKCEVerifier = null
    } catch (e) {
      this.error = String(e)
    } finally {
      this.loading = false
    }
  }
}

// Singleton instance — only one auth flow at a time.
export const account = new AccountStore()
