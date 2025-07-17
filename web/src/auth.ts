import { NextAuthOptions } from 'next-auth'
import NextAuth from 'next-auth'
import Google from 'next-auth/providers/google'
import GitHub from 'next-auth/providers/github'
import AzureAD from 'next-auth/providers/azure-ad'
import { getServerSession } from 'next-auth'

export const authOptions: NextAuthOptions = {
  secret: process.env.NEXTAUTH_SECRET,
  debug: process.env.NODE_ENV === 'development',
  providers: [
    ...(process.env.GOOGLE_CLIENT_ID && process.env.GOOGLE_CLIENT_SECRET ? [
      Google({
        clientId: process.env.GOOGLE_CLIENT_ID,
        clientSecret: process.env.GOOGLE_CLIENT_SECRET,
      })
    ] : []),
    ...(process.env.GITHUB_CLIENT_ID && process.env.GITHUB_CLIENT_SECRET ? [
      GitHub({
        clientId: process.env.GITHUB_CLIENT_ID,
        clientSecret: process.env.GITHUB_CLIENT_SECRET,
      })
    ] : []),
    ...(process.env.MICROSOFT_CLIENT_ID && process.env.MICROSOFT_CLIENT_SECRET ? [
      AzureAD({
        clientId: process.env.MICROSOFT_CLIENT_ID,
        clientSecret: process.env.MICROSOFT_CLIENT_SECRET,
        tenantId: process.env.MICROSOFT_TENANT_ID || 'common',
        authorization: {
          params: {
            scope: 'openid profile email User.Read',
            prompt: 'select_account',
          },
        },
        httpOptions: {
          timeout: 10000,
        },
      })
    ] : []),
  ],
  callbacks: {
    async signIn({ user, account, profile }) {
      console.log('Sign in attempt:', { 
        provider: account?.provider,
        email: user?.email,
        name: user?.name,
        userId: user?.id
      })
      
      // Check if we have the required user information
      if (!user?.email) {
        console.error('No email provided by OAuth provider')
        return false
      }
      
      // For Microsoft/Azure AD, sometimes the name might be in profile
      const userName = user.name || (profile as any)?.name || user.email.split('@')[0]
      
      if (account?.provider) {
        try {
          console.log(`OAuth sign-in successful for ${account.provider}:`, user.email)
          // Update user.name if it was missing
          if (!user.name && userName !== user.name) {
            user.name = userName
          }
          return true
        } catch (error) {
          console.error('OAuth sign-in error:', error)
          return false
        }
      }
      
      return false
    },
    async jwt({ token, user, account }) {
      // Initial sign in
      if (account && user) {
        return {
          ...token,
          provider: account.provider,
          providerAccountId: account.providerAccountId,
          // Don't store access token to reduce cookie size
          // accessToken: account.access_token,
        }
      }
      return token
    },
    async session({ session, token }) {
      return {
        ...session,
        provider: token.provider as string | undefined,
        providerAccountId: token.providerAccountId as string | undefined,
        user: {
          ...session.user,
          id: token.sub,
        }
      }
    },
    async redirect({ url, baseUrl }) {
      console.log('Redirect callback:', { url, baseUrl })
      
      // For Azure AD callbacks, use an intermediate page to ensure proper redirect
      if (url.includes('api/auth/callback/azure-ad')) {
        return `${baseUrl}/auth/callback?callbackUrl=${encodeURIComponent('/teams')}`
      }
      
      // Allows relative callback URLs
      if (url.startsWith('/')) return `${baseUrl}${url}`
      // Allows callback URLs on the same origin
      else if (new URL(url).origin === baseUrl) return url
      // Default redirect to teams page after sign in
      return `${baseUrl}/teams`
    },
  },
  pages: {
    signIn: '/auth/login',
    error: '/auth/error',
  },
  session: {
    strategy: 'jwt',
    maxAge: 24 * 60 * 60, // 24 hours
  },
  cookies: {
    sessionToken: {
      name: `next-auth.session-token`,
      options: {
        httpOnly: true,
        sameSite: 'lax',
        path: '/',
        secure: process.env.NODE_ENV === 'production',
      },
    },
  },
}

const handler = NextAuth(authOptions)
export { handler as GET, handler as POST }
export const auth = () => getServerSession(authOptions)