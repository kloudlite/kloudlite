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

  const devEmail = 'karthik@kloudlite.io'
  const devName = 'Karthik'

  // Always check PII DB first — use existing identity if present
  const existingUser = await getUserByEmail(devEmail)

  let userId: string
  let name: string
  let email: string

  if (existingUser) {
    // Use all data from PII DB as source of truth
    userId = existingUser.userId
    name = existingUser.name
    email = existingUser.email

    // Ensure user has at least one org
    try {
      const orgs = await getUserOrganizations(userId)
      if (orgs.length === 0) {
        const baseSlug = (name || email.split('@')[0])
          .toLowerCase()
          .replace(/[^a-z0-9-]/g, '-')
          .replace(/-+/g, '-')
          .replace(/^-|-$/g, '')
          .slice(0, 50)
        let slug = /^[a-z]/.test(baseSlug) ? baseSlug : `org-${baseSlug}`
        if (slug.length < 3) slug = `${slug}-org`

        await createOrganization(userId, `${name}'s Organization`, slug)
      }
    } catch (orgError) {
      console.error('Failed to auto-create organization for dev user:', orgError)
    }
  } else {
    // No PII record — create one
    userId = 'dev-user-id'
    name = devName
    email = devEmail

    try {
      await saveUser({
        userId,
        email,
        name,
        providers: ['github'],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      })

      // Auto-create a default organization
      try {
        const baseSlug = (name || email.split('@')[0])
          .toLowerCase()
          .replace(/[^a-z0-9-]/g, '-')
          .replace(/-+/g, '-')
          .replace(/^-|-$/g, '')
          .slice(0, 50)
        let slug = /^[a-z]/.test(baseSlug) ? baseSlug : `org-${baseSlug}`
        if (slug.length < 3) slug = `${slug}-org`

        await createOrganization(userId, `${name}'s Organization`, slug)
      } catch (orgError) {
        console.error('Failed to auto-create organization:', orgError)
      }
    } catch (error) {
      console.error('Failed to create dev user:', error)
    }
  }

  const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)

  const token = await new SignJWT({
    userId,
    email,
    name,
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
