import { NextRequest, NextResponse } from 'next/server'
import { isSubdomainAvailable, validateSubdomain } from '@/lib/console/storage'

// Use Node.js runtime for Supabase
export const runtime = 'nodejs'

/**
 * Check if subdomain is available for kli installation
 * This endpoint doesn't require authentication - anyone can check availability
 *
 * GET /api/installations/check-domain-kli?subdomain=mycompany
 *
 * Response:
 * {
 *   available: boolean,
 *   subdomain: string,
 *   reason?: "reserved" | "invalid" | "taken"
 * }
 */
export async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams
    const subdomain = searchParams.get('subdomain')

    if (!subdomain) {
      return NextResponse.json(
        { error: 'Subdomain is required', available: false },
        { status: 400 }
      )
    }

    // Validate subdomain format
    const validation = validateSubdomain(subdomain)
    if (!validation.valid) {
      const response = NextResponse.json({
        available: false,
        subdomain,
        reason: validation.reason,
      })
      response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
      return response
    }

    // Check availability
    const available = await isSubdomainAvailable(subdomain)

    const response = NextResponse.json({
      available,
      subdomain,
      reason: available ? undefined : 'taken',
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Check domain error:', error)
    return NextResponse.json({ error: 'Internal server error', available: false }, { status: 500 })
  }
}
