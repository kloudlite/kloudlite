'use server'

import { SignupCredentials, AuthResponse } from '@/lib/auth/types'
import { createUser, getUserByEmail, createMockSession } from '@/actions/auth/mock-data'
import { setSessionCookie } from '@/actions/auth/session'
import { redirect } from 'next/navigation'

export async function signupAction(credentials: SignupCredentials): Promise<AuthResponse> {
  try {
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

    // Create session
    const sessionId = createMockSession(user.id, false) // false for rememberMe on signup
    
    // Set session cookie
    await setSessionCookie(sessionId, false)

    // Redirect to dashboard (or email verification page in the future)
    redirect('/dashboard')
  } catch (error) {
    console.error('Signup error:', error)
    return {
      success: false,
      error: 'An error occurred during signup. Please try again.'
    }
  }
}