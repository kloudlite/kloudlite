import NextAuth from 'next-auth'
import Google from 'next-auth/providers/google'
import GitHub from 'next-auth/providers/github'
import MicrosoftEntraId from 'next-auth/providers/microsoft-entra-id'
import Credentials from 'next-auth/providers/credentials'
import type { NextAuthConfig } from 'next-auth'
import { authenticateUser } from '@/lib/actions/user-actions'
import { apiClient } from '@/lib/api-client'

export const authConfig: NextAuthConfig = {
  providers: [
    Credentials({
      name: 'credentials',
      credentials: {
        email: { label: 'Email', type: 'email' },
        password: { label: 'Password', type: 'password' }
      },
      async authorize(credentials) {
        if (!credentials?.email || !credentials?.password) {
          return null
        }

        try {
          // Call backend API to authenticate
          const response = await apiClient.post('/api/v1/auth/login', {
            email: credentials.email,
            password: credentials.password
          })

          if (response.token) {
            // Return user object that will be encoded in JWT
            return {
              id: response.email,
              email: response.email,
              name: response.email,
              roles: response.roles
            }
          }
          return null
        } catch (error) {
          console.error('Login error:', error)
          return null
        }
      }
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
      tenantId: process.env.MICROSOFT_ENTRA_TENANT_ID!,
    }),
  ],
  pages: {
    signIn: '/auth/signin',
    error: '/auth/error',
  },
  callbacks: {
    async jwt({ token, user, account }) {
      if (user) {
        // For credentials provider, user will have roles
        if ('roles' in user) {
          token.roles = user.roles
        }
        // For OAuth providers
        if (account) {
          token.provider = account.provider
          token.providerId = account.providerAccountId
        }
      }
      return token
    },
    async session({ session, token }) {
      if (session.user) {
        session.user.id = token.sub!
        if (token.provider) {
          session.user.provider = token.provider as string
        }
        if (token.roles) {
          session.user.roles = token.roles as string[]
        }
      }
      return session
    },
    async signIn({ user, account, profile }) {
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
  },
  session: {
    strategy: 'jwt',
  },
  secret: process.env.NEXTAUTH_SECRET,
}

export const { handlers, signIn, signOut, auth } = NextAuth(authConfig)