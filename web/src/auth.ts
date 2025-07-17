import NextAuth from 'next-auth'

export const { handlers, auth, signIn, signOut } = NextAuth({
  secret: process.env.NEXTAUTH_SECRET,
  providers: [],
  session: {
    strategy: 'jwt',
  },
})