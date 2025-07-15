'use server'

import { AuthResponse, LoginCredentials } from '@/lib/auth/types'
import { mockUsers, createMockSession } from './mock-data'
import { setSessionCookie } from './session'

export async function loginAction(credentials: LoginCredentials): Promise<AuthResponse> {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 500))
  
  // Find user by email
  const user = mockUsers.find(u => u.email === credentials.email)
  
  if (!user) {
    return {
      success: false,
      error: 'Invalid email or password',
    }
  }
  
  // Check password (in real app, this would be hashed)
  if (user.password !== credentials.password) {
    return {
      success: false,
      error: 'Invalid email or password',
    }
  }
  
  // Check if user is verified
  if (!user.verified) {
    return {
      success: false,
      error: 'Please verify your email before logging in',
    }
  }
  
  // Create session
  const sessionId = createMockSession(user.id, credentials.rememberMe)
  
  // Set cookie
  await setSessionCookie(sessionId, credentials.rememberMe)
  
  // Return user without password
  const { password, ...userWithoutPassword } = user
  
  return {
    success: true,
    user: userWithoutPassword,
  }
}