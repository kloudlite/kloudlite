'use server'

import { cookies } from 'next/headers'
import { Session } from '@/lib/auth/types'
import { validateMockSession } from './mock-data'

const SESSION_COOKIE_NAME = 'kloudlite-session'

export async function getSession(): Promise<Session | null> {
  const cookieStore = await cookies()
  const sessionId = cookieStore.get(SESSION_COOKIE_NAME)?.value
  
  if (!sessionId) return null
  
  const user = validateMockSession(sessionId)
  if (!user) return null
  
  return {
    user,
    expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000), // 1 day from now
  }
}

export async function setSessionCookie(sessionId: string, rememberMe: boolean = false) {
  const cookieStore = await cookies()
  
  cookieStore.set(SESSION_COOKIE_NAME, sessionId, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    maxAge: rememberMe ? 30 * 24 * 60 * 60 : 24 * 60 * 60, // 30 days or 1 day
  })
}

export async function clearSessionCookie() {
  const cookieStore = await cookies()
  cookieStore.delete(SESSION_COOKIE_NAME)
}