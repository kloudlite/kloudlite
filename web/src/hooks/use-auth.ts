'use client'

import { useSession } from 'next-auth/react'
import { User } from '@/lib/auth/types'

export function useAuth() {
  const { data: session, status } = useSession()

  const user: User | null = session?.user ? {
    id: session.user.id,
    email: session.user.email,
    name: session.user.name,
    verified: true, // OAuth users are considered verified
    createdAt: new Date(),
    updatedAt: new Date(),
  } : null

  return {
    user,
    isLoading: status === 'loading',
    isAuthenticated: status === 'authenticated',
    isUnauthenticated: status === 'unauthenticated',
  }
}