import { NextRequest, NextResponse } from 'next/server'
import { cookies } from 'next/headers'
import { jwtVerify } from 'jose'
import { reserveSubdomain } from '@/lib/registration/storage-service'

/**
 * Reserve subdomain for user
 */
export async function POST(request: NextRequest) {
  try {
    // Verify registration session
    const cookieStore = await cookies()
    const token = cookieStore.get('registration_session')?.value

    if (!token) {
      return NextResponse.json(
        { error: 'Not authenticated' },
        { status: 401 }
      )
    }

    const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
    const { payload } = await jwtVerify(token, secret)

    const userEmail = payload.email as string
    const userName = payload.name as string
    const userId = payload.userId as string

    if (!userEmail) {
      return NextResponse.json(
        { error: 'Invalid session' },
        { status: 401 }
      )
    }

    // Get subdomain from request body
    const body = await request.json()
    const { subdomain } = body

    if (!subdomain) {
      return NextResponse.json(
        { error: 'Subdomain is required' },
        { status: 400 }
      )
    }

    // Reserve subdomain
    const reservation = await reserveSubdomain(
      subdomain,
      userId,
      userEmail,
      userName
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
  } catch (error: any) {
    console.error('Reserve subdomain error:', error)

    if (error.message === 'Subdomain is not available') {
      return NextResponse.json(
        { error: 'Subdomain is not available' },
        { status: 409 }
      )
    }

    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
