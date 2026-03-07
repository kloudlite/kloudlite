import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

/**
 * Console middleware
 * Handles security headers for the console app (billing, installation management)
 */
export async function proxy(req: NextRequest) {
  return addSecurityHeaders(NextResponse.next(), req)
}

/**
 * Add security headers to the response
 */
function addSecurityHeaders(response: NextResponse, _req: NextRequest): NextResponse {
  const scriptSrc = [
    "'self'",
    "'unsafe-inline'",
    ...(process.env.NODE_ENV === 'development' ? ["'unsafe-eval'"] : []),
    'https://challenges.cloudflare.com',
    'https://static.cloudflareinsights.com',
    'https://js.stripe.com',
  ].join(' ')

  response.headers.set(
    'Content-Security-Policy',
    [
      `script-src ${scriptSrc}`,
      `style-src 'self' 'unsafe-inline'`,
      `connect-src 'self' https://api.stripe.com https://challenges.cloudflare.com https://static.cloudflareinsights.com https://cloudflareinsights.com`,
      `frame-src 'self' https://js.stripe.com https://hooks.stripe.com`,
    ].join('; '),
  )

  return response
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico).*)',
  ],
}
