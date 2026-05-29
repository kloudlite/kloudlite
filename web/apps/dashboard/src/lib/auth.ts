import NextAuth from 'next-auth'
import Credentials from 'next-auth/providers/credentials'
import type { NextAuthConfig, NextAuthResult } from 'next-auth'
import { createHmac } from 'crypto'
import type { NextRequest } from 'next/server'

interface SuperAdminTokenPayload {
  type: 'superadmin-login'
  installationId: string
  installationKey: string
  timestamp: number
  nonce: string
  expiresAt: number
}

function buildAuthConfig(): NextAuthConfig {
  return {
    trustHost: true,
    providers: [
      Credentials({
        name: 'super-admin-token',
        credentials: {
          superadminToken: { label: 'Super Admin Token', type: 'text' },
        },
        async authorize(credentials) {
          // Handle super-admin token login
          if (credentials?.superadminToken) {
            try {
              const token = credentials.superadminToken as string

              // Dev/test backdoor: skip HMAC validation for local testing and e2e tests
              if (process.env.ALLOW_DEV_SUPERADMIN === 'true' && token === 'dev-superadmin') {
                console.warn('[AUTH] Dev mode super-admin login bypass')
                return {
                  id: 'dev-installation',
                  email: 'admin@dev-installation',
                  name: 'Super Admin (Dev)',
                  username: 'superadmin',
                  roles: ['admin', 'super-admin'],
                  isActive: true,
                  provider: 'superadmin-login',
                }
              }

              const installationSecret = process.env.INSTALLATION_SECRET

              if (!installationSecret) {
                console.error('INSTALLATION_SECRET env var is not set')
                return null
              }

              // Parse token: base64url(payload).base64url(signature)
              const parts = token.split('.')
              if (parts.length !== 2) {
                console.error('Invalid super-admin token format')
                return null
              }

              const [payloadB64, signatureB64] = parts
              const payloadString = Buffer.from(payloadB64, 'base64url').toString('utf-8')

              // Verify HMAC signature
              const expectedSignature = createHmac('sha256', installationSecret)
                .update(payloadString)
                .digest('base64url')

              if (expectedSignature !== signatureB64) {
                console.error('Super-admin token signature verification failed')
                return null
              }

              // Parse and validate payload
              const payload = JSON.parse(payloadString) as SuperAdminTokenPayload

              if (payload.type !== 'superadmin-login') {
                console.error('Invalid super-admin token type:', payload.type)
                return null
              }

              if (payload.expiresAt < Date.now()) {
                console.error('Super-admin token has expired')
                return null
              }

              // Return user object with super-admin provider marker
              return {
                id: payload.installationId,
                email: `admin@${payload.installationId}`,
                name: 'Super Admin',
                username: 'superadmin',
                roles: ['admin', 'super-admin'],
                isActive: true,
                provider: 'superadmin-login',
              }
            } catch (error) {
              console.error('Super-admin login error:', error)
              return null
            }
          }

          return null
        },
      }),
    ],
    pages: {
      signIn: '/auth/signin',
      error: '/auth/error',
    },
    callbacks: {
      async jwt({ token, user }) {
        if (user) {
          // Store user info in JWT
          if ('username' in user) {
            token.username = user.username
          }
          if ('roles' in user) {
            token.roles = user.roles
          }
          if ('isActive' in user) {
            token.isActive = user.isActive
          }
          // Handle super-admin provider from credentials
          if ('provider' in user && user.provider === 'superadmin-login') {
            token.provider = 'superadmin-login'
          }
        }
        return token
      },
      async session({ session, token }) {
        if (session.user) {
          session.user.id = token.sub!
          session.user.username = token.username as string
          if (token.provider) {
            session.user.provider = token.provider as string
          }
          if (token.roles) {
            session.user.roles = token.roles as string[]
          }
          if (token.isActive !== undefined) {
            session.user.isActive = token.isActive as boolean
          }
          // Add cached work machine data
          if (token.namespace) {
            session.user.namespace = token.namespace as string
          }
          if (token.workMachineName) {
            session.user.workMachineName = token.workMachineName as string
          }
        }
        return session
      },
      async signIn({ account }) {
        return account?.provider === 'credentials'
      },
      async redirect({ url, baseUrl }) {
        // Allows relative callback URLs
        if (url.startsWith('/')) return `${baseUrl}${url}`
        // Allows callback URLs on the same origin
        else if (new URL(url).origin === baseUrl) return url
        // Default redirect to homepage after login
        return baseUrl
      },
    },
    session: {
      strategy: 'jwt',
      maxAge: 24 * 60 * 60, // 24 hours
    },
  }
}

// Lazy NextAuth singleton — survives HMR via globalThis
declare global {
  var __nextAuthInstance: NextAuthResult | undefined
}

function getNextAuth(): NextAuthResult {
  if (!globalThis.__nextAuthInstance) {
    globalThis.__nextAuthInstance = NextAuth(buildAuthConfig())
  }
  return globalThis.__nextAuthInstance
}

/**
 * Invalidate the cached NextAuth instance.
 * Call after changing auth-related runtime configuration.
 */
export function invalidateAuth() {
  globalThis.__nextAuthInstance = undefined
  console.log('[AUTH] NextAuth instance invalidated — will rebuild on next request')
}

// Stable proxy exports that delegate to the lazy singleton
export const handlers = {
  GET: (req: NextRequest) => getNextAuth().handlers.GET(req),
  POST: (req: NextRequest) => getNextAuth().handlers.POST(req),
}

type AuthFn = NextAuthResult['auth']
type SignInFn = NextAuthResult['signIn']
type SignOutFn = NextAuthResult['signOut']

export const auth: AuthFn = ((...args: Parameters<AuthFn>) => {
  const authImpl = getNextAuth().auth as AuthFn
  return authImpl(...args)
}) as AuthFn

export const signIn: SignInFn = ((...args: Parameters<SignInFn>) => {
  const signInImpl = getNextAuth().signIn as SignInFn
  return signInImpl(...args)
}) as SignInFn

export const signOut: SignOutFn = ((...args: Parameters<SignOutFn>) => {
  const signOutImpl = getNextAuth().signOut as SignOutFn
  return signOutImpl(...args)
}) as SignOutFn

// Export default for middleware (Edge Runtime compatible)
const authMiddleware: AuthFn = ((...args: Parameters<AuthFn>) => {
  const authImpl = getNextAuth().auth as AuthFn
  return authImpl(...args)
}) as AuthFn
export default authMiddleware
