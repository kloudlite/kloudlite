import { NextRequest, NextResponse } from 'next/server'
import { cookies } from 'next/headers'
import { jwtVerify } from 'jose'
import {
  reserveSubdomain,
  getUserInstallations,
  createInstallation,
} from '@/lib/console/supabase-storage-service'

/**
 * Reserve subdomain for user's installation
 */
export async function POST(request: NextRequest) {
  try {
    // Verify registration session
    const cookieStore = await cookies()
    const token = cookieStore.get('registration_session')?.value

    if (!token) {
      return NextResponse.json({ error: 'Not authenticated' }, { status: 401 })
    }

    const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
    const { payload } = await jwtVerify(token, secret)

    const userEmail = payload.email as string
    const userName = payload.name as string
    const userId = payload.userId as string
    const sessionInstallationKey = payload.installationKey as string | undefined

    if (!userEmail || !userId) {
      return NextResponse.json({ error: 'Invalid session' }, { status: 401 })
    }

    // Get subdomain from request body
    const body = await request.json()
    const { subdomain } = body

    if (!subdomain) {
      return NextResponse.json({ error: 'Subdomain is required' }, { status: 400 })
    }

    // Determine which installation to use
    let installationId: string
    const installations = await getUserInstallations(userId)

    if (sessionInstallationKey) {
      // Use the installation from session
      const installation = installations.find((i) => i.installationKey === sessionInstallationKey)
      if (!installation) {
        return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
      }
      installationId = installation.id
    } else {
      // Create a new installation if none exists
      const incompleteInstallation = installations.find((i) => !i.hasCompletedInstallation)
      if (incompleteInstallation) {
        installationId = incompleteInstallation.id
      } else {
        // Create new installation
        const generatedKey = crypto.randomUUID()
        const newInstallation = await createInstallation(userId, 'My Installation', undefined, generatedKey)
        installationId = newInstallation.id
      }
    }

    // Reserve subdomain for the installation
    const reservation = await reserveSubdomain(
      subdomain,
      installationId,
      userId,
      userEmail,
      userName,
    )

    const response = NextResponse.json({
      success: true,
      subdomain: reservation.subdomain,
      url: `https://${reservation.subdomain}.kloudlite.io`,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (err) {
    console.error('Reserve subdomain error:', err)

    const error = err instanceof Error ? err : new Error('Unknown error')
    if (
      error.message === 'Subdomain is not available' ||
      error.message === 'Subdomain is already reserved'
    ) {
      return NextResponse.json({ error: 'Subdomain is not available' }, { status: 409 })
    }

    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
