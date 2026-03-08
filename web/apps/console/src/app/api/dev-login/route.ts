import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { SignJWT } from 'jose'
import { cookies } from 'next/headers'
import { saveUser, getUserByEmail, createOrganization, getUserOrganizations } from '@/lib/console/storage'

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

    // Check if existing user has orgs — if not, create one (handles pre-migration users)
    try {
      const orgs = await getUserOrganizations(existingUser.userId)
      if (orgs.length === 0) {
        const baseSlug = (existingUser.name || existingUser.email.split('@')[0])
          .toLowerCase()
          .replace(/[^a-z0-9-]/g, '-')
          .replace(/-+/g, '-')
          .replace(/^-|-$/g, '')
          .slice(0, 50)
        let slug = /^[a-z]/.test(baseSlug) ? baseSlug : `org-${baseSlug}`
        if (slug.length < 3) slug = `${slug}-org`

        await createOrganization(existingUser.userId, `${existingUser.name}'s Organization`, slug)
        console.log('Auto-created organization for existing dev user')
      }
    } catch (orgError) {
      console.error('Failed to auto-create organization for existing dev user:', orgError)
    }
  } else {
    // Create new user
    console.log('Creating new dev user with userId:', devUser.userId)
    try {
      await saveUser({
        userId: devUser.userId,
        email: devUser.email,
        name: devUser.name,
        providers: ['github'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      })
      console.log('Dev user created successfully')

      // Auto-create a default organization for new users
      try {
        const baseSlug = (devUser.name || devUser.email.split('@')[0])
          .toLowerCase()
          .replace(/[^a-z0-9-]/g, '-')
          .replace(/-+/g, '-')
          .replace(/^-|-$/g, '')
          .slice(0, 50)
        let slug = /^[a-z]/.test(baseSlug) ? baseSlug : `org-${baseSlug}`
        if (slug.length < 3) slug = `${slug}-org`

        await createOrganization(devUser.userId, `${devUser.name}'s Organization`, slug)
      } catch (orgError) {
        // Log but don't block login — org creation is best-effort on signup
        console.error('Failed to auto-create organization:', orgError)
      }
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
