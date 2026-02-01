import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import type { Session } from 'next-auth'
import { auth } from '@/lib/auth'

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
  // NextAuth v5 uses 'authjs' prefix by default
  if (!session) {
    const cookieName =
      process.env.NODE_ENV === 'production'
        ? '__Secure-authjs.session-token'
        : 'authjs.session-token'

    const token = req.cookies.get(cookieName)?.value

    if (token) {
      try {
        // Use NextAuth's decode function to match how the token was encoded
        const payload = await decode({
          token,
          secret: process.env.NEXTAUTH_SECRET!,
          salt: cookieName,
        })

        // Check if this is a superadmin token
        if (payload?.provider === 'superadmin-login' && payload?.roles) {
          // Create a mock session for superadmin
          userRoles = payload.roles as string[]
          session = {
            user: {
              email: payload.email as string,
              name: payload.name as string,
              roles: userRoles,
            },
            expires: new Date((payload.exp as number) * 1000).toISOString(),
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

  // Get user roles and provider
  userRoles = session?.user?.roles || []
  const sessionProvider = (session?.user as { provider?: string })?.provider
  const hasUserRole = userRoles.includes('user')
  const hasAdminRole = userRoles.includes('admin') || userRoles.includes('super-admin')
  const isSuperAdminLogin = sessionProvider === 'superadmin-login' || userRoles.includes('super-admin')

  // Role-based routing logic
  if (pathname.startsWith('/admin')) {
    // Admin section - only allow admin/super-admin access
    if (!hasAdminRole) {
      return NextResponse.redirect(new URL('/', req.url))
    }
  } else if (pathname === '/' || pathname.startsWith('/(main)')) {
    // Super-admin logins should always go to admin section
    if (isSuperAdminLogin && hasAdminRole) {
      return NextResponse.redirect(new URL('/admin', req.url))
    }
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

  response.headers.set(
    'Content-Security-Policy',
    `script-src 'self' 'unsafe-inline' 'unsafe-eval'; connect-src ${connectSrc};`,
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
