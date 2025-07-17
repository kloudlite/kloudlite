'use client'

import { signOut } from 'next-auth/react'
import { Button } from '@/components/ui/button'
import { logoutAction } from '@/actions/auth/logout'

interface LogoutButtonProps {
  children?: React.ReactNode
  className?: string
  asChild?: boolean
}

export function LogoutButton({ children, className, asChild = false }: LogoutButtonProps) {
  const handleLogout = async () => {
    try {
      await logoutAction()
      await signOut({ callbackUrl: '/auth/login' })
    } catch (error) {
      console.error('Logout error:', error)
    }
  }

  if (asChild) {
    return (
      <div onClick={handleLogout}>
        {children}
      </div>
    )
  }

  return (
    <Button onClick={handleLogout} className={className}>
      {children || 'Logout'}
    </Button>
  )
}