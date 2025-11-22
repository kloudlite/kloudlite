import { auth } from '@/lib/auth'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { jwtVerify } from 'jose'
import type { Session } from 'next-auth'

/**
 * Console middleware
 * Handles tenant workspace management authentication and role-based access
 * This is for users inside their Kloudlite installation managing workspaces/workmachines
 */
export async function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl

  // Skip auth checks for auth pages, installation scripts, superadmin login, kltun install, and public assets
  if (
    pathname.startsWith('/auth') ||
    pathname.startsWith('/api') ||
    pathname.startsWith('/install') ||
    pathname.startsWith('/uninstall') ||
    pathname.startsWith('/kltun') ||
    pathname.startsWith('/superadmin-login') ||
    pathname.startsWith('/_next') ||
    pathname.includes('.')
  ) {
    return addSecurityHeaders(NextResponse.next(), req)
  }

  // Try to get session from NextAuth
  let session = await auth()
  let userRoles: string[] = []

  // If no NextAuth session, check for superadmin JWT token
  if (!session) {
    const cookieName =
      process.env.NODE_ENV === 'production'
        ? '__Secure-next-auth.session-token'
        : 'next-auth.session-token'

    const token = req.cookies.get(cookieName)?.value

    if (token) {
      try {
        const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
        const { payload } = await jwtVerify(token, secret)

        // Check if this is a superadmin token
        if (payload.provider === 'superadmin-login' && payload.roles) {
          // Create a mock session for superadmin
          userRoles = payload.roles as string[]
          session = {
            user: {
              email: payload.email as string,
              name: payload.name as string,
              roles: userRoles,
            },
            expires: new Date(payload.exp! * 1000).toISOString(),
          } as Session
        }
      } catch (error) {
        console.error('Failed to verify superadmin token:', error)
      }
    }
  }

  // Redirect to login if not authenticated
  if (!session) {
    return NextResponse.redirect(new URL('/auth/signin', req.url))
  }

  // Get user roles
  userRoles = session?.user?.roles || []
  const hasUserRole = userRoles.includes('user')
  const hasAdminRole = userRoles.includes('admin') || userRoles.includes('super-admin')

  // Role-based routing logic
  if (pathname.startsWith('/admin')) {
    // Admin section - only allow admin/super-admin access
    if (!hasAdminRole) {
      return NextResponse.redirect(new URL('/', req.url))
    }
  } else if (pathname === '/' || pathname.startsWith('/(main)')) {
    // Main dashboard section - redirect admin-only users to admin
    if (!hasUserRole && hasAdminRole) {
      return NextResponse.redirect(new URL('/admin', req.url))
    }
    // Require user role for main section
    if (!hasUserRole && !hasAdminRole) {
      return NextResponse.redirect(new URL('/auth/signin', req.url))
    }
  }

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
