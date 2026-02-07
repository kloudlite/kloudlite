import NextAuth from 'next-auth'
import Google from 'next-auth/providers/google'
import GitHub from 'next-auth/providers/github'
import MicrosoftEntraId from 'next-auth/providers/microsoft-entra-id'
import Credentials from 'next-auth/providers/credentials'
import type { NextAuthConfig, NextAuthResult } from 'next-auth'
import { authenticateUser } from '@/app/actions/user-auth.actions'
import { getOAuthConfig } from '@/lib/oauth-config'
import bcrypt from 'bcryptjs'
import { createHmac } from 'crypto'

interface SuperAdminTokenPayload {
  type: 'superadmin-login'
  installationId: string
  installationKey: string
  timestamp: number
  nonce: string
  expiresAt: number
}

function buildAuthConfig(): NextAuthConfig {
  const oauthConfig = getOAuthConfig()

  // Build OAuth providers conditionally based on config
  const oauthProviders = []

  if (oauthConfig.google.clientId && oauthConfig.google.clientSecret) {
    oauthProviders.push(
      Google({
        clientId: oauthConfig.google.clientId,
        clientSecret: oauthConfig.google.clientSecret,
      })
    )
  }

  if (oauthConfig.github.clientId && oauthConfig.github.clientSecret) {
    oauthProviders.push(
      GitHub({
        clientId: oauthConfig.github.clientId,
        clientSecret: oauthConfig.github.clientSecret,
      })
    )
  }

  if (oauthConfig.microsoft.clientId && oauthConfig.microsoft.clientSecret) {
    oauthProviders.push(
      MicrosoftEntraId({
        clientId: oauthConfig.microsoft.clientId,
        clientSecret: oauthConfig.microsoft.clientSecret,
        issuer: `https://login.microsoftonline.com/${oauthConfig.microsoft.tenantId || 'common'}/v2.0`,
      })
    )
  }

  return {
    trustHost: true,
    providers: [
      Credentials({
        name: 'credentials',
        credentials: {
          email: { label: 'Email', type: 'email' },
          password: { label: 'Password', type: 'password' },
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

          // Handle normal email/password login
          if (!credentials?.email || !credentials?.password) {
            return null
          }

          try {
            // Dynamic import K8s client (only loads in Node.js/Bun runtime, not Edge Runtime)
            const { userRepository } = await import('@kloudlite/lib/k8s')

            // Get user from Kubernetes by email
            let user
            try {
              user = await userRepository.getByEmail(credentials.email as string)
            } catch (error) {
              console.error('User lookup failed for:', credentials.email, error)
              return null
            }

            // Check if user is active
            if (!user.spec.active) {
              console.error('User is not active:', credentials.email)
              return null
            }

            // Verify password hash
            const passwordHash = user.spec.password
            if (!passwordHash) {
              console.error('User has no password set:', credentials.email)
              return null
            }

            // Decode base64 password hash
            const passwordHashBuffer = Buffer.from(passwordHash, 'base64')
            const passwordHashString = passwordHashBuffer.toString('utf-8')

            // Compare password with bcrypt hash
            const isPasswordValid = await bcrypt.compare(
              credentials.password as string,
              passwordHashString
            )

            if (!isPasswordValid) {
              return null
            }

            // Return user object - NextAuth will generate JWT
            return {
              id: user.spec.email,
              email: user.spec.email,
              name: user.spec.displayName || user.metadata!.name || user.spec.email,
              username: user.metadata!.name!,
              roles: user.spec.roles || ['user'],
              isActive: user.spec.active,
            }
          } catch (error) {
            console.error('Login error:', error)
            return null
          }
        },
      }),
      ...oauthProviders,
    ],
    pages: {
      signIn: '/auth/signin',
      error: '/auth/error',
    },
    callbacks: {
      async jwt({ token, user, account }) {
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

          // For OAuth providers, fetch user info from K8s
          if (account && account.provider !== 'credentials' && user.email) {
            token.provider = account.provider
            token.providerId = account.providerAccountId

            try {
              const { userRepository } = await import('@kloudlite/lib/k8s')
              const k8sUser = await userRepository.getByEmail(user.email)
              token.username = k8sUser.metadata?.name || user.email
              token.roles = k8sUser.spec?.roles || ['user']
              token.isActive = k8sUser.spec?.active ?? true
            } catch (error) {
              console.error('Failed to get user from K8s:', error)
            }
          }

          // Fetch and cache work machine namespace (only on initial login, skip for superadmin)
          const username = token.username as string
          if (username && !token.namespace && token.provider !== 'superadmin-login') {
            try {
              const { workMachineRepository } = await import('@kloudlite/lib/k8s')
              const workMachine = await workMachineRepository.getByOwner(username)
              if (workMachine) {
                token.namespace = workMachine.spec?.targetNamespace || 'default'
                token.workMachineName = workMachine.metadata?.name
              }
            } catch (error) {
              console.error('Failed to fetch work machine for session:', error)
            }
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
      async signIn({ user, account }) {
        // For credentials provider, we've already authenticated in authorize()
        if (account?.provider === 'credentials') {
          return true
        }

        // For OAuth providers, check if user exists in backend
        if (!user.email || !account) {
          console.error('Sign-in failed: Missing email or account information')
          return false
        }

        console.log(`Sign-in attempt: ${user.email} via ${account.provider}`)

        const result = await authenticateUser({
          email: user.email,
          name: user.name,
          image: user.image,
          provider: account.provider,
          providerId: account.providerAccountId,
        })

        // Only allow sign-in if user exists in backend
        if (!result.success) {
          console.warn(`Sign-in blocked: ${user.email} - ${result.error}`)
          // Redirect to error page with specific error
          return `/auth/error?error=AccessDenied&message=${encodeURIComponent(result.error || 'User not found')}`
        }

        console.log(`Sign-in successful: ${user.email} via ${account.provider}`)
        return true
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
 * Call after saving OAuth config so the next request rebuilds with fresh providers.
 */
export function invalidateAuth() {
  globalThis.__nextAuthInstance = undefined
  console.log('[AUTH] NextAuth instance invalidated — will rebuild on next request')
}

// Stable proxy exports that delegate to the lazy singleton
export const handlers = {
  GET: (req: any) => getNextAuth().handlers.GET(req),
  POST: (req: any) => getNextAuth().handlers.POST(req),
}
export const auth: NextAuthResult['auth'] = ((...args: any[]) =>
  (getNextAuth().auth as Function)(...args)) as any
export const signIn: NextAuthResult['signIn'] = ((...args: any[]) =>
  (getNextAuth().signIn as Function)(...args)) as any
export const signOut: NextAuthResult['signOut'] = ((...args: any[]) =>
  (getNextAuth().signOut as Function)(...args)) as any

// Export default for middleware (Edge Runtime compatible)
const authMiddleware: NextAuthResult['auth'] = ((...args: any[]) =>
  (getNextAuth().auth as Function)(...args)) as any
export default authMiddleware
