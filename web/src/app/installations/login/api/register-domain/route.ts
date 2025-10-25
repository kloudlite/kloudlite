import { NextRequest, NextResponse } from 'next/server'
import { auth } from '@/lib/registration/auth-config'
import { reserveSubdomain, getUserByEmail } from '@/lib/registration/storage-service'

export async function POST(request: NextRequest) {
  try {
    // Check authentication
    const session = await auth()
    if (!session || !session.user || !session.user.email) {
      return NextResponse.json({ success: false, error: 'Unauthorized' }, { status: 401 })
    }

    const body = await request.json()
    const { subdomain } = body

    if (!subdomain) {
      return NextResponse.json({ success: false, error: 'Subdomain is required' }, { status: 400 })
    }

    // Validate subdomain format
    const subdomainRegex = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/
    if (!subdomainRegex.test(subdomain)) {
      return NextResponse.json(
        { success: false, error: 'Invalid subdomain format' },
        { status: 400 },
      )
    }

    // Check if user already has a domain
    const existingRegistration = await getUserByEmail(session.user.email)
    if (existingRegistration?.subdomain) {
      return NextResponse.json(
        {
          success: false,
          error: 'You already have a domain reserved',
          subdomain: existingRegistration.subdomain,
        },
        { status: 400 },
      )
    }

    // Reserve the subdomain
    // Use the existing user's userId from registration, or construct from email if needed
    const userId = existingRegistration?.userId || `auth-${session.user.email}`
    const reservation = await reserveSubdomain(
      subdomain,
      userId,
      session.user.email,
      session.user.name || session.user.email,
    )

    return NextResponse.json({
      success: true,
      message: 'Domain reserved successfully',
      reservation: {
        subdomain: reservation.subdomain,
        fullDomain: `${reservation.subdomain}.kloudlite.io`,
        reservedAt: reservation.reservedAt,
      },
    })
  } catch (error) {
    console.error('Error registering domain:', error)

    // Check if it's an availability error
    if (error instanceof Error && error.message.includes('not available')) {
      return NextResponse.json({ success: false, error: error.message }, { status: 409 })
    }

    return NextResponse.json(
      {
        success: false,
        error: 'Failed to register domain',
        details: error instanceof Error ? error.message : 'Unknown error',
      },
      { status: 500 },
    )
  }
}
