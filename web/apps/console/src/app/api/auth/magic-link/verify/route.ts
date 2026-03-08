/**
 * API endpoint to verify a magic link token
 * GET /api/auth/magic-link/verify?token=xxx
 */

import { NextRequest, NextResponse } from 'next/server'
import {
  verifyMagicLinkToken,
  markTokenAsUsed,
} from '@/lib/console/storage/magic-links'
import {
  getUserByEmail,
  saveUser,
} from '@/lib/console/storage/users'
import { createOrganization } from '@/lib/console/storage'
import type { User } from '@/lib/console/storage/types'
import { SignJWT } from 'jose'

const JWT_SECRET = new TextEncoder().encode(
  process.env.NEXTAUTH_SECRET || 'your-secret-key'
)

/**
 * Generate JWT session token
 */
async function generateSessionToken(user: User): Promise<string> {
  const token = await new SignJWT({
    userId: user.userId,
    email: user.email,
    name: user.name,
    provider: 'email',
    providers: user.providers,
  })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime('30d')
    .sign(JWT_SECRET)

  return token
}

/**
 * Extract default name from email
 */
function getDefaultName(email: string): string {
  const username = email.split('@')[0]
  // Capitalize first letter
  return username.charAt(0).toUpperCase() + username.slice(1)
}

export async function GET(request: NextRequest) {
  const baseUrl = process.env.NEXT_PUBLIC_CONSOLE_URL || 'http://localhost:3002'

  try {
    const { searchParams } = request.nextUrl
    const token = searchParams.get('token')

    if (!token) {
      return NextResponse.redirect(
        new URL('/login?error=invalid_link', baseUrl)
      )
    }

    // Verify the token
    const email = await verifyMagicLinkToken(token)

    if (!email) {
      // Token expired, used, or doesn't exist
      // Don't reveal which specific error for security
      return NextResponse.redirect(
        new URL('/login?error=invalid_link', baseUrl)
      )
    }

    // Mark token as used
    await markTokenAsUsed(token)

    // Get or create user
    let user = await getUserByEmail(email)

    if (!user) {
      // Create new user with email provider
      const userId = `email-${email}`
      const now = new Date().toISOString()

      user = {
        userId,
        email,
        name: getDefaultName(email),
        providers: ['email'],
        createdAt: now,
        updatedAt: now,
      }

      await saveUser(user)

      // Auto-create a default organization for new users
      try {
        const baseSlug = (user.name || user.email.split('@')[0])
          .toLowerCase()
          .replace(/[^a-z0-9-]/g, '-')
          .replace(/-+/g, '-')
          .replace(/^-|-$/g, '')
          .slice(0, 50)
        let slug = /^[a-z]/.test(baseSlug) ? baseSlug : `org-${baseSlug}`
        if (slug.length < 3) slug = `${slug}-org`

        await createOrganization(user.userId, `${user.name}'s Organization`, slug)
      } catch (orgError) {
        // Log but don't block login — org creation is best-effort on signup
        console.error('Failed to auto-create organization:', orgError)
      }
    } else {
      // Add 'email' to providers if not already present
      if (!user.providers.includes('email')) {
        user.providers.push('email')
        await saveUser(user)
      }
    }

    // Generate session token
    const sessionToken = await generateSessionToken(user)

    // Create response with redirect
    const response = NextResponse.redirect(
      new URL('/installations', baseUrl)
    )

    // Set session cookie
    response.cookies.set('registration_session', sessionToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 30 * 24 * 60 * 60, // 30 days
      path: '/',
    })

    return response
  } catch (error) {
    console.error('Magic link verification error:', error)
    return NextResponse.redirect(
      new URL('/login?error=server_error', baseUrl)
    )
  }
}
