// Type declarations for optional runtime dependencies.
//
// @vercel/kv and resend are NOT installed in this monorepo
// (they're only available in the Vercel deployment
// environment). The API routes use dynamic import() so the
// dev environment works without them, but TypeScript still
// tries to resolve the types. These declarations tell the
// type checker "trust me, these are valid imports" without
// actually pulling in the packages.

declare module '@vercel/kv' {
  export interface KV {
    set(key: string, value: string, opts?: { ex?: number }): Promise<unknown>
    get(key: string): Promise<{ value: string } | null>
    del(key: string): Promise<unknown>
  }
  export const kv: KV
}

declare module 'resend' {
  export interface EmailRequest {
    from: string
    to: string
    subject: string
    html: string
  }
  export class Resend {
    constructor(apiKey: string)
    emails: {
      send(req: EmailRequest): Promise<unknown>
    }
  }
}
