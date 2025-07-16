'use client'

import { signOut } from 'next-auth/react'
import { Button } from '@/components/ui/button'
import { logoutUser } from '@/actions/auth/auth-server-actions'

interface LogoutButtonProps {
  children?: React.ReactNode
  className?: string
}

export function LogoutButton({ children, className }: LogoutButtonProps) {
  const handleLogout = async () => {
    try {
      await logoutUser()
      await signOut({ callbackUrl: '/auth/login' })
    } catch (error) {
      console.error('Logout error:', error)
    }
  }

  return (
    <Button onClick={handleLogout} className={className}>
      {children || 'Logout'}
    </Button>
  )
}