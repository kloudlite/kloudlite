import NextAuth from 'next-auth'

export const { handlers, auth, signIn, signOut } = NextAuth({
  secret: process.env.NEXTAUTH_SECRET || 'fallback-secret-for-development',
  providers: [],
  session: {
    strategy: 'jwt',
  },
})