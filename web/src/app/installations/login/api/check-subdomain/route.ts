import { NextRequest, NextResponse } from 'next/server'
import { isSubdomainAvailable } from '@/lib/registration/storage-service'

export async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams
    const subdomain = searchParams.get('subdomain')

    if (!subdomain) {
      return NextResponse.json(
        { success: false, error: 'Subdomain parameter is required' },
        { status: 400 },
      )
    }

    // Validate subdomain format
    const subdomainRegex = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/
    if (!subdomainRegex.test(subdomain)) {
      return NextResponse.json(
        {
          success: false,
          available: false,
          error: 'Invalid subdomain format',
        },
        { status: 400 },
      )
    }

    // Check length
    if (subdomain.length < 3 || subdomain.length > 63) {
      return NextResponse.json(
        {
          success: false,
          available: false,
          error: 'Subdomain must be between 3 and 63 characters',
        },
        { status: 400 },
      )
    }

    // Check availability
    const available = await isSubdomainAvailable(subdomain)

    return NextResponse.json({
      success: true,
      available,
      subdomain,
      message: available ? 'Subdomain is available' : 'Subdomain is already taken',
    })
  } catch (error) {
    console.error('Error checking subdomain:', error)
    return NextResponse.json(
      {
        success: false,
        error: 'Failed to check subdomain availability',
        details: error instanceof Error ? error.message : 'Unknown error',
      },
      { status: 500 },
    )
  }
}
