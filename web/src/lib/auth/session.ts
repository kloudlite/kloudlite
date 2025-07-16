import { auth } from '@/auth'
import { User } from '@/lib/auth/types'
import { getCurrentUser } from '@/actions/auth/auth-server-actions'

export async function getServerSession(): Promise<{ user: User } | null> {
  const session = await auth()
  
  if (!session?.user?.id) {
    return null
  }

  try {
    const user = await getCurrentUser()
    if (!user) {
      return null
    }

    return {
      user: {
        id: user.id,
        email: user.email,
        name: user.name,
        verified: user.verified,
        createdAt: user.createdAt,
        updatedAt: user.updatedAt,
      }
    }
  } catch (error) {
    console.error('Error getting server session:', error)
    return null
  }
}