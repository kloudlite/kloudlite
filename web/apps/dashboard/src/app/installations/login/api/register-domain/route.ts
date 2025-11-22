import { NextRequest, NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { reserveSubdomain, getUserInstallations } from '@/lib/console/supabase-storage-service'

export async function POST(request: NextRequest) {
  try {
    // Check authentication
    const session = await getRegistrationSession()
    if (!session?.user) {
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

    // Check if user already has installations with reserved domains
    const installations = await getUserInstallations(session.user.id)
    const installationWithDomain = installations.find((i) => i.subdomain)

    if (installationWithDomain) {
      return NextResponse.json(
        {
          success: false,
          error: 'You already have a domain reserved',
          subdomain: installationWithDomain.subdomain,
        },
        { status: 400 },
      )
    }

    // Get an incomplete installation or use the first one
    const incompleteInstallation = installations.find((i) => !i.hasCompletedInstallation)
    const targetInstallation = incompleteInstallation || installations[0]

    if (!targetInstallation) {
      return NextResponse.json(
        { success: false, error: 'No installation found. Please create an installation first.' },
        { status: 400 },
      )
    }

    // Reserve the subdomain for the installation
    const reservation = await reserveSubdomain(
      subdomain,
      targetInstallation.id,
      session.user.id,
      session.user.email!,
      session.user.name || session.user.email!,
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
