import { NextResponse } from 'next/server'
import { SignJWT } from 'jose'
import { cookies } from 'next/headers'

/**
 * Development-only backdoor login
 * WARNING: This should NEVER be deployed to production
 */
export async function GET() {
  // Only allow in development
  if (process.env.NODE_ENV === 'production') {
    return NextResponse.json({ error: 'Not available in production' }, { status: 403 })
  }

  const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)

  // Create JWT token for karthik@kloudlite.io
  const token = await new SignJWT({
    userId: 'dev-user-id',
    email: 'karthik@kloudlite.io',
    name: 'Karthik',
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
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 30, // 30 days
    path: '/',
  })

  // Redirect to installations page
  return NextResponse.redirect(new URL('/installations', process.env.NEXT_PUBLIC_BASE_URL || 'http://localhost:3002'))
}
