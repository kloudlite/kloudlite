import { NextRequest, NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { isSubdomainAvailable } from '@/lib/console/storage'

/**
 * Check if subdomain is available
 */
export async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams
    const subdomain = searchParams.get('subdomain')

    if (!subdomain) {
      return apiError('Subdomain is required', 400)
    }

    const available = await isSubdomainAvailable(subdomain)

    const response = NextResponse.json({
      available,
      subdomain,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Check subdomain error:', error)
    return apiError('Internal server error', 500)
  }
}
