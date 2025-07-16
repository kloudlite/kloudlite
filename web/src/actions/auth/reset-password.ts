'use server'

import { ResetPasswordData, AuthResponse } from '@/lib/auth/types'
import { mockResetTokens, mockUsers } from './mock-data'
import { redirect } from 'next/navigation'

export async function resetPasswordAction(data: ResetPasswordData): Promise<AuthResponse> {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 1000))

  // Validate token
  const tokenData = mockResetTokens.get(data.token)
  
  if (!tokenData) {
    return {
      success: false,
      error: 'Invalid or expired reset token'
    }
  }

  // Check if token is expired
  if (new Date() > tokenData.expiresAt) {
    mockResetTokens.delete(data.token)
    return {
      success: false,
      error: 'Reset token has expired'
    }
  }

  // Find user and update password
  const userIndex = mockUsers.findIndex(u => u.email === tokenData.email)
  if (userIndex === -1) {
    return {
      success: false,
      error: 'User not found'
    }
  }

  // Update password (in real app, this would be hashed)
  mockUsers[userIndex].password = data.password

  // Delete used token
  mockResetTokens.delete(data.token)

  // Redirect to login page with success message
  redirect('/auth/login?reset=success')
}