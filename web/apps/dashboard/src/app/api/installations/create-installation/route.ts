import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { createInstallation } from '@/lib/console/supabase-storage-service'
import { SignJWT } from 'jose'
import crypto from 'crypto'

export const runtime = 'nodejs'

/**
 * Create a new installation with name and description
 * This is called when user submits the name/description form
 */
export async function POST(request: Request) {
  try {
    const session = await getRegistrationSession()

    if (!session?.user) {
      return NextResponse.json({ error: 'Not authenticated' }, { status: 401 })
    }

    const body = await request.json()
    const { name, description } = body

    if (!name || typeof name !== 'string' || name.trim().length === 0) {
      return NextResponse.json({ error: 'Installation name is required' }, { status: 400 })
    }

    // Generate a new installation key
    const installationKey = crypto.randomUUID()

    // Create the installation
    const installation = await createInstallation(
      session.user.id,
      name.trim(),
      description?.trim() || undefined,
      installationKey,
    )

    // Update the session cookie with the installation key
    const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
    const token = await new SignJWT({
      provider: session.provider,
      email: session.user.email,
      name: session.user.name,
      image: session.user.image,
      installationKey: installation.installationKey,
      userId: session.user.id,
    })
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('30d')
      .sign(secret)

    const response = NextResponse.json({
      success: true,
      installationKey: installation.installationKey,
      installationId: installation.id,
    })

    // Update the session cookie
    response.cookies.set('registration_session', token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 30 * 24 * 60 * 60, // 30 days
    })

    return response
  } catch (error) {
    console.error('Error creating installation:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
