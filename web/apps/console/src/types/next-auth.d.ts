import { DefaultSession, DefaultUser } from 'next-auth'
import { DefaultJWT } from 'next-auth/jwt'

declare module 'next-auth' {
  interface Session {
    user: {
      id: string
      email: string
      name: string
      username?: string
      provider?: string
      roles?: string[]
      isActive?: boolean
    } & DefaultSession['user']
  }

  interface User extends DefaultUser {
    username?: string
    roles?: string[]
    isActive?: boolean
  }
}

declare module 'next-auth/jwt' {
  interface JWT extends DefaultJWT {
    username?: string
    provider?: string
    providerId?: string
    roles?: string[]
    isActive?: boolean
  }
}
