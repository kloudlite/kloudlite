import 'next-auth'

declare module 'next-auth' {
  interface Session {
    user: {
      id: string
      email: string
      name: string
      image?: string
    }
  }

  interface User {
    id: string
    email: string
    name: string
    image?: string
  }
}

declare module 'next-auth/jwt' {
  interface JWT {
    sub: string
  }
}