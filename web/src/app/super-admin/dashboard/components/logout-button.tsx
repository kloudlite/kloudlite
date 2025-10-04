'use client'

import { signOut } from 'next-auth/react'
import { Button } from '@/components/ui/button'

export function LogoutButton() {
  const handleLogout = async () => {
    await signOut({ callbackUrl: '/auth/signin' })
  }

  return (
    <Button
      onClick={handleLogout}
      variant="outline"
      className="h-9 px-4 text-sm"
    >
      Sign out
    </Button>
  )
}