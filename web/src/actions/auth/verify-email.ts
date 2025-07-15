'use server'

import { AuthResponse } from '@/lib/auth/types'
import { mockVerificationTokens, mockUsers } from './mock-data'
import { redirect } from 'next/navigation'

export async function verifyEmailAction(token: string): Promise<AuthResponse> {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 1000))

  // Validate token
  const tokenData = mockVerificationTokens.get(token)
  
  if (!tokenData) {
    return {
      success: false,
      error: 'Invalid or expired verification token'
    }
  }

  // Check if token is expired
  if (new Date() > tokenData.expiresAt) {
    mockVerificationTokens.delete(token)
    return {
      success: false,
      error: 'Verification token has expired'
    }
  }

  // Find user and verify
  const userIndex = mockUsers.findIndex(u => u.email === tokenData.email)
  if (userIndex === -1) {
    return {
      success: false,
      error: 'User not found'
    }
  }

  // Mark user as verified
  mockUsers[userIndex].verified = true
  mockUsers[userIndex].updatedAt = new Date()

  // Delete used token
  mockVerificationTokens.delete(token)

  // Redirect to login page with success message
  redirect('/auth/login?verified=success')
}

export async function resendVerificationEmailAction(email: string): Promise<AuthResponse> {
  try {
    // Simulate network delay
    await new Promise(resolve => setTimeout(resolve, 1000))

    // Find user
    const user = mockUsers.find(u => u.email.toLowerCase() === email.toLowerCase())
    
    if (!user) {
      // Don't reveal if user exists
      return {
        success: true,
        message: 'If an account exists with this email, we\'ve sent a new verification link.'
      }
    }

    if (user.verified) {
      return {
        success: false,
        error: 'This email is already verified'
      }
    }

    // Generate new token
    const token = Math.random().toString(36).substring(2) + Date.now().toString(36)
    const expiresAt = new Date()
    expiresAt.setHours(expiresAt.getHours() + 24) // Token expires in 24 hours
    
    // Store token
    mockVerificationTokens.set(token, { email: user.email, expiresAt })
    
    // In real app, send email with verification link
    console.log(`Verification link: /auth/verify-email?token=${token}`)

    return {
      success: true,
      message: 'Verification email sent! Please check your inbox.'
    }
  } catch (error) {
    console.error('Resend verification error:', error)
    return {
      success: false,
      error: 'An error occurred. Please try again.'
    }
  }
}