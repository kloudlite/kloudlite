import NextAuth from 'next-auth'
import GitHub from 'next-auth/providers/github'
import Google from 'next-auth/providers/google'
import { Provider } from 'next-auth/providers'

// Microsoft Azure AD provider configuration
const AzureAD: Provider = {
  id: 'azure-ad',
  name: 'Microsoft',
  type: 'oidc',
  issuer: `https://login.microsoftonline.com/${process.env.MICROSOFT_ENTRA_TENANT_ID}/v2.0`,
  authorization: {
    url: `https://login.microsoftonline.com/${process.env.MICROSOFT_ENTRA_TENANT_ID}/oauth2/v2.0/authorize`,
    params: { scope: 'openid email profile' },
  },
  token: `https://login.microsoftonline.com/${process.env.MICROSOFT_ENTRA_TENANT_ID}/oauth2/v2.0/token`,
  userinfo: `https://graph.microsoft.com/oidc/userinfo`,
  clientId: process.env.MICROSOFT_ENTRA_CLIENT_ID,
  clientSecret: process.env.MICROSOFT_ENTRA_CLIENT_SECRET,
  profile(profile) {
    return {
      id: profile.sub,
      name: profile.name,
      email: profile.email,
      image: null,
    }
  },
}

export const { handlers, signIn, signOut, auth } = NextAuth({
  trustHost: true,
  providers: [
    GitHub({
      clientId: process.env.GITHUB_CLIENT_ID!,
      clientSecret: process.env.GITHUB_CLIENT_SECRET!,
    }),
    Google({
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
    }),
    AzureAD,
  ],
  callbacks: {
    async jwt({ token, user, account }) {
      // Add provider and user info to JWT token on first sign in
      if (account && user) {
        token.userId = user.id
        token.provider = account.provider
        token.email = user.email
        token.name = user.name
      }
      return token
    },
    async session({ session, token }) {
      // Add custom fields to session from JWT
      if (token && session.user) {
        session.user.id = token.userId as string
        session.user.provider = token.provider as string
      }
      return session
    },
    async redirect({ url, baseUrl }) {
      // After successful login, redirect to install page
      if (url.startsWith(baseUrl)) {
        return url
      }
      return `${baseUrl}/register/install`
    },
  },
  session: {
    strategy: 'jwt',
    maxAge: 30 * 24 * 60 * 60, // 30 days
  },
})
