import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

/**
 * Console middleware
 * Handles security headers for installation registration app
 */
export async function middleware(req: NextRequest) {
  return addSecurityHeaders(NextResponse.next(), req)
}

/**
 * Add security headers to the response
 */
function addSecurityHeaders(response: NextResponse, req: NextRequest): NextResponse {
  const hostname = req.headers.get('host') || ''
  const baseDomain = process.env.CLOUDFLARE_DNS_DOMAIN || 'khost.dev'

  // Extract subdomain from hostname
  const hostParts = hostname.split('.')
  const baseParts = baseDomain.split('.')
  let subdomain: string | null = null

  if (hostParts.length > baseParts.length) {
    subdomain = hostParts[hostParts.length - baseParts.length - 1]
  }

  // Build VPN check URL if we have a subdomain
  const vpnCheckUrl = subdomain ? `https://vpn-check.${subdomain}.${baseDomain}` : ''
  const connectSrc = vpnCheckUrl
    ? `'self' http://localhost:* https://localhost:* ws://localhost:* wss://localhost:* ${vpnCheckUrl}`
    : `'self' http://localhost:* https://localhost:* ws://localhost:* wss://localhost:*`

  response.headers.set(
    'Content-Security-Policy',
    `script-src 'self' 'unsafe-inline' 'unsafe-eval'; connect-src ${connectSrc};`,
  )
  response.headers.set('Permissions-Policy', 'interest-cohort=(), browsing-topics=()')

  return response
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!_next/static|_next/image|favicon.ico).*)',
  ],
}
