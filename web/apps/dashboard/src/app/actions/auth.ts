'use server'

import { signOut } from '@/lib/auth'

export async function signOutAction() {
  // Let NextAuth handle all cookie cleanup
  await signOut({ redirectTo: '/auth/signin' })
}
