'use client'

import { Button } from '@/components/ui/button'
import { signIn, signOut } from 'next-auth/react'
import { Session } from 'next-auth'

interface AuthButtonProps {
  session: Session | null
}

export function AuthButton({ session }: AuthButtonProps) {
  if (session?.user) {
    return (
      <div className="flex items-center gap-4">
        <span className="text-sm">
          Signed in as {session.user.email || session.user.name}
        </span>
        <Button
          variant="outline"
          onClick={() => signOut()}
        >
          Sign Out
        </Button>
      </div>
    )
  }

  return (
    <Button
      variant="outline"
      onClick={() => signIn()}
    >
      Sign In
    </Button>
  )
}