import { NextRequest, NextResponse } from 'next/server'
import {
  getInstallationByKey,
  isSubdomainAvailable,
  validateSubdomain,
  reserveSubdomain,
  getUserById,
  getOrgMembers,
} from '@/lib/console/storage'
import { CLOUDFLARE_DNS_DOMAIN } from '@/lib/console/cloudflare-dns'

// Use Node.js runtime for Supabase
export const runtime = 'nodejs'

/**
 * Reserve subdomain for kli installation
 * This endpoint doesn't require session authentication - uses installation key
 *
 * POST /api/installations/reserve-domain-kli
 * Body: {
 *   installationKey: string,
 *   subdomain: string
 * }
 *
 * Response: {
 *   success: boolean,
 *   subdomain: string,
 *   url: string,  // https://subdomain.khost.dev
 *   error?: string
 * }
 */
export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { installationKey, subdomain } = body

    // Validate required fields
    if (!installationKey) {
      return NextResponse.json(
        { success: false, error: 'installationKey is required' },
        { status: 400 }
      )
    }

    if (!subdomain) {
      return NextResponse.json(
        { success: false, error: 'subdomain is required' },
        { status: 400 }
      )
    }

    // Validate subdomain format
    const validation = validateSubdomain(subdomain)
    if (!validation.valid) {
      const response = NextResponse.json({
        success: false,
        subdomain,
        error: validation.reason === 'reserved'
          ? 'This subdomain is reserved'
          : 'Invalid subdomain format (3-63 chars, alphanumeric and hyphens only)',
      })
      response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
      return response
    }

    // Get installation
    const installation = await getInstallationByKey(installationKey)
    if (!installation) {
      return NextResponse.json(
        { success: false, error: 'Invalid installation key' },
        { status: 404 }
      )
    }

    // Check if installation already has a subdomain
    if (installation.subdomain) {
      const response = NextResponse.json({
        success: true,
        subdomain: installation.subdomain,
        url: `https://${installation.subdomain}.${CLOUDFLARE_DNS_DOMAIN}`,
        message: 'Installation already has a subdomain assigned',
      })
      response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
      return response
    }

    // Check availability
    const available = await isSubdomainAvailable(subdomain)
    if (!available) {
      const response = NextResponse.json({
        success: false,
        subdomain,
        error: 'Subdomain is not available',
      })
      response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
      return response
    }

    // Always resolve user info from the org owner — never trust body-provided identity
    const members = await getOrgMembers(installation.orgId)
    const ownerMember = members.find((m) => m.role === 'owner')
    if (!ownerMember) {
      return NextResponse.json(
        { success: false, error: 'Organization owner not found' },
        { status: 404 }
      )
    }
    const user = await getUserById(ownerMember.userId)
    if (!user) {
      return NextResponse.json(
        { success: false, error: 'User not found' },
        { status: 404 }
      )
    }
    const reservationUserId = ownerMember.userId
    const reservationEmail = user.email
    const reservationName = user.name

    // Reserve the subdomain
    try {
      await reserveSubdomain(
        subdomain,
        installation.id,
        reservationUserId,
        reservationEmail,
        reservationName
      )
    } catch (error) {
      // Handle race condition - subdomain was taken between check and reserve
      if (error instanceof Error && error.message.includes('already reserved')) {
        const response = NextResponse.json({
          success: false,
          subdomain,
          error: 'Subdomain was just taken, please try another',
        })
        response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
        return response
      }
      throw error
    }

    const response = NextResponse.json({
      success: true,
      subdomain: subdomain.toLowerCase(),
      url: `https://${subdomain.toLowerCase()}.${CLOUDFLARE_DNS_DOMAIN}`,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Reserve domain error:', error)
    return NextResponse.json(
      { success: false, error: 'Internal server error' },
      { status: 500 }
    )
  }
}
