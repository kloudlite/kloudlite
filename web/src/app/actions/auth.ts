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

  // Delete each cookie
  cookiesToDelete.forEach((cookieName) => {
    if (cookieStore.get(cookieName)) {
      cookieStore.delete({
        name: cookieName,
        path: '/',
      })
    }
  })

  // Then call NextAuth's signOut
  await signOut({ redirectTo: '/auth/signin' })
}
