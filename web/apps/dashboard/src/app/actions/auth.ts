'use server'

import { signOut } from '@/lib/auth'
import { cookies } from 'next/headers'

export async function signOutAction() {
  // Explicitly delete all authentication cookies before signing out
  const cookieStore = await cookies()

  const cookiesToDelete = [
    'next-auth.session-token',           // Development
    '__Secure-next-auth.session-token',  // Production
    'registration_session',              // Registration flow
  ]

  // Get the cookie domain from environment
  const cookieDomain = process.env.NEXT_PUBLIC_AUTH_COOKIE_DOMAIN
    ? `.${process.env.NEXT_PUBLIC_AUTH_COOKIE_DOMAIN}`
    : undefined

  // Delete each cookie - both with and without domain to ensure cleanup
  cookiesToDelete.forEach((cookieName) => {
    // Delete cookie without domain (old cookies)
    try {
      cookieStore.delete({
        name: cookieName,
        path: '/',
      })
    } catch {
      // Ignore errors
    }

    // Delete cookie with domain (new cookies with cross-subdomain support)
    if (cookieDomain) {
      try {
        cookieStore.delete({
          name: cookieName,
          path: '/',
          domain: cookieDomain,
        })
      } catch {
        // Ignore errors
      }
    }
  })

  // Then call NextAuth's signOut
  await signOut({ redirectTo: '/auth/signin' })
}
