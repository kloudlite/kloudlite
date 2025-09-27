import NextAuth from "next-auth"

declare module "next-auth" {
  interface Session {
    user: {
      id: string
      email?: string | null
      name?: string | null
      image?: string | null
      emailVerified: boolean
    }
    accessToken: string
    error?: string
  }

  interface User {
    id: string
    email: string
    name: string
    emailVerified: boolean
    accessToken: string
    refreshToken: string
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    id: string
    emailVerified: boolean
    accessToken: string
    refreshToken: string
    accessTokenExpires: number
    error?: string
  }
}