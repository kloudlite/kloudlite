import { auth } from '@/lib/auth'
import { cookies } from 'next/headers'
import { jwtVerify } from 'jose'
import type { Session } from 'next-auth'

/**
 * Get session from either NextAuth or superadmin JWT token
 * This function checks both regular NextAuth sessions and custom superadmin JWT tokens
 */
export async function getSession(): Promise<Session | null> {
  // Try NextAuth session first
  const session = await auth()
  if (session) {
    return session
  }

  // Check for superadmin JWT token
  const cookieStore = await cookies()
  const cookieName = process.env.NODE_ENV === 'production'
    ? '__Secure-next-auth.session-token'
    : 'next-auth.session-token'

  const token = cookieStore.get(cookieName)?.value

  if (token) {
    try {
      const secret = new TextEncoder().encode(process.env.JWT_SECRET)
      const { payload } = await jwtVerify(token, secret)

      // Check if this is a superadmin token
      if (payload.provider === 'superadmin-login' && payload.roles) {
        // Create a session object for superadmin
        return {
          user: {
            id: payload.sub as string,
            email: payload.email as string,
            name: payload.name as string,
            username: payload.username as string,
            roles: payload.roles as string[],
            isActive: payload.isActive as boolean,
          },
          expires: new Date(payload.exp! * 1000).toISOString(),
        } as Session
      }
    } catch (error) {
      console.error('Failed to verify superadmin token:', error)
    }
  }

  return null
}
