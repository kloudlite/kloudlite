import NextAuth from 'next-auth'
import Google from 'next-auth/providers/google'
import GitHub from 'next-auth/providers/github'
import MicrosoftEntraId from 'next-auth/providers/microsoft-entra-id'
import Credentials from 'next-auth/providers/credentials'
import type { NextAuthConfig } from 'next-auth'
import { authenticateUser } from '@/lib/actions/user-actions'
import { unauthenticatedApiClient } from '@/lib/api-client'
import { env } from '@kloudlite/lib'

interface LoginResponse {
  user: {
    username: string
    email: string
    name: string
    displayName?: string
    isActive: boolean
  }
  roles: string[]
}

interface UserInfoResponse {
  user: {
    username: string
    email: string
    name: string
    displayName?: string
    isActive: boolean
  }
  roles: string[]
}

interface SuperAdminValidateResponse {
  valid: boolean
  user: {
    email: string
    displayName?: string
    name?: string
  }
  roles: string[]
}

export const authConfig: NextAuthConfig = {
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
            const response = await fetch(`${env.apiUrl}/api/v1/superadmin-login/validate`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ token: credentials.superadminToken }),
            })

            if (!response.ok) {
              console.error('Super-admin token validation failed:', response.status)
              return null
            }

            const data = (await response.json()) as SuperAdminValidateResponse

            if (!data.valid || !data.user) {
              return null
            }

            // Return user object with super-admin provider marker
            return {
              id: data.user.email,
              email: data.user.email,
              name: data.user.displayName || data.user.name || data.user.email,
              username: data.user.email, // Use email as username for super-admin
              roles: data.roles,
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
          // Call backend API to validate credentials and get user info
          const response = (await unauthenticatedApiClient.post('/api/v1/auth/login', {
            email: credentials.email,
            password: credentials.password,
          })) as LoginResponse

          if (response.user) {
            // Return user object - NextAuth will generate JWT with shared secret
            return {
              id: response.user.email,
              email: response.user.email,
              name: response.user.name || response.user.displayName || response.user.email,
              username: response.user.username,
              roles: response.roles,
              isActive: response.user.isActive,
            }
          }
          return null
        } catch (error) {
          console.error('Login error:', error)
          return null
        }
      },
    }),
    Google({
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
    }),
    GitHub({
      clientId: process.env.GITHUB_CLIENT_ID!,
      clientSecret: process.env.GITHUB_CLIENT_SECRET!,
    }),
    MicrosoftEntraId({
      clientId: process.env.MICROSOFT_ENTRA_CLIENT_ID!,
      clientSecret: process.env.MICROSOFT_ENTRA_CLIENT_SECRET!,
      issuer: `https://login.microsoftonline.com/${process.env.MICROSOFT_ENTRA_TENANT_ID}/v2.0`,
    }),
  ],
  pages: {
    signIn: '/auth/signin',
    error: '/auth/error',
  },
  callbacks: {
    async jwt({ token, user, account }) {
      if (user) {
        // Store user info in JWT (NextAuth will sign with JWT_SECRET)
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

        // For OAuth providers, fetch user info from backend
        if (account && account.provider !== 'credentials' && user.email) {
          token.provider = account.provider
          token.providerId = account.providerAccountId

          try {
            const response = (await unauthenticatedApiClient.post('/api/v1/auth/user-info', {
              email: user.email,
            })) as UserInfoResponse

            if (response.user) {
              token.username = response.user.username
              token.roles = response.roles
              token.isActive = response.user.isActive
            }
          } catch (error) {
            console.error('Failed to get user info for OAuth user:', error)
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
    maxAge: 24 * 60 * 60, // 24 hours (match backend token expiry)
  },
  // Configure cookies with domain for cross-subdomain sharing (needed for VPN check)
  // Domain should be set to the tenant subdomain (e.g., beanbag.khost.dev) to allow
  // cookies to be sent to subdomains like vpn-check.beanbag.khost.dev
  cookies: process.env.NEXT_PUBLIC_AUTH_COOKIE_DOMAIN
    ? {
        sessionToken: {
          name:
            process.env.NODE_ENV === 'production'
              ? '__Secure-authjs.session-token'
              : 'authjs.session-token',
          options: {
            httpOnly: true,
            sameSite: 'lax',
            path: '/',
            secure: process.env.NODE_ENV === 'production',
            domain: process.env.NEXT_PUBLIC_AUTH_COOKIE_DOMAIN,
          },
        },
      }
    : undefined,
  secret: process.env.JWT_SECRET, // Shared secret for JWT signing/verification (same as backend)
}

export const { handlers, signIn, signOut, auth } = NextAuth(authConfig)
