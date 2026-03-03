import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { SignJWT } from 'jose'
import { cookies } from 'next/headers'
import { saveUserRegistration, getUserByEmail } from '@/lib/console/storage'

export const runtime = 'nodejs'

/**
 * Development-only backdoor login
 * WARNING: This should NEVER be deployed to production
 */
export async function GET() {
  // Only allow in development
  if (process.env.NODE_ENV === 'production') {
    return apiError('Not available in production', 403)
  }

  const devUser = {
    userId: 'dev-user-id',
    email: 'karthik@kloudlite.io',
    name: 'Karthik',
  }

  // Ensure dev user exists in user_registrations table
  const existingUser = await getUserByEmail(devUser.email)
  console.log('Dev login - existing user:', existingUser)

  if (existingUser) {
    // User exists - use their existing userId instead
    console.log('Using existing user with userId:', existingUser.userId)
    devUser.userId = existingUser.userId
  } else {
    // Create new user
    console.log('Creating new dev user with userId:', devUser.userId)
    try {
      await saveUserRegistration({
        userId: devUser.userId,
        email: devUser.email,
        name: devUser.name,
        providers: ['github'],
        registeredAt: new Date().toISOString(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      })
      console.log('Dev user created successfully')
    } catch (error) {
      console.error('Failed to create dev user:', error)
    }
  }

  const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)

  // Create JWT token for karthik@kloudlite.io
  const token = await new SignJWT({
    userId: devUser.userId,
    email: devUser.email,
    name: devUser.name,
    image: undefined,
    provider: 'development',
    installationKey: undefined,
  })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime('30d')
    .sign(secret)

  // Set the cookie
  const cookieStore = await cookies()
  cookieStore.set('registration_session', token, {
    httpOnly: true,
    secure: process.env.NODE_ENV !== 'development' && process.env.NODE_ENV !== 'test',
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 30, // 30 days
    path: '/',
  })

  // Redirect to installations page
  return NextResponse.redirect(new URL('/installations', process.env.NEXT_PUBLIC_BASE_URL || 'http://localhost:3002'))
}
