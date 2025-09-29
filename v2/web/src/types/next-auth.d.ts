import { DefaultSession, DefaultUser } from 'next-auth'
import { DefaultJWT } from 'next-auth/jwt'

declare module 'next-auth' {
  interface Session {
    user: {
      id: string
      email: string
      name: string
      provider?: string
      roles?: string[]
      backendToken?: string
      isActive?: boolean
    } & DefaultSession['user']
  }

  interface User extends DefaultUser {
    roles?: string[]
    backendToken?: string
    isActive?: boolean
  }
}

declare module 'next-auth/jwt' {
  interface JWT extends DefaultJWT {
    provider?: string
    providerId?: string
    roles?: string[]
    backendToken?: string
    isActive?: boolean
  }
}