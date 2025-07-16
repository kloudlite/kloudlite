'use server'

import { SignupCredentials, AuthResponse } from '@/lib/auth/types'
import { createUser, getUserByEmail, createMockSession, generateMockToken, mockVerificationTokens } from '@/actions/auth/mock-data'
import { setSessionCookie } from '@/actions/auth/session'
import { redirect } from 'next/navigation'

export async function signupAction(credentials: SignupCredentials): Promise<AuthResponse> {
  // Check if user already exists
  const existingUser = await getUserByEmail(credentials.email)
  if (existingUser) {
    return {
      success: false,
      error: 'An account with this email already exists'
    }
  }

  // Create new user
  const user = await createUser({
    name: credentials.name,
    email: credentials.email,
    password: credentials.password
  })

  // Generate verification token
  const verificationToken = generateMockToken()
  const expiresAt = new Date()
  expiresAt.setHours(expiresAt.getHours() + 24) // Token expires in 24 hours
  
  // Store verification token
  mockVerificationTokens.set(verificationToken, { email: user.email, expiresAt })
  
  // In real app, send verification email
  console.log(`Verification link: /auth/verify-email?token=${verificationToken}`)

  // For demo, we'll create a session but in real app, require email verification first
  const sessionId = createMockSession(user.id, false)
  await setSessionCookie(sessionId, false)

  // Redirect to email verification notice page
  redirect('/auth/verify-email?email=' + encodeURIComponent(user.email))
}