import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import {
  getUserInstallations,
  createInstallation,
} from '@/lib/console/supabase-storage-service'
import { SignJWT } from 'jose'
import crypto from 'crypto'

export const runtime = 'nodejs'

/**
 * Get or create an incomplete installation for the current user
 * This is called when user clicks "New Installation" button
 */
export async function POST() {
  try {
    const session = await getRegistrationSession()

    if (!session?.user) {
      return NextResponse.json({ error: 'Not authenticated' }, { status: 401 })
    }

    // Get user's installations
    const installations = await getUserInstallations(session.user.id)

    // Check if there's an incomplete installation
    const incompleteInstallation = installations.find((inst) => !inst.hasCompletedInstallation)

    let installationKey: string

    if (incompleteInstallation) {
      // Reuse existing incomplete installation
      installationKey = incompleteInstallation.installationKey
      console.log('Reusing incomplete installation:', installationKey)
    } else {
      // Create a new installation
      const generatedKey = crypto.randomUUID()
      const newInstallation = await createInstallation(
        session.user.id,
        'My Installation',
        undefined,
        generatedKey
      )
      installationKey = newInstallation.installationKey
      console.log('Created new installation:', installationKey)
    }

    // Update the session cookie with the installation key
    const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
    const token = await new SignJWT({
      provider: session.provider,
      email: session.user.email,
      name: session.user.name,
      image: session.user.image,
      installationKey: installationKey,
      userId: session.user.id,
    })
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('30d')
      .sign(secret)

    const response = NextResponse.json({
      success: true,
      installationKey,
      message: incompleteInstallation ? 'Reusing incomplete installation' : 'Created new installation',
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
    console.error('Error getting or creating installation:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
