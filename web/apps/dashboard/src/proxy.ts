import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

/**
 * Proxy for security headers only
 * Auth checks are handled in layouts (Server Components with Node.js runtime)
 */
export default async function proxy(req: NextRequest) {
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

  // Build CSP connect-src with allowed patterns
  // Allow VPN check for any tenant subdomain (vpn-check.X.khost.dev pattern)
  // Also get additional allowed subdomains from env var if set
  const allowedSubdomains = process.env.ALLOWED_TENANT_SUBDOMAINS?.split(',').map((s) => s.trim()) || []
  if (subdomain && !allowedSubdomains.includes(subdomain)) {
    allowedSubdomains.push(subdomain)
  }

  const connectSrcParts = [
    "'self'",
    'http://localhost:*',
    'https://localhost:*',
    'ws://localhost:*',
    'wss://localhost:*',
    // Allow single-level subdomains of base domain
    `https://*.${baseDomain}`,
    `wss://*.${baseDomain}`,
  ]

  // Add wildcard patterns for each known tenant subdomain (covers vpn-check.X.domain pattern)
  for (const sub of allowedSubdomains) {
    connectSrcParts.push(`https://*.${sub}.${baseDomain}`)
    connectSrcParts.push(`wss://*.${sub}.${baseDomain}`)
  }

  const connectSrc = connectSrcParts.join(' ')
  const scriptSrc = [
    "'self'",
    "'unsafe-inline'",
    ...(process.env.NODE_ENV === 'development' ? ["'unsafe-eval'"] : []),
  ].join(' ')

  response.headers.set(
    'Content-Security-Policy',
    `script-src ${scriptSrc}; connect-src ${connectSrc};`,
  )
  // Removed deprecated Permissions-Policy features (interest-cohort, browsing-topics)

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
