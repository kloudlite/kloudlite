'use server'

import { ForgotPasswordData, AuthResponse } from '@/lib/auth/types'
import { getUserByEmail, generateMockToken, mockResetTokens } from './mock-data'

export async function forgotPasswordAction(data: ForgotPasswordData): Promise<AuthResponse> {
  try {
    // Simulate network delay
    await new Promise(resolve => setTimeout(resolve, 1000))

    // Check if user exists
    const user = await getUserByEmail(data.email)
    
    // Always return success to prevent email enumeration
    // In a real app, send email only if user exists
    if (user) {
      // Generate reset token
      const token = generateMockToken()
      const expiresAt = new Date()
      expiresAt.setHours(expiresAt.getHours() + 1) // Token expires in 1 hour
      
      // Store token (in real app, store in database)
      mockResetTokens.set(token, { email: data.email, expiresAt })
      
      // In real app, send email with reset link
      console.log(`Reset link: /auth/reset-password?token=${token}`)
    }

    return {
      success: true,
      message: 'If an account exists with this email, we\'ve sent a password reset link.'
    }
  } catch (error) {
    console.error('Forgot password error:', error)
    return {
      success: false,
      error: 'An error occurred. Please try again.'
    }
  }
}