'use server'

import { signOut } from '@/lib/auth'
import { redirect } from 'next/navigation'

export async function signOutAction() {
  // Sign out without redirect parameter (prevents external redirect)
  await signOut({ redirect: false })
  // Use Next.js redirect to ensure we stay on dashboard
  redirect('/auth/signin')
}
